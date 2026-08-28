---
schema: 1
id: 6g3q4rtv8d0a
status: next-up
epic: 30-threads-and-task-dependency-graphs
description: Compose literal-YAML membership and dependency graphs for existing tasks into planning-space-bound resumable apply plans.
effort: 4-7 days
tier: 2
priority: high
autonomy_level: 3
tags: [threads, graph, cli, workflow]
created: "2026-08-25"
updated_at: "2026-08-28"
depends_on: [6g3q4rtmv4ak]
---
# Bulk-link existing tasks into Threads with resumable apply

## Objective

Let users describe membership and dependency relationships among existing tasks in literal YAML, then safely converge the planning repository through a durable apply plan.

## Scope

- Support one new Thread per manifest using local node keys and literal stable `task_id` references; `member: false` represents explicit external gates.
- Materialize a planning-space-bound stable-ID apply plan before mutation.
- Apply membership and repository-global dependency additions with dry-run, per-operation receipts, conflict detection, interruption recovery, and idempotent retry.
- Exclude inline task creation from V1; a future amendment must define creation provenance and already-applied identity before adding it.

## Acceptance criteria

- [ ] Compose validates every task reference, member role, local key, and proposed global edge union without mutation.
- [ ] Apply revalidates current repository health and planning-space identity inside the mutation guard.
- [ ] Compose refuses an ID-less planning space with an actionable identity-migration message; path identity is never substituted silently.
- [ ] Every interrupted write prefix remains graph-valid and retrying the same plan converges without duplicates.
- [ ] Omitted membership or dependencies never imply destructive removal.
- [ ] Human and machine receipts distinguish creates, updates, skips, conflicts, and completion.
- [ ] Existing task edits between retries do not conflict merely because the plan links those tasks; V1 owns only additive membership and dependency intent.
- [ ] One compound mutation capability applies dependency and Thread writes
  under a single repository guard without nesting the narrower mutation ports.
- [ ] Task dependency additions precede the Thread document, every operation
  prefix remains sound, and receipts report the exact durable operation prefix.
- [ ] Compound semantic writes use the caller-provided clock, while idempotent
  skips neither rewrite files nor advance timestamps.

## Stress tests

- Inject failure after every operation; retry each prefix to completion.
- Wrong or absent planning-space identity, edited/stale plans, already-present memberships/edges, concurrent direct dependency mutation, cross-kind ID conflict, and raw hand edits.

## Sequencing

Requires production dependency mutations and Thread persistence. Use the first released version to bulk-link the next naturally suitable initiative.

## Mutation-guard performance gate (2026-08-27)

Before releasing bulk apply, benchmark the real guarded path at representative planning-space and manifest sizes. The current pure prefix validator rebuilds the full graph for every task-file write (O(W × (V+E))) while holding the exclusive repository guard; the adversarial audit measured roughly 442 ms for 1,000 tasks × 300 writes. Keep the simple validator for direct dependency operations, but require an explicit latency budget and move to incremental prefix validation if bulk-scale lock time is material. Include contention and raw-editor-CAS-window observations in the benchmark.

## Compound mutation amendment (2026-08-27)

Apply is one dedicated compound capability, not orchestration across task depend and Thread commands. It takes the canonical-root guard once, reloads planning-space identity, the strict task graph, and relevant Thread state, validates the materialized intent, then invokes lock-free internal materializers. Nesting narrower guarded ports would fail callback exclusion and would not provide one authoritative plan.

For existing-task V1, dependency additions land in the plan's deterministic prefix-safe order and the Thread document lands last. A failure or raw-edit conflict returns the exact durable operation prefix; retry rebuilds current intent and converges without treating unrelated edits between invocations as stale frozen-plan versions. All real semantic writes use the caller clock and idempotent skips remain byte-identical.

The existing performance gate remains an exit criterion for this compound path, including lock-held validation time, callback-contention behavior, and the immediate per-target CAS window.
