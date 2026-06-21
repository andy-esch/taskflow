---
schema: 1
status: ready-to-start
epic: 20-cli-ux-and-ergonomics
description: Capture UI options unlocked by lipgloss v2 (tree/table/list) and stable bubbletea v2 once fang lands — research/decide, not build
effort: Unknown
tier: 2
priority: low
autonomy_level: 3
tags: [cli, ux, research]
created: "2026-06-21"
---
## Objective

Capture the UI/rendering enhancements that become cheap **once fang lands and
`lipgloss/v2` is in the module graph** (it now is — stable v2.0.4 — via the fang
spike, see `planning/research/2026-06-21-fang-evaluation-spike.md`). lipgloss v2
ships `table`/`tree`/`list` sub-packages, and **bubbletea v2 / bubbles v2 are now
stable (v2.1.0)**. This is a *capture-and-decide* task — enumerate options with a
recommendation; do not build here. Governing rule (epic 20): never compromise the
agent/pipeline contract — every visual nicety is TTY-gated, raw under `--json`/pipe.

## Options to weigh (seed material, not conclusions)

**lipgloss/v2 rendering primitives (CLI human face):**

1. **`tree` for `epic show` / planning hierarchy** — render epic → member tasks
   (grouped by status), or a `status` planning tree, as a styled tree. Pure
   net-new visual; lowest risk (new surface, no existing golden tests to break).
   Likely the first slice.
2. **`table` for `task list` / `epic list`** — replace/augment the hand-rolled
   `internal/cli/render` tables with `lipgloss/table` (borders, alignment, wide-
   char aware). ⚠️ Touches a tested, machine-relevant surface: must stay TTY-gated,
   `-o table`/`--json`/pipe output byte-identical, golden tests reworked. Higher
   risk — weigh carefully or decline.
3. **`list` for audit findings / candidate lists** — styled nested bullets. Minor.

**Stack consolidation (TUI, epic 18):**

4. **Migrate the TUI from bubbletea/bubbles/lipgloss v1 → v2** (now stable). Big
   effort, its own sprint. Upside beyond polish: it **consolidates the two
   lipgloss majors** the fang adoption introduces (TUI on v1 + fang on v2) onto a
   single v2, retiring that debt. `lipgloss/v2 compat` eases the port. Cross-ref
   epic [[18-tui-bubble-tea-interactive-planning-browser]].

**Adjacent charm tooling:**

5. **`charmbracelet/wish`** — serve the existing TUI over SSH with ~zero new UI
   code. Already noted in the web-companion research; cross-ref epic
   [[19-web-companion-apps-over-a-shared-core]].
6. **`vhs`** — script terminal GIFs of help/TUI/`status` for the README + the
   generated CLI docs. Tooling, not a runtime dep. Cheap docs win.
7. **`charmbracelet/log`** — styled structured logging. *Considered, likely
   declined:* the tool's contract is the `--json` envelope + strict stderr
   discipline; a logger risks muddying that. Note the reasons and move on.

## Acceptance criteria

- [ ] Doc/decision enumerating each option with a value/risk/effort note and a
      recommended first slice (likely lipgloss/v2 `tree` for `epic show`).
- [ ] Explicit call on `table` (adopt-gated vs decline) given the golden-test +
      machine-contract risk.
- [ ] The bubbletea-v2 TUI migration is scoped as its own follow-up under epic 18
      (with the lipgloss-consolidation upside called out), not folded in here.
- [ ] Follow-up tasks proposed for whatever is greenlit.

## Out of scope

- Any implementation (this is capture-and-decide).
- The fang adoption itself — that's
  [[evaluate-fang-for-styled-help-errors-and-manpages]]; this task assumes its
  dependency graph as the starting point.
- Domain-model / machine-contract changes.

## Related

- Source: `planning/research/2026-06-21-fang-evaluation-spike.md` (the spike that
  surfaced lipgloss v2 + bubbletea v2 availability).
- Epic [[20-cli-ux-and-ergonomics]] · TUI epic
  [[18-tui-bubble-tea-interactive-planning-browser]] · web epic
  [[19-web-companion-apps-over-a-shared-core]].
- Sibling human-face work: [[evaluate-fang-for-styled-help-errors-and-manpages]],
  [[glamour-render-markdown-bodies-in-show]].
