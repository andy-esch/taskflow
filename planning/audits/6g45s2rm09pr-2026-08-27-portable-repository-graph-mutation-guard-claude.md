---
schema: 1
id: 6g45s2rm09pr
bucket: closed
area: portable-repository-graph-mutation-guard-claude
date: "2026-08-27"
updated_at: "2026-08-27"
---

# Audit: portable-repository-graph-mutation-guard-claude — 2026-08-27

> Edit findings in place and flip each `**Status:**` as you work it.

Adversarial implementation-readiness review of the portable repository graph mutation
guard (task `6g3q4rt0wzkq`, epic `30-threads-and-task-dependency-graphs`, branch
`feat/portable-graph-mutation-guard`, uncommitted tree vs merge-base `509a7c3`).

Method: every claim below was executed, not inferred. Counterexamples ran through
`go test -overlay=` against probe files held outside the repository, so no production
or test file was modified — `git status --porcelain` was identical before and after
(19 modified, 5 new implementation files), and this audit is the only repository
write. Two mechanisms were additionally checked by **mutation testing** — deliberately
breaking a guard and re-running the committed suite to see whether anything fails.

A sibling audit `6g45qfv27vrm` covers the same slice. Where we independently reach the
same conclusion that is stated as corroboration with my own evidence; the majority of
what follows (H3, H4, M1, M4, M5, M6, L1, L3, L4, L5, L6, L7) is not in it, and my
verdict is harder.

## Executive verdict

**Not ready to merge; safe with a bounded set of amendments.** The design is right and
the hard parts work: cross-process serialization of `MutateTaskGraph` is real — I
confirmed it with actual subprocesses racing opposite edges, and the loser was rejected
with a cycle diagnosis while the final graph stayed healthy with exactly one edge. The
control-inverted boundary, the persistence-token hygiene, the prefix-safety rule, and
the interrupt/rerun convergence story all survived attack.

Three things block merge:

1. **The re-entry guard is instance-local while the lock it protects is root-global**
   (H1). A planner holding a second `*FS` on the same repository is not detected: a
   read escapes the snapshot and returns live state; a write **hangs forever** with no
   timeout and no error. `docs/ARCHITECTURE.md` states "Planner callbacks cannot
   re-enter the Store; nested access fails explicitly instead of self-deadlocking" —
   that sentence is false as written, and acceptance criterion 3 is ticked on it.
2. **`rejectGraphPlannerCall` cannot tell a re-entrant planner from an unrelated
   concurrent caller** (H2). Any goroutine sharing the `*FS` — which is the TUI's
   shape, one `Service` behind `tea.Cmd`s — gets `ErrConflict` reading "graph mutation
   planner cannot call Store methods" for a plain `GetTask`. Reads previously could not
   fail for concurrency reasons at all.
3. **The two mechanisms this slice actually adds have no test that fails when they are
   deleted** (H3). Replacing `processRepositoryLock` with a no-op leaves the entire
   `internal/store` suite green. Removing symlink resolution from `repositoryLockKey`
   leaves it green. `TestRepositoryLockSerializesIndependentStoresInOneProcess` cannot
   distinguish the new mutex from flock, because flock alone already blocks a second
   in-process handle on Linux/macOS. Acceptance criterion 1 says "explicit, **tested**
   … contract" and is ticked.

Separately, the cross-process and termination tests drive `platformWriteLock`, a helper
with **zero production callers** (H4) — so the product lock path has no cross-process
coverage at all; I had to write that test myself.

Two ticked acceptance criteria are additionally not met as written: AC2 ("without …
teaching the store graph semantics") — the store now owns cycle, prefix-safety, legacy
field, and health validation (M4); and AC6 ("remove or fold any duplicate or otherwise
unused `ReadTaskGraph` scan seam so lint and mutation cannot drift") — `Service.Lint`
still builds its own snapshot and `ReadTaskGraph` still has no callers (M5).

Classifications: **[pre-merge blocker]**, **[follow-up]**, **[monitor]**.

---

## Findings

#### H1. Planner re-entry detection is per-`*FS` while the guard is per-root: a second Store on the same repository deadlocks, or silently escapes the snapshot  · **Status:** fixed

**File:** internal/store/lock.go:51-69; internal/store/fsstore.go:45-49 | **Component:** store / concurrency
**Effort:** M · **Urgency:** acute

**[pre-merge blocker]**

`plannerActive` is an `atomic.Bool` field on the `FS` struct (`fsstore.go:45-49`), and
`rejectGraphPlannerCall` (`lock.go:64-69`) consults only the receiver. The lock it
protects — `processRepositoryLock` (`lock.go:38-49`) — is keyed by canonical root and
shared by every `FS` in the process. The two scopes do not match, so the failure the
guard exists to convert into an error is reachable again through a second handle.

Reproduction (executed):

```
TestADV_SecondFSWriteInPlannerDeadlocks
  DEADLOCK CONFIRMED: planner write via a second *FS on the same root hangs
  with no timeout and no attributable error                      (3.00s, aborted)

TestADV_SecondFSGivesPlannerLiveState
  snapshot priority="high"  live-read-through-second-FS priority="low" err=<nil>
```

The write case hangs on `mutex.Lock()` inside `processRepositoryLock` — no deadline, no
`ErrConflict`, no stack attribution; the operator sees a wedged process. The read case
is quieter and worse in kind: the planner obtained **live, unsynchronized filesystem
state** that contradicts the immutable snapshot it was handed, which is precisely the
isolation the control-inverted boundary sells. In my probe the mutation happened to
abort afterwards on the whole-snapshot CAS, but only because the probe also wrote; a
read-only escape leaves no trace.

The comment at `lock.go:14-18` names this scenario as the motivation ("a future
long-lived adapter may create more than one `*FS`") and concludes the keyed mutex
"makes the combined contract identical within one [process]". It makes *serialization*
identical; it leaves *re-entry detection* instance-local, which flips the failure mode
from an attributable conflict to a silent hang.

**Why existing tests miss it:**
`TestMutateTaskGraphRejectsPlannerStoreCallsAndNestedMutationWithoutHanging`
(`graphmutation_test.go:220-254`) captures the *same* `fs` in its planner closure — the
one shape the instance-local sentinel covers. No test constructs a second `NewFS(root)`.

**Recommendation:** move the sentinel next to the mutex — store the owning
`*FS` (or a mutation-generation counter) in `repositoryMutexes.byRoot` keyed by
`repositoryLockKey`, and have `rejectGraphPlannerCall` consult that. Same code volume,
and it closes both the deadlock and the live-read escape.

**Resolution:** Planner activity moved from one FS instance to the
canonical-root guard. A callback using a second FS now receives ErrConflict for
reads, writes, lint helpers, and nested mutation instead of reading live state
or hanging; regression coverage uses a second NewFS(root).

#### H2. `rejectGraphPlannerCall` cannot distinguish re-entry from legitimate concurrency, so unrelated operations fail with a false, misleading error  · **Status:** fixed

**File:** internal/store/lock.go:64-69, 86-89 | **Component:** store / concurrency
**Effort:** M · **Urgency:** acute

**[pre-merge blocker]**

Every Store entry point begins with `rejectGraphPlannerCall`, which tests one process-
wide-per-instance boolean. It has no way to know whether the caller is the planner
goroutine or an unrelated one. Executed:

```
TestADV_LegitimateConcurrencyMisattributed
  concurrent unrelated GetTask   during mutation -> graph mutation planner cannot
      call Store methods or begin a nested mutation: conflict   (errors.Is ErrConflict = true)
  concurrent unrelated SetFields during mutation -> (same error)
```

Two distinct harms:

- **Reads now fail.** `GetTask`, `ListTasks`, `ListEpics`, `GetAudit`, and the four
  `Resolve*Path` helpers are all guarded (`paths.go:9-36`). Before this slice a read
  could not fail for a concurrency reason. A TUI refresh landing during a graph
  mutation now surfaces exit-14 conflict text.
- **The attribution is wrong.** The message accuses the operator's planner of illegal
  re-entry when the real situation is ordinary contention that the lock would have
  handled by waiting.

This is not hypothetical for the roadmap: `docs/ARCHITECTURE.md` describes the TUI as
reading through one `core.Service` as `tea.Cmd`s, which run in separate goroutines over
one `*FS`; and the slice's own rationale cites "a future long-lived adapter". As soon as
`ship-guarded-dependency-mutations-and-graph-queries` or
`add-usage-informed-thread-views-to-the-tui` puts a mutation behind that Service, this
fires.

**Why existing tests miss it:** every concurrency test either calls from inside the
planner (expecting the conflict) or uses separate `NewFS` instances (which, per H1, are
not checked at all). Nothing exercises the same `*FS` from a non-planner goroutine.

**Recommendation:** record the planner's goroutine identity, or — simpler and
adapter-friendly — have the mutation publish the sentinel on the *root-keyed* guard
entry (H1's fix) and reject only a caller that already holds it, letting everyone else
block on the mutex as before. Whichever way, unrelated callers must wait, not error.

**Resolution:** The misleading instance-local error was replaced by an explicit
root-wide callback-exclusion contract. ADR-0006 records that both callback
re-entry and unrelated access fail fast during the brief planner phase;
long-lived adapters must retry contention or deliberately introduce a new
capability-bearing port.

#### H3. The two guards this slice adds have no test that fails when they are removed  · **Status:** fixed

**File:** internal/store/lock.go:24-49; internal/store/lock_unix_test.go:108-135 | **Component:** testing / concurrency
**Effort:** M · **Urgency:** acute

**[pre-merge blocker]**

I ran a mutation battery — break one mechanism, run the committed suite
(`./internal/store ./internal/core`, `-count=1`):

| mutation | committed suite result |
|---|---|
| `processRepositoryLock` → returns a no-op immediately | **PASS (survives)** |
| `repositoryLockKey` → skip `filepath.EvalSymlinks` | **PASS (survives)** |
| `rejectGraphPlannerCall` → always nil | FAIL (caught) |
| `SameSourceSnapshot` → always true | FAIL (caught) |
| `Task()` → stop clearing `SourceVersion` | FAIL (caught) |
| prefix validation → validate the final graph only | FAIL (caught) |

The same-process mutex is the headline addition (`lock.go:14-22` justifies it at
length) and the canonical-root key is the entire identity model on Windows. Neither is
covered.

`TestRepositoryLockSerializesIndependentStoresInOneProcess`
(`lock_unix_test.go:108-135`) is named for exactly this mechanism, but it cannot
isolate it — flock already blocks a second handle in the same process:

```
TestADV_ProcessMutexVersusFlockInProcess
  flock alone ALREADY blocks a second in-process handle; the existing
  serialization test cannot distinguish it from the new process mutex
```

Acceptance criterion 1 — "Every supported platform has an explicit, **tested**
cooperating-writer serialization contract" — is ticked on this evidence.

**Recommendation:** add a test that can only pass with the mutex — e.g. inject a
platform-lock stub via an existing test hook and assert `writeLock` still serializes —
plus a `repositoryLockKey` table test over `.`/`..`, trailing separators, a symlinked
root, and (asserted, not executed) the Windows lowercase rule. Keep the mutation
battery as a review gate for this file.

**Resolution:** A direct product-function test now fails if
processRepositoryLock becomes a no-op, canonical-root table coverage includes
dot, dot-dot, trailing separator, symlink alias, and case-folding, and the
cross-process tests exercise writeLock/MutateTaskGraph rather than an isolated
platform helper.

#### H4. The cross-process and termination lock tests drive `platformWriteLock`, which has no production callers  · **Status:** fixed

**File:** internal/store/lock.go:107-115; internal/store/lock_unix_test.go:30, 65 | **Component:** testing / concurrency
**Effort:** S · **Urgency:** acute

**[pre-merge blocker]**

`platformWriteLock` is documented as "kept for direct lock-contract tests"
(`lock.go:107-108`), and a repository-wide grep finds no caller outside `lock*.go` and
`lock_unix_test.go`. Product mutations go through `writeLock`/`checkedWriteLock`, which
additionally take the process mutex and the re-entry check. So
`TestRepositoryLockReleasesWhenProcessTerminates` and its subprocess helper prove a
contract about a function the product never invokes, and **there is no test that a real
mutation in one process excludes a real mutation in another.**

That gap matters because the product path is where the two halves interact: the process
mutex is released in the same closure as the platform lock (`lock.go:96-104`), and a
release-ordering mistake there would be invisible to the current tests.

I wrote the missing test and it passes — the contract does hold today:

```
TestADV_CrossProcessOppositeEdges          (two real child processes, 150ms window)
  process 1: REJECTED: validation failed: planned write prefix ending at task
             6g4500000001 would leave a broken graph: dependency cycle:
             6g4500000001 -> 6g4500000002 -> 6g4500000001
  process 2: APPLIED
  final health=healthy total edges=1
```

Note that `TestMutateTaskGraphConcurrentOppositeEdges`
(`graphmutation_test.go:122-173`) races **goroutines**, which the process mutex alone
serializes — it does not exercise flock.

**Recommendation:** re-point the two subprocess tests at `writeLock`, and commit a
cross-process `MutateTaskGraph` opposite-edge test like the one above. Then delete
`platformWriteLock`, which exists only to be tested.

**Resolution:** The test-only platformWriteLock seam was deleted. Process
termination now proves a product writeLock holds and releases the OS flock, and
two real child processes racing opposite MutateTaskGraph edges yield exactly one
applied edge and one cycle rejection.

#### M1. Multi-file plans verify every file before writing any, so a raw edit inside the apply loop is clobbered  · **Status:** fixed

**File:** internal/store/graphmutation.go:90-105 | **Component:** store / CAS
**Effort:** S · **Urgency:** soon

**[follow-up]**

The apply phase is two sequential loops: `verifyUnchanged` for every write
(`:90-94`), then `writeFileAtomic` for every write (`:95-105`). For a plan of *n*
files, file *n*'s CAS is validated before files 1…*n*−1 are written, so the
verify→write window for the last file is the duration of the whole apply loop. Every
other write path in the store keeps that window to a single file operation.

Reproduction (executed) — a non-cooperating editor changes the second target after its
verify passed:

```
TestADV_MultiFileVerifyWriteWindowClobbers
  mutation err=<nil>
  raw edit survived in the second-written file: false
  CLOBBER CONFIRMED: a raw edit that landed after verifyUnchanged was silently
  overwritten by the plan's later write
```

The mutation reported **success**. The advisory lock does exclude cooperating writers,
so this needs a raw editor — but that is exactly the threat model the whole-snapshot
CAS was added for, and bulk-linking will make the window seconds long (see N1).

**Why existing tests miss it:** `TestMutateTaskGraphCASRejectsRawEditBeforeApply`
(`:297-321`) injects its edit through `testHookBeforeGraphVerify`, i.e. strictly
*before* the verify pass. There is no hook usage that interleaves an edit between
verify and write.

**Recommendation:** interleave the loops — verify file *i* immediately before writing
file *i* — keeping the existing all-files pre-check as a cheap early abort. The prefix
contract already tolerates a partial apply, so failing at file *i* is well-defined.

**Resolution:** The all-target preflight remains, and every target is now
reverified immediately before its own atomic replacement. A regression test
injects a raw edit before the second target, returns the first durable
AppliedTaskIDs prefix with ErrConflict, and proves the raw edit survives.

#### M2. Reversing a dependency edge succeeds or fails purely on stable-ID ordering  · **Status:** fixed

**File:** internal/store/graphmutation.go:124-126, 178-187 | **Component:** store / plan validation
**Effort:** M · **Urgency:** soon

**[follow-up]**

Writes are sorted by `TaskID` (`:124-126`) and every prefix in that order must be
non-broken (`:183-187`). Application order is therefore lexicographic, not semantic. For
an edge reversal one order is safe (remove, then add) and the other transiently cyclic —
and which one you get depends on the two ids.

Executed, same plan shape both ways:

```
TestADV_EdgeReversalDependsOnIDOrdering
  reverse edge, dependent has SMALLER id -> <nil>            (accepted)
  reverse edge, dependent has LARGER  id -> validation failed: planned write prefix
      ending at task ekpd8vnydg2h would leave a broken graph: dependency cycle: …
  ASYMMETRY: identical semantic operation succeeds/fails on stable-ID ordering alone
```

For slice 3 this means `task depend` reversal-shaped operations will appear flaky to
users: the same command shape works on one pair of tasks and errors on another, with a
message about "write prefix" that names an implementation concept.

**Why existing tests miss it:**
`TestMutateTaskGraphRejectsBrokenIntermediatePrefixBeforeWriting` (`:366-397`)
deliberately normalizes `firstID`/`secondID` (`:369-374`) so the unsafe orientation is
always chosen. The test correctly pins that unsafe prefixes are rejected; the
normalization is what hides that the safe orientation is silently allowed.

**Recommendation:** order the apply sequence semantically — all dependency *removals*
before all *additions*, `TaskID` order within each phase — which makes every reversal
prefix-safe and stays deterministic. Corroborates sibling `6g45qfv27vrm` H1; the
executed asymmetry above is the part worth adding to it.

**Resolution:** Core preserves planner-provided semantic order rather than
sorting task writes by ID. Tests cover a removal-before-addition reversal that
succeeds in supplied order and the opposite unsafe prefix that fails before any
write.

#### M3. Windows lock identity is a path hash under the *user's* cache directory, and cross-compilation cannot validate any of it  · **Status:** fixed

**File:** internal/store/lock_windows.go:21-47 | **Component:** store / portability
**Effort:** M · **Urgency:** soon

**[follow-up]**

On Unix the lock is `flock` on the repository root's own inode, so every spelling of
the path — symlink, alias, bind mount, relative — lands on one lock. On Windows the
lock file is `os.UserCacheDir()/tskflwctl/locks/<sha256(repositoryLockKey(root))>.lock`,
so mutual exclusion holds **only** when two processes compute the identical key string
under the identical cache root. It does not when:

- two users (or a service account and an interactive user) share a repository —
  different `os.UserCacheDir()`, therefore different lock files, therefore no exclusion
  at all;
- the same share is reached as `Z:\repo` and `\\server\share\repo`, or through `subst`,
  or through a roaming vs local profile — `filepath.EvalSymlinks` resolves junctions and
  symlinks but not drive mappings;
- the cache directory is unavailable, where the code fails the mutation
  (`:29-32`) rather than falling back — correct, but it makes lock availability depend
  on a directory unrelated to the repository.

Lock files also accumulate under `locks/` and are never swept.

I verified `GOOS=windows go build ./...` and `GOOS=windows go vet ./...` are clean, and
that `go test` for `GOOS=windows` cannot run (`exec format error`). There is no
`lock_windows_test.go`. `docs/ARCHITECTURE.md` is honest — "Windows is compile-checked"
— but acceptance criterion 1 claims every supported platform has a *tested* contract and
is ticked.

**Recommendation:** either (a) declare Windows unsupported for now and route it to
`lock_other.go`'s explicit rejection, or (b) put the lock file inside the planning
repository (e.g. `.tskflwctl.lock`, gitignored) so identity is the filesystem's job
rather than a string hash, and add a Windows CI job. Corroborates sibling
`6g45qfv27vrm` M1; the UNC/`subst` and lock-file-accumulation aspects are additional.

**Resolution:** The untested per-user Windows cache lock was removed. Non-Unix
builds now reject repository mutation with ErrValidation; ADR-0006 limits the
supported release contract to native-tested macOS/Linux until Windows has a
cross-user identity and native CI.

#### M4. The store now owns graph semantics that acceptance criterion 2 says it must not  · **Status:** fixed

**File:** internal/store/graphmutation.go:118-196, 242-252 | **Component:** architecture
**Effort:** M · **Urgency:** soon

**[follow-up]**

AC2 reads "…within one repository guard without exposing filesystem locking **or
teaching the store graph semantics**", and is ticked. But `normalizeTaskGraphPlan`
(`:118-196`) — which lives in `internal/store` — validates stable-id shape, self-edges,
duplicate edges, prerequisite existence, `updated_at` format, every deterministic write
prefix's acyclicity, and final graph health; and `materializeTaskGraphPlan` (`:242-252`)
hard-codes the canonical and legacy field names. `taskGraphFromMap` (`:198-204`) calls
`core.NewTaskGraph` from the store.

The dependency *direction* is fine — `store` imports `core`, never the reverse — so this
is layering, not a cycle. The practical cost lands on the next two tasks: bulk-linking
and the dependency commands will want to plan, validate, and preview graph deltas
*without* a filesystem, and today that logic is only reachable by taking the repository
lock.

**Recommendation:** extract a pure `core.ValidateGraphPlan(graph, plan)
(TaskGraphMutationPlan, error)` returning the normalized plan and the prefix/final
verdict, and leave the store with lock, read, materialize, CAS, write. That is a move,
not a rewrite, and it makes `MutateTaskGraph` usable by slice 3 without redesign.

**Resolution:** Source-health, dependency-set, every-prefix, and final-health
validation moved to pure
core.ValidateTaskGraphMutationSource/ValidateTaskGraphMutationPlan. Store
retains only authoritative load, exclusion, frontmatter materialization, CAS,
and atomic replacement.

#### M5. Acceptance criterion 6 is not met: lint still builds its own snapshot and `ReadTaskGraph` is still unused  · **Status:** fixed

**File:** internal/core/service.go:265-269; internal/core/service_task.go:84-99 | **Component:** core / architecture
**Effort:** S · **Urgency:** soon

**[follow-up]**

AC6: "The store boundary consumes one canonical strict-snapshot loader; remove or fold
any duplicate or otherwise unused `ReadTaskGraph` scan seam so lint and mutation cannot
drift." Ticked. Actually:

- `Service.Lint` still scans via `ListTasksWithBodies` and constructs
  `NewTaskGraph(taskRecords, taskProblems)` itself (`service.go:265-269`). It does not
  call `LoadTaskGraph`. Two independent definitions of the authoritative snapshot remain
  — the drift the criterion says was closed.
- `ReadTaskGraph` was refactored to delegate to `LoadTaskGraph` but not removed, and a
  grep across `internal` and `cmd` finds **no caller**. It is still the unused seam the
  criterion names.
- `LoadTaskGraph`'s parameter is an inline anonymous interface (`service_task.go:92-94`)
  rather than a named port, so nothing type-checks that lint and mutation agree on the
  source.

The two paths agree today only because `ListTasks` and `ListTasksWithBodies` share
`scanDir` and `parseTask`.

**Recommendation:** name the loader's source as a small port
(`type TaskSource interface { ListTasks() … }`), have `Service.Lint` obtain its graph
from `LoadTaskGraph`, and delete `ReadTaskGraph`. Then untick or re-scope AC6.

**Resolution:** The unused Service.ReadTaskGraph seam was removed and
LoadTaskGraph now consumes a named TaskGraphSource. Lint's body-bearing scan and
mutation loading both feed the same NewTaskGraph strict projection and shared
store parser without adding a duplicate filesystem scan.

#### M6. Graph writes no longer stamp `updated_at`; a pure planner has no clock to supply one  · **Status:** fixed

**File:** internal/store/graphmutation.go:253-255; internal/core/store.go:64 | **Component:** store / write contract
**Effort:** S · **Urgency:** soon

**[follow-up]**

`updated_at` is applied only when the planner puts a value in
`TaskDependencyWrite.UpdatedAt` (`:253-255`). Every other mutation path in the store
injects `now` itself (`SetFields`, `Move`, `EditBody`, `EditTask`, `SetEpicFields`,
`SetResearchFields`). A dependency change is therefore the one mutation that can leave
a file's `updated_at` stale, and the responsibility sits with a callback the whole
design defines as pure — planners run in `core` and have no clock.

Executed:

```
TestADV_UpdatedAtNotStampedByStore
  after dependency write, beta frontmatter contains updated_at "2026-01-01": true
```

**Why existing tests miss it:** `TestMutateTaskGraphOwnsSemanticReadValidateWriteBoundary`
always passes `UpdatedAt: "2026-08-27"` and asserts it lands.
`TestMutateTaskGraphReturnsDurablePrefixAndRerunConverges` omits it and asserts nothing
about it — so the suite contains a passing example of the defect.

**Recommendation:** take `now time.Time` on `MutateTaskGraph` (matching `EditTask`,
`Move`, `Defer`) and stamp it in `materializeTaskGraphPlan` for every write; drop
`UpdatedAt` from `TaskDependencyWrite`, or keep it only as an explicit override.

**Resolution:** MutateTaskGraph now requires the caller's time and the store
stamps updated_at whenever graph-owned fields semantically change.
TaskDependencyWrite remains clock-free, and an idempotent rerun on a later date
neither writes nor advances the timestamp.

#### L1. Two `core.Store` port methods have no re-entry check and are callable from a planner  · **Status:** fixed

**File:** internal/store/fix.go:25; internal/store/danglers.go:18; internal/core/store.go:229, 237 | **Component:** store / re-entry guard
**Effort:** XS · **Urgency:** eventually

**[follow-up]**

Enumerating every exported `FS` method, exactly three lack `rejectGraphPlannerCall`:
`WatchPaths` (pure path arithmetic, harmless), `DanglingLinks`, and `FixFrontmatter`.
The latter two are declared members of the `core.Store` port
(`store.go:229, 237`). `FixFrontmatter(false)` is covered transitively because it takes
`writeLock`; `FixFrontmatter(true)` takes no lock, and `DanglingLinks` takes none at all.

Executed from inside a live planner:

```
TestADV_UnguardedStorePortMethodsInPlanner
  planner called FixFrontmatter(dryRun) -> 0 results, err=<nil>
  planner called DanglingLinks()        -> 0 results, err=<nil>
  mutation err=<nil>
```

Neither writes, so nothing corrupts — but AC3 says the contract "permits no nested Store
calls", and these two are the exceptions.

**Recommendation:** add the two guard calls; add a reflection or grep-based test that
every method in the `core.Store` port surface begins with the check, so the invariant
survives the next port addition.

**Resolution:** FixFrontmatter, including dry-run, and DanglingLinks now
participate in canonical-root callback exclusion. The nested-access regression
covers both methods through a second FS in addition to core Store
reads/writes/mutation.

#### L2. `SameSourceSnapshot` compares problem and legacy *counts*, not identity, and over-claims in its doc  · **Status:** fixed

**File:** internal/core/dependency_graph.go:735-750 | **Component:** core / CAS
**Effort:** S · **Urgency:** eventually

**[follow-up]**

The function ends with `len(g.problems) == len(other.problems) && len(g.legacy) == len(other.legacy)`.
Unreadable files never enter `ids`, `tasks`, or the path comparison, so two genuinely
different repositories compare equal whenever their problem counts match:

```
TestADV_SameSourceSnapshotFalsePositive
  left.health=broken right.health=broken
  SameSourceSnapshot(different unreadable file sets) = true
```

**This is currently unreachable through `MutateTaskGraph`**, which refuses a broken
snapshot before planning (`graphmutation.go:56-59`), and a *newly* unreadable file
between the two loads changes the problem count from 0 and is caught. I verified that
reasoning holds. The defect is in the exported API and its doc comment — "reports
whether two graphs came from the same exact task files … Health and paths catch new
unreadable/renamed entities" — which invites a future long-lived-service or query
consumer to rely on more than the code delivers.

**Recommendation:** compare a sorted digest of `(code, taskID, path, message)` for
problems and `(taskID, field, values)` for legacy, or narrow the doc comment to exactly
what is compared and note the broken-health precondition.

**Resolution:** SameSourceSnapshot compares exact deterministic problem and
legacy diagnostic content rather than counts. Coverage proves different
unreadable file sets with the same cardinality no longer compare equal.

#### L3. Entity creation into a not-yet-existing planning root now fails at the repository lock  · **Status:** fixed

**File:** internal/store/create.go:54-59; internal/store/lock_unix.go:22-24 | **Component:** store / compatibility
**Effort:** XS · **Urgency:** eventually

**[follow-up]**

`writeNewFile` now takes `s.writeLock()` before `writeNewFileUnlocked`, which is where
`os.MkdirAll(dir)` used to create the tree. `platformWriteLockChecked` opens `s.root`,
so the root must already exist:

```
TestADV_CreateIntoMissingRoot
  CreateTask into a not-yet-existing planning root ->
      open repo root for write lock: open …/fresh-planning-root: no such file or directory
  CreateTask into an existing but empty root -> <nil>
```

Not currently reachable through the CLI — repo discovery rejects a missing root first
("not a taskflow planning repo …"), which I confirmed — so this is a store-contract
change rather than a user-visible regression. It still contradicts AC5 ("ordinary write
behavior remains compatible") for any other adapter or test, and the error carries no
domain sentinel (see L4).

**Recommendation:** `os.MkdirAll(s.root, 0o755)` in `platformWriteLockChecked` before
opening, or in `writeNewFile` before locking. One line, and it restores the prior
contract.

**Resolution:** writeNewFile restores the previous missing-root contract by
creating the planning root before acquiring the directory-backed lock;
CreateTask coverage begins from a nonexistent root and verifies the entity
lands.

#### L4. Lock and unsupported-platform errors carry no domain sentinel  · **Status:** fixed

**File:** internal/store/lock_unix.go:23, 26; internal/store/lock_other.go:10; internal/store/lock_windows.go:23-46 | **Component:** store / error classification
**Effort:** XS · **Urgency:** eventually

**[follow-up]**

`CLAUDE.md` requires errors to wrap `ErrNotFound`/`ErrValidation`/`ErrAmbiguous`/
`ErrConflict` so the CLI maps them to exit codes 10/11/13/14. Every lock error is a bare
`fmt.Errorf`, including `lock_other.go`'s "repository mutation locking is unsupported on
this platform" — a permanent, categorical refusal that an agent cannot distinguish from
a transient I/O failure. AC4 says acquisition/release errors are "attributable"; they are
*described*, not classified.

**Recommendation:** wrap the unsupported-platform case in `domain.ErrValidation` (a
permanent refusal) and lock acquisition failures in `domain.ErrConflict` where they are
retryable; leave genuine I/O errors unwrapped.

**Resolution:** The categorical unsupported-platform refusal now wraps
ErrValidation. Genuine open/flock/close I/O failures retain their underlying OS
error and attributable operation context rather than being mislabeled as
retryable conflicts.

#### L5. Dry runs hold the exclusive repository guard and skip the whole-snapshot CAS  · **Status:** fixed

**File:** internal/store/graphmutation.go:37, 73-75 | **Component:** store / concurrency
**Effort:** S · **Urgency:** eventually

**[follow-up]**

`MutateTaskGraph(true, …)` takes `checkedWriteLock` like a real write and holds it
through load, plan, normalize (see N1), and materialize — then returns at `:73-75`
*before* the CAS re-read. So a preview blocks every writer in the repository for the
full validation cost while itself making no durability claim about the plan it prints.

Executed: `TestADV_DryRunHoldsExclusiveGuard` observes the planner running inside the
guard on a dry run.

**Recommendation:** keep the lock (a consistent snapshot needs it) but document that
`--dry-run` is exclusive, and note in the result that a dry-run plan is not
CAS-validated. If contention matters later, a shared/read lock mode is the fix.

**Resolution:** ADR-0006 and ARCHITECTURE now state that graph dry-run
intentionally holds the exclusive guard for an internally consistent
authoritative preview, skips pre-apply CAS because it writes nothing, and makes
no durability claim about a later invocation.

#### L6. `repositoryMutexes.byRoot` never evicts  · **Status:** fixed

**File:** internal/store/lock.go:19-22, 38-49 | **Component:** store / resource use
**Effort:** XS · **Urgency:** eventually

**[monitor]**

Entries are created on first use and never removed. For the CLI this is bounded by one
root per process. For the long-lived multi-repository adapter the file's own comment
cites as motivation, it is an unbounded map keyed by attacker-independent but
unbounded path strings.

**Recommendation:** none now — note it beside the comment so the future adapter's author
sees it. Reference-count and delete at zero if a service ever lands.

**Resolution:** The root-guard registry comment now records the CLI's one-root
bound and requires reference-counted idle eviction if a future long-lived
multi-space adapter makes the map unbounded.

#### L7. Lock tests use a fixed 100 ms timeout as the success branch  · **Status:** fixed

**File:** internal/store/lock_unix_test.go:78, 126 | **Component:** testing
**Effort:** XS · **Urgency:** eventually

**[follow-up]**

Both serialization tests treat `case <-time.After(100 * time.Millisecond)` as *proof*
that the second acquirer is blocked. A goroutine that is merely slow to be scheduled —
a loaded CI runner, `-race`, a cold `t.TempDir()` — produces the same observation, so
the assertion can pass without the property holding. I ran the concurrency tests
`-race -count=20` with no failures, so there is no flake today; the concern is a
vacuous pass, not an intermittent one.

**Recommendation:** have the second acquirer signal "about to block" before calling, and
assert the ordering of acquisition (e.g. a shared counter) rather than the absence of an
event within a wall-clock budget.

**Resolution:** The same-process mutex test uses TryLock state rather than
elapsed time. The child-process test uses a nonblocking flock probe against a
product-held lock; fixed delay is no longer the success assertion.

#### N1. Prefix validation rebuilds the whole repository graph once per planned write, inside the exclusive lock  · **Status:** tracked by 6g3q4rtv8d0a

**File:** internal/store/graphmutation.go:183-187, 198-204 | **Component:** store / performance
**Effort:** M · **Urgency:** soon

**[monitor]**

`normalizeTaskGraphPlan` calls `taskGraphFromMap` — a full `core.NewTaskGraph` over
*every* task in the repository, including `computeSound` and state derivation — once per
planned write. Cost is Θ(writes × repository size), all of it inside `checkedWriteLock`.
Measured (dry run, so this is validation cost alone):

| repo tasks | planned writes | `MutateTaskGraph` |
|---|---|---|
| 200 | 20 | 16.5 ms |
| 200 | 100 | 36.6 ms |
| 500 | 100 | 83.3 ms |
| 1000 | 100 | 162.6 ms |
| 1000 | 300 | 442.1 ms |

Linear in each factor, as expected. Today's planning repository is 279 tasks, so a
single `task depend add` is imperceptible. The consumer to watch is
`bulk-link-existing-tasks-into-threads-with-resumable-apply`: a 1000-write manifest over
a 2000-task space extrapolates to several seconds of lock-held CPU, which also widens
M1's clobber window proportionally.

**Recommendation:** don't optimize yet — but before bulk-linking, validate prefixes
incrementally (maintain one mutable adjacency map and re-check only reachability from
the changed node), or batch prefix checks per connected component. Corroborates sibling
`6g45qfv27vrm` M2; the measured table and the M1 interaction are additional.

---

**Resolution:** The bulk-link task now gates release on representative
end-to-end guarded-path benchmarks, a lock-latency budget, and incremental
prefix validation when measured O(W × (V+E)) cost is material; ADR-0006 records
the same sequencing decision.

## Assumptions that survived review

Each was attacked with an executed counterexample attempt, a mutation, or a subprocess
test — not read off the implementation notes.

1. **Cross-process serialization of `MutateTaskGraph` is real.** Two genuine child
   processes racing opposite edges with a 150 ms in-guard window: one applied, one was
   rejected with an attributable cycle diagnosis, final graph healthy with exactly one
   edge. This is the single most important claim in the slice and it holds. (It is also
   untested in the committed suite — see H4.)
2. **Cycle prevention for concurrent opposite edges works, in-process and across
   processes.** The loser re-reads inside the guard, sees the winner's edge, and its
   plan is refused.
3. **Canonical-root identity is correct on Unix.** `repositoryLockKey` maps a symlinked
   alias and a relative spelling to the same key, and a lock held on the real path
   blocks an `FS` opened on the alias. Independently, `flock` on the root inode makes
   Unix alias-safe regardless of the key.
4. **Panic and termination paths release everything.** A planner panic propagates while
   the deferred unlock releases both the platform lock and the process mutex and clears
   the sentinel; subsequent writes succeed. A killed process's flock is reclaimed.
   Verified by the suite and by follow-on operations in my probes.
5. **An injected release failure is attributed without leaving the guard held** — the
   error names "release repository graph mutation guard" and the next mutation succeeds.
6. **Persistence tokens do not leak.** `SourceVersion` is `yaml:"-"`, cleared by
   `TaskGraph.Task()`, absent from `wire.TaskJSON`, and absent from the JSON Schema —
   no golden file or `schema_version` change was needed, which I confirmed by the empty
   golden diff. The mutation battery shows a committed test kills its removal.
7. **The whole-snapshot CAS genuinely catches a raw edit outside the write set**, and
   the mutation battery confirms a test kills `SameSourceSnapshot`. Within
   `MutateTaskGraph` its count-based tail is unreachable (L2 explains why).
8. **Prefix validation genuinely prevents a cyclic durable prefix.** Replacing it with a
   final-graph-only check fails a committed test.
9. **Interruption, rerun, and idempotence work.** An empty plan is a clean no-op; a
   re-applied plan produces byte-identical files and an empty applied set; a plan
   interrupted after one durable write leaves a healthy graph and converges on rerun.
10. **Fail-closed on unhealthy graphs is consistent.** A broken snapshot is refused
    before the planner runs; a single unrelated legacy `blocked_by` anywhere in the
    repository refuses every ordinary dependency mutation with an actionable message
    ("1 legacy dependency field occurrence(s) remain; run the guarded migration").
11. **Dependency direction is clean.** `store` imports `core`; `core` never imports
    `store`. `TaskGraphMutationStore` is deliberately a sibling capability, not part of
    `Store`, so read-only and test adapters are unaffected. (The *placement* of
    validation logic is M4; the direction itself is right.)
12. **Rejecting rather than no-op'ing on unsupported platforms is the right call**, and
    `internal/store` does compile for `GOOS=js` and `GOOS=plan9`, so the branch is
    reachable in principle even though the full module does not build there.
13. **No data races or flakes.** `go test -race ./...` clean; the concurrency tests
    clean at `-race -count=20`.

---

## Traceability table

| Code | Severity | Classification | Falsifies | Destination |
|---|---|---|---|---|
| H1 | High | **Pre-merge blocker** | AC3; `docs/ARCHITECTURE.md` re-entry sentence | Reopen `6g3q4rt0wzkq` — move the sentinel onto the root-keyed guard |
| H2 | High | **Pre-merge blocker** | AC3, AC5 | Reopen `6g3q4rt0wzkq` — identify the planner goroutine; others wait |
| H3 | High | **Pre-merge blocker** | AC1 ("tested") | Reopen `6g3q4rt0wzkq` — mutex-killing test + `repositoryLockKey` table test |
| H4 | High | **Pre-merge blocker** | AC1, AC4 | Reopen `6g3q4rt0wzkq` — re-point subprocess tests at `writeLock`; drop `platformWriteLock` |
| M1 | Medium | Follow-up | — | New task: interleave verify/write in the apply loop |
| M2 | Medium | Follow-up | — | New task: removals-before-additions apply order |
| M3 | Medium | Follow-up | AC1 | New task: repo-local lock file **or** declare Windows unsupported; add a Windows CI job |
| M4 | Medium | Follow-up | AC2 | Fold into `6g3q4rt7mgjn` — extract `core.ValidateGraphPlan` |
| M5 | Medium | Follow-up | AC6 | Reopen `6g3q4rt0wzkq` — lint consumes `LoadTaskGraph`; delete `ReadTaskGraph` |
| M6 | Medium | Follow-up | — | Reopen `6g3q4rt0wzkq` — store stamps `updated_at` from an injected clock |
| L1 | Low | Follow-up | AC3 | New task: guard `FixFrontmatter`/`DanglingLinks`; port-surface invariant test |
| L2 | Low | Follow-up | — | ADR-0006 / doc amendment on `SameSourceSnapshot`'s exact contract |
| L3 | Low | Follow-up | AC5 | Reopen `6g3q4rt0wzkq` — `MkdirAll` the root before locking |
| L4 | Low | Follow-up | AC4 | New task: classify lock errors against the domain sentinels |
| L5 | Low | Follow-up | — | Doc-only: `--dry-run` is exclusive and not CAS-validated |
| L6 | Low | **Monitor** | — | Note beside `lock.go:19-22` for the future long-lived adapter |
| L7 | Low | Follow-up | — | Fold into H3's test work |
| N1 | Medium | **Monitor** | — | Revisit before `bulk-link-existing-tasks-into-threads-with-resumable-apply` |

Acceptance criteria that should be unticked pending amendment: **1** (H3, H4, M3),
**2** (M4), **3** (H1, L1), **5** (H2, L3), **6** (M5).

---

## Validation commands and results

Run from `/Users/andyeschbacher/git/andy-esch/taskflow-graph-mutation-guard`, branch
`feat/portable-graph-mutation-guard`, uncommitted tree, merge-base `509a7c3`. Probe
files lived outside the repository and were supplied with `go test -overlay=`;
`git status --porcelain` was identical before and after apart from this audit.

| Command | Result |
|---|---|
| `go build ./...` | pass |
| `go test ./...` | pass, 23 packages |
| `go test -race ./...` | **pass, exit 0**, no race reports |
| `go test -race -count=20 -run 'TestMutateTaskGraphConcurrentOppositeEdges\|TestRepositoryLock\|TestMutateTaskGraphRejectsPlannerStoreCalls' ./internal/store/` | pass, 6.3 s, no flakes |
| `just lint` (`golangci-lint run ./...`) | **0 issues** |
| `go vet ./...` | pass |
| `GOOS=windows GOARCH=amd64 go build ./...` / `go vet ./...` | pass (compile-only) |
| `GOOS=windows go test ./internal/store/` | cannot run: `exec format error` — no Windows runtime coverage exists (M3) |
| `GOOS=js GOARCH=wasm go build ./internal/store/` | pass (`lock_other.go` branch compiles); full module does not build on js/plan9 for unrelated TUI deps |
| `go mod tidy` | no change (`golang.org/x/sys` correctly promoted to a direct requirement) |
| `./bin/tskflwctl lint` | exit **0**, 6 advisory legacy findings |
| Mutation: `processRepositoryLock` → no-op | **suite still PASSES** — H3 |
| Mutation: `repositoryLockKey` → no `EvalSymlinks` | **suite still PASSES** — H3 |
| Mutation: `rejectGraphPlannerCall` → nil | suite FAILS (caught) |
| Mutation: `SameSourceSnapshot` → true | suite FAILS (caught) |
| Mutation: `Task()` keeps `SourceVersion` | suite FAILS (caught) |
| Mutation: prefix check → final graph only | suite FAILS (caught) |
| Probe: planner writes via a second `*FS` | **hangs > 3 s, no error** — H1 |
| Probe: planner reads via a second `*FS` | returns **live** state ≠ snapshot — H1 |
| Probe: unrelated goroutine `GetTask`/`SetFields` during a mutation | `ErrConflict` "planner cannot call Store methods" — H2 |
| Probe: two real child processes, opposite edges | 1 applied / 1 rejected, final healthy, 1 edge — survived claim 1 |
| Probe: raw edit between verify and write, 2-file plan | **silently clobbered, mutation reported success** — M1 |
| Probe: edge reversal, both id orientations | smaller-id dependent accepted; larger-id rejected — M2 |
| Probe: dependency write with no `UpdatedAt` | `updated_at` left at `2026-01-01` — M6 |
| Probe: `FixFrontmatter(true)` / `DanglingLinks()` inside a planner | both succeed, no re-entry rejection — L1 |
| Probe: `CreateTask` into a missing root | `open repo root for write lock: … no such file or directory` — L3 |
| Probe: `SameSourceSnapshot` on differing unreadable sets | `true` — L2 |
| Probe: symlinked alias root, relative vs absolute root | one key, alias blocks behind the real path — survived claim 3 |
| Probe: prefix-validation scaling (200–1000 tasks × 20–300 writes) | 16.5 ms → 442 ms, linear in each factor — N1 |
| Probe: empty plan / repeated plan / one legacy field present | clean no-op / byte-identical / refused with actionable message — survived claims 9, 10 |

---

## Candidate tasks

- ⏳ Reopen `6g3q4rt0wzkq` — pre-merge set: H1 (root-keyed re-entry sentinel), H2 (planner-goroutine identity so unrelated callers wait), H3 (tests that die when the mutex or the key normalization dies), H4 (test the product lock path across processes; delete `platformWriteLock`).
- ⏳ Reopen `6g3q4rt0wzkq` — small corrections: M5 (lint consumes `LoadTaskGraph`; delete `ReadTaskGraph`), M6 (store stamps `updated_at`), L3 (`MkdirAll` the root before locking), L7 (replace wall-clock success branches).
- ⏳ `tskflwctl task new "Interleave graph-mutation CAS with the apply loop" --epic 30-threads-and-task-dependency-graphs --tags graph,storage,concurrency` — M1, with a `testHookAfterGraphWrite` regression test.
- ⏳ `tskflwctl task new "Apply graph plans removals-first so edge reversal is order-independent" --epic 30-threads-and-task-dependency-graphs --tags graph,storage` — M2, plus a test that runs both id orientations.
- ⏳ `tskflwctl task new "Settle the Windows repository lock identity" --epic 30-threads-and-task-dependency-graphs --tags storage,concurrency,portability` — M3: repo-local lock file or an explicit unsupported declaration, and a Windows CI job.
- ⏳ `tskflwctl task new "Extract graph plan validation into a pure core validator" --epic 30-threads-and-task-dependency-graphs --tags graph,architecture` — M4; prerequisite for `6g3q4rt7mgjn` and bulk-linking.
- ⏳ `tskflwctl task new "Guard every Store port method against planner re-entry" --epic 30-threads-and-task-dependency-graphs --tags storage,concurrency` — L1, with a port-surface invariant test.
- ⏳ `tskflwctl task new "Classify repository lock errors against the domain sentinels" --epic 30-threads-and-task-dependency-graphs --tags storage,errors` — L4.
- ⏳ ADR-0006 amendment — L2 (`SameSourceSnapshot`'s exact contract and its broken-health precondition) and L5 (`--dry-run` is exclusive and not CAS-validated).
- ⚠️ N1 and L6 — monitor only; revisit N1 before `bulk-link-existing-tasks-into-threads-with-resumable-apply` lands.
