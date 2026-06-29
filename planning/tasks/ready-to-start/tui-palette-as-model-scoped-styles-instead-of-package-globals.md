---
schema: 1
status: ready-to-start
epic: 25-design-system-coherent-palette-and-selectable-themes
description: Package-global pal+applyTheme (from T2) -> Model-scoped palette/styles, for multi-session safety (wish/epic 19) and test isolation; fg/lipColor become methods. No visible change.
effort: M
tier: 4
priority: low
autonomy_level: 3
tags: [tui, design]
created: "2026-06-28"
---
## Objective
Replace the package-global `pal` + `applyTheme` (shipped in [[route-tui-chrome-through-the-palette]]) with a Model-owned palette + styles, so theming is per-instance rather than process-global.

## Why (the robustness case)
- **Multi-session safety:** package globals are shared process-wide; a `wish`/SSH server (epic [[19-web-companion-apps-over-a-shared-core]]) serving concurrent sessions would have one session's `applyTheme` clobber the others'. A Model-scoped palette is per-session by construction — and aligns with the repo's DI/no-globals architecture rule.
- **Test isolation:** a global can't render light- and dark-themed models in one process without racing; Model-scoped styles unlock themed-output assertions + `t.Parallel()`.

## Scope
- Move `pal` + the chrome styles (accent, borders, headings, find/match, edit box) onto the Model (or a `styles` struct built from a `design.Palette` in `New`).
- Make the free fns that read the global — `fg`, `dim`, `glyph`, `lipColor`, `statusText`, `priorityText` — methods on Model/styles; thread `m`/`m.styles` through the call sites.
- Keep behavior identical; structural refactor, not a visual change.

## Not urgent
No user-visible benefit today (single local TUI process). Value is unlocked by epic 19 (wish/multi-session) and helps the live-retheme task. Tier 4 / low.

## Reference
Design doc: `planning/research/2026-06-28-color-palette-and-theming-overhaul.md` (§5 recommended the Model-field shape). Supersedes the package-global stopgap.
