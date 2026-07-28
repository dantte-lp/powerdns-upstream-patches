# Development

Everything runs inside the Podman dev container. The host needs Podman and
`podman-compose` — nothing else. No Go, Terraform or linter on the host.

## Prerequisites

```sh
podman --version          # 5.x
podman-compose --version
```

## First run

```sh
make up          # build and start the dev container
make shell       # a shell inside it
make versions    # confirm the pinned toolchain
```

`make up` builds `golang:1.26-trixie` with the toolchain baked in and pinned by
build argument — Go 1.26.5, golangci-lint v2.12.2, Terraform 1.15.8, OpenTofu
1.12.5, Terragrunt 1.1.1, tfplugindocs v0.25.0, goreleaser v2.17.1, gopls, plus
the documentation linters. Versions live in one place,
`deployments/containers/Containerfile.dev`, and are mirrored in
`deployments/compose/compose.dev.yml`.

## The lab

Acceptance tests need PowerDNS. `make lab-up` brings up four containers:

| Service | Endpoint | Why it exists |
|---|---|---|
| `pdns-lab-auth-pg` | `http://127.0.0.1:18081/api/v1` | Authoritative on PostgreSQL — the common deployment |
| `pdns-lab-auth-lmdb` | `http://127.0.0.1:18091/api/v1` | Authoritative on LMDB — the **only** backend implementing views and networks |
| `pdns-lab-recursor` | `http://127.0.0.1:18082/api/v1` | Recursor with `api_dir` set — without it every write returns 422 |
| `pdns-lab-pg` | `127.0.0.1:15432` | backend for `auth-pg` |

Two authoritative instances is not redundancy. On gpgsql a view write returns
`422` while a view read returns `200` with an empty list, so a single-backend
fixture cannot distinguish "unsupported" from "not configured". See ADR 0005.

```sh
make lab-up       # start and wait for every API to answer
make lab-verify   # assert versions and backends match the pinned references
make lab-status   # container state and reported versions
make lab-down     # remove, including volumes
```

The lab is driven by `scripts/automation/lab.py` through **podman-py**, not by
shelling out to the CLI: a failure arrives as an exception with a status code
rather than a string to be parsed. `lab-verify` is the interesting target — it
asserts the fixture is the one the tests were written against, so a silently
upgraded image is caught before it produces a confusing test failure.

## The daily loop

```sh
make build        # compile the provider
make test         # unit tests, race detector on
make lint         # golangci-lint v2
make all          # the pre-PR gate
```

`make all` runs build, unit tests, lint, `terraform fmt -check`, the
documentation linters and `govulncheck`. It does not need the lab.

`make verify` is `make all` plus `lab-verify` and the acceptance suite. Run it
before any pull request that touches a resource.

## Tooling you are expected to use

- **`gopls`** is installed in the image. Use the LSP for navigation, rename and
  live diagnostics rather than grepping. `find-references` before changing a
  client method is not optional courtesy — the client is used from every
  resource.
- **`context7` MCP** for current library documentation before writing code
  against it. Training-data recall of a framework signature is not evidence.
- **The PowerDNS sources** at `../pdns-upstream`, tags `auth-5.1.3`,
  `rec-5.4.4`, `dnsdist-2.1.0`. This is the authority on API behaviour; the
  published OpenAPI is not (see
  [`standards/powerdns-api-discipline.md`](standards/powerdns-api-discipline.md)).

## Manual testing against a local build

```sh
make install
```

Then point Terraform at the local binary with a `dev_overrides` block in
`~/.terraformrc`:

```hcl
provider_installation {
  dev_overrides {
    "dantte-lp/powerdns" = "/root/go/bin"
  }
  direct {}
}
```

`terraform init` is skipped under `dev_overrides`; run `terraform plan`
directly.

## Branching

`main` is never committed to directly.

```sh
scripts/worktree.sh new fix/zone/ipv6-masters
cd ../.worktrees/fix/zone/ipv6-masters
make up && make shell
```

Branch naming per
[`standards/naming-conventions.md`](standards/naming-conventions.md) §4.

## Remotes

| Remote | Repository | Use |
|---|---|---|
| `origin` | `mmianl/terraform-provider-powerdns` | upstream — read, and the target of contribution PRs |
| `fork` | `dantte-lp/terraform-provider-powerdns` | ours — all pushes |

Pushing to `origin` is a mistake, not a shortcut.

## Troubleshooting

**The lab starts but an API never answers.** `make lab-status` reports each
container's state and version. A container that is `running` but unreachable is
usually a configuration parse failure — check `podman logs pdns-lab-<name>`.

**Recursor writes return 422.** `webservice.api_dir` is unset. The lab sets it;
a hand-rolled recursor will not have it. This is not a provider bug — see the
API discipline standard.

**Views or networks return 422 on `auth-pg`.** Expected. Those resources
require LMDB; use `PDNS_SERVER_URL=http://127.0.0.1:18091`.

**Rootless Podman cannot bind port 53.** The lab listens on 5300 inside the
container and publishes 15300. This is a property of the fixture, not a defect.
