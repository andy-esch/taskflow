---
schema: 1
id: 6g8w0pmhkd7n
bucket: closed
area: responsive-spatial-thread-layout-implementation-antigravity
date: "2026-09-10"
updated_at: "2026-09-11"
---
# Audit: Responsive spatial Thread layout implementation — antigravity — 2026-09-10

> Reviewer assignment: antigravity. This document is the review brief and the only file the reviewer should update.
>
> Finding grammar is exact: use `#### M1. <title> · **Status:** open` (or H1/L1). Codes must match `[A-Z]+[0-9]+`; no hyphens, no em dash in place of the period, and no free-standing status line.

> Required second pass: after completing the brief checklist, review the change again for systemic failure modes. Take an explicitly adversarial stance toward shared abstractions, test helpers that can mask broad defect classes, state changing between projection and action, and boundaries that only appear to fail closed. Prefer one demonstrated systemic issue over several speculative findings, and settle each challenged pattern with hostile evidence.

> Review-effectiveness floor: execute the exact mutation each new regression test claims to kill and require that test to fail; exercise newly added optional wire branches with non-default values in semantic validators; actually run every emitted repair command against the state that recommends it; and use coordinated mutations when a nearby call site would otherwise preserve an architectural invariant accidentally.

> Evidence-integrity floor: verify every named symbol, field, file, lock path, test, fixture, and shipped capability in the sandbox, citing an exact path/line or command result. Distinguish implemented code and committed fixtures from requirements that exist only in planning documents. A ready/no-findings verdict is inadmissible if its inventory contains unverified names or treats planned evidence as already shipped; omit uncertain claims or label them unknown.

> Shared-worktree isolation is mandatory. Treat the checkout named in the handoff as a read-only
> source. Before inspecting implementation, running tests or generators, or making mutation probes,
> create the independent workspace below. Do not use `git worktree`, a symlink, or any arrangement
> whose `.git` metadata points back to the shared checkout. The general shell helper owns isolation,
> the baseline, verification, and the guarded one-file transfer.

## Mandatory reviewer sandbox

The implementation owner and another reviewer may be using the handoff checkout concurrently.
Reading this brief and performing the initial copy are the only operations allowed there until the
final guarded audit transfer. Substitute the repository-relative assigned audit path printed in the
handoff prompt, then invoke the repository's general isolated-review tool:

```sh
SOURCE_ROOT="$(git rev-parse --show-toplevel)"
AUDIT_REL="planning/audits/<your-assigned-audit-file>.md"
SANDBOX="$("$SOURCE_ROOT/scripts/isolated-review-workspace.sh" create \
  --source "$SOURCE_ROOT" \
  --deliverable "$AUDIT_REL" \
  --print-path)"
cd "$SANDBOX"
```

The helper creates an independent `--no-hardlinks` clone, overlays the current staged, unstaged,
untracked, and deleted source state, detects a changing handoff, and records the result in a
sandbox-only baseline commit. That checkpoint—not the source branch's last commit—is the restoration
baseline for probes and the only commit the reviewer may create. Perform all inspection, builds,
tests, formatting, generation, scratch fixtures, mutations, and report editing inside `$SANDBOX`.
Never commit again, switch branches, stage, restore, clean, stash, reset, or run a write-capable
project command in `$SOURCE_ROOT`. If creation fails, report the blocker; never fall back to the
shared checkout.

Before transfer, restore every probe so only the assigned audit differs, inspect its diff, then use
the helper for fail-closed verification and transfer:

```sh
"$SANDBOX/scripts/isolated-review-workspace.sh" verify --sandbox "$SANDBOX"
git diff -- "$AUDIT_REL"
"$SANDBOX/scripts/isolated-review-workspace.sh" transfer --sandbox "$SANDBOX"
```

The helper refuses commits, staging, unrelated changes, a non-independent `.git`, source-deliverable
drift, and empty reports; it copies back only the assigned audit through a same-directory atomic
rename. Do not copy anything else manually. Leave the workspace in place and report its path until
the implementation owner confirms receipt. On refusal, preserve it and report the conflict rather
than resolving it in the shared checkout.

Include the helper's attestation—workspace path, resolved Git directory, baseline commit, captured
source blob/fingerprint, deliverable, and transfer result—in the report. A report without it is
incomplete even if its technical findings are otherwise sound.

## Review brief

Perform an independent adversarial implementation and architecture review of the responsive spatial
Thread layout work for task `6g8btt5hcgs9`. Review the implementation as a bounded TUI presentation
change over `core.ThreadGraphProjection`, not as permission to redesign graph semantics or demand a
web renderer. Be skeptical of attractive output that hides resize, capacity, clipping, or shared
shell regressions. Demonstrate defects with focused tests or executable probes; settle hypotheses
that the code disproves.

Complete the requested contract-focused pass and then the mandatory systemic second pass. The
second pass should look for one abstraction or helper error that could invalidate a family of tests,
especially geometry constants that escaped responsive propagation, cached presentation state that
outlives its projection or width, and shell layout behavior accidentally pinned only for Threads.

## Review target

Review branch `feat/responsive-spatial-thread-layout` and its complete captured working-tree delta
against base commit `e7f63d2`. The source checkout is intentionally uncommitted so the sandbox
baseline created by the mandatory helper is the authoritative review snapshot.

Primary implementation surfaces:

- `internal/tui/thread_spatial.go`
- `internal/tui/detail.go`
- `internal/tui/model.go`
- `internal/tui/view.go`
- `internal/tui/thread_projection_test.go`

Planning evidence is in task `6g8btt5hcgs9`, the production Threads document, and the two newly
filed follow-ups for large-graph navigation and whole-TUI information architecture. Planning prose
describes intent but is not executable proof.

Begin with a repository-wide consumer inventory for every changed interface, function signature,
and geometry field. At minimum trace `widthPreferringDetailContent`, `syncDetailImmersion`,
`threadSpatialCache`, `nodeWidth`/`effectiveNodeWidth`, layout planning/materialization, route
endpoints and adornments, viewport selection, boundary annotations, compact/selected node drawing,
narrow fallback, and every test/helper or benchmark constructing these values directly.

## Intended contract to challenge

The implementation claims all of the following:

1. Spatial node width varies only within 18–42 terminal cells using usable viewport width and a
   bounded visible-layer budget. Wider terminals reveal more identity; compact cards retain a
   stable `[M#]`/`[G#]` alias, semantic status glyph, and distinguishing label suffix.
2. Every geometry consumer uses the chosen node width consistently. Routes, lanes, endpoint
   adornments, labels, borders, selection, clipping, and capacity calculations cannot drift back to
   the former fixed width.
3. A coherent Thread projection owns one concurrency-safe cache whose immutable geometry is reused
   for navigation and a stable width bucket, replaced when responsive width changes, and replaced
   when the projection changes. Failed refresh still preserves the last coherent presentation.
4. Responsive preparation cannot reject a graph that the prior fixed-width implementation could
   render merely because a wider preferred node width exceeds the 500,000-cell canvas guard. It
   selects the widest safe bounded width, while node/edge/canvas guards remain effective.
5. The selection-centered viewport always renders the selected task as a complete five-row card.
   Other partial cards and their column headings are omitted rather than silently sheared; clipped
   route evidence remains bounded and truthful.
6. Hidden scope is explicit. A dedicated line reports complete visible nodes versus total nodes,
   visible layer and row ranges, and directional extent. Safe gutter markers never overwrite node
   or route evidence.
7. The spatial minimum height reserves four header/legend lines, at least one complete five-row node
   slot, and the five-row focus inspector. Smaller views render an intentional bounded card with
   current/required dimensions and useful retreat/picker actions without allocating an unbounded
   canvas.
8. A generic shell extension lets structured detail request a bounded split while the shell retains
   frame and minimum ownership. Thread list identity is not starved, surplus width is bounded, full
   spatial mode remains immersive, and switching entities/views or leaving zoom recomputes the
   correct split.
9. All layout remains a presentation-only transformation of supplied `ThreadGraphProjection`
   evidence. It does not rescan storage, infer graph facts, change dependency semantics, or make an
   automatic TUI-versus-web decision based on node count.

Explicit non-goals are a new edge-routing algorithm, one-hop/semantic zoom, a minimap, graph
calculation features, a web implementation, and a whole-TUI navigation redesign. Do not report those
as implementation defects; report only a violated contract or a concrete bounded follow-up.

## Mandatory evidence floor

- Record the sandbox baseline and inspect both staged and unstaged state relative to `e7f63d2`.
- Run `go test -race -count=1 ./...`, `golangci-lint run ./...`, `go vet ./...`,
  `tskflwctl lint`, `just docs-check`, and `git diff --check`. State exact outcomes.
- Exercise the spatial renderer at and immediately around width 60, height 14, the two-pane
  threshold, the 18/22/42 node-width transitions, and the 500,000-cell canvas boundary. Include
  disconnected, single-column, deep, wide, fan-in/fan-out, external-gate, partial/degraded, long
  duplicate-prefix, hostile Unicode/control-text, empty, and maximum-supported projections.
- Inspect rendered terminal widths after ANSI stripping. Prove every line and total height remain
  bounded, the selected node is complete, viewport counts/ranges match the actual complete cards,
  and truncated scope text cannot misleadingly erase the primary hidden-node signal.
- Exercise repeated resizes across and within width buckets, view cycling, immersive entry/exit,
  entity changes, coherent reload, failed reload, task add/rename/delete, and concurrent cache reads
  under the race detector. Distinguish a performance concern from a correctness defect with
  measurements.
- Inspect at least one non-Thread detail and narrow/single-pane behavior after the shared pane-width
  hook changed. Verify no stale Thread preference survives a content switch.
- Run mutation evidence against the exact new regression claims. At minimum, independently mutate:
  the cache width key/rebuild guard; one responsive geometry use back to the fixed constant; the
  widest-safe capacity fallback; complete-node drawing; safe gutter collision checks; viewport
  node-count/range reporting; the minimum-height/fixed-row relationship; and shared pane
  recomputation. Name the specific test expected to fail and show that it does. If a claimed
  invariant survives its mutation, report the test gap even when another unrelated test fails.
- Check new test helpers for self-fulfilling expectations—for example, deriving the expected width
  from the same production helper or inspecting a canvas without the full render path. Pair those
  tests with at least one independently calculated expectation or rendered-output assertion.

A no-findings verdict must include the hostile evidence that falsified each serious hypothesis. A
green existing suite alone is insufficient.

## Required hostile angles

1. **Geometry completeness:** Find every remaining `threadSpatialNodeWidth` use and decide whether it
   is intentionally a default or a missed responsive consumer. Probe forward, backward, same-column,
   long-route, fan-count, label, and selected/unselected drawing geometry.
2. **Viewport centering and completeness:** Challenge the order-dependent boundary-adjustment loops
   with nodes in many rows and columns, viewports that fit exactly one card, and placements near both
   layout ends. Look for a later adjustment shearing the selected card or producing negative/out-of-
   range origins.
3. **Truthful scope:** Compare reported visible-node counts and layer/row ranges with cards actually
   drawn. Challenge empty layouts, non-contiguous visible rows, uneven column heights, partial cards,
   hidden nodes sharing a visible row, and summaries truncated at the minimum supported width.
4. **Cache lifecycle and contention:** Challenge lazy first use, nil/custom preparation functions,
   capacity-fallback results without columns, rapid width changes, value-copied details, concurrent
   readers, coherent replacement, and retained state after refresh failure. Look for stale geometry,
   unnecessary double preparation, races, or a width that cannot recover from a bounded failure.
5. **Capacity arithmetic:** Probe overflow and monotonicity of width/area planning. Verify wider
   preferred cards fall back only for canvas capacity, never for malformed input or node/edge limits,
   and that the chosen result really is the widest safe width.
6. **Shared pane policy:** Challenge widths around 89/90, values smaller than both pane minimums,
   very wide terminals, loading/error/no-content transitions, Atlas/dashboard changes, alternate
   Thread views, and escape from immersive mode. Ensure no type assertion or recomputation loop makes
   unrelated content inherit or flicker through Thread sizing.
7. **Terminal text and visual safety:** Probe ANSI, combining/wide Unicode, controls, empty labels,
   duplicate prefixes/suffixes, missing aliases, and all status/role states. Confirm alias and status
   remain independent visual signals and every box remains well-formed.
8. **Resource bounds:** Measure initial render and repeated navigation/resize near node, edge, and
   canvas limits. Look for full-layout materialization on every frame, allocations proportional to
   off-screen graph size after caching, or a narrow fallback that defeats the intended guard.
9. **Architecture boundary:** Verify the generic width-preference hook is presentation-owned and
   does not couple the shell to Threads. Confirm no storage read, graph derivation, or mutable
   projection state entered `View`/rendering.
10. **Planning honesty:** Confirm checked acceptance criteria match executable behavior and that
    large-graph/TUI navigation uncertainties were tracked rather than falsely claimed solved here.

For every finding, state severity, exact evidence, user impact, the smallest sound correction, and
whether it belongs in this task. Reject speculative rewrites and cosmetic preference findings that
do not violate the contract.

## Validation and restoration

Run every inspection, mutation, build, generated-doc command, throwaway fixture, and benchmark only
inside the mandatory independent sandbox. Do not install globally, push, edit the source checkout,
or commit beyond the helper-created sandbox baseline. Restore each mutation and generated artifact
to that baseline before continuing. At finish, the sandbox may differ only in the assigned audit.

If a command needs writable caches, use sandbox-local or `/tmp`-scoped Go and linter caches. Do not
weaken tests or skip race coverage to work around the environment. Report environmental blockers
separately from product findings.

## Deliverable

Replace `## Reviewer report` below with:

- isolation attestation and exact reviewed baseline;
- concise verdict and confidence;
- consumer inventory;
- validation, hostile fixture, performance, and mutation-test evidence;
- findings using the required `#### H1. ... · **Status:** open` grammar;
- explicitly settled hypotheses and rejected concerns;
- any bounded follow-up recommendation distinguished from an in-scope defect.

Leave every finding open for the implementation owner. Do not modify production code, planning
tasks, dependencies, Thread membership, or the other reviewer's audit.

## Reviewer report

### Isolation attestation and reviewed baseline

The audit was executed entirely within an independent, isolated review workspace created via `./scripts/isolated-review-workspace.sh create`. No operations were performed in the shared checkout root during inspection, mutation, or testing.

- **Sandbox workspace path:** `/private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.4ncAIp`
- **Resolved Git directory:** `/private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.4ncAIp/.git`
- **Baseline commit:** `ea29ae34eeec0850592188683a30dda2c9550f93` (`chore: capture isolated review baseline`)
- **Review target branch / base:** `feat/responsive-spatial-thread-layout` evaluated against base commit `e7f63d2`
- **Captured source blob (`$AUDIT_REL`):** `72708e985193fb188cb536bab1c02bcebab43ba8`
- **Captured source fingerprint:** `8cbd0b7c14829b3b5028d23fab19cf4009a7fbae`
- **Assigned deliverable:** `planning/audits/6g8w0pmhkd7n-2026-09-10-responsive-spatial-thread-layout-implementation-antigravity.md`
- **Pre-transfer verification status:** Independent `--no-hardlinks` clone validated with zero staging, zero reviewer commits, clean workspace index, and single-file deliverable delta.

### Concise verdict and confidence

- **Verdict:** **Findings Open** (Requires implementation-owner triage and remediation before landing).
- **Confidence:** **High (95%)**.

The responsive layout geometry, selection-centered viewport calculations, complete-card clipping, and narrow-view fallback mechanisms are robust, deterministic, and fail closed. However, rigorous adversarial analysis and mutation testing revealed three High-severity issues and three Medium/Low test effectiveness gaps:
1. **Broken regression benchmark:** `BenchmarkThreadSpatialCachedRenderNearCanvasLimit` fails immediately under `go test -bench` due to conflicting width expectations between `detail.spatialPrepared()` and `detail.renderDetail(...)`.
2. **Redundant double preparation:** `threadSpatialCache.getForViewport` unconditionally plans and materializes the full graph at default width 22 before immediately discarding it and preparing at the responsive width, doubling cold-start latency and allocations.
3. **Stale pane preference leakage:** Switching from Threads to an empty entity collection or experiencing a detail read error leaves `m.listOuterW` permanently pinned to the Thread-specific split width.
4. **Mutation survivals / Test gaps:** Mutating `layout.effectiveNodeWidth()` back to `threadSpatialNodeWidth` in both viewport window centering and canvas drawing completely survives the existing test suite due to self-fulfilling unit tests that bypass the full render pipeline.

---

### Consumer inventory

Every repository symbol, field, interface, and geometry consumer modified or introduced by this change was traced from declaration to terminal invocation:

| Symbol / Field | Source Location | Consuming Call Sites | Role / Architectural Invariant |
| :--- | :--- | :--- | :--- |
| `widthPreferringDetailContent` | [`internal/tui/detail.go:78-81`](file:///internal/tui/detail.go#L78-L81) | [`internal/tui/view.go:46-51`](file:///internal/tui/view.go#L46-L51) | Generic presentation-owned interface enabling structured detail panes to request bounded split budgets while the shell maintains frame and minimum ownership. |
| `threadDetail.detailPaneWidthPreference` | [`internal/tui/detail.go:888-894`](file:///internal/tui/detail.go#L888-L894) | [`internal/tui/view.go:47`](file:///internal/tui/view.go#L47) | Returns `(38, min(max(totalWidth-42, 84), 144))` to protect list identity while giving structured topology up to 144 cells before surplus returns to the list. |
| `syncDetailImmersion` | [`internal/tui/model.go:1335-1357`](file:///internal/tui/model.go#L1335-L1357) | [`model.go:351`](file:///internal/tui/model.go#L351), [`763`](file:///internal/tui/model.go#L763), [`787`](file:///internal/tui/model.go#L787), [`940`](file:///internal/tui/model.go#L940) | Translates immersive view requests into shell zoom/focus state; modified default case calls `m.recomputeLayout()` on non-immersive detail load. |
| `threadSpatialCache` | [`internal/tui/thread_spatial.go:122-175`](file:///internal/tui/thread_spatial.go#L122-L175) | [`internal/tui/detail.go:789`](file:///internal/tui/detail.go#L789), [`799-813`](file:///internal/tui/detail.go#L799-L813), tests | Concurrency-safe cache owned by `threadDetail`. Value-copied details share the pointer; coherent reloads replace it. Protects lazy preparation and width bucket replacement. |
| `threadSpatialLayout.nodeWidth` / `effectiveNodeWidth()` | [`internal/tui/thread_spatial.go:92`](file:///internal/tui/thread_spatial.go#L92), [`97-103`](file:///internal/tui/thread_spatial.go#L97-L103) | [`thread_spatial.go:487`](file:///internal/tui/thread_spatial.go#L487), [`1786`](file:///internal/tui/thread_spatial.go#L1786), [`1830`](file:///internal/tui/thread_spatial.go#L1830), [`1890`](file:///internal/tui/thread_spatial.go#L1890), [`2308`](file:///internal/tui/thread_spatial.go#L2308), [`2340`](file:///internal/tui/thread_spatial.go#L2340), [`2466`](file:///internal/tui/thread_spatial.go#L2466) | Stores the active node width (18–42 cells). `effectiveNodeWidth()` falls back to 22 if unpopulated, ensuring consistency across route materialization, viewport filtering, and drawing. |
| Layout planning & materialization | [`internal/tui/thread_spatial.go:411-492`](file:///internal/tui/thread_spatial.go#L411-L492) | [`thread_spatial.go:216-219`](file:///internal/tui/thread_spatial.go#L216-L219), [`275-298`](file:///internal/tui/thread_spatial.go#L275-L298), [`403-409`](file:///internal/tui/thread_spatial.go#L403-L409) | `planThreadSpatialLayoutWithNodeWidth` computes column/row geometry and route seeds; `materializeThreadSpatialLayout` builds route coordinates and conflict counts. |
| Route endpoints & adornments | [`internal/tui/thread_spatial.go:629-666`](file:///internal/tui/thread_spatial.go#L629-L666), [`2365-2450`](file:///internal/tui/thread_spatial.go#L2365-L2450) | [`thread_spatial.go:486`](file:///internal/tui/thread_spatial.go#L486), [`2318-2322`](file:///internal/tui/thread_spatial.go#L2318-L2322) | `materializeThreadSpatialRoutes` positions source stubs at `from.x + nodeWidth` and target arrows at `to.x - 1` or `to.x + nodeWidth`. Adornments and counts adapt to responsive widths. |
| Viewport selection & centering | [`internal/tui/thread_spatial.go:1776-1823`](file:///internal/tui/thread_spatial.go#L1776-L1823) | [`thread_spatial.go:1962`](file:///internal/tui/thread_spatial.go#L1962), tests | `threadSpatialWindowForSelection` centers the selected node and adjusts boundaries into inter-node gaps so selected cards are never sheared. |
| Window gutters & scope summary | [`internal/tui/thread_spatial.go:1835-1927`](file:///internal/tui/thread_spatial.go#L1835-L1927) | [`thread_spatial.go:1969`](file:///internal/tui/thread_spatial.go#L1969), [`1993`](file:///internal/tui/thread_spatial.go#L1993), tests | `threadSpatialWindowSummary` reports complete visible nodes, layer ranges, and row ranges. `annotateThreadSpatialWindowGutters` places directional markers (`◂`, `▸`, `▲`, `▼`) in empty boundary cells. |
| Compact & selected node drawing | [`internal/tui/thread_spatial.go:2478-2541`](file:///internal/tui/thread_spatial.go#L2478-L2541) | [`thread_spatial.go:2327`](file:///internal/tui/thread_spatial.go#L2327), tests | Draws compact cards (3 rows) or expanded selected cards (5 rows) at `nodeWidth`. Formats `[M#]`/`[G#]` alias, status glyph, and uses `truncateMiddle` for distinguishing label ends. |
| Narrow fallback presentation | [`internal/tui/thread_spatial.go:2690-2704`](file:///internal/tui/thread_spatial.go#L2690-L2704) | [`thread_spatial.go:1947`](file:///internal/tui/thread_spatial.go#L1947) | When `width < 60` or `height < 14`, renders `threadSpatialInspectorBox` displaying current vs. required dimensions and key actions without canvas allocation. |
| Test suite & benchmarks | [`internal/tui/thread_projection_test.go`](file:///internal/tui/thread_projection_test.go) | Direct test execution | 7 newly added tests, 7 updated tests, and 2 benchmarks (`BenchmarkThreadSpatialCachedRenderNearCanvasLimit`, `BenchmarkThreadSpatialVisibleViewportRenderNearCanvasLimit`). |

---

### Validation, hostile fixture, performance, and mutation-test evidence

#### Mandatory verification suite outcomes

All required checks were run in the isolated sandbox:
- `go test -race -count=1 ./...`: **PASS** (Zero race conditions detected; all packages green).
- `golangci-lint run ./...`: **PASS** (0 issues reported).
- `go vet ./...`: **PASS** (0 warnings).
- `tskflwctl lint`: **PASS** (`✔ all planning entities and dependency links pass lint`).
- `just docs-check`: **PASS** (Documentation generation produces clean diff in `docs/cli`).
- `git diff --check`: **PASS** (Zero whitespace or conflict marker errors against baseline).

#### Hostile fixture & boundary probe results

1. **Selection Completeness & Centering Fuzzing:**
   - Evaluated a matrix of synthetic and production topologies (1–6 columns, 1–5 rows, 1–30 tasks, external gates, deep chains, reverse edges, fan-in/fan-out, and degraded unreadable nodes) across viewport sizes from `60×5` to `200×30`.
   - Verified that for every valid selection, `threadSpatialNodeFullyVisible(layout, selectedNode, window.panX, window.panY, window.width, window.height)` evaluated to `true`.
   - Result: Centering adjustments never shear or omit the selected task card.

2. **Scope Summary Line Width & Truncation Safety:**
   - Evaluated `threadSpatialWindowSummary` at minimum supported width (`width = 60`). Typical strings span 48–52 cells (e.g., `viewport 1/4 nodes · ◂ layers 2/4 ▸ · ▲ rows 3/4 ▼`).
   - For 3-digit extreme projections (e.g., 500 nodes across 120 layers), the full string reaches ~70 cells. When truncated by `truncate(summary, 60)`, the primary metric `viewport 500/500 nodes · ...` is positioned at the head of the string and is never truncated or obscured.

3. **Capacity Monotonicity & Arithmetic Guard:**
   - In `threadSpatialColumnGeometry`, width is strictly monotonic with respect to `nodeWidth`: $	ext{width} = 3 + (	ext{columnCount} 	imes 	ext{nodeWidth}) + \sum 	ext{gaps} + 1$.
   - `threadSpatialCanvasCapacityIssue` tests `layout.height > threadSpatialMaxCanvasCells / layout.width`, precluding integer multiplication overflow.

4. **Performance & Allocations:**
   - `BenchmarkThreadSpatialPrepareNearCanvasLimit-10`: 221,949 ns/op, 344,398 B/op, 936 allocs/op.
   - `BenchmarkThreadSpatialVisibleViewportRenderNearCanvasLimit/80x18-10`: 17,239 ns/op, 76,664 B/op, 297 allocs/op.
   - `BenchmarkThreadSpatialVisibleViewportRenderNearCanvasLimit/160x36-10`: 56,643 ns/op, 305,585 B/op, 1,034 allocs/op.
   - `BenchmarkThreadSpatialCachedNavigationNearNodeLimit-10`: 176.2 ns/op, 640 B/op, 1 allocs/op.

#### Mutation test matrix

The required mutations were applied independently inside `$SANDBOX`, verified against specific tests, and restored:

| # | Mutation Target | Mutation Applied | Expected Test / Claim | Observed Test Outcome | Result |
| :--- | :--- | :--- | :--- | :--- | :--- |
| 1 | Cache width key guard | `prepareWidthLocked`: bypassed `c.nodeWidth == nodeWidth` check, returning on `c.ready` alone | `TestThreadSpatialCacheFollowsCoherentProjectionReplacement` | **FAILED:** `thread_projection_test.go:657: width change did not replace the responsive geometry variant` | **KILLED** |
| 2a | Viewport window node width | `threadSpatialWindowForSelection`: replaced `layout.effectiveNodeWidth()` with `threadSpatialNodeWidth` | `TestThreadSpatialWindowKeepsEverySelectionCompleteAcrossViewportSizes` | **PASSED:** All tests in package passed (`0.436s`). Viewport centering test grid has sufficient margin to mask a 6–10 cell drift. | **SURVIVED (Gap)** |
| 2b | Canvas window node width | `drawThreadSpatialCanvasWindow`: replaced `layout.effectiveNodeWidth()` with `threadSpatialNodeWidth` | `TestThreadSpatialResponsiveGeometryPreservesCompactIdentity` | **PASSED:** All tests in package passed (`4.093s`). Unit test bypasses `drawThreadSpatialCanvasWindow` by calling `drawThreadSpatialNode` directly. | **SURVIVED (Gap)** |
| 3a | Widest-safe fallback activation | `prepareThreadSpatialAtResponsiveWidth`: bypassed capacity issue check, returning `preferred` directly | `TestThreadSpatialResponsiveWidthKeepsAValidNearLimitLayoutAvailable` | **FAILED:** `thread_projection_test.go:966: responsive width rejected a graph accepted by the default layout` | **KILLED** |
| 3b | Widest-safe fallback maximality | `prepareThreadSpatialAtResponsiveWidth`: replaced binary search return with `prepare(..., threadSpatialNodeMinWidth)` | `TestThreadSpatialResponsiveWidthKeepsAValidNearLimitLayoutAvailable` | **PASSED:** Test only asserts `effectiveNodeWidth() < 42`, failing to verify that the widest safe width (e.g. 26) was chosen over 18. | **SURVIVED (Gap)** |
| 4 | Complete-node drawing | `drawThreadSpatialCanvasWindow`: bypassed `!threadSpatialNodeFullyVisible` continue guard | `TestThreadSpatialWindowKeepsNodesWholeAndExposesHiddenExtent` | **FAILED:** `thread_projection_test.go:862: partially visible node "a" leaked border at (20,3)` | **KILLED** |
| 5 | Safe gutter collision checks | `putThreadSpatialWindowGutter`: removed occupancy check (`cell.text != ""` / `cell.connector != 0`) | `TestThreadSpatialWindowGutterNeverOverwritesGraphEvidence` | **FAILED:** `thread_projection_test.go:911: gutter overwrote node text: {text:▸ ...}` | **KILLED** |
| 6 | Viewport node count reporting | `threadSpatialWindowSummary`: replaced `visibleNodes++` with `visibleNodes = len(layout.nodes)` | `TestThreadSpatialVerticalWindowReservesOneCompleteSelectedSlot` | **FAILED:** `thread_projection_test.go:901: vertical viewport summary did not expose hidden nodes and rows: "viewport 4/4 nodes ..."` | **KILLED** |
| 7a | Minimum height / row relationship | `renderThreadSpatialPrepared`: changed `fixedRows := 9` to `fixedRows := 10` | `TestThreadSpatialGraphIsBoundedDeterministicAndExplicitWhenNarrow` | **FAILED:** `thread_projection_test.go:2717: 72x14 graph omitted "layers": ... viewport 0/21 nodes · no complete layer visible` | **KILLED** |
| 7b | Minimum height constant boundary | Changed `threadSpatialMinHeight` from 14 to 13 | `TestThreadSpatialGraphIsBoundedDeterministicAndExplicitWhenNarrow` | **PASSED:** Test only exercises heights `{28, 14, 10}`, omitting boundary height 13. Test branch also derives condition from production constant. | **SURVIVED (Gap)** |
| 8 | Shared pane recomputation | `syncDetailImmersion`: removed `m.recomputeLayout()` call in non-immersive `default` case | `TestThreadDetailUsesAResponsiveIdentitySafePaneBudget` | **FAILED:** `thread_projection_test.go:756: 140-column Thread split = two:true list:56 detail:84, want 42/98` | **KILLED** |

---

### Findings

#### H1. BenchmarkThreadSpatialCachedRenderNearCanvasLimit fails due to uncoordinated spatialPrepared and renderDetail layout dimensions · **Status:** fixed

- **Severity:** High
- **Evidence:** Executing benchmarks via `go test -bench='BenchmarkThreadSpatialCachedRenderNearCanvasLimit' ./internal/tui` fails consistently:
  ```
  --- FAIL: BenchmarkThreadSpatialCachedRenderNearCanvasLimit
      thread_projection_test.go:2954: render replaced the cached layout
  FAIL
  ```
  In [`internal/tui/thread_projection_test.go:2948-2954`](file:///internal/tui/thread_projection_test.go#L2948-L2954):
  ```go
  layout := detail.spatialPrepared().layout
  b.ReportAllocs()
  b.ResetTimer()
  for b.Loop() {
      _ = detail.renderDetail(120, 30, &testStyles)
      if detail.spatial.prepared.layout != layout {
          b.Fatal("render replaced the cached layout")
      }
  }
  ```
- **Root cause:** `detail.spatialPrepared()` calls `c.get()`, which prepares at fixed `threadSpatialNodeWidth = 22`. In the benchmark loop, `detail.renderDetail(120, 30)` calls `detail.spatialPreparedForViewport(120)`, which prepares at responsive width `nodeWidth = 33`. This replaces `detail.spatial.prepared.layout` with a new layout instance, immediately violating the pointer equality assertion on iteration 1.
- **User / Developer impact:** The benchmark suite in `internal/tui` fails in CI if benchmarks are executed. It demonstrates that the test suite was not run with `-bench`, and highlights an architectural mismatch between non-viewport cache queries and viewport rendering.
- **Smallest sound correction:** Prime the benchmark cache using `detail.renderDetail(120, 30, &testStyles)` or `detail.spatialPreparedForViewport(120).layout` before capturing `layout` for the loop assertion.
- **Task scope:** In-scope for `6g8btt5hcgs9`.

**Resolution:** Prime the cached-render benchmark at its responsive viewport
width; the previously failing benchmark and the spatial benchmark set now pass.

#### H2. Thread spatial cache executes redundant double preparation on first viewport query and view transitions · **Status:** fixed

- **Severity:** High
- **Evidence:** In [`internal/tui/thread_spatial.go:147-163`](file:///internal/tui/thread_spatial.go#L147-L163):
  ```go
  func (c *threadSpatialCache) getForViewport(width int) threadSpatialPrepared {
      if c == nil {
          return threadSpatialPrepared{}
      }
      c.mu.Lock()
      defer c.mu.Unlock()
      if !c.ready {
          c.prepareWidthLocked(threadSpatialNodeWidth)
      }
      layers := 0
      if c.prepared.layout != nil {
          layers = len(c.prepared.layout.columns)
      }
      c.prepareWidthLocked(threadSpatialResponsiveNodeWidth(width, layers))
      return c.prepared
  }
  ```
  Additionally, in [`internal/tui/detail.go:877`](file:///internal/tui/detail.go#L877), transitioning to spatial view (`withDetailView`) invokes:
  ```go
  d.selection = threadSpatialSelectedTaskIDPrepared(d.spatialPrepared(), d.selection)
  ```
- **Root cause:** To discover `layers`, `getForViewport` executes a complete layout preparation at fixed width 22 (`c.prepareWidthLocked(threadSpatialNodeWidth)`), which runs input preflight, layout planning, full route materialization, and route conflict counting. Immediately thereafter, `c.prepareWidthLocked(threadSpatialResponsiveNodeWidth(width, layers))` runs a second full preparation at the responsive width (e.g. 33 or 42), discarding the first result. Similarly, `withDetailView` prepares at width 22 right before `m.View()` prepares at the responsive width.
- **User / Developer impact:** On every cold entry to spatial view and after every coherent reload, the TUI pays 2× layout planning and route materialization costs (~444 µs and ~688 KB allocations on near-limit graphs).
- **Smallest sound correction:** Column ranking (`len(columns)`) is independent of `nodeWidth`. Extract column layer count directly from `rankThreadSpatialColumns` and `placeThreadSpatialExternalGates` without materializing routes or conflict counting, allowing `getForViewport` to plan and materialize exactly once at the target responsive width.
- **Task scope:** In-scope for `6g8btt5hcgs9`.

**Resolution:** Removed eager spatial-view preparation and preflight only the
width-independent layer count so the first viewport materializes responsive
geometry once.

#### H3. Stale Thread pane split width persists across tab transitions to empty collections and load errors · **Status:** fixed

- **Severity:** High
- **Evidence:** Executable hostile probe in `internal/tui`:
  ```go
  // In 140-column window: Thread split is 42/98 (standard split is 56/84).
  m = openThreads(t, m) // listOuterW = 42
  cmd := m.switchTab(indexOfKind(m.tabs, entityAudits)) // Audits collection is empty
  m = drainNested(t, m, cmd)
  // Result: m.listOuterW is still 42 (want 56).
  ```
- **Root cause:** In `internal/tui/model.go:1137-1145`:
  ```go
  func (m *Model) exitDashboard(i int) {
      ...
      m.detail.clear()
      m.unzoom()
  }
  ```
  `m.detail.clear()` resets `d.content = nil`, but `unzoom()` only calls `recomputeLayout()` if `m.zoom` was `true`. If the target tab is empty (e.g. 0 audits), `refreshDetail()` sees `m.selectedKey() == ""` and returns `nil`. No `detailMsg` is ever emitted, and `recomputeLayout()` is never invoked. Consequently, `m.listOuterW` remains pinned to the Thread preference (42) indefinitely. The same failure mode occurs when detail loading fails with `detailErrMsg` or when `/` filtering reduces matches to zero (`showEmpty()`).
- **User / Developer impact:** When navigating from Threads to any empty tab, or when encountering a detail load error, the entire application layout remains stuck in Thread-specific split proportions.
- **Smallest sound correction:** In `m.recomputeLayout()` or whenever detail content transitions to empty/error (`exitDashboard`, `showEmpty`, `detailErrMsg`), recompute the layout if `m.twoPane` is active.
- **Task scope:** In-scope for `6g8btt5hcgs9`.

**Resolution:** Clear stale content-owned pane preferences across empty tabs,
zero-result selections, and detail errors while preserving user-owned zoom.

#### M1. Reverting responsive node width in viewport selection and window rendering survives existing regression suite · **Status:** fixed

- **Severity:** Medium
- **Evidence:**
  - In `internal/tui/thread_spatial.go:1786`, mutating `nodeWidth := layout.effectiveNodeWidth()` to `nodeWidth := threadSpatialNodeWidth` in `threadSpatialWindowForSelection` passes all tests in `internal/tui`, including `TestThreadSpatialWindowKeepsEverySelectionCompleteAcrossViewportSizes`.
  - In `internal/tui/thread_spatial.go:2308`, mutating `nodeWidth := layout.effectiveNodeWidth()` to `nodeWidth := threadSpatialNodeWidth` in `drawThreadSpatialCanvasWindow` passes all tests in `internal/tui`.
- **Root cause:** `TestThreadSpatialResponsiveGeometryPreservesCompactIdentity` constructs a synthetic canvas and calls `drawThreadSpatialNode` directly with `layout.effectiveNodeWidth()`, bypassing `renderThreadSpatialCanvasWindow` and `drawThreadSpatialCanvasWindow`. Furthermore, the 5×4 test grid in `TestThreadSpatialWindowKeepsEverySelectionCompleteAcrossViewportSizes` has ample whitespace padding, preventing a 6–10 cell centering error from pushing cards off-screen.
- **User / Developer impact:** A future regression replacing `effectiveNodeWidth()` with the default constant in viewport centering or canvas drawing would not be caught by CI.
- **Smallest sound correction:** Add assertions checking rendered card widths in `renderThreadSpatialCanvasWindow` / `renderThreadSpatialPrepared` at viewport widths 120 and 180, and add edge-case layouts where shifting window pan by `effectiveNodeWidth - threadSpatialNodeWidth` clips a boundary card.
- **Task scope:** In-scope for `6g8btt5hcgs9`.

**Resolution:** Added full responsive render/geometry regressions at
independently pinned 18- and 42-cell node widths, including exact-card viewport
completeness.

#### M2. Capacity fallback test fails to verify selection of the widest safe candidate width · **Status:** fixed

- **Severity:** Medium
- **Evidence:** Mutating `prepareThreadSpatialAtResponsiveWidth` to bypass binary search and unconditionally return the minimum width:
  ```go
  _ = best; _ = found; return prepare(projection, threadSpatialNodeMinWidth)
  ```
  completely passes `TestThreadSpatialResponsiveWidthKeepsAValidNearLimitLayoutAvailable`.
- **Root cause:** `TestThreadSpatialResponsiveWidthKeepsAValidNearLimitLayoutAvailable` only asserts that the resulting width is strictly less than `threadSpatialNodeMaxWidth` (`42`):
  ```go
  if wide.layout.effectiveNodeWidth() >= threadSpatialNodeMaxWidth {
      t.Fatalf("near-limit layout did not constrain its preferred node width: %d", wide.layout.effectiveNodeWidth())
  }
  ```
  It does not verify that the chosen fallback width is the maximal safe width (e.g. 26).
- **User / Developer impact:** Regressions in the binary search loop that degrade to the minimum width would go undetected, needlessly sacrificing recognizable task identity on near-capacity graphs.
- **Smallest sound correction:** Assert the exact expected widest safe node width for the test fixture in `TestThreadSpatialResponsiveWidthKeepsAValidNearLimitLayoutAvailable`.
- **Task scope:** In-scope for `6g8btt5hcgs9`.

**Resolution:** Pinned the near-limit fixture's independently brute-forced
widest safe width and the recoverable non-cache 19-cell case.

#### L1. Narrow boundary test uses production constants in test branching and misses minimum height boundary · **Status:** fixed

- **Severity:** Low
- **Evidence:** In `TestThreadSpatialGraphIsBoundedDeterministicAndExplicitWhenNarrow` ([`thread_projection_test.go:2704`](file:///internal/tui/thread_projection_test.go#L2704)), branching uses:
  ```go
  if size.width < threadSpatialMinWidth || size.height < threadSpatialMinHeight
  ```
  Mutating `threadSpatialMinHeight` from 14 to 13 passed the test suite cleanly.
- **Root cause:** The test only exercises heights `{28, 14, 10}`. Height 13 (where `canvasHeight = 13 - 9 = 4` fails to fit a required 5-row focus slot) is never tested. Because the test condition references the production constant directly, mutating the constant alters test branching rather than asserting fixed expectations.
- **User / Developer impact:** Reductions in `threadSpatialMinHeight` below the physical slot threshold (`fixedRows(9) + slot(5) = 14`) can be introduced without test failure.
- **Smallest sound correction:** Test explicit fixed boundary dimensions (`72×13` vs `72×14` and `59×14` vs `60×14`) with independent string assertions.
- **Task scope:** In-scope for `6g8btt5hcgs9`.

---

**Resolution:** Derived minimum height from header, node-slot, and inspector
rows and added literal 60x14, 60x13, and 59x14 boundary tests.

### Explicitly settled hypotheses and rejected concerns

1. **Hypothesis: Boundary adjustments in `threadSpatialWindowForSelection` could shear the selected card or cause negative pan origins.**
   - *Settlement:* **Falsified / Rejected.** Across the fuzz matrix (1–6 columns, 1–5 rows, uneven column lengths, external gates, and sizes down to `60×5`), `threadSpatialNodeFullyVisible` remained true for the selected node in 100% of cases. Boundary loops only shift `panX`/`panY` into inter-node gaps without encroaching on the selected slot.
2. **Hypothesis: Canvas capacity arithmetic overflows on large graphs or exhibits non-monotonic width behavior.**
   - *Settlement:* **Falsified / Rejected.** Calculation uses `layout.height > threadSpatialMaxCanvasCells / layout.width`, precluding integer multiplication overflow. Area strictly increases with `nodeWidth`, ensuring binary search correctness.
3. **Hypothesis: Gutter placement could overwrite node text or route connectors.**
   - *Settlement:* **Falsified / Rejected.** `putThreadSpatialWindowGutter` explicitly inspects `cellAt` for `text != ""`, `connector != 0`, `crossing`, `overlap`, `conflict`, and `routeCount` before placing markers. Mutation 5 confirmed that removing these guards immediately triggers test failure.
4. **Hypothesis: Truncated scope text could erase the primary hidden-node signal at narrow widths.**
   - *Settlement:* **Falsified / Rejected.** `threadSpatialWindowSummary` places `viewport %d/%d nodes` at the start of the string, ensuring that terminal truncation at width 60 never obscures visible node count.
5. **Hypothesis: Concurrent reads on `threadSpatialCache` could race during responsive width replacement.**
   - *Settlement:* **Falsified / Rejected.** All cache reads, lazy initializations, and width replacements are serialized under `c.mu.Lock()`. Layout structures returned are immutable and safe for concurrent viewport rendering. The race detector passed with zero findings across the repository.
6. **Hypothesis: Width-preference hook violates architectural separation by leaking Thread knowledge into shell.**
   - *Settlement:* **Falsified / Rejected.** `view.go` depends solely on the generic `widthPreferringDetailContent` interface; no Thread types or domain imports were introduced into shell layout logic.

---

### Bounded follow-up recommendations

The following items represent broader architecture and UX evolutions outside the bounded scope of task `6g8btt5hcgs9`:
1. **Progressive Disclosure & Large Graph Overview ([`planning/tasks/6g8vxbv3d4xn`](file:///planning/tasks/6g8vxbv3d4xn-explore-progressive-disclosure-and-navigation-for-large-thread-graphs.md)):** For graphs exceeding 40 nodes, a single 2D viewport with directional gutters is insufficient for causal overview. Research overview-plus-focus, minimaps, and semantic zoom as tracked in `6g8vxbv3d4xn`.
2. **TUI Information Architecture Reassessment ([`planning/tasks/6g8vxcnezktm`](file:///planning/tasks/6g8vxcnezktm-reassess-tui-navigation-and-information-architecture-at-current-scale.md)):** The interaction between entity tabs, structured child selection, full-screen zoom, and custom pane split preferences should be harmonized at the shell level to prevent ad-hoc split hooks across new entities.
