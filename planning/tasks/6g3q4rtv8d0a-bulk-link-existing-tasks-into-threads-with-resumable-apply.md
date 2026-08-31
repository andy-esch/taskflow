---
schema: 1
id: 6g3q4rtv8d0a
status: completed
epic: 30-threads-and-task-dependency-graphs
description: Compose literal-YAML membership and dependency graphs for existing tasks into planning-space-bound resumable apply plans.
effort: 4-7 days
tier: 2
priority: high
autonomy_level: 3
tags: [threads, graph, cli, workflow]
created: "2026-08-25"
updated_at: "2026-08-31"
depends_on: [6g4wm2yf6tyj]
started_at: "2026-08-30"
completed_at: "2026-08-31"
---
# Bulk-link existing tasks into Threads with resumable apply

## Objective

Let users describe membership and dependency relationships among existing tasks in literal YAML, then safely converge the planning repository through a durable apply plan.

## Scope

- Support one new Thread per manifest using local node keys and literal stable `task_id` references;
  `member: false` supplies nonmember graph context without entering persisted membership.
- Materialize a planning-space-bound stable-ID apply plan before mutation.
- Apply membership and repository-global dependency additions with dry-run, per-operation receipts, conflict detection, interruption recovery, and idempotent retry.
- Exclude inline task creation from V1; a future amendment must define creation provenance and already-applied identity before adding it.

## Acceptance criteria

- [x] Compose validates every task reference, member role, local key, and proposed global edge union without mutation.
- [x] Apply revalidates current repository health and planning-space identity inside the mutation guard.
- [x] Compose refuses an ID-less planning space with an actionable identity-migration message; path identity is never substituted silently.
- [x] Every interrupted write prefix remains graph-valid and retrying the same plan converges without duplicates.
- [x] Omitted membership or dependencies never imply destructive removal.
- [x] Human and machine receipts distinguish creates, updates, skips, conflicts, and completion.
- [x] Existing task edits between retries do not conflict merely because the plan links those tasks; V1 owns only additive membership and dependency intent.
- [x] One compound mutation capability applies dependency and Thread writes
  under a single repository guard without nesting the narrower mutation ports.
- [x] Task dependency additions precede the Thread document, every operation
  prefix remains sound, and receipts report the exact durable operation prefix.
- [x] Compound semantic writes use the caller-provided clock, while idempotent
  skips neither rewrite files nor advance timestamps.
- [x] Existing-task V1 does not manufacture task-creation projection impacts. If inline
  `new_task` is later promoted, its compound planner must derive one receipt across task creation,
  dependency, and membership changes rather than reuse the single-task lifecycle impact shortcut.

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

The current task-lifecycle impact helper intentionally returns no Thread impacts for an isolated
task creation: a freshly minted task cannot already participate in a persisted Thread or dependency
edge. That shortcut is not valid for a future compound `new_task` apply which creates membership and
edges in the same plan; such an extension must compare the compound before/after repository
projection instead.

The existing performance gate remains an exit criterion for this compound path, including lock-held validation time, callback-contention behavior, and the immediate per-target CAS window.

## V1 production contract (2026-08-30)

The authoring manifest is a strict, literal YAML/JSON document. `schema` may be omitted for the V1
authoring shape or set to `1`; unknown fields are rejected. Every node has one unique local `key`,
one exact stable `task_id`, and optional `member` (default true). Inline `new_task`, partial task
references, duplicate task declarations, duplicate edges, and shell interpolation are not accepted.
At least one member is required. A `member: false` node must be an actual transitive prerequisite of
a declared member in the proposed repository graph; otherwise compose rejects it as invalid
nonmember graph context. Only direct prerequisites outside the final membership set are external
gates in ordinary Thread projections. Deeper nonmembers remain visible through causal graph queries
but are not persisted as a Thread role.

Compose emits a strict schema-1 materialized plan containing the durable planning-repository ID,
compose date, one preallocated unstarted Thread (including its fully rendered body and sorted stable
member IDs), and sorted stable-ID dependency additions. The output path is created with no-clobber
semantics; compose never replaces a recovery token. Global `--dry-run` performs the same reads and
validation but does not create the output file. The apply command accepts only a durable file path,
not stdin.

Apply treats dependency arcs and Thread creation as additive intent. It reloads identity, the strict
task graph, and every Thread under one canonical-root guard; unions intended edges with each current
task's dependencies; validates the complete graph and every deterministic task-write prefix; and
creates the Thread last. Existing intended edges are skips, unrelated task edits and additional
dependencies are preserved, and a fully identical pre-existing planned Thread is an already-applied
skip. A same-ID Thread with different metadata, membership, lifecycle timestamps, or body is a
conflict rather than an overwrite. Raw edits are checked before the first write and again per target
where possible. Apply independently rejects a hand-edited memberless plan, preserving bulk compose's
narrower at-least-one-member rule even though ordinary empty Threads remain legal.

Receipts list every edge and the Thread in deterministic order with `pending`, `applied`, or
`skipped` state. Dry-run leaves needed operations pending. A real successful or already-converged
apply is complete; a failure reports the exact durable prefix and remaining pending operations.
Cleanup failure after the final durable write is classified as committed and is never retried.

## Implementation progress (2026-08-30)

Production now exposes `thread compose --from <manifest> --out <plan>` and
`thread apply <durable-plan>`. Compose uses strict single-document YAML/JSON decoding with unknown
field rejection, resolves only exact existing task IDs, validates `member: false` as real transitive
upstream graph context, renders the ordinary Thread template, and writes a schema-1 mode-`0600`
no-clobber recovery plan. Dry-run performs the same semantic validation and prints the plan without
creating it.

Apply is one `ThreadApplyMutationStore` operation. The CLI injects live config rediscovery into the
filesystem adapter; while holding the canonical-root guard, the adapter verifies physical root and
durable planning ID, reloads the strict task graph and complete Thread bodies, invokes the pure
additive planner behind the callback re-entry sentinel, materializes dependency-owner task writes,
and creates the Thread last. Whole-source and immediate-target CAS checks cover observed raw edits;
a final convergence re-read also runs when the Thread already exists, preventing an undone repair
from being reported complete.
Receipts retain every edge plus the Thread as pending/applied/skipped, distinguish complete from
committed, and survive both an interrupted durable prefix and post-final cleanup failure. Automatic
retry is limited to a pre-commit conflict with work still pending.

Focused coverage includes strict/unknown/multi-document input, no-clobber plans, missing/wrong/moved
identity evidence, unsafe hand-edited slugs, cross-kind ID collision, cycles, disconnected or
downstream nonmember context, additive unrelated task state, already-present edges and Thread,
byte-identical no-op retry,
injected failure and successful retry after every durable operation prefix, raw edits before and after
a durable prefix, a raw edit undoing a repair when the Thread already exists,
whole-source task and Thread races before the first write, an exact concurrent Thread create,
same-ID/different-body and post-lifecycle conflicts, planner re-entry, and serialization with a real
direct dependency mutation through another store.

The representative guarded benchmark measures 1,000 tasks and 300 dependency-owner writes. Review
showed that the generic per-prefix graph rebuild consumed roughly 59% of the planning phase despite
being redundant for pure edge supersets. The validator now uses final-graph validation for that
monotone case while retaining per-prefix checks for removals and legacy clearing. On an Apple M5,
dry-run read/validate/materialize is now approximately 0.22 seconds and real apply approximately
3.24 seconds. The ADR retains a one-second planning-phase and five-second total reference budget at
that scale; atomic file durability remains the dominant total lock hold.
