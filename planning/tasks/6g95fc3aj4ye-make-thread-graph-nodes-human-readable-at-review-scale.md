---
schema: 1
id: 6g95fc3aj4ye
status: next-up
epic: 30-threads-and-task-dependency-graphs
description: Make exported Thread nodes scannable without losing stable identity, state, role, or deterministic output.
effort: M
tier: 2
priority: high
autonomy_level: 3
tags: [threads, cli, graph, ux, mermaid, dot]
created: "2026-09-11"
depends_on: [6g914ybs9t87]
updated_at: "2026-09-11"
---
# Make Thread graph nodes human-readable at review scale

## Objective

Make Mermaid and DOT Thread graph nodes scannable by a human reviewing a diagram without sacrificing
stable identity, state, role, or deterministic output. The current slug-first, ID-heavy,
full-description labels turn even modest graphs into walls of text and make the dependency structure
harder to read.

## Acceptance criteria

- [ ] Establish and document a compact information hierarchy for a node: recognizable human label,
  lifecycle state, Thread role, stable identity, and optional description context.
- [ ] Default Mermaid and DOT nodes foreground a concise human-readable label while retaining enough
  stable identity to disambiguate duplicate or similar task names.
- [ ] Long slugs and descriptions are bounded deterministically; omitted text is visibly indicated and
  remains recoverable through an explicit verbose mode or adjacent command.
- [ ] Member and external-prerequisite roles remain understandable alongside the semantic legend, and
  completed external context is not presented as currently blocking work.
- [ ] Empty, duplicate-label, long-label, Unicode, hostile-label, member-only, and external-prerequisite
  fixtures cover both formats without changing graph topology or health metadata.
- [ ] Any projection-contract change needed to expose a better display title is justified for CLI, TUI,
  and future web consumers; renderer-only presentation choices remain outside the core projection.

## Design attention

Start with information hierarchy rather than arbitrary truncation. Compare the task heading, slug, and
stable ID as display and disambiguation sources, and decide whether compact output should be the default
with a verbose opt-in. Preserve copy/paste usefulness and accessibility while reducing visual density.

## Out of scope

- Spatial layout, neighborhood selection, interactive TUI zoom, or changing dependency and Thread
  membership semantics.

## Related

- Thread [Refine Thread and TUI navigation](../threads/6g90h4pg0q7n-refine-thread-and-tui-navigation.md)
- Follows [semantic Thread graph legends](6g914ybs9t87-add-a-semantic-legend-to-thread-graph-exports.md).
- Complements [bounded Thread neighborhood exports](6g9150nrt4p9-export-bounded-thread-neighborhoods-around-a-task.md).
