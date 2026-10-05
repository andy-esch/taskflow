---
schema: 1
id: 6ggkdbg0816h
status: next-up
epic: 21-code-quality-architecture-hardening
description: Show core-owned lifecycle remedy and Thread impact warnings even when no downstream task impacts exist.
effort: 1-2 hours
tier: 2
priority: medium
autonomy_level: 3
tags: [tui, recovery, threads]
audit_sources: [2026-10-04-core-owned-dependency-impact-recovery-implementation-codex]
created: "2026-10-04"
updated_at: "2026-10-04"
depends_on: [6ggdkztshmzz]
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

- [ ] A real sole-member completed-Thread reopen retains the exact core remedy when task impacts are empty and the task transition is durable.
- [ ] Thread-only, task-only, combined, and empty receipt cases have accurate summaries without invented advice; Thread warning meaning comes from core.
- [ ] A committed cleanup failure still surfaces its error and core guidance and schedules a reload.
- [ ] An exact regression probe reintroducing the task-count guard fails the Thread-only assertion; restored tests pass.
- [ ] Focused TUI tests, full tests/race/lint/build, and planning/audit lint pass; relevant consumer documentation is updated if needed.

## Sequencing

Follow the core-owned impact/recovery contract (6ggdkztshmzz) in the TUI navigation/refinement Thread. Do not add this pre-existing consumer bug to the adapter-neutral refactor closure path.

## Related

- [Core impact/recovery contract](6ggdkztshmzz-make-dependency-impact-and-recovery-semantics-core-owned.md)
- [TUI refinement Thread](../threads/6g90h4pg0q7n-refine-thread-and-tui-navigation.md)
