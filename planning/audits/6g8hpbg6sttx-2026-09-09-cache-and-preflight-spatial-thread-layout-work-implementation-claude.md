---
schema: 1
id: 6g8hpbg6sttx
bucket: closed
area: cache-and-preflight-spatial-thread-layout-work-implementation-claude
date: "2026-09-09"
updated_at: "2026-09-10"
---
# Audit: Spatial Thread layout preflight and cache implementation — claude — 2026-09-09

> Reviewer assignment: claude. This document is the review brief and the only file the reviewer should update.
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
AUDIT_REL="planning/audits/6g8hpbg6sttx-2026-09-09-cache-and-preflight-spatial-thread-layout-work-implementation-claude.md"
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
3. Instrument or mutate the phase implementation to prove over-limit nodes and edges return before
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

After the checklist, choose the three most plausible ways this optimization could silently regress
in a future refresh or performance change. For each, create a sandbox-only hostile test or
instrumentation probe, make the corresponding destructive mutation, and record whether existing
tests detect it. At least one must concern preflight work ordering and one cache invalidation or
immutability. Do not fix findings during review.

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

### 1. Executive verdict

**Ready after fixes.** Every behavioural claim in the intended contract held under hostile probing.
Instrumented phase counters prove that an over-limit node or edge count returns before *any* of
ranking, external-gate placement, aliasing, route seeding, lane/track geometry, route
materialization, or canvas allocation; that an over-canvas plan is rejected before route
materialization and before a single terminal cell is allocated; and that entering the spatial view
prepares exactly once and that ten subsequent renders plus five directional moves add zero layout
work. Invalidation is correct for add, rename, dependency-only change, deletion of the selected and
unselected task, coherent degradation, recovery, a different Thread, a stale generation, and a
transient failure both before and after first preparation. The canvas guard's integer arithmetic is
exact and overflow-safe across an exhaustive sweep. Thirty-two concurrent value copies share one
layout pointer with no race-detector report. `internal/core`, `internal/domain`, `internal/store`
and `internal/wire` are byte-identical to `main`.

The fixes are about *protection and evidence*, not shipped behaviour. Three targeted mutations that
break the two headline acceptance criteria — moving the input preflight to *after* planning, making
rendering rebuild instead of reading the cache, and making navigation rebuild instead of reading the
cache — all survive the complete package suite, and the two assertions that name those exact
regressions cannot fail for that reason. A fourth mutation that puts an unbounded
`buildThreadSpatialLayout` call inside the over-limit fallback renderer — precisely the work the
guard exists to prevent — also survives. Separately, the rejected-projection path still scans the
whole projection twice per frame and twice per keypress, an over-canvas rejection retains up to
349 KB of route-planning work it discarded, and the two committed benchmarks use fixtures at 4% of
the node band and 1.2% of the edge band while measuring a selection that is pinned at the layout's
right edge for 10 of 12 iterations.

### 2. Isolation attestation

Creation (only `git rev-parse`, reading this brief, and the helper `create` ran in the source
checkout; no write-capable project command, staging, restore, clean, stash, reset, or branch switch
touched `$SOURCE_ROOT`):

```
sandbox_path=/private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.UDcujJ
git_dir=/private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.UDcujJ/.git
baseline_commit=f713e46d16fbaee3c5f886f101c84edf6e6474ad
source_blob=05766302467d0ea9085414bb339495b82f0247c8
source_fingerprint=8603f21d6b2c267c35a84a9183246c52a37ad0e7
deliverable=planning/audits/6g8hpbg6sttx-2026-09-09-cache-and-preflight-spatial-thread-layout-work-implementation-claude.md
deliverable_changed=false
transfer=pending
```

`assert_independent_clone` passed: in-tree `.git` directory, no `objects/info/alternates`, exactly
one worktree resolving to the sandbox, no `core.worktree`. Baseline `f713e46` sits directly on
`03e84a4` (the `main` merge named in the brief) and captures the complete uncommitted working-tree
snapshot, so inspection was never restricted to `git diff HEAD`.

Branch `perf/cache-spatial-thread-layout`; source `HEAD` `03e84a4`. Toolchain: `go1.26.6
darwin/arm64`, `golangci-lint 2.12.2`, `just 1.51.0`, Darwin 25.6.0, Apple M5. All Go build, test,
lint and coverage caches were directed outside the sandbox (`GOCACHE`/`GOLANGCI_LINT_CACHE` under
the session scratchpad) so no cache directory ever appeared as sandbox drift; `bin/tskflwctl` was
also built outside the sandbox. Every probe, instrumentation patch, mutation and fixture was applied
and then reverted with `git checkout -- internal/tui/`; `git status --porcelain` and `git diff
--stat` are both empty apart from this audit, and the committed suite is green on the restored tree.

### 3. Consumer, phase, ownership and invalidation inventory

**Construction and ownership.** `newThreadDetail` (`internal/tui/detail.go:777`) is the *only*
production constructor of `threadDetail`; it is called from exactly one place,
`loadThreadDetail` (`internal/tui/thread_projection.go:74`). It mints one
`*threadSpatialCache` (`internal/tui/thread_spatial.go:107`) per coherent projection read.
`threadDetail.spatial` (`detail.go:769`) is a pointer, so every value copy the reducer produces
retains it. The only other `threadDetail{}` literals in the tree are five test sites
(`thread_projection_test.go:519,537,992,2497`, `atlas_test.go:816,849`); `spatialPrepared`
(`detail.go:790-797`) covers them by falling back to an uncached `prepareThreadSpatial`.

**Production consumers of the prepared result** — all six read through `spatialPrepared()`, never
`prepareThreadSpatial` directly:

| Site | Consumes |
| --- | --- |
| `detail.go:861` `withDetailView("spatial")` | `threadSpatialSelectedTaskIDPrepared` |
| `detail.go:880-882` `renderDetail` | `threadSpatialSelectedTaskIDPrepared` + `renderThreadSpatialPrepared` |
| `detail.go:887` `detailSelectionKey` | `threadSpatialSelectedTaskIDPrepared` |
| `detail.go:894` `withDetailSelection` | `threadSpatialSelectedTaskIDPrepared` |
| `detail.go:926-931` `moveDetailSelectionDirection` | `threadSpatialSelectedTaskIDPrepared` + `threadSpatialMovePrepared` |
| `nav.go:196` follow/picker | `detailSelectionKey()` on the pane's own content |

**Compatibility helpers with no production caller** (verified by grep over non-test files):
`renderThreadSpatial` (`thread_spatial.go:1330`), `threadSpatialMove` (`:714`) and
`buildThreadSpatialLayout` (`:195`). Each re-enters the prepared path with a fresh, uncached
`prepareThreadSpatial`, so they are honest wrappers rather than a second code path — but they are
the reason `internal/tui` still contains an unguarded layout builder (see M1).

**Phase split**, proved by counters rather than by returned fields (§4):

| Phase | Function | Runs when |
| --- | --- | --- |
| count | `threadSpatialProjectionNodeCount` `:150` | always |
| input guard | `threadSpatialInputCapacityIssue` `:176`, called `:132` | always |
| plan | `planThreadSpatialLayout` `:199` — rank, gates, labels, aliases, seeds, column/row geometry, `byID`/`nodes` | only within input limits |
| canvas guard | `threadSpatialCanvasCapacityIssue` `:187`, called `:141` | only after plan |
| materialize | `materializeThreadSpatialLayout` `:268` → `materializeThreadSpatialRoutes` | only within canvas limit |
| canvas allocation | `newThreadSpatialCanvas` `:936`, via `renderThreadSpatialCanvasWindow` `:1658` | only in render, after both guards |

**Invalidation.** A new `detailMsg` is dropped unless it is the current selection *and* the current
generation (`model.go:346`), so an out-of-order load can never resurrect an obsolete cache;
`loadDetail` (`model.go:1260-1262`) stamps the generation. `detailPane.SetContent`
(`detail.go:363-397`) installs the fresh content — hence the fresh cache — and, only for the same
item, transplants the on-screen view and the *canonical task ID* onto it via `withDetailView` +
`withDetailSelection`. `restoreDetailNavigation` (`nav.go:293-313`) likewise goes through those two
methods, so it never rebuilds a literal and never orphans the cache. `SetRefreshError`
(`detail.go:428-431`) touches only `errMsg`/`loading`, so a transient failure retains both the last
coherent content and its cache.

### 4. Validation, instrumentation, benchmarks and hostile fixtures

**Baseline validation, all in the sandbox, all green:** `go build ./...`; `go vet ./...`; the focused
spatial suite `go test ./internal/tui/ -run 'ThreadSpatial|ThreadDetail|ThreadRegistry|ThreadTopology'
-count=3`; `go test -race ./...` (24 packages, `internal/tui` 9.457s); `golangci-lint run ./...` →
`0 issues`; `go mod tidy` produced no `go.mod`/`go.sum` delta; `just docs-check` → no `docs/cli`
drift; `git diff --check 03e84a4 f713e46` → clean; `bin/tskflwctl lint` → `all planning entities and
dependency links pass lint`.

**Phase instrumentation.** I inserted counters at the head of `planThreadSpatialLayout`,
`materializeThreadSpatialLayout`, `rankThreadSpatialColumns`, `placeThreadSpatialExternalGates`,
`labelThreadSpatialColumns`, `threadGraphAliases`, `buildThreadSpatialRouteSeeds`,
`threadSpatialColumnGeometry`, `threadSpatialRowGeometry`, `materializeThreadSpatialRoutes` and
`newThreadSpatialCanvas`, then drove them from the boundary fixtures:

```
nodes-512     issue=""                          layout=y routes=0    cells=104516 | aliases=1 colGeom=1 gates=1 label=1 materializeLayout=1 materializeRoutes=1 plan=1 rank=1 routeSeeds=1 rowGeom=1
nodes-513     "513 nodes exceeds the 512-node…" layout=n             cells=0      | (no phase work)
edges-2048    issue=""                          layout=y routes=2048 cells=512    | aliases=1 colGeom=1 gates=1 label=1 materializeLayout=1 materializeRoutes=1 plan=1 rank=1 routeSeeds=1 rowGeom=1
edges-2049    "2049 edges exceeds the 2048-…"   layout=n             cells=0      | (no phase work)
canvas-in-91  issue=""                          layout=y routes=29   cells=495392 | aliases=1 colGeom=1 gates=1 label=1 materializeLayout=1 materializeRoutes=1 plan=1 rank=1 routeSeeds=1 rowGeom=1
canvas-out-92 "904×554 layout exceeds the 500000-cell…" layout=y routes=0 cells=500816 | aliases=1 colGeom=1 gates=1 label=1 plan=1 rank=1 routeSeeds=1 rowGeom=1
```

The `(no phase work)` rows are the executable proof the brief asked for: over-limit inputs reach
neither ranking nor gate placement nor aliasing nor seeds nor geometry nor materialization nor
canvas allocation. The `canvas-out-92` row proves the canvas guard precedes route materialization
(`materializeLayout=0`, `materializeRoutes=0`, `routes=0`) *and* canvas allocation
(`canvasAlloc=0`) while retaining placements for stable identity.

**Repeated-work counts through the real runtime methods**, not the prepared helpers:

```
summary + topology only            → aliases=1            (no plan, no materialize: laziness holds)
entering the spatial view          → plan=1 materializeLayout=1 materializeRoutes=1 rank=1 …
10 renders after entry             → canvasAlloc=10       (zero layout work)
5 × (render + directional move)    → canvasAlloc=5        (zero layout work)
```

**Canvas-guard arithmetic.** Exhaustive sweep of `w ∈ [1,4000]` against `⌊M/w⌋−1, ⌊M/w⌋, ⌊M/w⌋+1,
⌊M/w⌋+2` compared with the true 64-bit product: **0 mismatches**. `w=h=2³¹` is rejected correctly
(the naive product would wrap to 4611686018427387904), confirming the division form is
multiplication-safe. `width==0` accepts any height, but `planThreadSpatialLayout:246` writes
`max(layoutWidth, 1)`, so that state is unreachable.

**Boundary and adversarial fixtures exercised:** nodes 512/513; edges 2048/2049; canvas 495,392 /
500,816 cells; a wave-only identity beyond the node band; empty task IDs in `Nodes` and in
`Waves`; repeated raw node records (600 records / 300 identities); 2,000,000 duplicate wave entries;
a 512-node/2048-edge external-gate chain; a 32-node long-route chain; and a bipartite 512-node sweep
at 16/64/128/256/512/1024/2048 edges. That sweep found the practically reachable envelope is much
narrower than the advertised bands — 512 nodes is accepted at 16 edges (276,954 cells) and rejected
by the canvas guard from 64 edges upward — which is why the O(nodes × edges) direct-neighbour search
in `threadSpatialMovePrepared:758-762` never becomes expensive in practice (measured 136 ns/op).

**Benchmarks** (`-benchtime=300x -count=3`, Apple M5, arm64; medians):

| Benchmark | ns/op | B/op | allocs/op | Fixture |
| --- | --- | --- | --- | --- |
| `BenchmarkThreadSpatialPrepareNearCanvasLimit` (committed) | 209,641 | 314,148 | 848 | 120 nodes, 29 edges, 495,392 cells |
| `BenchmarkThreadSpatialCachedRenderAndNavigation` (committed) | 424,479 | 1,268,880 | 4,659 | 21 nodes, 25 edges, 23,572 cells |
| `BenchmarkProbeNearCanvasCachedRender` (mine) | 1,932,171 | 27,235,746 | 20,724 | one cached frame on the *prepare* benchmark's own fixture |
| `BenchmarkProbeAtLimitsNonStationaryMove` (mine) | 136 | 640 | 1 | 512-node rejected projection |

What they prove: preparing a near-canvas-limit layout from scratch costs 210 µs / 314 KB / 848
allocs, and the cache removes exactly that, once per coherent projection. What they do *not* prove:
that this dominates a frame. On the *same* fixture a single cached render still costs 1.93 ms /
27.2 MB / 20,724 allocs, because `renderThreadSpatialCanvasWindow:1658` allocates
`newThreadSpatialCanvas(layout.width, layout.height)` — the whole bounded canvas, not the visible
window — on every frame. Preparation is ~11% of one frame's time and ~1% of its allocation at
supported dimensions. Neither committed benchmark surfaces this (M3).

**Rejected-projection frame cost** (input guard fires, `layout == nil`):

```
nodes=1000    frame=214.667µs  keypress=8.167µs    allocs/frame=186
nodes=50000   frame=425.333µs  keypress=322.375µs  allocs/frame=186
nodes=400000  frame=5.392916ms keypress=2.539292ms allocs/frame=186
```

**Preflight cost against duplicate raw records:** 2,000,000 raw `Nodes` short-circuits in 125 ns;
2,000,000 duplicate wave entries take 11.96 ms (100,000 → 726 µs; 1,000 → 8.8 µs).

**Concurrency:** 32 goroutines each taking an independent value copy of the same `threadDetail`,
calling `spatialPrepared()` and `renderDetail`, repeated 5× under `-race`: no report, and all 32
observed one identical `*threadSpatialLayout`. `sync.Once` correctly publishes `c.prepared`.

**Memory retention of a rejected plan** (`runtime.MemStats`, 100 live results, forced GC): as
shipped 610,866 B each; with the layout struct detached from the enclosing plan 261,841 B each —
**349,024 B per over-canvas rejection** held only because `prepared.layout = &plan.layout`
(`thread_spatial.go:140`) makes the compiler keep the whole plan object alive. `go build -gcflags='-m
-m'` confirms: `internal/tui/thread_spatial.go:144:2: moved to heap: plan`, `flow: prepared ← &plan`.

**Coverage of the new code** (committed suite only, probes removed). Uncovered:
`thread_spatial.go:112-114` (nil-receiver guard in `get`), `:724-726` (the new rejected-input guard
in `threadSpatialMovePrepared`), `:803-804` (empty task ID in the fallback node loop), `:813-826`
(the entire wave half of the fallback selection scan), `:1359-1363` (the `layout == nil` fallback in
`renderThreadSpatialPrepared`), `:1897-1899` (the empty-selection inspector box).

### 5. Mutation table

Each mutation was applied to the restored baseline, built, run against the full `internal/tui`
package suite, then reverted.

| # | Challenged invariant | Mutation | Result | Named test |
| --- | --- | --- | --- | --- |
| M1 | Preflight precedes planning (AC1) | move the input guard to after `planThreadSpatialLayout` | **SURVIVED** | — |
| M2 | Canvas guard precedes materialization | materialize routes before the canvas check | KILLED | `TestThreadSpatialPreflightBoundsNodesEdgesRoutesAndCanvas/canvas_boundary_precedes_route_materialization` |
| M3 | Rendering reuses the cache (AC2) | `renderDetail` calls `prepareThreadSpatial` instead of `spatialPrepared` | **SURVIVED** | — |
| M4 | Navigation reuses the cache (AC2) | `moveDetailSelectionDirection` calls `prepareThreadSpatial` | **SURVIVED** | — |
| M5 | A refresh installs a new cache (AC3) | `newThreadDetail` reuses one process-wide cache | KILLED | `TestThreadSpatialCacheFollowsCoherentProjectionReplacement`, `TestThreadSpatialCacheRemainsLazyUntilSpatialView` |
| M6 | A transient failure retains the cache | `SetRefreshError` replaces the cache | KILLED | `TestThreadSpatialCacheFollowsCoherentProjectionReplacement` |
| M7 | Preparation is lazy | `newThreadDetail` prepares eagerly | KILLED | `TestThreadSpatialCacheRemainsLazyUntilSpatialView` |
| M8 | Selection is keyed by stable task ID | normalize to `node.Label` instead | KILLED | `TestThreadSpatialReloadAddsAndRenamesNodesWithoutLosingSelection` + 4 others |
| M9 | Navigation honours a rejected projection | drop `prepared.issue != ""` from the move guard | **SURVIVED** | — |
| M10 | Rejected selection keeps wave-only identities | delete the wave loop in `threadSpatialSelectedTaskIDPrepared` | **SURVIVED** | — |
| M11 | Node preflight dedupes identities | count raw records only | **SURVIVED** | — |
| M12 | Canvas guard is exact at the boundary | `+1` off-by-one in the guard | KILLED | `TestThreadSpatialPreflightBoundsNodesEdgesRoutesAndCanvas/canvas_boundary_precedes_route_materialization` |
| M14 | One preparation per cache | remove `sync.Once`, rebuild on every `get()` | KILLED | `TestThreadSpatialCacheFollowsCoherentProjectionReplacement` |
| M15 | Entering spatial normalizes selection | `withDetailView` skips normalization | KILLED | `TestThreadSpatialCacheRemainsLazyUntilSpatialView` |
| M16 | The input guard exists at all | delete `threadSpatialInputCapacityIssue`'s early return | KILLED | `TestThreadSpatialGraphFailsOpenToWavesBeyondPrototypeCapacity` + 3 subtests |
| M17 | The fallback stays bounded | fallback inspector calls `buildThreadSpatialLayout(projection)` | **SURVIVED** | — |

M1 was additionally instrumented: with the guard relocated, a 513-node projection still returns
`issue="513 nodes exceeds the 512-node prototype limit"` and `layout=nil`, but a call counter shows
`planThreadSpatialLayout calls=1`. The suite passes (`ok github.com/andy-esch/taskflow/internal/tui
3.069s`). The guard's *ordering* — the entire point of AC1 — is invisible to the tests.

### 6. Findings

#### H1. The two acceptance criteria this task exists to establish have no executable protection · **Status:** fixed

**Severity:** high (no shipped defect; the change is correct today, but its central contract can be
reverted in a future refactor without a single test turning red).

**Consequence.** AC1 ("preflight limits run before ranking, gate fixed-point placement, lane
assignment, and route materialization") and AC2 ("one immutable spatial layout is reused across
selection normalization, directional navigation, and rendering … instead of being rebuilt multiple
times per frame or keypress") are both ticked, and neither is pinned. M1, M3 and M4 each break one
of them and survive the whole package suite.

**Reproduction and evidence.**

- *Ordering (M1).* Relocate `thread_spatial.go:132-134` to after `plan :=
  planThreadSpatialLayout(projection)` at `:139`. Output is byte-identical, so nothing fails.
  Instrumenting `planThreadSpatialLayout` with a counter shows the mutated build runs the planner
  once for a 513-node projection while still reporting `layout=nil` and the right issue string.
  `TestThreadSpatialPreflightBoundsNodesEdgesRoutesAndCanvas/nodes_at_and_beyond_limit`
  (`thread_projection_test.go:2079-2083`) asserts only `prepared.layout != nil` and
  `strings.Contains(prepared.issue, "nodes exceeds")` — exactly the "observing a nil layout or
  fallback text" that the brief calls insufficient. The same is true of the edges subtest
  (`:2110-2119`) and the wave-only subtest (`:2091-2099`).
- *Reuse (M3/M4).* Change `detail.go:880` and `detail.go:926` from `d.spatialPrepared()` to
  `prepareThreadSpatial(d.projection)`. The suite passes.
  `TestThreadSpatialCacheFollowsCoherentProjectionReplacement:650-655` reads
  ```go
  _ = movedDetail.renderDetail(100, 28, &testStyles)
  if movedDetail.spatialPrepared().layout != original.layout {
      t.Fatal("render rebuilt instead of sharing the coherent projection layout")
  }
  ```
  `renderDetail` has a value receiver and returns a string; it cannot mutate `movedDetail.spatial`,
  and `spatialPrepared()` re-reads that `sync.Once`-guarded cache. The assertion is therefore
  incapable of failing for the reason it names, whatever `renderDetail` does internally. The
  identical structure appears at `:650` for navigation and at
  `BenchmarkThreadSpatialCachedRenderAndNavigation:2216-2219`.

**Recommendation (bounded).** Add one counter seam rather than a framework: a package-level
`threadSpatialLayoutBuilds` incremented inside `planThreadSpatialLayout`, asserted to be `0` after
an over-limit `prepareThreadSpatial`, and asserted to stay at `1` across a render + directional move
+ resize + view cycle on one `threadDetail`. That single variable kills M1, M3, M4 and M17 at once.

**Resolution:** Added explicit test seams: an injected planner proves node/edge
rejection occurs before planning, and an injected sentinel prepared result makes
render or navigation cache bypasses observably fail. The runtime cache is still
the only production preparation owner.

#### M1. The over-limit path is bounded only by convention: it rescans the whole projection every frame and keypress, and an unbounded fallback passes the suite · **Status:** fixed

**Severity:** medium.

**Consequence.** The input guard bounds *layout* work but not per-event work, and nothing prevents a
future edit from putting the guarded work back into the fallback renderer.

**Reproduction and evidence.**

- When `prepared.layout == nil`, `threadSpatialSelectedTaskIDPrepared`
  (`thread_spatial.go:801-826`) walks every `projection.Nodes` entry and every `wave.TaskIDs` entry.
  It runs **twice per frame** — once at `detail.go:881` and again at `thread_spatial.go:1349`, on the
  already-normalized value — and **twice per keypress**, at `detail.go:929` and `:927`. Measured:
  1,000 nodes → 215 µs/frame, 8 µs/keypress; 50,000 → 425 µs / 322 µs; 400,000 → **5.39 ms/frame and
  2.54 ms/keypress**. The second call per frame is pure redundancy on every path, cheap on the
  healthy path (a `byID` lookup) and a full rescan here.
- M9 (drop `prepared.issue != ""` from `threadSpatialMovePrepared:724`) and M10 (delete the wave loop
  at `:813-826`) both survive: coverage shows `:724-726` and `:813-826` are never executed by the
  committed suite. `TestThreadSpatialPreflightBoundsNodesEdgesRoutesAndCanvas/wave-only_nodes…`
  checks only `nodeCount`/`issue`, never the wave-only *selection* fallback the brief asks about.
- M17 is the sharpest: replacing the deliberately bounded branch at `thread_spatial.go:1897-1907`
  with `threadSpatialInspector(projection, buildThreadSpatialLayout(projection), …)` — i.e. building
  the full layout for a projection the guard just rejected — passes the entire suite.
  `buildThreadSpatialLayout` (`:195`) survives as an unguarded entry point with no production caller.

**Recommendation (bounded).** Compute the fallback identity once during `prepareThreadSpatial` and
store it on `threadSpatialPrepared` (it already pays one O(N) pass in
`threadSpatialProjectionNodeCount`), so rejected-path selection becomes a field read; drop the
redundant re-normalization at `thread_spatial.go:1349` since `renderDetail` already normalized; and
assert in a test that a rejected projection's render performs a constant number of projection scans.

**Resolution:** Capacity preflight now stores one bounded fallback task ID.
Rejected-path normalization no longer receives or scans the projection, and
render/movement each normalize only once.

#### M2. A canvas rejection retains the route-planning work it discarded — up to 349 KB, contradicting the type's own comment · **Status:** fixed

**Severity:** medium.

**Consequence.** The documented promise that a capacity issue lets callers "fail open to the wave
reader without paying the work the guard exists to bound" is not met for the canvas band: the seeds,
boundary lanes and track map computed before the rejection stay reachable for the whole lifetime of
the loaded Thread detail.

**Reproduction and evidence.** `prepareThreadSpatial:140` does `prepared.layout = &plan.layout`
*before* the canvas check at `:141`. Taking the address of a field makes Go heap-allocate the whole
enclosing `threadSpatialLayoutPlan`, and a pointer into it keeps every sibling field alive.
`go build -gcflags='-m -m' ./internal/tui/` reports `thread_spatial.go:144:2: moved to heap: plan`
with `flow: prepared ← &plan: from &plan.layout (address-of) at thread_spatial.go:145:20`.
Measured with `runtime.MemStats` over 100 live results after a forced GC, on a 512-node/2048-edge
projection rejected at 4125×3395 cells: 610,866 B retained per result as shipped versus 261,841 B
when the layout struct is copied out of the plan — **349,024 B of pure overhead per rejection**
(2.3× the layout it actually needs). On the accepted path `:150` rebinds `prepared.layout` to a
fresh `layout` variable, so the plan is collectible there; only the rejection leaks.

**Recommendation (bounded).** Copy the layout out before returning on the rejection branch — `l :=
plan.layout; prepared.layout = &l` — or set `prepared.layout` only after the guard, once from each
branch. Either is a two-line change and preserves the stable-identity behaviour the fallback relies
on.

**Resolution:** Over-canvas rejection now points at a standalone copy of the
placement layout, allowing rejected route seeds and lane/track maps in the
enclosing plan to be collected.

#### M3. Neither committed benchmark approaches the guarded dimensions, and one measures a selection pinned at the layout's edge · **Status:** fixed

**Severity:** medium.

**Consequence.** AC4 claims benchmark or regression coverage of "near-limit nodes, edges, route
length, canvas cells, and repeated render/navigation work". The *regression tests* do cover
nodes (512/513), edges (2048/2049), canvas (495,392/500,816) and route length (≥800), so AC4 is
satisfied in substance — but the benchmarks named in the implementation checkpoint as the
"repeatable evidence" for this work measure neither the guarded bands nor a representative frame,
so the residual per-frame cost is invisible and a navigation regression would barely register.

**Reproduction and evidence.**

- `BenchmarkThreadSpatialCachedRenderAndNavigation:2205` uses `hostileThreadGraphProjection()`:
  **21 nodes (4.1% of the 512 band), 25 edges (1.2% of the 2048 band), 23,572 cells (4.7% of the
  500,000 band)**.
- Replaying its exact loop, the selection trace for 12 successive `dx=+1` moves is
  `[…, vkfmzv3rggx9, vkfmzv3rggx9, vkfmzv3rggx9, vkfmzv3rggx9]` — **permanently stationary from
  iteration 10 of 12**. At the rightmost column the filter at `thread_spatial.go:759` rejects every
  candidate before `threadSpatialDirectNeighbor` is ever called, so the loop skips the only
  superlinear step in navigation.
- `BenchmarkThreadSpatialPrepareNearCanvasLimit` uses 120 nodes / 29 edges: near-limit in *cells*
  only. Running one cached render on that same fixture gives 1,932,171 ns / 27,235,746 B / 20,724
  allocs versus the benchmark's own 209,641 ns / 314,148 B / 848 allocs for a full uncached
  preparation — the cache removes ~11% of a frame's time and ~1% of its allocation. This is the
  residual the predecessor audit's L4 explicitly deferred here; it is legitimately out of this
  task's scope, but nothing in the suite now records it.

**Recommendation (bounded).** Point the cached-render benchmark at a fixture that reaches at least
one guarded band (`threadSpatialNearCanvasProjection(91)` already exists and is accepted), and
alternate `dx=+1`/`dx=-1` so the navigation stays in the non-stationary path. Optionally add one
`b.Run` for a rejected projection so the fallback frame cost in M1 is also tracked.

**Resolution:** Split the benchmark into near-canvas uncached preparation,
near-canvas cached rendering, and alternating active navigation at the 512-node
limit. The residual full-canvas render allocation is tracked by task
6g8k47wg8xks.

#### L1. `threadSpatialPrepared` is documented and relied on as immutable, but is a shallow handle over shared maps and slices · **Status:** fixed

**Severity:** low (no production caller mutates it today; verified by grep — the only writes to
`layout.byID`/`layout.nodes` in non-test code are `thread_spatial.go:265-266` inside the planner).

**Consequence.** `layout := *prepared.layout`, which appears at `thread_spatial.go:727`, `:799`,
`:1364` and `:1895`, reads like a defensive copy and is not: `byID`, `nodes`, `columns`, `columnX`,
`columnLabels` and `routes` are all shared with the cache. The comment at `:86-89` ("one immutable
presentation result") and AC2's "one immutable spatial layout" are conventions, not properties, so a
future consumer that annotates or reorders a layout copy would silently corrupt every other
consumer — and, since the maps are shared, would race under the concurrency `:98-99` anticipates.

**Reproduction and evidence.** Take two independent value copies from `spatialPrepared()`; through
the second, write `shallow.byID[victim] = poisoned` and `shallow.nodes[0].alias = "POISONED"`. A
third `spatialPrepared()` observes `alias "G1" → "POISONED"`, `x 3 → -9999`, and the next
`renderDetail` output contains `POISONED`. The `sync.Once` publication itself is sound: 32
concurrent goroutines under `-race`, repeated 5×, share one pointer with no report — the hazard is
mutation, not publication.

**Recommendation (bounded).** Give `threadSpatialPrepared` accessor methods
(`placement(id) (threadSpatialNode, bool)`, `routes() iter.Seq[...]`) and unexport or drop the
`layout` field so consumers cannot obtain the shared maps at all; or, cheaper, add a doc line at
`:86` stating that the layout is shared and must be treated as read-only, plus a test that a render
does not alter `reflect.DeepEqual` of the prepared layout.

**Resolution:** Documented that the shared layout is read-only by contract and
added a regression comparing the prepared layout with a fresh deterministic
layout after render and navigation.

#### L2. The node preflight counts raw records, so duplicate evidence is rejected and misreported, and the wave half has no raw-record short-circuit · **Status:** fixed

**Severity:** low — **neither shape is reachable from `core.ProjectThreadGraph` today** (verified:
`thread_projection.go:136-141` dedupes members, `:172-183` builds external gates from a
`map[string]bool` that excludes members, and `thread_graph.go:101-107` derives waves from a
topological partition in which each ID appears exactly once). This is a defence-in-depth and
accuracy finding, not a live bug.

**Consequence.** Two asymmetries between the comment at `thread_spatial.go:151-152` and the code.
First, `:153-154` returns `len(projection.Nodes)` when the raw slice is over the band, so 600 node
records carrying 300 distinct tasks are refused with "600 nodes exceeds the 512-node prototype
limit" and the fallback at `:1638` prints that same inflated 600 — a graph that plans to 300 nodes
and would render fine. The pre-change guard used `len(layout.nodes)` (deduped), so this is a
behaviour change for that shape. Second, the comment claims raw-record cost is rejected "without
allocating a proportional set", but only `Nodes` is short-circuited: the wave loop at `:161-173`
exits early only when a *new unique* ID pushes the count over the band, so duplicates are scanned in
full.

**Reproduction and evidence.** `threadSpatialProjectionNodeCount` on 2,000,000 raw `Nodes` returns
in **125 ns**; on one node plus 2,000,000 duplicate wave entries it takes **11.96 ms**
(100,000 → 726 µs, linear). A 600-record/300-identity projection reports `preflightCount=600` while
`planThreadSpatialLayout` places 300. M11 (count raw records only) survives the suite.

**Recommendation (bounded).** Mirror the `Nodes` short-circuit for waves — return early once the
number of *scanned* wave entries exceeds `threadSpatialMaxNodes`, since a partition can never need
more — and either count deduped identities in the over-band case too or reword the fallback line at
`:1638` to say "records" rather than "nodes" so the number it prints matches what it means.

**Resolution:** Raw node records remain intentionally bounded as input work
rather than silently deduplicated, now with accurate record wording. Wave
objects and task records gain equivalent early bounds so malformed duplicate
evidence cannot force an unbounded scan.

#### L3. The change adds three branches that cannot be reached and are not covered · **Status:** wontfix

**Severity:** low.

**Consequence.** Dead paths in a hot, security-of-bounds-relevant function invite a future reader to
assume `prepared.layout == nil` is a real state with `issue == ""`, which it is not.

**Reproduction and evidence.** `prepareThreadSpatial` always sets exactly one of `issue` or
`layout` (verified against zero, empty-`Nodes`, nil-`Waves` and wave-only projections), so
`renderThreadSpatialPrepared:1359-1363` — the `"spatial layout is unavailable"` fallback — is
unreachable from any production path; only a zero-value `threadSpatialPrepared` produces that state,
and no production code constructs one. Likewise `(*threadSpatialCache).get:112-114` guards a nil
receiver that `spatialPrepared:791` already excludes, and `threadSpatialInspectorPrepared:1897-1899`
handles an empty selection that no rejected-path caller produces. Coverage from the committed suite
confirms all three are never executed.

**Recommendation (bounded).** Delete the nil-receiver guard and the `layout == nil` fallback, or keep
one of them and add the single test that produces the state; leave the empty-selection box only if a
projection with no readable identity at all is considered reachable.

**Resolution:** Retained the defensive zero/nil states. They are constant-cost
internal totality guards, do not weaken the production invariant, and remain
useful for malformed projections and direct renderer tests.

### 7. Hypotheses disproved, and contracts that survived attack

- **"The guard is decorative — expensive work still happens before it."** Disproved. Phase counters
  show `(no phase work)` for 513 nodes and 2049 edges: no ranking, gate fixed-point, aliasing, route
  seeds, lane/track geometry, materialization or canvas allocation. Behaviour is right; only its
  protection is missing (H1).
- **"Route state or a full canvas leaks past an over-canvas rejection."** Disproved for routes and
  cells (`materializeLayout=0`, `materializeRoutes=0`, `canvasAlloc=0`, `routes=0` at 500,816 cells);
  the *plan's* seed and lane data does leak, which is M2.
- **"The canvas guard's integer arithmetic is wrong or overflows."** Disproved. Exhaustive sweep over
  `w ∈ [1,4000]` at four heights each: 0 mismatches against the 64-bit product. `2³¹ × 2³¹` rejects
  correctly where a naive product would wrap. `width==0` is unreachable (`:246` clamps to ≥1).
  M12 (a `+1` off-by-one) is killed.
- **"`sync.Once` is copied, or concurrent access races."** Disproved. `threadSpatialCache` is only
  ever held as a pointer (`detail.go:769`, `:783`); `go vet` reports no lock copies; 32 concurrent
  value copies of the enclosing `threadDetail` calling `spatialPrepared()` + `renderDetail`, 5×
  under `-race`, produce no report and one shared layout pointer.
- **"A stale or cross-Thread message can resurrect an obsolete cache."** Disproved. `model.go:346`
  drops any `detailMsg` whose `gen != m.detailGen` or whose selection has moved, and a different
  Thread gets a fresh cache and a fresh selection (`thread-b` → `b1`, cache pointer differs).
- **"A rename or dependency-only change leaves stale geometry."** Disproved. A dependency-only
  refresh installs a new layout (routes 0 → 1) and keeps the selected canonical ID; a rename
  refreshes `byID["a"].node.Label` while `detailSelectionKey()` still returns `"b"`. M8 confirms
  identity is the task ID, not the label.
- **"A transient failure discards the cache, or blocks later preparation."** Disproved.
  `SetRefreshError:428-431` touches only `errMsg`/`loading`; a failure before first preparation
  retains the content and lazy preparation still succeeds afterwards; a failure after preparation
  keeps the same layout pointer.
- **"Selection becomes unusable when preparation is rejected."** Disproved. Stable identity survives:
  the 513-node fallback renders `focus` with `task-0000`, and the over-canvas fallback still shows
  `focus [alias]` plus needs/unlocks from the retained placements. The residual concern is cost, not
  correctness (M1).
- **"An adversarial shape inside the bands triggers unacceptable fixed-point or navigation cost."**
  Disproved. A 512-node/2048-edge external-gate chain prepares in 1.19 ms and is then rejected by the
  canvas guard; the envelope sweep shows 512 nodes are accepted only up to ~16 edges, so the
  O(nodes × edges) direct-neighbour search at `:758-762` measures 136 ns/op. The canvas guard is the
  binding constraint and it fires first.
- **"The optimization escapes the presentation adapter."** Disproved. `git diff 03e84a4 f713e46`
  touches only `internal/tui/{detail,thread_projection,thread_spatial}.go`, their tests, and
  planning files. `internal/core/thread_graph.go` is byte-identical to `main`. No global cache, no
  repository write, no readiness or wave semantics cached, no terminal geometry in `core`/`domain`/
  `store`. `just docs-check` and `bin/tskflwctl lint` are clean.
- **Ordering helpers are non-destructive.** `orderedThreadGraphNodes:1289` and
  `orderedThreadGraphEdges:1295` copy before sorting, so no consumer reorders the core projection's
  slices — the one aliasing route by which a presentation cache could have corrupted shared evidence.

### 8. Residual uncertainty

- **Real terminal behaviour was not exercised.** All render measurements go through
  `renderDetail`/`renderThreadSpatialPrepared` in-process. I did not drive a live `tea.Program`, a
  real resize, or a watcher-driven reload against a real repository, so the frame costs in §4 are
  CPU costs, not observed input latency.
- **Concurrency is only *anticipated*.** The comment at `:98-99` reasons about "a future shell" that
  renders and navigates from different goroutines. Today Bubble Tea drives `Update`/`View` on one
  goroutine, and `loadThreadDetail` constructs the cache in a `tea.Cmd` goroutine without preparing
  it, so the only cross-goroutine hand-off is the `detailMsg` itself. My 32-goroutine probe shows
  the `sync.Once` publication is sound, but it does not establish that the *shared mutable maps*
  behind L1 would be safe under a genuinely concurrent shell.
- **The 400,000-node figure in M1 is a scaling probe, not a realistic Thread.** At plausible
  over-limit sizes (1,000–5,000 member tasks) the rejected-path rescan is well under a millisecond.
  The finding rests on the cost being unbounded and unprotected, not on today's magnitudes.
- **Duplicate node/wave evidence (L2) has no producer today.** I traced `ProjectThread` and
  `ProjectThreadGraph` and could not construct the shape through the service. If a future projection
  source (an import path, a repaired-document mode) can emit repeated identities, L2's severity
  rises.
- **Not run:** `just vulncheck`, `just release-snapshot`, and any cross-platform (linux/amd64)
  execution — none is implicated by this change, and I make no claim about their results.
- **Mutation coverage is targeted, not exhaustive.** Sixteen mutations were executed; five survived.
  A surviving mutation I did not think to write remains possible, so "protected" in §5 means
  "protected against these mutations".
