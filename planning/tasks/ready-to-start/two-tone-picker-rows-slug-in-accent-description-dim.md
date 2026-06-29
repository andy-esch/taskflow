---
schema: 1
status: ready-to-start
epic: 25-design-system-coherent-palette-and-selectable-themes
description: 'Deferred from T6: render picker options with the slug in the theme accent and the description dimmed. Fragile because huh re-styles each row over the label; wants visual iteration.'
effort: S
tier: 3
priority: low
autonomy_level: 3
tags: [cli, design]
created: "2026-06-29"
---
## Objective
Render interactive-picker rows two-tone: the **slug in the theme accent**, the **description dimmed** (the "slug one color / description another" idea; two-line optional). Deferred from [[theme-discovery-commands-glamour-polish-and-a-second-theme]] (T6).

## Why it was deferred (the fragility)
`prompt.Option` is `{Label, Value string}` and `SelectOne` hands the label to `huh.Select`. huh re-styles **every row** through its own `Option` / `SelectedOption` styles (`Foreground(accent).Bold` on the current row) — applied OVER the whole label. So embedding per-part ANSI inside the label gets overridden (especially on the selected row, which huh paints entirely in the accent), and huh's width/truncation is ANSI-sensitive. Reliable two-tone needs care + visual iteration, not a blind change.

## Approach
- Add a `Desc string` to `prompt.Option` (callers populate slug=Label, description=Desc), so `SelectOne` can compose a colored label from the parts rather than callers pre-formatting one string.
- Compose `accent(slug) + "  " + dim(desc)`, truncated ANSI-aware to one line — OR a two-line option (slug line + dim desc line) if huh height/measurement cooperates.
- Verify against huh v2: does `Option`/`SelectedOption` Foreground override embedded ANSI? does the selected row still read? does filtering/width stay correct? Iterate visually (`task start` with no arg triggers the picker).
- Pull the accent/dim from the active theme (the prompter already holds `design.Theme`).

## Reference
Deferred picker polish from T6; the picker accent caret/selection already lands via [[route-the-interactive-picker-theme-through-the-palette]] (T4). Update the callers that build picker labels (grep `prompt.Option` / `SelectOne`).
