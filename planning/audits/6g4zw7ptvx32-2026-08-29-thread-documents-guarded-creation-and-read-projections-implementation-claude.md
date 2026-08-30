---
schema: 1
id: 6g4zw7ptvx32
bucket: closed
area: thread-documents-guarded-creation-and-read-projections-implementation-claude
date: "2026-08-29"
updated_at: "2026-08-29"
---
# Audit: Thread documents, guarded creation, and read projections — Claude adversarial review — 2026-08-29

> Edit findings through `tskflwctl audit finding` so status and resolution metadata stay queryable.

Independent adversarial review of the uncommitted implementation on
`feat/thread-documents-read-projections` against
[ADR-0006](../adrs/0006-adopt-threads-as-task-dags.md) and
[`6g3q4rtmv4ak`](../tasks/6g3q4rtmv4ak-add-thread-documents-guarded-creation-and-read-projections.md).
The whole slice is in the working tree (79 files, ~1160 insertions); `main..HEAD` is empty, so the
review target is `git diff` plus the untracked `internal/{domain,core,store,cli,wire}/thread*`,
`internal/cli/render/thread.go`, goldens, fixtures, and generated CLI docs.

Coverage: Thread domain invariants and document parsing; stable identity, membership
canonicalization, and hand-edited malformed state; projection semantics (shared membership,
deprecated/deferred members, nominal vs sound completion, direct external gates, completed-Thread
drift, unhealthy graphs, fail-closed frontier); guarded creation (snapshot boundaries, cross-kind
identity on every task-creation path, raw-writer CAS, cooperating-writer serialization, planner
re-entry, retry, dry-run equivalence, committed post-write failure); Projects retirement; lint and
diagnostics; CLI/wire contracts; and architectural fan-out.

Executable validation performed: `go test ./...`, `go test -race ./...` (both green),
`golangci-lint run ./...` (0 issues), `go mod tidy -diff`, `git diff --check`, and regeneration of
`docs/cli/` (working-tree docs match `just docs` byte-for-byte — `just docs-check` only fails
because the branch is uncommitted). Eight scratch planning repositories were driven through the
real binary for behavioural repros: a projection lab, a malformed-document lab, a `lint --fix`
comparison lab, a legacy-repo lab, a 24-way parallel create stress, a graph-cycle lab, a six-case
Projects-retirement matrix, and an adversarial-input lab.

**Verdict:** not ready to close. The architecture is sound and the concurrency work is genuinely
strong — cross-process serialization, whole-snapshot CAS, planner re-entry rejection, committed-outcome
receipts, and fail-closed dispatch all behave as specified under direct attack. Two defects block.
**H1** makes a fully drained Thread report `inconsistent` because a *deprecated* member's prerequisite
is still counted as an outstanding external gate, which is exactly the "contradictory completion UX"
the acceptance criteria forbid. **H2** leaves `planning/threads/` out of `lint --fix` while
`thread new` refuses to run at all if any Thread document is invalid, so one hand-broken Thread file
wedges Thread creation with no supported repair path — and contradicts this branch's own CLAUDE.md.
The remaining findings are hardening in this slice, plus one explicitly-deferred list.

## Findings

### High

#### H1. A deprecated member's prerequisite still gates sound closure, contradicting the deprecated-excluded rollup  · **Status:** fixed

**File:** `internal/core/thread_projection.go:128-135,157-160` | **Component:** Thread projection semantics
**Effort:** S · **Urgency:** acute

**Class:** implementation defect. · **Disposition:** blocking in this slice.

The rollup deliberately excludes deprecated members (`thread_projection.go:115-118`) because withdrawn
work is not the initiative's work — `total`, `done`, and `drained` all skip them. The external-gate
set, however, is derived from **every** `memberIDs` entry including withdrawn ones
(`thread_projection.go:129-135`). A withdrawn member's prerequisites therefore remain outstanding
external gates forever, and `thread_projection.go:157-160` makes any outstanding external gate mark a
completed Thread inconsistent.

Reproduced end to end with the real binary. Tasks `alpha → beta → gamma` (chain) plus `delta` depending
on `epsilon`; Thread membership `{alpha, beta, gamma, delta}`; `alpha/beta/gamma` completed, `delta`
deprecated, `epsilon` left `ready-to-start`; Thread status set to `completed`. `thread show --json`:

```
rollup {'done': 3, 'total': 3, 'drained': 3, 'deprecated': 1}
health healthy   inconsistent True
problems []      graph_problems []
gates [('6g4zs3j61jzt', 'epsilon', 'ready-to-start', True)]
members [('alpha','completed'), ('beta','completed'), ('gamma','completed'), ('delta','deprecated')]
```

Human output is worse, because nothing on screen connects the warning to its cause:

```
status:  completed  ⚠ inconsistent
progress:  3/3 done · 3/3 drained · 1 deprecated
graph:  healthy · 0 frontier · 1 external gate(s)
```

Every live member is soundly complete, the graph is healthy, there are no problems, and the Thread is
still permanently un-closable. `epsilon` is reachable only through work the author explicitly withdrew.
The acceptance criterion "Views expose nominal `done / total`, sound `drained / total`, graph health,
and exact outstanding external gates **without contradictory completion UX**" is what fails here; the
implementation matches the literal wording of the implementation contract but not its intent, because
that contract also states `total` excludes deprecated members.

**Recommendation:** build the external set from members that count toward the rollup — skip a member
whose graph role is `RoleWithdrawn` in the `thread_projection.go:129` loop. If withdrawn members'
prerequisites are still worth *listing*, keep them but force `Outstanding: false` (or a third role) so
they never enter the closure test. Add a projection test: completed Thread, one deprecated member whose
only prerequisite is external and unfinished, expect `inconsistent == false`.

**Resolution:** External gates now derive only from non-deprecated members. A
completed Thread with a deprecated member and unfinished outside prerequisite
remains consistently drained; regression coverage pins the rule and ADR-0006
records it.

#### H2. `lint --fix` never visits `planning/threads/`, so one hand-broken Thread file wedges Thread creation with no supported repair  · **Status:** fixed

**File:** `internal/store/fix.go:131-143`; `internal/core/thread_creation.go:62-98`; `CLAUDE.md` hygiene bullet; `README.md` lint line | **Component:** lint repair path / guarded-creation preconditions
**Effort:** S · **Urgency:** acute

**Class:** implementation defect plus documented-vs-shipped mismatch. · **Disposition:** blocking in this slice.

`FixFrontmatter` repairs `tasksDir`, `epicsDir`, `auditsDir`, and `researchDir` (`fix.go:131-143`).
`threadsDir` is absent. This is not a deliberate exclusion pattern: this branch *did* add `threadsDir`
to the stale-temp sweep (`fix.go:155`) and to `referencesTo` (`fix.go:219`), so the repair list looks
like a missed third edit.

Meanwhile `ValidateThreadCreationSource` hard-fails on the **first** unreadable
(`thread_creation.go:66-68`) or invalid (`thread_creation.go:77-80`) Thread document. The two combine
into a state the tool cannot leave:

```
# planning/threads/6g4zsfffffff-missingid.md — frontmatter with no `id:`

$ tskflwctl lint
missingid
  id: missing or invalid stable Thread id

$ tskflwctl lint --fix
fixed .../planning/research/6g4nm7801ana-ref.md      # research: id assigned (was missing)
  - id: assigned (was missing)
could not auto-repair:
missingid
  id: missing or invalid stable Thread id

$ tskflwctl thread new "Anything" --description d --goal g
error: validation failed: existing Thread .../6g4zsfffffff-missingid.md is invalid:
       validation failed: thread id "" is not a stable id                    [exit 11]
```

An identically-broken **research** file in the same repository and the same `lint --fix` run is
repaired; the Thread is not. The same wedge reproduces with an unparseable Thread — `description: needs:
quoting` — which `lint --fix` normalizes for every other kind and leaves untouched under `threads/`.
`thread list`/`show`/`frontier` keep degrading correctly throughout, so reads stay diagnostic; only the
write path is stuck, and the only exit is hand-editing YAML — precisely what the tool's contract tells
users not to do.

This also contradicts documentation shipped in this branch. `CLAUDE.md` now reads "`--fix` to auto-repair
ordinary frontmatter; **Thread membership and graph defects remain deliberate repairs**", which asserts
ordinary Thread frontmatter *is* auto-repaired. `README.md`'s unchanged `lint --fix # auto-repair
frontmatter (quote ':' values, normalize lists, backfill ids)` is now wrong for one of the five kinds.

**Recommendation:** add `fixDir(s.threadsDir, true)` next to the research call. Threads are flat and
id-led with a co-located `id:` field, so text normalization plus filename-derived id backfill are exactly
as safe as for tasks, audits, and research; membership and graph defects stay out of `--fix` as
documented. Cover it with a `lint --fix` test for a Thread missing `id:` and one with an unquoted-colon
value. If the omission was in fact deliberate, correct CLAUDE.md and README instead, and give
`ValidateThreadCreationSource` a remedy string naming the file and the hand-repair.

**Resolution:** FixFrontmatter now visits threads for safe scalar normalization
and filename-owned id backfill while leaving tasks membership syntax untouched.
Tests cover colon repair, missing id, membership preservation, and cross-kind
collision refusal.

### Medium

#### M1. `inconsistent` is an unexplained boolean — the payload never says why a completed Thread is inconsistent  · **Status:** fixed

**File:** `internal/core/thread_projection.go:157-160`; `internal/wire/thread.go:78`; `internal/cli/render/thread.go:66-68,36-38` | **Component:** projection diagnostics / wire contract
**Effort:** S · **Urgency:** soon

**Class:** contract gap. · **Disposition:** hardening in this slice.

`Inconsistent` collapses four distinct conditions (empty or effectively empty, `drained != total`, an
outstanding external gate, broken graph/member evidence) into one bool. In the H1 repro the entire
payload is `health: healthy`, `problems: []`, `graph_problems: []`, `inconsistent: true` — a machine
consumer must re-implement all four disjuncts against the rest of the view to explain the flag, and the
human renderer prints a bare `⚠ inconsistent` with no reason. Everywhere else in this codebase a
diagnosis carries stable vocabulary (`GraphProblemCode`, `ThreadProblemCode`, `BlockerReason`); this one
does not.

This is also a process finding: a reason code for "outstanding external gate" would have made H1 visible
the moment the deprecated-member case was exercised.

**Recommendation:** extend `ThreadProblemCode` with `completed-thread-empty`,
`completed-thread-undrained`, `completed-thread-external-gate`, and `completed-thread-broken-evidence`,
append the applicable ones to `Problems`, and keep `inconsistent` as the derived boolean for
compatibility. Render the reason in `ThreadShowHuman` and in `ThreadsHuman`'s diagnostics block.

**Resolution:** Completed inconsistency now emits stable empty, undrained,
external-gate, and unhealthy-evidence reason codes. Human diagnostics render
them and wire output carries them alongside the compatibility boolean.

#### M2. `threadFields` is not a faithful Thread serializer and silently drops `updated_at` / `started_at` / `ended_at`  · **Status:** fixed

**File:** `internal/store/threadcreation.go:141-160`; `internal/domain/thread.go:61-63` | **Component:** materialization seam
**Effort:** S · **Urgency:** soon

**Class:** latent defect in the seam this slice exists to establish. · **Disposition:** hardening in this slice.

`threadFields` emits `schema, id, status, description, goal, [target_date], created, [tags], tasks`.
`domain.Thread` also persists `Updated`, `StartedAt`, and `EndedAt`; `ValidateThreadDocument`
(`domain/thread.go:93-104`) validates them and `LintThread` (`domain/lint.go:249-256`) lints them, but
nothing writes them.

Today this is inert because `ValidateThreadCreationPlan` rejects all three at creation
(`core/thread_creation.go:108-110`). The risk is structural: `materializeThreadCreation` is explicitly
advertised — in the task's implementation contract, in the code comment at `threadcreation.go:113-114`,
and in the acceptance criterion about "lock-free internal Thread materialization reusable by a later
compound bulk capability" — as the seam the membership/lifecycle and bulk-apply slices will reuse. Those
slices exist precisely to set `started_at`/`ended_at`/`updated_at`. Any reuse that rebuilds a Thread
document through `threadFields` will erase them silently, and whole-file rebuilds also violate the
repository rule that frontmatter is edited surgically to preserve unknown fields, comments, and key order.

**Recommendation:** either serialize every persisted field in `threadFields` with a round-trip test
(write → `parseThread` → compare), or rename it `threadCreationFields`, document that it is create-only,
and make the follow-on slice reach existing Thread files through the surgical `updateFrontmatter` path
rather than through this builder.

**Resolution:** The helper is now explicitly named and documented as a
creation-only field builder. The follow-on membership/lifecycle task requires a
lock-free surgical updater that preserves timestamps, unknown fields, comments,
and key order.

#### M3. `ThreadView.Health` conflates repository-graph health with Thread-document defects  · **Status:** fixed

**File:** `internal/core/thread_projection.go:80-88,103,150-156`; `internal/wire/thread.go:77` | **Component:** wire contract
**Effort:** S · **Urgency:** eventually

**Class:** contract ambiguity. · **Disposition:** hardening in this slice.

`view.Health` starts as `graph.Health()` and is then overwritten with `GraphBroken` when the Thread
document fails validation (`:83`) or references a missing task (`:103`). The wire field documents itself
as `healthy|degraded|broken`, which every other surface uses for the repository DAG.

Verified: in a repository whose task graph is perfectly healthy, a Thread with a malformed document
reports `"health":"broken"` alongside `"graph_problems":[]`. A dispatcher reading `health` cannot
distinguish "the repository DAG is unsafe" from "this one Thread file is malformed" without also
inspecting `problems`.

The fail-closed frontier behaviour driven off this field (`:150`) is correct and should stay — an
unsound Thread must not dispatch. The issue is that one field answers two questions.

**Recommendation:** keep `health` as the repository-graph verdict and add a derived field for the
projection verdict (e.g. `projection_health`, or a `dispatchable` boolean) that folds in Thread-local
defects and drives the frontier gate. If the merged meaning is intended, say so explicitly in the
`jsonschema` description and expose the raw graph verdict separately.

**Resolution:** Thread views now expose separate graph_health and
projection_health verdicts. Thread-local defects break only projection health,
and frontier dispatch is gated on that combined projection verdict.

#### M4. `thread list --json` repeats the whole repository graph-problem set once per Thread  · **Status:** fixed

**File:** `internal/core/thread_projection.go:81`; `internal/core/service_thread.go:127-131`; `internal/cli/render/thread.go:53-59` | **Component:** wire payload / rendering
**Effort:** S · **Urgency:** eventually

**Class:** efficiency and readability. · **Disposition:** hardening in this slice.

`ProjectThread` attaches `graph.Problems()` — the complete repository problem list — to every view, and
`ListThreadViews` calls it once per Thread against the same graph. Payload size is O(threads ×
problems). Verified with 12 Threads and one injected missing-dependency problem: 12 identical
`graph_problems` entries. The human renderer dedupes within a single view but re-prints the whole set
under each Thread's `Diagnostics` heading.

Not a correctness issue and not urgent, but this is the machine surface agents are meant to poll, and
this repository treats `--json` token cost as a first-class concern.

**Recommendation:** hoist repository graph problems to `ThreadsEnvelope` (beside `unreadable`) and keep
only Thread-local `problems` per view. `thread show` and `thread frontier` render one Thread, so they
can keep the embedded copy.

**Resolution:** ThreadListView hoists repository graph health and problems to
one list-level payload. Per-Thread list rows retain local diagnostics without
repeating the global problem set; show and frontier retain complete
single-Thread evidence.

### Low

#### L1. Thread `tags` are re-sorted in machine output, diverging from the file and from every other entity  · **Status:** fixed

**File:** `internal/wire/thread.go:32,36-40`; compare `internal/wire/dto.go:50,224,318` | **Component:** wire contract
**Effort:** XS · **Urgency:** eventually

**Class:** inconsistency. · **Disposition:** hardening in this slice.

`ToThreadJSON` runs both `Tags` and `Tasks` through `nonNilSortedStrings`. Sorting `Tasks` is correct —
the contract requires a sorted, duplicate-free membership set. Sorting `Tags` is required by nothing and
silently reorders author-supplied metadata: a Thread written with `tags: [z, a, m]` reports
`["a","m","z"]` (verified). `ToTaskJSON`, `ToResearchJSON`, and `ToEpicJSON` all pass tags through
verbatim with `omitempty`, so Threads are the only kind whose tag order and emptiness behaviour differ.

**Recommendation:** emit `Tags` verbatim. Keeping it non-`omitempty` so consumers can `len()` it is
defensible, but then it should be a deliberate, documented deviation rather than a side effect of
reusing the membership helper.

**Resolution:** Thread JSON now preserves authored tag order while continuing to
sort the semantic task membership set. A mapper regression test pins both
behaviors.

#### L2. Init prints a green "updated" for a run whose only outcome was preserving legacy Projects content  · **Status:** fixed

**File:** `internal/cli/init.go:230-246` | **Component:** init human output
**Effort:** XS · **Urgency:** eventually

**Class:** misleading output. · **Disposition:** hardening in this slice.

The early-return branch now correctly accounts for `removed` and `LegacyProjects`
(`init.go:230`), but when `created`, `removed`, and `tracked` are all empty and only `LegacyProjects` is
set, control falls through to the verb switch, `len(created) == 0` selects `updated`, and the run prints
`✔ updated <root>` followed by the preservation warning — a success marker for a run that changed
nothing.

**Recommendation:** when nothing was created, removed, or tracked, use the `· already initialized` line
(or a neutral verb) and let the ⚠ carry the message.

**Resolution:** An explicit init repair that only discovers preserved legacy
Projects content now emits a neutral already-initialized receipt plus the
warning, not a green updated claim.

#### L3. Bare `tskflwctl init` never mentions that the Threads/Projects scaffold repair is available  · **Status:** fixed

**File:** `internal/cli/init.go:56-58,105-133` | **Component:** upgrade discoverability
**Effort:** S · **Urgency:** eventually

**Class:** rollout gap. · **Disposition:** hardening in this slice.

`initTopologyFlagsChanged` routes a bare `init` on a configured repo to `runInitExisting`, an identity
read. Verified across six scratch repositories: with `threads/` absent and a retirable
`projects/.gitkeep` present, `tskflwctl init` prints `· already initialized` and changes nothing; only
`tskflwctl init --taskflow-root planning` performs the repair. `runInitExisting` already surfaces pending
*configuration* migrations with a `→ tskflwctl config migrate` hint, but says nothing about the scaffold.

This is discoverability, not breakage — reads tolerate a missing `threads/` and `thread new` creates it
on demand (both verified). But every pre-existing planning repository is in this state, and the new
`Long` help text is currently the only place the repair is documented.

**Recommendation:** have `runInitExisting` detect a missing entity directory or a retirable legacy
`projects/` and print a `→ tskflwctl init --taskflow-root <configured>` hint in the same shape as the
existing migration hint.

**Resolution:** Bare init performs a read-only scaffold preview. Human output
prints the exact matching taskflow-root repair command, and JSON exposes
scaffold_repair_available plus scaffold_repair_command without mutating the
tree.

## Verified correct under attack

These were probed adversarially and behaved as specified; recorded so the next reviewer does not
re-derive them.

- **Guarded creation, cooperating writers.** 12 concurrent `thread new` plus 12 concurrent `task new`
  against one repository, separate processes: 24/24 exit 0, 12 Threads and 13 tasks on disk, no
  duplicate ids, `lint` clean. Cross-process flock plus the process-local guard serialize correctly.
- **Raw-writer CAS.** `sameThreadSourceSnapshot` keys on `FilenameID + Path + SourceVersion` and
  `slices.Equal` over `FileProblem`s; combined with `SameSourceSnapshot` for the task graph, any raw edit
  between plan and write conflicts. Covered by `TestThreadCreationRejectsRawTaskRaceByWholeSnapshotCAS`
  and `...RawThreadRaceByWholeSnapshotCAS`.
- **Planner re-entry.** `callThreadCreationPlanner` takes `enterRepositoryPlanner`; every `FS` read entry
  point calls `rejectRepositoryPlannerCall`, converting re-entry into an attributable `ErrConflict`
  instead of a self-deadlock.
- **Committed post-write failure.** `result.Committed` is set before the deferred unlock can fail; the
  service retry loop is gated on `!result.Committed`, and the CLI wraps it into
  `threadCreationCommandFailure` so `--json` errors carry `thread_mutation` with `committed:true`, the
  path, and the workspace.
- **Cross-kind identity, both directions.** `thread new` rejects a task-owned id inside the guard;
  `CreateTask` and the create branch of `MutateTaskLifecycle` both call `ensureTaskIDNotThread` while
  holding the same canonical-root lock. Those are the only two task-creation write paths
  (`writeNewFile`/`writeNewFileUnlocked` call sites confirm it). Hand-edited collisions are caught by lint
  in both directions.
- **Fail-closed dispatch.** With a hand-made dependency cycle, `thread new` refuses with an attributable
  message and a repair remedy; `thread list`/`show`/`frontier` still render, report `health: broken`, and
  return an empty frontier. `task new` (ungated ordinary creation) is unaffected.
- **Dry-run equivalence.** `thread new --dry-run --json` produces the same document shape with
  `changed:true, dry_run:true, committed:false` and writes nothing; the real run differs only in id and
  `committed:true`.
- **Projects retirement, all six shapes.** Empty dir → removed. Lone regular `.gitkeep` → both removed.
  Any other content → preserved with the remedy. `projects` as a symlink to a directory → preserved
  (`os.Lstat` + `!IsDir`). `.gitkeep` as a symlink or as a directory → preserved (`DirEntry.Info()` is
  Lstat-based, `Mode().IsRegular()` false). Dry run previews the removals and touches nothing.
- **Malformed hand-edited Threads.** Missing frontmatter, non-id-led filename, invalid status, unsorted
  membership, duplicate member, unknown member, id drift, and duplicate Thread ids are each reported with
  field-attributed lint messages; `lint --fix` correctly refuses all of them (no unsafe automatic repair).
- **Adversarial input.** A `--body-file` whose content is itself a frontmatter block is stored as body
  only; `description`/`goal` with newlines are rejected; values containing `:` are correctly quoted on
  write and round-trip through `lint`.
- **Legacy repositories.** With `threads/` absent, `thread list` and `lint` succeed, `thread new` creates
  the directory on demand, and the TUI watcher tolerates the missing path (`newWatcher` skips failed
  `Add`s).
- **Fan-out.** `schema thread`, `schema --json` (`thread_statuses`), `schema --json-schema` (11 Thread
  `$defs`, four envelopes registered and required), `template list`, all four shell completions
  (`thread show`, `--status`, `--task`, `--template`), five golden files, the `testdata/planning/threads`
  fixture, `WatchPaths`, `spacehealth.planningRootEmpty`, and generated `docs/cli/` are all present and
  consistent. Working-tree `docs/cli/` matches `just docs` byte-for-byte.
- **Ordering and empty-array behaviour.** `ListThreadViews` sorts by id then path; members sort by stable
  task id; external gates sort by id. `thread list --json` on an empty repository emits
  `{"threads":[],"unreadable":[]}`; empty human output matches every other `list` command. `Removed` is
  normalized to `[]` on every init path including `runInitExisting`.

## Explicitly deferred — not defects in this slice

- Membership and Thread lifecycle mutation verbs, and affected-Thread receipts on task lifecycle
  transitions → `ship-guarded-thread-membership-and-lifecycle-mutations`, per the readiness checkpoint.
- Bulk linking → waits on that second slice plus dependency operations.
- Thread surfaces in the TUI, `board`, and `status`. The store/port/projection seams are ready; no TUI
  work is claimed by this task.
- `thread list -o table|csv` and `--json -c` column projection, which `task list` has and `thread list`
  does not. Not required by the implementation contract; worth a follow-up for agent parity.
- Cross-kind identity between Threads and research/audit ids. The contract narrows global uniqueness to
  task↔Thread and this slice delivers exactly that; research↔task collisions were equally unchecked
  before.
- Pre-existing, unchanged by this branch: `config.Init`'s dry run under-reports `<dir>/.gitkeep`
  (`isEmptyDir` is false for a directory that does not exist yet), so
  `--dry-run` shows `created:["threads"]` where the real run shows `["threads","threads/.gitkeep"]`.
- Pre-existing, unchanged by this branch: `repairInvalidID`'s target-collision check is directory-local
  (`store/fix.go:200`), so a canonicalized task id could in principle collide with a Thread id. Vanishing
  probability for 12-character random ids, and lint catches it afterwards.

## Candidate tasks

- ⏳ `tskflwctl task new "Exclude withdrawn members from Thread external gates" --epic 30-threads-and-task-dependency-graphs --tags threads,core` — H1: derive external gates from rollup-counted members only.
- ⏳ `tskflwctl task new "Repair Thread frontmatter with lint --fix" --epic 30-threads-and-task-dependency-graphs --tags threads,store,lint` — H2: add `threadsDir` to `FixFrontmatter` and cover missing-id/colon repair.
- ⏳ `tskflwctl task new "Give completed-Thread inconsistency a reason vocabulary" --epic 30-threads-and-task-dependency-graphs --tags threads,wire` — M1.
- ⏳ `tskflwctl task new "Make the Thread materialization seam serialize every persisted field" --epic 30-threads-and-task-dependency-graphs --tags threads,store` — M2; land before the membership/lifecycle slice reuses it.
- ⏳ `tskflwctl task new "Separate Thread projection health from repository graph health" --epic 30-threads-and-task-dependency-graphs --tags threads,wire` — M3.
- ⏳ `tskflwctl task new "Hoist repository graph problems out of per-Thread views" --epic 30-threads-and-task-dependency-graphs --tags threads,wire` — M4.
- ⏳ `tskflwctl task new "Fix Thread tag ordering and init upgrade diagnostics" --epic 30-threads-and-task-dependency-graphs --tags threads,cli` — L1, L2, and L3 together; all are XS/S and touch adjacent output paths.
