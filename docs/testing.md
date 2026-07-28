# Testing

Three layers, and one structural rule that is unusual enough to state first.

## The rule: one backend is not enough

PowerDNS views and networks are unimplemented by the generic PostgreSQL
backend. On gpgsql a write returns `422`; a read returns `200` with an empty
collection. So:

- a fixture running only LMDB tests the resources that use views, but never the
  failure path that most real installations will hit;
- a fixture running only PostgreSQL never tests those resources at all.

Upstream's harness is the first case. This project runs both (ADR 0005).

Every resource declares which backends it supports. Its acceptance test runs on
each of them. Where a resource is expected to fail on a backend, the test
asserts the **diagnostic**, not merely the failure — a bare `422` reaching the
user is itself the defect.

## Layer 1 — unit tests

No network. Cover schema validation, plan modifiers, and request and response
marshalling.

```sh
make test          # -race -count=1
make test-run RUN=TestZone_ValidateMasters PKG=./internal/resources/zone
```

`t.Parallel()` is mandatory here and enforced by `paralleltest`.

## Layer 2 — acceptance tests

Real plan, apply, refresh and destroy through `terraform-plugin-testing`
against the lab.

```sh
make lab-up
make lab-verify    # assert the fixture matches the pinned reference points
make testacc
```

Requirements per resource:

- at least one acceptance test;
- an `ImportState` step with `ImportStateVerify: true`;
- `plancheck.ExpectEmptyPlan` after apply, so idempotency is asserted rather
  than assumed;
- `CheckDestroy` leaving zero residue;
- test objects namespaced `tf-acc-<RUN_ID>`.

`t.Parallel()` is **forbidden** here — the lab instances are shared.

### The network exception

`terraform destroy` on a `powerdns_network` cannot fully remove it: PowerDNS
has no `DELETE` for a network, so removal is a `PUT` with an empty view and the
entry remains listed. `CheckDestroy` for that resource asserts the view is
empty, not that the entry is gone. This is a property of the API, and the test
documents it rather than working around it.

## Layer 3 — behaviour verification

An HTTP `200` proves the API accepted a request. It does not prove DNS works.
The dev image carries `dnsutils` and `postgresql-client` so a test can assert
the thing the user actually cares about:

```sh
dig @127.0.0.1 -p 15300 www.example.com A +short
psql -h 127.0.0.1 -p 15432 -U pdns -d pdns -c 'select name,type,content from records'
```

Use this for resources where the API response is a weak signal — record
content, zone publication, DNSSEC signing.

## What CI runs

| Job | Layer |
|---|---|
| `build-and-test` | unit, with coverage |
| `acceptance (gpgsql)` | acceptance against PostgreSQL |
| `acceptance (lmdb)` | acceptance against LMDB |
| `lint-go` | golangci-lint v2 |
| `vulncheck` | govulncheck |
| `terraform` | `terraform fmt -check` on examples |
| `lint-docs` | markdownlint, yamllint, cspell |
| `commitlint` | PR title |

## Writing a test that makes a claim about PowerDNS

If a test encodes an assumption about server behaviour, the assumption is cited
in the test's comment with the source and the observed response — see
[`standards/powerdns-api-discipline.md`](standards/powerdns-api-discipline.md).
A test that asserts the wrong behaviour confidently is worse than no test.
