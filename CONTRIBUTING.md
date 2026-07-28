# Contributing

Read [`AGENTS.md`](AGENTS.md) first. It is the canonical guide; this file is the
short path through it.

## Before you start

```sh
make up && make shell     # dev container — no host toolchain
make lab-up               # PowerDNS fixture, both backends
```

## The loop

1. `scripts/worktree.sh new <type>/<scope>/<name>` — `main` is never committed
   to directly.
2. Develop in the container. Use `gopls` for navigation and the `context7` MCP
   for current library documentation.
3. Establish PowerDNS behaviour from the sources plus a live round-trip, not
   from the OpenAPI specification — it diverges from the implementation in both
   directions.
4. `make all` before pushing. `make verify` if you touched a resource.
5. Update `CHANGELOG.md` under `[Unreleased]`.
6. Open a pull request whose title is a Conventional Commit subject.

## Standards

All of [`docs/standards/`](docs/standards/) is normative. The four that catch
people out:

- **Validate what the server will reject** — a configuration the API cannot
  accept should fail at `plan`, not at `apply`.
- **Two backends** — views and networks only work on LMDB; an acceptance test on
  one backend is half a test.
- **Status before body** — check `resp.StatusCode` before decoding.
- **No AI attribution** anywhere in code, documentation, commits or PR text.
  Enforced by a commit hook.

## Contributing upstream instead

If your change fixes an inherited defect and does not depend on the framework
migration, it probably belongs in
[`mmianl/terraform-provider-powerdns`](https://github.com/mmianl/terraform-provider-powerdns)
rather than here. See [`docs/release.md`](docs/release.md) § "Contributing
upstream" for the procedure — it branches from a different remote and follows
upstream's conventions, not ours.
