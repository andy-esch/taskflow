---
status: completed
epic: 18-tui-bubble-tea-interactive-planning-browser
description: 'Entity registry plus : command-jump and tab strip to switch tasks/epics/audits, each its own list with per-tab cursor and delegates'
effort: Unknown
tier: 3
priority: medium
autonomy_level: 3
tags: [tui, bubble-tea]
created: "2026-06-11"
updated_at: "2026-06-11"
started_at: "2026-06-11"
completed_at: "2026-06-11"
id: 6fb7ym401n1y
---

# TUI sprint 2a entity navigation (registry, tabs, command-jump)

> ℹ️ **Proposed S2 split (filed this session).** The original
> [tui-sprint-2-multi-entity-navigation-and-search](6faxn1802k3e-tui-sprint-2-multi-entity-navigation-and-search.md) grew large, so it's split
> into 2a (this — structure) and [tui-sprint-2b-search-status-views-and-interactive-sort](6fb7ym400n5q-tui-sprint-2b-search-status-views-and-interactive-sort.md)
> (refinement). The implementing agent should confirm the seam at sprint start;
> the original task is deprecated, not lost.

## Objective

The structural half of S2: make the browser **multi-entity**. Build the registry
+ switching surface that everything in 2b (filter/sort/status-views) then
operates on. After this you can move between tasks/epics/audits; after 2b you can
search/sort/scope within them. Decisions in
[18-tui-bubble-tea-interactive-planning-browser](../epics/18-tui-bubble-tea-interactive-planning-browser.md).

## Scope

- [x] **Entity registry** — a list of `{name, load Cmd, item delegate}` so
      entities are data-driven (tasks/epics/audits now; Projects/ADRs/Research
      register later with zero keybinding/layout change). This is the keystone
      abstraction; build it first so 2b's per-entity state hangs off it.
- [x] **`:` command-jump (entity targets)** — `textinput`-backed `:tasks` /
      `:epics` / `:audits` with completion; key routing gated while the input is
      focused (mirror the existing `SettingFilter()` guard in `model.go`). The
      `:` infra lands here; 2b registers status-view targets onto it.
- [x] **Tab strip + `[`/`]` cycle** — the discoverable affordance alongside `:`;
      collapses to `[Entity ▾]` under ~60 cols (reuse the responsive thresholds).
- [x] **Per-entity lists** — one `list.Model` per entity with its own
      `ItemDelegate` and a **per-tab cursor**. Epics delegate → rollup bars
      (reuse `Style.Bar`/`Summary` semantics); audits delegate → finding counts.
- [x] Loads run as `tea.Cmd`s per the registry (no I/O in `Update`/`View`),
      reusing the S1 async/stale-guard pattern.

## Acceptance

- [x] Switch entities via `:tasks`/`:epics`/`:audits`, the tab strip, and
      `[`/`]`; each lists correctly with its own delegate; the cursor is
      preserved per tab; the tab strip collapses under ~60 cols. Suite + lint
      green.

## Done (2026-06-11)

Implemented this session; full suite + lint green, rendered against the real
planning at 118×24 (tasks + epics tabs verified visually).

- **Entity registry** (`entity.go`): `entityKind` + `entityTab{name, list,
  loadList, loadItem, loaded, problems}`; `newEntityTabs()` declares
  tasks/epics/audits. Adding Projects/ADRs later = one entry. An `entityItem`
  interface (`list.Item` + `id()`) lets the model preserve cursors and
  stale-guard detail loads generically across entities.
- **Per-entity delegates** (`item.go`): shared `row()` cursor convention +
  `taskDelegate` (glyph/⚠/slug/date), `epicDelegate` (rollup bar + colored
  percent + done/total + id/desc, via a new `miniBar` mirroring
  `render.Style.Bar`), `auditDelegate` (bucket marker + slug + open/total).
- **Content-agnostic detail** (`detail.go`): `detailContent{Title; Render(width)}`
  with task/epic/audit renderers; the pane re-wraps on resize. The right pane now
  works on every tab (not just tasks).
- **`:` command-jump** (`command.go`): `bubbles/textinput` bar; captures all keys
  while open (global hotkeys gated, like the `/` filter), Tab-completes entity
  names (longest-unique-prefix), Enter dispatches, unknown reopens with an inline
  error.
- **Tab strip + `[`/`]`** (`model.go`): accent-highlighted strip above the panes,
  cycles with `[`/`]`, collapses to `[entity ▾]` under 60 cols. Layout reworked
  to tab-strip(1) + panes + footer(1); every tab's list is pre-sized so a switch
  needs no relayout. Per-tab cursor preserved (each tab owns its `list.Model`).
- **Loaders** (`commands.go`): per-entity list + detail Cmds returning
  `listLoadedMsg{kind,…}` / `detailMsg{kind,id,content}`; the stale-guard keys on
  **kind + id**. A background-tab load lands in the right tab even after a switch.
- **Tests** (`model_test.go`): seeded an epic + audit; added cycle-loads-entity,
  `:`-jump, command-bar capture/completion, unknown-command error, per-tab cursor
  preservation, and tab-strip collapse — alongside the migrated S1 suite.

**Deferred to S2b (unchanged):** filter chip, `:` status views, sort, detail vim
find. **Known minor:** the per-list title ("Tasks") is now redundant with the tab
strip — left as-is to keep the `/` filter prompt working; tidy in S2b.

### Post-implementation fixes (2026-06-11, from live use)

- **Chrome-cropping layout bug** (tab strip + `:` command bar vanished on the
  tasks tab at shorter heights, fine on sparse epics/audits). Root cause:
  `bubbles/list` renders its pagination `••` row **one line beyond** its
  `SetHeight`, so a paginated list overflowed its pane and the outer height clamp
  ate the footer. Fixed at two levels: (1) **reserve the pagination line** in the
  list's height budget; (2) **hard-clamp the body region** to its budget so the
  tab strip and command line — the load-bearing chrome — can never be pushed
  off-screen, whatever a child renders. Regression test
  `TestModel_ChromeVisibleWhenListPaginates` pins it (20 tasks at 100×14).
- **`:` shorthands** — `:t`/`:e`/`:a` (and `:task`/`:epic`/`:audit`) via a
  per-entity `aliases` list; future entities just declare their own (e.g.
  `adr`). Tab-completion still completes to the canonical name. Tested in
  `TestModel_CommandAliases`.

### Self-review fixes (2026-06-11)

- **Empty tab no longer sits on a perpetual "loading…"** — this repo has no
  `audits/`, so `:a` showed an empty list with the detail stuck loading forever.
  Extracted `refreshDetail()` (loads the selection, or settles to "(nothing
  selected)" when the tab is empty) and routed switch/reload through it. Test:
  `TestModel_EmptyTabShowsNothingSelected`.
- **De-duped the async stale-guard** into `isCurrentSelection(kind, id)` (was
  inline in both `detailMsg` and `detailErrMsg`).
- **Added renderer coverage** (`TestEntityDetailRenderers`) — the epic/audit
  detail renderers (rollup %, finding counts) had no direct test.

Out-of-scope findings folded forward:
[tui-sprint-2b-search-status-views-and-interactive-sort](6fb7ym400n5q-tui-sprint-2b-search-status-views-and-interactive-sort.md) (redundant list title,
filter-to-empty stale detail), [tui-sprint-3-fsnotify-live-reload](6faxn1800qb2-tui-sprint-3-fsnotify-live-reload.md) (reload all
loaded tabs, not just the active one),
[clirender-polish-batch-audit-flags-color-tty-wide-char-width](6fb7ym4038pj-clirender-polish-batch-audit-flags-color-tty-wide-char-width.md) (dedup the
progress-bar math into `theme.BarFill`).

### Second review pass — two subagents (2026-06-11)

Independent adversarial review (TUI-correctness lens + Go-idiom/boundary lens);
every finding re-verified before acting. Fixed in-scope:

- **Long detail title broke the pane border** (chrome corruption). The title was
  the one Join input not run through `truncate`; a slug wider than the (narrow,
  two-pane) detail pane wrapped to 2 rows, grew the pane past budget, and the
  body clamp ate its bottom border. Now truncated to the pane width. Pinned by
  `TestModel_LongTitleKeepsDetailBorder` (50+ char slug at 90 cols).
- **Fatal `errMsg` never cleared** → a transient load failure bricked the session
  permanently (even a later successful `r` left the error screen). Now cleared on
  any successful list load. Test `TestModel_RecoversFromFatalError`.
- **`selectByID` used an unfiltered index with `list.Select`**, which indexes the
  *visible* (filtered) set — so cursor restore landed on the wrong row when a `/`
  filter was active. Now ranges `VisibleItems()`. Matters for S3 (reload while
  filtered); confirmed against the `bubbles` source.
- **Single reload path** — `r` now emits `reloadMsg` (was a duplicate inline of
  the dead `reloadMsg` case); fsnotify reuses the same seam in S3. Test
  `TestModel_RefreshFiresReloadMsg`.
- **Dedup**: `taskDate` (was copy-pasted in `tui/item.go` and CLI `render.go`) →
  shared `theme.TaskDate`; `detailLabel` → the existing `dimStyle`.
- **`theme` package test coverage** — added `theme_test.go` pinning the shared
  glyph/color decision table (status/bucket/priority + the 34/100 percent edges),
  the one package whose justification is "single source of truth."
- Marked `Model.root` as reserved-for-S3 (was an unexplained unread field).

Low-severity findings folded to
[tui-sprint-2b-search-status-views-and-interactive-sort](6fb7ym400n5q-tui-sprint-2b-search-status-views-and-interactive-sort.md) (debounce detail
loads, alias-aware Tab-completion, epic-rollup computed twice). Reviewers
confirmed clean: boundary hygiene (no `store`/fs in non-test TUI code), the
value-receiver/pointer-helper mutation pattern, the async stale guards, layout
math, and key-routing gating order.

## Dependencies / ordering

- Lands before [tui-sprint-2b-search-status-views-and-interactive-sort](6fb7ym400n5q-tui-sprint-2b-search-status-views-and-interactive-sort.md) (2b's
  filter-chip, status views, and per-entity sort assume the registry + per-entity
  lists exist).
- Independent of S3 (live reload) and S4 (mutations).

## Out of scope (here — see 2b)

- The persistent filter chip, `:` **status** views (`:completed`/`:all`), the
  `s` status toggle, interactive sort, and detail-pane vim find — all in 2b.
- Body-content search beyond the detail find (deferred); live reload (S3);
  mutations (S4).

## Related

- Epic [18-tui-bubble-tea-interactive-planning-browser](../epics/18-tui-bubble-tea-interactive-planning-browser.md)
- Sibling [tui-sprint-2b-search-status-views-and-interactive-sort](6fb7ym400n5q-tui-sprint-2b-search-status-views-and-interactive-sort.md)
- Supersedes (with 2b) [tui-sprint-2-multi-entity-navigation-and-search](6faxn1802k3e-tui-sprint-2-multi-entity-navigation-and-search.md)
