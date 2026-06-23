---
schema: 1
status: ready-to-start
epic: 20-cli-ux-and-ergonomics
description: 'Page long show output via $PAGER like git: TTY-gated off the machine contract; configurable per-repo via .tskflwctl.toml [pager] + --no-pager.'
effort: Unknown
tier: 3
priority: medium
autonomy_level: 3
tags: [cli, ux, config]
created: "2026-06-23"
---
# Git-style auto-pager for long human output

Page long human output (the `show`-type commands) through `$PAGER` like git does,
**hard-gated off the machine contract** and configurable per-repo via
`.tskflwctl.toml`. Source: the auto-pager idea raised reviewing `epic show`'s new
tree (output runs off the top of the screen). Governing rule (epic 20): never
compromise the agent/pipeline contract — paging is a TTY-only human nicety.

## The gate (non-negotiable, reuse the existing pattern)
Page ONLY when the human face is active: **stdout is a TTY AND not `--json` AND
not `--no-input`/`TSKFLW_NO_INPUT`** — the same conditions `gateOpen`
(`internal/cli/color.go`) already computes for prompting/color. Off a TTY (pipe,
redirect, agent, CI) or under `--json`, output is printed RAW exactly as today —
the byte-stable machine contract is untouched by construction.

## Config — mirror git into `.tskflwctl.toml`
git keys paging off `core.pager` + `pager.<cmd>`; we put a `[pager]` table in the
repo config (parsed by `internal/config`, alongside `taskflow_root` etc.):

```toml
[pager]
enabled = true          # default: page long human output on a TTY
command = "less -FRX"   # override the pager program
```

Precedence (mirror git's GIT_PAGER > core.pager > PAGER > less):
- **program:** `TSKFLW_PAGER` env → `[pager].command` → `$PAGER` → default `less -FRX`.
- **on/off:** `--no-pager` flag (force off) → `--paginate` flag (force on) →
  `[pager].enabled` → default on. The TTY/`--json` gate above always wins.

`less -FRX`: `F` = don't page if it fits one screen (short `epic show` won't trap
you), `R` = keep our ANSI colors, `X` = don't clobber scrollback.

## Scope
1. A pager seam: on the gated human path, wrap `app.Out` with a spawned `$PAGER`
   process (stdin = our output); flush/close on command end. Everything else
   (machine path) passes the original writer through unchanged.
2. Apply to the long `show` commands first: `epic show`, `task show`, `audit
   show`, `schema` (+ `schema <kind>`/`--json-schema`). `list`/`status` can opt in
   later if wanted.
3. `--no-pager` (+ optional `--paginate`) persistent flags; `[pager]` config keys.

## Open questions
- [ ] **Which config in the pointer case?** When run from an impl repo whose
      `planning_repo` points elsewhere, does the pager setting come from the impl's
      config or the resolved planning repo's? (Pager is a local-terminal concern →
      likely the config Discover lands on, i.e. the parsed `cfg`; confirm.)
- [ ] Per-command toggles (git's `pager.<cmd>`) — ship now or defer? Default: a
      single global `[pager].enabled` first; per-command later if needed.
- [ ] Default on vs off out of the box (git defaults on). Recommend **on**.

## Acceptance criteria
- [ ] Long `show` output pages on a TTY; `--no-pager`, `[pager].enabled=false`,
      `$PAGER`/`TSKFLW_PAGER`/`[pager].command` all respected with the right precedence.
- [ ] Off a TTY / under `--json`: output is byte-identical to today (no pager, no
      hang) — a test pins this (the agent contract).
- [ ] `internal/config` parses `[pager]`; schema/docs updated for the new keys.
- [ ] Suite + lint green; the buffer-capturing CLI tests are unaffected (non-TTY).

## Related
- Surfaced 2026-06-23 from the `epic show` tree work
  ([[epic-show-lipgloss-v2-tree-rendering]]).
- Pattern precedent: `gateOpen` / the color + prompt TTY-gating in
  `internal/cli/color.go`. Epic [[20-cli-ux-and-ergonomics]].
