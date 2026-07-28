# Changelog conventions

[`CHANGELOG.md`](../../CHANGELOG.md) follows
[Keep a Changelog 1.1.0](https://keepachangelog.com/en/1.1.0/).

## Rules

- **Human-written and curated.** Not a `git log` dump — describe user impact.
- **`[Unreleased]` accumulates** at the top. Every pull request with a
  user-visible change adds an entry there; the PR template checks it.
- **Group by change type** in this order: `Added`, `Changed`, `Deprecated`,
  `Removed`, `Fixed`, `Security`.
- **Newest release first.** Each release heading is
  `## [X.Y.Z] — YYYY-MM-DD` (ISO 8601 date, em dash).
- **Breaking changes** go under `Changed` with a `BREAKING:` prefix and a
  migration note.
- **Link the diff** at the bottom with `[Unreleased]` and `[X.Y.Z]` compare
  links.

## Entries cite evidence

An entry that fixes a defect names it. The defect register is the capability
map's `CM-04` §4, so entries reference `D-nn` and, where one exists, the
upstream issue:

```markdown
### Fixed

- `powerdns_zone` accepts IPv6 addresses in `masters`; previously any IPv6
  literal was rejected as a malformed `<ip>:<port>` pair (D-01, upstream #73).
- `masters` is now validated on update as well as on create, so a value the
  provider rejects at create time can no longer enter state by way of an
  edit (D-02).
```

## Fork provenance

The changelog starts fresh at `0.1.0`. Upstream history up to `2.3.0` is not
copied in; it remains in the upstream repository and in Git history. The first
entry states the fork point explicitly so a reader can locate it.

## Release cut

1. Rename the `## [Unreleased]` content into a new dated
   `## [X.Y.Z] — YYYY-MM-DD` section; recreate an empty `[Unreleased]`.
2. Update `VERSION`.
3. Commit `chore(release): X.Y.Z`.
4. Tag `vX.Y.Z`, annotated and GPG-signed.
5. The release workflow extracts the section between two `## [` headings with
   `awk` and feeds it to goreleaser as the GitHub Release notes.

Because step 5 parses the file mechanically, keep the heading format exact:
`## [X.Y.Z] — YYYY-MM-DD`, version in square brackets.
