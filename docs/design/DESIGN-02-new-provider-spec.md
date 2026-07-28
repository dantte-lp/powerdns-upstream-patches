# DESIGN-02 — Specification for a new PowerDNS provider

Supersedes the migration plan in `plan.md` phases 3–6. The decision taken on
2026-07-28 is to build a new provider rather than finish migrating the
inherited one.

## 1. Why this is now the right call

`DESIGN-01` recommended finishing the strangler. That recommendation rested on
one thing: state continuity for users of the published `mmianl/powerdns`.

The counter-argument that carries: **we are not those users' maintainer.**
Upstream continues to exist and continues to serve them. Our obligation to them
is discharged by contributing the eight portable defect fixes back — which does
not require us to carry their architecture.

With that removed, every remaining advantage of the strangler evaporates:

| Claimed advantage | Reality after sprint 2 |
|---|---|
| Reuse the inherited client | It caused the import cycle (S2-11), covers 18 of 42 operations and 12 of 24 zone attributes, and phase 4 already scheduled its replacement |
| Keep 4 188 lines of tests | They cannot run behind the mux; the harness needs rewriting either way |
| Ship throughout | Shipping a half-migrated provider nobody uses yet is not a benefit |
| Smaller risk steps | The steps are smaller but there are more of them, and each fights the legacy |

## 2. Identity

### 2.1 The registry constrains this more than it appears

The Terraform Registry requires the repository to be named
`terraform-provider-{TYPE}`, and publishes it as `{github-namespace}/{TYPE}`.
`{TYPE}` becomes the identifier in every line of consumer HCL:

```hcl
provider "TYPE" {}
resource "TYPE_zone" "example" {}
provider::TYPE::fqdn("www.example.com")
```

So the "project name" is not free-form: it is the type, and the type is the
prefix a user types dozens of times.

### 2.2 The field

| Registry address | Stars | Plugin API | Resources | Last push |
|---|---:|---|---:|---|
| `pan-net/powerdns` | 53 | SDKv2 | 8 | 2024-05-15, 28 open issues |
| `mmianl/powerdns` | 9 | SDKv2 | 10 | active — our upstream |
| `gonzolino/powerdns` | 5 | framework + SDKv2 | 2 | active |
| `joelMuehlena/pdns` | 2 | framework | 2 | active |
| `mikelaws/pdns`, `pyama86/powerdns`, others | 0 | — | — | dormant |

Nobody covers DNSSEC, TSIG, autoprimaries, views or networks. Nobody uses
actions, ephemeral resources or provider functions. The most complete provider
in the field is the SDKv2 one we forked.

### 2.3 Recommendation: type `powerdns`, namespace `dantte-lp`

`registry.terraform.io/dantte-lp/powerdns`, repository
`dantte-lp/terraform-provider-powerdns`.

Reasoning, stated plainly because the instinct is to invent something:

- **A provider is named for the platform it manages.** That is HashiCorp's own
  design principle and it is what users search for. A distinctive brand name in
  the type field costs discoverability and gains nothing — a user looking for
  PowerDNS support will not search for a codename.
- **`pdns` is rejected**, despite being PowerDNS's own abbreviation and shorter
  in HCL, because `joelMuehlena/pdns` is actively developed and framework-native.
  Two similar providers sharing a type across namespaces is a trap for users.
- **Namespace does the disambiguation.** Seven `powerdns` types already coexist;
  that is what namespaces are for.

### 2.4 The name collision to resolve first

`dantte-lp/terraform-provider-powerdns` is currently the **fork**. GitHub does
not allow two repositories with one name under one account.

Proposal: the fork's only remaining job is carrying the eight defect fixes to
upstream, so rename it to what it now is —
`dantte-lp/powerdns-upstream-patches` — and create the new provider at the
freed name. GitHub redirects the old URLs, so the open pull requests survive.

## 3. Scope

**In scope:** PowerDNS Authoritative Server and PowerDNS Recursor, over their
HTTP APIs.

**Out of scope**, each for a stated reason:

| Excluded | Reason |
|---|---|
| dnsdist | Its API writes exactly two things — `PUT /config/allow-from` and `DELETE /api/v1/cache`. Rules, pools and downstreams are Lua or YAML. ADR 0006 stands |
| `pdnsutil` and other non-API management | A provider talks to an API |
| Zone content import from BIND files | A one-off migration task, not desired state |

## 4. Target surface

Mapped against the 42 authoritative and 16 recursor operations from the
capability map. The inherited provider covers 18 and 5.

### 4.1 Resources — 11

| Resource | API object | Notes |
|---|---|---|
| `powerdns_zone` | zone | Includes the DNSSEC and TSIG attributes the inherited struct omits |
| `powerdns_record` | RRset via `PATCH /zones/{id}` | Comments and `disabled` supported |
| `powerdns_zone_metadata` | metadata | |
| `powerdns_tsigkey` | TSIG key | Secret is write-only or generated and exposed ephemerally |
| `powerdns_zone_cryptokey` | cryptokey | Private key never enters state — §6.2 |
| `powerdns_autoprimary` | autoprimary | Flat; create/read/delete, no update |
| `powerdns_view_zone` | view membership | **LMDB only**, enforced with a diagnostic |
| `powerdns_network` | network → view | **LMDB only**; destroy semantics documented — the API has no DELETE |
| `powerdns_recursor_zone` | recursor forward zone | Requires `webservice.api_dir` |
| `powerdns_recursor_acl` | `allow-from`, `allow-notify-from` | The only two writable recursor settings; the type is closed, not free-form |
| `powerdns_record_soa` | SOA RRset | Kept as a distinct type: SOA has structure the generic record does not |

**Two inherited resources are deliberately dropped**, replaced by functions
(§4.4): `powerdns_reverse_zone` and `powerdns_ptr_record` exist largely to
compute a name. A zone is a zone; the arithmetic belongs in a function that can
be used anywhere, including outside this provider's resources.

### 4.2 Data sources — 13

`powerdns_zone`, `powerdns_zones`, `powerdns_record`, `powerdns_zone_metadata`,
`powerdns_zone_export`, `powerdns_tsigkey`, `powerdns_tsigkeys`,
`powerdns_cryptokeys`, `powerdns_autoprimaries`, `powerdns_views`,
`powerdns_networks`, `powerdns_server`, `powerdns_statistics`,
`powerdns_search`.

`powerdns_search` and `powerdns_statistics` are what make brownfield adoption
practical, and nothing in the field exposes them.

### 4.3 Actions — 5

Terraform 1.14 and later. These are the operations `CM-04` §3 wrongly wrote off
as not fitting the model.

| Action | Operation |
|---|---|
| `powerdns_notify_zone` | `PUT /zones/{id}/notify` |
| `powerdns_axfr_retrieve` | `PUT /zones/{id}/axfr-retrieve` — uses `SendProgress`, it can be slow |
| `powerdns_rectify_zone` | `PUT /zones/{id}/rectify` |
| `powerdns_flush_cache` | `PUT /cache/flush` |
| `powerdns_recursor_flush_cache` | recursor `PUT /cache/flush` |

`api_rectify` also exists as a zone attribute; the action is for the one-off
case, the attribute for the standing policy.

### 4.4 Provider functions — 5

Pure, offline, no client.

| Function | Example |
|---|---|
| `fqdn(name)` | `www.example.com` → `www.example.com.` |
| `is_fqdn(name)` | validation in a `precondition` |
| `reverse_zone_name(cidr)` | `192.0.2.0/24` → `2.0.192.in-addr.arpa.` |
| `ptr_name(ip)` | `192.0.2.15` → `15.2.0.192.in-addr.arpa.` |
| `soa_serial(date, counter)` | `2026-07-28`, `1` → `2026072801` |

IPv6 is handled by the same two address functions, which is where the inherited
`ip.go` already spends 333 lines.

### 4.5 Ephemeral resources — 2

`powerdns_tsigkey_secret` and `powerdns_cryptokey_material`: key material that
exists for the duration of a run and is never written to state.

## 5. Coverage target

| Component | Operations | Target | Inherited |
|---|---:|---:|---:|
| Authoritative | 42 | 38 | 18 |
| Recursor | 16 | 7 | 5 |
| dnsdist | 10 | 0 | 0 |

The four authoritative operations not targeted are `GET /error`, the two
discovery endpoints, and `GET /config/{name}`, which is documented but not
implemented by the server.

## 6. Architecture

### 6.1 Layout

```text
internal/
  client/pdns/        transport and one file per API domain
    client.go         HTTP, retry policy, status examined before body
    errors.go         typed errors; backend-capability classification
    zone.go rrset.go metadata.go cryptokey.go tsigkey.go
    autoprimary.go view.go network.go server.go search.go
  client/rec/         recursor client — a different API, not a mode of the same one
  provider/           schema, Configure, registration
  resources/<object>/ one package per API object
  datasources/<object>/
  actions/
  ephemeral/
  functions/
  testutil/           lab wiring shared by acceptance suites
```

Three rules, each earned from a defect in the inherited code:

1. **Nothing outside `internal/client/*` constructs an HTTP request.** The
   inconsistent status handling (`A-03`) exists because request construction
   leaked into resources.
2. **The client classifies backend capability.** "Views need LMDB",
   "recursor writes need `api_dir`" are properties of the server, discovered
   from a `422`. One classifier in `errors.go`; every resource gets the same
   diagnostic. Per-resource handling guarantees drift (`D-03`, `D-05`).
3. **Authoritative and recursor are separate clients.** They are separate
   products with separate APIs; the inherited code puts both in one 1 529-line
   file and it is the least readable part of it.

### 6.2 Secrets

Terraform state is not encrypted. The rule is therefore mechanical rather than
a judgement call:

| Value | Mechanism |
|---|---|
| TSIG secret supplied by the user | write-only attribute |
| TSIG secret generated by the server | ephemeral resource |
| DNSSEC private key | ephemeral resource; never an attribute of `powerdns_zone_cryptokey` |
| Provider API key | `Sensitive`, from environment |

This closes the plan's sharpest open risk without a design debate.

### 6.3 Errors and retry

- Status code examined before the body is decoded. Always.
- Transient transport failures and `5xx` retry with exponential backoff, five
  attempts, 1–16 s. `4xx` fails fast: `404`, `409` and `422` are answers.
- `401`/`403` fail fast with a hint naming the `api_key` argument.
- A `422` from views, networks or a recursor write is translated into a
  diagnostic naming the requirement, not relayed as "unprocessable entity".

### 6.4 Version floor

Terraform **1.11** for the provider (write-only attributes), with actions gated
at **1.14** through client-capability reporting so an older CLI loses the
actions rather than the provider. OpenTofu 1.12 is a co-equal target.

## 7. What carries over from the fork

| Carried | Discarded |
|---|---|
| The whole standards set under `docs/standards/` | `powerdns/` — the SDKv2 package |
| `AGENTS.md`, `CODEX.md`, `CLAUDE.md` | The mux server and `terraform-plugin-mux` |
| ADRs 0001, 0002, 0004, 0005, 0006, 0007 | ADR 0003 — superseded, kept as a record |
| `docs/plan.md` structure and discipline | Phases 3–6 as written |
| The four-service lab and `lab.py` | |
| Taskfile, golangci-lint config, semgrep, pin checker, AI-attribution guard | |
| `internal/provider`, `internal/client`, `internal/resources/zone` | |
| The capability map, as a cited sibling | |

The zone resource carries over intact, including the three normalisation fixes
the acceptance suite found. It is roughly a day of work that does not need
redoing.

## 8. Should the fork's changes be published?

Three answers, because they are three different artefacts.

**Yes — the eight defect fixes, to upstream.** They fix real problems for real
users of `mmianl/powerdns`, including an open issue with a named reporter
(#73), and a missing TLS floor. Five pull requests, per `release.md`. This is
the obligation that made state continuity look important, and discharging it
directly is better than carrying an architecture to serve it indirectly.

**No — the fork as a provider.** Its value was the migration path. Publishing a
half-migrated provider would advertise something we are not going to finish.
Rename it to `powerdns-upstream-patches`, leave the branches for the PRs, and
say in its README what it is and where the work went.

**Yes — the capability map, unchanged.** It stands alone, is already accurate,
and is the evidence base the new provider cites. One correction is owed:
`CM-04` §3 on actions.

## 9. Milestones

| # | Deliverable | Gate |
|---|---|---|
| M0 | Repository, standards carried over, lab, CI | `task all` green on an empty provider |
| M1 | `internal/client/pdns` complete, unit-tested against recorded responses | All 38 targeted authoritative operations implemented |
| M2 | `powerdns_zone`, `powerdns_record`, `powerdns_zone_metadata` + data sources | Acceptance on both backends; parity with the inherited provider's useful half |
| M3 | Functions and the two dropped resources' replacements | `reverse_zone_name` and `ptr_name` verified against the inherited 333-line implementation |
| M4 | DNSSEC and TSIG, with ephemeral key material | Signed zone verified with `dig +dnssec` |
| M5 | Autoprimaries, views, networks, with backend diagnostics | Negative tests assert the diagnostic, not just the failure |
| M6 | Actions | Verified against Terraform 1.14+ and gated below it |
| M7 | Recursor | `powerdns_recursor_zone`, `powerdns_recursor_acl` |
| M8 | Release 0.1.0, registry submission | Signed release, registry listing live |

M1 before any resource: the client being second-class is what produced every
structural problem in the inherited provider.

## 10. Open questions

1. **Namespace.** `dantte-lp` or a new organisation? An organisation reads as
   more maintained and allows co-maintainers; a personal namespace is one less
   thing to administer. Not blocking until M8.
2. **`powerdns_record` versus `powerdns_recordset`.** PowerDNS's object is an
   RRset. `gonzolino` calls it `recordset`, upstream calls it `record`. The API
   name is `rrsets`. Decide at M2; it is a breaking name to change later.
3. **Whether `powerdns_record_soa` survives.** It exists upstream because SOA
   has structure. If the generic record handles it cleanly, one fewer type.
4. **Catalog zones** as a first-class resource versus an attribute. Currently an
   attribute; PowerDNS models producer and consumer zones as zone kinds.
