# Security policy

## Reporting a vulnerability

Report privately through GitHub Security Advisories on this repository. Do not
open a public issue for an unfixed vulnerability.

If the vulnerability is in PowerDNS itself rather than in this provider, report
it to the PowerDNS security team — see their advisory process. This provider is
a client; a server-side flaw is not ours to embargo.

## What this provider handles

- **API keys** for the PowerDNS HTTP API, from provider configuration or
  `PDNS_API_KEY`.
- **TSIG keys** and **DNSSEC private keys** once those resources exist.

### The state-file problem

Terraform state is not encrypted. Any attribute the provider reads back is
written to state in plain text, including secrets marked `Sensitive` —
`Sensitive` redacts console output, nothing more.

The project's response, in order of preference:

1. **Write-only attributes** (Terraform 1.11+) for values that need no drift
   detection.
2. **Ephemeral resources** for values that exist only for the duration of a run.
3. Where neither applies, the resource documentation states plainly that the
   value lands in state, so the operator can protect the backend accordingly.

A DNSSEC private key retrieved from `GET /cryptokeys/{id}` is the sharp case and
is settled in its design review before the resource is implemented.

## Supply chain

- GitHub Actions are pinned by commit SHA, not by tag. A tag is mutable, and a
  workflow that holds a signing key is not the place for a mutable reference.
- Go dependencies are pinned to exact versions with `go.sum` verification.
- `govulncheck` and `osv-scanner` run in CI; an allow-listed advisory needs a
  documented reason.
- Release archives are GPG-signed.

## Lab credentials

The lab API key `labapikey` is a deliberately public test value in a fixture
that binds to loopback. It is not a secret and is never reused. Do not point
acceptance tests at a production PowerDNS.
