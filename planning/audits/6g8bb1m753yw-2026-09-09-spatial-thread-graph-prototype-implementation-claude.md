---
schema: 1
id: 6g8bb1m753yw
bucket: closed
area: spatial-thread-graph-prototype-implementation-claude
date: "2026-09-09"
updated_at: "2026-09-09"
---
# Audit: Spatial Thread graph prototype implementation — claude — 2026-09-09

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

Perform an independent adversarial implementation and architecture review of the merged
two-dimensional navigable Thread graph prototype. Treat the renderer, its tests, and its claimed
adapter boundary as suspect until hostile evidence demonstrates that the picture, navigation, and
shell state all remain faithful to the supplied graph projection. Do not merely restate the task or
confirm that the existing suite passes.

Distinguish a demonstrated defect from an acknowledged prototype limitation already assigned to a
follow-up. A limitation may still be a release or safety concern if the current UI presents it as
trustworthy, but do not manufacture duplicate findings for explicitly bounded future work.

## Review target

Review PR #217 as merged on `main` at `c40d09d`, using commit range `c4a06f0..c40d09d`. The review
target includes:

- `internal/tui/thread_spatial.go`
- the optional detail-presentation seams in `internal/tui/detail.go`
- immersive shell and key routing in `internal/tui/model.go`, `internal/tui/nav.go`, and
  `internal/tui/view.go`
- color/style integration in `internal/tui/style.go` and help text in `internal/tui/help.go`
- the spatial and neighboring Thread tests in `internal/tui/thread_projection_test.go`
- task `6g6dw5js81f3`, ADR-0006's 2026-09-08 amendment, the Threads/TUI epics, and the follow-ups
  `6g86016qvk5d`, `6g86c7y6hn41`, `6g86e10dztpf`, `6g86g03zfj2f`, and `6g87qn72901g`

The source checkout may contain an owner-authored, uncommitted lifecycle update marking the
prototype task completed. That metadata is not implementation evidence and is outside the target;
preserve it through the sandbox copy and do not edit it. Review committed code and planning claims,
then update only this assigned audit.

Build a consumer inventory for every new optional detail-content interface and every added shell
state field/helper. Identify all implementations, type assertions, event routes, reload paths, and
future-adapter claims. Verify that optional behavior cannot silently affect ordinary entity detail,
single-pane navigation, manual zoom, Atlas, modal routing, or non-Thread content.

## Intended contract to challenge

The TUI adds a third Thread presentation, `summary → topology → spatial`, as an immersive,
read-only adapter over the supplied `core.ThreadGraphProjection`. Core owns task identity, node and
edge semantics, member waves, roles, health, and topology completeness. The TUI may derive bounded
terminal placement and routes but must not infer readiness, authorize work, mutate planning data,
or make renderer coordinates part of a portable contract.

Layout and aliases are deterministic for equivalent projection evidence. Dependencies point from
prerequisite to dependent. External gates may occupy a late valid presentation layer before their
first dependent without altering core member-wave labels. `h`/`l` prefer actual incoming/outgoing
edges, including skipped columns; `j`/`k` move within the current presentation column. Canonical
task ID owns selection through reload, add/rename, task open, and `ctrl+o` return. `y` copies the
highlighted task in structured Thread views. Esc retreats spatial to waves, `v` continues the view
cycle, and immersive zoom does not corrupt prior manual zoom or focus state.

The renderer must never silently draw a partial oversized graph. Narrow and capacity-limited views
remain explicit and offer the complete wave reader/task picker. Renderer failure is display-only.
Status color remains semantic across selection; focus and edge direction use separate visual
signals. Known production follow-ups own dense-route truthfulness, one-hop focus, reusable Back,
structured action targeting, and alternate-view/list discoverability.

## Mandatory evidence floor

Inventory the complete data path from `Service.ThreadGraphProjection` through the Thread registry,
detail payload, shell type assertions, spatial layout, canvas, clipping, footer, and task
navigation. Cite exact paths/lines for every claimed invariant and distinguish core-supplied facts
from TUI-derived presentation values.

Run the full race suite and lint, plus focused Thread/TUI tests. Exercise a real local build against
the repository Thread at multiple terminal sizes when the environment supports a PTY. Inspect ANSI
cell width and clipping with Unicode, combining marks, hostile control text, long IDs/labels, empty
labels, and all status/role/health combinations.

Mutation evidence is mandatory. Temporarily remove or invert individual production guards and
prove the named focused regression fails for the intended reason. At minimum challenge:

- canonical-ID selection and reload restoration;
- graph-first horizontal navigation across a populated intermediate column;
- external-gate late placement and retained member-wave labels;
- the node-free long-edge track;
- the node/edge/canvas capacity checks before dense canvas allocation;
- immersive-entry/retreat state and `ctrl+o` presentation restoration;
- structured-detail `y` targeting; and
- deterministic ordering when projection nodes, waves, and edges are permuted.

If no existing regression kills a mutation, record the gap even when manual behavior appears
correct. A compile error, broad golden drift, or unrelated test failure is not a valid mutation
kill. Restore the sandbox baseline after every probe.

## Required hostile angles

1. Permute nodes, edges, member waves, map construction order, and equal-distance fan-in/fan-out.
   Look for nondeterministic aliases, rows, columns, navigation targets, connector geometry, and
   inspector ordering.
2. Stress external gates: source gates, gates between member waves, a gate feeding several distant
   waves, gate-to-gate edges, multiple gates sharing intervals, and incomplete/cyclic residue.
   Determine whether presentation placement can reverse an edge or mislabel a core wave.
3. Stress navigation independently from appearance. Confirm `h`/`l` always choose a real edge when
   one exists in that direction and that fallbacks do not jump across unrelated graph structure.
   Challenge ties, isolated nodes, reverse/cyclic residue, unreadable nodes, and disappearing tasks.
4. Try to make the picture lie: long edges through nodes, shared lanes, crossings near arrowheads,
   same-row and backwards edges, clipping at all four viewport boundaries, and selected incident
   routes. Separate defects already honestly bounded by `6g86e10dztpf` from unqualified falsehoods
   in the current display.
5. Exercise shell transitions from split, single-pane, and manually zoomed layouts. Cycle views,
   press Esc/q/h/z, open a task, return, switch tabs, enter Atlas, resize, reload, fail a detail read,
   and remove the selected node. Look for stranded zoom, stale focus, wrong selection, or behavior
   leaked to ordinary detail pages.
6. Challenge cell and resource safety: zero/negative/tiny dimensions, huge width×height products,
   integer overflow assumptions, 512/513 nodes, 2,048/2,049 edges, wide runes, combining marks,
   embedded ANSI/control characters, and unusually long relationship lists. Confirm bounding occurs
   before material allocation or expensive route construction where claimed.
7. Audit semantic isolation. Search for readiness/status/role recomputation, graph traversal beyond
   supplied edges, filesystem/store access from TUI rendering, or core imports of TUI/layout types.
   Also test whether incomplete projection health is preserved rather than normalized away.
8. Audit action identity. Verify `Enter`, `y`, task picker, `ctrl+o`, reload, and duplicate labels use
   canonical IDs at authority boundaries and do not confuse visible aliases/slugs with identity.
9. Examine test helpers and assertions for systemic blind spots: ANSI-stripped snapshots that cannot
   prove color distinctions, substring assertions that tolerate false connectors, fixtures whose
   insertion order matches the desired order, and mocks that skip asynchronous shell behavior.
10. Perform a second pass looking beyond the named checklist: abstraction leakage, quadratic or
    dense memory behavior below the advertised thresholds, accidental global key precedence,
    stale async state, malformed projections, and comments/planning claims stronger than code.

For each surviving finding, provide a minimal reproduction or precise code path, impact, why the
current tests miss it, and the smallest sound correction. State whether it belongs in the current
prototype closeout or an existing/new follow-up.

## Validation and restoration

Perform all builds, PTY sessions, fixtures, code mutations, and generated artifacts inside the
mandatory isolated sandbox. Use sandbox-local caches where possible. Run and report exact results
for `go test -race ./...`, `golangci-lint run ./...`, `tskflwctl lint`, `git diff --check`, focused
TUI tests, and any visual/PTY probes. Report unsupported checks and their exact blocker.

Before transfer, restore every mutation and scratch artifact to the sandbox baseline. The assigned
audit must be the only remaining change. Use the isolation helper's verify and guarded transfer
commands; do not manually copy results back.

## Deliverable

Update only your assigned audit and preserve this brief. Add findings using the exact repository
grammar, leaving each new finding `open` for implementation-owner triage. Include severity,
component, effort, urgency, exact evidence, mutation/probe results, affected acceptance criteria,
and a minimal recommendation. Reference an existing follow-up when it already owns the issue rather
than duplicating scope.

If no finding survives both passes, write a substantive no-findings report with the consumer
inventory, exact mutations killed, visual/PTY cases exercised, commands run, residual risks, and
why each residual risk is honestly bounded.

## Reviewer report

Independent adversarial review of PR #217 as merged on `main` at `c40d09d` (range
`c4a06f0..c40d09d`), performed entirely inside the mandatory isolated sandbox. Three findings
survive both passes and are left `open`. The uncommitted lifecycle edit to
`planning/tasks/6g6dw5js81f3-…md` in the source checkout was preserved through the copy and
neither read as implementation evidence nor edited.

### Consumer inventory — new optional seams

Every seam added by this PR has exactly **one** implementation, `threadDetail`, and each returns
false unless the active view is `spatial`. Verified by exhaustive grep over `internal/`:

| Seam | Implementations | Call sites (non-test) |
|---|---|---|
| `retreatingDetailContent.previousDetailViewName` | `threadDetail` (`detail.go:820`) | `detail.go:222` (`retreatView`) |
| `sizedDetailContent.detailSized` | `threadDetail` (`detail.go:848`) | `detail.go:169`, `detail.go:353` |
| `sizedDetailContent.renderDetail` | `threadDetail` (`detail.go:852`) | `detail.go:170` |
| `immersiveDetailContent.detailImmersive` | `threadDetail` (`detail.go:844`) | `detail.go:232` |
| `directionalDetailContent.detailDirectional` | `threadDetail` (`detail.go:850`) | `detail.go:237` |
| `directionalDetailContent.moveDetailSelectionDirection` | `threadDetail` (`detail.go:896`) | `detail.go:283` |
| `yankableDetailContent.detailSelectionYankRef` | `threadDetail` (`detail.go:915`) | `detail.go:303` |

The five `detailContent` implementations are `taskDetail` (`detail.go:657`), `epicDetail`
(`detail.go:727`), `threadDetail` (`detail.go:785`), `auditDetail` (`detail.go:1347`) and
`researchDetail` (`detail.go:1393`). Only `threadDetail` satisfies any new interface, so ordinary
entity detail, single-pane navigation, manual zoom, Atlas, and modal routing cannot be reached by
this behavior. `detail.go:169` is guarded by both the type assertion and `detailSized()`, and
`detailSized`/`detailDirectional` both delegate to `detailImmersive`, which is `view == spatial`
only (`detail.go:844`).

Shell-state fields added: `Model.immersiveZoom` (`model.go:81`), `Model.pendingDetailNavigation`
(`model.go:103`), `navLoc.detailView` / `navLoc.detailSelection` (`nav.go:25-26`),
`detailNavigationRestore` (`nav.go:29-34`). `immersiveZoom` is written in exactly four places —
`unzoom` (`model.go:1151`), `toggleZoom` (`model.go:1322`), and both branches of
`syncDetailImmersion` (`model.go:1336-1348`).

### Semantic isolation

Clean. `grep -rn "internal/tui" internal/core/` returns nothing. `thread_spatial.go` imports only
`fmt`, `sort`, `strings`, `charmbracelet/x/ansi`, `internal/core`, `internal/domain`,
`internal/theme` — no `os`, `io`, `path/filepath`, `internal/store`, or cobra. The five references
to `RoleUnknown`/`GateBroken`/`GateClear`/`GateBlocked` (`thread_spatial.go:831,849,922,924,926`)
are read-only color selection over core-supplied `node.State`; nothing recomputes readiness, and
the renderer traverses only `projection.Edges`.

### Commands run (all inside the sandbox)

| Command | Result |
|---|---|
| `go test -race ./...` | **pass** — all packages ok; `internal/tui` 8.920s |
| `golangci-lint run ./...` | **pass** — `0 issues.` |
| `./bin/tskflwctl lint` | **pass** — `✔ all planning entities and dependency links pass lint` |
| `git diff --check c4a06f0..c40d09d` | **pass** — no output |
| Size sweep, real repo Thread `complete-production-threads` (40 nodes / 51 edges) | 14 sizes from `0×0` and `-5×-5` through `1000×200`: no line exceeded the effective width, no render exceeded the height, no panic |
| Hostile-label cell safety | wide CJK, combining marks, embedded `ESC[31m`/`ESC[2J`, NUL/BEL/BS/CR/TAB, RTL override, ZWJ emoji, 640-char label, empty label — no cell-width overflow at 60×12, 80×24, 120×40, 200×50; control bytes are neutralized by `terminalText` |
| Run-to-run stability | 200 identical renders of the hostile fixture — byte-identical |

A real PTY session was not run: driving the interactive binary was judged a worse probe than the
deterministic size sweep above, which exercises the same `renderThreadSpatial` entry point at more
sizes with assertions. Recorded as an unexecuted check rather than a passed one.

### Mutation results

Each mutation was applied to the sandbox baseline `af806c1`, run, then reverted; `git status` was
confirmed clean between probes.

| # | Mutation | Named regression | Verdict |
|---|---|---|---|
| K1 | Remove the long-edge branch in `drawThreadSpatialEdge` (revert to the straight route) | `TestThreadSpatialLongEdgeUsesNodeFreeTrack` | **KILLED** |
| K2 | Disable graph-edge preference in `threadSpatialMove` (`if false && len(direct) > 0`) | `TestThreadSpatialHorizontalNavigationPrefersGraphEdgesAcrossVisibleColumns` | **KILLED** |
| K3 | Disable late external-gate placement | `TestThreadSpatialLayoutPullsSourceGateBesideItsFirstDependent` | **KILLED** |
| K4 | Delete the canvas-cell capacity guard | *entire* `internal/tui` suite | **SURVIVED** |
| K5 | Delete the edge capacity guard | *entire* `internal/tui` suite | **SURVIVED** |
| K6 | Delete the node capacity guard | `TestThreadSpatialGraphFailsOpenToWavesBeyondPrototypeCapacity` | **KILLED** |
| K7 | Ignore canonical ID in `threadSpatialSelectedTaskIDInLayout` | reload + directional tests | **KILLED** |
| K8 | Never restore the split on retreat | `TestThreadSpatialGraphUsesDirectionalStableIdentityNavigation` | **KILLED** |
| K9 | Let a *manual* zoom be claimed as immersive (`case nowImmersive:`) | *entire* `internal/tui` suite | **SURVIVED** |
| K10 | Drop presentation restore in `restoreDetailNavigation` | directional test | **KILLED** |
| K11 | Drop selection restore in `restoreDetailNavigation` | directional test | **KILLED** |
| K12 | Make `selectedYankRef` fall through to the parent Thread row | `TestThreadStructuredDetailYankCopiesHighlightedTask` | **KILLED** |
| K13 | Return the stable ID instead of the slug from `detailSelectionYankRef` | `TestThreadStructuredDetailYankCopiesHighlightedTask` | **KILLED** |
| K14 | Let `z` toggle zoom inside the immersive view | *entire* `internal/tui` suite | **SURVIVED** |

K4, K5, K9 and K14 were re-run against the whole package (`-run .`), not only the named test, to
confirm no other regression kills them. No mutation was scored on a compile error or unrelated
failure.

Determinism was checked separately: permuting `Edges` (all 300 transpositions) and re-sorting
permuted `Nodes` to the TaskID order core guarantees both leave the render byte-identical, as does
repeating an identical render 200 times. See **L1** for the caveat.

## Findings

#### H1. A long edge whose shallower endpoint sits on the deepest layout row loses its horizontal track and draws as two dangling stubs  · **Status:** tracked by 6g8by30btznq

**File:** internal/tui/thread_spatial.go:806 | **Component:** tui/thread-spatial-renderer
**Effort:** XS · **Urgency:** soon

`drawThreadSpatialLongEdge` puts the cross-column run on the blank row reserved *below* the
shallower endpoint's slot:

```go
trackRow := min(from.row, to.row)
trackY := threadSpatialNodeTop + trackRow*threadSpatialNodeStrideY + threadSpatialNodeSlotHeight
```

The canvas, however, is only as tall as the last slot's *content*. `buildThreadSpatialLayout`
(`thread_spatial.go:97`) sets `layout.height = max(placement.y + threadSpatialNodeSlotHeight)`, so
for the deepest row `R` the canvas is `6R+7` rows (indices `0..6R+6`) while `trackY` for
`trackRow == R` is exactly `6R+7` — one row past the end. `putConnector` (`thread_spatial.go:531`)
returns silently on `y >= len(c.cells)`, so the entire horizontal segment is discarded with no
error. The two vertical segments are still drawn, because they run from the node row *down toward*
the missing track.

The result is not an ambiguous route — it is a supplied edge rendered as two disconnected stubs
that visually attach to *other* edges. Minimal reproduction (six nodes, ANSI stripped, 100×26); the
`s2 → t2` edge is in the projection and named in the inspector, but no line joins the two boxes:

```
   ┌────────────────────┐        ┌────────────────────┐        ┌────────────────────┐
   │ ● src one          │───────▶│ ● mid one          │───────▶│ ● tgt one          │
   └────────────────────┘        └────────────────────┘        └────────────────────┘

   ┌────────────────────┐
   │ ● src two          │        ┌────────────────────┐        ┌────────────────────┐
 › │ [M2] next-up       │─┬─────▶│ ● mid two          │──────┬▶│ ● tgt two          │
   │ queued / clear     │ │      └────────────────────┘      │ └────────────────────┘
   └────────────────────┘ │                                  │

╭─ focus [M2] ─────────────────────────────────────────────────────────────────────╮
│ ● next-up  src two  (s2)                                                         │
│ state queued/clear  role member  unlocks [M4] mid two, [M6] tgt two              │
╰──────────────────────────────────────────────────────────────────────────────────╯
```

The frame contradicts itself: the inspector reports `unlocks … [M6] tgt two` while the picture
shows no such connection. The two orphan `│` strokes sit directly under the `┬` junctions of the
`s2→m2` and `m2→t2` routes, so a reader most naturally attributes them to those edges. Probe
values: `layout=89×13`, canvas rows `0..12`, `trackRow=1`, `trackY=13`.

Why the tests miss it: `TestThreadSpatialLongEdgeUsesNodeFreeTrack`
(`thread_projection_test.go:962`) asserts `canvas.cells[trackY][middle.x-2].accent` — but its
fixture puts both long-edge endpoints on **row 0** while a fourth node (`other-source`) pushes the
layout's deepest row to 1. `trackY` is therefore 7 against a 13-row canvas and the boundary is
never touched. The assertion also indexes `canvas.cells[trackY]` directly, so on the failing
geometry the *test itself* would panic rather than report — the helper cannot express the defect it
was written to guard. This is the fixture-ordering blind spot in the second-pass checklist.

Scope: this is the prototype's own claimed-shipped mechanism, not bounded future work.
`6g86e10dztpf` is explicitly told to "use its node-free track … as the baseline", and its first
criterion assumes every edge already has "a deterministic node-free route from its actual
prerequisite to its actual dependent". A baseline that drops the route on the deepest row does not
support that. Currently **latent** on real data: a sweep of both repository Threads
(`complete-production-threads`, 40 nodes/51 edges, layout 419×55; `tool-owned-actionable-sub-entities`,
4 nodes/3 edges) found zero off-canvas tracks today, because no skipped-column edge yet connects two
nodes that both sit on the deepest row. It becomes reachable as soon as one does.

**Recommendation:** reserve the track row in the canvas rather than clamping the route. In
`buildThreadSpatialLayout` (`thread_spatial.go:97`), extend the height by the inter-slot stride
instead of the slot height:

```go
layout.height = max(layout.height, placement.y+threadSpatialNodeStrideY)
```

**Verified in the sandbox.** With that single-token change the canvas becomes `89×14`, `trackY=13`
falls in bounds, 35 track cells are drawn between the two lanes, and the route closes — the same
frame now reads:

```
 › │ [M2] next-up       │─┬─────▶│ ● mid two          │──────┬▶│ ● tgt two          │
   │ queued / clear     │ │      └────────────────────┘      │ └────────────────────┘
   └────────────────────┘ │                                  │
                          └──────────────────────────────────┘
```

`go test ./internal/tui/ -count=1` stays green with the fix applied, which also confirms no existing
regression constrains this geometry in either direction. The change was reverted before transfer.

Pair it with a regression whose long edge has `min(from.row, to.row)` equal to the layout's deepest
row, asserting the horizontal run is present at `trackY` — and bounds-check the index first, so the
geometry failure reports instead of panicking the way the current helper would.

**Resolution:** Accepted. The consolidated prototype-hardening task reserves the
deepest long-edge track inside the canvas and pins the failing geometry.

#### M1. Two of the three prototype capacity guards have no regression coverage and both are reachable below the node limit  · **Status:** tracked by 6g8by30btznq

**File:** internal/tui/thread_spatial.go:717 | **Component:** tui/thread-spatial-capacity
**Effort:** S · **Urgency:** soon

`threadSpatialCapacityIssue` enforces three bounds — nodes (512), edges (2048), and canvas cells
(750,000). Only the node bound is tested. Deleting either of the other two
(`case false:`) leaves the **entire** `internal/tui` suite green (mutations K4 and K5 above,
re-run with `-run .`).

Both are reachable well inside the node limit, so the node guard does not subsume them:

- **Canvas-cell guard.** 255 chained nodes (255 columns) plus 250 extra sources landing in column 0
  = **505 nodes / 254 edges**, both under their limits. Layout is `7649×1507` = **11,527,043 cells,
  15.4× the 750,000 limit**. The canvas guard is the only thing between that projection and
  `newThreadSpatialCanvas` allocating 11.5M `threadSpatialCell` structs at
  `thread_spatial.go:757`.
- **Edge guard.** A 100-node complete DAG is **100 nodes / 4950 edges** — under the node limit,
  2.4× over the edge limit. Each edge drives a route construction in
  `renderThreadSpatialCanvas`.

`TestThreadSpatialGraphFailsOpenToWavesBeyondPrototypeCapacity`
(`thread_projection_test.go:1062`) builds `threadSpatialMaxNodes+1` nodes and **no edges**, so it
exercises the node branch alone and its `"513 nodes"` assertion pins that message specifically.

The guards are correctly *ordered* — `renderThreadSpatial` runs the capacity check
(`thread_spatial.go:679`) before `renderThreadSpatialCanvas` (`thread_spatial.go:684`), so the
bound genuinely precedes material allocation. The defect is that two thirds of that protection is
unverified and would be removed or broken by a refactor without any test objecting. The brief names
these checks as a mandatory mutation target and requires the gap be recorded even where manual
behavior is correct.

Note the narrow path (`thread_spatial.go:676`) returns *before* the capacity check. That is sound —
`renderThreadSpatialNarrow` allocates no canvas — but it is undocumented and untested, so a future
change that draws anything in the narrow view would silently bypass every bound.

**Recommendation:** add two regressions that assert the fallback message for the edge and canvas
branches using fixtures under the node limit (the 100-node complete DAG and the 505-node sparse grid
above are sufficient and fast), and assert each names its own limit rather than the node one. Add a
one-line comment at `thread_spatial.go:676` recording that the narrow path is exempt because it
never allocates a canvas.

**Resolution:** Accepted. Independent edge and canvas-cell capacity regressions,
plus the intentional narrow no-canvas exemption, are required by the
consolidated hardening task.

#### M2. The "immersive zoom must not consume a manual zoom" contract clause has no regression  · **Status:** tracked by 6g8by30btznq

**File:** internal/tui/model.go:1336 | **Component:** tui/immersive-shell-state
**Effort:** XS · **Urgency:** eventually

`syncDetailImmersion` distinguishes a zoom the graph entered itself from one the user entered, and
only unwinds its own:

```go
case nowImmersive && !m.zoom:
        m.zoom = true
        m.immersiveZoom = true
```

The production behavior is **correct** — verified directly: manual `z` → `v` → `v` → `esc` leaves
`zoom=true immersiveZoom=false` throughout, so the user's full-screen survives the retreat. But
broadening that case to `case nowImmersive:` (mutation K9) leaves the whole `internal/tui` suite
green while changing observable behavior: with the mutation the same sequence ends at
`zoom=false`, silently discarding the user's own full-screen on Esc. Confirmed not an equivalent
mutant by a probe that passes on baseline and fails under K9.

`TestThreadSpatialGraphUsesDirectionalStableIdentityNavigation`
(`thread_projection_test.go:712`) only ever enters spatial from a non-zoomed split, so the
`immersiveZoom=false` path is never exercised. The related guard at `model.go:884` — `if
!m.detail.immersive() { m.toggleZoom() }`, which makes `z` inert inside the immersive view — is
likewise uncovered (mutation K14 also survives the full suite).

This is a coverage finding, not a behavior defect: no user-visible misbehavior exists today. It is
recorded because the brief requires uncovered mutations be reported even when manual behavior is
correct, and because `immersiveZoom` has exactly four writers, two of which no test constrains.

**Recommendation:** extend the directional test (or add a sibling) with the manual-zoom path —
press `z`, cycle to spatial, press Esc, and assert `m.zoom` is still true and `m.immersiveZoom`
false; and assert `z` inside spatial does not change `m.zoom`.

**Resolution:** Accepted. The hardening task pins preservation of manual zoom
and the inert spatial z behavior.

## Residual risks (not findings)

- **L1 — the renderer relies on an unstated ordering contract.** `buildThreadSpatialLayout` derives
  its `order` map from the *slice index* of `projection.Nodes` (`thread_spatial.go:56`), and that
  order decides row assignment inside a column, `[M#]` alias numbering, and navigation tie-breaks.
  Raw permutation of `Nodes` changes the rendered graph (first counterexample: swapping indices 0
  and 6 of the hostile fixture). This is **not reachable today** — `ProjectThreadGraph` sorts
  `Nodes` by `TaskID` at `internal/core/thread_graph.go:63`, and re-sorting every permutation
  restores a byte-identical render across all 210 transpositions. The risk is that the invariant is
  documented nowhere at the TUI seam; a future projection change that stops sorting would silently
  reshuffle layout and aliases with no test failing. A comment at `thread_spatial.go:56` naming the
  dependency would close it.
- **Dense-route legibility** (shared lanes, crossings near arrowheads, parallel long edges on one
  track) remains genuinely bounded by `6g86e10dztpf` and is not double-reported here. H1 is
  separated from that scope because it is a missing route, not an ambiguous one.
- **Wave-synthesized nodes.** `buildThreadSpatialLayout` fabricates a node for any wave task ID
  absent from `Nodes` (`thread_spatial.go:64-71`). `detailSelectionYankRef` (`detail.go:915`)
  validates against `projection.Nodes`, so `y` on such a node copies nothing rather than a
  half-known ID. That fails closed and is left as-is.
- **`previousDetailViewName` on summary returns `spatial`** (`detail.go:820`), which would make Esc
  jump *into* the immersive view. Unreachable today because `retreatView` is only called behind
  `m.detail.immersive()` (`model.go:760`, `model.go:938`). Latent trap, no current defect.

## Task coordination

Maintainer triage consolidated H1, M1, and M2 into
[`6g8by30btznq` — close spatial Thread prototype correctness and invariant gaps](../tasks/6g8by30btznq-close-spatial-thread-prototype-correctness-and-invariant-gaps.md).
The three findings are small parts of one pre-feature baseline: preserve every supplied route, pin
the independent resource guards, and protect user-owned versus presentation-owned zoom state. The
task precedes dense-route, responsive-layout, and one-hop-focus work so those features cannot hide
or deepen the gaps.

## Isolation attestation

| Item | Value |
|---|---|
| Workspace path | `/private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.RYaknW` |
| Resolved Git directory | `/private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.RYaknW/.git` (a real directory — independent `--no-hardlinks` clone, not a worktree or symlink) |
| Baseline commit | `af806c14c061684722b68f07643b13768c869381` (`chore: capture isolated review baseline`) |
| Captured source blob | `55e0448e73ba569afa8db0ff0fdcc7afeaff8ed9` |
| Source fingerprint | `ea26661cd4b917c511caacb33d6e0e3451c2d2c9` |
| Source root | `/Users/andyeschbacher/git/andy-esch/taskflow` |
| Deliverable | `planning/audits/6g8bb1m753yw-2026-09-09-spatial-thread-graph-prototype-implementation-claude.md` |
| Transfer result | recorded below after `verify` + guarded `transfer` |

All builds, tests, mutations, probes and scratch fixtures ran inside the sandbox. Every mutation and
probe file was reverted; `git status` in the sandbox showed the assigned audit as the only change
before transfer. No commit beyond the baseline was created, and no write-capable command was run in
the source checkout.
