# Release

Two distinct flows: cutting a release of this fork, and contributing a change
upstream. They use different conventions and different version series.

## Cutting a fork release

1. Confirm the gate is green, including the lab:

   ```sh
   task verify
   ```

2. Move `[Unreleased]` into a dated section in `CHANGELOG.md`:
   `## [X.Y.Z] — YYYY-MM-DD`. Recreate an empty `[Unreleased]`. Keep the
   heading format exact — the workflow parses it with `awk`.
3. Update `VERSION`.
4. Commit `chore(release): X.Y.Z`.
5. Tag `vX.Y.Z`, annotated and GPG-signed:

   ```sh
   git tag -s -a v0.2.0 -m 'v0.2.0'
   git push fork v0.2.0
   ```

6. The release workflow runs goreleaser, which builds the platform archives,
   signs the checksum file with the release key, and publishes the GitHub
   Release using the changelog section as its notes.

Version impact per [`standards/versioning.md`](standards/versioning.md).
While `0.x`, a breaking change bumps the minor and carries a `BREAKING:`
changelog entry with a migration note.

## Registry publication

The Terraform Registry requires an OSI licence (MPL-2.0, inherited), a
top-level `terraform-registry-manifest.json` declaring the protocol version, a
GPG-signed release with a `_SHA256SUMS.sig`, archives named
`terraform-provider-powerdns_<version>_<os>_<arch>.zip`, and example programs.

The manifest declares protocol `6.0` from the start of the framework migration
(ADR 0003), because a provider served through the mux server speaks 6.0 even
while some resources are still SDKv2 underneath.

## Contributing upstream

Fixes that do not depend on the framework migration go to
`mmianl/terraform-provider-powerdns`. All eight inherited defects qualify
(`audit/AUDIT-01-fork-baseline.md` §3).

```sh
git fetch origin
git checkout -b fix/ipv6-masters origin/main
git cherry-pick <commit>
git push fork fix/ipv6-masters
gh pr create --repo mmianl/terraform-provider-powerdns --base main
```

Rules for these branches, which override this repository's own:

- Branch from `origin/main`, never from `fork/main` — the fork's tree has
  diverged structurally and a PR from it would be unreviewable.
- Follow **upstream's** conventions, not this project's standards set.
- Do not carry across the module-path change, the layout change, or anything
  that assumes the plugin framework.
- One defect per pull request. D-01 and D-02 are the exception: they are the
  same bug seen from two paths and belong together.
- Cite the evidence in the PR body the same way as here — upstream has no
  obligation to trust an assertion.

The reciprocal courtesy: a fix accepted upstream is recorded in this
repository's changelog with the upstream PR number, so the two histories stay
traceable to each other.

## Reporting the PowerDNS specification defect

The published OpenAPI for Authoritative 5.1.3 documents an endpoint that does
not exist and omits one that does
([`standards/powerdns-api-discipline.md`](standards/powerdns-api-discipline.md) §2).
That belongs in the `PowerDNS/pdns` issue tracker, not in any provider. It is
not a blocker for anything here — the project already treats the sources as
authoritative — but reporting it is the neighbourly thing to do.
