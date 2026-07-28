# ADR 0006 — dnsdist is out of scope pending a contract

- **Status:** accepted
- **Date:** 2026-07-28
- **Deciders:** PM, architect

## Context

Upstream issue #34 asks for dnsdist support and PR #35 claims "Rules and
Statistics Management". The dnsdist 2.1.0 HTTP API registers ten endpoints, of
which exactly two write: `PUT /api/v1/servers/localhost/config/allow-from` and
`DELETE /api/v1/cache`. Rules, pools, downstream servers and dynamic blocks are
configured in Lua or YAML and are not reachable over HTTP.

## Decision

dnsdist is **out of scope** for this fork until one of the following holds:

1. dnsdist gains a write API for the objects a provider would manage; or
2. a deliberate decision is taken to drive dnsdist over its console or Lua
   configuration, which is a different transport with its own authentication
   model and would need its own ADR.

If a minimal scope is later wanted, it is a `powerdns_dnsdist_allow_from`
resource and read-only data sources for statistics — and nothing more, because
nothing more exists.

## Consequences

- Issue #34 cannot be satisfied as written; the honest answer is a rescoping,
  not a schedule.
- PR #35 must be examined for what it actually drives before any assessment of
  it is meaningful.
- The provider's stated domain stays "PowerDNS Authoritative and Recursor".
