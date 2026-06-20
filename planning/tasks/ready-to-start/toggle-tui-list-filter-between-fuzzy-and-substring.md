---
status: ready-to-start
epic: 18-tui-bubble-tea-interactive-planning-browser
description: Toggle the TUI list filter between fuzzy and substring via a discoverable keybinding shown in the help footer, plus a visible mode indicator
effort: Unknown
tier: 3
priority: medium
autonomy_level: 3
tags: [tui, ux, filter]
created: "2026-06-19"
---
## Objective

Let a TUI user **toggle the list filter** between **fuzzy** (forgiving, good for
exploration — the current default) and **substring/exact** (predictable "contains",
better for structured slugs), with a **clearly discoverable keybinding shown on the
page** so users know it exists.

Context: the CLI picker deliberately uses substring matching (predictable for slug
identifiers — see `substringFilter` in `internal/cli/prompt/picker.go`); the TUI
keeps bubbles/list's default fuzzy (forgiving, exploratory). Both are reasonable
defaults for their context — this task adds the *choice* in the TUI rather than
forcing one.

## Design notes

- `bubbles/list`'s `Filter` field is a settable `FilterFunc`; toggling swaps it
  between `list.DefaultFilter` (fuzzy) and a substring `FilterFunc`. The CLI
  already has a substring implementation — **extract `substringFilter` into a small
  shared helper** (e.g. `internal/listfilter`) and reuse it in both the picker and
  the TUI, so the two can't drift.
- Register the toggle as a `key.Binding` with a `Help()` string so it **shows up in
  the list's help footer** (short + full help) — discoverability is the point.
- Show the **active mode** in the UI (e.g. the filter prompt reads `Filter
  (fuzzy)` / `Filter (exact)`, or a small indicator), so the user can see which is
  on, not just that a toggle exists.
- Re-run the filter on toggle so the visible results update immediately.
- Decide whether the choice persists for the session (likely yes) and whether it's
  remembered across tabs.

## Acceptance criteria

- [ ] A keybinding toggles the active list filter (fuzzy ↔ substring) live.
- [ ] The shortcut is visible in the list help footer (not hidden/undocumented).
- [ ] The active mode is indicated on-screen.
- [ ] Default stays **fuzzy** (the TUI's exploratory default is unchanged).
- [ ] The substring matcher is the SAME one the CLI picker uses (shared helper).

## Out of scope

- Changing the CLI picker (stays substring) or the TUI default (stays fuzzy).
- Multi-term / AND filtering, regex, or scoring tweaks — just the two modes.

## Related

- Epic [[18-tui-bubble-tea-interactive-planning-browser]]
- Builds on the substring filter added for the CLI picker in
  [[interactive-prompt-layer-gh-style-pickers]].
