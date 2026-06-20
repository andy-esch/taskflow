---
status: completed
epic: 18-tui-bubble-tea-interactive-planning-browser
description: Toggle the TUI list filter between fuzzy and substring via a discoverable keybinding shown in the help footer, plus a visible mode indicator
effort: Unknown
tier: 3
priority: medium
autonomy_level: 3
tags: [tui, ux, filter]
created: "2026-06-19"
updated_at: "2026-06-20"
started_at: "2026-06-20"
completed_at: "2026-06-20"
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

## Shipped (2026-06-20)

`F` toggles the TUI list filter fuzzy ⇄ substring.

- **Shared matcher:** extracted the CLI picker's `substringFilter` into
  `internal/listfilter.Substring` (a `list.FilterFunc`); both the picker
  (`prompt/picker.go`) and the TUI now use it, so they can't drift. Its tests
  moved with it.
- **Toggle:** `F` (`keys.FilterMode`) → `Model.toggleFilterMode` swaps each tab's
  `list.Filter` between `list.DefaultFilter` (fuzzy) and `listfilter.Substring`,
  **session-wide across every tab** (consistent no matter where you filter), and
  re-runs the visible filter live via `SetFilterText` so results update on toggle.
- **Indicator:** the filter-input prompt names the mode while typing
  (`filter (fuzzy): ` / `filter (exact): `); the title chip annotates an applied
  filter as `filter(exact):…` when non-default (fuzzy stays the silent default,
  matching the chip's `view:`/`sort:` philosophy).
- **Discoverability:** `F` is in the `?` help overlay ("filter mode: fuzzy ⇄
  substring (default fuzzy)").
- **Decisions:** persists for the session and is remembered across tabs (a single
  session-wide mode, not per-tab); default stays **fuzzy**.
- Tests: `listfilter` unit (moved) + a TUI test (`F` flips the mode + prompt on
  every tab, defaults fuzzy, toggles back).

## Acceptance criteria

- [x] A keybinding toggles the active list filter (fuzzy ↔ substring) live.
- [x] The shortcut is visible in the list help footer (the `?` overlay).
- [x] The active mode is indicated on-screen (filter prompt + chip when exact).
- [x] Default stays **fuzzy** (the TUI's exploratory default is unchanged).
- [x] The substring matcher is the SAME one the CLI picker uses
      (`internal/listfilter`).

## Out of scope

- Changing the CLI picker (stays substring) or the TUI default (stays fuzzy).
- Multi-term / AND filtering, regex, or scoring tweaks — just the two modes.

## Related

- Epic [[18-tui-bubble-tea-interactive-planning-browser]]
- Builds on the substring filter added for the CLI picker in
  [[interactive-prompt-layer-gh-style-pickers]].
