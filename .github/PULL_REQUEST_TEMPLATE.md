## What and why

<!-- What changes, and the reason. Link the issue or the defect id (D-nn). -->

## Evidence

<!--
Required for any claim about PowerDNS behaviour (AGENTS.md rule 3):
a source reference in PowerDNS/pdns at the pinned tag, and a live round-trip.
Quote the status code and body, not a summary of them.
-->

## Checklist

- [ ] `make all` is green; output quoted below if this changes behaviour
- [ ] Resource changes: `make verify` green on **every backend the resource supports**
- [ ] `CHANGELOG.md` `[Unreleased]` updated
- [ ] Registry docs regenerated (`make docs`) if the schema changed
- [ ] New PowerDNS behaviour claims cited against the sources and the lab
- [ ] Title is a Conventional Commit subject
- [ ] No AI attribution anywhere in the diff, commits or this description

## Backends exercised

- [ ] gpgsql (PostgreSQL)
- [ ] lmdb
- [ ] recursor with `api_dir`
- [ ] not applicable — no resource behaviour changed
