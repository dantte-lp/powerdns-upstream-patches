# Naming conventions

Consistent names make a repository searchable, sortable and legible. This
standard adapts two general references —
[Harvard HMS file-naming conventions](https://datamanagement.hms.harvard.edu/plan-design/file-naming-conventions)
and [IT Glue naming best practices](https://www.itglue.com/blog/naming-conventions-examples-formats-best-practices/)
— to a Terraform provider code base.

## 1. Universal rules

Derived from both references:

- **Alphanumeric plus `-` and `_` only.** No spaces, no
  `~ ! @ # $ % ^ & * ( ) : < > ? , [ ] { } ' " |`.
- **One separator per context** (see §2); never mix within one identifier.
- **ISO 8601 dates** (`YYYY-MM-DD`) so names sort chronologically. Timestamps
  `YYYY-MM-DDThhmm`.
- **Most significant token first**, so the default lexical sort is useful.
- **Concise but descriptive** — aim for ≤ 40–50 characters in file names.
- **No vague descriptors** (`final`, `new`, `latest`, `copy`, `v2`). Git and
  SemVer exist for that.
- **No undocumented abbreviations.** Coining one means adding it to §6 and to
  `.cspell.json` in the same commit.

## 2. Case by context

| Context | Case | Example |
|---|---|---|
| Go files | `snake_case.go` | `zone_metadata.go`, `client_test.go` |
| Go packages | short, lowercase, no underscores | `zonemetadata`, `pdnsclient` |
| Go exported identifiers | `PascalCase` | `ZoneResource`, `ClientBundle` |
| Go unexported identifiers | `camelCase` | `buildHTTPClient`, `providerModel` |
| Markdown | `kebab-case.md` | `naming-conventions.md` |
| ADRs | `NNNN-kebab.md` | `0001-methodology.md` |
| Audit records | `AUDIT-NN-kebab.md` | `AUDIT-01-fork-baseline.md` |
| Directories | `kebab-case` or one word | `deployments/`, `internal/client/` |
| YAML / config | `kebab-case.yml` | `compose.dev.yml`, `dependabot.yml` |
| Terraform files | `snake_case.tf` | `main.tf`, `variables.tf` |
| Environment variables | `SCREAMING_SNAKE_CASE` | `PDNS_SERVER_URL`, `TF_ACC` |

## 3. Terraform-facing names — the public contract

These names **are** the provider's public API. Changing one is a breaking
change (see [`versioning.md`](versioning.md)). They follow HashiCorp's
convention and mirror the PowerDNS API where practical.

- **Resource and data-source types:** `powerdns_<noun>`, `snake_case`,
  singular — `powerdns_zone`, `powerdns_record`, `powerdns_zone_metadata`,
  `powerdns_tsigkey`, `powerdns_zone_cryptokey`, `powerdns_autoprimary`.
- **Attributes:** `snake_case`, matching the PowerDNS field name where the API
  name is not actively confusing. Where PowerDNS itself is inconsistent, the
  API name still wins: `soa_edit_api` and `soa_edit` are distinct fields and
  both keep their names.
- **Boolean attributes describe an action and default to `false`**:
  `api_rectify`, `nsec3narrow`, `presigned` — never `disable_*`.
- **IDs are `id`** (computed); typed references elsewhere carry the type —
  `zone_id`, `tsigkey_id`.
- **Timestamps:** RFC 3339 strings suffixed `_at`.

### Inherited names that violate this standard

Two upstream names predate the standard and are load-bearing in released
state. They are **not** renamed outside a major version:

| Name | Problem | Disposition |
|---|---|---|
| `masters` / `slave` zone kinds | PowerDNS itself moved to primary/secondary | Keep; the API still accepts and returns both. Revisit at 1.0. |
| `powerdns_view_zone_association` | Names a join, not a PowerDNS object | Keep; renaming breaks state. Documented in the resource page. |

Recording them here is the point: the standard describes the target, and a
deliberate exception is cheaper than a silent inconsistency.

## 4. Git branches

| Kind | Pattern | Example |
|---|---|---|
| Sprint | `sprint/<id>-<scope>` | `sprint/S3-test-harness` |
| Feature | `feat/<scope>/<name>` | `feat/dnssec/cryptokey-resource` |
| Fix | `fix/<scope>/<name>` | `fix/zone/ipv6-masters` |
| Chore | `chore/<scope>/<name>` | `chore/standards/bootstrap-2026-07-28` |
| Build | `build/<scope>/<name>` | `build/deps/plugin-framework` |

`<scope>` is one of the commit scopes enumerated in `.commitlintrc.yaml`.

## 5. Commits, versions, tags

- Commit subjects per [`commits.md`](commits.md).
- Versions per [`versioning.md`](versioning.md); the source of truth is
  `VERSION`.
- Tags `vX.Y.Z`, annotated and GPG-signed.
- Release archives, mandated verbatim by the Terraform Registry:
  `terraform-provider-powerdns_<version>_<os>_<arch>.zip`.

## 6. Glossary of accepted abbreviations

| Abbreviation | Meaning |
|---|---|
| `pdns` | PowerDNS |
| `auth` | PowerDNS Authoritative Server |
| `rec` | PowerDNS Recursor |
| `rrset` | resource record set |
| `soa` | start of authority |
| `tsig` | transaction signature |
| `axfr` | full zone transfer |
| `acc` | acceptance (test) |
| `crud` | create/read/update/delete |
| `lmdb` | Lightning Memory-Mapped Database backend |
| `gpgsql` | generic PostgreSQL backend |

Add entries here and to `.cspell.json` in the same commit that introduces them.
