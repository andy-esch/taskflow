---
schema: 1
id: 6gn83987gt45
bucket: open
area: arch-failure-and-recovery
date: "2026-10-19"
updated_at: "2026-10-19"
---
# Weekly Architecture Audit: failure-and-recovery — 2026-10-19

> Create findings through `tskflwctl audit finding new`; update them through
> `tskflwctl audit finding` so identity, status, resolution, and managed candidate
> metadata stay synchronized. Never hand-edit those fields.
> Architecture audits are propose-only — no code, docs, or ADR edits.

Routine: `weekly-architecture-audit` · lens `failure-and-recovery` · ISO week
`2026-W43` (mod 4 = 3). On-schedule Monday run.
ADRs consulted: ADR-0006 (§2, amendments 2026-08-27 ×2, 2026-08-28, 2026-09-05),
ADR-0003 (rename amendment), ADR-0008 (§agent obligations, §3 classification) ·
Files read: `internal/store/atomic.go`, `lock.go`, `lock_unix.go`, `lock_other.go`,
`cas.go`, `graphmutation.go`, `lifecyclemutation.go`, `graphrepair.go`,
`threadapply.go`, `threadcreation.go`, `threadmutation.go`, `fix.go`,
`internal/core/retry.go`, `mutation_policy.go`, `dependency_operations.go`,
`dependency_repair.go`, `internal/cli/exit.go`, `fserror.go`, `moves.go`,
`internal/tui/read_retry.go`, `internal/config/config.go`,
`internal/userconfig/paths.go` · Sources cited: 6

## Executive summary

This lens was last audited four weeks ago (`2026-09-21-arch-failure-and-recovery`,
now closed, 4/4 findings done) and the substrate has held up: the guard topology,
the write-order-as-recovery-data contract, the whole-snapshot plus per-file CAS
coverage, and the typed partial-durability receipts are all intact, and every
`writeFileAtomic` call site in `internal/store` still has a paired CAS verify.
The one new defect is an internal inconsistency inside the broken-graph repair
path: `MutateTaskGraphRepair`'s preflight preserves a genuine filesystem read
error, but its per-write loop collapses five such errors into `domain.ErrConflict`
and discards the cause (H1). That is the *same* defect class that shipped task
`6g1dpb5gwtt9` fixed in `cas.go` on 2026-09-02, surviving in the one capability
that is only ever reached when the repository is already broken. The second
finding (M1) is the per-envelope verification that open design task
`6g7wsr8yvt1a` Q5 explicitly asked for, and is tracked there. Two divergences
found during the comparison — the per-filesystem `flock` guarantee and the
three-way `writeFileAtomic` durability split — are already owned by open tasks
and are deliberately **not** re-raised as findings here. Read H1 first; it is
small, local, and in actively-changed code.

## State of the architecture (this lens)

**The durable write primitive.** `writeFileAtomic` (`internal/store/atomic.go:48-62`)
stages a temp file in the destination's own directory, `fsync`s it, `chmod`s it,
renames it over the target, then best-effort `fsync`s the parent directory
(`syncDir`, `atomic.go:68-73`). Overwrite preserves the destination's existing
mode rather than resetting to `perm` (`atomic.go:49-51`), so a user-tightened
`0600` task is not silently widened. `createFileAtomic` (`atomic.go:84-126`) is a
separate primitive: `O_CREATE|O_EXCL` *is* the `ifVersion == ""` CAS precondition,
so creates need no companion verify, and every in-process failure path
(write/chmod/sync/close) removes the partial destination. Crash orphans under the
tool's own `.tskflwctl-*.tmp` prefix are swept by age ≥ 1h (`atomic.go:142-158`),
reachable only from `lint --fix` (`fix.go:170-183`).

**The guard.** Two layers at one canonical-root scope. `repositoryGuards`
(`lock.go:28-31`) is a process-wide map of per-root `sync.Mutex`, keyed through
`normalizeRepositoryLockKey` (`lock.go:37-49`: `EvalSymlinks` → `Abs` → `Clean`,
lowercased on Windows). `platformWriteLockChecked` (`lock_unix.go:21-39`) adds
`flock(LOCK_EX)` on the repository root directory; `lock_other.go:14-16` refuses
mutation outright on non-Unix rather than no-op locking, because "Silent no-op
locking would make a successful CAS claim untrue." `checkedWriteLock`
(`lock.go:115-137`) orders invocation authorization → planner-reentry rejection →
process mutex → platform lock, and returns release failures so a control-inverted
boundary can attribute them. `writeLock` (`lock.go:103-109`) is the legacy form
that discards the release error.

**Planner exclusion.** `enterRepositoryPlanner` / `rejectRepositoryPlannerCall`
(`lock.go:69-98`) mark a callback window at canonical-root scope; every Store
entry point checks it, so re-entry becomes an attributable `ErrConflict` rather
than a self-deadlock. The flag is cleared by `defer leave()` even on planner panic
(`graphmutation.go:118`).

**CAS.** Two tiers. Whole-snapshot equality compares complete source evidence
including unreadable revisions and duplicate-ID shadows
(`verifyTaskGraphSourceSnapshot`, `verifyThreadSourceSnapshot`, `cas.go:33-97`;
`SourceVersion != ""` fails closed on missing evidence). Per-target equality is
`verifyUnchanged` (`cas.go:107-141`), which re-resolves through an explicitly
supplied resolver key and compares a SHA-256 of exact bytes computed on read and
never persisted (`hashContent`, `cas.go:25-28`). Crucially, `verifyUnchanged`
distinguishes drift from failure: `ErrNotFound` and `os.IsNotExist` become
`ErrConflict`, while ambiguity and other read errors keep their own sentinels
(`cas.go:117-135`).

**Guarded mutation shape.** Five families share one skeleton — guard, authoritative
snapshot load, pure core planner over an immutable snapshot, plan validation,
materialization, whole-snapshot CAS, per-target CAS, atomic write, prefix recorded
into a typed receipt: `MutateTaskGraph` (`graphmutation.go:26-111`),
`MutateTaskLifecycle` (`lifecyclemutation.go:20-140`), `MutateThreadCreation`
(`threadcreation.go`), `MutateThread` (`threadmutation.go`), `ApplyThreadPlan`
(`threadapply.go`), plus `MutateTaskGraphRepair` (`graphrepair.go`) as the only
capability admitting `GraphBroken`. Multi-file plans are explicitly resumable
rather than transactional (`graphmutation.go:21-25`).

**Retry.** `retryOnConflict` (`core/retry.go:62-72`) re-runs only on
`domain.ErrConflict`, bounded at 4 attempts (`retry.go:15`) with full-jitter
backoff capped at 50ms (`retry.go:23-31`), never for a dry run. The graph and
repair families hand-roll the same loop with an added durability gate:
`dependency_operations.go:159-163` stops the moment `len(result.AppliedTaskIDs) > 0`,
and `dependency_repair.go:649` stops on `result.Committed || len(result.AppliedSources) > 0`.

**Error surfacing.** `domain.Classify` feeds a four-row class→code table
(`cli/exit.go:53-60`) into a seven-active-code process taxonomy
(`exit.go:33-47`; 12 published-but-reserved). `WriteError` (`exit.go:137-222`)
assembles the `--json` failure envelope: a class/path/retryable filesystem block
from `*fs.PathError`/`*os.LinkError` (`fserror.go:21-63`) plus up to eight typed
recovery receipts. Three of those branches additionally force
`Filesystem.Retryable = false` once their receipt reports durable work
(`exit.go:162`, `:180`, `:196`). The TUI resolves the contention half separately:
`withReadConflictRetry` (`tui/read_retry.go:57-68`) holds the *first* conflict per
read generation and retries once, explicitly noting that a conflict from a
*mutation* is not transient at all.

**Batch semantics.** `runMoves` (`cli/moves.go:32-88`) attempts every slug as its
own guarded transaction, renders a per-item result, and returns a summary error
that prefers a sentinel-bearing cause so the exit code names the most actionable
failure rather than the first in argv.

## ADR reconciliation

### ADR-0006 §2 — the scan, validation and write must be one guarded critical section

**Quoted:** "The final repository scan, graph validation, and write must occur
inside one planning-repository mutation guard. Per-file optimistic concurrency
alone is insufficient: two individually valid concurrent edge additions can form a
cycle when combined."
**Implementation:** follows — `MutateTaskGraph` takes `checkedWriteLock` before
`core.LoadTaskGraph` and holds it through the last write
(`graphmutation.go:38-109`). Same ordering in all five sibling families.
**Class if divergent:** n/a.

### ADR-0006 amendment 2026-08-27 §1 — planner write order is recovery data

**Quoted:** "A pure core validator preserves the planner-provided order,
canonicalizes only semantic sets such as `depends_on`, and rejects the complete
plan before writing unless every supplied prefix and the final state are sound."
**Implementation:** follows — `ValidateTaskGraphMutationPlan` runs before
materialization (`graphmutation.go:65-68`), and the applied prefix is recorded
write-by-write into `result.AppliedTaskIDs` (`graphmutation.go:103`).
**Class if divergent:** n/a.

### ADR-0006 amendment 2026-08-27 §2 — planner exclusion is scoped to the canonical root

**Quoted:** "During the callback, every Store entry point for that root—including
through a second `FS`—fails fast with `ErrConflict`. … a future TUI/server adapter
should treat it as retryable contention or deliberately revise the port."
**Implementation:** follows, and the clause's "future" is now present. The TUI
takes the first branch deliberately and documents why
(`tui/read_retry.go:42-56`): a read landing inside a planner window "has not found
a broken repository — it has found a busy one," so the first conflict per read
generation is retried once and a still-contended repository surfaces an ordinary
durable error instead of spinning.
**Class if divergent:** n/a — worth noting as an ADR clause whose deferred
decision has since been made in code rather than left implicit.

### ADR-0006 amendment 2026-08-27 §3 — only runtime-tested platforms claim mutation support

**Quoted:** "Only runtime-tested release platforms claim mutation support. macOS
and Linux use the canonical-root process mutex plus root-directory `flock`, with
real same-process and child-process tests over the production path. Windows and
other non-Unix source builds fail closed."
**Implementation:** drift in a known, already-tracked way. The gate is `GOOS`
(`lock_unix.go` build tag `//go:build unix` vs `lock_other.go`), but `flock(2)`'s
cross-process guarantee is a property of the *filesystem*, not the OS. Nothing in
`internal/`, `docs/` or `planning/adrs/` mentions NFS, FUSE, `local_lock` or
synced mounts in a locking context — while `atomic.go:64-67` explicitly
anticipates "network and FUSE mounts in particular" for durability.
**Class if divergent:** ADR gap — **already raised and tracked.** This is
finding M3 of `2026-09-21-arch-failure-and-recovery`, tracked by
`planning/tasks/6g6yjm16dkgt-design-bounded-and-observable-repository-lock-acquisition.md`.
Deliberately not re-raised as a new finding this run; re-verified as still
accurate against current code.

### ADR-0006 amendment 2026-08-27 (dependency-operation contracts) §4 — receipts distinguish convergence from success

**Quoted:** "A failure after a durable prefix carries that prefix in typed human
and JSON diagnostics rather than collapsing it into a prose-only error."
**Implementation:** follows at the receipt layer — `DependencyMutationReceipt`
carries `PlannedTaskIDs`/`AppliedTaskIDs`/`RemainingTaskIDs` plus a core-owned
`Remedy` that is explicitly "not permission to retry"
(`core/dependency_operations.go:40-55`), and all eight receipts reach the error
envelope (`cli/exit.go:164-218`). **Silent at the hint layer:** the ADR says
nothing about whether the generic transient-filesystem `retryable` flag in the
same envelope must agree with the receipt. See M1.
**Class if divergent:** ADR gap.

### ADR-0006 amendment 2026-09-05 — broken-graph repair operates on source declarations

**Quoted:** "`task depend repair` is a separate mutation capability and the only
supported command family that may begin from `GraphBroken`. Repair is
removal-only, source-declaration-addressed, guarded by the repository lock plus
whole-snapshot/per-file CAS, and validated by separate progress and preservation
proofs. … Receipts must make residual problems and partial durability explicit."
**Implementation:** follows structurally — `graphrepair.go:104-158` reauthorizes
each next source group against a freshly read durable snapshot, proves
componentwise non-regression per step, and advances `AppliedSources`/`Committed`
plus `Residual` into the receipt. **Diverges on error fidelity:** the per-step
reauthorization reports a genuine filesystem read failure as a *conflict*, so the
receipt is explicit about durability while the error it travels with misnames its
own cause. See H1.
**Class if divergent:** implementation drift.

### ADR-0008 — agent-facing obligations

**Quoted:** "The machine contract is designed for agents as well as conventional
programs. That creates additional obligations beyond naming a version: return the
durable identifier a later operation needs while retaining readable names, allow
bounded projections, and report actionable structured errors."
**Implementation:** partial. The exit taxonomy is closed, published and
self-describing via `schema` (`exit.go:122-129`), and `filesystemDetails`
deliberately reads `*fs.PathError`/`*os.LinkError` structure "rather than matching
on message text" (`fserror.go:12-20`). Both H1 and M1 are failures of this clause
specifically: H1 removes the structured filesystem block entirely by dropping the
`PathError` from the chain, and M1 lets two fields of one envelope contradict each
other.
**Class if divergent:** implementation drift (H1) / ADR gap (M1).

### Not governed by any ADR

The **error-class → exit-code mapping itself** is owned by `internal/cli/exit.go`
and its test, not by an ADR. ADR-0008 §3 names "error code" among the things whose
change is `NOT ADDITIVE`, which governs *changing* a code's meaning, but no
accepted decision states the predicate for *assigning* a class — i.e. when a
failure is a conflict rather than an I/O error. That absence is the soil H1 grew
in: the invariant exists, is honoured in `cas.go:117-135` and in
`graphrepair.go:91-98`, and is enforced nowhere.

## Best-practice comparison

### 1. Wrap to expose, and let callers inspect the cause

**Source:** <https://go.dev/blog/go1.13-errors> (blocked by this environment's
network policy; quoted via search excerpt, not fetched)
**Guidance:** `fmt.Errorf`'s `%w` makes the returned error "have an `Unwrap`
method returning the argument of `%w`", and the rule of thumb is to "wrap an error
to expose it to callers, and do not wrap an error when doing so would expose
implementation details."
**This codebase:** partial. `%w` is used universally and `domain.Classify` routes
on sentinels rather than text. But `graphrepair.go:110-157` wraps
`domain.ErrConflict` *instead of* the error it actually caught, which is neither
wrapping nor deliberate hiding — the cause is destroyed, so no caller can inspect
it. Every sibling family does this correctly (`threadapply.go:115-123`,
`threadcreation.go:88-95`, `threadmutation.go:89-96`,
`lifecyclemutation.go:105-112`), as does `graphrepair.go`'s own preflight
(`:91-98`).
**Justified?** No. There is no project constraint that makes the repair path's
errors less worth preserving than the lifecycle path's — the opposite, since
repair runs on an already-damaged tree.

### 2. A permanent failure must stop a retry loop, with its cause intact

**Source:** <https://pkg.go.dev/github.com/cenkalti/backoff/v4>
**Guidance:** `Permanent(err)` produces a `*PermanentError` that "signals that the
operation should not be retried"; `Retry` then does not retry and "the wrapped
error is returned". The wrapper implements `Unwrap()` and `Is()`, so the cause
survives the classification.
**This codebase:** partial, and in the inverse direction. taskflow's retry
predicate is "is this a conflict" rather than "is this permanent", which is a
sound simplification for one bounded local workload — `retryOnConflict`
(`core/retry.go:67`) retries *only* `ErrConflict`, and non-conflict errors return
immediately, so permanent failures stop by default. The divergence is that
`graphrepair.go` manufactures the one sentinel that *is* retryable from errors
that are permanent (EACCES, EIO, ENOSPC), inverting the default: `RepairTaskGraph`
(`core/dependency_repair.go:649`) then burns all four backoffs before failing.
**Justified?** No — this is the failure mode the cited library's `Permanent`
wrapper exists to prevent, and the one shipped task `6g1dpb5gwtt9` already
eliminated in `cas.go`.

### 3. Error responses to agents must be actionable, not opaque

**Source:** <https://www.anthropic.com/engineering/writing-tools-for-agents>
**Guidance:** "prompt-engineer your error responses to clearly communicate
specific and actionable improvements," "rather than opaque error codes or
tracebacks"; "error responses can steer agents towards more token-efficient
tool-use behaviors."
**This codebase:** partial. The envelope is unusually good — a closed error-code
vocabulary, a structured filesystem block with four genuinely distinct next
actions ("permission (grant access), not-found (create the parent / fix the path),
read-only (remount or choose another tree), no-space (free space)",
`fserror.go:17-20`), and core-owned prose remedies. Both findings degrade exactly
this: under H1 a permission-denied repair returns `code: "conflict"`, no
`error.filesystem` object at all, and the text "retry" — three signals, all wrong,
with no path to the actual next action. Under M1 an agent that trusts
`filesystem.retryable` can be told to repeat a call that already committed work.
**Justified?** No.

### 4. Write-fsync-rename, and fsync the directory

**Source:** <https://lwn.net/Articles/323430/> (2009 ext4 delayed-allocation
thread; commentary, not spec)
**Guidance:** the portable safe pattern is to write a new temp file, `fsync` it,
then rename it over the original — "you have to call fsync on the new file to keep
its data across a crash." POSIX is read as not requiring rename durability without
an `fsync` on the containing directory.
**This codebase:** follows in `internal/store` — `writeFileAtomic`
(`atomic.go:48-62`) performs all four steps, and `syncDir` (`:68-73`) tolerates
rather than assumes the directory `fsync`, which is the correct posture for the
FUSE and network mounts it names. Partial elsewhere: `internal/config/config.go:1205-1231`
and `internal/userconfig/paths.go:71-96` both `fsync` the temp and rename but omit
the directory `fsync`.
**Justified?** The divergence is deliberate and already owned. Task
`6g63jj1dh0sb-unify-the-three-divergent-writefileatomic-implementations.md`
documents the three-way table, establishes that atomicity and durability are
distinct guarantees most callers resolve differently, and measures the macOS
`F_FULLFSYNC` cost of unifying on the strictest behaviour. Not re-raised as a
finding here.

### 5. flock is advisory, per-open-file-description, and per-filesystem

**Source:** <https://manpages.ubuntu.com/manpages/jammy/man2/flock.2.html>
**Guidance:** "flock() places advisory locks only"; separate `open()` calls on one
file are "treated independently by flock()", so a lock "may be denied by a lock
that the calling process has already placed via another file descriptor"; "In
Linux kernels up to 2.6.11, flock() does not lock files over NFS", and a
compatibility mode exists — "see the discussion of the local_lock option in
nfs(5)".
**This codebase:** follows on the first two points, deliberately and in advance.
`lock_unix.go:21` opens a fresh descriptor per acquisition, which is precisely the
independent-open-file-description hazard; `repositoryGuards` (`lock.go:22-31`)
exists to neutralize it before any long-lived adapter can trigger it, and the
comment says so. Advisory-only is stated where it matters
(`lock_unix.go:15-17`: a raw hand-edit "isn't blocked, but the content hash still
catches that on the next verify"). Diverges on the third: see the ADR-0006 §3
reconciliation above — tracked, not re-raised.
**Justified?** Partly. Single-user local-first legitimises a whole-repo advisory
lock over per-row concurrency. It does not legitimise an unstated filesystem
precondition, which is why the existing task is the right owner.

### 6. A partially-completed operation is ambiguous to its caller unless the API says otherwise

**Source:** <https://stripe.com/blog/idempotency> (blocked by this environment's
network policy; quoted via search excerpt, not fetched)
**Guidance:** a call "could fail midway while the server is fulfilling the
operation, leaving the work in limbo"; in that state "the success of the operation
is ambiguous from the perspective of the client," which is what makes a blind
retry unsafe.
**This codebase:** follows at the receipt layer and is unusually explicit about
it — `Changed`/`Committed`/`Applied*`/`Remaining*` on every guarded receipt, a
core-owned `Remedy` string per outcome class
(`dependency_operations.go:211-222`), resumability promised only where the planner
is convergent (`graphmutation.go:21-25`), and `runMoves` attempting every item so a
batch never silently partials (`cli/moves.go:21-31`). Partial at the hint layer:
five of eight envelope branches let a transient `retryable: true` sit beside a
receipt reporting committed work. See M1.
**Justified?** No — and the project already agrees: the fix shipped for ordinary
creates (`exit.go:158-163`) and task `6g7wsr8yvt1a` Q5 asks for the remaining
envelopes to be verified. M1 is that verification.

## Tensions and trade-offs

**Whole-repository locking vs. agent throughput.** `flock` on the root directory
serializes every cooperating writer for the whole tree. For a single user and "a
handful of cron agents, not a thundering herd" (`core/retry.go:14`) this is
correct and the comment defends it explicitly. The cost surfaces only as the
4-attempt retry budget: because the budget is shared between genuine contention
and anything *misclassified* as contention, a classification defect like H1 does
not merely mislabel — it consumes the resource that real contention needs. That
coupling is why H1 is a recovery finding rather than a cosmetic one.

**Resumable-not-transactional multi-file writes.** The project deliberately
chose convergent, resumable multi-file plans over a transaction log or a staging
directory, and pays for it with a per-family durability predicate plus a receipt
vocabulary rich enough to describe every partial state. On a filesystem-backed
single-user store this is the right trade — a WAL would be a database in
disguise — but it is the direct cause of the "spelled four ways across seven
loops" problem that `6g7wsr8yvt1a` Q5 now owns. The architecture is sound; the
*uniformity* of its expression is the debt.

**Durability vs. macOS barrier cost.** `syncDir` is correct per the POSIX
reading, and on darwin `File.Sync()` issues `F_FULLFSYNC`, a whole-drive barrier.
A bulk `thread apply` therefore pays 2N full barriers on the maintainer's own
platform. Task `6g63jj1dh0sb` frames this honestly — the planning tree is
git-tracked, so losing the last write to a power cut is recoverable while showing
a half-written task file is not, which argues for atomicity everywhere and
durability only where chosen. Worth restating because it is the one place in this
lens where "follow the strictest best practice" is the wrong answer.

## Findings

<!-- Example grammar; let `audit finding new` allocate and render real findings: -->

```
#### H1. <title>  · **Status:** open

**File:** <path:line> | **Component:** <component>
**Effort:** <XS|S|M|L> · **Urgency:** <acute|soon|eventually>

<what's wrong, why it matters, evidence>

**Recommendation:** <minimum fix>

**Resolution:** <how it was resolved — written by `audit finding --note`, not by hand>
```

#### H1. Broken-graph repair reports genuine filesystem read failures as retryable conflicts, discarding the cause · **Status:** open

**File:** internal/store/graphrepair.go:110 | **Component:** store/graph-repair
**Effort:** S · **Urgency:** soon

`MutateTaskGraphRepair` collapses five distinct failure conditions into
`domain.ErrConflict` and throws the original error away:

- `graphrepair.go:110` — `evidenceErr` from `s.ReadThreads()`
- `graphrepair.go:115` — `evidenceErr` from `core.LoadTaskGraph(s)`
- `graphrepair.go:120-122` — `readErr` from `os.ReadFile(write.path)`
- `graphrepair.go:152` — post-commit `evidenceErr` from `s.ReadThreads()`
- `graphrepair.go:157` — post-commit `evidenceErr` from `core.LoadTaskGraph(s)`

Each is written `if evidenceErr != nil || verify…(…) != nil { return …
domain.ErrConflict }`, so a permission failure, an `EIO`, or a short read is
reported with the same sentinel and the same "retry" wording as a real
concurrent-edit conflict, and the cause is not wrapped at all.

These are not parse problems. `scanDir` is explicitly tolerant — "A parse failure
becomes a FileProblem (the file is skipped, not fatal) … only a real read error is
fatal" (`internal/store/resolve.go:113-117`) — so a non-nil error out of
`ReadThreads`/`LoadTaskGraph` *is* a filesystem failure by construction.

The inconsistency is local, which is what makes it a defect rather than a
simplification. The same function's own preflight gets it right 12 lines earlier:
`graphrepair.go:91-98` returns `fmt.Errorf("re-read Thread evidence before repair:
%w", err)` and only then compares snapshots at `:100`. All four sibling families
do the same (`threadapply.go:115-123`, `threadcreation.go:88-95`,
`threadmutation.go:89-96`, `lifecyclemutation.go:105-112`), and `verifyUnchanged`
(`cas.go:117-135`) states the policy outright: "Preserve actionable ambiguity/I/O
errors instead of inviting futile retries."

Three consequences follow:

1. **Futile retries against a possibly-failing disk.** `RepairTaskGraph` retries
   on `ErrConflict` while nothing is durable
   (`core/dependency_repair.go:649`). A permission error on the first repair source
   therefore costs 1 + 4 attempts, each re-scanning every task *and* every Thread,
   before failing. This is precisely the cost shipped task `6g1dpb5gwtt9`
   eliminated in `cas.go`.
2. **The structured filesystem block vanishes.** `filesystemDetails`
   (`internal/cli/fserror.go:21-39`) needs an `*fs.PathError` or `*os.LinkError`
   in the chain. Because the cause is discarded rather than wrapped, `errors.As`
   finds nothing and `error.filesystem` is omitted from the `--json` envelope —
   so the agent loses `operation`, `path`, `class` and `retryable` together, and
   with them the four distinct next actions `fserror.go:17-20` exists to name.
3. **The exit code asserts something false.** Exit 14 is published as "a write
   collided with existing or concurrently changed state"
   (`internal/cli/exit.go:42`). For `EACCES` or `ENOSPC` that is simply untrue,
   and an agent routing on the code has no way to discover otherwise.

The blast radius is bounded — `task depend repair` only — but this is the one
capability permitted to begin from `GraphBroken` (ADR-0006, 2026-09-05), i.e. it
runs exactly when the tree is already damaged and an underlying I/O or permission
fault is most plausible. The recovery path is the worst place to misreport a
cause.

**Compounding cost:** the invariant "a load failure is not a conflict" is honoured
in six places and enforced in none — no shared helper, no lint rule, no test. Every
new guarded mutation family is one more hand-written `if err != nil || verify() !=
nil` away from repeating it, and `6gm9ny98zyf2` (completed 2026-10-16) already
rewrote this loop's retry semantics without the classification surfacing. Fixed
now it is a five-line change in one function; left alone it is the default shape
the next family copies, and each copy also silently enlarges the set of permanent
errors that burn the shared 4-attempt retry budget.

**Recommendation:** In the per-write loop, split each load/read error from its snapshot-equality verdict exactly as graphrepair.go:91-98 and the four sibling mutation families already do: return the underlying error with %w, and reserve domain.ErrConflict for an actual evidence mismatch.

#### M1. The transient-filesystem retry hint is commit-aware per call site, so five of eight recovery envelopes can contradict their own receipt · **Status:** tracked by 6g7wsr8yvt1a

**File:** internal/cli/exit.go:150 | **Component:** cli/error-envelope
**Effort:** S · **Urgency:** eventually

`WriteError` (`internal/cli/exit.go:137-222`) attaches a generic
`error.filesystem` block whose `retryable` field is documented as "whether
repeating the call unchanged could succeed"
(`internal/wire/envelopes.go:1154-1157`). It is true only for `EAGAIN`, `EINTR`
and `EBUSY` (`internal/cli/fserror.go:59-60`).

Three branches then override it to `false` once their own receipt reports durable
work:

- `exit.go:158-163` — committed ordinary creation
- `exit.go:176-181` — `planningErr.result.Committed`
- `exit.go:190-197` — `renameErr.receipt.DestinationWritten`

Five do not, although every one of their receipts already carries the signal:

- `dependencyErr` (`exit.go:165-169`) — `AppliedTaskIDs` / `RemainingTaskIDs`
  (`core/dependency_operations.go:48-50`)
- `repairErr` (`exit.go:170-174`) — `Committed`, `AppliedFiles`
  (`core/dependency_repair.go:103,113`)
- `lifecycleErr` (`exit.go:183-188`) — `Committed`
  (`core/task_lifecycle.go:108`)
- `threadApplyErr` (`exit.go:214-218`) — `Committed`, `Complete`
  (`core/thread_apply.go:154-155`)
- the two thread mutation/creation branches (`exit.go:199-212`)

So an envelope can carry `error.filesystem.retryable: true` beside
`error.dependency_mutation.applied_task_ids: [...]` and a core-owned remedy that
says the opposite — `dependency_operations.go:217` returns "inspect the current
graph before deciding whether to retry; the planned dependency writes already
landed", and the field is documented as "core-owned inspection/recovery guidance,
not permission to retry" (`dependency_operations.go:52-53`). Two fields of one
object disagree, and the one an agent is most likely to branch on is the wrong
one.

Reachability is narrow and worth stating honestly: it needs a post-commit failure
whose cause is an `*fs.PathError`/`*os.LinkError` carrying `EAGAIN`, `EINTR` or
`EBUSY` — e.g. `EBUSY` from `os.Rename` on a later target after an earlier one
committed (`graphmutation.go:100`). No data-loss path; the damage is a misled
caller. Guard-release failures do not reach it at all, since `syscall.Flock`
returns a bare errno with no `PathError` to match.

This is the verification that open design task `6g7wsr8yvt1a` Q5 explicitly asks
for — "a transient filesystem cause must not advertise whole-command retry when a
mutation receipt reports committed or partially durable work. Ordinary creates now
suppress this hint; verify the remaining mutation recovery envelopes in this design
pass" — and its 2026-09-27 progress entry records the same intent. Overlap is FULL;
this finding supplies the per-branch inventory rather than opening a second owner.

**Compounding cost:** correctness here is a convention repeated inside a
20-branch `errors.As` chain, not a property of the receipt type. Adding a ninth
guarded mutation family means remembering an easily-missed three-line `if`, and
nothing fails if it is forgotten — no test asserts the invariant, and the branch
count has grown monotonically (eight receipts today, three with the override).
Deriving it from one durable-outcome predicate makes the next family correct by
default; every release it stays a convention is one more branch to audit by hand.

**Recommendation:** Derive the suppression from the receipt instead of repeating it per branch: give every guarded receipt one durable-outcome predicate and clear Filesystem.Retryable once in WriteError when it reports true — the shared contract task 6g7wsr8yvt1a Q5 is already chartered to name.

## What audited clean

- **Every `writeFileAtomic` in `internal/store` has a paired CAS verify.** All 15
  production call sites cross-check against `verifyUnchanged`,
  `verifyFixPath` or an exclusive-create precondition; no lost-update hole
  remains in the ordinary mutation surface — ADR-0006 §2.
- **`verifyUnchanged` still refuses to misname a failure.** `cas.go:117-135`
  converts only `ErrNotFound`/`os.IsNotExist` into `ErrConflict` and preserves
  ambiguity and I/O errors, the contract shipped task `6g1dpb5gwtt9` established.
  H1 is the sibling path this policy never reached, not a regression here.
- **Non-Unix fails closed rather than pretending.** `lock_other.go:14-16` refuses
  mutation because silent no-op locking "would make a successful CAS claim
  untrue" — ADR-0006 2026-08-27 §3.
- **The in-process guard closes `flock`'s descriptor-independence hole before any
  adapter can trigger it.** `repositoryGuards` (`lock.go:22-31`) plus
  `normalizeRepositoryLockKey` (`:37-49`) — exactly the self-denial hazard
  `flock(2)` documents.
- **Planner re-entry is an attributable conflict, not a deadlock, and survives a
  panic.** `lock.go:69-98` with `defer leave()` at `graphmutation.go:118` —
  ADR-0006 2026-08-27 §2.
- **The whole-snapshot CAS covers evidence a semantic graph discards.**
  `cas.go:33-97` compares unreadable revisions and duplicate occurrences and
  fails closed when `SourceVersion` is empty, so a raw edit cannot hide behind an
  unchanged parse error — ADR-0006 2026-09-05.
- **Materialization is refused unless it reloads to the planned semantics.**
  `graphmutation.go:170-179` reparses the candidate bytes and rejects a write
  whose canonical set or legacy clearing did not actually materialize;
  `lifecyclemutation.go:225-228` and `body.go:44-58` do the same, the latter also
  refusing a body that ends inside an unterminated fence.
- **Materialization deliberately follows authorization.** `readTaskLifecycleBody`
  (`lifecyclemutation.go:142-145`) reads only the policy input needed to validate,
  "so an untrusted planner cannot trigger YAML surgery before its semantic plan is
  authorized."
- **The durability gate is present in every hand-rolled retry loop.**
  `dependency_operations.go:159-163` and `dependency_repair.go:649` both stop
  retrying the instant a prefix is durable — the behaviour is correct everywhere
  even though its spelling is not uniform.
- **Retry backoff is bounded and fully jittered for a stated reason.**
  `core/retry.go:11-38` — full jitter to de-correlate cron writers that woke on
  the same schedule, capped at 4 attempts so real contention surfaces loudly.
- **Dry runs never claim durability and are never retried.** `retry.go:64-66`,
  plus graph dry-run holding the same guard for an authoritative preview while
  explicitly promising nothing about a later apply — ADR-0006 2026-08-27 §6.
- **The TUI decided the contention question the ADR deferred.**
  `tui/read_retry.go:42-68` holds the first conflict per read generation, retries
  once, and distinguishes read contention from a mutation CAS conflict in a
  comment — ADR-0006 2026-08-27 §2.
- **Mutation policy is an explicit composition choice with no implicit-allow
  mode.** `core/mutation_policy.go:9-16,58-73` — zero value invalid, decisions
  never cached, authorization re-checked after a blocking guard acquisition
  (`fix.go:52-55`).
- **`lint --fix` locks for the batch, refuses what it cannot prove, and retains
  its prefix.** `fix.go:36-55` takes the write guard as "a batch SECOND writer";
  `:71-79` reports a graph-owned repair as `Skipped` rather than manufacturing an
  unchecked edge; `:148-167` returns completed rows alongside a mid-run error.
- **Crash orphans are swept conservatively.** `atomic.go:128-158` is scoped by the
  tool's own filename prefix *and* a one-hour age, so it can touch neither a user
  file nor a live write.
- **Batch partial failure leaves a coherent receipt and the most actionable exit
  code.** `cli/moves.go:32-88` attempts every slug and prefers a sentinel-bearing
  error over whichever failure came first in argv.
- **The exit taxonomy is closed, published, and defensive.** `exit.go:33-47`
  publishes code 12 as reserved-and-never-emitted, and `activeExitCode`
  (`:85-91`) refuses to let an unpublished number become a contract by accident.

## Candidate tasks

<!-- candidate-tasks:v1 · ○ open · ● in-progress · ✔ fixed · → tracked · ◌ deferred · ◌ superseded · ✘ wontfix -->
<!-- Add or replace one row with `tskflwctl audit finding <audit> <code> --candidate "<one line>"`; an empty value removes it. -->

- ○ H1 · open — tskflwctl task new "Preserve filesystem causes through graph-repair reauthorization" --epic 21-code-quality-architecture-hardening --tier 2 --priority medium
- → M1 · tracked — Tracked in planning/tasks/6g7wsr8yvt1a-define-the-next-shared-entity-integrity-foundations.md

## Related-task observations (propose-only)

- **Same-class defect, shipped fix, unreached sibling:**
  `planning/tasks/6g1dpb5gwtt9-verifyunchanged-collapses-every-resolve-error-into-a-futile-retry-conflict.md`
  is `completed` (2026-09-02) and fixed this exact defect class in `cas.go`; H1 is
  the same class surviving in `graphrepair.go`. Overlap is PARTIAL, not FULL — the
  completed task's scope was `verifyUnchanged`. Deliberately not annotated: adding
  an `audit_sources` entry to a completed task would imply it reopened. Its
  framing ("The same function is careful to distinguish `os.IsNotExist` from a
  genuine read error seven lines later, so the inconsistency is the defect rather
  than a deliberate simplification") applies verbatim to `graphrepair.go`, whose
  own preflight at `:91-98` gets it right.
- **Adjacent recent work in the same loop:**
  `planning/tasks/6gm9ny98zyf2-pin-explicit-graph-repair-source-selections-across-retries.md`
  is `completed` 2026-10-16 — three days before this run — and changed the retry
  semantics of the very loop H1 lives in without the classification surfacing.
  That raises H1's urgency: this is actively-edited code, and the next family to
  copy the pattern will copy the defect.
- **Still-accurate tracked divergences, re-verified not re-raised:** finding M3 of
  `2026-09-21-arch-failure-and-recovery` (per-filesystem `flock` guarantee, tracked
  by `6g6yjm16dkgt`) and task `6g63jj1dh0sb` (three-way `writeFileAtomic`
  durability split) both still match current code exactly. No new findings created
  for either.
- **Possibly closeable audit:** `planning/audits/6ggqwdz14nzx-2026-10-05-arch-data-model-and-storage.md`
  reports `ready_to_close: true` (1 finding, 0 open) in `audit list --json`.
  Closing it is outside this routine's allowed writes; flagging for a maintainer.
