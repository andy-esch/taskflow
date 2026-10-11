---
schema: 1
id: 6gjn7axv0gmy
status: completed
epic: 30-threads-and-task-dependency-graphs
description: Carry legacy dependency causes and migration guidance into shared Thread diagnostics without changing fail-closed eligibility.
effort: S
tier: 3
priority: medium
autonomy_level: 3
tags: [threads, diagnostics, ux]
created: "2026-10-11"
updated_at: "2026-10-11"
started_at: "2026-10-11"
completed_at: "2026-10-11"
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

- [x] A resolvable legacy field outside the Thread explains degraded health and names the supported migration command instead of reporting an unexplained empty frontier.
- [x] Shared projections and human/JSON Thread reads preserve portable legacy diagnostic identity and deterministic ordering without reloading task files.
- [x] Missing/ambiguous legacy references and structural graph defects retain their existing severity and do not print duplicate causes.
- [x] Healthy behavior, fail-closed dispatch, in-flight context, and external prerequisite roles remain unchanged.
- [x] Machine-contract changes are classified and schema/golden tests cover both legacy-only degradation and combined broken evidence.

## Sequencing

Independent diagnostic-comprehension slice in Refine Thread and TUI navigation; no artificial
dependency on the candidate-metadata implementation. It does not block that contained change.

## Implementation checkpoint — 2026-10-11

Shared Thread projections now carry repository-wide legacy occurrences from the same graph
snapshot. Lists hoist them once, including empty selections; paginated continuations bind that
context. Source IDs and opaque locations remain separate from optional local paths.

Human CLI and TUI diagnostics share field-level summaries and migration/repair guidance.
Missing/ambiguous references are grouped for display without hiding structural defects;
JSON retains both typed legacy evidence and the original graph problems. Machine revision
**1.87 is ADDITIVE**; persisted documents, eligibility, and mutation authorization are unchanged.
Typed lists and opt-in projected pages gain context; the legacy unbounded column projection
keeps its existing envelope shape. Mutation and repair receipts reuse the shared projection.

Regression coverage includes pathless and duplicate source occurrences, outside-Thread causes,
empty streams, stale cursors, populated JSON-schema validation, actual migration/refusal, and
real repair preview/commit receipts. Full race tests, lint, build, generated-doc checks, and
planning lint pass. Implementation is complete and both independent reviews are triaged;
the linked audit dispositions preserve the evidence and review limits for PR delivery.

Review briefs: [Codex](../audits/6gjqmf7c7jhj-2026-10-11-thread-legacy-degradation-diagnostics-implementation-codex.md)
and [Antigravity](../audits/6gjqmf7n3n89-2026-10-11-thread-legacy-degradation-diagnostics-implementation-antigravity.md).
Both require independent sandboxes and concrete negative/mutation evidence. Codex completed
its review; M1 identified a pre-existing terminal-sanitization gap in shared structural
diagnostics. Both messages and opaque locations are now sanitized only at the human renderer;
produced pathless/local-path regressions cover frontier, blockers, and unblocks while checking
that raw core/JSON evidence survives. M1 is fixed and the Codex audit closed. Antigravity reported
no new findings; owner triage records that its blanket terminal-safety claim missed M1 and its
transfer attestation remains incomplete. Both audits are closed with those limits preserved.

## Related

- [Candidate context](6g6wdvfp2ksa-make-thread-frontier-help-choose-among-independent-candidates.md)
- [Status/lint degradation reporting](6g697mp8s4tx-report-graph-degradation-in-status-and-lint.md)
- [Refine Thread and TUI navigation](../threads/6g90h4pg0q7n-refine-thread-and-tui-navigation.md)
