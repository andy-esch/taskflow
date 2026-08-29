---
schema: 1
id: 6g3qxqav7yed
bucket: closed
area: threads-dag-implementation-readiness
date: "2026-08-25"
updated_at: "2026-08-29"
---
# Audit: Threads / task-DAG implementation readiness — 2026-08-25

Bounded adversarial review of the accepted ADR-0006 design, epic 30, and the eight
production tasks, stress-tested against the current production codebase (not the spike)
before implementation begins. Read the ADR, epic, and tasks first; formed an independent
assessment; then inspected the completed spike report and the archived
`spike/threads-dag-mcp` branch as evidence.

**Verdict: revise before implementation.** The central model survives scrutiny — one
task-owned global DAG with many Thread views is coherent, the storage shape fits ADR-0003,
and declining a graph dependency for V1 is well supported. But four contracts must be
settled before task `6g3q4rst78qy` freezes its analysis interface, because each one changes
the shape of that interface rather than its implementation: the mutation guard is not
reentrant (H1), the prototyped sound-completion algorithm is exponential (H2), two
first-class product paths bypass every gate the design assumes (H3), and the legacy
migration is a graph write scheduled into the slice that forbids graph writes (M3). None of
these is a reason to stop or to retreat to flat Projects.

Findings are classified in each entry:
**[before-task-1]** · **[before-slice-N]** · **[monitor]** · **[deferred]**.

## Findings

### High

#### H1. The repository mutation guard is not reentrant, so "one critical section" cannot wrap the existing store writes  · **Status:** tracked by 6g3q4rt0wzkq

**File:** internal/store/lock_unix.go:21 | **Component:** store / mutation guard
**Effort:** L · **Urgency:** acute

**[before-task-1 — it decides the analysis-interface shape; must land in slice 2]**

ADR-0006 §2 and §7 require the final repository scan, graph validation, and write to occur
inside one planning-repository mutation guard. `FS.writeLock()` opens the repo root fresh
and takes `flock(LOCK_EX)` on that new file descriptor. Every store write already calls it
individually — `fsstore.go:236` (move), `fsstore.go:308` (set fields), `edit.go:158`,
`edit.go:184`, `epicstore.go:105/164/201`, `body.go:98/134`, `rename.go:110`,
`auditstore.go:123`, `researchstore.go:152/188/226`, `fix.go:30`.

flock is per open file description, so a second acquisition inside the same process blocks
on the first. Measured directly with a faithful reproduction of `writeLock`: an outer guard
plus a nested acquisition blocked for the full 2s timeout — self-deadlock, not reentry.
Any implementation that takes a repo guard and then calls `store.SetFields` /
`store.CreateTask` per file will hang on the first write.

The spike did not hit this because it side-stepped the public store entirely:
`internal/store/threadspike.go:177` takes the lock once and then calls **unexported**
helpers — `r.fs.writeNewFile` (line 256), `writeFileAtomic` + `verifyUnchanged` (lines
275-278). That works only because the whole operation lives inside package `store`. This is
the clearest case of the spike validating a contract by simplifying away a production
concern: it proved "one critical section is possible", not "one critical section is
possible from where the ADR puts the logic".

The design is now pinched between two of its own constraints. ADR §9 says dependency
analysis belongs in core and "filesystem code … does not decide frontier, sound completion,
or topological semantics"; task `6g3q4rt0wzkq` says to expose the critical section "without
leaking filesystem locking into core services". Satisfying both means core logic must run
*inside* a lock it does not hold — an inversion nobody has written down.

**Recommendation:** Adopt control inversion explicitly and record it as an ADR amendment.
The store exposes one operation shaped like
`WithGraphMutation(func(GraphSnapshot) ([]PlannedWrite, error)) (Receipt, error)`: it takes
the guard, produces the strict snapshot, calls the core-supplied pure planner, applies the
returned writes through internal lock-free helpers, and releases. Add the rule the spike
relies on implicitly but never states: **no `Store` method may be called from inside the
callback.** As defence in depth, make `writeLock` depth-counted behind a process-local
mutex so a nested acquisition is a loud programming error rather than a hang. Fold the
resulting API sketch into task `6g3q4rt0wzkq` before task `6g3q4rst78qy` freezes the
analysis interface it must plug into.

#### H2. Prototyped sound completion is exponential; the ADR and spike both claim linear  · **Status:** tracked by 6g3q4rst78qy

**File:** internal/threadspike/graph.go:191 (spike/threads-dag-mcp) | **Component:** graph analysis
**Effort:** S · **Urgency:** acute

**[before-task-1]**

`Gate` calls `soundlyCompleted(prerequisite, map[string]bool{})` with a **fresh** visiting
map per prerequisite (line 198), and `soundlyCompleted` (lines 212-233) memoizes nothing —
`visiting` is only a cycle guard, unwound by `defer delete` on return. Every reconvergent
path is therefore re-walked in full. `Blockers` (line 382) does the same, once per node in
the traversal.

Measured with a faithful port of the function against a plain diamond chain (two parallel
tasks per stage converging on the next — an entirely ordinary planning shape):

| depth | vertices | edges | `soundlyCompleted` calls |
|---|---|---|---|
| 4 | 13 | 16 | 61 |
| 8 | 25 | 32 | 1,021 |
| 12 | 37 | 48 | 16,381 |
| 16 | 49 | 64 | 262,141 |
| 20 | 61 | 80 | 4,194,301 |
| 24 | 73 | 96 | 67,108,861 |

73 vertices is a small planning repository; this one already costs 67 million calls. The
ADR's Consequences section says "The expected complexity is linear in tasks plus edges, but
performance should be measured rather than assigned a speculative sub-millisecond promise" —
the caution was right and the estimate was wrong. The spike's assumption ledger goes
further and asserts as fact that "The algorithms are linear in vertices plus edges and the
fixture exposed no credible concern"; the fixture was simply too small and too shallow to
reveal it.

Task `6g3q4rst78qy`'s stress tests name "deep chains" and "wide frontiers" as separate
cases. Neither catches this. The killer shape is **reconvergent** — deep *and* wide at once.

**Recommendation:** Require memoized sound-completion/gate evaluation (one result per task
id per snapshot, giving O(V+E)) in `6g3q4rst78qy`'s scope, and add an explicit
reconvergent-diamond case to its stress tests with an assertion on visit count, not just
wall time. This is a one-line fix if it lands with the contract and a rewrite if it lands
after four consumers depend on the interface.

#### H3. `task edit` and `task set` are supported, tested product paths that bypass every gate the design assumes  · **Status:** tracked by 6g3q4rte8kc1

**File:** internal/store/edit_test.go:205; internal/core/service_task.go:235 | **Component:** cli / core / lifecycle
**Effort:** M · **Urgency:** acute

**[before-task-1 for the `depends_on` half; before-slice-3 for the status half]**

ADR §3 enumerates the paths into `in-progress` that must share one eligibility guard: "the
`task start` verb, generic `task move`, `task new --start`, and future TUI actions". Two
existing first-party paths are missing from that list, and both are *by design*, with tests
pinning the behaviour:

1. **`task edit` writes any status.** `FS.EditTask` (internal/store/edit.go:144) accepts
   edited content on `parseTask` + entity-id validity alone. `TestEditTask_StatusEdit_Writes`
   (internal/store/edit_test.go:205-220) exists specifically to pin that editing
   `status:` through the editor is "an ordinary write". It never reaches `Service.Move`, so
   it can never reach a core eligibility guard. The TUI's open-in-editor action inherits
   this.
2. **`task set` writes list fields with no referential or structural validation.**
   `Service.SetFields` (internal/core/service_task.go:235-292) special-cases only `status`,
   `updated_at`, and `epic` (existence-checked). Everything else is coerced by type and
   written. Verified in a throwaway space:
   `task set alpha-task --set blocked_by=beta-task,6fjangd7kvh0` succeeded and wrote
   `blocked_by: [beta-task, 6fjangd7kvh0]` — a slug and a nonexistent id, unvalidated. The
   moment task `6g3q4rst78qy` adds `depends_on` to `domain.taskFields` (which its scope
   requires), `task set --set depends_on=…` becomes a legal, unguarded, un-cycle-checked
   graph write. `--set … --force` already bypasses the registry entirely: it wrote
   `depends_on: 6g3qkmrcvjpj` as a bare scalar.

The spike's own fan-out table anticipated half of this ("raw generic `task set` cannot
bypass global validation") but it was never carried into the ADR's frozen §3 or into any
task's acceptance criteria. As written, `6g3q4rte8kc1` AC 1 — "Every current path into
`in-progress` calls one policy and cannot bypass dependency checks accidentally" — is not
achievable.

**Recommendation:** Two small decisions, recorded as an ADR amendment to §3:
(a) reject `depends_on` in `Service.SetFields` the way `status` is rejected today, with a
message pointing at `task depend add/remove` — and make `--force` not an escape from it,
since `--force` exists for unknown custom fields, not for corrupting the graph;
(b) declare `task edit` either a product path (re-run the eligibility guard on an accepted
edit whose status changed, and re-run graph validation on an accepted edit whose
`depends_on` changed) or an explicit hand-edit equivalent that only `lint` diagnoses. Pick
one and say so; today it is silently the second while the ADR reads as if it were the
first. Restate AC 1 of `6g3q4rte8kc1` against whichever list results.

#### H4. A bulk apply cannot converge after any ordinary edit to a task it created  · **Status:** tracked by 6g3q4rtv8d0a

**File:** internal/threadspike/manifest.go:398 (spike/threads-dag-mcp) | **Component:** bulk compose/apply
**Effort:** M · **Urgency:** soon

**[before-slice-5]**

`PrepareApply` treats a planned task that already exists as already-applied only when
`equivalentTask` holds (manifest.go:398-404). `equivalentTask` (lines 533-540) is a full
`reflect.DeepEqual` over the entire task record — status, epic, description, tier,
priority, autonomy, effort, created, `updated_at`, `started_at`, `revisit_at`, tags — plus
`depends_on` **and the whole body**. Anything else is `ErrConflict: planned task id X
already exists with different content`, with no reconcile or adopt path.

Concrete failure: apply creates new task T, then is interrupted before the Thread document
lands (the ADR mandates Thread-last, so this is the *expected* interruption window). T is
now visible in `task list`. The user starts it, or fixes a typo in its description, or any
tool path stamps `updated_at`. Retrying the same plan now fails permanently. The
materialized plan is the documented recovery state — ADR §7 says "safe to resume with the
same plan" and the Consequences section tells users to "keep the materialized plan until
its receipt is complete" — but there is nothing left it can do, and the Thread must be
rebuilt by hand.

This is stricter than the ADR itself asks for. §7 says apply conflicts on "a same-ID/
different-identity collision or stale conflicting edit". A changed `updated_at` is neither.
Task `6g3q4rtv8d0a` AC 3 ("retrying the same plan converges without duplicates") is
contradicted by that same task's own stress test list, which requires "concurrent direct
dependency mutation" to be exercised.

**Recommendation:** Define the already-applied predicate as **identity**, not content:
same id, same slug/title, same kind. Membership and edges stay set-additive (they already
are). Everything else is the user's to change after creation. Write that predicate into
`6g3q4rtv8d0a`'s acceptance criteria explicitly, and add a test that starts, edits, and
re-tags a created task between two applies of the same plan and still converges.

### Medium

#### M1. Thread progress and `thread complete` disagree, so a Thread can read 100% and refuse to close  · **Status:** tracked by 6g3q4rtmv4ak

**File:** internal/threadspike/graph.go:275 (spike/threads-dag-mcp) | **Component:** thread projection
**Effort:** S · **Urgency:** soon

**[before-slice-4]**

`ViewThread` computes `Done` from `taskView.Status == domain.StatusCompleted` (line 290) —
*nominal* completion — over members only; external gates never enter `Total` (lines
303-317), exactly as ADR §1 requires. But completion and inconsistency are judged on
`Drained` (line 320), and `Drained` is `soundlyCompleted`, which recurses through the
**global** graph — including external prerequisites and their prerequisites.

So: a Thread whose five members are all `completed`, one of which depends on an unfinished
task outside the Thread, renders `5/5` and then refuses `thread complete`. ADR §4 states
the rollup rule and the completion rule in adjacent bullets without noticing that one
excludes external gates and the other includes them transitively. Task `6g3q4rtmv4ak` AC 2
says an outside prerequisite "appears as an external gate without entering progress totals"
and is silent on completion.

The same asymmetry makes `Done` counted from a status that may be unsound: a member whose
prerequisite was reopened still counts toward `Done`.

**Recommendation:** Amend §4 to state both halves explicitly — external gates are excluded
from the denominator *and* required for completion — and require the Thread view to carry a
separate `drained` count beside `done` so the human output can say "5/5 complete · 4/5
drained · 1 external gate outstanding" rather than contradicting itself. Cheapest correct
alternative if that reads badly in practice: keep `done` nominal, and make `thread complete`
report the exact outstanding external gate rather than a bare refusal.

#### M2. `clear | blocked | broken` has no precedence rule, no home for `deferred`, and no word for "blocked by a completed task"  · **Status:** tracked by 6g3q4rst78qy

**File:** planning/adrs/0006-adopt-threads-as-task-dags.md:187 | **Component:** eligibility policy
**Effort:** S · **Urgency:** soon

**[before-slice-3]**

Three gaps in the frozen §3 definitions:

1. **Overlap.** `blocked` is "at least one prerequisite is not soundly completed";
   `broken` is "an upstream path contains a missing, unreadable, or withdrawn prerequisite".
   A task with one `next-up` prerequisite and one `deprecated` prerequisite satisfies both.
   The spike resolves it broken-first (graph.go:199-201) but nothing says that is the
   contract.
2. **`deferred` is unplaced.** §3 says parked tasks "never satisfy downstream dependencies",
   and `broken` names only missing/unreadable/withdrawn — so a deferred prerequisite is
   `blocked`. That is defensible, but a task deferred with no `revisit_at` is indefinite
   and materially closer to broken. State the choice.
3. **No vocabulary for the confusing case.** A prerequisite that is itself `completed` but
   unsound leaves the dependent `blocked` — blocked by a task the user can see is done.
   `Blockers` returns it with no reason attached (graph.go:373-391). This is the single most
   confusing output the design can produce and it has no name.

`6g3q4rte8kc1` AC 5 says "Deferred and deprecated prerequisites follow ADR-0006 semantics
consistently", which defers to prose that does not settle it.

**Recommendation:** Amend §3 with: broken outranks blocked; a deferred prerequisite is
blocked (and a Thread view may flag an indefinitely-deferred gate separately); and every
blocker entry carries a reason token (`not-started` · `in-flight` · `unsound-completed` ·
`withdrawn` · `missing` · `parked`). The reason token costs one field and removes the whole
class of "why is this blocked by something that's done" support questions.

#### M3. The legacy migration is a multi-file graph write scheduled inside the slice that forbids graph writes — and the live values are slugs, not IDs  · **Status:** tracked by 6g3q4rt7mgjn

**File:** planning/tasks/6g3q4rst78qy-establish-canonical-task-dependencies-and-strict-graph-reads.md:26 | **Component:** migration / sequencing
**Effort:** M · **Urgency:** acute

**[before-task-1]**

Task `6g3q4rst78qy` scope line 26 says "Migrate or actionably diagnose the live legacy
`blocked_by`, `dependencies`, and `blocks` vocabulary", and AC 6 says "No public dependency
or Thread mutation is introduced by this task". ADR §10 slice 1 asks for both at once
("replace and migrate the legacy dependency vocabulary … Do not expose graph writes"). A
migration that rewrites `blocked_by` into `depends_on` across several files *is* a graph
write, and by the ADR's own §2 rule it should happen inside the mutation guard — which is
task `6g3q4rt0wzkq`, a sibling with no declared ordering relationship to task 1.

Separately, the migration is harder than "rename the key". The six live values are **slugs**:

- `planning/tasks/6fgq1n0006y3-theme-config-table-and-selection-plumbing.md:12`
  → `blocked_by: [route-tui-chrome-through-the-palette, route-progress-bars-and-the-cli-ansi-map-through-the-palette, route-the-interactive-picker-theme-through-the-palette]`
- four more single-slug values in the `6fgq1n*` theme tasks and `6ffr4wc01gtv`.

ADR §2 requires `depends_on` to be "stable task IDs, never slugs", so the migration is a
slug→ID resolution pass that can hit ErrNotFound (a renamed task) or ErrAmbiguous
(duplicate slugs are legal under the flat layout). Neither the ADR nor the task's ACs
mention resolution at all; AC 4 only asks for "body/frontmatter preservation".

Verified state: all six referents resolve, all seven involved tasks are `completed`, and
the resulting union is acyclic — so the migration is tractable *today*. That is exactly why
it should be pinned by fixture now.

**Recommendation:** Split the slice explicitly. Task 1 ships **diagnosis only**: a lint
issue naming each legacy field, its resolved target id (or the resolution failure), and the
`--fix` that is coming. The rewrite ships in task `6g3q4rt7mgjn` as a guarded mutation
reusing the same critical section. Add an AC to task 1 covering unresolvable and ambiguous
slugs. Amend ADR §10 slice 1 to say "diagnose" rather than "migrate".

#### M4. Apply plans bind to a planning-space identity that is optional and frequently absent  · **Status:** tracked by 6g3q4rtv8d0a

**File:** internal/config/config.go:78 | **Component:** bulk apply / config
**Effort:** S · **Urgency:** soon

**[before-slice-5]**

ADR §7 states the materialized plan "is bound to the planning-space identity", and task
`6g3q4rtv8d0a` AC 2 requires apply to revalidate it. But the identity is optional by
design: `configFile.ID` is documented as "Empty in a pointer config and in any repo
predating the mint — absence is legal and silent" (internal/config/config.go:78-81), and
`Config.ID` repeats "Empty for a repo that predates the mint — never an error, and never
inferred".

The spike passes it straight through — `store.NewThreadSpikeRepository(app.Cfg.Root,
app.Cfg.ID)` (internal/cli/threadspike_enabled.go:46) — and `PrepareApply` rejects an empty
one as a conflict (internal/threadspike/manifest.go:348-350). In an id-less space that
produces `ErrConflict: apply plan belongs to planning space "", current space is ""` on
every apply: a nonsensical message for a repository that is working perfectly well. This
repository happens to carry `id = "6g1xbp8fvcrm"`, so dogfooding will not surface it.

**Recommendation:** Decide and record which: compose refuses in an id-less space with an
actionable "run `tskflwctl init` / mint an id first" message; or the plan records
`planning_space_id: null` and apply verifies path identity instead, with a warning. Either
is fine; silently failing with a conflict on two empty strings is not. Add the case to
`6g3q4rtv8d0a`'s stress tests beside "wrong planning space".

#### M5. Which reads fail closed is contradicted between the ADR and the task that implements them  · **Status:** tracked by 6g3q4rt7mgjn

**File:** planning/adrs/0006-adopt-threads-as-task-dags.md:242 | **Component:** read contracts
**Effort:** S · **Urgency:** soon

**[before-slice-2]**

ADR §5 item 2 requires a "Status-aware Thread frontier, **failing closed** for defects in
the relevant graph". Task `6g3q4rt7mgjn` AC 5 requires that "Lint and **queries** remain
useful for hand-edited broken repositories while mutation fails closed". The frontier is a
query. The spike implements a third behaviour: `Gate` never fails at all — a missing task
degrades to `GateBroken` (graph.go:192-195) and the projection keeps rendering.

This matters most for `task list --thread <t> --unblocked` (ADR §8), which would make one
flag on an ordinarily-resilient listing switch the command's failure semantics. It also
decides what dogfood checkpoint 1 actually runs.

**Recommendation:** Settle it as: **mutations fail closed; reads degrade and report.** A
read over a defective graph returns its answer plus the problem list and a non-zero
`graph_health` marker in the wire envelope, and human output says which tasks were
excluded. That keeps `lint`'s repair loop usable, keeps `--unblocked` from becoming a
trapdoor, and preserves the ADR's real intent (no *write* on an unsound graph). Amend §5
item 2 accordingly and align `6g3q4rt7mgjn` AC 5.

#### M6. Lifecycle moves silently rewrite the downstream graph, and no receipt says so  · **Status:** tracked by 6g3q4rte8kc1

**File:** planning/adrs/0006-adopt-threads-as-task-dags.md:548 | **Component:** lifecycle / receipts
**Effort:** M · **Urgency:** soon

**[before-slice-3]**

The ADR requires blast-radius disclosure for *dependency edits* ("A dependency edit is
global even when initiated while looking at one Thread. Human and machine output must make
that blast radius explicit"). It requires nothing equivalent for *status* moves, which have
a comparable radius under these semantics:

- `task deprecate X` makes every transitive descendant's gate `broken` (a withdrawn task is
  `broken` at graph.go:217, propagated at 199-201) — including descendants that are already
  completed, which become permanently inconsistent until an edge is removed. §3's stated
  remedy ("correct the edge when the constraint is no longer real") is a second, separate
  command the user is never told they need.
- `task defer X` makes descendants blocked for as long as X stays parked.
- Reopening any completed task invalidates completed descendants and can make a completed
  Thread inconsistent — and §4 already mandates that those changes "must report the affected
  Threads", which requires scanning every Thread document on every task status change. No
  task's acceptance criteria carry that requirement; `6g3q4rtmv4ak` mentions completed-thread
  drift only in its stress tests.

**Recommendation:** Add one acceptance criterion to `6g3q4rte8kc1`: every transition that
can change downstream gate state (`deprecate`, `defer`, and any move out of `completed`)
returns the count of affected descendants and the affected Thread ids, in both human and
machine receipts, and names the remedy. Decide at the same time whether that Thread scan
runs on every task move or only in `lint` / `thread show` — the ADR currently implies the
former without costing it.

#### M7. `--force` is one boolean carrying two different gates, and the paths that need it do not have it  · **Status:** tracked by 6g3q4rte8kc1

**File:** internal/cli/task.go:713 | **Component:** cli / lifecycle
**Effort:** S · **Urgency:** soon

**[before-slice-3]**

`Service.Move(slug, to, dryRun, force)` threads a single `force` down to `store.moveTask`,
where today it means exactly one thing: bypass the acceptance-criteria gate on a completion
(internal/store/fsstore.go:166-173). ADR §3 gives the same flag a second meaning on `start`
("bypasses only the dependency gate"). Two gates, two verbs, one boolean.

Meanwhile the surfaces that will need it do not expose it:

- `task move <task>... <status>` hardcodes `force=false` (internal/cli/task.go:718) and has
  no `--force` flag. After slice 3, `task move foo in-progress` on a blocked task fails with
  no override on that verb at all.
- `task start` has no `--force` flag either: `newTransitionCmd` adds one only when
  `to == domain.StatusCompleted` (internal/cli/task.go:770-777), with a comment explaining
  that offering it elsewhere "would advertise a check that does not exist there". That
  comment stops being true in slice 3.

**Recommendation:** Replace the boolean with a small typed override set (or at minimum
document `force` as "bypass whichever gate applies to this destination"), add `--force` to
`task start` and to `task move`, and confirm that `complete --force` continues to mean
acceptance criteria only — completing an ineligible task is legal by design under §3 and
must not start requiring a flag. Add both flags to `6g3q4rte8kc1`'s scope.

#### M8. The epic contradicts itself and the ADR about which slice does the dependency dogfood  · **Status:** fixed 2026-08-26

**File:** planning/epics/30-threads-and-task-dependency-graphs.md:26 | **Component:** planning docs
**Effort:** XS · **Urgency:** soon

**[before-slice-2]**

Epic line 26-27: "These bootstrap edges are prose until the guarded dependency-write surface
lands. **Slice 3** must persist them through the production commands". Epic line 94:
"2. **Slice 2** uses production dependency commands to sequence all remaining epic tasks."
ADR §10.1: "After guarded dependency writes land, sequence the remaining tasks in this
epic" — and guarded dependency writes are ADR slice 2. Line 26 is wrong under the epic's own
slice table, where slice 3 is eligibility enforcement.

The root cause is a numbering collision worth fixing once: ADR §10 and the epic table both
count **seven slices**, but the epic lists **eight production tasks**, because ADR slice 2
("Portable guarded dependency writes") is split into `6g3q4rt0wzkq` (guard) and
`6g3q4rt7mgjn` (operations). "Slice N" therefore means different things in different
paragraphs.

**Recommendation:** Fix line 26 to say slice 2, and add one line to the epic mapping slice
numbers to task ids (slice 2 = `6g3q4rt0wzkq` + `6g3q4rt7mgjn`). Prefer task ids over slice
numbers in prose from here on.

### Low

#### L1. `task depend add` idempotency is undecided, and the two prototyped surfaces already disagree  · **Status:** tracked by 6g3q4rt7mgjn

**File:** internal/threadspike/graph.go:177 (spike/threads-dag-mcp) | **Component:** dependency mutation
**Effort:** XS · **Urgency:** soon

**[before-slice-2]**

`WithEdges` rejects an already-present edge as a validation error
(`dependency %s -> %s already exists`, graph.go:177-179), while `PrepareApply` skips it as
`already-applied` (manifest.go:382-386). Task `6g3q4rt7mgjn` AC 1 correctly demands that
"idempotency behavior is explicit" but does not say which. Two answers in one release will
be discovered by an agent scripting retries.

**Recommendation:** Idempotent no-op with a `skipped` receipt entry on both surfaces, matching
how `thread add` on an existing member and `Move` to the current status already behave.

#### L2. A concurrency-induced cycle rejection is indistinguishable from a user error  · **Status:** tracked by 6g3q4rt7mgjn

**File:** internal/core/retry.go:62 | **Component:** error contracts
**Effort:** XS · **Urgency:** eventually

**[monitor]**

When two writers add opposite edges, the loser is rejected with `ErrValidation` (exit 11) and
the message "dependency cycle: …" (spike graph.go:118, exercised by
`TestThreadSpikeConcurrentOppositeEdgesCannotCommitACycle`). Exit 11 is arguably the right
code — the union really is cyclic and `retryOnConflict` (internal/core/retry.go:62-72) would
be wrong to retry it — but the user typed a request that was valid when they typed it, and
the message gives no hint that someone else moved first.

**Recommendation:** Keep exit 11; add the concurrent-origin detail to the message ("… the
prerequisite edge B → A was added after this command started"). One sentence, no contract
change.

#### L3. `task blockers` returns a set, but the design's motivating question asks for a chain  · **Status:** tracked by 6g3q4rst78qy

**File:** internal/threadspike/graph.go:373 (spike/threads-dag-mcp) | **Component:** graph queries
**Effort:** S · **Urgency:** eventually

**[monitor]**

ADR §Context asks "*What prerequisite chain explains why a task is blocked?*". `Blockers`
returns a flat, id-sorted set of transitive blockers with no path, no distance, and no
reason (graph.go:373-391). Task `6g3q4rt7mgjn` AC 3 improves on this by requiring direct and
transitive to be distinguished, which is necessary but still not a chain.

**Recommendation:** Carry a `depth` (or the shortest blocking path) per blocker entry in the
machine contract, and let human output render the nearest chain. Cheap now, awkward to add
after the DTO ships. Combine with M2's reason token.

#### L4. Dogfood checkpoint 1 has nothing to read  · **Status:** fixed 2026-08-26

**File:** planning/epics/30-threads-and-task-dependency-graphs.md:93 | **Component:** rollout policy
**Effort:** XS · **Urgency:** eventually

**[monitor]**

"Slice 1 runs strict read-only analysis against this planning repository." This repository
currently contains **zero** `depends_on` edges across 279 task files, so the analysis returns
an empty graph and proves nothing beyond "the scan does not crash". Its only real content
would be the six migrated legacy edges — which M3 recommends deferring out of slice 1.

Confirmed baseline for the checkpoint: `tskflwctl lint --json` reports `unreadable: 0,
issues: 0`, so the strict snapshot's fail-closed default will not immediately block work.

**Recommendation:** Restate checkpoint 1 as what it can honestly deliver — a clean strict
snapshot over 279 real tasks, timing measured, plus the legacy-vocabulary diagnostic
report — and move the "explanatory queries" claim to checkpoint 2, where real edges exist.

#### L5. Batch transitions will rebuild the whole graph once per task  · **Status:** tracked by 6g3q4rte8kc1

**File:** internal/cli/moves.go:24 | **Component:** cli / performance
**Effort:** S · **Urgency:** eventually

**[monitor]**

`runMoves` loops slugs calling `Service.Move` once each (internal/cli/moves.go:24-43). Once
eligibility is centralized in core, each call rebuilds a strict snapshot over the whole
repository. The tool's own documented idiom is
`task list -q --tag tui | xargs tskflwctl task start` (internal/cli/task.go:199) — N full
scans of 279 files, and N chances for the answer to change mid-batch.

**Recommendation:** Nothing structural now; note it in `6g3q4rte8kc1` and measure during
dogfooding. If it bites, the fix is a per-invocation snapshot cache passed through the batch,
not a different policy.

#### L6. `thread complete` on a Thread with no live members is vacuously true  · **Status:** tracked by 6g4wm2yf6tyj

**File:** planning/adrs/0006-adopt-threads-as-task-dags.md:225 | **Component:** thread lifecycle
**Effort:** XS · **Urgency:** eventually

**[before-slice-4]**

§4 requires `thread start` to have "at least one non-withdrawn member" and states completion
as "every non-withdrawn member is soundly completed" — vacuously satisfied by zero members.
§6 says an empty Thread "cannot start or complete", so the intent is clear; the rule as
stated in §4 does not encode it. A Thread whose every member was later deprecated reaches
the same state.

**Recommendation:** State the `≥ 1 non-withdrawn member` precondition on `complete` as well
as `start`, in `6g3q4rtmv4ak`'s acceptance criteria.

**Resolution:** The 2026-08-29 Thread readiness split moved start/complete
lifecycle enforcement into the dedicated guarded Thread mutation task.

#### L7. Thread ids and task ids share a mint but not a namespace check  · **Status:** tracked by 6g3q4rtmv4ak

**File:** internal/store/resolve.go:104 | **Component:** identity
**Effort:** XS · **Urgency:** eventually

**[deferred]**

Resolution is per-directory (`flatCandidates(dir)`), so a Thread and a task carrying the same
12-char id would resolve independently rather than collide — but every graph command takes a
bare id, and the output would be quietly wrong. The spike guards it explicitly
(manifest.go:360-362 and 443-445). Collision probability is negligible; the guard costs one
map lookup.

**Recommendation:** Keep the spike's cross-kind collision check when the Thread store lands.
No ADR change needed.

## Sequencing critique

The epic's drawn graph (lines 29-44) is close to right. Three corrections:

1. **`6g3q4rte8kc1` (eligibility) is drawn as a prerequisite of `6g3q4rtmv4ak` (Thread
   entity), but only half of it really is.** The eligibility task bundles two separable
   things: *derivation* (lifecycle role, gate state, sound completion, eligibility,
   inconsistency — all pure functions over a snapshot) and *enforcement* (routing every
   transition through one guard, `--force` receipts). Threads need only the derivation, for
   frontier and rollup. The derivation also belongs naturally beside the traversal primitives
   already in `6g3q4rst78qy`'s scope ("blocker/downstream traversal, and topological waves").
   **Move gate/role/sound-completion derivation into `6g3q4rst78qy`**, leave `6g3q4rte8kc1`
   as pure enforcement, and the Thread entity then depends on task 1 only — letting
   enforcement and the Thread entity proceed in parallel and shrinking the largest task
   (`6g3q4rtmv4ak`, 4-7 days) by removing a false wait.

2. **The migration edge runs the wrong way (M3).** As scoped, `6g3q4rst78qy` performs a
   multi-file graph write, which by the ADR's own rule needs `6g3q4rt0wzkq`'s guard — making
   the first task depend on the second. Splitting migration into diagnose (task 1) and
   rewrite (task 3, `6g3q4rt7mgjn`) removes the inversion without adding a task.

3. **`6g3q4rt0wzkq` (guard) cannot actually be designed "alongside" task 1 (its own
   Sequencing line 41).** H1 shows the guard's shape — inversion of control, with core's
   planner running inside the store's critical section — determines the signature of the
   analysis contract task 1 delivers. Either sequence the guard's *API decision* before task
   1 starts, or accept that task 1's interface will be revised in task 2. Say which.

Not defects, for the record: bulk linking's dependency on dependency operations is present
transitively through the Thread entity; generated views correctly sit off the critical path;
and the TUI genuinely is last.

## Assumptions that survived review

Recorded so this audit is not read as uniformly negative — these were checked and hold.

- **The core model holds.** One task-owned global DAG with many Thread views is coherent
  under shared membership: membership lives in the Thread document, edges live on the
  dependent task, and removing a member changes no edge. Option C's objection (two acyclic
  Threads whose union is cyclic) is real and correctly avoided.
- **`Drained ≡ soundly completed` is a true identity**, not just a convenient phrasing:
  completed + (every prerequisite soundly completed) is exactly the recursive definition.
- **Every persisted prefix of a validated edge set is acyclic**, so ADR §7's "every persisted
  prefix must itself remain a sound repository graph" holds for existing-task edge writes
  without extra machinery; the topological-wave requirement is needed only for *new* tasks,
  which is exactly where the ADR puts it.
- **Routing eligibility through `core.Service.Move` covers the TUI for free.** Every TUI
  transition already goes through it (internal/tui/entity.go:239, internal/tui/model.go:844,
  873, 893) — no separate TUI enforcement path is needed, and `6g3q4rv89vzw` AC 4's "no
  TUI-local readiness logic" is achievable.
- **The legacy-field survey in the ADR is accurate.** Exactly six live `blocked_by` users; all
  six referents resolve to existing files; all seven involved tasks are `completed`; the
  union is acyclic. `blocks` has zero users, `dependencies` has zero users on tasks, and the
  task-level `projects` field has zero users — so "deprecate `projects`, do not rename it to
  `threads`" costs nothing.
- **The Projects scaffold is genuinely empty.** `planning/projects/` contains only
  `.gitkeep`; ADR-0002 was never implemented. §9's "may remove an empty placeholder" is safe
  provided "empty" is defined as "contains only `.gitkeep`" — worth one line in
  `6g3q4rtmv4ak`.
- **`lint --fix` already repairs a scalar written into a list field** (internal/store/fix.go:353-355),
  so a `depends_on` scalar written by `--set … --force` is recoverable rather than fatal.
- **Lint already covers archived tasks** for missing/unknown status and id drift
  (internal/core/service.go:254-264), so the graph's vertex set (all tasks) and lint's
  coverage are much closer than the "lint is active-only" comments suggest.
- **This repository is graph-ready today**: `lint --json` → `unreadable: 0, issues: 0` across
  279 task files, so fail-closed mutation will not be blocked on day one.
- **Declining a graph dependency for V1 is well supported.** The comparison in the spike is
  honest, `dominikbraun/graph`'s unstable v0 API is a real cost, and taskflow would own every
  semantic contract regardless. Keeping the bounded contract-test bake-off inside slice 1
  rather than as a second research spike is the right call.
- **Two-phase compose/apply with IDs minted once at compose is the right shape**, and the
  spike's evidence for it (interruption, retry, no duplicate mints) is sound apart from H4's
  equivalence predicate.
- **flock does serialize goroutines within one process** (separate open file descriptions), so
  `TestThreadSpikeConcurrentOppositeEdgesCannotCommitACycle` is valid evidence as far as it
  goes — it just does not cover cross-process, non-Unix, or the nested-lock case in H1.

## Coverage the acceptance evidence does not include

Not a defect in the spike — its Prototype boundary said so — but material when reading
ADR-0006 as "validated". The spike's tagged CLI exposes only `compose`, `apply`, `list`,
`show`, `plan` (internal/cli/threadspike_enabled.go). It never implemented:

- Thread lifecycle (§4): `start` / `complete` / `abandon` / `reopen`. `ThreadStatus` exists as
  data set by fixtures; `validatePlannedThread` requires a new Thread to be `unstarted`
  (manifest.go:505-507). Every §4 rule is unexercised.
- Membership mutation (§8): `thread add` / `remove`.
- The dependency mutation verbs (§8): `task depend add` / `remove` — only the in-memory
  `WithEdges` was tested.
- Eligibility enforcement (§3): no transition guard exists in the prototype at all; only the
  derived projection was validated.
- The legacy migration (§2).

The two areas carrying the most unresolved semantics in this review — §4 Thread lifecycle
(M1, L6) and §3 single-policy enforcement (H3, M2, M6, M7) — are precisely the two that
received no executable validation.

## Questions requiring decider input

1. **Guard shape (H1):** store-owned callback with core's planner running inside the lock, or
   a store-owned graph mutation method that learns graph semantics? One of ADR §9 or
   `6g3q4rt0wzkq`'s "no locking in core services" has to give.
2. **`task edit` (H3):** product path that re-runs the eligibility and graph guards, or
   declared hand-edit equivalent that only `lint` diagnoses?
3. **`task set --set depends_on=` (H3):** rejected like `status`, including under `--force`?
4. **External gates and `thread complete` (M1):** does an unfinished external prerequisite
   block completion? If yes, does the rollup gain a separate `drained` count?
5. **Gate precedence (M2):** broken over blocked; and is a `deferred` prerequisite blocked or
   broken?
6. **Apply already-applied predicate (H4):** identity-only, or content equality?
7. **Planning-space id (M4):** refuse compose in an id-less space, or bind by path with a
   warning?
8. **Read failure semantics (M5):** does `thread frontier` / `task list --unblocked` fail
   closed on a defective graph, or degrade and report?
9. **Migration timing (M3):** diagnose in task 1 and rewrite in task 3, or move the guard
   ahead of task 1?
10. **Downstream disclosure (M6):** does every task status move scan `threads/` to report
    affected Threads, or is that a `lint` / `thread show` responsibility?

## Triage resolution (2026-08-26)

ADR-0006's 2026-08-26 amendment records the accepted mutation-boundary, analysis-complexity,
generic-edit, gate/read, lifecycle-impact, Thread-completion, migration, bulk-linking, identity, and
idempotency contracts. Existing production tasks were tightened as the tracked destinations; no
additional planning tasks were necessary.

Two review-only inconsistencies were fixed directly: M8 now names task IDs and maps ADR slice 2 to
the guard plus dependency-operation tasks, while L4 now describes the honest edge-free dogfood
checkpoint. H4 is tracked by the bulk-link task through the explicit exclusion of inline task
creation from V1; a future task-creation amendment must solve provenance/equivalence before restoring
that path. M5 was resolved more conservatively than the original recommendation: diagnostic reads
degrade and report, while dispatch-oriented frontier selectors return no eligible work on an unsound
relevant graph.
