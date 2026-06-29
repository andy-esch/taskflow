---
schema: 1
status: completed
epic: 20-cli-ux-and-ergonomics
description: Audits have no modified timestamp (date is immutable identity). Add a mutable updated field bumped on edit/append, surfaced in show + JSON. Split from audit-editing.
effort: S
tier: 3
priority: low
autonomy_level: 3
tags: [cli]
created: "2026-06-25"
updated_at: "2026-06-29"
deferred_at: "2026-06-26"
completed_at: "2026-06-29"
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

- [x] Audits gain an `updated_at` frontmatter field, bumped on `audit edit` +
      `audit append` (NOT on `close/reopen/defer` — a bucket move is a pure
      relocation; decided with the user). Epics get it too (set/edit/status-move).
- [x] `date` stays immutable and slug-forming; `updated_at` never affects the slug.
- [x] Surfaced in the `--json` audit + epic payloads (additive; schema_version
      1.21→1.22) and in human `audit show` (beside `date`). Human `epic show` stays
      minimal (it shows no dates at all), so the epic field is JSON-only there.
- [x] go build/test/lint green; docs/golden updated.

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

## Completed 2026-06-29 — `updated_at` on epics + audits, uniform "any edit bumps"

Decisions (with the user): both epics AND audits get the field; named `updated_at`
(matching tasks); **every content write bumps it** — set, append, AND the `$EDITOR`
edit paths (incl. `task edit`, which previously wrote verbatim) — while pure
relocation (audit `close`/`reopen`/`defer` bucket moves), dry-runs, and no-ops do
NOT. An epic's stored `updated_at` (own edits) stays distinct from the DERIVED
`EpicSummary.LastUpdated` (max member-task activity); a member-task change does not
touch it.

Shipped:
- **domain**: `Updated` (`yaml:"updated_at"`) on `Epic` + `Audit`; audit `date`
  stays immutable (the slug).
- **wire**: `updated_at` on `EpicMetaJSON` + `AuditJSON` (omitempty); schema_version
  1.21→1.22; goldens / schema_comments / docs regenerated.
- **store**: `editFile` now stamps `updated_at` on an accepted change (so task/epic/
  audit edit all bump uniformly); `AppendAuditBody` stamps (switched off the no-stamp
  `replaceBody`, which is now deleted as dead); `MoveEpic` stamps on a *real* status
  change only; `MoveAudit` deliberately does NOT (pure relocation). The clock is
  threaded as a `now` param (consistent with `moveTask`/`EditBody`); `SetEpicFields`
  injects `updated_at` into the map (mirroring task `SetFields`) and rejects a
  caller setting it.
- **tests**: store edit/append/move tests assert the stamp + date-immutability;
  `TestFS_MoveEpic_NoOp_NoStamp` and the `TestFS_MoveAudit` no-stamp assertion pin the
  exclusions; the old `TestAppendAuditBody_AppendsNoUpdatedAtStamp` was inverted to
  `…_StampsUpdatedAt`. build/vet/test/lint green; smoke-tested live.
