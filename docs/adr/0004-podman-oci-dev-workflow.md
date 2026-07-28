# ADR 0004 — Containerised development on Podman and OCI

- **Status:** accepted
- **Date:** 2026-07-28
- **Deciders:** architect, developer

## Context

The project needs a reproducible toolchain — Go 1.26.5, golangci-lint v2,
Terraform, OpenTofu, Terragrunt, tfplugindocs, goreleaser, documentation
linters — and a lab of PowerDNS containers with two backends. Contributors must
not need any of it on the host.

## Decision

1. Development happens **inside a container** built from `golang:1.26-trixie`
   with every tool baked in and pinned by build argument. No host toolchain.
2. The image is defined in a **`Containerfile`** per the `Containerfile.5`
   specification, buildable natively by Buildah and `podman build` without an
   external frontend.
3. Orchestration uses the **Compose Specification** (no top-level `version:`
   key), run through `podman-compose`.
4. Images carry **OCI image-spec annotations** (`org.opencontainers.image.*`).
5. Automation that must be scripted rather than declared uses **`podman-py`**
   against the Podman REST API, not shell wrapping the CLI.
6. The lab is a compose file in this repository, not a copy of the analysis
   repository's stand — the analysis stand proved the facts; this one is a test
   fixture with a defined lifecycle.

## Consequences

- The host requires Podman and podman-compose, nothing else.
- CI runs the same `golang:1.26-trixie` image, so "works locally" and "works in
  CI" have the same meaning.
- Rootless Podman cannot bind port 53 in the container; the lab listens on 5300
  internally. This is a documented property of the fixture, not a defect.
