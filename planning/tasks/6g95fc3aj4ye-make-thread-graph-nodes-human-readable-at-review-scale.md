---
schema: 1
id: 6g95fc3aj4ye
status: completed
epic: 30-threads-and-task-dependency-graphs
description: Make exported Thread nodes scannable without losing stable identity, state, role, or deterministic output.
effort: M
tier: 2
priority: high
autonomy_level: 3
tags: [threads, cli, graph, ux, mermaid, dot]
created: "2026-09-11"
depends_on: [6g914ybs9t87]
updated_at: "2026-09-12"
started_at: "2026-09-11"
completed_at: "2026-09-12"
---
# Make Thread graph nodes human-readable at review scale

## Objective

Make Mermaid and DOT Thread graph nodes scannable by a human reviewing a diagram without sacrificing
stable identity, state, role, or deterministic output. The current slug-first, ID-heavy,
full-description labels turn even modest graphs into walls of text and make the dependency structure
harder to read.

## Acceptance criteria

- [x] Establish and document a compact information hierarchy for a node: recognizable human label,
  lifecycle state, Thread role, stable identity, and optional description context.
- [x] Default Mermaid and DOT nodes foreground a concise human-readable label while retaining enough
  stable identity to disambiguate duplicate or similar task names.
- [x] Long slugs and descriptions are bounded deterministically; omitted text is visibly indicated and
  remains recoverable through an explicit verbose mode or adjacent command.
- [x] Member and external-prerequisite roles remain understandable alongside the semantic legend, and
  completed external context is not presented as currently blocking work.
- [x] Empty, duplicate-label, long-label, Unicode, hostile-label, member-only, and external-prerequisite
  fixtures cover both formats without changing graph topology or health metadata.
- [x] Any projection-contract change needed to expose a better display title is justified for CLI, TUI,
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

## Implementation progress (2026-09-11)

The default Mermaid and DOT hierarchy is now authored task title, lifecycle state plus human role, then stable ID. Labels and descriptions are whitespace-normalized and truncated on rune and useful word boundaries; descriptions are omitted from compact output and restored in bounded form by `thread graph --details`. The visible legend advertises that opt-in, and completed non-member context says `external prerequisite` rather than implying an active gate.

The filesystem adapter derives an optional task title from the first non-fenced Markdown H1 during its existing read. `domain.Task` and `ThreadGraphNode` carry that adapter-neutral value, while `Label` retains the stable slug-era fallback for bodyless or remote adapters. Wire output adds optional `title` without replacing `label` or `description`, so the additive machine-contract change advances `schema_version` from 1.63 to 1.64. Mermaid/DOT render options remain presentation-only.

Coverage includes exact authored titles, bodyless slug fallback, duplicate titles disambiguated by stable ID, empty/member/external/degraded projections, Unicode and long-value boundaries, hostile controls, compact-versus-details behavior, CLI flag conflicts, additive JSON, historical Thread compatibility, and generated schema/docs. Full tests, the full race suite, vet, golangci-lint, tidy-check with a sandbox-local Go cache, planning lint, audit lint, and `git diff --check` pass. Generated CLI docs are current; `just docs-check` will become clean once that intentional generated diff is staged or committed.

## Adversarial review closeout (2026-09-12)

[Antigravity's implementation audit](../audits/6g96h83t3w26-2026-09-11-readable-thread-graph-nodes-implementation-antigravity.md) found four bounded defects, all fixed with regression coverage: Unicode word-boundary truncation now compares rune offsets, detailed renderers describe their active mode instead of recommending `--details`, authored title casing is preserved, and invalid formats are rejected before repository reads. The related Markdown probe also exposed optional ATX closing hashes; body-derived titles now strip a valid closing sequence without stripping literal unseparated hashes. Claude's unused review scaffold was removed rather than retained as misleading no-findings evidence.

Final validation passes the full race suite, focused domain/formatter/CLI tests, vet, golangci-lint (zero issues), module-tidiness check, planning and audit lint, regenerated CLI docs, and `git diff --check`.
