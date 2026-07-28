# Versioning

The provider follows [Semantic Versioning 2.0.0](https://semver.org/spec/v2.0.0.html).
The binary, its JSON schema and its published documentation share one version.
The source of truth is [`VERSION`](../../VERSION); the Git tag `vX.Y.Z` is cut
from it at release time.

## What "the public API" means for a provider

A change is **breaking** if an unmodified, previously valid Terraform
configuration or state would newly error, silently change meaning, or force
replacement. Concretely:

- removing or renaming a resource, data source or attribute;
- making an optional attribute required, or removing a default;
- narrowing an attribute's accepted values or its type;
- changing an attribute from updatable to `RequiresReplace`;
- changing the shape of `id`, or of state produced by `terraform import`;
- raising the minimum Terraform version or the provider protocol version.

Non-breaking: adding a resource, adding an optional attribute, adding a computed
output, widening accepted input, and bug fixes that restore documented
behaviour.

### The awkward case: validation that rejects what used to be accepted

Adding a `ValidateFunc` that rejects a value the provider previously accepted is
formally breaking — `terraform plan` newly errors on an unmodified
configuration. Three of the planned changes are exactly this shape, because the
provider currently accepts input the *server* then rejects.

The rule for this project: **such a change is MINOR, not MAJOR, when the
rejected configuration could not previously have applied successfully.** It is
recorded in `CHANGELOG.md` under `Changed` with a `BREAKING:` prefix and a
migration note regardless, because the user-visible symptom does move — from a
failed apply to a failed plan.

Where the configuration *could* previously apply, the change is MAJOR and waits
for one.

## Rules

| Component | Bumps when |
|---|---|
| MAJOR (`X`) | Breaking change — **after `1.0.0`**. |
| MINOR (`Y`) | Backward-compatible feature; **also breaking changes while `0.x`** (SemVer §4). |
| PATCH (`Z`) | Backward-compatible bug fix. |

## Fork versioning

The fork starts at `0.1.0`, not at upstream's `2.3.0`. Reasons:

- The plugin API changes from protocol 5.0 to 6.0 and the module path changes;
  continuing upstream's series would imply a compatibility relationship that
  does not hold.
- While the framework migration is in flight the schema is not yet frozen, and
  `0.x` is the honest signal for that.
- Changes intended for upstream are cherry-picked into pull requests against
  `mmianl/terraform-provider-powerdns` and take **its** version series, not
  this one.

## Pre-1.0 policy

While `0.x` the schema may change between minors. Per SemVer §4 anything MAY
change in `0.x`; this project constrains that to: breaking changes bump the
MINOR and carry a `BREAKING:` entry with a migration note. Patch releases never
break.

## Pre-releases

`vX.Y.Z-rc.N` for release candidates, `-alpha.N` / `-beta.N` earlier. A
pre-release never becomes a stable GitHub Release or a Registry version; it is
validated against the lab first.

## State-upgrade obligation

Any schema change that alters stored state shape ships a
`ResourceWithUpgradeState` implementation so existing state migrates cleanly.
This is part of the per-resource Definition of Done. The SDKv2-to-framework
migration is itself such a change and is covered by ADR 0003.

## Dependency version policy

"Always latest, pinned" — track the newest releases while pinning exact versions
for reproducibility:

- `go.mod` `go` directive: exact patch, currently `1.26.5`.
- Go dependencies: exact tags; Dependabot proposes weekly bumps.
- Tool versions in the dev image: pinned as `ARG`s in `Containerfile.dev` and
  mirrored in `compose.dev.yml`.
- GitHub Actions: pinned by commit SHA with a `# vX.Y.Z` trailing comment.
  Floating tags such as `@v6` are not acceptable — they are mutable references
  in a supply-chain position.
- Base image `golang:1.26-trixie` intentionally floats within the `1.26` minor
  so patch releases are picked up on rebuild.
