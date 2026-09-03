---
schema: 1
id: 6g6db7yt62r4
bucket: closed
area: tui-thread-list-detail-implementation-claude
date: "2026-09-03"
updated_at: "2026-09-03"
---
# Audit: TUI Thread list and detail implementation — claude — 2026-09-03

> Reviewer assignment: claude. This document is the review brief and the only file the reviewer should update.
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

**ready with tracked follow-ups.**

The central architectural claim of this change holds under hostile probing. Threads really are an
ordinary registry entity: one list/detail owner, one generation family, canonical-ID identity, the
shared contention policy, and no surviving parallel cache. I tried to break each of those and could
not — five of eight required mutation probes were killed, several by more than one test.

Two demonstrated presentation defects remain, both in the compact `▶ ✓ ×` work triple and its
narrow-width rendering, and three required mutations survived, meaning the suite does not currently
defend rules the change depends on. None is data-corrupting; all are triageable follow-ups rather
than blockers. No finding requires reverting the design.

### Reviewed state

| | |
|---|---|
| Branch | `feat/tui-thread-list-detail` |
| HEAD | `e319de3` ("wip") |
| Base | `main` @ `a38e694` |
| Worktree at start | clean except the two untracked 2026-09-03 audit files |
| Worktree at finish | identical, plus my edits to this file only (`git status --short`, `git diff` verified empty) |
| Runtime | go1.26.6 darwin/arm64 |
| Diff scope | 25 files, +945 −440 |

**Validation, exact commands and results**

| Command | Result |
|---|---|
| `just build` | ok |
| `go test -race -count=1 ./...` (uncached) | all 24 packages ok, 14.9 s wall |
| `go test -race -count=1 ./internal/tui` (post-restore) | ok, 8.7 s |
| `just lint` (`golangci-lint run ./...`) | 0 issues |
| `just docs-check` | no drift |
| `just tidy-check` (`go mod tidy -diff`) | clean |
| `git diff --check` | clean |
| `tskflwctl lint` | all planning entities pass |
| `tskflwctl audit lint` | all findings pass |

No pre-existing unrelated failures were observed, so nothing needed separating from regressions.

### State-flow and capability re-derivation

Re-derived from production code, not comments:

**List.** `entityThreads` is a registry entry with `loadList: loadThreadList`. `loadThreadList`
captures `t.loadGen` at command-build time, calls `svc.ListThreadViews()`, and returns a
`listLoadedMsg` carrying complete `core.ThreadView` rows plus a `threadListDiagnostics` side-car.
`handleListLoaded` drops the message on `msg.gen != tab.loadGen`, then runs `validateEntityItems`
before installing rows, so an adapter emitting empty or duplicate canonical keys fails visibly
instead of collapsing rows.

**Detail.** `refreshDetail` → `loadDetail(key)` bumps `m.detailGen`, calls the registry's
`loadItem` (`loadThreadDetail`), and centrally stamps `gen` and `label` onto both `detailMsg` and
`detailErrMsg`. The whole thing is wrapped in `withReadConflictRetry` at `readEntityDetail` — the
same surface tasks/epics/audits use. `readThreadList`/`readThreadDetail` are gone from `readSurface`.

**Capability split.** `loadThreadDetail` resolves `svc.ThreadPath(id)` and `svc.ShowThread(id)`
independently. A path failure becomes an explanatory `pathIssue` field and never blanks semantic
content; a semantic failure still carries `localPath` onto `detailErrMsg`, which `SetError` retains
as `errorPath`. `threadItem.path()` returns `""` unconditionally, so the portable record is never
mistaken for the path port.

**Action surface.** Threads declare no `transitions` and no `applyMove`, so the footer (`view.go`),
`:` dispatch, and completion all omit lifecycle verbs by construction rather than by a hidden hint.
`selectedPath()` returns a Thread path only when `m.detail.loadedKey == m.selectedKey()`, so
`Y`/`E` are capability-aware and cannot act on a path belonging to another Thread.

### Findings

#### M1. The row's blocked count is not "not dispatchable", and silently loses members on an unhealthy graph  · **Status:** fixed

**File:** internal/tui/item.go (activityForThread, threadDelegate.Render); internal/tui/detail.go (renderThreadMeta "work") | **Component:** projection fidelity
**Effort:** S · **Urgency:** soon

`activityForThread` counts a member as blocked only when its role is queued/candidate **and**
`State.Gate != GateClear`. Dispatchable comes from `len(view.Frontier)`. Those two predicates do not
partition pending work, because eligibility is graph-derived: `Eligible` additionally requires the
whole repository graph to be healthy. On a degraded or broken graph a member can be pending with a
*clear* gate and still absent from `Frontier` — counted as neither dispatchable nor blocked, and so
invisible in the row.

Demonstrated on a throwaway space with one legacy `blocked_by` field (which degrades the graph) and
a three-member Thread:

```
graph=degraded projection=broken members=3 rollup=0/3
row triple: inFlight=1 dispatchable(len Frontier)=0 blocked=0
pending members with a CLEAR gate that are NOT in Frontier: 2
accounted=1 of 3 members  => UNACCOUNTED=2
   member cleara  role="queued"    gate="clear" eligible=false
   member clearb  role="candidate" gate="clear" eligible=false
   member inflight role="in-flight" gate="clear" eligible=false
```

The row renders `▶1 ✓0 ×0` for a Thread with two pending members. A reader reasonably concludes
"one item active, nothing else waiting". The `d:0/3` rollup and the `g~/v!` health mark are the only
contradicting signals, and neither says where the two members went. The detail's `work` line repeats
the same triple, though its "Members (persisted order)" section does list all three, so detail is
recoverable and the list row is not.

Note the CLI makes no equivalent claim — `thread list` shows `0 eligible` and `degraded/degraded`
without asserting a blocked count — so this is a claim introduced by this change.

**Recommendation:** Either label the third count for what it measures (gate-blocked) and add the
remainder, or derive it as `pending − dispatchable` so the triple accounts for every pending member.
A row that cannot partition its members should say so rather than reporting zeros.

**Resolution:** Pending work now partitions against the authoritative core
frontier: pending roles minus len(view.Frontier) are reported as not
dispatchable, including clear-gated members suppressed by unhealthy repository
evidence. A degraded-graph regression proves all pending members remain
accounted for.

#### M2. Narrow rows can drop the work triple and render two different Threads identically  · **Status:** fixed

**File:** internal/tui/item.go (threadDelegate.Render) | **Component:** responsive presentation
**Effort:** S · **Urgency:** soon

Every width branch builds a line and lets the shared row clipper truncate it. Because the label sits
*before* the work triple in the `<64` branches, a long Thread slug consumes the budget and the
`▶ ✓ ×` distinction disappears entirely. Two Threads whose slugs share a long prefix then render as
the same bytes:

```
w= 80 identical=false
   A: › ● in-progress g~/v!  migrate-configuration-subsystem-phase-one  d:0/3 s:0/3  …
   B: › ● in-progress g~/v!  migrate-configuration-subsystem-phase-two  d:0/3 s:0/3  …
w= 40 identical=true
   A: › ● g~/v!  migrate-configuration-subsys…
   B: › ● g~/v!  migrate-configuration-subsys…
w= 24 identical=true
   A: › ● g~/v!  migrate-conf…
   B: › ● g~/v!  migrate-conf…
```

This is the "truncation erases the only distinction between two rows" case. It also contradicts
acceptance criterion 6 as written ("Normal, narrow, and minimum supported terminal tests prove that
essential state remains readable") and the ADR-0006 text added here ("The list uses responsive
representations of … in-flight members, dispatchable frontier, and blocked pending work"), both of
which are ticked/asserted. The identity-hint mechanism does not help: it only fires on *duplicate*
slugs, and these are distinct.

Cell-width safety itself is sound — see the settled concerns; the row never exceeds the terminal
even with double-width CJK. The defect is what survives truncation, not overflow.

**Recommendation:** Give the work triple priority over the label tail at narrow widths (truncate the
label to a budget and keep the counts), or elide the label's middle so its distinguishing tail
survives.

**Resolution:** Fixed state and work columns now precede an explicitly budgeted
identity tail. Grapheme-aware middle elision preserves distinguishing slug tails
at 72, 48, and 26 cells, while the work triple remains visible.

#### M3. Nothing pins that dispatchable work comes from `view.Frontier`  · **Status:** fixed

**File:** internal/tui/item.go | **Component:** test coverage
**Effort:** XS · **Urgency:** soon

Required mutation probe: *"derive frontier from task status or count clear pending work
independently of `view.Frontier`."* I replaced `len(view.Frontier)` with a TUI-derived count of
pending members whose gate is clear, and **the entire `internal/tui` suite stayed green**.

This is the coverage counterpart of M1: the very substitution that would reintroduce a TUI-owned
eligibility rule is undetected, and on a degraded graph it produces different numbers.

**Recommendation:** Assert row/detail counts against `len(view.Frontier)` on a fixture whose
graph-derived eligibility and gate-clearness disagree — the degraded-graph fixture above is
sufficient and small.

**Resolution:** A degraded-graph regression now proves the displayed
dispatchable count equals len(view.Frontier), even when pending members have
clear local gates but global graph health suppresses eligibility.

#### M4. Nothing pins the Thread tab's read-only guarantee  · **Status:** fixed

**File:** internal/tui/entity.go (registry entry) | **Component:** test coverage / action surface
**Effort:** XS · **Urgency:** soon

Required mutation probe: *"expose lifecycle/membership mutation hints on the read-only Thread tab."*
Adding `transitions: taskTransitions, applyMove: moveTask` to the Threads registry entry left the
**entire suite green**. The mutation is behaviourally meaningful, not a dead field: `transitions`
gates the footer hint (`view.go:443`), `:`-verb dispatch (`command_dispatch.go:59,263`), and verb
completion — so a future edit could arm task lifecycle verbs on Threads, targeting a Thread's
canonical ID, with nothing failing.

Current production code is correct (the entry declares neither field). This is a coverage finding,
which the brief explicitly counts as a finding.

**Recommendation:** One assertion that the Threads registry entry offers no transitions and that the
footer/`:`-completion on that tab contains no lifecycle verb.

**Resolution:** A registry/action-surface regression now requires nil Thread
transitions and applyMove, no move/edit footer hints, no task lifecycle verb
completion, and an inert lifecycle action key.

#### L1. Nothing pins nominal-done versus soundly-drained in the detail  · **Status:** fixed

**File:** internal/tui/detail.go (renderThreadMeta progress line) | **Component:** test coverage
**Effort:** XS · **Urgency:** eventually

Required mutation probe: *"collapse nominal and sound progress."* Rendering `Rollup.Done` in place of
`Rollup.Drained` — so "0/2 soundly drained" reads as "1/2" — left the suite green. The distinction is
one the intended contract calls out explicitly, and the completed-but-unsound fixture exists, so this
is a cheap gap to close.

**Recommendation:** Assert both numbers separately on the completed-but-unsound Thread.

**Resolution:** A completed-but-unsound fixture now asserts 1/1 nominally done
and 0/1 soundly drained as distinct detail values.

#### L2. Inline edit is a silent no-op on a Thread whose detail has not loaded  · **Status:** fixed

**File:** internal/tui/model.go (keys.Edit branch) | **Component:** action-surface honesty
**Effort:** XS · **Urgency:** eventually

The `e` branch falls through to `else if m.selectedPath() != ""` to flash "no inline edit here — press
E to edit in $EDITOR". For Threads `selectedPath()` is empty until the detail for that row has
landed, so pressing `e` immediately after moving the cursor produces no feedback at all. Every other
entity exposes a path from the row itself and always gets the hint. `E` and `Y` handle this state
explicitly ("Thread path is still loading"); `e` does not.

**Recommendation:** Give the `e` branch the same kind-aware fallback the editor and yank paths
already have.

**Resolution:** Lowercase e now explains both a still-loading Thread path and a
pathless portable reader instead of silently doing nothing; focused tests cover
both states.

### Acceptance-criteria traceability

| AC | Verdict | Evidence |
|---|---|---|
| 1. First-class tab, palette, filter/sort/restore, path copy/open degrades explicitly | **met** | `TestThreadsUseOneRegistryListAndDetailOwner` asserts palette route+jump by canonical ID; `TestThreadRegistryPreservesSelectionFilterAndSortState`; pathless degradation proven by `TestThreadRouteSurvivesSplitPathlessCapabilities` and my killed `path_mandatory` mutation. Minor gap at L2. |
| 2. One list request, one generation/error owner, one canonical detail selection, one reload path | **met** | Re-derived above; `readThreadList`/`readThreadDetail` deleted; `threadProjectionState` and both Thread message types gone; `drop_list_gen_guard` and `drop_identity_validation` mutations both killed. |
| 3. Rows distinguish lifecycle, sound progress, health, active/eligible work without recomputing | **partially met** | Lifecycle/health/progress are faithful. The blocked count *is* TUI-derived and diverges from dispatchability on an unhealthy graph — M1, M3. |
| 4. Detail presents members, gates, in-flight, frontier, missing/broken, inconsistency, body | **met** | Verified against the probe space; `Members (persisted order)`, `Immediate external gates (not members)`, and `Diagnostics` sections all render; `partial_row_view` mutation killed. |
| 5. Unsound/cancelled/empty/shared/gated/healthy remain distinguishable | **met** | All six shapes built in the probe space and confirmed distinct via `thread list` parity plus `TestThreadProjectionStatesRemainVisuallyDistinct`. |
| 6. Normal/narrow/minimum widths keep essential state readable | **not met as written** | M2: at w≤40 with long slugs the work triple is truncated away and two distinct Threads render identically. |
| 7. No TUI-local traversal, eligibility rule, fs read/write, or mutation introduced | **met, with one caveat** | No traversal or fs access; no mutation surface. The blocked count is an aggregation over supplied role/gate rather than a traversal, but its implied semantics are TUI-invented — see M1. |

### Hostile-evidence ledger

**Probe space** (built outside the repo, deleted): 5 Threads / 9 task IDs covering an in-progress
member, both eligible pending statuses (`next-up` *and* `ready-to-start`, both eligible — confirming
eligibility is graph-derived, not status-derived), a blocked pending member, an external gate,
shared membership, a cancelled Thread, completed-but-unsound (`1/2 done · 0/2 drained` with
`completed-thread-undrained`), an empty Thread (`0/0`), and a missing member
(`healthy/broken`, `missing-thread-member`). TUI values cross-checked against
`thread list`, `thread show`, and `thread frontier`.

**Mutation ledger** — eight required probes, all applied to a backed-up copy and restored; final
`git diff` empty.

| # | Probe | Result | Killed by |
|---|---|---|---|
| 1 | Second Thread list/detail owner / dedicated retry surface | **killed** | `TestThreadsUseOneRegistryListAndDetailOwner` (DeepEqual over retained rows + diagnostics) |
| 2 | Remove list gen stale guard | **killed** | `TestModel_StaleReloadDoesNotStealRestore` |
| 2b | Remove canonical-key identity validation | **killed** | `TestEntityRegistryRejectsEmptyOrDuplicateCanonicalKeys` |
| 3 | Count clear pending work instead of `view.Frontier` | **SURVIVED** | — → finding M3 |
| 4 | Replace retained `core.ThreadView` with a partial summary | **killed** | `TestThreadsUseOneRegistryListAndDetailOwner`, `TestThreadRouteSurvivesSplitPathlessCapabilities` |
| 5 | Make `ThreadPath` mandatory for semantic detail | **killed** | `TestThreadRouteSurvivesSplitPathlessCapabilities` |
| 6 | Route identity through slug instead of canonical ID | **killed** | 3 tests: single-owner, path-survives-failure, selection/filter/sort |
| 7a | Drop repository graph/read diagnostics | **killed** | single-owner, split-capabilities |
| 7b | Collapse nominal and sound progress | **SURVIVED** | — → finding L1 |
| 8 | Expose lifecycle mutation on the read-only Thread tab | **SURVIVED** | — → finding M4 |

**Completion-order probes** — driven by constructing messages directly against the production
reducer, not through the synchronous helper:

| Case | Result |
|---|---|
| Older list generation lands after a newer one | dropped; rows intact |
| Older detail generation, same Thread | dropped; `loadedKey` and body unchanged |
| Detail for a *different* Thread at the current generation | dropped (guarded by `isCurrentSelection`) |
| Background-tab load while Threads active | active selection and detail untouched (`msg.kind != m.cur().kind` early return) |
| Task-only edit and Thread-document edit | both reload the tab — `TestThreadRegistryReloadsOnTaskAndThreadChanges` |

**Rendering evidence** — visible cell width after `ansi.Strip`, with a double-width CJK slug and a
long accented description:

```
w=100 cells=100 runes=88    w=64 cells=64 runes=52    w=42 cells=42 runes=30
w= 34 cells= 34 runes=23    w=24 cells=24 runes=18
```

Cells never exceed the terminal width; rune counts diverge sharply from cell counts, confirming the
clipper is cell-aware rather than rune-aware. Overflow is settled; what survives truncation is M2.

### Separation of concerns

**Demonstrated defects:** M1 (numeric divergence reproduced), M2 (byte-identical rows reproduced).

**Source-supported risks / coverage:** M3, M4, L1 — surviving mutations against currently-correct
production code. L2 — read directly from the `keys.Edit` branch.

**Unverified concerns, stated as such:** I ran no live PTY or interactive resize. Width evidence is
renderer-level (delegate output measured in cells), not a real terminal, as the brief permits.
I did not build a multi-thousand-Thread space, so finding 10's scale question is unaddressed rather
than cleared. I did not exercise a real concurrent workspace switch under in-flight Thread reads
beyond the session-generation code path, which I read but did not execute adversarially.

### Settled concerns

Recorded so they are not re-investigated.

- **Mid-enum insertion of `entityThreads`.** `entityKind` is never serialized — no JSON tag, config
  key, golden file, or on-disk session. Every lookup goes through `indexOfKind`; the only numeric
  uses are relative (`len(m.tabs)-1`, `m.active = 0`). Renumbering audits/research is inert.
- **Canonical identity.** Routing identity through the slug is killed by three tests; empty and
  duplicate canonical keys are rejected by `validateEntityItems` *before* rows are installed.
- **A dedicated Thread retry surface.** Genuinely removed: `readSurface` has only
  list/detail/dashboard, and Thread detail rides `readEntityDetail` with the shared conflict policy.
- **Diagnostics as a second cache.** `threadListDiagnostics` holds three repository-level fields, is
  written only in `handleListLoaded` from the same generation-guarded message, and is read for
  display. It is metadata, not a state machine — and dropping it is caught.
- **`ThreadPath` TOCTOU.** The path is captured with the detail and could go stale after a rename.
  But an active-tab list load calls `refreshDetail()`, so the watcher's debounce bounds the window,
  and `selectedPath()` is gated on `loadedKey == selectedKey()` so it can never return another
  Thread's path. With the watcher off the path can age — but that is the same staleness class as
  every other entity's row path, not something this change introduces.
- **Width/Unicode overflow.** Settled by the cell-width measurements above.
- **Session state.** `threadProjectionState` is fully removed from `spaceSession`; Thread state now
  rides `tabs`, which the session already saved and restored.

### If you fix only one thing

M1 and M3 are the same defect seen from production and from the test suite. Closing M3 with a
degraded-graph fixture would have caught M1 before it shipped, and the fixture costs a few lines.
