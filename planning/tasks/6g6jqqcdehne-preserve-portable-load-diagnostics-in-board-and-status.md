---
schema: 1
id: 6g6jqqcdehne
status: ready-to-start
epic: 30-threads-and-task-dependency-graphs
description: Retain task identity and optional locations when graph-backed dashboards report unreadable records through non-filesystem adapters.
effort: 1-2 days
tier: 3
priority: low
autonomy_level: 3
tags: [threads, architecture, diagnostics, ports]
created: "2026-09-03"
depends_on: [6g5vm4efjcdv, 6g697mp8s4tx]
updated_at: "2026-09-03"
---

# Preserve portable load diagnostics in board and status

## Objective

Preserve record identity when graph-backed dashboards report unreadable entities through a
database, service, cache, or other non-filesystem adapter. The newly explicit
`PlanningSummarySource` accepts neutral `TaskGraphLoadProblem` values, but `Board` and `Summary`
immediately collapse them back to `domain.FileProblem{Path, Message}`. A pathless adapter therefore
loses the task ID and slug that the graph source supplied, even though graph analysis itself retains
them correctly.

## Scope

- Reuse the repository-level neutral load-diagnostic vocabulary established by
  `make-repository-lint-load-diagnostics-adapter-neutral`; do not create a competing dashboard-only
  type.
- Carry entity kind, optional stable ID/slug, optional opaque repair location, and message through
  `Board`, `Summary`, `SpaceOverview`, CLI human/JSON output, and TUI overview/atlas attention.
- Keep local filesystem diagnostics equally actionable and preserve partial-result exit behavior.
- Preserve one task scan per board/summary load and the explicit `TaskGraphSource` boundary.
- Advance the wire schema deliberately, documenting compatibility for existing `{path,message}`
  consumers rather than silently changing the unreadable-record shape.

## Acceptance criteria

- [ ] A pathless task-graph load problem retains its task ID/slug through board and current/cross-space
      status core results and JSON.
- [ ] Explicit record identity wins over a misleading location; no core code parses an opaque
      location to recover identity.
- [ ] Local board/status human output still names actionable file locations and retains current
      non-zero behavior for unreadable records.
- [ ] Mixed task, epic, and audit load failures have deterministic kind/identity/location attribution
      in a summary without extra scans.
- [ ] TUI overview and atlas continue to flag unreadable data without depending on filesystem paths.
- [ ] Schema comments, generated JSON Schema, compatibility notes, and machine-contract fixtures are
      updated together.

## Stress tests

Pathless remote-style failures; explicit identity plus a contradictory location; invalid local
filenames; mixed readable/unreadable entity kinds; stable ordering; current status, cross-space
status, board, and TUI reloads; scan-count assertions.

## Out of scope

- Changing graph health, mutation, or lint severity policy.
- Redesigning guarded local-file snapshot comparison or `lint --fix`.
- Implementing a database or HTTP adapter.
- Redesigning unrelated entity-list envelopes unless the shared diagnostic contract requires a
  compatible mechanical mapping.

## Sequencing

This task follows both the status graph-health work and
`make-repository-lint-load-diagnostics-adapter-neutral`. The latter should establish the shared
multi-entity diagnostic value and wire compatibility policy first; this task then carries it through
dashboard projections without duplicating that design.

## Related

- Epic [30-threads-and-task-dependency-graphs](../epics/30-threads-and-task-dependency-graphs.md)
- [Report graph degradation in status and lint](6g697mp8s4tx-report-graph-degradation-in-status-and-lint.md)
- [Make repository lint load diagnostics adapter-neutral](6g5vm4efjcdv-make-repository-lint-load-diagnostics-adapter-neutral.md)
