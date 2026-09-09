---
schema: 1
id: 6g8e0pyn885r
bucket: closed
area: dense-thread-graph-routes-implementation-antigravity
date: "2026-09-09"
updated_at: "2026-09-09"
---
# Audit: Dense Thread graph route hardening implementation — antigravity — 2026-09-09

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

Perform an independent adversarial implementation and architecture review of the dense Thread
graph route-hardening change. Treat every rendered connector, crossing, shared segment, endpoint,
gate placement, viewport omission, and capacity claim as untrusted until hostile evidence proves it
faithfully represents the supplied projection. Do not merely confirm that the existing tests pass
or repeat the implementation checkpoint.

This is a correctness review of a deliberately bounded TUI renderer, not an invitation to redesign
Thread semantics or demand a general graph-layout engine. Distinguish a demonstrated defect from a
subjective aesthetic preference. Still report a visual treatment if it can cause a reasonable user
to infer a dependency, endpoint, direction, wave, or graph state that the projection does not
contain.

## Review target

Review the uncommitted implementation snapshot on branch
`feat/dense-thread-route-trustworthiness`, based on `main` at `93b4965`. The isolated-workspace
helper captures the working-tree implementation; do not restrict inspection to `git diff HEAD` or
assume the branch tip contains it.

Primary target:

- `internal/tui/thread_spatial.go`
- spatial-route and neighboring projection tests in `internal/tui/thread_projection_test.go`
- task `planning/tasks/6g86e10dztpf-make-dense-thread-graph-routes-visually-trustworthy.md`
- the completed predecessor lifecycle update in
  `planning/tasks/6g8by30btznq-close-spatial-thread-prototype-correctness-and-invariant-gaps.md`
- the source projection contract in `internal/core/thread_graph.go`
- the spatial design findings and task mapping in
  `planning/audits/6g8bbjgcx9tj-2026-09-09-spatial-thread-graph-tui-experience-design-review.md`
- the two prior prototype implementation audits whose bounded findings seeded this task

Build a consumer inventory for `threadSpatialLayout`, route seeds/routes/segments, canvas route
cells, window filtering and boundary annotation, external-gate placement, capacity fallback, and
the spatial renderer entry point. Trace each consumer and all assumptions it makes about canonical
edge order, endpoint presence, node geometry, viewport dimensions, ANSI cutting, and
`ThreadGraphProjection`. Verify that no route-derived state has leaked into core/domain/wire/store
contracts and that no neighboring TUI presentation silently depends on the replaced fixed geometry.

## Intended contract to challenge

Every supplied projection edge whose endpoints are present receives one deterministic, node-free
route from prerequisite to dependent. Adjacent, skipped-layer, same-layer, reverse, cyclic-residue,
and self-edge shapes must terminate at their actual dependent. Long and reverse routes may use
dedicated tracks and waypoints, but an intermediate node or arrowhead must never appear to own them.

Independent perpendicular routes render as crossings rather than junctions. Geometry shared by
routes with a real common endpoint, parallel routes, fan-in, and fan-out remains distinguishable
without manufacturing connectivity. Fan counts describe the routes actually presented at the
visible endpoint. Drawing order cannot erase selected-route emphasis or semantic node colors;
yellow markers remain direction/focus signals rather than status substitutes.

Panning renders all routes incident to a visible node and all routes incident to the selected node,
while omitting a route whose two endpoint nodes are both outside the viewport. A selected incident
route that crosses a viewport boundary identifies its offscreen endpoint alias. Clipping, Unicode
width, overlapping labels, multiple clipped fan-in/out routes, and narrow terminals must not leave
an unattributed line or overwrite a different node/route with a false claim.

External gates remain presentation-only nodes inside a valid open interval between included
prerequisites and dependents. Cascaded gates settle deterministically regardless of supplied slice
order; the bounded fixed-point pass must converge for chains, forks, joins, partial/cyclic evidence,
and maximum supported inputs without reversing an edge. Presentation columns are called layers;
only core-provided member generations may be called waves.

The renderer remains bounded before canvas allocation and fails open to the complete wave reader.
Equivalent projection evidence yields identical layout, aliases, routes, annotations, and rendered
text. No route logic may infer readiness, mutate planning data, redefine graph health, or weaken the
adapter-neutral `ThreadGraphProjection` boundary.

## Mandatory evidence floor

1. Build the consumer inventory above with exact file/line evidence. Include indirect callers and
   tests; do not infer consumers from names.
2. Run the focused spatial suite repeatedly, the full race suite, lint, vet, generated-doc drift,
   module-tidiness, diff checks, and planning lint. Use sandbox-local Go and lint caches where host
   cache writes are unavailable.
3. Exercise at least one real rendered production Thread at approximately 80x24 and 140x36, or
   construct equivalent viewport-level evidence if a PTY is unavailable. Inspect stripped text and
   cell metadata; a screenshot alone is not proof of connectivity.
4. Independently construct hostile fixtures for: a dense layered DAG; skipped same-row edges;
   fan-in and fan-out; duplicate/parallel supplied edges; unrelated perpendicular crossings; three
   or more routes colliding at one cell; same-column cycles; a self-edge; an explicit reverse route;
   external-gate chains presented in the least favorable canonical order; empty/partial projection;
   narrow and capacity-boundary canvases; and selected endpoints clipped on every boundary.
5. For each newly added regression family, execute a targeted mutation that removes or corrupts the
   exact invariant it claims to pin and require the named test to fail for the intended reason.
   At minimum challenge unique lane assignment, node-free track placement, crossing-vs-junction
   classification, selected emphasis, endpoint arrows, window route filtering, boundary aliases,
   gate fixed-point repetition, honest layer labels, and pre-allocation capacity fallback. Use a
   coordinated mutation when a nearby helper would otherwise preserve the behavior accidentally.
6. Measure or tightly bound layout plus render work near the supported node/edge/canvas limits.
   Look beyond canvas cell count for route-length multiplication, fixed-point iteration, memory
   retained per cell, and work performed before the capacity check.
7. Verify equivalent-evidence determinism by permuting nodes, edges, waves, and equal duplicate
   edges repeatedly. Do not accept same-input-twice as sufficient proof.
8. Compare every implementation claim in the task checkpoint with code and executable evidence.
   Treat planned responsive layout and one-hop focus behavior as unimplemented, not as protection
   supplied by this change.

## Required hostile angles

- Try to create false junctions from crossing order, a third colliding route, shared fan-in/out
  stubs, corridor/waypoint overlap, arrow overlap, and count-marker overwrite. Determine whether the
  cell's limited route ownership metadata is sufficient for every supported collision.
- Try to make a real edge disappear or acquire the wrong endpoint through duplicate edges, omitted
  endpoints, self-edges, cyclic residue, reverse geometry, zero-length segments, external-gate
  compaction, and route adornment drawing order.
- Attack the viewport rule with partially visible node boxes, endpoints just outside each boundary,
  vertical-only visible segments, several selected incident routes sharing a boundary row, a route
  that leaves and re-enters the viewport, and offscreen-to-offscreen routes crossing visible space.
- Attack deterministic lane and track allocation with reordered evidence, tied rows, equal edges,
  map iteration, changing selection, and live reload. Verify selection changes presentation only,
  not geometry or aliases.
- Attack gate convergence with reverse-ordered chains, forks, joins, cycles, multiple outgoing
  intervals, incoming gates that move late, and empty compacted columns. Prove the `len(nodes)+1`
  bound or exhibit a valid supported case that needs more passes or settles invalidly.
- Attack resource guards with many adjacent parallel edges, many long routes, many same-column
  routes, maximum rows/columns, and layouts whose area is just below and just above the limit.
  Include work done while merely building an oversized layout.
- Inspect terminal semantics: double-width/control-bearing labels, ANSI clipping, route glyph width,
  truncated legends/headings, color-disabled output, and whether color-independent glyphs retain
  direction, crossing, shared-route, focus, and state meaning.
- Look for systemic Go design risks rather than only visual bugs: accidental quadratic behavior,
  mutable slice aliasing during gate compaction, inconsistent old/new route helpers, hidden map
  nondeterminism, stale geometry constants, oversized single-file coupling, and tests that inspect
  implementation details while missing the user-visible invariant.
- Confirm the change remains a presentation adapter. Reject any hidden second graph model that
  reinterprets core waves, readiness, roles, health, or dependency direction.

For Antigravity specifically: the mandatory second pass is substantive. Rebuild the route/canvas
state machine from first principles and attempt at least three hostile multi-route collision or
viewport cases not already named by the tests. A no-findings verdict must include those concrete
fixtures, mutation results, and why the limited per-cell route metadata cannot lie in them.

## Validation and restoration

All inspection, tests, temporary fixtures, profiling, mutation probes, and edits must occur inside
the mandatory isolated sandbox. Restore every probe to the sandbox baseline before writing the
final report. Do not format or regenerate unrelated files. Do not commit beyond the helper-created
baseline, push, modify the source checkout, or transfer anything except the assigned audit through
the helper. If a command cannot run, record the exact failure and avoid claiming that evidence.

## Deliverable

Replace the reviewer-report placeholder in the assigned audit with:

1. an executive verdict: ship, ship after fixes, or do not ship;
2. the required isolation attestation;
3. the consumer and boundary inventory with exact paths/lines;
4. commands and hostile fixtures actually run, including repeat counts and resource observations;
5. a mutation table mapping each challenged invariant to the named test and observed failure;
6. findings in the exact `#### H1./M1./L1. ... · **Status:** open` grammar, with severity,
   consequence, reproduction, evidence, and a bounded recommendation;
7. explicit corroboration of important contracts that survived attack; and
8. residual uncertainty, including anything visual or terminal-dependent that was not executable.

Leave all findings `open` for implementation-owner triage. Do not edit tasks, ADRs, source, tests,
or another audit in the source checkout.

## Reviewer report

### Executive verdict

**Verdict:** ship after fixes

The dense Thread graph route-hardening implementation achieves substantial architectural and visual progress over the initial prototype:
1. It replaces unanchored direct-line segments with deterministic, collision-avoiding inter-row routing tracks and assigned boundary lanes.
2. It enforces presentation invariance under equivalent-evidence slice permutations across cyclic, multi-wave, and dense topologies.
3. It keeps column layout layers honestly labeled and distinct from core member waves.
4. It bounds canvas allocation before rendering with independent node, edge, and cell capacity guards, and fails open cleanly to the complete wave reader.

However, substantive adversarial inspection and coordinated mutation probes revealed two High-severity visual correctness bugs, one Medium-severity layout defect, and two test-integrity gaps that must be remediated:
- **H1:** The fan-out route count marker unconditionally overwrites reverse and self-edge directional arrowheads (`◀`), completely erasing incoming and cyclic dependency indicators.
- **H2:** Viewport boundary route alias labels blindly overwrite adjacent node box borders and interior text, while multiple offscreen routes entering along the same boundary row overwrite each other into illegible string fragments (e.g. `[[[[[[[[[[[[[[[[M4]…`).
- **M1:** Independent perpendicular routes that share a destination endpoint (such as a long route's corridor crossing an adjacent route's vertical approach lane) are misclassified as shared geometry, rendering false junction diamonds (`◆`) that manufacture connectivity.
- **M2:** Dead legacy route drawing helpers (`drawThreadSpatialEdge`, `drawThreadSpatialLongEdge`) remain in the codebase and are the only callers tested by `TestThreadSpatialConnectorGeometryUsesElbowsAndAccentFocus`, leaving the real production arrowhead adornment pipeline completely unpinned (commenting out `putRouteArrow` in `drawThreadSpatialRouteAdornments` causes zero test failures).
- **L1:** `TestThreadSpatialLayoutPlacesCascadedExternalGatesAtOneStableFixedPoint` passes on pass 0 with 0 moves because its 3-member fixture already places all gates in their final positions during initial longest-path ranking, failing to exercise multi-pass fixed-point propagation.
- **L2:** The full spatial layout is computed twice on every render frame (once for selection key resolution and once for rendering) before any capacity guards are evaluated.

All findings are left `open` below for implementation-owner triage.

---

### Mandatory reviewer sandbox isolation attestation

Review performed inside an independent local clone created and validated via `scripts/isolated-review-workspace.sh`:

- **Workspace path:** `/private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.a9cZO1`
- **Resolved Git directory:** `/private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.a9cZO1/.git`
- **Baseline commit:** `03863f1b9cd59d2fdd3c7c9e546d1a7db2ab16bb`
- **Captured source blob:** `a55ff048876f5e8c7b14e708418817f924d112e0`
- **Captured source fingerprint:** `5af9084d9aed8205e5ef519c5d2beeba40f807b2`
- **Assigned deliverable:** `planning/audits/6g8e0pyn885r-2026-09-09-dense-thread-graph-routes-implementation-antigravity.md`
- **Isolation integrity:** All builds, full race tests, linters, doc checks, hostile fixtures, and mutation probes executed strictly within `$SANDBOX`. The source checkout `/Users/andyeschbacher/git/andy-esch/taskflow` remained read-only throughout review.

---

### Consumer and boundary inventory

Exact file and line tracing for all spatial route and layout components:

- **`threadSpatialLayout` struct:** `internal/tui/thread_spatial.go:72-81`
  - Constructed by `buildThreadSpatialLayout(projection core.ThreadGraphProjection)` at `internal/tui/thread_spatial.go:83-149`.
  - Direct callers:
    - `threadSpatialMove`: `internal/tui/thread_spatial.go:588-641` (invoked by TUI navigation handler in `internal/tui/detail.go:901`).
    - `threadSpatialSelectedTaskID`: `internal/tui/thread_spatial.go:655-658` (invoked by `internal/tui/detail.go:862, 869`).
    - `renderThreadSpatial`: `internal/tui/thread_spatial.go:1036-1097` (entry point invoked by `internal/tui/detail.go:857`).
    - `renderThreadSpatialCanvas`: `internal/tui/thread_spatial.go:1216-1221` (invoked in regression tests `internal/tui/thread_projection_test.go:1092, 1417, 1456, 1514`).
    - `renderThreadSpatialCanvasWindow`: `internal/tui/thread_spatial.go:1223-1246` (invoked by `renderThreadSpatial` at line 1063 and `renderThreadSpatialCanvas` at line 1217).
    - `threadSpatialCapacityIssue`: `internal/tui/thread_spatial.go:1181-1193` (invoked by `renderThreadSpatial` at line 1051 and tests `internal/tui/thread_projection_test.go:1193, 1613, 1634`).
    - `threadSpatialRoutesForWindow`: `internal/tui/thread_spatial.go:1248-1266` (invoked by `renderThreadSpatialCanvasWindow` at line 1234 and tests `internal/tui/thread_projection_test.go:1294`).
    - `annotateThreadSpatialRouteBoundaries`: `internal/tui/thread_spatial.go:1099-1131` (invoked by `renderThreadSpatial` at line 1067).
    - `threadSpatialInspector`: `internal/tui/thread_spatial.go:1472-1515` (invoked by `renderThreadSpatial` at line 1095, `renderThreadSpatialCapacityFallback` at line 1212, and `renderThreadSpatialNarrow` at line 1581).
    - `threadSpatialConnections`: `internal/tui/thread_spatial.go:1537-1549` (invoked by `threadSpatialInspector` at line 1499 and tests `internal/tui/thread_projection_test.go:1495, 1496`).

- **Route seeds, routes, segments, and waypoints:**
  - `threadSpatialRouteSeed`: `internal/tui/thread_spatial.go:63-70`
  - `threadSpatialRoute`: `internal/tui/thread_spatial.go:54-61`
  - `threadSpatialRouteSegment`: `internal/tui/thread_spatial.go:48-52`
  - Generated by:
    - `buildThreadSpatialRouteSeeds`: `internal/tui/thread_spatial.go:151-176`
    - `materializeThreadSpatialRoutes`: `internal/tui/thread_spatial.go:262-316`
    - `newThreadSpatialRoute`: `internal/tui/thread_spatial.go:318-343`
  - Consumed by:
    - `threadSpatialColumnGeometry`: `internal/tui/thread_spatial.go:189-227` (allocates boundary lanes `gaps[boundary] = max(8, len(uses)+2)`).
    - `threadSpatialRowGeometry`: `internal/tui/thread_spatial.go:229-260` (allocates node-free horizontal tracks `trackY`).
    - `drawThreadSpatialRouteSegments`: `internal/tui/thread_spatial.go:1268-1281`.
    - `drawThreadSpatialRouteAdornments`: `internal/tui/thread_spatial.go:1283-1289`.
    - `drawThreadSpatialRouteCounts`: `internal/tui/thread_spatial.go:1291-1333`.

- **Canvas route cells and drawing state machine:**
  - `threadSpatialCell`: `internal/tui/thread_spatial.go:730-743`
  - `threadSpatialCanvas`: `internal/tui/thread_spatial.go:762-773`
  - State mutation methods:
    - `putText`: `internal/tui/thread_spatial.go:775-803`
    - `putRouteConnector`: `internal/tui/thread_spatial.go:805-848`
    - `putRouteArrow`: `internal/tui/thread_spatial.go:856-873`
    - `putRouteWaypoint`: `internal/tui/thread_spatial.go:875-892`
    - `putRouteCount`: `internal/tui/thread_spatial.go:894-905`
    - `routeHorizontal`: `internal/tui/thread_spatial.go:907-921`
    - `routeVertical`: `internal/tui/thread_spatial.go:923-936`
    - `renderLine`: `internal/tui/thread_spatial.go:980-1034`

- **Window filtering and boundary annotation:**
  - `threadSpatialRoutesForWindow`: `internal/tui/thread_spatial.go:1248-1266`
  - `annotateThreadSpatialRouteBoundaries`: `internal/tui/thread_spatial.go:1099-1131`
  - `threadSpatialVisibleRouteExtent`: `internal/tui/thread_spatial.go:1133-1158`
  - `threadSpatialRouteAlias`: `internal/tui/thread_spatial.go:1160-1165`
  - `putThreadSpatialBoundaryLabel`: `internal/tui/thread_spatial.go:1167-1179`

- **External gate placement:**
  - `placeThreadSpatialExternalGates`: `internal/tui/thread_spatial.go:351-424`
  - `threadSpatialOrderedIDs`: `internal/tui/thread_spatial.go:426-438`
  - `removeThreadSpatialID`: `internal/tui/thread_spatial.go:440-448`

- **Boundary enforcement and adapter purity:**
  - Package isolation: `internal/core`, `internal/domain`, `internal/wire`, and `internal/store` contain zero imports of `internal/tui`.
  - The adapter consumes `core.ThreadGraphProjection` strictly as an immutable input; no graph health, wave definition, readiness, or planning state is mutated or inferred by the routing logic.

---

### Verification and hostile test execution

#### 1. Test suite and toolchain verification
- `go test -race ./...`: 0 race conditions, all unit and integration tests passed across all packages.
- `go vet ./...`: 0 issues.
- `golangci-lint run ./...`: 0 issues reported.
- `git diff --check`: 0 whitespace or formatting anomalies.
- `go mod tidy -diff`: 0 module drift.
- `go run ./internal/tools/docgen -out docs/cli && git diff --exit-code docs/cli`: 0 CLI documentation drift.
- `go run ./cmd/tskflwctl lint`: 0 planning entity or dependency errors.
- `go test -v -count=5 -run TestThreadSpatial ./internal/tui`: 5 repeated clean runs (all 16 tests passing deterministically).

#### 2. Real production Thread execution
- Loaded Thread `6g503c6pfqeb` ("Complete production Threads") via `core.NewService(store.NewFS("planning"))` with the active repository task DAG.
- **Topology:** 42 nodes, 56 edges, 14 waves.
- **Layout metrics:** 470 terminal columns wide, 75 rows tall, 15 presentation layers, 56 materialized route structures.
- **Viewport 80x24:** Rendered 24 terminal lines (max width 80 chars). Verified focus on external gate `[G1]` (`ship-guarded-dependency-mutations-and-graph-queries`) with 5-edge fan-in bundle `────◆◆━━━━◆5▶│` entering `add-a-guarded-repair-path-for-broken-dependency-graphs`.
- **Viewport 140x36:** Rendered 36 terminal lines (max width 140 chars). Inspected stripped text and cell metadata; verified layer headers (`layer 7 · wave 6`, `layer 9 · wave 8 · gate`), gate double-line borders (`╔══╗`), member single-line borders (`┌──┐`), and focus inspector card layout.

#### 3. Hostile fixtures executed
1. **Fan-out with reverse edge / self-edge:** Probed `A -> B`, `A -> C`, `B -> A` and `A -> A`, `A -> B`. Demonstrated that `putRouteCount` places `"2"` at `(A.x+22, A.y+2)`, obliterating the incoming arrowhead `◀` of the reverse/self-edge.
2. **Multi-route offscreen fan-in:** Probed 20 offscreen prerequisite routes entering a single target node. Demonstrated boundary label collision on the bottom viewport boundary resulting in corrupted string fragments (`[[[[[[[[[[[[[[[[M4]…`) and total erasure of preceding prerequisite aliases (`[M1]…`, `[M2]…`, `[M3]…`).
3. **Node boundary label collision:** Probed an offscreen incident route entering at row `y=4` adjacent to a node at `x=3..24`. Centering label `[M1]…` (width 5) at lane `x=26` wrote into column 24, overwriting the node box right border `│` with `[`.
4. **Perpendicular crossing with common endpoint:** Probed long edge corridor crossing an adjacent vertical approach lane for two edges terminating at the same task (`z -> d` corridor at row 7 crossing `c -> d` vertical lane 56 at row 7). Demonstrated that `putRouteConnector` sets `shared = true` because `sharesEndpoint` is true, rendering a false junction `◆` instead of a crossing `╳`.
5. **Dense layered DAG (`TestThreadSpatialDenseLayeredFixtureIsBoundedAndDeterministic`):** 5 layers × 4 rows (20 nodes, 76 edges). Layout 171×32 cells, 0 capacity issues, verified 10 identical render iterations.
6. **Cyclic residue loops (`TestThreadSpatialCyclicResidueUsesDistinctDeterministicLoops`):** 3-node cycle with self-edge. Verified 4 distinct loop lanes (`x=26, 27, 28, 29`) and reverse arrow markers `◀`.
7. **Cascaded gate fixed point:** Instrumented `placeThreadSpatialExternalGates` with pass logging. Demonstrated that `TestThreadSpatialLayoutPlacesCascadedExternalGatesAtOneStableFixedPoint` exits on pass 0 with 0 moves.

---

### Mutation table

| Challenged Invariant | Targeted Mutation | Named Test | Observed Failure Output |
| :--- | :--- | :--- | :--- |
| **Unique lane assignment** | In `threadSpatialColumnGeometry` (`thread_spatial.go:221`), removed `+ index` so all boundary routes share lane 0 | `TestThreadSpatialCyclicResidueUsesDistinctDeterministicLoops` | `FAIL: cyclic edges shared indistinguishable loop lanes: lanes=map[26:true] routes=4` |
| **Node-free track placement** | In `buildThreadSpatialRouteSeeds` (`thread_spatial.go:171`), forced `needsTrack: false` for all routes | `TestThreadSpatialLongEdgeUsesNodeFreeTrack` | `FAIL: selected long edge did not use the node-free inter-row track` |
| **Crossing-vs-junction classification** | In `putRouteConnector` (`thread_spatial.go:831`), disabled crossing classification by forcing `if false && perpendicular` | `TestThreadSpatialDenseRoutesDistinguishCrossingsBundlesAndCounts` | `FAIL: fixture did not expose an unrelated perpendicular route crossing` |
| **Selected emphasis** | In `putRouteConnector` (`thread_spatial.go:843`), disabled `style.selected` highlighting | `TestThreadSpatialDenseRoutesDistinguishCrossingsBundlesAndCounts` | `FAIL: selected incident route did not retain emphasis through a crossing` |
| **Endpoint arrows (production pipeline)** | In `drawThreadSpatialRouteAdornments` (`thread_spatial.go:1288`), commented out `canvas.putRouteArrow` | `TestThreadSpatialDenseRoutesDistinguishCrossingsBundlesAndCounts` / entire suite | **SURVIVED (0 failures across all 16 spatial tests).** Exposed that production arrow drawing is untested; only dead helper `drawThreadSpatialEdge` had an arrow test. |
| **Window route filtering** | In `threadSpatialRoutesForWindow` (`thread_spatial.go:1258`), removed endpoint visibility filter, appending all routes | `TestThreadSpatialViewportOmitsRoutesBetweenTwoOffscreenNodes` | `FAIL: viewport routes=[{id:unattributed-highway ...} {id:selected-route ...}] want only the selected incident route` |
| **Boundary aliases** | In `annotateThreadSpatialRouteBoundaries` (`thread_spatial.go:1105`), added immediate return | `TestThreadSpatialClippedIncidentRoutesNameTheirOffscreenEndpoint` | `FAIL: selected a omitted clipped route endpoint "…▶[M4]"` and `selected d omitted clipped route endpoint "[M1]…"` |
| **Gate fixed-point repetition** | In `placeThreadSpatialExternalGates` (`thread_spatial.go:381`), limited loop to single pass (`pass < 1`) | `TestThreadSpatialLayoutPlacesCascadedExternalGatesAtOneStableFixedPoint` | **SURVIVED (Test passed with 0 failures).** Test fixture has 3 members and already settles on pass 0 with 0 moves. |
| **Honest layer labels** | In `labelThreadSpatialColumns` (`thread_spatial.go:552`), formatted label as `wave %d` instead of `layer %d` | `TestThreadSpatialColumnLabelsDistinguishLayoutLayersFromMemberWaves` | `FAIL: column 1 label "wave 1 · wave 1" misrepresents a layout layer as a wave` |
| **Pre-allocation capacity fallback** | In `renderThreadSpatial` (`thread_spatial.go:1051`), bypassed capacity issue check | `TestThreadSpatialGraphFailsOpenToWavesBeyondPrototypeCapacity` | `FAIL: capacity fallback omitted "bounded prototype fallback"` |

---

### Findings

#### H1. Route count marker overwrites reverse and self-edge arrowheads · **Status:** fixed

- **Severity:** High
- **Consequence:** When a task has multiple outgoing edges (fan-out count $\ge 2$) and also receives an incoming reverse edge, same-column cyclic edge, or self-edge terminating from the right (`arrowRune == '◀'`), the directional arrowhead is drawn at `(node.x + threadSpatialNodeWidth, node.y + slotHeight/2)`. Immediately afterward, `drawThreadSpatialRouteCounts` writes the source fan-out label (e.g. `"2"`) at that exact coordinate (`route.segments[0].from`), unconditionally clearing the cell connector and overwriting the arrowhead `◀`. The incoming dependency loses its directional pointer and re-entry indicator entirely, misleading the user into perceiving an incoming edge as an outgoing fan-out branch. In the case of self-edges (`A -> A` alongside `A -> B`), the self-loop re-enters the node with `"2"` instead of `◀`, completely concealing the self-dependency.
- **Reproduction:** Construct a projection with nodes `A, B` and edges `A -> A` and `A -> B` (or `A -> B`, `A -> C`, and `B -> A`). In `renderThreadSpatialCanvas`, cell `canvas.cells[A.y+2][A.x+22]` contains text `"2"` instead of `"◀"`.
- **Evidence:**
  `internal/tui/thread_spatial.go:1308-1310`:
  ```go
  source.point = route.segments[0].from
  source.count++
  sources[route.edge.From] = source
  ```
  `internal/tui/thread_spatial.go:903`:
  ```go
  cells := &c.cells[point.y][point.x]
  cells.text, cells.connector, cells.crossing = label, 0, false
  ```
- **Bounded recommendation:** Offset the source fan-out count marker onto the outgoing stub (e.g. `fromPoint.x + 1`) rather than placing it at `fromPoint` (`node.x + threadSpatialNodeWidth`), or prevent `putRouteCount` from overwriting any cell that already contains a directional arrowhead (`◀` or `▶`).

---

**Resolution:** Source fan-out and side-specific target counts occupy reserved
cells beyond the arrowheads; self, reverse, and mixed-direction regressions pin
both arrows and counts.

#### H2. Viewport boundary route labels overwrite node boxes and corrupt on multi-route entries · **Status:** fixed

- **Severity:** High
- **Consequence:**
  1. **Node Box Overwrite:** `putThreadSpatialBoundaryLabel` centers the boundary label on the route's boundary anchor coordinate (`anchor.x - labelWidth/2`). When an offscreen incident route enters near a node boundary (e.g. at lane `anchor.x = 26` next to a node occupying columns 3..24), the centered label spans columns 23..28, overwriting the node box right border `│` and interior text with `[` and alias characters.
  2. **Boundary Label Corruption:** When multiple offscreen incident routes enter the viewport across a boundary at adjacent columns (e.g. inter-row tracks entering across the bottom viewport row), each route independently writes its centered label. Consecutive labels shifted by 1 column repeatedly overwrite one another, rendering corrupted strings of stacked brackets (e.g. `[[[[[[[[[[[[[[[[M4]…`).
  3. **Fan-In Alias Erasure:** When multiple offscreen prerequisites enter a selected node on the same horizontal row (e.g. forward-adjacent edges entering `(panX, target.y+2)`), each route calls `putThreadSpatialBoundaryLabel` at the identical coordinate. The last processed route completely overwrites earlier labels, causing all but one offscreen prerequisite alias to vanish.
- **Reproduction:**
  1. Render a node at `x=3..24` with an offscreen route anchor at `x=26`: `canvas.cells[4][24].text` becomes `"["` instead of `"│"`.
  2. Render a selected node with 20 offscreen incident routes entering across the bottom boundary: renders `[[[[[[[[[[[[[[[[M4]…` on the boundary row.
  3. Render a selected node with 3 offscreen forward-adjacent prerequisites: only `[M3]…` is visible; `[M1]…` and `[M2]…` are overwritten.
- **Evidence:**
  `internal/tui/thread_spatial.go:1167-1179`:
  ```go
  func putThreadSpatialBoundaryLabel(canvas *threadSpatialCanvas, anchor threadSpatialPoint, label string, panX, panY, width, height int) {
      if anchor.y < panY || anchor.y >= panY+height {
          return
      }
      labelWidth := ansi.StringWidth(label)
      x := min(max(anchor.x-labelWidth/2, panX), max(panX+width-labelWidth, panX))
      canvas.putText(x, anchor.y, label, theme.ColorYellow, true)
  }
  ```
- **Bounded recommendation:** Make boundary label placement collision-aware: prevent writes into node box bounding boxes (`[node.x, node.x+threadSpatialNodeWidth)`), bundle or deduplicate overlapping offscreen aliases entering the same boundary row (e.g. `[M1,M2,M3]…` or `[3 tasks]…`), and suppress labels when adjacent columns would produce smeared character runs.

---

**Resolution:** Boundary annotations now group deterministic alias sets by
boundary and endpoint meaning, compact large groups, and search only unoccupied
visible cells instead of overwriting nodes or prior labels.

#### M1. Perpendicular crossings of routes with common endpoint render as false junctions · **Status:** fixed

- **Severity:** Medium
- **Consequence:** In `putRouteConnector`, when two perpendicular routes intersect, the code evaluates `sharesEndpoint := cell.routeFrom == style.from || cell.routeFrom == style.to || cell.routeTo == style.from || cell.routeTo == style.to`. If the two routes terminate at the same target task (such as a long route's horizontal corridor bypassing toward its arrival lane while an adjacent route travels vertically on its own approach lane toward the same target), `sharesEndpoint` evaluates to `true`. This bypasses the crossing branch (`if perpendicular && !sharesEndpoint`) and executes the bundle branch (`cell.shared = true; cell.connector |= directions`). Because the cell has both horizontal and vertical directions, `threadSpatialSharedConnectorGlyph` renders `◆` (shared junction diamond). The two routes do not join at this intersection; they are merely crossing paths to reach their respective entry lanes. Rendering `◆` manufactures visual connectivity between independent paths.
- **Reproduction:** Construct a fixture where `z-source -> d-target` (long route with corridor at `y=7` entering at lane 57) intersects `c-mid -> d-target` (vertical lane 56 extending from `y=10` to `y=4`). At coordinate `(56, 7)`, the vertical and horizontal segments cross. The cell renders `◆` with directions `Left|Right|Up|Down` instead of `╳`.
- **Evidence:**
  `internal/tui/thread_spatial.go:828-840`:
  ```go
  sharesEndpoint := cell.routeFrom == style.from || cell.routeFrom == style.to ||
      cell.routeTo == style.from || cell.routeTo == style.to
  perpendicular := (existingHorizontal && incomingVertical) || (existingVertical && incomingHorizontal)
  if perpendicular && !sharesEndpoint {
      cell.connector = 0
      cell.corridor = false
      cell.shared = false
      cell.crossing = true
  } else {
      cell.connector |= directions
      cell.corridor = cell.corridor || style.corridor
      cell.shared = true
  }
  ```
- **Bounded recommendation:** Differentiate true shared geometry (where routes share a directional segment from a common node) from perpendicular intersections. If one route is purely horizontal and the other is purely vertical at the intersection point, the cell must remain a crossing `╳` regardless of whether the routes eventually terminate at the same task downstream.

---

**Resolution:** Straight perpendicular intersections remain crossings even when
routes share a distant endpoint; only a turn that actually joins their common
endpoint stub uses shared-bundle grammar.

#### M2. Dead legacy edge drawing helpers remain while production arrow pipeline is unasserted · **Status:** fixed

- **Severity:** Medium
- **Consequence:** The functions `drawThreadSpatialEdge` (`thread_spatial.go:1335-1364`) and `drawThreadSpatialLongEdge` (`thread_spatial.go:1370-1394`) are uncalled by any production code in `internal/tui`. Production canvas rendering exclusively routes through `renderThreadSpatialCanvasWindow` -> `drawThreadSpatialRouteSegments` / `drawThreadSpatialRouteAdornments`. The sole caller of `drawThreadSpatialEdge` in the repository is `TestThreadSpatialConnectorGeometryUsesElbowsAndAccentFocus` (`thread_projection_test.go:1049-1073`), which tests obsolete prototype midpoint geometry (`middle := fromX + max(1, (toX-fromX)/2)`). Consequently, the actual production arrowhead drawing logic in `drawThreadSpatialRouteAdornments` is completely unasserted by tests: commenting out `canvas.putRouteArrow` in `drawThreadSpatialRouteAdornments` passes 100% of the test suite.
- **Reproduction:**
  1. `git grep drawThreadSpatialEdge` shows only `thread_spatial.go:1335` and `thread_projection_test.go:1053`.
  2. Comment out `canvas.putRouteArrow` in `drawThreadSpatialRouteAdornments` (`thread_spatial.go:1288`). Run `go test -v -run TestThreadSpatial ./internal/tui`; all tests pass.
- **Evidence:**
  `internal/tui/thread_spatial.go:1335-1394`
  `internal/tui/thread_projection_test.go:1049-1073`
- **Bounded recommendation:** Delete `drawThreadSpatialEdge` and `drawThreadSpatialLongEdge`. Refactor `TestThreadSpatialConnectorGeometryUsesElbowsAndAccentFocus` to render through `renderThreadSpatialCanvas` and assert that arrowheads and connector elbows are actually rendered onto the production canvas.

---

**Resolution:** Deleted the dead legacy drawing pipeline and made production
canvas regressions assert every materialized route arrowhead and semantic color.

#### L1. Cascaded external gate test succeeds on pass 0 and fails to test multi-pass convergence · **Status:** fixed

- **Severity:** Low
- **Consequence:** `TestThreadSpatialLayoutPlacesCascadedExternalGatesAtOneStableFixedPoint` (`thread_projection_test.go:911-960`) includes a docstring claiming:
  `Canonical presentation order visits the upstream gate first. Its downstream gate moves later in pass one, forcing the upstream placement to settle on a subsequent fixed-point pass.`
  In reality, the test fixture defines only 3 member tasks (`memberA -> memberB -> memberC`), which places `memberC` at column 2. Initial longest-path ranking puts `upstreamGate` at column 0 and `downstreamGate` at column 1. On pass 0 of `placeThreadSpatialExternalGates`, `upstreamGate`'s dependent is at column 1 (`desired = 0`), and `downstreamGate`'s dependent is at column 2 (`desired = 1`). Neither gate moves. The loop exits after pass 0 with `moved == false`. If `placeThreadSpatialExternalGates` is mutated to allow only 1 pass (`pass < 1`), the test still passes, failing to pin the multi-pass fixed-point invariant.
- **Reproduction:** Add a log statement to `placeThreadSpatialExternalGates` on loop exit (`println("finished after pass", pass)`). Run `TestThreadSpatialLayoutPlacesCascadedExternalGatesAtOneStableFixedPoint`; it prints `finished after pass 0`.
- **Evidence:**
  `internal/tui/thread_projection_test.go:911-960`
  `internal/tui/thread_spatial.go:381-408`
- **Bounded recommendation:** Extend the member chain in the test fixture to 4 tasks (`memberA -> memberB -> memberC -> memberD`), with `downstreamGate` gating `memberD`. In that configuration, `downstreamGate` advances to column 2 on pass 0, leaving `upstreamGate` at column 0 until pass 1 advances it to column 1, genuinely exercising multi-pass fixed-point convergence.

---

**Resolution:** The gate fixture now has four member layers, forcing the
downstream gate to advance on the first pass and the earlier-visited upstream
gate on a subsequent pass.

#### L2. Layout built redundantly on every render frame before capacity guards · **Status:** tracked by 6g8ezj5e51hg

- **Severity:** Low
- **Consequence:** When `renderDetail` is called (`detail.go:857`), it evaluates `d.detailSelectionKey()`, which calls `threadSpatialSelectedTaskID` (`detail.go:862`), which in turn calls `buildThreadSpatialLayout(projection)` (`thread_spatial.go:656`). Then `renderDetail` calls `renderThreadSpatial(...)`, which executes `buildThreadSpatialLayout(projection)` a second time (`thread_spatial.go:1043`). On navigation keypresses, `threadSpatialMove` invokes `buildThreadSpatialLayout` a third time. Furthermore, `buildThreadSpatialLayout` performs all longest-path ranking, gate placement, geometry assignment, seed generation, and route materialization before `threadSpatialCapacityIssue` is checked in `renderThreadSpatial`. While canvas cell allocation (`750,000` cells) is protected, layout construction work is unbudgeted and executed repeatedly on every frame.
- **Reproduction:** Instrument `buildThreadSpatialLayout` with an invocation counter; every single `renderDetail` call executes layout construction twice.
- **Evidence:**
  `internal/tui/detail.go:857, 862`
  `internal/tui/thread_spatial.go:656, 1043`
- **Bounded recommendation:** Cache the layout on `threadDetail` or pass `threadSpatialLayout` across selection key determination and rendering. Add an early node and edge count guard at the top of `buildThreadSpatialLayout` before running multi-pass ranking and route materialization.

---

**Resolution:** The current patch bounds cell memory and lowers the canvas
ceiling. The linked follow-up owns pre-layout node/edge guards and one coherent
immutable layout reused across render and navigation without stale reloads.

### Corroboration of surviving invariants

The following core invariants were subjected to targeted adversarial testing and confirmed to hold robustly:

1. **Deterministic Equivalence Invariance:**
   `buildThreadSpatialLayout`, `threadGraphAliases`, `threadSpatialConnections`, and `renderThreadSpatial` produce identical layout coordinates, route structures, inspector connections, and rendered characters across arbitrary slice reorderings of `projection.Nodes`, `projection.Edges`, and `projection.Waves`, as well as duplicate edges. Verified via `TestThreadSpatialPresentationIsInvariantUnderEquivalentProjectionPermutations`, `TestThreadSpatialCyclicResidueUsesDistinctDeterministicLoops`, and hostile permuted fixtures.

2. **Node-Free Inter-Row Corridors for Long Edges:**
   Long and skipped-layer routes reserve explicit inter-row tracks (`trackY`) that run in the whitespace between node slots. Hostile probes verifying every discrete cell coordinate along all route segments confirmed that long routes never intersect node bounding boxes (`TestThreadSpatialRoutesStayOutsideNodeBoxes`).

3. **Viewport Offscreen-to-Offscreen Route Omission:**
   `threadSpatialRoutesForWindow` omits all routes whose source and target nodes are both outside the current viewport, ensuring that panning over dense graphs does not render unowned highways of unattributed lines across the screen (`TestThreadSpatialViewportOmitsRoutesBetweenTwoOffscreenNodes`).

4. **Deepest-Row Track Canvas Invariant:**
   Horizontal routing tracks for edges whose shallower endpoint is on the deepest layout row remain strictly within the allocated canvas height, resolving predecessor audit finding 6g8bb1mf9hw0 H1 (`TestThreadSpatialLongEdgeKeepsDeepestRowTrackInsideCanvas`).

5. **Layer vs. Wave Label Honesty:**
   Presentation column headers consistently identify `layer N` rather than claiming presentation columns are member waves, preserving core wave semantics (`TestThreadSpatialColumnLabelsDistinguishLayoutLayersFromMemberWaves`).

6. **Pre-Allocation Capacity Guard:**
   `threadSpatialCapacityIssue` correctly enforces the limits (512 nodes, 2048 edges, 750,000 canvas cells) before canvas matrix allocation, cleanly failing open to the wave reader (`TestThreadSpatialGraphFailsOpenToWavesBeyondPrototypeCapacity`, `TestThreadSpatialCapacityGuardsEdgesAndCanvasIndependently`).

7. **Narrow Terminal Handling:**
   When terminal width < 60 or height < 12, `renderThreadSpatialNarrow` allocates 0 canvas cells and provides an explanatory message directing users to resize or return to the wave reader (`TestThreadSpatialNarrowFallbackAllocatesNoCanvasBeforeCapacityChecks`).

8. **Semantic Status Colors and Focus Separation:**
   Focus selection uses a yellow pointer (`›`) and palette-accented touching edge strokes, while node box borders strictly retain their semantic status colors (`domain.Status` mapping to `theme.Status`), ensuring status meaning is never overwritten by selection state (`TestThreadSpatialSelectedNodeExpandsInsideStableSlot`).

---

### Residual uncertainty

1. **Real Terminal Emulator Window Resizing:**
   All viewport window and panning tests were executed headlessly via Bubble Tea model message updates and ANSI string slices. Real PTY interaction under rapid mouse-wheel scrolling or terminal window drag-resize events was not directly exercised on an interactive terminal emulator.
2. **Double-Width Unicode / CJK Task Labels in Boundary Labels:**
   While node labels are truncated using `ansi.StringWidth`, task aliases used in boundary labels (`[M1]…`, `[G1]…`) are ASCII. If future aliases or boundary markers include double-width or combining characters, centering calculations (`anchor.x - labelWidth/2`) could introduce terminal cell alignment offsets.
