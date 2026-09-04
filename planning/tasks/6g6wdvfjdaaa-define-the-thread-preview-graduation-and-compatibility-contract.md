---
schema: 1
id: 6g6wdvfjdaaa
status: next-up
epic: 30-threads-and-task-dependency-graphs
description: Define the evidence, compatibility promises, and remaining gates required before Threads graduate from preview.
effort: 1-2 days
tier: 2
priority: medium
autonomy_level: 3
tags: [threads, release, compatibility, architecture]
created: "2026-09-04"
depends_on: [6g6scc9jgxae]
updated_at: "2026-09-04"
---

# Define the Thread preview graduation and compatibility contract

## Objective

Turn “Threads preview” from an open-ended label into an explicit product and compatibility decision.
Use evidence from the v0.18 CLI and v0.19 TUI checkpoints to define what must be stable, what may
still evolve, and what observable gates must pass before the preview notice can be removed.

## Scope

- Inventory the persisted Thread/task graph model, lifecycle and mutation commands, human CLI,
  machine envelopes/schema, core ports and projections, and TUI behavior exposed during preview.
- Classify each surface as stable, compatibility-managed preview, or internal implementation detail;
  document deprecation and migration expectations for persisted data, CLI changes, and wire changes.
- Define measurable graduation gates for graph integrity and recovery, lifecycle safety, adapter
  neutrality, dogfood coverage, documentation, upgrades, and release validation.
- Decide explicitly whether guarded graph repair, portable multi-entity diagnostics, and the spatial
  graph experiment are graduation requirements, follow-ups, or intentionally non-blocking.
- Amend ADR-0006 and the release/documentation guidance with a checklist that can produce a clear
  graduate-or-remain-preview decision.

## Acceptance criteria

- [ ] A contract matrix names every public or persisted Thread surface, its stability classification,
      and the compatibility or migration promise attached to it.
- [ ] Graduation gates are observable and linked to owning task IDs or concrete verification; no
      gate is merely “seems stable” or a version/date threshold.
- [ ] Guarded repair, portable diagnostics, the spatial prototype, and known preview limitations are
      each classified as required, explicitly non-blocking, or tracked by a named follow-up.
- [ ] Persisted-document, CLI, and JSON/wire evolution have deliberate deprecation and migration
      rules that work for local files and future service/database adapters.
- [ ] ADR-0006, the README preview notice, architecture guidance, and release checklist agree on what
      graduation means and which interfaces remain experimental afterward.
- [ ] The result provides an executable decision path for either removing the preview label or
      retaining it with specific unmet gates and owners.

## Stress tests

- Existing v0.18 data opened by v0.19+, shared tasks, external gates, malformed/degraded graphs,
  pathless adapters, CLI automation pinned to JSON fields, TUI behavior, and a future web adapter.

## Out of scope

- Implementing every graduation gate inside this task.
- Declaring the entire project 1.0-stable or promising that preview interfaces can never change.
- Making the two-dimensional graph view mandatory merely because it is visually desirable.

## Sequencing

Start after the v0.19.0 checkpoint so the contract is based on two shipped dogfood slices rather
than aspiration. It may then refine the order and release significance of the post-v0.19 tasks, but
must not retroactively describe an unimplemented capability as stable.

## Related

- Epic [30-threads-and-task-dependency-graphs](../epics/30-threads-and-task-dependency-graphs.md)
- ADR [0006 — Adopt Threads as task DAGs](../adrs/0006-adopt-threads-as-task-dags.md)
- Thread [Complete production Threads](../threads/6g503c6pfqeb-complete-production-threads.md)
- Release checkpoint [v0.19.0 TUI preview](6g6scc9jgxae-cut-v0.19.0-as-a-tui-threads-preview.md)
