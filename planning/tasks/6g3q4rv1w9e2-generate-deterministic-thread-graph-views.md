---
schema: 1
id: 6g3q4rv1w9e2
status: next-up
epic: 30-threads-and-task-dependency-graphs
description: Render explanatory Mermaid and DOT Thread graphs from shared runtime projections without persisting derived output.
effort: 2-3 days
tier: 3
priority: medium
autonomy_level: 4
tags: [threads, graph, cli, rendering]
created: "2026-08-25"
depends_on: [6g3q4rtv8d0a, 6g5f1d23jy1b, 6g5fy1m967ka, 6g5gbk5a5bt0]
updated_at: "2026-08-31"
started_at: "2026-08-31"
---
# Generate deterministic Thread graph views

## Objective

Render explanatory Thread graphs from the shared runtime projection without storing derived diagrams or introducing scheduling semantics.

## Scope

- Add one taskflow-owned, adapter-neutral graph projection in core for CLI, TUI, and future web
  consumers; it must not contain renderer, framework, filesystem, or third-party graph types.
- Add deterministic Mermaid and DOT output with explicit member versus external-gate roles.
- Put Mermaid and DOT formatters in a pure `internal/graphfmt` package with no CLI/TUI/HTTP or
  styling dependencies, so another primary adapter can export the same projection without invoking
  Cobra, importing Lip Gloss, or parsing CLI output.
- Polish topological waves as explanatory planning output.
- Keep ASCII/Unicode optional and exclude critical path, slack, forecasting, and transitive reduction.

## Acceptance criteria

- [ ] Mermaid and DOT encode the same nodes, edges, roles, and ordering as the shared projection.
- [ ] CLI output and reusable formatters consume an adapter-neutral core projection that can also
  be used directly by TUI and future web adapters.
- [ ] Output is generated at runtime and never persisted into Thread documents.
- [ ] Titles and metadata are escaped safely and output remains deterministic across repository scan order.
- [ ] The core projection carries raw labels; each `internal/graphfmt` formatter owns its
  format-specific escaping and is directly testable without a primary adapter.
- [ ] Core and machine contracts remain renderer-, framework-, and graph-library-neutral.

## Stress tests

- Golden fixtures for external gates, shared tasks, disconnected members, hostile labels, empty/minimal graphs, and large deep/wide graphs.

## Sequencing

Requires stable Thread projections and adapter-neutral graph-load diagnostics. It is not a
prerequisite for bulk linking or eligibility enforcement.
