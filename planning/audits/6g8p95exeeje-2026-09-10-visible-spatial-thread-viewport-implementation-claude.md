---
schema: 1
id: 6g8p95exeeje
bucket: closed
area: visible-spatial-thread-viewport-implementation-claude
date: "2026-09-10"
updated_at: "2026-09-10"
---
# Audit: Visible spatial Thread viewport implementation — claude — 2026-09-10

> Reviewer assignment: claude. This document is the review brief and the only file the reviewer should update.
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

### Isolation attestation

All inspection, builds, tests, benchmarks, generation, mutation probes and report editing were
performed inside the mandatory isolated workspace. The handoff checkout was read only for this
brief and for the helper's initial copy; no write-capable command ran there.

```
sandbox_path=/private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.Tuy1yG
git_dir=/private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.Tuy1yG/.git
baseline_commit=efe995ff3ccba007a74737a37940b7e4d0b36173
source_blob=0e2881c9b8c12c593afc654155084f8458867c99
source_fingerprint=95869742ddd16a6c53781e466b99c2429421e2ec
deliverable=planning/audits/6g8p95exeeje-2026-09-10-visible-spatial-thread-viewport-implementation-claude.md
deliverable_changed=true
transfer=succeeded
```

Independence verified directly: `git rev-parse --absolute-git-dir` resolves inside the sandbox,
`.git/objects/info/alternates` is absent, and `git worktree list --porcelain` names exactly the one
sandbox checkout. Every mutation was reverted with `git checkout -- internal/tui/thread_spatial.go`
and re-verified against the baseline commit; all scratch probe files (`internal/tui/zz_*.go`) were
deleted. `git status --porcelain --untracked-files=all` was empty before this report was written.

An independent reference tree was cloned **from the sandbox** (never from the handoff checkout) at
`/private/tmp/claude-501/.../scratchpad/reference`, checked out at `3d3a34b` (the pre-change merge
of PR #221). It is outside both repositories and is the oracle referenced throughout.

### Reviewed snapshot

- Branch `perf/visible-spatial-thread-viewport`, working-tree snapshot captured as sandbox baseline
  `efe995f`, parent `3d3a34b` ("Merge pull request #221 from andy-esch/perf/cache-spatial-thread-layout").
- Diff under review: `internal/tui/thread_spatial.go` (+243/-66), `internal/tui/thread_projection_test.go`
  (+267), plus the two planning files and the two audit briefs.
- Toolchain: go1.26.6 darwin/arm64, golangci-lint present, sandbox-local `GOCACHE`.

### Consumer inventory

Runtime path (single chain; `internal/tui/detail.go:880` is the only production entry):

| Symbol | Production call sites | Notes |
|---|---|---|
| `renderThreadSpatialPrepared` | `detail.go:880` | the one runtime consumer |
| `renderThreadSpatialCanvasWindow` (`thread_spatial.go:1816`) | `thread_spatial.go:1497` | allocates the viewport canvas |
| `drawThreadSpatialCanvasWindow` (`:1832`) | `:1828` only | shared with the parity test |
| `annotateThreadSpatialRouteBoundaries` (`:1535`) | `:1501` | post-composition pass |
| `renderLine` (`:1393`) | `:1528` | `ansi.Cut` removed at this site |
| `threadSpatialRoutesForWindow` (`:1858`) | `:1841` | route admission |
| `routeHorizontal`/`routeVertical` (`:1240`/`:1261`) | `:1888`/`:1890` | clipped iteration |
| `putText` (`:1027`) | 10 sites (`drawThreadSpatialNode`, column labels, inspector-adjacent) | |
| `putAccentText` (`:1070`) | `:1781` **only** (boundary labels) | |
| `putRouteConnector`/`putRouteArrow`/`putRouteCount*` | `:1252`/`:1273`, `:1899`, `:1958`/`:1959` | |
| `cellAt`/`localPoint`/`containsY` | 6 / 3 / 2 | the single clipping choke point |
| `routeConflictCount` (`:1369`) | `:1382` → header at `:1516` | **now viewport-scoped** (see M1) |
| `threadSpatialVisibleRouteExtent` (`:1667`) | `:1582` | analytical replacement |

Test-only full-canvas helpers — **verified zero production callers**, so their `layout.width ×
layout.height` allocation is correctly *not* a runtime finding:

- `newThreadSpatialCanvas` (`:990`): 0 production, 14 test call sites.
- `renderThreadSpatialCanvas` (`:1809`): 0 production, 6 test call sites.
- `renderThreadSpatial` (`:1453`): 0 production callers; a test/oracle wrapper over
  `renderThreadSpatialPrepared`, which is why the old-vs-new oracle below exercises the true
  production composition path.

### Independent oracle

The brief forbids trusting the cell-for-cell legacy comparison at `thread_projection_test.go:1757`,
because its "legacy" canvas is built by the **same** `drawThreadSpatialCanvasWindow` at origin
`(0,0)`. I therefore built an oracle that does not share code with the patch at all: the pre-change
implementation in the reference tree. A dump harness (identical source in both trees) rendered
**3,144 production frames** through `renderThreadSpatial` — 8 fixtures (empty, single node, ASCII
chain, wide/CJK/emoji chain, the near-canvas benchmark fixture, a dense fan-in/fan-out graph with a
self loop and a 2-cycle, a long-route chain, and a degraded/inconsistent projection) × every
selection × 12 terminal sizes (including `threadSpatialMinWidth-1` and `threadSpatialMinHeight-1`)
× with and without a path issue — and the two dumps were diffed.

Result: 570 frames differ only in trailing whitespace (a consequence of `renderLine` trimming inside
the viewport instead of the full layout row; visually identical). **32 frames differ substantively**,
in exactly three classes — one improvement (below), and the two changes reported as M2 and M3.

**Verified improvement — left-edge double-width alignment is now correct.** Where a double-width
glyph straddles `panX`, the removed `ansi.Cut(line, panX, panX+width)` re-emitted the whole glyph,
shifting the entire row one column right. `wide-chain sel="chain-02" 60x14`:

```
row5  ──────────┐        │ ● éclair café näi… │        ┌──────────   (both)
row6  ビュー完…  │━━━━━━━▶│ [M3] next-up       │━━━━━━━▶│ ● plain…   OLD  ← │ at col 11
row6   ュー完…  │━━━━━━━▶│ [M3] next-up       │━━━━━━━▶│ ● plain 0  NEW  ← │ at col 10
row7  ──────────┘        │  /                 │        └──────────   (both)
```

In OLD the node's right border (col 11) no longer lines up with its own `┐`/`┘` corners (col 10);
NEW restores the box. `TestThreadSpatialViewportClipsWideTextWithoutMovingFollowingCells`
(`:1881`) pins this. Good change, correctly justified by the comment at `thread_spatial.go:1049-1051`.

### Findings

#### M1. The routing-conflict diagnostic is now viewport-scoped, so a real renderer-invariant failure disappears from the header when it is panned offscreen · **Status:** fixed

`routeConflictCount` (`thread_spatial.go:1369-1378`) counts `conflict` cells by walking `c.cells`.
Before the change `c.cells` was the whole layout, so every conflict along a windowed route was
counted; now `c.cells` is only the viewport, and `putRouteConnector` (`:1085-1088`) never writes an
offscreen cell, so the conflict is never even detected. The header at `:1516-1519` consumes that
count. A `conflict` cell is documented at `:1119-1122` as "a routing invariant failure, not a
legitimate crossing" — a diagnostic that exists precisely to expose renderer bugs.

**Reproduction (production path, old-vs-new).** Randomized projections do reach real conflicts
(3 distinct fixtures found in 4,000 trials). Rendering them through `renderThreadSpatial` in both
trees and comparing the header:

```
                                            OLD (3d3a34b)          NEW (efe995f)
trial=3046 sel=t03 160x40    conflicts=  1 routing conflict         none        <-- regression
trial=3046 sel=t03 200x60    conflicts=  2 routing conflicts        2 routing conflicts
trial=3046 sel=t03 400x60    conflicts=  3 routing conflicts        3 routing conflicts
```

Isolated on the canvas primitives, using the exact geometry of the existing regression test
`TestThreadSpatialRoutingConflictIsNeutralAndReportedOutOfBand` (`:1387`):

```
full-layout canvas summary   = "1 routing conflict"
viewport(originX=4) summary  = ""
viewport(originY=4) summary  = ""
```

**Why no test catches it.** `:1387` builds its canvas with `newThreadSpatialCanvas(9, 8)` — origin
`(0,0)` — so it cannot observe the scope change. Nothing else asserts on the summary.

**Note on prior art.** The count was *already* window-dependent before this change, because
`threadSpatialRoutesForWindow` only ever drew windowed routes. This finding is the further
narrowing from "conflicts anywhere along a windowed route" to "conflicts inside the viewport
rectangle", demonstrated above at 160×40.

**Recommended disposition.** Decide the intended contract and pin it. Either (a) accept
viewport-scoped counting and add a regression at nonzero origin documenting that the header reports
*visible* conflicts, or (b) preserve the layout-wide diagnostic by detecting conflicts from route
geometry rather than from composed cells. Silent narrowing of a renderer-invariant alarm is the
part worth deciding deliberately.

**Resolution:** Cached a layout-wide routing-conflict index built by replaying
only route turn candidates through the production merge grammar; the header is
now stable across pans, with curated and 300-layout composition-oracle
regressions.

#### M2. Clipping an endpoint's preferred count cell now triggers the node-border congestion fallback, so a fan count appears on a node that the pre-change renderer left bare · **Status:** fixed

`putRouteCountAlong` gained an early return when the preferred endpoint cell is outside the viewport
(`thread_spatial.go:1203-1208`). Its comment says the boundary annotation "owns that offscreen
evidence" — but the caller treats `false` as "stub congested" and falls through to
`putRouteCountFallback` (`drawThreadSpatialRouteCounts`, `:1955-1960`), which writes the count onto
the node's top border (`threadSpatialCountFallback`, `:1968-1991`) with an unconditional
`*cell = threadSpatialCell{...}` (`:1233-1235`). So the count is not withheld; it is relocated.

**Reproduction (production path, old-vs-new).** Dense fixture, `sel=""` (resolves to `c`), 60×14:

```
OLD   │ ● node c           │           ┌────────────────────┐
NEW   │ ● node c           │           ┌──────────2─────────┐
```

Traced exactly: `panX=0, panY=1, width=60, canvasHeight=6`. Node `a` sits at `x=36,y=2` (visible);
routes `a→h` and `a→hub` share first segment origin `{58,4}`, so the preferred count cell is
`{x:60, y:4}` — one column outside the 60-wide viewport. OLD wrote the `2` at global `(60,4)` on the
full canvas and cropped it away, so `putRouteCountAlong` reported success and no fallback ran. NEW
returns `false` and the fallback paints `2` at `{x:47, y:3}`, node `a`'s top-border centre.

**Effect on the grammar.** In the established dense-route semantics the border count means "every
safe stub cell was congested" (`thread_projection_test.go:1368-1383`). It now also means "the stub
was scrolled off", and the two are indistinguishable to a reader; counts will appear and disappear
on node borders purely as a function of pan position.

**Sub-concern settled, not a defect.** I checked whether a relocated count can under-report: it
cannot. `threadSpatialRoutesForWindow` (`:1858-1878`) admits every route incident to a *visible*
node, and the fallback only fires on a node whose border is inside the canvas, so the tally is
complete.

**Why no test catches it.** `TestThreadSpatialViewportPreservesPartiallyClippedEndpointCounts`
(`:1828`) was written for exactly this boundary, but it compares against a legacy canvas built by the
same `drawThreadSpatialCanvasWindow` at origin `(0,0)`; in its fixture both sides agree, so the
divergence never appears. No fixture places a preferred count cell just outside the viewport while
its node stays visible.

**Recommended disposition.** Decide whether clipping should relocate the count. If yes, keep it and
add a regression fixing the behaviour (and consider reconciling the comment at `:1203-1205`, which
states the opposite intent). If no, have the caller distinguish "clipped" from "congested" so the
fallback is skipped.

**Resolution:** Distinguished placed, clipped, and congested count placement
outcomes so a clipped endpoint count is omitted instead of triggering the
node-border congestion fallback; added a mutation-killing boundary regression.

#### M3. Vertical route clipping's `originY` translation is unpinned: removing it passes the entire test suite while silently erasing route strokes from 364 of 3,144 production frames · **Status:** fixed

`routeVertical` (`thread_spatial.go:1265-1270`) clips against `c.originY`; `routeHorizontal`
(`:1245-1249`) clips against `c.originX`. The horizontal translation **is** pinned — mutating
`max(x0, c.originX)/min(x1, c.originX+c.width-1)` to `max(x0, 0)/min(x1, c.width-1)` fails
`TestThreadSpatialVisibleRouteExtentHandlesLeaveAndReentryWithoutPathScanning` (`:1892`) and
`TestThreadSpatialViewportPreservesPartiallyClippedEndpointCounts`. The vertical one is not:

```
mutation: visibleY0 := max(y0, 0); visibleY1 := min(y1, len(c.cells)-1)   [originY dropped]
go test ./internal/tui/   ->  ok   github.com/andy-esch/taskflow/internal/tui   3.125s
independent old-vs-new oracle -> 364 / 3144 production frames differ
example (dense 60x12, row 5):
  BASE  │ [M3] in-progress   │━━━┓       │ ● node a           │◀═
  MUT   │ [M3] in-progress   │━━━━       │ ● node a           │◀═
```

The elbow `┓` degrades to `━`: the vertical stroke was never drawn, so the route silently loses its
turn. Nothing panics or index-errors because `cellAt` (`:1015-1022`) fail-closes on every write —
the cell-level guard converts a coordinate-translation bug into invisible data loss, which is
exactly the "boundaries that only appear to fail closed" failure mode.

**Root cause of the blind spot.** Every new nonzero-origin regression uses a nonzero `originX` with
`originY = 0`: `:1892` uses `newThreadSpatialViewportCanvas(10, 0, 10, 10)`, `:1881` uses
`(2, 0, 3, 1)`. The one test that composes vertical strokes at nonzero `originY` —
`TestThreadSpatialViewportAnnotatesVerticalOffscreenEndpoints` (`:1963`, origin `(10,10)`) — calls
`drawThreadSpatialRouteSegments` and then asserts **only** on boundary-label text, never on the
route cells it just drew, so the stroke can vanish entirely and the test still passes. And
`TestThreadSpatialCanvasWindowAllocatesOnlyTheViewportAndMatchesLegacyClipping` (`:1757`) does use
nonzero `panY`, but its comparison side is built at origin `(0,0)`, where the mutation is the
identity — the shared helper reproduces the correct result on the reference side and the defective
result only on the side under test, and its fixture's vertical extents happen not to reach the
diverging rows.

**Recommended disposition.** Add a cell-level assertion at nonzero `originY` — the cheapest fix is to
assert in `:1963` that the vertical segment's visible cells actually carry connectors, mirroring the
`originX` coverage at `:1892`.

**Resolution:** Added direct connector assertions for every visible cell of a
selected vertical route at nonzero originY; dropping the translation now fails
the focused test.

#### L1. Segment clipping is a 200× optimisation with no regression guard of any kind · **Status:** fixed

Removing the clipping ranges from both `routeHorizontal` and `routeVertical` (iterating the whole
segment and letting `cellAt` reject each write) passes the full TUI package and is invisible to the
repository's own benchmarks:

```
                                   shipped        clipping removed
CachedRenderNearCanvasLimit     270879 ns/op        253891 ns/op     (219773 vs 219806 B/op)
VisibleViewportRender/80x18     104348 ns/op        105423 ns/op
VisibleViewportRender/160x36    147512 ns/op        137193 ns/op
```

The near-canvas fixture's routes are short relative to the layout, so path-length work is noise there.
On the case the clipping exists for — oversized segments crossing a small viewport (20 horizontal +
20 vertical segments of length 100,000 into an 80×18 window at origin (4000,4000)) — it is worth
200×:

```
with clipping (shipped)   25568 ns/op    74229 B/op   19 allocs/op
clipping removed        5114986 ns/op    74176 B/op   19 allocs/op
```

Only the *extent* calculation is pinned (`threadSpatialVisibleSegmentExtent`, `:1915`); the drawing
clip is not. A future refactor can reintroduce O(path-length) scanning with a green suite.

**Recommended disposition.** Add one bounded assertion or benchmark over an oversized segment, or
accept the gap explicitly.

**Resolution:** Route drawing now returns its composed-cell count, and a
deterministic oversized-segment regression proves horizontal and vertical work
is bounded to viewport width and height; removing either clip fails.

#### L2. The checkpoint's per-viewport figures conflate cell storage with total per-op allocation, and the benchmark's `viewport-cells/op` metric is a restatement of its inputs · **Status:** fixed

`planning/tasks/6g8k47wg8xks-...md` states: "Direct canvas benchmarks allocate exactly 1,440 cells
for 80×18 (about 123 KB) and 5,760 cells for 160×36 (about 352 KB)." The cell counts are right, but
`sizeof(threadSpatialCell)` is 48 bytes (measured in-sandbox), so the cells are 69,120 B and
276,480 B. The 123 KB / 352 KB figures are the whole benchmark op (122,952 and 351,822 B/op),
which also carries route metadata, styles and render strings. The brief asks explicitly that exact
terminal-cell allocation be separated from the remaining bounded scans; as written the prose reads
as though cells account for all of it.

Relatedly, `b.ReportMetric(float64(size.width*size.height), "viewport-cells/op")`
(`thread_projection_test.go:2548`) reports the product of the benchmark's own inputs. It is a label,
not a measurement — it cannot move when allocation regresses, and the benchmark's only real guard is
the `canvas.width`/`len(canvas.cells)` assertion at `:2556`.

The underlying claim is nonetheless **supported**: cells are 56% of the 80×18 op and 79% of the
160×36 op, and the size-to-size delta (228,870 B) is dominated by the extra 4,320 cells
(207,360 B). Memory does scale with the viewport.

**Recommended disposition.** Correct the two parenthetical byte figures in the checkpoint, or state
that they are total per-op allocations.

**Resolution:** Removed the derived viewport-cells benchmark metric and rewrote
the checkpoint to separate exact 48-byte cell storage from total benchmark
allocation.

#### L3. The positive half of the incident-route contract is unpinned: dropping the selected-route clause passes the whole suite · **Status:** fixed

Contract item 4 and the ticked acceptance criterion "Panning and resizing cannot ... silently omit an
incident route that the existing visibility contract requires" depend on the
`route.edge.From == selected || route.edge.To == selected` clause in `threadSpatialRoutesForWindow`
(`:1868-1869`). Removing that clause passes `go test ./internal/tui/` cleanly.

`TestThreadSpatialViewportOmitsRoutesBetweenTwoOffscreenNodes` (`:1739`) looks like the guard and does
kill the opposite mutation (admitting every route), but its fixture places the selected node `focus`
at `x=63` inside the `panX=50, width=40` window, so `visibleNode("focus")` is already true and the
selected-route clause is never load-bearing. No fixture has a selected node **offscreen** with a
segment crossing the viewport.

The clause itself is unchanged by this diff, so this is a pre-existing coverage gap — but it is a gap
under an acceptance criterion this task ticks and claims regression evidence for.

**Recommended disposition.** Move the fixture's `focus` node outside the window (e.g. `x=3` with
`panX=50`) in a second sub-case so the clause is actually exercised.

**Resolution:** Added an offscreen-endpoints fixture whose selected incident
route crosses the viewport; removing the selected-route admission clause now
fails.

### Mutation results

Every mutation was applied inside the sandbox and reverted to `efe995f` immediately after
measurement; the tree was confirmed clean after the last one.

| # | Mutation | Targeted evidence | Observed | Restored |
|---|---|---|---|---|
| 1 | `renderThreadSpatialCanvasWindow` allocates `layout.width × layout.height` | `...MatchesLegacyClipping`, `...PartiallyClippedEndpointCounts` | **FAIL** — `size=904x110 want 37x9` on all 5 windows | ✓ |
| 2a | `putText` drops the `originX` translation | `...ClipsWideTextWithoutMovingFollowingCells` + 2 others | **FAIL** — `="a界" want " bc"`, plus 5 cell mismatches | ✓ |
| 2b | `routeVertical` drops the `originY` translation | whole TUI package | **PASS (survives)** — oracle: 364/3144 frames corrupted → **M3** | ✓ |
| 3a | `routeHorizontal` drops the `originX` translation | whole TUI package | **FAIL** — `...LeaveAndReentry...` 4 points + endpoint counts; oracle 1046/3144 | ✓ |
| 3b | horizontal segment clipping removed entirely | whole TUI package + benchmarks | **PASS (survives)**, benchmarks flat → **L1** | ✓ |
| 3c | vertical segment clipping removed entirely | whole TUI package + benchmarks | **PASS (survives)**, benchmarks flat → **L1** | ✓ |
| 3d | 3b+3c together, oversized-segment benchmark | scratch benchmark | 25,568 → 5,114,986 ns/op (200×) | ✓ |
| 4a | visible extent = first visible segment only | `...LeaveAndReentryWithoutPathScanning` | **FAIL** — `last=(19,2)` want re-entry `(15,4)` | ✓ |
| 4b | visible extent = raw route endpoints only | same | **FAIL** — `first=(0,2)` want clipped entry `(10,2)` | ✓ |
| 5 | layout-intersection cap replaced by a fixed 512 cap (bounded substitute, so the probe cannot request 10¹² cells) | `...CapsSyntheticTerminalSizeToLayoutBounds` | **FAIL** — `canvas=512x512 want 12x4` | ✓ |
| 6a | coordinated, below the shared helper: column labels drawn at `y=-1` | shared parity test, then whole package | **PASS (survives) — including the parity test**; oracle 876/3144 | ✓ |
| 6b | coordinated: window admits offscreen↔offscreen routes | `...OmitsRoutesBetweenTwoOffscreenNodes` | **FAIL** — both routes returned | ✓ |
| 6c | coordinated: selected-incident clause removed | whole TUI package | **PASS (survives)** → **L3** | ✓ |

Mutations 2b, 6a and 6c are the answer to the brief's shared-helper challenge: a defect introduced
*below* `drawThreadSpatialCanvasWindow` is reproduced identically on both sides of
`...MatchesLegacyClipping` (6a) or is an identity transform on its origin-`(0,0)` reference side
(2b), so the parity test cannot see it. In all three cases only the pre-change reference tree
detected the change. The parity test is a useful anti-drift check but must not be read as coverage.

### Benchmark results

`go test ./internal/tui/ -run '^$' -bench ... -benchmem -benchtime 200x`, go1.26.6 darwin/arm64.

| Benchmark | Tree | ns/op | B/op | allocs/op |
|---|---|---|---|---|
| `CachedRenderNearCanvasLimit` (120×30) | reference `3d3a34b` | 1,951,395 | 27,235,855 | 20,724 |
| `CachedRenderNearCanvasLimit` (120×30) | under review | 270,879 | 219,773 | 2,708 |
| `VisibleViewportRenderNearCanvasLimit/80x18` | under review | 104,348 | 122,952 | 2,710 |
| `VisibleViewportRenderNearCanvasLimit/160x36` | under review | 147,512 | 351,822 | 3,206 |

The headline claim is **real and attributable**: 27.2 MB → 220 KB (124×) and 20,724 → 2,708 allocs
per frame, measured on both trees from the same fixture. Attribution separated as the brief requires:
`sizeof(threadSpatialCell) = 48 B`, so 80×18 = 1,440 cells = 69,120 B (56% of that op) and
160×36 = 5,760 cells = 276,480 B (79%); the 228,870 B delta between the two viewport sizes is
dominated by the 207,360 B of additional cells. The remainder is bounded O(nodes/edges/routes)
metadata and render strings, which does **not** scale with viewport area. See L2 for the checkpoint's
prose imprecision on these numbers.

### Settled hostile angles (challenged, no finding)

- **Clipped connector arms (angle 3).** A segment crossing all four edges keeps its arms: with
  `originX=10, originY=10, 6×6`, a horizontal run `0→100` at `y=12` and a vertical run `0→100` at
  `x=12` render `──┼───`; the left/right/top/bottom boundary cells are `─`/`│`, not endpoints or
  turns. `routeHorizontal`/`routeVertical` compute `directions` from the **unclipped** `x0/x1`,
  `y0/y1` (`:1250-1256`, `:1271-1277`), which is what makes this correct.
- **Route geometry shapes.** Over 34,798 segments from 1,500 randomized production layouts: 21,502
  horizontal, 13,296 vertical, **0 zero-length, 0 diagonal**. The `default:` branch of
  `threadSpatialVisibleSegmentExtent` (`:1725-1727`) is unreachable in production, so replacing the
  cell walk with analytical intersection loses nothing. Forward/reverse traversal order is pinned at
  `:1915`.
- **Accent without text.** `putAccentText` can accent a cell `putText` refused (`abcd界ef` into a
  5-wide viewport leaves `x=4` empty with `accent=true`). Not reachable: `putAccentText` has exactly
  one caller, `putThreadSpatialBoundaryLabel:1781`, whose `available()` guard (`:1766-1779`)
  guarantees the whole label fits inside the canvas, and boundary labels are generated aliases
  (`[M1,M2]…▶`) with no double-width runes.
- **Combining marks at boundaries.** A combining mark whose base rune is clipped is dropped rather
  than reattached to the wrong cell (`last = -1` at `:1046`/`:1055`); a wide base clipped at the left
  edge yields `" xyz"`, preserving downstream coordinates. Verified for precomposed, decomposed, and
  wide-base-plus-combining sequences.
- **Integer boundaries.** `renderThreadSpatialCanvasWindow(..., math.MaxInt, math.MaxInt)` returns a
  12×4 canvas with no panic, and the subsequent annotate pass and segment-extent call behave. The
  layout intersection at `:1825-1826` is what makes this safe (mutation 5).
- **Layout smaller than the terminal.** `annotateThreadSpatialRouteBoundaries` and
  `putThreadSpatialBoundaryLabel` receive the raw terminal `width`/`height` while the canvas may be
  narrower, but `putThreadSpatialBoundaryLabel` re-clamps to `canvas.originX + canvas.width`
  (`:1752-1754`) and `'b'`-side rows can never be selected when `height > layout.height`. Confirmed
  by oracle frames at 200×60 and 400×80 over small layouts: no divergence.
- **Fallback compatibility (angle 9).** Oracle frames at 10×6, 59×20 (`threadSpatialMinWidth-1`),
  45×11 (`threadSpatialMinHeight-1`), plus the empty, degraded/inconsistent and single-node fixtures
  are byte-identical to the reference tree. `TestThreadSpatialNarrowFallbackAllocatesNoCanvasBeforeCapacityChecks`
  (`:2581`) still passes.
- **Architecture and refresh (angle 8).** The diff introduces no package-level `var`, `const` or
  `init` (`git diff 3d3a34b efe995f -- internal/tui/thread_spatial.go | grep -E '^\+(var|func init|const)'`
  → empty). The change stays inside the TUI presentation adapter; no core, CLI, store or repository
  symbol is touched. All 22 predecessor cache/refresh/degraded/narrow/capacity tests pass, including
  `TestThreadSpatialCacheFollowsCoherentProjectionReplacement`,
  `TestThreadSpatialCacheRemainsLazyUntilSpatialView` and
  `TestThreadSpatialRuntimeMethodsConsumeTheCachedPreparedResult`.
- **`var routes` instead of a preallocated slice** (`:1866`). The comment's rationale holds; the
  append-growth cost is invisible at 2,708–3,206 allocs/op, and for a visible hub node every incident
  route is admitted anyway.
- **`putRouteCountAlong`'s new `break`** (`:1213-1215`). Candidates march monotonically by
  `step * attempt`, so `break` and `continue` are equivalent; `step == 0` is already excluded by the
  pre-loop check.

### Residual risks (not findings)

- The three substantive oracle divergence classes were found from 3,144 frames over 8 fixtures. That
  is broad but not exhaustive; fixtures with heavier crossing/bundle density at boundaries could
  surface more. The oracle harness is cheap to re-run against any future refactor.
- M1's production reachability was demonstrated with randomized projections, not with a
  repository-realistic Thread. Conflicts are rare (3 fixtures in 4,000 trials) but not synthetic.
- The `ansi.Cut` removal changes trailing-whitespace behaviour on 570 frames. All are visually
  identical in a terminal, but any downstream consumer that compares rendered lines byte-for-byte
  would see them.

### Validation results

All commands run inside the sandbox with a sandbox-local `GOCACHE`.

| Check | Command | Result |
|---|---|---|
| Build | `go build ./...` | pass |
| Vet | `go vet ./...` | pass |
| Focused spatial TUI | `go test ./internal/tui/ -run 'ThreadSpatial'` | pass |
| Full TUI package | `go test ./internal/tui/` | pass (3.2 s) |
| Race, whole repo | `go test -race ./...` | pass (all 30 packages) |
| Predecessor cache/refresh | `go test -run 'Cache\|Cached\|Refresh\|Degraded\|Prepared\|Replace\|Capacity\|Narrow' -v` | 22/22 pass |
| Benchmarks | `-bench 'VisibleViewportRenderNearCanvasLimit\|CachedRenderNearCanvasLimit' -benchmem` | pass, table above |
| Lint | `golangci-lint run ./...` | `0 issues.` |
| Formatting | `gofmt -l cmd internal` | no output |
| Module tidy | `go mod tidy -diff` | no diff |
| CLI docs drift | `go run ./internal/tools/docgen -out docs/cli && git diff --exit-code docs/cli` | no drift |
| Whitespace | `git diff --check` | clean |
| Planning lint | `./bin/tskflwctl lint` | `✔ all planning entities and dependency links pass lint` |
| Audit lint | `./bin/tskflwctl audit lint 6g8p95exeeje` | `✔ all audit findings pass lint` |

No check was skipped and no check was blocked.

### Recommendation

**Ship after triaging M1 and M2; M3 should land with it.**

The core of this change is sound and the performance claim is verified end-to-end against the real
pre-change implementation (27.2 MB → 220 KB per frame, 124×). The coordinate transform is a genuine
presentation-only transform: layout coordinates stay global and read-only, `cellAt`/`localPoint` are
a single well-placed clipping choke point, connector arms survive clipping, no diagonal or
zero-length segments exist to break the analytical extent replacement, every fallback path is
byte-identical to the reference tree, and the change actually *fixes* a pre-existing one-column row
shift at the left viewport edge.

M1 and M2 are behavioural changes the diff did not intend and no test observes — both need an owner
decision on the intended contract rather than a reflexive revert. M3 is the one I would not ship
without: a coordinate-translation bug in vertical route clipping is invisible to the entire suite
while corrupting 12% of sampled frames, and it costs one assertion to close. L1–L3 are coverage and
documentation debt.

The acceptance criteria are all ticked; on the evidence above, criterion 2 ("...endpoint arrows and
counts ... remain faithful when partially or wholly clipped") is contradicted by M2, criterion 3
("cannot ... silently omit an incident route that the existing visibility contract requires") is
unverified by the suite per L3, and criterion 5's claim of bounded regression evidence has the
`originY` hole described in M3.

All findings are left `open` for implementation-owner triage.
