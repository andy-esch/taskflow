---
schema: 1
status: active
description: Validate and, if accepted, implement Threads as initiative views over a planning-space task DAG, with global dependencies, lifecycle gating, bulk composition, and generated projections.
priority: medium
tags: [planning-model, threads, graph, workflow]
created: "2026-08-24"
---
# Threads and task dependency graphs

**Goal.** Validate and, if ADR-0006 is accepted, implement Threads as initiative views over a planning-space task DAG without overcommitting to speculative graph features.

## Why this is its own epic

Threads are not only a new first-class document. They change task dependency ownership, lifecycle eligibility, repository-wide graph integrity, multi-file composition, CLI and wire projections, and eventually the TUI. That cross-cutting domain deserves a coherent home rather than being split between the generic entity, storage, and CLI epics.

The first work in this epic is deliberately a decision spike. Production implementation should not be scoped until the spike gives ADR-0006 enough evidence to accept, revise, or abandon.

## Decision gate

The spike must leave one explicit recommendation:

- accept ADR-0006 and scope implementation slices;
- revise named decisions or contracts, then re-evaluate; or
- abandon the idea and record why the simpler existing model wins.

## Out of scope

- Treating the spike as production implementation.
- Critical-path, slack, forecasting, transitive reduction, or scheduler features.
- Autonomous multi-agent or worktree orchestration.
- TUI implementation before the domain, CLI, and wire projections are proven.

## Related

- [ADR-0006](../adrs/0006-adopt-threads-as-task-dags.md)
- [Vertical MVP decision spike](../tasks/6g3a1wtx4zrr-spike-a-vertical-threads-and-global-task-dag-mvp.md)
- [First-class entities epic](28-first-class-entities-new-planning-nouns.md).
