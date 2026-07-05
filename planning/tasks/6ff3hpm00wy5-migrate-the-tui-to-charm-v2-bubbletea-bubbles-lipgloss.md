---
schema: 1
status: completed
epic: 18-tui-bubble-tea-interactive-planning-browser
description: Move internal/tui + render to bubbletea/bubbles/lipgloss v2; retire the v1 lipgloss major; harvest the free perf wins incl. the OSC-11 init removal.
effort: Unknown
tier: 3
priority: high
autonomy_level: 3
tags: [tui, perf, migration]
created: "2026-06-23"
updated_at: "2026-06-23"
started_at: "2026-06-23"
completed_at: "2026-06-23"
id: 6ff3hpm00wy5
---
# Migrate the TUI to charm v2 (bubbletea + bubbles + lipgloss)

The foundation for the v2 sprint. Move `internal/tui` (~7k LOC, all v1) to
bubbletea v2 + bubbles v2 + lipgloss v2, and consolidate `internal/cli/render`
onto lipgloss v2 so the v1 major is fully retired (fang already brought v2 in).
Plan: `planning/research/2026-06-23-tui-v2-migration-plan.md`.

## Why (beyond keeping current)
- **Perf, broad:** v1 ships a package-level `init()` that pings the terminal
  (OSC-11) for background color and can hang ~5s when unanswered; v2 removed it.
  bubbletea is in EVERY binary's import graph, so this can tax even `task list`.
  **Step 1: measure the v1 latency on a TTY** for real before/after evidence.
- Free with v2: the Cursed renderer (faster, low-bandwidth), synchronized output
  (no flicker on live-reload), wide-Unicode (no layout breakage), auto color
  downsampling.
- Retires the two-lipgloss-majors debt the fang adoption introduced.

## Scope
1. Port `internal/tui`: `View()`→`tea.View` struct; `KeyMsg`→`KeyPressMsg`
   (`.Type`→`.Code`, `.Runes`→`.Text`, `" "`→`"space"`); mouse types→interfaces;
   bubbles getter/setters (`SetWidth`/`Width`). Charm's old→new lookup table is
   search-and-replaceable.
2. Move `internal/cli/render` to lipgloss/v2 (consolidation). lipgloss v2 `Color`
   is now `image/color.Color`; `AdaptiveColor` is gone (manual light/dark — the
   `internal/theme` pkg already centralizes this, so it's contained).
3. Confirm a glamour release compatible with lipgloss v2 (markdown `show` + TUI).

## Acceptance criteria
- [ ] TUI + render build on bubbletea/bubbles/lipgloss v2; the v1 lipgloss major
      is gone from `go.mod`.
- [ ] **Machine contract intact:** `-o table`/`csv`/`json` byte-identical;
      goldens untouched (or regenerated only for deliberate human-output changes).
- [ ] OSC-11 latency measured before/after (the perf claim, evidenced).
- [ ] Suite + lint green; TUI teatest cases pass on v2.

## Out of scope
- The harvest tasks (progress bars, the designed feature wins) — they depend on this.

## Related
- Plan: `planning/research/2026-06-23-tui-v2-migration-plan.md`; decision
  `planning/research/2026-06-23-lipgloss-v2-charm-ecosystem.md`.
- [[18-tui-bubble-tea-interactive-planning-browser]].

## OSC-11 spike result (2026-06-23)

Step-1 spike done: `planning/research/2026-06-23-osc11-startup-latency-spike.md`. The mechanism IS present in our pinned deps (bubbletea v1.3.10 tea_init.go init() → lipgloss.HasDarkBackground() → termenv, OSCTimeout=5s) and fires on every invocation (bubbletea in the import graph). BUT it's hard TTY-gated: termenv short-circuits `if !o.isTTY()` before the query, so agents/pipes/redirects/CI pay nothing (measured non-TTY startup 7–32ms). The up-to-5s worst case is interactive-TTY-only AND only on terminals that don't answer OSC-11.

**Re-rank:** OSC-11 is a real-but-NICHE bonus (helps some interactive users), NOT a headline/agent-wide perf win. The migration's primary justifications stand: consolidate the two lipgloss majors + the v2 feature wins (progress bars/layers/clipboard) + the new renderer. No clean v1 interim fix — v2 is the fix.
