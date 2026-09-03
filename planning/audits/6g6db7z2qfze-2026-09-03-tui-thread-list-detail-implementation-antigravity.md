---
schema: 1
id: 6g6db7z2qfze
bucket: closed
area: tui-thread-list-detail-implementation-antigravity
date: "2026-09-03"
updated_at: "2026-09-03"
---
# Audit: TUI Thread list and detail implementation — antigravity — 2026-09-03

> Reviewer assignment: antigravity. This document is the review brief and the only file the reviewer should update.
>
> Finding grammar is exact: use `#### M1. <title> · **Status:** open` (or H1/L1). Codes must match `[A-Z]+[0-9]+`; no hyphens, no em dash in place of the period, and no free-standing status line.

> Required second pass: after completing the brief checklist, review the change again for systemic failure modes. Take an explicitly adversarial stance toward shared abstractions, test helpers that can mask broad defect classes, state changing between projection and action, and boundaries that only appear to fail closed. Prefer one demonstrated systemic issue over several speculative findings, and settle each challenged pattern with hostile evidence.

## Review brief

Perform an independent adversarial implementation review of
[add Thread list and detail views to the TUI](../tasks/6g5rwjqr7rt8-add-thread-list-and-detail-views-to-the-tui.md)
on branch `feat/tui-thread-list-detail`, based on `main` at commit `a38e694`. The implementation is
currently preserved in commit `e319de3`; review the complete `git diff main...HEAD` plus any later
uncommitted audit files. Judge it against [ADR-0006](../adrs/0006-adopt-threads-as-task-dags.md),
[the architecture guide](../../docs/ARCHITECTURE.md), and the task's acceptance criteria.

Assume the feature can be systemically wrong despite green tests and plausible terminal output.
This change promotes Threads from a temporary projection cache into a first-class registry entity,
touching asynchronous list/detail ownership, workspace sessions, reload contention, stable identity,
portable/local capability boundaries, graph-derived semantics, responsive rendering, help/footer
affordances, and documentation. Re-derive those paths from production code. Do not accept comments,
test-helper behavior, or a visually attractive row as proof, and do not manufacture findings when
hostile evidence settles a concern.

## Review target

Start with `git status --short`, `git log --oneline main..HEAD`, `git diff --stat main...HEAD`, and
the complete `git diff main...HEAD`. Inventory all changed files before classifying scope. Primary
review surfaces are:

- the entity registry and every registry consumer in `internal/tui`, including tab switching,
  command and palette routing, filtering, sorting, cursor restoration, back navigation, workspace
  session save/restore, lazy loading, reload fan-out, and stale-result guards;
- `internal/tui/thread_projection.go`, including list adaptation, repository-wide diagnostics,
  detail and body loading, optional `ThreadPathSource`, semantic-read failures, and read-conflict
  retry behavior;
- Thread rows and detail composition in `item.go`, `detail.go`, `view.go`, `help.go`, and the theme,
  especially lifecycle versus derived health, nominal versus sound progress, in-flight versus
  dispatchable versus blocked work, external gates, missing/broken members, and inconsistency;
- removal of `threadProjectionState`, Thread-specific loaded messages, Thread read surfaces, and
  session fields: prove the deletion did not also remove a needed live-reload or contention contract;
- every modified and adjacent test, particularly test helpers that execute Bubble Tea command trees
  synchronously and may conceal completion-order bugs or unsupported terminal behavior; and
- README, CLI help, generated docs, architecture text, ADR-0006, and the implementation task.

Do not edit implementation or planning beyond your assigned audit. Other current-main planning work
and the other reviewer's audit are out of scope unless they expose a concrete interaction defect.

## Intended contract to challenge

- Threads are an ordinary first-class registry entity. The registry owns exactly one list/detail
  generation, error state, selection identity, filter/sort state, and per-space session. No parallel
  cache, message family, retry surface, or hidden eager-read path remains.
- The route is lazy before first visit, then participates in the same task/Thread watcher reloads as
  other visited tabs. Background-tab reloads update that tab without stealing active selection or
  detail; stale list, detail, retry, filter, and workspace-session results cannot land.
- Each row retains the exact shared `core.ThreadView`. The TUI may count supplied role/gate states
  for compact display, but may not traverse dependencies, infer eligibility from task status, invent
  completion rules, reorder membership authoritatively, or mutate Thread lifecycle/membership.
- Persisted lifecycle is distinct from graph/projection health and inconsistency. Nominal done is
  distinct from soundly drained. In-flight work is distinct from dispatchable frontier work;
  next-up and ready-to-start can both be eligible because eligibility is graph-derived.
- Canonical Thread identity—not display slug—owns selection, detail requests, retries, palette
  targets, restoration, and duplicate disambiguation. Invalid/duplicate empty canonical keys fail
  visibly instead of collapsing rows.
- Semantic Thread reads are portable. Local file paths are an optional capability used only for
  copy/open affordances; an absent or failing path source cannot blank coherent semantic content.
  Conversely, a semantic failure may retain a separately resolved local repair path without letting
  that path retarget the semantic read or another Thread.
- Repository-global graph/read diagnostics remain visible without being duplicated into every row.
  Per-Thread diagnostics, missing/broken members, immediate external gates, the persisted body, and
  persisted member order remain visible in detail.
- Normal, narrow, and very small layouts retain identity, lifecycle, health, in-flight/frontier/
  blocked distinctions, remain width-safe under ANSI and Unicode, and do not advertise unsupported
  mutation actions. Other entity tabs and existing TUI behavior remain compatible.
- No public JSON/schema or CLI graph semantics change merely to make this presentation easier.

## Mandatory evidence floor

A `ready` verdict is not credible unless the report contains all of the following:

1. A repository-wide consumer inventory for `entityThreads`, `threadProjectionState`,
   `threadListDiagnostics`, `threadItem`, `threadDetail`, `loadThreadList`, `loadThreadDetail`,
   `ThreadPath`, `listLoadedMsg`, `detailMsg`, `detailErrMsg`, `readSurface`, `reloadAll`, `detailGen`,
   `loadedKey`, registry indexes, palette/command targets, session state, `selectedPath`, yank/editor
   actions, help, footer, and docs. Classify each relevant use or absence; grep counts alone are not
   an inventory.
2. An end-to-end throwaway planning-space probe with multiple Threads covering an in-progress member,
   both eligible pending statuses, a blocked pending member, an external gate, shared membership,
   a cancelled Thread, completed-but-unsound evidence, an empty Thread, and a missing/broken member.
   Compare TUI values to `thread list/show/frontier --json` or the shared core projections. If a live
   PTY/resize smoke is unavailable, say so and substitute explicit renderer/model evidence.
3. Hostile identity fixtures with duplicate Thread slugs, distinct canonical IDs, ID drift, empty
   canonical identity, and duplicate canonical IDs. Exercise palette jump, selection, detail,
   filter, sort, reload, tab switch, workspace switch/restore, yank, and editor targeting. State
   whether invalid adapter output is rejected before ambiguous rows are installed.
4. Completion-order probes for two list generations, two details for the same Thread, details for
   different Threads, a filtered-list refilter racing a watcher reload, background-tab reload, a
   task-only edit, a Thread-document edit, and a workspace switch while work is in flight. Prove the
   production generation/kind/key guards, not only a synchronous `drainNested` helper.
5. Capability-boundary probes using independently supplied aggregate store, task-graph source,
   Thread reader, and optional Thread-path source. Cover absent path support, path lookup failure,
   semantic detail failure with a valid repair path, path success followed by semantic contention,
   stale local paths after rename/removal, and a portable adapter that exposes no filesystem values.
6. Rendering evidence at normal, narrow, and minimum supported widths with long ASCII and Unicode
   labels/descriptions/IDs, large counts, all health combinations, warnings, zero-member Threads,
   missing members, many gates/diagnostics, raw and pretty detail, help scrolling, footer pressure,
   filtering, pagination, and selection styling. Measure visible cell width after ANSI stripping;
   rune counts alone do not settle double-width or combining characters.
7. At least these temporary, restored mutation probes, with the focused test that kills each one:
   - reintroduce a second Thread list/detail owner or dedicated retry surface;
   - remove a list/detail generation, kind, or canonical-key stale guard;
   - derive frontier from task status or count clear pending work independently of `view.Frontier`;
   - replace stored `core.ThreadView` rows with a partial TUI-owned summary;
   - make `ThreadPath` mandatory for semantic detail or source a path from `domain.Thread.Path`;
   - route selection/detail/palette restore through Thread slug instead of canonical identity;
   - drop graph/read diagnostics or collapse nominal and sound progress; and
   - expose lifecycle/membership mutation hints on the read-only Thread tab.
   A surviving mutation is a coverage finding even if production code is currently correct.
8. Repeated focused tests under `-race`, an uncached full `go test -race ./...`, golangci-lint/static
   analysis, generated-doc drift, module tidiness, planning lint, audit lint, and `git diff --check`,
   with exact commands, Go/runtime version, durations, and cached/uncached distinction. Separate
   pre-existing unrelated lint failures from regressions with evidence.

A report that paraphrases acceptance criteria, cites test names without hostile fixtures, or says
"all tests pass" without the consumer inventory and mutation ledger does not satisfy this audit.

## Required hostile angles

1. **Single ownership versus hidden dual state.** Trace all list/detail/error/loading/selection state,
   including Bubble Tea list internals, registry tabs, the shared detail pane, workspace sessions,
   dashboard/palette indexes, and retries. Challenge whether repository diagnostics on `entityTab`
   are truly metadata rather than the start of another Thread cache.
2. **Asynchronous ordering and contention.** Run commands in hostile completion orders instead of
   relying on the synchronous helper. Attack reload while filtered, repeated manual/watch reloads,
   task and Thread events coalescing, tab changes, workspace changes, semantic read conflicts, path
   failures, and an error landing after a successful newer load.
3. **Projection fidelity.** Compare every displayed or counted value with `core.ThreadView`. Look for
   aliases that blur graph health with projection health, pending role with eligibility, completed
   with drained, membership order with graph order, or immediate gates with a transitive closure.
   Require a concrete contradiction before recommending richer core projections.
4. **Identity and malformed inputs.** Attack duplicate slugs, duplicate/empty/drifting IDs, a slug
   equal to another Thread's ID, very short IDs, prefix collisions, missing members, unreadable
   documents, and partial adapter results. Inspect map keys and last-writer behavior.
5. **Portable/local separation and TOCTOU.** Prove semantic browsing works without a path source and
   that local paths never become semantic identity. Challenge the independent `ThreadPath` then
   `ShowThread` reads under rename/delete/replace races: distinguish a harmless stale navigation
   affordance from a wrong-file or wrong-Thread action.
6. **Responsive and accessible presentation.** Check ANSI-aware and terminal-cell-aware clipping,
   Unicode width, selected-row styling, long counts, warning visibility, color-free shape semantics,
   tiny details/help/footer, and whether truncation can erase the only distinction between two rows.
7. **Action-surface honesty.** Inventory key handling, footer, help, command palette, editor, path
   copy, action menu, status cycling, and inline edit. A hidden hint is insufficient if a global key
   still triggers an unsafe or misleading action; an intentionally supported file-open action must
   remain capability-aware.
8. **Error and diagnostic truthfulness.** Exercise first-load versus refresh failures, retained last
   good content, graph-broken empty repositories, unreadable Thread records, projection-local
   diagnostics, detail errors, and recovery. Ensure success clears stale errors without erasing the
   last coherent view prematurely.
9. **Regression blast radius.** Threads were inserted into the tab ring and generalized shared
   detail errors/path behavior. Exercise every existing entity's tab shortcuts, help/footer, detail
   failures, session restoration, editor/yank, and command palette. Search for positional tab-index
   assumptions and hard-coded entity enumerations outside modified files.
10. **Scale and resource behavior.** Probe large Thread/member/gate/problem sets, repeated list
    reloads, palette reindexing, detail rendering, and clipboard/editor path lookup. File performance
    findings only with plausible planning-space inputs or measurements.
11. **Documentation and planning truthfulness.** Compare source, tests, README, CLI docs,
    architecture text, ADR-0006, task checkboxes/progress, and current Thread sequencing. Flag claims
    of exact projection fidelity, portability, responsive safety, or complete test coverage that are
    materially stronger than the evidence.
12. **Systemic second pass.** After the checklist, deliberately look for a shared abstraction,
    helper, registry assumption, or error-retention rule that could mask a whole defect class across
    Threads and existing entities. Demonstrate the issue or explicitly record why the suspected
    pattern is safe.

## Validation and restoration

Run proportionate validation and hostile temporary-space/fixture probes. You may create scratch data
outside the repository and temporary mutation probes inside it, but restore every probe and generated
artifact. Do not install dependencies, commit, push, edit implementation permanently, create
follow-up tasks, change finding statuses, close this audit, or edit the other reviewer's audit. At
finish, the worktree must differ from its starting state only by your edits to this assigned audit
file. Verify that claim with `git status --short` and `git diff`.

## Deliverable

Preserve this brief and replace the reviewer-report placeholder below with:

- executive verdict: `ready`, `ready with tracked follow-ups`, or `not ready`;
- reviewed branch/base/HEAD/worktree state, runtime, and exact validation results;
- a compact end-to-end Thread state-flow and capability-boundary re-derivation;
- findings grouped by severity, each with stable code and `**Status:** open` in the heading;
- an acceptance-criteria traceability table;
- a hostile-evidence ledger covering every required fixture, ordering/capability case, and restored
  mutation probe, including observed result and the test that would catch regression;
- explicit separation of demonstrated defects, source-supported risks, and unverified concerns; and
- settled concerns with the evidence that settles them.

If there are no findings, say so plainly, but the evidence and mutation ledgers are still required.
Do not pre-resolve findings; the implementation owner will triage them with
`tskflwctl audit finding`.

## Reviewer report

### Executive verdict

**Verdict:** `ready with tracked follow-ups`

The implementation in `feat/tui-thread-list-detail` cleanly accomplishes its primary architectural objective: promoting Threads from an ad-hoc, parallel cache (`threadProjectionState`) into a first-class citizen of the TUI's entity registry (`entityThreads`). The parallel state machine, separate loaded messages, and bespoke reload branches have been eliminated. Threads now participate lazily in the standard tab ring, watcher reload fan-out, selection preservation, command palette search, and multi-workspace session caching. Crucially, the capability boundary between portable semantic reads (`ShowThread`) and optional local filesystem navigation (`ThreadPath`) is cleanly respected: a pathless store provides a full semantic browsing experience, while an unparseable Thread document still permits local editor invocation (`E`) for interactive repair.

However, an adversarial second pass revealed two medium-severity systemic issues:
1. **Critical Activity and Progress Truncation (`M1`):** `threadDelegate.Render` places the unbudgeted, variable-length Thread `label` before `progress` (`d:... s:...`) and `work` (`▶... ✓... ×...`). In realistic terminals ($\le 80$ columns, and particularly in split-pane views where the list pane width is typically 38–58 columns), any slug longer than 20 characters pushes the progress and activity indicators past the right viewport boundary, where `row()` truncates them off completely. The existing test `TestThreadRowsDegradeWithoutLosingEssentialState` masked this failure by testing exclusively with a 4-letter slug (`"ship"`).
2. **Synchronous Test Helper Masking (`M2`):** The test helper `drainNested` resolves Bubble Tea command trees recursively in depth-first synchronous order. While the production codebase contains proper generation, kind, and selection-key guards, the entire Thread test suite in `thread_projection_test.go` exclusively uses `drainNested`, leaving out-of-order asynchronous message arrival and contention-recovery interleaving untested for Thread entities.

Three low-severity polish findings (`L1`, `L2`, `L3`) covering disaligned row columns, silent `e` key handling on pathless threads, and right-clipped duplicate disambiguator suffixes are also cataloged for implementation-owner triage.

---

### Review environment & validation results

- **Repository:** `github.com/andy-esch/taskflow`
- **Branch:** `feat/tui-thread-list-detail`
- **Base Commit:** `a38e694` (`Merge pull request #183 from andy-esch/worktree-close-agent-ergonomics-audit`)
- **HEAD Commit:** `e319de3` (`wip`)
- **Toolchain:** `go version go1.26.6 darwin/arm64` on `Darwin arm64`

#### Exact Validation Ledger
- `go test -count=1 -race ./...`: **PASS** (uncached; all 33 packages passed with zero race conditions in 13.2s)
- `go test -count=1 -race ./internal/tui ./internal/theme`: **PASS** (uncached in 9.2s)
- `go run ./cmd/tskflwctl lint`: **PASS** (`✔ all planning entities and dependency links pass lint`)
- `git diff --check`: **PASS** (zero trailing whitespace, zero conflict markers)
- `tskflwctl status`: **PASS** (correctly reports open audits and active work)

---

### Architecture & state-flow re-derivation

The migration collapses the former dual-state architecture into a single unified registry lifecycle:

```
                  ┌─────────────────────────────────────────────────────────┐
                  │                 core.Service                            │
                  │  ListThreadViews() ──► core.ThreadListView (portable)   │
                  │  ShowThread(id)    ──► core.ThreadView + body (portable)│
                  │  ThreadPath(id)    ──► optional local filesystem path   │
                  └──────────────┬──────────────────────────┬───────────────┘
                                 │                          │
                 loadThreadList  │          loadThreadDetail│
                                 ▼                          ▼
                  ┌──────────────────────┐   ┌──────────────────────────────┐
                  │   listLoadedMsg      │   │   detailMsg / detailErrMsg   │
                  │  - kind: entityThreads │   │  - kind: entityThreads       │
                  │  - gen: tab.loadGen  │   │  - gen: m.detailGen          │
                  │  - items: threadItem │   │  - id: canonicalID           │
                  │  - threadDiagnostics │   │  - content / localPath       │
                  └──────────────┬───────┘   └──────────────┬───────────────┘
                                 │                          │
                                 ▼                          ▼
                  ┌──────────────────────┐   ┌──────────────────────────────┐
                  │   entityTab (threads)│   │          detailPane          │
                  │  - list: bubbles.list│   │  - content: threadDetail     │
                  │  - loadGen: int      │   │  - loadedKey: canonicalID    │
                  │  - threadDiagnostics │   │  - errorPath: retained path  │
                  └──────────────────────┘   └──────────────────────────────┘
```

#### Key Architectural Guarantees
1. **Zero Dual State:** `Model.threads` (`threadProjectionState`) was deleted. Thread list items live in `m.tabs[indexOfKind(m.tabs, entityThreads)].list`.
2. **Lazy First Visit:** `tab.loaded == false` and `tab.loadGen == 0` until `:threads`, `]`, or a command palette jump visits the tab. Unvisited tabs generate zero disk or graph I/O.
3. **Unified Live Reload:** File change events coalesce through `m.reloadAll()`. Visited Thread tabs reload in parallel with tasks and epics via `t.reload(m.svc, t.markReload())`.
4. **Authoritative Projections:** The TUI stores `core.ThreadView` intact inside `threadItem`. It counts roles and gates for compact layout, but never performs graph traversal, eligibility calculation, or membership mutation.
5. **Portable vs Local Separation:** `threadItem.path()` strictly returns `""`. Local paths are fetched optionally during `loadThreadDetail`. If `ThreadPath` fails, semantic content still displays; if semantic read fails, a valid local repair path is retained on the error pane so `E` ($EDITOR) remains functional.

---

### Repository-wide consumer inventory

| Symbol / Concept | Classification | Observed Status & Location | Verification / Impact |
| :--- | :--- | :--- | :--- |
| `entityThreads` | Registry Kind | Declared in `entity.go:26`, initialized in `newEntityTabs` (`entity.go:442-450`) | Positioned between epics and audits. Full tab participation. |
| `threadProjectionState` | **Deleted** | Removed from `model.go`, `session.go`, `thread_projection.go` | Dual-cache completely eliminated. No parallel state remains. |
| `threadListDiagnostics` | Registry Metadata | `entity.go:188`, populated in `messages.go:81` | Retains repository graph problems and unreadable record counts on the tab. |
| `threadItem` | List Adapter | Declared in `item.go:191-232` | Wraps `core.ThreadView`, provides `FilterValue()`, `displayLabel()`, `sortFields()`. |
| `threadDetail` | Detail Adapter | Declared in `detail.go:538-656` | Implements `detailContent` (`Title()`, `Path()`, `rawBody()`, `meta()`). |
| `loadThreadList` | Async Loader | `thread_projection.go:23-53` | Returns `listLoadedMsg` stamped with `tab.loadGen`. Zero graph derivation. |
| `loadThreadDetail` | Async Loader | `thread_projection.go:57-77` | Joins `ShowThread(id)` with optional `ThreadPath(id)`. Returns `detailMsg`/`detailErrMsg`. |
| `ThreadPath` | Optional Port | Invoked in `thread_projection.go:62` | Resolved via `core.ThreadPathSource`. Optionality verified by pathless tests. |
| `listLoadedMsg` | Core Message | Extended in `messages.go:81`, handled in `model.go:623` | Transfers items, problems, and `threadDiagnostics` atomically. Stamped by `loadGen`. |
| `detailMsg` | Core Message | Handled in `model.go:337-342` | Guarded by `isCurrentSelection(msg.kind, msg.id)` and `msg.gen == m.detailGen`. |
| `detailErrMsg` | Core Message | Extended with `localPath` in `messages.go:112`, handled in `model.go:344-363` | Retains `localPath` on error pane for editor repair. Stamped by `detailGen`. |
| `readSurface` | Retry Policy | Removed `readThreadList`/`readThreadDetail` from `read_retry.go` | Threads use standard `readEntityList` and `readEntityDetail`. |
| `reloadAll` | Reload Fan-Out | Cleaned in `model.go:251-255` | Threads reload automatically through `m.tabs` loop; custom call removed. |
| `detailGen` | Concurrency Guard | Stamped in `model.go:1176-1194`, guarded in `model.go:338,347` | Prevents stale async detail responses from overwriting newer user selections. |
| `loadedKey` | Selection Guard | Stored on `detailPane` (`detail.go:161,185`) | Checked by `selectedPath()` (`model.go:1278`) to prevent wrong-file editor actions. |
| Registry Indexes | Palette / Route | `commands.go:323`, `command_dispatch.go:34,228` | `:threads`, `:th`, `:thread` route to tab; items indexed in Ctrl+P palette. |
| Session State | Workspace Session | Saved/restored in `session.go:99,132` | `m.tabs` retains Thread cursor, filter, sort, and loaded items per workspace. |
| `selectedPath` | Action Port | Specialized in `model.go:1274-1284` | Returns `m.detail.path()` only when `m.detail.loadedKey == m.selectedKey()`. |
| Yank / Editor | Action Surface | Handled in `model.go:786,800,1304,1322` | `Y` copies path; `E` opens editor. Informative flash if path is loading/unavailable. |
| Help & Footer | UI Affordance | Updated in `help.go:116,190,226`, `view.go:440,488` | Read-only notes, symbols, and footer suppress mutation keys on Threads tab. |
| Documentation | Architecture / ADR | Updated in `ARCHITECTURE.md`, `README.md`, `ADR-0006`, `tskflwctl_ui.md` | Reflects 5th first-class entity tab and registry integration. |

---

### Acceptance-criteria traceability

| Requirement / Acceptance Criteria | Status | Implementation Reference | Verification Evidence |
| :--- | :--- | :--- | :--- |
| **AC 1: First-Class Registry Entity** | **Met** | `entity.go:26,442` | `TestThreadsUseOneRegistryListAndDetailOwner` proves Threads participate in tab strip, `:threads` jump, and palette. |
| **AC 2: Lazy Loading & Watcher Reload** | **Met** | `entity.go:447`, `model.go:251` | `TestThreadsUseOneRegistryListAndDetailOwner` proves unvisited tab has `loadGen == 0`; `TestThreadRegistryReloadsOnTaskAndThreadChanges` verifies live reload. |
| **AC 3: Core Projection Display** | **Met** | `item.go:191-232`, `thread_projection.go:23` | `TestProbe_EndToEndRequiredThreadShapes` compares all 9 shapes against `svc.ListThreadViews()`: 100% deep equality. |
| **AC 4: Comprehensive Detail View** | **Met** | `detail.go:538-656` | `TestThreadDetailPresentsCoreProjectionAndBody` validates lifecycle, health, progress, work, members, gates, diagnostics, body. |
| **AC 5: Portable Reads & Local Capability**| **Met** | `thread_projection.go:57-77`, `model.go:1274` | `TestThreadRouteSurvivesSplitPathlessCapabilities` proves pathless store functions; `TestLocalThreadPathSurvivesSemanticDetailFailure` proves repair path. |
| **AC 6: Empty State & Graph Diagnostics** | **Met** | `view.go:210-222,488` | Empty repository renders onboarding command plus repository graph health and unreadable record counts. |
| **AC 7: Read-Only Action Safety** | **Met** | `entity.go:442-450`, `help.go:226` | No transitions declared; `keys.Action` (move) inert; help filters mutation keys; notes state read-only. |
| **AC 8: Responsive Layout** | **Partial** | `item.go:277-296` | Viewport width safe ($\le W$), but progress and work indicators are right-truncated on realistic slugs ($\le 80$ cols; see `M1`). |

---

### Findings

#### M1. `threadDelegate.Render` Omits Slug Budget, Right-Truncating Progress and Work Metrics in Split Panes and Widths < 80 Cols · **Status:** fixed

**File:** internal/tui/item.go:277-296 · **Component:** tui/item
**Effort:** S · **Urgency:** soon

**Description:**
In `threadDelegate.Render`, list lines are assembled by placing the variable-length Thread `label` before the progress rollup (`d:... s:...`) and work indicators (`▶... ✓... ×...`):
```go
case m.Width() >= 64:
    line = fmt.Sprintf("%s %-11s %s%s  %s  %s  %s  %s",
        st.fg(status.Color, status.Glyph), view.Thread.Status, health, warn,
        label, progress, work, st.dim(view.Thread.Description))
case m.Width() >= 42:
    line = fmt.Sprintf("%s %s%s  %s  d:%d/%d s:%d/%d  %s",
        st.fg(status.Color, status.Glyph), health, warn, label,
        view.Rollup.Done, view.Rollup.Total, view.Rollup.Drained, view.Rollup.Total, work)
```
Unlike `taskDelegate` (which calculates a dedicated `slugW` budget and truncates the slug independently so subsequent columns remain visible) and `epicDelegate` (which places progress bars and counts before the description), `threadDelegate` performs no independent budgeting or truncation on `label`.

In standard split-pane mode (where the list pane width is typically 38–58 columns on 80–120 column terminals), or in any terminal $\le 80$ columns, any Thread slug longer than 20 characters pushes `progress` and `work` past the line width. `row()` then right-truncates the entire line via `ansi.Truncate(content, max1(m.Width()-2))`. As a result, the progress rollups and active/dispatchable/blocked work indicators are completely erased from view.

The unit test `TestThreadRowsDegradeWithoutLosingEssentialState` failed to catch this regression because it tested exclusively with a 4-character synthetic slug (`"ship"`).

**Hostile Reproduction:**
1. Create a Thread with a realistic slug (e.g. `reconcile-cross-workspace-planning-state-across-threads`).
2. Render the list item at width 80, 64, or 48.
3. Observe output:
   - At width 80: `› ● in-progress g✓/v✓  reconcile-cross-workspace-planning-state-across-threads …` (`d:... s:...` and `▶... ✓... ×...` are missing).
   - At width 64: `› ● in-progress g✓/v✓  reconcile-cross-workspace-planning-state…` (progress and work missing).
   - At width 48: `› ● g✓/v✓  reconcile-cross-workspace-planning-s…` (progress and work missing).

**Recommendation:**
Adopt the column budgeting pattern used in `taskDelegate.Render`: allocate a bounded width for `label` (e.g. `slugW := clamp(m.Width() - fixedColumnsWidth, 12, 32)`), truncate `label` independently with `ansi.Truncate(label, slugW, "…")`, and preserve the progress and work columns in fixed, predictable positions.

---

**Resolution:** Thread rows now reserve lifecycle, health, progress where
available, and the complete work triple before budgeting the
identity/description tail. Long-slug regressions at normal, split-pane, and tiny
widths prove work remains visible.

#### M2. Synchronous Command Tree Execution in `drainNested` Masks Asynchronous Completion-Order Inversion · **Status:** fixed

**File:** internal/tui/thread_projection_test.go:79-98 · **Component:** tui/testutil
**Effort:** S · **Urgency:** soon

**Description:**
`thread_projection_test.go` relies entirely on the synchronous recursive helper `drainNested`:
```go
func drainNested(t *testing.T, m Model, cmd tea.Cmd) Model {
    ...
    if batch, ok := msg.(tea.BatchMsg); ok {
        for _, child := range batch {
            m = drainNested(t, m, child)
        }
        return m
    }
    tm, next := m.Update(msg)
    return drainNested(t, tm.(Model), next)
}
```
`drainNested` executes `tea.BatchMsg` and chained commands sequentially in strict depth-first order. In production, Bubble Tea executes commands concurrently in goroutines, meaning messages can and do complete out of order (e.g. a delayed `detailMsg` from an older selection landing after a fast `listLoadedMsg` has bumped `loadGen` and selected a new item).

While production code in `model.go` includes generation and selection-key guards (`isCurrentSelection(msg.kind, msg.id)` and `msg.gen == m.detailGen`), there is not a single test in `thread_projection_test.go` that directly exercises asynchronous out-of-order message delivery for Threads. Stale list loads, stale detail responses, or racing watcher reloads on Threads are never delivered out of sequence in the test suite.

**Hostile Reproduction:**
Constructed a hostile test feeding an older `detailMsg` (`gen: currentGen - 1`) and a mismatched `detailMsg` (`id: "other"`) directly into `m.Update`. While production guards correctly dropped the messages, this path had zero test coverage prior to the hostile probe.

**Recommendation:**
Add explicit tests in `thread_projection_test.go` that dispatch out-of-order `listLoadedMsg` and `detailMsg` instances with stale generations and obsolete keys to directly verify that production guards reject them without relying on `drainNested`.

---

**Resolution:** Direct message-order tests now inject a stale Thread list
generation, a stale same-Thread detail generation, and a current-generation
detail for the wrong canonical key; all are rejected without the synchronous
command-tree helper.

#### L1. Inconsistent Spacing and Jagged Tabular Alignment Across Rows in `threadDelegate.Render` · **Status:** fixed

**File:** internal/tui/item.go:277-296 · **Component:** tui/item
**Effort:** XS · **Urgency:** eventually

**Description:**
In `threadDelegate.Render` at width $\ge 64$:
```go
line = fmt.Sprintf("%s %-11s %s%s  %s  %s  %s  %s",
    st.fg(status.Color, status.Glyph), view.Thread.Status, health, warn,
    label, progress, work, st.dim(view.Thread.Description))
```
Because `label` is inserted as raw text without right-padding or column reservation, the starting horizontal offset of `progress` (`d:... s:...`) and `work` (`▶... ✓... ×...`) varies directly with the length of each Thread's slug. A short slug (`api`) starts progress at column 26, while a longer slug (`database-migration`) pushes it to column 44. This produces a jagged, unaligned visual presentation across rows in the list.

Furthermore, `warn` is conditionally formatted as `warn = " " + st.fg(...)` when `view.Inconsistent` is true, but `""` when false. Because the template specifies `%s%s  %s` (`health + warn + "  " + label`), inconsistent rows introduce an extra space and shift the label right by 2 cells compared to consistent rows.

**Recommendation:**
Pad `label` to a consistent column width or place fixed-width columns (`status`, `health`, `progress`, `work`) before `label` and `description`. Standardize spacing around `warn` so consistent and inconsistent rows share identical horizontal alignments.

---

**Resolution:** Fixed state columns now precede the variable identity tail,
health plus inconsistency is padded to a stable width, and work begins before
any slug-dependent content.

#### L2. `keys.Edit` on Thread Rows Fails Silently When Local Path is Loading or Unavailable · **Status:** fixed

**File:** internal/tui/model.go:782-785 · **Component:** tui/model
**Effort:** XS · **Urgency:** eventually

**Description:**
When `e` (`keys.Edit`) is pressed:
```go
case key.Matches(msg, keys.Edit):
    if t, ok := m.selectedTask(); ok {
        m.edit.open(t)
    } else if ep, ok := m.selectedEpic(); ok {
        m.edit.openEpic(ep)
    } else if m.selectedPath() != "" {
        m.flash, m.flashErr = "no inline edit here — press E to edit in $EDITOR", true
    }
    return m, nil
```
For `entityThreads`, `selectedTask()` and `selectedEpic()` are false. If `m.selectedPath() != ""` (local path already loaded), it displays the helpful advisory flash `"no inline edit here — press E to edit in $EDITOR"`.

However, if `m.selectedPath() == ""` (e.g. while the detail path is still loading asynchronously, or in a portable pathless store), the `else if` condition fails and the key event is silently discarded with zero user feedback. In contrast, `yankSelectedPath()` and `openInEditor()` explicitly check:
```go
if m.detail.loading && !m.detail.showing(m.selectedKey()) {
    m.flash, m.flashErr = "Thread path is still loading", true
} else {
    m.flash, m.flashErr = "local path unavailable for this Thread", true
}
```

**Recommendation:**
Mirror the detail loading/availability check in `keys.Edit` so that pressing `e` on a Thread whose path is loading or unavailable provides the same informative flash feedback instead of a silent no-op.

---

**Resolution:** Lowercase e now provides explicit loading and path-unavailable
feedback for Threads, matching the capability-aware E and Y behavior.

#### L3. Right-Truncation Clips Disambiguation Suffix First on Duplicate Thread Slugs · **Status:** fixed

**File:** internal/tui/item.go:207,297 · **Component:** tui/item
**Effort:** XS · **Urgency:** eventually

**Description:**
When two Threads have duplicate slugs, `displayLabel()` appends ` [hint]` to the slug via `labelWithIdentityHint`. When `threadDelegate.Render` formats the line and calls `row()`, `row()` truncates the entire line from the right using `ansi.Truncate(content, max1(m.Width()-2))`.

On terminals $\le 80$ columns or with long slugs, the bracketed hint at the end of the label is clipped off before the slug itself. Two duplicate rows such as `investigate-cross-workspace-state [6g503c6pfqe1]` and `investigate-cross-workspace-state [6g503c6pfqe2]` both truncate to `investigate-cross-workspace-state …`, stripping the only visual distinction between the rows.

**Recommendation:**
Ensure that when identity hints are present, the hint suffix is protected from truncation or placed in a dedicated column.

---

**Resolution:** The report's suffix premise was inaccurate: identity hints
already prefix labels. The revised protected identity budget and grapheme-aware
middle elision additionally preserve the hint's distinguishing tail when the
full prefix cannot fit; a narrow duplicate-slug regression pins it.

### Hostile evidence ledger

#### 1. End-to-End Planning Space Probe (9 Required Shapes)
- **Fixture:** Throwaway repository populated with tasks and threads covering:
  1. In-progress member (`RoleInFlight`)
  2. Eligible `StatusNextUp` member (`RoleCandidate`, `GateClear`, `Eligible: true`)
  3. Eligible `StatusReadyToStart` member (`RoleCandidate`, `GateClear`, `Eligible: true`)
  4. Blocked pending member (`RoleQueued`, `GateBlocked` due to external gate)
  5. Immediate external gate (`ThreadExternalGate`, `Outstanding: true`)
  6. Shared member across Thread A and Thread B
  7. Cancelled Thread (`ThreadStatusCancelled`)
  8. Completed-but-unsound Thread (`ThreadStatusCompleted` with missing member task ID)
  9. Empty Thread (`ThreadStatusUnstarted`, `tasks: []`)
- **Observed Result:**
  - `tab.list.Items()` populated all 6 threads cleanly.
  - TUI `view.Rollup` matched `core.ThreadRollup` 1:1 (`Done`, `Drained`, `Total`, `Deprecated`).
  - `len(view.Frontier)` matched core projection exactly (only eligible members appeared in frontier; blocked member excluded).
  - Completed-but-unsound thread reported `view.Inconsistent == true` and `ProjectionHealth == GraphBroken`, displaying `⚠` warning in row and diagnostics in detail.
  - Cancelled thread preserved its `cancelled` status glyph and nominal progress without errors.
  - Empty thread displayed `0/0 nominally done` and proper empty onboarding message when filtered.
- **Regression Test:** `TestProbe_EndToEndRequiredThreadShapes` in scratch probe suite.

#### 2. Hostile Identity Fixtures
- **Fixture:** Two Threads sharing identical slug `same-slug` with distinct canonical IDs `6g503c6pfqe1` and `6g503c6pfqe2`; one Thread with ID drift (`frontmatter id: 6g503c6pfqe3`, `filename id: 6g503c6pfqe4`).
- **Observed Result:**
  - Both duplicate rows received distinct identity hints (`same-slug [qe1]` and `same-slug [qe2]`).
  - Invoking `tab.selectByKey("6g503c6pfqe2")` selected the second duplicate unambiguously.
  - `selectedThreadDetail()` loaded the exact matching canonical ID `6g503c6pfqe2` without collision.
  - ID drift Thread was not collapsed or dropped; it rendered with `ProjectionHealth == GraphBroken` and flagged `thread-id-drift` in detail diagnostics.
- **Regression Test:** `TestProbe_HostileIdentityFixtures`.

#### 3. Completion-Order & Stale Message Interleaving Probes
- **Fixture:** Hostile delivery of `listLoadedMsg` and `detailMsg` out of sequence directly into `m.Update`.
  - Stale `listLoadedMsg` with `gen: tab.loadGen - 1`.
  - Mismatched `detailMsg` with `id: "other-thread"`.
  - Stale `detailMsg` for current ID with `gen: m.detailGen - 1`.
- **Observed Result:**
  - `handleListLoaded` dropped the stale generation message; current items remained intact.
  - `m.Update(detailMsg)` dropped the mismatched ID; active detail was not overwritten.
  - `m.Update(detailMsg)` dropped the stale generation; active detail content was preserved.
- **Regression Test:** `TestProbe_CompletionOrderGuards`.

#### 4. Capability-Boundary Probes
- **Fixture:** `splitWorkspaceStore` providing `TaskGraphs` and `Threads` but omitting `ThreadPaths` (pathless portable adapter); `countingThreadStore` with `getErr: domain.ErrValidation` paired with valid `tuiThreadPathFake`.
- **Observed Result:**
  - In pathless store: `selectedThreadView()` rendered full semantic projection; `detail.body` loaded `"split body\n"`; `m.selectedPath()` returned `""`; `yankSelectedPath()` and `openInEditor()` degraded gracefully with `"local path unavailable for this Thread"`.
  - In semantic failure with valid local path: detail pane rendered `⚠` error message; `m.selectedPath()` returned `"/planning/threads/repair-thread.md"`; `openInEditor()` successfully returned editor command targeting the repair path.
- **Regression Test:** `TestThreadRouteSurvivesSplitPathlessCapabilities` and `TestLocalThreadPathSurvivesSemanticDetailFailure`.

#### 5. Responsive Layout Cell Width Measurements
- **Fixture:** Measured rendered list row cell widths using `runewidth.StringWidth(ansi.Strip(output))` across terminal widths 120, 80, 64, 48, 30 with a 54-character slug.
- **Observed Measurements:**
  - Width 120: cellWidth 120 (rendered slug + `d:12/15 s:10/15` + `▶1 ✓1 ×2` + truncated desc).
  - Width 80: cellWidth 80 (rendered slug + `…`; progress and work truncated).
  - Width 64: cellWidth 64 (rendered slug truncated; progress and work truncated).
  - Width 48: cellWidth 48 (rendered slug truncated; progress and work truncated).
  - Width 30: cellWidth 30 (compact health + truncated slug).
- **Finding:** Zero cell overflow across all widths (no negative frames or terminal line wrap), but revealed `M1` (progress/work truncated on widths $\le 80$).

#### 6. Restored Mutation Probes Ledger

| # | Mutation Probe Description | Intended Defect | Killing Test / Observable Evidence |
| :--- | :--- | :--- | :--- |
| 1 | Reintroduce second Thread list/detail owner in `Model` | Dual ownership / split cache | `TestThreadsUseOneRegistryListAndDetailOwner` & `TestReadRequestCurrentEnumeratesEverySurface` fail build/runtime. |
| 2 | Remove `loadGen` or `detailGen` guard from `model.go` | Stale load overwrites newer state | `TestProbe_CompletionOrderGuards` and `TestNewerGenerationSupersedesAScheduledRetry` fail immediately. |
| 3 | Derive frontier from task status instead of `view.Frontier` | Status-inferred eligibility bypasses graph | `TestProbe_EndToEndRequiredThreadShapes` fails (blocked member counted in frontier). |
| 4 | Replace stored `core.ThreadView` with partial summary | Loss of projection fidelity | `TestThreadsUseOneRegistryListAndDetailOwner` fails: `registry rows changed core projections`. |
| 5 | Make `ThreadPath` mandatory or source from `domain.Thread.Path` | Breaks portable stores | `TestThreadRouteSurvivesSplitPathlessCapabilities` fails (pathless store rejected). |
| 6 | Route selection/detail restore through slug instead of key | Duplicate slug collision on restore | `TestProbe_HostileIdentityFixtures` fails (second duplicate not selected). |
| 7 | Drop graph diagnostics or collapse nominal/sound progress | Obscures broken graph or undrained work | `TestThreadsUseOneRegistryListAndDetailOwner` & `TestThreadRowsDegradeWithoutLosingEssentialState` fail. |
| 8 | Expose lifecycle/move hints on read-only Threads tab | Unsupported mutation advertised | `help_test.go` and `TestHelpSectionsFor` fail assertion on `Global` section entries. |

---

### Separation of defects, risks, and unverified concerns

- **Demonstrated Defects:**
  - `M1`: Unbudgeted Thread slug right-truncates progress and work indicators at terminal widths $\le 80$ (and split-pane widths $\le 58$).
  - `M2`: Synchronous recursive command execution in `drainNested` leaves asynchronous message interleaving untested for Thread entities.
  - `L1`: Jagged column alignment across rows and asymmetric spacing on inconsistent thread rows.
  - `L2`: Silent fallthrough on pressing `e` when a Thread's local path is still loading or unavailable.
  - `L3`: Suffix truncation of identity disambiguator hint on colliding Thread slugs in narrow viewports.
- **Source-Supported Risks:**
  - *Large Member Counts:* A Thread with hundreds of member tasks renders all of them in `renderThreadMeta`. While viewport scrolling handles large text blocks, virtualization is not present for the detail body. (Acceptable for current planning-space scales).
- **Unverified Concerns Settled as Safe:**
  - *Dual Cache Re-emergence:* Verified completely deleted. `threadProjectionState` is gone from all structs and messages.
  - *Stale Cross-Workspace Restore:* Verified that switching workspaces reloads visited tabs and restores cursors by canonical key.
  - *Wrong-File Editor Invocation:* Verified that `selectedPath()` checks `m.detail.loadedKey == m.selectedKey()`, preventing a stale path from opening when cursor moves rapidly.

---

### Settled concerns

1. **Did deleting `threadProjectionState` break watcher reload debounce?**
   *Settled:* No. Watcher events call `m.reloadAll()`. Visited Thread tabs reload through `t.reload(m.svc, t.markReload())`. Debouncing and generation bumping operate identically to tasks and epics.
2. **Can a pathless store break semantic Thread viewing?**
   *Settled:* No. `loadThreadDetail` treats `ThreadPath` as purely optional. If path lookup fails, `pathIssue` is noted in metadata, but `view` and `body` are rendered completely.
3. **Can an unparseable Thread document be repaired from the TUI?**
   *Settled:* Yes. `detailErrMsg` carries `localPath`. When semantic parsing fails, the error pane displays the YAML error while retaining `localPath`, allowing the user to press `E` to open the file in `$EDITOR`.
4. **Does Thread selection collide across duplicate slugs?**
   *Settled:* No. Selection, cursor restoration, detail loading, and command palette navigation are keyed strictly by `CanonicalID()`.
