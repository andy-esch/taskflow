---
schema: 1
id: 6g392b0rps7w
status: ready-to-start
epic: 21-code-quality-architecture-hardening
description: SetFindingStatus rides the interactive editor loop and skips the content-hash CAS every other programmatic mutation performs.
effort: S
tier: 3
priority: medium
autonomy_level: 3
tags: [store, core, occ, concurrency]
created: "2026-08-24"
updated_at: "2026-08-24"
---
# Give finding-status writes a non-interactive CAS path

## Objective

`core.SetFindingStatus` routes through `store.EditAudit`, which is `editFile` — the loop
built for interactive `$EDITOR` sessions. Every other programmatic mutation
(`SetTaskFields`, `MoveAudit`, `AppendAuditBody`, `SetCriterionState`) instead performs a
content-hash compare-and-swap wrapped in `retryOnConflict`.

Two consequences, from audit finding M1 of
[2026-08-24-planning-state-vocabulary](../audits/6g38xqpa1ef6-2026-08-24-planning-state-vocabulary.md):

- **No OCC.** `editFile` takes its advisory lock only after the edit callback returns and
  re-checks that the file has not MOVED, but does not verify the content hash. A
  concurrent automated write between read and write is a lost update. The interactive path
  can live with that — a human editor session is long and a person sees the result — but a
  scripted `audit finding --status` in a loop cannot.
- **A workaround stands in for the mechanism.** Because `editFile` re-invokes its callback
  on a rejected edit, the current code passes a closure holding an `attempted` bool to
  refuse the second call. That is a guard against an editor affordance this caller never
  wanted, not conflict handling.

## Acceptance criteria

- [ ] A non-interactive audit body-replace exists on the store (sibling to
  `AppendAuditBody`), taking the write lock across read→verify→write and performing the
  same content-hash CAS the other programmatic mutations use.
- [ ] `core.SetFindingStatus` uses it inside `retryOnConflict`, and the `attempted` closure
  flag disappears — the mechanism replaces the workaround rather than sitting beside it.
- [ ] A conflicting concurrent write surfaces `domain.ErrConflict` (exit 14) rather than
  silently winning or losing.
- [ ] `--dry-run` still validates and computes without taking the lock.
- [ ] A test drives two writers at one audit and asserts the loser is told, mirroring
  whatever the existing OCC tests do rather than inventing a second shape.

## Out of scope

- Changing `editFile` itself, or the interactive `audit edit` path — the editor loop is
  correct for the editor.
- Extending CAS to any other surface that currently lacks it; if there are more, that
  survey is its own task.

## Related

- Epic [21-code-quality-architecture-hardening](../epics/21-code-quality-architecture-hardening.md)
- Audit finding M1: [2026-08-24-planning-state-vocabulary](../audits/6g38xqpa1ef6-2026-08-24-planning-state-vocabulary.md)
- Introduced by [audit-finding-write-surface-status-write-and-candidate-list-sync](6feeygw00jmx-audit-finding-write-surface-status-write-and-candidate-list-sync.md)
