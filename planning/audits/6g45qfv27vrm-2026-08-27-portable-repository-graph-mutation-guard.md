---
schema: 1
id: 6g45qfv27vrm
bucket: closed
area: portable-repository-graph-mutation-guard
date: "2026-08-27"
updated_at: "2026-08-27"
---

# Audit: Portable Repository Graph Mutation Guard — 2026-08-27

Adversarial implementation-readiness review of the portable repository graph mutation guard in Taskflow (task `6g3q4rt0wzkq`, epic `30-threads-and-task-dependency-graphs`, branch `feat/portable-graph-mutation-guard`), evaluated against ADR-0006, `docs/ARCHITECTURE.md`, upcoming dependency mutations (`6g3q4rt7mgjn`), and multi-platform concurrency requirements.

**Executive Verdict: Safe with amendments.** The core control-inverted boundary is well-designed: the store owns the repository-wide critical section and atomic file operations while invoking pure core planners; `core.LoadTaskGraph` unifies diagnostic and mutation snapshot loading; `SourceVersion` enables whole-repository CAS without leaking persistence tokens to planners or wire contracts; and panic/termination paths cleanly release locks. However, two high-severity defects must be amended before production usage:
1. **Destructive write re-sorting (H1):** `normalizeTaskGraphPlan` forcefully re-sorts multi-file write plans by arbitrary `TaskID` strings (`graphmutation.go:124`), destroying planner-provided topological write sequences. When an atomic graph update involves both edge additions and removals, arbitrary ID sort order can place an addition ahead of a removal, causing intermediate prefix validation to fail with a spurious cycle error and rejecting valid plans.
2. **Re-entry sentinel scope gaps (H2):** Tracking `plannerActive` on the specific `*FS` instance struct (`lock.go:51`) leaves a self-deadlock vulnerability if a planner indirectly uses a secondary `*FS` instance for the same root (e.g. via a sub-service), and causes legitimate concurrent reads on a shared `*FS` instance in multi-goroutine environments to fail spuriously with `ErrConflict`.

Findings are classified as **[pre-merge blocker]**, **[pre-mutation prerequisite]**, or **[tracked follow-up]**.

---

## Findings

### High

#### H1. Enforcing arbitrary `TaskID` sort order on multi-file write plans rejects valid prefix-safe mutations · **Status:** fixed

**File:** internal/store/graphmutation.go:124-126, 183-187 | **Component:** store / graph mutation
**Effort:** M · **Urgency:** acute

**[pre-merge blocker / pre-mutation prerequisite]**

`normalizeTaskGraphPlan` unconditionally re-sorts `plan.TaskWrites` by `TaskID` string (`lines 124-126`):
```go
sort.Slice(normalized.TaskWrites, func(i, j int) bool {
    return normalized.TaskWrites[i].TaskID < normalized.TaskWrites[j].TaskID
})
```
Each intermediate prefix in that sorted order is then validated against `prefixGraph.Health() == core.GraphBroken` (`lines 183-187`).

When an atomic graph mutation contains both edge additions and edge removals (for example, reversing a dependency `A -> B` into `B -> A`, or restructuring a cyclic subgraph into a DAG), the planner knows the safe topological execution sequence: execute removals first (breaking the old edge), then additions (introducing the new edge).

Because `normalizeTaskGraphPlan` discards the planner's write ordering and enforces lexical `TaskID` ordering:
1. If the dependent task's random Crockford-32 ID happens to sort before the prerequisite's ID, the addition write is evaluated before the removal write.
2. Prefix 1 introduces the reverse edge while the old edge still exists, creating a temporary cycle `A <-> B`.
3. `normalizeTaskGraphPlan` aborts with `planned write prefix ending at task ... would leave a broken graph: dependency cycle`.
4. The success or failure of valid multi-file mutations becomes entirely dependent on random task ID hash collisions.

**Why tests missed it:** `TestMutateTaskGraphRejectsBrokenIntermediatePrefixBeforeWriting` intentionally constructed a test where ID sorting caused a cycle, but concluded that "a planner with an unsafe mixed rewrite must split it into prefix-safe operations" rather than allowing the planner to supply the deterministic safe write order.

**Recommendation:** Respect the planner-supplied write order in `TaskGraphMutationPlan` by default, validating each prefix in the planner's provided sequence. Alternatively, automatically partition writes into a removal wave followed by an addition wave when ordering multi-file operations.

**Resolution:** Core now preserves the planner-provided deterministic write
order and validates every supplied prefix; regression tests prove
removal-before-addition edge reversal succeeds while the unsafe order is
rejected before writing.

#### H2. Instance-local `plannerActive` sentinel allows self-deadlock via secondary `FS` instances and rejects legitimate concurrent reads · **Status:** fixed

**File:** internal/store/lock.go:51-69, 87-95, internal/store/fsstore.go:49 | **Component:** store / concurrency & re-entry
**Effort:** M · **Urgency:** acute

**[pre-merge blocker]**

`plannerActive` is stored as an `atomic.Bool` field on the specific `*FS` instance (`fsstore.go:49`). Every Store method checks `s.rejectGraphPlannerCall()` (`lock.go:64-69`).

This design has two symmetric flaws:

1. **Self-Deadlock via Secondary Store Instance:** If a planner callback invokes code or a constructor that creates a new `*FS` instance rooted at the same planning directory (for example, `fs2 := store.NewFS(s.root)` or a helper service), `fs2.plannerActive` is `false`. When `fs2` attempts any write operation (such as `fs2.SetFields`), `fs2.rejectGraphPlannerCall()` passes. `fs2.checkedWriteLock()` then calls `processRepositoryLock(s.root)`, which attempts to acquire the non-reentrant `sync.Mutex` in `repositoryMutexes.byRoot[key]`. Because that mutex is already held by the same goroutine in `fs1`, the process **self-deadlocks and hangs permanently**.
2. **Spurious Conflict on Unrelated Concurrent Reads:** In a long-lived process (such as a future daemon, web server, or TUI with background refresh routines) sharing a single `*FS` instance across goroutines, while Goroutine 1 is executing a planner inside `fs.MutateTaskGraph`, any concurrent read in Goroutine 2 (e.g. `fs.GetTask`, `fs.ListTasks`, `fs.ResolveTaskPath`) calls `fs.rejectGraphPlannerCall()` and fails with `ErrConflict` ("graph mutation planner cannot call Store methods"). Non-locking concurrent reads should not be rejected simply because another goroutine is planning a mutation.

**Why tests missed it:** `TestMutateTaskGraphRejectsPlannerStoreCallsAndNestedMutationWithoutHanging` only tested nested Store calls made directly on the exact same `fs` variable from within the planner goroutine.

**Recommendation:** Associate the planner re-entry check with the repository lock state for that canonical root key (or goroutine-local context) rather than an instance-wide struct field. Ensure read methods that do not take the write lock are permitted from separate goroutines.

---

**Resolution:** Planner activity is now recorded on the canonical-root guard, so
same- and second-FS calls fail promptly rather than deadlocking or escaping the
snapshot. ADR-0006 explicitly adopts root-wide fail-fast contention because Go
cannot reliably distinguish callback and unrelated goroutines without a new
capability-bearing port.

### Medium

#### M1. Windows lock file placement in `UserCacheDir` fails cross-user / multi-user repository concurrency · **Status:** fixed

**File:** internal/store/lock_windows.go:29-38 | **Component:** store / platform lock (Windows)
**Effort:** S · **Urgency:** soon

**[pre-mutation prerequisite / platform]**

On Windows, `platformWriteLockChecked` locates the lock file under `os.UserCacheDir()` (`%LocalAppData%\tskflwctl\locks\<hash>.lock`).

If a planning repository on a shared drive, multi-user workstation, or CI agent is accessed by multiple OS user accounts or service accounts, each user resolves a different `UserCacheDir`. User A acquires `LockFileEx` on `C:\Users\Alice\AppData\Local\...`, while User B acquires `LockFileEx` on `C:\Users\Bob\AppData\Local\...`. The OS-level locking guarantee is completely bypassed across user boundaries.

**Why tests missed it:** Tests run in single-user Unix and CI environments.

**Recommendation:** Locate repository lock files in a repository-local directory (such as `meta/.lock` or `.taskflow.lock`) or use a Windows Named Mutex (`windows.CreateMutex` with a global namespace name derived from the canonical repository path).

**Resolution:** Windows no longer claims an untested per-user cache lock
guarantee. The non-Unix build fails closed with ErrValidation until a cross-user
lock identity has native runtime coverage; macOS/Linux remain the supported
release matrix.

#### M2. Quadratic graph reconstruction $\mathcal{O}(K \times (V+E))$ during prefix validation limits scalability of bulk operations · **Status:** tracked by 6g3q4rtv8d0a

**File:** internal/store/graphmutation.go:183-187, 198-204 | **Component:** store / performance & scale
**Effort:** M · **Urgency:** soon

**[tracked follow-up / monitor]**

In `normalizeTaskGraphPlan`, for every write $i \in [1..K]$ in a plan, `taskGraphFromMap` constructs a full `NewTaskGraph` (`graphmutation.go:203`). Each `NewTaskGraph` executes full topological wave analysis, Tarjan strongly connected components cycle detection, and sound completion memoization over all $V$ tasks in the repository.

For a bulk operation involving $K$ writes in a repository with $V$ tasks and $E$ edges, prefix validation takes $\mathcal{O}(K \times (V + E))$ time while holding the exclusive repository write lock. For large planning repositories ($V > 2,000$) and large bulk manifests ($K > 200$), this introduces noticeable write-lock holding times.

**Why tests missed it:** Unit test suites tested small graphs with $V \le 5$ and $K \le 2$.

**Recommendation:** Use incremental cycle and edge validation for intermediate prefix checks, reserving full `NewTaskGraph` construction for the initial snapshot and the final graph state.

---

**Resolution:** The bulk-link task now requires representative guarded-path
benchmarks, an explicit lock-latency budget, and incremental prefix validation
if the measured O(W × (V+E)) cost is material.

### Low

#### L1. Unreadable file changes during planning are only coarsely detected by slice length in `SameSourceSnapshot` · **Status:** fixed

**File:** internal/core/dependency_graph.go:751 | **Component:** core / graph snapshot
**Effort:** XS · **Urgency:** eventually

**[tracked follow-up]**

`SameSourceSnapshot` verifies task file byte hashes via `SourceVersion`, but checks `problems` only by slice length: `len(g.problems) == len(other.problems)`. If an unreadable file `tasks/bad1.md` is replaced by a different corrupted file `tasks/bad2.md` during planning, the problem count remains unchanged.

While per-file CAS and final reload mitigate corrupted writes, `SameSourceSnapshot` should strictly match problem file paths and codes.

**Why tests missed it:** CAS test cases tested task content edits, deletions, and additions, rather than same-count unreadable file replacements.

**Recommendation:** Compare `problems` slice contents (or problem file paths) in `SameSourceSnapshot`.

---

**Resolution:** SameSourceSnapshot now compares the full deterministic
graph-problem and legacy-diagnostic identities, including paths, codes,
messages, cycles, resolutions, candidate IDs, and projected edges; a
same-count/different-unreadable-set regression test fails the old behavior.

## Assumptions That Survived Adversarial Review

These claims and invariants were independently verified under adversarial conditions and found sound:

1. **Control-Inversion Boundary & Pure Core Separation:**
   - Store exclusively owns filesystem locking, snapshot loading, YAML parsing/frontmatter surgery, and atomic replacement (`writeFileAtomic`).
   - Planner callbacks receive immutable `*core.TaskGraph` projections and return pure `core.TaskGraphMutationPlan` values; no graph library types or persistence handles leak into domain or wire layers.
2. **Whole-Snapshot CAS via `SourceVersion`:**
   - Every scanned task receives an internal SHA-256 byte hash (`SourceVersion`).
   - `SameSourceSnapshot` verifies that all task files and paths remain byte-identical before writes begin.
   - `graph.Task(id)` strips `SourceVersion`, preventing persistence tokens from leaking to planners, YAML, or wire DTOs.
3. **Resumable Multi-File Atomic Replacements:**
   - Every task file is written atomically via temporary files and `os.Rename`.
   - `TaskGraphMutationResult.AppliedTaskIDs` accurately records the durable applied prefix if an error or interruption occurs.
   - Re-running the planner over a partial prefix converges idempotently (`TestMutateTaskGraphReturnsDurablePrefixAndRerunConverges`).
4. **Panic and Error Lock Recovery:**
   - Defers in `MutateTaskGraph` (`graphmutation.go:41-50`) guarantee that `unlock()` is executed even if the planner callback panics or returns an error (`TestMutateTaskGraphPlannerPanicReleasesGuard`).
5. **Unix Lock Semantics & Process Termination Cleanup:**
   - macOS and Linux use root-directory `flock(LOCK_EX)` with an in-process mutex.
   - Subprocess termination tests confirm the OS immediately releases locks upon unexpected process termination (`TestRepositoryLockReleasesWhenProcessTerminates`).
6. **Explicit Platform Rejection for Unsupported Targets:**
   - Non-Unix, non-Windows platforms (e.g. WebAssembly) fail explicitly with `"repository mutation locking is unsupported on this platform"`, avoiding silent no-op CAS loss.
7. **Entity Creation Guard Participation:**
   - `CreateTask`, `CreateAudit`, `CreateEpic`, and `CreateResearch` now acquire `writeLock()`, preventing creation collisions and race conditions against concurrent mutations.
8. **Legacy Migration Support:**
   - `MutateTaskGraph` accepts `GraphDegraded` snapshots and allows `ClearLegacy: true` writes to transition the repository to `GraphHealthy` in one guarded operation.
   - Broken graph snapshots (`GraphBroken`) are rejected before the planner is invoked.

---

## Sequencing & Architectural Critique

1. **Readiness for Slice 2 (`6g3q4rt7mgjn`):** `TaskGraphMutationStore` provides the exact mutation boundary needed by `task depend add/remove` and legacy migration. However, resolving **H1** (write ordering) is a strict prerequisite so that multi-edge and reverse-edge operations do not fail on prefix cycle checks.
2. **Integration with `core.LoadTaskGraph`:** Moving snapshot construction to `core.LoadTaskGraph` successfully unifies diagnostic reads and mutations, eliminating scan drift.

---

## Traceability Table

| Finding | Severity | Classification | Action / Target Destination |
|---|---|---|---|
| **H1** (Destructive write re-sorting) | High | Pre-merge blocker | Reopen `6g3q4rt0wzkq` / fix in `graphmutation.go` to preserve planner write order |
| **H2** (Instance-local `plannerActive` sentinel) | High | Pre-merge blocker | Reopen `6g3q4rt0wzkq` / bind re-entry check to root lock state |
| **M1** (Windows `UserCacheDir` isolation) | Medium | Pre-mutation prerequisite | Tracked by `6g3q4rt0wzkq` Windows platform task |
| **M2** ($\mathcal{O}(K \times (V+E))$ prefix rebuilding) | Medium | Tracked follow-up / monitor | Tracked by bulk-operations performance task |
| **L1** (Problem comparison in `SameSourceSnapshot`) | Low | Tracked follow-up | Tracked by follow-up CAS refinement |

---

## Validation Commands and Results

All checks executed in worktree `/Users/andyeschbacher/git/andy-esch/taskflow-graph-mutation-guard` on branch `feat/portable-graph-mutation-guard`:

1. **Full Test Suite:**
   ```bash
   go test ./...
   ```
   *Result:* Passed (all 25 packages passed).

2. **Race Detection Test Suite:**
   ```bash
   go test -race ./...
   ```
   *Result:* Passed (all packages passed with zero data races detected).

3. **Store Package In-Depth Tests:**
   ```bash
   go test -v ./internal/store/...
   ```
   *Result:* Passed (120+ tests including fuzzing, concurrent opposite edges, and crash recovery passed).

4. **Audit Lint Validation:**
   ```bash
   tskflwctl audit lint 2026-08-27-portable-repository-graph-mutation-guard
   ```
   *Result:* Passed.

## Candidate tasks

- ⏳ `tskflwctl task new "Preserve planner write order in graph mutation prefix validation" --epic 30-threads-and-task-dependency-graphs --tags storage,graph` — Stop forcing TaskID lexical sort order on TaskWrites in normalizeTaskGraphPlan (H1)
- ⏳ `tskflwctl task new "Bind planner re-entry check to root lock state instead of FS instance" --epic 30-threads-and-task-dependency-graphs --tags storage,concurrency` — Fix secondary FS self-deadlock and avoid rejecting concurrent reads from separate goroutines (H2)
- ⏳ `tskflwctl task new "Fix cross-user repository lock path on Windows" --epic 30-threads-and-task-dependency-graphs --tags storage,windows` — Use repo-local lock file or global named mutex instead of UserCacheDir (M1)
