# ADR 0002 — Fork relationship and licence

- **Status:** accepted
- **Date:** 2026-07-28
- **Deciders:** PM, architect

## Context

Work proceeds on a fork of `mmianl/terraform-provider-powerdns` with the stated
intent of contributing changes back. The upstream provider is published to the
Terraform Registry under `mmianl/powerdns` at v2.3.0 and is licensed MPL-2.0.

## Decision

1. The fork keeps **MPL-2.0**. A fork cannot relicense inherited code, and
   contributions must stay MPL-2.0-compatible to be acceptable upstream.
2. The fork restarts versioning at **0.1.0**, not 2.3.1. It changes the plugin
   protocol and the module path; continuing upstream's series would imply a
   compatibility relationship that does not hold.
3. Git remotes are asymmetric and not interchangeable: `origin` is upstream and
   is read-only for us plus the target of contribution PRs; `fork` is ours and
   receives all pushes.
4. Portable fixes — those not depending on the framework migration — are
   cherry-picked onto a branch cut from `origin/main` and opened against
   upstream under **upstream's** conventions.

## Consequences

- Two version series exist. A fix may ship as `0.2.0` here and as `2.3.1`
  upstream; the changelog entry names both where relevant.
- The module path changes to `github.com/dantte-lp/terraform-provider-powerdns`,
  which is itself a reason the fork cannot share upstream's version series.
- Registry publication under a different namespace is a separate decision, not
  taken here.
