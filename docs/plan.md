# Delivery plan

Living document. The method is in [`methodology.md`](methodology.md); this is
its execution record. Task status is updated **in the same commit as the work**,
never retrospectively in a batch — a plan updated after the fact is a report,
not a control.

**Status:** **phases 3–6 superseded** by
[`design/DESIGN-02-new-provider-spec.md`](design/DESIGN-02-new-provider-spec.md)
on 2026-07-28. Phases 1–2 and sprints 0–2 stand as the record of what was
done and what it taught; the remaining work moves to the new provider.
Phase 6's upstream contributions (R-02…R-07) survive unchanged — they are
the reason the fork still exists.
**Last updated:** 2026-07-28

## Legend

| Mark | Meaning |
|---|---|
| `[x]` | done, gate green, evidence in the commit body |
| `[~]` | in progress |
| `[ ]` | not started |
| `[!]` | blocked — the blocker is named in the row |
| `[-]` | dropped — the reason is named in the row |

Roles per [`methodology.md`](methodology.md) § Roles: **PM**, **ARC**
(architect), **DEV**, and where the work is test- or infrastructure-shaped,
**DEV** wearing the QA or OPS hat. The rule that matters: the author of a
resource does not write its acceptance test, and nobody approves their own ADR.

---

## Phase 1 — Audit · `[x]` closed 2026-07-28

Exit gate: baseline accepted, defect register agreed with the capability map.

| ID | Task | Role | Status |
|---|---|---|---|
| A-01 | Inventory the inherited code base: scale, layout, dependencies | ARC | `[x]` |
| A-02 | Structural findings S-01…S-06 | ARC | `[x]` |
| A-03 | Portability assessment for each inherited defect D-01…D-08 | PM | `[x]` |
| A-04 | Further findings A-01…A-05 not in the capability map | ARC | `[x]` |
| A-05 | Test-coverage blind spot | DEV/QA | `[x]` |

Output: [`audit/AUDIT-01-fork-baseline.md`](audit/AUDIT-01-fork-baseline.md).
All eight defects are portable, so the upstream contribution stream is not
blocked by the migration.

---

## Phase 2 — Design · `[x]` closed 2026-07-28

Exit gate: ADRs frozen; standards set published; public surface rules agreed.

| ID | Task | Role | Status |
|---|---|---|---|
| D-01 | ADR 0001 — delivery methodology | PM + ARC | `[x]` |
| D-02 | ADR 0002 — fork relationship, licence, version series | PM | `[x]` |
| D-03 | ADR 0003 — migrate by muxing, not rewriting | ARC | `[x]` |
| D-04 | ADR 0004 — Podman, OCI, Compose Specification | ARC | `[x]` |
| D-05 | ADR 0005 — two-backend acceptance matrix | ARC | `[x]` |
| D-06 | ADR 0006 — dnsdist out of scope pending a contract | PM + ARC | `[x]` |
| D-07 | Standards: naming, versioning, commits, changelog | PM | `[x]` |
| D-08 | Standards: Go 1.26, provider best practices, Terragrunt | ARC | `[x]` |
| D-09 | Standard: PowerDNS API discipline | ARC | `[x]` |
| D-10 | Standard: Python tooling — uv, ruff, ty | ARC | `[x]` |
| D-11 | ADR 0007 — Task replaces Make as the command interface | ARC | `[x]` |

---

## Phase 3 — Migration · `[~]` open

Exit gate: every resource served by the framework; unit and lab acceptance
green on every backend the resource supports; the mux server removed.

### Sprint 0 — Infrastructure · `[x]` closed 2026-07-28

Goal: the toolchain and the fixture the rest of the work depends on.

| ID | Task | Role | Depends | Status |
|---|---|---|---|---|
| S0-01 | `AGENTS.md` + `CODEX.md` + `CLAUDE.md` | PM | — | `[x]` |
| S0-02 | Dev container on `golang:1.26-trixie`, tools pinned by build arg | OPS | — | `[x]` |
| S0-03 | Compose Specification files: dev container and lab | OPS | S0-02 | `[x]` |
| S0-04 | Lab fixture, four services, two backends | OPS | S0-03 | `[x]` |
| S0-05 | `lab.py` on podman-py: up, down, status, verify | OPS | S0-04 | `[x]` |
| S0-06 | `golangci-lint` v2, 82 linters + 3 formatters | OPS | — | `[x]` |
| S0-07 | Python gate: uv, ruff, ty, `pyproject.toml` | OPS | S0-02 | `[x]` |
| S0-08 | CI: acceptance matrix across both backends, Actions pinned by SHA | OPS | S0-04 | `[x]` |
| S0-09 | Commit-message guard against AI attribution | OPS | — | `[x]` |
| S0-10 | Worktree helper | OPS | — | `[x]` |

Evidence: lab brought up from nothing and verified — gpgsql view write `422`,
LMDB `204`, recursor zone write `201`, seven tables in the PostgreSQL schema,
teardown to zero containers. Python gates clean under ruff 0.16.0 and ty 0.0.64.

### Sprint 1 — Module identity and dependencies · `[x]` closed 2026-07-28

Goal: the module is what it claims to be, on current dependencies, without
`vendor/`. No behaviour change.

| ID | Task | Role | Depends | Status |
|---|---|---|---|---|
| S1-01 | Module path → `github.com/dantte-lp/terraform-provider-powerdns` (S-01) | DEV | — | `[x]` |
| S1-02 | Go directive → 1.26.5; drop the stale `check.yml` pinning 1.26.1 (S-06) | DEV | S1-01 | `[x]` |
| S1-03 | Remove `vendor/` and `GOFLAGS=-mod=vendor` (S-03) | DEV | S1-01 | `[x]` |
| S1-04 | Dependencies to latest pinned; `go mod tidy` | DEV | S1-03 | `[x]` |
| S1-05 | Gate green on the new module path | DEV | S1-04 | `[x]` |
| S1-06 | `CHANGELOG.md` entry | PM | S1-05 | `[x]` |
| S1-07 | **Added mid-sprint.** Remove `GNUmakefile` — GNU make prefers it over `Makefile`, so the new one was dead on arrival | OPS | — | `[x]` |
| S1-08 | **Added mid-sprint.** Fix two reachable CVEs found by `govulncheck` | DEV | S1-04 | `[x]` |
| S1-09 | **Added mid-sprint.** Decide how to carry the inherited lint debt | ARC | S1-05 | `[x]` |
| S1-10 | **Added mid-sprint.** Replace `Makefile` with a Taskfile (ADR 0007) | OPS | S1-07 | `[x]` |
| S1-11 | **Added mid-sprint.** Enforce verified identifiers: pin checker, standard, skill | OPS | — | `[x]` |

Not a release: nothing user-visible changes. The import-path change is invisible
to a consumer because the provider is consumed as a binary.

Three things were discovered during the sprint rather than planned, and each is
recorded above rather than folded silently into an existing task.

**S1-07** was a defect in sprint 0's own work: `Makefile` was added while
upstream's `GNUmakefile` was still present, and GNU make resolves `GNUmakefile`
first. Every `make` target added in sprint 0 was unreachable until this landed.

**S1-08** — `govulncheck` reported two vulnerabilities reachable from this
code, not merely present in the dependency graph:

| Advisory | Module | Found | Fixed |
|---|---|---|---|
| GO-2026-6061 | `google.golang.org/grpc` | v1.79.3 | v1.82.1 |
| GO-2026-5970 | `golang.org/x/text` | v0.37.0 | v0.39.0 |

Both are indirect dependencies, which is why the direct-dependency check came
back clean and the vulnerability check did not. Bumped explicitly; the scan now
reports no vulnerabilities.

**S1-11** came from a defect in this project's own authoring, not in the
inherited code: a commit SHA for a GitHub Action was written from recall rather
than looked up, twice, on consecutive days. Both were caught by chance. The
response is a check rather than a resolution — `scripts/check-action-pins.sh`
runs as a pre-commit hook, inside `task all`, and as a CI job. On its first run
it found four floating tags in `release.yml`, which no one had touched and which
holds the signing key.

**S1-10** followed directly from S1-07. Having removed one make entry point
because two of them collide silently, keeping a `Makefile` alongside a
`Taskfile` would repeat the mistake in a different key — so the `Makefile`
is replaced rather than wrapped. ADR 0007 records what is given up.

**S1-09** — the full gate against the inherited `powerdns/` package produced
**729 findings, 275 of them error-tier**. Fixing them in this sprint was
rejected: it would mix a mechanical clean-up into an unrelated change, and it
would rewrite files that sprints 2-4 replace outright. The package is excluded
in `.golangci.yml` with that reasoning written at the exclusion. New code under
`internal/` is linted from its first line; the exclusion narrows as each
resource ports, and S4-10 deletes it. Its removal is the gate that proves the
migration is complete.

### Sprint 2 — Mux server and the first resource · `[~]` in progress

Goal: prove the migration mechanism end to end on one resource before
committing to it for ten.

| ID | Task | Role | Depends | Status |
|---|---|---|---|---|
| S2-01 | `main.go`: `tf5to6server.UpgradeServer` + `tf6muxserver` | DEV | S1-05 `[x]` | `[x]` |
| S2-02 | Manifest declares protocol 6.0 (S-02) | DEV | S2-01 | `[x]` |
| S2-03 | Layout: `internal/provider`, `internal/client`, `internal/resources/<area>` | ARC + DEV | S2-01 | `[x]` |
| S2-04 | Port `powerdns_zone` to the framework | DEV | S2-03 | `[x]` |
| S2-05 | Fix D-01 and D-02 in the ported resource — `net.SplitHostPort`, validator on both paths | DEV | S2-04 | `[x]` |
| S2-06 | State-continuity acceptance test: apply on SDKv2, refresh on framework | QA | S2-04 | `[!]` blocked — needs the released 2.3.0 provider as an external provider; see S2-11 |
| S2-07 | Acceptance on both backends | QA | S2-04 | `[x]` |
| S2-08 | Registry docs regenerated | DEV | S2-04 | `[ ]` |
| S2-09 | **Added mid-sprint.** Schema-parity test between the two mux halves | QA | S2-01 | `[x]` |
| S2-11 | **Added mid-sprint.** Inherited acceptance harness cannot run behind the mux — import cycle | ARC | — | `[!]` blocked on the client leaving `powerdns/` |
| S2-10 | **Added mid-sprint.** Semgrep in the gate; fix what it found | OPS | — | `[x]` |

`powerdns_zone` first because it is the resource every other one depends on and
the one carrying two open defects. If the mechanism is wrong, it is cheapest to
learn here.

The zone port surfaced an ordering constraint worth recording: `internal/client`
exists because a resource needs the client bundle and the provider needs the
resource, so holding the bundle in `internal/provider` is an import cycle. It
was cheaper to learn this on the first port than on the fifth — which is the
argument ADR 0003 makes for doing one resource before committing to ten.

**S2-11** is the sprint's most consequential finding and it changes the plan.

The inherited acceptance tests cannot be moved behind the mux: an in-package
test in `powerdns/` cannot import `internal/provider`, because that package
imports `powerdns` — a cycle Go rejects for in-package tests. Nine of those
tests declare a `powerdns_zone` in their configuration, which the SDKv2 half no
longer serves, so they cannot pass as written either.

The cycle exists because `internal/provider` reuses `powerdns.Config` and the
client types. That reuse looked like thrift and is now the constraint. It clears
only when the client moves out of `powerdns/` — which phase 4 already schedules
as C-02, "extend ZoneInfo to the full 24 attributes", and which DNSSEC and TSIG
require regardless.

Until then the acceptance job is scoped to `./internal/...`, and both the CI
workflow and the Taskfile say why at the point of exclusion rather than in a
commit message nobody will find.

**S2-10** added semgrep to the gate on the operator's suggestion. It is not a
second golangci-lint: it reasons across expressions and it scans `powerdns/`,
which the Go linter skips until migration. Two findings on the first run, both
real — a `tls.Config` with no `MinVersion` in the HTTP client, and a dynamic
URL reaching `urlopen` in the lab automation. The first is a genuine hardening
and is portable upstream; the second is suppressed with the reason recorded at
the call site, because the URLs are constants and adopting `requests` to
satisfy a pattern would be the worse trade.

**S2-09** was not planned and is the most useful thing in the sprint so far.
`terraform-plugin-mux` refuses to serve halves whose **provider** schemas
differ — and the comparison includes attribute descriptions, not just names and
types. Nothing in the build catches a drifting description; it surfaces at the
plugin handshake, as a Terraform error inside a consumer's `terraform plan`.

`TestMuxServer_ProviderSchemasMatch` fetches `GetProviderSchema` through the
mux and fails on any reported error. Constructing the server is not enough on
its own — `NewMuxServer` defers the comparison until the schemas are fetched.
The test was verified by deliberately drifting one description and confirming
it fails, then restoring it. A companion test names any type served by both
halves, which is the one-way rule from ADR 0003 made mechanical.

### Sprint 3 — Records and metadata · `[ ]`

| ID | Task | Role | Depends | Status |
|---|---|---|---|---|
| S3-01 | Port `powerdns_record` | DEV | S2-07 | `[ ]` |
| S3-02 | Port `powerdns_record_soa` | DEV | S3-01 | `[ ]` |
| S3-03 | Port `powerdns_ptr_record` | DEV | S3-01 | `[ ]` |
| S3-04 | Port `powerdns_reverse_zone` | DEV | S3-01 | `[ ]` |
| S3-05 | Port `powerdns_zone_metadata` | DEV | S2-07 | `[ ]` |
| S3-06 | Client: check `resp.StatusCode` before decoding, everywhere (D-08, A-03) | DEV | — | `[ ]` |
| S3-07 | Acceptance for each, both backends | QA | S3-01…S3-05 | `[ ]` |

### Sprint 4 — Backend-dependent resources and diagnostics · `[ ]`

Goal: the resources that only work on some installations say so, at the right
moment.

| ID | Task | Role | Depends | Status |
|---|---|---|---|---|
| S4-01 | Port `powerdns_view_zone_association` | DEV | S3-07 | `[ ]` |
| S4-02 | Port `powerdns_network` | DEV | S4-01 | `[ ]` |
| S4-03 | `422` from views/networks → diagnostic naming the LMDB requirement (D-03) | DEV | S4-02 | `[ ]` |
| S4-04 | Port `powerdns_recursor_forward_zone` | DEV | S3-07 | `[ ]` |
| S4-05 | Port `powerdns_recursor_config` with a validator on `name` (D-04) | DEV | S4-04 | `[ ]` |
| S4-06 | `422 api-config-dir` → diagnostic naming `webservice.api_dir` (D-05) | DEV | S4-05 | `[ ]` |
| S4-07 | Negative acceptance tests asserting each diagnostic, not merely the failure | QA | S4-03, S4-06 | `[ ]` |
| S4-08 | Document the `powerdns_network` destroy semantics (A-04) | DEV/DOC | S4-02 | `[ ]` |
| S4-09 | Port the six data sources | DEV | S3-07 | `[ ]` |
| S4-10 | Remove the mux server; drop `terraform-plugin-mux` | DEV | S4-09 | `[ ]` |

S4-10 closes phase 3. Until it lands, the provider is half SDKv2 by design, not
by neglect.

---

## Phase 4 — Coverage · `[ ]`

Exit gate: each new resource meets the per-resource Definition of Done.

| ID | Task | Role | Depends | Status |
|---|---|---|---|---|
| C-01 | Design review: cryptokey schema and private-key handling | ARC | S4-10 | `[ ]` |
| C-02 | Client: six cryptokey operations; extend `ZoneInfo` to the full 24 attributes (A-02) | DEV | C-01 | `[ ]` |
| C-03 | `powerdns_zone_cryptokey`; fix the `dnssec` JSON tag as it is first used (D-06, A-01) | DEV | C-02 | `[ ]` |
| C-04 | Zone attributes `dnssec`, `nsec3param`, `nsec3narrow`, `api_rectify`, `presigned` | DEV | C-02 | `[ ]` |
| C-05 | `powerdns_tsigkey`; zone `master_tsig_key_ids` / `slave_tsig_key_ids` | DEV | S4-10 | `[ ]` |
| C-06 | `powerdns_autoprimary` | DEV | S4-10 | `[ ]` |
| C-07 | Read-only data sources: statistics, search, server config, zone export | DEV | S4-10 | `[ ]` |
| C-08 | Decide the cache-versus-refresh interaction; triage PR #71 (A-05) | ARC | — | `[ ]` |

C-01 exists as its own task because the private key returned by
`GET /cryptokeys/{id}` lands in Terraform state in plain text unless a
write-only attribute or an ephemeral resource is used. That is a design
decision, not an implementation detail, and it is taken before code is written.

C-06 is the cheapest item in the phase — a flat object, create/read/delete, no
update — and is kept as relief work between the two large ones.

---

## Phase 5 — Verification · `[ ]`

| ID | Task | Role | Status |
|---|---|---|---|
| V-01 | Version matrix: acceptance against auth 5.0.x as well as 5.1.3 | QA | `[ ]` |
| V-02 | Drift: out-of-band change reported by `plan -refresh-only`, per resource | QA | `[ ]` |
| V-03 | Negative suite: conflict, not-found, transport failure | QA | `[ ]` |

---

## Phase 6 — Release and contribution · `[ ]`

| ID | Task | Role | Status |
|---|---|---|---|
| R-01 | Signed `v0.2.0` release | PM | `[ ]` |
| R-02 | Upstream PR: D-01 + D-02, IPv6 masters and update-path validation | DEV | `[ ]` |
| R-03 | Upstream PR: D-08 + A-03, status-code handling | DEV | `[ ]` |
| R-04 | Upstream PR: D-04, recursor config name validator | DEV | `[ ]` |
| R-05 | Upstream PR: D-03 + D-05, backend and `api_dir` diagnostics and docs | DEV | `[ ]` |
| R-06 | Upstream PR: D-06 + D-07, JSON tag and README | DEV | `[ ]` |
| R-07 | Issue to `PowerDNS/pdns`: the OpenAPI divergence | ARC | `[ ]` |

R-02…R-06 branch from `origin/main` and follow **upstream's** conventions, not
this project's — see [`release.md`](release.md). They do not wait for the
migration; all eight defects are portable.

R-07 is not a blocker for anything here. The project already treats the sources
as authoritative; reporting it is a courtesy to the next person.

---

## Risk register

| Risk | Effect | Response |
|---|---|---|
| State shape shifts during a resource port | Existing state breaks on upgrade | S2-06 state-continuity test on the first resource, repeated per port |
| Mux server rejects a duplicated type name | A half-migrated resource cannot ship | Migration is one-way per resource, in one commit (ADR 0003) |
| PowerDNS 5.2 ships during the work | Pinned claims go stale | Re-verify against the sources; `lab-verify` fails loudly on a version change |
| `ty` blocks more than it catches | Python gate becomes noise | Recorded per [`standards/python-tooling.md`](standards/python-tooling.md); removal would be an ADR |
| Upstream diverges under the contribution PRs | Cherry-picks stop applying | One defect per PR, opened early, rebased rather than batched |
| DNSSEC private key in plain-text state | Secret exposure | C-01 settles the mechanism before implementation |
| The `powerdns/` lint exclusion outlives the migration | 729 findings quietly permanent | S4-10 deletes the exclusion; it cannot be closed while the path is still listed |
| An exact identifier written from recall | Fails in a later pull request, blamed on infrastructure | `scripts/check-action-pins.sh` in three gates; the general rule in `standards/verified-identifiers.md` |

## How this document is maintained

- A task's status changes in the commit that does the work, not afterwards.
- A task that turns out to be wrong is marked `[-]` with the reason, not
  deleted — a plan that only ever contained things that worked is not evidence
  of anything.
- A new task discovered mid-sprint is added to the sprint it belongs to with a
  note of where it came from.
- Phase and sprint closures are recorded in `CHANGELOG.md`.
