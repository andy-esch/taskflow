---
status: planning
description: Bubble Tea TUI (tskflwctl ui), a second primary adapter over core.Service; read-only browser first, mutations later
priority: high
tags: [tui, bubble-tea, ux]
created: "2026-06-10"
---

# TUI Bubble Tea interactive planning browser

**Goal.** Bubble Tea TUI (tskflwctl ui), a second primary adapter over core.Service; read-only browser first, mutations later

## Why this is its own epic

The TUI is a **second primary adapter** over the same `core.Service` the CLI uses
— a distinct, multi-sprint surface (new package, Bubble Tea event model, new
testing style) that the dashboard sprint deliberately set up (`Summary()` + the
render concepts). Epic 17 is the CLI port; this is the interactive front-end.

## Architecture (non-negotiable)

- TUI goes through `core.Service` (`ListTasks`/`ShowTask`/`ListEpics`/
  `ListAudits`/`Summary`, later `Move`/`SetFields`) — **never `store`/the fs**.
  `Service` re-scans on each call (no cache), so all reads run as `tea.Cmd`s
  returning custom `tea.Msg`s — **never I/O in `Update`/`View`**.
- Launched by `tskflwctl ui` (a cobra command in `root.go`); `app.Svc` is passed
  to `tui.Run(svc, cfg.Root)`. `Cfg.Root` is for fsnotify only.
- New `internal/tui/` package (replace the 53-line stub): root `Model` owns
  `svc` + size + focus + tab; per-pane sub-models (list/detail/search) are pure
  view/state and never see the service.

## Decisions (locked 2026-06-10, after two research agents)

- **Entity switching is a `:` command-jump (k9s-style), not tabs.** A thin tab
  strip + `[`/`]` is the *discoverable* affordance; `:tasks`/`:epics`/`:audits`
  (+ `:in-progress` filters) is the muscle-memory path. Both read from one
  **entity registry**, so adding Projects/ADRs/Research later is a registration,
  not new keybindings or layout cost. (Answers the "extensible to ~6 entities"
  ask.)
- **Vim-first keys** (maintainer knows k9s): global `: / ? Esc Tab/Shift-Tab
  [ ] q Ctrl+c`; list `j/k Ctrl+d/u g/G Enter|l h s`; detail scroll + `h/Esc`.
  Keys route through the **focused pane only** (so `j/k` = move in list, scroll
  in detail). **Don't bind `1`–`6` to tabs** (lazygit's documented regret).
- **`bubbles`**: `list` (550 items fine; custom `ItemDelegate` for the glyph
  row), `viewport` (detail; wrap manually to inner width), `textinput` (`/`;
  gate routing while focused), `help` (footer), `key.Binding`, `spinner`
  (only while loading). New deps: `bubbles`, `x/teatest`, `fsnotify`.
- **Shared theme, not shared rendering.** Extract a dependency-free `theme`
  package (status/bucket/priority → `{Glyph, Color hex}`, imports only `domain`)
  consumed by **both** the CLI `render` (→ ANSI) and the TUI (→ lipgloss). One
  place decides "in-progress is yellow ●".
- **Focus = two signals** (accent border + pane-title marker). **Responsive:**
  ≥100 cols two-pane (40/60), 60–99 single-pane drill, <60 collapse the tab
  strip. Truncate, never wrap, in bordered panes; subtract border/padding frame
  from child sizes (the #1 lipgloss sizing bug).
- **Surface `Task.Misfiled()` as a `⚠` row glyph** — a free correctness signal.
- **List UX (from use, 2026-06-10):** default **working-set order**
  (in-progress → next-up → ready-to-start), not directory order (S1);
  **interactive k9s-style sort** by status/priority/updated/tier/slug with a
  header indicator + reverse (S2); a **persistent filter/view chip** that stays
  visible until `Esc` (S2); **`:` status views** (`:completed`/`:all`) so
  archived tasks are reachable without cluttering the default (S2).
- **Testing:** ~80% message-injection unit tests (send msgs to `Update`, assert
  state/`View()`; `core.Store` is an interface → fake store), ~20% `teatest`
  golden-view tests for layout regressions.

## Sprint roadmap

- **S0 — foundation** [[tui-sprint-0-foundation-shared-theme-test-harness]]:
  clear the stub, extract `theme`, add `ui` command + model skeleton + resize +
  `q`, async-load one task list, test harness. *Proves the adapter wiring.*
- **S1 — read-only browser** [[tui-sprint-1-read-only-two-pane-task-browser]]:
  two-pane tasks list + detail (frontmatter + body via `ShowTask`), vim nav,
  focus highlight, lazy body + spinner, responsive, empty/no-repo states,
  manual `r` refresh.
- **S2 — multi-entity + search** *(split 2026-06-11; the original
  [[tui-sprint-2-multi-entity-navigation-and-search]] grew large)*:
  - **S2a — entity navigation**
    [[tui-sprint-2a-entity-navigation-registry-tabs-command-jump]]: entity
    registry + `:` jump + tab strip (tasks/epics/audits), per-entity lists.
  - **S2b — search, views & sort**
    [[tui-sprint-2b-search-status-views-and-interactive-sort]]: `/` filter +
    persistent chip, `:` status/archived views, interactive sort, detail vim find.
- **S3 — live reload** [[tui-sprint-3-fsnotify-live-reload]]: fsnotify watch
  (debounced) → reload preserving cursor by slug. (~½–1 day; plumb the
  `reloadMsg` in S1.)
- **S4 — mutations** [[tui-sprint-4-mutations-and-actions]]: lifecycle via
  `Service.Move`/`SetFields` with confirmation; reconsider multi-select then.
- **S5 — pretty markdown** [[tui-glamour-markdown-rendering-with-rawpretty-toggle]]:
  glamour-render the body (cached, never in `View()`) with an `R` raw/pretty toggle.
- **S6 — cross-link navigation** [[tui-cross-link-navigation-between-epics-and-tasks]]:
  follow epic↔task references with a jump + back-stack (body wikilinks/peek deferred).
- **Polish** [[tui-s2b-polish-find-occurrences-highlight-fidelity-per-entity-sort]]:
  S2b fresh-eyes follow-ups (occurrence-level find, highlight fidelity, per-entity sort).

## References

- `research/2026-06-09-tui-ux-design-and-navigation-spec.md` — the input UX spec
  (good bones; over-reached on Projects-tab + multi-select — superseded by the
  decisions above).
- `research/2026-06-10-tui-design-decisions.md` — the **build reference**
  (full keybinding matrix, package structure, bubbles verdicts, testing,
  footguns), distilled from two 2026-06-10 research agents (UX patterns from
  k9s/lazygit/gh-dash/gitui; Bubble Tea architecture/testing). Includes the
  **Layout discipline** checklist (audited 2026-06-11) every pane must follow —
  locked by `TestModel_ViewFitsTerminal`.
- Epic [[17-pm-go-cli]] (the CLI port this builds on).

## Out of scope

- Projects/ADRs/Research entities (don't exist yet — the registry is built to
  accept them; the entities ship with their features).
- Multi-select / bulk actions until a concrete bulk need appears (S4+).
- Mouse support (keyboard-first).
