---
schema: 1
id: 6g41amrnje2j
bucket: closed
area: canonical-task-dependency-read-foundation-claude
date: "2026-08-26"
---

# Audit: canonical-task-dependency-read-foundation-claude — 2026-08-26

> Edit findings in place and flip each `**Status:**` as you work it.

Adversarial implementation review of the canonical task-dependency read foundation
(task `6g3q4rst78qy`, epic `30-threads-and-task-dependency-graphs`, branch
`feat/canonical-task-dependency-reads`, uncommitted working tree vs `main` at
`43b3044`). Every claim below was checked against code and executed against the
compiled package; counterexamples were constructed with `go test -overlay` so that
no production or test file in the repository was modified. This audit is a second,
independent pass — a sibling audit `6g417v97bx8s` exists for the same slice; where
we agree that is noted as corroboration, and where its "survived adversarial
review" list is wrong that is stated explicitly.

## Executive verdict

**Not ready as-is; safe with amendments.** The architecture is right and the
fail-closed posture genuinely holds: no exercised path let an unsound graph state
reach a supported command, `-race` is clean, snapshot immutability is real, and the
blocker path projection is not merely deterministic but lexicographically minimal
and insertion-order independent. Sound completion really is memoized once per task.

But three classes of defect should not merge into a foundation that seven downstream
tasks will build on:

1. **Two exported read APIs are misleading in the exact way a future authorization
   check will consume them.** `Blockers()` returns an empty slice for a task whose
   own record is hard-broken (H2), and `TopologicalWaves()` returns
   `complete = true` over a silently truncated edge set on a `broken` snapshot (H3).
   Both are "no news" answers that mean "damaged", and the natural slice-3 predicate
   (`len(Blockers(id)) == 0` → allow start) reads them as permission.
2. **Two write paths still change graph-owned fields.** `lint --fix` normalizes
   `depends_on` at the text level and can *create* edges that did not exist
   (H4, verified end-to-end), and the `task edit` guard is inverted on malformed
   frontmatter: repairing a file while *preserving* its dependency is rejected while
   repairing it by *deleting* the dependency succeeds silently (H5).
3. **The legacy edge plan slice 2 is meant to consume is not validated.** Legacy
   references that resolve to a self-edge or to a mutual cycle are reported as
   cleanly `resolved` at `degraded` health (M1).

Cycle diagnostics are also incomplete (H1) — corroborating the sibling audit's H1
with an executable counterexample, and adding that lint attributes a cycle to
exactly one member (M4).

The overbuilt/underbuilt call: the `DAGAnalyzer` seam is **overbuilt** for what
shipped (one implementation, one three-case contract, and the comparison adapter is
not in the repository at all — L2), while cycle attribution, blocker honesty, and
the legacy plan's own validity are **underbuilt** relative to what ADR-0006 promises.
Nothing here is founded on an incorrect assumption; the `depends_on`-as-stable-ID-set
model and the strict/resilient split are sound and worth keeping.

Findings are classified as **[merge-blocking]**, **[pre-mutation prerequisite]**, or
**[tracked follow-up]**.

---

## Findings

#### H1. `deterministicCycles` misses cycle members, so `blocker.Reason` lies about tasks that can never start  · **Status:** fixed 2026-08-27

**File:** internal/core/dependency_graph.go:806-821 | **Component:** core / graph analysis
**Effort:** M · **Urgency:** acute

**[pre-mutation prerequisite]**

The DFS at `dependency_graph.go:806-821` handles only `state[dependent] == 0`
(unvisited) and `== 1` (on stack). There is no `case 2`: an edge into an already
finished node is dropped. Every *cycle* still yields at least one back edge, but not
every *node on a cycle* is reached by one, so `g.cycleMembers` is incomplete.

Executed counterexample (three tasks, prerequisite→dependent edges
`A→B, A→C, B→A, C→B`, i.e. `A.depends_on=[B]`, `B.depends_on=[A,C]`,
`C.depends_on=[A]`, IDs chosen so `A < B < C`):

```
health=broken
problem code=cycle task=000000000001 cycle=[000000000001 000000000002 000000000001]
cycleMembers=map[000000000001:true 000000000002:true]
GAP: 000000000003 is on cycle A->C->B->A but is not a recorded cycle member
blocker: id=000000000003 reason=not-started path=[Z 000000000003]
```

`C` sits on the cycle `A→C→B→A` and is reported to consumers with reason
`not-started`. ADR-0006 (as amended by this slice) states that `cycle` is forensic
vocabulary that lets diagnostics "explain damage … without collapsing it into a
vague missing/blocked result" — this is precisely that collapse, in the opposite
direction: an unstartable task is described as merely unstarted.

Fail-closed is preserved (`gate(C)` still returns `broken`, because `C`'s
prerequisite `A` *is* a flagged cycle member and `computeSound` propagates
`broken`), so this is a diagnostic-honesty defect, not a safety hole. That is why it
is a pre-mutation prerequisite rather than merge-blocking.

**Why current tests miss it:** `TestTaskGraphCycleBlockerReason`
(`dependency_graph_test.go:249-264`) uses a two-node cycle, the one shape where a
single back edge covers every member. Worse, its assertion body is
`if (blocker.TaskID == a.ID || blocker.TaskID == b.ID) && blocker.Reason != BlockerCycle`
inside a `range` — it passes vacuously if `Blockers` returns nothing at all.
`dagcontract.Run`'s "attributable cycle" case is a single 3-cycle with no chords.

**Recommendation:** replace the ad-hoc DFS with Tarjan SCC to mark membership
(`cycleMembers` = every node in a non-trivial SCC, plus self-loops), and keep the
existing back-edge walk only to render one representative path per SCC. That is a
contained change inside `deterministicCycles` and does not widen the analyzer
surface.

#### H2. `Blockers()` reports nothing for a task whose own record is broken, and a lifecycle check will read that as permission  · **Status:** fixed 2026-08-27

**File:** internal/core/dependency_graph.go:646-698 | **Component:** core / graph read API
**Effort:** S · **Urgency:** acute

**[merge-blocking]**

`Blockers` explains *prerequisites*. It never inspects the queried task's own
validity, so a task that is `hardBroken` for a reason local to itself — invalid
status, duplicate `depends_on` entry, ID drift, missing frontmatter ID — returns an
empty slice:

```
state(B)={TaskID:…02 Role:unknown Gate:broken …}   Blockers(B)=[] (len=0)
state(C)={TaskID:…03 Role:candidate Gate:broken …} Blockers(C)=[] (len=0)
```

(`B` has status `nonsense`; `C` has `depends_on: [A, A]`. Both are `GateBroken`.)

**Realistic failure:** epic 30's next slices are
`enforce-dependency-eligibility-across-every-task-start-path` and
`ship-guarded-dependency-mutations-and-graph-queries`. The obvious implementation of
"may this task start?" is `blockers := graph.Blockers(id); if len(blockers) == 0 { … }`,
and `task why-blocked <slug>` is the obvious CLI. Both would authorize / report
"nothing blocking" for a task the snapshot has already classified as broken. The
correct predicate exists (`State(id).Gate == GateClear`, or
`state.Eligible`), but nothing in the type system, the doc comment
("returns every reachable unsound prerequisite"), or the tests steers a caller to it.

**Why current tests miss it:** every `Blockers` test queries a *well-formed* task
whose prerequisites are the damaged ones. No test queries a hard-broken task.

**Recommendation:** smallest fix — emit a self-blocker
(`Blocker{TaskID: taskID, Reason: <local reason>, Path: []string{taskID}, Direct: true}`)
when `g.hardBroken[taskID]` or the task's own status is invalid, so an empty slice
means exactly one thing. Alternatively rename to `PrerequisiteBlockers` and document
that `State().Gate` is the authorization predicate. Do this before the API is
consumed, not after.

#### H3. `TopologicalWaves()` returns `complete = true` on a `broken` snapshot, over an edge set that silently dropped the broken edges  · **Status:** fixed 2026-08-27

**File:** internal/core/dependency_graph.go:331-366, 758-760 | **Component:** core / graph read API
**Effort:** S · **Urgency:** acute

**[merge-blocking]**

Edges are appended to the analyzer input only in the `default` arm of the switch at
`dependency_graph.go:331-357`; duplicate, invalid-ID, and missing prerequisites are
recorded as problems and then *omitted* from `edges`. `Analyze` therefore sees a
strictly smaller graph, finds no cycle, and returns `TopologicalComplete: true`.
`TopologicalWaves()` hands that straight out with no health qualifier:

```
cyclic snapshot:      waves=[]                  complete=false health=broken
broken-but-acyclic:   waves=[[…0a …0b]]         complete=true  health=broken
```

In the second case task `…0a` declares `depends_on: [<missing id>]`. The returned
wave plan places it in wave 0 alongside an unrelated task, and the boolean the API
offers as its trust signal says the ordering is complete.

**Realistic failure:** `generate-deterministic-thread-graph-views` and
`add-usage-informed-thread-views-to-the-tui` are the natural consumers of waves. A
rendered Thread diagram on a partially migrated or hand-edited repository would show
a confidently wrong ordering, with the missing prerequisite invisible rather than
marked.

**Why current tests miss it:** `TestTaskGraphTopologicalWavesAndDownstream` uses a
fully healthy graph. No test calls `TopologicalWaves` on a snapshot with a
non-cycle problem.

**Recommendation:** return `complete = false` whenever `g.health != GraphHealthy`
(one line at `dependency_graph.go:758`), or drop the bool and require callers to
check `Health()` first. `complete` must never mean "no cycle among the edges I
happened to keep".

#### H4. `lint --fix` rewrites `depends_on` and can create dependency edges through an unguarded path  · **Status:** fixed 2026-08-27

**File:** internal/store/fix.go:353; internal/domain/fields.go:33 | **Component:** store / repair
**Effort:** S · **Urgency:** acute

**[merge-blocking]**

Registering `depends_on` as a `"list"` field (`fields.go:33`) enrols it in
`domain.IsListField`, which is exactly the predicate `fixValue` uses at
`fix.go:353` to rewrite a scalar into a YAML flow list. `FixFrontmatter` consults
`domain.IsGraphOwnedTaskField` nowhere. Verified end-to-end on a scratch planning
repo:

```
$ tskflwctl -C $R lint
! …/tasks/000000000001-alpha.md
    validation failed: malformed frontmatter: field "depends_on" must be a YAML list,
    but it is a string ("000000000002, 000000000003")

$ tskflwctl -C $R task show 000000000001 --json     # before: unreadable, ZERO edges
{"schema_version":"1.49","error":{"code":"validation",…}}

$ tskflwctl -C $R lint --fix
  - depends_on: normalized to a YAML list

$ tskflwctl -C $R task show 000000000001 --json     # after: TWO edges now exist
{…,"depends_on":["000000000002","000000000003"]}
```

`lint --fix` therefore performed a repository-global graph mutation — with no
referential check, no self-edge check, and no cycle check — via the one write path
the slice never guarded. The same mechanism applies to the legacy `dependencies` and
`blocks` fields (both already `"list"`), though not to `blocked_by`, which is
registered as `"list"` too but is the field this repo actually uses. This is also
the *only* CLI way to touch these fields today, which interacts badly with M2.

**Why current tests miss it:** `fix_test.go` predates the field and tests
normalization on `tags`/`dependencies` as generic list behaviour, never as a graph
write. No test asserts that `FixFrontmatter` leaves graph-owned fields alone.

**Recommendation:** exclude `domain.IsGraphOwnedTaskField(key)` from the
list-normalizing branch in `fixValue`, and have `FixFrontmatter` report those files
as `Skipped` with the reason, exactly as `repairInvalidID` already does for a
referenced id. If normalization is genuinely wanted, it belongs in the guarded
migration where the result can be validated.

#### H5. The `task edit` dependency guard is inverted on malformed frontmatter: preserving an edge is rejected, deleting it is accepted  · **Status:** fixed 2026-08-27

**File:** internal/store/edit.go:158-178, 228-252 | **Component:** store / interactive edit
**Effort:** M · **Urgency:** acute

**[pre-mutation prerequisite]**

`dependencyValues` (`edit.go:228`) recovers the pre-edit baseline with a narrow YAML
decode. Its doc comment claims this "can recover the graph baseline even when an
unrelated typed field (for example tier) is malformed" — true, and a good idea. But
the narrow struct still declares `DependsOn []string`, so it is defeated by
malformation *of the guarded field itself*, which is the case that matters. When
`dependenciesReadable` is false, `edit.go:167-176` accepts the edit only if the
result has **no** dependency fields at all.

Verified against `store.FS.EditTask` directly (the CLI requires a TTY):

```
dependencyValues readable=false fields={dependsOn:[] …}
repair-keeping-deps:  changed=false err=validation failed: cannot verify the original
                      graph-owned fields while repairing malformed frontmatter…
repair-deleting-deps: changed=true  err=<nil>   <-- edge silently deleted via task edit
resulting DependsOn=[]
```

So on a file with `depends_on: A, B` (scalar, the realistic hand-edit): repairing it
into `depends_on: [A, B]` is **refused**, and repairing it by **deleting the line
entirely is accepted with no diagnostic**. The guard permits exactly the destructive
outcome and forbids the corrective one. The error text ("repair them directly, run
lint") then points the operator at a raw filesystem edit — the behaviour the slice
set out to discourage — or at `lint --fix`, which is H4.

**Why current tests miss it:**
`TestEditTaskRejectsDependencyDeltaButAllowsReordering` and its legacy twin only
exercise well-formed originals, where `dependenciesReadable` is always true. The
`!dependenciesReadable` branch has no test at all.

**Recommendation:** make the narrow decode tolerant — decode into
`map[string]yaml.Node` (or `[]string`-or-`string`) so a scalar/duplicated value is
still captured verbatim as the baseline, and compare raw text when it cannot be
typed. Failing that, invert the escape hatch: accept an edit that leaves the
dependency text **byte-identical**, and reject one that removes it.

#### M1. The resolved legacy edge plan is never validated — self-edges and mutual cycles resolve "clean" at `degraded` health  · **Status:** fixed 2026-08-27

**File:** internal/core/dependency_graph.go:436-499 | **Component:** core / legacy migration diagnosis
**Effort:** M · **Urgency:** soon

**[pre-mutation prerequisite]**

`resolveLegacyDiagnostics` classifies each reference as `resolved` / `missing` /
`ambiguous` purely on *cardinality of the slug/ID match* (`dependency_graph.go:475-491`).
Nothing checks whether the resulting `ref.Edge` is legal, or whether the union of all
resolved edges with the existing canonical edges is still a DAG. Verified:

```
# legacy blocked_by naming the task's own slug
health=degraded problems=0
legacy blocked_by ref="alpha" resolution=resolved edge={From:000000000001 To:000000000001}

# two tasks each blocked_by the other's slug
health=degraded problems=0
legacy blocked_by on …01 ref="beta"  -> resolved edge=…02->…01
legacy blocked_by on …02 ref="alpha" -> resolved edge=…01->…02
```

Both snapshots report zero problems and `degraded` — the health state whose stated
meaning is "every legacy reference resolves exactly but has not yet been migrated".

**Realistic failure:** `6g3q4rt7mgjn` is specified to consume this plan. A guarded
migration that trusts `Resolution == LegacyResolved` and writes
`ref.Edge` would produce a self-dependency or a cycle in one atomic batch, taking the
repository from `degraded` straight to `broken` — and, because migration is a bulk
write, potentially after an arbitrary write prefix has already landed.

**Why current tests miss it:**
`TestTaskGraphLegacyResolutionHealthAndDirection` uses one prerequisite and one
dependent with no feedback edge;
`TestTaskGraphLegacyMissingAndAmbiguousAreBroken` covers only cardinality 0 and >1.

**Recommendation:** after building `g.legacy`, run the *projected* edge set
(canonical edges ∪ resolved legacy edges) through the same `analyzer.Analyze`, and
downgrade to `broken` with a distinct problem code (e.g.
`legacy-reference-unsafe`) when the projection self-loops or cycles. This is the
cheapest possible pre-flight for slice 2 and reuses machinery already present.

#### M2. Ordinary `lint` now fails permanently on this repository, and the guard removed the only CLI way to fix it  · **Status:** fixed 2026-08-27

**File:** internal/core/service.go:327-380 | **Component:** core / lint · operations
**Effort:** S · **Urgency:** acute

**[merge-blocking — operational]**

Verified on the working tree:

```
$ ./bin/tskflwctl lint
… 6 item(s) with issues
error: validation failed: 6 item(s) with issues, 0 unreadable file(s)   # exit 11
```

All six are *resolvable* `blocked_by` occurrences — health is `degraded`, structural
problems are zero. This is deliberate and documented in the task's stress-test
criterion, but three consequences were not weighed:

1. `CLAUDE.md` states "Keep `planning/` lint-clean" as a standing invariant. This
   slice makes that impossible until `6g3q4rt7mgjn` ships, so the invariant and any
   CI gate built on it are now permanently red.
2. Once lint is expected to fail, it stops functioning as a signal. A genuine new
   defect lands inside a known-failing command whose exit code nobody reads.
3. The same slice made `task set`, `task set --force`, `task set --unset` and
   `task edit` refuse these fields (correctly). The remaining CLI path is
   `lint --fix`, which is H4 — an unguarded text rewrite. The operator's only
   sanctioned option is a raw filesystem edit.

This corroborates the sibling audit's M1; the CLAUDE.md conflict and the
`lint --fix` interaction are additional.

**Recommendation:** make a *fully resolved* legacy diagnostic advisory — reported in
`lint` output and in `--json`, but not exit-code-bearing — until the migration lands,
and keep exit 11 for `legacy-reference-missing` / `legacy-reference-ambiguous`, which
are genuinely actionable today. That preserves the diagnostic without spending the
repository's lint-clean invariant on a state the tool refuses to let anyone fix.

#### M3. `Blockers` is superlinear in chain depth and makes the lint pass roughly cubic; the "O(V+E)" acceptance criterion over-claims  · **Status:** fixed 2026-08-27

**File:** internal/core/dependency_graph.go:669; internal/core/service.go:346-360 | **Component:** core / performance
**Effort:** M · **Urgency:** soon

**[pre-mutation prerequisite]**

`dependency_graph.go:669` copies the whole accumulated path for every edge examined
(`path := append(append([]string(nil), current.path...), prerequisite)`), so one
`Blockers` call over a chain of depth *d* does Θ(d²) work even when it returns no
blockers. `dependencyLintIssues` (`service.go:346`) then calls `Blockers` once per
`Inconsistent` task, which on a reopened chain is every task on it.

Measured (single reopened thread, all downstream tasks `completed`):

| chain depth | `dependencyLintIssues` |
|---|---|
| 100 | 10 ms |
| 200 | 40 ms |
| 300 | 84 ms |
| 500 | 345 ms |
| 1000 | 3.0 s |
| 2000 | **84 s** |

`Blockers` alone: n=1000 → 6.4 ms, n=8000 → 139 ms.

To be fair about when this bites: it is driven by **depth, not task count**. 2000
tasks arranged as 200 threads of depth 10 cost **8 ms**, and this repository's
current `lint` over 279 tasks takes **0.09 s**. The shape that hurts is one long
Thread whose root is reopened — which is both a listed future failure scenario and
the literal purpose of epic 30.

The acceptance criterion "Sound completion and derived graph state memoize one
result per task per snapshot and are O(V+E)" is true of `computeSound`
(independently confirmed) but not of the blocker projection, which is inherently
Ω(Σ path lengths) because ADR-0006 requires a path per blocker.

**Why current tests miss it:**
`TestTaskGraphSoundCompletionMemoizesReconvergentDiamonds` asserts visit counts for
`computeSound` only, and the deep-chain test
(`TestOwnedDAGAnalyzerDeepWideAndDisconnected`) exercises `Analyze`, never
`Blockers`. There is no benchmark and no test that calls `dependencyLintIssues` at
scale.

**Recommendation:** store a parent pointer per visited node and materialize the path
once per emitted blocker, not per edge. Then amend the acceptance criterion (or
ADR-0006) to say "state derivation is O(V+E); blocker path projection is O(V + Σ
path lengths)" rather than leaving an unachievable claim ticked.

#### M4. A cycle is attributed to exactly one task; every other member is unattributed in `lint`  · **Status:** fixed 2026-08-27

**File:** internal/core/dependency_graph.go:374; internal/core/service.go:340-346 | **Component:** core / lint attribution
**Effort:** S · **Urgency:** soon

**[pre-mutation prerequisite]**

`GraphProblem{Code: ProblemCycle, TaskID: cycle[0], …}` records one problem per
cycle, keyed to the lexicographically smallest member.
`dependencyLintIssues` buckets by `problem.TaskID`, so `lint` prints the cycle under
one slug only. Verified on a plain two-node cycle plus the H1 chorded shape:

```
lint issue for 000000000001: [depends_on] dependency cycle: …01 -> …02 -> …01
GAP: task 000000000002 participates in a cycle but ordinary lint attributes nothing to it
GAP: task 000000000003 participates in a cycle but ordinary lint attributes nothing to it
```

ADR-0006 as amended by this slice requires that "every structural defect and legacy
occurrence must also appear during an ordinary `lint` call with deterministic
attribution". An operator inspecting the *other* task in the cycle — or a filter
like `lint --json` keyed by slug — sees a clean record. Combined with H1, a cycle
member can be both unattributed *and* mislabelled `not-started`.

**Recommendation:** emit one `ProblemCycle` per member (same `Cycle` payload, each
with its own `TaskID` and `Path`), and let the message name the representative path.
Deduplication for human output belongs in the renderer, not in the attribution.

#### M5. A duplicate stable task ID silently discards the second file's edges and mirrors the first file's issues onto both slugs  · **Status:** fixed 2026-08-27

**File:** internal/core/dependency_graph.go:302-307; internal/core/service.go:272 | **Component:** core / identity
**Effort:** M · **Urgency:** soon

**[pre-mutation prerequisite]**

On a duplicate id the second record is dropped with `continue`
(`dependency_graph.go:306`) — its `depends_on` never becomes edges and its own
defects are never diagnosed. Meanwhile `service.go:272` appends
`graphIssues[canonicalTaskID(t)]` for **every** task, and both files share that key,
so the first file's issues are reprinted under the second file's slug. Verified:

```
problem duplicate-task-id  task=…01 path=tasks/…01-second.md
problem missing-dependency task=…01 path=tasks/…01-first.md
dependencies[…01] = [000000000009]      # second file's …07 edge is gone
lint issue attributed to BOTH slugs: [id]         duplicate stable task id …
lint issue attributed to BOTH slugs: [depends_on] task …01 depends on missing task …09
```

`second` is told it depends on a missing task it never referenced, and its real
reference to `…07` is reported nowhere. This is simultaneously the "silently
dropping" and "misleadingly duplicating" failure the review set out to test for.
Duplicate task ids are not hypothetical: `domain.DuplicateIDIssues` exists precisely
because they occur, and it is currently wired to research only, not tasks.

**Recommendation:** two small changes — (a) key `graphIssues` by file path rather
than by task id in `service.go`, so a duplicate's issues land on the file that owns
them; (b) keep the dropped record in a `duplicates []domain.Task` side list and emit
its edge-level problems too, so the operator sees both files' claims before choosing
which to rename.

#### M6. Unreadable task files carry no identity, so a blocker on one is reported as `missing`  · **Status:** fixed 2026-08-27

**File:** internal/core/dependency_graph.go:269-273; internal/core/service.go:333 | **Component:** core / diagnostics
**Effort:** S · **Urgency:** soon

**[tracked follow-up]**

`ProblemUnreadable` is constructed with `Path` only; `TaskID` stays empty even though
task filenames are id-led (`tasks/<id>-<slug>.md`) and `store.splitFlatName` already
recovers the id. `dependencyLintIssues` then skips the code entirely
(`service.go:333`), so the two halves are never joined:

```
problem unreadable-task    task=""     path=tasks/000000000002-broken.md
problem missing-dependency task="…01"  msg=task …01 depends on missing task 000000000002
blocker id=000000000002 reason=missing
lint issues for A: [{Field:depends_on Message:task …01 depends on missing task …02}]
```

The prerequisite is not missing — its file is right there and unparseable. A human
can join the two lines by eye because the id leads the filename; a machine consumer,
and any Thread view, cannot. Corroborates the sibling audit's M2; the concrete
remedy below is additional.

**Recommendation:** populate `GraphProblem.TaskID` from the filename in the
`unreadable` loop, register those ids in a `g.unreadable` set, and add a
`BlockerUnreadable` reason so `blockerReason` distinguishes "file absent" from "file
present but unparseable". Both are broken; only one is fixed by re-creating the task.

#### M7. `Blockers` recurses through a withdrawn prerequisite and attributes its upstream to the querying task  · **Status:** fixed 2026-08-27

**File:** internal/core/dependency_graph.go:678-687 | **Component:** core / graph read API
**Effort:** S · **Urgency:** eventually

**[tracked follow-up]**

Traversal continues past any node that exists, including one whose reason is
`withdrawn`. Verified:

```
blocker …0u reason=not-started path=[T …0w …0u] direct=false
blocker …0w reason=withdrawn   path=[T …0w]     direct=true
```

`T` depends on the deprecated `W`, which depends on `U`. Once `W` is withdrawn, `U`
is irrelevant to `T`: finishing `U` changes nothing. A "what must I finish to unblock
this?" surface would list work that cannot help. The same applies to recursing past
an `invalid-status` node.

**Recommendation:** stop traversal at a terminal blocker (`withdrawn`,
`invalid-status`, `cycle`, `missing`) and keep only the direct one, or add a
`Terminal bool` to `Blocker` so consumers can prune. Decide before slice 3 consumes
the list.

#### L1. `invalid-dependency-id` has a problem code but no blocker reason, so it degrades to `missing`  · **Status:** fixed 2026-08-27

**File:** internal/core/dependency_graph.go:701-705 | **Component:** core / vocabulary
**Effort:** XS · **Urgency:** eventually

**[tracked follow-up]**

`blockerReason` returns `BlockerMissing` for anything absent from `g.tasks`, which
includes values that are not stable ids at all:

```
blocker id="not-a-stable-id" reason=missing path=[000000000001 not-a-stable-id]
```

The snapshot already knows better — it emitted `ProblemInvalidDependencyID` for the
same value. The blocker vocabulary the ADR now pins (eight tokens) has no way to say
so. **Recommendation:** add `BlockerInvalidReference` and check `id.Valid` in
`blockerReason` before falling back to `missing`; amend the ADR's token list in the
same change.

#### L2. The "shared contract" exercises one implementation, and lives where `vet`/`golangci-lint` cannot see it  · **Status:** fixed 2026-08-27

**File:** internal/core/dependency_graph_contract_test.go:10; internal/core/testdata/dagcontract/contract.go | **Component:** testing / architecture
**Effort:** S · **Urgency:** eventually

**[tracked follow-up]**

The acceptance criterion "One shared contract suite covers the owned analyzer and
the bounded library adapter" is ticked, but:

- `dagcontract.Run` has exactly one call site, with `core.OwnedDAGAnalyzer{}`.
- `dominikbraun/graph` appears nowhere in `go.mod`, and the implementation record
  states the adapter "remained under `/tmp`". The comparison is therefore not
  reproducible, not re-runnable when the analyzer changes, and not auditable.
- The contract itself is three cases: a 4-node diamond, an unchorded 3-cycle, and a
  self-loop. It does not cover multiple cycles, chorded cycles (H1's shape), edges to
  unknown nodes, duplicate edges, or empty input.
- `go list ./... | grep -c dagcontract` → `0`. A package under `testdata/` is
  excluded from `./...`, so `go vet ./...` and `golangci-lint run ./...` never
  analyze it. It is type-checked only as a transitive dependency of `core_test`.

The verdict to retain the owned analyzer is defensible on the reasoning given
(deterministic attributable cycle paths, owned wave derivation, owned tie-breaking,
pre-v1 upstream API). Nothing important is being reimplemented poorly, and deferring
critical-path/weighted features is correct for V1. The problem is that the
`DAGAnalyzer` seam and the `testdata` package are abstraction cost paid for a
comparison the repository cannot demonstrate.

**Recommendation:** either commit the adapter behind a build tag so the contract has
two implementations and the AC is true, or delete the `DAGAnalyzer` interface and the
`testdata` package and call `deterministicCycles` directly — then record the bake-off
as prose in the ADR. Meanwhile move the contract into `internal/core/dagcontract`
(no `testdata`) so it is linted, and grow it with the shapes above.

#### L3. `schema --json` advertises all four graph-owned fields with no machine-readable "not settable" marker  · **Status:** tracked by 6g3q4rt7mgjn

**File:** internal/domain/fields.go:33, 89-97 | **Component:** wire / agent contract
**Effort:** S · **Urgency:** eventually

**[tracked follow-up]**

`schema --json` `task_fields` now lists `depends_on`, `blocked_by`, `dependencies`,
and `blocks` alongside genuinely settable fields, while all four are rejected:

```
$ tskflwctl task set 6g3q4rst78qy --set depends_on=6fjangd7kvh0
error: validation failed: depends_on is graph-owned and cannot be changed with `task set` …
```

The 1.49 note states the contract "recognizes the persisted list while generic
mutation remains forbidden" — the recognizing half is machine-readable, the
forbidding half is not. In fairness there is precedent: `status` has always been in
`task_fields` and has never been settable, so the registry has never meant
"settable". But this slice quadruples the exception set, and `schema` is explicitly
the surface agents route on.

**Recommendation:** add a `writable` (or `owner: graph|lifecycle`) boolean to the
`task_fields` entries and bump to 1.50 when convenient. Low urgency; the error
message is clear and names the future command.

#### L4. `ReadTaskGraph()` has no production caller, and `Service.Lint` re-implements the scan it wraps  · **Status:** tracked by 6g3q4rt0wzkq

**File:** internal/core/service_task.go:84-90; internal/core/service.go:246-250 | **Component:** core / architecture
**Effort:** XS · **Urgency:** eventually

**[tracked follow-up]**

`ReadTaskGraph` is the documented snapshot factory, but the only production consumer
of the graph — `Service.Lint` — calls `s.store.ListTasks()` and `NewTaskGraph`
itself. `ReadTaskGraph` has zero callers and zero direct tests
(`grep -rn ReadTaskGraph internal` returns only its own definition), and
`TaskGraph.Task()` shows 0% coverage.

The core/store split is otherwise well prepared for slice 2: the graph is pure
in-memory analysis over `domain.Task` with no filesystem or lock dependency, so a
store-owned `WithGraphMutation` critical section can take the write lock, build a
snapshot, plan, and write without core learning about locking. That part needs no
undoing. **Recommendation:** have `Lint` call `ReadTaskGraph` (it needs
`[]TaskWithBody` for other checks, so extract the projection) so the seam is
exercised by the one command that uses it.

#### L5. `store.FS.CreateTask` serializes `depends_on` with no validation  · **Status:** fixed 2026-08-27

**File:** internal/store/create.go:83-87, 112 | **Component:** store / create
**Effort:** XS · **Urgency:** eventually

**[tracked follow-up]**

`taskFields` writes `t.DependsOn` sorted whenever it is non-empty, with no
`id.Valid` check, no existence check, and no self-edge check. No CLI reaches it
today (`task new` has no dependency flag and `NewTaskParams` has no such field), so
this is not currently exploitable — but the create path is exactly where
`task new --depends-on` will land, and the guard that exists on `SetFields` and
`EditTask` has no counterpart here. **Recommendation:** reject a non-empty
`DependsOn` in `CreateTask` until the guarded create arrives, mirroring
`FS.SetFields` (`fsstore.go:262-266`).

#### L6. Test-suite gaps that create false confidence  · **Status:** fixed 2026-08-27

**File:** internal/core/dependency_graph_test.go | **Component:** testing
**Effort:** M · **Urgency:** soon

**[tracked follow-up]**

Concrete weaknesses, beyond the per-finding notes above:

- **Vacuous pass.** `TestTaskGraphCycleBlockerReason:259-263` asserts inside a
  `range` over `Blockers`; an empty result passes. Given H2, that is not theoretical.
- **Containment instead of equality.** Both problem-code tests use
  `slices.Contains` over a `wantCodes` list (`:45-49`, `:81-85`). Neither asserts the
  exact set, the count, or the attribution, so a dropped, duplicated, or
  misattributed problem is invisible.
- **Unrepresentative ids.** `testutil.TaskID` hashes a seed into the Crockford
  alphabet, producing ids with no shared prefix. Real ids are time-ordered
  (`id.New`), so tasks minted in one session share a long prefix and their
  lexicographic order *is* creation order. Every tie-break and sort assertion is
  exercised against a distribution the product never produces.
- **White-box coupling.** `TestTaskGraphSoundCompletionMemoizesReconvergentDiamonds`
  reaches into the unexported `graph.soundVisits`, which is incremented only on cache
  miss — so `visits == 1` is close to tautological given the early return. It does
  demonstrate memoization; it does not demonstrate a complexity bound.
- **Untested seam.** No test injects a non-owned `DAGAnalyzer` into `newTaskGraph`,
  so the abstraction's only justification is untested (see L2).
- **Missing shapes.** No fuzz or property test over `NewTaskGraph`; no concurrency
  test on `TaskGraph` (the suite is `-race`-clean but nothing exercises parallel
  readers — I added one ad hoc and it passed); no benchmark; no test of
  `dependencyLintIssues` on a real repository fixture; no `-o csv` / TUI coverage of
  the new field; `TaskGraph.Task()` uncovered.

**Recommendation:** convert the two containment assertions to exact
`[]GraphProblem` comparisons, restructure the cycle-reason test to assert a
count first, add a `testutil.SequentialTaskID` helper for order-sensitive tests, and
add one property test (random DAG → `TopologicalWaves` respects every edge; random
graph with a planted cycle → every planted member is a `cycleMembers` entry). The last
one is the regression test for H1.

---

## Claims I tried to falsify and found sound

Each of these was attacked with a constructed counterexample or a direct
measurement, not read off the implementation record.

1. **`depends_on` really is a duplicate-free stable-ID set at the boundaries.**
   The reader deliberately retains raw duplicates (`store` round-trip preserves
   `[second, first]` unsorted), `CreateTask` writes sorted, `ToTaskJSON` sorts a
   *copy* and provably does not mutate the domain record, and the graph dedupes via
   `sortedUnique` while still detecting duplicates — the `seen` map at
   `dependency_graph.go:325` runs over the raw list, so assigning the deduped set
   first does not hide them. Verified.
2. **Generic mutation is genuinely closed.** All four fields, both spellings
   (`--set` and `--unset`), with and without `--force`, at both the core
   (`Service.SetFields`) and store (`FS.SetFields`) layers. Checked through the real
   CLI, not just the unit tests. The TUI inline editor uses an explicit field list
   that excludes them.
3. **Concurrent reads are safe.** 16 goroutines × 200 tasks × 8 accessors under
   `-race`: clean. The reason is structural, not accidental — `computeSound` is
   driven to completion during construction and is never reached again from an
   exported method, so `g.sound` is frozen; `Blockers`/`Downstream` guard their
   caches with `g.mu`; and returned slices are copies (I mutated a returned
   `Task().Tags` and the snapshot was unaffected).
4. **Blocker paths are shortest, deterministic, *and* lexicographically minimal.**
   Not just "stable for the happy shape": on a graph with two equal-length competing
   routes through different intermediates, 50 randomized input permutations produced
   byte-identical `[]Blocker`. This follows from FIFO BFS over sorted adjacency, which
   keeps each frontier in lexicographic path order.
5. **Legacy edge direction is right, including the `blocks` inversion.**
   `blocked_by`/`dependencies` produce `From: candidate, To: task`; `blocks` produces
   `From: task, To: candidate`. Exact-ID match takes precedence over slug match, and
   resolution is exact-only (no prefix/fuzzy), so it is deterministic. Duplicate
   slugs — legal in this repo — correctly go `ambiguous`.
6. **The one existing path that could change graph identity already fails closed.**
   `repairInvalidID` refuses to canonicalize a misspelled id that is referenced
   anywhere else in the planning tree, and its `referencesTo` substring scan covers
   `depends_on` values for free. `lint --fix` cannot orphan an inbound edge by
   renaming. (Contrast H4, which is the *content* path, not the identity path.)
7. **Reopen invalidation works without touching disk.** Flipping an upstream task
   back to `ready-to-start` immediately makes the completed downstream
   `SoundlyCompleted: false`, `Drained: false`, `Gate: blocked`, `Inconsistent: true`,
   with its frontmatter unchanged. Broken-over-blocked precedence holds; `deferred`
   is `parked`/blocked and `deprecated` is `withdrawn`/broken, as the ADR says.
8. **Deep graphs do not blow the stack.** A 100 000-node chain builds and answers
   correctly (`sound=true broken=false health=healthy`). The recursion in
   `computeSound` and `deterministicCycles` is bounded by depth and Go's growable
   stacks absorb it; the sibling audit's L3 is a real but very low-priority concern.
   Depth costs time (M3), not correctness.
9. **The 1.49 wire change is additive and correctly versioned.** `depends_on` is
   `omitempty`, sorted, present in the JSON Schema and the schema-comment map, and no
   existing field or envelope changed. A 1.48 validator will reject 1.49 payloads
   because `TaskJSON` is `additionalProperties: false` — that is inherent to any
   additive change under a closed schema and is exactly what the version bump is for.
10. **The omitted surfaces are consistent, not accidental.** `depends_on` is absent
    from `render.TaskColumns()`, so `-c depends_on` is unavailable in `-o table/csv`
    and in `--json -c` projections — but `tags` has always been absent for the same
    reason (list-valued fields are not string columns). Full `--json` carries it. This
    is a deliberate deferral matching existing precedent, not a gap.
11. **Multi-repository planning spaces are not broken by this design.** Edges are
    plain stable ids resolved within one planning root, and ids are globally unique by
    construction, so a planning space coordinating several implementation checkouts
    works unchanged. The real limitation is that an edge cannot *name* another
    planning space — but nothing in this slice forecloses adding a qualifier later,
    and adding one now would be speculative complexity. Correctly not built.
12. **The core/store split is the right preparation for slice 2.** The graph is pure
    analysis over `domain.Task` with no I/O and no locking, so the store-owned
    mutation critical section can wrap it without inverting the dependency. Nothing
    here will need undoing (see L4 for the one loose thread).

Two claims from the sibling audit's own "survived adversarial review" list do **not**
survive: its item 2 extends the O(V+E) finding to `Blockers` (falsified by M3), and
its item 3 states `task edit` "correctly permits formatting / order-only adjustments"
without qualifying the malformed-frontmatter branch (falsified by H5). Its item 6
covers legacy direction and cardinality correctly but does not test edge *legality*
(M1).

---

## Traceability table

| Finding | Severity | Classification | Destination |
|---|---|---|---|
| H1 cycle members under-reported | High | Fixed | `6g3q4rst78qy`: Tarjan SCC membership + representative-cycle tests |
| H2 `Blockers` silent on self-brokenness | High | Fixed | `ExplainGate` couples derived authorization state, local problems, and frontier |
| H3 `TopologicalWaves` complete on broken snapshot | High | Fixed | completeness is true only at healthy graph health |
| H4 `lint --fix` writes graph-owned fields | High | Fixed | graph-owned normalization is skipped noisily without writing |
| H5 `task edit` guard inverted on malformed YAML | High | Fixed | an unreadable graph baseline rejects every edited candidate |
| M1 legacy edge plan unvalidated | Medium | Fixed | canonical ∪ resolved-legacy SCC validation + ADR amendment |
| M2 permanent `lint` failure on this repo | Medium | Fixed | safe legacy debt is an advisory with exit zero; unsafe debt remains fatal |
| M3 blocker/lint cost cubic in chain depth | Medium | Fixed | parent-pointer traversal, action-frontier lint, honest output-sensitive contract |
| M4 cycle attributed to one member only | Medium | Fixed | one deterministic problem per SCC member |
| M5 duplicate id drops + mirrors issues | Medium | Fixed | every record is validated and lint attribution is keyed by source path |
| M6 unreadable files lose identity | Medium | Fixed | stable ID recovery from id-led filenames + `unreadable` blocker reason |
| M7 traversal past withdrawn prerequisites | Medium | Fixed | causal closure traverses; action frontier stops at withdrawn/terminal damage |
| L1 no `invalid-reference` blocker token | Low | Fixed | distinct `invalid-reference`, `unreadable`, and `invalid-task` reasons |
| L2 contract covers one implementation | Low | Fixed | speculative analyzer interface/testdata package removed; direct adversarial tests retained |
| L3 `schema` cannot express "not settable" | Low | Tracked | acceptance criterion on `6g3q4rt7mgjn` |
| L4 `ReadTaskGraph` unused, scan duplicated | Low | Tracked | acceptance criterion on `6g3q4rt0wzkq` |
| L5 `CreateTask` writes unvalidated edges | Low | Fixed | ordinary create rejects non-empty graph-owned fields |
| L6 test-suite gaps | Low | Fixed | non-vacuous exact SCC, projection, identity, repair, and legacy counterexamples |

ADR-0006 amendments required: **M1** (define `degraded` to require a *legal* projected
edge set), **M2** (state which legacy states are exit-code-bearing), **M3** (state the
blocker-path complexity honestly), **L1** (extend the blocker token list).

Current-task reopening recommended for: **H1–H5, M2, M4** (all in-scope defects of
`6g3q4rst78qy`, not new scope).

---

## Validation commands and results

Run from `/Users/andyeschbacher/git/andy-esch/taskflow-canonical-task-dependency-reads`
on branch `feat/canonical-task-dependency-reads` (uncommitted tree, merge-base
`43b3044`). All counterexamples were compiled with `go test -overlay=…` against files
held outside the repository, so no production or test file was modified — verified by
`git status --porcelain` before and after (48 entries both times; the only new entry
is this audit).

| Command | Result |
|---|---|
| `go build ./...` | pass |
| `go test ./...` | pass, 23 packages |
| `go test -race ./...` | **pass, exit 0**, no race reports |
| `just lint` (`golangci-lint run ./...`) | **0 issues** |
| `go vet ./...` | pass |
| `go list ./... \| grep -c dagcontract` | **0** — contract package excluded from `./...` (L2) |
| `./bin/tskflwctl lint` | **exit 11**, 6 legacy `blocked_by` issues, 0 unreadable (M2) |
| `time ./bin/tskflwctl lint` | 0.09 s over 279 tasks |
| `./bin/tskflwctl task set … --set/--unset depends_on\|blocked_by\|dependencies\|blocks [--force]` | all rejected, exit 11 (sound) |
| `go test ./internal/core -run 'TestTaskGraph\|TestOwnedDAG' -cover` | 26.1% package; `TaskGraph.Task` 0%, `TaskIDs` 0% |
| Overlay: cycle-member coverage counterexample | **GAP confirmed** — H1, M4 |
| Overlay: `Blockers` on hard-broken task | **`[]` returned** — H2 |
| Overlay: `TopologicalWaves` on broken-but-acyclic snapshot | **`complete=true`** — H3 |
| Overlay: legacy self-edge / mutual-cycle resolution | **`resolved`, `degraded`, 0 problems** — M1 |
| Overlay: duplicate task id | **second file's edges dropped; issues mirrored** — M5 |
| Overlay: unreadable file vs dependency id | **`TaskID=""`, blocker `missing`** — M6 |
| Overlay: `EditTask` on malformed frontmatter | **preserve rejected, delete accepted** — H5 |
| CLI: `lint --fix` on scalar `depends_on` in a scratch repo | **0 edges → 2 edges**, unguarded — H4 |
| Overlay: 50-permutation blocker determinism on competing equal-length paths | byte-identical (sound) |
| Overlay: 16-goroutine concurrent reader under `-race` | clean (sound) |
| Overlay: 100 000-node chain | builds, correct, no stack overflow (sound) |
| Overlay: `dependencyLintIssues` depth scaling | 100→10 ms · 500→345 ms · 1000→3.0 s · **2000→84 s** — M3 |
| Overlay: 2000 tasks / 200 threads of depth 10 | 8 ms (M3 is depth-driven, not size-driven) |

---

## Remediation disposition (2026-08-27)

The current slice absorbed every merge-blocking and pre-mutation correctness finding. L3 is
tracked by an explicit machine-schema ownership criterion on `6g3q4rt7mgjn`; L4 is tracked by an
explicit canonical-snapshot-loader criterion on `6g3q4rt0wzkq`. No speculative standalone task or
graph-library abstraction was created. Final post-remediation validation is recorded in task
`6g3q4rst78qy`.
