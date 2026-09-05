---
schema: 1
id: 6g392b0rps7w
status: completed
epic: 21-code-quality-architecture-hardening
description: Give audit finding edits a store-owned body-transform CAS and bounded retry path, replacing their remaining interactive-editor workaround without risking concurrent audit writes.
effort: S
tier: 3
priority: medium
autonomy_level: 3
tags: [store, core, occ, concurrency]
created: "2026-08-24"
updated_at: "2026-09-05"
audit_sources: [planning/audits/6g7731f5kjkw-2026-09-05-tool-owned-sub-entity-writes-implementation-claude.md, planning/audits/6g7731f8zzjq-2026-09-05-tool-owned-sub-entity-writes-implementation-antigravity.md]
started_at: "2026-09-05"
completed_at: "2026-09-05"
---
# Give finding-status writes a non-interactive CAS path

## Objective

Move scripted `audit finding` edits off the interactive `$EDITOR` callback loop. `store.EditAudit` now correctly detects concurrent content changes, so this is no longer a lost-update fix: the remaining work is to give audit finding edits a non-interactive transform operation and bounded automatic retry equivalent to other agent-facing mutations.

## Current state

The earlier premise that `EditAudit` lacked content OCC is obsolete. It hashes the bytes supplied to the callback and verifies that version under the write lock; a conflicting writer already returns `domain.ErrConflict` without losing either write. `core.EditFinding` still carries an `attempted` closure flag solely to escape the editor retry loop, and callers still receive transient conflicts that sibling scriptable verbs retry automatically.

## Acceptance criteria

- [x] The audit-store port exposes a non-interactive body-transform operation whose input is derived from the exact audit snapshot protected by its content CAS.
- [x] `core.EditFinding` uses that operation inside `retryOnConflict`, and its interactive-loop `attempted` closure flag disappears.
- [x] A race between `audit append` and `audit finding` preserves both changes; transient conflicts are reapplied, while exhaustion still surfaces `domain.ErrConflict` (exit 14).
- [x] `--dry-run` validates and computes without locking or writing, and an already-applied finding edit remains a no-op.
- [x] Focused store/core tests pin callback errors, no-op behavior, CAS rejection, retry, and concurrent append preservation.

## Out of scope

- Changing `EditAudit` or the interactive `audit edit` loop.
- Task acceptance-criteria mutation, owned by `close-the-lost-update-window-in-task-acceptance-criteria-writes`.
- A generic callback abstraction spanning unrelated entity types unless the implementations prove genuinely identical.

## Related

- Epic [21-code-quality-architecture-hardening](../epics/21-code-quality-architecture-hardening.md)
- Audit [2026-09-04 concurrency and atomicity](../audits/6g6qvrj15x97-2026-09-04-concurrency-and-atomicity.md)
- Task [Close the task acceptance-criteria lost-update window](6g6yhyrhp643-close-the-lost-update-window-in-task-acceptance-criteria-writes.md)
