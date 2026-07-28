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
  configured in `pyproject.toml` and gated by `make py` as part of `make all`.
- `docs/standards/python-tooling.md` recording the Python rules, including why
  the ruff selection is an allowlist rather than `ALL` and how a pre-1.0 type
  checker is treated in a merge gate.
- `docs/plan.md` — the live delivery plan. Task status changes in the same
  commit as the work; phase and sprint closures are recorded here.
- CI workflow with an acceptance matrix across both backends and a Python
  lint job; all GitHub Actions pinned by commit SHA.

## [0.1.0] — 2026-07-28

### Added

- Initial fork.

[Unreleased]: https://github.com/dantte-lp/terraform-provider-powerdns/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/dantte-lp/terraform-provider-powerdns/releases/tag/v0.1.0
