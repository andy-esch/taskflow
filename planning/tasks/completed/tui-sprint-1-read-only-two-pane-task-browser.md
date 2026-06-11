---
status: completed
epic: 18-tui-bubble-tea-interactive-planning-browser
description: Two-pane task list plus detail preview with vim navigation, focus highlighting, async load, and lazy body loading
effort: Unknown
tier: 3
priority: medium
autonomy_level: 3
tags: [tui, bubble-tea]
created: "2026-06-10"
updated_at: "2026-06-11"
started_at: "2026-06-10"
completed_at: "2026-06-10"
---

# TUI sprint 1 read-only two-pane task browser

## Objective

The v1 milestone: a genuinely usable read-only task browser — list + preview +
vim navigation. Decisions/architecture in
[[18-tui-bubble-tea-interactive-planning-browser]].

## Scope

- [ ] **Two panes:** left = task list (custom `list.ItemDelegate`: status glyph
      + slug + relative date + `⚠` if `Misfiled()`), right = detail (viewport)
      showing frontmatter fields + markdown body (plain text, wrapped to inner
      width).
- [ ] **Default order = working-set, not directory order:** in-progress →
      next-up → ready-to-start (then the rest). A "what am I doing" view leads
      with active work, not the backlog. (Service returns dir order; the TUI
      sorts for display. Consider whether the CLI `task list` should adopt the
      same default — likely yes, as a follow-on.)
- [ ] **Nav:** `j/k`, `Ctrl+d/u`, `g/G`; `Enter`/`l` focus detail; `h`/`Esc`
      back; keys route to the **focused pane only**. `q` context-quit,
      `Ctrl+c` hard.
- [ ] **Focus = two signals** (accent border + pane-title marker); selected row
      reverse-video when focused, dim when not.
- [ ] **Async + lazy:** load body via `ShowTask` Cmd on selection change;
      `spinner` while loading; **stale-result guard by slug**.
- [ ] **Responsive:** `WindowSizeMsg` → two-pane ≥100 cols, single-pane drill
      60–99; subtract border/padding frame from child sizes; truncate rows.
- [ ] **Footer:** `bubbles/help` driven by `(focus)` — context-sensitive hints.
- [ ] **States:** empty repo, no in-progress, unreadable files (surface the
      `FileProblem`s), and a manual `r` refresh. Plumb a `reloadMsg` path now
      (fsnotify wired in S3).
- [ ] Tests: selection-loads-body, focus routing, resize sizing (unit); one
      `teatest` golden at 120×40.

## Acceptance

- [x] Browse tasks, scroll a task's body, navigate by keyboard, resize without
      breakage; `r` refreshes. Suite + lint green; rendered against the real
      ~7-task planning at 118×22 (two bordered panes, working-set order).

## Done (2026-06-10)

- **Shared-primitive move:** `RelativeDate` → `theme` (so the TUI doesn't import
  the CLI's render package); CLI stays byte-stable.
- **Two-pane browser** (`internal/tui`): `bubbles/list` (custom delegate: glyph
  + ⚠-if-misfiled + slug + relative date), `viewport` detail (frontmatter fields
  + plain markdown body, wrapped). **Working-set order** (in-progress → next-up →
  ready-to-start) via `sortWorkingSet`.
- **Focus** = accent border + bold title; `l`/`Enter`/`Tab` → detail, `h`/`Esc`
  → list; keys route to the focused pane only (`j/k/g/G` to the list,
  `j/k`/`ctrl+d/u` to the viewport).
- **Async + lazy:** body loads via a `loadBody` Cmd on selection change, with a
  **stale-result guard by slug**. **Responsive:** two-pane ≥90 cols, single-pane
  drill below; borders subtracted from child sizes; tiny terminals don't panic.
- **States:** loading, error, empty (`No active tasks…`), unreadable-file count
  in the footer. **`r` refresh** preserves the cursor by slug; the `reloadMsg`
  path is plumbed for S3's fsnotify.
- Tests: working-set order, selection-loads-body + stale guard, focus routing,
  responsive sizing, `teatest` quit. Full suite + lint green.

## Post-completion fixes (2026-06-10, from live use)

- **Per-task errors no longer brick the UI:** a duplicate slug across status
  dirs made `ShowTask` return `ErrAmbiguous`, which had blanked the whole screen.
  Body-load failures now surface in the detail pane (`bodyErrMsg`); the list
  stays navigable.
- **`g`/`G` scroll the detail pane** (viewport has no default binding; handled
  explicitly).
- **`/` filter enabled early:** the list's built-in fuzzy filter now works (was
  deliberately off for S1). Also fixed a real bug — the model dropped every
  unhandled message, so the list's async `FilterMatchesMsg` never applied;
  unhandled messages now forward to the list. (S2 adds the persistent chip,
  detail-pane vim find, status views, sort.)
- Tests added for all three; full suite + lint green.
- **Layout audit (2026-06-11):** a clipped-top-border report traced to rendering
  before the first `WindowSizeMsg` (negative frame dims corrupt height
  tracking). Hardened per the **Layout discipline** checklist now in the build
  reference: size guard in `View()`, frame sizes from `GetFrameSize()` (not a
  hardcoded `2`), ANSI/width-aware `truncate`, and a final `MaxWidth/MaxHeight`
  clamp. Locked by `TestModel_ViewFitsTerminal` (View() == terminal height,
  every line ≤ width). Validated by a research agent as idiomatic.

## Deferred to later sprints (noted in S2)

- Animated spinner (loads are instant), `bubbles/help` footer, single-pane
  *drill* polish, and the S2 list-UX (sort/filter-chip/status views).

## Out of scope

- Epics/audits tabs + search (S2); live reload (S3); any mutation (S4).

## Related

- Epic [[18-tui-bubble-tea-interactive-planning-browser]]
