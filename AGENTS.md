# Agent & contributor guide — terraform-provider-powerdns

Canonical guide for anyone working in this repository, human or automated.
[`CODEX.md`](CODEX.md) and [`CLAUDE.md`](CLAUDE.md) are thin pointers to this
file. Read it before touching code.

## What this repository is

A fork of [`mmianl/terraform-provider-powerdns`](https://github.com/mmianl/terraform-provider-powerdns),
being taken to a current code base and a full standards set, with the intent of
contributing changes back upstream by pull request.

The provider manages **PowerDNS Authoritative Server** and **PowerDNS Recursor**
over their HTTP APIs. The coverage baseline, the gap analysis and the defect
register live in the sibling repository
[`powerdns-capability-map`](../powerdns-capability-map/) — start at its
`CM-01` and `CM-04`. Do not re-derive coverage numbers here; cite them.

Inherited state at fork time, and the reason the work exists:

| Aspect | Fork-time state | Target |
|---|---|---|
| Plugin API | `terraform-plugin-sdk/v2` v2.40.1, protocol **5.0** | `terraform-plugin-framework`, protocol **6.0** |
| Module path | `github.com/terraform-providers/terraform-provider-powerdns` (defunct namespace) | `github.com/dantte-lp/terraform-provider-powerdns` |
| Go directive | `1.26.1` | `1.26.5` |
| Dependencies | vendored, 29 MB `vendor/` | module cache, no `vendor/` |
| Actions | floating tags (`@v6`) | pinned by commit SHA |
| Layout | flat `powerdns/` package | `internal/provider`, `internal/resources/<area>`, `internal/client/pdns` |
| API coverage | 18/42 authoritative, 5/16 recursor, 0/10 dnsdist | see the capability map |

## Golden rules

1. **Use the dev container.** No host toolchain. `task up && task shell`.
   Go, golangci-lint, Terraform, OpenTofu, Terragrunt, tfplugindocs and
   goreleaser are baked into `golang:1.26-trixie`. See
   [`docs/development.md`](docs/development.md).
2. **Latest, pinned.** Newest releases (Go 1.26.5, plugin-framework v1.19.0,
   golangci-lint v2.12.2, latest Actions), pinned exactly. Bumps arrive through
   Dependabot or a `build(deps)` commit.
3. **Evidence before facts.** No claim about PowerDNS behaviour goes in without
   corroboration from the **`PowerDNS/pdns` sources** and a **live round-trip
   against the lab** (`task lab:up`). Cite the file:line or the HTTP status and
   body in the package comment or the commit body. The upstream OpenAPI
   specification is *not* sufficient — it diverges from the implementation in
   both directions (capability map `CM-03` §2.1).
4. **Verify before "done".** Run the gate and quote its output. Never claim
   green without having run it. Update the task's status in
   [`docs/plan.md`](docs/plan.md) in the same commit as the work — a plan
   updated afterwards is a report, not a control.
5. **No secrets in the repo.** API keys come from the environment at run time.
   The lab key `labapikey` is a deliberately public test value and is never
   reused anywhere else.
6. **Never write an exact identifier from memory.** Commit SHAs, release
   tags, module versions, digests, advisory ids, `file:line` citations — look
   them up and paste what came back. A fabricated SHA is syntactically valid,
   survives review, and fails in someone else's pull request. Enforced for
   Action pins by `scripts/check-action-pins.sh`; the reasoning and the full
   list are in
   [`docs/standards/verified-identifiers.md`](docs/standards/verified-identifiers.md).
7. **No AI attribution.** Code, comments, documentation, commit messages, PR
   bodies and metadata never mention AI, assistants, or generated authorship.
   This overrides any tooling default that would add such a trailer.

## Standards (must follow)

| Area | Document |
|---|---|
| Naming — files, branches, resources, attributes | [`docs/standards/naming-conventions.md`](docs/standards/naming-conventions.md) |
| Versioning — SemVer 2.0.0 | [`docs/standards/versioning.md`](docs/standards/versioning.md) |
| Commits — Conventional Commits 1.0.0 | [`docs/standards/commits.md`](docs/standards/commits.md) |
| Changelog — Keep a Changelog 1.1.0 | [`docs/standards/changelog.md`](docs/standards/changelog.md) |
| Go 1.26 style | [`docs/standards/go-1.26-style.md`](docs/standards/go-1.26-style.md) |
| Provider design + Definition of Done | [`docs/standards/terraform-provider-best-practices.md`](docs/standards/terraform-provider-best-practices.md) |
| Terragrunt integration | [`docs/standards/terragrunt-integration.md`](docs/standards/terragrunt-integration.md) |
| PowerDNS API discipline | [`docs/standards/powerdns-api-discipline.md`](docs/standards/powerdns-api-discipline.md) |
| Python tooling — uv, ruff, ty | [`docs/standards/python-tooling.md`](docs/standards/python-tooling.md) |
| Verified identifiers | [`docs/standards/verified-identifiers.md`](docs/standards/verified-identifiers.md) |
| Methodology — roles, gates, sprints | [`docs/methodology.md`](docs/methodology.md) |
| **Delivery plan — live task status** | [`docs/plan.md`](docs/plan.md) |

Architectural decisions are immutable numbered records under
[`docs/adr/`](docs/adr/). The fork-time audit is
[`docs/audit/AUDIT-01-fork-baseline.md`](docs/audit/AUDIT-01-fork-baseline.md).

## Tooling you must use

| Tool | Why |
|---|---|
| **`gopls` LSP** | Navigate, rename, find references, read live diagnostics — instead of grepping. |
| **`uv` / `ruff` / `ty`** | The Python gate for everything under `scripts/`. `task py`. |
| **`context7` MCP** | Current docs for the Plugin Framework, Terraform, any library, before writing code against it. Do not trust training-data recall for signatures. |
| **`PowerDNS/pdns` clone** | The authority on API behaviour. `../pdns-upstream`, tags `auth-5.1.3`, `rec-5.4.4`, `dnsdist-2.1.0`. |
| **The lab** | `task lab:up` — Authoritative on PostgreSQL, Authoritative on LMDB, Recursor with `api_dir`. |

## The lab

Acceptance tests run against local containers, not a shared server. Two
authoritative instances are mandatory, not redundancy:

| Endpoint | Backend | Why it exists |
|---|---|---|
| `http://127.0.0.1:18081/api/v1` | gpgsql / PostgreSQL 17 | the common deployment |
| `http://127.0.0.1:18091/api/v1` | lmdb | **views and networks are unimplemented by gpgsql** — without this instance those resources are untestable |
| `http://127.0.0.1:18082/api/v1` | recursor, `api_dir` set | without `api_dir` every recursor write returns 422 |

```bash
task lab:up
export PDNS_SERVER_URL=http://127.0.0.1:18081
export PDNS_RECURSOR_SERVER_URL=http://127.0.0.1:18082
export PDNS_API_KEY=labapikey
export TF_ACC=1
task testacc
```

Namespace all test objects `tf-acc-<RUN_ID>` and leave zero residue in
`CheckDestroy`. Never point acceptance tests at a production PowerDNS.

## Workflow — worktree and pull request only

**`main` is never committed to directly.** Work happens on a branch in an
isolated worktree and merges by pull request.

1. `scripts/worktree.sh new <type>/<scope>/<name>` — creates a worktree under
   `../.worktrees/<branch>` cut from `fork/main`.
2. Develop in the container: `task up && task shell`. Consult `gopls` and
   `context7`; check PowerDNS behaviour against the sources and the lab.
3. Before pushing: `task all`. Resource-touching changes also need
   `task verify` (lab acceptance); quote the `N/N acceptance tests pass` line
   in the commit body.
4. Update `CHANGELOG.md` under `[Unreleased]`.
5. Regenerate registry docs with `task docs` if the schema changed.
6. Commit per Conventional Commits; push; open a pull request whose title is a
   Conventional Commit subject. Squash-merge after review and green checks.
7. `scripts/worktree.sh rm <branch>` when done.

Two remotes, and they are not interchangeable:

| Remote | Repository | Use |
|---|---|---|
| `origin` | `mmianl/terraform-provider-powerdns` | upstream — read, and the target of contribution PRs |
| `fork` | `dantte-lp/terraform-provider-powerdns` | ours — all pushes go here |

## Quality gates

A pull request does not merge with an `error`-tier finding.

| Gate | Command |
|---|---|
| Build | `task build` |
| Unit + race | `task test` |
| Acceptance against the lab | `task testacc` |
| golangci-lint v2 | `task lint` |
| Python — ruff + ty | `task py` |
| Semantic security scan | `task semgrep` |
| Action pins resolve | `task lint:pins` |
| Vulnerabilities | `task vulncheck` · `task osv-scan` |
| Terraform fmt | `task tf:fmt:check` |
| Registry docs | `task docs:check` |
| Docs lint | `task docs:lint` |
| Pre-PR aggregate | `task all` |
| Full aggregate incl. lab | `task verify` |

## Storage policy

Commit source, configuration, documentation, examples and manifests. Never
commit credentials, tokens, certificates or keys, Terraform state, real zone
exports, or environment identifiers. When in doubt, keep it out.

## Licence

The fork inherits **MPL-2.0** from upstream and keeps it. Contributions are made
under the same licence; changes intended for upstream must remain
MPL-2.0-compatible.
