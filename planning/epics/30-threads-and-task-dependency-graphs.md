---
schema: 1
status: active
description: Implement accepted Threads as initiative views over a planning-space task DAG, with global dependencies, lifecycle gating, bulk linking, and generated projections.
priority: medium
tags: [planning-model, threads, graph, workflow]
created: "2026-08-24"
---
# Threads and task dependency graphs

**Goal.** Implement accepted Threads as initiative views over a planning-space task DAG without overcommitting to speculative graph features.

## Why this is its own epic

Threads are not only a new first-class document. They change task dependency ownership, lifecycle eligibility, repository-wide graph integrity, multi-file composition, CLI and wire projections, and eventually the TUI. That cross-cutting domain deserves a coherent home rather than being split between the generic entity, storage, and CLI epics.

The first work in this epic was deliberately a decision spike. It recommended accepting ADR-0006 and scoping the production implementation; no named risk required another spike. ADR-0006 was accepted on 2026-08-25, and the production tasks below now own delivery.

## Decision gate

ADR-0006 is **accepted**. Work follows the delivery sequence below; this epic does not treat all
Thread work as one implementation task.

## Production task graph

These bootstrap edges are prose until the guarded dependency-write surface lands. Slice 3 must
persist them through the production commands, making this epic the first real dependency dogfood:

```text
6g3q4rst78qy strict reads ---------\
                                      -> 6g3q4rt7mgjn dependency operations
6g3q4rt0wzkq portable mutation guard /                 |
                                                         v
                                      6g3q4rte8kc1 eligibility
                                                         |
                                                         v
                                      6g3q4rtmv4ak Thread entity
                                         |               |
                                         v               v
                              6g3q4rtv8d0a bulk link   6g3q4rv1w9e2 generated views
                                                             |
                                                             v
                                                   6g3q4rv89vzw TUI
```

- [6g3q4rst78qy — strict dependency reads](../tasks/6g3q4rst78qy-establish-canonical-task-dependencies-and-strict-graph-reads.md)
- [6g3q4rt0wzkq — portable mutation guard](../tasks/6g3q4rt0wzkq-make-repository-graph-mutations-portable-and-serializable.md)
- [6g3q4rt7mgjn — dependency operations and queries](../tasks/6g3q4rt7mgjn-ship-guarded-dependency-mutations-and-graph-queries.md)
- [6g3q4rte8kc1 — eligibility enforcement](../tasks/6g3q4rte8kc1-enforce-dependency-eligibility-across-every-task-start-path.md)
- [6g3q4rtmv4ak — Thread entity and projections](../tasks/6g3q4rtmv4ak-add-the-thread-entity-lifecycle-and-graph-projections.md)
- [6g3q4rtv8d0a — resumable bulk linking](../tasks/6g3q4rtv8d0a-bulk-link-existing-tasks-into-threads-with-resumable-apply.md)
- [6g3q4rv1w9e2 — generated graph views](../tasks/6g3q4rv1w9e2-generate-deterministic-thread-graph-views.md)
- [6g3q4rv89vzw — usage-informed TUI](../tasks/6g3q4rv89vzw-add-usage-informed-thread-views-to-the-tui.md)

## Delivery sequence and gates

```text
strict read model -> guarded edge writes -> eligibility enforcement
                                      \-> Thread entity -> bulk linking -> generated views -> TUI
```

Eligibility enforcement and the Thread entity share the same graph foundation. They can be scoped
separately after guarded writes stabilize, but bulk linking waits for both dependency mutation and
Thread persistence.

| Order | Slice | Exit gate | Highest-value stress tests |
|---|---|---|---|
| 1 | Strict dependency read foundation and legacy migration | One deterministic, fail-closed graph snapshot and lint contract; no public graph write yet | malformed/unreadable tasks, ID drift, unknown status, duplicate/self/missing edges, cycles, migration preserving body/frontmatter |
| 2 | Portable guarded edge writes and read queries | Final scan, validation, and write share one store-owned critical section on every supported platform | concurrent opposite edges, direct write versus bulk apply, stale CAS, same edge twice, removal during concurrent reads |
| 3 | Eligibility enforcement | Every route into `in-progress` uses one policy and produces the same blocker/force result | all task statuses, direct/transitive blockers, withdrawn/missing prerequisites, reopen after downstream completion, forced inconsistent work |
| 4 | Thread entity and projections | Membership and lifecycle persist independently from global edges; CLI and wire consume one projection | shared tasks, external gates and rollup denominators, empty/start/complete rules, abandoned/completed drift, membership conflicts |
| 5 | Existing-task bulk linking | One literal-YAML manifest can create a Thread, add memberships and global edges, and converge after interruption | failure after every write prefix, retry/idempotency, wrong planning-space identity, edited/stale plan, concurrent edge mutation |
| 6 | Generated Mermaid/DOT and explanatory UX | Stable ordering and explicit member/external roles; nothing generated is persisted | snapshot/golden output, escaping hostile titles, large/deep/wide readable graphs |
| 7 | Usage-informed TUI | TUI is a consumer of core/wire behavior, not a second graph engine | watcher reload during mutation, parity with CLI state, narrow/small-terminal degradation |

### Design attention

The mutation guard and strict-versus-resilient repository read split are the highest-risk design
work. They protect graph truth; no library can supply them. Lifecycle consistency is next: recursive
sound completion, forced starts, reopen behavior, and every status transition need one authoritative
policy. Bulk apply is a convergence protocol rather than a transaction, so interruption testing must
inject failure after every operation and prove that the same plan repairs the prefix.

The graph library is lower risk. Keep a taskflow-owned interface and run a bounded contract-test
bake-off between the spike's small implementation and `dominikbraun/graph` during slice 1. Do not
create another open-ended research spike, and do not let library features pull critical path, slack,
or other deferred graph analysis into V1.

## Dogfood checkpoints

This epic is the first production consumer of its own capabilities:

1. Slice 1 runs strict read-only analysis against this planning repository.
2. Slice 2 uses production dependency commands to sequence all remaining epic tasks.
3. Slice 4 creates a real Thread for the remaining initiative and observes its frontier and external
   gates during normal implementation work.
4. Slice 5 uses bulk linking on the next naturally suitable initiative rather than a synthetic demo.
5. Every dogfood finding is recorded in the active task; contract changes also amend ADR-0006.

The experimental spike binary is limited to disposable planning spaces and does not satisfy these
checkpoints. Dogfooding begins when the corresponding production slice passes its exit gate.

## Out of scope

- Treating the spike as production implementation.
- Critical-path, slack, forecasting, transitive reduction, or scheduler features.
- Autonomous multi-agent or worktree orchestration.
- TUI implementation before the domain, CLI, and wire projections are proven.

## Related

- [ADR-0006](../adrs/0006-adopt-threads-as-task-dags.md)
- [Vertical MVP decision spike](../tasks/6g3a1wtx4zrr-spike-a-vertical-threads-and-global-task-dag-mvp.md)
- [First-class entities epic](28-first-class-entities-new-planning-nouns.md).
