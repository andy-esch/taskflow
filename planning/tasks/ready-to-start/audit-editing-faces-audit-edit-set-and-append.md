---
schema: 1
status: ready-to-start
epic: 20-cli-ux-and-ergonomics
description: Audits lack the editing faces tasks have; add audit edit ($EDITOR), audit set (area/date), and audit append (add a finding), reusing the picker + editor infra.
effort: M
tier: 3
priority: low
autonomy_level: 3
tags: [cli]
created: "2026-06-25"
updated_at: "2026-06-25"
---
# Audit editing faces: `audit edit` + `audit set` + `audit append`

## Objective

Audits lack the editing faces tasks have. Findings are *authored* by hand-editing
the audit file, yet there's no `audit edit` to open it — you must hunt the path
yourself. There's also no `audit set` (area/date) or `audit append` (add a
finding). Give audits the same agent + human mutation faces as tasks.

## Context

`audit` today is `{new, list, show, findings, lint, close/reopen/defer}`. Templates
and infra already exist:
- `task edit` ($EDITOR, re-validated), `task set` (fields), `task append` (body) —
  mirror these.
- The picker (`resolveOne` + `auditOptions`) and editor flow are in place from the
  2026-06-25 pickers/width work.
- Audit frontmatter is `area` + `date` (bucket = directory); the body holds
  findings, parsed by `domain.ParseFindings` and checked by `audit lint`.

Relates to epic 20. Siblings: epic-mutation, lint-covers-epics.

## Acceptance criteria

- [ ] `audit edit [slug]` — $EDITOR on the audit file, picker on bare invocation,
      exit-11 non-interactively, re-validated on save (run `audit lint` on the
      result and surface issues). Mirror `task edit` (incl. `--dry-run` rejection).
- [ ] `audit set <slug> [--area|--date]` — surgical, validated (date YYYY-MM-DD),
      atomic, `--dry-run`, JSON envelope; picker on bare invocation.
- [ ] `audit append <slug> --body|--body-file|-` — append markdown to the body
      (a finding section), like `task append`; picker on bare invocation.
- [ ] go build/test/lint green; docs/cli regenerated; schema_version bumped if an
      envelope/field is added.

## Implementation sketch

- `core.SetAuditFields` / `AppendAuditBody`, mirroring the task ones; store
  `GetAudit`/atomic write helpers already exist.
- `audit.go`: `newAuditEditCmd`/`newAuditSetCmd`/`newAuditAppendCmd`; reuse
  `resolveOne`, `auditOptions`, `fillSelect`, and the editor wiring.

## Risks / gotchas

- **Slug = `date-area`**, which is the filename. `set --area`/`--date` would imply a
  RENAME — decide: either update frontmatter only and leave the slug (simplest, but
  the slug then disagrees with area/date), or rename the file (more correct, more
  work). Recommend documenting the choice; a rename touches completion + open handles.
- `append` should leave finding *correctness* to `audit lint` (append raw markdown,
  like `task append`) rather than trying to validate the finding grammar inline.
- `edit` re-validation: a non-open audit with a still-open finding must surface the
  bucket↔state lint issue (don't silently accept).

## Done when

`audit edit`, `audit set`, and `audit append` all work — validated, atomic, JSON,
picker — so authoring/fixing an audit no longer means leaving the tool.

## Decision 2026-06-25 — drop 'audit set', date is immutable

Scope to **audit edit + audit append only**. The audit 'date' (and 'area') ARE the slug (<date>-<area>), so they're immutable identity — there is no 'audit set' (the slug-rename problem dissolves). A mutable 'updated' timestamp is split into [[add-a-mutable-updated-timestamp-to-audits]]. Retitle/rescope this task to edit+append when picked up.
