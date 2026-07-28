# DESIGN-01 — From-scratch versus rewrite, and the 2026 target architecture

Written 2026-07-28, at the point where sprint 2 hit a structural obstacle
(`plan.md` S2-11) and the question of restarting became a fair one.

## 1. The comparison is narrower than it looks

Porting a resource from `terraform-plugin-sdk/v2` to
`terraform-plugin-framework` **is a rewrite of that resource**. The schema
types, the CRUD signatures, the value system and the diff model share no code.
There is no incremental refactor of an SDKv2 resource into a framework one;
`powerdns_zone` was not refactored, it was written again.

So the implementation cost is the same on both paths:

| Work | From scratch | Rewrite in place |
|---|---:|---:|
| 9 remaining resources | ~3 200 lines | ~3 200 lines |
| 6 data sources | ~900 lines | ~900 lines |
| Client | ~1 000 lines | ~1 000 lines — the inherited one covers 18 of 42 operations and 12 of 24 zone attributes, and DNSSEC and TSIG need it extended regardless |

What differs is only the **transition strategy**. That is the whole decision.

## 2. Where the two paths actually diverge

| Axis | From scratch | Rewrite in place (current) |
|---|---|---|
| **State continuity** | A published `mmianl/powerdns` user upgrading gets no path; state must be re-imported or the upgrade is breaking | The mux moves one resource at a time, each with a state-continuity test and a state upgrader where the shape changes |
| **Upstream contribution** | No shared ancestry — cherry-picks are impossible, so the 8 portable defects never reach upstream users | Preserved; `origin/main` is still a branch point |
| **Inherited tests** | 4 188 lines discarded | Kept — but **not** as intact as it looked: they cannot run behind the mux (S2-11), so part of that value is already lost |
| **Time to first shippable** | Nothing ships until full parity | Shipping throughout |
| **Lint debt** | Zero from line one | 729 findings quarantined behind an exclusion that shrinks per port |
| **Risk shape** | One large cutover | Many small ones, each independently verified |
| **Legacy coupling** | None | Real, and biting — the import cycle in S2-11 exists precisely because `internal/provider` reused `powerdns.Config` |

## 3. Recommendation

**Neither, in pure form. Keep the mux; rewrite the client now; treat the legacy
package as a shell that is deleted, not maintained.**

The reasoning:

- The **one thing** a from-scratch rewrite genuinely sacrifices is state
  continuity, and it is the one thing a user feels. Fifteen versions are
  published; a silent state break is not an acceptable upgrade story.
- The **one thing** the rewrite-in-place path is currently paying for is reuse
  of the legacy client — and that reuse is now the blocker, not the saving. It
  created the import cycle, it caps coverage at 18 of 42 operations, and phase 4
  already schedules its replacement as `C-02`.
- Therefore: bring `C-02` forward. Rewriting the client is not extra work; it is
  scheduled work moved to where it unblocks rather than where it is convenient.

This is the strangler pattern applied honestly: the new implementation grows
inside the old shell, and the shell is measured by how fast it shrinks. The
gate that proves it finished already exists — `S4-10`, the deletion of the
`powerdns/` lint exclusion.

## 4. A conclusion of mine that is now wrong

The capability map's `CM-04` §3 sorts the 24 uncovered operations into three
groups and says of four of them — `notify`, `axfr-retrieve`, `rectify`,
`cache/flush` — that they are "imperative by nature and do not fit Terraform's
declarative model", to be left outside the provider.

**That was correct when written against SDKv2 and is no longer correct.**
Terraform 1.14 introduced **actions**, and `terraform-plugin-framework`
implements them (`ProviderWithActions`, `action.Action` with
`Schema`/`Metadata`/`Invoke`, plus `SendProgress` for long-running work).

An action is exactly the shape of those four operations: an invocable unit with
a config schema, diagnostics and progress reporting, which does not own state.
`powerdns_notify_zone` and `powerdns_rectify_zone` are natural actions, and
`SendProgress` suits an AXFR retrieval that may take a while.

`CM-04` and `CM-05` need revising on this point. Recording the error here rather
than quietly editing them is the standard this project set for itself.

## 5. The 2026 stack

### 5.1 Language and runtime

| Choice | Version | Why |
|---|---|---|
| Go | 1.26.5 | `new(expr)` for the pointer-heavy schema defaults; `errors.AsType[T]` for typed API errors; Green Tea GC; `reflect` iterators for schema tooling |
| Module layout | `internal/` throughout | Nothing in this provider is a library. Anything outside `internal/` is a public API by accident |
| No vendoring | — | `go.sum` is the reproducibility mechanism; vendoring predates it |

### 5.2 Provider framework — what 2026 actually offers

This is where the design changes most, and the inherited provider uses none of
it because SDKv2 cannot.

| Capability | Minimum Terraform | Use here |
|---|---|---|
| **Actions** | 1.14 | `notify`, `axfr_retrieve`, `rectify`, `cache_flush` — the four operations previously written off |
| **Resource identity** (`ResourceWithIdentity`) | 1.12 | A zone's identity is its canonical name and never changes; import by identity rather than by a guessed ID string |
| **Ephemeral resources** | 1.10 | TSIG secret and DNSSEC private key material that exists for one run and must never reach state |
| **Write-only attributes** | 1.11 | Secrets that are supplied but need no drift detection; the framework nullifies them before every response |
| **Provider-defined functions** | 1.8 | Pure, offline computation — see §5.3 |
| **Protocol 6** | 1.0 | Nested attributes, and the only protocol the framework speaks natively |

**On the DNSSEC private key**, which `C-01` was scheduled to agonise over: the
answer is now mechanical rather than a judgement call. The key is an
**ephemeral resource** or a **write-only attribute**, and it does not go into
state at all. That removes the sharpest open risk in the plan.

### 5.3 Provider functions absorb work currently smeared into resources

Two inherited resources exist partly to *compute a name*:

- `powerdns_reverse_zone` derives `2.0.192.in-addr.arpa.` from `192.0.2.0/24`
- `powerdns_ptr_record` derives a PTR owner name from an IP

That is pure, offline computation with no remote object — the exact definition
of a provider function, and a case where HashiCorp's own design principle
("anything needing the network is a data source, not a function") points the
other way for once: these need nothing.

```hcl
locals {
  rev = provider::powerdns::reverse_zone_name("192.0.2.0/24")   # 2.0.192.in-addr.arpa.
  ptr = provider::powerdns::ptr_name("192.0.2.15")              # 15.2.0.192.in-addr.arpa.
  fqdn = provider::powerdns::fqdn("www.example.com")            # www.example.com.
}
```

The resources remain for the objects they create; the arithmetic stops being a
resource's side effect.

### 5.4 Engines and orchestration

| Tool | Version | Position |
|---|---|---|
| Terraform | 1.15.8 | Primary target; actions need ≥ 1.14 |
| OpenTofu | 1.12.5 | Co-equal target, tested in CI. State encryption and provider mocking in tests are genuine advantages for a provider that handles key material |
| Terragrunt | 1.1.1 | The consumer-side orchestration; units and stacks, `run`, stack dependencies |

Declaring a minimum of Terraform 1.14 is a real cost — it excludes older
installations — and it buys actions. The proposal is to declare **1.11 as the
floor** for the provider generally (write-only attributes) and to gate the four
actions on 1.14 through the framework's own client-capability reporting, so an
older CLI loses the actions rather than the provider.

### 5.5 Target layout

```text
internal/
  client/pdns/            transport, typed errors, one file per API domain
    client.go             HTTP, retry policy, status-before-body
    error.go              typed errors; backend-capability classification
    zone.go rrset.go metadata.go cryptokey.go tsigkey.go
    autoprimary.go view.go network.go server.go recursor.go
  provider/               provider schema, Configure, registration
  resources/<area>/       one package per API object
  datasources/<area>/
  ephemeral/              tsigkey, cryptokey — key material that never persists
  actions/                notify, axfr_retrieve, rectify, cache_flush
  functions/              fqdn, reverse_zone_name, ptr_name
```

Two structural rules earned the hard way:

**The client owns capability classification.** `D-03` and `D-05` — views and
networks needing LMDB, recursor writes needing `api_dir` — are properties of the
server, discovered from a `422`. Classifying them in `internal/client/pdns/error.go`
means one implementation of "this backend cannot do that", and every resource
gets the same diagnostic. Doing it per-resource guarantees drift.

**Nothing outside `internal/client/pdns` builds an HTTP request.** The reason
the current code has inconsistent status handling (`A-03`) is that request
construction leaked. One chokepoint makes "check the status before decoding the
body" enforceable rather than aspirational.

### 5.6 Testing

| Layer | Tool | Scope |
|---|---|---|
| Unit | stdlib, `-race` | schema, validators, plan modifiers, payload marshalling |
| Acceptance | `terraform-plugin-testing` v1.16 | plan/apply/refresh/destroy on **both backends**, `plancheck.ExpectEmptyPlan`, `ImportStateVerify` |
| Behaviour | `dig`, `psql` | a `200` proves the API accepted a request, not that DNS answers |
| Semantic | semgrep | across expressions, and over the packages golangci-lint skips |

Sprint 2 settled the argument for the acceptance layer: three server-side
normalisations — `soa_edit_api` defaulting, `kind` title-casing, IPv6 zero
compression — were invisible to every unit test and would each have shipped as
a permanent diff.

## 6. What this implies for the plan

| Change | Effect |
|---|---|
| Bring `C-02` (client rewrite) forward into phase 3, before further resource ports | Clears the S2-11 import cycle; unblocks the inherited harness; lifts the 18/42 operation ceiling |
| Add actions for the four imperative operations | Revises `CM-04` §3 and `CM-05`; four operations move from "out of scope" to scheduled |
| Add provider functions for the pure computations | New, small, no state, no migration risk |
| Settle DNSSEC key handling as ephemeral/write-only | Closes the plan's sharpest open risk without a design debate |
| Declare Terraform 1.11 as the floor, gate actions at 1.14 | Costs older CLIs the actions, not the provider |

None of this argues for starting over. It argues for finishing the strangler in
the right order — client first, because that is what is actually in the way.
