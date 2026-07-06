---
status: completed
epic: 17-pm-go-cli
description: TTY-aware ANSI color, status glyphs, dim table headers and count footers in the render layer, with --color and NO_COLOR support
effort: Unknown
tier: 3
priority: medium
autonomy_level: 3
tags: [pm-tooling, go, cli, ergonomics]
created: "2026-06-09"
completed_at: "2026-06-09"
updated_at: "2026-06-09"
id: 6fakbec017kj
---

# CLI color, glyphs, table headers (render styling)

## Objective

The flagship CLI ergonomics upgrade (companion to autocomplete): colorize human
output with status glyphs, dim table headers, and count footers — TTY-aware so
pipes and agents stay plain.

## Done

- `render/style.go`: a `Style` value (zero = plain/disabled) with `Status`
  (colored glyph + label), `Priority`, `Bucket`, `Percent`, `Bold`/`Dim`/
  `Green`/`Red`, plus an **ANSI-aware `writeTable`** (text/tabwriter miscounts
  escape bytes → broke alignment, so columns pad on `visibleWidth`).
- All human renderers take a `Style`: `task list`/`epic list`/`audit list` get
  dim headers + colored glyphs + a count footer; show/moves/lint/fix/created
  colorized. JSON renderers untouched.
- `cli/color.go` + `--color=auto|always|never` (persistent) honoring `NO_COLOR`;
  auto = colored only on a TTY (`isTerminal` via `os.ModeCharDevice`, no x/term
  dep). `app.Style` set in `PersistentPreRunE`.
- Plain output stays byte-stable for non-TTY (glyphs gated on color), so every
  existing test passes unchanged; empty list still prints nothing.

## Acceptance

- [x] `--color=always` emits ANSI; default (piped/non-TTY) and `--color=never`
      are plain; `NO_COLOR` respected. Tests cover all three + `visibleWidth` +
      colored-table alignment. Full suite + lint green.
- [x] Demoed: glyph'd, aligned, headed `task list`/`epic list`.

## Out of scope

- Relative dates, terminal-width truncation, `glamour` markdown bodies — the
  "readability polish" follow-ups, separate from the color core.

## Related

- Epic [17-pm-go-cli](../epics/17-pm-go-cli.md); the keyboard-ergonomics sibling is
  [[fuzzy-partial-slug-resolution]]; the at-a-glance view is
  [at-a-glance-dashboard-status-board](6fakbec01wt8-at-a-glance-dashboard-status-board.md).
