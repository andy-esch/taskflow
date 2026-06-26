---
schema: 1
status: ready-to-start
epic: 20-cli-ux-and-ergonomics
description: Audits have no modified timestamp (date is immutable identity). Add a mutable updated field bumped on edit/append, surfaced in show + JSON. Split from audit-editing.
effort: S
tier: 3
priority: low
autonomy_level: 3
tags: [cli]
created: "2026-06-25"
updated_at: "2026-06-26"
deferred_at: "2026-06-26"
---
# Add a mutable updated/modified timestamp to audits

## Objective

An audit's `date` is **immutable** — it's identity (the slug is `<date>-<area>`).
But there's no field for *when the audit was last touched*. Tasks carry
`updated_at`; audits carry nothing. Add a mutable `updated` (modified) timestamp so
edits/lifecycle moves are dated without disturbing the immutable audit `date`.

## Context

Split out from the audit-editing work (the `audit date should be immutable; a
modified date is not` decision, 2026-06-25). Sibling:
[[audit-editing-faces-audit-edit-set-and-append]] (which adds `audit edit`/`append`
that would bump this field).

## Acceptance criteria

- [ ] Audits gain an `updated` (or `updated_at`, matching the task field name)
      frontmatter field, set/bumped on `audit edit`, `audit append`, and the
      `close/reopen/defer` moves (decide which mutate it — at least edit/append).
- [ ] `date` stays immutable and slug-forming; `updated` never affects the slug.
- [ ] Surfaced in `audit show` + the `--json` audit payload (additive; bump
      schema_version); `audit list -c` projection optional.
- [ ] go build/test/lint green; docs/golden updated.

## Risks / gotchas

- Timestamp source: the tool is otherwise time-agnostic in tests — thread a clock
  the way tasks' `updated_at` is set (find that path) so tests stay deterministic.
- Don't break the existing audit golden fixtures (the fixture audits have no
  `updated` — decide: omitempty, or backfill the fixtures).

## Done when

Editing/appending to an audit stamps a mutable `updated` field, the immutable
`date` is untouched, and both surface in show/JSON.

## Scope expansion 2026-06-25 — also epics

The adversarial review flagged the same gap for epics: tasks carry `updated_at`, but `epic set`/`MoveEpic` stamp nothing and epics have no modified field either. Treat this as ONE decision across the non-task entities: do epics + audits get a mutable modification timestamp (bumped on set/move/edit/append), and what is it named (`updated_at` to match tasks)? Audit `date` stays immutable; this is a separate, mutable field.
