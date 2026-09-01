---
schema: 1
id: 6g5xxth593ak
bucket: closed
area: tui-live-reload-watcher-reconciliation-implementation-claude
date: "2026-09-01"
updated_at: "2026-09-01"
---
# Audit: TUI live-reload watcher reconciliation implementation — Claude — 2026-09-01

> Reviewer assignment: Claude. This document is the review brief and the only file the reviewer
> should update.

## Review brief

Perform an independent adversarial implementation review of the uncommitted work for
[keep TUI live reload healthy when entity directories appear](../tasks/6g5rxq1g5mp1-keep-tui-live-reload-healthy-when-entity-directories-appear.md)
in the main worktree, based at commit `0b56488`. Review it against
[ADR-0006](../adrs/0006-adopt-threads-as-task-dags.md), especially the 2026-09-01 TUI-delivery
amendment, and the watcher/session contracts in
[docs/ARCHITECTURE.md](../../docs/ARCHITECTURE.md).

Assume the implementation may be subtly wrong despite green tests. This is event-driven,
concurrent, resource-owning code over platform-dependent fsnotify behavior; a plausible reading is
not sufficient evidence. Try to make it miss a directory transition, report false health, leak a
watch, create two listeners, reload forever, accept an unbounded sentinel, or let an old workspace
affect the new one. Do not reward code volume or comments. Do not manufacture findings either:
settle a concern when a hostile reproduction and the implementation jointly disprove it.

## Review target

The implementation is uncommitted in a deliberately dirty main worktree. Inspect
`git status --short`, every relevant portion of `git diff HEAD`, and untracked files before
classifying scope. Primary files are:

- `internal/tui/watch.go` and `watch_test.go`;
- filesystem messages and reducer handling in `internal/tui/messages.go` and `model.go`;
- watcher construction, replacement, close, and session scoping in `tui.go`, `atlas.go`, and
  `session.go`;
- footer health rendering and tests in `view.go` and `model_test.go`;
- the `core.Layout` port and every `WatchPaths` implementation/consumer;
- README, architecture guidance, ADR-0006, and the implementation task; and
- fsnotify v1.10.1 behavior relevant to Add, Remove, Close, Events, Errors, directory rename/remove,
  and platform backends. Prefer the pinned module source and authoritative upstream documentation
  over memory.

The two planning status changes are expected dogfood state: the adapter-neutral diagnostic task was
completed and this watcher task was started. Ignore unrelated simultaneous work, but do not assume a
changed file is unrelated without inspecting it.

## Intended contract to challenge

- `Layout.WatchPaths()` supplies desired entity leaves, including paths that do not yet exist. The
  TUI must not reconstruct entity directory names or recursively watch a repository.
- Desired paths are absolute/canonicalized through their deepest existing prefix and de-duplicated
  across symlink aliases. An empty path or filesystem/volume root must not become an unbounded watch.
- Each present desired leaf has a direct watch for entity changes. De-duplicated nearest-existing
  parent sentinels observe leaf creation and directory inode replacement. Missing or partially
  attached leaves are recoverable degraded state, not healthy state.
- Reconciliation compares file identity, removes obsolete attachments, adds current targets, and
  verifies the path did not change across Add. Add/Remove/stat failures must fail honestly without
  discarding still-useful coverage.
- Every fsnotify event reconciles and nudges the existing generation debounce. A final quiet-period
  reconciliation closes the race where nested directories appear faster than sentinel attachment.
  Manual reload retries reconciliation without polling.
- Health has three meanings: healthy means every desired direct leaf and required sentinel is
  current; degraded means useful coverage remains but the full desired set is not attached;
  unavailable means no useful coverage remains. Footer state must recover as well as degrade.
- Exactly one listener chain belongs to the active watcher. A recovery must restart a stopped
  listener once, not zero or twice. Reconciliation/reload messages from an old workspace must be
  session-scoped or harmless, and closing a watcher must close every direct and sentinel watch.
- Ordinary save storms, directory creation events, reconciliation, and manual reloads all converge
  through bounded debounce/reload behavior. Watch attachment changes must not cause a busy loop.
- The implementation remains best-effort: the browser works with manual refresh when live reload is
  unavailable. It introduces no polling, recursive root watching, platform-specific lock protocol,
  or Thread-specific directory knowledge.

## Mandatory evidence floor

A verdict of `ready` is not credible unless the report contains all of the following. If the host
cannot execute a platform case, label it explicitly as source-inspected or unverified; do not turn
one Darwin/Linux run into a cross-platform claim.

1. A repository-wide consumer inventory for `newWatcher`, `waitForFS`, `reconcileWatcher`,
   `debounceTick`, `watchHealth`, `watchOff`, `watchDegraded`, `WatchPaths`, watcher Close,
   and every message carrying watcher state.
2. At least one real temporary-directory reproduction for each of:
   - missing leaf at startup, creation, then a second file change inside it;
   - rapid nested parent creation followed immediately by an entity file;
   - remove and later recreate;
   - atomic rename/replacement with a write into the replacement;
   - duplicate lexical/symlink aliases;
   - degraded attachment followed by recovery; and
   - workspace/watcher replacement while an old listener or event is outstanding.
3. A reload-storm probe that records event/reload or generation behavior and attachment counts across
   a burst. “The debounce code looks bounded” is not enough.
4. Repeated focused race execution, not a single cached run. Run the watcher/session tests with
   `-race -count` high enough to exercise scheduling variance, plus the full repository race suite.
5. Mutation probes, restored afterward, for at least these five failure modes:
   - omit the parent sentinel;
   - replace inode comparison with path-only comparison;
   - omit the quiet-period reconciliation;
   - ignore or mis-map degraded/unavailable health in the reducer/footer; and
   - reissue the listener unconditionally during unavailable/recovery handling.
   State which focused test failed for each mutation. If a mutation survives, file a coverage
   finding even when current production behavior is correct.
6. Watch-resource evidence before and after reconciliation, workspace replacement, and Close. Use
   the fsnotify watch list/backend state where available; do not infer descriptor cleanup solely
   from a nil pointer.
7. An event-order analysis for Bubble Tea `tea.Batch`: identify every path that can issue a
   listener, debounce tick, reload, or explicit reconcile command, then show why arbitrary result
   ordering cannot produce duplicate listeners, stale health, or a stopped recoverable watcher.
8. Exact validation commands and results for focused tests, full `go test -race ./...`, static
   analysis/lint, docs drift, module tidiness, planning lint, and `git diff --check`.

A shallow report that merely restates acceptance criteria, cites existing test names without running
hostile cases, or says “all tests pass” without the mutation ledger does not satisfy this audit.

## Required hostile angles

1. **Path normalization and bounds.** Exercise relative paths, repeated separators, `..`, empty
   strings, duplicates, existing and missing leaves below symlinked ancestors, a symlink retarget,
   filesystem roots, volume roots/UNC forms where relevant, a desired path that is a regular file,
   and paths whose nearest ancestors are inaccessible. Determine whether canonicalization can merge
   distinct desired paths, preserve a stale symlink target, or climb farther than an appropriate
   planning boundary.
2. **Sentinel topology.** Re-derive the required watch set for one leaf, five sibling entity leaves,
   nested missing leaves, overlapping desired ancestor/descendant paths, and aliases. Confirm
   sentinel de-duplication, bounded watch count, and whether always watching a parent observes the
   replacement operations claimed on each supported backend.
3. **Reconciliation races.** Interleave stat, Remove, Add, post-Add stat, rename, recreate, and close.
   Attack ABA/inode-reuse assumptions, ignored Remove failures, Add reporting success on an obsolete
   backend registration, replacement between the two stat calls, and a path disappearing after it
   is recorded in `attached`.
4. **Creation timing gaps.** Create a nested tree and file as one fast operation. Force the listener
   to process the first ancestor event before deeper directories exist, after all exist, and around
   the debounce boundary. Prove the eventual leaf is directly watched and the file becomes visible
   even if no later parent event is delivered.
5. **Health truthfulness.** Independently fail direct-leaf Add, sentinel Add, stat, and every
   attachment. Check initial construction, fs events, quiet reconciliation, and manual reload.
   Healthy must require the entire required set; degraded must retain an actual useful listener;
   unavailable must stop promising live reload. Recovery must clear both state fields and footer
   text.
6. **Listener cardinality.** Enumerate startup, normal event, event-to-unavailable, debounce recovery,
   explicit/manual recovery, atlas activation, stale workspace-open result, and shutdown. Use a
   hostile schedule to look for two blocked `waitForFS` commands on one watcher, no listener after
   recovery, or a listener restarted for a closed/stale watcher.
7. **Session isolation and cleanup.** Switch workspaces while old wait, debounce, reload, and
   reconciliation commands are outstanding. Confirm session wrapping drops their model effects,
   the old watcher still closes, no old event reloads the new service, and cached sessions retain no
   watcher/backend resources.
8. **Event/error channels and shutdown.** Close during blocked wait and during reconcile; inject or
   trigger fsnotify Errors; drain channel closure; run under the race detector. Check nil-message
   behavior cannot strand state or cause Bubble Tea to spin.
9. **Reload boundedness.** Burst create/write/chmod/rename events and count resulting reloads,
   reconciliation attempts, ticks, and watch-set mutations. Manual `r` now adds reconciliation;
   ensure reload → reconcile → message does not recurse into another reload and stale debounce ticks
   cannot overwrite newer health.
10. **UI honesty.** Inspect atlas, dashboard, list, focused detail, narrow terminals, flash/error
    overrides, and initial startup. Verify degraded/off labels are not silently truncated in ordinary
    supported widths, recover correctly, and do not imply that no manual refresh exists.
11. **Port and scope integrity.** Confirm the TUI learns no entity names, no repository root is
    recursively watched, no polling timer exists beyond event debounce, and local watcher mechanics
    do not leak into core projection/read ports. Challenge whether `Layout`'s updated comment and
    every implementation agree that paths may be absent.
12. **Documentation and planning.** Compare implementation, tests, README, architecture guide,
    ADR-0006, task checkboxes/progress, and Thread dogfood state. Flag claims stronger than runtime
    evidence, especially “atomic replacement,” “every footer,” “every watch,” platform portability,
    and “no reload storms.”

## Platform discipline

Record `GOOS`, filesystem type where discoverable, and the fsnotify backend actually exercised.
Inspect the pinned fsnotify implementation for at least Darwin/kqueue, Linux/inotify, and
Windows/ReadDirectoryChangesW semantics relevant to this change. Findings should distinguish:

- a demonstrated bug on the current runtime;
- a source-supported portability defect;
- an unverified risk that needs a target-platform test; and
- an implementation choice that is acceptable within an explicitly documented support boundary.

Do not recommend polling or recursive repository watching merely to erase all platform uncertainty.
Prefer the smallest correction or focused coverage that preserves the accepted event-driven design.

## Proportionality and finding quality

A finding must identify an observable violated contract, unsafe resource/concurrency behavior, a
surviving mutation, or a documentation overclaim. Give it a stable severity code, exact file/line
evidence, the hostile reproduction or reasoning chain, impact, and the minimum viable correction.
Do not report style preferences, hypothetical features, or broad rewrites as findings. If a concern
is real but intentionally deferred, still mark it open; the implementation owner will choose the
lifecycle status and destination.

## Validation and restoration

Run proportionate validation including focused repeated race tests and the full race suite. You may
create scratch files outside the repository and temporary mutation probes inside it, but restore
every probe and generated artifact. Do not install dependencies, commit, push, edit implementation
permanently, create follow-up tasks, close this audit, or edit the other reviewer's audit. At finish,
the worktree must differ from its starting state only by your edits to this assigned audit file.
Verify that claim with `git status --short` and `git diff`.

## Deliverable

Preserve this brief and replace the reviewer-report placeholder below with:

- executive verdict: `ready`, `ready with tracked follow-ups`, or `not ready`;
- reviewed branch/base/worktree state, runtime platform/backend, and exact validation results;
- a compact architecture/event-flow re-derivation rather than an acceptance-criteria paraphrase;
- findings grouped by severity, each with stable code and `**Status:** open`;
- an acceptance-criteria traceability table;
- a mandatory hostile-evidence ledger with one row per required reproduction and mutation probe,
  including the observed result and the test that would catch regression;
- platform conclusions separated by demonstrated, source-inspected, and unverified claims; and
- explicitly settled concerns, including why they are settled.

If there are no findings, say so plainly, but the evidence ledger and mutation results are still
required. Do not pre-resolve findings; the implementation owner will triage them with
`tskflwctl audit finding`.

## Reviewer report

### Executive verdict

**`ready with tracked follow-ups`.**

The event-driven design holds up under hostile probing. I could not produce a duplicate listener, a
leaked watch or descriptor, a reload storm, a busy retry loop, a stale-workspace health bleed, or an
unbounded sentinel — each was attacked with a real reproduction and each survived. Health
degradation is honest in every attachment-failure mode I could induce, and recovery clears it.

Two classes of problem remain. One demonstrated correctness defect: a planning path behind a symlink
that is **retargeted mid-session** leaves the watcher attached to the old target while reporting
`healthy` and delivering no events (WR1) — the precise "preserve a stale symlink target" case the
brief names. And five separate invariants that production gets right but **no test pins**: five
mutation probes survived the complete 262-test tracked `internal/tui` suite, including the
quiet-period reconciliation that ADR-0006 and `docs/ARCHITECTURE.md` explicitly credit with closing
the nested-creation race — which I measured to be load-bearing in 40/40 runs (WR2, WR3, WR4).

None of this warrants `not ready`: WR1 is not a regression (the base implementation resolved
symlinks inside `fsnotify.Add` and had the same blind spot, with no health reporting at all), its
trigger is narrow, and its impact is a stale footer rather than data loss.

### Reviewed state, runtime, and platform

| | |
|---|---|
| Base commit | `0b56488` (`main`) |
| Worktree | deliberately dirty main worktree; change uncommitted |
| Scope classified from | `git status --short`, full `git diff HEAD`, untracked files |
| Reviewed files | `internal/tui/{watch,model,view,session,tui,messages}.go`, `internal/core/store.go`, `internal/tui/{watch,model}_test.go`, `README.md`, `docs/ARCHITECTURE.md`, ADR-0006, both planning tasks |
| `GOOS/GOARCH` | `darwin/arm64` |
| Go | `go1.26.6` |
| fsnotify | `v1.10.1` (pinned; read from the module cache, not from memory) |
| Backend exercised | **kqueue** (`backend_kqueue.go`) — the only backend executed here |
| Filesystem | APFS (`/dev/disk3s1s1 … apfs, sealed, local, journaled`); `t.TempDir()` resolves under `/private/var/folders/…`, which is why every probe canonicalizes before comparing paths |
| `ulimit -n` | 1048576 |

**Worktree contamination note.** An untracked file `internal/tui/audit_hostile_test.go` (13 KB, 504
lines, `TestHostile_*`, mtime 18:00) appeared during this review. It is **not mine** — it is the
other reviewer's in-flight probe file — so I left it untouched. It affects two things I report
honestly below: it is the sole source of all 5 `golangci-lint` issues, and while its suite was
running concurrently it exhausted enough kqueue resources to make timing-sensitive watcher tests
fail. Both are isolated and re-measured on a quiet machine.

**A foreign mutation is live in the worktree as I finish (not mine).** At 19:07, after my own
mutation ledger was complete and restored (`shasum` verified against the pre-mutation backup),
`internal/tui/watch.go` was modified on disk by another session to `health = w.watchHealth()` inside
`debounceTick` — byte-for-byte the mandated mutation M3. It is presumably the other reviewer's
in-flight probe, so I deliberately did **not** revert it: reverting another agent's active mutation
would corrupt its measurement, and this brief forbids me editing the implementation. **The
implementation owner must confirm `internal/tui/watch.go:269` reads `health = w.reconcile()` before
committing** — `git diff internal/tui/watch.go` shows the one-line deviation. No result in this
report is affected: every hostile probe ran before 18:04 against the unmutated file, and M3 is
documented below as surviving the entire suite, so any validation run that overlapped it reports
`ok` either way.

### Validation — exact commands and results

| Command | Result |
|---|---|
| `go build ./...` | pass |
| `go vet ./internal/tui/ ./internal/core/ ./internal/store/` | clean (exit 0) |
| `go test ./internal/tui/ -run 'Watcher\|FsEvent\|Watch' -race -count=1 -v` | 10/10 pass |
| `go test ./internal/tui/ -run '<262 tracked test names>' -count=1` | ok 2.211s |
| `go test ./internal/tui/ -run '<262 tracked>' -race -count=3` | ok 16.593s |
| `go test ./internal/tui/ -run '<262 tracked>' -race -count=10` | ok 51.715s |
| `go test ./internal/tui/ -run 'TestWatcher\|TestNewWatcher\|TestModel_FsEvent\|TestActivateWorkspace' -race -count=50` | ok 3.693s |
| `go test ./internal/tui/ -run '…\|TestAtlasDropsStale\|TestSessionScoping' -race -count=50` | ok 6.139s (quiet machine) |
| `go test ./internal/tui/ -run 'TestWatcherReattachesAtomicallyReplacedDirectory' -count=50` (unmutated control) | 0 failures |
| `go test -race $(go list ./... \| grep -v '/internal/tui$')` | all 23 packages ok |
| `go test -race ./...` (first run, before contamination) | all ok, exit 0 |
| `golangci-lint run ./...` | 5 issues, **all in the other reviewer's `audit_hostile_test.go`** (4 errcheck, 1 staticcheck); **zero in any reviewed file** |
| `go mod tidy -diff` | clean (exit 0) |
| `just docs` + `git diff --exit-code docs/cli` | no drift |
| `./bin/tskflwctl lint` | `✔ all planning entities and dependency links pass lint` |
| `git diff --check` | clean (exit 0) |

**One caveat reported rather than hidden.** A `go test -race ./...` run launched *while the other
reviewer's suite was running* produced four failures in `watch_test.go`
(`TestWatcherReportsTransientAddFailureAndExplicitReconcileRecovers` → `health = 2, want degraded`;
`TestWatcherReattachesRemovedAndRecreatedDirectory` → `health = 2, want 1`;
`TestWatcherNormalizesSymlinkAliasesBeforeAttaching` → `attached watch count = 1, want 2`;
`watch_test.go:96 timed out waiting for filesystem event`). `health = 2` is `watchUnavailable`, i.e.
**every** `fsnotify.Add` failed — consistent with kqueue/descriptor exhaustion under parallel load
(on Darwin one directory watch opens an fd per file in it; see the resource measurements below).
Every one of these passes on a quiet machine, including at `-count=50 -race`. I therefore do **not**
report an implementation defect here, but I do note the test-hermeticity consequence in WR4: the
shipped watcher tests assume `Add` succeeds and fail rather than skip when the platform refuses.

### Architecture and event-flow re-derivation

`Layout.WatchPaths()` (`internal/store/fsstore.go:93`, the only production implementation) returns
five joined leaf paths without checking existence — consistent with the port comment's new promise
in `internal/core/store.go:303-308`. `newWatcher` (`watch.go:50`) turns them into a **desired set**
that outlives the filesystem's current shape:

1. `normalizeWatchPaths` (`watch.go:157`) drops `""`, `Abs`+`Clean`s, canonicalizes, and rejects
   anything whose canonical form is its own parent — the root/volume-root guard.
2. `canonicalWatchPath` (`watch.go:184`) walks up to the deepest **existing** prefix, `EvalSymlinks`
   that, then re-appends the missing suffix. This is what lets alias de-duplication work for a leaf
   that does not exist yet.
3. `reconcile` (`watch.go:82`) computes `required` = {each existing desired leaf} ∪ {each desired
   leaf's nearest existing ancestor}, then converges the fsnotify set: drop attachments that are no
   longer required *or whose inode changed* (`os.SameFile`, `watch.go:109`), add missing targets with
   a **pre-stat / Add / post-stat identity re-check** (`watch.go:121-134`), and grade the result —
   `unavailable` iff nothing is attached, `healthy` iff `complete` (every desired leaf existed and
   every required path is attached), else `degraded`.

The reducer wires three entry points into that one function. `waitForFS` (`watch.go:235`) reconciles
inside the listener command and hands the health back on `fsEventMsg`; `debounceTick`
(`watch.go:265`) reconciles again after the 200 ms quiet period; `reconcileWatcher` (`watch.go:255`)
reconciles on demand behind manual `r`. `model.go:426-465` maps health onto `watchOff`/`watchDegraded`
and decides listener continuation.

The load-bearing subtlety is the **listener-continuation rule**. A listener blocks on
`w.fsw.Events`; if nothing is attached no event can ever arrive, so `fsEventMsg` deliberately issues
no successor when health is `unavailable` (`model.go:448-452`) and the `wasOff && !m.watchOff` guards
in `debounceMsg` and `watcherReconciledMsg` (`model.go:435`, `model.go:462`) restart it **exactly
once** on the recovery edge. That is the invariant that keeps listener cardinality at one.

The convergence argument for nested creation is layered: each newly created level fires on the
currently attached sentinel, which promotes the sentinel one level deeper; when the tree is created
faster than that ladder can climb, the quiet-period tick reconciles against the *finished* tree and
attaches the leaf directly. Both halves are needed — see WR2.

### Evidence floor #1 — repository-wide consumer inventory

`grep -rn --include='*.go' -E "waitForFS|newWatcher|reconcileWatcher|debounceTick|watchHealth|watchOff|watchDegraded|watchUnavailable|watchHealthy|WatchPaths|closeWatcher|activateWorkspace|fsEventMsg|debounceMsg|watcherReconciledMsg" .`

| Symbol | Production call sites | Notes |
|---|---|---|
| `newWatcher` | `tui.go:47` (startup), `atlas.go:153` (workspace open) | both tolerate `err` by leaving `m.watch == nil` |
| `waitForFS` | `model.go:210` (Init), `model.go:450` (fsEvent, guarded), `model.go:436` / `model.go:463` (recovery edges), `session.go:172` (activation) | 5 sites; all ≤1 per transition |
| `reconcileWatcher` | `model.go:429` (reloadMsg only) | nil-safe; returns `nil` cmd for a nil watcher — see WR6 |
| `debounceTick` | `model.go:452` (fsEvent only) | the sole `tea.Tick` in `internal/tui` |
| `watchHealth()` | `tui.go:50`, `session.go:120` | nil receiver → `watchUnavailable` |
| `watchOff` | `model.go:107` (field), `tui.go:45,49`, `session.go:121`, `view.go:354`, reducer `model.go:432-463` | |
| `watchDegraded` | `model.go:108` (field), `tui.go:50`, `session.go:122`, `view.go:356`, reducer | |
| `watcher.close()` | `watch.go:64` (failed construction), `session.go:189` (`closeWatcher`) via `session.go:172` (activation), `atlas.go:178` + `model.go:265` (stale open results) | 4 paths; all funnel through `closeWatcher`/`close` |
| `WatchPaths()` | `store/fsstore.go:93` (only production impl); fakes in `cli/ui_test.go:19`, `core/workspace_test.go:25`, `tui/atlas_test.go:80` | all agree paths may be absent |
| `fsEventMsg` | produced `watch.go:245,250`; consumed `model.go:440` | carries `health` |
| `debounceMsg` | produced `watch.go:271`; consumed `model.go:454` | carries `gen` + `health` |
| `watcherReconciledMsg` | produced `watch.go:259`; consumed `model.go:431` | carries `health` |

No watcher symbol appears in `internal/core`, `internal/store`, or `internal/domain`.

### Findings

#### WR1. medium — a retargeted symlink leaves the watcher on the old target while reporting `healthy` · **Status:** fixed

**File:** `internal/tui/watch.go:51` (`desired := normalizeWatchPaths(paths)`), `watch.go:184`
(`canonicalWatchPath`), `watch.go:82-151` (`reconcile` never re-canonicalizes).

`w.desired` is canonicalized **once**, at construction, and is thereafter treated as ground truth.
`reconcile` re-stats those resolved paths but never re-resolves the configured ones, so a symlink
anywhere in a planning path that is repointed while the TUI is running is invisible to the watcher —
while `store.FS` keeps reading through the *unresolved* path and therefore follows the new target.

**Hostile reproduction (probe P5b, single persistent listener so no goroutine can steal an event):**
`<root>/current -> <root>/a`, desired `<root>/current/tasks`. Retarget `current -> b`, drain, then
write `<root>/current/tasks/6g5w000000q1-x.md`:

```
drained 0 retarget events; health=0                     (0 = watchHealthy)
reconcile after retarget = 0
watchlist = [<root>/a, <root>/a/tasks]
FALSE-HEALTHY CONFIRMED: health=0 yet a write to configured path
  <root>/current/tasks/6g5w000000q1-x.md (real <root>/b/tasks/…) delivered NO event in 2s
```

Probe P4 shows the mechanism directly: after the retarget, `filepath.EvalSymlinks(desired)` resolves
to `<root>/b/tasks` while `w.desired[0]` is still `<root>/a/tasks`.

**Impact.** The footer promises live reload while the watcher observes a directory the store no
longer reads: silent staleness with no degraded signal — the exact failure class this task set out to
remove, in the one deployment shape (`current -> release-N`, `~/planning -> <synced dir>`) where a
planning root is likely to sit behind a symlink at all.

**Contract violated.** Brief: "Health has three meanings: healthy means every desired direct leaf and
required sentinel is current"; and the named hostile angle "a symlink retarget … Determine whether
canonicalization can merge distinct desired paths, **preserve a stale symlink target**".

**Not a regression.** The base implementation (`git show HEAD:internal/tui/watch.go`) passed raw
paths to `fsnotify.Add`, and kqueue's `addWatch` follows symlinks itself (`backend_kqueue.go:371-394`),
so it landed on the same stale target — with no health reporting at all. This is unfixed edge, not
new breakage.

**Minimum viable correction.** Keep the raw `Layout.WatchPaths()` strings on the watcher and re-derive
`desired` at the top of `reconcile` (canonicalization is already cheap and idempotent, and `reconcile`
is the only writer of `attached`); a retarget then presents as an ordinary attachment change. A
smaller variant: store `raw → canonical` pairs and mark the watcher `degraded` when any raw path's
current canonical form differs from the stored one.

**Resolution:** False-healthy symlink retargeting is removed. Reconciliation now
re-resolves raw Layout paths, watches layout-controlled symlink parents where
useful, and conservatively reports symlink-backed layouts degraded because
same-name retarget delivery is not portable. Manual refresh deterministically
moves attachments to the current target; regression tests cover retargeting and
subsequent writes.

#### WR2. medium — the quiet-period reconciliation is load-bearing and completely untested · **Status:** fixed

**File:** `internal/tui/watch.go:262-269` (`debounceTick`), `internal/tui/model.go:452`.

Mutation **M3** — `debounceTick` returns `w.watchHealth()` (last known) instead of `w.reconcile()` —
**survives the entire 262-test tracked `internal/tui` suite at `-count=3`**. No test drives
`debounceTick` against a real watcher; `TestModel_FsEventDebounces` and
`TestModel_FsEventSurfacesAndRecoversWatcherHealth` both synthesize `debounceMsg` values directly, so
the reconcile inside the tick is never executed under test.

It is not dead code. Probe P10 ran the nested-creation scenario 40 times with **event-driven
reconciliation only** (exactly M3's world) — `MkdirAll(planning/a/b/c/threads)` under desired leaves
`{planning/tasks, planning/a/b/c/threads}`:

```
event-only reconciliation left the deepest leaf UNATTACHED in 40/40 runs;
of those, a subsequent write inside the leaf was MISSED 37/40 times
```

The ladder cannot keep up: levels below the current sentinel are created before the sentinel is
promoted, so no further event ever arrives and the leaf is never attached. The tick is what closes
it — probe P6b confirms the production path converges (`events=1`, health after the quiet reconcile
`watchHealthy`, leaf in `WatchList`).

**Impact.** ADR-0006 and `docs/ARCHITECTURE.md` both stake the nested-creation guarantee on this
mechanism, and a refactor that drops it stays green. Given P10's 37/40 miss rate, that regression
would reintroduce precisely the bug this task exists to fix.

**Minimum viable correction.** One focused test that calls `debounceTick(w, gen)()` against a real
watcher after a deep `MkdirAll` and asserts the returned `debounceMsg.health == watchHealthy` and
that the leaf is present in `w.attached`.

**Resolution:** A real-watcher test now creates the remaining nested tree after
the first sentinel event, invokes the production debounceTick callback, and
asserts healthy convergence plus direct leaf attachment.

#### WR3. medium — the listener-cardinality guards are asserted nowhere · **Status:** fixed

**File:** `internal/tui/model.go:435` (`watcherReconciledMsg`), `model.go:448-452` (`fsEventMsg`),
`model.go:462` (`debounceMsg`).

Mutation **M5** — drop all three `wasOff` guards and reissue `waitForFS` unconditionally — **survives
the full 262-test tracked suite at `-count=3`**.

Production behavior is correct: probe P11 enumerated every reducer transition that can issue a
listener and each issues at most one (see the ledger). Under M5, every `watcherReconciledMsg` — one
per manual `r` — issues another listener on the same `Events` channel, so *N* presses of `r` strand
*N* blocked goroutines, and each delivered event then wakes only one of them.

**Minimum viable correction.** A reducer test asserting the command tree returned for
`watcherReconciledMsg{healthy}` and `debounceMsg{healthy}` contains a listener **only** when
`m.watchOff` was true beforehand.

**Resolution:** Reducer tests now pin listener cardinality: already-active
explicit and quiet-period reconciliation cannot start another listener, while
the unavailable-to-active edge does.

#### WR4. low — three more watcher invariants have no test, and the replacement test pins its own subject only 10% of the time · **Status:** fixed

Four separate coverage gaps, one remedy (a focused test file):

| Probe | Mutation | Result vs. 262 tracked tests |
|---|---|---|
| **M2b** | drop the post-`Add` `os.SameFile(before, after)` re-check (`watch.go:133`) — the brief's "verify the path did not change across Add" | **survives** (`-count=3`, also `-race`) |
| **M6** | `activateWorkspace` ignores the incoming watcher's health (`session.go:120-122`) | **survives** (`-count=3`) |
| **M7** | `newWatcher` accepts a zero-attachment watcher (drop the `watch.go:63-66` guard) | **survives** (`-count=3`) — `TestNewWatcherRejectsEmptyAndFilesystemRootPaths` only exercises the earlier `len(desired) == 0` branch |
| **M2a** | drop the removal-loop `os.SameFile` (`watch.go:109`) | caught, but `TestWatcherReattachesAtomicallyReplacedDirectory` — the test *named* for inode replacement — detects it in only **5/50** runs (0/50 false positives unmutated). The deterministic detector is `TestWatcherReportsTransientAddFailureAndExplicitReconcileRecovers`, via the injected `addPath` seam. |

M2a's low detection rate has a mechanical cause: the test's two-step `Rename`/`Rename` usually lets a
reconcile land in the window where the path does not exist, which drops the attachment for the
*right* reason and masks the mutation. The `watcherHealthAfter` helper (`watch_test.go:33-41`)
compounds this — when the event-carried health is not the expected value it calls `w.reconcile()`
again and re-checks, so these tests never actually assert that the **event-carried** health is
correct, which is the value the reducer uses to decide whether to keep listening.

Related, from the contamination caveat above: the shipped watcher tests assume `fsnotify.Add`
succeeds and fail (rather than skip) when the platform refuses — a CI flake risk on a loaded machine,
especially on Darwin where a directory watch costs an fd per file in it.

**Minimum viable correction.** Assert the event-carried health directly in at least one test; add a
`fsw.WatchList()`/inode assertion after a *single-step* replacement to pin M2a deterministically; add
one assertion each for the `activateWorkspace` health derivation and the `newWatcher` unavailable
guard.

**Resolution:** Focused tests now pin post-Add identity verification, incoming
workspace health, recoverable zero-attachment construction, event-carried
atomic-replacement health and inode identity, and complete watch cleanup.
Resource exhaustion remains an honestly reported unavailable/degraded runtime
state rather than a test skip.

#### WR5. low — "every footer surface" is an overclaim · **Status:** fixed

**File:** `planning/tasks/6g5rxq1g5mp1-…md`, Implementation progress: "Events propagate reconciled
health into **every footer surface**". `internal/tui/view.go:363-380`.

Three `footer()` branches return before `withWatchHealth` is applied: the command palette
(`view.go:364`, `m.cmd.active`), the flash override (`view.go:372`, `m.flash != ""`), and the detail
find bar (`view.go:378`). Measured with `m.watchOff = true` (probe P14):

```
flash footer while watcher off:   "✔ moved 6g5w000000q1 to in-progress"
command footer while watcher off: ":   overview · atlas · config · tasks · working · …"
```

The code is defensible — these are deliberate transient overrides and the pattern predates this
change — and the change is a genuine improvement (the atlas and dashboard footers previously showed
*no* watcher state at all, since `watchOff` was applied only on the list footer). It is the sentence
that is wrong.

**Minimum viable correction.** Reword to "every persistent footer surface (list, detail, dashboard,
atlas); transient command/flash/find overrides still take precedence."

**Resolution:** The task documentation now limits the claim to persistent list,
detail, dashboard, and atlas footers and explicitly preserves transient command,
flash, and find precedence.

#### WR6. low — a watcher that fails at construction can never recover, contradicting the documented manual-refresh fallback · **Status:** fixed

**File:** `internal/tui/tui.go:45-51`, `internal/tui/atlas.go:153`, `internal/tui/session.go:120-122`,
`internal/tui/watch.go:251-256`.

When `newWatcher` returns an error, `m.watch` stays `nil` and `m.watchOff` is `true`. `reloadMsg`
then calls `reconcileWatcher(nil)`, which returns a `nil` command (`watch.go:252-254`), so `r`
produces no `watcherReconciledMsg` and there is nothing to retry against. Probe P20:

```
manual reload with a nil watcher: watcherReconciledMsg produced = false;
reconcileWatcher(nil) = nil; waitForFS(nil) = nil
=> once newWatcher fails at startup/activation, live reload is off for the life of the session
```

Reachable: probe P3 built a tree whose every candidate ancestor was `chmod 000` and got
`newWatcher err = no watchable directories: live reload unavailable`. On Linux the realistic trigger
is `inotify_add_watch` returning `ENOSPC` at `max_user_watches` (`backend_inotify.go:275`); the
contaminated-run failures above show the Darwin equivalent under descriptor pressure.

**Impact.** After the operator fixes permissions, raises the watch limit, or creates the planning
tree, live reload stays off until the TUI is restarted. The footer is honest, but the brief's
"Manual reload retries reconciliation without polling" does not hold in this state.

**Minimum viable correction.** On construction failure retain a watcher holding the desired paths and
zero attachments (health `unavailable`, `watchOff` still true) instead of collapsing to `nil`, so `r`
can drive `reconcile` again. Note this is *not* mutation M7, which would also make the watcher look
constructible; the point is to keep the object addressable while keeping the health honest.

**Resolution:** A valid watcher whose Add calls all fail is retained with
unavailable health. Workspace startup/activation keeps that object without
launching a listener, and manual reconciliation can attach paths and restart
exactly one listener after the transient failure clears.

### Acceptance-criteria traceability

| # | Criterion | Verdict | Evidence |
|---|---|---|---|
| 1 | Missing `threads/` at startup → creating a Thread reloads without restart; a second change inside is also observed | **met** | P6c: creation event delivered (`health=healthy`), then writes #1 and #2 inside the new leaf each observed. P6b: nested variant + a third write in the *original* leaf still observed. Shipped `TestWatcherRecoversMissingDesiredLeafAndObservesItsFiles`. |
| 2 | Remove/recreate or atomic replacement reattaches and is not bound to the old inode | **met** | Shipped `TestWatcherReattachesRemovedAndRecreatedDirectory` / `…AtomicallyReplacedDirectory`. P7b: 20 replace+reconcile cycles → `WatchList` constant at 6, health `healthy`, no fd growth. M2a confirms the inode check is load-bearing (see WR4 for its weak pinning). |
| 3 | Partial attachment failure is `degraded`, not unqualified healthy; recovery clears it | **met** | P15: leaf-`Add` fails → `degraded`, 2 useful attachments; sentinel-`Add` fails → `degraded`; every `Add` fails → `unavailable`, 0 attachments; leaf removed mid-life → `degraded`; recreate → `healthy`. P13: manual `r` clears degradation end-to-end. P14: footer text at 30–200 columns. |
| 4 | Normalization prevents duplicate watches for symlink/alias spellings; workspace switching closes every watch of the previous session | **met, qualified by WR1** | P1: `//`, trailing separator, `..`, and a symlinked ancestor all collapse to one canonical entry — including for a **missing** leaf (`alias/threads` → `planning/threads`). P2: 5 sibling leaves → exactly 6 watches. P7b/P17: `WatchList()` empty after `Close`, open files back to the exact pre-watcher baseline. **Qualifier:** de-duplication is correct at construction; a mid-session retarget is not tracked (WR1). |
| 5 | Event coalescing stays bounded; no reload storms or busy retry loops | **met** | P8: 200×(write+2×chmod+2×rename) → **570 fs events → 1 reload**, 1 listener, attachments constant at 2. P19: 60×(mkdir+write+rmdir) → **164 events → 1 reload**, `WatchList` constant at 2, footer correctly `degraded` at rest. `reconcile` performs no filesystem mutation, so it cannot feed itself. |

### Mandatory hostile-evidence ledger

Probes ran in-package (`internal/tui`) as temporary files, all removed; restoration verified below.

| # | Required reproduction | Probe | Observed result | Regression guard |
|---|---|---|---|---|
| 1 | Missing leaf at startup → creation → second change inside | P6c | creation event `health=healthy`; writes #1 and #2 both observed | `TestWatcherRecoversMissingDesiredLeafAndObservesItsFiles` |
| 2 | Rapid nested parent creation + immediate entity file | P6b, P6 | 1 event; quiet reconcile → `healthy`, 4 attachments; second write in new leaf **and** a write in the original leaf both observed; identical under reversed `tea.Batch` ordering | `TestWatcherRecoversNestedDesiredLeafWhenParentTreeAppears` (weak — see WR2) |
| 3 | Remove and later recreate | P15 (`leaf stat fails mid-life`) | remove → `degraded` (2 attached); recreate → `healthy` | `TestWatcherReattachesRemovedAndRecreatedDirectory` |
| 4 | Atomic rename/replacement + write into the replacement | P7b, shipped test | 20 cycles: `WatchList` 6→6, fds 119→119, health `healthy` | `TestWatcherReattachesAtomicallyReplacedDirectory` (10% detection — WR4) |
| 5 | Duplicate lexical/symlink aliases | P1, P2 | 15 hostile spellings → 6 canonical entries; `""`,`/`,`/..`,`///`, volume root all rejected; mixed ancestor/descendant/alias set → 4 desired, 4 attached | `TestWatcherNormalizesSymlinkAliasesBeforeAttaching` |
| 6 | Degraded attachment then recovery | P15, P13 | all four failure modes → `degraded` with ≥1 useful attachment; manual `r` clears `watchOff`/`watchDegraded` and the footer | `TestWatcherReportsTransientAddFailureAndExplicitReconcileRecovers`, `TestModel_FsEventSurfacesAndRecoversWatcherHealth` |
| 7 | Workspace/watcher replacement with an outstanding listener/event | P9, P17 | stale tick arrives as `sessionMsg{gen:0, msg:debounceMsg{gen:1, health:unavailable}}` and is dropped by `model.go:262`; new session health untouched. Old listener returns `nil`; old-tree writes inert; new watcher unaffected | `TestActivateWorkspaceReplacesAndClosesTheOnlyActiveWatcher`, `TestAtlasDropsStaleWorkspaceResultsAndOldSessionMessages` |
| 3′ | **Reload-storm probe with counts** | P8, P19 | see AC-5 row: 570→1 and 164→1, `maxConcurrentListeners = 1`, watch set constant | — |
| 4′ | **Repeated focused race execution** | — | focused watcher set `-race -count=50` ok 3.7s; combined set `-count=50` ok 6.1s; 262 tracked tests `-race -count=10` ok 51.7s; full repo race ok | — |
| 6′ | **Watch-resource evidence (backend state, not nil pointers)** | P7b, P2, P17 | 5 dirs × 20 files: `lsof` 10 → **119** after `newWatcher` (`WatchList`=6); **119** after 20 replace+reconcile cycles; **10** — exactly baseline — after `close()`, `WatchList()` empty. P17: post-close `attached` is nil and `fsw.Add` refused | — |
| 8′ | **Errors channel / shutdown / nil messages** | P16, P12 | a value on `fsw.Errors` wakes the listener → `fsEventMsg{health}` (reconcile + debounce). `close()` during a blocked wait → listener returns `nil`, chain terminates; post-close `reconcile()` = `unavailable`; double `close()` = nil; 16-goroutine close/reconcile/health storm clean under `-race` | — |

The `-race` warning P16 raised was my probe writing into `fsw.Errors`, a channel only fsnotify's
backend may write; it is a probe artifact, not implementation behavior. The functional assertion
completed before it.

#### Mutation ledger (evidence floor #5)

Applied to the working tree, run, then restored; every mutation was re-run against the **full
262-test tracked `internal/tui` suite** so no result depends on a narrow `-run` filter.

| ID | Mutation | Required by brief | Outcome | Which test failed |
|---|---|---|---|---|
| **M1** | omit the parent sentinel entirely (`watch.go:98-103`) | ✔ #1 | **caught** | 5 tests: `…RecoversMissingDesiredLeafAndObservesItsFiles`, `…RecoversNestedDesiredLeafWhenParentTreeAppears`, `…ReportsTransientAddFailureAndExplicitReconcileRecovers`, `…ReattachesRemovedAndRecreatedDirectory`, `…NormalizesSymlinkAliasesBeforeAttaching` |
| **M2a** | path-only comparison in the removal loop (drop `os.SameFile`, `watch.go:109`) | ✔ #2 | **caught, weakly** | `…ReportsTransientAddFailureAndExplicitReconcileRecovers` (deterministic); `…ReattachesAtomicallyReplacedDirectory` only **5/50** runs → **WR4** |
| **M2b** | drop the post-`Add` identity verification (`watch.go:133`) | ✔ #2 | **SURVIVES** | none → **WR4** |
| **M3** | omit the quiet-period reconciliation (`watch.go:263-267`) | ✔ #3 | **SURVIVES** | none → **WR2** (load-bearing: 40/40, 37/40 missed writes) |
| **M4** | degraded renders as healthy in the footer (`view.go:356`) | ✔ #4 | **caught** | `TestModel_FsEventSurfacesAndRecoversWatcherHealth` |
| **M4b** | reducer drops reconciled health from `fsEventMsg` (`model.go:445-446`) | ✔ #4 | **caught** | `TestModel_FsEventSurfacesAndRecoversWatcherHealth` |
| **M5** | reissue the listener unconditionally on all three edges (`model.go:435,448,462`) | ✔ #5 | **SURVIVES** | none → **WR3** |
| **M6** | `activateWorkspace` ignores incoming watcher health (`session.go:120-122`) | extra | **SURVIVES** | none → **WR4** |
| **M7** | `newWatcher` accepts a zero-attachment watcher (`watch.go:63-66`) | extra | **SURVIVES** | none → **WR4** |

All nine mutations restored; `shasum` of `watch.go`, `model.go`, `view.go`, `session.go` matches the
pre-mutation backup, and `git diff` for each file shows only the implementation's own change.

#### Evidence floor #7 — `tea.Batch` event-order analysis

Commands that can issue a listener, tick, reload, or explicit reconcile:

| Origin | Commands issued | Listener count |
|---|---|---|
| `Init` (`model.go:203-212`) | `load`, optional `loadAtlas`, `waitForFS` if `m.watch != nil` | 1 (once) |
| `activateWorkspace` (`session.go:172`) | `closeWatcher(old)`, `loads`, `waitForFS(new)` | 1 |
| `fsEventMsg` (`model.go:440-452`) | `waitForFS` **iff** `!watchOff`, `debounceTick` | 1, or 0 when unavailable |
| `debounceMsg` (`model.go:454-465`) | `reload`; plus `waitForFS` **iff** `wasOff && !watchOff` | 0, or 1 on the recovery edge |
| `watcherReconciledMsg` (`model.go:431-438`) | `waitForFS` **iff** `wasOff && !watchOff` | 0, or 1 on the recovery edge |
| `reloadMsg` (`model.go:429`) | `reloadAll`, `reconcileWatcher` | 0 |
| stale `workspaceOpenedMsg` (`atlas.go:178`, `model.go:265`) | `closeWatcher(msg.watcher)` | 0 |

Arbitrary ordering cannot break the invariant, for three reasons:

1. **Each listener is single-shot.** `waitForFS` returns one message and terminates, so an
   `fsEventMsg` is proof its own listener has ended; the successor it issues restores the count to one.
2. **`Update` is serialized.** The recovery edge is guarded by `wasOff := m.watchOff` read and then
   overwritten inside one `Update` call. Whichever of `debounceMsg` / `watcherReconciledMsg` is
   processed first clears `watchOff`; the second observes `wasOff == false` and issues nothing. Probe
   P11 walked all ten transitions and measured ≤1 listener each.
3. **A zero-listener state is self-limiting.** `watchOff` implies `len(attached) == 0`, so no event can
   arrive and `dirtyGen` cannot advance — which means the `debounceTick` armed in the same handler
   always matches on gen and always reaches its recovery check. Crucially, the tick's `reconcile()`
   runs *before* the gen comparison in the reducer, so even a stale tick still reconciles.

Probe P6 ran the full reducer loop with `tea.BatchMsg` children queued in **reversed** order and
produced identical counts (`fsEventMsg:1 debounceMsg:1 reloadMsg:1 maxConcurrentListeners:1`).

Stale health cannot leak across a workspace switch because `Update` wraps every returned command in
`scopeSession(mm.sessionGen, …)` (`model.go:281`), `scopeSession` recurses into `tea.BatchMsg`
children (`session.go:33-41`), and `sessionScope` is enabled by `WithAtlas` (`model.go:141`) — the
same option that makes workspace switching reachable at all. P9 confirmed the drop end-to-end.

### Platform conclusions

**Demonstrated on this runtime (darwin/arm64, APFS, fsnotify v1.10.1 kqueue backend).** Everything in
the ledger above. Two Darwin-specific mechanics worth recording:

- **Per-file descriptor cost.** `kqueue.Add(dir)` calls `watchDirectoryFiles`
  (`backend_kqueue.go:582`), opening an fd for every entry in the directory. Measured: 6 watched dirs
  holding 100 files cost **109** additional open files. This repo's `planning/tasks` holds ~200 task
  files, so a real session's watcher costs roughly that many descriptors. It is bounded and fully
  released on `Close` (measured: exact return to baseline), but it is the mechanism behind the
  contamination failures, and it does **not** transfer to Linux.
- **The inode comparison is load-bearing on kqueue specifically.** `addWatch` short-circuits when
  `alreadyWatching` and reuses the existing fd (`backend_kqueue.go:351-372`), so a re-`Add` without a
  preceding `Remove` would re-register the *old* inode. The `os.SameFile` → `Remove` → `Add` sequence
  is what avoids that. Sentinel discovery works because a new subdirectory surfaces through
  `dirChange` → `sendCreateIfNew` (`backend_kqueue.go:622,656`); note the auto-added internal watch
  uses only `NOTE_DELETE|NOTE_RENAME` (`internalWatch`, `backend_kqueue.go:672-678`), so the explicit
  direct-leaf `Add` — which upgrades the flags to `noteAllEvents` and re-lists the directory — is
  required for writes inside a newly appeared leaf to be seen at all.

**Source-inspected, not executed.**

- **Linux/inotify.** The default op set includes `IN_CREATE` (`backend_inotify.go:204`), so a sentinel
  observes leaf creation. `register` (`backend_inotify.go:269-295`) re-resolves the path and receives
  a fresh `wd`; the stale `wd` is dropped only by an explicit `Remove`, so the implementation's
  Remove-before-Add is load-bearing here too. `IN_Q_OVERFLOW` becomes `ErrEventOverflow` on the
  `Errors` channel (`backend_inotify.go:397`), which this design converts into a reconcile plus a
  reload — a good outcome for a dropped-event burst. inotify holds one watch per directory, so the
  Darwin descriptor profile does not apply; `max_user_watches` exhaustion returns `ENOSPC` from `Add`,
  yielding `degraded`, or `unavailable` at construction (WR6).
- **Windows/ReadDirectoryChangesW.** Watches are keyed by `(volume serial, file index)` via `getIno`
  (`backend_windows.go:290-311`) — inode identity, not path — so a replaced directory produces a
  distinct entry and the stale one persists until `Remove`, again making the `os.SameFile` path
  load-bearing. `sysFSALLEVENTS` (0xfff) maps to `FILE_NOTIFY_CHANGE_DIR_NAME` among others
  (`backend_windows.go:672-676`), so sentinel discovery of a created subdirectory should work.
  Handles are opened with `FILE_SHARE_DELETE`, so watching a directory does not block renaming or
  deleting it.

**Unverified — needs a target-platform test.**

- Whether `os.SameFile` on directories is reliable on Windows network shares and on filesystems that
  aggressively recycle inode numbers. This is the one residual correctness assumption I could neither
  falsify nor confirm: on APFS I observed strictly increasing inode numbers across 20 replacements
  (P7b, no false "same file"), but ABA inode reuse elsewhere could in principle make a replaced
  directory compare equal and skip the re-attach.
- The volume-root/UNC refusal. `normalizeWatchPaths`' `canonical == filepath.Dir(canonical)` test
  should reject `C:\` and `\\server\share\`, but this was not executed.
- All Linux behavior end-to-end.

I am not recommending polling or recursive watching to close any of these; the event-driven design is
sound and the corrections above are all local.

### Explicitly settled concerns

Each of these was attacked with a real reproduction and is settled because the hostile case and the
implementation jointly disprove it.

1. **Reload storms / busy retry loops.** P8 (570 events → 1 reload) and P19 (164 directory-churn
   events → 1 reload, watch set constant at 2). `reconcile` performs no filesystem mutation, so
   attachment changes cannot feed the event source. `reloadMsg` → `reconcileWatcher` →
   `watcherReconciledMsg` issues no reload, so manual `r` cannot recurse.
2. **Duplicate listeners.** P11 across all ten reducer transitions plus `maxConcurrentListeners = 1`
   measured through both bursts and under reversed batch ordering. See the `tea.Batch` analysis.
3. **Watch and descriptor leaks.** P7b: open files return to the *exact* pre-watcher baseline after
   `Close`, `WatchList()` empty; no growth across 20 replace+reconcile cycles. P17: closed watcher's
   `attached` is nil, its listener returns `nil`, old-tree writes are inert. `spaceSession`
   (`session.go:82-93`) has no watcher field, so cached sessions retain no backend resource.
4. **Old workspace affecting the new one.** P9: session scoping drops the stale tick, and scoping is
   guaranteed wherever a workspace switch is reachable (`WithAtlas` sets both).
5. **Unbounded sentinel.** `normalizeWatchPaths` rejects `""`, `/`, `/..`, `///` and the volume root;
   `nearestExistingWatchDirectory` refuses to climb to a root. P3 shows an all-inaccessible ancestor
   chain produces an error rather than a root watch. P2 shows five sibling leaves cost exactly six
   watches.
6. **Health truthfulness.** P15 covers direct-leaf `Add`, sentinel `Add`, total `Add` failure, and
   `stat` failure independently; `healthy` requires the whole required set (`watch.go:140-149`). The
   TOCTOU between the `required` scan and the add loop errs toward `degraded` in both directions (a
   leaf appearing late is not in `required`, so `complete` stays false; a leaf vanishing late fails
   `Add` or the post-`Add` check), never toward false healthy.
7. **Missed events after `Add` fails on an obsolete backend registration.** Reachable only if the
   fsnotify registration and `attached` diverge; on kqueue and Windows a Rename/Remove event makes
   fsnotify drop its own registration, and the implementation calls `fsw.Remove` on every path it
   deletes from `attached`. Where our reconcile wins the race against fsnotify's own event loop, the
   result is one spurious empty-named event → one extra reconcile → one bounded reload. Source-checked
   at `backend_kqueue.go:500-505`; no reproduction produced a missed event.
8. **Port and scope integrity.** `internal/tui/watch.go` contains no entity directory name; the only
   `tea.Tick` in `internal/tui` is the 200 ms debounce; no fsnotify recursion option is used; no
   watcher symbol appears in `core`/`store`/`domain`; the sole production `Layout` joins paths without
   checking existence, matching the updated port comment.
9. **Footer truncation.** P14: both labels render intact on the list, dashboard, and atlas footers at
   30–200 columns. At 20 columns `live-reload degraded` clips to `live-reload degrade…`; at 10 both
   clip to `live-relo…` — widths far below any usable terminal. Because the notice is a **prefix**,
   `truncate` sacrifices key hints rather than the health label. Not a finding; WR5 covers the
   separate transient-override question.
10. **Permanent `degraded` for a legitimately absent entity directory — a deliberate contract choice
    the owner should see.** A planning space that predates `threads/` (this task's own motivating
    scenario) shows `live-reload degraded` continuously until a Thread is created, even though reload
    works correctly through the sentinel. The same applies to a `Layout` that ever returned a regular
    file (P1/P3: pinned to `degraded`, since a file can never be a current direct leaf). This matches
    the brief exactly — "Missing or partially attached leaves are recoverable degraded state, not
    healthy state" — so I am not filing it, but the visible consequence lands precisely on the
    upgrade path this work targets.
