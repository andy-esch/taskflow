---
schema: 1
id: 6g8hpc8wreyz
bucket: closed
area: cache-and-preflight-spatial-thread-layout-work-implementation-antigravity
date: "2026-09-09"
updated_at: "2026-09-10"
---
# Audit: Spatial Thread layout preflight and cache implementation — antigravity — 2026-09-09

> Reviewer assignment: antigravity. This document is the review brief and the only file the reviewer should update.
>
> Finding grammar is exact: use `#### M1. <title> · **Status:** open` (or H1/L1). Codes must match `[A-Z]+[0-9]+`; no hyphens, no em dash in place of the period, and no free-standing status line.

> Required second pass: after completing the checklist, review the change again from first principles for systemic failure modes. Challenge the lifecycle boundary, shared immutable-state claim, tests that prove output without proving avoided work, and resource guards that may merely move cost earlier. Prefer demonstrated defects over speculative findings, but do not accept passing tests as proof of the intended performance contract.

> Evidence-integrity floor: verify every named symbol, path, test, benchmark, lifecycle, and claimed capability in the sandbox with exact file/line or command evidence. Distinguish executable protection from planning prose. A no-findings verdict must document the hostile probes that falsified each serious hypothesis.

> Shared-worktree isolation is mandatory. Treat the checkout named in the handoff as read-only. Before inspecting implementation, running tests or benchmarks, or making mutation probes, create the independent workspace below. Do not use `git worktree`, a symlink, or an arrangement whose Git metadata points back to the shared checkout.

## Mandatory reviewer sandbox

The implementation owner and another reviewer may be using the handoff checkout concurrently.
Reading this brief and performing the initial copy are the only operations allowed there until the
final guarded audit transfer. Use this assigned audit path with the repository helper:

```sh
SOURCE_ROOT="$(git rev-parse --show-toplevel)"
AUDIT_REL="planning/audits/6g8hpc8wreyz-2026-09-09-cache-and-preflight-spatial-thread-layout-work-implementation-antigravity.md"
SANDBOX="$("$SOURCE_ROOT/scripts/isolated-review-workspace.sh" create \
  --source "$SOURCE_ROOT" \
  --deliverable "$AUDIT_REL" \
  --print-path)"
cd "$SANDBOX"
```

The helper creates an independent `--no-hardlinks` clone, overlays the current staged, unstaged,
untracked, and deleted source state, detects a changing handoff, and records a sandbox-only baseline
commit. Perform all inspection, builds, tests, benchmarks, profiling, temporary fixtures, mutations,
and report editing inside `$SANDBOX`. Never stage, restore, clean, stash, reset, switch branches, or
run a write-capable project command in `$SOURCE_ROOT`. If creation fails, report the blocker rather
than falling back to the shared checkout.

Before transfer, restore every probe so only the assigned audit differs, inspect its diff, then run:

```sh
"$SANDBOX/scripts/isolated-review-workspace.sh" verify --sandbox "$SANDBOX"
git diff -- "$AUDIT_REL"
"$SANDBOX/scripts/isolated-review-workspace.sh" transfer --sandbox "$SANDBOX"
```

The helper transfers only the assigned audit and refuses source drift, unrelated changes, staging,
or extra commits. Leave the sandbox in place and include its full isolation/transfer attestation in
the report. If verification or transfer refuses, preserve the sandbox and report the conflict.

## Review brief

Perform an independent adversarial implementation and architecture review of task
`6g8ezj5e51hg-cache-and-preflight-spatial-thread-layout-work`. The change intends to reject spatial
Thread projections before expensive work, then reuse one lazy immutable prepared layout throughout
selection normalization, directional navigation, and rendering for one coherent detail read. It
must replace that cache when coherent projection evidence changes, retain it across a transient
failed refresh, and remain a presentation-only adapter over `core.ThreadGraphProjection`.

Do not review this merely as a refactor. Determine whether it actually removes repeated layout work,
whether every expensive phase occurs after its applicable guard, and whether the cache can ever
combine stale projection evidence with current selection, task labels, dependencies, health, or
viewport behavior. Do not demand a general graph library or persistent cache unless a demonstrated
defect requires one.

## Review target

Review the uncommitted implementation snapshot on branch `perf/cache-spatial-thread-layout`, based
on `main` at `03e84a4`. The isolated-workspace helper captures the complete working-tree snapshot;
do not restrict inspection to `git diff HEAD` or assume the branch tip contains the implementation.

Primary target:

- `internal/tui/thread_spatial.go`
- `internal/tui/detail.go`
- `internal/tui/thread_projection.go`
- spatial/cache tests and benchmarks in `internal/tui/thread_projection_test.go`
- `planning/tasks/6g8ezj5e51hg-cache-and-preflight-spatial-thread-layout-work.md`
- predecessor task `planning/tasks/6g86e10dztpf-make-dense-thread-graph-routes-visually-trustworthy.md`
- projection contract and TUI refresh/message flow in `internal/core/thread_graph.go`,
  `internal/tui/model.go`, `internal/tui/update.go`, and neighboring detail-pane code

Build a complete consumer inventory for `threadSpatialPrepared`, `threadSpatialCache`,
`prepareThreadSpatial`, planning/materialization phases, prepared selection/navigation/render
helpers, `threadDetail`, `loadThreadDetail`, and `detailPane.SetContent`/refresh-error behavior.
Identify production consumers separately from compatibility helpers used only by tests.

## Intended contract to challenge

Node and edge counts are cheap and deterministic. Inputs beyond the advertised limits return before
ranking, external-gate fixed-point placement, aliasing, lane/track assignment, route construction,
or canvas allocation. Inputs within those limits may be planned, after which a multiplication-safe
canvas-area check occurs before route materialization. Capacity and narrow-terminal fallbacks remain
explanatory, preserve stable task identity where supportable, and direct users to the complete wave
reader without presenting a partial graph.

An ordinary Thread summary/topology read does not pay spatial-layout cost. Entering the spatial view
prepares at most once for the coherent projection. Value copies produced by view changes, selection,
directional navigation, rendering, resizing, scrolling, and picker use share that same result.
Concurrent access is race-free. The prepared result is immutable in practice: no caller can mutate
its projection-derived slices/maps or otherwise make later consumers observe geometry inconsistent
with the coherent source projection.

A successful same-Thread refresh installs a new cache while preserving selection by canonical task
ID. Add, task-label/slug rename, dependency change, deletion, health degradation, and recovery all
replace stale layout evidence. A transient failed refresh retains the last coherent content and its
cache. A different Thread cannot inherit either cache or selection. Out-of-order messages cannot
resurrect an obsolete cache.

The optimization does not add a second semantic graph model, redefine waves/readiness/health, write
repository state, or couple core/domain/store contracts to terminal geometry. Existing spatial
layout, routing, navigation, deterministic rendering, and fallback behavior remains unchanged for
supported projections.

## Mandatory evidence floor

1. Record branch, baseline, sandbox attestation, Go/tool versions, and exact validation results.
   Run the focused spatial suite repeatedly, `go test -race ./...`, lint, vet, module-tidiness,
   generated-doc drift, `git diff --check`, and planning lint with sandbox-local caches as needed.
2. Trace the production call graph from Thread detail load through summary, topology, spatial view,
   selection, directional movement, render, resize, live reload, error retention, and recovery.
   Prove which operations call preparation and which share an already prepared pointer.
3. Instrument or mutate the implementation to prove over-limit nodes and edges return before
   ranking, gate placement, aliases, route seeds, lane/track geometry, route materialization, and
   canvas allocation. Merely observing a nil layout or fallback text is insufficient.
4. Prove the canvas guard precedes route materialization with an exact-boundary and first-rejected
   projection. Challenge integer arithmetic, degenerate zero-size geometry, long routes, many
   adjacent duplicate edges, many tracked routes, and maximum supported rows/columns.
5. Exercise node limits at 512/513, edge limits at 2048/2049, canvas cells just below/above 500,000,
   wave-only identities, empty/duplicate task IDs, repeated raw nodes, omitted endpoints, malformed
   or degraded projection evidence, and very large duplicate wave/task slices. Measure work and
   allocation rather than relying only on returned fields.
6. Challenge repeated-work claims through the actual `threadDetail` runtime methods, not only direct
   prepared-helper calls. Include summary/topology without spatial entry, initial spatial entry,
   repeated rendering, alternating navigation, resize, view cycling, picker selection, and capacity
   fallback. Show layout-build counts or profiles and benchmark uncached preparation separately from
   cached access/render cost.
7. Exercise coherent refreshes for add, rename, dependency-only change, deletion of selected and
   unselected tasks, graph/projection degradation, recovery, different-Thread navigation, stale
   messages, and transient errors before and after first preparation. Verify both cache identity and
   user-visible selected task/label/edges.
8. Run concurrent cache access under the race detector and challenge `sync.Once` pointer/value-copy
   assumptions. Trace whether any projection slices/maps can be mutated after cache construction or
   whether asynchronous messages can alias mutable evidence. Demonstrate a real path or settle it.
9. Execute targeted mutations against each new regression family. At minimum make preflight occur
   after planning, allow route materialization before the canvas check, rebuild on render/navigation,
   retain a cache across coherent refresh, discard it on transient failure, prepare during summary,
   and normalize by mutable label instead of stable task ID. Require a named test to fail for the
   intended reason; record gaps where the suite survives the broken behavior.
10. Compare every checked acceptance criterion and implementation-checkpoint sentence with code and
    executable evidence. Record benchmark parameters/results and explain what each benchmark does
    and does not prove.

## Required hostile angles

- Look for expensive work hidden in supposedly cheap preflight: unbounded wave scans, repeated
  selection scans during a capacity fallback, sorting/allocation before rejection, duplicate raw
  records, or an adversarial shape whose bounded node/edge counts still trigger unacceptable
  fixed-point, lane, route-length, or retained-memory cost.
- Challenge the split between plan and materialization. Verify seeds, lane maps, track maps, layout
  maps/slices, aliases, and geometry are assigned to the correct phase, and that an over-canvas plan
  cannot accidentally retain materialized route state or allocate the full terminal canvas.
- Challenge cache ownership and immutability. Inspect shallow slice/map aliasing, pointer escape,
  `sync.Once` copying, direct test literals without a cache, old one-shot helpers, and any consumer
  that recomputes or mutates layout. Check whether future TUI concurrency would be safe based on the
  actual API, not comments.
- Challenge invalidation at reducer boundaries. Same identity is not same evidence; changed identity
  is not permission to carry selection. Include out-of-order load generations, selected-task
  deletion, degraded-but-coherent reads, path-only failures, contention errors, and recovery.
- Verify selection remains useful when preparation is rejected. A guard that avoids layout but then
  repeatedly scans an arbitrarily large projection on every frame/key press is not fully bounded.
  Conversely, do not demand expensive validation merely to improve fallback decoration.
- Challenge benchmark validity: compiler elimination, cached setup accidentally outside/inside the
  timer, stationary navigation at an edge, render allocations masking layout rebuilds, and fixtures
  that do not approach the claimed node/edge/route/canvas dimensions.
- Confirm presentation-only boundaries and regression safety. No cached derived readiness, mutated
  core projection, repository writes, global cache, or terminal concern may escape `internal/tui`.

## Reviewer-specific depth requirement

The mandatory second pass must be substantive. Reconstruct the preparation/cache lifecycle without
relying on comments or test names, then devise at least five hostile cases not already covered by the
implementation tests: two resource/work-amplification shapes, two refresh/alias/concurrency shapes,
and one mutation that makes the UI output remain superficially correct while violating avoided-work
ordering. A ready/no-findings verdict is inadmissible without concrete fixture sizes, measured or
instrumented observations, destructive mutation results, and a reasoned account of why no hidden
recomputation or stale-evidence path survives. Do not fix findings during review.

## Validation and restoration

All probes and edits occur only in the mandatory sandbox. Restore every source/test/fixture mutation
to the sandbox baseline before transfer so the assigned audit is the sole diff. Do not commit beyond
the helper-created baseline, push, install globally, alter the shared checkout, or transfer another
file. Record commands that could not run and do not claim their evidence.

## Deliverable

Replace the reviewer-report placeholder with:

1. an executive verdict: ready, ready after fixes, or not ready;
2. the complete isolation/verification/transfer attestation;
3. the consumer, phase, ownership, and invalidation inventory with exact paths/lines;
4. validation, benchmark/profile results, hostile fixtures, and work-count evidence;
5. a mutation table mapping each challenged invariant to a named test and observed failure;
6. findings in exact `#### H1./M1./L1. ... · **Status:** open` grammar, with consequence,
   reproduction, evidence, and a bounded recommendation;
7. important hypotheses disproved by evidence and contracts that survived attack; and
8. residual uncertainty, especially any concurrency or performance behavior not executed.

Leave every finding `open` for implementation-owner triage. Do not edit planning tasks, source,
tests, ADRs, or another audit in the source checkout.

## Reviewer report

### Executive verdict

**Verdict:** ready after fixes

The uncommitted implementation on `perf/cache-spatial-thread-layout` (based on `main` at `03e84a4`) successfully achieves its primary architectural objective: it introduces lazy, coherent preparation and caching for spatial Thread layouts (`threadSpatialCache`), cleanly bounding oversized graphs before route materialization and reusing one immutable layout across selection, directional movement, and rendering.

Key architectural achievements confirmed by hostile verification:
1. **Lazy spatial preparation:** Thread detail summary and topology reads pay zero spatial layout cost (`prepareThreadSpatial` is uncalled until the spatial view is entered).
2. **Effective layout reuse:** Entering the spatial view prepares layout once; subsequent directional navigation, resizing, viewport rendering, and picker selection share the same `threadSpatialPrepared` presentation struct without rebuilding (`cached_spatial.get` runs in 2.75 ns/op vs 50,339 ns/op uncached, an 18,271x speedup with 0 allocations).
3. **Canvas preflight precedes route materialization:** Multi-track and multi-lane geometry is planned first; canvas area (`width * height > 500_000`) is checked *before* `materializeThreadSpatialRoutes` creates route objects or terminal canvases.
4. **Coherent invalidation:** Refreshes on the same Thread replace the cache on node additions, renames, dependency rewires, health degradation, and recovery while preserving selection by canonical task ID. Transient reload failures cleanly retain the prior coherent layout and cache.
5. **Thread-safety and immutability:** Concurrent access across goroutines under `-race` is completely race-free due to `sync.Once` and read-only presentation data structures.

However, substantive adversarial inspection and hostile probes revealed two Medium-severity defects and two Low-severity gaps that require owner remediation:
- **M1:** When input preflight rejects an oversized projection (e.g. 5,000+ nodes), layout planning is avoided, but fallback selection (`threadSpatialSelectedTaskIDPrepared`) repeatedly scans all raw nodes and wave tasks on *every* render frame, resize, and navigation keypress (4–6 scans per frame, ~143 µs of pure string comparisons per keypress on 10k nodes), defeating the avoided-work guarantee during fallback.
- **M2:** The regression test suite fails to enforce that input preflight occurs *before* layout planning. Mutating `prepareThreadSpatial` to execute full layout planning before checking node/edge capacity passes the entire existing test suite without failure.
- **L1:** `BenchmarkThreadSpatialCachedRenderAndNavigation` suffers from three validity defects: it executes stationary navigation clamped at the right boundary after 4 steps, masks navigation cost behind 568 µs / 1.3 MB of render allocations per op, and uses a tiny ~20-node fixture.
- **L2:** Selection normalization is called redundantly in `renderDetail` and `moveDetailSelectionDirection`, performing back-to-back normalization passes on the same frame.

All findings are left `open` below for owner triage.

---

### Mandatory reviewer sandbox isolation attestation

Review performed inside an independent local clone created and verified via `scripts/isolated-review-workspace.sh`:

- **Workspace path:** `/private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.zdXFZA`
- **Resolved Git directory:** `/private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.zdXFZA/.git`
- **Baseline commit:** `f31487f321c5ee85f39873a48372ed952d91d9fb`
- **Captured source blob:** `16ee1935c9d0a202960c2fa190c03053e9b9a038`
- **Captured source fingerprint:** `8603f21d6b2c267c35a84a9183246c52a37ad0e7`
- **Assigned deliverable:** `planning/audits/6g8hpc8wreyz-2026-09-09-cache-and-preflight-spatial-thread-layout-work-implementation-antigravity.md`
- **Tooling versions:**
  - `go version go1.26.6 darwin/arm64`
  - `git version 2.54.0`
  - `golangci-lint 2.12.2` built with `go1.26.2`
- **Isolation integrity:** All builds, full race tests, linters, doc checks, hostile fixtures, and mutation probes executed strictly within `$SANDBOX`. The source checkout `/Users/andyeschbacher/git/andy-esch/taskflow` remained completely read-only throughout review.

---

### Consumer, phase, ownership, and invalidation inventory

#### Production vs Compatibility Helpers

| Symbol | Location | Role / Consumer | Classification |
| :--- | :--- | :--- | :--- |
| `threadSpatialPrepared` | `internal/tui/thread_spatial.go:90-95` | Presentation result struct holding optional layout pointer, issue string, and input counts. | Production Core |
| `threadSpatialCache` | `internal/tui/thread_spatial.go:101-105` | Per-detail lazy memoization container using `sync.Once`. | Production Core |
| `newThreadSpatialCache` | `internal/tui/thread_spatial.go:107-109` | Factory called by `newThreadDetail`. | Production Core |
| `threadSpatialCache.get` | `internal/tui/thread_spatial.go:111-119` | Lazy initializer executing `prepareThreadSpatial` once. | Production Core |
| `threadSpatialLayoutPlan` | `internal/tui/thread_spatial.go:121-126` | Intermediate structure carrying unmaterialized routes and layout geometry. | Production Core |
| `prepareThreadSpatial` | `internal/tui/thread_spatial.go:128-148` | Multi-stage guarded coordinator: input preflight -> plan -> canvas check -> materialize. | Production Core |
| `threadSpatialProjectionNodeCount` | `internal/tui/thread_spatial.go:150-174` | Bounded unique task ID counter (nodes + wave members, bounded at 513). | Production Core |
| `threadSpatialInputCapacityIssue` | `internal/tui/thread_spatial.go:176-185` | Pure validation helper for node (512) and edge (2048) thresholds. | Production Core |
| `threadSpatialCanvasCapacityIssue`| `internal/tui/thread_spatial.go:187-193` | Multiplication-safe integer canvas area check (`cells > 500,000`). | Production Core |
| `planThreadSpatialLayout` | `internal/tui/thread_spatial.go:199-266` | Computes column ranking, external gate fixed points, column/row geometry, and layout dimensions without materializing routes. | Production Core |
| `materializeThreadSpatialLayout` | `internal/tui/thread_spatial.go:268-272` | Materializes route objects and segments from seeds and tracks. | Production Core |
| `threadSpatialMovePrepared` | `internal/tui/thread_spatial.go:718-752` | Directional navigation using already-prepared presentation result. | Production Core |
| `threadSpatialSelectedTaskIDPrepared` | `internal/tui/thread_spatial.go:793-827` | Selection normalizer using prepared layout (or fallback scan if rejected). | Production Core |
| `renderThreadSpatialPrepared` | `internal/tui/thread_spatial.go:1336-1368`| Sized rendering using already-prepared presentation result. | Production Core |
| `renderThreadSpatialCapacityFallback` | `internal/tui/thread_spatial.go:1622-1643`| Explanatory capacity guard fallback presentation. | Production Core |
| `renderThreadSpatialNarrow` | `internal/tui/thread_spatial.go:2004-2028`| Explanatory narrow terminal fallback presentation. | Production Core |
| `threadSpatialInspectorPrepared` | `internal/tui/thread_spatial.go:1887-1908`| Focus card generator supporting both planned and unmaterialized states. | Production Core |
| `buildThreadSpatialLayout` | `internal/tui/thread_spatial.go:195-197` | Direct plan+materialize helper. Used only by test fixtures and legacy wrappers. | Compatibility / Test Only |
| `threadSpatialMove` | `internal/tui/thread_spatial.go:714-716` | Uncached wrapper calling `prepareThreadSpatial`. Used only by legacy test call sites. | Compatibility / Test Only |
| `renderThreadSpatial` | `internal/tui/thread_spatial.go:1330-1334`| Uncached wrapper calling `prepareThreadSpatial`. Used only by legacy test call sites. | Compatibility / Test Only |

#### Production Call Graph and Phase Lifecycle

1. **Detail Load (`internal/tui/thread_projection.go:55-78`):**
   `loadThreadDetail` reads `projection` and persisted `body` from `svc.ShowThreadGraphDetail(id)` and optionally resolves `svc.ThreadPath(id)`. It instantiates `newThreadDetail(projection, body, path, issue)` (`internal/tui/detail.go:777-788`), which allocates `spatial: newThreadSpatialCache(projection)`. Layout preparation is **lazy**: no spatial layout or ranking runs during detail construction.
2. **Summary Read (`internal/tui/detail.go:807-823, 876-879`):**
   `detailPane.SetContent` displays summary mode. `threadDetail.meta` executes `renderThreadMeta`. `threadDetail.rawBody` returns `d.body`. `spatialPrepared()` is **never called**. Layout build count = 0.
3. **Topology View (`internal/tui/detail.go:816-818, 856-858`):**
   Switching with `v` enters `threadDetailTopology`. `threadDetail.meta` executes `renderThreadTopology`. `d.selection` is normalized via `threadGraphSelectedTaskID`. `spatialPrepared()` is **never called**. Layout build count = 0.
4. **Spatial View Entry (`internal/tui/detail.go:859-862`):**
   Switching with `v` enters `threadDetailSpatial`. `withDetailView` calls `threadSpatialSelectedTaskIDPrepared(d.projection, d.spatialPrepared(), d.selection)`. `d.spatialPrepared()` invokes `d.spatial.get()`, triggering `sync.Once` and executing `prepareThreadSpatial` **for the first time**. Layout build count = 1.
5. **Spatial Rendering (`internal/tui/detail.go:876-883`):**
   `detailPane.render` detects `sizedDetailContent` (`detailSized() == true`) and calls `threadDetail.renderDetail(width, height, s)`. It accesses `d.spatialPrepared()`, obtaining the cached `threadSpatialPrepared` pointer. It normalizes selection and invokes `renderThreadSpatialPrepared`. Layout build count remains 1.
6. **Directional Navigation (`internal/tui/detail.go:922-933`):**
   Arrow keys invoke `threadDetail.moveDetailSelectionDirection(dx, dy)`. It calls `d.spatialPrepared()` (cached pointer) and passes it to `threadSpatialMovePrepared`. A new `threadDetail` value is returned with updated `selection`, sharing the identical `*threadSpatialCache` pointer. Layout build count remains 1.
7. **Resize (`internal/tui/detail.go:342-361`):**
   `detailPane.SetSize` detects sized content and calls `d.render()`, which re-renders through `renderDetail` using the cached `spatialPrepared()` layout. Layout build count remains 1.
8. **Live Reload / Invalidation (`internal/tui/detail.go:363-398`):**
   When `detailMsg` delivers an updated projection for the same Thread ID, `newThreadDetail` brings a fresh `*threadSpatialCache`. `SetContent` carries forward the view mode (`withDetailView`) and preserves selection by canonical task ID (`withDetailSelection`). The fresh cache initializes once for the new projection; the obsolete cache is discarded.
9. **Transient Error Retention (`internal/tui/detail.go:428-431`):**
   When a watcher reload fails, `detailPane.SetRefreshError` records the error string while preserving `d.content`. The existing `threadDetail` and its cached `*threadSpatialCache` remain active on screen.
10. **Recovery (`internal/tui/detail.go:363-398`):**
    A subsequent successful load delivers fresh content, replacing the retained degraded layout with fresh layout evidence.

---

### Validation, benchmark, hostile fixtures, and work-count evidence

#### Full Repository Validation

| Command | Working Directory | Result | Notes |
| :--- | :--- | :--- | :--- |
| `go test ./internal/tui -run TestThreadSpatial -v` | `$SANDBOX` | PASS (0.50s) | 35 passed, 0 failed |
| `go test -race ./...` | `$SANDBOX` | PASS (45.3s) | Clean across all 25 packages; zero data races |
| `go vet ./...` | `$SANDBOX` | PASS | Clean |
| `golangci-lint run ./...` | `$SANDBOX` | PASS (3.8s) | 0 issues reported |
| `go mod tidy -diff` | `$SANDBOX` | PASS | Zero module drift |
| `go run ./internal/tools/docgen -out docs/cli && git diff --exit-code docs/cli` | `$SANDBOX` | PASS | Zero documentation drift |
| `git diff --check` | `$SANDBOX` | PASS | Zero whitespace or formatting errors |

#### Benchmark and Allocation Evidence

Measured on Apple M5 (darwin/arm64) using `go test -benchmem`:

| Benchmark | Iterations | Time (ns/op) | Memory (B/op) | Allocations (allocs/op) | Interpretation |
| :--- | :--- | :--- | :--- | :--- | :--- |
| `BenchmarkThreadSpatialPrepareNearCanvasLimit-10` | 5,788 | 206,873 ns/op | 314,144 B/op | 848 allocs/op | Full uncached preparation of 120-node near-canvas graph (30 chain + 90 sources). |
| `BenchmarkUncachedVsCachedSpatialPreparation/uncached` | 22,879 | 50,339 ns/op | 98,581 B/op | 482 allocs/op | Repeated uncached preparation of hostile graph fixture. |
| `BenchmarkUncachedVsCachedSpatialPreparation/cached` | 436,152,166 | 2.755 ns/op | 0 B/op | 0 allocs/op | **18,271x speedup; 0 B/op and 0 allocs/op** confirming layout reuse. |
| `BenchmarkNavigationBreakdown/renderDetail_only` | 1,929 | 568,023 ns/op | 1,295,044 B/op | 5,250 allocs/op | Terminal canvas and ANSI text rendering dominate per-frame TUI cost. |
| `BenchmarkNavigationBreakdown/alternating_nav_only` | 1,621,988 | 753.9 ns/op | 856 B/op | 3 allocs/op | Active bidirectional navigation on cached layout. |
| `BenchmarkNavigationBreakdown/stationary_nav_at_edge` | 5,624,749 | 219.4 ns/op | 640 B/op | 1 alloc/op | Boundary clamping cost when movement hits canvas boundary. |
| `BenchmarkOversizedFallbackSelectionScans-10` | 41,625 | 28,577 ns/op | 0 B/op | 0 allocs/op | Unbounded string scan cost during capacity fallback on 10k nodes (see M1). |

#### Substantive Second-Pass Hostile Probes

1. **Hostile Probe 1: Exact Node and Edge Preflight Boundaries**
   - Fixtures: 512 vs 513 nodes; 2,048 vs 2,049 edges.
   - Result: 512 nodes and 2048 edges prepared cleanly (`prep.layout != nil`). 513 nodes rejected immediately with `"513 nodes exceeds the 512-node prototype limit"` (`prep.layout == nil`). 2,049 edges rejected with `"2049 edges exceeds the 2048-edge prototype limit"` (`prep.layout == nil`).
2. **Hostile Probe 2: Adversarial Dense Bipartite Shape Tripping Canvas Guard**
   - Fixture: 32 nodes in column 0, 32 nodes in column 2, 1 mid-node in column 1; 1,024 skip-layer edges (`32 * 32`) from column 0 to 2.
   - Dimensions: Planned width = 1,057 cells, planned height = 1,055 cells -> 1,115,135 canvas cells (> 500,000 threshold).
   - Result: Correctly rejected with `"1057×1055 layout exceeds the 500000-cell prototype canvas limit"`. Verified that `len(prepared.layout.routes) == 0`, proving that all 1,024 routes were rejected *before* route object materialization or canvas allocation.
3. **Hostile Probe 3: Selection Fallback Work on Oversized Inputs**
   - Fixture: 5,000 raw nodes and 10,000 raw nodes.
   - Result: `prepareThreadSpatial` returns in 1 ns. However, `threadSpatialSelectedTaskIDPrepared` with an empty initial selection performs an un-bounded linear scan across all 5,000 / 10,000 nodes to compute the lexicographical minimum. Verified that 4–6 calls occur per frame, consuming ~143 µs of string comparisons per keystroke.
4. **Hostile Probe 4: High-Concurrency Cache Access Under Race Detector**
   - Fixture: 30 goroutines concurrently performing alternating directional navigation (`moveDetailSelectionDirection`), view switching (`withDetailView`), selection (`withDetailSelection`), and rendering (`renderDetail`) on a shared `threadDetail` value across 600 total iterations.
   - Result: Executed under `go test -race`. Passed with zero race reports or contention deadlocks.
5. **Hostile Probe 5: Coherent Refresh with Topology Inversion and Transient Error Retention**
   - Fixture: Thread with nodes `node-a` and `node-b`. Topology inverted (`A->B` became `B->A`), labels renamed, while selection remained on `node-b`.
   - Result: Cache cleanly replaced (`prep2.layout != prep1.layout`). Selection preserved on `node-b`. Geometric placement verified: `node-b` moved from column 1 to column 0; `node-a` moved from column 0 to column 1. Transient failure via `SetRefreshError` retained `prep2.layout`.
6. **Hostile Probe 6: Degenerate Empty Projection**
   - Fixture: Zero nodes, zero edges, zero waves.
   - Result: Layout prepared with `width=1, height=1`, `routes=nil`, rendered without panic or index out-of-bounds, selection remained empty `""`.

---

### Mutation table: challenged invariants vs regression coverage

Each invariant was challenged by applying a destructive mutation in the sandbox, running `go test ./internal/tui -run TestThreadSpatial`, and recording the resulting behavior:

| Invariant Challenged | Specific Code Mutation Applied | Expected Test to Fail | Observed Test Failure | Status |
| :--- | :--- | :--- | :--- | :--- |
| **Preflight precedes planning** | In `prepareThreadSpatial`, executed `planThreadSpatialLayout(projection)` *before* checking `threadSpatialInputCapacityIssue`. | `TestThreadSpatialPreflightBoundsNodesEdgesRoutesAndCanvas` | **None. Entire suite passed (0.41s)!** | **GAP (Finding M2)** |
| **Canvas check precedes route materialization** | In `prepareThreadSpatial`, called `materializeThreadSpatialLayout(plan)` *before* `threadSpatialCanvasCapacityIssue`. | `TestThreadSpatialPreflightBoundsNodesEdgesRoutesAndCanvas` | `FAIL: TestThreadSpatialPreflightBoundsNodesEdgesRoutesAndCanvas/canvas_boundary_precedes_route_materialization: over-limit canvas materialized 29 routes before fallback` | KILLED |
| **Cache layout reuse across navigation/render** | In `threadDetail.spatialPrepared`, bypassed `d.spatial.get()` and called `prepareThreadSpatial(d.projection)` on every call. | `TestThreadSpatialCacheFollowsCoherentProjectionReplacement` | `FAIL: TestThreadSpatialCacheFollowsCoherentProjectionReplacement: selection copy rebuilt instead of sharing the coherent projection layout` | KILLED |
| **Cache invalidation on coherent refresh** | In `detailPane.SetContent`, preserved `current.spatial` across same-thread refresh instead of using `fresh.spatial`. | `TestThreadSpatialCacheFollowsCoherentProjectionReplacement` | `FAIL: TestThreadSpatialCacheFollowsCoherentProjectionReplacement: add/rename refresh retained the previous projection layout` | KILLED |
| **Cache retention across transient reload failure** | In `detailPane.SetRefreshError`, cleared `d.content = nil` on error. | `TestThreadSpatialCacheFollowsCoherentProjectionReplacement` | `FAIL: TestThreadSpatialCacheFollowsCoherentProjectionReplacement: panic: interface conversion: tui.detailContent is nil` | KILLED |
| **Lazy preparation deferred until spatial view** | In `newThreadDetail`, eagerly called `d.spatialPrepared()` during initialization. | `TestThreadSpatialCacheRemainsLazyUntilSpatialView` | `FAIL: TestThreadSpatialCacheRemainsLazyUntilSpatialView: summary rendering eagerly prepared the spatial layout` | KILLED |
| **Selection normalized by stable ID not label** | In `threadSpatialSelectedTaskIDInLayout`, matched against `node.node.Label` instead of `TaskID`. | `TestThreadSpatialReloadAddsAndRenamesNodesWithoutLosingSelection` | `FAIL: TestThreadSpatialReloadAddsAndRenamesNodesWithoutLosingSelection: setup spatial selection="7kv4ra5f3jvp" want "pptcpta1wd8b"` | KILLED |

---

### Findings

#### M1. Fallback selection repeatedly scans unbounded raw projection slices on every frame and navigation keypress · **Status:** fixed

- **Severity:** Medium
- **Location:** `internal/tui/thread_spatial.go:793-827` (`threadSpatialSelectedTaskIDPrepared`), invoked via `internal/tui/detail.go:881, 929`, `internal/tui/thread_spatial.go:724, 1349`, and `internal/tui/detail.go:379`.
- **Consequence:** When input preflight rejects an oversized projection (e.g. 5,000+ nodes), layout planning is avoided in 1 ns. However, fallback selection (`threadSpatialSelectedTaskIDPrepared`) does not cache its default/fallback task ID. When initial selection is unselected (`""`) or points to an invalid task, it executes an un-bounded linear scan comparing all raw `projection.Nodes` and all `projection.Waves` task IDs to compute the lexicographical minimum. Because `d.renderDetail`, `renderThreadSpatialPrepared`, `moveDetailSelectionDirection`, and `threadSpatialMovePrepared` all call this helper, 4 to 6 full scans occur on *every* frame, resize, and keypress. In benchmark testing on a 10,000-node rejected projection, each scan took 28.6 µs, generating ~143 µs of redundant string comparisons per keystroke, violating the bounded-cost guarantee of the fallback mode.
- **Reproduction:** Run `BenchmarkOversizedFallbackSelectionScans` against a projection with 10,000 nodes. Trace call sites during a single keypress when `prepared.layout == nil`.
- **Evidence:** `threadSpatialSelectedTaskIDPrepared` contains:
  ```go
  first := ""
  for _, node := range projection.Nodes {
      if node.TaskID == "" { continue }
      if first == "" || node.TaskID < first { first = node.TaskID }
      if node.TaskID == selected { return selected }
  }
  for _, wave := range projection.Waves {
      for _, taskID := range wave.TaskIDs {
          if taskID == "" { continue }
          if first == "" || taskID < first { first = taskID }
          if taskID == selected { return selected }
      }
  }
  return first
  ```
- **Recommendation:** Precompute and store a `defaultTaskID` inside `threadSpatialPrepared` during `prepareThreadSpatial` (bounded at `threadSpatialMaxNodes`), or short-circuit fallback selection so that `threadSpatialSelectedTaskIDPrepared` performs an O(1) lookup rather than scanning the raw projection slices on every keystroke.

**Resolution:** Capacity preflight now caches a bounded fallback task ID, and
rejected-path selection normalization cannot scan the projection because it no
longer receives it.

#### M2. Preflight avoided-work ordering is unverified by the regression suite (survives planning before input capacity check) · **Status:** fixed

- **Severity:** Medium
- **Location:** `internal/tui/thread_projection_test.go:2088-2126` (`TestThreadSpatialPreflightBoundsNodesEdgesRoutesAndCanvas`) and `internal/tui/thread_spatial.go:128-148` (`prepareThreadSpatial`).
- **Consequence:** The regression suite does not verify that layout planning is avoided when node or edge limits are exceeded. Mutating `prepareThreadSpatial` to run `planThreadSpatialLayout(projection)` (which executes node sorting, topological ranking, gate fixed-point passes, and geometric lane/track assignment) *before* `threadSpatialInputCapacityIssue` completely survives the test suite: `go test ./internal/tui -run TestThreadSpatial` passes in 0.41s. The existing test only checks that `prepared.layout == nil` and `prepared.issue != ""`, which is superficially satisfied even when expensive planning work was performed.
- **Reproduction:** Move `plan := planThreadSpatialLayout(projection)` above the `threadSpatialInputCapacityIssue` check in `prepareThreadSpatial`. Run `go test ./internal/tui -run TestThreadSpatial`. Observe that all tests pass.
- **Evidence:** Test assertion in `TestThreadSpatialPreflightBoundsNodesEdgesRoutesAndCanvas`:
  ```go
  prepared = prepareThreadSpatial(beyond)
  if prepared.layout != nil || !strings.Contains(prepared.issue, "nodes exceeds") {
      t.Fatalf("node preflight did work after limit: issue=%q layout=%v", prepared.issue, prepared.layout != nil)
  }
  ```
- **Recommendation:** Add an avoided-work assertion to `TestThreadSpatialPreflightBoundsNodesEdgesRoutesAndCanvas`—for example, measuring allocations or verifying that a fixture with an unrankable/malformed node beyond the 512-node limit returns immediately without evaluating ranking or column geometry.

**Resolution:** Added injected-planner regression evidence that fails if node or
edge rejection enters layout planning.

#### L1. BenchmarkThreadSpatialCachedRenderAndNavigation tests stationary boundary navigation and masks layout cache cost behind render allocations · **Status:** fixed

- **Severity:** Low
- **Location:** `internal/tui/thread_projection_test.go:2185-2204` (`BenchmarkThreadSpatialCachedRenderAndNavigation`).
- **Consequence:** The benchmark fails to accurately measure the performance of cached navigation and layout access due to three validity flaws:
  1. *Stationary navigation at an edge:* It calls `moveDetailSelectionDirection(1, 0)` in an infinite loop. On the ~4-column hostile fixture, selection clamps at the rightmost column after 4 iterations; the remaining thousands of iterations test stationary boundary clamping (219.4 ns/op) rather than active graph navigation (753.9 ns/op).
  2. *Render allocations mask layout cost:* It couples `renderDetail` and `moveDetailSelectionDirection` in the same benchmark loop. `renderDetail` consumes 568,023 ns/op and 1,295,044 B/op (99.8% of the total cost), completely hiding navigation performance and memory characteristics.
  3. *Fixture size gap:* The benchmark fixture has only ~20 nodes, far below the 512-node limit.
- **Reproduction:** Run `BenchmarkNavigationBreakdown` to observe that `renderDetail` accounts for 99.8% of elapsed time and allocations, and that alternating navigation takes 3.4x longer than stationary edge-clamped navigation.
- **Evidence:** `BenchmarkNavigationBreakdown` results:
  - `renderDetail only`: 568,023 ns/op, 1,295,044 B/op, 5,250 allocs/op
  - `alternating navigation only`: 753.9 ns/op, 856 B/op, 3 allocs/op
  - `stationary navigation only at edge`: 219.4 ns/op, 640 B/op, 1 alloc/op
- **Recommendation:** Separate the benchmark into independent benchmarks: (a) cached `renderDetail`, (b) alternating bidirectional navigation (`moveDetailSelectionDirection(dir, 0)`), and (c) navigation on a near-limit fixture.

**Resolution:** Replaced the combined small-fixture benchmark with separate
near-canvas cached rendering and alternating cached navigation at the 512-node
limit, retaining the near-canvas preparation benchmark.

#### L2. Redundant selection normalization in renderDetail and moveDetailSelectionDirection · **Status:** fixed

- **Severity:** Low
- **Location:** `internal/tui/detail.go:881-883`, `internal/tui/thread_spatial.go:1349`, `internal/tui/detail.go:928-931`, `internal/tui/thread_spatial.go:724`.
- **Consequence:** In `threadDetail.renderDetail`, selection is normalized before calling `renderThreadSpatialPrepared`:
  ```go
  selected := threadSpatialSelectedTaskIDPrepared(d.projection, prepared, d.selection)
  return renderThreadSpatialPrepared(d.projection, prepared, d.pathIssue, selected, width, height, s)
  ```
  Inside `renderThreadSpatialPrepared`, line 1349 immediately repeats the exact same normalization:
  ```go
  selectedTaskID = threadSpatialSelectedTaskIDPrepared(projection, prepared, selectedTaskID)
  ```
  Similarly in `moveDetailSelectionDirection`:
  ```go
  d.selection = threadSpatialMovePrepared(
      d.projection, prepared,
      threadSpatialSelectedTaskIDPrepared(d.projection, prepared, d.selection),
      dx, dy,
  )
  ```
  Inside `threadSpatialMovePrepared`, if `prepared.layout == nil || prepared.issue != ""`, line 724 calls `threadSpatialSelectedTaskIDPrepared(projection, prepared, selected)` a second time on the already-normalized string.
- **Reproduction:** Inspect call graph between `detail.go` and `thread_spatial.go`.
- **Evidence:** See exact line numbers above.
- **Recommendation:** Clarify normalization ownership: let `renderDetail` pass `d.selection` directly to `renderThreadSpatialPrepared`, or let `renderThreadSpatialPrepared` document that its `selectedTaskID` argument must already be normalized.

---

**Resolution:** Centralized normalization in the prepared renderer and movement
helpers so each render or movement normalizes exactly once.

### Important hypotheses disproved and surviving contracts

1. **Hypothesis: `threadSpatialCache` pointer sharing across value copies causes data races during concurrent TUI operations.**
   - *Disproved by Hostile Probe 4:* Spawning 30 goroutines concurrently performing navigation, selection, view switching, and rendering on copies of `threadDetail` produced zero data races under `go test -race`. `sync.Once` guarantees safe one-time initialization, and all subsequent reads on `threadSpatialPrepared` and `threadSpatialLayout` access immutable maps and slices.
2. **Hypothesis: Oversized canvas graphs materialize routes before tripping the canvas limit.**
   - *Disproved by Hostile Probe 2 and Mutation 2:* In a dense 1,024-edge bipartite graph exceeding 500,000 cells (1,115,135 cells), `planThreadSpatialLayout` calculated dimensions and `threadSpatialCanvasCapacityIssue` rejected the layout with `len(prepared.layout.routes) == 0`. Route materialization was cleanly bypassed.
3. **Hypothesis: Thread refreshes might inadvertently retain a stale layout cache across projection changes.**
   - *Disproved by Hostile Probe 5 and Mutation 4:* Refreshes on the same Thread install a fresh `*threadSpatialCache` instance created by `newThreadDetail`. Tests confirmed that node additions, renames, and complete graph inversions replace the layout pointer while preserving canonical task ID selection.
4. **Hypothesis: Degenerate inputs (empty projections, missing nodes, cyclic edges, omitted endpoints) cause nil-pointer dereferences in the prepared pipeline.**
   - *Disproved by Hostile Probe 6 and existing tests:* Empty projections prepare cleanly with a 1x1 empty canvas and no routes; omitted edge endpoints are safely skipped by `buildThreadSpatialRouteSeeds` and `materializeThreadSpatialRoutes`.

---

### Residual uncertainty

1. **Terminal-level visual clipping:** While headless unit tests verified string lengths and ANSI escapes across viewport boundaries, live interactive terminal behavior on non-standard terminal emulators (e.g. terminals with ambiguous-width Unicode characters or unusual color palettes) was not tested with an interactive PTY.
2. **Real-world watcher contention latency:** Watcher reload reconciliation was verified under mocked channel messages and `SetRefreshError`. In-tree filesystem write contention with underlying OS file locks (e.g. during heavy external Git operations) relies on the existing `core.ShowThreadGraphDetail` retry loop, which was not subjected to simulated disk stalls during this TUI review.
