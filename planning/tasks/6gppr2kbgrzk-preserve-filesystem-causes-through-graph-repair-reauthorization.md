---
schema: 1
id: 6gppr2kbgrzk
status: ready-to-start
epic: 21-code-quality-architecture-hardening
description: MutateTaskGraphRepair's per-write loop turns genuine read failures into retryable ErrConflict and drops the cause; wrap the real error and keep conflict for evidence mismatch.
effort: 2-3 hours
tier: 2
priority: medium
autonomy_level: 3
tags: [store, errors, recovery]
created: "2026-10-23"
audit_sources: [planning/audits/6gn83987gt45-2026-10-19-arch-failure-and-recovery.md]
updated_at: "2026-10-23"
---
# Preserve filesystem causes through graph-repair reauthorization

## Objective

Five checks in `MutateTaskGraphRepair` read `if evidenceErr != nil || verify(...) != nil { return
ErrConflict }`. A permission or I/O failure is therefore reported as a concurrent edit: it burns
the retry budget, loses the structured filesystem block, and exits 14. The same function's
preflight and every sibling mutation family already keep the two apart. Make the loop match them.

## Acceptance criteria

- [ ] Each load or read error in the repair loop is returned wrapped with `%w`, and `ErrConflict` is reserved for an actual snapshot or evidence mismatch.
- [ ] A regression injects a non-conflict read failure during repair and asserts no retry, the original cause in the chain, and a populated `error.filesystem` block in `--json`.
- [ ] A guard-removal probe of the split fails the new test.
- [ ] Genuine concurrent drift during repair still reports a conflict.

## Out of scope

- A shared "a load failure is not a conflict" helper across all families, and commit-aware retry hints (owned by the integrity-foundations design).

## Related

- Epic [21-code-quality-architecture-hardening](../epics/21-code-quality-architecture-hardening.md)
- Audit [2026-10-19 failure and recovery](../audits/6gn83987gt45-2026-10-19-arch-failure-and-recovery.md), finding H1
- Same defect class fixed in `cas.go` by [6g1dpb5gwtt9](6g1dpb5gwtt9-verifyunchanged-collapses-every-resolve-error-into-a-futile-retry-conflict.md)
