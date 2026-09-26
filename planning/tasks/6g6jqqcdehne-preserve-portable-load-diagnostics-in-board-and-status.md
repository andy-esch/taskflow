---
schema: 1
id: 6g6jqqcdehne
status: completed
epic: 30-threads-and-task-dependency-graphs
description: Retain task identity and optional locations when graph-backed dashboards report unreadable records through non-filesystem adapters.
effort: 1-2 days
tier: 3
priority: medium
autonomy_level: 3
tags: [threads, architecture, diagnostics, ports]
created: "2026-09-03"
depends_on: [6g5vm4efjcdv, 6g697mp8s4tx]
updated_at: "2026-09-26"
started_at: "2026-09-25"
completed_at: "2026-09-26"
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

- [x] A pathless task-graph load problem retains its task ID/slug through board and current/cross-space
      status core results and JSON.
- [x] Explicit record identity wins over a misleading location; no core code parses an opaque
      location to recover identity.
- [x] Local board/status human output still names actionable file locations and retains current
      non-zero behavior for unreadable records.
- [x] Mixed task, epic, and audit load failures have deterministic kind/identity/location attribution
      in a summary without extra scans.
- [x] TUI overview and atlas continue to flag unreadable data without depending on filesystem paths.
- [x] Schema comments, generated JSON Schema, compatibility notes, and machine-contract fixtures are
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
dashboard projections without duplicating that design. It is one of three parallel concrete seams
after that foundation in the adapter-neutral data-access Thread; all three feed the shared entity-read
design rather than allowing the design to proceed from lint alone.

## Related

- Epic [30-threads-and-task-dependency-graphs](../epics/30-threads-and-task-dependency-graphs.md)
- [Report graph degradation in status and lint](6g697mp8s4tx-report-graph-degradation-in-status-and-lint.md)
- [Make repository lint load diagnostics adapter-neutral](6g5vm4efjcdv-make-repository-lint-load-diagnostics-adapter-neutral.md)
- Thread [Make planning data access adapter neutral](../threads/6gcwd78p9r04-make-planning-data-access-adapter-neutral.md)

## Implementation outcome (2026-09-25)

`Board`, `Summary`, retained cross-space summaries, wire envelopes, CLI partial-result errors, and
TUI health indicators now share `LintLoadProblem`. `TaskGraphLoadProblem` preserves an explicit
opaque location separately from its local repair path, and whole-snapshot comparison includes that
evidence. Local adapters recover task/audit stable identity during their existing scans and recover
legacy epic identity at the filesystem boundary; core performs no filename parsing. Board and
summary tests prove pathless identity, canonical within-kind ordering, and one task/epic/audit scan.

Schema 1.75 extends board/current-status/cross-space-status unreadable records with kind, optional
ID/slug, and location while retaining required `path` and `message`; opaque locations leave `path`
empty unless the adapter also supplied a distinct local repair path. Human output retains non-zero
partial-result behavior: current and cross-space status identify affected records, and cross-space
status renders their location, repair path, and message beneath the owning space. The remaining
aggregate `SummaryStore` `FileProblem` seam is explicitly transitional and remains owned by the
sequenced portable entity-read design task.

The independent [Codex](../audits/6gdpcag50p97-2026-09-25-portable-board-status-diagnostics-implementation-codex.md)
and [Antigravity](../audits/6gdpcagd8fk6-2026-09-25-portable-board-status-diagnostics-implementation-antigravity.md)
reviews drove the closeout hardening. Codex's ordering, dual-location, cross-space detail, and
no-inference findings were fixed in this slice. Antigravity's surviving mutation identified an
inert location copy in lint's private graph input; the implementation now states and enforces that
this copy carries unreadable identity only, while the original portable diagnostic remains the
single user-facing source.

Validation: `just test` (race-enabled), `golangci-lint`, `go mod tidy -diff`, generated schema-comment
drift, machine-golden regeneration, planning lint, and `git diff --check` all pass.
