---
schema: 1
id: 6g6qvrj15x97
bucket: open
area: concurrency-and-atomicity
date: "2026-09-04"
updated_at: "2026-09-04"
---

# Code Quality Audit: concurrency-and-atomicity — 2026-09-04

> Edit findings through `tskflwctl audit finding` so status and resolution
> metadata stay queryable. Never hand-edit a `**Status:**` line.

Routine: `code-quality-audit` · lens `concurrency-and-atomicity` · ISO week `2026-W36`,
slot `Fri` (index 1).

## Punch list

One line per finding for fast triage. Order: Critical → High → Medium → Low.

- `H1`. `task ac` silently discards a concurrent body write  (effort: S · urgency: acute)
- `M1`. `task ac` and `audit finding` are the only scriptable mutations without OCC auto-retry  (effort: XS · urgency: soon)
- `M2`. Repository lock acquisition is unbounded and undiagnosed  (effort: S · urgency: eventually)
- `L1`. `store.writeFileAtomic` replaces a symlink instead of its target  (effort: S · urgency: eventually)
- `L2`. `materializeTaskGraphPlan` misattributes an absent task as `ErrConflict`  (effort: XS · urgency: eventually)

## Files audited

- **Signal**: `internal/store/lifecyclemutation.go` — highest non-test churn in the lens
  surface (4 commits/30d); every task status change enters here from the CLI, the TUI, and
  guarded create-and-start.
- **Signal**: `internal/store/graphmutation.go` — 3 commits/30d; owns the exclusive
  canonical-root planner window that every concurrent `Store` call collides with.
- **Adjacency** (from `lifecyclemutation.go`, across the core↔store seam): `internal/core/retry.go`
  — the bounded OCC retry that wraps these mutations. The lens asks directly about "retry
  loops that can retry a *committed* write", so the guard belongs beside the guarded thing.
- **Adjacency** (from `graphmutation.go`, to the guard it acquires): `internal/store/lock.go`
  (with `lock_unix.go` / `lock_other.go`).
- **Random**: `internal/store/threadstore.go` (drawn from the lens surface; the first draw,
  `internal/store/mdlinks.go`, was outside the lens and redrawn).

Read for context, not audited as picks: `store/cas.go`, `store/atomic.go`,
`store/threadapply.go`, `store/body.go`, `store/edit.go`, `store/resolve.go`,
`core/finding.go`, `core/service_task.go`, `core/service_thread.go`,
`core/dependency_operations.go`, `core/service_thread_apply.go`, `core/thread_apply.go`,
`core/dependency_graph.go`, `tui/read_retry.go`, `tui/model.go`, `tui/watch.go`.

## Commands run

- `go build -o bin/tskflwctl ./cmd/tskflwctl` — clean (`just` is not installed in the runner;
  the Justfile recipe is the same command plus ldflags).
- `go test ./...` — all packages pass. **Green before this run**; no pre-existing red.
- `go test -race ./internal/store/... ./internal/core/... ./internal/tui/...` — passes
  (`just test` is `go test -race ./...`).
- `./bin/tskflwctl audit list --json` — 1 open audit, so no backpressure triage.
- Churn: `git log --since="30 days ago" --name-only --pretty=format: -- internal/store internal/core internal/tui`.
- Blast radius: `grep -rln "<ExportedName>" internal cmd` per mutation-store symbol.
- `grep -rn "os.WriteFile\|os.Create(\|os.OpenFile" internal cmd --include='*.go'` — confirms
  no raw non-atomic write in a planning write path.
- `grep -rn "LOCK_NB\|Timeout\|deadline" internal/store/lock*.go` — no hit in production code;
  `LOCK_NB` appears only in `lock_unix_test.go`. This is the evidence for M2.

**Mutation probe for H1** (run against a throwaway planning repo under the session
scratchpad — nothing in this checkout was written): built a repo with 3 000 task files to
widen the resolve window, then raced `task append` against `task ac --check 1` twelve times.

```
=== runs=12  lost_appends=3 ===
run 6: APPEND LOST (PROGRESS-MARKER-6 not in file)
run 7: APPEND LOST (PROGRESS-MARKER-7 not in file)
run 8: APPEND LOST (PROGRESS-MARKER-8 not in file)
```

The same probe against the audit twin — `audit append` raced with `audit finding --status` —
lost **0 of 12** (6 of 12 correctly exited 14). That contrast is the finding and also the fix.

## Findings

### Critical

(none)

### High

#### H1. `task ac` silently discards a concurrent body write  · **Status:** open

**File:** `internal/core/service_task.go:205` (and `:268`, `:287`) | **Component:** core/store OCC
**Effort:** S · **Urgency:** acute

All three `task ac` write paths — `SetAcceptanceCriterion` (`--check`/`--uncheck`),
`setCriteriaState` (`--defer`/`--wontfix`/`--tracked`/`--na`), and `EditCriteria`
(`--add`/`--remove`/`--replace`) — share one shape:

1. `s.store.GetTask(slug)` reads the file (**read #1**) and returns `body`;
2. the transform runs in core, producing `newBody` from read #1;
3. `s.store.EditBody(slug, newBody, false, …)` reads the file **again** (read #2,
   `internal/store/body.go:129`), computes `ifVersion = hashContent(content)` from read #2,
   and replaces the **whole body** with the bytes derived from read #1.

The content-hash CAS therefore guards only read #2 → write. Anything committed between
read #1 and read #2 passes the CAS and is then overwritten by the stale body. Because
`appendMode` is `false`, this is a full-body replacement: the concurrent writer's work is
not merged, it is erased.

This is exactly the pattern the lens flags as high-yield — "reads a snapshot, validates
against it, and writes without re-checking the snapshot" — and it defeats the premise of
epic 24, whose whole point is that a lost update is impossible.

**Failing scenario:** task `T` has body `B0`. Agent A runs `tskflwctl task ac T --check 1`
and reads `B0`. Agent B runs `tskflwctl task append T --body "progress note"` and commits
`B0 + note`. Agent A's `EditBody` then reads `B0 + note`, hashes *that*, passes the CAS, and
writes `B0`-with-criterion-1-ticked. **Agent B's note is gone, and agent A exits 0.**
Measured at 3 of 12 races above; the window is two full resolve+read cycles wide, so it
grows with repository size.

**Why tests didn't catch it:** the OCC tests in `store/occ_test.go` drive two writers
through *one* store method and assert the loser is told. No test spans the core-level
read → store-level read gap, because from the store's own point of view the CAS did hold.

**Recommendation:** move the transform inside the CAS window, exactly as the audit twin
already does. `core.EditFinding` (`internal/core/finding.go:263`) passes a **callback** to
`store.EditAudit` and re-applies `edit.apply(current, code)` to `current` — the content the
store read under its own `ifVersion`. Give the task body-replace path the same callback
shape (or thread read #1's hash through as an explicit `ifVersion`) so the criterion
transform is computed from CAS-covered bytes. Then wrap it in `retryOnConflict` (see M1) so
the loser re-derives rather than failing.

**Tightening (adjacent):** `AcceptanceCriteria` (`:191`) is read-only and fine, but the
three writers duplicate the read→transform→`EditBody` sequence three times; folding them
onto one helper is what makes the fix land once instead of three times.

**Follow-up:** survey any other core use case that computes final bytes from its own read
and hands them to a store writer that re-reads. The task at
`planning/tasks/6g392b0rps7w-…` explicitly deferred that survey ("Extending CAS to any other
surface that currently lacks it … is its own task"); this finding is its first hit.

### Medium

#### M1. `task ac` and `audit finding` are the only scriptable mutations without OCC auto-retry  · **Status:** open

**File:** `internal/core/finding.go:245` · `internal/core/service_task.go:267` | **Component:** core
**Effort:** XS · **Urgency:** soon

`retryOnConflict` (`internal/core/retry.go:62`) exists so scriptable agents don't
reimplement a read→re-apply→rewrite loop. It wraps `AppendBody`, `SetFields`, `MoveAudit`,
`AppendAuditBody`, `AppendResearchBody`, and the epic writers. It does **not** wrap
`EditFinding`/`SetFindingStatus` or any of the `task ac` writers — both instead ride paths
built for the interactive `$EDITOR` loop, where refusing to retry is correct because a human
is watching.

Both `task ac` and `audit finding` are documented in `CLAUDE.md` as the *agent* face of a
closed vocabulary ("never hand-edit them; `task ac` owns them"), and this routine's own
step 10 tells agents to run `audit finding` in a loop. Under contention they surface exit 14
where their siblings self-heal.

**Failing scenario:** two cron agents resolving findings in the same audit. In the probe
above, `audit finding --status` exited 14 on **6 of 12** races against a concurrent
`audit append` — correct behaviour for the interactive path, but an agent scripting
`audit finding` over a punch list sees half its calls fail and must implement its own
backoff, which is the duplication `retryOnConflict` was introduced to prevent.

**Why tests didn't catch it:** exit 14 is a *correct* outcome, so no assertion is violated.
The gap is a policy inconsistency between sibling verbs, not a broken invariant.

**Context (severity vs urgency):** graded Medium rather than High because nothing is lost —
the CAS holds and the caller is told. It outranks M2 on urgency because agents hit it today.

**Recommendation:** wrap both in `retryOnConflict`. For `audit finding` this is the second
acceptance criterion of the existing task `6g392b0rps7w` and needs no new design. For
`task ac` it must land **after** H1 — retrying a lost-update path just loses the update on a
later attempt.

**Follow-up:** the `attempted` closure flag at `internal/core/finding.go:262` is the
workaround that stands in for the mechanism; it should disappear with the retry, as that
task already specifies.

#### M2. Repository lock acquisition is unbounded and undiagnosed  · **Status:** open

**File:** `internal/store/lock_unix.go:26` | **Component:** store guard
**Effort:** S · **Urgency:** eventually

`platformWriteLockChecked` calls `syscall.Flock(fd, syscall.LOCK_EX)` with no `LOCK_NB`, no
timeout, and no progress signal; `processRepositoryLock` (`lock.go:63`) likewise blocks on a
bare `sync.Mutex`. Every guarded mutation therefore waits an unbounded time to *acquire*
the guard, with no output.

This contradicts the design intent stated one file over: `retry.go:11` explains the retry
bound as "a genuinely contended file should surface a loud `ErrConflict` (exit 14) quickly,
not spin". The CAS retry is bounded and loud; the lock wait ahead of it is unbounded and
silent. `grep` confirms `LOCK_NB` appears only in `lock_unix_test.go`, never in production.

**Failing scenario:** `MutateThreadApply` holds the guard across repeated full-repository
scans — `LoadTaskGraph` plus `listThreadApplyThreads`, re-run by `reprepareThreadApply`
after every changed dependency prefix. On a large planning tree over a network or FUSE mount
(a case `atomic.go:66` explicitly contemplates), a concurrent `tskflwctl task start` prints
nothing and appears hung for the holder's whole duration. `flock` releases on process death,
so an agent's own timeout kills the waiter safely — the cost is the silence, not corruption.

**Why tests didn't catch it:** `lock_unix_test.go` asserts that the lock *is held* (using
`LOCK_NB` as a probe) and that waiters eventually proceed within a 5s deadline. Nothing
asserts anything about what a waiter reports while waiting, because today it reports nothing.

**Context (severity vs urgency):** Medium severity but `eventually` urgency — I could not
construct a *stuck* holder in today's code (planners are pure, and the `$EDITOR` session is
deliberately outside the lock), so this is a robustness and observability gap rather than a
live hang.

**Recommendation:** acquire with `LOCK_EX|LOCK_NB` first; on `EWOULDBLOCK`, emit one
"waiting for the repository lock…" line to the injected `io.Writer` and fall back to the
blocking acquire, optionally under a deadline that returns `ErrConflict`. Keep the blocking
path as the default so cooperating serialization is unchanged.

### Low

#### L1. `store.writeFileAtomic` replaces a symlink instead of its target  · **Status:** tracked by 6g63jj1dh0sb

**File:** `internal/store/atomic.go:49` | **Component:** store atomic write
**Effort:** S · **Urgency:** eventually

`writeFileAtomic` carries the destination's mode forward with `os.Stat(path)`, which
**follows** a symlink, then commits with `os.Rename(tmp, path)`, which **does not**. A
symlinked entity file would therefore be converted into a regular file, leaving the link
target stale. `internal/config/config.go:1148` does the same job correctly
(`atomicWriteDestination`: `Lstat` → `EvalSymlinks` → rename onto the resolved target) — and
its comment at `config.go:1146` asserts "**The store has the same idea**", which is not true.

**Failing scenario:** none reachable today, and I checked rather than assumed:
`markdownDoc` (`internal/store/resolve.go:23`) gates both `scanDir` and `flatCandidates` on
`e.Type().IsRegular()`, so a symlinked task/audit/epic/research/thread file is never listed
or resolved, and `writeFileAtomic` is never handed one. The defect is latent: the safety
depends entirely on that `IsRegular()` gate, which `atomic.go` never mentions. Relaxing
`markdownDoc` — a plausible change if symlinked entity files are ever wanted for cross-repo
sharing — would silently turn this into data loss.

**Why tests didn't catch it:** `atomic_test.go` never constructs a symlinked destination,
and no test would, since the scanner cannot produce one.

**Recommendation:** adopt `config`'s `Lstat`/`EvalSymlinks` destination resolution, or — if
the store deliberately refuses symlinked entities — say so in a comment at `atomic.go:48`
and correct the false cross-reference at `config.go:1146`.

**Resolution:** Fully covered by
unify-the-three-divergent-writefileatomic-implementations, whose comparison
table already records store/atomic.go as non-symlink-preserving (os.Stat).
Latent only: markdownDoc gates entity scans on IsRegular(), so a symlinked
entity file is never resolved today.

#### L2. `materializeTaskGraphPlan` misattributes an absent task as `ErrConflict`  · **Status:** open

**File:** `internal/store/graphmutation.go:132` | **Component:** store graph mutation
**Effort:** XS · **Urgency:** eventually

`task, _ := graph.Task(planned.TaskID)` discards the `ok`. For a task absent from the
snapshot, `task.Path` is `""`, so the very next guard (`if path != task.Path`) reports
`task %s changed path during graph snapshot: ErrConflict` — telling the caller to retry a
condition no retry can fix. The lifecycle sibling handles the identical case correctly:
`prepareTaskLifecycleMaterialization` (`lifecyclemutation.go:186`) checks `ok` and returns
`ErrNotFound`.

**Failing scenario:** not reachable through the guarded API — `ValidateTaskGraphMutationSource`
rejects a non-strict snapshot and `ValidateTaskGraphMutationPlan` rejects a plan naming an
unknown task, so both gates fire first. This is a defense-in-depth and consistency finding,
reported as a smell rather than a defect. Kept because it sits in a signal-picked file and is
adjacent to H1's theme: the exit code the CLI shows must name the real condition.

**Recommendation:** check the `ok` and return `ErrNotFound`, mirroring the lifecycle sibling.

## What audited clean

This lens found the write *transactions* in very good shape; the defects are all at their
edges. Specifically:

- **`internal/store/lifecyclemutation.go`** — the transaction shape is right: guard →
  authoritative snapshot → pure plan → materialize → whole-snapshot CAS over **both** the
  graph and the Thread set → per-target CAS → atomic write. Two details are better than they
  had to be: `ensureTaskIDNotThread` is deliberately re-checked *after* the CAS (`:115`), and
  the create path leans on `writeNewFileUnlocked`/`O_EXCL` rather than trusting its own
  earlier `os.Stat`, so the stat is a courtesy and not a TOCTOU.
- **`internal/core/retry.go` and all five guarded retry loops** — I went looking specifically
  for the lens's "retry a committed write" bug and did not find it. Every loop guards:
  `runTaskLifecycleMutation` and `runThreadMutation` and `runThreadCreationMutation` on
  `!result.Committed`, `runDependencyMutation` on `len(result.AppliedTaskIDs) > 0`, and
  `ApplyThreadPlan` on `result.Committed || result.Complete`. Each also captures `now` once
  *outside* its loop, which is what keeps a retry idempotent rather than re-stamping dates.
- **`internal/store/graphmutation.go`** (aside from L2) — the two-tier CAS is correct: a
  whole-snapshot preflight before the first file, then a per-file check that keeps the
  raw-editor window bounded to one replacement after a durable prefix exists. The
  resumable-not-atomic contract is stated honestly at `:22` rather than being quietly assumed.
- **`internal/store/lock.go`** (aside from M2) — the keyed process mutex plus `flock` is a
  sound combination, and `enterRepositoryPlanner` converts what would be a self-deadlock into
  an attributable `ErrConflict`. `normalizeRepositoryLockKey` canonicalizes through
  `EvalSymlinks`+`Abs`+`Clean` so two spellings of one root share a guard.
- **`internal/store/threadstore.go`** (the random pick) — nothing notable, and the reason is
  structural: it performs no mutation, and `parseThread` stamps `SourceVersion` from the exact
  bytes it parsed, which is what lets `sameThreadSourceSnapshot` fail **closed** on an empty
  token (`threadcreation.go:189`). Both whole-snapshot comparators treat a missing version as
  a mismatch rather than a match — the right default.
- **`internal/store/threadapply.go`** — I specifically checked the unchecked index at `:164`
  (`finalDecision.Operations[len-1]`) and it is safe: `PrepareThreadApply` appends exactly one
  `Kind: "thread"` operation last on every successful return path. It also explicitly guards
  the silent `bytes.Equal → continue` skip that `materializeTaskGraphPlan` leaves implicit
  (`:82`), which is the more careful of the two call sites.
- **`internal/tui/read_retry.go` + `model.go`** — the stale-result story holds. The
  generation guard `readRequestCurrent` is checked on **both** `readConflictMsg` and
  `readRetryMsg`, and the outer `sessionMsg` stamp drops results from a superseded workspace
  *before* the surface guard is consulted, so a retry closure bound to an old `core.Service`
  cannot deliver into a new session. Only the first conflict per request is held, so a
  persistently contended repository surfaces a durable error instead of spinning.
- **`internal/store/atomic.go`** (aside from L1) — `writeFileAtomic` and `createFileAtomic`
  are correct: temp + fsync + rename + best-effort directory fsync, `O_EXCL` for creates, and
  cleanup on *every* failure path including a failed `Close`. No raw `os.WriteFile` hides in
  any planning write path (verified by grep across `internal` and `cmd`).

## External research

Not done — three Medium-or-higher findings were surfaced (H1, M1, M2), so step 11 does not
apply.

## Candidate tasks (human to triage)

- `tskflwctl task new "Close the lost-update window in the task ac write path" --epic 21-code-quality-architecture-hardening --tags store,core,occ,concurrency --tier 1 --priority high --description "task ac computes the new body from a GetTask read that no CAS covers, then hands finished bytes to EditBody, which CASes only its own later read; a concurrent body write is silently erased."`
- `tskflwctl task new "Emit a waiting signal and a non-blocking first attempt for the repository lock" --epic 21-code-quality-architecture-hardening --tags store,concurrency,ux --tier 3 --priority medium --description "flock(LOCK_EX) has no LOCK_NB, timeout, or progress output, so a waiter appears hung; the bounded CAS retry beside it is loud and fast by design."`
- `tskflwctl task new "Return ErrNotFound rather than ErrConflict for a task missing from the graph-mutation snapshot" --epic 21-code-quality-architecture-hardening --tags store,errors --tier 4 --priority low --description "materializeTaskGraphPlan discards graph.Task()'s ok, so an absent task is reported as a path change; the lifecycle sibling checks it."`
- ⚠️ M1 partially tracked in `planning/tasks/6g392b0rps7w-give-finding-status-writes-a-non-interactive-cas-path.md` — that task owns the `audit finding` half (its `retryOnConflict` acceptance criterion is still open); the `task ac` half is new and depends on H1.
- ⏳ L1 tracked in `planning/tasks/6g63jj1dh0sb-unify-the-three-divergent-writefileatomic-implementations.md`

## Related-task observations (propose-only)

- **Partly already-shipped, and one premise now false:**
  `planning/tasks/6g392b0rps7w-give-finding-status-writes-a-non-interactive-cas-path.md`.
  Its objective says `editFile` "does not verify the content hash. A concurrent automated
  write between read and write is a lost update." That is no longer true: `store.EditAudit`
  computes `ifVersion := hashContent(orig)` (`edit.go:295`) and passes `verifyUnchanged`
  (`:301`) under `s.writeLock`. My probe raced `audit finding` against `audit append` twelve
  times and lost **zero** writes, with six correct exit-14s — so its third acceptance
  criterion ("a conflicting concurrent write surfaces `domain.ErrConflict` … rather than
  silently winning or losing") already holds. What remains live is the `retryOnConflict`
  wrapper and the `attempted` closure flag. Suggest re-scoping the description to the retry
  gap; **not changing status/tier/priority here.**
- **Scope conflict:** the same task lists `SetCriterionState` among the mutations that
  "instead perform a content-hash compare-and-swap wrapped in `retryOnConflict`". Finding H1
  shows `SetCriterionState` does neither, and is in fact the worse offender of the two. The
  task's comparison should not be used as evidence that the `task ac` path is safe.
