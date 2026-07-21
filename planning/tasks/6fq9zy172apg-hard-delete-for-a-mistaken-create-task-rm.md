---
schema: 1
id: 6fq9zy172apg
status: ready-to-start
epic: 20-cli-ux-and-ergonomics
description: No task rm/audit rm; undoing a mistaken create needs deprecate (tombstone) or a raw rm (bypasses the CLI).
effort: Unknown
tier: 3
priority: low
autonomy_level: 3
tags: [lifecycle]
created: "2026-07-18"
---
# Hard-delete for a mistaken create (task rm / audit rm)

## Objective

A duplicate or mistaken `task new` can only be undone with `deprecate` (leaves a
tombstone) or by `rm`-ing the file (bypasses the CLI, against the "drive the
lifecycle via tskflwctl" principle). There is no hard-delete verb. Confirmed: no
`task rm`/`audit rm` today.

## Acceptance criteria

- [ ] A `task rm` (and `audit rm`) removes a never-committed mistake through the CLI
- [ ] Guard-railed so it's clearly for mistakes, not a routine lifecycle move (vs deprecate)

## Notes

- Source: https://github.com/andy-esch/taskflow/issues/105 (P5, Low)
