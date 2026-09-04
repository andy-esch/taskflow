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
updated_at: "2026-09-04"
audit_sources: [planning/audits/6g6qvrj15x97-2026-09-04-concurrency-and-atomicity.md]
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

Cross-referenced by audit 2026-09-04-concurrency-and-atomicity: M1 (partial overlap). Two corrections that audit surfaced, for the human triaging this task — no status/scope change made here.

1. The 'No OCC' premise above is no longer accurate. `store.EditAudit` computes `ifVersion := hashContent(orig)` (internal/store/edit.go:295) and passes `verifyUnchanged` (:301) under `s.writeLock`. A probe racing `audit append` against `audit finding --status` twelve times lost zero writes and produced six correct exit-14s, so the third acceptance criterion already holds. What remains live is the `retryOnConflict` wrapper and the `attempted` closure flag at internal/core/finding.go:262.

2. The objective lists `SetCriterionState` among the mutations that already 'perform a content-hash compare-and-swap wrapped in retryOnConflict'. It does neither. Finding H1 of that audit shows the `task ac` writers compute the new body from a `GetTask` read that no CAS covers, then hand finished bytes to `EditBody`, which CASes only its own later read — measured at 3 lost appends in 12 races. Please do not read this task's comparison as evidence that the task ac path is safe.
