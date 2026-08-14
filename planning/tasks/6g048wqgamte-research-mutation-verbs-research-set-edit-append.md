---
schema: 1
id: 6g048wqgamte
status: ready-to-start
epic: 28-first-class-entities-new-planning-nouns
description: 'Give research the agent+human mutation faces every other entity has: field-level set, editor-based edit, body append; created stays immutable (the id encodes it).'
effort: Unknown
tier: 3
priority: medium
autonomy_level: 3
tags: [cli, core]
created: "2026-08-14"
---

# Research mutation verbs — `research set` / `edit` / `append`

## Objective

`research` landed read-mostly: `new|list|show|path`. Editing a doc today means
`research path` + `$EDITOR`, with no re-validation on save and no agent-facing field
write. Close the gap so research has the same two faces of mutation every other entity
has — **agent** (field-level, scriptable, atomic) and **human** (`$EDITOR` on the whole
file, re-validated on save).

## Why it was deferred, and what changed

`research set` was deliberately skipped in the first pass: the main reason `set` exists
for tasks is maintaining cross-reference fields (`epic`, `blocked_by`, …), and research
has none by design. But the thin contract still has writable fields — `description` and
`tags` — and `description` is empty on all 28 migrated docs, so there is no scriptable
way to backfill them today.

## Acceptance criteria

- [ ] `research set <slug> --description/--tags` (+ `--set key=value` / `--unset`),
      surgical: unknown keys, comments, and key order preserved.
- [ ] `research edit <slug>` — `$EDITOR` on the whole file, parse-before-accept, loops
      on a broken edit; stamps `updated_at` on a changed save.
- [ ] `research append <slug>` — body append in one atomic validated write, stamping
      `updated_at` (`created` stays immutable — the id is minted from it).
- [ ] `created` is NOT settable via `set`: the id encodes it, so changing one silently
      desynchronizes them. Reject with `ErrValidation` naming why.
- [ ] Ports added to `core.ResearchStore`; `--json` mutation envelope mirrors
      `task_mutation` / `audit_mutation`; `schema_version` bumped.

## Out of scope

- Any lifecycle verb — research has no status (epic 28 decision).
- `research rename` (slug churn + link cascade) — its own call if wanted.

## Related

- Epic [28-first-class-entities-new-planning-nouns](../epics/28-first-class-entities-new-planning-nouns.md)
