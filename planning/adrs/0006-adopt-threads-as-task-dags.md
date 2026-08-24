---
status: proposed
date: "2026-08-21"
deciders: [andy-esch]
tags: [adr, planning-model, graph, workflow, threads]
supersedes:
  - ADR-0002
superseded_by: null
---

# ADR-0006: Adopt Threads as Initiative Views over a Global Task DAG

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
> **Acceptance gate:** This ADR remains proposed until the bounded
> [vertical MVP spike](../tasks/6g3a1wtx4zrr-spike-a-vertical-threads-and-global-task-dag-mvp.md)
> reports whether its core contracts should be accepted, revised and re-tested, or abandoned.

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

Membership does not own dependency truth. Removing a task from a Thread, completing or abandoning
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
- The existing unmodelled `dependencies`, `blocked_by`, and `blocks` fields are legacy vocabulary.
  Implementation must migrate the six current `blocked_by` users or report them with actionable
  lint, then converge on `depends_on` alone. The unused task-level `projects` field is deprecated
  separately: Thread membership belongs only in Thread documents, so it must not be renamed to
  `threads` on tasks.

### 3. Persisted Lifecycle and Derived Graph State

Dependencies do not add a persisted `blocked` status. Task `status` remains the author-declared
lifecycle; graph state is a derived projection that cannot drift from the edges and current task
statuses.

`ready-to-start` means the task itself is adequately scoped and ready to undertake if its external
constraints permit. It does **not** promise that those constraints are currently satisfied.
`eligible` is the derived word for that stronger claim. This avoids adding another persisted status
while preserving room for the separate draft/finalized readiness work already contemplated by the
project.

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
  path contains a missing, unreadable, or withdrawn prerequisite.

Named views are compositions of those fields:

- **Eligible / Frontier:** candidate plus a clear gate. Only eligible Thread members appear in the
  dispatchable frontier.
- **Drained:** nominally complete plus a clear gate; equivalently, soundly completed.
- **Inconsistent:** in-flight or nominally complete with a blocked/broken gate—for example a forced
  start, a completed task whose prerequisites never caught up, or completed downstream work whose
  prerequisite was reopened.

A queued task can therefore be queued-and-blocked without contradiction, while a ready task can be
candidate-and-clear (eligible) or candidate-and-blocked. Parked and withdrawn tasks are never
eligible and never satisfy downstream dependencies.

Every product path that enters `in-progress`—the `task start` verb, generic `task move`,
`task new --start`, and future TUI actions—must call the same core eligibility guard. The transition
refuses an ineligible task by default. `--force` is an explicit escape hatch: it bypasses only the
dependency gate, does not remove dependencies, and returns a receipt that says the transition was
forced and names the outstanding blockers. The task remains derived as inconsistent until its
prerequisites become soundly completed or the graph is corrected.

Hand edits, or completing a force-started task before its prerequisites catch up, can still produce
a nominally completed-but-unsound task. Such a task does not unblock descendants; `lint`, Thread
views, and blocker queries surface the inconsistency. V1 does not invent a separate dependency-waiver
entity: correct the edge when the constraint is no longer real.

### 4. Thread Lifecycle

Thread lifecycle remains explicit: `unstarted -> in-progress -> completed | abandoned`.

- `thread start` requires at least one non-withdrawn member, stamps `started_at`, and changes only
  the Thread.
- Thread progress reuses the existing `TaskRollup` rule: deprecated members are reported but
  excluded from `done / total`, deferred members remain in the denominator, and external gates do
  not count. Graph health is reported beside that nominal rollup rather than folded into it.
- `thread complete` succeeds only when every non-withdrawn member is soundly completed and no
  member path is broken or inconsistent. Deferred members prevent completion. A deprecated member
  cannot remain as an unsatisfied prerequisite of live member work.
- `thread abandon` is terminal and membership-immutable for the initiative and never mutates member
  tasks; they may belong to other Threads or remain useful off-Thread. V1 does not reopen an
  abandoned Thread.
- A completed Thread is not membership-mutable until an explicit `thread reopen`, which returns it
  to `in-progress` and clears its terminal stamp. Task status and task-owned dependencies can still
  change through their global commands; those changes can make a completed Thread inconsistent and
  must report the affected Threads. The exact reopen diagnosis/repair UX deserves focused
  follow-up but is not a blocker to adopting the model.

### 5. Required Graph Projections and Package Boundary

The first implementation needs a small, deterministic graph-analysis contract:

1. Global cycle validation with an attributable cycle/path error.
2. Status-aware Thread frontier, failing closed for defects in the relevant graph.
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

Graph implementation/package selection is a pre-implementation spike, not an ADR commitment.
Evaluate off-the-shelf packages against the operations above, determinism, API stability,
dependency weight, cycle diagnostics, and render support. Keep the selected package behind a
taskflow-owned analysis interface and use plain task IDs/domain values at every boundary.

### 6. Document Layout & Frontmatter Specification

Threads are stored in `planning/threads/<id>-<slug>.md` using flat, ID-addressed storage
([ADR-0003](0003-stable-key-id-addressed-storage.md)).

```yaml
---
schema: 1
id: 6g2b01v9ck2w
status: in-progress      # unstarted | in-progress | completed | abandoned
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

### 7. Bulk Composition is a First-Class Use Case

Scoping often discovers a complete predicted task graph at once. Bulk composition is therefore
part of the first useful Thread release, not an afterthought. An authoring manifest may mix existing
tasks with complete specifications for new tasks:

```yaml
thread:
  title: Unified navigation hub
  description: Consolidate configuration lifecycle into one hub
  goal: Ship unified config CLI and TUI routes
nodes:
  - key: config-model
    new_task:
      title: Define the unified configuration model
      epic: 17-pm-go-cli
      tags: [config, domain]
  - key: existing-cli
    task_id: 6fjangd7kvh2
  - key: legacy-gate
    task_id: 6fjangd7kvh0
    member: false
dependencies:
  - {from: legacy-gate, to: config-model}
  - {from: config-model, to: existing-cli}
```

- Every node has a manifest-local, unique `key`. Exactly one of `task_id` or `new_task` is
  required. A new-task specification uses the same validation and defaults as `task new`; V1 may
  create it in `ready-to-start` or `next-up`, but not directly in an already-executing or terminal
  state.
- A V1 manifest creates exactly one new Thread. Extending an existing Thread remains available
  through the ordinary membership/dependency commands until real use justifies update manifests.
- `member` defaults to true. `member: false` permits an explicitly declared existing external gate;
  output still discovers undeclared external gates from the global graph.
- Dependency arcs point from prerequisite to dependent. They add repository-global `depends_on`
  relations; the manifest does not own those edges and omission never removes an existing edge.
- Local keys are authoring conveniences only. Persisted Thread membership and task dependencies use
  stable IDs.

#### Compile, Then Apply

A one-shot command that generates fresh IDs during each retry is not safe enough for a multi-file
operation. Bulk composition uses two phases:

```bash
tskflwctl thread compose --from thread-plan.yaml --out thread-apply.yaml
tskflwctl thread apply thread-apply.yaml
tskflwctl thread apply thread-apply.yaml --dry-run --json
```

1. **Compose** reads the repository, validates the entire authoring manifest, mints all new Thread
   and task IDs, resolves every local key, validates exact task/epic references and normal
   task-creation invariants, and checks the proposed edge union for cycles. It writes a
   materialized apply plan before any planning entity is mutated. The plan is bound to the
   planning-space identity and records stable IDs, intended creations, additive membership/edge
   changes, and the preconditions used to calculate them.
2. **Apply** revalidates the complete materialized plan, then converges the repository toward it.
   A missing planned entity is created with its preallocated ID; an identical creation or already
   present set addition is skipped; a same-ID/different-identity collision or stale conflicting
   edit stops with `ErrConflict`. Reapplying the same plan cannot mint duplicate tasks.

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
tskflwctl thread start|complete|abandon|reopen <thread>
tskflwctl thread list [--status <status>]
tskflwctl thread show <thread>               # Rollup, blockers, external gates, frontier
tskflwctl thread frontier <thread>           # Machine list of currently eligible members
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
be scoped into the following delivery slices without requiring the deferred features:

1. **Dependency foundation:** add modeled `Task.DependsOn`, replace the legacy dependency field
   registry, migrate the six current `blocked_by` files to stable IDs, add graph loading and
   fail-closed global integrity/lint, and complete the graph-package spike.
2. **Dependency operations and eligibility:** add dependency mutation/query commands, deterministic
   analysis projections, the shared transition guard across CLI/core call paths, and `--force`
   receipts.
3. **Thread entity:** add document/store/core/wire/CLI support, many-valued membership, lifecycle,
   rollup, external gates, frontier, and initialization migration from the unused Projects scaffold.
4. **Bulk composition and generated views:** ship compose/apply manifests with resumable receipts,
   explanatory plans, and Mermaid/DOT rendering. This is part of V1 because known-ahead scoping is a
   primary use case.
5. **Planned TUI follow-up:** after CLI and wire behavior have usage feedback, add a Thread list/tab
   and detail view showing lifecycle, rollup, frontier, member/external distinction, and a readable
   graph. The first TUI slice should consume the same projections and should not introduce direct
   graph editing or a separate readiness calculation.

### 11. Explicitly Out of Scope for This Decision

- critical path, slack, effort weighting, target-date forecasting, and bottleneck scoring;
- autonomous dispatch, multi-agent/worktree orchestration, merge ordering, retries, or barriers;
- transitive reduction or graph simplification commands;
- soft, OR, conditional, time-based, or cross-planning-space dependency edges;
- stored diagrams, stored rollups, or a shared execution log inside the Thread file;
- all-files rollback or claims of transactionality across Markdown documents; and
- choosing a graph package before the required-operation spike.

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

## Amendments

_None yet (proposed)._

## Related

- Supersedes: [0002-adopt-projects](0002-adopt-projects.md).
- ADR format standard: [0001-adopt-adrs](0001-adopt-adrs.md).
- Stable-key ID storage foundation: [0003-stable-key-id-addressed-storage](0003-stable-key-id-addressed-storage.md).
- Validation home: epic [30-threads-and-task-dependency-graphs](../epics/30-threads-and-task-dependency-graphs.md).
- Acceptance-gate spike: [spike-a-vertical-threads-and-global-task-dag-mvp](../tasks/6g3a1wtx4zrr-spike-a-vertical-threads-and-global-task-dag-mvp.md).
- First-class entity roadmap: epic [28-first-class-entities-new-planning-nouns](../epics/28-first-class-entities-new-planning-nouns.md).
- Adjacent readiness work: [task-readiness-state-draft-vs-finalized-in-frontmatter](../tasks/6fbj87001m03-task-readiness-state-draft-vs-finalized-in-frontmatter.md).
- Graph package selection: deferred to the required-operation spike in this ADR.
