---
schema: 1
status: completed
epic: 25-design-system-coherent-palette-and-selectable-themes
description: styles is a ~10KB struct (13 lipgloss.Style fields); render helpers are value receivers copying it per call in hot row paths — switch to pointer receivers to match the *styles storage elsewhere
effort: Unknown
tier: 3
priority: medium
autonomy_level: 3
tags: [design, tui, refactor]
created: "2026-06-29"
updated_at: "2026-06-30"
started_at: "2026-06-30"
completed_at: "2026-06-30"
id: 6fh1b7w01zfa
---

# styles render helpers: pointer receivers instead of value

## Objective

<why / what — one short paragraph>

## Acceptance criteria

- [ ] <observable outcome>

## Out of scope

- <explicitly excluded>

## Related

- Epic [25-design-system-coherent-palette-and-selectable-themes](../epics/25-design-system-coherent-palette-and-selectable-themes.md)

**Done 2026-06-30.** Flipped all `styles` render helpers (lipColor/fg/glyph/dim/miniBar/segBar/statusText/priorityText) from value to pointer receivers, and every free fn/method that took `s styles`/`st styles` by value (view/cell/hint/enumInline/meta/renderTaskMeta·EpicMeta·AuditMeta/detailField/helpLines/helpBox/symbolsFor/helpSectionsFor/relDateCells/highlightLine/row/epicGlyph/epicStatusNote/setSummary/urgencyLine/componentLine/dash.view/palette·action·follow·edit·command.view) now takes `*styles`. Call sites drop the `*d.st`/`*m.st` value-derefs (e.g. `st := d.st`); tests pass `&testStyles`. No more ~10KB struct copy per render call; coherent with the `*styles` storage on Model/delegates/detailPane. Behavior unchanged — gofmt/vet/full `go test ./...` green.
