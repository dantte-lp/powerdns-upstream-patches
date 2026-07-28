# ADR 0003 — Migrate to plugin-framework by muxing, not rewriting

- **Status:** accepted
- **Date:** 2026-07-28
- **Deciders:** architect, developer

## Context

The inherited provider is built on `terraform-plugin-sdk/v2` v2.40.1 and
declares protocol 5.0 in `terraform-registry-manifest.json`. SDKv2 is in
maintenance; the framework is the supported path and is required for write-only
attributes and ephemeral resources, both of which this provider needs for
DNSSEC private keys and TSIG secrets.

Ten resources and six data sources are in scope.

## Decision

Migrate **incrementally behind a multiplexed server**. `main.go` serves both:
`tf5to6server.UpgradeServer` lifts the existing SDKv2 provider to protocol 6,
and `tf6muxserver.NewMuxServer` combines it with the new framework provider.
Resources move one at a time; each move ships with its acceptance test.

The manifest declares protocol 6.0 from the start of the migration.

## Alternatives rejected

- **Big-bang rewrite.** A long-lived branch where nothing ships and every
  resource is simultaneously unverified. Rejected: it conflicts with the
  trunk-based rule and with keeping the provider shippable.
- **Stay on SDKv2.** Rejected: no write-only attributes, no ephemeral
  resources, and a deprecated foundation for work that will outlive it.

## Consequences

- `terraform-plugin-mux` becomes a dependency for the duration of the
  migration, and is removed when the last resource lands.
- A resource may not exist in both servers at once — the mux server rejects
  duplicate type names. The migration is therefore strictly one-way per
  resource, in one commit.
- State shape must be preserved across the move, or a state upgrader shipped.
  This is checked by an acceptance test that applies with the old resource and
  refreshes with the new.
