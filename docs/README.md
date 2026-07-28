# Documentation

Reading order is encoded in the directory names: the guide, then the method,
then the standards, then the evidence.

| Document | Contents |
|---|---|
| [`../AGENTS.md`](../AGENTS.md) | **Start here.** Golden rules, workflow, gates. |
| [`methodology.md`](methodology.md) | Delivery method, roles, phase gates, sprints. |
| [`plan.md`](plan.md) | **Live delivery plan.** Task status, updated with the work. |
| [`development.md`](development.md) | Dev container, the lab, the daily loop. |
| [`testing.md`](testing.md) | Test layers and the two-backend matrix. |
| [`release.md`](release.md) | Cutting a release; contributing upstream. |

## Standards

Normative. Read the standard before changing the thing it governs.

| Standard | Governs |
|---|---|
| [`standards/naming-conventions.md`](standards/naming-conventions.md) | Files, branches, resource and attribute names. |
| [`standards/versioning.md`](standards/versioning.md) | SemVer, what "breaking" means for a provider. |
| [`standards/commits.md`](standards/commits.md) | Conventional Commits, evidence in the body. |
| [`standards/changelog.md`](standards/changelog.md) | Keep a Changelog, release cut. |
| [`standards/go-1.26-style.md`](standards/go-1.26-style.md) | Go patterns, antipatterns, tooling. |
| [`standards/terraform-provider-best-practices.md`](standards/terraform-provider-best-practices.md) | Provider design and Definition of Done. |
| [`standards/terragrunt-integration.md`](standards/terragrunt-integration.md) | How consumers orchestrate this provider. |
| [`standards/powerdns-api-discipline.md`](standards/powerdns-api-discipline.md) | How to establish a fact about PowerDNS. |
| [`standards/python-tooling.md`](standards/python-tooling.md) | uv, ruff, ty for the automation scripts. |
| [`standards/verified-identifiers.md`](standards/verified-identifiers.md) | Never write a SHA, version or citation from memory. |

## Decisions

Immutable, numbered. A reversal adds a superseding record rather than editing
one.

| ADR | Decision |
|---|---|
| [`adr/0001`](adr/0001-methodology.md) | Gated-iterative delivery. |
| [`adr/0002`](adr/0002-fork-and-upstream-relationship.md) | Fork relationship, licence, version series. |
| [`adr/0003`](adr/0003-plugin-framework-migration.md) | Migrate by muxing, not rewriting. |
| [`adr/0004`](adr/0004-podman-oci-dev-workflow.md) | Podman, OCI, Compose Specification. |
| [`adr/0005`](adr/0005-two-backend-test-matrix.md) | Acceptance runs on two backends. |
| [`adr/0006`](adr/0006-dnsdist-scope.md) | dnsdist out of scope pending a contract. |

## Audit

| Record | Contents |
|---|---|
| [`audit/AUDIT-01-fork-baseline.md`](audit/AUDIT-01-fork-baseline.md) | State at the fork point: 6 structural findings, 8 inherited defects, 5 further findings, the test blind spot. |

## Related repositories

| Repository | Holds |
|---|---|
| `powerdns-capability-map` | Coverage analysis, gap matrix, defect register. Cited here, never duplicated. |
| `pdns-upstream` | Clone of `PowerDNS/pdns` at the pinned tags. The authority on API behaviour. |
