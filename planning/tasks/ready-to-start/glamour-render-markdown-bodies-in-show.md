---
status: ready-to-start
epic: 20-cli-ux-and-ergonomics
description: Render task/epic/audit show bodies as styled markdown via glamour on a TTY (raw under --json/pipe); glamour already in the module graph
effort: Unknown
tier: 3
priority: medium
autonomy_level: 3
tags: [cli, ux, output]
created: "2026-06-19"
---
## Objective

`task/epic/audit show` print the raw markdown body verbatim. On a TTY, render it
with [glamour](https://github.com/charmbracelet/glamour) (headings, lists, code
fences, bold/emphasis, links) for a far nicer read — the same engine gh/glab/gitea
CLIs use. glamour is **already in our module graph** (transitive via the TUI), so
this adds no new third-party org.

Strictly a human-face nicety: it applies only when we'd already colorize (TTY +
`--color`), and NEVER to `--json` (the `body` field stays raw) or piped/`--color=never`
output. The agent/porcelain contract is untouched.

**Decision (2026-06-19):** render on a TTY **by default** (gated on the existing
color decision), and add a **`--raw`** escape hatch for humans who want the source.

## Acceptance criteria

- [ ] `show` on a TTY renders the body via glamour by default; `--json` body is
      unchanged (raw markdown), and piped / `--color=never` output is the raw body.
- [ ] `--raw` forces the unrendered source even on a TTY.
- [ ] Style follows the existing color decision (`wantColor`) — one gate, no extra
      "render?" flag beyond `--raw`.
- [ ] Width respects the terminal (glamour `WithWordWrap`) and the existing
      width detection; downsample colors via lipgloss if needed.
- [ ] Tests assert raw-under-json and rendered-under-TTY (golden or ANSI-present check).

## Out of scope

- Rendering list output (this is `show` only).
- A full pager/TUI browser (that's epic 18).

## Related

- Epic [[20-cli-ux-and-ergonomics]]
- Sibling human-face work: [[evaluate-fang-for-styled-help-errors-and-manpages]]
