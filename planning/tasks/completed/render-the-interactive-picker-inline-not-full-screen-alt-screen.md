---
schema: 1
status: completed
epic: 20-cli-ux-and-ergonomics
description: The fillSelect/resolveOne picker (bare epic show, task edit, etc.) takes over the whole terminal via alt-screen; render it as a compact inline selector menu below the prompt (gh/huh-style) instead.
effort: M
tier: 3
priority: medium
autonomy_level: 3
tags: [cli, tui]
created: "2026-06-25"
updated_at: "2026-06-25"
started_at: "2026-06-25"
completed_at: "2026-06-25"
---
# Render the interactive picker inline, not full-screen (alt-screen)

## Objective

The `fillSelect`/`resolveOne` picker (bare `epic show`, `task edit`, `audit close`,
…) takes over the WHOLE terminal via alt-screen — it feels jarring for a quick
"which one?" choice. Render it instead as a compact selector menu **below the
prompt** (gh / huh-Select style): a few rows, inline, leaving the surrounding
output in place.

## Context

The picker lives in `internal/cli/prompt/picker.go` (a `pickerModel` over
`bubbles/list`, run via `tea.NewProgram` at ~line 97). Two things make it
full-screen:
- `View()` sets `v.AltScreen = true` (declarative alt-screen in bubbletea v2).
- `list.New(items, d, 0, 0)` + `SetSize(msg.Width, msg.Height)` size the list to
  the whole terminal.

Alt-screen was a deliberate choice ("keeps the picker from scrolling the
surrounding output, restored cleanly on exit") — so the refinement must preserve a
**clean teardown** without it. It was built by the completed
[[interactive-prompt-layer-gh-style-pickers]]; this is its ergonomic follow-up.
The bubbles/list component was chosen over `huh.Select` because huh's own filter
misbehaved — so keep bubbles/list, just render it inline.

## Acceptance criteria

- [ ] The picker renders INLINE below the prompt (no alt-screen takeover): a
      compact window (~7–10 rows) showing the title + the visible options + the
      `/` filter, with the surrounding shell output untouched above it.
- [ ] On selection OR abort, the terminal is left clean — no leftover menu lines,
      no scroll pollution (handle bubbletea v2's inline final-frame / clear).
- [ ] The `/` fuzzy filter and scrolling still work (the reason bubbles/list was
      picked over huh.Select).
- [ ] Non-TTY / `--no-input` / `--json` paths unchanged (picker only on a TTY;
      exit-11 otherwise).
- [ ] go build/test/lint green; picker tests updated.

## Implementation sketch

- Set `v.AltScreen = false`; cap the list height to `min(len(items)+chrome, ~10)`
  instead of the full terminal height (still handle resize).
- Consider a single-line list delegate (the DefaultDelegate is 2 lines/item) for a
  tighter menu.
- Ensure the final `View` (post-selection) clears the menu region so nothing is
  left behind — check bubbletea v2 inline-mode semantics for the exit frame.

## Risks / gotchas

- Inline rendering can leave artifacts if the menu exceeds the remaining terminal
  height, or on resize mid-pick — alt-screen sidestepped both, so test those.
- Hard to verify headlessly; needs a real-terminal check. Pin what you can in tests
  (height capped, AltScreen off, selection still resolves).

## Done when

Bare `tskflwctl epic show` (and the other pickers) drop a small selector menu under
the prompt, you arrow/filter to a choice, and on enter the menu vanishes leaving a
clean prompt — no full-screen flash.

## Research 2026-06-25 — huh.Select reshapes this task

Deep web research (verified vs the pinned huh v2.0.3 source) found the original
reason we hand-rolled the bubbles/list picker — huh's filter being broken — is
**largely obsolete**: `huh.Select` now has first-class `.Filtering(true)` (substring
match, same policy as our `listfilter.Substring`) + filter keymaps. And **huh forms
render inline-vertical below the prompt BY DEFAULT** (alt-screen is opt-in via
`WithViewHook`) — i.e. huh already gives the "menu below the prompt" UX this task
wants. So two paths:

- **Path A (incremental):** keep bubbles/list; set `v.AltScreen=false` + cap the
  list height. Smallest diff, exact current behavior, fewer free features.
- **Path B (huh.Select):** replace picker.go's bubbles/list with
  `huh.NewSelect[T]().Filtering(true)` — inline by default, FEWER lines, and free
  **descriptions** (epic id + desc — we already gather it in `labeledOption`),
  themes (ThemeDracula, already used by the Text prompt), validation.

**Verify before Path B:** GitHub huh #510 — Select hides its title while filtering
(cosmetic); a claim that v2.0.3 hardened the filter UI was REFUTED. So spike
huh.Select's filter in a real terminal; fall back to Path A if rough.

⚠️ `huh.Select.Inline(true)` is a DIFFERENT thing (single-line HORIZONTAL, Height 1,
left/right nav) — NOT what we want. Plain vertical huh.Select is the inline menu.

**Free toggles to bundle either way** (bubbles/list): `SetShowStatusBar` +
`SetStatusBarItemName("epic","epics")` (item count), `SetShowHelp`/
`AdditionalShortHelpKeys` (the picker has NO on-screen key hints today),
descriptions as a 2nd line.

**Spun off (separate tasks worth filing):** multi-select for the variadic verbs
(`task complete a b c`, `audit close x y` pick only ONE today) via
`huh.MultiSelect[T]` (`.Limit(n)`, `Option.Selected(true)`) — the Prompter interface
already anticipates `SelectMany`; textinput autocomplete (`ShowSuggestions`); TUI
`view.WindowTitle`/`ReportFocus`/OSC `ProgressBar`; `form.WithAccessible(env)`.
