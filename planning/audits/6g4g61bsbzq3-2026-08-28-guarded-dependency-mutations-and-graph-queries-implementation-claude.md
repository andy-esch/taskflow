---
schema: 1
id: 6g4g61bsbzq3
bucket: closed
area: guarded-dependency-mutations-and-graph-queries-implementation-claude
date: "2026-08-28"
updated_at: "2026-08-28"
---

# Audit: guarded-dependency-mutations-and-graph-queries-implementation-claude — 2026-08-28

> Edit findings in place and flip each `**Status:**` as you work it.

Adversarial implementation review of the guarded task-dependency operations slice
(task `6g3q4rt7mgjn`, epic `30-threads-and-task-dependency-graphs`, branch
`feat/guarded-dependency-operations`, uncommitted tree over merge-base `90bfabe`),
evaluated against ADR-0006 (including the new 2026-08-27 amendment), the readiness
audit `6g4aj7v60syg`, and `docs/ARCHITECTURE.md`.

Method: every claim below was executed against a built binary and throwaway
repositories in a scratch directory, never against the real planning tree. Two
invariants were additionally checked by **mutation testing** — deliberately breaking
the implementation in a copy of the tree outside the repository and re-running the
committed suite to see whether anything fails. `git status --short` reports the same
69 entries before and after this review; this audit file is the only repository write.

## Executive verdict

**Ready with required fixes.** The hard parts are right and survived attack. Reference
resolution genuinely happens inside the authoritative snapshot and is byte-identical to
`store.resolveID` across exact/case/prefix/substring/ambiguity/unsafe-name/duplicate-slug
and unreadable-file inputs. Two real OS processes racing opposite edges produced exactly
one commit and one cycle refusal with an acyclic final graph. An interrupted multi-file
migration (forced with a real `EPERM` on the second rename, not a test hook) left a
degraded-but-sound durable prefix, emitted structured `applied`/`remaining` recovery data
on the error envelope, printed human retry guidance, left no temp files, and converged on
rerun. No-op adds and removes are byte-identical with no `updated_at` advance. The
migration write order really is prefix-safe, and the real planning tree's six legacy
occurrences converted to eight canonical edges with correct direction.

Four things should be fixed before this task closes:

1. **The two new explanatory commands report an affirmative all-clear on a `degraded`
   graph** (H1). `task blockers` prints a green `✔ no blockers` and returns
   `"blockers":[]` for a task whose resolvable legacy prerequisite is not started;
   `task unblocks` returns `"unblocks":[]` for that prerequisite. After `task depend
   migrate` the same repository reports the constraint correctly. Degraded is exactly
   the state this slice exists to leave, and `task start` does not yet enforce
   eligibility, so a false all-clear is actionable today.
2. **The migration's prefix-safe write order — the one invariant standing between an
   interrupted `blocks` migration and permanent, silent edge loss — has no test** (H2).
   Reversing the wave comparison leaves all 23 packages green, and the resulting binary
   destroys the edge irrecoverably under the same injected failure the correct binary
   survives.
3. **`task blockers` carries no derived state for the queried task** (M1), so the only
   thing an agent can read from a clean result is the empty blocker list — precisely the
   inference `docs/ARCHITECTURE.md` says must never be made.
4. **User-facing guidance still says the guarded commands are unavailable** (M3):
   `task set` answers "`task depend add/remove` once available" and `lint` answers
   "canonical migration is intentionally deferred to guarded dependency operations",
   both of which are now false.

The remaining findings are genuine but lower-stakes, and two of them are better as
separate follow-ups.

## Findings

#### H1. Degraded snapshots make `task blockers` and `task unblocks` report a false all-clear  · **Status:** fixed

**File:** internal/core/dependency_graph.go:369,389-390,414-416 | **Component:** core/graph, cli/render
**Effort:** M · **Urgency:** acute

`g.dependencies` and `g.outgoing` are populated only from canonical `depends_on`
(dependency_graph.go:369, 389-390, 404-405). Resolved legacy edges are appended to
`projectedEdges` for `analyzeDAG` alone (dependency_graph.go:414-416), so they influence
cycle attribution but never enter the traversals that `projectBlockers`
(dependency_graph.go:890) and `DownstreamImpact` (dependency_graph.go:1039) walk. Both new
commands therefore answer as if the legacy constraints did not exist, while reporting
`health: degraded` and a separate `legacy_dependencies` array.

Reproduced end to end. Two tasks, `alpha` (`ready-to-start`) and `beta`
(`ready-to-start`, `blocked_by: [alpha]`):

```
$ tskflwctl task blockers beta
beta  (6g0000000b02)
graph:  degraded
view:  frontier
✔ no blockers
⚠ legacy blocked_by on 6g0000000b02; run task depend migrate

$ tskflwctl task blockers beta --causal --json
… "health":"degraded","blockers":[], …

$ tskflwctl task unblocks alpha
alpha  (6g0000000a01)
graph:  degraded
view:  downstream impact
• no downstream tasks

$ tskflwctl task depend migrate >/dev/null
$ tskflwctl task blockers beta
• alpha  not-started  direct
$ tskflwctl task unblocks alpha
• beta  candidate/blocked  direct
```

The same repository, same constraint, opposite answers on either side of a migration
that ADR-0006 explicitly says preserves semantics. `--causal` — the projection sold as
"the full forensic closure" — is equally empty.

**Why it matters.** Degraded is not an exotic state: it is the state of any repository
that has not yet run `migrate`, that a colleague's older binary or a hand edit
reintroduces, or that a partially-applied migration leaves behind (H2's recovery path
lands there by design). The failure mode is a green checkmark and an empty machine array
in answer to "is anything blocking this?", which is the one question where a false
negative causes work to start on an unsatisfied prerequisite. `task list --unblocked`
does fail closed on degraded, and the eligibility guard for `task start` is the *next*
slice (`6g3q4rte8kc1`), so nothing else catches this today. The `legacy_dependencies`
array is a real mitigation only for a consumer who already knows to distrust `blockers`.

This is arguably inherited projection behavior, but this slice is what turns it into a
public product answer, so it belongs here.

**Recommendation:** Smallest correction: stop emitting an affirmative empty result when
the snapshot is not healthy. In `render.TaskBlockersHuman`/`TaskUnblocksHuman`
(cli/render/dependency.go:61-63, 89-91) replace `✔ no blockers` / `• no downstream
tasks` with an explicit "no canonical constraints; N legacy occurrence(s) are not
projected into this view — run `task depend migrate`" whenever `Health != healthy`, and
add a boolean such as `legacy_constraints_unprojected` to both envelopes so a machine
consumer sees the same caveat. The larger correction — projecting resolved legacy edges
into the diagnostic blocker/downstream traversals so degraded reads are simply correct —
is a defensible follow-up, but the misleading affirmative must not ship.

**Resolution:** Projected every exactly resolved legacy edge into blocker,
downstream, gate, and sound-completion reads; focused core and CLI tests cover
degraded snapshots.

#### H2. The prefix-safe migration write order has no test; reversing it silently destroys edges  · **Status:** fixed

**File:** internal/core/dependency_operations.go:302-326 | **Component:** core/migration planner, store tests
**Effort:** S · **Urgency:** acute

`planLegacyDependencyMigration` orders writes by descending topological wave
(dependency_operations.go:320-326) so a dependent gains its canonical edge before the
prerequisite that declared it via legacy `blocks` clears that declaration. The comment at
302-305 states exactly why. The implementation is correct — I confirmed the plan order
on a `blocks`-only fixture:

```
$ tskflwctl task depend migrate --dry-run --json | jq .planned_task_ids
["6g0000000b02", "6g0000000a01"]     # dependent first, clear-owner second
```

**No test defends it.** I copied the tree outside the repository, changed line 323 from
`return leftWave > rightWave` to `return leftWave < rightWave`, and ran the committed
suite: 23 packages ok, zero failures. The three migration tests in
`internal/store/dependency_operations_test.go` cannot detect the reversal, because in
every fixture the edge is declared from *both* directions (`blocks` on one task and
`blocked_by` on the other) or from `blocked_by` alone — so a wrong order still converges
on retry.

The reversal is not benign. Building the mutant binary and forcing the second rename to
fail (`chflags uchg` on the second target, an ordinary `EPERM`, no test hook) on a
`blocks`-only fixture:

```
mutant plan:  ["6g0000000a01", "6g0000000b02"]    # clear-owner first
after failure: (no dependency fields anywhere)
retry:        changed=False  edges=[]
final:        health=healthy, blockers=[]
```

The legacy declaration is cleared, the canonical edge was never written, the retry
reports nothing to do, and the graph is reported **healthy**. That is silent,
unrecoverable loss of a planning constraint. The correct binary under the identical
failure leaves `blocks: [beta]` on the owner plus `depends_on: [alpha]` on the
dependent — degraded, diagnosable, and convergent on rerun.

**Why it matters.** Acceptance criterion "Migration remains graph-sound after every
injected write failure and retry converges from every durable prefix" is ticked, but the
single ordering case where soundness is load-bearing is untested. Any future refactor of
the ordering block — including a plausible "simplify to plain ID order" cleanup — ships
green.

**Recommendation:** Add one store-level test: a `blocks`-only fixture (owner declares
`blocks: [dependent]`, dependent declares nothing), `testHookAfterGraphWrite` failing
after the first write, asserting the dependent's `depends_on` contains the prerequisite,
the owner still carries `blocks`, health is `degraded` not `healthy`, and the rerun
converges. That test fails under the reversed comparison.

**Resolution:** Added the blocks-only store interruption case: dependent writes
first, owner retains legacy evidence after failure, degraded prefix is sound,
and retry converges.

#### M1. `task blockers` exposes no derived state for the queried task  · **Status:** fixed

**File:** internal/wire/dependency.go:107-115,143-150 | **Component:** wire/query envelopes, cli/render
**Effort:** S · **Urgency:** soon

`TaskBlockersEnvelope` and `TaskUnblocksEnvelope` carry `GraphTaskJSON` (id, slug,
status, epic) for the queried task and a full `TaskGraphStateJSON` for every *other*
task in the payload. The queried task's own role, gate, and `eligible` are absent, and
the human header (cli/render/dependency.go:108-116) prints only name, graph health, and
projection.

`docs/ARCHITECTURE.md` (this branch, line ~155) says "Eligibility is read from derived
state, never inferred from an empty blocker list", and the task's stress-test list says
"authorization must never be inferred from an empty blocker result". The blockers
envelope offers no other signal, so it invites the inference it forbids:

```
$ tskflwctl task blockers parked --json
{… "task":{"task_id":"6g0000000d01","slug":"parked","status":"deferred"},
   "projection":"frontier","health":"healthy","blockers":[], …}
$ tskflwctl task blockers parked
✔ no blockers
```

`parked` is `deferred`: role `parked`, gate `clear`, `eligible: false`, and correctly
excluded from `task list --unblocked`. Nothing in the blockers answer says so. A
consumer can recover it from `status` plus the role table, but the derived field the
architecture doc points at is not in the payload it points at.

**Recommendation:** Add `"state": TaskGraphStateJSON` for the queried task to both
envelopes (the data is already in hand — `graph.State(taskID)` is one call in
`Service.TaskBlockers`/`TaskUnblocks`), and render `role/gate` in `graphQueryHeader`
alongside graph health. Both are additive and land inside the existing schema/golden
gates.

**Resolution:** Both query results and wire envelopes now carry queried-task
state; human headers print role, gate, and eligibility. Wire schema advanced to
1.51.

#### M2. No product path repairs a broken dependency graph  · **Status:** tracked by 6g4g8gatbnrs

**File:** internal/core/dependency_graph_mutation.go:14-23 | **Component:** core/store, recovery UX
**Effort:** M · **Urgency:** soon

Once the graph is `broken`, every write path refuses:

```
$ tskflwctl task depend remove cyca --on cycb
error: validation failed: repository task graph is broken; repair it before mutation:
dependency cycle: … (2 additional problem(s); run lint for the full sweep)

$ tskflwctl lint --fix --dry-run
nothing to fix

$ tskflwctl task set cyca --unset depends_on --force
error: validation failed: depends_on is graph-owned …

$ tskflwctl task edit cyca      # rejects any dependency delta (store/edit.go:175)
```

`ValidateTaskGraphMutationSource` rejects the whole repository, not the offending
component, so even a removal that would *repair* the cycle is refused; and a dangling
`depends_on` value cannot be removed at all, because it does not resolve to a task.
`fix.go:349` deliberately skips graph-owned keys, so `lint --fix` cannot help either.
The only remedy is a raw filesystem edit — which the tool otherwise treats as the
non-cooperating writer it defends against.

Failing closed is the specified behavior and I am not disputing it. The defect is the
guidance: "repair it before mutation … run lint for the full sweep" reads as though a
repair path exists inside the tool, and `lint --fix` then says "nothing to fix".

**Recommendation:** In this task, fix the message: name the offending file and field
(the `GraphProblem` already carries `Path` and `Field`) and say plainly that graph
repair is a direct file edit today, then rerun `lint`. A guarded repair mode — a removal
that is admitted when it strictly reduces the problem set, or `lint --fix --graph` — is
genuinely separable work and belongs in its own task.

**Resolution:** Current failures now name the first file and field and state
that direct frontmatter repair plus lint is required. A monotonic guarded repair
path is separately scoped in task 6g4g8gatbnrs.

#### M3. Generic-path guidance still says the guarded commands are unavailable  · **Status:** fixed

**File:** internal/core/service_task.go:332-336 | **Component:** core/cli messages
**Effort:** XS · **Urgency:** acute

Four different phrasings answer "you cannot change this here", two of them now false:

```
$ tskflwctl task set beta --set depends_on=6g0000000a01 --force
error: … depends_on is graph-owned and cannot be changed with `task set`
(including --force); use guarded dependency operations (`task depend add/remove` once available)

$ tskflwctl lint                       # degraded repo
  blocked_by: legacy dependency field: … ; canonical migration is intentionally
  deferred to guarded dependency operations

$ tskflwctl task depend add beta --on alpha     # degraded repo
error: … 1 legacy dependency field occurrence(s) remain; run the guarded migration

$ tskflwctl task list --unblocked               # degraded repo
error: … 1 legacy dependency field occurrence(s) remain; run `task depend migrate`
```

`service_task.go:334` says "once available" and `service.go:401` says "intentionally
deferred" about commands that ship in this very branch. The task's own command contract
requires that `task set` and `task edit` "direct users to `task depend add/remove`"; the
parenthetical actively tells an agent the command does not exist yet. The readiness
audit's M1 raised precisely this ("lint messages cannot cite an exact command name") and
its resolution was accepted.

Separately, `graphMutationHealthDetail` (dependency_graph_mutation.go:107-119) and
`taskGraphHealthDetail` (dependency_operations.go:437-449) are near-duplicate helpers
that differ only in this sentence — which is how the two spellings drifted.

**Recommendation:** Name the real commands in all four places, and collapse the two
health-detail helpers into one. `internal/core/setfields_coercion_test.go:121` asserts
only the substring "guarded dependency", so no test pins the stale text.

**Resolution:** All graph-owned-field, degraded-health, and lint guidance names
the exact supported tskflwctl dependency commands; the duplicate health helper
was consolidated.

#### L1. Snapshot resolution silently picks a winner where Store resolution reports ambiguity  · **Status:** fixed

**File:** internal/core/dependency_graph.go:752-806 | **Component:** core/graph resolution
**Effort:** XS · **Urgency:** eventually

`ResolveTaskID` builds one candidate per canonical task ID (line 763, over `g.ids`,
which `newTaskGraph` deduplicates at line 344). `store.resolveID` builds one candidate
per *file*. For two files declaring the same `id:`, the two disagree:

```
$ tskflwctl task show 6g0000000a01
error: "6g0000000a01" matches 2 tasks: one (6g0000000a01), two (6g0000000a01): ambiguous match

$ tskflwctl task blockers 6g0000000a01 --json
{… "task":{"task_id":"6g0000000a01","slug":"one", …}, "health":"broken", …}
```

Acceptance criterion 1 claims parity "for exact IDs/slugs, case-insensitive unique
prefixes, unique substrings, missing values, **and ambiguity**". Every other tier I
tested — exact ID, exact slug, case-insensitive exact, ID prefix, slug prefix, substring,
duplicate slug, unreadable file addressed by ID or by filename slug, `../escape`, empty —
matches the Store byte for byte including the error text. Duplicate ID is the one gap.

Impact is contained: a duplicate ID makes the graph `broken`, so `ValidateTaskGraphMutationSource`
rejects every mutation before the planner runs. Only the diagnostic queries are affected,
and they report `health: broken` plus the `duplicate-task-id` problem. But the answer is
about an arbitrary one of two files without saying which.

**Recommendation:** Keep every duplicate-ID entry in the candidate list (the graph already
knows them via `idCounts`/`idPaths`), so the tier returns `ErrAmbiguous` exactly as the
Store does. Add the case to `TestTaskGraphResolveTaskIDMatchesRepositoryReferenceTiers`.

**Resolution:** TaskGraph retains one resolver candidate per source file, so
duplicate stable IDs remain ambiguous exactly as Store resolution does;
regression coverage added.

#### L2. Downstream and causal query output is quadratic and unbounded; the stress test does not cover it  · **Status:** deferred

**File:** internal/core/dependency_graph.go:1039-1076 | **Component:** core/graph, wire output size
**Effort:** S · **Urgency:** eventually

`DownstreamImpact` materializes one full path per reachable dependent and caches it, so a
chain of depth *n* produces Θ(n²) path elements. Measured on a synthetic 1,500-task linear
repository:

```
$ /usr/bin/time -l tskflwctl task unblocks 6g0000000000 --json > out.json
0.09 real, 138,608,640 maximum resident set size
$ wc -c < out.json
17232679                      # 17.2 MB of JSON for one query
$ tskflwctl task blockers 6g0000001499 --causal --json | wc -c
17267175
```

Latency is fine; the payload is not, for a contract whose whole point is machine
consumption. `TestTaskGraphSupportedDeepChainEnvelope`
(internal/core/dependency_graph_test.go:327-360) builds the 4,096-edge chain but only
calls `State` and `BlockingFrontier`, which return one blocker — so the acceptance
criterion "Deep-chain stress establishes the supported graph-depth envelope" is
established for the cheap projection and not for the two that amplify. Extrapolated,
4,096 would emit roughly 120 MB.

Real planning graphs are shallow (the production tree's deepest chain is 6), so this is
scale robustness, not a present defect.

**Recommendation:** Extend the existing stress test to call `DownstreamImpact` and
`CausalBlockers` and assert the measured envelope, so the number is recorded rather than
assumed. Capping or gating path materialization (`--paths=false`) is a reasonable
follow-up if a real repository ever approaches it.

**Resolution:** A 512-edge test now records the quadratic causal/downstream
full-path amplification while the 4096-edge structural/frontier envelope remains
covered. Output capping is deferred until dogfood exceeds the live depth of six.

#### L3. Present-but-empty legacy keys are invisible to lint, health, and migrate  · **Status:** fixed

**File:** internal/core/dependency_graph.go:551-553 | **Component:** core/graph, migration
**Effort:** S · **Urgency:** eventually

`resolveLegacyDiagnostics` skips a field whose value list is empty, so a task carrying
`blocked_by: []` / `blocks: []` produces no diagnostic, keeps the repository `healthy`,
and is never touched by migration:

```
$ tskflwctl lint --json
{"schema_version":"1.50","unreadable":[],"issues":[]}
$ tskflwctl task depend migrate --json
{… "changed":false, "cleared_legacy_fields":[], …}
$ grep blocked_by planning/tasks/6g0000000a01-alpha.md
blocked_by: []
```

ADR-0006 §2 says implementation must "converge on `depends_on` alone"; an empty legacy
key survives forever, and a later hand edit that fills it in reintroduces legacy
vocabulary. There is no correctness consequence today (`task set` and `task edit` both
refuse to write the key), so this is completeness, not a bug.

**Recommendation:** Either treat a present-but-empty legacy key as a clearable occurrence
(needs a parser signal distinguishing "absent" from "empty"), or state the carve-out in
the migration's documentation. Reasonable as a follow-up rather than current scope.

**Resolution:** Parser-side field-presence metadata makes empty or null legacy
keys degraded lint/migration occurrences instead of invisible content.

#### L4. Migration receipts under-report cleared legacy fields  · **Status:** fixed

**File:** internal/core/dependency_operations.go:243-245 | **Component:** core/receipts
**Effort:** XS · **Urgency:** eventually

`ClearedLegacyFields` is built from the legacy *diagnostics*, but the materializer clears
all three legacy keys whenever `ClearLegacy` is set (store/graphmutation.go:150-154). A
task with one diagnosed field and one empty legacy key loses both and reports one:

```
# alpha declares  blocks: [beta]  and  dependencies: []
$ tskflwctl task depend migrate --json | jq '.cleared_legacy_fields[] | select(.task_id=="6g0000000a01")'
{"task_id":"6g0000000a01","field":"blocks"}
# file afterwards: both keys gone
```

The task's contract says migration receipts "identify planned, applied, skipped, and
remaining work"; the cleared-field list is the one place the receipt is not a faithful
description of the write.

**Recommendation:** Either restrict `ClearLegacy` to the fields that actually had
diagnostics, or derive `ClearedLegacyFields` from the materialized delta. The first is
smaller and keeps the planner pure.

**Resolution:** Every present legacy key now produces a diagnostic, so migration
receipts list every key the materializer removes, including empty companions;
store regression coverage added.

#### L5. Mutation receipts omit the resulting derived state, so a new edge can silently strand a task  · **Status:** tracked by 6g3q4rte8kc1

**File:** internal/cli/render/dependency.go:22-33 | **Component:** cli/receipts
**Effort:** S · **Urgency:** eventually

`task depend add` will happily point a live task at a `deprecated` prerequisite. ADR-0006
says withdrawn tasks "never satisfy downstream dependencies", so the dependent's gate
becomes permanently broken — reported with a green tick and no hint:

```
$ tskflwctl task depend add live --on dead
✔ added 6g0000000a01 -> 6g0000000b02
applied task files: 6g0000000b02

$ tskflwctl task blockers live --json | jq '.blockers[] | {id:.task.task_id, reason}'
{"id":"6g0000000a01","reason":"withdrawn"}
$ tskflwctl task list --unblocked --json
{"schema_version":"1.50","tasks":[]}
```

The graph stays `healthy` — this is a legal edge, not a defect — but the receipt is the
moment to say what just happened. The same applies to adding a not-yet-complete
prerequisite to a `completed` task, which makes it `inconsistent`.

**Recommendation:** Include the dependent's post-plan `TaskGraphState` in the mutation
receipt (the planner already holds the prospective graph via
`ValidateTaskGraphMutationPlan`), or at minimum warn in the human receipt when the
planned edge leaves the dependent's gate non-clear. Separable from the current slice; a
good pairing with M1's envelope change.

**Resolution:** The eligibility/lifecycle slice now owns a reusable before-after
graph-state impact shape and post-plan state in dependency add/remove receipts.

#### L6. Graph queries print every repository problem, repeated once per cycle member  · **Status:** fixed

**File:** internal/core/dependency_operations.go:376-377,412 | **Component:** core/queries, cli/render
**Effort:** XS · **Urgency:** eventually

`TaskBlockers`/`TaskUnblocks` attach `graph.Problems()` — the whole repository sweep —
to a single-task query, and `analyzeDAG` emits one `ProblemCycle` per SCC member with an
identical message, so the human renderer repeats it:

```
$ tskflwctl task blockers cyca
⚠ cycle: dependency cycle: 6g0000000a01 -> 6g0000000a02 -> 6g0000000a01
⚠ cycle: dependency cycle: 6g0000000a01 -> 6g0000000a02 -> 6g0000000a01
⚠ missing-dependency: task 6g0000000a03 depends on missing task 6g0000000zzz
```

The third line concerns a task unrelated to the query. `TaskGraph.LocalProblems` exists
for exactly this and has no production caller. Reporting the full sweep is a defensible
choice for a diagnostic read — a 20-member SCC producing 20 identical lines is not.

**Recommendation:** Deduplicate identical cycle problems in the human renderer, and
consider ordering the queried task's local problems first. Follow-up.

**Resolution:** Human graph diagnostics deduplicate identical code-and-message
entries with regression coverage. Structured machine output intentionally
retains the attributable repository-wide sweep.

#### L7. The new list-returning query has no `-o`/`-q` output mode  · **Status:** deferred

**File:** internal/cli/task_dependency.go:130-149 | **Component:** cli ergonomics
**Effort:** XS · **Urgency:** eventually

`task unblocks` returns a list but offers only `--json`, while `task list` offers
`-o human|json|name|table|csv` and `-q`. The README's own idiom is
`tskflwctl task list -q --tag tui | xargs tskflwctl task start`; the natural
`task unblocks <id> -q | xargs …` is unavailable. `task blockers` has the same shape.

**Recommendation:** Add `-q` (ids one per line) to both, reusing the existing name-mode
plumbing. Follow-up; no correctness impact.

**Resolution:** No correctness or current dogfood workflow requires terse output
yet; revisit after real blockers/unblocks scripting demonstrates the desired ID
versus slug contract.

## Rejected concerns

Investigated deliberately and found sound. Recorded so a later reviewer does not repeat
the work.

- **Resolver parity.** Tested `add-retry` (ambiguous), `ADD-RETRY-ALPHA` (case-insensitive
  exact slug), `6g0000000a` and `6g0000000a0` (ambiguous ID prefixes), `retry-alpha`
  (substring), `6G0000000A01` (case-insensitive exact ID), duplicate slug `dup`,
  `../zeta`, `""`, an unreadable file by ID and by filename slug. `task show` and
  `task blockers` produce identical results and identical error strings in every case
  except L1.
- **No Store re-entry, no pre-lock TOCTOU.** Planners receive only `*TaskGraph`;
  `ResolveTaskID` is pure; `MutateTaskGraph` re-reads and whole-snapshot-CASes
  (`SameSourceSnapshot`) before the first write and per-file-CASes before each later one.
- **Concurrent opposite edges, real processes.** Two `tskflwctl task depend add`
  processes racing `A→B` and `B→A`: one exit 0, one exit 11 with
  `planned write prefix … would leave a broken graph: dependency cycle …`, final graph
  acyclic with exactly one edge.
- **Interrupted migration recovery.** Forced a genuine `EPERM` on the second rename.
  Result: durable prefix `["6g0000000b02"]`, remaining `["6g0000000a01"]`, both present in
  `error.dependency_mutation` with workspace identity; human text "durable dependency
  prefix applied to … retry the same command to converge remaining tasks …"; graph
  degraded not broken; no `.tskflwctl-*.tmp` left behind; rerun converged to healthy.
  The all-applied variant produces the distinct "all planned dependency task files were
  durably applied" wording.
- **Idempotence and the clock.** `task depend add` of a present edge and `remove` of an
  absent edge are byte-identical (`bytes.Equal` asserted in
  `internal/cli/task_dependency_test.go:97,126`) and do not advance `updated_at`;
  `materializeTaskGraphPlan` compares content *before* stamping, so the stamp cannot be
  the only change. `s.now()` is taken once per operation, outside the retry loop.
- **`task unblocks` excludes its source even in a cycle.** `seen` is seeded with the
  queried ID (dependency_graph.go:1046), so it can never be re-enqueued;
  `TestTaskGraphCycleBlockerReason` pins it.
- **`task list --unblocked` fails closed.** Errors with `ErrValidation` on both `degraded`
  and `broken`, and `State().Eligible` independently requires `health == healthy`, so the
  filter is doubly gated.
- **Generic paths cannot bypass graph ownership.** `core.SetFields` (service_task.go:273,
  before the `--force` branch), `store.FS.SetFields` (fsstore.go:281), `CreateTask`,
  `EditTask`'s sorted before/after comparison including the malformed-frontmatter carve
  (edit.go:160-177), and `fix.go:349` all refuse. Reordering `depends_on` in the editor is
  correctly still allowed; duplicating an entry is not.
- **Nil-versus-empty and schema fidelity.** Every new array is built with
  `make(…, 0, n)` / `append([]string{}, …)`, so empty renders `[]`, never `null`; all are
  `required` in the reflected schema. `envelopes_test.go` validates a `broken` blockers
  envelope carrying problems and legacy diagnostics against the schema. `GraphProblemJSON.Cycle`
  is the only `omitempty` array and is genuinely optional.
- **Contract churn.** All 22 modified goldens change only `schema_version` 1.49→1.50;
  three new goldens added; `go run ./internal/tools/docgen` reproduces `docs/cli/`
  byte-identically; `go mod tidy -diff` clean; `git diff --check` clean.
- **Boundary hygiene.** No graph-library type crosses core/CLI/wire; the analyzer is
  taskflow-owned; `TaskDependencyWrite` names only task IDs and a canonical set.
- **Absolute paths in `problems[].path`.** Matches the existing envelope convention
  (`PathEnvelope`, `domain.FileProblem`), not new leakage.
- **Migration semantics on the real tree.** Six legacy `blocked_by` occurrences became
  eight canonical edges with direction preserved: `6ffr4wc01gtv→6fgq1n00235z`,
  `6fgq1n0006y3→{6fgq1n000pca,6fgq1n0016kj,6fgq1n003wty}`,
  `6fgq1n000pca/6fgq1n0016kj/6fgq1n003wty→6fgq1n002skz`, `6fgq1n00235z→6fgq1n0006y3`.
  Re-running `task depend migrate --dry-run` on the live tree now reports
  `changed:false`. The only remaining `dependencies:` matches in `planning/` are inside a
  fenced manifest example in the spike task's body.
- **Bootstrap sequence direction.** `6g3q4rt7mgjn` depends on both completed foundations;
  the five later slices form one chain, each depending on its predecessor. `task unblocks
  6g3q4rt7mgjn` returns all five with correct shortest paths and `queued/blocked` state.
- **Already-canonical and doubly-declared legacy edges.** A `blocks` declaration whose
  canonical edge already exists produces an `outcome:"skipped"` edge and only the
  clear-owner is written; a task declaring the same edge through `blocked_by` *and*
  `dependencies` produces one edge and two cleared fields.
- **Lock-held validation cost.** `ValidateTaskGraphMutationPlan` rebuilds the whole graph
  per planned write, i.e. O(k·(V+E)) under the repository lock. Negligible at k≤8 today,
  and the 2026-08-26 ADR amendment already commits to benchmarking this before the
  bulk-linking slice; not re-raised here.
- **Duplicate `--on` operands.** `--on tb --on 6g000000000b` is a hard validation error
  rather than a dedup. That is the specified behaviour ("validate … duplicate/self/missing
  edges") and reads correctly.

## Traceability against acceptance criteria

| # | Criterion (abbreviated) | Verdict | Evidence |
|---|---|---|---|
| 1 | Snapshot-local resolution matches ordinary resolution, no Store call | **Partial** | Byte-identical across 10 reference shapes; duplicate-ID ambiguity diverges (L1) |
| 2 | Add/remove validate endpoints, duplicate/self/missing, union, cycles, prefixes, final health | **Met** | Self/duplicate/ambiguous/cycle all rejected; concurrent opposite edges refused |
| 3 | Present adds and absent removals are receipt-bearing no-ops, no byte or timestamp change | **Met** | `bytes.Equal` assertions plus reproduced by hand |
| 4 | `--dry-run` runs the same resolution/planning/validation with zero replacements | **Met** | Dry-run receipt carries plan; no file written; degraded/broken still rejected |
| 5 | `migrate` converts the six live legacy occurrences, preserving frontmatter/body | **Met** | Real tree: 6 occurrences → 8 edges; comments and bodies preserved; rerun is a no-op |
| 6 | Migration graph-sound after every injected write failure; retry converges | **Partial** | True in behaviour; the load-bearing `blocks` ordering case is untested (H2) |
| 7 | Receipts distinguish applied/skipped; partial-failure diagnostics preserve applied and remaining | **Partial** | Applied/remaining correct and structured; cleared-field list under-reports (L4) |
| 8 | `blockers` defaults to frontier, `--causal` full closure, deterministic reason/path/health | **Partial** | Correct on healthy graphs; both projections empty on degraded (H1) |
| 9 | `unblocks` reports all transitive dependents without counterfactual claims | **Partial** | Correct and non-counterfactual on healthy graphs; misses legacy edges on degraded (H1) |
| 10 | `task list --unblocked` filters derived eligibility, no work on an unsound graph | **Met** | Fails closed on degraded and broken; `Eligible` independently gated on health |
| 11 | `task set --force` and `task edit` cannot bypass graph ownership | **Met** | Rejected on both spellings, both paths, plus `CreateTask` and `lint --fix` |
| 12 | Envelopes enter the reflected schema, human/JSON parity, no graph-library types | **Met** | Schema `$defs` present and `required`; broken-state envelope schema-validated |
| 13 | Semantic changes use the Service clock exactly once; no-ops do not advance `updated_at` | **Met** | Clock taken once outside the retry loop; content compared before stamping |
| 14 | Deep-chain stress establishes the supported graph-depth envelope | **Partial** | 4,096 chain covers `State`/`BlockingFrontier` only; the amplifying queries are unmeasured (L2) |
| 15 | Production commands persist bootstrap edges and exercise the queries on the real graph | **Met** | Bootstrap chain correct; live `blockers`/`unblocks`/`--unblocked` all verified |

## Validation commands and results

```
GOCACHE=/tmp/taskflow-review-go-cache go test ./...
  → ok, 23 packages, 0 failures

GOCACHE=/tmp/taskflow-review-go-cache go test -race ./...
  → ok, 23 packages, 0 failures, 0 data races

GOLANGCI_LINT_CACHE=/tmp/taskflow-review-lint-cache golangci-lint run ./...
  → 0 issues

git diff --check
  → clean

go mod tidy -diff
  → clean

go run ./internal/tools/docgen -out /tmp/docgen-check && diff -rq /tmp/docgen-check docs/cli
  → identical (generated CLI reference in sync)

tskflwctl task depend migrate --dry-run --json      # real planning tree, read-only
  → changed:false, no legacy occurrences remain

mutation test: internal/core/dependency_operations.go:323  `>` → `<`   (copy outside the repo)
  → go test ./...  ok, 23 packages, 0 failures     ← the reversal is undetected (H2)
```

## Candidate tasks

Disposition after implementation hardening:

- ✅ `6g4g8gatbnrs` — guarded repair of an already-broken graph is separately scoped with a monotonic problem-reduction proof (M2).
- ✅ Resolved legacy projection and empty-key convergence landed in the reviewed slice rather than creating a redundant follow-up (H1, L3, L4).
- ✅ `6g3q4rte8kc1` now owns the shared before/after graph-state receipt shape and post-dependency-mutation state warning (L5).
- ✅ Repeated human cycle lines are deduplicated here; repository-wide structured problems remain deliberate. Terse query output is deferred until dogfooding establishes the needed ID/slug contract, so no speculative task was created (L6, L7).
