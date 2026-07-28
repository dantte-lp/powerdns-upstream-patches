# Commit conventions

Commits follow [Conventional Commits 1.0.0](https://www.conventionalcommits.org/en/v1.0.0/),
enforced by `commitlint` (config [`.commitlintrc.yaml`](../../.commitlintrc.yaml))
in the `commit-msg` hook and against pull-request titles in CI.

## Shape

```text
<type>(<scope>): <summary, imperative, ≤ 72 chars, lower case>

<body — WHY this change; wrap at 72; blank line before>

<footer — issue refs, BREAKING CHANGE:>
```

## Types

| Type | Use |
|---|---|
| `feat` | New user-facing capability — resource, attribute, auth method. |
| `fix` | Bug fix. |
| `docs` | Documentation only. |
| `chore` | Repository hygiene, non-product files. |
| `build` | Build system: `Containerfile`, `Makefile`, compose, `go.mod`. |
| `ci` | GitHub Actions and workflows. |
| `refactor` | Behaviour-preserving code change. |
| `perf` | Performance change. |
| `test` | Tests and acceptance harness only. |
| `style` | Formatting only. |
| `revert` | Revert of a prior commit. |

## Scopes

Enumerated in `.commitlintrc.yaml`: `provider`, `zone`, `record`, `metadata`,
`view`, `network`, `dnssec`, `tsig`, `autoprimary`, `recursor`, `dnsdist`,
`client`, `auth`, `docs`, `examples`, `lab`, `ci`, `build`, `deps`, `release`,
`test`, `lint`, `repo`, `standards`.

## Breaking changes

Either `feat(scope)!: …` / `fix(scope)!: …`, or a footer:

```text
BREAKING CHANGE: <what breaks and the migration path>
```

Every breaking change also gets a `CHANGELOG.md` entry under `Changed` with a
`BREAKING:` prefix. Version impact per [`versioning.md`](versioning.md).

## Evidence in the body

The body is where the *why* lives, and for this project that includes the
evidence required by `AGENTS.md` rule 3. A commit that relies on a claim about
PowerDNS behaviour cites it:

```text
fix(zone): parse IPv6 masters with net.SplitHostPort

strings.Split(s, ":") treated every colon as the port separator, so any
IPv6 master was rejected with "more than one colon in <ip>:<port> string".
The update path never validated masters at all, which is why the documented
workaround — create with IPv4, then edit — appeared to work.

Verified against auth-5.1.3: PUT /servers/localhost/zones accepts a bare
IPv6 literal in masters[] and returns 204.

4/4 acceptance tests pass against the lab.

Closes: #73
```

## Examples

```text
feat(dnssec): add powerdns_zone_cryptokey resource

docs(standards): add PowerDNS API discipline

build(deps): bump terraform-plugin-framework to v1.19.0

chore(release): 0.2.0
```

## Prohibited content

Commit messages, bodies, footers and trailers never mention AI, assistants, or
generated authorship — see `AGENTS.md` golden rule 6. This overrides any tool
default that would append such a trailer. Human co-authorship uses the ordinary
`Co-Authored-By: Name <email>` trailer.
