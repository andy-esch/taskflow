---
schema: 1
id: 6g7s6hr3qnfq
status: ready-to-start
epic: 28-first-class-entities-new-planning-nouns
description: CreateResearch scans for duplicate IDs before taking the write lock, so concurrent different-slug creates can both commit the same stable ID.
effort: 1-2 hours
tier: 2
priority: medium
autonomy_level: 4
tags: [research, identity, concurrency, store, robustness]
created: "2026-09-07"
---
# Serialize research ID collision checks with creation

## Objective

Close the race exposed while hardening audit-ID creation. `CreateResearch` correctly detects an
existing same-kind stable ID, but it performs that scan before `writeNewFile` acquires the
repository lock. Two concurrent creates with one ID and different slugs can therefore both pass the
scan and commit, producing the ambiguous and unwritable state the guard exists to prevent.

## Acceptance criteria

- [ ] The non-dry-run research ID scan and atomic create execute under one repository write lock.
- [ ] Dry-run performs the same semantic collision check without writing.
- [ ] A cooperating-writer race test starts two different-slug creates with one stable ID and proves
      exactly one succeeds, one returns `ErrConflict`, and one file exists.
- [ ] Existing collision diagnostics and day-derived ID regeneration behavior remain unchanged.
- [ ] Focused race tests, the full suite, lint, and diff hygiene pass.

## Out of scope

- Deciding planning-space-wide versus per-kind stable-ID uniqueness.
- Changing ID generation or adding automatic duplicate-ID repair.
- Refactoring every create path behind one abstraction without another concrete caller.

## Related

- Audit-ID hardening [detect-duplicate-audit-ids-at-creation-and-lint-time](6g7s4k845fsb-detect-duplicate-audit-ids-at-creation-and-lint-time.md)
- Prior research-ID fix [duplicate-stable-ids-brick-research-docs](6g1dnnfgyjap-duplicate-stable-ids-brick-research-docs-silently-and-unrecoverably.md)
- Epic [28-first-class-entities-new-planning-nouns](../epics/28-first-class-entities-new-planning-nouns.md)
