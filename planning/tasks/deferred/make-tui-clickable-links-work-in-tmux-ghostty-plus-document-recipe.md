---
schema: 1
status: deferred
epic: 18-tui-bubble-tea-interactive-planning-browser
description: TUI OSC 8 clickable links work in bare terminals but not in tmux+ghostty. Apply terminal-features hyperlinks + tmux restart + shift-cmd-click; then document the recipe in the README.
effort: Unknown
tier: 3
priority: medium
autonomy_level: 3
tags: [tui, docs]
created: "2026-06-24"
updated_at: "2026-06-25"
started_at: "2026-06-25"
deferred_at: "2026-06-25"
---
## Objective

The TUI's clickable detail-title links (OSC 8) work in a bare terminal (iTerm) but
not in tmux+ghostty. This is terminal/tmux config, NOT a tskflwctl bug — the app
emits correct OSC 8 (proven by bare iTerm). Apply the config locally, confirm, and
document the recipe in the README's TUI notes (next to the clipboard/pager tmux
caveats). Deferred only because the render fix needs `tmux kill-server`, which would
drop the current session.

## Diagnose first (raw test in a tmux pane)

`printf '\e]8;;https://example.com\e\\Click Me\e]8;;\e\\\n'`
- "Click Me" underlined -> rendering works; it's clickability -> Layer 2.
- Plain text -> tmux isn't passing OSC 8 -> Layer 1.

## Acceptance criteria

- [ ] **Layer 1 (render):** tmux >= 3.4 + `set -as terminal-features ',xterm-ghostty:hyperlinks'` (match real `$TERM`); then a full `tmux kill-server` -- a `source-file` reload often does NOT apply terminal-features. Verify: `tmux info | grep -i hyperlink`.
- [ ] **Layer 2 (click):** with tmux `mouse on`, open links via **shift+cmd+click** (Shift bypasses tmux's mouse capture; ensure Ghostty `mouse-shift-capture` is not `true`). Alternative: `set -g mouse off` -> plain cmd+click (loses tmux mouse scroll/select).
- [ ] Document the working recipe in the README TUI section, beside the clipboard/pager tmux notes.

## Notes

- Our link is on the short detail title (slug), so it avoids tmux's multi-line-URL parsing bug.
- Likely the live symptom is Layer 2 (cmd-click without Shift while tmux mouse is on).

## Related

- Research: GnixAij "Fix Clickable Hyperlinks in Tmux With Ghostty"; tmux 3.4 CHANGES; ghostty #11907 / #9108.
- Built in [[tui-v2-polish-window-title-and-clickable-detail-path]].
- Epic [[18-tui-bubble-tea-interactive-planning-browser]].
