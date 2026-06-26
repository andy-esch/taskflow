---
schema: 1
status: ready-to-start
epic: 18-tui-bubble-tea-interactive-planning-browser
description: 'Bring the CLI due-for-revisit surfacing to the TUI: an on-open banner, a non-emoji row marker, and due-deferred tasks sorted to the top. Several design forks left open (see task body).'
effort: M
tier: 3
priority: medium
autonomy_level: 3
tags: [tui]
created: "2026-06-26"
---

# Surface "due for revisit" deferred tasks in the TUI

## Objective

The CLI surfaces deferred tasks whose `revisit_at` has arrived — the `status`
nudge and `task list --revisit-due` — but the TUI, the human browse surface, has
ZERO awareness of it: no banner, no marker, no use of `domain.IsRevisitDue`. A
human sitting in the `:deferred` view can't tell which tasks have come due. Bring
the feature to the second primary adapter in three layers: **awareness** (a banner
on open) → **identify** (a per-row marker) → **prioritize** (due ones sorted to the
top).

Hard constraint that drives the open decisions: the default TUI task view is the
ACTIVE working set (in-progress → next-up → ready-to-start). Deferred tasks only
appear via `:deferred` / `:all` / the `s`/`S` cycle — they are NOT in the default
view. So "sort due ones to the top of the list" must pick *which* list.

## OPEN DECISIONS — resolve before/while building (defaults noted)

> These are the design forks from the discussion. Each has a recommended DEFAULT;
> confirm or flip before implementing the affected piece.

1. **[BIG FORK] Where do the row marker + sort-to-top apply?**
   - **(A) Keep the default active view pure; banner is the un-bury.** The banner
     shows on open regardless of view; a jump key takes you to a dedicated
     `:revisit` view (or the existing `:deferred` view) where due tasks are marked
     and sorted to the top. Preserves "view = status". — **DEFAULT / recommended.**
   - (B) Also pin due-deferred tasks to the top of the DEFAULT active list, so
     they are the first rows on open (zero keystrokes). Maximally un-buried, but
     injects a `deferred` row into "what am I working on" and muddies the
     working-set model; if chosen, style those rows so they read as
     snoozed-but-due, not active.

2. **Row indicator glyph — NO emoji (standing user preference).**
   - `◷` monochrome clock-face glyph (the "alarm clock" nod, not an emoji). — **DEFAULT.**
   - `due` short colored text tag (most unambiguous; renders everywhere).
   - `!` / `▲` terser alert markers.
   Render it in its OWN leading marker column so existing columns don't shift;
   color it with the existing warn style.

3. **Banner form.**
   - Sticky one-line header while the due count > 0, auto-hides at zero
     (can't-miss; costs one line of height). — **DEFAULT.**
   - Transient flash on open (free, but easy to miss).
   Wording e.g. "N deferred due for revisit — press <key> to review"; needs a jump key.

4. **Also de-emoji the CLI `status` nudge?** It still uses an ⏰ emoji
   (`internal/cli/render/render.go`). Swap it for the same monochrome marker for
   consistency with the no-emoji preference? — **DEFAULT: yes.**

## Acceptance criteria

- [ ] On open, the TUI shows a banner when ≥1 deferred task is due for revisit
      (`revisit_at` ≤ today via `domain.IsRevisitDue` + the injected core clock);
      it self-hides at zero. (form per decision #3)
- [ ] A non-emoji indicator marks due-for-revisit rows (per decision #2), in a
      leading column so the existing columns keep their positions.
- [ ] Due-for-revisit tasks sort to the top of the relevant view (per decision #1).
- [ ] A way to jump to the due tasks — a `:revisit` view and/or the banner's jump
      key — the TUI mirror of `task list --revisit-due`. (scope per decision #1)
- [ ] All read-side: reuses `domain.IsRevisitDue` + the core clock + `Summary`'s
      `revisit_due`; no new core mutations; the TUI keeps reading through
      `core.Service` (no store access in `Update`/`View`).
- [ ] If decision #4 = yes: the CLI `status` nudge uses the non-emoji marker;
      regenerate any affected goldens.
- [ ] Tests are deterministic (inject the clock): banner count, marker on due rows,
      sort order, and exclusion of not-due / no-date / non-deferred-with-stale-date.

## Out of scope

- Auto-resuming due tasks — the snooze stays nudge/surface-only; resume is manual
  (`task next`/`task ready`).
- Changing how `revisit_at` is set or cleared (already shipped).
- A persistent due / review-by date on active tasks — that's a separate field/concept,
  its own task.

## Related

- Epic [[18-tui-bubble-tea-interactive-planning-browser]]
- TUI mirror of [[add-a-revisit-due-filter-to-task-list-for-deferred-task-triage]]
- Builds on [[set-a-revisit-date-when-deferring-snooze-and-surface-what-is-due]]
