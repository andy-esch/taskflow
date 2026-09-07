---
schema: 1
id: 6g77rn6n6b86
status: next-up
epic: 21-code-quality-architecture-hardening
description: Five merged behaviours survive deletion with the suite green; two doc comments were captured by inserted declarations, one already in a generated artifact.
effort: 4-6 hours
tier: 3
priority: medium
autonomy_level: 3
tags: [audit, store, tests, hygiene]
created: "2026-09-05"
---

# Pin the unpinned audit write and lint behaviours, and unify the body-transform twins

## Objective

Lock down the merged audit body-write contract before removing its duplicated implementation.
Mutation review found five observable or operational behaviors that can be deleted while the full
suite stays green, and a second review showed that `TransformAuditBody`'s parsed return value and
parse-before-write boundary are also unasserted. Add focused tests that fail for those exact
mutations, repair the two doc comments captured by newly inserted declarations and regenerate their
derived artifact, then extract the proven-common task/audit transform sequence behind thin typed
wrappers without changing its concurrency, dry-run, validation, or error semantics.

Treat this as two ordered review units: behavior/documentation pinning first, shared-helper
extraction second. The refactor is justified only after the tests can distinguish each entity's
noun, resolver, parser, and compare-and-swap recheck—the substitutions most likely to be crossed
while unifying the twins.

## Acceptance criteria

- [ ] `EditAudit` has a direct test proving that a newly introduced near-miss finding header is
  rejected while an unrelated edit to a body with pre-existing drift remains allowed; deleting the
  edit guard makes the test fail.
- [ ] `IntroducedNearMissHeaders` is pinned by occurrence text rather than line number: inserting
  content above an existing near miss does not make it new, while adding another occurrence of the
  same heading is reported exactly once.
- [ ] Rendering tests exercise the truly empty-audit branch so changing the `Findings == 0`
  condition back to the old percentage path fails. They avoid freezing the separate
  `6g77rn6em6n8` decision about how unparsed headings should be presented.
- [ ] `FixFindingHeaders` performs no body transform or write for a clean audit, demonstrated with
  a counting/spying store rather than inferred only from byte-identical output.
- [ ] The audit transform's fence/validation failure names an `audit`, not a `task`, and a targeted
  test fails if the noun passed to the shared body writer is crossed.
- [ ] `TransformAuditBody` tests assert that the returned `domain.Audit` is the parsed post-transform
  entity (including stable identity and updated finding counts) and that an invalid transformed
  document is rejected before any disk write.
- [ ] The comments for `NearMissHeader`, `LintFindings`, `auditProgressCell`, and `auditStateNote`
  each document the declaration they precede. `internal/wire/schema_comments.json` is regenerated
  and no longer attributes `LintFindings` semantics to `NearMissHeader`.
- [ ] Only after the preceding pins are green, `TransformAuditBody` and `TransformTaskBody` delegate
  their common read/transform/normalize/timestamp/lock/CAS/write sequence to one internal typed
  helper; the public/store-port method signatures and entity-specific wrappers remain unchanged.
- [ ] Shared table-driven tests run the helper contract through both task and audit wrappers,
  covering callback failure, no-op, dry-run, timestamp stamping, parse failure, concurrent-content
  conflict, and a successful populated return value without weakening the existing focused tests.
- [ ] The behavior-pinning and helper-extraction changes are separate logical commits, and the full
  race-enabled test suite, lint, module-tidiness check, generated-artifact check, and
  `git diff --check` all pass.

## Out of scope

- Changing which headings count as near-miss findings; `6g77rn6b9wf8` owns classifier semantics.
- Changing the user-facing distinction between empty and unparsed audits; `6g77rn6em6n8` owns that
  rendering/data contract.
- Guarding audit creation or repairing `EditAudit`'s unreadable-frontmatter baseline; those write
  gaps are owned by `6g77rn6hvmh8`.
- Generalizing finding near-miss detection for candidate-task markers; `6g3ag8py12y9` owns that
  sub-entity convention.
- Introducing a repository-wide persistence abstraction or changing atomic-write durability,
  retry ceilings, lock scope, or CAS semantics.

## Related

- Epic [21-code-quality-architecture-hardening](../epics/21-code-quality-architecture-hardening.md)
- Predecessor: [give finding status writes a non-interactive CAS path](6g392b0rps7w-give-finding-status-writes-a-non-interactive-cas-path.md)
- Independent implementation audits:
  [Claude](../audits/6g7731f5kjkw-2026-09-05-tool-owned-sub-entity-writes-implementation-claude.md)
  and
  [Antigravity](../audits/6g7731f8zzjq-2026-09-05-tool-owned-sub-entity-writes-implementation-antigravity.md)
- Adjacent ownership:
  [recognizer semantics](6g77rn6b9wf8-narrow-the-near-miss-finding-recognizer-and-re-measure-across-every-entity-type.md),
  [audit read surfaces](6g77rn6em6n8-report-unparsed-findings-on-the-audit-read-surfaces.md),
  [remaining write guards](6g77rn6hvmh8-close-the-remaining-audit-body-write-guard-gaps.md), and
  [candidate-task convention](6g3ag8py12y9-decide-the-candidate-list-convention-and-make-the-tool-own-it.md)
