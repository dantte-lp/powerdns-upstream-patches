# AUDIT-01 — Fork baseline

State of `mmianl/terraform-provider-powerdns` at the fork point, commit
`0dac0e7`, release `v2.3.0`. This is the record phase 1 of the methodology
exits on. Every figure is reproducible from the commands quoted.

## 1. Scale

| Measure | Value |
|---|---|
| Go source, excluding vendor and tests | ≈ 4 300 lines |
| Test source | 5 546 lines |
| Total Go, excluding vendor | 9 825 lines |
| Vendored dependencies | 29 MB |
| Resources | 10 |
| Data sources | 6 |
| Files importing `terraform-plugin-sdk` | 35 |
| Files importing `terraform-plugin-framework` | 0 |
| Licence | MPL-2.0 |

Test-to-source ratio is healthy. The problem is not test volume; it is what the
tests can reach — see §5.

## 2. Structural findings

Ordered by cost of leaving them.

### S-01 — Module path names a defunct namespace

`go.mod` declares `github.com/terraform-providers/terraform-provider-powerdns`.
The repository is `mmianl/terraform-provider-powerdns`; the
`terraform-providers` GitHub organisation was retired when HashiCorp moved
community providers to their authors' namespaces. The path resolves to nothing.

Effect: `go get` of the module path fails; `goimports.local-prefixes` cannot be
set meaningfully; the import path in `main.go` documents a repository that does
not exist.

### S-02 — SDKv2 and protocol 5.0

`terraform-registry-manifest.json` declares `"protocol_versions": ["5.0"]`;
`main.go` serves through `terraform-plugin-sdk/v2/plugin`. SDKv2 is in
maintenance. Protocol 5 has no write-only attributes and no ephemeral
resources — precisely the two mechanisms needed to handle DNSSEC private keys
and TSIG secrets without writing them to plain-text state.

This is the single finding that determines the shape of the whole project
(ADR 0003).

### S-03 — Vendored dependencies

29 MB of `vendor/`, with `GOFLAGS=-mod=vendor` forced in the makefile. Vendoring
predates module checksums and now mainly obscures diffs and inflates the
repository. `go.sum` provides the reproducibility guarantee it was there for.

### S-04 — Flat package layout

All ten resources, all six data sources and the HTTP client live in one
`powerdns/` package, 35 files. `client.go` alone is over 1 500 lines and mixes
transport, authoritative-API and recursor-API concerns. Nothing is under
`internal/`, so every symbol is importable by third parties and therefore
formally part of the public surface.

### S-05 — GitHub Actions pinned by floating tag

Workflows reference `actions/checkout@v6`, `actions/setup-go@v6`,
`golangci/golangci-lint-action@v9`, `goreleaser/goreleaser-action@v7`,
`crazy-max/ghaction-import-gpg@v7`, `hashicorp/setup-terraform@v4`. A tag is
mutable. For a workflow holding a GPG signing key this is a supply-chain
position that deserves a commit SHA.

### S-06 — Go directive behind

`go 1.26.1`; current is `1.26.5`. The workflow additionally hard-codes
`go-version: 1.26.1` in one place and `go-version-file: go.mod` in another, so
the two can drift.

## 3. Functional defects inherited

The register is maintained in the capability map, `CM-04` §4. Reproduced here
with the audit's assessment of portability — whether a fix can go upstream
without depending on the framework migration.

| ID | Defect | Location | Portable |
|---|---|---|---|
| D-01 | IPv6 in `masters` parsed as `<ip>:<port>` by `strings.Split` | `powerdns/resource_powerdns_zone.go:78-82` | yes |
| D-02 | `masters` validated only on create, never on update | `powerdns/resource_powerdns_zone.go:181-183` | yes |
| D-03 | views and networks silently unusable off LMDB | resources + docs | yes |
| D-04 | `powerdns_recursor_config` accepts any setting name; API supports two | `powerdns/resource_powerdns_recursor_config.go` | yes |
| D-05 | recursor resources require `webservice.api_dir`, undocumented | recursor resources + docs | yes |
| D-06 | JSON tag `json:"dnsssec"` — three `s` | `powerdns/client.go:205` | yes |
| D-07 | README lists a non-existent `powerdns_view` resource | `README.md` | yes |
| D-08 | `ListZones` decodes the body without checking the status code | `powerdns/client.go:340-366` | yes |

All eight are portable. That matters for sequencing: the upstream contribution
stream does not have to wait for the migration.

## 4. Additional findings from this audit

Not in the capability map, found while reading for the fork.

### A-01 — `DNSSec` field is dead as well as misspelled

`ZoneInfo.DNSSec` is declared and never read or written anywhere in the
package. D-06 records the misspelling; the audit adds that the field is
entirely unused, so the tag is currently harmless and will become a silent
data-loss bug the moment DNSSEC work begins. Fix the tag in the same commit
that first uses the field, or delete the field.

### A-02 — `ZoneInfo` carries 12 of the API's 24 zone attributes

Absent: `dnssec`, `nsec3param`, `nsec3narrow`, `presigned`, `soa_edit`,
`api_rectify`, `master_tsig_key_ids`, `slave_tsig_key_ids`, `notified_serial`,
`edited_serial`, `last_check`. The struct is not merely incomplete for the
resources that exist — it is the reason adding DNSSEC or TSIG touches the
client before it touches any resource.

### A-03 — Status-code handling is inconsistent, not merely absent

D-08 names `ListZones`. Reading the whole client, the pattern varies: some
methods switch on `resp.StatusCode` with a full case list, others decode
directly. The inconsistency is worse than uniform absence would be, because it
makes the correct methods look like the rule.

### A-04 — `DeleteNetwork` is not a delete

`client.go:1237-1244` sends `PUT` with `{"view": ""}` because the API has no
`DELETE` for a network. The workaround is correct and unavoidable, but the
consequence is not documented: after `terraform destroy` the network is still
listed by `GET /networks` with an empty view. A user reasonably expects destroy
to remove it.

### A-05 — Provider-level cache is undocumented in its interaction with drift

`cache_requests`, `cache_mem_size` and `cache_ttl` cache API responses for a
default 30 s. Nothing states how that interacts with `terraform plan
-refresh-only`, which is the operation whose whole purpose is to observe
out-of-band change. Open pull request #71 refactors this area; the interaction
needs deciding before that lands, not after.

## 5. Test coverage assessment

5 546 lines of tests, and a structural blind spot.

The acceptance harness (`acc-tests/docker-compose.yml`) starts **one**
authoritative instance, on LMDB. Consequences:

- `powerdns_view_zone_association` and `powerdns_network` are exercised, but
  only on the backend where they work. The `422` path on gpgsql — the common
  deployment — is never executed.
- Every other resource is exercised only on LMDB, not on the PostgreSQL backend
  most installations run.
- The recursor container has no `api_dir`, so the recursor resources cannot
  perform a successful write in the harness at all.

This is why ADR 0005 makes the two-backend matrix a structural decision rather
than a test-writing task.

## 6. What is in good shape

Recording this matters as much as the defects; the fork is not a rescue.

- Test volume and discipline are above average for a community provider.
- `tflog` is used throughout; there is no `log` or `fmt.Print` in library code.
- Sentinel errors are defined and wrapped consistently in the recursor client.
- Zone metadata, record comments and the `disabled` flag are handled correctly,
  including the round-trip that upstream issue #67 fixed.
- The upstream maintainer's trajectory is coherent — CRUD, then recursor, then
  quality, then 5.0 features — and the code reads as though one person held the
  whole design in mind.

## 7. Reproducing the figures

```sh
find powerdns pathorcontents -name '*.go' | xargs wc -l | tail -1
find . -name '*_test.go' -not -path './vendor/*' | xargs wc -l | tail -1
grep -rl "terraform-plugin-sdk" powerdns/ | wc -l
du -sh vendor
jq -r '.metadata.protocol_versions[]' terraform-registry-manifest.json
grep -hE 'uses:' .github/workflows/*.yml | sort -u
```

Backend and API claims are reproduced by the lab and are recorded in
`powerdns-capability-map/docs/03-src/SRC-01-lab-evidence.md`.
