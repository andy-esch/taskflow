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
---
# Generate deterministic Thread graph views

## Objective

Render explanatory Thread graphs from the shared runtime projection without storing derived diagrams or introducing scheduling semantics.

## Scope

- Add deterministic Mermaid and DOT output with explicit member versus external-gate roles.
- Polish topological waves as explanatory planning output.
- Keep ASCII/Unicode optional and exclude critical path, slack, forecasting, and transitive reduction.

## Acceptance criteria

- [ ] Mermaid and DOT encode the same nodes, edges, roles, and ordering as the shared projection.
- [ ] Output is generated at runtime and never persisted into Thread documents.
- [ ] Titles and metadata are escaped safely and output remains deterministic across repository scan order.
- [ ] Machine contracts remain graph-library-neutral.

## Stress tests

- Golden fixtures for external gates, shared tasks, disconnected members, hostile labels, empty/minimal graphs, and large deep/wide graphs.

## Sequencing

Requires stable Thread projections. It is not a prerequisite for bulk linking or eligibility enforcement.
