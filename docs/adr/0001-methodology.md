# ADR 0001 — Gated-iterative delivery

- **Status:** accepted
- **Date:** 2026-07-28
- **Deciders:** PM, architect, developer

## Context

The fork faces two kinds of work at once: a one-off architectural migration
(SDKv2 protocol 5 to plugin-framework protocol 6, module path, layout) and an
open-ended coverage backlog (DNSSEC, TSIG, autoprimaries, data sources).

The sibling analysis `powerdns-capability-map/CM-06` recommends Scrumban for
the *upstream* repository, where work is a stream of independent fixes carried
by a solo maintainer and volunteers.

## Decision

Adopt **gated-iterative delivery**: phase-gated macro-lifecycle, iterative
two-week sprints inside implementation, trunk-based development, a per-resource
Definition of Done, and evidence discipline.

Scrumban is retained as the recommendation for upstream. The two methods
coexist deliberately; contribution pull requests follow upstream's process, not
this one.

## Rationale

A protocol change and a module-path change cannot be flowed through a board one
card at a time — reversing them after resources are ported is expensive, so
they need a design gate that closes before implementation opens. Coverage work
afterwards is genuinely iterative and suits sprints.

## Consequences

- ADRs must be frozen before phase 3 opens.
- The provider must stay shippable during migration, which forces the mux
  approach in ADR 0003.
- Two processes exist in the project's orbit; `AGENTS.md` states which applies
  where.
