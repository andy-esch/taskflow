---
schema: 1
id: 6ggkdbg0816h
status: completed
epic: 21-code-quality-architecture-hardening
description: Show core-owned lifecycle remedy and Thread impact warnings even when no downstream task impacts exist.
effort: 1-2 hours
tier: 2
priority: medium
autonomy_level: 3
tags: [tui, recovery, threads]
audit_sources: [2026-10-04-core-owned-dependency-impact-recovery-implementation-codex]
created: "2026-10-04"
updated_at: "2026-10-10"
depends_on: [6ggdkztshmzz]
started_at: "2026-10-10"
completed_at: "2026-10-10"
---
# Preserve Thread-only lifecycle recovery guidance in the TUI

## Objective

Keep the core lifecycle receipt visible in TUI feedback even when a change affects only a Thread projection. Rendering must not decide whether recovery advice matters from the downstream task count.

## Evidence

[Codex core impact/recovery implementation review](../audits/6ggjnz40teda-2026-10-04-core-owned-dependency-impact-recovery-implementation-codex.md), M1, reproduces a pre-existing defect. Reopen the only completed member of a completed Thread to ready-to-start, with no descendants. The real moveTask command returns a committed receipt with zero task impacts, one newly inconsistent Thread impact, and a nonempty core remedy. Model.Update reports only the task movement because internal/tui/model.go nests remedy output inside len(msg.lifecycle.Impacts) > 0.

The implementation owner independently reproduced the failing real-filesystem command/message/update path on 2026-10-04. This is not a new defect introduced by the core ownership task, nor a data-safety blocker for that refactor closure.

## Scope

- Display any nonempty lifecycle Remedy independently of task-impact counts.
- Summarize changed Thread projections and newly inconsistent warnings using the core-owned decision, without new gate/string policy.
- Preserve committed cleanup warnings, flash severity conventions, and reload behavior.
- Keep it a bounded feedback fix, not a new receipt framework, persistent notification system, unreadable-marker redesign, or TUI navigation redesign.

## Acceptance criteria

- [x] A real sole-member completed-Thread reopen retains the exact core remedy when task impacts are empty and the task transition is durable.
- [x] Thread-only, task-only, combined, and empty receipt cases have accurate summaries without invented advice; Thread warning meaning comes from core.
- [x] A committed cleanup failure still surfaces its error and core guidance and schedules a reload.
- [x] An exact regression probe reintroducing the task-count guard fails the Thread-only assertion; restored tests pass.
- [x] Focused TUI tests, full tests/race/lint/build, and planning/audit lint pass; relevant consumer documentation is updated if needed.

## Sequencing

Follow the core-owned impact/recovery contract (6ggdkztshmzz) in the TUI navigation/refinement Thread. Do not add this pre-existing consumer bug to the adapter-neutral refactor closure path.

## Related

- [Core impact/recovery contract](6ggdkztshmzz-make-dependency-impact-and-recovery-semantics-core-owned.md)
- [TUI refinement Thread](../threads/6g90h4pg0q7n-refine-thread-and-tui-navigation.md)

## Implementation and self-review (2026-10-10)

- TUI movement feedback treats downstream task impacts, Thread impacts and the nonempty core Remedy independently. Thread projection counts and newly inconsistent warning counts use ThreadProjectionImpact.NewlyInconsistent; the adapter does not invent recovery policy.
- Existing successful-move severity, revisit text, cursor restoration and reload behavior remain intact. A committed cleanup failure still carries its original error and exact core guidance, reloads both task and loaded Thread views, and is never retried.
- TestTUITaskReopenSurfacesThreadOnlyRecovery drives the actual task action menu and scoped command/message/update path against a real filesystem adapter. Reopening a completed sole member creates zero task impacts and one newly inconsistent Thread; the task is durable and the Thread document is byte-for-byte unchanged. The post-commit conflict fixture injects only cleanup failure after the real adapter commits.
- TestTUILifecycleFeedbackUsesIndependentReceiptFields covers task-only, Thread-only, combined, empty/nil receipts, standalone remedy, already-inconsistent and restored Thread cases, revisit text and clearing a previous error flash.
- In an isolated source copy, restoring the exact old outer guard (lifecycle != nil && len(lifecycle.Impacts) > 0) compiled and made the real sole-member test fail at the missing feedback assertion. Restoring the fix passed. Focused and full TUI tests, just test (race), just lint, just build, just docs-check, planning/audit lint and diff whitespace checks passed.
- The architecture recovery contract and reducer comment document the ownership rule. No core port, persisted entity or machine-contract shape changed. This bounded presentation fix was self-reviewed; no external review is required for completion.

## Presentation boundary

The existing one-line footer still clips long feedback on narrow terminals. Full core text is retained, and a wide-footer assertion proves it is rendered when space allows; this does not claim complete recovery text is readable at every width. Evidence and presentation options are now recorded in [the queued TUI navigation design](6g8vxcnezktm-reassess-tui-navigation-and-information-architecture-at-current-scale.md), rather than introducing a notification framework in this fix. Completion records local implementation and validation, not a merge or release.
