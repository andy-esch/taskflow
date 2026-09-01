---
schema: 1
id: 6g3q4rv89vzw
status: deprecated
epic: 30-threads-and-task-dependency-graphs
description: Expose production-proven Thread lifecycle, rollup, frontier, gates, and graph information through read-oriented TUI views.
effort: 3-5 days
tier: 3
priority: medium
autonomy_level: 3
tags: [threads, tui, graph, ux]
created: "2026-08-25"
updated_at: "2026-09-01"
depends_on: [6g3q4rv1w9e2, 6g5m69wpydzw]
deprecated_at: "2026-09-01"
---
# Add usage-informed Thread views to the TUI

## Objective

Expose production-proven Thread lifecycle, rollup, frontier, member/external roles, and readable graph information in the TUI without creating a second graph engine.

## Superseded by scoped slices

Deprecated on 2026-09-01 after the v0.18.0 CLI preview and clean-space dogfood made the risk centers
clearer. This combined task is retained as planning history; its work moved into three independently
reviewable, dependency-ordered tasks:

1. [Contention-safe Thread projection loading](6g5rwjqeh6a6-wire-thread-projections-into-the-tui-with-contention-safe-reloads.md)
   owns adapter parity, split-capability composition, and coherent watcher reloads.
2. [Thread list and detail views](6g5rwjqr7rt8-add-thread-list-and-detail-views-to-the-tui.md)
   owns the first useful read-only interface and responsive navigation.
3. [Dogfooded Thread graph presentation](6g5rwjr0dz4p-add-dogfooded-thread-graph-presentation-to-the-tui.md)
   owns only topology questions that remain after list/detail usage.

This split keeps graph-layout judgment behind a real usage gate and prevents transient repository
contention from being buried inside a broad visual implementation task.

The split also exposed four prerequisite hardening tasks that were not honest visual sub-slices:
[stable TUI identity](6g5rxq17px59-make-tui-entity-navigation-use-stable-identities.md),
[recoverable directory watches](6g5rxq1g5mp1-keep-tui-live-reload-healthy-when-entity-directories-appear.md),
[optional local Thread paths](6g5ryqqx5ab7-split-local-thread-path-resolution-from-portable-thread-reads.md),
and [neutral Thread diagnostics](6g5rxq1ravd3-make-thread-read-diagnostics-adapter-neutral.md).
They precede the relevant replacement slices through repository-global dependencies.

## Original scope

- Add Thread list/navigation and a detail view using existing core/wire projections.
- Begin read-oriented; direct graph editing is not required for the first TUI slice.
- Incorporate CLI dogfood feedback before freezing layout and interaction choices.

## Acceptance criteria

- [ ] TUI state matches CLI/core projections for lifecycle, progress, frontier, gates, and inconsistency. · **tracked:** split into 6g5rwjqeh6a6 and 6g5rwjqr7rt8
- [ ] Watcher reload handles Thread and task dependency changes coherently. · **tracked:** split into 6g5rxq1g5mp1 and 6g5rwjqeh6a6
- [ ] Narrow terminals degrade readably and navigation/completion remain consistent with other first-class entities. · **tracked:** split into 6g5rwjqr7rt8 and 6g5rwjr0dz4p
- [ ] No TUI-local readiness, graph traversal, or direct filesystem mutation logic is introduced. · **tracked:** enforced across 6g5rwjqeh6a6, 6g5rwjqr7rt8, and 6g5rwjr0dz4p
- [ ] Planner-phase ErrConflict from canonical-root callback exclusion is
  treated as transient watcher contention and retried or debounced without
  rendering an empty or permanently broken view. · **tracked:** owned by 6g5rwjqeh6a6

## Stress tests

- Watcher reload during mutations, completed-thread drift, missing/broken members, large Thread lists, graph overflow, small terminals, and CLI/TUI parity fixtures.

## Original sequencing

Last planned V1 slice, after real CLI usage of Thread projections and generated views.

## Mutation-contention amendment (2026-08-27)

The graph-mutation planner phase excludes every Store call at canonical-root scope. A watcher
refresh that lands in that brief window can therefore receive ErrConflict even though no repository
defect exists. Treat that result as transient contention: retain the last coherent model, debounce
or retry after the mutation events settle, and never replace the view with empty data or a permanent
broken-state banner.

A multi-file mutation may also emit several watcher events for graph-valid durable prefixes.
Coalesce them and reload the shared core projection after the burst rather than deriving intermediate
graph state in the TUI.
