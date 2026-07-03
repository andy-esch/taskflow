---
schema: 1
status: completed
epic: 25-design-system-coherent-palette-and-selectable-themes
description: Convert the global lipgloss chrome vars (accent, pane/help/action/edit borders, dashHeading, find highlights) into palette-built styles; lipColor -> palette truecolor.
effort: M
tier: 3
priority: medium
autonomy_level: 3
tags: [tui, design]
created: "2026-06-28"
blocked_by: [design-package-foundation-palette-theme-registry-and-the-neon-default]
updated_at: "2026-06-28"
started_at: "2026-06-28"
completed_at: "2026-06-28"
id: 6fgq1n0016kj
---
## Objective
Make the TUI's structural chrome derive from the palette instead of scattered lipgloss literals.

## Scope
- In: convert the global style vars into a `styles` struct built from a `Palette` in `New` — `tui/style.go` (accent, paneActive/Inactive, selected/dim/activeTab), `dashboard.go` dashHeading, `help.go` border/heading, `action.go` action/danger border+heading, `find.go` match/current, `edit.go` edit-box border. Reimplement `lipColor` as `pal.Of(c).True`. Thread `design.Default().For(dark)` for now (config selection is a later task).
- Out: bars/CLI ANSI, picker, config wiring. Glyphs + border *style* unchanged.

## Done when
No stray chrome `lipgloss.Color(...)` literals in the TUI; styles built from the palette; `build/test/lint` green; TUI substring tests pass unchanged.

## Reference
Design doc §1,§5. Depends on the design foundation. Parallel with the bars task.

**Implementation 2026-06-28 (branch feat/tui-chrome-palette).** All TUI chrome now routes through the shared palette — zero `lipgloss.Color("…")` literals remain in internal/tui. Approach: a package-level `pal design.Palette` (default neon-night dark at init) plus an `applyTheme(p)` recolor hook in style.go that repoints every chrome style in place; `Run` detects the background once and calls `applyTheme(design.Default().For(dark))` before the first render (the same `dark` also picks the markdown style). `lipColor` now returns `pal.Of(c).Color()`. Chose the package-global pattern (consistent with the existing chrome globals + the free fns fg/lipColor) over a Model-field styles struct — far less churn, and the applyTheme hook is exactly what the future live-retheme task calls on BackgroundColorMsg. Routed: accent, pane borders, dashHeading, help/action/danger borders+headings, find match/current, edit box, lipColor. DELIBERATELY OUT OF SCOPE (T3): progressbar gradient/segments + miniBar/segBar + the CLI ansiCode map — untouched. Verified: gofmt clean, `go build ./...`, `go vet`, full `go test ./...` all green. `just lint` (golangci-lint) still owed on host (not in container).
