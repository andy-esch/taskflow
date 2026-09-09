---
schema: 1
id: 6g8bbjgcx9tj
bucket: closed
area: spatial-thread-graph-tui-experience-design-review
date: "2026-09-09"
updated_at: "2026-09-09"
---
# Audit: Spatial Thread graph TUI experience design review — 2026-09-09

> Reviewer assignment: independent product/interaction design reviewer. This document is the review
> brief and the only repository file the reviewer should update.
>
> This is an experience-first design review, not a substitute for the two implementation audits.
> Do not claim visual or interaction quality from source inspection alone. Observe the rendered TUI
> in color through a real PTY, screenshot/GIF, or deterministic rendered-frame harness. If the
> environment cannot provide viewable output, report that as a blocker rather than inventing a
> design verdict.
>
> Finding grammar is exact: use `#### H1. <title> · **Status:** open` (or M1/L1). Leave every new
> finding open for implementation-owner triage.

## Mandatory reviewer sandbox

The source checkout is shared and contains owner work. Reading this brief and creating the isolated
copy are the only operations allowed there. Do not run the TUI, VHS, builds, tests, generators, or
design probes in the source checkout. Substitute this audit's repository-relative path below:

```sh
SOURCE_ROOT="$(git rev-parse --show-toplevel)"
AUDIT_REL="planning/audits/<this-design-audit-file>.md"
SANDBOX="$("$SOURCE_ROOT/scripts/isolated-review-workspace.sh" create \
  --source "$SOURCE_ROOT" \
  --deliverable "$AUDIT_REL" \
  --print-path)"
cd "$SANDBOX"
```

The helper makes an independent no-hardlinks clone, overlays the exact handoff state, and creates a
sandbox-only baseline commit. Perform all interaction, fixture, screenshot, recording, rendering,
and report work there. Do not create further commits. Before transfer, restore every generated GIF,
temporary tape, fixture, source probe, and cache-visible change so the audit is the only diff, then:

```sh
"$SANDBOX/scripts/isolated-review-workspace.sh" verify --sandbox "$SANDBOX"
git diff -- "$AUDIT_REL"
"$SANDBOX/scripts/isolated-review-workspace.sh" transfer --sandbox "$SANDBOX"
```

Leave the sandbox in place and include the helper's isolation/transfer attestation in the report.
If verification or transfer refuses, preserve the sandbox and report the conflict; do not work
around the guard.

## Review purpose

Evaluate whether the merged spatial Thread graph is a compelling, comprehensible first usable
prototype and identify the smallest design changes needed before it becomes a flagship production
TUI feature. Pay equal attention to reaching the experience: the Thread list currently leads with
compact status/health symbols and truncated slugs, while alternate Thread and Atlas presentations
depend on discovering the quiet `v` binding.

This review should answer:

1. Can a first-time user find the relevant Thread and recognize what it represents?
2. Can that user discover which views exist, identify the current view, and predict how to switch?
3. Does the spatial graph communicate dependency direction, state, roles, focus, and immediate
   context faster than the wave reader?
4. Do selection, movement, task opening, and return feel spatially and conceptually coherent?
5. What should be hardened now, and what should remain deliberately experimental?

## Review target and context

Review PR #217 as merged at `c40d09d`, especially the actual TUI experience implemented in
`internal/tui/thread_spatial.go` and the surrounding Thread list/detail, shell, footer, help, and
Atlas presentations. Read these planning artifacts for intent, but do not treat them as evidence
that the experience succeeds:

- `planning/tasks/6g6dw5js81f3-prototype-a-two-dimensional-navigable-thread-graph-view.md`
- `planning/tasks/6g87qn72901g-make-alternate-tui-views-and-thread-identity-discoverable.md`
- `planning/tasks/6g86e10dztpf-make-dense-thread-graph-routes-visually-trustworthy.md`
- `planning/tasks/6g86c7y6hn41-add-a-one-hop-focus-subgraph-mode-to-the-tui-thread-graph.md`
- `planning/tasks/6g86016qvk5d-add-a-reusable-back-action-for-tui-entity-navigation.md`
- `planning/tasks/6g86g03zfj2f-define-consistent-action-targets-for-structured-tui-detail-selections.md`
- ADR-0006's 2026-09-08 spatial-presentation amendment

The source handoff may include an owner-authored lifecycle-only change marking the prototype task
completed. Preserve it; it is not design evidence and is outside this review.

## Required visual evidence protocol

Build the sandbox-local binary and keep color enabled:

```sh
GOCACHE="$SANDBOX/.cache/go-build" just build
unset NO_COLOR
export TERM=xterm-256color
```

Preferred evidence is a live PTY session at approximately 140×36, 100×28, and 72×24 cells. Use the
repository planning space for the large production Thread and `assets/demo-threads` for the smaller
touring-bike story. At minimum observe and capture:

- Atlas default and alternate views, before and after `v`;
- the Thread list with several rows, a long selected identity, and narrow truncation;
- Thread summary, waves, and spatial views, including how the current/available views are signaled;
- spatial focus on a root, fan-out, fan-in, external gate, long edge, and leaf;
- `hjkl`, `Enter`, `y`, `ctrl+o`, `Esc`, `v`, resize, and reload transitions; and
- default plus one alternate color theme if available.

For reproducible image evidence, regenerate only the Threads recording inside the sandbox and
inspect its actual frames with an image-capable tool:

```sh
env -u NO_COLOR PATH="$SANDBOX/bin:$PATH" vhs assets/vhs/threads.tape
```

The current tape reaches the spatial view near the end. For Atlas, use a live PTY or make a
sandbox-only temporary tape derived from `assets/vhs/atlas.tape` that pauses before and after `v`.
Restore generated assets before delivery. If VHS is unavailable, use the agent's PTY/screenshot
facility. If neither path yields output the reviewer can actually see, stop the visual verdict and
document the limitation; source code and stripped ANSI strings are insufficient substitutes.

## Task-based walkthroughs

Perform these as user jobs rather than isolated key checks:

1. **Find the initiative.** Starting outside detail, locate `complete-production-threads` and state
   what it is before opening it. Record which text survives scanning and truncation, what every
   leading symbol appears to mean, and whether selected-row expansion would help or distract.
2. **Discover the presentations.** Without first reading global help, attempt to enumerate and enter
   every Thread and Atlas view. Then consult footer/help. Record whether the UI visually exposes the
   current view, alternatives, order, and `v` action at the moment of need.
3. **Read the graph.** Give the screen ten seconds, then describe dependency direction, active work,
   external gates, selected task, its needs/unlocks, and where work converges. Compare spatial and
   wave comprehension rather than assuming the spatial view wins.
4. **Ride an edge.** Select several nodes and use `h`/`l`; predict each destination before moving.
   Include a skipped-column edge and fan-in/out. Note whether geometry and navigation tell the same
   story.
5. **Inspect and return.** Open a task, copy its name, return to the same graph context, retreat to
   waves, switch views, and return to the list. Record surprises around focus, Back, Esc, and
   parent-versus-child actions.
6. **Degrade the viewport.** Repeat discovery and graph reading at medium and narrow sizes. Identify
   what disappears first, whether the fallback is honest, and whether identity remains primary.

## Design lenses to challenge

- **Information hierarchy:** recognizable Thread identity before codes; status, graph health,
  progress, description, and goal ordered by decision value.
- **Wayfinding and view discovery:** current location, available presentations, cycle order, and
  reversible exits visible without memorized keys. Evaluate chips/tabs/segmented labels or a better
  alternative; do not assume `v` must disappear as the fast binding.
- **Focus and selection:** semantic state color remains stable while selected state is unmistakable;
  expanded cards or boxes add useful detail without causing disorienting layout shifts.
- **Graph comprehension:** edge direction, crossings, roles, waves/layers, long jumps, and local
  context are visually honest. Separate route correctness from aesthetic preference.
- **Interaction mapping:** `hjkl` behavior matches what the picture promises; open/copy/return/back
  actions act on the object the user believes is selected.
- **Responsive behavior:** wide, ordinary, and narrow terminals retain identity and next action;
  fallbacks do not look like renderer failures.
- **Accessibility:** color is reinforced by glyph/label/shape, contrast is adequate, text and Unicode
  remain legible, and animation/expansion is not the sole source of context.
- **Consistency and extensibility:** the treatment fits existing TUI language and can serve Atlas or
  future multi-view entities without a Thread-specific shell abstraction.
- **Delight versus density:** judge whether the status legend, magenta incident routes, yellow
  arrows, expanded selected node, and inspector feel coherent, noisy, flat, or redundant.

## Deliverable

Update only this audit. Preserve the brief and add:

1. an executive verdict: compelling prototype, promising but blocked, or direction needs revision;
2. an evidence matrix listing terminal size, theme, scenario, capture method, and observed result;
3. a concise journey map for find → open → discover view → comprehend → navigate → inspect → return;
4. prioritized findings using exact audit grammar, each with observed evidence, user consequence,
   and the smallest design recommendation;
5. compact text wireframes for any proposed Thread row/focused card and view switcher, including
   wide and narrow variants;
6. a mapping from each recommendation to the existing follow-up task that owns it, or a clearly
   bounded candidate task only when none fits; and
7. residual questions that require maintainer taste rather than pretending they are objective bugs.

Do not edit implementation, fixtures, tapes, tasks, ADRs, or the other reviewers' audits. Do not
close this audit or pre-resolve findings.

## Findings

Every finding below is anchored to a rendered frame captured in the sandbox at a stated cell size.
Frame ids (`w2-04`, `d1-06`, …) refer to the evidence matrix in the reviewer report.

#### H1. Thread and Atlas alternate presentations are not visible before the user knows to switch them  · **Status:** tracked by 6g87qn72901g

**Surface:** Threads list footer / `?` overlay / `ctrl+p` palette / tab strip | **Component:** tui shell
**Effort:** M · **Urgency:** acute

Observed at 140×36, neon dark. On the Threads list the footer reads
`: cmd · a atlas · / filter · E editor · [ ] tabs · l / ⏎ detail · z full · ? help · q quit`
(`w1-02`) — no `v`, no view name. Pressing `v` there changes nothing and says nothing (`w3-01`,
identical to `w1-02` pixel-for-pixel). The Threads `?` overlay documents `j/k`, `g/G`, `d/u`,
`l/⏎`, `h`, the whole status/health symbol vocabulary and three notes, and never mentions `v` or
that alternate views exist (`w3-02`, `w3-03` after scrolling to the end of the overlay). The
command palette — the shell's own stated "discovery surface" — answers `view` with `:overview`
plus fuzzy entity slugs (`w8-02`) and `spatial` with three audit/research slugs (`w8-03`); no
view-switch command is indexed at all. `v` first appears only after `l` focuses the detail pane,
as the trailing hint `v topology` (`w1-04`).

The Atlas partially solves the location half of this problem: its tab strip renders `atlas work` /
`atlas spaces` as the tab label (`a1-01`, `a1-03`), its body header repeats `atlas · spaces · 3
spaces · order name ↑`, and its `?` overlay lists `v — cycle this screen's alternate views` under
Global plus a `both: v switch view` line (`a1-04`). It still does not show which alternate Atlas
views exist until the user cycles or opens help. The Thread detail is weaker: its tab remains
`threads` in all three presentations and neither the current view nor the available set is visible
at the shell level.

**User consequence.** The flagship feature of this change is reachable only by guessing an
unadvertised key while the detail pane happens to hold focus. A first-time user who does the
reasonable thing — look at the footer, press `?`, then try the palette — is told three times that
the spatial graph does not exist.

**Recommendation (smallest).** Generalize the Atlas's useful current-view indicator into a shared
shell contract that also exposes the available set. Render the current view in the tab strip (for
example, `threads · spatial` or `atlas · spaces`) and put a segmented chip line on the detail pane's
top border showing every presentation with the current one marked (wireframe §5.2). Keep `v` as the
fast cycle, add the contextual `v` row to help, and index view names as palette commands. No new
binding is required, and Atlas is part of the fix rather than merely the precedent.

**Resolution:** Accepted. The shared view-discovery task now treats Atlas as a
partial current-view precedent and requires both Atlas and Thread detail to
expose the available presentation set.

#### H2. The graph canvas clips nodes and wave labels mid-glyph, with no off-canvas affordance  · **Status:** tracked by 6g8btt5hcgs9

**Surface:** `internal/tui/thread_spatial.go` canvas viewport | **Component:** tui / spatial graph
**Effort:** M · **Urgency:** acute

Two distinct clipping defects, both visible at the review's largest size.

*Wave labels and nodes are clipped by arbitrary cell panning.* Whenever the canvas is horizontally scrolled, the
leftmost column header loses exactly its first two characters: `ve 6` (`w2-02`), `ve 5` (`w2-03`),
`ve 4` (`w2-04`), `ve 3` (`w2-05`, `w2-06`), `ve 7` (`w5-06`), and `gate` where the label was
`wave 9 + gate` (`w6-04`). At scroll origin the same label renders in full with a four-cell left
indent (`w6-01`: `wave 1 + gate`). Source reconciliation after the visual review showed that headers
and nodes actually share the same `3 + column*stride` origin. The failure is instead that one
cell-level `panX` is applied to the combined canvas, so it can bisect both a label and its node
column. The rendered evidence remains valid; the review's original different-origin diagnosis did
not.

*Nodes are cut mid-box.* At the right edge, `preserve-portab…` and `pin-thread-docu…` are drawn
with top and bottom borders running into the frame and no right border at all (`crop-right`, a 6×
enlargement of `w2-02`). At the left edge in `w6-04` four nodes read `a-guarded-r…`,
`-repository…`, `ne-the-thre…` and `-thread-fro…` with no left border — the status glyph and the
first characters of every slug are gone. At 72×24 the same thing happens on both edges at once
(`n1-04`: `d-thread-list…`, `eserve-cohere…`; reproduced through a live tmux resize in `e1-03`).
Nothing anywhere in the frame indicates that waves 1–5 and 11–13 exist off-canvas.

**User consequence.** The most common state of the production Thread's graph looks like a broken
renderer rather than a scrolled viewport. The cut characters are the identifying prefix of the
slug and the status glyph, so an edge-column node cannot be identified or classified at all, and
the user has no cue that panning would reveal more.

**Recommendation (smallest).** Pan and clip on complete node/layer boundaries rather than arbitrary
cells, render a boundary gutter instead of a silently bisected node, and add a one-line rail above the columns —
`◀ waves 1–5 │ wave 6 │ … │ waves 9+10 │ waves 11–13 ▶` (wireframe §5.3). Align the wave-label

**Resolution:** Accepted with corrected diagnosis. Headers and nodes share an
origin; the responsive-layout task owns complete-unit panning, boundary gutters,
and off-canvas orientation, while route behavior at that boundary remains
coordinated with 6g86e10dztpf.

#### H3. Long dependency edges render as unattributed full-width horizontal rules  · **Status:** tracked by 6g86e10dztpf

**Surface:** spatial edge router | **Component:** tui / spatial graph
**Effort:** M · **Urgency:** acute

In the empty lower lanes of `w2-05` there are two grey horizontal lines spanning the entire
viewport width with no arrowhead, no endpoint, and no node at either end (`crop-empty`, a 1.5×
enlargement). They are long dependency routes passing through vacant corridors, but every visual
property they have — full bleed, uniform weight, sitting in the gap between node rows — is the
property of a table rule. The same line at y-row 213 appears in `w2-02` (grey), `w6-04` (grey) and
`w2-03` (magenta, because it touches focus); only the recolouring reveals that it is an edge at
all. In `crop-mid3` a magenta long edge and a magenta vertical route cross with no jump or bridge
marker, so at the intersection the two focus-coloured routes are indistinguishable from a junction.

**User consequence.** The user is asked, in ten seconds, to read dependency direction from the
picture. The picture contains full-width lines that are either separators or dependencies, and
crossings that are either crossings or joins. Both readings are available and the graph gives no
way to choose, which is worse than omitting the edge.

**Recommendation (smallest).** Give corridor routes a distinct de-emphasised style (dashed or
half-intensity) plus an endpoint tick where they leave and re-enter a node column, and draw a jump
arc at crossings. Where a route would span the full viewport with both endpoints off-canvas,
replace it with a boundary stub carrying the neighbour's code (`… ▸ [M12]`) rather than a rule.

**Resolution:** Accepted. Dense-route hardening will test crossings, endpoint
stubs, waypoints, and skip routes against the invariant that rendered geometry
cannot invent connectivity.

#### M1. Thread list rows spend their width on codes and mangle the only thing that identifies a Thread  · **Status:** tracked by 6g87qn72901g

**Surface:** Threads list rows | **Component:** tui / thread list
**Effort:** M · **Urgency:** soon

At 140×36 the row is `› ● g✓/v✓  ▶0 ✓6 ×1  complete-production-threads  De…` (`w1-02`). Roughly 24
of the ~63 columns available to the list pane are consumed before the name starts, and the trailing
description column is three cells wide — it renders the literal string `De…`. The second row is
already middle-elided at this size: `tool-owned-actio…le-sub-entities`. At 100×28 (`m1-01`) the
description column is gone and *both* identities are destroyed: `complete-p…n-threads` and
`tool-owned…–entities`, while `g✓/v✓ ▶0✓6×1` survives intact.

**User consequence.** Walkthrough 1 — "locate `complete-production-threads` and say what it is
before opening it" — fails at ordinary terminal width. The text that survives truncation is the
part that carries no identity, and the description column is noise at every size the review
covered.

**Recommendation (smallest).** Lead with a human title, right-align the compact codes, and drop
the description column entirely below ~110 columns rather than rendering a stub; expand only the
selected row into a fixed-height three-line card so row pitch never shifts (wireframe §5.1).

**Resolution:** Accepted. Thread identity-first rows and bounded focused cards
are explicit acceptance criteria.

#### M2. Spatial node labels are capped at ~15 characters regardless of the width available  · **Status:** tracked by 6g8btt5hcgs9

**Surface:** spatial node rendering | **Component:** tui / spatial graph
**Effort:** S · **Urgency:** soon

In `d1-03` (140×36, the four-node-wide touring-bike Thread) the graph occupies about 60% of the
frame and leaves roughly 40 columns of empty canvas to the right, yet every node still truncates:
`inspect-touring…`, `true-and-tensio…`, `run-a-loaded-sh…`. The node's inner text budget is fixed
at ~17 cells whether three columns are shown or five. That the width exists is demonstrated inside
the same app at the same size: the `f` follow picker renders the identical slugs in full —
`ship-guarded-dependency-mutations-and-graph-queries`,
`preserve-coherent-atlas-summaries-across-transient-per-space-refresh-failures` (`w7-04`).

**User consequence.** Reading the graph becomes reading the focus inspector: the only place a full
name appears is the panel at the bottom, so the user navigates blind and confirms afterwards,
which is the opposite of what a spatial view is for.

**Recommendation (smallest).** Derive the per-node text budget from viewport width divided by the
number of columns actually shown, with a floor; when the budget is short, wrap the slug across two
lines (head + tail) instead of a hard ellipsis so the distinguishing suffix survives.

**Resolution:** Accepted. The responsive-layout task owns adaptive label budgets
and an honest code/legend or multi-line fallback when full identity cannot fit.

#### M3. The vertical viewport admits partial node slots against the focus inspector boundary  · **Status:** tracked by 6g8btt5hcgs9

**Surface:** spatial focus inspector | **Component:** tui / spatial graph
**Effort:** S · **Urgency:** soon

In `d1-06` the node `● receive-front-r…` is drawn with its top and side borders intact and its
bottom border replaced by the inspector's top border (`crop-overlap`, 4× enlargement) — the box is
left open and the row is half-hidden. The same happens to `decouple-thread…` and
`model-graph-own…` in `w6-01`. Source reconciliation after the review showed that
`renderThreadSpatial` already reserves five inspector rows through `fixedRows := 8`; the inspector
is not simply painted over an unbudgeted canvas. The remaining defect is that the vertically panned
window is cell-based and may end on part of a five-row node slot immediately above the separately
appended inspector. The rendered evidence remains valid; the original compositing diagnosis did not.

**User consequence.** A whole lane of real work can be partially or wholly hidden behind the panel
that is supposed to explain the selection, with no scroll cue and no indication that anything is
missing.

**Recommendation (smallest).** Keep the existing inspector reservation, but calculate the visible
window and vertical pan in complete node-slot units. A node must either fit above the inspector or be
represented by an explicit boundary cue; it must never appear to lose its bottom border into the
inspector.

**Resolution:** Accepted with corrected diagnosis. Inspector space is already
reserved; the task owns slot-aware vertical panning so a partial node cannot
collide visually with that boundary.

#### M4. Six statuses are carried by four distinguishable glyphs  · **Status:** tracked by 6g87qn72901g

**Surface:** spatial status legend and node glyphs | **Component:** tui / design tokens
**Effort:** S · **Urgency:** soon

The legend renders `● active · ● next · ○ ready · ✓ done · ◌ deferred · ✗ deprecated`
(`crop-legend`, a 3× enlargement of the legend rows of `w2-02`). `active` and `next` use the
identical filled-circle glyph and differ only in hue (olive vs blue); `ready` and `deferred` differ
only by a solid versus dotted ring, which at a 14px cell is close to a rendering difference. So of
six states, four are separated by colour alone or near-alone. The role encoding is better: member
nodes use a single border and external gates a double border, preserved under focus (`crop-node`,
`crop-focusnode`), which is a genuinely accessible shape channel.

**User consequence.** For a colour-blind user, or on a projector, the two states that matter most
for "what is being worked on now" (`active`) and "what is next" (`next`) are the pair that cannot
be told apart without colour.

**Recommendation (smallest).** Give the six states six glyphs — e.g. `◐ active`, `▸ next`,
`○ ready`, `✓ done`, `◌ deferred`, `✗ deprecated` — keeping the existing hues as reinforcement.

**Resolution:** Accepted. The discoverability work owns a shared status
vocabulary that does not rely on color alone.

#### M5. Semantic `h`/`l` dependency walks are mislabeled, ambiguous at branches, and silent at dead ends  · **Status:** tracked by 6g86c7y6hn41

**Surface:** spatial navigation | **Component:** tui / spatial graph
**Effort:** M · **Urgency:** soon

The footer says `hjkl node`, even though the intentional interaction contract makes `h`/`l` walk
actual prerequisite/dependent edges rather than adjacent screen cells. From `[M21]
add-thread-list-and-detail-views-to-the-tui` in row 1 (`w2-03`), `h` lands on
`[M20] wire-thread-projections…` in row 2 of the previous column (`w2-04`) — a prerequisite, chosen
from the two the inspector lists (`needs [M20] …, [M5] …`) with no visible rule and no way to reach
the other. From `[M20]` (`needs [M17], [M18]`) the next `h` lands on `[M18]` (`w2-05`), the nearer
wave, so the tie-break is plausible but unstated. Separately, `j` on the single-node wave 5 changes
nothing and reports nothing: `w2-06` is identical to `w2-05`. `g` on an unscrolled canvas is
likewise a silent no-op (`w2-01` vs `w2-02`) although the footer advertises `g/G top/bottom`.

**User consequence.** Walkthrough 4 — "predict each destination before moving" — is not winnable.
The user cannot tell whether a dead key means "no neighbour", "not bound here", or "the app is
stuck", and cannot reach a node's second prerequisite at all.

**Recommendation (smallest).** Preserve graph-first navigation and name it accurately in the footer
(`h/l prereq · dependent · j/k lane`). Briefly flash the traversed edge, expose when more than one
edge-neighbor is available and provide a deterministic way to select among them, and emit a one-line
status message ("no prerequisite in this direction") instead of absorbing the key.

**Resolution:** Accepted in part with the interaction direction settled: h/l
remain semantic prerequisite/dependent walks. The focus task owns branch choice
and accurate navigation labels; reusable unsupported-action feedback is
coordinated with 6g86g03zfj2f.

#### M6. A column is not a wave, but it is labelled as one  · **Status:** tracked by 6g86e10dztpf

**Surface:** spatial column headers | **Component:** tui / spatial graph
**Effort:** S · **Urgency:** soon

`w6-01` labels five adjacent columns `wave 1 + gate`, `waves 1+2`, `waves 2+3`, `wave 3`, `wave 4`
— wave 3 is named by two adjacent columns, and waves 1 and 2 each straddle two. `w2-02` shows
`wave 9 + gate` immediately followed by `waves 9+10`. The layout's horizontal axis is a layer
position, which is a legitimate and different thing from a wave, but the header claims otherwise.

**User consequence.** The header block promises `focus wave 8 + gate` and the columns promise
waves, so the user reads horizontal distance as wave distance; two nodes in the same wave sitting
in different columns then read as sequential when they are concurrent.

**Recommendation (smallest).** Label columns as positions (`col 3 · waves 2–3`) or collapse to one
column per wave, and keep exact wave membership on the node badge and in the inspector.

**Resolution:** Accepted. Route/layout hardening must label layout layers
honestly while preserving actual wave membership on nodes and in the inspector.

#### M7. Leaving the spatial view leaves the pane in an alternate presentation with no indicator or control  · **Status:** tracked by 6g87qn72901g

**Surface:** Thread detail view state / list focus | **Component:** tui shell
**Effort:** S · **Urgency:** soon

`Esc` from the spatial view returns to the wave reader and carries the selection with it — `[G1]`
is still the cursor row under **External gates** (`e1-06`), which is good. A second `Esc` returns
focus to the list, and the detail pane stays in the topology view while the footer reverts to the
list hints (`e1-07`): no `v`, no view name, no way to know the pane is not showing the summary. The
`view: topology · member waves plus bounded dependencies` line inside the body is the only trace,
and it is the third row of a key/value block.

**User consequence.** The user is left in a mode they cannot see, cannot name and cannot leave
without rediscovering H1.

**Recommendation (smallest).** Render the view chip strip under list focus too (it is pane state,
not focus state), and either bind `v` there or reject it with a message naming where the binding
lives.

**Resolution:** Accepted. Current and available presentations must remain
visible under list and detail focus for Thread and Atlas surfaces.

#### L1. The static 40/60 pane split can waste list space while constraining Thread detail  · **Status:** tracked by 6g8btt5hcgs9

**Surface:** shell pane split | **Component:** tui shell
**Effort:** S · **Urgency:** later

Source reconciliation corrected the review's original “50/50” label: the implementation allocates
two fifths to the list and three fifths to detail. At 140×36 the list pane nevertheless holds two
short rows and 32 blank lines while the detail pane wraps
constantly: every frontier row breaks so that its `queued/clear` state lands alone on the next
line, and the goal breaks mid-phrase (`w1-02`). Pressing `z` proves the content fits — the same
summary renders one row per item with no wrapping at all (`w9-01`), and the footer there correctly
leads with the mode name `full-screen`.

**Recommendation (smallest).** Make the split responsive to the available width and the active
content budget, within stable bounds, rather than changing to another fixed ratio. The list must
retain recognizable Thread identity while surplus width can move to detail when its content would
otherwise wrap heavily.

**Resolution:** Accepted after correcting the factual premise from a 50/50 split
to the current static 40/60 split. The task owns responsive bounded pane
allocation rather than another fixed ratio.

#### L2. At 72 columns the chrome truncates exactly the parts that say where you are and how to leave  · **Status:** tracked by 6g87qn72901g

**Surface:** tab strip and footer | **Component:** tui shell
**Effort:** S · **Urgency:** later

At 72×24 the list pane is dropped and the detail pane opens focused, which is a reasonable
fallback. What is not reasonable is what the surviving chrome loses. The tab strip renders
`[isolated-review…] · atlas · overview · tasks · epics · …` — the current tab, `threads`, is off
the right edge, so nothing on screen says which entity you are looking at (`n1-01`). The footer
ends `… · R raw/pre…`, cutting `h/esc back`: the exit affordance is the casualty of truncation. In
the spatial view the header truncates to `prerequisite ⟶ depe…` and the legend to `magenta lines
touch focus › fo…` (`n1-04`), while both legend rows still consume 2 of 24 rows.

**Recommendation (smallest).** Below ~90 columns render the tab strip as `‹ threads ›` (current tab
only), order footer hints so the exit key is never the one dropped, and collapse the two legend rows
to a single `? legend` hint.

**Resolution:** Accepted. Narrow view discovery must retain current location,
recognizable identity, and the exit affordance before secondary chrome.

#### L3. Stacked fan-in arrowheads collide into an ambiguous glyph pair  · **Status:** tracked by 6g86e10dztpf

**Surface:** spatial edge arrowheads | **Component:** tui / spatial graph
**Effort:** S · **Urgency:** later

Where several routes converge on one node the arrowheads are laid side by side in the same cell
run and read as one compound glyph: `›▶` immediately left of `add-thread-li…` (`crop-mid3`) and of
`run-a-loaded-sh…`, which has three incoming edges (`d1-06`). A lone `›` also precedes the focused
node in `w2-02`, so the same mark means "one edge arrives here" and "several do".

**Recommendation (smallest).** Draw one arrowhead at the node boundary and, when the in-degree is
greater than one, annotate the entry with a count or give each route a distinct vertical entry row.

**Resolution:** Accepted. Dense-route hardening owns unambiguous fan-in entry
and arrowhead treatment.

### What is already working

Recorded so the hardening tasks do not regress it:

- **Return is coherent.** `Enter` on a spatial node opens the task on the tasks tab with a
  breadcrumb footer `‹ ctrl+o complete-production-threads (1)` (`w5-02`), and `ctrl+o` restores the
  spatial view with the same node focused and the same viewport (`w5-06`).
- **State survives reload and resize.** `r` keeps the spatial view and focus `[G1]` (`e1-05`); a
  live tmux resize 140→100→72→140 reflows cleanly with no stale frame or garbage (`e1-02`, `e1-03`,
  `e1-04`).
- **Focus causes no layout shift.** The focused node expands to three lines but rows keep a fixed
  pitch, so moving the cursor does not reflow the grid (`w2-02` vs `w2-04`).
- **Role encoding is a shape, not a colour.** External gates keep their double border under focus
  (`crop-focusnode`) and unfocused (`crop-node`).
- **`y` is honest.** It copies the focused node's slug and confirms in the footer — `✓ copied slug:
  ship-guarded-dependency-mutations-and-graph-queries` (`w4-02`), verified against `pbpaste`.
- **The small graph is genuinely good.** At 140×36 the touring-bike Thread reads left-to-right in
  one glance, with the outstanding external gate visibly the thing holding up the shakedown ride
  (`d1-03`). This is the proof the direction is right.
- **Themes adapt.** `catppuccin` (`t1-01`) and a light terminal background (`t2-01`) preserve every
  semantic distinction; edge routes measure 5.5:1 contrast on dark (`#929399` on `#1c1e28`) and
  3.6:1 on light (`#7b7e8b` on `#eef0f4`) — above the 3:1 non-text floor, but the light theme has
  little headroom.

## Task coordination

Maintainer triage on 2026-09-09 reconciled the visual findings with the implementation and filed
the one missing bounded owner. The mappings below are the intended scopes; shared rows identify a
deliberate seam, not permission for two tasks to reimplement the same behavior.

| Finding | Owning task | Why it fits |
| --- | --- | --- |
| H1, M7 | `6g87qn72901g` make-alternate-tui-views-and-thread-identity-discoverable | AC 4 explicitly covers both Thread detail and Atlas. Atlas is a partial current-view precedent, not a completed solution: the shared contract must expose the available set in both surfaces. |
| M1 | `6g87qn72901g` | AC 1 (rows lead with a human-identifiable title, not status syntax and a truncated slug) and AC 3 (bounded multi-line focused card). |
| M4 | `6g87qn72901g` | AC 2 — compact symbols "self-explanatory or supported by a nearby legend … without making color the only carrier of meaning". Extends the same shared status vocabulary from the list to the graph nodes. |
| L2 | `6g87qn72901g` | AC 6 covers narrow identity and alternate-view orientation; shell chrome and the exit affordance are explicitly part of the shared view-discovery contract rather than a Thread-row-only concern. |
| H2 (node/header boundary), M2, M3, L1 | `6g8btt5hcgs9` make-spatial-thread-layout-responsive-to-available-space | Owns complete node/layer viewport units, adaptive identity budgets, slot-aware vertical panning, and responsive pane allocation. It does not own edge geometry. |
| H2 (edge boundary), H3, L3 | `6g86e10dztpf` make-dense-thread-graph-routes-visually-trustworthy | Owns whether routes, crossings, fan-in/out, endpoints, and selected incident paths remain truthful at and across the viewport boundary. It consumes the responsive task's visible-window contract rather than choosing node visibility. |
| M5 | `6g86c7y6hn41` add-a-one-hop-focus-subgraph-mode-to-the-tui-thread-graph | AC 4 requires an explicit interaction contract. The accepted direction preserves semantic edge walks; this task owns neighbor choice and focus-mode consistency, not a switch to geometric `h`/`l`. |
| M6 | `6g86e10dztpf` | AC 4 ("external gates … without … misrepresenting member waves") plus AC 5's deterministic layout for skipped-layer cases; the header is the place the layer/wave distinction leaks. |
| M5 dead-key silence, and `m` in the wave reader (silent no-op, `w7-03`) | `6g86g03zfj2f` define-consistent-action-targets-for-structured-tui-detail-selections | AC 3 owns unsupported-action feedback. The focus task owns navigation semantics; this task supplies the reusable shell response when an action has no valid target. |
| Positive regression guard for `Enter`/`ctrl+o` (`w5-02`, `w5-06`) | `6g86016qvk5d` add-a-reusable-back-action-for-tui-entity-navigation | AC 1 already pins this behaviour; recording it here gives the task its first observed baseline. |

## Wireframes

### 5.1 Thread list row

Current, 140×36 (`w1-02`) — 24 columns of codes, then a middle-elided slug, then a 3-cell stub:

```
› ● g✓/v✓  ▶0 ✓6 ×1  complete-production-threads       De…
  ● g✓/v✓  ▶0 ✓1 ×0  tool-owned-actio…le-sub-entities  Ma…
```

Proposed, wide (≥110 cols) — identity first, codes right-aligned, selected row expands into a
fixed-height slot so pitch never shifts:

```
› Complete production Threads                    ● in-progress   28/36   6 ready
    complete-production-threads · graph ✓ · projection ✓ · 1 blocked
    Deliver production Threads through the CLI preview and a usage-informed TUI
  Tool-owned actionable sub-entities              ● in-progress    3/4   1 ready
```

Proposed, narrow (<90 cols) — the title is the last thing to be cut; codes fold to line two of the
selected card only:

```
› Complete production Threads            ● 28/36
    graph ✓ · proj ✓ · 6 ready · 1 blocked
  Tool-owned actionable sub-entities      ● 3/4
```

### 5.2 View switcher

Current: the tab strip says `threads` in all three views; the only current-view text is row 3 of a
key/value block (`view: topology · member waves plus bounded dependencies`) and the footer names
the *destination* (`v spatial`).

Proposed, wide — segmented chips on the detail pane's top border, mirroring the Atlas's existing
`atlas work` / `atlas spaces` tab labelling:

```
[space] · atlas · overview · tasks · epics · threads · spatial · audits · research
╭ complete-production-threads ─────────── summary · waves · ▏SPATIAL▕ · v ⟳ ──╮
│ complete · focus wave 8 + gate · prerequisite ⟶ dependent · graph ✓ proj ✓  │
```

Proposed, narrow (<90 cols) — position replaces the enumeration, identity still first:

```
‹ threads ›
╭ complete-production-… ── SPATIAL 3/3 ‹v› ╮
```

### 5.3 Spatial canvas: off-canvas rail, node budget, reserved inspector

Current, 140×36 (`w2-02`, `w6-04`) — the leftmost header is cut to `ve 6`, the edge nodes lose a
border, and the last partial node slot collides visually with the inspector boundary:

```
ve 6              wave 7          wave 8 + gate     wave 9 + gate      waves 9+10
┌───────────────┐ ┌──────────────┐ ╔══════════════╗                    ┌──────────
│× add-usage-in…│ │✓ add-thread-…│ ║✓ ship-guarde…║ ┌───────────────┐  │○ preserve
└───────────────┘ └──────────────┘ ╚══════════════╝ └───────────────┘  └──────────
╭ focus [G1] ──────────────────────────────────────────────────────────────────╮
```

Proposed — a boundary rail, node text budgeted from the width actually free, and a slot-aware
vertical viewport so no partial node row is admitted above the already-reserved inspector:

```
◀ waves 1–5 │ wave 6 │ wave 7 │ wave 8 + gate │ wave 9 + gate │ waves 9+10 │ 11–13 ▶
┌─────────────────────┐  ┌────────────────────────┐  ╔═════════════════════════╗
│ ✗ add-usage-informed│─▸│ ✓ add-thread-list-and- │─▸║ ✓ [G1] ship-guarded-    ║
│   -thread-views     │  │   detail-views-to-tui  │  ║   dependency-mutations  ║
└─────────────────────┘  └────────────────────────┘  ╚═════════════════════════╝
   ⋯ long route from [M12] ────────────────────────────────────────▸ [M25] ⋯
────────────────────────────────────────────────────────────────────────────────
╭ focus [G1] ── gate · done · unlocks [M25] ───────────────────────────────────╮
```

Key changes shown: `◀ ▶` rail naming the hidden waves; whole-node clipping with a dimmed gutter
column; two-line node labels so the distinguishing tail survives; corridor routes drawn dashed with
named endpoint stubs instead of a full-bleed rule; a solid divider marking the canvas floor above
the reserved inspector band.

## Reviewer report

### Executive verdict

**Promising but blocked.**

The direction is right and there is one frame that proves it: on the touring-bike Thread at 140×36
(`d1-03`) the spatial view answers "what is holding up the shakedown ride" in about two seconds —
the outstanding external gate sits low-left, its magenta route runs into the rack work, and the
converging fan into the loaded ride is visible as geometry. The wave reader cannot do that; it can
only list. Selection stability, `ctrl+o` return, reload and live resize are all already sound, and
the role encoding is a genuine shape channel rather than a colour.

It is blocked on two things, both of which show up on the repository's own production Thread at
every size tested. First, nobody can find it: the footer under list focus, the `?` overlay and the
command palette all omit the view cycle (H1), so the feature is reachable only by prior knowledge.
Second, once found, the picture is not yet trustworthy on a 37-member graph: nodes and wave labels
are cut mid-glyph at the viewport edges with no off-canvas cue (H2), long routes are drawn as
full-width rules indistinguishable from separators (H3), and node labels are capped at ~15
characters even when 40 columns of canvas sit empty (M2), which pushes all reading into the focus
inspector.

None of these is a direction problem. Atlas provides a useful current-view precedent for H1 but
shares the undiscoverable-alternatives gap. The fixes now have explicit owners: responsive layout
owns complete viewport units and identity budgets (H2/M2/M3/L1); dense-route hardening owns edge
truthfulness (H2/H3/M6/L3); and shared discoverability owns Thread and Atlas wayfinding
(H1/M1/M4/M7/L2). The maintainer has also settled M5's central taste question: `h`/`l` remain
semantic prerequisite/dependent walks and must be labelled and explained as such. One-hop focus
remains a complementary local lens, not a replacement for the full graph.

### Isolation attestation

| Field | Value |
| --- | --- |
| Workspace | `/private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.WA4Zey` |
| Resolved Git directory | `/private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.WA4Zey/.git` (independent `--no-hardlinks` clone; no `git worktree`, no symlink) |
| Source root | `/Users/andyeschbacher/git/andy-esch/taskflow` |
| Baseline commit | `a7714ef0464c41daaaf774b5e11bd11208584fc8` — "chore: capture isolated review baseline", on top of `c40d09d` (PR #217 merge) |
| Captured source blob | `79eaef2e896aefbe78535f9c1d299d655331a9ac` |
| Source fingerprint | `b66b9801f3ba7f32f959686f75e2335cfb8d45e7` |
| Deliverable | `planning/audits/6g8bbjgcx9tj-2026-09-09-spatial-thread-graph-tui-experience-design-review.md` |
| Commits created by reviewer | none beyond the helper's baseline |
| Probes restored before transfer | tape/screenshot scratch directory `.review/`, VHS output GIF, sandbox Go build cache, `bin/tskflwctl` |
| Verify / transfer result | recorded below |

The owner-authored lifecycle-only change to
`planning/tasks/6g6dw5js81f3-prototype-a-two-dimensional-navigable-thread-graph-view.md` was
carried into the sandbox by the overlay and left untouched.

### Evidence matrix

Capture method for every row: sandbox-local `bin/tskflwctl` (built with `GOCACHE="$SANDBOX/.cache/go-build" just build`,
`v0.20.0-29-ga7714ef`) driven through a real PTY by VHS 0.11.0 with `TERM=xterm-256color` and
`NO_COLOR` unset, recorded as PNG frames and read back as images. Cell geometry was calibrated by
`stty size` inside the recorder (140×36 = 1291×586 px, 100×28 = 932×472 px, 72×24 = 674×407 px at
FontSize 14, Padding 0). Enlargements are `ffmpeg` nearest-neighbour crops of those same frames.

| Frame | Size | Theme | Scenario | Observed |
| --- | --- | --- | --- | --- |
| `w1-01`…`w1-08` | 140×36 | neon dark | launch → `:threads` → `l` → `v`×4 | 3-view cycle confirmed: summary → topology → spatial → summary |
| `w1-02` | 140×36 | neon dark | Thread list, both rows | 24 cols of codes before identity; row 2 middle-elided; description = `De…`; footer has no `v` |
| `w1-04` | 140×36 | neon dark | detail focused (summary) | content unchanged from unfocused; only border colour and footer differ; `v topology` first appears |
| `w1-05` | 140×36 | neon dark | topology / wave reader | `view: topology …` is the only current-view label; footer names destination `v spatial` |
| `w2-01`…`w2-08` | 140×36 | neon dark | spatial: `g`, `h`×3, `j`, `l`×2 | `g` no-op; `h` walks to a prerequisite across lanes; `j` on a single-node wave silent |
| `w2-05` / `crop-empty` | 140×36 | neon dark | empty lower lanes | two full-width grey rules, no arrowheads, no endpoints |
| `w2-03` / `crop-mid3` | 140×36 | neon dark | focus `[M21]`, 6× enlargement | magenta long edge crosses magenta vertical with no jump marker; `›▶` stacked arrowheads |
| `crop-left` / `crop-right` | 140×36 | neon dark | 6× edge enlargements of `w2-02` | header `ve 6`; right-edge nodes drawn without a right border |
| `crop-node` / `crop-focusnode` | 140×36 | neon dark | gate node, unfocused vs focused | double border preserved in both states |
| `crop-legend` | 140×36 | neon dark | 3× enlargement of legend rows | `●` shared by active and next; `○`/`◌` near-identical |
| `w3-01` | 140×36 | neon dark | `v` under list focus | pixel-identical to `w1-02`: silent no-op |
| `w3-02`, `w3-03` | 140×36 | neon dark | Threads `?` overlay, scrolled to end | documents symbols and list keys; no `v`, no views |
| `w4-02` | 140×36 | neon dark | `y` in spatial | footer toast `✓ copied slug: ship-guarded-…`; `pbpaste` matched |
| `w5-02`, `w5-05`, `w5-06` | 140×36 | neon dark | `Enter` → `y` → `ctrl+o` | opens task on tasks tab with `‹ ctrl+o … (1)` breadcrumb; returns to spatial with `[M25]` still focused |
| `w6-01` | 140×36 | neon dark | `h`×8 → root, wave 1 | headers `wave 1 + gate / waves 1+2 / waves 2+3 / wave 3 / wave 4`; bottom row cut by inspector |
| `w6-04` | 140×36 | neon dark | `l`×14 → leaf, wave 13 | four left-edge nodes cut mid-box; header cut to `gate`; the six dispatchable tasks cluster in the last column |
| `w7-03`, `w7-04` | 140×36 | neon dark | `m` and `f` on a wave-reader task row | `m` silent no-op; `f` opens a picker targeting the highlighted task, showing full untruncated slugs |
| `w8-02`, `w8-03`, `w8-04` | 140×36 | neon dark | `ctrl+p` palette: `view`, `spatial`, `graph` | no view-switch command indexed |
| `w9-01`, `w9-02` | 140×36 | neon dark | `z` full-screen summary and waves | no wrapping; footer leads with mode name `full-screen` |
| `m1-01`…`m1-06` | 100×28 | neon dark | full flow | both identities destroyed in list; spatial shows 3 columns × 2 rows with arrow stubs into the frame edge |
| `n1-01`…`n1-06` | 72×24 | neon dark | full flow | list pane dropped, detail auto-focused; current tab off-screen; footer cuts `h/esc back`; nodes clipped on both edges |
| `e1-01`…`e1-04` | 140×34 → 100×28 → 72×24 → 140×34 | neon dark | live tmux `resize-window` | reflows cleanly at every step; no stale frame; matches the static captures |
| `e1-05` | 140×34 | neon dark | `r` refresh in spatial | view and focus `[G1]` preserved |
| `e1-06`, `e1-07` | 140×34 | neon dark | `Esc`, `Esc` from spatial | first → waves with selection carried; second → list focus with the pane still in topology and no indicator |
| `a1-01`, `a1-03`, `a1-04` | 140×36 | neon dark | Atlas default, after `v`, `?` | current view named in the tab strip (`atlas work` / `atlas spaces`) and body header; `?` documents `v` |
| `d1-01`…`d1-07` | 140×36 | neon dark | `assets/demo-threads` touring-bike Thread | small graph reads well; labels still truncated with 40 columns free; inspector cuts the bottom node |
| `crop-overlap` | 140×36 | neon dark | 4× enlargement of `d1-06` | node bottom border replaced by the inspector's top border |
| `t1-01` | 140×36 | catppuccin dark | spatial | all semantic distinctions preserved |
| `t2-01` | 140×36 | neon on light background | spatial | light palette engages; edge contrast 3.6:1 vs 5.5:1 on dark |

Two captures were discarded rather than reported: an early `y` frame that appeared to navigate away
(`w3-06`) and an early Atlas `v` frame that appeared to do nothing (`a1-02`). Both were screenshot
timing artefacts and both failed to reproduce under `w5-05` and `a1-03` respectively.

### Journey map

| Step | What the user does | What the UI gives back | Verdict |
| --- | --- | --- | --- |
| **Find** | Opens the Threads tab, scans for the initiative | Two rows led by `● g✓/v✓ ▶0 ✓6 ×1`; the name is truncated at 100 cols and middle-elided at 140 for the longer Thread; the description column renders `De…` | ✗ M1 — the surviving text is the part with no identity |
| **Recognize** | Reads the row to say what the Thread *is* before opening it | The symbols are undocumented in the row; `?` explains them well, but only after you think to ask | ~ Partial; M4 for the glyph collisions |
| **Open** | `l` | Border recolours; content is unchanged; footer swaps to detail hints | ~ Focus change is subtle but consistent with the rest of the shell |
| **Discover the view** | Looks for other presentations | Footer under list focus: nothing. `v` under list focus: nothing. `?`: nothing. Palette: nothing. Only `v` *after* `l`, labelled by destination | ✗ H1 — the blocker |
| **Comprehend** | Ten seconds on the spatial graph | Direction, roles, focus, magenta incident routes and the frontier cluster all land; but ~15-char labels, edge-clipped nodes, cut wave headers and full-width corridor rules do not | ✗ H2, H3, M2, M6 — wins on shape, loses on identity |
| **Navigate** | `hjkl`, predicting each destination | `h`/`l` walk the dependency graph across lanes with an unstated tie-break; `j` with no neighbour is silent | ✗ M5 |
| **Inspect** | `Enter`, `y` | Opens the right task on the tasks tab; `y` copies the right slug with a footer toast | ✓ |
| **Return** | `ctrl+o`, then `Esc`, `Esc` | `ctrl+o` restores the spatial view, node and viewport exactly. `Esc` → waves with selection carried. Second `Esc` → list focus with the pane silently still in topology | ✓ then ✗ M7 |
| **Degrade** | 100 then 72 columns | 100: identity destroyed in the list, graph down to 3×2 with arrow stubs into the frame. 72: list pane dropped (reasonable), current tab off-screen and `h/esc back` truncated away (not reasonable) | ✗ L2 — identity and the exit are the first things lost |

### Comparison: spatial vs the wave reader

Asked the same question of both — *where does the work converge and what is holding it up* — the
spatial view wins on the small graph and loses on the large one, for a reason worth naming. On the
touring-bike Thread (`d1-03`) convergence is geometry: three lanes narrow into one node. On the
37-member Thread the same geometry exists (`w6-04` shows all six dispatchable tasks stacked in the
final column, which is a genuinely good picture) but the labels that would tell you *which* six are
truncated to 15 characters, so the answer has to be assembled by moving the cursor and reading the
inspector six times. The wave reader (`w1-05`) prints all six names, states and ids in one screen
without any navigation. Fixing M2 and H2 gives the spatial view a fair test on both graphs; it does
not predetermine that one representation must win every planning question.

### Residual questions for maintainer taste

1. **Should `v` remain the only cycle, and should the chips be operable?** The brief explicitly
   does not want `v` removed. A chip strip could be display-only (cheapest, matches the Atlas) or
   clickable/numbered (`1`/`2`/`3`). Display-only is my recommendation, but that is a preference.
2. **Should the two legend rows be permanent?** They cost 2 of 24 rows at narrow sizes and are pure
   overhead once learned. Collapsing them behind `?` is a density judgement, not a defect.
3. **Is the description column worth keeping in Thread rows at all?** It never rendered more than
   three cells at any size tested. Removing it is one option; giving it the space freed by
   right-aligning the codes is another.
4. **Is 3.6:1 edge contrast on the light background acceptable?** It clears the WCAG non-text floor
   of 3:1 with little headroom, and the graph's whole argument rests on those lines being readable.
   Whether to darken them for the light palette is a taste call, not a compliance failure.
5. **Is the magenta incident-route highlight doing too much?** It reads well for short edges
   (`d1-06`) and becomes indistinguishable from decoration on long ones (`crop-empty`). Whether
   that is fixed by H3's routing changes alone, or also needs a quieter highlight, wants a look at
   the hardened picture first.

Settled during maintainer triage: `h`/`l` remain semantic prerequisite/dependent walks. The owning
tasks must expose that contract and remove branch/dead-end ambiguity rather than changing the keys
to geometric column movement.

## Appendix A — Design research

_Added 2026-09-09 after the review was delivered, at the owner's request. This appendix documents
external research into how other systems solve the problems the findings describe. It is not itself
review evidence: nothing here was observed in the rendered TUI. Subsequent maintainer triage is
called out where it corrected the research interpretation, source diagnosis, or task mapping.
Claims are labelled **[verified locally]** when checked against this repository or its pinned
dependencies, and **[read]** when taken from published documentation or literature that was not
independently reproduced._

### A.1 Method and evidence quality

Nine web searches plus four page fetches across four strands: graph-drawing literature, terminal
rendering primitives, existing ASCII/TUI graph renderers, and keybinding-discoverability patterns.
Where a source made a claim about a library this repository already depends on, the claim was
re-checked against the pinned module in the local Go module cache rather than trusted from release
notes — that check caught one discrepancy, recorded in A.4.

Three tiers of confidence are used below:

- **Verified locally.** Checked against `go.mod`, the module cache, or this repository's source.
  Cited with a path and version.
- **Documented.** Stated in a project's own README or reference docs. Reliable about intent and API,
  not independently exercised. Implementation details behind these claims were not read.
- **Literature.** Peer-reviewed or preprint results. Reliable as a prior, with the usual caveats
  about generalising from a controlled experiment to this specific product (see A.13).

### A.2 Graph size, density, and task—not a twenty-node crossover

Maintainer triage subsequently read the cited 2017 replication rather than relying on the original
appendix summary. The combined research does **not** establish a node-count threshold at which a
Thread should stop using a node-link presentation.

**[Literature]** Ghoniem, Fekete and Castagliola compared synthetic graphs at 20, 50, and 100 nodes
and unusually high densities of 0.2, 0.4, and 0.6. They found node-link stronger for small sparse
connectivity tasks and matrices stronger across large, dense conditions, with path-finding remaining
a node-link strength. “Above roughly twenty vertices” was therefore an overgeneralization: size,
density, and task varied together in the reported result.

**[Literature, read during maintainer triage]** Okoe, Jianu and Kobourov's 2017 *Revisited
Experimental Comparison of Node-Link and Matrix Representations* evaluated interactive views of a
much larger but sparse real-world network: 258 nodes, 1,090 edges, and density 0.016. Across 557
participants and 14 task types, node-link was better for topology and connectivity questions and
comparable for group and memorability tasks. The paper explicitly attributes some disagreement with
earlier work to network sparsity, structure, interaction, and task choice, while warning that its
single dataset is not universally generalizable.

The reviewed Threads are both DAGs and the production example is sparse under the later paper's
stated `edges / nodes²` convention:

| Thread | Nodes | Edges | `edges / nodes²` | Review's observation |
| --- | ---: | ---: | ---: | --- |
| `touring-bike-departure` (`assets/demo-threads`) | 8 | 8 | 0.125 | Reads in one glance at 140×36; `d1-03` proves the direction is promising |
| `complete-production-threads` (this repo) | 40 | 51 | 0.032 | Fixed labels, cell clipping, and ambiguous long routes obstruct comprehension |

The visual evidence still proves that the current 40-node rendering is harder to read. It does not
prove that node-link itself crossed a size ceiling: every observed failure has a concrete identity,
viewport, or routing mechanism, and the later study found node-link useful at far larger sparse
graphs for the same topology/path questions Threads are meant to answer. `ascii-dag`'s warning that
very dense graphs with hundreds of crossing edges become unreadable remains sensible, but it does
not describe this 40-node, 51-edge projection.

**What this implies.** Keep the full spatial graph as a flagship representation worth hardening.
Keep the wave reader as a complementary ordered/status scan rather than declaring it the at-scale
winner. Keep `6g86c7y6hn41` (one-hop focus) high priority because focus+context is useful for local
causal inspection, not because it is the only node-link mode “correct by construction.” Do not
automatically change representations at an arbitrary node count; use dogfooding evidence across
graph shape, density, viewport, and user question before introducing any threshold or semantic zoom.

### A.3 Focus+context and degree-of-interest

**[Literature]** The general focus+context technique has a long line of work behind it.
A *degree-of-interest* (DOI) function scores each node against a model of what the user currently
cares about, and the view then filters or de-emphasises by that score — the basis for fisheye views
and for Card and Nation's Degree-of-Interest Trees, which show large taxonomies by filtering to the
elements a model of current interest deems relevant. *Focus+context* displays keep a detailed view
of the high-interest region while retaining surrounding context so the analyst stays oriented, with
spatial continuity preserved across zoom and pan. *Semantic zoom* goes further, changing the encoding
itself at different scales — showing meta-nodes for clusters when zoomed out and expanding them when
zoomed in.

Three of these map onto decisions already in front of this codebase:

1. **DOI as the subgraph rule.** The one-hop rule in `6g86c7y6hn41` is a DOI function with a
   hard cutoff at distance 1. That is a defensible starting point, and the literature suggests the
   generalisation (distance-weighted emphasis rather than binary inclusion) if one hop proves too
   tight — dimming two-hop neighbours rather than removing them.
2. **Context retention.** The task's acceptance criterion already requires stating "how many nodes
   are shown versus hidden." Focus+context work suggests going one step further and keeping hidden
   neighbours visible as boundary stubs, so spatial continuity survives the transition.
3. **Semantic zoom as a later experiment, not the presumed whole-graph answer.** Collapsed wave or
   layer meta-nodes could eventually provide overview context for genuinely large/dense Threads.
   The current 40-node graph does not establish that need, and introducing synthetic meta-nodes now
   would complicate the view's promise to render the supplied projection faithfully. Keep it out of
   the present hardening tasks unless dogfooding after the concrete clipping/routing fixes still
   demonstrates a representation ceiling.

### A.4 Rendering primitives already present in the pinned dependency set

**[Verified locally]** against `go.mod` and `$(go env GOMODCACHE)`. This matters because several of
the review's findings were framed as needing new machinery, and they do not.

`charm.land/lipgloss/v2 v2.0.6` ships a cell-addressable canvas and a z-ordered layer system:

- `lipgloss.NewCanvas(width, height) *Canvas` — documented in `canvas.go` as "a cell-buffer that can
  be used to compose and draw `uv.Drawable`s like `Layer`s", implementing `uv.Screen` and
  `uv.Drawable`. Composed drawables draw in order, so later ones appear on top.
- `(*Canvas).CellAt(x, y) *uv.Cell` and `(*Canvas).SetCell(x, y, *uv.Cell)` — per-cell read and
  write.
- `(*Canvas).Draw(scr uv.Screen, area uv.Rectangle)` and `(*Canvas).Bounds() uv.Rectangle` — an
  explicit clip rectangle.
- `(*Canvas).Compose(drawer uv.Drawable) *Canvas` and `(*Canvas).Render() string`.
- `lipgloss.NewLayer(content string, layers ...*Layer) *Layer` with `.X(int)`, `.Y(int)`, `.Z(int)`,
  `.ID(string)`, `.AddLayers(...)`, `.GetLayer(id)`, `.MaxZ()` — real z-ordering and named lookup.
- `lipgloss.Blend1D(steps int, stops ...color.Color) []color.Color` and
  `Blend2D(width, height int, angle float64, stops ...color.Color) []color.Color` in `blending.go`.
- `NormalBorder()`, `RoundedBorder()`, `ThickBorder()`, `DoubleBorder()`, `BlockBorder()`,
  `OuterHalfBlockBorder()`, `InnerHalfBlockBorder()`, `HiddenBorder()`, `MarkdownBorder()`,
  `ASCIIBorder()` — including a heavy weight and an ASCII fallback set.

**This repository already uses that exact pattern.** `internal/tui/help.go:346-349` composes
`lipgloss.NewLayer(bg)` with `lipgloss.NewLayer(fg).X(x).Y(y).Z(1)` onto a `lipgloss.NewCanvas(width,
height)` to render the help overlay. The spatial view does not use it. These primitives could help
compose fixed chrome, boundary gutters, and a clipped drawing area, but they do not by themselves
solve H2 or M3: source reconciliation showed that the inspector rows are already reserved and the
remaining failure is admission of partial node slots through cell-level panning. The current custom
canvas also owns graph-specific connector merging and route highlighting, so replacing it with Lip
Gloss Canvas is an implementation option to test, not a prerequisite or evidence that the fix is
trivial.

`github.com/charmbracelet/ultraviolet` is already present as an indirect dependency
(`v0.0.0-20260811164956-006e29f97886`) — **[documented]** it provides the terminal primitives that
power Bubble Tea v2 and Lip Gloss v2, including the `ScreenBuffer` type the Canvas is built on, and
is usable standalone.

`github.com/charmbracelet/harmonica v0.2.0` is likewise already an indirect dependency — Charm's
spring-physics animation library, which is the disciplined way to animate viewport panning if the
graph gains smooth motion between focused nodes.

`internal/design` already carries the token system this would build on: a `Palette` with a semantic
`map[theme.Color]Hue`, chrome slots (`Accent`, `BorderActive`, `BorderIdle`, `Danger`, `Heading`,
`Match`, `Track`, `Surface`), and a `Gradient []Hue` documented as "the rollup bar fill stops (the
deliberate truecolor exception; purple -> cyan -> pink)" that "degrades per-cell on low-color
terminals." Contrast is already pinned by tests — `TestFindHighlightContrastAA`,
`TestChromeSurfaceContrastAA`, `TestNeonLightHighlight`.

**Discrepancy caught by local verification.** The Lip Gloss v2 announcement discussion (#506)
enumerates the v2 changes — deterministic styles, I/O control, colour downsampling,
`HasDarkBackground()`, `LightDark()`, `Complete()`, `color.Color` — and does **not** mention canvas,
layers, z-ordering, clipping, or gradients. Those APIs are nonetheless present in the pinned v2.0.6
source as listed above. Anyone planning this work from the release notes alone would wrongly
conclude the primitives are unavailable.

### A.5 Edge routing: techniques worth borrowing for H3 and L3

**Line jumps (the crossing/junction disambiguation).** **[Documented]** Line jumps — also called
edge jumps or bridges — exist specifically "to show that lines cross each other, but do not connect,"
and are used in electrical diagrams and anywhere crossing connectors would otherwise be ambiguous.
The convention descends from schematic drawing, where designers drew "wire bumps" wherever a line
crossed another to show visibly that the lines cross rather than join. Modern tools expose a jump
*style*: draw.io added line-jump support with configurable shapes, and Visio's documented styles
include "a smooth arc, a **gap**, a square, or a multi-sided arc."

In a character grid, the gap is the only one of those four that is representable. The concrete
technique for H3 is therefore:

1. When the router is about to write a segment into a cell that already holds a perpendicular
   segment, write **nothing** for that one cell (a gap) instead of a junction glyph.
2. Reserve `┼ ├ ┤ ┬ ┴` **exclusively** for true junctions present in the supplied
   `ThreadGraphProjection`.

This is directly enabled by `CellAt(x, y)` from A.4 — the router can read the destination cell before
writing it. It also gives task `6g86e10dztpf` a testable invariant matching its AC 2: a junction
glyph appears if and only if the projection contains a real junction at that point.

A secondary channel is available if the gap alone proves too subtle: `ThickBorder()` alongside
`NormalBorder()` means line weight can express "over/under", the same way heavy-versus-light strokes
do in printed schematics.

**Visible routing waypoints.** **[Documented]** `ascii-dag` routes multi-level edges through explicit
waypoints and can render them with a `◍` marker at each routing point (`show_dummy_nodes`). This
speaks directly to H3's root cause: the review's `crop-empty` frame shows corridor routes with no
arrowhead, no endpoint and no attachment anywhere on screen, which is why they read as table rules.
A dim waypoint glyph where a route enters and leaves a corridor restores the "this is an edge going
somewhere" reading without needing the endpoints to be visible.

**Label overflow policy.** **[Documented]** `ascii-dag` handles labels that do not fit by "omission or
legend collection" rather than by truncating in place. That is a better answer to M2 than fitting
more characters into a node box, and this codebase is unusually well set up for it: the spatial view
already assigns and displays short codes (`[M25]`, `[G1]`, visible in every review frame). The honest
narrow-width node is `[M25]` with the code roster available, not `add-a-guarded-r…`.

**Orthogonal routing theory.** **[Literature]** In orthogonal drawings edge routes are sequences of
alternating horizontal and vertical segments, which is what a character grid natively supports.
Wybrow, Marriott and Stuckey's *Orthogonal Connector Routing* is the standard reference for
computing such routes with obstacle avoidance; yWorks' polyline/orthogonal router documentation
covers the same ground from a product perspective, including how commercial layout engines treat
crossings and bundling. `ascii-dag` states **[documented]** that it routes strictly orthogonally with
no diagonals, "constraining the visual representation but keeping it scannable in text format" —
which is the correct trade for this product.

### A.6 Layered-layout prior art

**[Literature/documented]** The layout family in use here is Sugiyama's: assign nodes to layers,
reorder within layers to minimise crossings, then route edges — with all edges except back-edges
pointing one way. `ascii-dag` describes its pipeline as cycle breaking → layering → crossing
reduction → edge routing, with configurable crossing-reduction presets (`FAST`, `STANDARD`,
`QUALITY`), a pluggable `CrossingReducer`, and four layout directions; it also accepts cyclic input
by internally reversing problem edges and rendering them dashed.

Two details are relevant to specific findings:

- **Skip-level edges.** `ascii-dag` routes edges that skip levels "along the outer edges of the
  graph" rather than through the middle. That is an alternative to this prototype's corridor routing
  and would sidestep H3's failure mode entirely by never putting a long route in a position where it
  reads as a row separator.
- **The layer/wave distinction (M6).** The literature is explicit that layers are a layout construct.
  This review's M6 observed headers reading `waves 1+2` next to `wave 3`, meaning one wave straddles
  two columns. That is the layout's layer index leaking into a label that promises wave semantics —
  a naming problem, not a layout bug, and it resolves by labelling the axis for what it is.

Edge bundling is the documented next step when crossing minimisation alone is insufficient: it has
been suggested that crossing minimisation be preceded by an edge concentration or bundling step, and
that metro-line crossing-minimisation techniques applied to bundles preserve Sugiyama style while
producing a more readable view. This is probably beyond what `6g86e10dztpf` should attempt, but it
is the named escape hatch if the bounded router cannot meet its contract.

### A.7 Terminal DAG renderers surveyed

| Project | Language | Approach | Relevance |
| --- | --- | --- | --- |
| `ascii-dag` | Rust | Sugiyama in a fixed-width character grid; zero-dependency, `no_std`; grid positions snapped to cells | Closest analogue to this prototype; source of the waypoint, legend-collection and skip-level techniques above |
| `graphs-tui` | Rust | Terminal renderer for Mermaid and D2 — flowcharts, state diagrams — as Unicode or ASCII | Demonstrates the density ceiling in practice |
| `mermaid-ascii`, `mermaid2term`, `mermaid-ascii-diagrams` | Go / various | Mermaid → ASCII/Unicode in the terminal | Same class; configurable spacing and padding |
| `serie` | Rust | Rich git commit graph rendered via **terminal image protocols** | See A.10 — instructive as a rejected approach |
| `tig` | C | Git DAG in a character grid (`graph-v1.c`, `graph-v2.c`) | The most battle-tested lane-assignment code in this space |
| `git-foresta`, `git-graph-drawing` | Perl / various | Text git graph viewers; a collection of drawing implementations | Reference implementations for lane packing |

**[Documented]** A recurring caveat across the Mermaid/D2 renderers: readability is the binding
constraint, and once a diagram grows, ASCII rendering "can turn into a carpet of characters," with
practical value concentrated in smaller structures rather than large architecture diagrams. That is
the same conclusion as A.2 reached from a different direction.

For lane assignment specifically, pvigier's *Commit Graph Drawing Algorithms* is the clearest
public writeup: it notes that commit-graph drawing is simplified by the graph being directed,
acyclic and timestamped, and that a particular visit order during topological sorting is used to
minimise crossing edges. Thread graphs share the directed-acyclic property and have a wave index
that plays the role timestamps play there, so the technique should transfer.

### A.8 Wayfinding and keybinding discoverability (H1, M7, M5)

Two complementary patterns, addressing two different questions.

**"Where am I?" — a persistent view indicator.** Atlas already names the current view more clearly
than Thread detail, per H1's evidence: frames `a1-01` and `a1-03` show `atlas work` / `atlas spaces`
in the tab strip. It still does not answer the distinct “what else is available?” question without
cycling or opening help, so both surfaces belong to the shared fix. **[Documented]** k9s's convention is the general
form of this: always show the current navigation path as a breadcrumb, and provide a `:resource`
command mode for direct jumps in deep hierarchies. This codebase already has the `:` command mode —
H1's evidence (`w8-02`, `w8-03`) is that view names are simply not indexed in it.

**"What can I do here?" — which-key / minor modes.** **[Documented]** Helix pops up a menu of
available commands when a prefix key is pressed, entering a "minor mode." Users credit this
specifically with giving Helix "a sense of discoverability that is often not present in terminal
UIs," and with letting them rediscover bindings rather than memorise them. This is
**[literature]** progressive disclosure applied to keybindings — the pattern of deferring advanced
features to a secondary surface and revealing them when relevant to the current task.

The two are not substitutes. The chip strip answers H1's "the views are invisible"; a which-key
popup would answer M5's "the dead key told me nothing" and would generalise beyond this screen.

**[Documented]** A published TUI design skill dissects lazygit, k9s, btop, helix, yazi, atuin, htop,
Posting, Harlequin, Claude Code and starship as exemplars for focus and discoverability, and codifies
seven canonical layouts (multi-panel, miller columns, drill-down stack, dashboard, IDE three-panel,
overlay, tabbed-within-panel) along with the standard cross-app bindings this product already
follows (`q`, `?`, `/`, `Esc`, `hjkl`, `Tab`). Its non-negotiables list — alt screen, panic-safe
restoration, `SIGWINCH`, `SIGTSTP` — is worth checking against, though the fetched summary did not
elaborate the per-app lessons.

### A.9 Colour accessibility (M4)

**[Documented]** Three usable rules emerged:

1. **Blue/orange as the primary pair.** Blue and orange are distinguishable under all common forms of
   colour vision deficiency, making them the safe choice for the most critical status distinction.
   The **Okabe-Ito** palette is the standard eight-colour qualitative set designed for
   colour-universal design and is a safe default for categorical data.
2. **Prefer lightness contrast over hue contrast.** When choosing between two similar hues, take the
   one with greater lightness difference — lightness remains perceptible when hue discrimination is
   reduced. This is the directly actionable rule for M4: `● active` (olive) and `● next` (blue)
   currently differ in hue at similar lightness and share a glyph. Separating them on lightness *and*
   glyph fixes the finding on two independent channels.
3. **Never let colour be the only carrier.** Redundant coding — glyph, label, shape or pattern
   alongside hue — is the standard remedy. The review noted the spatial view already does this
   correctly for *roles* (member = single border, external gate = double border, preserved under
   focus per `crop-node` and `crop-focusnode`); the gap is statuses.

Bloomberg's account of redesigning the Bloomberg Terminal for colour accessibility is the closest
published case study to this product's situation: a dense, professional, heavily colour-coded
terminal interface with an existing user base.

### A.10 Terminal graphics protocols — considered and rejected

**[Documented]** `serie` renders git commit graphs as genuine images through the iTerm2 inline-images
protocol and the kitty terminal graphics protocol (both standard and Unicode-placeholder modes), with
`--graph-width` (`auto`/`double`/`single`) and `--graph-style` (`rounded`/`angular`) options. The
results are visually excellent and it is the strongest demonstration of what a terminal graph *could*
look like.

It is the wrong model for `tskflwctl`, and the project says so itself: universal terminal
compatibility is an explicit **non-goal**. Sixel is unsupported, Windows is unsupported, terminal
multiplexers other than tmux (kitty Unicode mode only) are unsupported, and **there is no Unicode
fallback** — if the terminal cannot display images, `serie` cannot function.

That is incompatible with this product's constraints as they already exist in the codebase: a
`--color auto|always|never` contract, `NO_COLOR` support, an `ASCIIBorder()` fallback path, and
review evidence that the TUI is expected to work inside tmux (frames `e1-01`–`e1-04` were captured
through a live tmux resize). Protocol support is also uneven in 2026 — kitty's protocol is fully
implemented in kitty and largely in WezTerm, partially in Konsole and wayst; the *Are We Sixel Yet?*
tracker is the current state of the other protocol.

**Recommendation:** treat terminal graphics as a possible progressive enhancement layered over a
complete Unicode renderer, never as the substrate. `go-termimg` is the Go library to reach for if
that is ever attempted; it supports kitty (with virtual images and z-index, marked experimental) and
sixel with dithering, and documents a Bubble Tea `ImageWidget` integration.

### A.11 Subcell rendering via braille — partially applicable

**[Documented]** Unicode braille (U+2800–U+28FF) gives a 2×4 dot grid per cell, so a braille canvas
addresses eight independently settable points per character — two times wider and four times taller
resolution than a character canvas. `mum4k/termdash` ships a `canvas/braille` package implementing
exactly this in Go, with the documented workflow of setting sub-cell points then copying the result
to a regular character canvas; `drawille-go` is a further Go port of the original `drawille`.

**Assessment for this view.** Braille is the right tool for continuous curves — sparklines, plots,
the smooth diagonal edges `serie` achieves with images. It is a poor fit for *this* graph for two
reasons: colour is applied per character cell, not per dot, so a braille edge cannot change colour at
a crossing or carry the magenta focus highlight at sub-cell granularity; and braille strokes render
as dotted lines, which would conflict with using dashes to mean "reversed" or "long-route" edges.
Recorded here so the option is on the table with its trade-offs, not as a recommendation.

### A.12 Research → finding → task mapping

Maintainer triage connected the useful techniques to the corrected findings and their concrete task
owners. Techniques remain options to test against rendered evidence; appearing here does not make a
specific implementation mandatory.

| Finding | Technique from research | Owning task |
| --- | --- | --- |
| H1, M7 | Persistent current-view indicator plus a visible available-view set for both Thread and Atlas; index view names in `:`/palette; which-key popup only as a possible broader generalisation | `6g87qn72901g` |
| H2 | Complete node/layer viewport units and a boundary rail for nodes/headers; route endpoint stubs and incident-edge continuity at the same boundary | `6g8btt5hcgs9` (layout boundary) and `6g86e10dztpf` (route boundary) |
| H3 | Line-jump **gap** at crossings; `┼` reserved for real junctions; visible routing waypoints (`◍`); skip-level edges routed on the outer boundary | `6g86e10dztpf` |
| L3 | Single arrowhead at the node boundary with fan-in count; distinct vertical entry rows | `6g86e10dztpf` |
| M2 | Label budget derived from viewport ÷ visible layers; **legend collection** keyed on existing `[M25]`/`[G1]` codes where full identity cannot fit | `6g8btt5hcgs9` |
| M3 | Slot-aware vertical panning above the already-reserved inspector; Lip Gloss layers remain optional compositing machinery, not the root fix | `6g8btt5hcgs9` |
| L1 | Responsive bounded pane allocation rather than another fixed ratio | `6g8btt5hcgs9` |
| M4 | Six glyphs for six states; separate `active`/`next` on lightness as well as hue; Okabe-Ito as the categorical reference | `6g87qn72901g` |
| M5 | Preserve and name semantic edge walks in the footer; expose branch choice; explicit message on a dead key | `6g86c7y6hn41` (navigation contract) and `6g86g03zfj2f` (unsupported-action feedback) |
| M6 | Label the axis as layout layers, not waves; keep wave membership on the node and inspector | `6g86e10dztpf` |
| M1, L2 | Identity-first rows; drop rather than stub the description column; responsive chrome | `6g87qn72901g` |
| — (roadmap framing, not a finding) | No numeric representation cutoff: harden the full sparse-DAG view, retain waves for ordered scanning, and add focus as a complementary local lens | `6g86e10dztpf`, `6g8btt5hcgs9`, and `6g86c7y6hn41` retain parallel, non-replacement scopes |

### A.13 Caveats on this appendix

1. **There is no established numeric crossover for this product.** Ghoniem et al. tested dense
   synthetic graphs; Okoe et al. later found node-link stronger for topology/connectivity on one
   much larger sparse real-world graph. Neither tested a layered planning DAG in a character-cell
   terminal. Graph shape, density, interaction, task, and viewport all matter alongside size.
2. **Implementation details behind the ASCII renderers were not read.** The `ascii-dag`, `tig` and
   `serie` claims here come from documentation, not source. The techniques are sound and
   well-established; the specific implementations are unverified.
3. **The library API claims were verified; their behaviour was not exercised.** A.4's symbols were
   read out of the pinned module source in the local module cache. No code was written against them
   during this review, and no build or test was run for this appendix.
4. **Nothing in this appendix was observed in the rendered TUI.** The review's findings rest on the
   captured frames in the evidence matrix. This appendix rests on external sources and local source
   inspection, and is explicitly not visual evidence.
5. **One source is a vendor tracker and one is a personal blog post.** *Are We Sixel Yet?* and the
   TUI-design-skill post are cited for orientation, not as authorities.

### A.14 Reference index

**Graph readability and representation choice**

- Ghoniem, Fekete & Castagliola — *On the Readability of Graphs Using Node-Link and Matrix-Based
  Representations: A Controlled Experiment and Statistical Analysis* (Information Visualization,
  2005) — <http://www-sop.inria.fr/orion/COGC/teams/INSITUghoniem-fivj05-final.pdf>
- Ghoniem, Fekete & Castagliola — *A Comparison of the Readability of Graphs Using Node-Link and
  Matrix-Based Representations* (InfoVis 2004) —
  <https://www.researchgate.net/publication/221005943_A_Comparison_of_the_Readability_of_Graphs_Using_Node-Link_and_Matrix-Based_Representations>
- Okoe, Jianu & Kobourov — *Revisited Experimental Comparison of Node-Link and Matrix
  Representations* (2017 preprint; read during maintainer triage) —
  <https://arxiv.org/pdf/1709.00293>
- Henry, Fekete & McGuffin — *NodeTrix: Hybrid Representation for Analyzing Social Networks* —
  <https://arxiv.org/pdf/0705.0599>
- *Matrices or node-link diagrams: which visual representation is better for visualising
  connectivity models?* — <https://dl.acm.org/doi/10.1057/palgrave.ivs.9500116>

**Focus+context, degree-of-interest, semantic zoom**

- Card & Nation — *Degree-of-Interest Trees: A Component of an Attention-Reactive User Interface* —
  <https://courses.ischool.berkeley.edu/i247/f05/readings/Card_DOITrees_AVI02.pdf>
- Heer & Shneiderman — *Interactive Dynamics for Visual Analysis* —
  <https://www.cs.umd.edu/~ben/papers/Heer2012Interactive.pdf>
- *iSphere: Focus+Context Sphere Visualization for Interactive Large Graph Exploration* —
  <https://dl.acm.org/doi/pdf/10.1145/3025453.3025628>
- *Semantic Zoom View: A Focus+Context Technique* —
  <https://summit.sfu.ca/_flysystem/fedora/sfu_migrate/11587/etd6479_DDunsmuir.pdf>
- *Responsive Matrix Cells: A Focus+Context Approach for Exploring and Editing Multivariate Graphs* —
  <https://arxiv.org/pdf/2009.03385>
- Progressive disclosure (overview) — <https://en.wikipedia.org/wiki/Progressive_disclosure>

**Layered layout and orthogonal edge routing**

- *The Sugiyama Method — Layered Graph Drawing* — <https://blog.disy.net/sugiyama-method/>
- Reference implementation — <https://github.com/auroraptor/sugiyama-algorithm>
- *Improving Layered Graph Layouts with Edge Bundling* —
  <https://link.springer.com/chapter/10.1007/978-3-642-18469-7_30>
- Wybrow, Marriott & Stuckey — *Orthogonal Connector Routing* (GD'09) —
  <https://people.eng.unimelb.edu.au/pstuckey/papers/gd09.pdf>
- yWorks — *Drawing Orthogonal Diagrams* — <https://www.yworks.com/pages/drawing-orthogonal-diagrams>
- yFiles — *Polyline Edge Routing* —
  <https://docs.yworks.com/yfiles-html/dguide/layout/polyline_router.html>
- *Orthogonal Edge Routing for the EditLens* — <https://arxiv.org/pdf/1612.05064>

**Crossing versus junction disambiguation**

- draw.io — *Now Supports Line Jumps* —
  <https://drawio-app.com/blog/draw-io-now-supports-line-jumps/>
- Microsoft — *Add or remove connector line jumps* (documents gap/arc/square jump styles) —
  <https://support.microsoft.com/en-us/office/add-or-remove-connector-line-jumps-1a5516db-0212-4352-90dd-fc9cfd4018ed>
- Javelin — *How to display a Wire Bump / Cross Over in SOLIDWORKS PCB* (the schematic convention) —
  <https://www.javelin-tech.com/blog/2017/08/solidworks-pcb-display-cross-over/>
- Unicode Box-drawing characters — <https://en.wikipedia.org/wiki/Box-drawing_characters>
- Box Drawing block reference — <https://jrgraphix.net/r/Unicode/2500-257F>

**Terminal DAG and diagram renderers**

- `ascii-dag` — <https://github.com/AshutoshMahala/ascii-dag> · crate listing —
  <https://lib.rs/crates/ascii-dag>
- `graphs-tui` (Mermaid + D2 → Unicode/ASCII) — <https://github.com/decisiongraph/graphs-tui>
- `mermaid-ascii` — <https://github.com/AlexanderGrooff/mermaid-ascii>
- `mermaid2term` — <https://github.com/watzon/mermaid2term>
- `serie` (image-protocol git graph) — <https://github.com/lusingander/serie>
- pvigier — *Commit Graph Drawing Algorithms* —
  <https://pvigier.github.io/2019/05/06/commit-graph-drawing-algorithms.html>
- `git-graph-drawing` (collection of implementations) —
  <https://github.com/indigane/git-graph-drawing>
- `git-foresta` — <https://github.com/takaaki-kasai/git-foresta>

**Rendering primitives (Charm ecosystem)**

- Lip Gloss — <https://github.com/charmbracelet/lipgloss> · v2 API docs —
  <https://pkg.go.dev/charm.land/lipgloss/v2> · v2.0.0 release —
  <https://github.com/charmbracelet/lipgloss/releases/tag/v2.0.0>
- *Lip Gloss v2: What's New* (discussion #506 — note the omissions described in A.4) —
  <https://github.com/charmbracelet/lipgloss/discussions/506>
- Ultraviolet — <https://github.com/charmbracelet/ultraviolet>

**Subcell / braille rendering**

- `termdash` braille canvas —
  <https://pkg.go.dev/github.com/mum4k/termdash@v0.6.1/canvas/braille>
- `drawille-go` — <https://github.com/exrook/drawille-go>
- `drawille` (original) — <https://github.com/asciimoo/drawille>

**Terminal graphics protocols**

- *Are We Sixel Yet?* — <https://www.arewesixelyet.com/>
- `go-termimg` — <https://github.com/blacktop/go-termimg>
- *Terminal Graphics Protocols: Kitty, Sixel, iTerm2, and Beyond* —
  <https://akmatori.com/blog/terminal-graphics-protocols>

**TUI wayfinding and discoverability**

- Helix — *Are you happy with helix's keybinding model?* (minor modes / which-key discussion) —
  <https://github.com/helix-editor/helix/discussions/1511>
- Helix — key remapping reference — <https://docs.helix-editor.com/remapping.html>
- *Helix: Why (And How) I Use It* — <https://jonathan-frere.com/posts/helix/>
- *Teaching Claude to design TUIs* (layouts, non-negotiables, exemplar app list) —
  <https://griffen.codes/post/tui-design-skill-claude/>
- `awesome-tuis` (survey of terminal UIs) — <https://github.com/rothgar/awesome-tuis>

**Colour accessibility**

- Bloomberg UX — *Designing the Terminal for color accessibility* —
  <https://www.bloomberg.com/ux/2021/10/14/designing-the-terminal-for-color-accessibility/>
- *Colorblind-Safe Color Palettes for Designers* (Okabe-Ito reference) —
  <https://colorblind.io/guides/colorblind-safe-palettes>
- *Color Blind Colors: Safe Palettes for Deuteranopia, Protanopia & Tritanopia* —
  <https://www.colorcontrast.org/color-blind-colors/>
- *Accessible Color Palette Guide — WCAG Color Systems for UI* —
  <https://accessibility.build/guides/accessible-color-palettes>
