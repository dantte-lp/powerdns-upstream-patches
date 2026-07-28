# Development methodology

The delivery method for this fork, chosen jointly by the PM, architect and
developer roles.

## The choice: gated-iterative delivery

**Phase-gated macro-lifecycle, iterative sprints inside implementation,
trunk-based development, a per-resource Definition of Done, and evidence
discipline.**

### Why this and not the Scrumban proposed for upstream

The capability map's `CM-06` proposes Scrumban for the **upstream** repository,
and that remains the right call there: upstream is a maintained provider with a
solo maintainer and volunteer contributors, where nobody can commit to a sprint
volume and the flow is a stream of independent fixes.

This fork is a different problem and takes a different method:

| Factor | Upstream | This fork |
|---|---|---|
| Nature of work | stream of independent fixes | one architectural migration plus coverage |
| Contract stability | already shipped, must not break | schema is being re-cut; freeze it once |
| Who decides | maintainer, on volunteer capacity | this project, on its own schedule |
| Right method | Scrumban — flow with a release cadence | gated-iterative — design gates plus sprints |

A protocol change and a module-path change are not work you flow through a
Kanban board one card at a time. They need a design gate that closes before
implementation opens, because reversing them after resources are ported is
expensive. Coverage work afterwards is genuinely iterative and gets sprints.

Pure Waterfall would under-serve the per-resource discovery loop; pure Agile
would under-serve the contract. The hybrid fits.

## Roles

One person may wear several hats. The separation exists so that the same person
does not silently approve their own contract decisions, not to imply headcount.

| Role | Owns |
|---|---|
| **PM** | Scope, sprint goals, phase gates, changelog and release cadence, the risk register. |
| **Architect** | Provider, schema and state contract; ADRs; transport and retry design; non-goals. |
| **Developer** | Resource implementation, tests, documentation, gates green. |

Hard rule: the developer does not approve their own ADR, and the author of a
resource does not write its acceptance test.

## Macro-phases

| Phase | Output | Exit gate |
|---|---|---|
| 1. Audit | Fork baseline: inherited defects, stale dependencies, coverage. | `AUDIT-01` accepted; defect register agreed with the capability map. |
| 2. Design | Target layout, schema shapes, migration strategy, error taxonomy, ADRs. | ADRs 0001–0006 frozen; `terraform-registry-manifest.json` declares protocol 6.0; public surface frozen for the minor. |
| 3. Migration | SDKv2 → framework, resource by resource behind a mux server. | Every resource ported; unit plus lab acceptance green on every backend it supports. |
| 4. Coverage | DNSSEC, TSIG, autoprimaries, read-only data sources. | Each meets the per-resource Definition of Done. |
| 5. Verification | Version matrix across PowerDNS releases, negative tests, drift. | Matrix green. |
| 6. Release and contribution | Signed release; upstream pull requests for the portable fixes. | Registry listing live; upstream PRs opened. |

A downstream phase does not start until the upstream gate is signed off in the
sprint log.

## Why the migration is incremental, not a rewrite

Phase 3 uses `tf5to6server.UpgradeServer` with `tf6muxserver` so the SDKv2
provider and the framework provider serve side by side. Resources move one at a
time, each with its acceptance test, and the provider is shippable throughout.
This is ADR 0003; the alternative — a big-bang rewrite with a long red branch —
was rejected there.

## Iteration mechanics

- **Sprints are two weeks.** Sprint work lands on `sprint/<id>-<scope>`; the
  sprint pull request cites the exit criteria it satisfies.
- **Trunk-based development.** `main` is always shippable; short-lived
  branches; squash-merge; no long-running divergent forks.
- **The per-resource Definition of Done is the unit of progress** — see
  [`standards/terraform-provider-best-practices.md`](standards/terraform-provider-best-practices.md) §6.
- **Quality gates** (`make all`, `make verify`) are the automated part of the
  gate. A resource is not done until they are green with lab evidence quoted.

## Evidence discipline

Every factual claim about PowerDNS behaviour that a resource depends on carries
a source, per
[`standards/powerdns-api-discipline.md`](standards/powerdns-api-discipline.md).
Corroboration is a source reference plus a live round-trip, cited in the package
comment or the commit body. Unverified assumptions are labelled and kept
non-load-bearing.

This is not ceremony. The published OpenAPI specification for PowerDNS is wrong
in both directions, and a `200` from a collection endpoint does not mean the
backend supports it — two facts that would each have produced a broken resource
if taken on trust.

## Definition of ready

Before a resource enters a sprint:

- Endpoint and method identified **in the PowerDNS sources**, not only in the
  specification.
- Backend and configuration preconditions known and recorded.
- Schema shape drafted and reviewed against the API object.
- Test approach sketched, including which backends it must run on.

## Contribution back upstream

Fixes that are portable — those not depending on the framework migration — are
cherry-picked onto a branch cut from `origin/main` and opened as pull requests
against `mmianl/terraform-provider-powerdns`. They carry upstream's conventions
and version series, not this fork's. The defect register in the capability map
marks which defects are portable.

## Cadence and tracking

- ADRs are immutable numbered records under [`adr/`](adr/).
- The audit baseline is [`audit/AUDIT-01-fork-baseline.md`](audit/AUDIT-01-fork-baseline.md).
- Release cadence per [`release.md`](release.md).
