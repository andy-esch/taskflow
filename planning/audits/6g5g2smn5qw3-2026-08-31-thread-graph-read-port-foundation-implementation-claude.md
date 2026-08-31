---
schema: 1
id: 6g5g2smn5qw3
bucket: closed
area: thread-graph-read-port-foundation-implementation-claude
date: "2026-08-31"
updated_at: "2026-08-31"
---

# Audit: Thread graph read-port foundation — Claude — 2026-08-31

> Reviewer assignment: Claude. This document is the review brief and the only file the reviewer
> should update.

## Review brief

Perform an independent adversarial implementation and architecture review of the uncommitted work
for task
[`6g5fy1m967ka`](../tasks/6g5fy1m967ka-decouple-thread-graph-reads-from-the-aggregate-planning-store.md)
on branch `feat/thread-graph-read-ports`, against
[ADR-0006](../adrs/0006-adopt-threads-as-task-dags.md) and
[`docs/ARCHITECTURE.md`](../../docs/ARCHITECTURE.md). This is a foundational port-boundary change:
judge whether it genuinely supports CLI, TUI, future web, and split/read-only adapters rather than
merely making the current filesystem implementation easier to test.

Assume the design may be subtly incomplete even though the local suite is green. Look for hidden
aggregate-store coupling, inconsistent snapshots, unsafe mutation reuse, false portability claims,
nil-capability panics, accidental framework leakage, and seams that will harden into liabilities
when deterministic graph views are implemented. Do not reward abstraction or test volume by itself.
Equally, do not manufacture findings: settle a concern when code inspection and an adversarial
reproduction disprove it.

## Review target

The branch is based at `9d857cc`. The implementation is unstaged and includes source, tests,
architecture guidance, ADR amendments, and dogfood planning changes, so inspect
`git status --short`, `git diff HEAD`, and relevant untracked files. In-scope files are:

- `internal/core/service.go`, `service_task.go`, `dependency_operations.go`, `service_thread.go`,
  `service_thread_apply.go`, and their changed tests;
- `internal/core/workspace.go`, `workspace_test.go`, and `internal/workspacestore/fs.go`;
- `docs/ARCHITECTURE.md`, ADR-0006, the completed task `6g5fy1m967ka`, the amended graph-view task
  `6g3q4rv1w9e2`, and the production Thread membership change.

Ignore unrelated concurrent work under `planning/meta/`, `routines/`,
`planning/tasks/6g54ay3njm8y-adopt-branch-protection-ruleset-as-code-.github-ruleset.json.md`, and
`planning/tasks/6g5erdkd5pk4-make-audit-findings-unforgeable-a-finding-writer-command-and-near-miss-lint.md`.
The separate `show-in-flight-work-before-dispatchable-thread-frontier-members` task is a previously
filed CLI UX follow-up, not implementation evidence for this review. The two review-audit documents
are scaffolding, not evidence.

The intended contract is:

- `core.Service` owns a narrow, independently injectable `TaskGraphSource`; ordinary complete
  adapters default it from `Store`, while a graph-only service may use it with a nil aggregate
  store.
- Task blocker/downstream queries, Thread list/show projections, and Thread compose use that narrow
  source. `ThreadStore` remains independently injectable.
- `WorkspaceSource` may carry task-graph and Thread read capabilities separately even when the local
  filesystem implements every capability on one value.
- Guarded dependency, lifecycle, membership, creation, and apply mutations continue loading and
  revalidating their own authoritative snapshots under the repository guard. An injected read
  projection must never authorize a write.
- Complete-store construction and existing CLI behavior remain backward compatible. Missing narrow
  capabilities fail explicitly rather than panicking.
- Core values remain taskflow-owned and framework-neutral. The next graph-view slice must produce a
  neutral core projection consumable by CLI, TUI, and future web; Mermaid/DOT formatters belong in a
  reusable output-adapter package rather than core or a Cobra command.
- No graph view or derived renderer output is persisted.

## Required hostile angles

1. Inventory every `Service` graph read and every `LoadTaskGraph` caller. Prove each read use case
   uses the intended source and each guarded mutation still acquires a fresh store-owned snapshot.
   Look especially for lint, lifecycle impact, Thread create/mutate/apply, and future-call-site traps
   that make the split only partial or easy to bypass.
2. Attack `NewService` option/default semantics: nil aggregate store, explicit override, complete
   store fallback, option ordering, typed-nil interfaces, partial fakes, and a store that implements
   graph reads but not Threads (and vice versa). Identify panics, misleading capability errors, or
   an inability to deliberately override/disable a discovered capability.
3. Challenge whether `TaskGraphSource` is the correct consumer-owned port. Assess its method shape,
   package ownership, returned mutability/aliasing, error and unreadable-file semantics, scan cost,
   and whether future database or remote adapters can implement it without adopting filesystem
   assumptions or unrelated task CRUD.
4. Inspect split-source consistency. `ThreadView` combines a task snapshot and a Thread snapshot
   from separate calls and potentially separate adapters. Determine what consistency is actually
   promised under concurrent cooperating writes, raw edits, database transactions, or remote
   reads. Flag documentation or tests that claim a coherent snapshot where none is guaranteed, and
   recommend the minimum correction if an atomic read capability is already necessary.
5. Stress `WorkspaceSource` and `WorkspaceService.Open`. The workspace still requires a complete
   `Store` and `Layout`; decide whether that contradicts claims about read-only, split, TUI, or
   future served adapters. Check nil fallback behavior, concrete-adapter leakage, identity drift,
   watcher ownership, and whether the test's embedded nil aggregate port hides a production flaw.
6. Trace Thread compose versus apply end to end. Compose may use injected narrow reads to create an
   advisory plan, but apply must bind identity and re-authorize everything under the guarded store.
   Look for any read-source value, resolved reference, task identity, or graph conclusion that is
   trusted across that boundary.
7. Attack errors and machine behavior. Verify missing graph/Thread capabilities fail explicitly and
   consistently, do not panic, and do not become misleading user validation failures. Check whether
   callers can distinguish adapter misconfiguration from repository corruption and whether current
   CLI/wire exit behavior remains stable.
8. Verify backward compatibility at every composition root, not only unit fakes: CLI construction,
   TUI workspace opening/reload, filesystem store, built-in template-only service, tests or tools
   constructing `Service`, and any future-facing workspace interfaces. Look for silently nil ports,
   changed method behavior, extra scans, or new public assumptions.
9. Challenge package boundaries for the next slice. Determine whether the amended graph-view task
   is sufficiently precise to prevent Cobra/Bubble Tea/HTTP, renderer, filesystem, or third-party
   graph types from entering core/wire contracts. Assess whether Mermaid/DOT placement and reuse are
   actionable or merely prose, and whether one neutral projection can express stable ordering,
   health, roles, external gates, hostile labels, and future UI interaction without renderer logic.
10. Assess test quality rather than only coverage. Use targeted mutation probes or temporary local
    reversions where useful, restore them afterward, and report which tests actually fail. Look for
    tests coupled through shared fakes, fallback paths that pass accidentally, missing negative
    cases, typed-nil holes, aliasing, race assumptions, and assertions too weak to prove port
    independence.
11. Reconcile implementation, task acceptance criteria, ADR, architecture docs, and dogfood state.
    Call out any stronger guarantee in prose than the code provides. Verify the completed foundation
    task really clears the graph-view prerequisite and that normal planning/audit lint is clean.
12. Consider likely evolution without designing speculative infrastructure: multiple planning repos,
    read replicas, a long-lived web process, TUI reloads, pagination/large graphs, caching, and a
    later graph library. Identify only decisions that would be expensive to reverse after graph-view
    machine contracts ship; distinguish those from safely deferred enhancements.

Run proportionate validation, including the full tests, focused race tests, vet/static analysis,
planning and audit lint, and `git diff --check`. Record exact commands and outcomes.

## Deliverable

Update this audit in place after the review. Preserve this brief, then add:

- an executive verdict (`ready`, `ready with tracked follow-ups`, or `not ready`);
- the reviewed commit/worktree state and commands run;
- findings grouped by severity, each with a stable code, `**Status:** open`, file/line evidence,
  impact or reproduction, and a concrete recommendation;
- a short acceptance-criteria traceability table; and
- explicitly settled concerns that looked suspicious but were disproved.

Do not modify implementation files, the ADR, tasks, Thread, generated artifacts, or the other
reviewer's audit. Do not create follow-up tasks or pre-resolve findings; the implementation owner
will triage them after both independent reviews arrive.

## Executive verdict

**Ready with tracked follow-ups.**

The change does the hard part correctly. Guarded mutation is genuinely untouched — no store file
changed, every store-side `LoadTaskGraph(s)` still loads its own snapshot under the repository
guard, and no precomputed graph is passed into any planner, so an injected read projection cannot
authorize a write. Backward compatibility holds at the real composition roots, not just in fakes.
And the test suite is unusually honest: seven independent mutation probes on the new seams each
produced a failing test, including the workspace wiring and the nil guard.

Three findings matter. The split is **partial** in a way the acceptance criteria do not catch —
`task list --unblocked` builds a `TaskGraph` from the aggregate store and panics on the very
graph-only service this task exists to enable. The nil-capability guard catches only a nil
*interface*, so a typed-nil or embedded-nil `Store` — precisely the shape the repo's own workspace
fake uses — defaults `taskGraphs` to a non-nil value and panics rather than failing explicitly. And
the split now spans two independently injectable ports with **no stated consistency relationship**,
where skew surfaces to the user as a repository-corruption diagnostic on a healthy system.

None of these is a data-safety defect, and none blocks the graph-view slice from starting. All three
are cheaper to fix now than after graph views turn this projection into a machine contract.

## Reviewed state and commands

Worktree: branch `feat/thread-graph-read-ports`, based at `9d857cc`, unstaged
(`14 files changed, 202 insertions(+), 25 deletions(-)`; 22 `git status --short` lines including
untracked). Machine: Apple M5, darwin 25.6.0. The worktree was restored to exactly this state after
every probe, verified by `shasum -c` over the six touched source files and by re-checking the
diffstat.

| Command | Outcome |
| --- | --- |
| `go build ./...` · `go vet ./...` | clean |
| `go test ./...` | all packages pass |
| `go test -race -count=1 ./internal/core/... ./internal/workspacestore/... ./internal/cli/... ./internal/tui/...` | pass, no race reports |
| `just lint` (golangci-lint) | 0 issues |
| `./bin/tskflwctl lint` · `./bin/tskflwctl audit lint` | both clean |
| `git diff --check` | clean |
| CLI smoke (12 commands, real planning tree) | `thread list/show/frontier`, `task blockers/unblocks`, `task list --unblocked`, `board`, `status`, `lint`, `--json`, `template list`, `schema thread` — all exit 0 |

Mutation probes — each applied, run, then reverted and checksum-verified:

| Probe | Reversion | Tests that failed |
| --- | --- | --- |
| 1 | graph queries back to `s.store` | `TestServiceThreadReadsComposeIndependentGraphAndThreadPorts`, `TestServiceGraphQueriesDoNotRequireThreadSupport` |
| 2 | Thread list+show back to `s.store` | `TestServiceThreadReadsComposeIndependentGraphAndThreadPorts`, `TestWorkspaceService_OpenAssemblesRuntimeAndPreservesSelection` |
| 3 | compose back to `s.store` | `TestServiceComposeThreadApplyRendersDefaultTemplate` |
| 4 | drop `taskGraphs: store` default | 2 core + 5 cli tests incl. `TestGolden_MachineContract` |
| 5 | `Open` drops `WithTaskGraphSource` | `TestWorkspaceService_OpenAssemblesRuntimeAndPreservesSelection` |
| 6 | `Open` drops `WithThreadStore` | `TestWorkspaceService_OpenAssemblesRuntimeAndPreservesSelection` |
| 7 | remove `LoadTaskGraph` nil guard | `TestServiceThreadReadsFailExplicitlyWithoutTaskGraphSource` |

Behavioural probes used a temporary `internal/core/zz_review_probe_test.go`, deleted afterward
(confirmed absent; suite re-run green).

## Findings

### Medium

#### M1. The read split is partial — `task list --unblocked` and `board` still build graph reads on the aggregate store  · **Status:** fixed

**File:** `internal/core/service_task.go:43-49`, `internal/core/board.go:26` | **Component:** core/service
**Effort:** S · **Urgency:** soon

`ListTasks` reads `s.store.ListTasks()` at `:43` and then, when `f.Unblocked`, constructs
`NewTaskGraph(all, problems)` at `:49` and gates every row on `graph.State(t.ID).Eligible`. That is a
graph read use case by definition — and its task read is *literally the `TaskGraphSource` method
signature*. `Board()` has the same shape at `board.go:26`. Neither was routed.

**Reproduction.** With the composition the task's own objective names —
`NewService(nil, WithTaskGraphSource(graphs), WithThreadStore(threads))`:

```
ListTasks(TaskFilter{Unblocked: true}) -> PANIC: invalid memory address or nil pointer dereference
ListTasks(TaskFilter{})                -> PANIC: invalid memory address or nil pointer dereference
Board()                                -> PANIC: invalid memory address or nil pointer dereference
```

The objective is "Remove the aggregate planning `Store` as an accidental prerequisite of read-only
graph use cases", and AC 3 requires that unavailable capabilities "are explicit and do not panic".
For the three call sites that were routed, both hold. For the most-used read paths in the tool, the
aggregate store is still an accidental prerequisite and the failure mode is a panic.

A compounding factor: `Store` structurally satisfies `TaskGraphSource` (identical method set), so
`LoadTaskGraph(s.store)` compiles anywhere in `Service`. Nothing type-level, and no lint rule,
prevents the next read use case from re-coupling exactly the way these two already have. The
boundary is convention-only at the moment the next slice is about to add read surface next to it.

**Recommendation:** Route the task reads in `ListTasks` and `Board` through `s.taskGraphs`. The only
genuine aggregate dependency in `ListTasks` is `s.store.ListEpics()` for `--epic` validation, which
is already conditional on `f.Epic != ""` and can fail explicitly when the store is absent. Then state
in the `TaskGraphSource` doc comment that `LoadTaskGraph(s.store)` is not an acceptable call form
inside `Service`, so the next reviewer has a rule to check against.

**Resolution:** Task listing and Board now read tasks through the injected
TaskGraphSource; only an explicitly requested epic filter consults the aggregate
Store, with an unavailable-capability error when absent. Graph-only tests pin
list, unblocked filtering, and board behavior.

#### M2. Typed-nil and embedded-nil stores default `taskGraphs` to a non-nil value and panic instead of failing explicitly  · **Status:** fixed

**File:** `internal/core/service.go:158`, `internal/core/service_task.go:100-103`, `internal/core/workspace.go:82` | **Component:** core/service
**Effort:** S · **Urgency:** soon

`NewService` sets `taskGraphs: store` unconditionally, and the new guard in `LoadTaskGraph` tests
`source == nil` — which is true only for a **nil interface**. A typed-nil pointer, or a struct value
embedding a nil `Store`, is a non-nil interface that passes every guard and panics on first use.

**Reproduction:**

```
NewService(nilInterface)              -> taskGraphs == nil : true   (explicit error, correct)
NewService((*probeStore)(nil))        -> taskGraphs == nil : false  -> TaskBlockers PANICS
NewService(probeStore{})              -> taskGraphs == nil : false  -> TaskBlockers PANICS
        // probeStore is `struct{ Store }` — an embedded nil interface
```

That last shape is not hypothetical: it is exactly `workspaceCapabilitiesFake struct{ Store }` in
`internal/core/workspace_test.go:22-24`. `WorkspaceService.Open`'s completeness check at
`workspace.go:82` is `source.Store == nil`, which cannot see it. An adapter that returns such a value
with a nil `TaskGraphs` therefore reaches production graph reads and panics, while `Open` reports the
capabilities as complete. The stated contract is "Missing narrow capabilities fail explicitly rather
than panicking."

The current tests do not hide the *wiring* (probes 5 and 6 prove `Open` passes both capabilities),
but the fake's shape means no test exercises the defaulting path with a store that is non-nil and
unusable.

**Recommendation:** Default `taskGraphs` only when `store != nil`, so the nil-interface case stays
exact, and treat an unusable capability as a first-class outcome rather than trusting a nil check —
either by having `Open` verify the capabilities it accepts, or by documenting in `WorkspaceSource`
that adapters must never return a partially-nil capability value and asserting that in the fake.

**Resolution:** Service construction and narrow-port options now recognize
typed-nil capabilities, normalize a typed-nil aggregate store, and return
explicit unavailable errors. Workspace validation rejects typed-nil Store/Layout
values, and its test uses operational separate fakes. A non-nil adapter that
internally delegates through nil remains an adapter contract violation that core
cannot introspect without invoking arbitrary methods.

#### M3. Split reads promise no consistency, and skew is reported as repository corruption  · **Status:** fixed

**File:** `internal/core/service_thread.go:214-219` and `:242-247`; `internal/core/thread_projection.go:120-134` | **Component:** core/thread-projection
**Effort:** M · **Urgency:** soon

`ListThreadViews` and `ShowThread` take a Thread snapshot and a task snapshot in two separate calls
to two independently injectable ports, then join them in `ProjectThread`. A member present in the
Thread snapshot but absent from the task snapshot becomes `ThreadProblemMissingMember` and forces
`ProjectionHealth = GraphBroken`.

**Reproduction** — a Thread store that knows a member the graph replica has not yet observed:

```
err=<nil>
GraphHealth=healthy   ProjectionHealth=broken   Inconsistent=false
PROBLEM missing-thread-member: "thread 6g3q4rtmv4ak references missing task g45ar6exsvvf"
Rollup={Done:0 Total:1 Drained:0 Deprecated:0}

// same skew on a completed Thread:
Inconsistent=true, 3 problems: missing-thread-member,
  completed-thread-unhealthy-evidence, completed-thread-undrained
```

The command succeeds and tells the operator the repository is damaged.

This class existed before the change, but as a narrow local race: both reads came from one
filesystem tree, and reading Threads first meant the task snapshot was always the newer one. The
entire point of this change is that the two capabilities may now be different adapters — a database,
a read replica, a remote service — where skew is structural rather than momentary and replication
lag inverts the incidental ordering property. Nothing in the port doc comments, the ADR amendment,
or `docs/ARCHITECTURE.md` states what relationship an adapter pair must guarantee. The architecture
text describes `TaskGraph` as "an immutable read projection over one repository scan", which is true
of the graph alone and says nothing about the joined view.

Impact today is a misleading diagnostic rather than a wrong exit code — the CLI renders health at
`internal/cli/render/thread.go:81,117` and does not gate on it. The reason to fix it now is that the
next slice makes this same projection a machine graph contract.

**Recommendation:** State the required consistency where the ports are defined: the task snapshot
must be no older than the Thread snapshot, and an adapter pair that cannot guarantee that must
supply a combined atomic read. If that guarantee is not going to be required, then distinguish
"member absent from this snapshot" from "member missing from the repository" in
`ThreadProblemMissingMember`, so lag cannot masquerade as corruption in the graph-view output.

**Resolution:** Thread list/show and compose now read Thread data before tasks,
tests pin that ordering, and the paired-port contract requires the later task
snapshot to be no older than the Thread snapshot. Lagging/split backends must
coordinate a compatible snapshot; these point-in-time reads remain diagnostic
and never authorize mutation.

### Low

#### L1. The narrow port's problem channel is filesystem-shaped, and core parses paths to attribute graph problems  · **Status:** fixed

**File:** `internal/core/service_task.go:97-99`, `internal/core/dependency_graph.go` `taskIdentityFromPath`, `internal/domain/problem.go:5-8` | **Component:** core/dependency-graph
**Effort:** M · **Urgency:** eventually

`TaskGraphSource` returns `[]domain.FileProblem{Path, Message}`, and `newTaskGraph` recovers which
task is unreadable by running `taskIdentityFromPath(problem.Path)` — `filepath.Base`, strip `.md`,
require a `<12-char-id>-<slug>` stem. The port that is supposed to free an adapter from filesystem
assumptions carries one in its error channel.

**Reproduction** — the same unreadable record reported four ways:

```
filesystem-shaped path   health=broken  problemTaskID="6g3q4rtmv4ak"
database row key         health=broken  problemTaskID=""
opaque remote id         health=broken  problemTaskID=""
empty                    health=broken  problemTaskID=""
```

Health is `broken` in every case, so this is fail-closed and mutations still refuse — the loss is
diagnosability. A non-filesystem adapter cannot say *which* record is bad without synthesising a
string that parses as `<id>-<slug>.md`, and `g.hardBroken` / `g.referenceCandidates` never learn the
ID, so dangling-reference resolution against that task degrades. `domain.Task.Path` is load-bearing
too: it breaks ordering ties and names duplicate sources, degrading to
`duplicate stable task id "…" across <unknown path>, <unknown path>` without it.

**Recommendation:** Not urgent — but decide before graph views publish problem and attribution
vocabulary in a machine contract, because that is the point at which the shape becomes external. The
minimal move is a neutral record-problem type carrying an optional task ID directly, with a path as
one presentation form rather than the identity channel.

**Resolution:** Task 6g5gbk5a5bt0 replaced the filesystem-shaped source contract
with TaskGraphRead and TaskGraphLoadProblem. Pathless adapters now provide
stable ID/slug directly; local FileProblem identity parsing occurs only in an
explicit conversion boundary, with existing CLI/wire behavior preserved.

#### L2. No option can disable a capability discovered from the aggregate store  · **Status:** wontfix

**File:** `internal/core/service.go:46-52` | **Component:** core/service
**Effort:** XS · **Urgency:** eventually

`WithTaskGraphSource` ignores a nil argument, so `NewService(fullStore, WithTaskGraphSource(nil))`
leaves `taskGraphs` bound to the store (verified). A composition root cannot deliberately build a
Service that refuses graph reads — to prove a code path never reads the graph, or to construct a
Thread-only service over a complete adapter.

This is consistent with every sibling option and the nil guard is the right default (an accidental
nil must not silently disable production reads), so this is a design note the brief asked for rather
than a defect.

**Recommendation:** If deliberate disabling is ever wanted, add an explicit `WithoutTaskGraphSource()`
rather than relaxing the nil guard.

**Resolution:** Ignoring nil options is the established safe default and
prevents accidental capability removal in production composition. If deliberate
disabling gains a real caller, it will receive an explicit
WithoutTaskGraphSource option rather than overloading nil.

#### L3. Architecture text claims a read-only/split workspace that `WorkspaceService.Open` refuses  · **Status:** fixed

**File:** `docs/ARCHITECTURE.md:148-152` vs `internal/core/workspace.go:82` | **Component:** docs/architecture
**Effort:** XS · **Urgency:** soon

The new `WorkspaceSource` sentence says a "split or read-only adapter can supply Thread projections
without implementing unrelated entity persistence". `Open` returns
`workspace adapter returned incomplete capabilities` whenever `source.Store == nil`, so through the
workspace path — the TUI's only path — the complete aggregate `Store` remains mandatory. The claim
is true of direct `NewService` composition only. The sentence three lines above it ("used when a
primary adapter needs a complete entity service") pulls the opposite way inside the same paragraph.

**Recommendation:** Scope the sentence to `Service` composition, or relax `Open` to require `Store`
only when the narrow capabilities are absent.

**Resolution:** Architecture guidance now states that WorkspaceService
intentionally assembles the complete Store and Layout required by the local TUI;
read-only primary adapters compose Service directly from narrow ports.

#### L4. The graph-view package rule is satisfiable by a package that defeats it  · **Status:** fixed

**File:** `planning/tasks/6g3q4rv1w9e2-generate-deterministic-thread-graph-views.md` scope/AC; ADR-0006 2026-08-31 amendment item 4 | **Component:** planning/next-slice
**Effort:** XS · **Urgency:** soon

The rule is "outside the CLI command package" (task) and "outside the core and CLI command packages"
(ADR). `internal/cli/render` satisfies both readings literally — and it imports
`charm.land/lipgloss/v2` and `charm.land/lipgloss/v2/tree`. Mermaid/DOT formatters placed there
would bind a future web or served adapter to lipgloss, which is the exact outcome the rule exists to
prevent. No neutral output package exists yet to land them in.

Also unstated: where hostile-label escaping lives. Mermaid and DOT escape differently, so escaping
belongs to each formatter; the AC "Titles and metadata are escaped safely" does not say the neutral
projection must carry raw text, which is the constraint that actually keeps the projection
renderer-neutral.

**Recommendation:** Name the destination package in the task rather than describing it by exclusion,
and add an acceptance criterion that the core projection carries unescaped text with escaping owned
by each formatter.

**Resolution:** ADR-0006 and the graph-view task now name pure internal/graphfmt
as the Mermaid/DOT package, prohibit CLI/TUI/HTTP and styling dependencies
there, require raw labels in the core projection, and assign format-specific
escaping to each formatter.

## Acceptance-criteria traceability

| # | Criterion (abbreviated) | Verdict | Evidence |
| --- | --- | --- | --- |
| 1 | `Service` accepts injected `TaskGraphSource`, defaults from non-nil `Store`, production construction unchanged | **Qualified** | Option and default verified; probe 4 proves the default is pinned; 12 CLI commands green. Defaulting is not conditional on `store != nil` — M2. |
| 2 | Thread list/show and blocker/downstream queries work with separate minimal fakes while the aggregate store is nil | Met | `TestServiceThreadReadsComposeIndependentGraphAndThreadPorts`; probes 1 and 2 both fail on reversion. |
| 3 | Compose uses the narrow source plus `ThreadStore`; unavailable-capability errors explicit, no panic | **Qualified** | True for compose (`TestServiceComposeThreadApplyReportsEachMissingReadCapability`, probe 3). Not true service-wide: `ListTasks`/`Board` panic on the same composition — M1; typed-nil stores panic — M2. |
| 4 | A workspace adapter can provide separate graph and Thread ports; `Workspace.Planning` projects without concrete-adapter checks | Met | `workspacestore/fs.go` supplies all four; probes 5 and 6 each fail on reversion. Workspace still requires a complete `Store`, which is a docs mismatch — L3. |
| 5 | Guarded mutation callbacks and ports unchanged; no weakening of under-lock revalidation | Met | `git status internal/store/` empty; all nine store-side `LoadTaskGraph(s)` call sites unchanged; no graph value crosses into any planner. |
| 6 | Architecture and ADR name the reusable projection boundary for graph views, TUI, and future web | **Qualified** | Both documents amended and specific about type neutrality. The package rule is satisfiable by `internal/cli/render` — L4; the workspace sentence overstates — L3. |

## Settled concerns

Chased and disproved by code inspection plus reproduction.

1. **A read projection authorizing a write.** The sharpest risk in a read/write port split. It does
   not happen: no file under `internal/store/` changed, every guarded mutation still calls
   `core.LoadTaskGraph(s)` on its own adapter under the repository guard, and no `*TaskGraph` value
   is passed into `TaskGraphPlanner`, `ThreadCreationPlanner`, `ThreadMutationPlanner`, or
   `ThreadApplyPlanner`. The injected source is unreachable from any write path.
2. **Weak tests that would pass with the seam removed.** Seven mutation probes, seven failures —
   including the two workspace wiring options and the `LoadTaskGraph` nil guard. This is the
   opposite of the "green suite proves nothing" pattern; every new behaviour is pinned by a test
   that fails when it is reverted.
3. **Aliasing through the new port.** `newTaskGraph` stores `cloneTask(task)` and rebuilds
   `DependsOn` with `append([]string(nil), …)`. Probe: mutating the adapter's slice after
   `LoadTaskGraph`, and mutating the value returned by `graph.Task()`, both left the graph
   unchanged. A caching adapter that reuses buffers cannot corrupt graph state.
4. **Backward compatibility at real composition roots, not just fakes.** `internal/cli/root.go:419`
   (`core.NewService(fs)`), the TUI via `workspace.Planning` (`tui/session.go:117`, `tui/tui.go:19`),
   `workspacestore/fs.go`, and `core.NewBuiltinTemplateService()` all behave as before; 12 CLI
   commands including `--json` and the golden machine contract are green.
5. **Extra repository scans.** In production `taskGraphs == store`, so every read performs exactly
   the same number of `ListTasks` calls as before. No path loads the graph twice.
6. **Compose trusting a read conclusion across into apply.** Only the plan document crosses.
   `MutateThreadApply` re-reads identity, the strict graph, and every Thread under the guard and
   re-validates each referenced task. A composition root that paired one repository's ID with
   another's graph source would mint a plan that apply rejects — fails closed, though the message
   would be `planned prerequisite … does not exist` rather than a repository mismatch.
7. **Capability errors becoming misleading user validation failures.** They are unwrapped errors, so
   they exit 1 ("unexpected") and stay distinct from `ErrValidation`'s 11 and `ErrConflict`'s 14. An
   agent can tell adapter misconfiguration from repository corruption by exit code. This matches the
   pre-existing "unavailable from this store" style rather than introducing a new one.
8. **Lint's separate graph construction being a missed call site.** `service.go:348,367` builds its
   graph from `ListTasksWithBodies` because lint needs bodies; routing it through the narrow port
   would force a second full scan. The doc comment on `TaskGraphSource` says exactly this. Correct
   as written.
9. **The dogfood prerequisite not really clearing.** `task blockers 6g3q4rv1w9e2` reports
   `in-flight/clear`, `✔ no blockers`; `thread frontier complete-production-threads` is
   `healthy · projection healthy`. Planning and audit lint are both clean.

## Candidate tasks

- ✅ Route `ListTasks` and `Board` through `TaskGraphSource` — fixed with M1.
- ✅ Reject typed-nil capabilities explicitly — fixed with M2.
- ✅ Define and pin causally compatible paired reads — fixed with M3.
- ✅ Task
  [`6g5gbk5a5bt0`](../tasks/6g5gbk5a5bt0-make-task-graph-load-diagnostics-adapter-neutral.md)
  replaced filesystem-shaped task-load attribution before graph views — fixed L1.
- ⛔ Add a nil option that disables discovered graph reads — intentionally rejected with L2.
- ✅ Correct workspace scope and name `internal/graphfmt` plus the escaping owner — fixed with
  L3/L4.
