---
schema: 1
id: 6g7wxs43g7nh
status: completed
epic: 24-data-model-evolution-stable-key-storage-read-model-content-occ
description: Guard RenameTask planning and cascade writes with repository snapshots/CAS so concurrent renames cannot duplicate IDs or discard cooperating writes.
effort: 1-2 days
tier: 2
priority: high
autonomy_level: 3
tags: [rename, concurrency, occ, store, hardening]
created: "2026-09-07"
updated_at: "2026-09-08"
started_at: "2026-09-07"
completed_at: "2026-09-08"
---
# Guard RenameTask against stale cascade plans and concurrent identity duplication

## Objective

Make `RenameTask` authorize and commit its target rename and inbound-link cascade against one guarded repository snapshot so concurrent cooperating mutations cannot create duplicate task identities or silently lose updates.

## Demonstrated hazards

The shared-create adversarial review reproduced two pre-existing failures:

- Two concurrent renames of one task can both pass the target check and leave two filenames carrying the same stable task ID.
- A concurrent guarded field edit can report success and then be overwritten by the rename cascade compiled from stale bytes.

The earlier target-filename collision guard only checks the target before the repository lock. `RenameTask` then walks and plans all edits before acquiring that lock, and its clobbering atomic writes have no content CAS.

## Scope

- Put the target-path decision and cascade authorization under the canonical repository guard, or introduce an equivalent snapshot plus per-file CAS/recheck that closes the same race.
- Preserve Scheme-2 task identity and inbound-link cascade behavior.
- Fail loudly and attributably before corrupting identity when the target or source snapshot changes.
- Define the multi-file partial-durability/receipt behavior if a later cascade write fails.
- Keep dry-run non-durable while ensuring its diagnostics do not overclaim a reservation.
- Add deterministic coordinated concurrency tests; do not rely on timing or scheduler luck.

## Acceptance criteria

- [x] Concurrent renames of one task to different slugs cannot both commit; disk state contains one stable-ID owner and the loser receives an attributable conflict.
- [x] A concurrent guarded task field/body mutation is either included by the rename plan or causes the rename to conflict; no successful cooperating write is silently lost.
- [x] The target filename is rechecked after acquiring the authoritative guard (or protected by an equivalent CAS).
- [x] Every cascade file write is authorized against the bytes used to plan it.
- [x] Partial multi-file failure semantics and operator recovery diagnostics are explicit and tested.
- [x] Existing link-cascade, no-op, dry-run, and exact-target collision behavior remains covered.
- [x] Race-enabled focused tests and the full validation suite pass.

## Out of scope

- A universal backend-independent transaction framework.
- Reworking the Scheme-2 link format.
- Serializing or preventing raw-editor and Git-merge conflicts beyond detecting stale state at the guarded mutation boundary.

## Evidence

- [Shared entity create guard Claude audit](../audits/6g7wqndnnvs2-2026-09-07-shared-entity-create-guard-implementation-claude.md), finding M1.
- [Earlier target collision guard](6fkkhs47abx1-guard-renametask-against-target-filename-collision.md).
- [Next shared entity-integrity foundations](6g7wsr8yvt1a-define-the-next-shared-entity-integrity-foundations.md).
- [Claude implementation review](../audits/6g81f73jh6d9-2026-09-08-task-rename-snapshot-and-recovery-implementation-claude.md).
- [Antigravity implementation review](../audits/6g81f73v3q8f-2026-09-08-task-rename-snapshot-and-recovery-implementation-antigravity.md).

## Implementation progress (2026-09-08)

`RenameTask` now captures the caller's source version before waiting, takes the canonical repository guard, rechecks the target and source under that guard, and recompiles the cascade from current bytes. Cascade replacements and final source removal use content CAS; destination creation uses an exclusive no-clobber write. Deterministic coordinated tests prove that concurrent renames produce one winner, cooperating guarded edits survive, and raw target/cascade/source races fail closed.

Core now returns an adapter-neutral durable-prefix receipt. A retryable cascade prefix, the destination-written/source-retained inspection state, and a fully committed unlock failure have distinct tested recovery guidance. CLI JSON failures expose that receipt under `error.task_rename`; the wire schema is 1.62. ADR-0003 and the architecture guide record the ordering and recovery contract.

Validation: focused rename tests passed 20 repetitions; focused race tests passed 10 repetitions; `just test`, `just lint`, `just docs-check`, `just tidy-check`, planning lint, schema-comment freshness, wire validation, and CLI goldens pass.

## Adversarial review closeout (2026-09-08)

Two isolated reviews confirmed the guarded ordering and recovery model, then exposed gaps worth fixing. The suite now independently pins a source change while waiting, a raw source change before destination creation, a target appearing after the final source CAS, and a dangling-symlink target. Mutation probes proved that removing each guarded boundary makes its named regression test fail.

Moved destinations now retain the source file permission bits exactly; a 0600 regression fixture prevents widening. Successful receipts no longer contain failure-oriented recovery prose, and an exact title/slug no-op reports zero planned/applied documents without writing. The separate success/dry-run JSON receipt proposed by Claude is tracked by [expose-successful-task-rename-receipts-in-json](6g81npee5kv2-expose-successful-task-rename-receipts-in-json.md), sequenced after this task.

Post-review validation: the focused rename suite passed 20 repetitions under the race detector; `just test`, `just lint`, `just docs-check`, `just tidy-check`, planning lint, schema validation, and CLI goldens pass.
