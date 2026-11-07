---
schema: 1
id: 6gv0qm4src7v
bucket: open
area: concurrency-and-atomicity
date: "2026-11-06"
updated_at: "2026-11-06"
---

# Audit: concurrency-and-atomicity — 2026-11-06

> Create findings with `audit finding new`; update them with `audit finding`.

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

#### L1. editFile defers the repository guard release inside its editor retry loop · **Status:** open

**File:** internal/store/edit.go:120 | **Component:** store / repository guard
**Effort:** XS · **Urgency:** eventually

`editFile` is the shared interactive-editor write path behind `EditTask`, `EditAudit`,
`EditEpic` and `EditResearch`. Its `lock()` call sits **inside** the `for` retry loop that
reopens the editor on a parse failure, and the guard is released with `defer unlock()`
(`internal/store/edit.go:120`). In Go a `defer` is scoped to the enclosing *function*, not
the loop iteration, so a release registered in one attempt does not run until `editFile`
returns.

The comment three lines above states the invariant this path is built around:

> Take the write lock only NOW — after the (long, interactive) editor returned and the edit
> parsed — so the flock covers just the brief verify→write, never the whole editor session.

That invariant holds **today**: the loop's only `continue` is the parse-failure branch, which
sits *before* `lock()`, and every statement after `lock()` returns from the function. So the
guard is in fact held for one brief verify→write and the audit found no way to make it fail.
This is a latent-trap observation, not a live defect.

It is worth recording because of what the trap costs if it is ever tripped. The guard is
`flock(LOCK_EX)` on the **planning root**, repo-wide, and acquisition is unbounded
(tracked separately by `6g6yjm16dkgt`). A future `continue` placed after the lock — e.g.
reopening the editor when `recheck()` reports a CAS conflict, which is the natural next
feature for this loop, since a conflict is exactly the case where the user would want their
text back — would hold the repository-wide lock across the *next* interactive `$EDITOR`
session. Every other writer in the repository, including the cron agents the lock comment
names as the motivating concurrent writers, would block for as long as the editor stayed
open. The failure would not look like a bug in `editFile`; it would look like the tool
hanging.

For contrast, `writeBody` (`internal/store/body.go:67`) uses the same lock/recheck/write
tail with no enclosing loop, so its `defer unlock()` is already correctly scoped. The
divergence is only in `editFile`.

**Recommendation:** Release explicitly on each exit path, or lift the post-parse lock-verify-write tail into a helper so its defer is scoped to one attempt.

## Run context

Routine: `code-quality-audit` · lens `concurrency-and-atomicity` · ISO week `2026-W45`,
slot `Fri` (index 1 = (45 × 2 + 1) mod 6). Five audits open at the backpressure check
(threshold 10), so this is an authoring run.

`just` is not installed in the routine's container; the Justfile recipes were replaced
with their underlying Go commands and that substitution is recorded under **Commands run**.

## Punch list

- `L1. editFile defers the repository guard release inside its editor retry loop  (effort: XS · urgency: eventually)`

No Critical, High, or Medium findings. One Low. The lens's core surface — atomic
replacement, the repository guard, version-CAS, and the guarded multi-write planners —
audited clean, and this run's main result is the evidence for that claim rather than the
single Low.

## Files audited

- **Signal**: `internal/core/store.go` — highest churn in the lens surface (12 commits
  in 30 days) and the highest blast radius in the tree: 236 files import `internal/core`,
  and this file declares every guarded mutation port (`TaskGraphMutationStore`,
  `TaskLifecycleMutationStore`, `TaskGraphRepairStore`, `ThreadApplyMutationStore`) plus
  the planner-callback contract they all share.
- **Signal**: `internal/store/lock.go` + `lock_unix.go` / `lock_other.go` — the lens's
  declared `*lock*` surface and the single point where the same-process mutex, the
  planner re-entry sentinel, and the platform `flock` compose into one guard.
- **Adjacency** (from `core/store.go`, across the hexagonal seam) —
  `internal/store/graphmutation.go`, `lifecyclemutation.go`, `graphrepair.go`,
  `threadapply.go`: the four store-side implementations of those ports. Read together
  because the lens's high-yield question ("reads a snapshot, validates against it, and
  writes without re-checking the snapshot") can only be answered by comparing them.
- **Adjacency** (from `lock.go`) — `internal/store/atomic.go` and `internal/store/cas.go`:
  the lock-free write primitive and the version-CAS the guard exists to make atomic.
- **Random** (from the lens surface, after the quality filter) —
  `internal/tui/watch.go`: the fsnotify/debounce half of the lens.

Supporting reads for context, not audited as primary targets: `internal/store/edit.go`,
`body.go`, `rename.go`, `fix.go`, `fix_relocation.go`, `create.go`, `fsstore.go`,
`internal/core/retry.go`, `dependency_operations.go`, `service_thread_apply.go`,
`dependency_graph.go`, `dependency_repair.go`, `internal/tui/model.go`, `entity.go`,
`commands.go`, `read_retry.go`, `internal/cli/command_safety.go`,
`internal/appwiring/wiring.go`.

## Commands run

```sh
git status --porcelain                 # empty before the run; only planning/ after
go build -o bin/tskflwctl ./cmd/tskflwctl    # `just` unavailable in this container
TZ=America/New_York date "+%A %Y-%m-%d %V"   # Friday 2026-11-06 45 → index 1
./bin/tskflwctl audit list --json            # 5 open audits — under the 10 threshold
./bin/tskflwctl epic list                    # urgency scoring context
git log --since="30 days ago" --name-only --pretty=format: \
  -- internal/store/ internal/core/ internal/tui/   # churn signal
grep -rn "os\.WriteFile\|os\.Create(" internal cmd --include=*.go | grep -v _test
                                       # → only internal/config and internal/tools/*gen*;
                                       #   no raw write anywhere in store/core/cli/tui
grep -rn "go func" internal/ cmd/ --include=*.go | grep -v _test   # → no matches
go vet ./internal/store/... ./internal/core/... ./internal/tui/...
go test -race ./internal/store/... ./internal/core/... ./internal/tui/...
./bin/tskflwctl -C . lint
./bin/tskflwctl -C . audit lint 2026-11-06-concurrency-and-atomicity
```

Results are recorded in the PR body. No mutation probe was run against `internal/`: this
routine is propose-only and has no write access there, so every claim below was settled by
reading, with the one claim I could not settle that way labelled as such.

## What audited clean

- `internal/store/atomic.go` — clean. `writeFileAtomic` implements the full
  temp-in-same-dir → `fsync` → `chmod` → `rename` → directory-`fsync` recipe, and
  `stageTemp` applies the mode *before* the rename rather than after, which is the
  ordering mistake the pattern is usually gotten wrong by. `syncDir` deliberately
  discards its error with a comment naming network/FUSE mounts as the reason.
  `createFileAtomicWithMode` cleans up the partial new file on the write, chmod, sync
  **and** close paths — the close path is the one most implementations miss.
  `sweepStaleTemps` is bounded by both the tool's own filename prefix and a one-hour age,
  so it cannot reach a user file or a live write.
- `internal/store/lock.go`, `lock_unix.go`, `lock_other.go` — clean. Lock ordering is
  consistent everywhere (registry mutex → per-root mutex → `flock`), and
  `repositoryLockKey` is computed *before* the registry mutex is taken, so no filesystem
  call happens under it. `lock_other.go` refuses mutation on platforms without a tested
  primitive instead of silently no-op locking, which is the right failure direction: a
  successful CAS claim that was never serialized would be worse than a refusal.
- `internal/store/cas.go` — clean. `hashContent` is SHA-256 over exact on-disk bytes,
  computed on read and never stored, and the reasoning for a *strong* validator (two
  writers touching different fields must both be caught) is correct. `verifyUnchanged`
  re-resolves the key *and* compares the path before comparing content, so a record that
  moved is a conflict rather than a write to the wrong file, and it correctly preserves
  `ErrAmbiguous` / I/O errors instead of flattening them into a retry invitation.
- The four guarded planner paths (`graphmutation.go`, `lifecyclemutation.go`,
  `graphrepair.go`, `threadapply.go`) — clean on the lens's high-yield question. Every one
  of them re-reads the whole-repository snapshot and re-verifies it *after* planning and
  *before* the first durable byte, and then CASes each individual file again immediately
  before its own replacement. `graphrepair.go` additionally re-verifies after every
  committed write and advances its expected snapshot, which is stricter than its siblings.
  `graphrepair.go` is the only one that inlines its per-file hash check rather than calling
  `verifyUnchanged`; that is safe here because the immediately preceding whole-snapshot
  comparison includes `LocalPath`, so path drift is already caught.
- Retry-after-commit — clean, and checked at every call site, because this is the lens
  question most likely to be gotten wrong. `retryOnConflict` (`internal/core/retry.go:62`)
  documents that callers must ensure a conflict precedes any durable write, and each
  multi-write caller enforces exactly that with its own guard rather than relying on the
  shared helper: `runDependencyMutation` refuses to retry once
  `len(result.AppliedTaskIDs) > 0`, `ApplyThreadPlan` refuses once `result.Committed`, and
  `RenameTask` — which can fail with `ErrConflict` after a committed cascade prefix — is
  deliberately not wrapped in a retry at all and returns a typed
  `TaskRenameFailure` carrying the durable prefix instead.
- `internal/store/fsstore.go` — clean, and load-bearing for everything above. `FS` has no
  mutable field after construction (no caches, no memoized resolver), so concurrent reads
  and writes through one `*FS` share no writable state. `internal/core/service.go` is the
  same: every assignment to a `Service` field lives in `NewService` or an `Option`. This is
  what makes the TUI's concurrent `tea.Cmd` goroutines safe without any locking in `core`.
- `internal/tui/watch.go` (the random pick) — clean. `reconcile` and `close` are
  serialized on the same mutex, `closed` is checked first so a reconcile cannot resurrect
  a closed watcher, and `w.fsw` is never reassigned after construction, so `waitForFS`
  reading it without the mutex is not a race. Attachment uses `os.SameFile` identity rather
  than path spelling, so an atomically replaced directory is re-attached.
  `nearestExistingWatchDirectory` explicitly refuses a filesystem root, which prevents one
  broken layout path from turning into an unbounded event source.
- `internal/tui/` generation discipline — clean, and notable given that this is the
  classic Bubble Tea race. Every list loader (`loadTaskList`, `loadEpicList`,
  `loadAuditList`, `loadResearchList`, `loadThreadList`) snapshots `t.statusView` and
  `t.loadGen` into locals *before* returning the closure, so no `tea.Cmd` goroutine reads
  through the `*entityTab` pointer that `Update` and `View` mutate. The debounce path is
  generation-stamped (`dirtyGen`) and I could not construct an interleaving that produces a
  duplicate outstanding `waitForFS` listener. There are no explicit goroutines anywhere in
  non-test code.
- Already-tracked items re-confirmed, deliberately **not** re-reported as findings:
  `writeFileAtomic` replacing a symlink rather than its target (tracked by
  `6g63jj1dh0sb`); unbounded, undiagnosed repository-lock acquisition (tracked by
  `6g6yjm16dkgt`); the `flock`-on-NFS concern, which the 2026-10-16 run of this same lens
  already recorded as proactive research and which `6g6yjm16dkgt` covers. The unbounded
  `repositoryGuards.byRoot` map was checked again and still matches its two prior
  dispositions (`6g45s2rm09pr` L6 `fixed` — documented rather than changed; and an explicit
  *rejected* in `6g7wqndnnvs2`), so it stays closed.

## External research

Fewer than three Medium-or-higher findings, so this run did the conditional research. The
useful result is negative in the productive sense: three of the four patterns searched are
ones this codebase already gets right, which is corroboration for the clean findings above
rather than new work.

- **Atomic file replacement on POSIX.** The convergent practitioner recipe is
  temp-in-target-dir → `fsync` → `rename` → parent-directory `fsync`, with permissions
  applied *before* the rename (applying them after can briefly expose the replaced file
  with the wrong mode), and with directory-`fsync` errors tolerated because some systems
  refuse it outright. `atomic.go` already does all four, in that order.
  — <https://awesome.ecosyste.ms/projects/github.com%2Fcplieger%2Fatomicfile>,
  <https://pkg.go.dev/github.com/BackendStack21/odek/internal/fsatomic>,
  <https://github.com/Widthdom/CodeIndex/issues/3409>,
  <https://pypi.org/project/atomicreplace/>
- **`flock` semantics for multiple descriptors in one process.** Locks belong to the open
  file *description*, not the process, so two independent `open` calls on the same file in
  one process contend with each other; descriptors shared via `dup`/`fork` do not. This is
  exactly the hazard `lock.go`'s same-process mutex is documented as covering, and it
  confirms the layered guard is not redundant. Also confirms the NFS emulation caveat that
  `6g6yjm16dkgt` already tracks. — <https://man.archlinux.org/man/flock.2.en.txt>,
  <https://manpages.debian.org/jessie-backports/manpages-dev/flock.2.en.html>
- **Bubble Tea `tea.Cmd` data races.** The reported real-world failures are all the same
  shape: a `tea.Cmd` closure reads model state through a pointer that `Update` writes
  (`BlinkCmd` on a cursor, `nextFrame` against `SetPercent`). The community guidance is to
  capture values and return them in a message rather than touch shared state from the
  command. This codebase's loaders already snapshot eagerly, which is why the random-pick
  surface came back clean. — <https://github.com/charmbracelet/bubbletea/issues/909>,
  <https://gist.github.com/cardil/5762ff149f10fe1d76ddadc8b01c9f8d>,
  <https://github.com/charmbracelet/bubbletea/discussions/374>
- **`defer` inside a loop.** This is the one search that found a pattern the tree *does*
  contain, and it is finding L1. A deferred call runs at function return, not at the end of
  the iteration, so each pass pushes onto the defer stack and resources stay held for the
  whole loop; static analyzers (JetBrains `GoDeferInLoop`, Xygeni `go.defer_in_loop`) flag
  it, and the standard remedies are an immediately-invoked closure per iteration or an
  explicit release at each exit. Note that neither analyzer would flag `editFile` as written
  today, since the pattern only becomes a defect once a `continue` can follow the lock —
  which is precisely why L1 is filed as a latent trap rather than a live defect.
  — <https://jetbrains.com/help/inspectopedia/GoDeferInLoop.html>,
  <https://detectors.xygeni.io/xydocs/quality/detectors/go/go.defer_in_loop.html>,
  <https://github.com/golang/go/issues/30130>

## Related-task observations (propose-only)

- Possibly worth a scope note, not a status change: `6g6yjm16dkgt`
  (*design bounded and observable repository lock acquisition*) is the natural home for L1's
  consequence as well as its own subject — an unbounded wait becomes materially worse if the
  holder can ever be an interactive editor. The two are independent fixes; flagging the
  adjacency only so whoever picks up `6g6yjm16dkgt` sees it. No dependency added: L1 is not
  a prerequisite for that task.

## Candidate tasks

<!-- candidate-tasks:v1 · ○ open · ● in-progress · ✔ fixed · → tracked · ◌ deferred · ◌ superseded · ✘ wontfix -->
<!-- Add or replace one row with `tskflwctl audit finding <audit> <code> --candidate "<one line>"`; an empty value removes it. -->

- ○ L1 · open — `tskflwctl task new "Scope editFile guard release to one editor attempt" --epic 21-code-quality-architecture-hardening --tags store,concurrency --tier 4 --priority low --description "Move editFile lock/recheck/write tail into a per-attempt helper so defer unlock() cannot outlive one iteration, and add a regression that fails if a post-lock continue is introduced"`
