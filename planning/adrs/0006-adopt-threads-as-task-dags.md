---
status: accepted
date: "2026-08-21"
deciders: [andy-esch]
tags: [adr, planning-model, graph, workflow, threads]
supersedes:
  - ADR-0002
superseded_by: null
---

# ADR-0006: Adopt Threads as Initiative Views over a Global Task DAG

> ✔ **Accepted 2026-08-25 — finalized.** The vertical MVP spike validated the central model and
> identified the production safety gates captured below. Decision sections are now frozen; record
> later clarifications under `## Amendments` and reverse the decision through a superseding ADR.

> Follows the ADR format established in [0001-adopt-adrs](0001-adopt-adrs.md). Formally
> **supersedes [0002-adopt-projects](0002-adopt-projects.md)** ("Adopt Projects"), replacing
> flat task buckets with named views over a planning-repository task DAG. Builds upon the flat,
> ID-addressed storage foundation of
> [0003-stable-key-id-addressed-storage](0003-stable-key-id-addressed-storage.md).
>
> **Grounding note — revised 2026-08-24:** This proposal was rewritten after stress-testing it
> against the current domain, store, CLI, wire, initialization, and TUI seams. Obsolete
> library-first, scheduler, stored-diagram, and project-bucket claims were removed. Remaining
> uncertainties are called out as deferred decisions rather than implied implementation.
>
> **Acceptance evidence:** The bounded
> [vertical MVP spike](../tasks/6g3a1wtx4zrr-spike-a-vertical-threads-and-global-task-dag-mvp.md)
> recommended acceptance after executable filesystem, concurrency, interruption, and manual-use
> validation. Production rollout remains governed by the dependency-ordered slices in Section 10.

## Context and Problem Statement

`tskflwctl` currently structures work across two primary abstractions:

1. **Epics ([`planning/epics/`](../epics)):** Long-lived thematic domains (e.g. `cli-ux`,
   `storage-engine`). Tasks belong to **exactly one** epic, representing their permanent
   taxonomic home. Epics never truly "complete."
2. **Tasks ([`planning/tasks/`](../tasks)):** Atomic units of execution with lifecycle
   statuses (`ready-to-start`, `in-progress`, `completed`, etc.) and acceptance criteria.

### The Limitation of Flat "Projects" (ADR-0002)

[ADR-0002](0002-adopt-projects.md) proposed **Projects** as cross-cutting groupings of tasks
targeting a shared milestone. However, ADR-0002 was never implemented (0 project files exist
beyond scaffolding placeholders). More fundamentally, **a flat project bucket lacks causal
structure, dependencies, and execution sequencing**.

In real-world software engineering and autonomous AI agent workflows, non-trivial initiatives
are not unordered bags of tasks:

- Core schemas must be designed before storage adapters can be written.
- Storage adapters must land before CLI commands or TUI screens can consume them.
- Backend APIs and frontend components must merge before end-to-end integration tests run.

A flat project cannot answer critical operational questions:

- *Which tasks are unblocked and ready for immediate parallel execution?*
- *What prerequisite chain explains why a task is blocked?*
- *If a task is delayed, what downstream work is impacted?*
- *How is one cross-epic initiative progressing toward its finish line?*

We need a planning primitive that captures both the **cohesive container** of an initiative
and the **causal dependency graph (DAG)** connecting its tasks.

## Considered Options

- **Option A — Implement ADR-0002 as written (Flat Projects).**
  Provides basic cross-epic grouping with `done / total` rollups.
  *Rejected:* Leaves execution ordering and dependency management to external prose, forcing
  agents and humans to manually script execution loops and deduce blockers by hand.
- **Option B — Decentralized task-level dependencies only (`depends_on: []` on tasks, no workflow entity).**
  Tasks declare prerequisites directly in their frontmatter.
  *Rejected as the complete model:* Lacks a cohesive initiative-level container (goal, target
  date, lifecycle, and dedicated rollup). The dependency graph has no named finish line.
- **Option C — Store dependencies inside each Thread.**
  Each Thread document owns both its members and edges.
  *Rejected:* A task may belong to multiple Threads, while a prerequisite is a global constraint
  within the planning repository. Two individually acyclic Threads could declare contradictory
  edges whose global union is cyclic, and `task blockers` would have no single answer.
- **Option D — Threads own membership; tasks own global dependencies (Chosen).**
  The planning repository has one task DAG. A Thread is a first-class, many-valued initiative
  view over a subset of that graph, with its own goal, lifecycle, and rollup. An edge is stored
  once, on the dependent task.

The implementation choice between an off-the-shelf Go graph library and a small taskflow-owned
algorithm package is deliberately **not decided here**. It follows from the required operations
and a focused package evaluation; no third-party graph type may leak into the domain, persistence,
or wire contracts.

## Decision

Adopt **Threads** as the primary cross-cutting initiative abstraction in `taskflow`, together
with one canonical, global task-dependency relation inside each planning repository. This
formally supersedes ADR-0002.

### 1. Conceptual Model: One Repository DAG, Many Thread Views

For one planning repository, let $G = (V, E)$ be the global Directed Acyclic Graph:

- **Vertices ($V$):** every task in that planning repository, addressed by stable task ID.
- **Edges ($E$):** $(u, v)$ exists when task $v$ declares `depends_on: [u]`. The edge is a hard,
  AND-style prerequisite: every declared predecessor must be satisfied before ordinary work on
  $v$ may start.
- **Scope:** the graph belongs to the planning repository, not to an implementation repository.
  One planning space may coordinate multiple implementation repos (as in desirelines); that is
  still one planning graph. A Thread never spans planning spaces.

A **Thread** pairs an initiative definition with a member set $M \subseteq V$:

- A task may belong to zero, one, or many Threads.
- The Thread's internal graph is the projection of $G$ over $M$.
- A direct prerequisite outside $M$ remains an **external gate**. It affects member readiness and
  appears in graph/query output with properties such as task ID, slug, status, epic, and
  `external: true`, but it does not enter the Thread's progress denominator.
- Full blocker queries may traverse beyond those immediate boundary vertices through the global
  graph; Thread graph views may keep the default display to members plus immediate external gates.
- A flat project is the special case of a Thread whose members have no internal dependency edges.

Membership does not own dependency truth. Removing a task from a Thread, completing or cancelling
a Thread, or deleting a Thread must not remove task dependencies that may constrain other Threads.

```text
                    ┌───────► Task 2 (UI Model) ───────┐
                    │                                   ▼
Task 1 (Core Schema)┤                        Task 5 (E2E Integration)
                    │                                   ▲
                    └───────► Task 3 ──► Task 4 ────────┘
                             (Store)   (CLI Client)
```

### 2. Dependency Ownership and Integrity

The dependent task is the single source of truth:

```yaml
depends_on: [6fjangd7kvh0, 6fjangd7kvh2]
```

- `depends_on` is a duplicate-free semantic set of stable task IDs, never slugs. It is serialized
  in stable ID order for reviewable diffs. There is no stored inverse `blocks` field; downstream
  relations are computed.
- `task depend add <task> --on <prerequisite>...` and `task depend remove ...` are the direct
  mutation surface. A Thread-oriented edge command, if retained as convenience syntax, must say
  that it changes a repository-global task dependency and must require both tasks to be Thread
  members; it does not create a Thread-owned edge.
- Every dependency mutation validates exact task resolution, duplicate/self edges, and acyclicity
  against the repository graph before writing. `--dry-run` performs the same validation.
- The final repository scan, graph validation, and write must occur inside one planning-repository
  mutation guard. Per-file optimistic concurrency alone is insufficient: two individually valid
  concurrent edge additions can form a cycle when combined. The existing Unix repository lock is
  a starting point, but the implementation must define equivalent correctness on platforms where
  that lock is currently a no-op.
- Graph mutation fails closed when the repository dependency graph cannot be read soundly. `lint`
  remains the fail-open, full-sweep diagnostic surface for hand-edited missing IDs, unreadable
  tasks, legacy fields, and cycles.
- Every immutable snapshot reports one repository-level health marker. `healthy` means the
  canonical graph is valid and no legacy dependency vocabulary remains. `degraded` means every
  legacy reference resolves exactly but has not yet been migrated; diagnostic reads may explain
  it, while mutation and dispatch still fail closed. `broken` means any task is unreadable, an ID
  or status is invalid, an edge is duplicate/self/invalid/missing, a legacy reference is missing or
  ambiguous, or the graph is cyclic. `broken` takes precedence over `degraded`.
- Supported commands prevent graph-invalid states rather than relying on lint after the fact.
  Generic set/edit paths treat both `depends_on` and the legacy dependency fields as graph-owned,
  including under `--force`; guarded dependency commands are the only product write path. Direct
  filesystem edits and older binaries remain possible, so every structural defect and legacy
  occurrence must also appear during an ordinary `lint` call with deterministic attribution.
- The existing unmodelled `dependencies`, `blocked_by`, and `blocks` fields are legacy vocabulary.
  Implementation must migrate the six current `blocked_by` users or report them with actionable
  lint, then converge on `depends_on` alone. The unused task-level `projects` field is deprecated
  separately: Thread membership belongs only in Thread documents, so it must not be renamed to
  `threads` on tasks.

### 3. Persisted Lifecycle and Derived Graph State

Dependencies do not add a persisted `blocked` status. Task `status` remains the author-declared
lifecycle; graph state is a derived projection that cannot drift from the edges and current task
statuses.

`next-up` and `ready-to-start` are both pending-work states. `next-up` says the task is selected into
the working queue; `ready-to-start` additionally says its scope is suitable for handoff. Neither
status promises that external constraints are currently satisfied, and `ready-to-start` is not a
mandatory authorization hop before execution. `eligible` is the derived word for pending work whose
repository graph is healthy and whose dependency gate is clear. This preserves the useful
author-declared readiness distinction without letting it override or duplicate graph truth.

A task is **soundly completed** when its own status is `completed` and every direct prerequisite is
also soundly completed. Because the graph is acyclic, that recursive definition terminates. It makes
reopen behavior explicit: reopening an upstream task invalidates completed descendants as dependency
satisfiers until the chain is repaired, even though their persisted status remains `completed`.

V1 exposes lifecycle role and dependency health as separate derived fields rather than collapsing
them into a shadow status vocabulary:

- **Lifecycle role** follows the persisted status: queued (`next-up`), candidate
  (`ready-to-start`), in-flight (`in-progress`), parked (`deferred`), nominally complete
  (`completed`), or withdrawn (`deprecated`).
- **Gate state** is `clear` when every prerequisite is soundly completed; `blocked` when the graph
  is readable but at least one prerequisite is not soundly completed; or `broken` when an upstream
  path contains a missing, unreadable, withdrawn, invalid-status, cyclic, or otherwise recursively
  broken prerequisite.

Blocker projections use stable reason tokens: `not-started`, `in-flight`, `parked`, `withdrawn`,
`missing`, `unsound-completed`, `invalid-status`, and `cycle`. The last two are forensic vocabulary,
not states that supported commands may create: they let lint and diagnostic reads explain damage
from direct filesystem edits, old binaries, or an interrupted external writer without collapsing it
into a vague missing/blocked result. Every blocker also carries one deterministic shortest path from
the queried task to that blocker.

Named views are compositions of those fields:

- **Eligible / Frontier:** queued or candidate plus a clear gate in a healthy graph. The Thread
  frontier is the eligible subset of its members; upstream sound completion, not a mandatory
  `ready-to-start` hop, determines which pending tasks reach it.
- **Drained:** nominally complete plus a clear gate; equivalently, soundly completed.
- **Inconsistent:** in-flight or nominally complete with a blocked/broken gate—for example a forced
  start, a completed task whose prerequisites never caught up, or completed downstream work whose
  prerequisite was reopened.

A queued or candidate task may independently have a clear, blocked, or broken gate. Either role is
eligible when the graph is healthy and its gate is clear. In-flight, parked, nominally complete,
and withdrawn tasks are not work waiting to start, so they stay outside the frontier; parked and
withdrawn tasks also never satisfy downstream dependencies.

Every product path that enters `in-progress`—the `task start` verb, generic `task move`,
`task new --start`, and future TUI actions—must call the same core eligibility guard. An existing
task may enter from either `next-up` or `ready-to-start` when its gate is clear. The transition
refuses an ineligible task by default. `--force` is an explicit escape hatch available from either
pending-work state: it bypasses only the dependency gate, does not bypass broken graph evidence or
other lifecycle roles, does not remove dependencies, and returns a receipt that says the transition
was forced and names the outstanding blockers. The task remains derived as inconsistent until its
prerequisites become soundly completed or the graph is corrected.

Hand edits, or completing a force-started task before its prerequisites catch up, can still produce
a nominally completed-but-unsound task. Such a task does not unblock descendants; `lint`, Thread
views, and blocker queries surface the inconsistency. V1 does not invent a separate dependency-waiver
entity: correct the edge when the constraint is no longer real.

### 4. Thread Lifecycle

Thread lifecycle remains explicit, with successful completion distinct from intentionally stopping
the initiative:

```text
unstarted ──> in-progress ──> completed
    │              │
    └──────────────┴────────> cancelled

completed ──thread reopen──> in-progress
```

- `thread start` requires at least one non-withdrawn member, stamps `started_at`, and changes only
  the Thread.
- Thread progress reuses the existing `TaskRollup` rule: deprecated members are reported but
  excluded from `done / total`, deferred members remain in the denominator, and external gates do
  not count. Graph health is reported beside that nominal rollup rather than folded into it.
- `thread complete` succeeds only when every non-withdrawn member is soundly completed and no
  member path is broken or inconsistent. Deferred members prevent completion. A deprecated member
  cannot remain as an unsatisfied prerequisite of live member work.
- `thread cancel` may stop an unstarted or in-progress initiative without claiming successful
  completion. Cancellation is terminal and membership-immutable and never mutates member tasks;
  they may belong to other Threads or remain useful off-Thread. V1 does not reopen a cancelled
  Thread.
- A completed Thread is not membership-mutable until an explicit `thread reopen`, which returns it
  to `in-progress` and clears its terminal stamp. Task status and task-owned dependencies can still
  change through their global commands; those changes can make a completed Thread inconsistent and
  must report the affected Threads. The exact reopen diagnosis/repair UX deserves focused
  follow-up but is not a blocker to adopting the model.

### 5. Required Graph Projections and Package Boundary

The first implementation needs a small, deterministic graph-analysis contract:

1. Global cycle validation with an attributable cycle/path error.
2. Lifecycle-filtered, graph-derived Thread frontier, failing closed for defects in the relevant
   graph.
3. Full upstream blocker and downstream impact queries over global dependencies.
4. A deterministic topological plan. Its waves/generations are **explanatory**, not an execution
   scheduler or barrier protocol.
5. Runtime graph export. Mermaid and DOT are straightforward textual projections; ASCII/Unicode is
   included only if the selected renderer makes it maintainable. Generated output is never stored
   in the Thread document.

All otherwise-equal ordering uses stable task ID as the tie-breaker. Thread plans rank member tasks
only; they list outside prerequisites as marked gates rather than silently treating external work as
Thread-owned execution. Section 11 bounds the deliberately deferred graph features, while the
implementation slices preserve the planned TUI follow-up.

The validation spike completed enough package research to avoid a second open-ended research task.
V1 defaults to the small owned algorithms proven by the spike. During the dependency-foundation
slice, run one bounded implementation-time bake-off: execute the same taskflow-owned contract tests
against the owned implementation and a thin [`dominikbraun/graph`](https://github.com/dominikbraun/graph)
adapter. Adopt the dependency only if it materially removes maintained code while preserving stable
task-ID ordering, attributable cycle diagnostics, and taskflow's error contracts; its explicitly
unstable v0 API counts against marginal savings. Gonum is not a second mandatory candidate unless
the required analysis surface grows beyond V1. No graph-library type may cross the taskflow-owned
analysis interface into domain, persistence, or wire contracts.

### 6. Document Layout & Frontmatter Specification

Threads are stored in `planning/threads/<id>-<slug>.md` using flat, ID-addressed storage
([ADR-0003](0003-stable-key-id-addressed-storage.md)).

```yaml
---
schema: 1
id: 6g2b01v9ck2w
status: in-progress      # unstarted | in-progress | completed | cancelled
description: Consolidate configuration lifecycle into unified hub
goal: Ship unified config CLI and interactive TUI routes
target_date: "2026-09-01"
created: "2026-08-21"
started_at: "2026-08-21"
ended_at: null
tags: [cli, tui, config]
tasks:
  - 6fjangd7kvh0
  - 6fjangd7kvh1
  - 6fjangd7kvh2
  - 6fjangd7kvh3
  - 6fjangd7kvh4
---

# Thread: Unified Navigation Hub

## Context
Why this initiative exists, important constraints, and decisions that do not belong on one task.
```

The Thread file owns initiative metadata and membership only. `description`, `goal`, `status`,
`created`, and `tasks` are required; `target_date`, lifecycle timestamps, and tags are optional.
`target_date` is a human planning constraint, not a calculated forecast. An empty member set is
valid while an unstarted Thread is being scoped, but it cannot start or complete.

`tasks` is a duplicate-free semantic set serialized in stable ID order; list position carries no
execution meaning. Task files own `depends_on`, and graph commands join the two at runtime.
Thread IDs reuse ADR-0003's task/research conventions: 12-character stable ID in both filename
and frontmatter, exact-ID resolution, drift lint, and rename-safe identity. The body is free-form
narrative, not a generated graph or shared execution log.

### 7. Bulk Linking and Composition are First-Class Use Cases

Scoping often discovers the relationships among a set of already-created tasks at once. A manifest
can bulk-link those tasks into one new Thread and add their repository-global dependency edges in one
operation; it does not recreate or move the tasks. This is the V1 authoring path. Inline task
creation is explicitly outside V1: it would add creation provenance, epic revalidation, and a
different prefix-order problem before existing-task bulk linking has earned that complexity.

```yaml
thread:
  title: Unified navigation hub
  description: Consolidate configuration lifecycle into one hub
  goal: Ship unified config CLI and TUI routes
nodes:
  - key: config-model
    task_id: 6fjangd7kvh1
  - key: existing-cli
    task_id: 6fjangd7kvh2
  - key: legacy-gate
    task_id: 6fjangd7kvh0
    member: false
dependencies:
  - {from: legacy-gate, to: config-model}
  - {from: config-model, to: existing-cli}
```

- Every node has a manifest-local, unique `key` and one exact stable `task_id`. The task must already
  exist. Slugs, partial references, and inline `new_task` specifications are rejected.
- A V1 manifest creates exactly one new Thread. Extending an existing Thread remains available
  through the ordinary membership/dependency commands until real use justifies update manifests.
- `member` defaults to true. `member: false` declares existing nonmember graph context that must be
  transitively upstream of at least one member. It is an authoring role, not persisted Thread
  metadata. Runtime output derives direct external gates from the global graph regardless of which
  nonmembers were declared in the manifest.
- Dependency arcs point from prerequisite to dependent. They add repository-global `depends_on`
  relations; the manifest does not own those edges and omission never removes an existing edge.
- Local keys are authoring conveniences only. Persisted Thread membership and task dependencies use
  stable IDs.

An authoring manifest is literal YAML; taskflow does not interpolate shell variables in it. In the
primary existing-task workflow, `task_id` contains the copied stable task ID—not a slug, title, or
`$VARIABLE`—and CLI/docs must make those IDs easy to discover. Compose resolves local keys and emits
only stable IDs into the materialized apply plan. A later ADR amendment may introduce `new_task`
only after it defines creation provenance, already-applied identity, and recovery ordering.

#### Compile, Then Apply

A one-shot command that generates fresh IDs during each retry is not safe enough for a multi-file
operation. Bulk composition uses two phases:

```bash
tskflwctl thread compose --from thread-plan.yaml --out thread-apply.yaml
tskflwctl thread apply thread-apply.yaml
tskflwctl thread apply thread-apply.yaml --dry-run --json
```

1. **Compose** reads the repository, validates the entire authoring manifest, mints the new Thread
   ID, resolves every local key, validates exact existing-task references, and checks the proposed
   edge union for cycles. It writes a
   materialized apply plan before any planning entity is mutated. The plan is bound to the
   planning-space identity and records stable IDs, the exact intended Thread creation, and additive
   membership/edge changes. It is an intent log, not a frozen repository snapshot: unrelated
   repository changes between compose and apply remain legal when current-state revalidation passes.
2. **Apply** revalidates the complete materialized plan, then converges the repository toward it.
   The missing planned Thread is created with its preallocated ID; an identical Thread or already
   present edge addition is skipped; a same-ID/different-identity collision or invalid current graph
   stops with `ErrConflict` or validation. Apply rechecks every referenced task and Thread invariant
   rather than trusting compose-time validation. Reapplying the same plan cannot mint another ID.

Compose supports stdin for the authoring manifest, but a materialized plan must have a durable
path before apply begins. A later convenience command may compose and apply in one invocation only
if it first saves and reports that recovery plan.

The store provides content-version checks and atomic per-file replacement; it does not provide an
all-files transaction. Real apply holds the planning-repository mutation guard across its final
read, graph validation, and write sequence so cooperating taskflow writers cannot interleave a
globally invalid edge. It validates everything it can before the first write, reports a
per-entity/edge receipt, stops on the first conflict, and is safe to resume with the same plan.
It does not claim rollback or isolation from raw hand edits. `lint` must diagnose any remainder
after interruption.

Every persisted prefix must itself remain a sound repository graph. Existing-task V1 writes
dependency-owner task files in deterministic, prefix-validated order and creates the Thread last.
The Thread therefore never advertises sequencing that has not already become durable. A future
inline-task extension must solve vertices-before-edges ordering separately; it cannot inherit the
existing-task write order by assumption.

### 8. CLI Surface

```bash
# Thread authoring and membership
tskflwctl thread new "Title" --description "..." --goal "..." [--target-date YYYY-MM-DD]
tskflwctl thread add <thread> <task-id>...
tskflwctl thread remove <thread> <task-id>...
tskflwctl thread set|edit|rename <thread> ...
tskflwctl thread info|path <thread>

# Bulk graph composition
tskflwctl thread compose --from <yaml-or-json-file|-> --out <apply-plan>
tskflwctl thread apply <apply-plan> [--dry-run]

# Repository-global task dependencies
tskflwctl task depend add <task> --on <prerequisite>...
tskflwctl task depend remove <task> --on <prerequisite>...
tskflwctl task blockers <task-id>
tskflwctl task unblocks <task-id>

# Lifecycle and progress
tskflwctl task start <task-id> [--force]
tskflwctl thread start|complete|cancel|reopen <thread>
tskflwctl thread list [--status <status>]
tskflwctl thread show <thread>               # Rollup, blockers, external gates, frontier
tskflwctl thread frontier <thread>           # Machine list of graph-clear pending members
tskflwctl thread plan <thread>               # Explanatory topology/waves
tskflwctl thread graph <thread> [--format mermaid|dot|ascii]

# Existing task queries gain Thread filters
tskflwctl task list --thread <thread> [--unblocked]
```

Mutations participate in the existing global `--dry-run` and output-mode contracts. Machine output
must expose stable IDs, membership versus external-gate roles, derived graph state, blockers, and
forced-transition metadata without requiring callers to parse human rendering.

### 9. Planning Layout and Implementation Boundary

`planning/threads/` replaces the unimplemented `planning/projects/` scaffold:

- Add `ThreadsDir = "threads"`; `init` creates it and stops creating `ProjectsDir`.
- A repository migration may remove an empty `projects/` placeholder. It must not delete a non-empty
  legacy directory automatically; report it as deprecated content requiring an explicit migration.
- Update space-health/layout discovery so a planning repository remains one coherent planning space.
  The implementation repositories it coordinates do not gain their own copies of Thread state.

Threads are first-class documents, so the implementation must deliberately cover the same seams as
other entities: domain and validation, parser/store and optimistic concurrency, core use cases, CLI
and completion, schema/templates, wire envelopes and JSON schema, initialization/layout, lint, and
space health. The current entity descriptor reduces some enumeration, but store, core, render, and
TUI paths are still partly per-entity; that fan-out is implementation cost, not evidence that Thread
should be encoded as a task field.

Dependency analysis belongs in pure core/domain-facing code over task IDs and statuses. Filesystem
code parses and atomically mutates documents but does not decide frontier, sound completion, or
topological semantics. Introduce narrow consumer-owned dependency and Thread persistence ports;
keep the selected graph package, if any, behind the taskflow-owned analysis contract.

### 10. Implementation Slices

If the validation spike recommends acceptance and the decider accepts this ADR, production work can
be scoped into the following dependency-ordered slices without requiring the deferred features:

1. **Strict dependency read foundation:** add modeled `Task.DependsOn`, replace and migrate the
   legacy dependency vocabulary, define the taskflow-owned graph-analysis contract, and add strict
   graph snapshot/lint behavior. Run the bounded library bake-off here. Do not expose graph writes
   while repository health and deterministic diagnostics are still unsettled.
2. **Portable guarded dependency writes:** establish the cross-platform repository mutation guard,
   then add dependency add/remove, dry-run, blocker/downstream, and explanatory-plan operations. The
   final read, cycle validation, and write remain one store-owned critical section.
3. **Eligibility enforcement:** centralize the transition policy and route every path into
   `in-progress` through it, including generic move, start, create-and-start, and later TUI actions.
   Ship `--force` only with an explanatory receipt. Do not release partial enforcement that callers
   can bypass through another verb.
4. **Thread documents, guarded creation, and read projections:** add domain/store/core/wire/CLI
   support, guarded creation in the explicit `unstarted` state, a reusable lock-free materializer,
   nominal/sound rollups, external gates, frontier, and initialization migration from the unused
   Projects scaffold. This consumes the shared analysis contract rather than reimplementing graph
   state.
5. **Guarded Thread membership and lifecycle:** add member add/remove and start, complete, cancel,
   and reopen over one authoritative task/Thread snapshot; retain committed outcomes and augment
   task-lifecycle receipts with affected Thread IDs.
6. **Bulk linking of existing tasks:** ship YAML compose/apply for existing stable task IDs, with a
   durable materialized plan, planning-space binding, dry-run, resumable receipts, and safe retry.
   The prototype's inline `new_task` path is an optional follow-up and is not required to prove the
   primary bulk-linking workflow.
7. **Generated views:** add deterministic Mermaid/DOT and polish explanatory plans after the shared
   projection has CLI usage. Generated output is never persisted as Thread state.
8. **Planned TUI follow-up:** after CLI and wire behavior have usage feedback, add a Thread list/tab
   and detail view showing lifecycle, rollup, frontier, member/external distinction, and a readable
   graph. The first TUI slice should consume the same projections and should not introduce direct
   graph editing or a separate readiness calculation.

### 10.1. Dogfooding and Rollout Policy

Threads and task dependencies must manage their own remaining implementation work as soon as each
production safety boundary permits it. This is a product-validation policy, not an invitation to use
the spike adapter on canonical planning data:

- After the strict read foundation lands, run its graph health and explanatory queries against this
  planning repository before enabling writes.
- After guarded dependency writes land, sequence the remaining tasks in this epic with real
  dependencies using the public commands. Do not invent dependencies merely to exercise a feature.
- After Thread documents and guarded creation land, make this initiative one of the first real
  Threads and inspect its read projections. After guarded membership/lifecycle lands, manage that
  Thread through the production mutation verbs. Exercise shared membership and external gates when
  actual work naturally has those relationships.
- After bulk linking lands, use a literal-YAML manifest to establish or extend the next real
  initiative and retain its materialized plan until apply reports completion.
- Record confusing output, forced transitions, reopen behavior, merge conflicts, and recovery work
  in the owning implementation task. Contract-level findings amend this ADR before TUI behavior is
  treated as settled.

Dogfooding never bypasses the slice exit gates. In particular, canonical planning data must not be
written by the experimental `threadspike` adapter; production migration, lint, locking, and wire
contracts must land first.

### 11. Explicitly Out of Scope for This Decision

- critical path, slack, effort weighting, target-date forecasting, and bottleneck scoring;
- autonomous dispatch, multi-agent/worktree orchestration, merge ordering, retries, or barriers;
- transitive reduction or graph simplification commands;
- soft, OR, conditional, time-based, or cross-planning-space dependency edges;
- stored diagrams, stored rollups, or a shared execution log inside the Thread file;
- all-files rollback or claims of transactionality across Markdown documents; and
- adopting a graph package for speculative algorithms outside the V1 contract.

## Consequences

### Positive

- The planning repository has one dependency answer for every task, even when tasks are reused
  across Threads or executed off-Thread.
- Threads add an initiative goal, lifecycle, and rollup without changing the one-epic taxonomic
  ownership rule.
- Frontier, blockers, downstream impact, and explanatory sequencing become deterministic machine
  projections instead of prose conventions.
- External gates stay visible without inflating Thread progress, and one planning space can
  coordinate tasks implemented across multiple repositories.
- Bulk compose/apply supports known-ahead project scoping without sacrificing stable IDs or safe
  retries.
- Generated views avoid merge-prone cached diagrams and progress logs.

### Negative / Cost

- `Task`, its field registry, persistence, mutation ports, transition paths, lint, and wire contracts
  all change. This is not merely a new `Thread` model.
- A first-class Thread still fans out across domain, store, core, CLI, completion, schema/template,
  wire, init/layout, health, and eventually TUI code.
- Graph-sensitive mutations require a sound repository scan and cycle analysis. The expected
  complexity is linear in tasks plus edges, but performance should be measured rather than assigned
  a speculative sub-millisecond promise.
- Bulk apply is resumable but not transactionally atomic across files. Users and automation must keep
  the materialized plan until its receipt is complete.
- The canonical Thread document centralizes membership edits and can conflict under concurrent
  scoping. Ordinary task execution does not rewrite it, which limits but does not remove that risk.

### Semantic Risks

- A dependency edit is global even when initiated while looking at one Thread. Human and machine
  output must make that blast radius explicit.
- Reopening an upstream task can make completed descendants and completed Threads inconsistent.
  Recursive sound completion is conservative by design; the follow-up repair UX must make removal,
  replacement, or re-completion understandable.
- Hard AND-only prerequisites will not model every real workflow. They are intentionally the smallest
  enforceable contract; pressure for soft or alternative gates should be evaluated from actual use.
- The best graph package may provide useful later algorithms, but package capability must not expand
  V1 scope or leak library types into persisted/public contracts.

## Validation Spike Commentary (2026-08-24)

The bounded vertical prototype under `internal/threadspike` plus its experimental filesystem adapter
under `internal/store/threadspike.go` exercised the required scenarios against temporary Markdown
planning repositories. It is intentionally not wired into `core.Service`, production wire contracts,
ordinary CLI builds, or live planning data. A `-tags threadspike` build exposes a disposable manual
compose/apply/show/plan surface solely to exercise the prototype in throwaway spaces.

**Recommendation: accept and scope implementation**, retaining this ADR's proposed status until the
decider signs off. The central model survived: one task-owned global DAG supports shared Thread
membership, external gates, deterministic frontier/topology, and separate lifecycle/gate state.
Recursive sound completion also behaves mechanically as specified when an upstream task is reopened,
although the human repair UX remains a follow-up to validate through use.

The spike found five contracts that must be explicit in production:

1. **Every interrupted apply prefix must be graph-valid.** Existing-task dependency writes require
   deterministic prefix validation and the Thread lands last. Any future inline tasks need
   topological creation order (or deferred edge attachment) before joining this contract.
2. **Apply-time validation is authoritative.** Compose-time success cannot justify creating a task
   after its epic disappeared or accepting a hand-edited apply plan with invalid creation fields.
3. **Graph snapshots are stricter than ordinary resilient lists.** ID drift, unknown task status,
   malformed files, missing dependencies, and missing Thread members must make graph mutation fail
   closed even when ordinary listing can still return partial data for repair.
4. **The current lock proves the concurrent-cycle contract only on Unix.** Two opposite edge adds
   were serialized correctly in the test harness, with exactly one accepted, but `writeLock` is a
   documented no-op on non-Unix builds. Portable equivalent correctness is prerequisite work, not a
   post-launch polish item.
5. **The authoring manifest and apply plan are different artifacts.** The manifest is user-authored,
   literal YAML which bulk-links existing task IDs with local graph keys. Inline `new_task`
   definitions are not accepted in V1. The apply plan is generated,
   stable-ID-only recovery state. Shell interpolation may be useful in a test fixture, but it is not
   part of either file format and should not obscure the primary YAML-first workflow in product
   documentation.

The package comparison does not justify a V1 dependency. [`dominikbraun/graph`](https://github.com/dominikbraun/graph)
is the closest fit—generic IDs, acyclic traits, stable topological ordering, cycle prevention, DOT,
and no transitive dependencies—but its own documentation says the v0 API is unstable, and taskflow
would still own its required diagnostics and all status/Thread semantics. [Gonum's graph/topo
packages](https://pkg.go.dev/gonum.org/v1/gonum/graph/topo) are mature and actively maintained, but
their generalized node API and broader module add adaptation cost without replacing taskflow's hard
contracts. Keep a small owned adjacency/analysis implementation for V1, subject only to the bounded
contract-test bake-off in Section 5; reconsider a broader package search if real usage asks for
materially richer rendering or algorithms.

The spike's pure graph and compose/apply tests are promotable as contract tests. The experimental
mirror types, private store adapter, and failure-injection hook should be removed or rewritten behind
production domain and consumer-owned persistence ports as the implementation slices land.

## Amendments

### 2026-08-26: Implementation-readiness hardening

The adversarial [implementation-readiness audit](../audits/6g3qxqav7yed-2026-08-25-threads-dag-implementation-readiness.md)
confirmed the central model but found contracts that the validation spike had simplified or left
implicit. The following amendments supersede conflicting wording above:

1. **The store controls the graph-mutation critical section.** A graph mutation port takes the
   repository guard, loads the authoritative strict snapshot, invokes a core-supplied pure planner,
   applies the returned planned writes through internal lock-free primitives, and releases the
   guard. The planner cannot call another Store method. Nested graph/store mutation must fail
   explicitly rather than self-deadlock. Filesystem locking and graph semantics remain on their
   respective sides of this callback boundary.
2. **Snapshot analysis is memoized and linear.** Sound completion, gate state, blockers, downstream
   impact, and Thread projections cache one result per task per immutable snapshot. Reconvergent
   graphs must remain O(V+E), demonstrated by visit-count assertions rather than timing guesses.
3. **First-party generic editing cannot bypass graph or lifecycle policy.** `task set` and
   `--force` always reject `depends_on` and direct users to `task depend add/remove`. `task edit` is
   a product path, not a raw-hand-edit equivalent: a dependency delta must use the same guarded
   graph validation or be rejected, and a status delta must use the same lifecycle/eligibility
   policy. Direct filesystem edits remain possible but are diagnosed by lint and derived views.
4. **Gate derivation has explicit precedence and explanations.** `broken` outranks `blocked`; a
   deferred prerequisite is blocked, while a missing, unreadable, withdrawn, or recursively broken
   prerequisite is broken. Each blocker projection carries a stable reason token and one
   deterministic shortest explanatory path, so an unsound completed prerequisite is distinguishable
   from ordinary unfinished work.
5. **Diagnostic reads and actionable selectors have different failure behavior.** Mutations always
   fail closed. Diagnostic reads such as `thread show`, blockers, and lint degrade with an explicit
   problem list and graph-health marker. Dispatch-oriented selectors such as `thread frontier` and
   `task list --unblocked` include clear-gated `next-up` and `ready-to-start` work, return no eligible
   work when their relevant graph is unsound, and report why; they never silently dispatch from a
   partial graph.
6. **Legacy migration follows the mutation boundary.** The strict-read task diagnoses legacy
   `blocked_by`, `dependencies`, and `blocks`, including exact slug-to-ID resolution, ambiguity, and
   missing-target errors, but does not rewrite them. The guarded dependency-operations task performs
   the additive, cycle-validated migration with body/frontmatter preservation.
7. **Lifecycle receipts disclose graph impact.** Internal overrides are typed by gate even when the
   CLI presents the contextual `--force` spelling. `start` and a move to `in-progress` may force only
   the dependency gate; completion force continues to address unexplained acceptance criteria.
   Transitions report descendants whose gate state changed. Once Threads exist, the same receipt also
   reports affected Thread IDs rather than requiring a second discovery command.
8. **Thread rollup and closure expose nominal versus sound completion.** External gates stay outside
   the progress denominator but do block Thread completion. Views report nominal `done / total` and
   sound `drained / total`, plus exact outstanding external gates. Both `thread start` and
   `thread complete` require at least one non-withdrawn member.
9. **Bulk-linking V1 operates on existing tasks only.** Inline task creation is deferred, so apply
   convergence does not own or compare mutable task content. Compose refuses an ID-less planning
   space with an actionable identity-migration error; path identity is not a substitute for the
   durable planning-space ID. Additive membership and dependency operations treat an already-present
   value as an idempotent skip with a receipt entry.
10. **Identity and output remain deterministic.** Task and Thread IDs share a planning-space
    namespace and are checked for cross-kind collision. Public blocker/impact DTOs use stable task
    IDs, reason/path data, and taskflow-owned error vocabulary; concurrency attribution may enrich an
    error but does not change a cycle from validation into a retryable conflict.

### 2026-08-26: Dependency-foundation adversarial hardening

Two independent implementation audits—[Gemini](../audits/6g417v97bx8s-2026-08-26-canonical-task-dependency-read-foundation.md)
and [Claude](../audits/6g41amrnje2j-2026-08-26-canonical-task-dependency-read-foundation-claude.md)—found
places where the first read foundation was safe in aggregate but its individual APIs and repair paths
were too easy to misread or bypass. The following clarifications supersede conflicting wording above:

1. **Cycle identity is an SCC property.** Validation identifies every member of each non-trivial
   strongly connected component, plus a one-task component with a self-edge. Diagnostics emit one
   deterministic representative edge-following cycle per component and attribute cyclic membership
   to every affected task; they do not promise to enumerate every simple cycle. A self-edge is not
   also reported as an indistinguishable generic cycle on the same task.
2. **Graph health qualifies every projection.** Canonical edges and exactly resolved legacy edges
   are validated as one projected union before a snapshot may be called degraded. A resolvable
   legacy self-edge or cycle is broken, not merely unmigrated. A topological plan may return useful
   partial waves for diagnosis, but its completeness flag is true only for a healthy snapshot.
3. **Every first-party write honors graph ownership.** Generic set, edit, creation, and lint-repair
   paths may neither add, delete, nor normalize canonical or legacy dependency fields. Until guarded
   dependency creation exists, ordinary task creation rejects non-empty graph-owned fields. Malformed
   graph frontmatter fails closed and must be repaired through an explicitly guarded migration or
   deliberate filesystem edit; parser failure is never interpreted as an empty dependency set.
4. **Safe legacy debt is visible without poisoning routine lint.** An exactly resolved legacy
   reference whose projected edge is structurally legal is an advisory in normal human and JSON lint
   output and does not make lint exit non-zero. Missing, ambiguous, self-referential, or cyclic legacy
   projections remain errors. Snapshot health remains degraded and mutation/dispatch remains closed
   until the advisory debt is migrated.
5. **Authorization and explanation are separate contracts.** Lifecycle authorization uses the
   typed derived state (`Eligible` and its gate), never the length of a blocker list. `Eligible`
   admits both queued (`next-up`) and candidate (`ready-to-start`) work when the authoritative graph
   is healthy and the gate is clear; readiness remains metadata rather than an extra authorization
   gate. A gate explanation includes that state, task-local structural problems, and an
   action-oriented blocking frontier. The API also exposes a separately named full causal
   prerequisite projection for forensic queries; neither projection's empty result is itself
   permission to start work.
6. **Blocking projections declare their traversal semantics.** The causal projection may traverse
   through all reachable unsound prerequisites. The action frontier stops at terminal constraints
   such as missing, unreadable, withdrawn, invalid, or cyclic records and otherwise returns the
   deepest current constraints a user can act on. Both use stable reason tokens and deterministic
   shortest paths.
7. **Complexity claims include output cost.** Snapshot state derivation remains O(V+E). A path
   projection is O(V+E plus the size of the paths it returns); implementations keep predecessor
   links during traversal and materialize paths only for emitted results. Lint uses the bounded
   action frontier instead of expanding the full causal closure for every inconsistent task.
8. **Diagnostics preserve source identity.** Unreadable task filenames retain recoverable stable-ID
   identity, invalid dependency tokens remain distinguishable from missing valid IDs, and duplicate
   IDs are attributed to every source path without silently assigning one record's graph defects to
   another. Strict mutation still fails closed when no unique authoritative record exists.

### 2026-08-27: Portable mutation-guard adversarial hardening

Two independent implementation audits—[Gemini](../audits/6g45qfv27vrm-2026-08-27-portable-repository-graph-mutation-guard.md)
and [Claude](../audits/6g45s2rm09pr-2026-08-27-portable-repository-graph-mutation-guard-claude.md)—validated the
control-inverted boundary and cross-process cycle protection but exposed underspecified recovery,
contention, and platform contracts. The following clarifications supersede conflicting wording above:

1. **Planner write order is recovery data.** Stable-ID ordering is deterministic but not generally
   prefix-safe: an edge reversal must durably remove the old edge before adding its reverse. A pure
   core validator preserves the planner-provided order, canonicalizes only semantic sets such as
   `depends_on`, and rejects the complete plan before writing unless every supplied prefix and the
   final state are sound. Planners must therefore emit a deterministic, convergent sequence.
2. **Planner exclusion is scoped to the canonical planning root.** During the callback, every Store
   entry point for that root—including through a second `FS`—fails fast with `ErrConflict`. This
   converts invalid re-entry into an attributable error and prevents a callback from escaping its
   immutable snapshot. Go cannot reliably distinguish the callback goroutine from an unrelated
   caller without threading an explicit execution capability through every port, so an unrelated
   concurrent read or write may receive the same brief conflict. CLI use is unaffected; a future
   TUI/server adapter should treat it as retryable contention or deliberately revise the port.
3. **Only runtime-tested release platforms claim mutation support.** macOS and Linux use the
   canonical-root process mutex plus root-directory `flock`, with real same-process and child-process
   tests over the production path. Windows and other non-Unix source builds fail closed until a
   shared-repository lock identity is selected and exercised in native CI; a per-user cache lock is
   insufficient for cross-user repositories.
4. **Raw-editor detection is best effort, not transactional isolation.** Real apply performs one
   whole-snapshot content check before the first replacement and another content check immediately
   before each target replacement. This preserves an attributable durable prefix when a later raw
   edit is detected and substantially narrows the clobber window. A raw writer can still race the
   final verify-to-rename interval because it does not honor the advisory guard; the operation never
   claims an all-files transaction or isolation from direct filesystem edits.
5. **Core owns plan semantics; store owns persistence mechanics.** Source-health, dependency-set,
   prefix, and final-graph validation are pure core operations usable by command preview code. The
   store owns canonical loading, repository exclusion, frontmatter materialization, exact source
   comparison, immediate per-file CAS, and atomic replacement. Dependency writes stamp
   `updated_at` from the caller-provided clock only when graph-owned fields change.
6. **Graph dry-run is authoritative but not durable.** It takes the same exclusive repository guard
   through snapshot, planning, validation, and materialization so its preview is internally
   consistent. Because it performs no replacement, it does not run the pre-apply CAS and cannot
   promise that a later real invocation sees the same repository.
7. **Prefix validation cost is a bulk-apply gate, not premature V1 machinery.** Rebuilding the full
   graph per changed task is acceptable for direct dependency commands and remains intentionally
   simple. Before the bulk-linking slice is released, benchmark realistic repository/manifest sizes
   and replace it with incremental validation if lock-held latency is operationally significant.

### 2026-08-27: Dependency-operation command and recovery contracts

The focused [dependency-operations readiness audit](../audits/6g4aj7v60syg-2026-08-27-ship-guarded-dependency-mutations-and-graph-queries.md)
found that the guarded substrate was ready but several product contracts remained implicit. The
following clarifications govern the first production dependency commands:

1. **User references resolve inside the authoritative snapshot.** Dependency planners translate
   each task operand to a stable ID from the immutable `TaskGraph`; they never pre-resolve through
   the Store and carry a potentially stale choice into the guard. Resolution preserves Taskflow's
   ordinary task-reference tiers—exact ID or slug, then unique case-insensitive prefix, then unique
   case-insensitive substring—and returns typed missing or ambiguous errors. The matching policy is
   shared or parity-tested so guarded and ordinary commands cannot drift.
2. **Legacy migration is an explicit, repository-wide convergence command.** `task depend migrate`
   resolves every safe legacy occurrence and inherits the global `--dry-run` and `--json` modes. V1
   has no per-task selector. It rejects an unsafe source before writing and emits a deterministic,
   prefix-safe plan. The repository guard serializes the operation, but multiple file replacements
   are not one rollback transaction: an attributable sound prefix may remain after failure, and the
   same command must resume idempotently from that state.
3. **Public query meanings are narrow and named.** `task blockers` defaults to the actionable
   frontier and `--causal` requests the full forensic closure. `task unblocks` reports all transitive
   downstream dependents plus their current derived state; it does not claim that satisfying the
   source alone makes every result eligible. `task list --unblocked` is the first dispatch-oriented
   selector, includes clear-gated queued and candidate tasks, and returns no work with an explicit
   diagnosis on an unsound relevant graph. There is no repository-global `task plan` command in this
   slice; topological waves become public through the later Thread plan projection.
4. **Receipts distinguish convergence from success.** Edge receipts identify canonical endpoint
   IDs and applied versus idempotently skipped operations. Migration receipts expose planned,
   applied, skipped, and remaining work. A failure after a durable prefix carries that prefix in
   typed human and JSON diagnostics rather than collapsing it into a prose-only error. Diagnostic
   query envelopes include graph health and taskflow-owned problems, and all machine results carry
   planning-workspace identity where mutation receipts already require it.
5. **Generic editing remains intentionally narrow.** `task edit` continues to reject canonical or
   legacy dependency deltas and directs the user to `task depend add/remove`; V1 does not reinterpret
   an arbitrary editor-produced graph delta. Query and receipt DTO names remain implementation
   details, subject to the reflected schema and human/JSON parity gates, rather than being frozen in
   this ADR.

### 2026-08-28: Guarded dependency-operation implementation consequences

Shipping the first production dependency commands validated the graph model and planner boundary,
but also made the next lifecycle slice's persistence obligations concrete:

1. **Eligibility is a guarded lifecycle write, not a preflight query.** The store holds the
   canonical-root guard across authoritative graph loading, pure core authorization/impact planning,
   and persistence. Ordinary `Move` after a separate eligibility read is invalid because a
   prerequisite lifecycle or dependency mutation could commit between authorization and write.
2. **Every creation and editing path has an explicit lifecycle contract.** `task new --start`
   authorizes and creates the in-progress record in one guarded operation rather than composing
   create and move. The interactive portion of `task edit` stays outside the repository guard, but
   its ordinary writer may not persist a status delta. `task edit` rejects every status change and
   directs the user to an explicit lifecycle verb. Separating content editing from lifecycle may
   require one additional CLI call, but it keeps dates, gates, overrides, and impact receipts on the
   single guarded lifecycle path. `task new --start` remains the intentional combined operation.
3. **Lifecycle impact lands before Thread impact.** The eligibility slice owns deterministic
   before/after task gate state and descendant task receipts. The later Thread slice augments those
   receipts with affected Thread IDs after Thread documents exist; future data is not an eligibility
   completion dependency.
4. **Thread persistence follows the lifecycle seam.** Although pure graph derivations remain
   independently testable, Thread mutations implement after eligibility settles the first
   non-dependency reuse of the guard and lock-free materialization primitives. A readiness checkpoint
   then decides whether the broad Thread slice remains coherent or should split.
5. **Broken-graph repair remains parallel recovery hardening.** It does not block eligibility or the
   first Thread entity. Dogfooding may promote it when an actual broken repository prevents guarded
   work; otherwise reconsider it before bulk linking expands graph use.

### 2026-08-29: Dependency-eligibility implementation consequences

Two independent implementation audits—[Antigravity](../audits/6g4ta2mabq1y-2026-08-29-dependency-eligibility-enforcement-antigravity.md)
and [Codex](../audits/6g4tf2yr4nb0-2026-08-29-dependency-eligibility-enforcement-codex.md)—closed the lifecycle slice with these reusable constraints:

1. **Lifecycle durability is an explicit outcome.** A store result records whether its atomic write
   committed. A guard-release or cleanup failure after that point returns a typed failure carrying
   the committed receipt and must never enter a generic conflict retry loop. Human, machine, and TUI
   adapters preserve the receipt; the TUI reloads. Thread lifecycle mutations must adopt the same
   outcome distinction rather than treating every returned error as pre-commit failure.
2. **Ordinary creation cannot manufacture lifecycle-owned state.** Generic task creation accepts only
   the product's non-start creation states. Guarded create-and-start is the sole creation path into
   `in-progress`. Thread creation should likewise begin from one explicit initial state; combined
   create-and-start, if ever added, requires a semantic guarded operation rather than an arbitrary
   status on a generic create port.
3. **Cooperating-writer claims require cooperating-writer tests.** Raw-file CAS tests remain valuable,
   but each new guarded Thread capability must also race two real services over distinct store
   instances, prove the second operation waits on the canonical-root guard, and prove it authorizes
   from fresh task/Thread state after release.
4. **Failure receipts remain structured at the batch row.** Typed eligibility failures retain derived
   state, blockers, override eligibility, and remedy per item. Thread member/frontier failures must
   not be flattened into prose where a mixed batch cannot reconstruct attribution.
5. **Per-item batch commits and repeated graph scans remain deliberate V1 tradeoffs.** Batch lifecycle
   commands do not promise rollback. Repository scans stay correctness-first until dogfooding or a
   benchmark demonstrates operational cost; no invocation-local cache is justified yet.
6. **The first Thread slice is split at the materializer boundary.** The readiness checkpoint found
   that first-class document/schema/init fan-out plus guarded creation/read projections is one
   reviewable slice, while mutable membership/lifecycle, committed recovery, cooperating-writer
   races, and affected-Thread task receipts form a second guarded mutation slice. Bulk apply depends
   on both and composes the settled lock-free, creation-only Thread materializer for its final new
   document rather than inventing a third creation path. Existing-Thread mutations use surgical
   frontmatter updates so lifecycle timestamps and unowned text survive.
7. **Withdrawn members do not retain actionable boundary gates.** Deprecated members remain visible
   in the Thread rollup but are excluded from its progress denominator. Their own prerequisites are
   likewise excluded from the external-gate set and cannot make an otherwise drained completed Thread
   inconsistent. External gates are the direct outside prerequisites of non-withdrawn members.

### 2026-08-29: Dogfood correction — readiness is not graph authorization

The first production Thread exposed a semantic error in the accepted design: its next task was
`next-up` with a clear gate, but `thread frontier` returned no work until the task was manually moved
to `ready-to-start`. That made an author-declared handoff signal, rather than upstream graph truth,
control the DAG frontier.

The corrected contract treats `next-up` and `ready-to-start` as equally start-capable pending-work
roles. `Eligible` is true for either role only when the authoritative repository graph is healthy
and the task's gate is clear. `thread frontier` and `task list --unblocked` use that same derived
state. `task start`, generic moves to `in-progress`, and TUI lifecycle actions accept either source
role; `--force` may bypass a blocked gate from either role but still cannot bypass broken graph
evidence or start from in-flight, parked, completed, or withdrawn roles.

This does not delete `ready-to-start`. It remains useful metadata saying a task is sufficiently
scoped for handoff, while `next-up` says it has entered the working queue. Teams may use that extra
signal, but the CLI does not require the intermediate transition before guarded execution. A focused
[corrective task](../tasks/6g5075cga2nt-make-dependency-eligibility-graph-driven-for-queued-and-ready-tasks.md)
must update the shared graph derivation, every guarded start path, Thread projections,
`task list --unblocked`, wire/schema output, documentation, and parity tests before later Thread UX
builds on the narrower behavior.

### 2026-08-30: Guarded Thread mutation semantics

The implementation-readiness checkpoint for the first existing-Thread writer fixes three mutation
contracts before code makes them expensive to change:

1. **Multi-member add/remove is one atomic Thread mutation.** `thread add <thread> <task-id>...`
   and `thread remove <thread> <task-id>...` validate the complete request and Thread mutability
   inside one canonical-root guard. Any invalid member or forbidden lifecycle state rejects the
   whole command without a write. Add-existing and remove-absent intents are successful idempotent
   receipt entries, and one surviving change writes the sorted membership set once.
2. **Task lifecycle receipts report projection impact, not only direct membership.** The guarded
   task transition compares Thread projections before and after the task-graph change and reports
   every Thread whose derived view changes, including shared membership, downstream-member, and
   external-gate effects. Thread documents remain read-only during a task transition.
3. **Cancellation is the unsuccessful terminal outcome.** The earlier provisional
   `abandon`/`abandoned` vocabulary is replaced by `thread cancel` and persisted status
   `cancelled`. `thread complete` remains the successful project outcome and deliberately shares
   the familiar verb with tasks; the entity-qualified CLI keeps the operations unambiguous.
   Deferred member work is parked rather than terminal and therefore blocks Thread completion.
   Cancellation may enter from `unstarted` or `in-progress`, never mutates member tasks, and is not
   reopenable in V1. Because no production lifecycle writer has emitted `abandoned`, the guarded
   mutation slice replaces that provisional domain/wire/schema token directly rather than carrying
   two permanent synonyms; lint must diagnose any hand-authored legacy value with an actionable
   repair.

### 2026-08-30: Guarded Thread mutation implementation consequences

The first production existing-Thread writer validates the design with four concrete constraints:

1. **Thread writes use the repository transaction shape, not entity-local read/modify/write.**
   Membership and lifecycle acquire the canonical-root guard, load the complete strict task graph
   and Thread set, resolve references inside that snapshot, run the pure policy, verify whole-source
   versions, verify the target Thread immediately before replacement, and then perform one atomic
   surgical write. The update materializer changes only `tasks`, `status`, and lifecycle/update
   timestamps; comments, body, unknown fields, and key order remain owned by the author.
2. **Task lifecycle now treats the Thread set as authorization evidence.** It loads and validates
   every Thread under the same guard, compares all Thread projections before and after the task
   transition, and includes the Thread snapshot in its pre-write CAS. A malformed or concurrently
   edited Thread therefore blocks a task lifecycle write rather than allowing a receipt to claim
   incomplete impact attribution. This is an intentional correctness-first global constraint.
3. **Cooperating operations serialize; raw editors remain advisory outsiders.** Independent stores
   prove that dependency mutation, task lifecycle, and Thread mutation wait on the same guard and
   re-authorize from fresh state. Whole-snapshot and immediate-target hashes reject observed raw
   task/Thread changes, but cannot eliminate the final verify-to-rename window for an editor that
   ignores the advisory lock. The product must describe this as detection, not transactional
   exclusion of arbitrary filesystem writers.
4. **No new graph dependency was needed.** Membership/lifecycle implementation reused the existing
   immutable `TaskGraph` projections and deterministic comparisons. This reinforces the earlier
   decision to research a graph package only when a named operation exceeds the owned algorithms,
   not merely because another graph-backed command has shipped.

### 2026-08-30: Existing-task bulk apply V1 contract

The production bulk-link slice narrows the spike into a generated-intent workflow with explicit
recovery semantics:

1. **Both inputs are strict but serve different authors.** The authoring manifest accepts schema 1
   (or omitted schema), rejects unknown fields, uses unique local keys plus exact existing stable
   task IDs, and treats `member: false` as a claim that the node is real transitive upstream graph
   context for a member. Compose rejects disconnected or downstream nonmember context. The
   generated schema-1 plan contains only the durable planning-repository ID, compose date, one
   fully rendered/preallocated unstarted
   Thread, sorted member IDs, and sorted stable-ID edge additions. Its output is mode `0600` and
   no-clobber. Apply accepts a durable path, never stdin.
2. **Identity is live authorization evidence.** Compose refuses an ID-less repository. Production
   `FS` receives a configuration-rediscovery function from the CLI; apply invokes it while holding
   the canonical planning-root guard and verifies that both the physical root and durable ID still
   match the plan. A moved checkout remains valid, while a changed pointer, wrong repository, or
   removed identity fails closed.
3. **The plan is additive intent, not snapshot ownership.** Apply reloads the strict graph and every
   Thread under the guard, unions each intended edge into the dependent task's current canonical
   set, preserves unrelated edits and additional dependencies, and validates every task-file
   prefix. A hand-edited plan with no member tasks is rejected even though ordinary `thread new`
   permits an empty Thread; bulk apply retains compose's narrower at-least-one-member contract.
   Existing edges and an exactly matching planned Thread are skips. A same-ID Thread with different
   metadata, lifecycle state, membership, or body is a conflict, with an explicit diagnosis when
   the original Thread has simply advanced after a successful apply. Canonical slug checks prevent
   a hand-edited plan from turning the generated filename into a path escape.
4. **Receipts describe physical durability without pretending at rollback.** Edge entries are
   ordered by dependency-owner write and followed by the Thread create. A task-file replacement may
   make several edge entries durable together; each is marked `applied` at that shared commit
   boundary. Interruption leaves later entries `pending`; retrying the same plan recomputes current
   additive intent and converges. After every changed dependency prefix, apply re-reads identity,
   the graph, and Threads before declaring completion—even if the planned Thread was already an
   idempotent skip—so a raw edit that undoes a repaired edge remains pending rather than becoming a
   false success. A final-create or guard-cleanup error after all semantic writes reports
   `complete: true` and is never retried automatically. On a real successful apply, `changed`
   reports whether this invocation actually applied an operation; a dry-run uses it to report
   whether work would be required.
5. **The simple validator remains within the V1 budget, while atomic I/O dominates large applies.**
   An adversarial review showed that rebuilding the entire graph for every additive prefix consumed
   roughly 59% of the planning phase even though edge-addition safety is monotone: each prefix is an
   edge-subset of the validated final graph. The shared validator now skips those redundant rebuilds
   only for canonical edge supersets; removals and legacy-field clearing retain full prefix checks.
   On an Apple M5, the complete guarded dry-run at 1,000 tasks and 300 dependency-owner writes now
   takes approximately 0.22 seconds; the real 300-task-file apply plus final Thread create takes
   approximately 3.24 seconds. V1 sets a reference budget of one second for guarded
   read/validate/materialize and five seconds total at that scale. Crossing either budget on a
   supported development-class machine, or dogfooding ordinary manifests that approach the
   300-write case, triggers another profiling pass. The current real-apply time is mostly per-file
   atomic durability, while the independently budgeted planning phase now has substantial headroom;
   replacing the graph algorithm would not remove the dominant total lock hold. Contending taskflow
   writers wait on the same guard and re-authorize from a fresh snapshot after release.

### 2026-08-31: Nonmember graph context is not an external-gate role

Bulk-apply dogfooding and its implementation audit exposed an overloaded term. The manifest needs
to name every endpoint of a proposed dependency edge, including deeper prerequisites outside the
Thread, while the runtime projection deliberately names only the immediate membership boundary.
The corrected contract uses two distinct concepts:

1. **`member: false` declares compose-time nonmember graph context.** It is valid when the task is
   transitively upstream of a declared member in the proposed final repository graph. This retains
   multi-hop manifests such as `context -> boundary -> member`. The role is not written into the
   materialized plan or Thread document; only the dependency edges persist.
2. **An external gate is a direct boundary prerequisite.** Runtime Thread projections derive it
   from the global DAG when a nonmember directly gates a non-withdrawn member. Default `thread
   show`, rollup, completion, and graph views remain bounded to members plus those immediate gates.
3. **Deeper context remains queryable, not silently promoted.** Full causal blocker queries traverse
   beyond the direct gate and can explain the entire prerequisite chain. Deeper context does not
   enter Thread membership, progress denominators, or persisted metadata merely because compose
   referenced it.

This is a vocabulary and contract correction, not a schema or projection expansion. It resolves the
bulk-apply audit contradiction before deterministic graph views consume the shared projection.

## Related

- Supersedes: [0002-adopt-projects](0002-adopt-projects.md).
- ADR format standard: [0001-adopt-adrs](0001-adopt-adrs.md).
- Stable-key ID storage foundation: [0003-stable-key-id-addressed-storage](0003-stable-key-id-addressed-storage.md).
- Validation home: epic [30-threads-and-task-dependency-graphs](../epics/30-threads-and-task-dependency-graphs.md).
- Acceptance-gate spike: [spike-a-vertical-threads-and-global-task-dag-mvp](../tasks/6g3a1wtx4zrr-spike-a-vertical-threads-and-global-task-dag-mvp.md).
- First-class entity roadmap: epic [28-first-class-entities-new-planning-nouns](../epics/28-first-class-entities-new-planning-nouns.md).
- Adjacent readiness work: [task-readiness-state-draft-vs-finalized-in-frontmatter](../tasks/6fbj87001m03-task-readiness-state-draft-vs-finalized-in-frontmatter.md).
- Graph package selection: deferred to the required-operation spike in this ADR.
