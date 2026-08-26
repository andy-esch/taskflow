---
schema: 1
id: 6g3q4rst78qy
status: in-progress
epic: 30-threads-and-task-dependency-graphs
description: Model canonical depends_on data and deterministic fail-closed graph snapshots, lint, migration, and bounded library comparison.
effort: 3-5 days
tier: 1
priority: high
autonomy_level: 3
tags: [threads, graph, storage, migration]
created: "2026-08-25"
updated_at: "2026-08-25"
started_at: "2026-08-25"
---
# Establish canonical task dependencies and strict graph reads

## Objective

Introduce the production read foundation for one planning-repository task DAG without exposing graph mutations yet.

## Scope

- Model `depends_on` as a sorted, duplicate-free set of stable task IDs on `domain.Task` and in the task field/schema contracts.
- Define a narrow taskflow-owned graph analysis interface and strict repository snapshot distinct from resilient repair-oriented listing.
- Migrate or actionably diagnose the live legacy `blocked_by`, `dependencies`, and `blocks` vocabulary without disturbing task prose.
- Implement deterministic validation, cycle paths, blocker/downstream traversal, and topological waves.
- Run the bounded owned-versus-`dominikbraun/graph` contract-test bake-off and record the decision; do not expand V1 algorithms.

## Acceptance criteria

- [ ] Valid task frontmatter round-trips `depends_on` in stable ID order through domain, store, schema, and wire-facing task representations where applicable.
- [ ] A strict snapshot fails closed on malformed/unreadable tasks, ID drift, unknown status, duplicate/self/missing dependencies, and attributable cycles.
- [ ] Existing resilient list/lint repair behavior remains available and reports every graph problem deterministically.
- [ ] Legacy dependency vocabulary has an explicit migration or actionable diagnostic path, with fixtures proving body/frontmatter preservation.
- [ ] One shared contract suite covers the owned analyzer and the bounded library adapter; the retained choice and rationale are recorded.
- [ ] No public dependency or Thread mutation is introduced by this task.

## Stress tests

- Randomized input/map order produces byte-for-byte stable diagnostics and plans.
- Deep chains, wide frontiers, disconnected tasks, duplicate edges, missing IDs, self-edges, and exact cycle paths are covered.
- Ordinary repository lint and the full test, race, formatting, schema, and diff checks pass.

## Sequencing

First production slice. It unlocks guarded dependency writes and supplies the analysis contract consumed by eligibility and Threads.
