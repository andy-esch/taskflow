---
schema: 1
id: 6gm8g0f66k80
bucket: open
area: concurrency-and-atomicity
date: "2026-10-16"
updated_at: "2026-10-16"
---

# Audit: concurrency-and-atomicity — 2026-10-16

> Create findings through `tskflwctl audit finding new`; update them through
> `tskflwctl audit finding` so identity, status, resolution, and managed candidate
> metadata stay synchronized. Never hand-edit those fields.

Routine: `code-quality-audit` · lens `concurrency-and-atomicity` · ISO week
`2026-W42`, slot `Fri` (index 1 — `(42 * 2 + 1) mod 6`).

## Punch list

- M1. lint --fix id repair relocates a file with the clobbering writer, silently widening its mode  (effort: S · urgency: soon)
- M2. An interrupted id-repair relocation leaves a duplicate id that freezes every guarded mutation  (effort: S · urgency: soon)
- L1. A replaced watcher's final event can pin the TUI live-reload indicator to off  (effort: XS · urgency: eventually)

M1 and M2 are the same eight-line block in `internal/store/fix.go` seen through two
questions, and share one fix; M1 is reproduced, M2's end state is reproduced and its
trigger reasoned.

## Files audited

- **Signal**: `internal/core/thread_mutation.go` — churn 5 in the last 30 days and
  ~318 references to its exported surface from other files; top score
  (`log(churn+1) × log(blast+1)`) inside the lens surface, and it owns the
  membership/lifecycle policy that a store guard has to enforce.
- **Signal**: `internal/core/task_lifecycle.go` — churn 3, ~363 references; the other
  core guarded planner, and the one whose authorization reads a file body.
- **Adjacency** (from `thread_mutation.go`): `internal/store/threadmutation.go` —
  the hop across the hexagonal seam, from the invariant to the guard that enforces it.
- **Adjacency** (from `task_lifecycle.go`): `internal/store/lifecyclemutation.go` —
  the same hop; the lens's high-yield shape (read a snapshot, validate against it,
  write) lives here if anywhere.
- **Random**: `internal/store/auditstore.go` — seeded pick from the 35 non-test files
  in the lens surface over 50 lines.

Read as supporting context rather than as picks, because the lens questions pointed
there: `store/atomic.go`, `store/cas.go`, `store/lock.go` + `lock_unix.go` +
`lock_other.go`, `store/resolve.go` (the source-version token), `store/threadapply.go`
and `core/thread_apply.go` (the only multi-file durable path), `store/fix.go`,
`core/retry.go`, `tui/watch.go`, `tui/commands.go`, `tui/model.go`, `tui/session.go`,
and `internal/config` + `internal/userconfig` (their own copies of the atomic
helpers). Both Mediums are in `fix.go`, which the five picks did **not** surface — it
came out of the lens question "is every write atomic, or is there a raw write hiding
somewhere", which ranked every write path in `internal/store` by whether it pairs an
atomic write with a content CAS. `fix.go` was the only file with writes and no CAS.

## Commands run

- `go build -o bin/tskflwctl ./cmd/tskflwctl` — clean. `just` is **not installed** in
  this environment, so the spec's documented fallback was used; `just test` and
  `just lint` were likewise unavailable and the underlying recipes were run directly
  (see the PR's Validation section for what that covered and what it did not).
- `git status --porcelain` — empty before the run; only `planning/` afterwards.
- `TZ=America/New_York date "+%A %Y-%m-%d %V"` → `Friday 2026-10-16 42` → lens index 1.
- `./bin/tskflwctl audit list --json` — 9 audits in the `open` bucket, below the
  backpressure threshold of 10, so this is a normal authoring run.
- `./bin/tskflwctl epic list` — for the urgency scoring and real epic ids in the
  candidate rows.
- Churn: `git log --since="30 days ago" --name-only --pretty=format:` over the lens
  paths; blast radius by counting files referencing each file's exported identifiers.
- `git log -S'createFileAtomicExactMode'` / `-S'repairInvalidID'` — both resolve to
  the same merge (#253), so the asymmetry between the two relocation paths was *not*
  introduced by one landing after the other; it is simultaneous divergence.
- **M1 probe** (a throwaway planning repo under the session scratchpad, never the
  working tree): `tskflwctl init`; `epic new`; `task new`; renamed the task file to an
  id-led name whose id carries a non-Crockford `l` and matched the `id:` field;
  `chmod 600`; `lint --fix --dry-run` (previewed the repair); `lint --fix`. Result:
  source `-rw-------` → repaired destination `-rw-r--r--`. **M1 confirmed.**
- **M2 probe** (same repo): copied the repaired file to a second name carrying the
  same id to reproduce an interrupted relocation's end state. `task list` printed both
  rows; `task start <id>` refused with "repository task graph is broken: duplicate
  stable task id … no source is uniquely authoritative"; `lint` exited non-zero
  naming both files. **M2's consequence confirmed**; its trigger (an interruption
  between the two writes) was reasoned from the code, not forced.
- Probe cleanup: `tskflwctl space forget probe` and `rm -rf` the scratch repo;
  `git status --porcelain` re-checked clean afterwards.
- L1 was **not** probed — see its confidence note.

## Findings

#### M1. lint --fix id repair relocates a file with the clobbering writer, silently widening its mode · **Status:** open

**File:** internal/store/fix.go:95 | **Component:** store
**Effort:** S · **Urgency:** soon

`FixFrontmatter`'s id-repair branch relocates a file by writing the destination and
then removing the source:

    if err := writeFileAtomic(renamed, fixed, 0o644); err != nil { return err }
    if err := os.Remove(path); err != nil { ... }

`writeFileAtomic` is the **overwrite** primitive. Its own doc comment explains that it
preserves an existing destination's mode by stat'ing it first and that `perm` is only
"the fallback for a file that doesn't yet exist" — and in a relocation the destination
never exists, so the fallback `0o644` always wins and the source's mode is dropped.

The package already owns the right primitive for exactly this operation.
`createFileAtomicExactMode` is documented as "the no-clobber create used when an
existing source file is being relocated. Unlike an ordinary create, the destination
must retain the source's effective permission bits even under a different process
umask." `rename.go:121-137` — the other relocation path in this same package — stats
the source and passes `sourceInfo.Mode().Perm()` to it, and maps `os.IsExist` to
`ErrConflict`. The fixer does neither.

Secondary consequence: because the write is `os.Rename`-based rather than `O_EXCL`,
a destination that appears after `repairInvalidID`'s `os.Stat(target)` guard
(fix.go:211) and before the write is silently overwritten rather than refused. The
guard is a stat-then-write check of precisely the kind `createFileAtomic`'s doc
comment says `O_EXCL` exists to eliminate ("no stat/write race"). Cooperating writers
are excluded by the repository write lock this path holds, so only a raw editor can
reach the window — the same actor every other write path in this package defends
against with `O_EXCL` plus a content CAS.

**Failing scenario:** `chmod 600 planning/tasks/6fbj87000lt6-bad-id.md` (an id-led
name whose id contains a non-Crockford `l`), then `tskflwctl lint --fix`. Verified on
a scratch planning repo with the binary built from this checkout: the source was
`-rw-------`, and after the repair the file at the canonical name
`6fbj870001t6-bad-id.md` is `-rw-r--r--`. A mode the user (or a synced/encrypted
setup) deliberately narrowed is widened with no mention in the receipt.

**Why tests didn't catch it:** the mode invariant is pinned for the other two write
paths and not for this one. `rename_test.go:77-90` and `187-206` chmod 0600 and
assert the relocated file is still 0600; `audit_fixes_test.go:75-83` asserts the same
for an in-place audit body write. `fix_test.go:143` covers this exact repair
(`TestFixRepairsMisspelledIDInFilenameAndFrontmatter`) but asserts only the new
filename, the moved `id:` field, and the removal of the old path — and its fixture is
written at the default 0644, which masks the regression.

**Tightening (adjacent):** add the 0600 assertion to
`TestFixRepairsMisspelledIDInFilenameAndFrontmatter` so the invariant is pinned for
all three relocation/replacement paths, and have the `FixResult` note the destination
path rather than only the source.

**Recommendation:** In fix.go's rename branch, stat the source and call createFileAtomicExactMode(renamed, fixed, sourceMode) instead of writeFileAtomic, mapping os.IsExist to ErrConflict as rename.go does.

#### M2. An interrupted id-repair relocation leaves a duplicate id that freezes every guarded mutation · **Status:** open

**File:** internal/store/fix.go:91 | **Component:** store
**Effort:** S · **Urgency:** soon

The same relocation block is a two-step durable change — `writeFileAtomic(renamed,
…)` then `os.Remove(path)` — with nothing between them and no receipt for the half
that landed:

    if !dryRun {
        if err := writeFileAtomic(renamed, fixed, 0o644); err != nil { return err }
        if err := os.Remove(path); err != nil {
            return fmt.Errorf("remove %s after rename: %w", path, err)
        }
    }
    results = append(results, domain.FixResult{Path: path, Changes: changes})

Two things follow. First, the `results` row is appended only *after* both steps, so a
failing `os.Remove` returns the error with that file absent from the receipt — the
caller is told the run failed but not that the destination was already written. The
surrounding code is explicit that it means to report partial progress ("On a
mid-run write failure the results so far are returned ALONGSIDE the error", fix.go:132),
and this branch is the one place that drops its own row. Second, the source bytes are
never re-checked before the delete, so a raw edit landing between the two steps is
removed rather than refused.

`rename.go` solves both for the identical operation: it sets
`result.Committed`/`result.DestinationWritten` the moment the destination exists
(rename.go:140-141), then calls `verifyTaskRenamePath(source.path, source.version)`
before unlinking, with the comment "after the destination exists, a conflict
intentionally leaves both copies for explicit inspection" (rename.go:154-157).

**Failing scenario:** interrupt (SIGKILL, ENOSPC, an EPERM on the unlink from a
read-only parent or an immutable flag) between the two writes. Both names now exist
carrying the same stable id. Verified on a scratch planning repo by reproducing that
end state: `task list` happily prints both rows, but every guarded mutation refuses
repo-wide — `task start 6gm8f3ww4t1z` fails with "repository task graph is broken:
duplicate stable task id … no source is uniquely authoritative", and `lint` exits
non-zero for both files. So one interrupted cosmetic id repair blocks all lifecycle,
dependency, and Thread writes in the space until a human deletes one file.

This is loud rather than silent — `lint` names both paths and the remedy — which is
why it is Medium and not High. But the fixer is the tool a user reaches for *because*
the tree is already unhealthy, and it is the one write path in the package whose
interrupted state is neither re-checked nor reported.

**Context:** severity sits below urgency here. The probability is low (a misspelled
id is rare by the fixer's own account) but the blast radius is the whole planning
space, and the fix is the same small patch M1 already needs in the same block, so
doing them together costs one review instead of two.

**Recommendation:** Re-verify the source bytes immediately before os.Remove and record the written destination in the FixResult, so an interrupted repair reports the committed half instead of returning an error that omits it.

#### L1. A replaced watcher's final event can pin the TUI live-reload indicator to off · **Status:** open

**File:** internal/tui/model.go:576 | **Component:** tui
**Effort:** XS · **Urgency:** eventually

`waitForFS` closes over one `*watcher` and blocks on its `fsnotify` channels
(watch.go:331-349). An atlas navigation replaces the watcher and closes the previous
one through a batched command (`session.go:181-186` issues `waitForFS(m.watch)` for
the new watcher alongside `closeWatcher(oldWatcher)`), so for a moment both the old
and the new listener are outstanding.

The ordinary shutdown is clean: `close()` closes the fsnotify watcher, the blocked
receive observes `ok == false`, and the closure returns a nil message. The gap is the
other branch of that `select` — if an event was already queued, the *old* closure
returns `fsEventMsg{health: w.reconcile()}`, and `reconcile` on a closed watcher
reports `watchUnavailable` (watch.go:93-96). The reducer applies that health
unconditionally:

    case fsEventMsg:
        m.watchOff = msg.health == watchUnavailable
        m.watchDegraded = msg.health == watchDegraded

`m.watch` is the new, healthy watcher by then, so the model records live reload as
off for a watcher that is working. The same applies to `watcherReconciledMsg`.

Nothing is lost and no reload is missed: the new listener issued by the session swap
is still live, so the next real filesystem change delivers a healthy `fsEventMsg` and
clears the flag. The visible cost is a wrong "live reload off" indicator, which in a
quiet space persists until something in the tree changes — exactly the situation in
which the user would trust the indicator and start refreshing by hand.

The codebase already has the idiom this message type is missing: `debounceMsg` carries
`gen` and is dropped when `msg.gen != m.dirtyGen` (model.go:590-592), and `errMsg` /
`listLoadedMsg` are gated on `msg.gen == m.tabs[i].loadGen`. Only the two watcher
messages carry no identity for the resource they describe.

**Confidence: unverified — needs a probe.** The window is a queued fsnotify event
racing a watcher swap, and I did not construct it. A probe would stub `addPath`
(the seam `newWatcherWithPathAdder` already exposes), push an event into the old
watcher's channel, run the atlas-navigation swap, and assert `watchOff` is false
afterwards. Labelled Low on that basis rather than Medium.

**Recommendation:** Carry the watcher generation on fsEventMsg and watcherReconciledMsg and ignore a message whose generation is not current, matching the guard debounceMsg and errMsg already use.

## What audited clean

- `internal/core/thread_mutation.go` (signal) — **clean for this lens.** A pure
  planner: no filesystem, no clock (`now` is an injected semantic input, validated
  non-zero), no Cobra. The one thing worth checking twice was
  `analysis.Changed = !slices.Equal(thread.Tasks, after.Tasks)`, which compares
  on-disk order against a sorted `after` and would therefore report a change for a
  pure reorder — but `domain.ValidateThreadDocument` refuses an unsorted `tasks`
  list (`domain/thread.go:116-119`), so a readable Thread is always sorted when it
  reaches the planner and the comparison is exact. Callback-owned slices are re-cloned
  before the planner sees them (`threadmutation.go:120-122`), so a hostile planner
  cannot rewrite the snapshot that authorizes its own plan.

- `internal/core/task_lifecycle.go` (signal) — **clean for this lens.** Also pure.
  The read/validate/write seam it feeds is sound: `MutateTaskLifecycle` reads the body
  for acceptance-criteria policy at `lifecyclemutation.go:68` and re-reads the file for
  materialization at `:210`, which looked like a drift window until the token is traced
  — `scanDirSourcesWithReader` hashes the *whole file bytes* before parsing
  (`resolve.go:58-64`), so a body-only raw edit between the two reads changes the
  graph's `SourceVersion` and is caught by the pre-write snapshot CAS at `:113`.
  `SameSourceSnapshot` additionally requires both versions to be non-empty and equal
  and compares unreadable-file problems too (`dependency_graph.go:1022-1036`), so a
  raw edit cannot hide behind an unchanged parse error.

- `internal/store/threadmutation.go` (adjacency, across the hexagonal seam) —
  **clean.** The full sequence is present and in the right order: repository guard,
  snapshot, planner under a re-entry guard, pure validation, materialization, whole-
  repository re-read and CAS, per-file `verifyUnchanged`, single `writeFileAtomic`.
  Release failures are joined into the returned error rather than swallowed, and the
  `Committed` flag lets `service_thread.go:93` stop the OCC retry loop on a durable
  write.

- `internal/store/auditstore.go` (random) — **clean.** `MoveAudit` reads and validates
  *outside* the lock and then takes it only around verify→write, which initially reads
  like the lens's high-yield smell; it is not, because the CAS token is
  `hashContent(content)` from that earlier read, so any drift is rejected rather than
  overwritten. The lock only has to make verify→write atomic, which is exactly what it
  is scoped to do.

- Write-path sweep (lens question: "is there a raw `os.WriteFile` hiding somewhere")
  — every entity write in `internal/store` goes through `writeFileAtomic` /
  `createFileAtomic`, and all but `fix.go` pair it with a content CAS. Outside the
  store, `internal/config` and `internal/userconfig` keep their own copies of the
  helpers (they cannot import `store`) and lock the stable config *directory* rather
  than the file, since atomic writes replace the inode. `backfillID` /
  `backfillPointerID` do an unlocked read-modify-write and were the best lost-update
  candidate found, until their only caller turned out to be `migrateUnlocked`, which
  `migrateWithIDGen` wraps in `writeLock(dir)` after re-reading under it
  (`migrate.go:66-83`) — the naming makes the precondition explicit.
  `AddSpace`/`ForgetSpace`/`SetPresentation`/`appendTrackedRepo` all re-read inside
  their critical section.

- Retry sweep (lens question: "retry loops that can retry a *committed* write") —
  every guarded loop gates on durability: `!result.Committed` in `service_task.go:656`,
  `service_thread.go:93` and `:179`, `len(result.AppliedTaskIDs) > 0` in
  `dependency_operations.go:161`, `result.Committed || result.Complete` in
  `service_thread_apply.go:69`. The generic `retryOnConflict` has no such gate and says
  so in its contract ("Callers must also ensure a conflict precedes any durable
  write"); its twelve call sites each wrap exactly one single-file transform whose
  conflict is raised before the write, including the one inside a loop
  (`finding.go:323`), where only the single audit's transform is retried.

- `internal/store/threadapply.go` (read as context for the multi-write question) —
  the one genuinely multi-file durable path, and deliberately "resumable rather than
  transactionally atomic". Its per-file writes cannot collide: `PrepareThreadApply`
  builds `TaskWrites` from a de-duplicated `owners` set (`thread_apply.go:418-432`),
  so no two writes in a batch target the same file with the same stale `ifVersion`,
  and the plan is re-prepared after the dependency prefix lands so an undone edge
  cannot produce a false `Complete`. `Operations[len-1]` at `threadapply.go:164`
  assumes the thread operation is last; it always is, because it is appended
  unconditionally after the dependency loop — a positional coupling, not a defect, and
  `markThreadApplyThreadState` searches by `Kind` for the same thing.

- TUI in-place filters (`commands.go:113`, `:274`, `:319`) compact with
  `out := records[:0]`, aliasing the caller's backing array. Safe as written:
  `taskGraphRecords` returns a freshly built or explicitly copied slice on every path
  (`service_task.go:210-226`) and nothing caches it, and the pre-filter `fullRefs` are
  materialized before the compaction runs.

## External research

Two Medium findings surfaced, so this is the step-11 pass: patterns checked against
the audited lens, including ones today's code does *not* violate.

- **Mode preservation on atomic replace is a named, first-class concern, not a nit.**
  Google's `renameio/v2` — the reference Go implementation of temp-file-plus-rename —
  exposes `WithExistingPermissions`, which "configures the pending file to try to use
  the permissions from an already existing target file", and documents that its
  default is deliberately restrictive (`0600 & ^umask`) precisely because the temp
  file's mode, not the target's, is what survives a naive rename.
  `cplieger/atomicfile` ships the same idea as `WithPreserveMode()`, stat'ing the
  target and falling back when it cannot. taskflow has both halves of this already
  (`writeFileAtomic` stats the destination; `createFileAtomicExactMode` carries the
  source's bits through a foreign umask) — M1 is purely that one relocation call site
  reaches for neither.
  <https://pkg.go.dev/github.com/google/renameio/v2> ·
  <https://pkg.go.dev/github.com/cplieger/atomicfile/v2>

- **Antipatterns this code does not have.** `renameio`'s documented failure modes are
  a temp file on a different mount point (rename cannot cross filesystems) and a
  missing `fsync` before the rename leaving a zero-length file on POSIX. `stageTemp`
  avoids both: `os.CreateTemp(dir, …)` stages in the destination's own directory and
  the file is `Sync`'d before `Close` and rename, with a best-effort `syncDir`
  afterwards for the directory entry itself. Tailscale's `atomicfile` is the
  counter-example worth noting — a plain `os.Rename` on non-Windows with no mode
  handling at all.
  <https://pkg.go.dev/github.com/google/renameio/v2>

- **Proactive: the repository flock will fail, loudly, on an NFS-mounted planning
  space.** `platformWriteLockChecked` takes `flock(LOCK_EX)` on a read-only handle to
  the repo root *directory*. On Linux that is fine locally, but `flock(2)` records
  that since 2.6.12 NFS clients emulate `flock()` as whole-file `fcntl()` byte-range
  locks, and that "an exclusive lock also requires the file to be opened for writing"
  — which a directory handle can never be. So on NFS every mutation would abort with
  `acquire write lock: …` rather than a diagnosed "locking is unsupported on this
  mount". That fails *closed*, which is the same deliberate choice `lock_other.go`
  makes for non-unix targets, so it is a diagnostic-quality question and not a safety
  one — worth a sentence in the error, not a code change. (`local_lock` in `nfs(5)`
  is the escape hatch, and it makes locks local, i.e. unsafe across hosts.)
  <https://manpages.ubuntu.com/manpages/noble/man2/flock.2.html>

- **The stale-async-result guard is the established Bubble Tea idiom.** The Bubble Tea
  docs are explicit that commands "run asynchronously in goroutines" and that `Batch`
  gives "no ordering guarantees about the results", with all state changes confined to
  `Update` — which is the mechanism behind L1 and the reason the fix belongs in the
  reducer rather than in the watcher. taskflow already applies the counter idiom to
  `debounceMsg`, `errMsg`, and `listLoadedMsg`; L1 is only asking the two watcher
  messages to join them.
  <https://www.mintlify.com/charmbracelet/bubbletea/concepts/commands>

## Related-task observations (propose-only)

- Possibly ready to close: `planning/audits/6ggqwdz14nzx-2026-10-05-arch-data-model-and-storage.md`
  reports `ready_to_close: true` with 0 open findings in `audit list --json`. Not
  touched by this run, which was not a backpressure run (9 open audits against a
  threshold of 10) — noting it because authoring this audit brings the open count to
  10, so the next run will enter triage with this one already terminal.
- No open task was found to own M1, M2, or L1. The nearest fingerprint matches —
  `6fq9zy126tas-invalid-non-crockford-entity-id-is-mishandled` (the task that built
  `repairInvalidID`), `6fhnydm021es-scheme-2-references-and-rename-verb-with-link-cascade`
  (the task that built `rename.go`'s relocation sequence), `6fbj87000m0a-store-write-path-hardening`,
  `6fakbec03cct-triage-external-review-fix-fixer-comment-o-excl-config-anchor-diagnostics`
  and `6fes83r03s90-crash-recovery-signal-guard-and-duplicate-slug-repair` — are all
  **completed**, so the overlap is NONE (keyword hits only) and nothing was annotated.
  Worth flagging for the reader: `6fq9zy126tas`'s implementation notes describe the id
  repair's refusal cases in detail and never discuss the write primitive, so M1/M2 read
  as a gap in that work rather than a trade-off it considered and accepted.

## Candidate tasks

<!-- candidate-tasks:v1 · ○ open · ● in-progress · ✔ fixed · → tracked · ◌ deferred · ◌ superseded · ✘ wontfix -->
<!-- Add or replace one row with `tskflwctl audit finding <audit> <code> --candidate "<one line>"`; an empty value removes it. -->

- ○ M1 · open — `tskflwctl task new "lint --fix id repair must relocate through the no-clobber, mode-preserving create" --epic 21-code-quality-architecture-hardening --tags store --tier 2 --priority high --description "FixFrontmatter.repairInvalidID relocates via writeFileAtomic, so the repaired file loses the source mode and an existing destination would be clobbered; adopt rename.go createFileAtomicExactMode sequence and pin the 0600 invariant."`
- ○ M2 · open — `tskflwctl task new "Make the lint --fix id-repair relocation resumable and self-describing" --epic 21-code-quality-architecture-hardening --tags store --tier 3 --priority medium --description "The create-then-remove pair in FixFrontmatter has no pre-remove CAS and no receipt for the written destination, so an interrupted repair leaves a duplicate stable id that breaks the task graph repo-wide; mirror rename.go verifyTaskRenamePath plus a DestinationWritten-style receipt."`
- ○ L1 · open — `tskflwctl task new "Generation-guard the TUI watcher messages like the debounce tick" --epic 18-tui-bubble-tea-interactive-planning-browser --tags tui --tier 4 --priority low --description "fsEventMsg and watcherReconciledMsg carry no watcher identity, so a closed watchers last event can set watchOff for the live replacement; stamp a watcher generation and drop stale messages as debounceMsg already does."`
