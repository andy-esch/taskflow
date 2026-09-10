---
schema: 1
id: 6g8e0pyczpns
bucket: closed
area: dense-thread-graph-routes-implementation-claude
date: "2026-09-09"
updated_at: "2026-09-09"
---
# Audit: Dense Thread graph route hardening implementation — claude — 2026-09-09

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

### 1. Executive verdict

**Do not ship.** The routing rewrite is well-factored, deterministic under permutation, bounded
before canvas allocation, and free of any leak into `core`/`domain`/`wire`/`store`. But its central
claim — that a rendered connector faithfully represents the supplied projection — fails on the
smallest non-planar shape the layout can produce. A cell keeps **one** route identity and **one**
merged connector bitmask, so it cannot distinguish a route *turning* from a route *passing through*.
Every turn that lands on a cell another route already occupies is reclassified as a crossing and the
turn is erased. On a two-edge crossing (`a→d`, `b→c`) the picture reads as `a→c` and `b→d`: both
edges acquire the wrong endpoint. This reproduces on the repository's own
`complete-production-threads` Thread (7 severed corners across 56 routes) and inside the change's own
headline regression fixture, where the rendered row reads `c ─▶ e` for a projection that contains no
`c→e` edge.

Three of the ten invariants the brief asked me to challenge are **not pinned by any test** — proven
by mutation, not inspection: endpoint fan counts (M11b), the shared-bundle grammar (M15), and the
external-gate fixed-point repetition (M8b, where deleting the entire gate-placement function leaves
the new cascaded-gate test green). Two production-dead helpers (`drawThreadSpatialEdge`,
`drawThreadSpatialLongEdge`) are still exercised by a live test, which reports on geometry the user
never sees.

Fixes are bounded and local to `internal/tui/thread_spatial.go`; nothing here argues for redesigning
Thread semantics or building a general layout engine.

### 2. Isolation attestation

All inspection, builds, tests, generators, fixtures, mutations and report editing were performed in
the independent `--no-hardlinks` clone below. `$SOURCE_ROOT` was read for the brief and the initial
copy only; no state-changing command ran there. Go build cache, golangci-lint cache and scratch
binaries were kept **outside** the sandbox
(`/private/var/folders/…/T/isolated-review-caches/`) so the workspace holds no artefacts of my own.
Every mutation probe was reverted with `git checkout -- internal/tui/thread_spatial.go` and every
scratch test file (`internal/tui/zz_review_probe*_test.go`, 10 files) was deleted before the report
was written; `git status --porcelain --untracked-files=all` was empty afterwards, and
`go test ./internal/tui/` was re-run green on the restored baseline.

```
sandbox_path=/private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.mR7nsP
git_dir=/private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.mR7nsP/.git
baseline_commit=54e6e0a9276c59b707d716ddd8f7de8b0ae5c52f
source_blob=008ccacd90fb7e5e054f30e7667426f36da591ed
source_fingerprint=5af9084d9aed8205e5ef519c5d2beeba40f807b2
deliverable=planning/audits/6g8e0pyczpns-2026-09-09-dense-thread-graph-routes-implementation-claude.md
deliverable_changed=false   (verify, before the report was written)
transfer=pending
```

The transfer attestation is appended in §9.

### 3. Consumer and boundary inventory

Every consumer below was located with `grep -rn --include="*.go" "<symbol>("` over the whole
sandbox, not inferred from naming.

**Production consumers of the spatial adapter (all in `internal/tui/detail.go`):**

| Call site | Symbol | Assumption it makes |
| --- | --- | --- |
| `detail.go:857` | `renderThreadSpatial(d.projection, d.pathIssue, d.detailSelectionKey(), width, height, s)` | terminal `width`/`height` may be ≤ 0; output must be ≤ `height` lines and ≤ `width` cells |
| `detail.go:862`, `detail.go:869` | `threadSpatialSelectedTaskID` | a selection key survives reload, or falls back to the first selectable node |
| `detail.go:901` | `threadSpatialMove` | arrow/`hjkl` navigation; no capacity guard on this path |
| `detail.go:845-851` | `detailImmersive/Sized/Directional` | the spatial view alone is sized and directional |
| `detail.go:909` | `threadGraphTask` (not spatial) | selection key is a projection task ID, so `threadSpatialMove`'s return value must always be one |

**Shared helpers the layout depends on (`internal/tui/detail.go`):** `orderedThreadGraphNodes:1259`
and `orderedThreadGraphEdges:1265` (canonical `TaskID` / `From,To` sort — the *only* thing that makes
layout order-independent), `threadGraphAliases:1276` (`G<n>`/`M<n>`/`?<n>`, ASCII-only, so boundary
labels are never wide-rune), `terminalText:1362` (control-rune scrubbing). `truncate`
(`internal/tui/style.go:162`) and `padRight` (`:172`) are `ansi.StringWidth`-based.

**Internal layout pipeline** (`internal/tui/thread_spatial.go`), each with its sole caller:
`buildThreadSpatialLayout:83` ← `:589`, `:656`, `:1043`; `rankThreadSpatialColumns:454` ← `:115`;
`placeThreadSpatialExternalGates:351` ← `:116`; `labelThreadSpatialColumns:530` ← `:117`;
`buildThreadSpatialRouteSeeds:151` ← `:119`; `threadSpatialColumnGeometry:189` ← `:120`;
`threadSpatialRowGeometry:229` ← `:125`; `materializeThreadSpatialRoutes:262` ← `:147`;
`threadSpatialRoutesForWindow:1248` ← `:1234`; `annotateThreadSpatialRouteBoundaries:1099` ← `:1067`;
`threadSpatialCapacityIssue:1181` ← `:1051`.

**Boundary check — no route-derived state escaped the TUI.** `git diff --name-only 93b4965 HEAD`
returns exactly `internal/tui/thread_spatial.go`, `internal/tui/thread_projection_test.go` and four
`planning/` files. `internal/core/thread_graph.go` is byte-identical to `main`; nothing in
`core`/`domain`/`wire`/`store` changed. The renderer reads `ThreadGraphProjection` fields only and
never writes back. **This contract holds and survived attack.**

**Neighbouring TUI presentation.** `threadSpatialNodeStrideY:24` is now referenced only from
`thread_spatial.go:1381` (inside dead code, see L3) and from
`thread_projection_test.go:1422` and `:1452`. No other TUI view reads spatial geometry.

**Test-only consumers that mask the shipped path.** `renderThreadSpatialCanvas:1216` is dead in
production; its four callers (`thread_projection_test.go:1092,1417,1456,1514`) all pass the window as
the *whole* layout, so every canvas-level assertion in the suite bypasses
`threadSpatialRoutesForWindow` entirely.

### 4. Commands and hostile fixtures actually run

All in `$SANDBOX` with `GOCACHE`/`GOLANGCI_LINT_CACHE` pointed outside it.

| Command | Result |
| --- | --- |
| `go test ./...` | all packages `ok` (go1.26.6 darwin/arm64) |
| `go test -race ./internal/tui/ ./internal/core/` | `ok` 8.842s / 1.494s |
| `go test ./internal/tui/ -run ThreadSpatial -count=5` | `ok` (5 repeats, stable) |
| `go vet ./...` | clean |
| `golangci-lint run ./...` | `0 issues.` |
| `go mod tidy -diff` | exit 0 |
| `gofmt -l cmd internal` | empty |
| `go run ./internal/tools/docgen -out docs/cli` + `git diff --stat -- docs/cli` | no drift |
| `go build ./cmd/tskflwctl` then `tskflwctl lint` | `✔ all planning entities and dependency links pass lint` |
| `git diff --check` | clean |

All of the implementation checkpoint's validation claims reproduce. The failures below are *new*
evidence, not failures of the shipped suite.

**Hostile fixtures I constructed** (scratch tests, since deleted; each named result is reproducible
from the description):

1. **Two-edge crossing** — `a,b` × `c,d`, edges `a→d`, `b→c`. Both corners severed (H1).
2. **Three-edge reversal** — `a→z`, `b→y`, `c→x`, found by exhaustive search over all 502
   two-to-four-edge subsets of a 3×3 bipartite fixture. Smallest false-bundle case (H2).
3. **Dense layered DAG** — 4 layers × 4 rows, complete between adjacent layers, 48 routes:
   **45/48 routes have a severed corner**; 0 lost arrowheads.
4. **8-node / 8-edge and 8-node / 10-edge crossbars** — 4 and 9 defect instances respectively.
5. **Fan-in at focus** — `a→c`, `b→c`, selected `c`: count cell renders `›` not `2` (H3).
6. **Mixed-direction fan-in** — `r→x`, `x→y`, `y→x` (cyclic residue): `2` marker placed beside the
   single right-entry `◀` while two routes converge (M1).
7. **Self-edge** — `a→a` plus `a→b`, and `a→a` twice: the `◀` arrowhead is overwritten by the
   fan-out count in both (M2).
8. **Duplicate/parallel edges** — three identical `a→b`: three routes, one stroke, `3` at both ends.
9. **Non-contiguous wave set** — gate chains inflating columns without inflating waves produced
   `layer 5 · waves 1–5` for a layer holding exactly waves {1,3,5} (L1).
10. **Permutation determinism** — 400 Fisher–Yates permutations of nodes, edges *and* waves over a
    7-node/9-edge fixture with a duplicate edge and a two-gate chain: `reflect.DeepEqual` on the
    layout and byte-equality on `renderThreadSpatial(…,120,30)` held for all 400.
11. **Wide labels** — CJK node labels at 80×24, panned to three different selections: no line
    exceeded 80 display cells; `putText` continuation cells kept `ansi.Cut` aligned.
12. **Degenerate projections** — empty; nodes-only; edges naming absent endpoints; waves naming
    tasks absent from `Nodes`; an empty `TaskID`. Rendered at 0×0, 1×1, 40×8, 60×12, 80×24, 200×60.
    No panic, no overflow, line count never exceeded `height`.
13. **Viewport, integrated** — 10-node chain plus an `n0→n9` highway, layout 304×8, panned to
    `panX=124`: the highway and all five other non-incident routes were correctly **omitted** while
    the four routes incident to visible nodes were drawn.
14. **Endpoints clipped on each boundary** — 9-node fixture at 62×16, selections `c0,c2,c4,v0,v3`:
    offscreen prerequisites and dependents were named (`[M2]…`, `…▶[M4]`) on every boundary; no
    line overflowed.
15. **Real production Threads** — `core.NewService(store.NewFS("planning"))` →
    `ShowThreadGraph("complete-production-threads")` (42 nodes, 56 edges, 14 waves, layout 470×75,
    15 layers) and `"tool-owned-actionable-sub-entities"`, rendered at **80×24 and 140×36**. Stripped
    text and per-cell metadata inspected. 7 severed corners; 0 lost arrowheads; no width or line
    overflow.

**Resource observations** (`runtime.MemStats` deltas around a single `renderThreadSpatial`,
`unsafe.Sizeof`):

- `threadSpatialCell` is **96 bytes**; `routeID`/`routeFrom`/`routeTo` are 48 of them.
- The 750,000-cell guard therefore permits **~69 MB of canvas per frame**. Measured at the boundary
  (440 gates, 504 nodes, 13255×56 = 742,280 cells): **73.9 MB and 11.1 ms for one render**, inside
  the guard. At 775,880 cells the guard fires and the fallback costs 1.8 MB / 1.3 ms.
- The node and edge limits are effectively unreachable: a 512-node dense layout is already
  1,255,508 cells, so the canvas limit fires first for anything but a very sparse graph.
- Pre-guard work (`buildThreadSpatialLayout` runs at `:1043`, before the check at `:1051`):
  2.3 ms at 512 nodes, 7.8 ms at 2,048, 31.8 ms at 8,192. `threadSpatialMove` has **no** guard at
  all: 3.0 ms / 22.1 ms / 244.7 ms for the same three sizes.

### 5. Mutation table

Each mutation was applied to the restored baseline, compiled, run against the named test, then
reverted. "Coordinated" marks mutations that also had to keep a neighbouring symbol live so the
package still compiled.

| # | Invariant challenged | Mutation | Named test | Observed |
| --- | --- | --- | --- | --- |
| M1b | unique lane assignment | drop `+ index` in `lanes[boundary][routeID]` (coordinated: `_ = index`) | `TestThreadSpatialCyclicResidueUsesDistinctDeterministicLoops` | **FAIL** — `cyclic edges shared indistinguishable loop lanes: lanes=map[26:true] routes=4` |
| M2 | node-free track placement | `trackY = rowY[row] + index` (drop `+ slotHeight`) | `TestThreadSpatialDenseLayeredFixtureIsBoundedAndDeterministic`, `TestThreadSpatialLongEdgeUsesNodeFreeTrack` | **FAIL** — `dense route l0-r0 -> l2-r0 crossed node l1-r0 at {x:47 y:2}` |
| M3 | crossing-vs-junction | disable the `perpendicular && !sharesEndpoint` branch | `TestThreadSpatialDenseRoutesDistinguishCrossingsBundlesAndCounts` | **FAIL** — `fixture did not expose an unrelated perpendicular route crossing` |
| M4 | selected emphasis | `putRouteConnector` never sets `accent`/`bold` | same, and `…ConnectorGeometryUsesElbowsAndAccentFocus` | **FAIL** — `selected incident route did not retain emphasis through a crossing` |
| M5 | endpoint arrows | `putRouteArrow` returns immediately | `TestThreadSpatial…` (suite) | **FAIL** — `dependency endpoint="" want arrowhead` |
| M6b | window route filtering | include every route (coordinated: keep `visibleNode` live) | `TestThreadSpatialViewportOmitsRoutesBetweenTwoOffscreenNodes` | **FAIL** — `viewport routes=[unattributed-highway …] want only the selected incident route` |
| M7 | boundary aliases | `annotateThreadSpatialRouteBoundaries` returns immediately | `TestThreadSpatialClippedIncidentRoutesNameTheirOffscreenEndpoint` | **FAIL** — both `…▶[M4]` and `[M1]…` missing |
| M8 | gate fixed-point **repetition** | `for pass := 0; pass < 1` | `TestThreadSpatialLayoutPlacesCascadedExternalGatesAtOneStableFixedPoint` | **PASS — not pinned** |
| M8b | gate placement at all | `placeThreadSpatialExternalGates` returns `columns` unchanged | same test | **PASS — not pinned** (M8c shows the *older* `…PullsSourceGateBesideItsFirstDependent` does fail here) |
| M9 | honest layer labels | `"layer %d"` → `"wave %d"` | `TestThreadSpatialColumnLabelsDistinguishLayoutLayersFromMemberWaves` | **FAIL** — `column 1 label "wave 1 · wave 1" misrepresents a layout layer as a wave` |
| M10 | pre-allocation capacity fallback | `threadSpatialCapacityIssue` always `""` | `TestThreadSpatialGraphFailsOpenToWavesBeyondPrototypeCapacity`, `…CapacityGuardsEdgesAndCanvasIndependently` | **FAIL** — `capacity fallback omitted "bounded prototype fallback"` |
| M11b | endpoint fan counts | `drawThreadSpatialRouteCounts` returns immediately | `TestThreadSpatialDenseRoutesDistinguishCrossingsBundlesAndCounts` | **PASS — not pinned** |
| M12 | fan-count placement | `target.point.y += 3` | whole `TestThreadSpatial` suite | **PASS — not pinned** |
| M13 | corridor attribution | `corridor: false` in `newThreadSpatialRoute` | dense-routes and reverse-route tests | **FAIL** — `skipped-layer route omitted corridor attribution: corridors=0` |
| M14 | (diagnostic) column-label text | blank the label `putText` | dense-routes test | **PASS** — confirms the `"2"` assertion does not depend on labels |
| M15 | shared-bundle grammar | `cell.shared = false` at `:839` | whole `TestThreadSpatial` suite | **PASS — not pinned** (the surviving `shared` cells come from `putRouteArrow:862`) |
| M16 | corridor waypoints | `putRouteWaypoint` returns immediately | dense-routes test | **FAIL** — `waypoints=0` |

M11b's result is decisive on its own: with *no* fan-count marker drawn anywhere, the assertion
`if routeCounts < 2 { t.Fatalf("fan-in/fan-out endpoint multiplicity was not annotated…") }`
(`thread_projection_test.go:1138-1140`) still passes, because it counts any canvas cell whose text is
the literal `"2"` — including the `layer 2 · …` heading and the focused node's `[M2]` alias line.
I confirmed by classification that the fixture does contain two genuine markers at `(61,4)` and
`(25,10)`, and two decoys at `(39,0)` and `(7,10)`; the decoys alone satisfy the assertion.

### 6. Findings

#### H1. A turning route's corner is reclassified as a crossing, so edges acquire the wrong endpoint · **Status:** fixed

**Severity:** high. **Consequence:** on the most elementary non-planar shape — two edges that cross —
the rendered graph shows the opposite dependencies from the ones in the projection. This is the
change's headline contract ("Every supplied projection edge … receives one deterministic, node-free
route from prerequisite to dependent") failing at the presentation layer.

`putRouteConnector` (`internal/tui/thread_spatial.go:805-848`) decides "crossing" from the *merged*
connector bitmask of the cell:

```go
existingHorizontal, existingVertical := threadSpatialConnectorAxes(cell.connector)   // :826
incomingHorizontal, incomingVertical := threadSpatialConnectorAxes(directions)        // :827
perpendicular := (existingHorizontal && incomingVertical) || (existingVertical && incomingHorizontal) // :830
if perpendicular && !sharesEndpoint { … cell.crossing = true }                        // :831-835
```

A corner cell of route *R* carries **both** axes for *R* alone (`└` is `Up|Right`). When any
unrelated route then draws a straight stroke through that cell, `perpendicular` is true and the
branch fires — even though nothing is passing through *R*: *R* is turning. `cell.connector` is zeroed
(`:832`) and, because `case cell.crossing:` at `:816-818` is a no-op, *R* can never re-assert the
turn. The turn is gone; the `╳` reads as a pass-through.

**Reproduction** (four nodes, two edges, no gates, no cycles, default 22×5 geometry):

```go
projection := core.ThreadGraphProjection{
    Nodes: []core.ThreadGraphNode{{TaskID:"a",Role:core.ThreadTaskMember}, {TaskID:"b",…},
                                  {TaskID:"c",…}, {TaskID:"d",…}},
    Edges: []core.ThreadGraphEdge{{From:"a",To:"d"}, {From:"b",To:"c"}},
}
```

renders (colour stripped, 90×24, selection `a`):

```
   ┌────────────────────┐        ┌────────────────────┐
   │ • a                │─┐┌────▶│ • c                │
   └────────────────────┘ ││     └────────────────────┘
                          ││
   ┌────────────────────┐ ││     ┌────────────────────┐
   │ • b                │─╳╳────▶│ • d                │
   └────────────────────┘        └────────────────────┘
```

`a→d`'s corner at `(26,10)` and `b→c`'s corner at `(27,10)` are both `crossing:true,
connector:0`. Two vertical strokes descend and terminate in crossing marks with nothing below them,
while an unbroken horizontal runs from `b`'s border into `d`'s arrowhead. The reading is `a→c` and
`b→d`. The inspector line underneath simultaneously and correctly says `unlocks [M4] d` — the graph
and the inspector contradict each other.

**Evidence:**

- Exhaustive search over the 3×3 bipartite fixture space found `{a→z, b→y, c→x}` — a pure reversal —
  drawn as three straight horizontals `a→x`, `b→y`, `c→z`; two of three edges misattributed.
- 4-layer × 4-row dense fixture: **45 of 48 routes** have at least one corner on a `crossing:true`
  cell. 8-edge crossbar: 4 of 8 (`a→h` at `(27,22)`, `b→g` at `(29,16)`, `c→f` at `(30,16)`,
  `d→e` at `(32,22)`).
- **Real data:** `ShowThreadGraph("complete-production-threads")` (42 nodes, 56 edges) → **7 severed
  corners**. At 140×36 the visible row reads
  `│ ✔ preserve-cohere… │─╳╳────▶` and `──┘┌─╳◆━━━━2▶║`.
- The change's own fixture in `TestThreadSpatialDenseRoutesDistinguishCrossingsBundlesAndCounts`
  (`thread_projection_test.go:1075-1091`) renders row `y=4` as
  `│ • c │─╳╳◆━━2▶│ • e │`. Per-cell ownership: `x=55 ─ [c→f]`, `x=56 ╳ [b→e, c→f]`,
  `x=57 ╳ [b→e, c→f]`, `x=58 ◆ [b→e, d→e]`, `x=59,60 ━ [b→e, d→e]`, `x=61 "2"`, `x=62 ▶`.
  There is no `c→e` edge in the projection; `c`'s only outgoing edge is `c→f`, whose corner at
  `(57,4)` was severed. The test asserts `crossings > 0` and therefore **pins the defect**.

**Recommendation (bounded):** classify at the *route* level, not the cell level. `putRouteConnector`
already receives the incoming `directions` for one route; a cell is a genuine crossing only when the
incoming route contributes a single axis **and** the resident route contributed a single axis. Track
the resident route's own axes (one extra `threadSpatialConnector` field, ~8 bits) rather than
inferring them from the merged mask; when either route turns, keep the turning route's elbow and
render the other as an over/under break (or nudge the turn one cell along its lane — lanes and tracks
are already unique per route, so a free neighbouring cell always exists). Then update the
`crossings > 0` assertion to a fixture with a real pass-through.

**Resolution:** Assigned separate source and target lanes, preserve legitimate
endpoint merges, and render unrelated turn collisions as explicit invariant
failures; production crossed-edge and dense fixtures pin every corner and
arrowhead.

#### H2. Independent routes that merely overlap are drawn with the "shared bundle" grammar · **Status:** fixed

**Severity:** high. **Consequence:** the legend teaches `◆ shared` as "these strokes belong to routes
that meet here". The implementation applies it to any two collinear routes that happen to occupy the
same cell, with no endpoint in common — manufacturing connectivity the projection does not contain.
This is the second half of the intended contract ("Geometry shared by routes with a real common
endpoint … remains distinguishable **without manufacturing connectivity**").

`sharesEndpoint` is computed at `:828-829` but is consulted **only** inside the `perpendicular`
guard at `:831`. The `else` arm at `:836-840` sets `cell.shared = true` unconditionally, so two
parallel, unrelated routes sharing a cell are bundled:

```go
if perpendicular && !sharesEndpoint {
    … crossing …
} else {
    cell.connector |= directions
    cell.corridor = cell.corridor || style.corridor
    cell.shared = true            // :839 — no endpoint check on this path
}
```

**Reproduction:** the same three-edge reversal `{a→z, b→y, c→x}` (six nodes, no gates). Cell
`(27,16)` renders `━` (`threadSpatialSharedConnectorGlyph`, `:968-978`) and is occupied by `a→z` and
`c→x`, which share no task:

```
   │ • a                │─┐ ┌───▶│ • x                │
   │ • b                │─╳─╳───▶│ • y                │
   │ • c                │─╳━╳───▶│ • z                │
```

**Evidence:** in the 10-edge crossbar (`a,b,c,d` × `e,f,g,h` plus `a→h, d→e, b→g, c→f, a→f, d→g`),
9 cell/route-pair occurrences; e.g. `(30,22)` renders `━` and carries `a→h`, `d→e` and `d→g` — no
task is common to all three. M15 shows no test pins this grammar at all: replacing `:839` with
`cell.shared = false` leaves the entire `TestThreadSpatial` suite green, because the `shared == 0`
guard in the dense-routes test is satisfied by `putRouteArrow`'s separate `cell.shared = true`
(`:862`) when two fan-in arrows land on one cell.

**Recommendation (bounded):** gate `:839` on `sharesEndpoint` as well. Overlap without a common
endpoint is a *bundle of independent routes*, which needs its own grammar (or, better, avoidance) —
not the same glyph as genuine fan-in. Then add a regression that asserts the glyph *and* the
occupying routes' endpoints, so M15 fails.

**Resolution:** Retained compact integer route ownership and classify unrelated
collinear overlap separately from true common-endpoint shared geometry, with
exact crossing, shared-stub, merge, and overlap regressions.

#### M1. Fan counts are placed on one side of a node while counting both sides · **Status:** fixed

**Severity:** medium. **Consequence:** the count marker claims N routes converge where fewer are
drawn, and none at all is drawn where the rest arrive. The contract is explicit: "Fan counts describe
the routes actually presented at the visible endpoint."

`drawThreadSpatialRouteCounts` (`:1291-1333`) keys `targets` on `route.edge.To` alone. Forward routes
put their arrow at `toLeft` (`▶`, node's left border) and same-column / reverse / self routes at
`toRight` (`◀`, right border) (`:275-312`). The map therefore accumulates one count across both
sides and stores whichever `route.arrow` came last (`:1316-1321`).

**Reproduction:** `r→x`, `x→y`, `y→x` (a two-node cycle in the residue column with one ranked
prerequisite):

```
   │ • r                │───────▶│ • x                │◀2┐
                                 │ • y                │◀◆┘
```

`x` has two incoming routes — `r→x` arriving from the **left** (`▶` at `x=32`) and `y→x` from the
**right** (`◀` at `x=55`). The `2` is written at `x=56`, beside the single right-entry arrow; the
left side carries no count.

The same call also overwrites unrelated geometry: `putRouteCount` (`:894-905`) does
`cells.text, cells.connector, cells.crossing = label, 0, false` with no ownership check. In the
fixture above, `(56,4)` was route `x→y`'s corner (`┐`); the `2` erases it, so `x→y` now appears to
originate from the count marker.

**Evidence:** M12 (moving every target marker three rows away) and M11b (removing all markers) both
leave the whole `TestThreadSpatial` suite green — count placement is completely unpinned.

**Recommendation (bounded):** key `sources`/`targets` on `(taskID, arrowRune)` so each side gets its
own count, and make `putRouteCount` skip a cell that already carries another route's arrow, waypoint
or corner (place the digit one cell further out along the approach instead).

**Resolution:** Fan counts are keyed by endpoint role and entry side, so forward
and reverse target counts remain separate from source fan-out counts.

#### M2. The focus pointer and the fan-out count silently erase route endpoints · **Status:** fixed

**Severity:** medium. **Consequence:** the selected node — the one the user is reading — is exactly
where its fan-in count disappears; and a self-edge loses its only direction marker.

Two unconditional writes collide with cells that `drawThreadSpatialRouteCounts` and `putRouteArrow`
have already claimed, because nodes are drawn last (`:1242-1244`) and counts after adornments
(`:1241`):

1. `drawThreadSpatialNode` writes the focus caret at `placement.x-2, placement.y+2` (`:1446`). For a
   `▶` arrow, `target.point.x` is `route.arrow.x - 1 = to.x - 2` and `target.point.y = to.y + 2`
   (`:1316-1318`) — **the identical cell**.
2. `sources[…].point` is `route.segments[0].from` (`:1308`), the node's exit cell — which is also
   exactly where a self-edge's `◀` arrow lands (`toRight == fromPoint` when `From == To`).

**Reproduction A** — `a→c`, `b→c`, render with selection `c` then with selection `a`:

```
selected a:   │ • a … │──◆━━━2▶│ • c … │      ← count 2 present
selected c:   │ • a … │──◆━━━›▶│ [M3] … │     ← count replaced by ›
```

Cell `(c.x-2, c.y+2)` is `"2"` when `c` is not selected and `"›"` when it is. Both fan-in routes
still share the single `▶`, so the focused node reads as having one prerequisite instead of two.

**Reproduction B** — `a→a` plus `a→b`: the self-loop's `◀` at `(25,4)` renders as `"2"`. With two
self-edges the row is `│22┐` — no arrowhead anywhere, so the loop has no direction marker at all.

**Evidence:** verified per-cell on the rendered canvas in both cases; the production Thread render
shows the same collision at the focused gate (`… ║ [G1] completed ║────◆◆━━━━◆5▶│`, where the caret
sits on the incoming count cell).

**Recommendation (bounded):** move the focus caret to `placement.x-3` (the gap is ≥ 8 cells wide) or
suppress it when the cell is already an endpoint marker, and have `putRouteCount` refuse to overwrite
an arrow cell (see M1).

**Resolution:** Reserved distinct cells for the focus pointer, endpoint arrows,
side-aware fan-in counts, and fan-out counts; the mixed-direction regression
pins all six markers.

#### M3. Three of the change's own regression families do not pin what they name · **Status:** fixed

**Severity:** medium. **Consequence:** the checkpoint's central claims are unverified by the suite
that is supposed to protect them, so a later refactor can silently remove them.

- **Gate fixed-point repetition.** `TestThreadSpatialLayoutPlacesCascadedExternalGatesAtOneStableFixedPoint`
  (`thread_projection_test.go:911-968`) documents itself as forcing "the upstream placement to settle
  on a subsequent fixed-point pass". It does not: for that fixture the longest-path ranking already
  yields `upstream=0, downstream=1, member-c=2`, so `desired <= current` at `:397` and **no gate
  moves at all**. M8 (single pass) passes; M8b (deleting the entire body of
  `placeThreadSpatialExternalGates`) also passes. The `len(nodes)+1` bound and the repeated-pass loop
  at `:381` are unexercised. (The pre-existing `TestThreadSpatialLayoutPullsSourceGateBesideItsFirstDependent`
  does fail under M8c, so single-hop movement remains covered.)
- **Fan counts.** M11b/M12 above.
- **Shared-bundle grammar.** M15 above.

**Recommendation (bounded):** for the gate test, use a fixture where the initial ranking is wrong —
e.g. `g1→g2→g3→member` with `member` also reachable through a longer member chain, so `g1` must move
after `g2` has moved — and assert the number of passes indirectly by comparing against a
single-pass-limited expectation. For counts and bundles, assert on the specific cell coordinates and
their occupying routes rather than on "a cell somewhere containing the digit".

**Resolution:** Replaced incidental numeral and dead-helper assertions with
semantic cell metadata and production-path tests for fan counts, shared bundles,
endpoint arrows, and genuine multi-pass gate convergence.

#### L1. A non-contiguous set of waves in one layer is labelled as a contiguous range · **Status:** fixed

**Severity:** low. **Consequence:** the heading claims member generations occupy a presentation layer
that do not.

`labelThreadSpatialColumns` renders three or more distinct waves as `waves %d–%d` using only the
first and last (`:563-567`).

**Reproduction:** members `p` (wave 1), `q` (wave 3), `r` (wave 5) pushed into one column by external
gate chains of different lengths yields `layer 5 · waves 1–5` for a column whose actual wave set is
{1, 3, 5}. Waves 2 and 4 sit in layers 2/3 and 3/4.

**Evidence:** produced from a 15-node fixture (`g1→g2→g3→g4→p`, `h1→h2→q1→q2→q`, `r1→r2→r3→r4→r`);
column 4 = `[p q r]`, label `"layer 5 · waves 1–5"`. The two-wave case correctly prints `waves 1+3`.

**Recommendation (bounded):** print the range only when `waveNumbers` is contiguous; otherwise use a
compact list with an ellipsis (`waves 1,3,5` or `waves 1+3+2 more`).

**Resolution:** Layer headings enumerate the exact supplied wave set instead of
implying a contiguous range; non-contiguous 1, 3, and 5 evidence is pinned.

#### L2. The legend omits three glyphs the renderer actually emits · **Status:** fixed

**Severity:** low. **Consequence:** colour-independent readers cannot decode strokes that carry
meaning.

`threadSpatialRoleLegend` (`:1466-1470`) names `┄ long · ╳ crossing · ◆ shared`. The renderer also
emits `━` and `┃` for shared horizontal/vertical runs (`threadSpatialSharedConnectorGlyph:968-978` —
`◆` is only the both-axes case, and in practice `━` is by far the most common), `◇` for corridor
waypoints (`putRouteWaypoint:883`), and bare digits/`+` for fan counts (`putRouteCount:898-901`).
None of the three appears in either legend row. `━` is visually a *bolder* `─`, which without a
legend entry reads as emphasis rather than as "more than one route here".

**Recommendation (bounded):** extend the role legend to `┄ long · ╳ crossing · ━ ◆ shared · ◇ track ·
2 fan`, truncating as it already does.

**Resolution:** The compact route legend now explains bidirectional arrows, fan
counts, tracks, turns, crossings, shared strokes, overlaps, and invariant
conflicts.

#### L3. Two superseded route helpers survive as dead production code and one live test · **Status:** fixed

**Severity:** low. **Consequence:** a green test reports on geometry no user can see, and a geometry
constant is kept alive only by that dead path — exactly the "inconsistent old/new route helpers"
risk.

`drawThreadSpatialEdge:1335` is never called from production; its only caller is
`TestThreadSpatialConnectorGeometryUsesElbowsAndAccentFocus`
(`thread_projection_test.go:1049-1073`), which therefore asserts elbow and accent behaviour of the
*replaced* mid-column algorithm. `drawThreadSpatialLongEdge:1370` is reachable only from
`drawThreadSpatialEdge:1345`. `threadSpatialNodeStrideY:24` is referenced only from `:1381` (inside
that dead helper) and from `thread_projection_test.go:1422` / `:1452`, where
`TestThreadSpatialLongEdgeUsesNodeFreeTrack` and
`TestThreadSpatialLongEdgeKeepsDeepestRowTrackInsideCanvas` recompute the *old* fixed track formula
`threadSpatialNodeTop + row*threadSpatialNodeStrideY + threadSpatialNodeSlotHeight` instead of
reading the new per-route `trackY` map. Those two tests agree with the live geometry only while every
row carries at most one track (`gaps[row] == 1`, `index == 0`); the moment a row needs a second track
they will silently assert on an unrelated cell.

Separately, `renderThreadSpatialCanvas:1216` is production-dead and always renders the window as the
entire layout, so all four canvas-level test call sites bypass `threadSpatialRoutesForWindow`. And
the `if cell.text != "▶" && cell.text != "◀"` guard at `:814` is vestigial: in the live path all
segments are drawn before any adornment, so no connector is ever written after an arrow exists.

**Recommendation (bounded):** delete `drawThreadSpatialEdge`, `drawThreadSpatialLongEdge`,
`threadSpatialNodeStrideY` and the `:814` guard; re-point
`TestThreadSpatialConnectorGeometryUsesElbowsAndAccentFocus` at
`renderThreadSpatialCanvasWindow`; and have the two long-edge tests read `layout.routes[i].segments`
rather than recomputing geometry.

**Resolution:** Removed both superseded edge-drawing helpers and the stale fixed
row stride; long-route and arrow assertions now exercise materialized production
routes.

#### L4. The capacity guard is expressed in cells but the cost is 96 bytes per cell, and the pre-guard path is unbounded · **Status:** tracked by 6g8ezj5e51hg

**Severity:** low (a bounded prototype, and the guard does fail open correctly). **Consequence:** a
graph the guard *admits* still allocates ~74 MB and 11 ms per frame, and the navigation path has no
guard at all.

- `threadSpatialCell` (`:730-743`) is 96 bytes, of which `routeID`, `routeFrom` and `routeTo` are 48.
  `threadSpatialMaxCanvasCells = 750_000` (`:30`) therefore admits ~69 MB of canvas. Measured at
  742,280 cells (504 nodes, inside the node limit): **73.9 MB / 11.1 ms per `renderThreadSpatial`**,
  re-allocated on every keystroke and resize.
- `buildThreadSpatialLayout` runs at `:1043`, before the guard at `:1051` — by necessity, since the
  guard needs `layout.width`/`height`. Measured: 2.3 ms (512 nodes) → 31.8 ms (8,192 nodes).
- `threadSpatialMove:588` and `threadSpatialSelectedTaskID:655` build a full layout with **no**
  capacity check: 3.0 ms / 22.1 ms / 244.7 ms at 512 / 2,048 / 8,192 nodes. `detail.go` compounds
  this — `renderDetail:857` builds twice per frame (once in `detailSelectionKey:862`, once inside
  `renderThreadSpatial`), and `moveDetailSelectionDirection:901` builds twice more per keypress.
- The node (512) and edge (2,048) bands are effectively unreachable: a 512-node dense layout is
  already 1,255,508 cells, so the canvas band always fires first.

**Recommendation (bounded):** replace the three per-cell strings with a single `int32` index into
`layout.routes` (cell drops to ~40 bytes, canvas budget to ~29 MB) and lower
`threadSpatialMaxCanvasCells` accordingly; memoise the layout on `threadDetail` keyed by projection
identity so a frame and a keypress each build once.

**Resolution:** This patch replaces three per-cell route strings with one int32
identity, caps cells at 500,000, and pins cell storage below 24 MiB. Early input
preflight and coherent per-refresh layout reuse remain in the linked follow-up.

### 7. Contracts that survived attack

Stated as corroborated, with the evidence that would have broken them.

- **Adapter neutrality.** No route-derived state reached `core`/`domain`/`wire`/`store`;
  `internal/core/thread_graph.go` is byte-identical to `main` and only two Go files changed. The
  renderer never mutates the projection, infers readiness, or redefines graph health. No hidden
  second graph model exists: columns are longest-path over the supplied node/edge set and are
  labelled `layer`, distinct from core `Waves` (M9 pins this).
- **Determinism under equivalent evidence.** 400 randomised permutations of `Nodes`, `Edges` **and**
  `Waves` (including a duplicate edge and a cascaded gate pair) produced `reflect.DeepEqual` layouts
  and byte-identical `renderThreadSpatial` output. Lane and track values are computed from
  seed-order indices, so the two `range` loops over `uses` (`:218`) and `tracks` (`:252`) are
  order-independent despite iterating maps. Selection changes presentation only — geometry and
  aliases are computed before `selected` is consulted.
- **Node-free routes.** Verified structurally, not just by the shipped test: `trackY ∈ [rowY[r]+5,
  rowY[r+1]-1]` and lane `x ∈ [columnX[b]+23, columnX[b+1]-2]` by construction, and I re-checked
  every cell of every route against every node box on the 20-node/76-edge fixture, the 4×4 dense
  fixture, and both production Threads: zero intrusions. M2 fails as intended.
- **Lane and track uniqueness.** Distinct seeds always receive distinct lanes within a boundary and
  distinct track rows within a row band; corridors of different routes provably never collide.
  M1b fails as intended.
- **Viewport rule, integrated.** With a 304×8 layout panned to `panX=124`, the `n0→n9` highway whose
  corridor physically crosses the visible window was omitted along with five other non-incident
  routes, while all four routes incident to a visible node were drawn. M6b fails as intended.
- **Boundary aliases.** Three simultaneously clipped fan-out routes each named a distinct offscreen
  endpoint (`…▶[M5]`, `…▶[M6]`, `…▶[M7]`) on distinct rows; clipping on left, right, top and bottom
  boundaries all produced correct `[M2]…` / `…▶[M4]` labels. `putThreadSpatialBoundaryLabel:1167-1179`
  clamps to `panX + width - labelWidth`, so a label can never be cut by `ansi.Cut`. M7 fails as
  intended.
- **Gate placement soundness.** I could not construct a supported case that settles invalidly or
  needs more than `len(nodes)+1` passes. Each move strictly increases a gate's column, gates never
  move left, and `desired <= latestIncoming` (`:397`) keeps every gate strictly inside the open
  interval, so no supplied edge is reversed. Reverse-ordered chains, forks, joins, a gate with
  multiple outgoing intervals, and cyclic residue all converged. The *implementation* is right; only
  its regression test is inert (M3).
- **Capacity fail-open.** The narrow explanation precedes the capacity check by design (`:1045-1050`)
  so a small terminal always retreats to the wave reader; the fallback names the issue and the node
  and edge counts. Line counts never exceeded `height` and display width never exceeded `width` at
  0×0, 1×1, 40×8, 60×12, 62×16, 70×24, 80×24, 90×24, 120×30, 140×36 and 200×60. M10 fails as
  intended.
- **Terminal safety.** CJK labels at 80×24 across three pan positions produced no overflow;
  `putText`'s continuation cells (`:798-800`) keep `ansi.Cut` aligned, and zero-width runes are
  appended to the preceding cell rather than consuming one. `terminalText` scrubs control runes from
  every label, description and alias that reaches the canvas. Degenerate projections (empty, missing
  endpoints, phantom wave members, empty `TaskID`) neither panicked nor produced a route.
- **Arrowheads.** Across the 4-layer dense fixture (48 routes) and both production Threads, no
  forward or reverse arrowhead was lost. The only loss is the self-edge case in M2.
- **Selected emphasis and semantic colour.** A later non-selected route cannot clear `cell.accent`
  (`:845` only sets `color` when it is `ColorNone`, and `renderLine:994-998` gives accent priority),
  and node boxes keep `theme.Status(node.Status).Color` under selection (`:1441-1445`). Yellow stays
  a direction/focus signal. M4 fails as intended.
- **Checkpoint validation claims.** Every command listed in the task's Implementation checkpoint
  (`go test -race ./...`, `go vet`, tidiness, doc drift, `golangci-lint`, `git diff --check`,
  `tskflwctl lint`) reproduced green in the sandbox, and the "20-node/76-edge layered graph" fixture
  is real (5 layers × 4 rows, 64 + 12 edges).

### 8. Residual uncertainty

- **Not executable here.** No PTY was available, so I never drove the real Bubble Tea event loop;
  everything above is `renderThreadSpatial`/canvas-level evidence at fixed sizes plus cell metadata.
  I did not verify true terminal emulator behaviour for `╳`, `◆`, `━`, `┃`, `┄`, `◇` or `›` — glyph
  availability and width in the user's font are unverified, and `━`-vs-`─` legibility in particular
  is a judgement I could only make from stripped text.
- **Colour rendering** was exercised through `testStyles`, not a real palette; I checked that
  accent/bold/colour attributes are *set* correctly and that stripped output stays meaningful, but
  not how magenta-on-theme actually looks.
- **Severity of H2 relative to H1.** The false-bundle glyph is a weaker lie than the severed corner —
  it says "more than one route here" where the routes are genuinely distinct but unrelated. I rate it
  high because it shares H1's root cause and because no test constrains it at all; a maintainer may
  reasonably triage it as medium once H1 is fixed, since fixing H1 changes which cells reach the
  `shared` branch.
- **The 8,192-node measurement is synthetic.** I did not establish that a Thread that large is
  reachable through `thread new` / bulk-link; the 512-node figure (2.3 ms build, 3.0 ms move) is the
  one that matters for the stated prototype band, and the 74 MB/frame figure at 504 nodes is inside
  it.
- **Not attempted.** Live reload under concurrent store mutation, and Atlas/immersion interactions
  outside `threadDetail`; both are outside this change's diff. I did not evaluate the planned
  responsive layout or one-hop focus behaviour — per the brief I treated them as unimplemented, and I
  confirmed no code in this change provides them.

### 9. Transfer attestation

The helper's own output, from `isolated-review-workspace.sh verify --sandbox "$SANDBOX"` run on the
finished report immediately before the guarded copy-back:

```
sandbox_path=/private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.mR7nsP
git_dir=/private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.mR7nsP/.git
baseline_commit=54e6e0a9276c59b707d716ddd8f7de8b0ae5c52f
source_blob=008ccacd90fb7e5e054f30e7667426f36da591ed
source_fingerprint=5af9084d9aed8205e5ef519c5d2beeba40f807b2
deliverable=planning/audits/6g8e0pyczpns-2026-09-09-dense-thread-graph-routes-implementation-claude.md
deliverable_changed=true
transfer=pending
```

`verify` passed fail-closed: HEAD still at the baseline commit, no staged changes, no untracked
files, exactly one unstaged delta (this audit), `git diff --check` clean, and the source deliverable
still hashing to the recorded `source_blob`. The transfer itself necessarily runs *after* this file
is finalised, so its `transfer=succeeded` / `source_deliverable=` / `sandbox_retained=` lines cannot
appear inside the transferred bytes; they are reproduced verbatim in the handoff message. The
workspace is retained at the `sandbox_path` above until the implementation owner confirms receipt.
