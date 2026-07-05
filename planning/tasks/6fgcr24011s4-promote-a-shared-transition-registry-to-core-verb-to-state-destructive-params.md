---
schema: 1
status: completed
epic: 21-code-quality-architecture-hardening
description: One core verb→state registry (domain.Transition) the CLI+TUI consume, carrying the destructive flag and an optional-date param marker so defer stops being special-cased per adapter.
effort: L
tier: 3
priority: high
autonomy_level: 3
tags: [architecture, core]
created: "2026-06-27"
updated_at: "2026-06-29"
completed_at: "2026-06-29"
id: 6fgcr24011s4
---
Audit 2026-06-27-consumer-data-flow-architecture H3 (+M4,+M5). One verb->destination table in core/domain that CLI/TUI/web all consume (the missing peer of AllStatuses/AllAuditBuckets). Carry the destructive flag and an optional param spec so defer's revisit-date stops being special-cased in three layers and the destructive-confirm signal is shared. Builds on completed make-tui-lifecycle-action-machinery-registry-driven.

**Update 2026-06-28:** M5's clock-consistency half landed — the revisit-date relative-offset parse now uses svc.Now() (the injected clock) not time.Now(), at all 3 sites (cli/task.go, tui/edit.go, tui/model.go). Remaining in this task: the H3 shared transition registry + M4 defer-param + M5 destructive-confirm consolidation.

**Progress 2026-06-28 (M4 atomicity half).** `DeferTask` is now one atomic write: a new `Store.Defer` port method records `revisit_at` inside the same relocation write (shared `moveTask` in fsstore.go; a re-defer rewrites it in place), retiring the Move-then-SetFields two-write hazard; the core also validates the date up front. Pinned by TestFS_Defer + TestDeferTask_*. Remaining in this task: the H3 shared verb->state registry + M4`s defer-param-on-descriptor (so `defer` stops being special-cased per adapter) + M5`s destructive-confirm consolidation.

## Completed 2026-06-29

All three audit findings are now closed:

- **H3 (shared registry) — landed earlier (in main).** `domain.Transition` +
  `TaskTransitions()`/`AuditTransitions()` are the one verb→state table; the CLI
  builds its verb commands from them and the TUI action menu is a thin
  `fromDomain(...)` view. (Epics stay TUI-local by design — they have no CLI verb
  vocabulary yet; folding them in belongs with [[epic-lifecycle-named-verbs-epic-retire-deprecate-activate]].)
- **M5 (destructive-confirm) — landed earlier (in main).** `Transition.Destructive`
  lives in the registry and flows to the TUI's y/n gate via `fromDomain`; the
  non-interactive CLI ignores it. The clock-consistency half also already landed.
- **M4 (defer-param on the descriptor) — this change.** Added a minimal typed
  marker `TransitionParam` (`ParamNone` / `ParamOptionalDate`) + a `Param` field on
  `domain.Transition`; the task `defer` row carries `ParamOptionalDate`. The two
  adapters now read the registry signal instead of hardcoding which verb is
  special: the CLI routes to its `--until` builder on `tr.Param == ParamOptionalDate`
  (was `tr.Verb == "defer"`), and the TUI opens its revisit-date prompt on
  `tr.optionalDate` (was `kind == tasks && tr.to == StatusDeferred`, at both the
  menu-enter and `beginTransition` sites). HOW each collects the date (a flag vs a
  TUI widget) stays adapter-specific — irreducible — so a full param *spec* would be
  over-build; the marker is the honest minimum. Pinned by the extended
  TestTaskTransitions/TestAuditTransitions (defer carries the param, audits never
  do). build/vet/test/lint green; only schema_comments.json moved among generated
  files (the new TransitionParam doc comment).
