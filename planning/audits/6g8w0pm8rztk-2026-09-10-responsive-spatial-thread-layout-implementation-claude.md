---
schema: 1
id: 6g8w0pm8rztk
bucket: closed
area: responsive-spatial-thread-layout-implementation-claude
date: "2026-09-10"
updated_at: "2026-09-11"
---
# Audit: Responsive spatial Thread layout implementation — claude — 2026-09-10

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

Reviewer: claude (Opus 5), 2026-09-10. Independent adversarial implementation and architecture
review, contract pass followed by the mandatory systemic second pass.

### Isolation attestation and reviewed baseline

Created with `scripts/isolated-review-workspace.sh create --source <checkout> --deliverable
planning/audits/6g8w0pm8rztk-…-claude.md --sandbox-parent <session scratchpad>`:

- `sandbox_path=/private/tmp/claude-501/-Users-andyeschbacher-git-andy-esch-taskflow/7f674cc4-3bf8-48ff-ae8f-580466271ab0/scratchpad/isolated-review.IY6qxq`
- `git_dir=<sandbox_path>/.git` (independent in-tree `.git`, `--no-hardlinks` clone, no alternates)
- `baseline_commit=c19d9feb858cf1398b498faa071e7124b6264189` (sandbox-only "capture isolated review
  baseline" commit whose parent is `e7f63d2`)
- `deliverable=planning/audits/6g8w0pm8rztk-2026-09-10-responsive-spatial-thread-layout-implementation-claude.md`
- `source_blob=5cbe3d20b9d60c3fe61f2207076b67fbdc64f798`
- `source_fingerprint=8cbd0b7c14829b3b5028d23fab19cf4009a7fbae`
- Transfer: performed only after `verify` succeeds. The helper's `transfer=succeeded` line cannot be
  written into this file before it exists, so it is reported with the handoff alongside this document.

Reviewed delta: `git diff --stat e7f63d2 c19d9fe` covers 14 files (+1344/−81). Code: `detail.go`,
`model.go`, `view.go`, `thread_spatial.go`, `thread_projection_test.go`. Planning: task
`6g8btt5hcgs9`, `6g8k47wg8xks` (moved to `completed`), `6g86c7y6hn41` and `6g87qn72901g`
(new `depends_on`), two new follow-ups `6g8vxbv3d4xn`/`6g8vxcnezktm` (both `next-up`), Thread
`6g503c6pfqeb` membership, and the two reviewer briefs. All inspection, builds, probes, and
mutations ran in the sandbox. Every probe and mutation was restored. `git status --short` was
empty before this report was written.

### Verdict

**Not ready as claimed. Four medium and two low findings, all open.** Confidence: high. Every
finding has a deterministic reproduction in the sandbox.

What holds up under hostile evidence:
- The selected card is always complete.
- Viewport counts match the cards actually drawn.
- Line and height bounds hold.
- The cache is race-clean and follows width changes through the real shell.
- The capacity search picks the widest safe width on the production cache path.
- The pane hook stays presentation-owned.

What does not hold up:
1. **Window choice (M1).** The greedy window-adjustment loop routinely shows fewer complete layers
   and rows than fit. On the repo's own production Thread at 140×36 it averages 2.39 complete cards
   against 3.67 achievable.
2. **Test coverage of the geometry contract (M2).** Contract 2 ("every geometry consumer uses the
   chosen node width") has no test behind it. Nineteen fixed-width reverts survive the suite, because
   the whole legacy geometry suite runs at the one width (22) where such a revert changes nothing.
3. **Wide-rune labels (M3).** CJK labels overflow compact cards by one cell, which breaks the box
   and overwrites a route stub.
4. **Stale pane split (M4).** The Thread pane split outlives Thread content.
5. **Broken benchmark (L1).** The committed cached-render benchmark now fails. It is the evidence
   cited by task `6g8k47wg8xks`, which this delta marks completed.
6. **Narrow card (L2).** Below about 55 columns the card drops the required size and the picker
   action.

### Consumer inventory (verified in sandbox, line numbers at baseline `c19d9fe`)

- **Shell hook:** `widthPreferringDetailContent` is declared at `detail.go:78` and implemented only
  by `threadDetail.detailPaneWidthPreference`, which returns `38, min(max(total-42,84),144)` at
  `detail.go:888-894`. Its only consumer is a type assertion in `recomputeLayout` at `view.go:46`.
  Recompute is triggered by the new `default:` branch of `syncDetailImmersion` (`model.go:1350-1356`),
  which is reached only from `detailMsg` (`model.go:351`) and view cycling/retreat
  (`model.go:763,787,940`). The content-clearing paths do **not** trigger it:
  - `exitDashboard` → `detail.clear()` at `model.go:1142`
  - `SetError` at `detail.go:419`
  - `showEmpty` at `detail.go:472`
  - the empty-key branch of `refreshDetail` at `model.go:1240-1243`
- **Cache:** `threadSpatialCache` (mutex, `nodeWidth`, `ready`) is at `thread_spatial.go:122-175`.
  - `get()` prepares at the default 22.
  - `getForViewport()` first prepares at 22 to count layers, then at the responsive width.
  - `prepareWidthLocked()` keys on the *requested* width, not the fallback-chosen width.
  - Runtime callers:
    - `threadDetail.spatialPrepared()` at `detail.go:799`, used for selection and navigation
      (`detail.go:877,911,918,950`)
    - `spatialPreparedForViewport()` at `detail.go:808`, used by `renderDetail` at `detail.go:905`
  - `renderDetail` is reached only from `detailPane.render()` at `detail.go:174-180`, which runs on
    the Update side (SetSize/SetContent/cycleView), never from `View`.
- **Preparation:**
  - `prepareThreadSpatialForViewport` (`thread_spatial.go:188-210`) is a separate non-cache path. Its
    production caller is only the `d.spatial == nil` branch (`detail.go:812`, unreachable via
    `newThreadDetail`). It also backs `renderThreadSpatial` (`thread_spatial.go:1770`), which 10
    tests use.
  - `prepareThreadSpatialAtResponsiveWidth` does the binary search (`thread_spatial.go:221-255`).
  - `threadSpatialResponsiveNodeWidth` is at `thread_spatial.go:265-273`.
  - `planThreadSpatialLayoutWithNodeWidth` (`:411`) clamps to 18–42 and stores `layout.nodeWidth`.
- **Width consumers:**
  - column pitch and route lanes (`:576,:581`), layout width (`:592`)
  - route source/`toRight` endpoints (`:643,:645`)
  - window centering and adjustment (`:1786-1806`)
  - `threadSpatialNodeFullyVisible` (`:1830`), gutters (`:1890-1894`)
  - column headings (`:2311-2314`), routes-for-window (`:2340-2344`)
  - count fallback (`:2466-2472`), node borders and inner width (`:2497,:2507`)
- **Remaining `threadSpatialNodeWidth` uses** (`:101,:142,:154,:185,:198,:204,:267,:408`) are all
  defaults: the zero-width layout fallback, lazy first preparation, the default preparation, the
  unknown-layer default, and the default planner. A literal sweep of `thread_spatial.go` found no
  remaining numeric width geometry.
- **Test helpers still on 22:** `buildThreadSpatialLayout`, `prepareThreadSpatial` and
  `renderThreadSpatialCanvas`, plus `effectiveNodeWidth()`'s zero fallback for hand-built
  `threadSpatialLayout{}` literals. Together these account for 40 call sites in
  `thread_projection_test.go`. Only 19 call sites use `prepareThreadSpatialForViewport`,
  `renderThreadSpatial` or `renderDetail`.

### Validation evidence (sandbox)

- `go test -race -count=1 ./...` → all packages ok, exit 0 (`internal/tui` 9.6s).
- `golangci-lint run ./...` → `0 issues.`, exit 0.
- `go vet ./...` → exit 0.
- `just build` → ok; `./bin/tskflwctl lint` → `✔ all planning entities and dependency links pass
  lint`.
- `just docs-check` → `git diff --exit-code docs/cli` clean, exit 0.
- `git diff --check e7f63d2` → exit 0.

### Hostile fixture and render evidence

- **Full-render sweep.** Temporary probe `TestReviewProbeFullRenderSweep`, removed afterwards.
  - Fixtures (8): the existing hostile fixture (external gate, unreadable/partial topology, missing
    member, control/ANSI/CJK slugs, deep chain, fan-out/fan-in, disconnected), a 30-deep chain with
    duplicate prefixes, 12 disconnected nodes, a 10-wide fan, a lane-heavy long-edge graph, a 5×4
    grid, a single node, and an empty projection.
  - Widths (17): 59/60/61/62/64/80/86/87/88/89/90/91/118/120/140/180/200.
  - Heights (9): 13/14/15/19/20/22/24/31/36.
  - Every node selected in turn: ≈17,000 full `renderThreadSpatialPrepared` frames.
  - Result: 0 frames over height, 0 lines over width after ANSI stripping, 0 selected cards clipped
    or malformed, and 0 mismatches between `viewport N/M` and the distinct `[M#]/[G#]` card identity
    lines actually drawn.
  - The selected-card check reads the five card rows out of the stripped frame with
    `ansi.Cut(line, x-panX, x-panX+nw)`, independently of the canvas.
- **Resize through the real shell.** 200→64→65→80→200→62→140 with immersive spatial.
  - Cached node widths were 42/25/25/33/42/24/42.
  - These match hand-computed expectations for two layers, `(inner−4−8)/2` clamped.
  - The first rendered node border in `View()` had exactly that width at every step.
- **Race detector.** 16 goroutines sharing value-copied `threadDetail`s ran 3,200 interleaved
  `spatialPreparedForViewport`/`spatialPrepared`/render calls across widths 60–209 with no race
  report.
- **Capacity.** On the near-canvas fixture with 60/75/91/100/110/120 rows and widths 64/87/120/180,
  the cache path picked the brute-force widest safe width in all 24 cases. Examples:
  - rows 91: 22
  - rows 100: 19, where the old fixed-22 build rejected the graph
  - rows 110+: issue at every width, same as before
  - The non-cache `prepareThreadSpatialForViewport` diverged once (rows 100, width 87): it returned
    an issue where 19 fits. This is folded into M2.
- **Performance** (`-benchtime=20x`, near-canvas 91-row fixture):
  - Default preparation: 0.22–0.29 ms.
  - First `getForViewport(180)` on a fresh cache: **1.53 ms and 7 preparations** (widths
    22,42,29,23,20,21,22).
  - First `getForViewport(120)`: 1.30 ms. First `getForViewport(64)`: 0.47 ms. That is 2
    preparations, and the 22 one exists only to count layers.
  - Each resize crossing between two buckets: 0.66 ms, re-running the search.
  - Same-bucket render at 121×30: 0.16 ms.
  - A never-fitting 110-row graph costs 8 plan-only preparations per bucket.
  - All bounded and sub-2 ms, so this is a performance note, not a correctness defect.
- **Dogfood** (`complete-production-threads` loaded from the sandbox `planning/`: 46 nodes, 61
  edges, 17 layers). Terminal sizes are mapped to the immersive region (width−2, height−5).
  | Terminal | Node width | Selections short on layers | Short on rows | Only the selected card visible |
  |---|---|---|---|---|
  | 80×24 | 19 | 26/46 | 0/46 | 15/46 |
  | 100×30 | 26 | 29/46 | 5/46 | 9/46 |
  | 140×36 | 39 | 26/46 | 18/46 | 5/46 |
  | 200×50 | 42 | 26/46 | 25/46 | 0/46 |
- **Pane policy.** At 200 columns, Threads measure 56/144. See M4 for what happens after leaving
  them.

### Mutation evidence

Each mutation was applied to the sandbox, run with `go test ./internal/tui -count=1`, and restored
with `git checkout -- <file>` plus a clean-diff check. Mutations that name "the fixed constant"
replace the responsive width with `threadSpatialNodeWidth`. Where a single edit left `nodeWidth`
unused and broke the build, the function's width source was mutated instead (a coordinated
mutation).

**Killed:**

| Mutation | Test that failed |
|---|---|
| Cache ignores width, rebuild guard is `if c.ready` | `TestThreadSpatialCacheFollowsCoherentProjectionReplacement` |
| Cache always rebuilds | That test, plus `…ResponsiveWidthKeepsAValidNearLimitLayoutAvailable` and `…RuntimeMethodsConsumeTheCachedPreparedResult` |
| Capacity fallback disabled | `TestThreadSpatialResponsiveWidthKeepsAValidNearLimitLayoutAvailable` |
| Complete-node drawing guard removed | `TestThreadSpatialWindowKeepsNodesWholeAndExposesHiddenExtent` |
| Column-heading guard removed | The same test |
| Gutter `text` check removed; gutter `connector` check removed | `TestThreadSpatialWindowGutterNeverOverwritesGraphEvidence` |
| Viewport counts every node | `…WindowKeepsNodesWhole…` and `…VerticalWindowReservesOneCompleteSelectedSlot` |
| `▼` and `▸` markers disabled | Killed |
| `fixedRows` 9→10 | `TestThreadSpatialGraphIsBoundedDeterministicAndExplicitWhenNarrow` |
| Min height 14→15 | `TestThreadSpatialHeaderReportsLayoutConflictOutsideTheViewport` |
| Narrow-card dimensions text | Both narrow tests |
| Four pane mutations: sync recompute removed, preference ignored, 144 cap removed, list floor 38→28 | `TestThreadDetailUsesAResponsiveIdentitySafePaneBudget` |
| Width helper layer budget 3→4 | Only `TestThreadSpatialResponsiveGeometryPreservesCompactIdentity` (literal 18/33/42). The cache test derives its `wantWidth` from the same helper (`thread_projection_test.go:659`), so it is self-fulfilling. |

**Survived:** listed under M2 and L2.

### Findings

#### M1. Greedy window adjustment hides complete layers and rows that fit · **Status:** fixed

**Evidence.** In `threadSpatialWindowForSelection` (`thread_spatial.go:1776-1823`), the window first
centres on the selection. It then makes one pass over `layout.nodes` in column-major order:
- A straddled node on the left pushes the window's left edge to the node's right edge.
- A straddled node on the right pulls the window's right edge to the node's left edge.

The rule always excludes the straddled node and never considers including it. Later nodes undo
earlier fixes, and the last node processed wins. The result:
- one neighbour ends up sheared, and is therefore omitted by the new complete-node rule;
- the neighbour on the other side is excluded;
- even when an aligned window would show both.

A brute-force oracle maximized complete layers and rows over every window that still contains the
selected slot. Examples:
- **`fan10` at 80×24 selecting `mid04`.** Layout columns are at x=3/46/89 (node width 20, gaps
  23). The window ping-pongs 16→23→9 and shows only the selected card: `viewport 2/12 nodes · ◂
  layers 2/3 ▸`. Routes arrive from and leave into blank space labelled `[M11]…` and `…▶[M12]`.
  Yet x=3 (root and selection) or x=46 (selection and sink) each fit two complete layers.
- **Twelve disconnected nodes at 80×20** (11-row canvas, row pitch 6). All 12/12 selections show
  one card with five blank rows. A window starting at `selected.y−6` fits two complete rows.
- **Real production Thread.** See the dogfood table above. At 140×36 the mean is 2.39 complete
  cards against 3.67 achievable (26/46 selections short on layers, 18/46 on rows). At 80×24, 15/46
  selections show only the selected card.

**Impact.** This works against the task's core premise of spending space on complete units (AC 2
and AC 3) and against planning's own diagnosis that "only a few complete nodes fit". The picture
is truthful: counts and gutters are right, and the selection is never clipped. But the
prerequisite and dependent of a fan node disappear even though the terminal has room for them.
This is a regression from `e7f63d2`: the earlier code drew those neighbours, clipped.

**Smallest sound correction.** Pick each axis's pan from unit boundaries instead of a mutating
pass:
- Candidates are windows left-aligned at a column start (`columnX[i]` minus a margin) or
  right-aligned at a column end (`columnX[j]+nw−width`), clamped to the layout, that contain the
  selected column.
- Choose the candidate with the most complete columns, breaking ties by distance from the centred
  pan.
- Do the same over `rowY` with the slot height.

This is O(columns + rows) and deterministic. Add a regression that compares complete
layer and row counts to a brute-force oracle over a fan and a disconnected fixture at 80×24 and
80×20, rendered end to end.

**Belongs in this task:** yes. The planning already defers large-graph overview/focus design to
`6g8vxbv3d4xn`; this bug is inside the window this task owns.

**Resolution:** Replaced order-dependent window adjustment with deterministic
unit-boundary optimization; independent brute-force oracles cover fan layers and
disconnected rows.

#### M2. The responsive-geometry contract has no tests; the suite is pinned to the default width · **Status:** fixed

**Evidence.** This is the systemic finding from the second pass. Contract 2 says routes, lanes,
endpoints, borders, selection, clipping, and capacity "cannot drift back to the former fixed
width". Nothing in the suite enforces that. Each of the following reverts to `threadSpatialNodeWidth`
leaves the entire `internal/tui` suite green:
- **Single-site reverts (12):**
  - route source endpoint (`:643`) and reverse `toRight` endpoint (`:645`)
  - boundary lanes (`:581`), column pitch (`:576`)
  - node inner width (`:2507`), node borders (`:2497`)
  - count-fallback middle and right (`:2467,:2472`)
  - window centring (`:1787`), window adjustment extents (`:1797`)
  - column-heading truncation and its guard (`:2311,:2314`)
- **Coordinated function-level reverts (7):** `materializeThreadSpatialLayout`'s route width
  (`:487`), `threadSpatialWindowForSelection`, `threadSpatialNodeFullyVisible`,
  `annotateThreadSpatialWindowGutters`, `drawThreadSpatialCanvasWindow` (node drawing and
  headings), `threadSpatialRoutesForWindow`, and `threadSpatialCountFallback`.

The same pattern shows in the other claims:
- The "widest safe" claim has no test. Stopping the binary search at the first fitting midpoint
  (`:244`, `low = candidateWidth+1; high = low-1`) survives, because the near-limit test only
  asserts `< 42` and no issue.
- The relationship between min height and fixed rows has no test either:
  - `fixedRows` 9→8 survives, silently cutting the inspector's bottom border at every height.
  - `threadSpatialMinHeight` 14→13 or →12 survives, although a 3- or 4-row canvas cannot hold the
    selected five-row slot.
- Range upper bounds (`lastColumn+1`, `lastRow+1`, `:1849,:1871,:1875`) and the gutter's
  crossing, overlap, conflict, route-count and continuation guards (`:1920-1921`) survive. Those
  cells carry `connector=0` and empty text (`:1417-1434`), so the guards do real work and
  deserve tests.

**Why all of this survives.** Three things keep the whole geometry regression family at width 22,
where a responsive-versus-fixed confusion is invisible:
- The legacy helpers default to 22 (`prepareThreadSpatial`, `buildThreadSpatialLayout`,
  `renderThreadSpatialCanvas`).
- `effectiveNodeWidth()` quietly returns 22 for hand-built layouts.
- The new window and whole-node tests render at width 64, where node width 18 is *below* 22, so a
  revert only makes the checks stricter.

In addition, `renderThreadSpatial` goes through `prepareThreadSpatialForViewport`, not the
production cache. That function returns `base` early whenever the responsive width is 22
(`:198-200`) and so skips the capacity fallback the cache applies. The near-canvas 100-row
fixture at viewport 87 is rejected by the test path and rendered at node width 19 by the
production path.

**Impact.** A single future edit can split geometry between two widths: routes attached inside
cards, lanes over borders, nodes drawn sheared as "fully visible". The planning checkpoint and
AC 5 would still read as satisfied.

**Smallest sound correction.** No production change is needed for the widths themselves:
1. Add one end-to-end test that renders a fan, a reverse edge and a long-edge fixture at node
   widths 18 and 42 through `threadDetail.renderDetail` (the cache path). Assert, from the
   stripped frame:
   - every drawn card's borders are exactly `nw` cells wide;
   - each route's first cell sits at the card edge and its arrow at `x−1`;
   - no card cell lies outside the viewport.
2. Pin the widest-safe choice against a brute-force loop.
3. Derive `threadSpatialMinHeight` from `fixedRows + threadSpatialNodeSlotHeight`, and test at
   exactly min height and min−1 for a complete selected card plus the inspector's `╰` border.
4. Extend the gutter test to crossing, overlap, conflict, continuation and route-count cells.
5. Make `renderThreadSpatial` and `prepareThreadSpatialForViewport` delegate to the same
   `prepareThreadSpatialAtResponsiveWidth` policy as the cache, dropping the width-22 early
   return.

**Belongs in this task:** yes. These are the claimed regressions.

**Resolution:** Unified cache and non-cache responsive fallback, derived
physical height bounds, and added exact 18/42-width geometry, widest-safe
capacity, complete-window, and guarded-gutter regressions.

#### M3. Wide-rune labels overflow compact cards and overwrite the outgoing route · **Status:** fixed

**Evidence.** `drawThreadSpatialNode` now builds compact identity from
`identityPrefix + truncateMiddle(label, labelWidth)` (`thread_spatial.go:2512-2518`) and pads it
without truncating. `truncateMiddle` (`item.go:269-284`) exceeds its width budget by one cell
whenever `ansi.TruncateLeft` keeps a wide rune that straddles the cut.

For a 24-rune `移行…` label, `truncateMiddle(label, w)` returns width w+1 at w=4/8/12/16/20/24/28.
Examples: `"移行…移行"` is width 9 at w=8, and `"移…行"` is width 5 at w=4.

End to end, a chain `a→b→c` with that label at widths 77–79 (node width 19) renders:
- top border `┌─────────────────┐` (19 cells);
- middle row `│ [M2] • 移行…移行 │` (20 cells);
- the displaced `│` landing on the first cell of the selected route: `━━━━━━▶`, not
  `━━━━━━━▶` as at width 80.

Across the probe matrix (10 hostile labels × 4 aliases × node widths 18–42 × compact/selected),
244 rows had the right border in the wrong place. All the failures were compact rows with wide
runes. Selected cards are safe because `line()` truncates.

**Impact.** A CJK label (the existing hostile fixture already carries `移行` slugs) produces a
malformed box. It also erases route evidence at the card edge, which violates contract 7's "every
box remains well-formed".

**Smallest sound correction.** Make `truncateMiddle` width-exact: after `TruncateLeft`, drop one
more cell while `StringWidth(right) > rightWidth`. Alternatively, clamp the compact identity with
`truncate(…, insideWidth)` before `padRight`. Test `truncateMiddle` width for every w against
all-wide, mixed and emoji inputs, and assert the compact middle row is exactly `nodeWidth` cells.
The same helper also serves the Thread list's `threadIdentityLabel` (existing behaviour, not new
here), so fixing the helper corrects both.

**Belongs in this task:** yes. The spatial use is new in this delta.

**Resolution:** Made middle truncation honor exact terminal-cell budgets for
wide graphemes and added Unicode plus rendered-card border regressions.

#### M4. The Thread pane split survives after Thread content is gone · **Status:** fixed

**Evidence.** The split is recomputed only on `detailMsg` and on view cycling or retreat.
`exitDashboard` → `detail.clear()` (`model.go:1142`) and `SetError` (`detail.go:419`) set
`content=nil` without recomputing. `refreshDetail` returns `nil` for a tab with no selection
(`model.go:1240-1243`), so no later `detailMsg` arrives either.

Probe: `threadRepo`, 200×36, Threads loaded at list 56 / detail 144, then switch tab.

| Tab | Before load | After drain | A fresh `recomputeLayout()` gives |
|---|---|---|---|
| Tasks | 56/144 | 80/120 | 80/120 |
| Epics (empty) | 56/144 | 56/144 | 80/120 |
| Audits (empty) | 56/144 | 56/144 | 80/120 |
| Research (empty) | 56/144 | 56/144 | 80/120 |
| Thread load error | 56/144 | 56/144 | 80/120 |

For Tasks, the list renders at the Thread width until the task detail arrives, then jumps. For
empty tabs and error panes the Thread split persists until the next resize.

**Impact.** This directly violates contract 8 ("switching entities/views … recomputes the correct
split"). It is also the brief's explicit "no stale Thread preference survives a content switch"
check. Every Threads→other-tab switch visibly re-lays out the list once the detail loads.

**Smallest sound correction.** Recompute the layout whenever the detail content identity changes:
- `clear()`, `SetError`, and `showEmpty` in the model, or better,
- one "detail content changed" hook that `syncDetailImmersion`'s default branch also uses.

Add a model test: Threads at 200 wide → switch to an empty tab → expect 80/120, and the same
after a detail error.

**Belongs in this task:** yes.

**Resolution:** Recompute shell layout when detail becomes empty or errors and
when tab content is cleared; model regressions prove the default split is
restored.

#### L1. The committed cached-render benchmark now fails; task `6g8k47wg8xks` cites it · **Status:** fixed

**Evidence.** `BenchmarkThreadSpatialCachedRenderNearCanvasLimit` (`thread_projection_test.go:2943-2957`)
captures `detail.spatialPrepared().layout` (the default-22 preparation) and then renders at
120×30:
- At the sandbox baseline it fails: `thread_projection_test.go:2954: render replaced the cached
  layout`.
- At `e7f63d2` (checked out temporarily into the sandbox and then restored) it passes at 250 µs/op.

`go test` never runs benchmarks, so the green suite hides this. This delta flips `6g8k47wg8xks` to
`completed`, and that task's checked benchmark AC and prose cite this exact cached 120×30 render.

**Impact.** The committed performance evidence for the viewport work no longer runs, and there is
no benchmark for the new responsive cache path.

**Smallest sound correction.** Capture the layout after one `renderDetail(120, 30)` warm-up, or
from `getForViewport(120)`. Optionally add benchmarks for a first open and a bucket-crossing
resize. Run `go test -run '^$' -bench ThreadSpatial` as part of this task's evidence.

**Belongs in this task:** yes.

**Resolution:** Prime the cached-render benchmark at its responsive viewport
width; the Thread spatial benchmark set now runs successfully.

#### L2. The narrow card truncates away the required size and picker action · **Status:** fixed

**Evidence.** `renderThreadSpatialNarrow` (`thread_spatial.go:2683-2704`) spends its three
content lines on:
1. a 54-cell requirement sentence;
2. the decorative line "Give the map room to breathe…";
3. a 58-cell action line.

Rendered at height 20:
- width 54: `f picks` is lost;
- width 50 and below: `60×14` is lost (`…needs at least 60×…`);
- width 30: only `Esc r…` survives.

Width-driven fallback happens exactly when the terminal is narrower than about 62 columns, so the
action line is truncated in *every* width-driven case. The narrow test checks only 54×10, for
"needs at least" and "Current 54×10".

**Impact.** Contract 7 promises current and required dimensions plus retreat and picker actions.
At common split-terminal widths the user gets neither the target size nor the picker hint.

**Smallest sound correction.** Lead with compact facts (`need ≥60×14 · now 48×20`), then
compact actions (`Esc waves · f pick · resize`), and put flavour text last or drop it. Test at
widths 40/50/59 for both `60×14` and `f`.

**Belongs in this task:** yes.

**Resolution:** Reordered narrow-state content around compact required/current
dimensions and actions; explicit 40/59/60-column and 13/14-row boundaries retain
the guidance.

### Settled hypotheses and rejected concerns

- **Selected card sheared, or negative/out-of-range origins:** falsified. 0 of ≈17,000 frames had
  a clipped or malformed selected card, and the pan is clamped after every adjustment. The sweep found no case where M1's
  ping-pong clipped the selection.
- **Viewport scope untruthful:** falsified for counts. Reported complete-node counts equal the
  distinct card identity lines drawn in every frame. At the maximum supported size the summary
  (~71 cells) loses its trailing `rows … ▼` segment at width 60, but `viewport N/M nodes` comes
  first and survives, and `▲▼` gutters remain.
- **Stale geometry, double rebuilds, or races in the cache:** no correctness defect.
  - Value copies share the pointer; reload builds a new cache (`newThreadDetail`, `detail.go:786`).
  - `SetRefreshError` keeps the last coherent detail (`detail.go:434`).
  - The resize sequence tracks independent expectations.
  - Race-clean under 16 concurrent readers.
  - Repeated preparation on first open and bucket crossing is measured above (≤1.5 ms) and is a
    bounded performance follow-up, not a defect.
- **Capacity arithmetic unsound:** falsified. `threadSpatialColumnGeometry` width is linear in node
  width and height does not depend on it, so area grows monotonically and bisection is valid.
  Preflight node and edge issues do not depend on width and are returned before any fallback. The
  cache path matched the brute-force widest safe width in 24/24 cases. The one divergence is on
  the test-only path (folded into M2).
- **Geometry constant escaped responsive propagation in production:** falsified by the inventory
  above. Every remaining `threadSpatialNodeWidth` use is a default. The risk is the untested
  contract (M2), not current drift.
- **Shell coupled to Threads, or I/O in `View`:** falsified.
  - The hook is a presentation-owned optional interface.
  - Preparation happens in Update-side `render()`, never in `View`.
  - No store or `core.Service` call was added.
  - The node-count behaviour is presentation-only; nothing switches representation automatically.
- **89/90 and very wide split:** no violation. At 90 columns Threads get 38/52 against a default of
  36/54, which is the intended identity floor. At 400 columns the list takes the surplus (256).
  The comment "instead of growing either pane without bound" in `detail.go:890-892` overstates
  this for the list, but that is cosmetic and not reported.
- **Narrow path defeating the capacity guard:** falsified. The narrow card allocates no canvas
  and uses `threadSpatialInspectorPrepared`.
- **Planning honesty:**
  - The large-graph and whole-TUI uncertainties are tracked, not claimed solved: `6g8vxbv3d4xn`
    and `6g8vxcnezktm` are `next-up`, joined to Thread `6g503c6pfqeb`, and gate `6g86c7y6hn41`
    and `6g87qn72901g`.
  - AC 2, 3 and 5 of `6g8btt5hcgs9` are checked, but they overstate the executable evidence (see
    M1 and M2).

### Bounded follow-up (not a defect in this task)

Reuse one ranking plan across widths and compute the widest safe width arithmetically from
`sum(gaps)` and the column count before materializing, instead of re-planning and materializing
up to seven candidates per bucket. Also cache by chosen width, so crossing back into a bucket
does not repeat the search. This measurably cuts the 1.5 ms first open and 0.66 ms bucket crossing
on near-limit graphs, but it is not required for correctness.
