# Changelog

All notable changes to this project are documented here.

The format follows [Keep a Changelog 1.1.0](https://keepachangelog.com/en/1.1.0/)
and this project adheres to [Semantic Versioning 2.0.0](https://semver.org/spec/v2.0.0.html).

This is a fork of [`mmianl/terraform-provider-powerdns`](https://github.com/mmianl/terraform-provider-powerdns),
taken at commit `0dac0e7` (release `v2.3.0`). Upstream history is not reproduced
here; versioning restarts at `0.1.0` because the plugin protocol and the module
path both change — see [ADR 0002](docs/adr/0002-fork-and-upstream-relationship.md).

## [Unreleased]

### Added

- Standards set under `docs/standards/`: naming conventions, versioning,
  commits, changelog, Go 1.26 style, Terraform provider best practices,
  Terragrunt integration, and PowerDNS API discipline.
- `AGENTS.md` as the canonical contributor guide, with `CODEX.md` and
  `CLAUDE.md` as pointers.
- Development methodology (`docs/methodology.md`) and six architectural
  decision records under `docs/adr/`.
- Fork baseline audit (`docs/audit/AUDIT-01-fork-baseline.md`) recording the
  inherited defects, structural findings and test blind spot.
- Containerised development on `golang:1.26-trixie` with the toolchain pinned
  by build argument; Compose Specification files for the dev container and the
  acceptance lab.
- Acceptance lab with **two** authoritative backends (PostgreSQL and LMDB) plus
  a recursor with `api_dir` set, driven through podman-py
  (`scripts/automation/lab.py`). Views and networks are unimplemented by the
  PostgreSQL backend, so a single-backend fixture cannot cover this provider.
- `golangci-lint` v2 configuration with an explicit allowlist of 82 linters,
  3 formatters, and severity tiers.
- Python toolchain for the automation scripts: `uv` 0.11.33 as environment
  manager, `ruff` 0.16.0 as linter and formatter, `ty` 0.0.64 as type checker,
  configured in `pyproject.toml` and gated by `task py` as part of `task all`.
- `docs/standards/python-tooling.md` recording the Python rules, including why
  the ruff selection is an allowlist rather than `ALL` and how a pre-1.0 type
  checker is treated in a merge gate.
- `docs/plan.md` — the live delivery plan. Task status changes in the same
  commit as the work; phase and sprint closures are recorded here.
- CI workflow with an acceptance matrix across both backends and a Python
  lint job; all GitHub Actions pinned by commit SHA.
- Task ([taskfile.dev](https://taskfile.dev)) v3.52.0 replaces `Makefile` as
  the command interface, with namespaced tasks (`lab:up`, `py:typecheck`,
  `tf:fmt:check`) and real incremental builds through `sources`/`generates`.
  Rationale in [ADR 0007](docs/adr/0007-taskfile-over-make.md); the immediate
  prompt was the `GNUmakefile` shadowing incident. `task --list` is the index.

- `scripts/check-action-pins.sh` — verifies that every GitHub Action pinned by
  commit SHA resolves upstream, and rejects floating tags. Runs as a pre-commit
  hook, as `task lint:pins` inside `task all`, and as its own CI job. Written
  after two fabricated SHAs reached the repository within a day; a fabricated
  SHA is syntactically valid and fails only in a later pull request.
- `docs/standards/verified-identifiers.md` — the general rule the hook enforces:
  an identifier that must be exact is looked up, never recalled. Covers SHAs,
  release tags, module versions, digests, advisory ids and `file:line`
  citations.

- Taskfile guards: every task that shells into the dev container now carries a
  `preconditions` check that names the fix (`run: task up`) instead of failing
  with a podman-compose stack trace; `testacc` additionally requires the lab.

- The provider is served through `tf6muxserver` with the SDKv2 half lifted to
  protocol 6.0 by `tf5to6server`, so resources can move to
  `terraform-plugin-framework` one at a time while the provider stays
  shippable (ADR 0003). `internal/provider` holds the framework half.
- `TestMuxServer_ProviderSchemasMatch` — the mux rejects halves whose provider
  schemas differ, descriptions included, and that failure otherwise appears at
  the plugin handshake inside a consumer's `terraform plan`. Verified by
  drifting a description on purpose and confirming the test fails.

- Semgrep in the gate — `task semgrep`, inside `task all`, and a CI job.
  Rulesets `p/golang`, `p/security-audit`, `p/secrets`; 310 rules. It
  complements golangci-lint rather than duplicating it, and reaches
  `powerdns/`, which the Go linter deliberately skips until migration.

### Changed

- Module path is `github.com/dantte-lp/terraform-provider-powerdns`. The
  inherited path named the retired `terraform-providers` GitHub organisation
  and resolved to nothing. Invisible to consumers — the provider is used as a
  binary, not imported.
- Go directive raised to 1.26.5.
- `terraform-registry-manifest.json` declares protocol **6.0** instead of
  5.0. Both halves of the mux speak 6.0.

### Removed

- `vendor/` — 29 MB of vendored dependencies, along with
  `GOFLAGS=-mod=vendor`. `go.sum` provides the reproducibility guarantee
  vendoring was there for.
- `GNUmakefile`, superseded by `Makefile`. Both were present after the previous
  change, and GNU make resolves `GNUmakefile` first — so every target added
  then was unreachable until this removal. `Makefile` itself was subsequently
  replaced by `Taskfile.yml`; see ADR 0007.
- `.github/workflows/check.yml`, superseded by `ci.yml`. It pinned Go 1.26.1 in
  four places, which is how a toolchain bump would have silently failed to take
  effect.

### Security

- The PowerDNS HTTP client sets `MinVersion: tls.VersionTLS12` explicitly.
  Found by semgrep in `powerdns/config.go`, which built a `tls.Config` with no
  floor at all. Go's own default is already 1.2, so this states the intent and
  prevents a future default change from lowering it. Not 1.3: the API is
  frequently published through a front end that has not moved, and failing to
  reach the server is worse than the marginal difference.

- All four GitHub Actions in `release.yml` are now pinned by commit SHA instead
  of a floating tag. Found by the new pin check on its first run, in a workflow
  nobody had edited — and the one holding the GPG signing key.

- `google.golang.org/grpc` to v1.82.1 (GO-2026-6061) and `golang.org/x/text` to
  v0.39.0 (GO-2026-5970). Both were reachable from this code, not merely
  present in the dependency graph; `govulncheck` now reports none.

### Fixed

- The inherited `powerdns/` package is excluded from `golangci-lint` with the
  reasoning recorded at the exclusion. The gate against it produced 729
  findings, 275 error-tier; fixing them here would rewrite files that the
  framework migration replaces. New code under `internal/` is linted from its
  first line and the exclusion narrows as each resource ports.

## [0.1.0] — 2026-07-28

### Added

- Initial fork.

[Unreleased]: https://github.com/dantte-lp/terraform-provider-powerdns/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/dantte-lp/terraform-provider-powerdns/releases/tag/v0.1.0
