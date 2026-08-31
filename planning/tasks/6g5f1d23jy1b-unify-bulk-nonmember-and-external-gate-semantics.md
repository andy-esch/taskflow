---
schema: 1
id: 6g5f1d23jy1b
status: completed
epic: 30-threads-and-task-dependency-graphs
description: Distinguish transitive nonmember graph context in bulk manifests from the direct external gates exposed by Thread projections.
effort: 1-2 hours
tier: 2
priority: high
autonomy_level: 3
tags: [threads, graph, cli]
created: "2026-08-31"
updated_at: "2026-08-31"
started_at: "2026-08-31"
completed_at: "2026-08-31"
---

# Unify bulk nonmember and external-gate semantics

## Objective

Resolve the bulk-apply audit's vocabulary contradiction without weakening graph composition. A
`member: false` manifest node is nonmember graph context: it must be transitively upstream of a
declared Thread member, but only a direct prerequisite across the membership boundary is an
external gate in the ordinary Thread projection. Make that distinction explicit in validation,
diagnostics, documentation, and regression coverage before graph renderers consume the projection.

## Contract

- Preserve transitive nonmember nodes so one manifest can express multi-step dependency chains that
  enter a Thread.
- Reserve **external gate** for a direct prerequisite outside the Thread membership set. The default
  Thread projection remains members plus that immediate boundary.
- Treat deeper nonmembers as compose-time graph context. They persist only through the
  repository-global dependency edges, remain reachable through causal blocker queries, and do not
  become implicit Thread members or persisted Thread metadata.
- Do not widen `thread show`, rollups, completion gates, or the forthcoming default graph view to the
  entire transitive prerequisite closure.

## Acceptance criteria

- [x] A manifest chain `context-a -> boundary-b -> member-c` accepts both nonmembers, persists both
  global edges, and creates a Thread containing only `member-c`.
- [x] The resulting Thread projection reports only `boundary-b` as its direct external gate while a
  causal blocker query can still reach `context-a`.
- [x] A disconnected or downstream `member: false` node is rejected as invalid nonmember graph
  context without calling it an external gate.
- [x] ADR, architecture, bulk-task history, CLI help/manifest descriptions, and diagnostics consistently
  distinguish nonmember graph context from direct external gates.
- [x] The originating Claude audit finding is tracked by this task and is marked fixed only after
  implementation and regression coverage land.

## Stress tests

- Multiple members sharing one boundary gate; several nonmember hops; context connected only by
  proposed edges; an already-existing transitive path; disconnected and downstream nonmembers.
- Confirm deterministic compose/apply receipts do not imply nonmember context is persisted in the
  Thread document.

## Out of scope

- Transitive external-gate projection, graph-view traversal switches, schema-2 manifest changes,
  stored nonmember roles, and changes to Thread progress or completion semantics.

## Sequencing

This task follows bulk apply and directly gates deterministic Thread graph views so renderers inherit
one settled projection vocabulary.

## Related

- Epic [30-threads-and-task-dependency-graphs](../epics/30-threads-and-task-dependency-graphs.md)
- ADR [0006 — Adopt Threads as task DAGs](../adrs/0006-adopt-threads-as-task-dags.md)
- Audit [2026-08-30 bulk Thread apply implementation — Claude](../audits/6g5axb85endz-2026-08-30-bulk-thread-apply-implementation-claude.md) — M1

## Implementation outcome (2026-08-31)

Compose retains the existing transitive reachability rule but now names `member: false` tasks as
nonmember graph context in code, diagnostics, CLI help, architecture, the ADR, and the original bulk
task record. The wire schema explicitly defines `external_gates` as direct prerequisites outside the
membership boundary and points deeper context to causal blocker queries.

Core and CLI regressions exercise a proposed two-hop chain, an already-existing transitive chain,
Thread membership persistence, the direct external-gate projection, full causal blocker discovery,
and rejection of disconnected or downstream context. Validation passed with the full Go suite, vet,
golangci-lint, focused race tests, planning lint, audit lint, generated CLI documentation, and diff
hygiene.
