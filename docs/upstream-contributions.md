# Upstream contributions

Seven pull requests against
[`mmianl/terraform-provider-powerdns`](https://github.com/mmianl/terraform-provider-powerdns),
opened 2026-07-28. This is the register; the branches are in this repository.

## Why they exist separately from everything else here

The migration this fork began was justified largely by one obligation: state
continuity for users of the published upstream provider. `DESIGN-02` concluded
we are not those users' maintainer, and that the obligation is better
discharged directly — by sending the fixes upstream, where they reach those
users regardless of what we build next.

None of these depend on the framework migration. Each is branched from
`origin/main` and follows **upstream's** conventions, not this repository's:
lower-case commit subjects without a Conventional Commits prefix, the issue
number in the subject, vendored dependencies left intact.

## The register

| PR | Branch | Defect | Diff |
|---|---|---|---:|
| [#75](https://github.com/mmianl/terraform-provider-powerdns/pull/75) | `fix/ipv6-masters` | D-01, D-02 — closes upstream #73 | +155/−26 |
| [#76](https://github.com/mmianl/terraform-provider-powerdns/pull/76) | `fix/tls-min-version` | no TLS floor on the API client | +7/−1 |
| [#77](https://github.com/mmianl/terraform-provider-powerdns/pull/77) | `fix/dnssec-json-tag` | D-06 — `json:"dnsssec"` | +1/−1 |
| [#78](https://github.com/mmianl/terraform-provider-powerdns/pull/78) | `docs/readme-view-resource` | D-07 — README names a non-existent resource | +6/−1 |
| [#79](https://github.com/mmianl/terraform-provider-powerdns/pull/79) | `fix/recursor-config-name-validation` | D-04 — accepts settings the API cannot write | +18/−5 |
| [#80](https://github.com/mmianl/terraform-provider-powerdns/pull/80) | `fix/list-zones-status-check` | D-08, A-03 — body decoded without checking status | +24/−0 |
| [#81](https://github.com/mmianl/terraform-provider-powerdns/pull/81) | `fix/backend-requirement-diagnostics` | D-03, D-05 — `422` with no stated requirement | +103/−7 |

Defect identifiers are the capability map's `CM-04` §4; `A-03` is
`AUDIT-01` §4.

## How each was verified

Against upstream's own gates, not this repository's:

```sh
make fmtcheck                    # their GNUmakefile
golangci-lint run ./...          # defaults, no config file upstream
go mod tidy && go mod vendor     # upstream vendors; CI checks it is clean
go test ./powerdns/ -count=1     # unit
TF_ACC=1 go test ./powerdns/     # acceptance, against the lab
```

Acceptance ran against PowerDNS Authoritative 5.1.3 on both the `gpgsql` and
LMDB backends, and Recursor 5.4.4 with `api_dir` set.

## Three things worth recording

**#75 would not have closed the issue with only the reported fix.** Repairing
the `strings.Split` parsing was the obvious half. Running the reporter's exact
configuration afterwards showed PowerDNS stores the compressed form, so
`fd92:81e1:e314:ea7b:0000:1234:5678:60ab` returns as
`fd92:81e1:e314:ea7b:0:1234:5678:60ab` and every subsequent plan wants to change
the zone. Without the `StateFunc` the reporter would have traded an error for a
permanent diff.

**#80 was opened before its tests were run, and they failed.** Making
`ListRecords` strict about every non-200 broke every record acceptance test:
after `terraform destroy` the zone is gone, `GET /zones/{id}` answers `404`, and
`CheckDestroy` asks for the records of exactly that deleted zone. The baseline
was checked to confirm the breakage was ours, `404` was made an empty result
with the reason recorded at the call site, and the pull request body was
rewritten to state the exception. The lesson is ordering: run the suite, then
open the pull request.

**Upstream's error strings were preserved wherever tests depended on them.**
`ValidateMasterAddress` reports "values in masters list attribute must be valid
IPs" and "invalid port value in masters attribute" — the wording the create
path used — so `TestAccPDNSZoneSlaveWithInvalidMasters` and
`TestAccPDNSZoneSlaveWithMastersWithInvalidPort` pass untouched. A better
message would have been easy and would have turned a focused fix into a
diff across their test suite.

## Reported to PowerDNS itself

[`PowerDNS/pdns#17807`](https://github.com/PowerDNS/pdns/issues/17807) — the
OpenAPI specification diverges from the implementation in both directions:
`GET /config/{config_setting_name}` is documented with no handler behind it,
and `POST` on `cryptokeys/{cryptokey_id}` is registered but absent from the
specification. Verified on `master` `a74d89a8`, not only on the 5.1.3 tag, and
checked for duplicates first as their `CONTRIBUTING.md` asks.

This is the fact `docs/standards/powerdns-api-discipline.md` is built on, so
reporting it is not a courtesy — it is the thing that would let that standard
be relaxed if fixed.

### On the AI disclosure in that issue

`PowerDNS/pdns` publishes an [`AI_POLICY.md`](https://github.com/PowerDNS/pdns/blob/master/AI_POLICY.md)
that forbids AI-produced **code**, permits AI use for **bug reports**, and
requires that any such use be disclosed. The issue therefore carries a
disclosure block, which is a deliberate exception to this repository's own rule
against AI attribution.

The two rules do not actually conflict once their scopes are read: ours governs
what we publish in our own repositories, theirs governs what they accept in
theirs. Filing without disclosure would have broken their stated terms. The
seven pull requests to `mmianl/terraform-provider-powerdns` carry no such
disclosure because that project publishes no such policy — the rule is to
follow each recipient's terms, not to apply one habit everywhere.

## Status

Seven pull requests and one issue awaiting review. If a maintainer asks for changes, the branch is in this
repository and the lab that verified it is `task lab:up`.
