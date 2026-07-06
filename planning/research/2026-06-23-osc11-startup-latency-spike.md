---
status: reference
created: "2026-06-23"
tags: [tui, perf, bubbletea, spike, decision]
---

# OSC-11 startup-latency spike — is tskflwctl exposed?

Step 1 of [migrate-the-tui-to-charm-v2-bubbletea-bubbles-lipgloss](../tasks/6ff3hpm00wy5-migrate-the-tui-to-charm-v2-bubbletea-bubbles-lipgloss.md): verify, before
the port, whether the bubbletea-v1 init/OSC-11 latency the v2 docs tout actually
affects us. **Verified from our pinned deps + measured.**

## TL;DR — real but BOUNDED (human-TTY only; agents/pipes pay nothing)

The mechanism is present and fires on every invocation, but it is **hard TTY-
gated**: on a non-TTY (agent, pipe, redirect, CI) there is **no query and no
wait**. The worst case (up to 5s) is an *interactive-terminal* startup delay, and
only on terminals that don't answer OSC-11. So it's a niche human-face latency
risk — **not** the universal/agent-contract problem the headline framing implied.
The v2 migration removes it, but treat it as a **bonus**, not the primary
justification.

## Evidence (pinned versions)

- **The init exists and calls it.** `bubbletea@v1.3.10/tea_init.go`:
  ```go
  func init() {
      // ... Programs that use Lip Gloss/Termenv might hang while waiting for a
      // [termenv.OSCTimeout] while querying the terminal ... removed in v2.
      _ = lipgloss.HasDarkBackground()   // line 21
  }
  ```
  → `lipgloss@v1.1.1 HasDarkBackground()` → `termenv@v0.16.0 Output.BackgroundColor()`.
- **Timeout is 5s.** `termenv@v0.16.0/termenv_unix.go`: `OSCTimeout = 5 * time.Second`,
  used by `o.waitForData(OSCTimeout)` on the query path.
- **Fires on every command.** bubbletea is in tskflwctl's import graph
  (`go list -deps ./cmd/tskflwctl` → 1) via the `ui` command, and Go runs a
  package's `init()` whenever it's linked in — so it runs for `task list`,
  `epic show`, every subcommand, not just `ui`.
- **But it's TTY-gated.** `termenv Output.BackgroundColor()` guards the query:
  ```go
  if !o.isTTY() { return }   // → no OSC write, no waitForData, no 5s
  ```
  And the init's own comment: "only affect programs running on the default IO
  i.e. os.Stdout and os.Stdin."
- **Measured (non-TTY, this container):** `version` / `epic list` with stdout
  piped → **7–32 ms**. No OSC penalty when not a TTY. (The interactive-TTY
  worst case can't be reproduced here — no pty — but the source is conclusive.)

## What this means

| Context | Exposed? |
| :-- | :-- |
| Agent / pipe / redirect / CI (non-TTY) | **No** — short-circuits, ~0 cost. The machine contract is unaffected. |
| Interactive TTY, terminal answers OSC-11 (most modern terminals) | Negligible (~ms). |
| Interactive TTY, terminal does NOT answer (some tmux/screen/SSH/terminal combos) | **Up to 5s per command** — the real victim. |

No clean interim fix in v1: the `init()` is inside bubbletea and unconditional
(only termenv's internal `isTTY` gate stops it), and bubbletea is linked into the
single binary, so we can't skip it per-subcommand without a separate build. **The
v2 migration is the fix** (init removed; bubbletea v2 owns the terminal query).

## Decision

- Keep the migration's **primary** justifications: consolidate the two lipgloss
  majors + unlock the v2 feature wins (progress bars, layers, clipboard) +
  the new renderer.
- **Re-rank OSC-11 from "headline perf win" to "real-but-niche bonus"** — it helps
  interactive users on OSC-11-unresponsive terminals, nothing for agents.
- Correct the migration task brief accordingly (done).

## Update (2026-06-23) — the migration alone was NOT enough; huh v2 was

A correction the port surfaced: v1 `bubbletea` was linked into the binary by **two**
paths, not one. Besides the TUI's own (now-v2) usage, `internal/cli/prompt` used
`github.com/charmbracelet/huh` (v1) for the `Text` prompt, and huh v1 pulls v1
bubbletea/lipgloss/termenv transitively. So after porting *our* TUI to v2, the v1
`bubbletea.init()` (the OSC-11 query) was STILL in the binary — `go tool nm` showed
3 `charmbracelet/bubbletea.init` symbols.

The actual fix landed only when **huh moved to v2** (`charm.land/huh/v2` v2.0.3, on
the v2 charm stack). Migrating `tty.go` to it removed the last v1 path:
- `go list -m all` → no `github.com/charmbracelet/{bubbletea,bubbles,lipgloss,glamour}`
  or `muesli/{termenv,reflow}`.
- `go tool nm` → **0** `charmbracelet/bubbletea.init` symbols (was 3).

So the spike's "the v2 migration is the fix" is true, but only with **huh v2
included** — porting the TUI without retiring huh would have left the init in place.

## Sources
- `bubbletea@v1.3.10/tea_init.go`, `lipgloss@v1.1.1/renderer.go`,
  `termenv@v0.16.0/{termenv.go,output.go,termenv_unix.go}` (module cache).
- [2026-06-23-tui-v2-migration-plan](2026-06-23-tui-v2-migration-plan.md) · [2026-06-23-lipgloss-v2-charm-ecosystem](2026-06-23-lipgloss-v2-charm-ecosystem.md).
