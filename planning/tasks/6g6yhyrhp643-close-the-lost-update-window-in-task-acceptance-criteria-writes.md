---
schema: 1
id: 6g6yhyrhp643
status: completed
epic: 21-code-quality-architecture-hardening
description: Make task ac derive criterion edits from the same task-body snapshot protected by the store CAS, then retry transient conflicts without losing concurrent writes.
effort: S
tier: 1
priority: high
autonomy_level: 4
tags: [store, core, occ, concurrency]
created: "2026-09-04"
updated_at: "2026-09-04"
started_at: "2026-09-04"
audit_sources: [planning/audits/6g6qvrj15x97-2026-09-04-concurrency-and-atomicity.md]
completed_at: "2026-09-04"
---
# Close the lost-update window in task acceptance-criteria writes

## Objective

Eliminate the stale read-to-replace gap in every `task ac` mutation. The store must apply a pure body transform to the exact task snapshot guarded by its content CAS, while the core retries a rejected snapshot by reapplying the semantic edit to fresh content.

## Acceptance criteria

- [x] The task-store port exposes a non-interactive body-transform operation whose input is the current persisted body and whose output is covered by the same snapshot CAS.
- [x] Checkbox, explicit-state, and add/remove/replace criterion mutations use that operation inside the bounded `retryOnConflict` policy.
- [x] A concurrent task-body append is preserved when a criterion edit races it; the criterion edit either reapplies successfully or returns `domain.ErrConflict` after the retry bound.
- [x] Dry runs validate and compute without writing, and an already-satisfied edit remains a no-op without changing `updated_at`.
- [x] Focused store/core tests cover callback errors, no-op behavior, CAS rejection, retry, and the original lost-update interleaving.

## Out of scope

- Changing interactive `task edit` behavior.
- Repository-lock waiting policy or audit-finding mutation, which are tracked separately.

## Related

- Epic [21-code-quality-architecture-hardening](../epics/21-code-quality-architecture-hardening.md)
- Audit [2026-09-04 concurrency and atomicity](../audits/6g6qvrj15x97-2026-09-04-concurrency-and-atomicity.md)

## Implementation

Added an adapter-neutral `TransformTaskBody` port operation that applies pure semantic edits to the exact persisted body snapshot covered by the store content CAS. All `task ac` mutation forms share one core helper with bounded conflict retry and one captured timestamp. Regression coverage injects the original append-between-read-and-write interleaving and proves both the concurrent note and criterion transition survive; additional tests cover no-op, callback failure, dry-run, retry success, and retry exhaustion.

Validation: focused core/store tests, the full race-enabled suite, golangci-lint, module tidy check, generated CLI docs check, planning lint, and `git diff --check` pass.
