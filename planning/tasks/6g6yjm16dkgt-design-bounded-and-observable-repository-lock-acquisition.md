---
schema: 1
id: 6g6yjm16dkgt
status: ready-to-start
epic: 21-code-quality-architecture-hardening
description: Define an adapter-neutral cancellation and progress contract for both process-local and platform repository-lock waits before changing blocking behavior.
effort: M
tier: 3
priority: medium
autonomy_level: 3
tags: [store, concurrency, architecture, ux]
created: "2026-09-04"
audit_sources: [planning/audits/6g6qvrj15x97-2026-09-04-concurrency-and-atomicity.md]
updated_at: "2026-09-04"
---
# Design bounded and observable repository lock acquisition

## Objective

Make repository-lock contention bounded and diagnosable without coupling the filesystem store to terminal output. The contract must cover both the same-process guard and the cross-process platform lock so CLI, TUI, and future adapters can present or cancel waits consistently.

## Acceptance criteria

- [ ] Record the current acquisition semantics and decide the default wait, timeout, cancellation, and error-classification contract for cooperating mutations.
- [ ] One adapter-neutral mechanism carries wait progress and cancellation/deadline intent across the application/store boundary; the store does not receive an `io.Writer` or emit presentation text.
- [ ] The design covers both `repositoryGuard.write` and the platform lock rather than making only Unix `flock` observable.
- [ ] Unix and unsupported-platform behavior remain explicit, including how timeout or cancellation maps to `domain.ErrConflict` or another typed error.
- [ ] Deterministic tests hold each lock layer, observe the wait signal, cancel or time out the waiter, and prove all handles and process guards are released.
- [ ] Normal uncontended mutations retain their current serialization and output behavior.

## Stress tests

- Multiple `FS` values for one canonical root, separate processes, a long guarded Thread apply, cancellation racing lock release, and a non-CLI consumer with no progress renderer.

## Out of scope

- Per-file locks, lock-free graph mutations, or weakening the repository-wide mutation guard.
- Printing directly from `internal/store`.
- Treating non-cooperating raw filesystem edits as lock participants.

## Related

- Epic [21-code-quality-architecture-hardening](../epics/21-code-quality-architecture-hardening.md)
- Audit [2026-09-04 concurrency and atomicity](../audits/6g6qvrj15x97-2026-09-04-concurrency-and-atomicity.md)
