---
schema: 1
id: 6g2qtp0022t7
created: "2026-08-23"
description: Why the atlas shipped list-shaped, why it should become two views, and the sequencing that keeps the useful half ahead of the delightful half.
tags: [tui, atlas, design, multi-repo]
updated_at: "2026-08-23"
---
# The atlas as a dashboard of dashboards

## Question

The atlas shipped in v0.17.0 as a navigator: it answers *"where can I go?"* well and
*"what am I working on?"* not at all. The remaining audit findings (M2/M3/M4) each
proposed a patch to that surface — a rail, a filter, badges — without anyone asking
whether the surface itself is the right shape.

Two things forced the question. First, what shipped looks nothing like the sketch that
authorized it: the sketch drew a grid of bordered mini-dashboard tiles, and the
implementation is a full-width vertical list. Second, the stated goal for this screen is
a *birds-eye view of every project* — in progress, critical findings, size of work
remaining — and a list of cards is a weak instrument for that.

## Why the sketch and the implementation diverged

Not drift. **The sketch predates the model it renders.** It was drawn 2026-08-15; the
"one planning identity, several entry points" refinement landed 2026-08-20. A tile cannot
hold what that refinement added.

The arithmetic is decisive. A full entry-point row, measured against the real registry:

```
  └─ ● desirelines-deploy    ok    ~/git/andy-esch/desirelines-deploy
```

That is **81 columns**. A two-across tile grid on a 100-column terminal leaves roughly
**48 usable columns inside each tile's borders**, and the longest registered path is 36
columns on its own. Entry-point rows physically do not fit in a tile, and a space can have
several of them, each with a variable-length detail and remedy line when unhealthy — so
tiles in one row could not even agree on a height.

The implementation went list-shaped because the list was the only shape that fit. That is
worth recording, because the sketch in
[`6g0ajre026c6`](6g0ajre026c6-multi-space-home-registry-and-the-atlas.md) still shows the
grid, and anyone reading it would re-derive a layout that was already tried and rejected
by the constraints.

## The reframe: two questions, not one screen

The atlas is being asked to answer two questions that want opposite layouts:

| Question | Wants | Shape |
| --- | --- | --- |
| Where can I go? | spatial, comparative, navigation-first | tiles |
| What am I working on? | temporal, sorted, work-first | a flat list |

These are not two renderings of the same data. They are different data at different
resolutions, and cramming both into one 20-row viewport is what makes every proposed
layout feel like a compromise.

**Make them two views on one screen, cycled by a key.** This fits the existing vocabulary
(`o`/`O` cycles order, `s`/`S` cycles status views on entity tabs) rather than inventing a
concept.

The important consequence is architectural, not visual: **it decouples the work from the
layout.** As a band beneath the cards, the rail competes with tiles for vertical budget and
has to be built after them. As its own view it owns a full viewport, can ship first
against today's list, and survives the later tile rewrite untouched.

A view cycle that toggles *questions* is principled. A view cycle that toggles *density or
shape of the same content* is usually a way to avoid deciding, and should be resisted.

## What a cross-space summary can actually show

Nearly everything wanted from a birds-eye banner is already computed. `core.Summary`
carries, per space:

| Field | Answers |
| --- | --- |
| `InProgress` | what is actively being worked on |
| `Findings.Acute` | the acute findings **themselves**, not just a count |
| `ReadyToClose` | audits where every finding is resolved |
| `RevisitDue` | deferred tasks whose snooze has expired |
| `Problems` | unreadable files |
| `Epics[].Total/Done/Deprecated` | size of the work remaining |

`SpaceOverview.InProgress` already aggregates the first of these across spaces, with the
space id attached, and `status --all` renders it. So a cross-space summary is mostly
aggregation of existing fields rather than new core work.

**The redundancy trap.** A banner and a rail are the same information at two resolutions.
Spending two blocks of a 20-row viewport on one fact is expensive. Either the banner is
counts-only and a rail carries detail, or the banner absorbs the rail (top N plus
"+M more") and there is no separate rail.

The way to settle that is not argument. Build the work view alone, with no banner, and see
whether one is missed. A full-viewport work view may well make a banner redundant, and
that is cheap to find out and expensive to guess wrong.

## Component landscape: nothing to adopt

Surveyed on 2026-08-23, because a tile grid is the first thing this TUI has wanted that
looks like generic layout.

- **`76creates/stickers`** (FlexBox + Table for lipgloss) is the only real candidate.
  Rejected: its `go.mod` requires `github.com/charmbracelet/bubbletea v1.1.1`, while this
  repo is on `charm.land/bubbletea/v2 v2.0.7`. Its FlexBox is lipgloss-only, so importing
  just that subpackage would not link two runtimes, but it would still put a v1 requirement
  in `go.mod` for roughly twenty lines of layout math. Against this repo's depguard-enforced
  dependency posture, a bad trade.
- **`rivo/tview`** and the `tcell` family are a different paradigm entirely; adopting one
  means rewriting the TUI, not extending it.
- The broader third-party bubbles ecosystem is still largely on v1 import paths. The
  `charm.land` v2 migration is recent enough that compatibility, not quality, is the
  binding constraint.

**Everything needed is already here.** `charm.land/lipgloss/v2` is a direct dependency and
ships `Border`, `Width`, `JoinHorizontal`/`JoinVertical`, plus `table`, `tree`, and `list`
subpackages. The repo's own `internal/progressbar` provides both the gradient bar and the
segmented finding bar, already surfaced on the TUI's styles bundle as `s.miniBar(pct,
width)` and `s.segmentBar(...)`, already theme-aware. The sketched tile is a composition of
widgets that exist; the genuinely new parts are the card frame and the layout math.

## Entry points inside a tile

Given the 48-column ceiling, three options were considered:

1. **Selected entry only.** The tile footer shows the selected entry's role and path plus
   `N entry points · h/l`. Cheapest; `h`/`l` keeps working unchanged.
2. **Expand on focus.** The focused tile's entry points render full-width below the grid.
   Keeps more information, but the grid reflows as the cursor moves, which reads as jumpy.
3. **Off the atlas entirely**, into a detail pane or modal. Most expensive: `h`/`l` loses
   its home and entering a specific checkout becomes a two-step flow.

**Recommendation: (1).** It regresses a shipped acceptance criterion from
[`6g2jhr3g20ss`](../tasks/6g2jhr3g20ss-build-the-tui-atlas-and-navigate-into-spaces.md)
— "shows every entry-point directory in subdued text" — so that criterion should be
amended explicitly rather than quietly dropped. The full directory listing already exists
in `space list` and `doctor`, which is arguably where an inventory belongs; the atlas is
for navigation and orientation.

## Worktrees: deferred, with one condition

A git worktree of a *direct* planning checkout has its own copy of the planning files on
another branch — so two registered worktrees are two genuinely different planning states
sharing one repo id. They group into one card, and `preferredHealthyEntry` picks the first
healthy **direct** entry *in registry order*. That is arbitrary: the card could silently
summarize a feature branch's planning tree rather than the main one.

**Measured 2026-08-23: this cannot fire today.** All six registered spaces are `base`
checkouts; none is a worktree. Note also that pointer repos are structurally immune —
`desirelines-deploy` sits on a feature branch but resolves to `desirelines-planning/planning`
regardless, because its branch does not change where `planning_repo` points. Only worktrees
of a *direct* planning checkout are exposed.

So deferral is free, and clarifying: "which tree does this card summarize" is a different
question from "what shape is this card."

**The condition.** Deferring a *display element* in a width-starved layout is the one kind
of deferral that costs. Retrofitting a branch/worktree marker into a full 48-column tile
could force a re-layout. Reserve the slot now — the tile's title row, right-aligned, where
the sketch put the age — and leave it empty. Costs nothing today; saves a redesign later.

## Recommendation (2026-08-23)

Sequential, not parallel. Running the layout rewrite and the new view together destroys the
iteration signal: if the atlas feels better afterwards, nothing tells you which half did it
— and the expensive rendering rewrite gets paid before learning whether the cheap useful
half was the whole answer.

1. **Cross-space work view** — a view cycle plus a full-viewport in-progress list, built
   against today's list layout. No banner. Ships the payload with no rendering rewrite.
2. **Live with it**, and let that answer whether a banner earns its rows.
3. **Tiles for the spaces view** — option (1) above, plus per-space accents, plus the
   reserved worktree slot. The delight, once the useful half is already earning its keep.
4. **Worktree-aware entry selection** — correctness, then badges.

The triage the sequence encodes:

- **Useful** (answers something unaskable today): the work view · cross-space acute
  findings · worktree-aware selection.
- **Pragmatic** (cheap, reuses what exists): selected-entry-only tiles · `s.miniBar` per
  card · aggregate counts.
- **Delightful** (makes you want to open it): accents · tile borders · naming the single
  most urgent thing across every space.

The trap this ordering avoids is spending the delight budget on tile-grid mechanics before
the useful part exists.

## Open forks

1. **Banner, rail, or both?** Deliberately unanswered — step 2 above exists to answer it
   empirically.
2. **Two views or three?** A third, "Attention" (acute findings, ready-to-close audits,
   revisits due, unreadable files across all spaces), is tempting and the data is already
   there. Three views is roughly where a toggle starts needing its own help entry.
3. **Per-space accent derivation.** `core.SpaceEntryPoint.Accent` exists and is unused;
   epic 29 currently lists it as out of scope. Tiles are what make an accent pay off, so
   this reopens on purpose or not at all.

## Addendum, 2026-08-23 — the atlas is under-composed, not under-designed

Both views shipped flat, and reviewing them against the rest of the TUI shows why. This
codebase already has a widget vocabulary, and the dashboard — two keystrokes away — uses
all of it. The atlas uses almost none.

| Available | Dashboard | Atlas |
| --- | --- | --- |
| `s.miniBar(pct, w)` gradient bar | yes | no |
| `s.segBar(done, active, dropped, total, w)` | yes | no |
| `s.priorityText` / `s.statusText` | yes | no |
| `theme.Breakdown` (`N key · N key · +N more`) | yes | no |
| `relDateCells` (aligned, dimmed date column) | yes | hand-rolled |
| `countsWidth` + `rollupCounts` (aligned `6/14`) | yes | no |

The data gap is worse. Per space the atlas ignores aggregate `Done`/`Total`,
`Findings.Acute`, `ReadyToClose`, `RevisitDue`, `Problems`, and `Epics[].LastUpdated`. Per
task the work view ignores `Epic`, `Priority`, `Effort`, and `StartedAt`.

So "more delight" and "more pragmatism" are not in tension here. The flatness comes from
hand-rolling a text run instead of composing widgets that already exist and are already
theme-aware and already tested. Several of these changes make the code *smaller*.

### Spaces: a comparable table, and the entry-point problem solved properly

The dashboard's epic row is `{glyph} {bar} {%} {done/total} {date} {id}`, with columns
measured across the visible set. The same idiom applied to spaces:

```
  ● bike-workshop   ████░░░░░░  43%   6/14   ▸2   4 epics · 1 audit   ⚠3   1mo ago
› ● kitchen         ████░░░░░░  41%   9/22   ▸3   3 epics · 1 audit   ⚠2   yesterday
```

Two wins beyond appearance: spaces become **directly comparable** — the bar column is
scannable — and attention gets a home.

**`⚠` is a deliberate fold** of acute findings, audits ready to close, snoozed tasks now
due, and unreadable files. Confirmed 2026-08-23 as the right call: the atlas is an
aggregate surface, and digging in is what the overview is for. It is one number because
four columns of mostly-zeroes would be worse than one number that is usually absent.

**Entry points move to a pinned detail band** below the table, showing the focused space's
entries with `h`/`l` cycling there. This is a better answer than anything in the tile
analysis above: the table scrolls, the band does not move, and it reuses the pinned
header/status split already built for finding H2 rather than adding machinery. It also
retires the awkward choice between hiding entry points and letting them blow up a row's
height — a table row can be as wide as the terminal, and the band as tall as it needs.

### Work: structure, and the signal nothing surfaces

```
› ● bake-the-crackly-rice-flour-baguette   kitchen   02-banh-mi  high   started yesterday
  ● overhaul-bottom-bracket        bike-workshop     01-touring  medium started 1mo ago
```

**The largest single win is `started`, not `updated`.** The view currently shows
last-touched. For in-progress work the useful question is how long it has been in
progress — "started 5 weeks ago" is what says a thing is stuck — and `StartedAt` is
already stamped on every task that entered in-progress. Nothing in the tool surfaces it.
Colouring it by staleness (a small `theme.Staleness`, exactly analogous to the existing
`theme.Percent`) turns an inventory into a health signal.

`o`/`O` should also work here, cycling `started` (default) / `space` (grouped) /
`priority`. Grouping by space removes the repeated space label — noise becomes structure —
and earns its keep once several spaces are active.

### What this does to the tile decision

Most of what the tile task was carrying — bars, richer cards, visual weight — lands in the
**table** for a fraction of the cost, with no grid reflow math and no 48-column squeeze.
The 81-column entry-point arithmetic that ruled tiles out does not apply to a table: a row
can be as wide as the terminal.

So the tile question is **deliberately reopened rather than settled**: after the table
exists, either tiles shrink to a pure layout change over already-rich rows, or they stop
being necessary. Finding that out is cheaper than committing to the grid now.

### Revised sequence

1. Spaces table + attention fold + pinned entry band — the biggest visible win, almost
   entirely composition of existing widgets.
2. Work view structure — `started`/staleness, epic, priority, orders, grouping.
3. Re-decide tiles with the table in front of you.
