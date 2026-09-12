---
schema: 1
id: 6g9150nrt4p9
status: next-up
epic: 30-threads-and-task-dependency-graphs
description: Render truthful one- or two-edge Thread neighborhoods for review-sized CLI, PR, TUI, and future web context.
effort: 1-2 days
tier: 1
priority: high
autonomy_level: 2
tags: [threads, cli, graph, ux, ports, dogfood]
created: "2026-09-11"
depends_on: [6g8vxbv3d4xn]
updated_at: "2026-09-11"
---
# Export bounded Thread neighborhoods around a task

## Objective

Let adapters and planning-aware workflows render a truthful one- or two-edge neighborhood around
any node in a Thread projection. This gives PRs, terminals, and future interfaces a review-sized
causal picture without pretending the excerpt is the whole graph or rescanning repository
dependencies outside the supplied projection.

## Acceptance criteria

- [ ] Thread graph export accepts a focal member or external-gate task and an explicit depth of one
  or two edges while preserving full-graph export as the default.
- [ ] Neighborhood membership is defined deterministically from supplied incoming and outgoing
  edges: include the focal node, every node within the requested hop distance, and every supplied
  edge whose endpoints are both included.
- [ ] Mermaid, DOT, and JSON expose the same bounded node/edge set, stable identities,
  member/external-gate roles, graph and projection health, focal identity, requested depth,
  shown-versus-hidden scope, and deterministic boundary-continuation evidence.
- [ ] The output states that it is a bounded neighborhood and visibly discloses omitted graph scope
  or boundary continuations so it cannot be mistaken for the complete Thread.
- [ ] Missing, ambiguous, unreadable, or out-of-Thread focus references fail explicitly; degraded
  and partial projections retain their supplied diagnostics rather than being relabeled healthy.
- [ ] A pure adapter-neutral selector derives the bounded node/edge set and explicit scope metadata
  from `ThreadGraphProjection`; CLI, TUI, web, and PR tooling can reuse it without importing
  terminal geometry or GitHub policy into core graph facts.
- [ ] Fan-in, fan-out, chain, isolated member, external gate, shared neighbor, depth-one/depth-two,
  hostile label, and deterministic ordering regressions are covered.

## Design attention

Treat hop distance as undirected proximity over directed dependency evidence while preserving arrow
direction in the rendered result. The selector must distinguish a truthful bounded excerpt from a
complete graph and identify supplied edges that cross its boundary. Keep responsive coordinates,
aliases, keys, viewport history, and PR-description composition out of this shared seam.

## Out of scope

- Repository-wide blocker traversal beyond the Thread projection, arbitrary depth, critical-path
  calculations, automatic focal-node selection, or hard-coding PR description behavior into
  taskflow.

## Related

- Epic [30-threads-and-task-dependency-graphs](../epics/30-threads-and-task-dependency-graphs.md)
- Thread [Refine Thread and TUI navigation](../threads/6g90h4pg0q7n-refine-thread-and-tui-navigation.md)
- Follow [large-graph navigation research](6g8vxbv3d4xn-explore-progressive-disclosure-and-navigation-for-large-thread-graphs.md)
- Implements the shared seam selected in [Large Thread graph progressive disclosure and navigation](../research/6g95tfg02hm3-large-thread-graph-progressive-disclosure-and-navigation.md).
- Coordinate with the [one-hop TUI focus view](6g86c7y6hn41-add-a-one-hop-focus-subgraph-mode-to-the-tui-thread-graph.md).
