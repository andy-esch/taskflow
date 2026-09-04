---
schema: 1
id: 6g5rwjr0dz4p
status: completed
epic: 30-threads-and-task-dependency-graphs
description: After list/detail dogfooding, present Thread topology in the TUI without duplicating traversal logic or making narrow terminals unusable.
effort: 1-2 days
tier: 3
priority: medium
autonomy_level: 3
tags: [threads, tui, graph, ux, dogfood]
created: "2026-09-01"
depends_on: [6g5rwjqr7rt8]
updated_at: "2026-09-03"
started_at: "2026-09-03"
completed_at: "2026-09-03"
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

- [x] The task records a short list/detail dogfood note naming the topology questions that justify
  the selected presentation and alternatives deliberately rejected or deferred.
- [x] The TUI consumes `ThreadGraphProjection` directly and does not parse Mermaid/DOT, traverse the
  task DAG, or infer eligibility, waves, external gates, or topology health locally.
- [x] Members, immediate external gates, prerequisite-to-dependent direction, member waves, and
  incomplete/broken topology are unambiguous in the chosen view.
- [x] Deep, wide, disconnected, hostile-label, and incomplete projections render deterministically
  without panic; normal, narrow, and minimum supported terminal behavior is covered.
- [x] CLI/core/TUI parity fixtures prove that the same projection evidence produces compatible
  explanations even though the presentation formats differ.
- [x] `f` opens the established stable-identity picker over readable member and immediate-gate
  tasks; hidden lifecycle states widen the destination view and `ctrl+o` returns to the Thread.
- [x] The topology page has a visible stable-ID cursor; `j`/`k` traverses readable tasks in displayed
  gate/wave/unranked order, `Enter` opens the selection, and reload retains it when still present.
- [x] No critical-path, slack, forecasting, transitive reduction, graph editing, or Thread mutation
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

## Dogfood findings (2026-09-03)

The production `complete-production-threads` projection is healthy and complete with 19 bounded
nodes, 24 edges, eight member waves, and one satisfied immediate external gate. The shipped summary
answers lifecycle, health, active/frontier work, membership, and gate state, while `thread plan`
answers broad wave order. What remained expensive was causal reading: neither surface explains why
a task occupies its wave without opening a separate Mermaid/DOT export, so wave headings alone
would still present a sequence of task bags.

The first TUI pass combined the supplied waves with a full prerequisite-to-dependent edge ledger.
Dogfooding immediately showed that repeating two long task identities for every edge was verbose
without making the relationships easier to scan, and the read-only text offered no route into its
tasks. The revised presentation gives every node a compact deterministic member/gate alias and
lists incoming prerequisite aliases on that node; external gates remain outside member waves and
unranked members remain visible when topology is partial. `f` reuses the existing task picker over
all readable bounded nodes, including external gates, and navigation/back-stack operations retain
canonical identities. The shared picker windows large reference sets around its cursor so every
target remains visible in a short terminal and opens on the topology page's current selection. A
second dogfood pass exposed that a picker alone left the page itself feeling inert, so the page now
carries a visible stable-ID cursor: `j`/`k` moves in the same order the gate, wave, and unranked
sections are rendered, while `Enter` reuses canonical shell navigation. A free-form node canvas was
rejected for this slice; parsing Mermaid/DOT back into the TUI, adding layout dependencies, and
advanced graph calculations remain deferred. PTY dogfooding against the real Thread at wide and
narrow widths confirmed that the structured view remains readable through the ordinary detail
viewport and that stable IDs, roles, direction, and truncation survive the narrow single-pane
layout.

## Implementation notes

- Thread detail loads the body and `ThreadGraphProjection` through one paired core read so summary
  and topology cannot come from separate snapshots during an external edit.
- The generic detail pane owns an optional alternate-view seam. `v` switches Thread summary and
  topology without teaching the root model Thread-specific state, and a same-entity reload keeps
  the selected view. A sibling optional navigation seam lets the renderer own stable row identity
  and order while the shell owns `j`/`k`, history, and Enter-to-open; reload preserves the selected
  task when it remains readable.
- The topology renderer indexes and groups supplied nodes, waves, and edges for presentation only.
  It performs no graph traversal or readiness/scheduling derivation, and it neutralizes terminal
  controls in repository-authored text.
- A navigation target hidden by an initially unloaded tab's default lifecycle view now loads
  `:all`; this makes unloaded and already-loaded explicit jumps consistent instead of landing on an
  unrelated visible row.
- Stress coverage includes deep chains, wide fan-out/fan-in, a disconnected member, an external
  gate, a missing/unranked member, hostile Unicode/control text, reload, resize, and widths down to
  twelve cells. A parity fixture checks waves/nodes/edges against CLI plan and DOT renderings of the
  same core projection.
- Full validation passes: the race-enabled repository suite, `golangci-lint`, planning lint, and
  `git diff --check` are clean.
- Follow-up [two-dimensional navigable graph prototype](6g6dw5js81f3-prototype-a-two-dimensional-navigable-thread-graph-view.md)
  owns spatial layout and `hjkl` node navigation after this linear reader accumulates evidence.

## Related

- Epic [30-threads-and-task-dependency-graphs](../epics/30-threads-and-task-dependency-graphs.md)
- ADR [0006 — Adopt Threads as task DAGs](../adrs/0006-adopt-threads-as-task-dags.md)
- Thread [Complete production Threads](../threads/6g503c6pfqeb-complete-production-threads.md)
- Predecessor [Thread list and detail views](6g5rwjqr7rt8-add-thread-list-and-detail-views-to-the-tui.md)
- Supersedes part of [the original combined TUI slice](6g3q4rv89vzw-add-usage-informed-thread-views-to-the-tui.md)
