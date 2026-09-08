---
schema: 1
id: 6g7wxs43g7nh
status: ready-to-start
epic: 24-data-model-evolution-stable-key-storage-read-model-content-occ
description: Guard RenameTask planning and cascade writes with repository snapshots/CAS so concurrent renames cannot duplicate IDs or discard cooperating writes.
effort: 1-2 days
tier: 2
priority: high
autonomy_level: 3
tags: [rename, concurrency, occ, store, hardening]
created: "2026-09-07"
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

- [ ] Concurrent renames of one task to different slugs cannot both commit; disk state contains one stable-ID owner and the loser receives an attributable conflict.
- [ ] A concurrent guarded task field/body mutation is either included by the rename plan or causes the rename to conflict; no successful cooperating write is silently lost.
- [ ] The target filename is rechecked after acquiring the authoritative guard (or protected by an equivalent CAS).
- [ ] Every cascade file write is authorized against the bytes used to plan it.
- [ ] Partial multi-file failure semantics and operator recovery diagnostics are explicit and tested.
- [ ] Existing link-cascade, no-op, dry-run, and exact-target collision behavior remains covered.
- [ ] Race-enabled focused tests and the full validation suite pass.

## Out of scope

- A universal backend-independent transaction framework.
- Reworking the Scheme-2 link format.
- Serializing or preventing raw-editor and Git-merge conflicts beyond detecting stale state at the guarded mutation boundary.

## Evidence

- [Shared entity create guard Claude audit](../audits/6g7wqndnnvs2-2026-09-07-shared-entity-create-guard-implementation-claude.md), finding M1.
- [Earlier target collision guard](6fkkhs47abx1-guard-renametask-against-target-filename-collision.md).
- [Next shared entity-integrity foundations](6g7wsr8yvt1a-define-the-next-shared-entity-integrity-foundations.md).
