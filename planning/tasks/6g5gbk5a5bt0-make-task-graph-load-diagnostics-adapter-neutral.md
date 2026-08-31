---
schema: 1
id: 6g5gbk5a5bt0
status: completed
epic: 30-threads-and-task-dependency-graphs
description: Replace filesystem-shaped task-graph load attribution with neutral record identity before graph-view machine contracts ship.
effort: 1-2 days
tier: 2
priority: high
autonomy_level: 3
tags: [threads, architecture, ports, diagnostics]
created: "2026-08-31"
depends_on: [6g5fy1m967ka]
updated_at: "2026-08-31"
started_at: "2026-08-31"
completed_at: "2026-08-31"
---

# Make task-graph load diagnostics adapter-neutral

## Objective

Remove the remaining filesystem-shaped diagnostic contract from `TaskGraphSource` before generated
graph views publish it as a reusable machine projection. A database, remote service, cache, or
read-only adapter must be able to attribute an unreadable task record by stable identity without
synthesizing a Markdown filename that core parses for meaning.

## Current constraint

`TaskGraphSource.ListTasks` returns `[]domain.FileProblem`, and `NewTaskGraph` derives the affected
task ID and slug by parsing `problem.Path` as `<stable-id>-<slug>.md`. Non-filesystem sources still
fail closed, but lose task attribution, hard-broken identity, and useful reference candidates unless
they counterfeit a path. Ordinary task listing and lint legitimately retain file diagnostics; the
graph-read port should not require them.

## Scope

- Define one taskflow-owned task-graph load result/problem shape with optional stable task identity,
  neutral source/location metadata, and an actionable message.
- Make `TaskGraphSource` return that neutral shape, while adapting the aggregate `Store` and local
  filesystem at the composition boundary without weakening complete-store defaults.
- Preserve `NewTaskGraph` compatibility for body-aware lint and other callers that already own
  `FileProblem` values; filesystem path parsing belongs only in that adapter/conversion path.
- Keep unreadable task identity participating in hard-broken graph state, reference resolution,
  deterministic ordering, and explanatory diagnostics.
- Define the neutral graph-view mapping before Mermaid/DOT and machine contracts consume the
  problem vocabulary.

## Acceptance criteria

- [x] A `TaskGraphSource` implementation can report an unreadable record with a task ID and no
  filesystem path, and the resulting graph attributes the problem and fails closed for that ID.
- [x] Local filesystem and aggregate-store construction preserve current task, graph, lint, CLI,
  and wire behavior without duplicate scans.
- [x] File-path parsing used to recover legacy/local identity is isolated to an explicit adapter or
  conversion boundary rather than the neutral source contract.
- [x] Missing, duplicate, invalid, and unreadable task records retain deterministic health,
  resolution, and diagnostics across filesystem and non-filesystem fixtures.
- [x] Graph-view core and machine projections can map the neutral problem shape without importing
  filesystem or concrete-adapter types.

## Out of scope

- Implementing database/HTTP adapters, changing persisted task files, adding graph renderers,
  redesigning all entity `FileProblem` APIs, or changing guarded mutation policy.

## Sequencing

This follows the independent `TaskGraphSource` foundation and is a prerequisite of deterministic
Thread graph views because those views are the first new machine contract that would otherwise
cement the filesystem-shaped attribution channel.

## Related

- Epic [30-threads-and-task-dependency-graphs](../epics/30-threads-and-task-dependency-graphs.md)
- Foundation
  [decouple-thread-graph-reads-from-the-aggregate-planning-store](6g5fy1m967ka-decouple-thread-graph-reads-from-the-aggregate-planning-store.md)
- Downstream
  [generate-deterministic-thread-graph-views](6g3q4rv1w9e2-generate-deterministic-thread-graph-views.md)
- ADR [0006 — Adopt Threads as task DAGs](../adrs/0006-adopt-threads-as-task-dags.md)

## Implementation outcome

`TaskGraphSource` now returns one neutral `TaskGraphRead` containing task records and
`TaskGraphLoadProblem` values. Unreadable records carry optional stable ID/slug directly; `Path` is
only optional repair context and the analyzer never parses it for identity. The local filesystem
implements this port with the same single `ListTasks` scan, while complete stores that do not yet
implement it receive a compatibility adapter. `NewTaskGraph` preserves body-aware lint and existing
file-oriented callers through the same explicit conversion.

Tests cover pathless remote-style attribution, explicit identity winning over a misleading
location, unreadable blocker classification and reference resolution, exact local diagnostic
wording, aggregate fallback scan count, complete filesystem behavior, and all existing graph health
fixtures. No wire/schema change or guarded mutation policy change was required.
