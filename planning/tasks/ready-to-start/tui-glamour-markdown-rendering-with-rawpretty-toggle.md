---
status: ready-to-start
epic: 18-tui-bubble-tea-interactive-planning-browser
description: Render the detail-pane body with glamour (cached, never in View), with an R toggle between raw and pretty; persist the preference
effort: Unknown
tier: 3
priority: medium
autonomy_level: 3
tags: [tui, bubble-tea]
created: "2026-06-11"
---

# TUI: glamour markdown rendering with raw/pretty toggle

## Objective

The detail pane shows raw markdown. Render it with **glamour** for comfortable
reading, but keep a **raw toggle** — the CLI `task show` deliberately stayed raw
to preserve fidelity + avoid a heavy dep, so the TUI should offer both rather than
replace raw. This was always the plan (epic decision: "plain first, then glamour").

## Approach

- Add `glamour` as a dep. Render in the detail **load Cmd / `SetContent`**, cache
  the compiled string on the pane — **never call glamour in `View()`** (it's
  CPU-heavy; the documented footgun). Re-render on `SetSize` (width reflow).
- Cache both representations per item: the raw (current `renderTaskDetail`-style)
  string and the glamoured body. A toggle just swaps which the viewport shows; no
  re-render needed to flip.
- **`R` key** toggles raw ⇄ pretty; a **global preference** (on the Model) so it
  persists across selections and tabs. Show the mode in the detail title or footer.
- Frontmatter fields stay as the existing styled key/value block; glamour applies
  to the **body** only (the fields aren't markdown).

## Interaction with detail find (important)

S2b's `/` find highlights by line index over the rendered content. Glamour reflows
and indents, so: (a) match line indices differ between raw and pretty — recompute
matches on toggle; (b) the "rebuild matched line from stripped text" highlight drops
glamour styling on matched lines (the #4 issue in
[[tui-s2b-polish-find-occurrences-highlight-fidelity-per-entity-sort]], amplified).
Coordinate with that task — ideally land the ANSI-aware highlighter first.

## Acceptance criteria

- [ ] The detail body renders as styled markdown (headings, lists, code, emphasis)
      by default; `R` flips to raw and back; the mode is visible.
- [ ] Toggling and resizing never lag — glamour runs in `Update`/load, not `View`;
      flipping mode does no re-compile.
- [ ] The preference persists across selection and tab changes.
- [ ] Find still works in both modes (matches recomputed on toggle). Suite + lint green.

## Out of scope

- Glamour for epic/audit bodies beyond the same body-rendering path (apply uniformly
  via `detailContent`, but no per-entity markdown features).
- A custom glamour theme (use a sensible built-in; theming is a later polish).

## Related

- Epic [[18-tui-bubble-tea-interactive-planning-browser]]
- Interacts with [[tui-s2b-polish-find-occurrences-highlight-fidelity-per-entity-sort]]
