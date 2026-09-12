---
schema: 1
id: 6g914ybs9t87
status: completed
epic: 30-threads-and-task-dependency-graphs
description: Explain member and external-gate colors and solid or dashed borders directly in Mermaid and DOT Thread graphs.
effort: S
tier: 2
priority: medium
autonomy_level: 4
tags: [threads, cli, graph, ux, mermaid]
created: "2026-09-11"
depends_on: [6g3q4rv1w9e2]
updated_at: "2026-09-11"
started_at: "2026-09-11"
completed_at: "2026-09-11"
---
# Add a semantic legend to Thread graph exports

## Objective

Make rendered Thread graphs explain their own visual language. Mermaid currently distinguishes
Thread members from immediate external gates through fill color and solid versus dashed borders,
but a reader encountering the diagram in a PR or document has to infer those meanings.

## Acceptance criteria

- [x] Mermaid output includes a compact, structurally separate legend that labels the Thread-member
  and external-prerequisite treatments in words.
- [x] The legend explains both color and solid/dashed border meaning without requiring color
  perception as the only signal.
- [x] Legend samples cannot be mistaken for graph nodes, cannot create dependency edges, and do not
  alter projection node/edge counts or graph-health metadata.
- [x] DOT output receives an equivalent self-explaining treatment, or a documented renderer
  constraint and focused follow-up if exact parity is not sound.
- [x] Output remains deterministic and bounded for member-only, external-gate, empty, degraded, and
  hostile-label projections; snapshots cover both formats.
- [x] JSON and the adapter-neutral Thread projection remain unchanged because the legend is
  presentation metadata.

## Design attention

Prefer a small renderer-owned legend subgraph or cluster with textual role labels. Do not add fake
domain nodes to the projection or make downstream consumers parse styling to recover role semantics.

## Out of scope

- Redesigning the existing palette, adding interactive graph controls, or changing
  member/external-gate semantics.

## Related

- Epic [30-threads-and-task-dependency-graphs](../epics/30-threads-and-task-dependency-graphs.md)
- Thread [Refine Thread and TUI navigation](../threads/6g90h4pg0q7n-refine-thread-and-tui-navigation.md)
- Builds on [deterministic Thread graph views](6g3q4rv1w9e2-generate-deterministic-thread-graph-views.md).

## Implementation outcome (2026-09-11)

Mermaid and DOT now append one constant-size, structurally separate legend that names both
Thread-member and external-prerequisite roles and explains their color plus solid or dashed border
treatment. DOT task nodes now use the same blue/amber palette as Mermaid. Legend samples are never
connected to dependency edges; the adapter-neutral projection and JSON contract remain unchanged.

Coverage pins empty, member-only, degraded, external-gate, and hostile-label projections across
both formats, including deterministic output and unchanged dependency-edge counts. Full
`go test ./...`, focused race tests, `go vet ./...`, `golangci-lint run ./...`, `just docs-check`,
planning lint, and `git diff --check` pass.

Dogfood evidence: merged planning PR [#227](https://github.com/andy-esch/taskflow/pull/227) was
regenerated with the implementation and its body now embeds the renderer-produced legend.
