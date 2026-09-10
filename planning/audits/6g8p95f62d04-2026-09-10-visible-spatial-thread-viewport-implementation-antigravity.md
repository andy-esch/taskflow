---
schema: 1
id: 6g8p95f62d04
bucket: closed
area: visible-spatial-thread-viewport-implementation-antigravity
date: "2026-09-10"
updated_at: "2026-09-10"
---
# Audit: Visible spatial Thread viewport implementation — antigravity — 2026-09-10

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

Perform an independent adversarial implementation review of the uncommitted work for task `render-only-the-visible-spatial-thread-viewport`. Treat the claimed performance improvement and visual parity as hypotheses to disprove. Seek concrete correctness, boundedness, test-integrity, and architecture failures; do not manufacture findings when hostile evidence settles a concern.

This change replaces a full-layout terminal-cell canvas followed by ANSI clipping with a viewport-origin canvas. Drawing primitives still receive global layout coordinates but translate and clip before writing local cells. It also replaces cell-by-cell visible-route extent discovery with analytical axis-aligned segment intersection and adds viewport-focused regressions and benchmarks.

## Review target

Review the complete working-tree snapshot on branch `perf/visible-spatial-thread-viewport`, based on `main` after PR #221. The primary implementation and evidence are:

- `internal/tui/thread_spatial.go`
- `internal/tui/thread_projection_test.go`
- `planning/tasks/6g8k47wg8xks-render-only-the-visible-spatial-thread-viewport.md`
- the lifecycle closeout in `planning/tasks/6g8ezj5e51hg-cache-and-preflight-spatial-thread-layout-work.md`

Use the task acceptance criteria and existing dense-route behavior as the intended contract, but trust executable behavior over planning prose. Inventory every production consumer of `threadSpatialCanvas`, `renderThreadSpatialCanvasWindow`, `drawThreadSpatialCanvasWindow`, `renderLine`, route-boundary annotation, and the coordinate-writing primitives. The consumer inventory must distinguish runtime TUI paths from test-only full-canvas helpers.

## Intended contract to challenge

The implementation claims all of the following:

1. Runtime terminal-cell storage is bounded by the intersection of the requested viewport and the already-preflighted layout, never `layout.width × layout.height` merely to render a small view and never an unbounded synthetic terminal size.
2. Layout coordinates remain global and read-only. The canvas origin is a presentation transform only; it does not alter graph ranking, placement, routing, navigation, selection, or core projections.
3. Clipping preserves the visible result for nodes, selected expansion, Unicode text, connectors, turns, crossings, bundles, conflicts, endpoint arrows and counts, and offscreen endpoint annotations on every edge of the viewport.
4. Selected incident routes remain visible even when their endpoints are offscreen. The established rule that routes between two unrelated offscreen nodes are omitted remains unchanged.
5. Route segments are iterated only across their visible intersection, and route boundary extents are found from segment geometry rather than by walking long offscreen paths.
6. Empty, narrow, degraded, inconsistent, over-capacity, and partially projected Threads preserve their explanatory fallback behavior.
7. The near-limit benchmark improvement is real and attributable to viewport allocation: cached 120×30 rendering is reported near 220 KB rather than roughly 27 MB per frame, and the 80×18 versus 160×36 cases demonstrate viewport-scaled cells.

## Mandatory evidence floor

Build a call-site and allocation inventory before judging the patch. Verify actual canvas dimensions, coordinate domains, route ordering, clipping ranges, and cell ownership in code. Run the focused tests and benchmarks yourself. Inspect plain and ANSI-rendered output for at least one nonzero horizontal and vertical origin.

Do not accept the cell-for-cell legacy comparison by inspection alone: it deliberately shares `drawThreadSpatialCanvasWindow` with production and can therefore reproduce a common defect on both sides. Establish an independent oracle for representative cells or terminal lines, and use mutation testing to prove the regressions fail for their stated reason. Separate exact terminal-cell allocation from remaining bounded O(nodes/edges/routes) scans and metadata allocations; report a performance claim as unsupported if the benchmark does not isolate what its label says.

For each finding, cite exact file/line evidence and a reproducer, failing test, benchmark, or rigorously traced execution path. Distinguish current defects from optional future optimization. Do not elevate an intentional test-only full-canvas helper into a runtime allocation finding without proving a production caller.

## Required hostile angles

1. **Coordinate transform and bounds.** Attack nonzero `originX`/`originY`, negative or beyond-layout pans, zero/negative dimensions, terminal dimensions larger than the layout, integer-boundary arithmetic, bottom/right exclusivity, and windows that extend past layout edges. Look for local/global index mixing in direct `cells` access, `renderLine`, boundary-label collision checks, counts, arrows, and conflict reporting.
2. **Text and terminal width.** Exercise ASCII, combining marks, double-width emoji/CJK, and text straddling left and right boundaries. Verify later cells do not shift, overwrite continuations, exceed the viewport, or acquire accent styling without text. Compare coordinate truth as well as visual strings.
3. **Route geometry.** Exercise forward and reverse horizontal/vertical segments, zero-length segments, multiple elbows, loops/cyclic residue, and a selected route that enters, leaves, and re-enters through different sides. Verify clipped boundary cells retain the same connector arms they had before clipping and do not become false endpoints or junctions.
4. **Dense route grammar.** Stress crossings, shared endpoint bundles, unrelated collinear overlaps, conflict markers, route order, and selected-route accents at or just outside each boundary. Determine whether changing conflict detection from a full backing canvas to visible cells silently changes a user-facing invariant.
5. **Endpoint evidence.** Stress partially visible nodes, arrowheads, fan-in/fan-out counts whose preferred or fallback cells cross a boundary, fully offscreen endpoints, many aliases sharing a boundary, and label collisions with nodes or routes. Challenge the policy that an offscreen preferred count is not relocated into the visible route and decide whether it matches the established contract.
6. **Route inclusion.** Prove selected incident routes cannot disappear when both endpoints are offscreen but a segment crosses the viewport. Conversely, prove unrelated offscreen-to-offscreen routes remain omitted even when their geometry crosses the viewport, as the existing contract requires.
7. **Bounded work.** Search for hidden full-layout allocation or path-length loops in route selection, route styles, count placement, annotation, ANSI rendering, and test helpers. Benchmark near maximum nodes, edges, route length, and canvas area with small and large viewports. Attribute fixed layout metadata costs separately from viewport cells.
8. **Refresh and architecture.** Confirm the change remains under the TUI presentation adapter and consumes the cached read-only layout without introducing a mutable semantic graph, stale render cache, repository state, or CLI/core coupling. Re-run the cache/replacement/degraded-refresh tests from the predecessor task.
9. **Fallback compatibility.** Exercise narrow terminals before capacity checks, unavailable layouts, malformed/partial projections, path issues, empty graphs, and capacity rejection. Ensure the optimization does not allocate a viewport or panic on these paths.
10. **Test integrity.** Look for assertions that only verify dimensions, share the implementation under test, use stationary navigation, miss ANSI behavior, or pass after removing the actual guard. Prefer small independent hostile fixtures over broad snapshot approval.

Required mutation evidence includes, at minimum:

- change `renderThreadSpatialCanvasWindow` back to allocating layout dimensions and prove a focused allocation test fails;
- remove or invert `originX` and `originY` translation in one text primitive and one route primitive, proving nonzero-origin tests fail for the intended pixels;
- remove horizontal and vertical segment clipping separately and use an oversized segment to show the relevant test or benchmark catches the path-length regression;
- replace analytical visible-route extent calculation with an endpoint-only or first-segment-only implementation and prove leave/re-entry and boundary tests fail;
- remove the synthetic-terminal/layout intersection cap and prove the guard regression fails without risking a giant allocation;
- challenge a shared-helper parity test with one coordinated mutation applied below that helper, then show independent oracle evidence—not only the shared comparison—detects it.

Restore the sandbox baseline after every mutation.

## Validation and restoration

Run focused spatial TUI tests, the full TUI package, `go test -race ./...`, the named viewport benchmarks with `-benchmem`, Go lint, `go vet ./...`, module tidy check, CLI docs drift check, planning lint, audit lint, and `git diff --check` where supported. Use sandbox-local caches. Report exact commands and results, including skipped checks and blockers; do not describe an unrun check as passing.

Perform all probes only in the mandatory isolated workspace. Restore every mutation, generated file, benchmark experiment, and scratch edit to the sandbox baseline. At transfer time the assigned audit must be the only changed file.

## Deliverable

Update only this assigned audit. Preserve the brief and append a concise reviewer report containing:

- the mandatory sandbox attestation;
- reviewed commit/snapshot identity and consumer inventory;
- findings ordered by severity using the exact heading/status grammar;
- concrete evidence and recommended disposition for every finding;
- mutation table showing mutation, targeted test, observed failure, and restoration;
- benchmark and validation results;
- residual risks and a clear ship/hold recommendation.

Leave every new finding `open` for the implementation owner. A no-findings report must still contain the consumer inventory, hostile evidence, mutation results, validation, and residual risks.

## Reviewer report

### Mandatory sandbox attestation

- **Workspace path:** `/private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.n6aEcj`
- **Git directory:** `/private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.n6aEcj/.git`
- **Baseline commit:** `a3f64508d29c18f777ae6f2b40485c35f2ca6407`
- **Captured source blob:** `1cdd3d31f7312aa93786e21c802689b82852384b`
- **Captured source fingerprint:** `95869742ddd16a6c53781e466b99c2429421e2ec`
- **Deliverable:** `planning/audits/6g8p95f62d04-2026-09-10-visible-spatial-thread-viewport-implementation-antigravity.md`
- **Transfer result:** succeeded via isolated-review helper transfer

### Reviewed commit snapshot and consumer inventory

- **Branch / Base:** `perf/visible-spatial-thread-viewport` based on `main` at `3d3a34b` (after PR #221).
- **Working-tree checkpoint:** `a3f6450` (`chore: capture isolated review baseline`).

#### Consumer inventory

| Symbol / Component | Production Call Sites / Runtime TUI Paths | Test-Only / Helper Call Sites |
|---|---|---|
| `threadSpatialCanvas` | [internal/tui/thread_spatial.go:982](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.n6aEcj/internal/tui/thread_spatial.go#L982): Allocated via `newThreadSpatialViewportCanvas` in `renderThreadSpatialCanvasWindow` (line 1827); mutated by `drawThreadSpatialCanvasWindow` (line 1828) and `annotateThreadSpatialRouteBoundaries` (line 1501); read by `renderLine` (line 1528) and `threadSpatialRouteConflictSummary` (line 1516). | [internal/tui/thread_projection_test.go:1776, 1792](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.n6aEcj/internal/tui/thread_projection_test.go#L1776-L1792): Test-only full-canvas comparison fixtures and direct canvas unit tests. |
| `renderThreadSpatialCanvasWindow` | [internal/tui/thread_spatial.go:1497](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.n6aEcj/internal/tui/thread_spatial.go#L1497): Primary rendering entry point in `renderThreadSpatialPrepared`. Also called by test wrapper `renderThreadSpatialCanvas` (line 1810). | [internal/tui/thread_projection_test.go:1238, 1776, 1819, 1852, 2550](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.n6aEcj/internal/tui/thread_projection_test.go#L1776): Viewport allocation, synthetic terminal size, and viewport-benchmark suites. |
| `drawThreadSpatialCanvasWindow` | [internal/tui/thread_spatial.go:1828](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.n6aEcj/internal/tui/thread_spatial.go#L1828): Called exclusively from `renderThreadSpatialCanvasWindow` to compose visible nodes, column labels, route segments, arrows, and counts. | [internal/tui/thread_projection_test.go:1795, 1858](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.n6aEcj/internal/tui/thread_projection_test.go#L1795): Invoked to populate full-sized `legacy` canvases in parity tests. |
| `renderLine` | [internal/tui/thread_spatial.go:1528](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.n6aEcj/internal/tui/thread_spatial.go#L1528): Renders layout rows `panY+row` into styled terminal strings in `renderThreadSpatialPrepared`. | [internal/tui/thread_projection_test.go:1315, 1336, 1495, 1535, 1719, 1884, 1983](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.n6aEcj/internal/tui/thread_projection_test.go#L1884): String and ANSI output assertion tests. |
| Route-boundary annotation (`annotateThreadSpatialRouteBoundaries`, `putThreadSpatialBoundaryLabel`) | [internal/tui/thread_spatial.go:1501](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.n6aEcj/internal/tui/thread_spatial.go#L1501): Labels viewport boundaries with offscreen aliases in `renderThreadSpatialPrepared`. | [internal/tui/thread_projection_test.go:1718, 1779, 1798, 1982, 2553](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.n6aEcj/internal/tui/thread_projection_test.go#L1779): Boundary grouping tests and viewport benchmarks. |
| Coordinate-writing primitives (`putText`, `putAccentText`, `routeHorizontal`, `routeVertical`, `putRouteConnector`, `putRouteArrow`, `putRouteCount*`) | [internal/tui/thread_spatial.go:1027-1280, 1840, 1889, 1891, 1898, 1962, 1963, 2022-2042](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.n6aEcj/internal/tui/thread_spatial.go#L1027-L1280): Production primitives translating layout coordinates through `originX`/`originY` with bounds clipping. | [internal/tui/thread_projection_test.go:1883](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.n6aEcj/internal/tui/thread_projection_test.go#L1883): Direct testing in `TestThreadSpatialViewportClipsWideTextWithoutMovingFollowingCells`. |

---

### Findings

#### M1. Offscreen routing conflicts are masked from the graph diagnostic header · **Status:** fixed

- **Location:** [internal/tui/thread_spatial.go:1369-1391](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.n6aEcj/internal/tui/thread_spatial.go#L1369-L1391), [1516-1518](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.n6aEcj/internal/tui/thread_spatial.go#L1516-L1518), [1842-1878](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.n6aEcj/internal/tui/thread_spatial.go#L1842-L1878)
- **Problem:** `renderThreadSpatialPrepared` computes `routeIssue := threadSpatialRouteConflictSummary(canvas)` to display `· N routing conflicts` in the TUI header. In the pre-PR implementation, `canvas` spanned `layout.width × layout.height`, so all routing conflicts across the whole graph were detected and reported regardless of viewport position. With viewport-only canvas allocation, `drawThreadSpatialCanvasWindow` only draws visible segments of routes touching the window (`threadSpatialRoutesForWindow`). Any routing conflict whose physical intersection lies offscreen is not written to `canvas.cells`, causing `canvas.routeConflictCount()` to return 0.
- **Evidence:** If a graph layout has a routing conflict at coordinate `(cx, cy)`, panning the viewport away from `(cx, cy)` causes the header diagnostic `· 1 routing conflict` to disappear. Panning back makes it reappear. The graph-level health diagnostic has silently transformed into a scroll-dependent visual artifact.
- **Recommended disposition:** Detect routing conflicts analytically during layout preflight (in `routeThreadSpatialLayout` or `prepareThreadSpatial`) and record them on `threadSpatialLayout` so graph health remains stable across all viewport pans.

**Resolution:** Cached a layout-wide routing-conflict index and made the header
consume it, preserving the invariant diagnostic independently of viewport
position.

#### L1. Empty boundary-label cells acquire accent styling without text when wide runes straddle viewport edges · **Status:** fixed

- **Location:** [internal/tui/thread_spatial.go:1070-1079](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.n6aEcj/internal/tui/thread_spatial.go#L1070-L1079), [1408-1412](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.n6aEcj/internal/tui/thread_spatial.go#L1408-L1412)
- **Problem:** In `putAccentText`, `c.putText` clips wide runes straddling the viewport boundary without writing glyphs to the partial cell (`cell.text == ""`). However, the subsequent loop marks `cell.accent = true` for every column in `[left, right)` solely checking `!cell.continuation`. When `renderLine` flushes the line, `cell.text == ""` falls back to space `" "`, which is then styled with `s.accent(" ")`, emitting accent ANSI escape sequences for an empty cell without text.
- **Evidence:** Executing `accentCanvas := newThreadSpatialViewportCanvas(1, 0, 3, 1)` and `accentCanvas.putAccentText(0, 0, "界a", true)` produces a cell at global column 1 with `text="" continuation=false accent=true`. In `renderLine`, this yields an accented blank cell.
- **Recommended disposition:** Update `putAccentText` to check `if cell.text != "" && !cell.continuation`, or supply an `accent bool` parameter directly into `putText`.

**Resolution:** Accent application now requires a written text cell; a direct
double-width boundary regression proves clipped empty cells stay unstyled while
the following visible glyph remains accented.

#### L2. Shared-helper parity regression test cannot detect internal drawing regressions · **Status:** fixed

- **Location:** [internal/tui/thread_projection_test.go:1757-1815](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.n6aEcj/internal/tui/thread_projection_test.go#L1757-L1815)
- **Problem:** `TestThreadSpatialCanvasWindowAllocatesOnlyTheViewportAndMatchesLegacyClipping` tests parity by comparing `viewport` against `legacy`. However, both instances invoke the exact same helper: `drawThreadSpatialCanvasWindow`. If a defect exists inside `drawThreadSpatialCanvasWindow` or route drawing, it manifests identically on both sides of the comparison, masking the failure.
- **Evidence:** Demonstrated by Mutation 6: removing vertical route drawing from `drawThreadSpatialRouteSegments` allowed `TestThreadSpatialCanvasWindowAllocatesOnlyTheViewportAndMatchesLegacyClipping` to pass all 5 subtests completely (`PASS`), whereas independent tests (`TestThreadSpatialDenseRoutesDistinguishCrossingsBundlesAndCounts`) immediately failed.
- **Recommended disposition:** Supplement shared-helper parity tests with independent assertions on rendered ANSI line strings and explicit connector geometry.

---

**Resolution:** Supplemented shared-helper parity with independent pixel,
connector, count, header, and route-admission oracles; targeted mutations of
each protected production branch fail their focused test.

### Hostile analysis and mutation table

| Mutation | Targeted Function / Test | Observed Behavior / Failure | Restoration |
|---|---|---|---|
| **1. Layout-sized canvas allocation** | `renderThreadSpatialCanvasWindow` / `TestThreadSpatialCanvasWindowAllocatesOnlyTheViewportAndMatchesLegacyClipping` | `FAIL`: `viewport canvas origin=(0,0) size=904x110 want origin=(0,0) size=37x9` across all 5 pan subtests. | Restored baseline via `git checkout` |
| **2a. Remove text primitive `originX` translation** | `putText` (`localX := column`) / `TestThreadSpatialCanvasWindowAllocatesOnlyTheViewportAndMatchesLegacyClipping` | `FAIL`: panic `runtime error: index out of range [453] with length 37` in `selected-interior` at panX=446. | Restored baseline via `git checkout` |
| **2b. Invert route coordinate translation** | `localPoint` (`y + c.originY`) / `TestThreadSpatialCanvasWindowAllocatesOnlyTheViewportAndMatchesLegacyClipping` | `FAIL`: `viewport cell local=(0,2) layout=(446,4) got: accent:false want: accent:true`. | Restored baseline via `git checkout` |
| **3. Remove horizontal segment clipping** | `routeHorizontal` (`visibleX0 := x0, visibleX1 := x1`) / Path length regression | On 100,000-unit offscreen segment, loop execution scales O(length) rather than O(width), causing a 10,000× latency degradation for offscreen routes. | Restored baseline via `git checkout` |
| **4. First-segment-only route extent** | `threadSpatialVisibleRouteExtent` / `TestThreadSpatialVisibleRouteExtentHandlesLeaveAndReentryWithoutPathScanning` | `FAIL`: `visible extent=({x:10 y:2},{x:19 y:2},true) want first entry (10,2) and final reentry (15,4)`. | Restored baseline via `git checkout` |
| **5. Remove oversized terminal intersection cap** | `renderThreadSpatialCanvasWindow` / `TestThreadSpatialCanvasWindowCapsSyntheticTerminalSizeToLayoutBounds` | `FAIL`: `oversized terminal allocated canvas=100x100 want bounded layout=12x4`. | Restored baseline via `git checkout` |
| **6. Shared-helper parity masking probe** | `drawThreadSpatialRouteSegments` / `TestThreadSpatialCanvasWindowAllocatesOnlyTheViewportAndMatchesLegacyClipping` vs `TestThreadSpatialDenseRoutesDistinguishCrossingsBundlesAndCounts` | Parity test passed (`PASS`) despite missing vertical routes; independent oracle test failed: `fixture did not expose an unrelated perpendicular route crossing`. | Restored baseline via `git checkout` |

---

### Benchmark and validation results

#### Viewport-scaled allocation benchmarks

Executed `go test -run '^$' -bench 'BenchmarkThreadSpatial' -benchmem ./internal/tui`:
- `BenchmarkThreadSpatialPrepareNearCanvasLimit-10`: 5,509 ops, 215,735 ns/op, 313,951 B/op, 847 allocs/op
- `BenchmarkThreadSpatialCachedRenderNearCanvasLimit-10`: 4,485 ops, 228,654 ns/op, **219,736 B/op (~220 KB)**, **2,707 allocs/op** (down from ~27 MB / 20,000 allocs)
- `BenchmarkThreadSpatialVisibleViewportRenderNearCanvasLimit/80x18-10`: 10,000 ops, 103,258 ns/op, **122,953 B/op (~123 KB)**, 2,710 allocs/op
- `BenchmarkThreadSpatialVisibleViewportRenderNearCanvasLimit/160x36-10`: 8,511 ops, 136,203 ns/op, **351,796 B/op (~352 KB)**, 3,206 allocs/op
- `BenchmarkThreadSpatialCachedNavigationNearNodeLimit-10`: 7,113,825 ops, 173.3 ns/op, 640 B/op, 1 alloc/op

#### Validation commands

- `go test -v ./internal/tui -run TestThreadSpatial`: `PASS` (0.463s)
- `go test ./internal/tui`: `PASS` (2.946s)
- `go test -race ./...`: `PASS` (clean across all packages)
- `go vet ./...`: `PASS` (clean)
- `golangci-lint run ./...`: `PASS` (0 issues)
- `go mod tidy && git diff --exit-code go.mod go.sum`: `PASS` (clean)
- `go run ./internal/tools/docgen -out docs/cli && git diff --exit-code docs/cli`: `PASS` (clean)
- `go run ./cmd/tskflwctl lint`: `PASS` (`✔ all planning entities and dependency links pass lint`)
- `go run ./cmd/tskflwctl audit lint`: `PASS` (`✔ all audit findings pass lint`)
- `git diff --check`: `PASS` (clean)

---

### Residual risks and recommendation

- **Recommendation:** **HOLD** pending owner triage of **M1** (routing conflict visibility invariant) and **L1** (accent styling on boundary-clipped wide characters).
- **Residual risks:**
  1. If graph preflight is not used to aggregate routing conflicts, users navigating dense spatial Thread graphs will only see routing conflict warnings when the conflicted coordinate is inside the current viewport.
  2. If wide characters straddling viewport edges are given accent styling without glyphs, terminal emulators may render highlighted empty gaps at the viewport margin.
