---
schema: 1
id: 6gc7jd9aq1q9
bucket: open
area: arch-failure-and-recovery
date: "2026-09-21"
updated_at: "2026-09-21"
---

# Audit: arch-failure-and-recovery — 2026-09-21

> Create findings with `audit finding new`; update them with `audit finding`.

Routine: `weekly-architecture-audit` · lens `failure-and-recovery` · ISO week `2026-W39`
(mod 4 = 3). Run on schedule (Monday); `just` was unavailable in the runner, so the
build used `go build -o bin/tskflwctl ./cmd/tskflwctl` and the suite `go test -race ./...`
— the same commands the Justfile recipes wrap.

ADRs consulted: ADR-0003 (+ its 2026-09-08 rename amendment) · ADR-0004 §1–§2 ·
ADR-0006 (2026-08-27 portable mutation-guard hardening, 2026-08-27 dependency-operation
recovery contracts, 2026-08-28 lifecycle consequences, 2026-09-05 broken-graph recovery).
Files read end-to-end: `internal/store/atomic.go`, `lock.go`, `lock_unix.go`,
`lock_other.go`, `cas.go`, `rename.go`, `graphrepair.go`; `internal/core/retry.go`,
`task_lifecycle.go`, `task_rename.go`, `dependency_operations.go`, `dependency_repair.go`,
`store.go`; `internal/cli/exit.go`, `moves.go`; plus the edge reads
`internal/tui/entity.go` and `internal/domain/problem.go` / `lint.go` / `classify.go`.
Sources cited: 6.

## Executive summary

This lens is in good shape — materially better than most filesystem-backed tools ever get.
The write primitive implements the full crash-durable replacement sequence including the
parent-directory fsync that practitioners routinely omit; the repository guard correctly
pairs an in-process keyed mutex with `flock` to close the "one process, two descriptors"
hole that `flock(2)` warns about; and every guarded mutation family returns a typed receipt
that distinguishes a pre-commit failure from a durable partial prefix, which is exactly the
"partial completion is a first-class outcome, not Failed with a prose note" discipline the
distributed-systems literature asks for. No Critical or High finding surfaced.

The headline is a **shape** problem, not a defect: the predicate that decides *whether a
retry is safe* — "did anything become durable?" — is spelled four different ways across
seven hand-rolled retry loops, and one guarded mutation family does not carry the flag at
all (M1). Each of the seven is individually correct today; nothing in the type system or a
shared test holds the eighth to the same rule. Second, the diagnostic vocabulary that
`lint` emits carries no machine-readable kind, so the tool cannot prove which reportable
defects have a repair path and an agent must route on English prose (M2) — the one
closed-vocabulary-shaped surface in a codebase otherwise obsessive about closed
vocabularies. Third, the guard's supported matrix is stated per-platform while the
primitive's guarantee is per-filesystem (M3).

Read M1 first; it is the finding whose cost compounds with the roadmap.

## State of the architecture (this lens)

Failure handling is layered, and the layers are genuinely separable.

**The write primitive** (`internal/store/atomic.go`). `writeFileAtomic` stages into a temp
file in the *same directory* as the destination, `Sync()`s it, `chmod`s it, `os.Rename`s it
into place, then best-effort fsyncs the parent directory (`atomic.go:48-62`, `syncDir` at
`atomic.go:68-73`). Overwrites preserve the destination's existing mode rather than
resetting to `perm` (`atomic.go:49-51`). `createFileAtomic` is the exclusive-create sibling
using `O_CREATE|O_EXCL` and is explicitly documented as the `ifVersion == ""` CAS
precondition — create-must-not-exist (`atomic.go:81-86`). A crash-orphan sweeper
(`sweepStaleTemps`, `atomic.go:142-158`) reclaims the tool's own `.tskflwctl-*.tmp` files
older than an hour, scoped by prefix *and* age so it can never touch a user file.

**The CAS token** (`internal/store/cas.go`). Concurrency control is version-CAS: a SHA-256
over a file's exact bytes, computed on read and never stored (`hashContent`, `cas.go:26-29`).
`verifyUnchanged` (`cas.go:120-175`) runs immediately before every write and is unusually
careful about *error taxonomy*: it re-resolves by the canonical id rather than the caller's
fuzzy query (so a concurrently-created same-prefix file cannot fake a conflict), treats
`ErrNotFound` and `os.IsNotExist` as genuine conflicts, and deliberately refuses to collapse
`ErrAmbiguous` or a real I/O error into a conflict — because retrying those "would burn the
whole OCC budget and then hand back 'changed on disk; retry': advice that cannot work, for
a condition it misnames" (`cas.go:147-153`). Graph-aware mutations add a whole-snapshot CAS
(`verifyTaskGraphSourceSnapshot`, `cas.go:36-41`) that deliberately includes *unreadable*
source revisions, so a raw edit cannot hide behind an unchanged parse error.

**The repository guard** (`internal/store/lock.go`, `lock_unix.go`, `lock_other.go`). Two
halves. The cross-process half is `flock(LOCK_EX)` on the repo root directory
(`lock_unix.go:21-39`). The same-process half is a package-level map of per-canonical-root
mutexes (`repositoryGuards`, `lock.go:28-31`), keyed through `normalizeRepositoryLockKey`
(symlink-resolved, absolute, cleaned — `lock.go:37-49`). The comment states the reason
precisely: "Platform file-lock semantics vary for independently opened handles owned by one
process, and a future long-lived adapter may create more than one `*FS`" (`lock.go:23-25`).
On non-Unix, `platformWriteLockChecked` fails closed with `ErrValidation` rather than
silently no-op locking (`lock_other.go:14-16`).

Layered on top is the **planner-exclusion phase**: `enterRepositoryPlanner` (`lock.go:69-83`)
marks a canonical root as inside a control-inverted planner callback, and
`rejectRepositoryPlannerCall` (`lock.go:89-98`) converts any Store entry during that window
into an attributable `ErrConflict` instead of a self-deadlock. This is applied broadly and
consistently: **49** entry points call it across `create.go`, `edit.go`, `body.go`,
`fsstore.go`, `epicstore.go`, `auditstore.go`, `researchstore.go`, `threadstore.go`,
`paths.go`, `danglers.go`, `fix.go`, `rename.go` and the six guarded mutation files — reads
included, which is what makes the ADR's "every Store entry point for that root" claim true
rather than aspirational.

**Guarded mutation transactions.** Six families own a guarded critical section:
`TaskGraphMutationStore` (dependency edges), `TaskGraphRepairStore` (broken-graph repair),
`TaskLifecycleMutationStore` (status transitions), `ThreadCreationMutationStore`,
`ThreadMutationStore` (membership/lifecycle), `ThreadApplyMutationStore` (bulk plans), plus
`RenameTask` as the exceptional multi-document member of the ordinary task port. Each takes
`checkedWriteLock`, loads a strict snapshot, calls a pure core planner, CASes the whole
snapshot and then each target immediately before replacement, and returns a receipt
identifying the durable prefix. `graphrepair.go:142-148` is representative: write, then
`AppliedSources = append(...)`, `Changed = true`, `Committed = true`, then re-read Thread
and task evidence and fail with an explicit "committed a prefix; inspect before retrying"
conflict if either moved.

**Retry policy** (`internal/core/`). `retryOnConflict` (`retry.go:62-71`) is the generic
bounded auto-retry for single-file scriptable mutations: 4 attempts, capped exponential
backoff with full jitter (`retryBackoff`, `retry.go:23-31`), explicitly never applied to a
dry run. Guarded multi-file mutations do **not** use it; each re-implements the loop with
its own durability guard (M1).

**What the caller is told.** `cmd/tskflwctl/main.go:47-58` routes every failure through
`cli.WriteError` + `cli.ExitCode`. The exit taxonomy is a four-row table (`exit.go:22-30`:
10 not-found, 11 validation, 13 ambiguous, 14 conflict; 12 retired and reserved) driven by
`domain.Classify` (`classify.go:35-50`), which the TUI and a future web adapter share. Under
`--json`, `WriteError` (`exit.go:66-127`) attaches a *typed recovery envelope* per mutation
family — eight `errors.As` arms covering dependency mutation, graph repair, task lifecycle,
task rename, thread creation, thread update, thread policy, and thread apply — and every one
of those envelopes carries a `committed` boolean into the wire contract
(`wire/dependency.go:222`, `wire/task_rename.go:18`, `wire/thread.go:299,339,475`,
`wire/dependency_repair.go:67`). Batch transitions go through `runMoves`
(`cli/moves.go:32-88`): every slug is attempted, each gets a per-item result row, and the
summary error prefers a sentinel-bearing failure so the batch's exit code names the most
actionable cause rather than the first one in argv.

**The TUI** surfaces the same distinction: `moveTask` and `deferTaskCmd`
(`tui/entity.go:343-375`) unwrap `*core.TaskLifecycleMutationFailure` and return a
`movedMsg` carrying the committed receipt plus a warning — a reload, not a stale view —
where any other error becomes `actionErrMsg` with no reload.

**Repair.** `lint` reports; `lint --fix` repairs a deliberately narrow set
(`store/fix.go:17-24`): text-level frontmatter normalization and stable-id backfill from the
filename. It takes the repository write lock for the whole batch because it is "a batch
SECOND writer" (`fix.go:29-38`), and it refuses graph-owned repairs outright rather than
manufacturing unchecked edges, reporting the refusal as a `Skipped` result whose reason is
the useful payload (`fix.go:55-62`, `domain/problem.go:21-25`). Graph defects instead route
to `task depend repair` under `TaskGraphRepairStore`.

## ADR reconciliation

### ADR-0006 2026-08-27 §2 — planner exclusion is scoped to the canonical planning root

**Quoted:** "During the callback, every Store entry point for that root—including through a
second `FS`—fails fast with `ErrConflict`."
**Implementation:** follows — `rejectRepositoryPlannerCall` is invoked at 49 Store entry
points (`lock.go:89-98` and call sites across all 13 store files with public entry points),
and the guard is looked up by normalized canonical root (`lock.go:33-49`), so a second `*FS`
over the same tree shares it.
**Class if divergent:** n/a.

### ADR-0006 2026-08-27 §3 — only runtime-tested release platforms claim mutation support

**Quoted:** "Windows and other non-Unix source builds fail closed until a shared-repository
lock identity is selected and exercised in native CI; a per-user cache lock is insufficient
for cross-user repositories."
**Implementation:** follows for the platform axis — `lock_other.go:14-16` returns
`ErrValidation` rather than a silent no-op. But the clause is stated per-**platform**, and
`flock`'s guarantee is per-**filesystem**: a Linux or macOS build over a network-mounted or
cloud-synced planning root passes the gate while the cross-process exclusion silently
degrades.
**Class if divergent:** ADR gap (finding M3).

### ADR-0006 2026-08-27 §4 — raw-editor detection is best effort, not transactional isolation

**Quoted:** "A raw writer can still race the final verify-to-rename interval because it does
not honor the advisory guard; the operation never claims an all-files transaction or
isolation from direct filesystem edits."
**Implementation:** follows, and honestly — `lock_unix.go:11-20` and
`docs/ARCHITECTURE.md` both state the residual window rather than papering over it, and
`verifyUnchanged` is positioned immediately before each replacement to narrow it.
**Class if divergent:** n/a.

### ADR-0006 2026-08-27 §1 / 2026-08-28 §1 — planner write order is recovery data; eligibility is a guarded write

**Quoted:** "Stable-ID ordering is deterministic but not generally prefix-safe: an edge
reversal must durably remove the old edge before adding its reverse." / "Ordinary `Move`
after a separate eligibility read is invalid because a prerequisite lifecycle or dependency
mutation could commit between authorization and write."
**Implementation:** follows — `TaskLifecycleMutationStore` holds the guard across graph load,
authorization and persistence (`store/lifecyclemutation.go:28-32`, planner entered at
`:165`), and generic set/edit paths cannot change status.
**Class if divergent:** n/a.

### ADR-0006 2026-08-27 recovery contracts §4 — receipts distinguish convergence from success

**Quoted:** "A failure after a durable prefix carries that prefix in typed human and JSON
diagnostics rather than collapsing it into a prose-only error."
**Implementation:** follows at the adapter boundary — eight typed recovery envelopes in
`cli/exit.go:84-122`, each carrying `committed`. **Silent** on the core-side question of how
that durability is *represented* so a retry loop can read it; the ADR specifies the receipt's
obligations to the caller but never names the predicate the tool itself must consult.
**Class if divergent:** ADR gap (finding M1).

### ADR-0003 2026-09-08 — task rename is a guarded, resumable multi-document mutation

**Quoted:** "A real rename captures the caller's source version before waiting, takes the
canonical repository guard, rejects a changed source, and plans against the fresh guarded
tree." / "Same-path cascade rewrites commit first, destination creation second, and
old-source deletion last."
**Implementation:** follows precisely — `store/rename.go:31-80` snapshots the pre-lock source
version, takes `checkedWriteLock`, re-checks target availability *after* acquiring the guard,
and `verifyUnchanged`es the source before planning. Commit order and the
destination-written/source-retained recovery state are modelled explicitly
(`core/task_rename.go:12-24`, `taskRenameFailureRemedy` at `:79-92`). The amendment is
**silent** on whether a pre-commit conflict — one where nothing was written — should be
auto-retried like every sibling mutation; it is not (finding L1).
**Class if divergent:** ADR gap (finding L1).

### ADR-0006 2026-09-05 — broken-graph recovery operates on source declarations

**Implementation:** follows — `TaskGraphRepairStore` is a separate capability rather than a
mode flag on the ordinary mutation port (`core/store.go:105`), accepts a broken source graph
only for reauthorized operations, and composes each step's proof into the receipt
(`store/graphrepair.go:135-160`).
**Class if divergent:** n/a.

### ADR-0004 §1–§2 — git is the single canonical authority; the local default is "tool writes files, you commit"

**Implementation:** follows — `tskflwctl` never touches git, and the guard makes no claim
beyond one machine's cooperating writers. This clause is also what keeps M3 a Medium rather
than a High: multi-machine convergence is git's job by decision, not the lock's.

### No ADR governs the diagnostic vocabulary

`domain.FileProblem` (`problem.go:5-15`) and `domain.Issue` (`lint.go:39-45`) are the two
shapes in which the tool reports a defect, and neither is the subject of any accepted ADR.
ADR-0007 decides *planning state* vocabularies (statuses, finding statuses, criterion states)
and is scrupulous about them; the vocabulary in which the tool describes a **broken
repository** was never given the same treatment. That absence is itself the finding (M2).

## Best-practice comparison

### 1. Crash-durable atomic replacement requires a parent-directory fsync, attempted and tolerated

**Source:** https://github.com/ArcadeData/arcadedb/issues/7465
**Guidance:** "On a POSIX filesystem the rename is a directory-metadata update, and fsyncing
the file does not make that entry durable: after a power failure the file's data can be on
disk while the directory still points at the old name, or at nothing." The fix is to sync the
parent directory afterwards — but "Opening a directory as a `FileChannel` works on Linux and
macOS and throws on Windows, so it has to be attempted and its failure tolerated, not
assumed."
**This codebase:** follows, including the subtlety — `writeFileAtomic` runs
write→`Sync`→`Rename`→`syncDir` (`atomic.go:48-62`), and `syncDir` swallows the error with
the rationale that "some filesystems (network and FUSE mounts in particular) reject directory
fsync, and degraded durability there beats failing every write" (`atomic.go:64-73`). That is
the cited recommendation's *exact* posture, arrived at independently.
**Justified?** n/a — follows.

### 2. Stage the temp file on the same filesystem; fsync before the rename

**Source:** https://github.com/google/renameio
**Guidance:** "On POSIX operating systems, the `fsync` system call must be used to ensure
that the `os.Rename()` call will not result in a 0-length file." And: "The temporary file
must be created on the same file system (same mount point) for the rename to work."
**This codebase:** follows — `stageTemp` calls `os.CreateTemp(dir, ...)` with `dir` being the
destination's own directory (`atomic.go:13-14`, called from `atomic.go:52` as
`filepath.Dir(path)`), and `tmp.Sync()` precedes the rename (`atomic.go:27-30`).
**Justified?** n/a — follows.

### 3. Windows cannot be made atomic reliably; fail closed rather than pretend

**Source:** https://github.com/google/renameio (citing https://github.com/golang/go/issues/22397)
**Guidance:** "It is not possible to reliably write files atomically on Windows."
**This codebase:** follows — `lock_other.go:11-16` rejects repository mutation on non-Unix
with an explicit comment that "Silent no-op locking would make a successful CAS claim untrue
and is therefore less safe than rejecting mutation." Matches ADR-0006 2026-08-27 §3.
**Justified?** n/a — follows.

### 4. `flock` treats independently opened descriptors in one process as independent locks

**Source:** https://man7.org/linux/man-pages/man2/flock.2.html
**Guidance:** "If a process uses open() to obtain more than one file descriptor for the same
file, these file descriptors are treated independently by flock(). An attempt to lock the
file using one of these file descriptors may be denied by a lock that the calling process has
already placed via another file descriptor." The same page notes flock is advisory, and that
over NFS since Linux 2.6.12 flock locks are *emulated* as fcntl byte-range locks.
**This codebase:** follows for the descriptor pitfall — the keyed in-process mutex
(`lock.go:28-31`, `processRepositoryLock` at `:63-67`) exists precisely so that two `*FS`
values over one root serialize within a process rather than relying on `flock` semantics that
vary by platform. Diverges on the *filesystem* axis: nothing detects or declares that the
planning root is on a filesystem where the cross-process guarantee degrades.
**Justified?** Partly. ADR-0004 makes git the sanctioned multi-machine channel, so a synced
working tree is outside the decided model — but the model is not stated where a user or a
future served adapter would meet it. See M3.

### 5. Partial completion is a first-class outcome, not "failed" with a prose note

**Source:** https://github.com/Limes-Labs/limes-axis/issues/601 (and the surrounding
idempotency/durable-execution practice: https://github.com/mycelium-labs/mycelium/issues/113)
**Guidance:** "Partial completion is a first-class outcome, not Failed with a prose note."
"Automatic retry after unknown outcome requires stable operation identity and either provider
idempotency or authoritative proof of no-effect. Non-idempotent operations with no
reconciliation must not auto-retry after uncertain delivery."
**This codebase:** follows on representation, partially on *consumption*. Representation is
excellent: every guarded family returns planned-vs-applied counts, a durable prefix, and a
`Remedy` string, and all eight reach `--json` as typed recovery envelopes
(`cli/exit.go:84-122`). Consumption is where it frays — "authoritative proof of no-effect" is
the exact predicate each retry loop must evaluate, and it is evaluated seven times in four
different vocabularies (M1).
**Justified?** No. Nothing about single-user local-first scope requires the predicate to be
anonymous.

### 6. A diagnostic should carry a stable machine identifier and an explicit applicability for its fix

**Source:** https://github.com/rust-lang/rustc-dev-guide/blob/main/src/diagnostics.md
**Guidance:** "Most errors have an associated error code. Error codes are linked to long-form
explanations." "The compiler accepts an `--error-format json` flag to output diagnostics as
JSON objects (for the benefit of tools such as `cargo fix`)." And crucially, suggestions
carry "`rustc_errors::Applicability` confidence level to guide automated source fixes by
tools" — "The last argument provides a hint to tools whether the suggestion is mechanically
applicable or not."
**This codebase:** diverges. `tskflwctl lint --json` is the structural analogue of
`--error-format json`, and `lint --fix` the analogue of `cargo fix` — but the payload is
`{"path": ..., "message": ...}` (`domain/problem.go:5-15`) and `{"field": ..., "message":
..., "severity": ...}` (`domain/lint.go:39-45`). There is no code, no defect kind, and no
applicability. `Message` is populated straight from `err.Error()` at four sites
(`store/resolve.go:63`, `core/finding.go:187,366`, `core/service.go:496`).
**Justified?** No. This is the one surface where the project's own stated principle — "A
state the tool cannot write is a state nobody can be held to" (`docs/ARCHITECTURE.md`) —
was not applied to the vocabulary in which the tool reports trouble. See M2.

## Tensions and trade-offs

**Repo-wide locking vs. per-file locking.** `lock_unix.go:11-20` takes the whole repo root
for every write and says why: "writes are brief and infrequent, so serializing them is
imperceptible; per-file locking is a future refinement." For a single-user planning repo with
a handful of cron agents this is unambiguously the right call — per-file locking would buy
throughput nobody needs and cost a deadlock-ordering problem. Worth revisiting only if epic 19's
served adapter ever holds the guard across a request.

**Best-effort `syncDir` vs. strict durability.** Ignoring the directory-fsync error is a real
weakening of the durability claim, and it is the right weakening: failing every write on a
FUSE or network mount to protect against power loss would trade a common failure for a rare
one. Cited practice agrees (comparison §1). The cost is that the tool cannot *tell* a user
their planning root is on a filesystem where durability is degraded — the same blind spot M3
names for locking.

**Bounded retry vs. surfacing contention.** `defaultMaxRetries = 4` with a 50ms cap
(`retry.go:15,25`) is deliberately small: "a genuinely contended file should surface a loud
`ErrConflict` (exit 14) quickly, not spin." Correct for this workload. The tension is that
exit 14 alone cannot distinguish a retry-safe pre-commit conflict from a post-commit cleanup
failure that must *not* be retried; the project resolves this by putting `committed` in the
`--json` envelope rather than by splitting the exit code, which keeps the four-row taxonomy
stable (a published contract, `exit.go:19`) at the cost of making the human path's prose
carry the distinction. That is a defensible trade, and the prose does carry it
(`core/task_lifecycle.go:172-178`).

**Fail-closed lifecycle vs. fail-open lint.** `ValidateTaskLifecycleSource`
(`core/task_lifecycle.go:188-198`) refuses to authorize a transition on a degraded graph,
while `lint` lists and flags a bad status rather than refusing. The asymmetry is principled —
mutation needs a sound graph, diagnosis must work on a broken one — and is exactly what lets
`TaskGraphRepairStore` exist as the only capability permitted to accept `GraphBroken`.

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

#### M1. The retry-safety predicate — "did anything become durable?" — is spelled four ways across seven hand-rolled loops · **Status:** open

**File:** internal/core/retry.go:62 | **Component:** core/guarded-mutations
**Effort:** M · **Urgency:** soon

**Class:** ADR gap
**Anchored to:** ADR-0006 2026-08-27 recovery contracts §4; https://github.com/Limes-Labs/limes-axis/issues/601

Every guarded mutation must answer one question before retrying a `domain.ErrConflict`:
*did anything already become durable?* Retrying when the answer is yes replays a partial
multi-file plan and erases the recovery evidence the operator needs. The codebase answers
it correctly in all seven places — and spells it differently in four:

| Call site | Durability predicate |
| --- | --- |
| `core/retry.go:67` (`retryOnConflict`, generic) | *none* — retries unconditionally |
| `core/service_task.go:376` (lifecycle) | `!result.Committed` |
| `core/service_thread.go:93` (thread creation) | `!result.Committed` |
| `core/service_thread.go:179` (thread mutation) | `!result.Committed` |
| `core/service_thread_apply.go:69` (bulk apply) | `result.Committed \|\| result.Complete` |
| `core/dependency_operations.go:148` (graph edges) | `len(result.AppliedTaskIDs) > 0` |
| `core/dependency_repair.go:619` (graph repair) | `len(result.AppliedSources) > 0` |

The result types diverge with them. `TaskLifecycleMutationResult`,
`ThreadCreationMutationResult` and `ThreadMutationResult` carry `Committed`.
`ThreadApplyMutationResult` carries `Committed` *and* `Complete`.
`TaskRenameMutationResult` carries four (`Committed`, `Complete`, `DestinationWritten`,
`SourceRemoved`). `TaskGraphRepairMutationResult` carries `Committed` **and**
`AppliedSources` — and its retry loop reads the slice while the receipt it returns reports
the flag (`core/dependency_repair.go:92,619`; they are kept in sync only by adjacency at
`store/graphrepair.go:146-148`). `TaskGraphMutationResult` (`core/store.go:77-81`) carries
no durability flag at all; `len(AppliedTaskIDs)` is the only evidence available.

Nothing enforces the invariant. There is no shared interface, no named predicate, and no
test that asserts across families that a partially-applied result is never retried — each
loop is verified only by its own package's tests. `retryOnConflict`, the one *shared*
combinator, has no durability guard whatsoever; it is safe today purely because every
caller happens to be a single-file read-modify-write where the CAS fails before any write
(`core/retry.go:56-60` states that reasoning as a comment, not as a constraint the type
system checks).

This is not a live bug. All seven are correct as written. The finding is that correctness
here is a property of seven independent authors agreeing, not of the design.

**Compounding cost:** The eighth guarded family pays the full cost. Epic 19's served
adapter and epic 30's remaining Thread work both add guarded mutations, and ADR-0006
2026-08-27 §2 explicitly anticipates "a future long-lived adapter may retry after the
mutation" — i.e. more retry call sites, written by someone reading one of four precedents.
The TUI shows the shape of the tax already: `tui/entity.go:349,367` unwraps
`*core.TaskLifecycleMutationFailure` and nothing else, so the day the TUI gains Thread
mutations it must independently learn two more failure types, with no interface making
that obligation visible. Fixing this after N families costs N migrations plus the
archaeology of deciding which spelling each one meant; fixing it at 7 is one interface and
one combinator.

**Proposed ADR amendment:** ADR-0006 — add to the 2026-08-27 recovery contracts: "Every
guarded mutation result exposes durability through one shared predicate. A retry may
proceed only while that predicate is false; a result carrying durable work is returned to
the caller with its prefix, never replayed. Per-operation fields (`Complete`,
`DestinationWritten`) refine the receipt for the caller and never substitute for the
predicate."

**Recommendation:** Name the predicate once: a small interface (e.g. `DurableOutcome interface { Durable() bool }`) implemented by every guarded mutation result, plus one shared `retryUntilDurable` combinator the seven call sites delegate to. Add `Committed` to `TaskGraphMutationResult` so no family infers durability from a slice length, and one table test asserting every result type reports Durable()==true after a partial write.

#### M2. lint reports defects in prose, so repair coverage is unprovable and agents must route on English · **Status:** open

**File:** internal/domain/problem.go:5 | **Component:** domain/diagnostics
**Effort:** M · **Urgency:** soon

**Class:** ADR gap
**Anchored to:** https://github.com/rust-lang/rustc-dev-guide/blob/main/src/diagnostics.md; `docs/ARCHITECTURE.md` "A state the tool cannot write is a state nobody can be held to"

The tool reports repository defects through two types, and neither carries a machine
identifier:

    // internal/domain/problem.go:5-15
    type FileProblem struct { Path string; Message string; ... }

    // internal/domain/lint.go:39-45
    type Issue struct { Field string; Message string; Severity IssueSeverity }

`Message` is the payload, and at four of its construction sites it is `err.Error()` verbatim
— `store/resolve.go:63`, `core/finding.go:187`, `core/finding.go:366`, `core/service.go:496`.
Both fields are public wire contract (`json:"message"`, surfaced through the `unreadable`
and `issues` arrays of the `lint --json` envelope, `wire/envelopes.go:917,946`). The only
machine axis anywhere in the vocabulary is `Severity`, and it has exactly one value
(`IssueAdvisory`, `lint.go:34`) plus the empty default.

rustc is the direct precedent for what this surface is trying to be. Its diagnostics carry
a stable code ("Most errors have an associated error code"), it emits them machine-readably
("The compiler accepts an `--error-format json` flag to output diagnostics as JSON objects
(for the benefit of tools such as `cargo fix`)"), and — the part that matters most here —
each suggestion carries an explicit `Applicability` that "provides a hint to tools whether
the suggestion is mechanically applicable or not." `tskflwctl lint --json` is the structural
analogue of `--error-format json` and `lint --fix` the analogue of `cargo fix`, but the link
between a reported defect and its repair is missing at both ends.

Two concrete consequences, both visible in the repo today:

1. **Repair coverage cannot be stated, let alone tested.** `FixFrontmatter`
   (`store/fix.go:17-24`) repairs text-level frontmatter normalization and stable-id
   backfill. Everything else `lint` can report — a missing or unrecognized status, a
   non-id-led file under a scanned directory, a duplicate epic `NN` key
   (`domain/lint.go:74-79`), a graph-owned field it deliberately refuses
   (`store/fix.go:55-62`) — has a repair path somewhere, nowhere, or in a different verb
   (`task depend repair`), and there is no enumerable set to check that against. The
   question LENS-SCOPES asks of this lens — "whether a supported repair path exists for
   each defect `lint` can report" — is currently unanswerable from the code.

2. **An agent routes on prose.** The project's whole premise is an agent-first machine
   contract (ADR-0008), and `lint --fix` is the documented hygiene loop in `CLAUDE.md`. An
   agent that wants to know whether a reported problem is self-healing has only the English
   string to go on — including the refusal string "graph-owned repair refused (%s); repair
   deliberately, run lint, then use guarded dependency operations" (`store/fix.go:58`).

The project applies exactly the missing discipline everywhere else. ADR-0007 closes the
planning-state vocabularies; `domain/resolution.go` spells a shared word once and has drift
tests that fail if the two sides diverge; `docs/ARCHITECTURE.md` states the principle as "A
state the tool cannot write is a state nobody can be held to," with the rationale that the
alternative "is how a vocabulary drifts from its own documentation." The vocabulary in which
the tool describes a *broken repository* is the one closed-vocabulary-shaped surface that
never received that treatment, and no accepted ADR governs it.

**Compounding cost:** Every prose message is a de-facto contract the moment an agent or a
golden pins it. The CLI goldens under `internal/cli/testdata/golden/` already freeze
diagnostic text, so today's strings are load-bearing without being declared. Each new lint
check adds another unclassified string; retrofitting codes later means assigning them to a
corpus of messages whose exact wording is already depended on, rather than declaring them
at the point each check is written. Epic 26 is about to derive lint, `schema task`
guidance, and the `--json` contract from one declared field registry — that work will
generate diagnostics from the registry, which is precisely the moment a kind is cheap to
mint and expensive to add afterwards.

**Proposed ADR amendment:** ADR-0007 — extend the closed-vocabulary rule beyond planning
state: "A defect the tool can report is a closed vocabulary like any other. Each reportable
defect declares a stable kind and an explicit repair applicability; `lint --fix`'s
capability set is derived from that table rather than maintained beside it, and a kind with
neither a fix path nor a declared manual disposition is a lint failure in the tool's own
test suite."

**Follow-up:** The neutral-identity work in
`planning/tasks/6g5vm4efjcdv-make-repository-lint-load-diagnostics-adapter-neutral.md`
touches the same struct and envelope and is the natural vehicle, but its scope is entity
kind, stable identity and optional repair *location* — not defect kind or applicability,
and it explicitly lists "Changing lint rules" as out of scope. Sequencing the two together
would avoid two consecutive schema bumps on one envelope.

**Recommendation:** Give both diagnostic types a closed machine vocabulary: a `Kind` (or `Code`) field drawn from a domain-owned table, and an `Applicability` enum (`fix` | `manual` | `refused`) stating whether `lint --fix` can repair it. Emit both in the --json envelope behind a schema_version bump, and add the drift test the project already uses for status vocabularies: every declared kind either has a fix path or is declared manual.

#### M3. The repository guard's supported matrix is stated per-platform, but flock's guarantee is per-filesystem · **Status:** open

**File:** internal/store/lock_unix.go:21 | **Component:** store/repository-guard
**Effort:** S · **Urgency:** eventually

**Class:** ADR gap
**Anchored to:** ADR-0006 2026-08-27 §3; https://man7.org/linux/man-pages/man2/flock.2.html

ADR-0006's 2026-08-27 hardening amendment scopes mutation support by **platform**:

> "macOS and Linux use the canonical-root process mutex plus root-directory `flock`, with
> real same-process and child-process tests over the production path. Windows and other
> non-Unix source builds fail closed until a shared-repository lock identity is selected and
> exercised in native CI; a per-user cache lock is insufficient for cross-user repositories."

The code implements that faithfully (`lock_other.go:14-16` fails closed with
`ErrValidation`). But `flock`'s guarantee is not a property of the platform — it is a
property of the **filesystem the locked inode lives on**. `flock(2)` notes that over NFS,
`flock()` locks are emulated as `fcntl()` byte-range locks (Linux ≥ 2.6.12), interacting
with `fcntl` locks in ways the local path never does, and on older kernels did not lock over
NFS at all. FUSE-backed cloud-sync mounts and SMB shares vary similarly and are not covered
by any test in this repo.

So a planning root on a network share or a cloud-synced directory passes every gate the
project checks — a supported platform, a Unix build, `flock` returning success — while the
cross-process exclusion it claims silently degrades or disappears. `store/lock_unix.go:11-20`
documents the *advisory* limitation (a raw editor is not blocked) but not the *filesystem*
one, and `syncDir` (`store/atomic.go:64-73`) already acknowledges that this deployment shape
is real: it swallows directory-fsync errors precisely because "some filesystems (network and
FUSE mounts in particular) reject directory fsync."

Two properties keep this Medium rather than High. ADR-0004 §1–§2 decides that git is the
canonical multi-machine channel and the local default is "tool writes files, you commit," so
a synced working tree is outside the decided model. And the version-CAS still fires
independently of the lock: two writers on a degraded filesystem are far more likely to get a
loud `ErrConflict` than a silent lost update, because `verifyUnchanged` re-hashes the file
immediately before each replacement (`store/cas.go:158-173`). The residual exposure is the
verify→rename window the guard exists to close, which on such a mount is simply not closed.

**Compounding cost:** The clause's own last words are "a per-user cache lock is insufficient
for **cross-user repositories**" — and a cross-user repository lives on a shared filesystem
by definition. Epic 19's `serve` adapter and epic 29's multi-space registry both move toward
exactly that. When the shared-repository lock identity that ADR-0006 §3 defers is finally
chosen, the filesystem axis will have to be answered anyway; answering it now costs a
documented precondition and a `doctor` check, whereas answering it then means retrofitting a
precondition onto a released guard whose users have already put repos wherever they liked.

**Proposed ADR amendment:** ADR-0006 — extend the 2026-08-27 §3 supported matrix: "Mutation
support is claimed for a runtime-tested platform **on a local filesystem**. The repository
guard's cross-process exclusion is not claimed for network or synchronized mounts, where
`flock` is emulated or absent; version-CAS remains in force there and a degraded root is a
reportable `doctor` advisory rather than a hard refusal."

**Recommendation:** State the filesystem precondition where a caller meets it: document that the repository guard's cross-process exclusion assumes a local filesystem, and have `doctor` report the planning root's filesystem type with an advisory when it is a network or synced mount. No lock redesign — the CAS already covers the single-machine case, and ADR-0004 already names git as the multi-machine channel.

#### L1. task rename is the only scriptable mutation that does not auto-retry a pre-commit conflict · **Status:** open

**File:** internal/core/service_task.go:678 | **Component:** core/task-rename
**Effort:** XS · **Urgency:** eventually

**Class:** implementation drift
**Anchored to:** `docs/ARCHITECTURE.md` — "scriptable mutations auto-retry pre-commit conflicts in `core.Service` (bounded + jittered, so agents don't reimplement the loop)"; ADR-0003 2026-09-08

`core.Service.RenameTask` calls the store exactly once:

    // internal/core/service_task.go:678-686
    func (s *Service) RenameTask(slug, newTitle string, dryRun bool) (TaskRenameReceipt, error) {
        result, err := s.store.RenameTask(slug, newTitle, dryRun)
        receipt := taskRenameReceipt(result)
        if err != nil && result.Committed { ... }
        return receipt, err
    }

There is no retry loop, and `retryOnConflict` is not used. Every other scriptable mutation —
`set`, `append`, `move`, `defer`, audit finding writes, dependency edges, graph repair,
lifecycle transitions, all three Thread families — retries a bounded, jittered pre-commit
conflict.

The conflict rename can raise pre-commit is the ordinary kind. `store/rename.go:70-76`
captures the caller's source version *before* waiting on the guard, then rejects the
operation if the source moved while it waited:

    // The pre-lock source version is the intent boundary: a competing rename or
    // source edit that completed while this caller waited makes this operation stale
    if err := verifyUnchanged(s.resolvePath, source.id, source.path, source.version, "task", "rename"); err != nil {
        return result, err
    }

Nothing has been written at that point — `result.Committed` is false — so the same
re-read-and-re-plan that makes a lifecycle retry safe applies verbatim. In practice this
means `tskflwctl task rename` is the one mutation where two cron agents touching the same
repo hand the loser exit 14 instead of quietly succeeding on the second attempt, and the
agent has to reimplement the loop that `docs/ARCHITECTURE.md` says it should not need to.

ADR-0003's 2026-09-08 amendment specifies rename's recovery contract in unusual detail —
convergent durable prefix, resume by stable id, the destination-written/source-retained
state that must *not* be retried blindly — but says nothing about auto-retry of a failure
that wrote nothing. The silence reads as unconsidered rather than decided: the amendment's
careful separation of committed from pre-commit is exactly the distinction that would make
the retry safe.

**Compounding cost:** Small and slow — this is one call site, and the fix is a few lines
once M1's combinator exists. It is filed because it is the cheapest possible demonstration
of M1's thesis: with no shared retry combinator, adding a guarded mutation means remembering
to write the loop, and here it was simply not written. Left alone it also quietly erodes the
`ARCHITECTURE.md` guarantee agents are told to rely on, and that guarantee is the kind of
thing an agent only discovers is false under contention.

**Follow-up:** Worth confirming with the author whether rename's non-retry was a deliberate
"a destructive multi-document operation should always surface to a human" call. If so the
decision belongs in ADR-0003 as an explicit carve-out rather than as an absence — and
`ARCHITECTURE.md`'s blanket claim needs the exception noted.

**Recommendation:** Wrap the RenameTask call in the same bounded retry the other guarded mutations use, gated on the durability predicate M1 introduces (retry only while !result.Committed). A pre-commit rejection wrote nothing, so replaying it re-reads the tree and re-plans exactly as a lifecycle retry does; a committed prefix must keep today's no-retry behaviour.

## What audited clean

- **The crash-durable write sequence is complete, including the step most implementations
  miss.** `writeFileAtomic` does same-dir temp → `fsync(temp)` → `rename` → `fsync(parent dir)`
  (`store/atomic.go:48-62`). Satisfies the cited durability sequence in full
  (ArcadeData/arcadedb#7465; google/renameio).
- **The directory fsync is attempted and tolerated, not assumed.** `syncDir`
  (`store/atomic.go:64-73`) ignores the error with a stated rationale — exactly the posture
  the cited issue recommends ("it has to be attempted and its failure tolerated, not
  assumed") rather than the stricter-looking policy that would break FUSE and network mounts.
- **Non-Unix fails closed instead of pretending.** `lock_other.go:11-16` refuses repository
  mutation because "Silent no-op locking would make a successful CAS claim untrue." Matches
  ADR-0006 2026-08-27 §3 and the known-unreliability of atomic replacement on Windows
  (golang/go#22397 via google/renameio).
- **The in-process guard closes `flock`'s descriptor-independence hole.** The keyed mutex
  (`store/lock.go:28-31,63-67`) exists so two `*FS` values over one canonical root serialize
  within a process — the pitfall `flock(2)` warns about verbatim, anticipated before a
  long-lived adapter exists to trigger it.
- **Lock keys are canonicalized before use.** `normalizeRepositoryLockKey`
  (`store/lock.go:37-49`) resolves symlinks, absolutizes and cleans, so two spellings of one
  root cannot take two "exclusive" locks.
- **Planner exclusion is enforced at reads, not just writes.** 49 Store entry points call
  `rejectRepositoryPlannerCall`, which is what makes ADR-0006 2026-08-27 §2's "every Store
  entry point for that root" a fact rather than a claim — and it converts re-entry into an
  attributable `ErrConflict` instead of a self-deadlock (`store/lock.go:89-98`).
- **The CAS refuses to misname a failure.** `verifyUnchanged` (`store/cas.go:139-170`)
  deliberately does not collapse `ErrAmbiguous` or an I/O error into a conflict, because
  retrying those "would burn the whole OCC budget and then hand back 'changed on disk;
  retry': advice that cannot work." Error taxonomy treated as a correctness concern, not
  cosmetics.
- **The whole-snapshot CAS covers unreadable records.** `verifyTaskGraphSourceSnapshot`
  (`store/cas.go:31-41`) and `sameThreadSource` (`:82-98`) compare opaque source revisions
  and fail closed on missing evidence (`SourceVersion != ""`), so a raw edit cannot hide
  behind an unchanged parse error.
- **Retry backoff is jittered and bounded for the right reason.** `retryBackoff`
  (`core/retry.go:23-31`) uses full jitter to de-correlate cron writers that woke on the same
  schedule, capped at 4 attempts so genuine contention surfaces loudly rather than spinning
  (`core/retry.go:11-15`).
- **Dry runs never claim durability.** `retryOnConflict` returns immediately for a dry run
  (`core/retry.go:63-66`), and graph dry-run holds the same guard for an authoritative
  preview while explicitly making no CAS promise about a later apply — ADR-0006 2026-08-27 §6.
- **Post-commit failures are never auto-retried.** `TaskLifecycleMutationFailure`
  (`core/task_lifecycle.go:163-184`) keeps the cause wrapped so classification survives while
  "retry loops can stop," including when cleanup itself wraps `ErrConflict`. Satisfies
  "Non-idempotent operations with no reconciliation must not auto-retry after uncertain
  delivery."
- **Rename models the one state that must not be retried.** `taskRenameFailureRemedy`
  (`core/task_rename.go:79-92`) distinguishes complete, destination-written-but-source-retained,
  and convergent-prefix, and tells the operator which is which — ADR-0003 2026-09-08.
- **Every recovery receipt reaches the machine contract.** Eight typed envelopes in
  `cli/exit.go:84-122`, each carrying `committed` (`wire/dependency.go:222`,
  `wire/task_rename.go:18`, `wire/thread.go:299,339,475`, `wire/dependency_repair.go:67`), so
  an agent never parses prose to learn whether work became durable.
- **Batch partial failure leaves a coherent receipt.** `runMoves` (`cli/moves.go:32-88`)
  attempts every slug, reports each, and prefers a sentinel-bearing error so the summary exit
  code names the most actionable cause rather than the first in argv.
- **The TUI reloads on a committed failure instead of showing a stale view.**
  `tui/entity.go:343-375` unwraps the committed receipt into a `movedMsg` with a warning —
  ADR-0006 2026-08-28 lifecycle consequences.
- **`lint --fix` takes the write lock for the batch and refuses what it cannot prove.**
  `store/fix.go:29-38` locks because it is "a batch SECOND writer"; `:55-62` refuses
  graph-owned repairs rather than manufacturing unchecked edges, and reports the refusal as a
  `Skipped` result because "calling a refusal a fix is how a tool loses trust"
  (`domain/problem.go:21-25`).
- **Crash orphans are swept conservatively.** `sweepStaleTemps` (`store/atomic.go:138-158`)
  is scoped by the tool's own filename prefix *and* a one-hour age, so it cannot touch a user
  file or a live write.
- **Repair is a separate capability, not a mode flag.** `TaskGraphRepairStore`
  (`core/store.go:105`) is the only port permitted to accept a broken graph, and it proves
  structural non-regression per step (`store/graphrepair.go:135-160`) — ADR-0006 2026-09-05.

## Related-task observations (propose-only)

- ⚠️ **M2 partially adjacent** to
  `planning/tasks/6g5vm4efjcdv-make-repository-lint-load-diagnostics-adapter-neutral.md`
  (ready-to-start, epic 21, tier 3/low). That task replaces the shared unreadable-file bucket
  with entity kind, stable identity and an *optional repair location*, and touches the same
  struct and the same `lint --json` envelope M2 does. It is not the same change: its
  acceptance criteria say nothing about a defect kind or a repair applicability, and it lists
  "Changing lint rules" as out of scope. Both sides cross-referenced; M2 stays open.
  **Sequencing is the human's call** — landing them together would avoid two consecutive
  schema bumps on one envelope, but widening a ready-to-start task is not this audit's
  decision to make.
- Possible home for M2's eventual task: `26-frontmatter-schema-declared-validation-contract`
  (25% done) is about to derive lint, `schema task` guidance and the `--json` contract from
  one declared field registry — the moment a defect kind is cheapest to mint. Its design-first
  task `6fkkz41cax80-adr-close-frontmatter-schema-policy-questions` (next-up) is where the
  policy question would naturally be settled first.
- L1 is deliberately filed as a Low rather than folded silently into M1: if rename's
  non-retry turns out to be a considered "always surface a destructive multi-document
  operation to a human" decision, the right outcome is an explicit carve-out in ADR-0003 plus
  a corrected claim in `docs/ARCHITECTURE.md` — not a code change. That call belongs to the
  author, not to this audit.
- No FULL overlaps were found for any Medium+ finding, so **no task frontmatter or body was
  annotated this run**. The only file this audit changes is the audit itself.

## Candidate tasks

<!-- candidate-tasks:v1 · ○ open · ● in-progress · ✔ fixed · → tracked · ◌ deferred · ◌ superseded · ✘ wontfix -->
<!-- Add or replace one row with `tskflwctl audit finding <audit> <code> --candidate "<one line>"`; an empty value removes it. -->

- ○ M1 · open — tskflwctl task new "Give guarded mutations one named durability predicate and one retry combinator" --epic 21-code-quality-architecture-hardening
- ○ M2 · open — tskflwctl task new "Give lint diagnostics a closed kind vocabulary and a declared repair applicability" --epic 26-frontmatter-schema-declared-validation-contract
- ○ M3 · open — tskflwctl task new "Declare and diagnose the repository guard's local-filesystem precondition" --epic 21-code-quality-architecture-hardening
- ○ L1 · open — Fold into the M1 retry-combinator task — rename becomes the eighth call site
