---
schema: 1
id: 6g3a1wtx4zrr
status: completed
epic: 30-threads-and-task-dependency-graphs
description: Build a bounded, fixture-backed prototype of the dependency, Thread projection, and resumable bulk-composition contracts; recommend accepting, revising, or abandoning ADR-0006.
effort: 2-4 days
tier: 2
priority: high
autonomy_level: 3
tags: [spike, threads, graph, planning-model, adr]
created: "2026-08-24"
updated_at: "2026-08-25"
started_at: "2026-08-24"
completed_at: "2026-08-24"
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

- [x] A runnable, deterministic prototype or focused test harness demonstrates every required scenario against a temporary filesystem-backed planning repository.
- [x] The graph implementation/package comparison records API fit, determinism, diagnostics, dependency cost, render support, maintenance signals, and what taskflow must still own regardless of package.
- [x] The spike maps the actual production fan-out across domain, field registry/schema, store and mutation guard, core ports/use cases, CLI transitions, wire contracts, initialization/layout, lint, and later TUI work.
- [x] The spike records every ADR assumption as validated, falsified, or still open, with evidence and any proposed replacement contract.
- [x] ADR-0006 receives a concise spike-findings commentary section or proposed amendments, but its status is not changed by this task.
- [x] The final task report recommends exactly one outcome: accept and scope implementation, revise and re-spike named risks, or abandon and preserve the simpler current model.
- [x] If acceptance is recommended, the report proposes implementation slices with dependency order and identifies which prototype code should be promoted versus removed.
- [x] Repository tests, formatting, lint, and diff checks pass for retained code; disposable prototype artifacts are removed or clearly isolated.

## Out of scope

- Shipping the production Thread CLI, wire contract, migration, or TUI.
- Changing ADR-0006 from proposed to accepted.
- Critical path, slack, forecasting, transitive reduction, scheduling, or autonomous agent orchestration.
- Polishing graph visualization beyond enough output to validate the projection contract.
- Treating benchmark guesses as evidence; measure only if the fixture exposes a credible concern.

## Spike report (2026-08-24)

### Outcome

**Recommend exactly: accept and scope implementation.** Keep ADR-0006 proposed until decider sign-off,
but no named risk requires another spike before scoping the production slices. The core model is
implementable and useful; the important corrections are bounded persistence contracts, not a reason
to retreat to flat Projects.

The retained prototype is deliberately isolated in `internal/threadspike`; the adapter in
`internal/store/threadspike.go` reaches the production Markdown parser, frontmatter surgery, atomic
file replacement, content CAS, and repository lock without publishing Thread or dependency fields
through the current domain or wire APIs. Ordinary CLI builds remain unchanged; a
`-tags threadspike` binary adds only the disposable manual surface documented below. All fixtures
use temporary planning repositories.

### Executable evidence

- `TestThreadSpikeFilesystemProjectionCoversLifecycleExternalGatesAndValidation` loads real task and
  Thread Markdown and demonstrates two Threads sharing tasks; a visible, denominator-excluded
  external gate; queued-and-blocked, candidate-and-blocked, candidate-and-clear, force-started
  inconsistent, and soundly completed projections; reopened-upstream inconsistency; and rejection
  of self, missing, duplicate, and cyclic edges with attributable diagnostics.
- `TestThreadSpikeConcurrentOppositeEdgesCannotCommitACycle` starts opposite edge mutations from
  the same repository state. The final scan/validation/write is inside the Unix repo lock: exactly
  one write succeeds, one is rejected with the resulting cycle, and the repository remains a DAG.
- `TestComposeMaterializesExistingAndNewNodesIntoDurablePlan` compiles manifest-local keys into a
  planning-space-bound YAML plan with stable Thread/task IDs, one existing node, two new nodes, an
  external-only gate, and deterministic dependency order. YAML round-trip preserves the retry token.
- `TestThreadSpikeApplyIsDryRunnableAndConvergesAfterInterruption` exercises dry-run, injects failure
  after the first real file write, verifies that prefix is still a readable DAG, then retries the
  same plan to completion. A third apply performs only `already-applied` skips: no duplicate task,
  Thread membership, or edge is created.
- `TestThreadSpikeApplyIsBoundToPlanningSpaceAndFailsClosed` proves repository identity binding and
  stricter graph-snapshot health than ordinary resilient listing.
- `TestThreadSpikeTaggedCLIComposeApplyAndInspect` builds the opt-in command tree and exercises
  compose, dry-run apply, real apply, list/show/plan, idempotent retry, and persisted dependencies
  through the same CLI shape used by the manual playbook.

### Assumption ledger

| ADR assumption | Result | Evidence / replacement contract |
|---|---|---|
| One planning-repository DAG can back many Thread views. | Validated | Shared membership changes no edge ownership; both views derive from the same task snapshot. |
| An outside prerequisite can gate a member without entering Thread progress. | Validated | Typed `external: true` projection reports the direct gate while rollup counts members only. |
| Lifecycle and dependency health should be orthogonal. | Validated | Queued/blocked, candidate/clear, candidate/blocked, in-flight/inconsistent, and completed/drained remain unambiguous without a persisted blocked status. |
| Recursive sound completion reacts coherently to reopen. | Mechanically validated; UX open | Reopening an upstream task makes the completed descendant non-drained/inconsistent and a completed Thread inconsistent. Keep the rule; validate repair language through real use. |
| Per-file CAS plus final graph validation is sufficient. | Falsified | It is sufficient only when the final scan, validation, and write share one repository mutation guard. Separate CAS operations admit write skew. |
| The current repository lock proves that guard everywhere. | Falsified | Unix `flock` passes the concurrent opposite-edge test; the non-Unix implementation is explicitly a no-op. Production needs a portable equivalent or a clearly bounded platform contract. |
| Ordinary resilient task scans are sound graph snapshots. | Falsified | The normal parser deliberately returns ID drift/unknown statuses for diagnosis. Graph mutation must elevate those plus missing Thread members and bad dependency refs into fail-closed problems. |
| Any deterministic order is safe for interrupted bulk creation. | Falsified | Stable-ID order can write a dependent before its new prerequisite and make retry fail closed. New tasks must land in topological waves (or vertices first, edges later); existing-task edges follow; Thread last. |
| Compose-time validation can be trusted by apply. | Falsified | Epics or repository content can change and plans can be edited. Apply must revalidate current graph, exact create identity, all creation invariants, membership, and planning-space identity. |
| A repository snapshot hash/precondition is required for safe retry. | Not supported by evidence | Exact create intent plus additive operations and current-state revalidation safely tolerate unrelated edits. Conflict only when an intended identity changed or the current union is invalid. |
| A two-phase materialized plan prevents duplicate IDs after interruption. | Validated with the prefix-order amendment | IDs are minted once at compose; identical already-landed creations skip, collisions conflict, and retry converges. |
| A graph package could own the domain behavior. | Falsified | A package can supply traversal/topology/render helpers, but taskflow must own statuses, sound completion, external gates, membership, errors, storage, mutation guard, and apply convergence. |
| The graph size needs early optimization or benchmarking. | Still open, non-blocking | The algorithms are linear in vertices plus edges and the fixture exposed no credible concern. Measure only after real planning repositories provide scale evidence. |

### Graph implementation/package comparison

| Candidate | API fit and determinism | Diagnostics | Dependency/render cost | Maintenance signal | What taskflow still owns |
|---|---|---|---|---|---|
| Small owned adjacency model (prototype) | Exact string IDs; deterministic DFS, upstream/downstream traversal, and member-only topological waves; roughly 425 lines including domain projections. | Exact, stable cycle path plus task-specific self/missing/duplicate errors. | No module dependency; Mermaid/DOT are small textual adapters but were not polished in the spike. | Maintained with taskflow; smallest surface, but every algorithm is ours. | All semantic and persistence contracts. |
| [`dominikbraun/graph`](https://github.com/dominikbraun/graph) | Closest library fit: generics, directed/acyclic traits, prevention on edge add, stable topological sort, traversal, and transitive queries. | Useful library errors, but taskflow would still build stable user-facing cycle paths and broken-reference attribution. | Zero dependencies and built-in DOT support are attractive. | Active repository and substantial tests, but the maintainer explicitly says its v0 public API is not stable. | All lifecycle/gate/Thread projections, persistence, guard, receipts, and public error contract. |
| [Gonum `graph/topo`](https://pkg.go.dev/gonum.org/v1/gonum/graph/topo) | Stabilized topological sort, path queries, directed cycles, and SCCs are strong; generalized `int64` node identity needs an adapter for stable string task IDs. | Rich cycle/SCC data, still requiring deterministic task-ID translation and taskflow wording. | Mature but much broader module; DOT support exists elsewhere in Gonum. | Active, scheduled releases and cross-platform testing; still pre-v1. | The same taskflow-specific surface; replaces only generic algorithms. |

**Selection recommendation:** own the small V1 algorithms behind a taskflow analysis interface. Do
not add either dependency yet. A production-foundation task may spend at most a bounded bake-off on
a `dominikbraun/graph` adapter by running the same taskflow contract tests against both
implementations; adopt it only for meaningful code reduction without weaker determinism or
diagnostics. This is not another research spike. Gonum is justified only if analysis expands
substantially.

### Actual production fan-out

| Seam | Current concrete touchpoints | Required production work / risk |
|---|---|---|
| Domain and validation | `domain.Task`, status vocabulary, `ActiveTaskFieldErr`, entity descriptor | Model sorted `DependsOn`; add Thread/status/validation; keep gate state derived. Avoid renaming legacy `projects` to `threads`. |
| Field registry and schema | `domain/fields.go`, task struct/registry sync tests, entity `AuthoringFields`, `schema` render | Replace legacy `dependencies`/`blocked_by`/`blocks` vocabulary with canonical `depends_on`; advertise Thread authoring separately; preserve unknown-field migration behavior. |
| Filesystem store | `FS`, task parse/create/set/edit/move, entity scanners, atomic create/replace, content CAS | Add strict graph snapshot and Thread store. Dependency mutation must be surgical and additive/removal-aware; raw generic `task set` cannot bypass global validation. |
| Mutation guard | `store/lock_unix.go`, no-op `lock_other.go` | Expose a store-owned graph mutation operation/callback so final scan, validation, and write cannot be split across core calls. Close or explicitly bound the non-Unix correctness gap. |
| Core ports/use cases | Large `TaskStore`/`Store`, `Service.Move`, `NewTask` | Prefer narrow consumer-owned Thread/graph ports so every fake does not immediately grow. Centralize eligibility before every transition into `in-progress`; return forced blocker metadata. |
| CLI and transitions | `internal/cli/task.go`, lifecycle aliases, global dry-run/output modes, completion | Add dependency queries/mutations and Thread commands; route generic move, start, TUI move, and future starts through one guard; add compose/apply durable-plan UX. |
| Wire/schema/rendering | `internal/wire` DTO/envelope registries and schema-version goldens; `cli/render` | Add stable DTOs for role/gate/external/blockers/forced receipts/Thread views and bulk receipts; bump/regenerate contracts rather than leaking a graph-library type. |
| Initialization/layout/health | `domain/layout.go`, entity descriptor, `config.Init`, discovery, `FS.WatchPaths` | Scaffold/watch `threads/`, stop creating new `projects/`, and refuse automatic removal of non-empty legacy content. Planning identity—not implementation repo path—binds apply plans. |
| Lint and migration | resilient scans, core lint, frontmatter diagnose/fix, six live `blocked_by` users | Add global duplicate/self/missing/cycle/ID/status/member checks; provide an explicit stable-ID migration or actionable legacy lint before mutation goes live. |
| Later TUI/atlas | watcher-driven reload, task transition actions, per-entity list/detail rendering | Consume the same core projection and eligibility guard. Add Thread views after CLI feedback; do not recompute readiness or edit graphs directly in the TUI. |

### Implementation slices and prototype disposition

1. **Portable graph foundation and migration:** define the production graph snapshot/analysis
   interface and mutation-guard contract; add modeled `depends_on`, strict graph lint/load, and the
   legacy-field migration. Prove Unix plus the supported non-Unix strategy before public edge writes.
2. **Dependency use cases and transition eligibility:** ship guarded add/remove, blocker/unblocks and
   deterministic plan queries; enforce eligibility across every start path with explanatory `--force`
   receipts.
3. **Thread entity:** add descriptor/domain/store/core/CLI/wire coverage, many-valued membership,
   lifecycle/rollup/frontier/external-gate views, and Projects-scaffold migration behavior.
4. **Bulk compose/apply:** first promote existing-task bulk linking, materialized-intent validation,
   planning-space binding, idempotent retries, conflicts, dry-run, and machine receipts. Keep inline
   task creation and its topological prefix ordering optional until the simpler workflow has usage.
5. **Generated views, then TUI:** add Mermaid/DOT from the shared projection; only after usage feedback,
   add read-oriented Thread list/detail/frontier views to the TUI.

Promote the pure graph contracts and tests, the materialized-plan shape, apply-time validation rules,
topological prefix ordering, and the interruption/concurrency tests. Rewrite the experimental types
against production domain DTOs and narrow ports. Remove the private store adapter and `AfterWrite`
failure hook after their behavior is covered at production seams; they are spike scaffolding, not a
second persistence architecture.

### Prototype retention decision

The current build tag hides only the experimental CLI command. `internal/threadspike` and
`internal/store/threadspike.go` are untagged, so they compile—and their contract tests run—in an
ordinary build. They are internally scoped and unreachable from the normal command tree, but this is
not full build isolation and would create a second model/store path that can drift if retained
indefinitely.

Default integration disposition:

- retain and merge the ADR, epic, and this completed spike report;
- preserve the prototype on its spike branch as executable evidence;
- port the graph, concurrency, and interruption scenarios into production contract tests as their
  owning slices land; and
- rewrite useful algorithms behind production domain types and narrow ports rather than promoting
  the experimental package wholesale.

If a short-lived tagged binary must be retained in the main branch for further disposable-space
evaluation, first apply `//go:build threadspike` consistently to the graph package, store adapter,
and their tests, and give that scaffolding an explicit removal milestone. Even then, a tagged binary
can mutate any selected planning root, so it must not be used on canonical planning data.

### Validation

- `go test -race ./...` — passed, including the temporary-filesystem, interruption, and Unix
  concurrency harnesses.
- `golangci-lint run ./...` — passed with 0 issues.
- `go mod tidy -diff` — passed with no module-file changes; the spike added no graph dependency.
- `go run ./cmd/tskflwctl -C . lint` — all active tasks and epics pass planning lint.
- `git diff --check` — passed.
- `go test -tags threadspike ./...` — the opt-in CLI variant and its end-to-end playbook test pass;
  the untagged full suite separately proves ordinary builds remain unchanged.

### Manual throwaway-space playbook

The spike now provides an explicitly experimental `thread` command only when `tskflwctl` is built
with the `threadspike` tag. Ordinary builds remain unchanged. This surface is for disposable data;
it is not a production CLI or wire contract.

Build one reusable binary and initialize a throwaway space:

```bash
cd ../taskflow-threads-spike
go build -tags threadspike -o /tmp/tskflwctl-threads ./cmd/tskflwctl

export TSK=/tmp/tskflwctl-threads
export PLAY=/tmp/taskflow-threads-play
mkdir -p "$PLAY"
"$TSK" init --path "$PLAY" --no-register
```

Create two epics and five tasks (`jq` captures the stable IDs):

```bash
CORE=$("$TSK" -C "$PLAY" epic new "Core delivery" --description "Build the core in dependency order" --tags threads --json | jq -r '.created.id')
DOCS=$("$TSK" -C "$PLAY" epic new "Documentation" --description "Explain and validate the result" --tags threads --json | jq -r '.created.id')

GATE=$("$TSK" -C "$PLAY" task new "External decision" --epic "$CORE" --tags threads --json | jq -r '.created.id')
SCHEMA=$("$TSK" -C "$PLAY" task new "Define schema" --epic "$CORE" --tags threads --json | jq -r '.created.id')
STORE=$("$TSK" -C "$PLAY" task new "Build store" --epic "$CORE" --tags threads --json | jq -r '.created.id')
CLI=$("$TSK" -C "$PLAY" task new "Expose CLI" --epic "$CORE" --tags threads --json | jq -r '.created.id')
GUIDE=$("$TSK" -C "$PLAY" task new "Write guide" --epic "$DOCS" --tags threads --json | jq -r '.created.id')
```

The variables above belong to this shell walkthrough, not to the Thread manifest format. The
unquoted heredocs below substitute them while writing literal task IDs into each YAML file. A saved
manifest used directly with production `thread compose` must contain its actual existing-task IDs.
The spike explored inline `new_task`; production V1 deliberately rejects it until creation
provenance and recovery ordering receive a separate amendment.

Compose and apply a linear Thread with an external gate:

```bash
cat > "$PLAY/core.thread.yaml" <<YAML
thread:
  title: Core delivery
  description: Exercise a linear Thread with one external gate
  goal: Move schema, store, and CLI through a visible dependency chain.
  tags: [threads, spike]
nodes:
  - {key: gate, task_id: $GATE, member: false}
  - {key: schema, task_id: $SCHEMA}
  - {key: store, task_id: $STORE}
  - {key: cli, task_id: $CLI}
dependencies:
  - {from: gate, to: schema}
  - {from: schema, to: store}
  - {from: store, to: cli}
YAML

"$TSK" -C "$PLAY" thread compose --from "$PLAY/core.thread.yaml" --out "$PLAY/core.apply.yaml"
"$TSK" -C "$PLAY" thread apply "$PLAY/core.apply.yaml" --dry-run
"$TSK" -C "$PLAY" thread apply "$PLAY/core.apply.yaml"
```

Add a second Thread sharing tasks but using a fan-in shape:

```bash
cat > "$PLAY/docs.thread.yaml" <<YAML
thread:
  title: Cross-cutting guide
  description: Reuse core tasks in a second Thread
  goal: Show shared membership, external gates, and a two-input documentation task.
  tags: [threads, spike]
nodes:
  - {key: schema, task_id: $SCHEMA}
  - {key: cli, task_id: $CLI}
  - {key: guide, task_id: $GUIDE}
dependencies:
  - {from: schema, to: guide}
  - {from: cli, to: guide}
YAML

"$TSK" -C "$PLAY" thread compose --from "$PLAY/docs.thread.yaml" --out "$PLAY/docs.apply.yaml"
"$TSK" -C "$PLAY" thread apply "$PLAY/docs.apply.yaml"
```

Inspect the projections and then satisfy the first external gate:

```bash
"$TSK" -C "$PLAY" thread list
"$TSK" -C "$PLAY" thread show core-delivery
"$TSK" -C "$PLAY" thread plan core-delivery
"$TSK" -C "$PLAY" thread show cross-cutting-guide

"$TSK" -C "$PLAY" task complete "$GATE" --force
"$TSK" -C "$PLAY" thread show core-delivery

# A second apply is intentionally idempotent.
"$TSK" -C "$PLAY" thread apply "$PLAY/core.apply.yaml"
```

The tagged surface exercises persistence and projections only. Existing task lifecycle commands do
not yet enforce dependency eligibility; use `thread show` to observe the frontier and inconsistency
rules rather than treating this binary as the production transition guard.

## Related

- Epic [30-threads-and-task-dependency-graphs](../epics/30-threads-and-task-dependency-graphs.md)
- [ADR-0006](../adrs/0006-adopt-threads-as-task-dags.md)
- [First-class entities epic](../epics/28-first-class-entities-new-planning-nouns.md)
- [Task readiness state](6fbj87001m03-task-readiness-state-draft-vs-finalized-in-frontmatter.md).
