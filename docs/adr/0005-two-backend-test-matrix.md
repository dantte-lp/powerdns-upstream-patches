# ADR 0005 — Acceptance tests run on two backends

- **Status:** accepted
- **Date:** 2026-07-28
- **Deciders:** architect, developer

## Context

PowerDNS views and networks are unimplemented by the generic PostgreSQL
backend. On gpgsql a write returns `422` while a read returns `200` with an
empty collection; on LMDB both succeed. Two provider resources —
`powerdns_view_zone_association` and `powerdns_network` — depend on this.

Upstream's test harness runs a single LMDB instance, so those resources are
tested but the PostgreSQL failure mode never is; the other resources are the
mirror image.

## Decision

The acceptance matrix runs **two authoritative instances**: gpgsql on
PostgreSQL 17 and LMDB. Every resource declares which backends it supports, and
its acceptance test runs on each of them. A resource that is expected to fail
on a backend has a negative test asserting the *diagnostic*, not merely the
failure.

The recursor instance runs with `webservice.api_dir` set, plus a negative case
without it.

## Consequences

- The lab is four containers, not two, and CI runtime grows accordingly.
- Backend support becomes an explicit, testable property of a resource rather
  than folklore.
- A regression that makes a resource silently backend-dependent is caught.
