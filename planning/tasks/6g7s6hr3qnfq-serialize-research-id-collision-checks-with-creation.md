---
schema: 1
id: 6g7s6hr3qnfq
status: completed
epic: 28-first-class-entities-new-planning-nouns
description: CreateResearch scans for duplicate IDs before taking the write lock, so concurrent different-slug creates can both commit the same stable ID.
effort: 2-4 hours
tier: 2
priority: medium
autonomy_level: 4
tags: [research, identity, concurrency, store, robustness]
created: "2026-09-07"
started_at: "2026-09-07"
updated_at: "2026-09-07"
completed_at: "2026-09-07"
---
# Serialize research ID collision checks with creation

## Objective

Close the race exposed while hardening audit-ID creation. `CreateResearch` correctly detects an
existing same-kind stable ID, but it performs that scan before `writeNewFile` acquires the
repository lock. Two concurrent creates with one ID and different slugs can therefore both pass the
scan and commit, producing the ambiguous and unwritable state the guard exists to prevent.

The implementation survey found that this is an ordinary entity-creation boundary, not a research
special case. Give the filesystem adapter one check/allocate-and-create transaction so current and
future single-file entities cannot accidentally place repository scans before lock acquisition.
Graph-aware compound creation keeps its core planner and outer guarded transaction.

## Acceptance criteria

- [x] The non-dry-run research ID scan and atomic create execute under one repository write lock.
- [x] Dry-run performs the same semantic collision check without writing.
- [x] A cooperating-writer race test starts two different-slug creates with one stable ID and proves
      exactly one succeeds, one returns `ErrConflict`, and one file exists.
- [x] Existing collision diagnostics and day-derived ID regeneration behavior remain unchanged.
- [x] Task, audit, research, and epic ordinary creation all use the shared guarded primitive;
      identifier scans and epic-number allocation happen inside its real-write critical section.
- [x] Ordinary task creation rejects a same-ID/different-slug task, including when the existing
      owner is unreadable but its filename retains a valid identity.
- [x] Compound Thread and create-and-start flows retain their graph-aware planners and reuse only
      the lock-compatible no-clobber write layer rather than being weakened to a generic callback.
- [x] Focused race tests, the full suite, lint, and diff hygiene pass.

## Out of scope

- Deciding planning-space-wide versus per-kind stable-ID uniqueness.
- Changing ID generation or adding automatic duplicate-ID repair.
- Replacing portable core planners or database-native uniqueness with filesystem locking policy.

## Related

- Audit-ID hardening [detect-duplicate-audit-ids-at-creation-and-lint-time](6g7s4k845fsb-detect-duplicate-audit-ids-at-creation-and-lint-time.md)
- Prior research-ID fix [duplicate-stable-ids-brick-research-docs](6g1dnnfgyjap-duplicate-stable-ids-brick-research-docs-silently-and-unrecoverably.md)
- Epic [28-first-class-entities-new-planning-nouns](../epics/28-first-class-entities-new-planning-nouns.md)

## Implementation evidence (2026-09-07)

`createEntityFile` now owns preparation and atomic creation for ordinary task, epic, audit, and
research files. Real writes allocate or scan identity while holding the canonical repository guard;
dry-run evaluates the same preparation without creating a root or claiming a reservation. Threads
and create-and-start tasks retain their core-planned guarded transactions and finish through the
same lock-compatible no-clobber primitive.

The shared pass also closed two adjacent holes: ordinary task creation now treats a same-ID task
under another slug—including an unreadable filename owner—as a conflict, and epic sequence-number
allocation is serialized with persistence. Regression coverage pins the shared preparation boundary,
research collision race, task identity ownership, epic allocation, and task/Thread cross-kind path.
The full race suite, Go lint, module tidiness, generated-doc check, planning/audit lint, and diff
hygiene all pass.

## Adversarial review closeout (2026-09-07)

Two isolated external reviews accepted the shared ordinary-create transaction after repeated race tests, exact mutation kills, symlink and cross-process probes, and a repository-wide caller inventory. The [Claude review](../audits/6g7wqndnnvs2-2026-09-07-shared-entity-create-guard-implementation-claude.md) found one adjacent repair defect: canonical ID repair could create a same-kind duplicate under another slug. That finding is fixed here by reusing filename identity ownership and refusing the repair without changing either file. Its separate concurrent `RenameTask` stale-plan/lost-update finding is tracked by [guard RenameTask against stale cascade plans](6g7wxs43g7nh-guard-renametask-against-stale-cascade-plans-and-concurrent-identity-duplication.md). The [Antigravity review](../audits/6g7wqne275wg-2026-09-07-shared-entity-create-guard-implementation-antigravity.md) reported no findings. Both audits are closed.

Post-review validation passed: `go test -race ./...`, `just lint`, `just tidy-check`, `just docs-check`, planning lint, audit lint, and `git diff --check`.
