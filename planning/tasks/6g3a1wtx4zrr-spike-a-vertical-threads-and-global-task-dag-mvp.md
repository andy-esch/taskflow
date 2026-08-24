---
schema: 1
id: 6g3a1wtx4zrr
status: ready-to-start
epic: 30-threads-and-task-dependency-graphs
description: Build a bounded, fixture-backed prototype of the dependency, Thread projection, and resumable bulk-composition contracts; recommend accepting, revising, or abandoning ADR-0006.
effort: 2-4 days
tier: 2
priority: high
autonomy_level: 3
tags: [spike, threads, graph, planning-model, adr]
created: "2026-08-24"
---
# Spike a vertical Threads and global task-DAG MVP

> **Decision spike, not production rollout.** This task exists to challenge ADR-0006 with executable evidence before the ADR is accepted or implementation work is split out.

## Objective

Build the smallest credible vertical prototype that exercises the risky contracts in [ADR-0006](../adrs/0006-adopt-threads-as-task-dags.md) against real taskflow seams. Use it to identify incorrect assumptions, missing decisions, disproportionate implementation cost, and package opportunities. End with a clear recommendation to accept, revise, or abandon the proposal.

The prototype may be explicitly experimental, but it must cross a real filesystem-backed planning repository in a temporary test fixture. A pure in-memory graph demo is not enough to test persistence, optimistic concurrency, repository locking, stable IDs, or resumable multi-file application. Never mutate the live planning tree as prototype data.

## Hypotheses to test

1. One task-owned `depends_on` relation can remain coherent when tasks belong to multiple Threads and when a prerequisite is outside a Thread.
2. Lifecycle role and dependency health can be projected separately into useful `eligible`, `blocked`, `broken`, `drained`, and `inconsistent` answers without adding persisted statuses.
3. Recursive sound completion produces understandable behavior when an upstream task is reopened.
4. Cycle validation can be kept correct across concurrent graph mutations using a repository-wide mutation guard rather than per-file CAS alone.
5. A two-phase compose/apply plan can mix new and existing tasks, survive interruption, and converge on retry without duplicate IDs, memberships, or edges.
6. The required V1 algorithms and render projections fit behind a taskflow-owned interface, whether implemented locally or by an off-the-shelf Go package.

## Prototype boundary

- Model only the fields required for canonical task dependencies and Thread membership/metadata.
- Exercise graph load, exact-ID resolution, cycle and broken-reference diagnostics, deterministic topology, upstream/downstream queries, and Thread frontier.
- Exercise a materialized bulk plan with preallocated IDs and additive dependency/membership changes through the real filesystem mutation primitives in a temporary repository.
- Demonstrate external-gate properties and member-versus-external roles in a typed or JSON-shaped projection; do not build polished public rendering.
- Evaluate realistic off-the-shelf graph packages against the exact V1 operation list and a small owned implementation. Do not select a package because it offers unrelated advanced algorithms.
- Keep experimental code visibly bounded and document which pieces are disposable, promotable, or deliberately faked.

## Required scenarios

The runnable test or demo fixture must include:

- two Threads sharing at least one task;
- one incomplete prerequisite outside a Thread, excluded from rollup but visible as an external gate;
- queued-and-blocked, candidate-and-clear, force-started inconsistent, and soundly completed tasks;
- reopening an upstream task and observing the completed descendant and Thread become inconsistent;
- rejection of a self-edge, missing task reference, duplicate edge, and attributable cycle;
- two would-be concurrent edge additions whose union is cyclic, proving final validation occurs inside the mutation guard;
- a bulk plan containing at least one existing task and two new tasks; and
- an injected interruption after a partial bulk apply, followed by a retry that creates no duplicates and produces a complete receipt.

## Acceptance criteria

- [ ] A runnable, deterministic prototype or focused test harness demonstrates every required scenario against a temporary filesystem-backed planning repository.
- [ ] The graph implementation/package comparison records API fit, determinism, diagnostics, dependency cost, render support, maintenance signals, and what taskflow must still own regardless of package.
- [ ] The spike maps the actual production fan-out across domain, field registry/schema, store and mutation guard, core ports/use cases, CLI transitions, wire contracts, initialization/layout, lint, and later TUI work.
- [ ] The spike records every ADR assumption as validated, falsified, or still open, with evidence and any proposed replacement contract.
- [ ] ADR-0006 receives a concise spike-findings commentary section or proposed amendments, but its status is not changed by this task.
- [ ] The final task report recommends exactly one outcome: accept and scope implementation, revise and re-spike named risks, or abandon and preserve the simpler current model.
- [ ] If acceptance is recommended, the report proposes implementation slices with dependency order and identifies which prototype code should be promoted versus removed.
- [ ] Repository tests, formatting, lint, and diff checks pass for retained code; disposable prototype artifacts are removed or clearly isolated.

## Out of scope

- Shipping the production Thread CLI, wire contract, migration, or TUI.
- Changing ADR-0006 from proposed to accepted.
- Critical path, slack, forecasting, transitive reduction, scheduling, or autonomous agent orchestration.
- Polishing graph visualization beyond enough output to validate the projection contract.
- Treating benchmark guesses as evidence; measure only if the fixture exposes a credible concern.

## Related

- Epic [30-threads-and-task-dependency-graphs](../epics/30-threads-and-task-dependency-graphs.md)
- [ADR-0006](../adrs/0006-adopt-threads-as-task-dags.md)
- [First-class entities epic](../epics/28-first-class-entities-new-planning-nouns.md)
- [Task readiness state](6fbj87001m03-task-readiness-state-draft-vs-finalized-in-frontmatter.md).
