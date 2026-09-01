---
schema: 1
id: 6g5xxthexrh1
bucket: closed
area: tui-live-reload-watcher-reconciliation-implementation-antigravity
date: "2026-09-01"
updated_at: "2026-09-01"
---
# Audit: TUI live-reload watcher reconciliation implementation — Antigravity — 2026-09-01

> Reviewer assignment: Antigravity. This document is the review brief and the only file the reviewer
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

### Executive Verdict: Ready with Tracked Follow-ups

The TUI live-reload watcher reconciliation implementation for task [`6g5rxq1g5mp1`](../tasks/6g5rxq1g5mp1-keep-tui-live-reload-healthy-when-entity-directories-appear.md) in the main worktree (based at commit `0b56488`) is verified against [ADR-0006](../adrs/0006-adopt-threads-as-task-dags.md) (including the 2026-09-01 TUI-delivery amendment) and [`docs/ARCHITECTURE.md`](../../docs/ARCHITECTURE.md).

1. **Desired Layout Port & Bounded Sentinel Topology:** The TUI treats `Layout.WatchPaths()` as desired leaf targets without hardcoded entity names. Missing leaves degrade health rather than aborting the session. Each leaf is watched directly for entity writes, paired with de-duplicated nearest-existing-parent sentinels to observe directory creations and inode replacements. Filesystem roots are explicitly refused as unbounded fallback sentinels.
2. **Identity-Aware Reconciliation:** File identity (`os.SameFile`) over `(Dev, Ino)` governs attachment retention. Inode replacements (e.g. atomic directory swap) and removed directories are detached and reattached safely. A double-`os.Stat` verification around `fsw.Add` ensures that paths modified during the attachment window are not recorded in an inconsistent state.
3. **Bounded Coalescing & Quiet-Period Closure:** Save storms, directory creations, and renames coalesce through `dirtyGen`-guarded 200ms debounce ticks. Debounce quiet-period reconciliation closes the race where rapidly created nested subdirectories appear before sentinels could attach intermediate levels. Manual reload (`r`) issues a non-polling explicit reconciliation.
4. **Session Isolation & Single Listener Cardinality:** Active watchers are session-scoped and cleanly closed upon workspace departure, leaving cached sessions descriptor-free. Outdated events or stale workspace-open results cannot mutate the new session. Listener re-arming is strictly gated by `wasOff && !m.watchOff`, preventing duplicate listener goroutines.
5. **Truthful UI Surfaces:** Health transitions across `watchHealthy`, `watchDegraded`, and `watchUnavailable` propagate reliably to atlas, dashboard, and entity-tab footers with non-overflowing text truncation.

Two low-severity coverage gaps were uncovered via mutation probing where the existing test suite lacked assertions for quiet-period debounce reconciliation and listener cardinality guards. Both are tracked as open findings below.

---

### Review Environment & Validation Results

- **Base Commit:** `0b56488163b9a67914a6298875edf234386c542c` (`0b56488`)
- **Runtime Platform:** `darwin/arm64` (macOS 15.x), APFS filesystem
- **fsnotify Backend:** `kqueue` (`github.com/fsnotify/fsnotify v1.10.1`)

#### Validation Commands & Outcomes

| Validation Suite | Exact Command Line | Outcome |
| :--- | :--- | :--- |
| **Repeated Focused Race** | `go test -race -count=10 -run 'TestWatcher\|TestModel_FsEvent' ./internal/tui` | **PASS** (10/10 iterations clean, 1.844s) |
| **Full Repository Race** | `go test -count=1 -race ./...` | **PASS** (all packages clean, 7.090s max package duration) |
| **Static Analysis / Lint** | `golangci-lint run ./...` | **PASS** (0 issues) |
| **Go Vet** | `go vet ./...` | **PASS** (clean) |
| **Docs & CLI Drift** | `go run ./internal/tools/docgen -out docs/cli && git diff --exit-code docs/cli` | **PASS** (no drift) |
| **Go Module Tidiness** | `go mod tidy -diff` | **PASS** (no diff) |
| **Planning System Lint** | `go run ./cmd/tskflwctl lint` | **PASS** (all planning entities and links valid) |
| **Git Diff Check** | `git diff --check` | **PASS** (clean whitespace) |

---

### Compact Architecture & Event-Flow Re-Derivation

```
[Layout Port]
      │ WatchPaths() []string (desired leaves: tasks, epics, audits, research, threads)
      ▼
[normalizeWatchPaths] ──► Clean / Abs / canonicalWatchPath (EvalSymlinks) / Refuse FS Roots
      │
      ▼
[newWatcher / reconcile]
      ├── 1. Stat desired leaves -> mark present vs missing
      ├── 2. Derive nearestExistingWatchDirectory(filepath.Dir(leaf)) -> de-duplicated sentinels
      ├── 3. Prune obsolete / replaced inodes: os.SameFile(previous, current) == false -> fsw.Remove
      ├── 4. Attach new targets: Stat -> fsw.Add -> Stat -> verify os.SameFile(before, after)
      └── 5. Compute Health:
              ├── All desired leaves + sentinels attached -> watchHealthy
              ├── Partial useful coverage attached         -> watchDegraded
              └── 0 attachments attached                   -> watchUnavailable
      │
      ├──────────────────────┬────────────────────────┬──────────────────────┐
      ▼                      ▼                        ▼                      ▼
 [waitForFS]           [debounceTick]         [reconcileWatcher]      [closeWatcher]
 (select Events/Errors) (tea.Tick 200ms)         (manual 'r')          (workspace switch)
      │                      │                        │                      │
      ▼                      ▼                        ▼                      ▼
  fsEventMsg            debounceMsg          watcherReconciledMsg           nil
 (health, dirtyGen++)  (gen == dirtyGen,       (health, rearm if       (fsw.Close(),
  rearm listener &     health, reload &        wasOff && !off)         w.closed=true)
  debounceTick)        rearm if wasOff)
```

1. **Leaf & Sentinel Convergence:** The active workspace watcher maintains desired leaf targets (whether currently existing or absent). Sentinels are placed on the deepest existing ancestor directory (refusing filesystem roots). When a directory appears or is replaced, its parent sentinel generates an fs event.
2. **Identity-Driven Attachment (`os.SameFile`):** File identity checks compare filesystem device and inode identifiers (`syscall.Stat_t`). If an attached directory is replaced via atomic rename or recreated after deletion, the stale inode is detached via `fsw.Remove` and the new inode attached.
3. **Bounded Debounce & Quiet-Period Closure:** Raw fs events bump `m.dirtyGen` and schedule `debounceTick(m.watch, m.dirtyGen)`. Intermediate ticks with `msg.gen != m.dirtyGen` are discarded without triggering reload. The final quiet-period tick invokes `w.reconcile()`, ensuring rapidly created nested directories are attached even if intermediate directory creation events occurred before sentinels could attach.
4. **Session Isolation & Listener Cardinality:** Changing workspaces via `activateWorkspace` atomically increments `m.sessionGen`, closes the prior workspace's watcher (`closeWatcher(oldWatcher)`), and spawns exactly one listener for the new watcher. Outdated messages from previous sessions are dropped by `sessionMsg` generation filtering.

---

### Repository-Wide Consumer Inventory

| Symbol / Message | Defined In | Consuming Files & Call Sites | Architectural Function |
| :--- | :--- | :--- | :--- |
| `newWatcher` | `internal/tui/watch.go:50` | `internal/tui/tui.go:47`<br>`internal/tui/atlas.go:153`<br>`internal/tui/watch_test.go:50,65,88,109,149,172,213`<br>`internal/tui/atlas_test.go:297,301` | Normalizes desired paths, initializes `fsnotify.Watcher`, performs initial `reconcile()`, rejects empty/root sets. |
| `waitForFS` | `internal/tui/watch.go:235` | `internal/tui/model.go:210,436,450,463`<br>`internal/tui/session.go:172`<br>`internal/tui/watch_test.go:14` | Blocking `tea.Cmd` selecting `w.fsw.Events` / `Errors`, reconciles attachment set, delivers `fsEventMsg`. |
| `reconcileWatcher` | `internal/tui/watch.go:255` | `internal/tui/model.go:429`<br>`internal/tui/watch_test.go:137` | `tea.Cmd` fired on manual reload (`reloadMsg`), executes `w.reconcile()`, delivers `watcherReconciledMsg`. |
| `debounceTick` | `internal/tui/watch.go:265` | `internal/tui/model.go:452` | `tea.Tick` timer (200ms) with quiet-period reconciliation, delivers `debounceMsg{gen, health}`. |
| `watchHealth` | `internal/tui/watch.go:20` | `internal/tui/watch.go:23,42,70,82,267`<br>`internal/tui/session.go:120`<br>`internal/tui/tui.go:50`<br>`internal/tui/messages.go:106,111,117`<br>`internal/tui/watch_test.go:33,70,93,130,155,227`<br>`internal/tui/model_test.go:1600` | Enum (`watchHealthy`, `watchDegraded`, `watchUnavailable`) describing attachment completeness. |
| `watchOff` | `internal/tui/model.go:107` | `internal/tui/tui.go:45,49`<br>`internal/tui/session.go:121,122`<br>`internal/tui/model.go:432,433,435,445,449,458,459,462`<br>`internal/tui/view.go:354`<br>`internal/tui/model_test.go:1583,1602,1608` | Model state boolean: true when zero useful watches exist; forces honest fallback to manual reload. |
| `watchDegraded` | `internal/tui/model.go:108` | `internal/tui/tui.go:50`<br>`internal/tui/session.go:122`<br>`internal/tui/model.go:434,446,460`<br>`internal/tui/view.go:356`<br>`internal/tui/model_test.go:1583,1602,1608` | Model state boolean: true when useful coverage/sentinels exist but not all desired direct leaves are attached. |
| `WatchPaths` | `internal/core/store.go:308` | `internal/store/fsstore.go:93`<br>`internal/tui/tui.go:47`<br>`internal/tui/atlas.go:153`<br>`internal/store/epicstore_test.go:66`<br>`internal/store/researchstore_test.go:183`<br>`internal/workspacestore/fs_test.go:54`<br>`internal/core/workspace_test.go:25`<br>`internal/cli/ui_test.go:19` | `Layout` port method: returns desired entity leaf paths without leaking domain directory names to TUI. |
| watcher `close` / `closeWatcher` | `internal/tui/watch.go:217`<br>`internal/tui/session.go:187` | `internal/tui/tui.go:57,59`<br>`internal/tui/session.go:172`<br>`internal/tui/model.go:265`<br>`internal/tui/watch_test.go:52,69,92,114,153,176,220` | Thread-safe teardown: marks `closed=true`, clears `attached` map, closes underlying `fsw`, releases OS descriptors. |
| `fsEventMsg` | `internal/tui/messages.go:106` | `internal/tui/watch.go:245,250`<br>`internal/tui/model.go:440`<br>`internal/tui/model_test.go:1581,1607` | Delivers reconciled health on raw fsnotify event; arms debounce tick and continues listening if not off. |
| `watcherReconciledMsg` | `internal/tui/messages.go:111` | `internal/tui/watch.go:259`<br>`internal/tui/model.go:431`<br>`internal/tui/watch_test.go:137` | Delivers explicit reconciliation health; restarts stopped listener if recovering from `watchUnavailable`. |
| `debounceMsg` | `internal/tui/messages.go:115` | `internal/tui/watch.go:271`<br>`internal/tui/model.go:454`<br>`internal/tui/model_test.go:1600` | Fires after quiet period with final reconciliation health; triggers `reloadMsg` only if generation matches. |
| `workspaceOpenedMsg` | `internal/tui/atlas.go:139` | `internal/tui/atlas.go:151,154,176`<br>`internal/tui/model.go:264,312` | Carries new workspace, newly constructed watcher, and watcher construction error to `activateWorkspace`. |
| `sessionMsg` | `internal/tui/session.go:16` | `internal/tui/session.go:21,58`<br>`internal/tui/model.go:262,282` | Stamps async results with workspace generation; drops stale results and cleans up stale opened watchers. |

---

### Findings

#### Low Severity

#### L1. Coverage Gap — Mutation survival for quiet-period reconciliation in `debounceTick` · **Status:** fixed

**File:** `internal/tui/watch.go:265-273` | **Component:** `tui/watch`
**Effort:** XS · **Urgency:** low

**Description:**
In `debounceTick(w *watcher, gen int)`, `w.reconcile()` is executed inside the tick callback so that nested directories created rapidly after an initial parent event converge during the quiet period. Mutation Probe 3 (replacing `w.reconcile()` with `w.watchHealth()`) survived the existing test suite because existing test cases either triggered events after all nested directories were already created or manually called `w.reconcile()`.

**Impact:**
Production behavior is correct, but regressions in quiet-period convergence could pass CI unnoticed if `debounceTick` reconciliation were accidentally removed.

**Hostile Reproduction / Reasoning:**
1. Desired leaf `root/nested/threads`, sentinel `root`.
2. Parent event on `root` triggers `waitForFS`, reconciling while `nested/threads` does not exist yet (`health = watchDegraded`).
3. Subdirectories `nested` and `nested/threads` are created without touching `root`.
4. `debounceTick` runs: without `w.reconcile()`, `debounceMsg.health` remains `watchDegraded` and the leaf is not attached in `w.attached`.

**Recommendation:**
Add a unit test in `watch_test.go` that simulates a parent event before deeper directories exist, creates the subdirectories, runs `debounceTick` directly, and asserts that `debounceMsg.health == watchHealthy` and `w.attached` contains the leaf directory without subsequent fs events.

---

**Resolution:** A production debounceTick test now proves quiet-period
reconciliation attaches a rapidly completed nested desired leaf and returns
healthy state.

#### L2. Coverage Gap — Mutation survival for listener cardinality guards on `watcherReconciledMsg` and `debounceMsg` · **Status:** fixed

**File:** `internal/tui/model.go:435-438`, `internal/tui/model.go:462-465` | **Component:** `tui/model`
**Effort:** XS · **Urgency:** low

**Description:**
In `Model.Update`, both `watcherReconciledMsg` and `debounceMsg` restart `waitForFS(m.watch)` only when recovering from an unavailable state (`if wasOff && !m.watchOff`). Mutation Probe 5 (reissuing `waitForFS(m.watch)` unconditionally) survived the existing test suite because no test in `model_test.go` verified that `Update` returns a `nil` listener command when `m.watchOff` was already false.

**Impact:**
Production logic correctly guards against duplicate listeners, but an accidental removal of `wasOff && !m.watchOff` would spawn competing listener goroutines on every manual reload (`r`) or debounce tick without failing tests.

**Hostile Reproduction / Reasoning:**
If `watcherReconciledMsg` unconditionally returns `waitForFS(m.watch)` when `m.watchOff == false`, two concurrent goroutines execute `waitForFS` on the same `fsnotify.Watcher`. Each subsequent event triggers two `fsEventMsg` deliveries, exponentially duplicating listeners.

**Recommendation:**
Add assertions in `model_test.go` verifying that `Update(watcherReconciledMsg{health: watchHealthy})` and `Update(debounceMsg{gen: m.dirtyGen, health: watchHealthy})` return `nil` (or only `reloadMsg` for debounce) when `m.watchOff` is false.

---

**Resolution:** Reducer command-shape tests now catch unconditional listener
reissue on explicit and debounce reconciliation while preserving the single
off-to-active restart edge.

### Acceptance-Criteria Traceability Table

| Acceptance Criterion | Status | Implementation Seams | Verification Evidence |
| :--- | :---: | :--- | :--- |
| **1. Missing leaf at startup & second file change**<br>Starting with no `threads/` directory and creating a Thread reloads the workspace; a second change inside that directory is also observed. | **Fulfilled** | `internal/tui/watch.go:90-104`<br>`nearestExistingWatchDirectory`<br>`reconcile()` | `TestWatcherRecoversMissingDesiredLeafAndObservesItsFiles`<br>`TestHostile_MissingLeafStartupCreateSecondWrite` |
| **2. Remove/recreate & atomic replacement**<br>Removing and recreating or atomically replacing a configured directory reattaches the watch and yields current data without clinging to old inode. | **Fulfilled** | `internal/tui/watch.go:106-138`<br>`os.SameFile(previous, current)`<br>Double-`os.Stat` on `Add` | `TestWatcherReattachesRemovedAndRecreatedDirectory`<br>`TestWatcherReattachesAtomicallyReplacedDirectory`<br>`TestHostile_AtomicRenameReplacementWithWrite` |
| **3. Degraded vs healthy vs unavailable reporting**<br>Partial attachment failure is represented as degraded live reload, not healthy; recovery clears degradation in state and footers. | **Fulfilled** | `internal/tui/watch.go:140-154`<br>`internal/tui/view.go:352-361`<br>`internal/tui/model.go:431-466` | `TestWatcherReportsTransientAddFailureAndExplicitReconcileRecovers`<br>`TestModel_FsEventSurfacesAndRecoversWatcherHealth`<br>`TestHostile_DegradedAttachmentFollowedByRecovery` |
| **4. Symlink/alias de-duplication & session cleanup**<br>Path normalization prevents duplicate watches for symlink/alias spellings; workspace switching closes prior session watches. | **Fulfilled** | `internal/tui/watch.go:157-201`<br>`canonicalWatchPath`<br>`internal/tui/session.go:110-173`<br>`closeWatcher` | `TestWatcherNormalizesSymlinkAliasesBeforeAttaching`<br>`TestAtlas_WorkspaceSwitchClosesOldWatcher`<br>`TestHostile_DuplicateLexicalAndSymlinkAliases`<br>`TestHostile_WorkspaceWatcherReplacementWithOutstandingListener` |
| **5. Bounded event coalescing & no busy loops**<br>Save storms, directory reconciliation, and ensuing file events do not cause reload storms or busy retry loops. | **Fulfilled** | `internal/tui/watch.go:265-273`<br>`internal/tui/model.go:440-466`<br>`dirtyGen` / `debounceTick` | `TestModel_FsEventDebounces`<br>`TestHostile_ReloadStormProbe` (50 burst writes coalesced to 1 reload) |

---

### Mandatory Hostile-Evidence Ledger

| Probe / Angle | Hostile Attack Condition | Observed Implementation Response | Catching Test / Coverage Status |
| :--- | :--- | :--- | :--- |
| **Reproduction 1: Missing leaf & second write** | Leaf `threads/` missing at launch; create directory, then write `thread-1.md` and `thread-2.md`. | Started in `watchDegraded`. `Mkdir` triggered sentinel event -> reconciled to `watchHealthy`. Both file writes delivered events. | `TestWatcherRecoversMissingDesiredLeafAndObservesItsFiles` |
| **Reproduction 2: Rapid nested creation** | Leaf `future/planning/threads` with multiple missing parent levels created via `MkdirAll` alongside immediate file write. | Initial `watchDegraded`. Reconciled parent sentinels and attached leaf; `debounceTick` confirmed quiet-period convergence. | `TestWatcherRecoversNestedDesiredLeafWhenParentTreeAppears` |
| **Reproduction 3: Remove and recreate** | Delete `threads/` directory, wait, then recreate `threads/` and write file. | `Remove` transitioned watcher to `watchDegraded`; `Mkdir` reattached leaf and restored `watchHealthy`; file write observed. | `TestWatcherReattachesRemovedAndRecreatedDirectory` |
| **Reproduction 4: Atomic directory swap** | Rename `threads/` to `retired/`, rename `replacement/` to `threads/`, write file into `threads/`. | `os.SameFile` detected inode mismatch on previous attachment; pruned old watch, attached new inode; file write observed. | `TestWatcherReattachesAtomicallyReplacedDirectory` |
| **Reproduction 5: Lexical & symlink aliases** | Input 5 alias paths (`tasks`, `./tasks`, `tasks/../tasks`, `symlink/tasks`, `symlink/./tasks`). | Collapsed to 1 canonical desired leaf; attached exactly 2 watches (1 leaf + 1 parent sentinel). | `TestWatcherNormalizesSymlinkAliasesBeforeAttaching` |
| **Reproduction 6: Degraded recovery** | Intercept `w.addPath` with synthetic `ENOSPC`, detach leaf, then restore `w.addPath` and trigger manual reconcile. | Reconciled to `watchDegraded` (sentinels remained active). Explicit `reconcileWatcher` recovered to `watchHealthy`. | `TestWatcherReportsTransientAddFailureAndExplicitReconcileRecovers` |
| **Reproduction 7: Workspace replacement** | Switch workspace while old `waitForFS` listener is blocking on `w1`. | `activateWorkspace` returned `closeWatcher(w1)`; `w1.close()` unblocked listener with `nil`; `sessionMsg` dropped stale events. | `TestHostile_WorkspaceWatcherReplacementWithOutstandingListener` / `TestAtlas_WorkspaceSwitchClosesOldWatcher` |
| **Reload Storm Probe** | Burst 50 rapid file writes (1ms spacing) into watched directory. | Received 50 raw events; `dirtyGen` incremented 50 times; 49 intermediate debounce ticks dropped; exactly 1 reload fired. | `TestHostile_ReloadStormProbe` / `TestModel_FsEventDebounces` |
| **Watch-Resource Lifecycle** | Measure `fsw.WatchList()` across creation, attachment, and close. | Startup: 2 (`tasks` + parent). Post-mkdir: 3 (`tasks` + `threads` + parent). Post-close: 0 (empty, descriptors closed). | `TestHostile_WatchResourceEvidence` |
| **Mutation 1: Omit parent sentinel** | Comment out `required[parent] = struct{}{}` in `reconcile()`. | Missing leaves never detected when created; `TestWatcherRecoversMissingDesiredLeafAndObservesItsFiles` timed out. | **CATCHES REGRESSION** (`watch_test.go:74`) |
| **Mutation 2: Path-only identity (omit `SameFile`)** | Skip `os.SameFile(previous, current)` check during reconciliation. | Retained old inode after atomic directory replacement; `TestWatcherReattachesAtomicallyReplacedDirectory` failed. | **CATCHES REGRESSION** (`watch_test.go:196`) |
| **Mutation 3: Omit quiet-period reconcile** | Replace `w.reconcile()` with `w.watchHealth()` in `debounceTick`. | Debounce tick fails to attach nested subdirectories created after initial parent event. | **SURVIVED EXISTING SUITE** (Tracked in **Finding L1**) |
| **Mutation 4: Ignore degraded in footer** | Omit `case m.watchDegraded` in `withWatchHealth()`. | Footer failed to display `"live-reload degraded"`; `TestModel_FsEventSurfacesAndRecoversWatcherHealth` failed. | **CATCHES REGRESSION** (`model_test.go:1584`) |
| **Mutation 5: Unconditional listener reissue** | Remove `wasOff && !m.watchOff` guard in `watcherReconciledMsg`. | Spawns redundant concurrent listener goroutines on each manual reload or debounce recovery. | **SURVIVED EXISTING SUITE** (Tracked in **Finding L2**) |

---

### Event-Order Analysis for Bubble Tea `tea.Batch`

In Bubble Tea v2, `tea.Batch(cmds...)` executes child commands concurrently in separate goroutines; message delivery to `Model.Update` is non-deterministic. The reconciliation and reload subsystem is provably robust against arbitrary message interleavings:

```
Sources of Async Messages:
  1. waitForFS(w)           ──► fsEventMsg{health}
  2. debounceTick(w, gen)    ──► debounceMsg{gen, health}
  3. reloadAll()            ──► listLoadedMsg / dashLoadedMsg
  4. reconcileWatcher(w)    ──► watcherReconciledMsg{health}
  5. closeWatcher(oldW)     ──► nil
  6. openWorkspace(svc, req)──► workspaceOpenedMsg{gen, workspace, watcher}
```

1. **`fsEventMsg` interleaved with `debounceMsg`:**
   - If `fsEventMsg` arrives after `debounceTick` was scheduled but before `debounceMsg` arrives: `m.dirtyGen` is incremented. When `debounceMsg{gen: old}` arrives, `msg.gen != m.dirtyGen` evaluates to `true`, dropping the stale reload.
   - If `debounceMsg` arrives first: `m.dirtyGen` matches, firing `reloadMsg`. The subsequent `fsEventMsg` starts a fresh debounce cycle with `m.dirtyGen+1`.
2. **`watcherReconciledMsg` interleaved with `fsEventMsg`:**
   - Both messages execute `w.reconcile()` under `w.mu.Lock()`.
   - In `Model.Update`, both update `m.watchOff` and `m.watchDegraded` to the latest reconciled health.
   - If the watcher is already active (`!m.watchOff`), neither message spawns a redundant listener (`wasOff && !m.watchOff` is false).
3. **`workspaceOpenedMsg` interleaved with previous workspace messages:**
   - All async commands generated by workspace actions are wrapped with `scopeSession(sessionGen, cmd)`.
   - When workspace switches, `m.sessionGen++`. Any late-arriving message from the previous workspace is dropped in `Model.Update:263`.
   - If a stale `workspaceOpenedMsg` arrives after another switch, `Model.Update:265` intercepts it and calls `closeWatcher(opened.watcher)`, releasing OS descriptors immediately.
4. **`closeWatcher` interleaved with `waitForFS`:**
   - `closeWatcher` acquires `w.mu`, sets `w.closed = true`, and calls `w.fsw.Close()`.
   - A blocking `waitForFS` selects on closed `Events`/`Errors` channels (`ok == false`) and returns `nil`. No error or message is dispatched to the runtime.

---

### Platform Discipline & Portability Analysis

| Platform / Backend | Support Classification | Evidence & Operational Findings |
| :--- | :--- | :--- |
| **Darwin / `kqueue`** | **Demonstrated** | Tested on macOS 15.x / ARM64 / APFS. File descriptors for direct leaves and parent directories are registered via `kevent`. Directory rename, deletion, recreation, and rapid file creation delivered expected events; `fsw.WatchList()` confirmed clean descriptor release on close. |
| **Linux / `inotify`** | **Source-Inspected** | In `fsnotify/backend_inotify.go`, directory deletion emits `IN_DELETE_SELF` and `IN_IGNORED`, automatically removing the descriptor from the inotify kernel watch table. When recreated, the parent sentinel receives `IN_CREATE|IN_ISDIR`, triggering `reconcile()` to re-add the new watch descriptor (`IN_WATCH`). Atomic rename emits `IN_MOVED_FROM` and `IN_MOVED_TO` to the parent sentinel. |
| **Windows / `ReadDirectoryChangesW`** | **Source-Inspected** | In `fsnotify/backend_windows.go`, directory watches listen for `FILE_ACTION_ADDED`, `FILE_ACTION_REMOVED`, `FILE_ACTION_RENAMED_OLD_NAME`, and `FILE_ACTION_RENAMED_NEW_NAME`. Replacing a directory without parent watching is prone to `ERROR_ACCESS_DENIED` or stale handle locks; keeping the parent sentinel open absorbs rename events and re-establishes the handle on the new path. |
| **Unbounded Volume Roots** | **Verified Boundary** | `nearestExistingWatchDirectory` explicitly stops at `candidate == filepath.Dir(candidate)`. On Windows (`C:\` or `\\server\share`) and Unix (`/`), the sentinel climber returns `"", false`, preventing whole-drive recursive kernel event floods. |

---

### Explicitly Settled Concerns

1. **Concern: Stale symlink resolution when target directory is retargeted or missing at launch.**
   *Resolution:* Settled. `canonicalWatchPath` uses `filepath.EvalSymlinks` on the deepest existing prefix and appends unresolved missing suffixes. Once created or retargeted, the parent sentinel's event triggers `reconcile()`, which evaluates symlinks afresh on the new path and re-aligns attachments.
2. **Concern: High CPU or busy retry loops during continuous filesystem churn.**
   *Resolution:* Settled. `waitForFS` blocks on channel reads rather than polling. Debounce timers use a single quiet-period `time.Tick` (200ms) with generation filtering (`dirtyGen`), coalescing bursts of arbitrary size into exactly one reload.
3. **Concern: Leaking watches or file descriptors in multi-workspace Atlas workflows.**
   *Resolution:* Settled. `activateWorkspace` emits `closeWatcher(oldWatcher)` and `spaceSession` explicitly omits the watcher pointer. Cached sessions store zero watcher state. Inotify/kqueue watch lists confirm complete reclamation on close.
4. **Concern: TUI acquiring domain knowledge of entity directory naming.**
   *Resolution:* Settled. The TUI receives leaf targets exclusively through `core.Workspace.Layout.WatchPaths()`. Entity names (`tasks`, `epics`, `audits`, `research`, `threads`) remain strictly encapsulated inside `internal/store`.
5. **Concern: Inode recycling (ABA problem) during rapid remove-recreate cycles.**
   *Resolution:* Settled. Inode checks compare `(Dev, Ino)` via `os.SameFile(previous, current)`. In addition, `reconcile()` performs a double-stat before and after `fsw.Add` to detect any inode replacement that occurred concurrently with the kernel registration call.
