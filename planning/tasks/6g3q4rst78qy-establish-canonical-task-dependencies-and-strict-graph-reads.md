---
schema: 1
id: 6g3q4rst78qy
status: in-progress
epic: 30-threads-and-task-dependency-graphs
description: Model canonical depends_on data and deterministic strict graph snapshots, derived gate state, legacy diagnostics, and bounded library comparison.
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
- Diagnose—but do not rewrite—the live legacy `blocked_by`, `dependencies`, and `blocks` vocabulary, including slug-to-ID resolution failures and ambiguity.
- Implement deterministic validation, cycle paths, memoized sound completion/gate derivation, lifecycle roles, blocker/downstream traversal, and topological waves.
- Define blocker records with stable reason tokens and one deterministic shortest explanatory path.
- Prevent generic `task set`, including `--force`, and unguarded `task edit` from introducing or changing `depends_on`.
- Run the bounded owned-versus-`dominikbraun/graph` contract-test bake-off and record the decision; do not expand V1 algorithms.

## Acceptance criteria

- [ ] Valid task frontmatter round-trips `depends_on` in stable ID order through domain, store, schema, and wire-facing task representations where applicable.
- [ ] A strict snapshot identifies malformed/unreadable tasks, ID drift, unknown status, duplicate/self/missing dependencies, and attributable cycles for fail-closed consumers while diagnostic consumers retain the problem list.
- [ ] Existing resilient list/lint repair behavior remains available and reports every graph problem deterministically.
- [ ] Legacy dependency diagnostics report each resolvable target ID and actionable missing/ambiguous slug failures; no migration write occurs in this task.
- [ ] Gate derivation implements broken-over-blocked precedence, treats deferred prerequisites as blocked, and explains unfinished, parked, withdrawn, missing, and unsound-completed blockers.
- [ ] Sound completion and derived graph state memoize one result per task per snapshot and are O(V+E).
- [ ] Generic set always rejects `depends_on`, including under `--force`; interactive edit cannot land a dependency delta before guarded validation exists.
- [ ] One shared contract suite covers the owned analyzer and the bounded library adapter; the retained choice and rationale are recorded.
- [ ] No public dependency or Thread mutation is introduced by this task.

## Stress tests

- Randomized input/map order produces byte-for-byte stable diagnostics and plans.
- Deep chains, wide frontiers, disconnected tasks, duplicate edges, missing IDs, self-edges, and exact cycle paths are covered.
- Reconvergent diamond chains assert bounded visit counts, not merely acceptable wall-clock time.
- Ordinary repository lint and the full test, race, formatting, schema, and diff checks pass.

## Sequencing

First production slice. It unlocks guarded dependency writes and supplies the pure role/gate/sound-completion analysis consumed independently by eligibility enforcement and Threads.
