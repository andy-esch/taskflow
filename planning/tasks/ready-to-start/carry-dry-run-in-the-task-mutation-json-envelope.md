---
status: ready-to-start
epic: 20-cli-ux-and-ergonomics
description: task set/append/set --body emit TaskShowEnvelope under --json — no dry_run field, so a --dry-run preview is indistinguishable from a real write
effort: Unknown
tier: 3
priority: medium
autonomy_level: 3
tags: [cli, agents, json]
created: "2026-06-20"
---

# Carry dry_run in the task-mutation JSON envelope

## Objective

`task set`, `task append`, and `task set --body` all report under `--json` via
`render.TaskShowJSON` → `TaskShowEnvelope {schema_version, task, body}`
(`internal/cli/task.go:reportTaskMutation`). That envelope has **no `dry_run`
field**, so a `--dry-run` preview is byte-indistinguishable from a real write to
a `--json` consumer — even though schema_version 1.3 established "dry_run is
always present on mutation envelopes" (it is, on `moves`/`created`/`fix`/`init`,
just not on the task-mutation path). Surfaced by the 2026-06-20 adversarial
review. Pre-existing for `task set`; `task append`/`set --body` inherit it.

Secondary (same call site): `reportTaskMutation` hardcodes `body: ""`, so the
body-editing commands report an empty body even though the body is what changed.
A body op echoing the resulting body would let an agent confirm in one round-trip.

## Design fork (decide first)

`TaskShowEnvelope` is shared with `task show` (a read), where `dry_run` makes no
sense. Options:
1. A distinct mutation-result envelope (e.g. `TaskMutationEnvelope {schema_version,
   dry_run, task, body?}`) for set/edit-less body ops, leaving `task show` alone.
2. Add an `omitempty` `dry_run` to the shared envelope (simplest, but puts a
   mutation-only field on a read type).
Option 1 is cleaner; confirm before building. Either way it's additive →
schema_version minor bump, and the round-trip schema test must cover the new shape.

## Acceptance criteria

- [ ] `task set|append` `--json --dry-run` output is distinguishable from a real
      write (a `dry_run: true` marker).
- [ ] Body-editing commands optionally return the resulting body in `--json`
      (or an explicit decision to keep it omitted, recorded here).
- [ ] schema_version bumped; `schema --json-schema` + the round-trip test updated.
- [ ] Human output unchanged; suite + lint green.

## Out of scope

- The body replace/append behavior itself (shipped in
  [[agent-facing-cli-ergonomics-batch]]).

## Related

- Epic [[20-cli-ux-and-ergonomics]] ·
  [[agent-facing-cli-ergonomics-batch]] · [[publish-json-schema-for-the-json-envelopes]].
