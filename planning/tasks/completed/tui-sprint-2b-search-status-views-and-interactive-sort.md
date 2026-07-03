---
status: completed
epic: 18-tui-bubble-tea-interactive-planning-browser
description: 'Persistent / filter chip, : status/archived views with toggle, k9s-style sort with header indicator, and detail-pane vim find (n/N)'
effort: Unknown
tier: 3
priority: medium
autonomy_level: 3
tags: [tui, bubble-tea]
created: "2026-06-11"
updated_at: "2026-06-11"
started_at: "2026-06-11"
completed_at: "2026-06-11"
id: 6fb7ym400n5q
---

# TUI sprint 2b search, status views, and interactive sort

> ℹ️ **Proposed S2 split (filed this session).** Second half of the original
> [[tui-sprint-2-multi-entity-navigation-and-search]], which grew large. Builds
> on [[tui-sprint-2a-entity-navigation-registry-tabs-command-jump]] (the registry
> + per-entity lists). The implementing agent should confirm scope at sprint
> start; the original task is deprecated, not lost.

## Objective

The refinement half of S2: search, scope, and sort **within** the current
entity's list (plus find within a task body). 2a lets you move between entities;
this makes each list usable at scale. Decisions in
[[18-tui-bubble-tea-interactive-planning-browser]].

## Scope

- [x] **Extend the list `/` filter** — the built-in fuzzy filter landed in S1
      over slug+description; broaden the `FilterValue` to slug/description/tags.
- [x] **Persistent filter chip** — once a `/` filter is applied, show
      `filter: «keyword»` in the header until cleared with `Esc` (don't drop the
      active filter when leaving filter mode). Same chip for an active status view
      (`view: completed`).
- [x] **`:` status / archived views** — register status targets onto 2a's `:`
      surface (`:in-progress`, `:completed`, `:deprecated`, `:all`) so archived
      tasks are reachable without cluttering the default working-set view; a quick
      `s` toggle cycles (capital `S` reverse) over the same filtered list.
- [x] **Interactive sort (k9s-style)** — sort the current list by **status**
      (working-set default), **priority**, **updated** (recency), **tier**, or
      **slug**; sort key + indicator in the header; a pure in-memory reorder that
      **persists per entity tab**; reverse toggle. Extensible so new columns are
      cheap.
- [x] **Detail-pane vim find** — when the detail pane is focused, `/` searches
      the task body, `n`/`N` jump next/prev, matches highlighted, the viewport
      scrolls to the current match, `Esc` clears. `bubbles/viewport` has no
      built-in search, so this is real work: track match offsets in the rendered
      content, style them, scroll-to-match. List `/` vs detail `/` dispatched by
      focus. *(The fiddliest item — fine to split to a 2c if it runs long.)*

## Acceptance

- [x] `/` filters instantly and the active filter stays visible until cleared;
      reach archived tasks via `:completed`/`:all`; sort by
      status/priority/updated/slug (reversible) with a visible indicator;
      sort/filter persist per entity tab; detail `/`+`n`/`N` finds within a body.
      Suite + lint green.

## Dependencies / ordering

- Requires [[tui-sprint-2a-entity-navigation-registry-tabs-command-jump]] (the
  registry, per-entity lists, and the `:` input surface).
- Independent of S3/S4.

## Out of scope

- The registry / entity switching / tabs (2a).
- Live reload (S3); mutations (S4).

## Follow-ups folded in from S2a review (2026-06-11)

Small S2a leftovers that land naturally here:

- **Redundant per-list title** — the `bubbles/list` title ("Tasks") now duplicates
  the tab strip. Left in S2a to keep the `/` filter prompt rendering; revisit when
  reworking the list header for the filter chip (hide the title, or fold the chip
  into it). Reclaims 2 rows of list height.
- **Filter-to-empty leaves a stale detail** — when `/` narrows the list to zero
  matches, the detail pane keeps showing the last item (selection becomes empty
  but `updateList` only reloads on a non-empty change). Call `refreshDetail()` (it
  already handles the empty → "(nothing selected)" case) when the filtered
  selection goes empty.
- **Debounce detail loads** (low) — holding `j` fires one `ShowTask/ShowEpic/...`
  service read per row crossed; the `isCurrentSelection` guard drops all but the
  last result, but the I/O still happens. On a slow FS this lags scrolling.
  Debounce with a `tea.Tick` keyed by a generation counter (load only after the
  cursor settles).
- **Tab-completion ignores aliases** (low) — `complete()` is fed `entityNames()`
  (canonical only), while `:` accepts `t/e/a/task/…` via `matches()`. Harmless
  today (single-letter aliases are name prefixes), but a future entity whose alias
  isn't a name prefix won't Tab-complete. Feed the alias set too when it matters.
- **Epic rollup computed twice** (low) — the list row uses
  `core.EpicSummary.Percent()`; `renderEpicDetail` re-derives done/total/pct from
  its own task slice. They agree today (same join + `StatusCompleted` rule), but
  the "done" definition now lives in 3 places. Thread the `EpicSummary` (or a
  shared count helper) into the detail so they can't diverge.

## Progress Log

### 2026-06-11 — implemented (suite + lint green)

All scope landed, vim find included (it stayed tractable — no 2c needed):

- **Filter broadened** — `taskItem`/`epicItem` `FilterValue` now includes tags, so
  `/go` narrows by tag (`item.go`).
- **Status views** — `:` accepts `all`/`active`/any status word, and `s`/`S` cycle
  the tasks tab through them (`statusCycle`). The task loader reads the tab's
  `statusView` and re-queries via `core.TaskFilter` (`commands.go`); the loader
  signature changed to `loadList(*entityTab, *core.Service)` so it can. Working-set
  order applies only to the default view.
- **State chip** — the list's title slot now carries `view:… / sort:…↓ / filter:…`
  (`entityTab.chip()`), collapsing to nothing in the clean default — which also
  resolves the redundant-"Tasks"-title follow-up (and the `/` prompt still renders
  there). The persistent filter shows because `bubbles/list` keeps
  `FilterState==FilterApplied` after Enter.
- **Interactive sort** — `o` cycles status→priority→updated→tier→slug, `O` reverses;
  a generic comparator over a new `entityItem.sortFields()` (`sort.go`), reordering
  the loaded slice in place and persisting per tab (re-applied on every reload).
- **Detail vim find** — `/` in the detail pane opens a find input; `n`/`N` walk
  matches (highlighted; current one brighter), the viewport scrolls to the current
  match, `Esc` clears (first clears find, second leaves the pane). Matches tracked
  by wrapped-line index; matched lines rebuilt from stripped text so a highlight
  never splits an existing escape (`find.go`, `detail.go`). Find persists across
  selections (vim-like).
- **`?` help overlay** (requested mid-sprint) — a floating keybinding panel
  composited over the body via `ansi.Cut` (items stay visible around it), dismissed
  by any key (`help.go`).
- Folded S2a follow-ups: redundant title (done, via the chip), filter-to-empty
  stale detail (done — `detail.showEmpty()`), alias-aware `:` completion (done —
  `commandOptions()`).
- Tests: sort+chip, status cycle/filter, `:`-command status view, tag filter, help
  overlay (+ layout invariant with it open), detail find nav/scroll/clear,
  `highlightOccurrences`. Layout invariant (`ViewFitsTerminal`) still holds.

**Deferred (both low, noted for a later cleanup or S3):** debounce detail loads on
fast scroll; dedupe the epic done/total rollup (computed in row + detail). Neither
blocks S2b acceptance.

## Related

- Epic [[18-tui-bubble-tea-interactive-planning-browser]]
- Sibling [[tui-sprint-2a-entity-navigation-registry-tabs-command-jump]]
- Supersedes (with 2a) [[tui-sprint-2-multi-entity-navigation-and-search]]
