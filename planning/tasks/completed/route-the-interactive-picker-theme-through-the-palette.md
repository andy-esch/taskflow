---
schema: 1
status: completed
epic: 25-design-system-coherent-palette-and-selectable-themes
description: Build the huh picker theme from the palette (Accent caret/selection) instead of the hardcoded neon-purple stopgap.
effort: S
tier: 3
priority: medium
autonomy_level: 3
tags: [cli, design]
created: "2026-06-28"
blocked_by: [design-package-foundation-palette-theme-registry-and-the-neon-default]
updated_at: "2026-06-29"
started_at: "2026-06-28"
completed_at: "2026-06-29"
---
## Objective
Replace the picker's hardcoded `#b026ff` stopgap with the palette.

## Scope
- In: `cli/prompt/tty.go` `pickerTheme` builds the huh `*Styles` from the palette inside the `ThemeFunc(isDark)` closure — caret + selected row from `Palette.Accent`. Can start from `ThemeBase16`/`ThemeDracula` and migrate fields.
- Out: the slug/description two-tone picker polish (lands in the discovery/polish task against the real palette).

## Done when
No hex literals in `pickerTheme`; the caret/selection come from the palette; `build/test/lint` green.

## Reference
Design doc §5. Depends on the design foundation.

**Implementation 2026-06-28 (worktree, branch feat/picker-palette off main).** pickerTheme now draws the huh selection caret + current row from design.Default().For(isDark).Accent (the shared neon accent), resolved inside huh's ThemeFunc(isDark) closure — replacing the hardcoded #b026ff stopgap. The lipgloss import dropped (it was only the literal); added design. Base stays ThemeDracula (a fuller huh-base migration + the deferred slug/description two-line picker polish are T6). Verified: gofmt, go build ./..., go vet, full go test ./... green. Single-file change (internal/cli/prompt/tty.go); zero overlap with T2/T3.
