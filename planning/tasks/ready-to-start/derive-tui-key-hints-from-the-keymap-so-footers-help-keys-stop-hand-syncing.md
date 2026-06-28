---
schema: 1
status: ready-to-start
epic: 21-code-quality-architecture-hardening
description: The ? help table, three footer hint strings, and keys.go are hand-synced copies of the keybinding vocabulary; derive hints from one source (key.WithHelp on keyMap).
effort: M
tier: 3
priority: low
autonomy_level: 3
tags: [tui, render, maintainability]
created: "2026-06-28"
---
## Why

From the 2026-06-28 adversarial review of the legend/zoom/help-wrap work. The keybinding vocabulary is written out by hand in several places that must stay in sync:

- `internal/tui/help.go` `helpSections` (the `?` overlay table) — its own comment says "Keep it in sync with keys.go and the focus-routed handlers."
- `internal/tui/keys.go` `keyMap` (the bindings; only `FilterMode` uses `key.WithHelp`).
- `internal/tui/model.go` `footer()` — THREE hand-strung hint strings (list-focus, detail-focus, full-screen/zoom), each repeating `: cmd / m move / z full / ...`.

Adding the `z` zoom feature widened this to ~4 places and added a fourth distinct footer string. A rebind of any key today touches all of them; nothing derives.

## Goal

Make the keymap the single source of truth: attach `key.WithHelp(key, desc)` to every `keyMap` binding, then derive both the `?` table and the footer hints from it (filtered by focus/context). The footers can be generated from an ordered subset of bindings per context instead of literals.

## Notes / acceptance

- Non-breaking: `?` overlay + footers render the same vocabulary (goldens are TUI-render tests, not CLI goldens).
- Keep the context-specific ordering (active-pane keys first, Global last) and the page-specific Notes/Symbols sections.
- Pre-existing house style hand-maintained footers, so this is a deliberate DRY pass, not a regression fix.