---
schema: 1
id: 6g3q4rtv8d0a
status: next-up
epic: 30-threads-and-task-dependency-graphs
description: Compose literal-YAML membership and dependency graphs for existing tasks into planning-space-bound resumable apply plans.
effort: 3-5 days
tier: 2
priority: high
autonomy_level: 3
tags: [threads, graph, cli, workflow]
created: "2026-08-25"
---
# Bulk-link existing tasks into Threads with resumable apply

## Objective

Let users describe membership and dependency relationships among existing tasks in literal YAML, then safely converge the planning repository through a durable apply plan.

## Scope

- Support one new Thread per manifest using local node keys and literal stable `task_id` references; `member: false` represents explicit external gates.
- Materialize a planning-space-bound stable-ID apply plan before mutation.
- Apply membership and repository-global dependency additions with dry-run, per-operation receipts, conflict detection, interruption recovery, and idempotent retry.
- Keep inline task creation outside the required production path until usage justifies it.

## Acceptance criteria

- [ ] Compose validates every task reference, member role, local key, and proposed global edge union without mutation.
- [ ] Apply revalidates current repository health and planning-space identity inside the mutation guard.
- [ ] Every interrupted write prefix remains graph-valid and retrying the same plan converges without duplicates.
- [ ] Omitted membership or dependencies never imply destructive removal.
- [ ] Human and machine receipts distinguish creates, updates, skips, conflicts, and completion.

## Stress tests

- Inject failure after every operation; retry each prefix to completion.
- Wrong planning space, edited/stale plans, already-present memberships/edges, concurrent direct dependency mutation, same-ID conflict, and raw hand edits.

## Sequencing

Requires production dependency mutations and Thread persistence. Use the first released version to bulk-link the next naturally suitable initiative.
