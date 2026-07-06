---
status: deprecated
epic: 18-tui-bubble-tea-interactive-planning-browser
description: Entity registry with command-jump plus tab strip for tasks epics audits, and in-memory filter search
effort: Unknown
tier: 3
priority: medium
autonomy_level: 3
tags: [tui, bubble-tea]
created: "2026-06-10"
updated_at: "2026-06-11"
deprecated_at: "2026-06-11"
id: 6faxn1802k3e
---

# TUI sprint 2 multi-entity navigation and search

> 🔀 **Deprecated 2026-06-11 — split into two stages** (this session) because the
> scope grew large. Superseded by:
> - [tui-sprint-2a-entity-navigation-registry-tabs-command-jump](6fb7ym401n1y-tui-sprint-2a-entity-navigation-registry-tabs-command-jump.md) — structure
>   (registry, tabs, `:` command-jump, per-entity lists).
> - [tui-sprint-2b-search-status-views-and-interactive-sort](6fb7ym400n5q-tui-sprint-2b-search-status-views-and-interactive-sort.md) — refinement
>   (`/` filter + chip, status views, interactive sort, detail vim find).
>
> The full scope below is preserved verbatim as the source for the split; the
> implementing agent should confirm the seam at sprint start.

## Objective

Make the browser multi-entity and searchable. (Refine at sprint start with S0/S1
learnings.) See [18-tui-bubble-tea-interactive-planning-browser](../epics/18-tui-bubble-tea-interactive-planning-browser.md).

## Scope

- [ ] **Entity registry** — a list of `{name, load Cmd, item delegate}` so
      entities are data-driven (tasks/epics/audits now; Projects/ADRs/Research
      register later with zero keybinding/layout change).
- [ ] **`:` command-jump** (primary, scales to 6+): `:tasks`/`:epics`/`:audits`
      and status filters (`:in-progress`), backed by the registry; `textinput`
      with completion. Plus a thin **tab strip** + `[`/`]` cycle as the
      discoverable affordance (collapses to `[Entity ▾]` under ~60 cols).
- [ ] Epics tab → rollup bars (reuse `Style.Bar`/`Summary` semantics); audits
      tab → finding counts. One `list.Model` per entity (per-tab cursor).
- [x] **`/` in the list = filter** — in-memory fuzzy filter over the loaded
      slice (slug/description); key routing gated while the input is focused.
      *(Landed early in S1 via `bubbles/list`; S2 extends it to slug/desc/tags +
      the chip below.)*
- [ ] **`/` in the detail = vim-like text find** (context-dependent, per the
      maintainer): when the detail pane is focused, `/` searches the task body,
      `n`/`N` jump to next/prev match, matches highlighted, viewport scrolls to
      the current match, `Esc` clears. `bubbles/viewport` has no built-in search,
      so this is a real feature (track match offsets in the rendered content,
      style them, scroll-to-match). The list `/` and detail `/` are dispatched by
      focus.
- [ ] **Persistent filter chip** — once a `/` filter is applied, show
      `filter: «keyword»` in the header until cleared with `Esc` (don't hide the
      active filter when leaving filter mode). Same for an active `:` status
      view (`view: completed`).
- [ ] **Status / archived views** — `:` jump to a status (`:in-progress`,
      `:completed`, `:deprecated`, `:all`) so archived tasks are reachable
      without cluttering the default active view; a quick toggle (e.g. `s`
      cycles, capital for reverse) over the same filtered list.
- [ ] **Interactive sort (k9s-style)** — sort the current list by **status**
      (the working-set default), **priority**, **updated** (recency), **tier**,
      or **slug**; a sort key + indicator in the header (k9s uses `Shift`-letter
      bindings, e.g. sort-by-age/name). Sort is a pure in-memory reorder of the
      loaded slice; persists per entity tab; reverse toggle. Extensible so new
      columns are easy.

## Acceptance

- [ ] Switch entities via `:`/tabs; each lists correctly; `/` filters instantly
      and the active filter stays visible until cleared; reach archived tasks via
      `:completed`/`:all`; sort by status/priority/updated/slug (reversible) with
      a visible indicator; cursor preserved per tab. Suite + lint green.

## Out of scope

- Body-content search (defer); live reload (S3); mutations (S4).

## Related

- Epic [18-tui-bubble-tea-interactive-planning-browser](../epics/18-tui-bubble-tea-interactive-planning-browser.md)
