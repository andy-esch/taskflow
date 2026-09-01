---
schema: 1
id: 6g5rwjr0dz4p
status: next-up
epic: 30-threads-and-task-dependency-graphs
description: After list/detail dogfooding, present Thread topology in the TUI without duplicating traversal logic or making narrow terminals unusable.
effort: 1-2 days
tier: 3
priority: medium
autonomy_level: 3
tags: [threads, tui, graph, ux, dogfood]
created: "2026-09-01"
depends_on: [6g5rwjqr7rt8]
updated_at: "2026-09-01"
---

# Add dogfooded Thread graph presentation to the TUI

## Objective

Dogfood the Thread list/detail view, record what topology remains hard to understand, and add the
smallest terminal-native graph presentation that answers those observed questions. Reuse the
neutral core graph projection directly; do not import Mermaid/DOT output or build a TUI graph
engine merely because the projection makes one possible.

## Scope

- Exercise at least one real active Thread with multiple waves and an immediate external gate in
  normal, narrow, and wide terminals; record the specific unanswered questions before choosing a
  layout.
- Prefer a compact wave/edge/role explanation over a free-form node canvas unless dogfood evidence
  demonstrates that the simpler view is inadequate.
- Consume `ThreadGraphProjection` nodes, edges, waves, roles, health, and
  `topology_complete` verbatim. Presentation may group or elide secondary labels, but it must not
  derive new readiness, scheduling, or ownership semantics.
- Make truncation, scrolling, focus, edge direction, member/external roles, and incomplete topology
  explicit and deterministic at supported terminal sizes.
- Keep the graph view read-only and reachable from the Thread detail context without displacing the
  list/detail answers that proved useful during dogfooding.

## Acceptance criteria

- [ ] The task records a short list/detail dogfood note naming the topology questions that justify
  the selected presentation and alternatives deliberately rejected or deferred.
- [ ] The TUI consumes `ThreadGraphProjection` directly and does not parse Mermaid/DOT, traverse the
  task DAG, or infer eligibility, waves, external gates, or topology health locally.
- [ ] Members, immediate external gates, prerequisite-to-dependent direction, member waves, and
  incomplete/broken topology are unambiguous in the chosen view.
- [ ] Deep, wide, disconnected, hostile-label, and incomplete projections render deterministically
  without panic; normal, narrow, and minimum supported terminal behavior is covered.
- [ ] CLI/core/TUI parity fixtures prove that the same projection evidence produces compatible
  explanations even though the presentation formats differ.
- [ ] No critical-path, slack, forecasting, transitive reduction, graph editing, or Thread mutation
  feature enters this slice.

## Stress tests

- Deep chains, wide fan-out/fan-in, included external gates, disconnected members, shared tasks,
  hostile Unicode and diagram-significant labels, large graphs, partial projections, resize while
  focused, and a watcher reload that changes topology.

## Out of scope

- A general terminal graph-layout engine, persisted graph artifacts, TUI mutations, advanced graph
  analysis, web presentation, and changing the bounded member-plus-immediate-gate projection.

## Sequencing

Depends on the read-only Thread list/detail slice and begins with its dogfood gate. If list/detail
usage reveals that a compact wave explanation is sufficient, keep this task correspondingly small;
do not spend the estimate to justify a richer diagram after the fact.

## Related

- Epic [30-threads-and-task-dependency-graphs](../epics/30-threads-and-task-dependency-graphs.md)
- ADR [0006 — Adopt Threads as task DAGs](../adrs/0006-adopt-threads-as-task-dags.md)
- Thread [Complete production Threads](../threads/6g503c6pfqeb-complete-production-threads.md)
- Predecessor [Thread list and detail views](6g5rwjqr7rt8-add-thread-list-and-detail-views-to-the-tui.md)
- Supersedes part of [the original combined TUI slice](6g3q4rv89vzw-add-usage-informed-thread-views-to-the-tui.md)
