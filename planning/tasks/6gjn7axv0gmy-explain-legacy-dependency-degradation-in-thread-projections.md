---
schema: 1
id: 6gjn7axv0gmy
status: ready-to-start
epic: 30-threads-and-task-dependency-graphs
description: Carry legacy dependency causes and migration guidance into shared Thread diagnostics without changing fail-closed eligibility.
effort: S
tier: 3
priority: medium
autonomy_level: 3
tags: [threads, diagnostics, ux]
created: "2026-10-11"
---
# Explain legacy dependency degradation in Thread projections

## Objective

Explain why a Thread has no dispatchable frontier when resolvable legacy dependency fields
degrade the repository graph. Preserve fail-closed eligibility; improve the diagnosis, not the gate.

## Evidence

The degraded case in `TestThreadFrontierPlanningContextAcrossRealAdapter`
(`internal/cli/thread_test.go`) reproduces `graph_health=degraded`, an empty frontier,
and empty `problems`/`graph_problems` arrays. `TaskGraph` retains those causes in
`LegacyDiagnostics()`, but `projectThread` carries only `Problems()`.
Human Thread diagnostics likewise receive no legacy diagnostics. Status/lint already explain
these occurrences; reuse existing neutral diagnostic/summary contracts where suitable.

## Scope

Carry the cause and actionable migration guidance through the shared core Thread projection,
then expose it deliberately in human/JSON reads and relevant TUI health presentation.
Use the same graph snapshot, preserve portable record identity, and avoid duplicate repository
diagnostics in lists. Consider existing mutation-impact views so reads and receipts do not drift.
No adapter rescans, new storage capability, automatic migration, or eligibility changes.

## Acceptance criteria

- [ ] A resolvable legacy field outside the Thread explains degraded health and names the supported migration command instead of reporting an unexplained empty frontier.
- [ ] Shared projections and human/JSON Thread reads preserve portable legacy diagnostic identity and deterministic ordering without reloading task files.
- [ ] Missing/ambiguous legacy references and structural graph defects retain their existing severity and do not print duplicate causes.
- [ ] Healthy behavior, fail-closed dispatch, in-flight context, and external prerequisite roles remain unchanged.
- [ ] Machine-contract changes are classified and schema/golden tests cover both legacy-only degradation and combined broken evidence.

## Sequencing

Independent diagnostic-comprehension slice in Refine Thread and TUI navigation; no artificial
dependency on the candidate-metadata implementation. It does not block that contained change.

## Related

- [Candidate context](6g6wdvfp2ksa-make-thread-frontier-help-choose-among-independent-candidates.md)
- [Status/lint degradation reporting](6g697mp8s4tx-report-graph-degradation-in-status-and-lint.md)
- [Refine Thread and TUI navigation](../threads/6g90h4pg0q7n-refine-thread-and-tui-navigation.md)
