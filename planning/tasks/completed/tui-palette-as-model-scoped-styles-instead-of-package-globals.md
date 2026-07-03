---
schema: 1
status: completed
epic: 25-design-system-coherent-palette-and-selectable-themes
description: Package-global pal+applyTheme (from T2) -> Model-scoped palette/styles, for multi-session safety (wish/epic 19) and test isolation; fg/lipColor become methods. No visible change.
effort: M
tier: 4
priority: low
autonomy_level: 3
tags: [tui, design]
created: "2026-06-28"
updated_at: "2026-06-30"
started_at: "2026-06-29"
completed_at: "2026-06-30"
id: 6fgq1n00054d
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

**Review follow-ups (2026-06-29).** The pre-T5 adversarial review converged on a concrete shape for this refactor and added one decision to fold in:

- **`newChrome(p) chrome` constructor.** The shipped `applyTheme` hand-mirrors ~13 package-global chrome styles declared across 6 files; forget to add a new style to `applyTheme` and it silently renders default-dark — wrong only on light terminals (the hardest case to catch). Replace the globals with a `chrome` struct built by `newChrome(p design.Palette) chrome`; the compiler then forces every field to be palette-sourced, killing the silent-miss class. This is also exactly the shape this task wants (drop `chrome` onto the `Model`).
- **Decide `Hue.ANSI` on the TUI low-color path.** The palette carries curated 16-color ANSI slots for chrome (accent=13, gradient 5/14/13) that NOTHING reads — the TUI feeds raw truecolor to lipgloss and lets it auto-downsample, so neon purple/pink can collapse to the same magenta on a 16-color terminal. Either honor `Hue.ANSI` on the TUI's low-color profile (controlled degradation), or formally drop chrome `Hue.ANSI` as dead data and document the reliance on lipgloss downsampling. Practical impact is low (semantic colors downsample fine), so it can ride with this refactor.

**T5 note (2026-06-29).** T5 made the global `pal` LOAD-BEARING — it now holds a config/flag-selected theme (was always `design.Default().Dark`, an effective constant). `tui.Run(svc, layout, th design.Theme)` is the input seam this refactor should consume: the Model already receives the theme at construction via `Run`, so the refactor becomes "store `th` on the Model + thread `pal` through render calls." T6's `theme preview` is the first realistic second-caller that would stomp the global, raising the urgency.

**Done 2026-06-29 (branch refactor/tui-model-scoped-styles).** The package-global mutable `pal` + ~13 chrome style vars + `applyTheme` are gone. Replaced with a `styles` struct (palette + chrome styles + render helpers as methods), built by `newStyles(p)`. The Model holds `st *styles`; the list delegates share that SAME pointer (threaded via newEntityTabs); `Run` repopulates it once after background detection (`*m.st = newStyles(th.For(dark))`) instead of mutating globals — so each Run/session has isolated, per-Model theming (the multi-session/wish + test-isolation goal). ~26 files (mostly mechanical call-site threading to m.st/d.st). Tests updated via a shared `testStyles` helper. gofmt/vet/full `go test ./...` green; behavior unchanged (substring tests pass). NOTE: a sub-agent did the bulk; its connection dropped mid-edit.go, finished by hand (edit.go, view.go, all test call-sites).

**Adversarial review (3 agents, distinct lenses) — 2026-06-29.** Verdict unanimous: correct, behavior-preserving, goals met (per-session isolation, test isolation, no-globals architecture rule). No reachable nil-deref, no behavior drift vs the old applyTheme slots, no stale-copy hazard (delegates deref the shared pointer per render; Run sets it before first render). Cleanups applied from the review: (1) deleted dead `styles.accent` field (written once, never read — vestige of the old global); (2) reordered `enumInline` to styles-LAST to match every other free fn; (3) softened the live-retheme comment in tui.go (startup colors every surface, but a *runtime* swap would also need to refresh the dashboard/detail surfaces that snapshot styles by value). gofmt/vet/full `go test ./...` green after fixes. Two non-blocking follow-ups filed: value→pointer receivers on `styles` render helpers, and fold glamour markdown style into `styles`.
