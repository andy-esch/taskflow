---
status: completed
epic: 20-cli-ux-and-ergonomics
description: task set/append/set --body emit TaskShowEnvelope under --json — no dry_run field, so a --dry-run preview is indistinguishable from a real write
effort: Unknown
tier: 3
priority: medium
autonomy_level: 3
tags: [cli, agents, json]
created: "2026-06-20"
updated_at: "2026-06-21"
started_at: "2026-06-21"
completed_at: "2026-06-21"
id: 6fe4my002bp7
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

## Shipped (2026-06-21)

**Decisions (owner, 2026-06-21):** option 1 — a distinct **`TaskMutationEnvelope
{schema_version, dry_run, task, body?}`** — and **echo the resulting body** for the
body-editing commands.

- New `render.TaskMutationEnvelope` + `TaskMutationJSON`; `task set`, `task append`,
  and `task set --body` now emit it under `--json` (via `reportTaskMutation`).
  `dry_run` is always present (a preview is distinguishable from a real write);
  `body` is `omitempty` — populated for the body commands, omitted for field-only
  `set`. `task show` keeps its own envelope (no `dry_run`).
- To echo the body accurately (append computes old+addition in the store),
  `store.EditBody` → `core.ReplaceBody`/`AppendBody` now return the resulting body
  alongside the task (port + fake updated).
- `schema_version` 1.6 → 1.7; the new envelope is in `schema --json-schema`
  (with its doc-comment description), the round-trip test (now 18 envelopes), and
  the regenerated golden. Human output unchanged.
- Tests: cli (`append --json` echoes the body; `set --body --dry-run --json` →
  `dry_run:true` + would-be body, no write; field-set `--dry-run --json` →
  `dry_run:true`, body omitted). Verified end-to-end on the binary.

## Acceptance criteria

- [x] `task set|append` `--json --dry-run` output is distinguishable from a real
      write (a `dry_run: true` marker) — true for the field-set path too.
- [x] Body-editing commands return the resulting body in `--json` (decided: echo it).
- [x] schema_version bumped (1.6→1.7); `schema --json-schema` + the round-trip test
      + golden updated.
- [x] Human output unchanged; suite + lint green.

## Out of scope

- The body replace/append behavior itself (shipped in
  [agent-facing-cli-ergonomics-batch](6fbj87000anh-agent-facing-cli-ergonomics-batch.md)).

## Related

- Epic [20-cli-ux-and-ergonomics](../epics/20-cli-ux-and-ergonomics.md) ·
  [agent-facing-cli-ergonomics-batch](6fbj87000anh-agent-facing-cli-ergonomics-batch.md) · [publish-json-schema-for-the-json-envelopes](6fdtbb400n8x-publish-json-schema-for-the-json-envelopes.md).
