---
schema: 1
id: 6g8g43px88xr
bucket: closed
area: dense-thread-route-trustworthiness-design-review
date: "2026-09-09"
updated_at: "2026-09-09"
---
# Audit: Dense Thread route trustworthiness design review — 2026-09-09

> Reviewer assignment: independent product and interaction design reviewer. This document is the
> review brief and the only repository file the reviewer should update.
>
> This is an experience-first review of the rendered spatial Thread graph after dense-route
> hardening. Do not substitute source inspection, passing tests, or the prior implementation audits
> for seeing and using the TUI. If viewable color output is unavailable, report that limitation and
> do not invent a visual verdict.
>
> Finding grammar is exact: use `#### H1. <title> · **Status:** open` (or M1/L1). Leave every new
> finding open for implementation-owner triage.

## Mandatory reviewer sandbox

The source checkout is shared and may contain active owner work. Reading this brief and creating the
isolated copy are the only operations allowed there. Substitute this audit's repository-relative
path below:

```sh
SOURCE_ROOT="$(git rev-parse --show-toplevel)"
AUDIT_REL="planning/audits/<this-design-audit-file>.md"
SANDBOX="$("$SOURCE_ROOT/scripts/isolated-review-workspace.sh" create \
  --source "$SOURCE_ROOT" \
  --deliverable "$AUDIT_REL" \
  --print-path)"
cd "$SANDBOX"
```

The helper creates an independent no-hardlinks clone, overlays the exact handoff state, and records
a sandbox-only baseline commit. Perform every build, TUI session, screenshot, fixture mutation,
render probe, and report edit there. Do not create further commits. Before transfer, restore every
probe and generated artifact so the audit is the only diff, then run:

```sh
"$SANDBOX/scripts/isolated-review-workspace.sh" verify --sandbox "$SANDBOX"
git diff -- "$AUDIT_REL"
"$SANDBOX/scripts/isolated-review-workspace.sh" transfer --sandbox "$SANDBOX"
```

Leave the sandbox in place and include the helper's isolation and transfer attestation in the
report. If verification or transfer refuses, preserve the sandbox and report the conflict rather
than working around the guard.

## Review purpose

Evaluate whether hardened dense routes are visually trustworthy *and* practically comprehensible.
The implementation is intended to stop the graph from inventing endpoints or junctions, but a
technically faithful transit map can still be too noisy, too subtle, or too dependent on its legend
to help a person understand the work.

Answer these questions from observed use:

1. Can a user follow one dependency from its actual prerequisite to its actual dependent through
   turns, crossings, tracks, shared stubs, and viewport clipping without losing ownership?
2. Can a user distinguish a crossing from a true fan-in/fan-out bundle, and a normal route from an
   overlap or invariant-conflict warning, without guessing from color alone?
3. Do focus, semantic task state, direction, endpoint multiplicity, and offscreen attribution form
   a coherent visual hierarchy, or do their colors and glyphs compete?
4. Does the route presentation remain useful on the real production Thread at ordinary terminal
   sizes, or does added correctness turn density into excessive whitespace or line noise?
5. What is the smallest design refinement needed before this becomes a flagship TUI surface, and
   what should remain experimental?

## Review target and context

Review branch `feat/dense-thread-route-trustworthiness`, specifically commits:

- `6a4e4c4` — dense route layout, canvas grammar, viewport annotation, and regressions;
- `806ab90` — planning evidence, audit dispositions, and sequenced follow-up.

The branch is based on `main` at `93b4965`. Primary implementation is
`internal/tui/thread_spatial.go`, with behavior evidence in
`internal/tui/thread_projection_test.go`. Read these planning artifacts for intent and boundaries,
but do not treat their claims as proof of design quality:

- `planning/tasks/6g86e10dztpf-make-dense-thread-graph-routes-visually-trustworthy.md`
- `planning/audits/6g8e0pyczpns-2026-09-09-dense-thread-graph-routes-implementation-claude.md`
- `planning/audits/6g8e0pyn885r-2026-09-09-dense-thread-graph-routes-implementation-antigravity.md`
- `planning/audits/6g8bbjgcx9tj-2026-09-09-spatial-thread-graph-tui-experience-design-review.md`
- `planning/tasks/6g8btt5hcgs9-make-spatial-thread-layout-responsive-to-available-space.md`
- `planning/tasks/6g86c7y6hn41-add-a-one-hop-focus-subgraph-mode-to-the-tui-thread-graph.md`
- `planning/tasks/6g8ezj5e51hg-cache-and-preflight-spatial-thread-layout-work.md`
- `planning/threads/6g503c6pfqeb-complete-production-threads.md`

Keep the architectural boundary explicit: this renderer consumes `ThreadGraphProjection`; it must
not redefine graph direction, waves, readiness, health, or persistence. This review may recommend a
different presentation technique, but it must identify the demonstrated experience problem rather
than prescribing a graph library or layout engine by taste.

## Required visual evidence protocol

Build and run only inside the sandbox, with real terminal color enabled:

```sh
GOCACHE="$SANDBOX/.cache/go-build" just build
unset NO_COLOR
export TERM=xterm-256color
./bin/tskflwctl ui
```

Use the repository planning space and open `complete-production-threads`. Exercise its spatial view
at approximately 140×36, 100×28, and 72×24 cells, plus one alternate available theme. Capture
viewable frames or screenshots with the terminal size and theme recorded. Also inspect a
color-disabled rendering to determine whether glyph and shape still carry the meaning; that pass is
accessibility evidence, not a substitute for the color review.

If the production Thread does not naturally expose a required case, create a sandbox-only disposable
planning space or focused rendering probe containing known edges. It must include, at minimum:

- ordinary adjacent dependencies and a skipped-layer dependency;
- fan-in and fan-out of at least three routes;
- two unrelated perpendicular routes;
- two routes sharing a real endpoint stub;
- a same-layer or reverse route and a self-edge;
- a selected route whose other endpoint is clipped from each viewport side; and
- enough nodes and edges to require both horizontal and vertical panning.

Record the supplied edge list beside each synthetic frame so the reviewer can compare what the
picture suggests with what the graph actually contains. Do not leave the fixture or captures in the
source checkout or transfer them with the audit.

## Task-based walkthroughs

Perform these as user jobs rather than isolated glyph checks:

1. **Read before decoding.** Give the wide production graph ten seconds without reading the legend.
   State which task is selected, its state, what it needs and unlocks, where work converges, and
   which direction dependencies flow. Then consult the legend and record what changed.
2. **Ride real routes.** Pick a root, a fan-out, a fan-in, a long edge, an external gate, and a leaf.
   Trace each selected incident route visually, predict the next `h`/`l` destination, move, and
   compare navigation with the line that appeared to connect the nodes.
3. **Decode collisions.** Locate or construct crossing, shared-stub, overlap, and conflict grammar.
   Ask what relationship each glyph communicates before looking up its definition. Verify that a
   crossing never looks selectable or connected and that shared geometry does not conceal route
   direction.
4. **Pan through ownership.** Follow selected routes across every viewport boundary. Check whether
   grouped aliases remain associated with the correct line, whether fan counts help or distract,
   and whether a route leaving and re-entering the viewport can still be understood.
5. **Remove visual advantages.** Repeat representative tracing at medium and narrow sizes, with a
   non-default theme and with color disabled. Identify the first point at which comprehension fails
   and whether the wave reader fallback is discoverable and honest.
6. **Compare reading modes.** Answer one sequencing question and one local-dependency question in
   both waves and spatial views. Record which view is faster and why; do not assume the spatial view
   should replace the wave reader.

## Design lenses to challenge

- **Route continuity:** elbows, tracks, waypoints, crossings, and shared segments read as one owned
  path rather than disconnected line fragments or table rules.
- **Topology honesty:** arrows and junction grammar cannot reasonably imply an absent edge,
  endpoint, or direction even when several routes occupy nearby cells.
- **Visual hierarchy:** selected incident routes are prominent, task status remains semantic,
  direction stays distinct from focus, and warnings do not resemble ordinary graph content.
- **Density and whitespace:** lane separation buys comprehension rather than merely expanding the
  canvas; long corridors do not dominate nodes or create excessive panning.
- **Endpoint annotation:** fan counts and grouped offscreen aliases are readable, correctly located,
  and more useful than the detail they replace.
- **Legend burden:** common cases are understandable without constant lookup; rare warning grammar
  remains discoverable and concise at realistic widths.
- **Navigation congruence:** `hjkl` follows the route the selected node appears to own, including
  skipped layers, multiple candidates, cycles, and viewport changes.
- **Accessibility and terminal behavior:** meaning survives color loss, theme changes, ANSI
  clipping, Unicode glyphs, and ordinary font/rendering differences.
- **Extensibility:** recommendations stay inside the TUI presentation adapter and could support a
  future web renderer without leaking layout state into the graph/domain contract.

## Existing ownership and non-duplication

Do not reopen an already tracked concern merely because it remains visible. Map evidence to the
existing owner unless this route change materially changes its scope:

- complete-unit clipping and adaptive viewport composition: `6g8btt5hcgs9`;
- one-hop local focus/subgraph mode: `6g86c7y6hn41`;
- layout preflight, reuse, and performance budgets: `6g8ezj5e51hg`;
- alternate-view and Thread identity discoverability: `6g87qn72901g`;
- reusable Back navigation: `6g86016qvk5d`.

It is valid to corroborate those tasks without opening a duplicate finding. Open a new finding only
for a demonstrated route-design problem, an inadequate existing scope, or a conflict between two
planned treatments. Distinguish a correctness defect from an aesthetic preference and a
maintainer-taste question.

## Deliverable

Preserve this brief and replace only the reviewer-report placeholder below with:

1. an executive verdict: visually trustworthy, promising but needs refinement, or design direction
   needs revision;
2. the isolation and transfer attestation;
3. an evidence matrix containing frame id, terminal size, theme/color mode, graph fixture, selected
   node/route, and observation;
4. a compact comprehension scorecard for direction, endpoint ownership, crossings, shared routes,
   clipping, density, and navigation congruence;
5. prioritized findings in exact audit grammar, each with observed evidence, user consequence, and
   the smallest bounded recommendation;
6. a task mapping that points each accepted recommendation to an existing owner or proposes one
   clearly bounded new task;
7. compact text wireframes only where they make a recommendation materially clearer; and
8. residual questions that need maintainer taste or real-user observation.

Do not edit implementation, tests, fixtures, tasks, Threads, ADRs, or another audit. Do not close
this audit or pre-resolve findings.

## Reviewer report
### 1. Executive verdict

**Promising but needs refinement.**

The correctness work landed. I could not make this build invent an endpoint, a junction, or a
direction in any frame I captured. The three-edge reversal that broke the previous build
(`g1→h3, g2→h2, g3→h1`) now renders as three owned paths with real elbows and two honest crossings;
grouped offscreen aliases matched the projection exactly every time I checked them against
`thread graph --json`; and `hjkl` followed a real supplied edge on every move I tested. The renderer
also now refuses to lie by construction — collinear overlap and turn-versus-crossing conflicts are
classified rather than flattened.

It is not yet comprehensible. Three things stand between this and a flagship surface, and all three
are grammar/hierarchy problems rather than topology problems:

1. **The endpoint multiplicity marker is suppressed exactly where congestion makes it necessary.**
   `putRouteCount` bails when the cell before the arrowhead is a crossing, so the busiest fan-in in
   a frame is the one that shows no number. On the production Thread the node with **5**
   prerequisites was the only visible node with no count, while its two-prerequisite neighbours both
   showed `2`. A reader comparing "2, 2, nothing" concludes the opposite of the truth.
2. **Crossing, shared, overlap and conflict glyphs inherit the selected route's magenta accent**, so
   the marks that mean "another route is here" are painted in the colour that means "this is your
   route". A selected connector reads `─╳╳━›╳▶` in one unbroken magenta-bold span.
3. **The legend is 141 display cells and never fits.** It is truncated at 140, 100 and 72 columns,
   and the tokens it drops first (`╳ cross`, `━/┃/◆ shared`) are the ones actually on screen, while
   the tokens it protects (`≋ overlap`, `! conflict`) never appeared in any frame I captured.

Underneath all three sits a density problem that the existing responsive-layout task already owns:
at 140×36 the production Thread shows 5 of 43 nodes and roughly 55% empty canvas; at 100×28 it
shows 3 nodes and 9 of 12 graph rows are blank; at 72×24 it shows one whole node. In the
head-to-head reading test the wave reader beat the spatial view on **both** questions at every size.

Nothing here argues for a different layout engine. Every recommendation below is a grammar,
colour-role, or legend change inside the existing presentation adapter.

### 2. Isolation and transfer attestation

Every build, TUI session, fixture mutation, capture and report edit happened in the independent
clone below. The source checkout was read for this brief and the initial copy only. Go build cache,
the compiled binary, captured frames, and the ANSI→HTML tooling were kept **outside** the sandbox
(`/private/var/folders/…/T/dr-work/`). The disposable planning space
(`$SANDBOX/.review-fixture`, 21 tasks / 23 edges / 2 Threads) was created inside the sandbox as the
brief allows, then deleted; `tskflwctl init` had also registered it in the user-level space registry
at `~/.config/tskflwctl/spaces.toml`, so it was removed with `tskflwctl space forget
.review-fixture` (verified absent from `space list`). `git status --porcelain
--untracked-files=all` was empty before this report was written.

```
sandbox_path=/private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.iJU8Aw
git_dir=/private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.iJU8Aw/.git
baseline_commit=16a5589b2d60a2e0ea1471bf4d914e2f695ae40b
source_blob=323196426ae2ec591d92395532de838115e8a479
source_fingerprint=71989b89887e9d6aa7d7f45430b21aec8694492d
deliverable=planning/audits/6g8g43px88xr-2026-09-09-dense-thread-route-trustworthiness-design-review.md
deliverable_changed=true
transfer=pending
```

`verify` passed fail-closed on the finished report: HEAD still at the baseline commit, no staged
changes, no untracked files, exactly one unstaged delta (this audit), `git diff --check` clean, and
the source deliverable still hashing to the recorded `source_blob`. The transfer runs after this
file is finalised, so its `transfer=succeeded` / `source_deliverable=` / `sandbox_retained=` lines
cannot appear inside the transferred bytes; they are reproduced in the handoff message. The
workspace is retained at `sandbox_path` until the implementation owner confirms receipt.

**Visual method.** No graphical terminal was available, so frames were captured from a real PTY:
`tmux 3.6b` sessions sized exactly to each target, running the sandbox-built
`tskflwctl ui` under `TERM=xterm-256color`, captured with `tmux capture-pane -e` (escape sequences
preserved), then rendered to HTML with a truecolor SGR decoder and screenshotted in Chrome at
FiraCode Nerd Font 15px. Colour, bold and dim are therefore real and inspected visually; what is
**not** covered is a physical terminal's own font substitution and box-drawing hinting (see §8).

### 3. Evidence matrix

| Frame | Size | Theme / colour | Graph | Selected | Observation |
| --- | --- | --- | --- | --- | --- |
| F1 | 140×36 | neon, colour | production `complete-production-threads` (43 nodes / 57 edges / 15 waves, canvas 470×75) | `[G1]` ship-guarded-dependency-mutations (external gate) | 5 nodes visible; ~55% empty canvas; legend truncated at `! conf…`; run into `add-a-guarded-r…` is `─────────╳╳━━╳╳▶` with **no fan count** although 5 edges arrive; two neighbours show `┌╳2▶` |
| F2 | 140×36 | neon, colour | production | `[M25]` add-a-guarded-repair-path (fan-in 5) | Grouped alias `[M24,M13,M14]…` plus `[M7]…` correctly name 4 offscreen prerequisites (verified against `thread graph --json`); both labels sit 20 rows and ~100 columns from the arrowhead they explain; connector reads `─╳╳━›╳▶` — 6 glyph classes in 7 cells |
| F3 | 100×28 | neon, colour | production | `[G1]` | 3 nodes visible; 9 of 12 graph rows blank; legend cut after `◇ turn` while `╳` and `━` are both on screen |
| F4 | 72×24 | neon, colour | production | `[G1]` | 1 whole node; 2 half-clipped nodes losing their status glyph and left border; legend cut mid `▶/◀ directi…` |
| F5 | 140×36 | **catppuccin**, colour | production | `[G1]` | Same structure; accent/status hues change (`203;166;247` focus, `166;227;161` done) and the hierarchy problem is identical — crossings still adopt the accent |
| F6 | 140×36 | neon, **NO_COLOR=1** | production | `[G1]` | Only SGR `0/1/2` remain. Status glyphs, gate double-border and the expanded focus card survive. Selected route is **bold-only** and near-indistinguishable. Legend still reads `magenta=focus` |
| F7 | 140×36 | neon, colour | fixture `route-grammar` (15 nodes / 20 edges / 1 gate) | `[G1]` a4 (single edge a4→c3) | Selected route descends ~350px through three magenta `╳`; `…▶[M10]` correctly names c3; b1 row reads `━━3◆╳─╳──╳───╳2▶` — 8 semantic tokens in 12 cells |
| F8 | 140×36 | neon, colour | fixture | `[M14]` f1 (prereqs e1 visible, e2 offscreen) | Fan count `2` correct at f1; `[M11]…` correctly names e2 but is placed bottom-left, ~100 columns from f1 top-right; a 70-cell corridor `◇┄┄╳┄…┄╳╳┄…┄◇` reads as a horizontal rule |
| F9 / F10 | 140×36 | neon, colour | fixture | — | Summary and wave reader for the head-to-head comparison; wave reader shows 13 of 15 members with full `needs [M8], [M9], [M10]` lists in one screen |
| F11 | 140×36 | neon, colour | fixture `reversal` (g1→h3, g2→h2, g3→h1) | `[M1]` g1 | **Renders correctly** — real elbows, two honest crossings, no false bundle. 3 edges cost 4 `◇`, 2 dotted stubs and 3 `╳` |

Ground truth for every topology claim came from `tskflwctl thread graph <ref> --json`, not from the
picture. Two shapes the brief asks for are **unreachable through supported tooling**: `task depend
add` refuses a self-edge (`task … cannot depend on itself`) and refuses a cycle (`dependency cycle:
… run tskflwctl task depend repair`). Reverse and self-edge grammar is therefore defensive-only and
I did not manufacture it by hand-editing frontmatter.

### 4. Comprehension scorecard

Scored from observed use at 140×36 / 100×28 / 72×24 unless noted. ✔ works · ◐ works with effort ·
✘ fails.

| Dimension | Score | Basis |
| --- | --- | --- |
| **Direction** | ✔ | Yellow `▶`/`◀` at the dependent border plus the header's `prerequisite ─▶ dependent` were unambiguous at every size and in both themes; survives colour loss as a glyph |
| **Endpoint ownership** | ◐ | Aliases and grouped aliases were honest every time, but a label can sit 100 columns and 20 rows from the arrowhead it explains (F8), and it uses the same yellow as counts, carets and arrows |
| **Endpoint multiplicity** | ✘ | Correct where drawn, **silently absent** at the densest endpoint in two independent graphs (F1 production fan-in 5, F7 fixture fan-in 3) |
| **Crossings** | ◐ | Glyph is right and never became a junction; but it adopts the focus accent when the selected route passes through, and `╳` sits adjacent to `◇` turns and count digits so the cluster is unparseable |
| **Shared routes** | ◐ | `━`/`◆` correctly mark genuine common-endpoint bundles (verified on a2's fan-out and b1's fan-out); reads as "heavier line" = emphasis rather than "more than one route" |
| **Clipping / panning** | ✔ topology, ◐ reading | No unattributed selected route; offscreen endpoints always named; but at 72×24 partially visible nodes lose their status glyph and border (already owned by `6g8btt5hcgs9`) |
| **Density** | ✘ | 5 / 3 / 1 nodes visible at the three sizes against 43 in the graph; 55–75% of graph rows empty; corridors read as table rules |
| **Navigation congruence** | ✔ | `G1→l→c3→l→d1→l→e1`, `e1→h→d1→h→c1`, `c1→j→c2→k→c1` — every move followed a supplied edge or stayed in-column, matching the projection exactly |
| **Colour independence** | ◐ | Node status, role and focus survive; **selected-route identity does not** (bold only), and the legend still names a colour |
| **Legend burden** | ✘ | 141 cells, truncated at all three sizes; drops the on-screen tokens first |

**Ten-seconds-before-the-legend test (F1).** Without reading the legend I could state the selected
task, its state, that dependencies flow left-to-right, and that there were five layers. I could not
state what it needs or unlocks (I read the focus card), where work converges (the counts are 1-cell
digits inside glyph clusters), or what the long dotted horizontals at the bottom of the canvas were
— my first reading of them was "section rules". After reading the legend, the only thing that
changed was the corridors: `┄ track` reframed them as routes. Nothing else in the legend altered my
reading, which is the argument for shrinking it rather than expanding it.

### 5. Findings

#### H1. The fan count disappears at the densest endpoint, understating convergence · **Status:** fixed

**Observed.** F1, 140×36, production Thread, focus `[G1]`. The connector run into
`add-a-guarded-r…` renders `─────────╳╳━━╳╳▶` with no digit. `thread graph --json` gives that node
(`6g4g8gatbnrs`, alias M25) **five** incoming edges — G1, M7, M24, M13, M14 — and zero outgoing. In
the same frame `add-thread-list…` renders `┌╳2▶` and `preserv…` renders `┌╳2▶`, both correct at 2.
Reproduced independently in F7 on the fixture: `c1-node ──────╳╳▶ d1-node` where d1 has exactly
three prerequisites (c1, c2, c3), while c1 and c2 each show a correct `2`.

The mechanism is visible in `putRouteCount`
(`internal/tui/thread_spatial.go:942-945`): the marker is skipped when the target cell is already a
crossing, overlap, conflict, arrow or waypoint. That guard is the right instinct — it was added so
the digit stops erasing another route's geometry — but "skip" is the wrong resolution. Congestion
is precisely what makes a cell a crossing, and congestion is precisely when the count matters.

**User consequence.** The picture reports fewer converging dependencies than exist, and it does so
*selectively*: the busiest node in the viewport is the one that goes unmarked. A reader scanning for
"where does work converge" — the single question this view should answer better than the wave reader
— is steered away from the true convergence point. This is the one place where the hardened
renderer still says something false.

**Smallest bounded recommendation.** Walk outward along the approach segment for the first cell that
is not a crossing/overlap/conflict/arrow/waypoint and place the digit there (the lane gap is ≥ 8
cells, so a free cell always exists within two or three steps). If no cell is free within the stub,
render the count in the node's own border row instead of dropping it. Never let congestion silence
the marker.

**Resolution:** Fan counts now scan outward along the common endpoint stub, with
a reserved node-border fallback when every safe stub cell is congested;
regressions pin both paths without erasing crossing evidence.

#### H2. Crossing, shared and overlap glyphs wear the selected route's accent, inverting their meaning · **Status:** fixed

**Observed.** F2, 140×36, neon. The connector from `ship-guarded-de…` into the focused
`add-a-guarded-r…` is emitted as a single span
`\x1b[1m\x1b[38;2;234;92;226m─────────╳╳━━╳╳\x1b[38;2;201;211;100m▶`. Magenta-bold covers the route
stroke **and** both crossing pairs **and** the shared bar; only the arrowhead switches to yellow.
Zoomed, the run reads `─ ╳╳ ━ › ╳ ▶`: six glyph classes in seven cells, five of them in the same
colour. F7 shows the same at larger scale — the selected route from a4 descends through three
magenta `╳`.

The legend teaches `magenta=focus`, `╳ cross` and `━/┃/◆ shared`. In the rendering, magenta says
"this is your route" while `╳` says "a different route passes here". Painting the second in the
first tells the reader that the crossing belongs to the selected path.

**User consequence.** The moment a route is worth studying — you selected it — its most
information-dense cells become ambiguous. A reader cannot separate "my route turns/joins here" from
"my route merely passes over someone else's". This directly undercuts the change's premise: the
grammar is honest, but the emphasis channel overrides it.

**Smallest bounded recommendation.** Make accent a property of the *route stroke* only. Keep
`╳`, `≋` and `!` in a neutral colour (gray for crossing/overlap, red for conflict) regardless of
whether a selected route passes through, exactly as `putRouteArrow` already forces yellow for
direction. Optionally keep the shared glyphs accent-eligible, since a bundle containing the selected
route genuinely is part of it.

**Resolution:** Crossings, unrelated overlaps, and renderer conflicts now force
neutral gray and clear focus accent; genuine shared endpoint bundles remain
eligible for focus because they truly contain the selected route.

#### H3. Selected-route identity is colour-only, and the legend names a colour that can be absent · **Status:** fixed

**Observed.** F6, 140×36, `NO_COLOR=1`. The capture contains only SGR `0`, `1`, `2` — no colour at
all. Node status still reads (`✔ ○ ● ◌ ✘`), the external gate still reads (`╔═╗` vs `┌─┐`), and the
focused node still reads (expanded three-line card plus `›`). The selected route, however, is
distinguished **only** by SGR `1` (bold) on box-drawing characters; against the neighbouring plain
connector the difference is marginal in this rendering and disappears entirely in terminals that
implement bold as bright-colour. In the same frame the legend still reads `magenta=focus`.

**User consequence.** For a reader with colour disabled, in a monochrome profile, or with a
red/green or contrast-limited palette, the question "which lines touch the task I am looking at" —
the primary job of the focus mechanic — has no reliable answer, and the legend points at a channel
that is not there. This is the accessibility floor the brief asks about, and it currently fails.

**Smallest bounded recommendation.** Give the selected route a second, non-colour channel: render
its stroke with the heavy set (`━`/`┃`, already implemented for shared) or a distinct dash phase,
and reserve `─`/`│` for unselected routes. Change the legend token from `magenta=focus` to a glyph
demonstration (e.g. `━ focus route`) so it stays true with colour off.

**Resolution:** Focused routes now use heavy box-drawing strokes as a non-color
channel in addition to magenta, and the legend demonstrates the glyph instead of
naming a color.

#### M1. The legend is 141 cells wide and is truncated at every supported size, dropping the glyphs that are on screen · **Status:** fixed

**Observed.** `threadSpatialRoleLegend` renders
`roles  ┌ member  ╔ external gate  magenta=focus  › focus  ▶/◀ direction  2 fan  ┄ track  ◇ turn  ╳ cross  ━/┃/◆ shared  ≋ overlap  ! conflict`
— 141 display cells, measured. Token start columns: `┄ track` 81, `◇ turn` 90, `╳ cross` 98,
`━/┃/◆ shared` 107, `≋ overlap` 121, `! conflict` 132. So:

- 140×36 (F1) ends `≋ overlap  ! conf…`
- 100×28 (F3) ends `◇ turn  …` — `╳` and `━` are dropped while both are visible in the same frame
- 72×24 (F4) ends `▶/◀ directi…` — `2 fan`, `┄ track`, `◇ turn`, `╳ cross`, shared are all dropped
  while `╳`, `━` and a fan digit are all on screen

Meanwhile `≋ overlap` and `! conflict` — the two tokens the truncation protects longest — did not
appear in **any** frame I captured, across the 43-node production Thread, a 15-node/20-edge fixture
built to provoke them, and a deliberate three-edge reversal.

**User consequence.** The legend is unusable at exactly the sizes where the grammar is hardest to
read, and it spends its scarcest resource on grammar the user cannot currently encounter. A reader
at 100 columns sees `╳` and `━` with no way to look them up without resizing.

**Smallest bounded recommendation.** Split it: keep a one-line *always-true* legend of the common
grammar (member/gate, direction, focus, fan, track) that fits at 72; move the collision grammar
(`◇ ╳ ━/┃/◆ ≋ !`) behind the existing `?` help overlay, which already has a Symbols section and
room to spell each one out in a sentence. Drop tokens from the inline legend right-to-left by
*rarity*, not by string position.

**Resolution:** The inline grammar is reduced to member/gate, focus route,
direction, and fan count within 72 cells; uncommon shared/collision/conflict
symbols now have full descriptions in contextual Thread help.

#### M2. Route texture changes mid-path, so one dependency reads as several disconnected fragments · **Status:** fixed

**Observed.** F7 and F11, 140×36. A route that needs a track is drawn as: solid `│` vertical, then a
hollow `◇` turn marker, then a dotted `┄` corridor, then another `◇`, then solid `│` again. The F7
zoom shows clusters like `◇╳┄┄┄◇` and `◇┄┄┄╳◇` floating in otherwise empty space between the node
bands, and F8 contains a 70-cell dotted horizontal spanning nearly the whole viewport. In the
ten-second test I read those corridors as section rules, not routes.

Two compounding effects: `◇` is a *node-shaped* glyph used for a corner, so every turn looks like a
small waypoint station rather than a continuation; and the dotted/solid change means the eye must
re-acquire the path at each turn instead of following a stroke.

**User consequence.** The stated goal is that a reader can "follow one dependency from its actual
prerequisite to its actual dependent … without losing ownership". Ownership is lost not to a
falsehood but to fragmentation: three textures and two punctuation marks per long route. In F11 a
simple three-edge crossing costs 4 diamonds, 2 dotted stubs and 3 crossings — 9 extra marks for 3
edges.

**Smallest bounded recommendation.** Use real elbows (`┘ ┐ └ ┌`) for turns and reserve `◇` for
something that genuinely is not a corner, or drop the turn marker entirely now that conflict
detection catches the case it was defending against. Keep the corridor's subdued treatment as
*colour* (dim) rather than as a different stroke set, so a route keeps one texture end to end.

**Resolution:** Waypoint diamonds and dotted corridor strokes are removed. Long
and reverse dependencies retain one continuous line texture with real light or
heavy elbows while keeping internal track attribution for layout tests.

#### M3. A red `!` reports a renderer invariant failure in the same channel as broken task state · **Status:** fixed

**Observed.** `putRouteConnector` classifies `perpendicular && (existingTurns || incomingTurns)` as
`cell.conflict`, sets `theme.ColorRed` and bold, and the legend calls it `! conflict`
(`internal/tui/thread_spatial.go:860-865`). The source comment names it correctly: "This is a
routing invariant failure, not a legitimate crossing". Elsewhere in the same view red is the
semantic colour for an unreadable or broken task (`drawThreadSpatialNode` paints the box red for
`RoleUnknown`/`GateBroken`). I could not provoke a `!` in any fixture, so this is a design
observation about the treatment, not about its frequency.

**User consequence.** If it ever fires, a red `!` inside the graph will read as "this dependency or
task is broken" — a planning-data problem the user should act on — when it actually means the
layout engine could not route two paths cleanly. That is a renderer self-diagnosis leaking into the
product surface, in the one colour already reserved for real graph damage.

**Smallest bounded recommendation.** Keep the classification (it is what stops the renderer lying)
but present it out-of-band: draw the cell in the ordinary crossing grammar and surface the condition
once in the header (e.g. `· 2 routing conflicts`), or in the focus card, in a non-status colour.
Reserve red inside the canvas for graph health.

**Resolution:** Renderer routing conflicts are neutral rather than task-health
red and are counted explicitly in the spatial header, while a neutral local
marker preserves the failure location.

#### L1. Yellow carries four unrelated meanings · **Status:** fixed

**Observed.** In the neon palette (`38;2;201;211;100`) yellow is used for the direction arrowhead,
the fan count digit, the focus caret `›`, and the offscreen boundary alias `[M11]…` / `…▶[M10]`
(F1, F2, F7, F8). In F8 the sequence `───╳━›2▶` places the caret and the count adjacent in the same
hue, and in F7 `…▶[M10]` sits in that hue at the bottom of the canvas.

**User consequence.** Colour stops carrying category. A reader learning "yellow = where to look"
gets four different answers, and the caret in particular adds a glyph inside a connector run whose
cells are already contested.

**Smallest bounded recommendation.** Keep yellow for direction and multiplicity (both are
"about the edge"), and move the focus caret and boundary aliases to the accent hue, which already
means focus — the aliases only ever annotate a selected route.

**Resolution:** Yellow is now restricted to edge direction and multiplicity.
Selected offscreen endpoint aliases use the focus accent, and the redundant
yellow focus caret is gone.

#### L2. The focus caret sits inside the connector run rather than beside the node · **Status:** fixed

**Observed.** `drawThreadSpatialNode` places `›` at `placement.x-3`
(`internal/tui/thread_spatial.go:1561`). In F1 it renders as a lone chevron floating in whitespace
two cells from the gate box (`───┐  ›  ║ [G1] …`); in F2 and F8 it lands *inside* the incoming
connector (`━›╳▶`, `╳━›2▶`), splitting the stroke one or two cells before the arrowhead.

**User consequence.** Where there is no incoming route the caret reads as a stray glyph; where there
is one it interrupts it at the most contested point. Meanwhile the focused node is already
unmistakable from its expanded three-line card and accent border, so the caret is adding cost
without adding information.

**Smallest bounded recommendation.** Drop the caret, or move it to the node's own top-left corner
cell where it cannot collide with a route.

**Resolution:** The connector-adjacent focus caret is removed; the expanded node
card and heavy accent incident routes carry focus without interrupting route
geometry.

### 6. Task mapping

**Corroborated, already owned — no new finding opened:**

| Evidence | Owner |
| --- | --- |
| 5 / 3 / 1 nodes visible at 140/100/72; 55–75% of graph rows empty; canvas 470×75 for 43 nodes | `6g8btt5hcgs9` (responsive layout) — its AC3 already owns vertical viewport budgeting |
| At 72×24, `-thread-list…` and `✔ add-` lose their status glyph, prefix and border | `6g8btt5hcgs9` AC2, which names this case verbatim |
| Every node label truncated to ~16 chars at every size; no full task name readable in the graph | `6g8btt5hcgs9` AC1 |
| "Too much graph to hold at once" — the underlying reason the wave reader wins the head-to-head | `6g86c7y6hn41` (one-hop focus subgraph) |
| Layout rebuilt per frame and per keypress | `6g8ezj5e51hg` (cache and preflight) AC2 |
| Reaching the spatial view needs `v v` with no hint that it exists | `6g87qn72901g` (alternate-view discoverability) |

**Accepted recommendations needing an owner.** H1, H2, H3, M1, M2, M3, L1 and L2 are all route- and
grammar-scoped and none falls inside the tasks above — those own *how much space the graph gets* and
*how much graph is shown*, not *what a connector cell means*. I propose one bounded new task:

> **Give dense Thread route grammar a legible visual hierarchy** — restore the fan count at
> congested endpoints (H1); make collision grammar colour-independent of focus (H2); give the
> selected route a non-colour channel and a truthful legend token (H3); split the legend into an
> always-fitting inline row plus a `?`-overlay reference (M1); keep one stroke texture per route
> (M2); move the routing-conflict signal out of the canvas's status-red channel (M3); and separate
> the four current uses of yellow (L1, L2). Presentation-adapter only; no change to
> `ThreadGraphProjection`, waves, readiness, health, or persistence.

Sequencing note: H1 and H2 are worth doing before `6g8btt5hcgs9`, because responsive layout will
change *how many* of these cells are visible without changing what they say — and both are small,
local edits to `putRouteCount` and `putRouteConnector`.

### 7. Wireframes

**H1/H2/L2 — the connector into a focused fan-in node.** Observed today (F2, magenta shown as `▓`,
yellow as `▒`):

```
   ║ [G1] completed     ║▓▓▓▓▓▓▓▓▓╳╳━━╳╳▒│ ✔ add-a-guarded-r… │
                          └── all magenta-bold: stroke, crossings and bundle are one colour
                                         └── no count, though 5 edges arrive here
```

Proposed — count restored by walking outward, collision glyphs held neutral, caret dropped:

```
   ║ [G1] completed     ║▓▓▓▓▓▓╳▓╳▓━▓5▒│ ✔ add-a-guarded-r… │
                              │ │  │ └── fan count survives congestion
                              │ │  └──── bundle may keep the accent (it contains your route)
                              └─┴─────── crossings stay gray: not yours
```

**M1 — legend split.** Inline row that fits at 72 columns:

```
roles  ┌ member  ╔ gate  ━ focus route  ▶ direction  2 fan  ┄ track
```

with `? → Symbols` gaining a spelled-out block for `◇ ╳ ━/┃/◆ ≋ !`, where each has room for a
sentence rather than two words.

### 8. Residual questions

- **Real-terminal glyph rendering is unverified.** My frames came from a tmux PTY re-rendered in
  Chrome at FiraCode Nerd Font 15px. `╳` rendered as a bowtie/hourglass noticeably heavier than the
  lines it crosses, which is part of why collisions dominate — but that is font-specific. Whether
  `╳`, `◇`, `≋`, `━`/`┃` and `┄` keep their intended relative weight in the maintainer's actual
  terminal and font needs one look from a real session. If `╳` is lighter there, M2 softens; if
  `━` is not visibly heavier than `─`, H3's proposed fix needs a different channel.
- **Bold-only focus under colour loss** (H3) depends on terminal behaviour I could only reason about:
  some terminals synthesise bold on box-drawing glyphs, others map bold to bright colour (nothing,
  with colour off). Worth confirming in the target terminals before choosing the second channel.
- **`≋ overlap` and `! conflict` were never observed.** I could not construct either through
  supported tooling, and cycles and self-edges are refused by the dependency guard. Whether they are
  reachable at all in a healthy repository is a question for the implementation owner; if they are
  not, M1's recommendation to move them into the help overlay becomes strictly better, and M3 becomes
  low-priority.
- **Maintainer taste: is the fan count the right convergence signal at all?** It is a single digit
  competing for the most contested cell in the frame. An alternative — thickening the arrowhead, or
  annotating convergence in the column heading (`layer 4 · wave 4 · 3 converge`) — would remove the
  H1 class of problem entirely rather than relocating it. I did not test that and would not
  prescribe it.
- **Real-user observation needed on the corridor band.** My reading of the long dotted horizontals as
  "section rules" is one reader's first impression. Whether that generalises, and whether the
  responsive-layout work incidentally fixes it by shrinking the canvas, should be checked with
  someone who has not read this renderer's source.
- **Not evaluated:** live reload under concurrent edits, the Atlas surface, and the summary view's
  own design; all outside this change.
