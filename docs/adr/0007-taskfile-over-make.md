# ADR 0007 — Task replaces Make as the command interface

- **Status:** accepted
- **Date:** 2026-07-28
- **Deciders:** developer, architect

## Context

The repository needs one entry point for build, test, lint and lab commands.
It started with a `Makefile` copied in shape from the sibling
`terraform-provider-cvp`.

That `Makefile` immediately collided with the `GNUmakefile` inherited from
upstream: GNU make resolves `GNUmakefile` first, so every target added was
unreachable until the inherited file was deleted (plan `S1-07`). The incident
is small but it is a fair illustration of what Make costs — the resolution
order is implicit, and nothing reported the problem.

## Decision

Adopt [Task](https://taskfile.dev) as the command interface. `Taskfile.yml`
replaces `Makefile`; the `Makefile` is removed rather than kept as a wrapper.

Task is pinned to v3.52.0 in the dev image, as an `ARG` mirrored into
`compose.dev.yml`, per the dependency policy.

## Rationale

- **One entry point, not two.** A `Makefile` delegating to a `Taskfile` is two
  places to change and two to drift. Having just removed a file for exactly
  that reason, adding another would be perverse.
- **The rest of the repository is already YAML.** Compose, CI, golangci-lint,
  yamllint, markdownlint, commitlint. One fewer syntax to hold.
- **Namespaced tasks.** `lab:up`, `py:typecheck`, `tf:fmt:check` group
  naturally; `task --list` is a usable index, which a `Makefile` only achieves
  with a hand-maintained `help` target that goes stale.
- **`sources` and `generates`** give real incremental builds. Make's own
  mechanism is timestamps on file targets, which a `.PHONY`-heavy Makefile
  abandons entirely.
- **`requires: vars`** turns a missing argument into a clear message instead of
  a shell test embedded in the recipe.
- **No tab significance.** Not a deep argument, but the class of error it
  removes is real.

## What is given up

Make is everywhere; Task must be installed. That cost is close to zero here,
because the golden rule is already that work happens inside the dev container,
where the toolchain is ours to define. A contributor who has Podman has
everything.

CI does not run Task for the Go and Python jobs — it invokes the tools directly
so a workflow failure points at the tool rather than at a task wrapper. Task is
used in CI only for the lab lifecycle, where the orchestration is the point.

## Consequences

- `Makefile` is deleted. Documentation referring to `make <target>` is updated
  in the same change; a stale `make` reference is a documentation bug.
- Task joins the pinned toolchain and the version-mirroring rule.
- The command names change shape — the old `lab-up` is now `lab:up`, and
  `typecheck-py` is now `py:typecheck`. `all` and `verify` keep their names,
  because they are the two that matter most.
