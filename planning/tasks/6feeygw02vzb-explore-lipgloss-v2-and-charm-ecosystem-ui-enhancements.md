---
schema: 1
status: completed
epic: 20-cli-ux-and-ergonomics
description: Capture UI options unlocked by lipgloss v2 (tree/table/list) and stable bubbletea v2 once fang lands — research/decide, not build
effort: Unknown
tier: 2
priority: low
autonomy_level: 3
tags: [cli, ux, research]
created: "2026-06-21"
updated_at: "2026-06-23"
started_at: "2026-06-23"
completed_at: "2026-06-23"
id: 6feeygw02vzb
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
   epic [18-tui-bubble-tea-interactive-planning-browser](../epics/18-tui-bubble-tea-interactive-planning-browser.md).

**Adjacent charm tooling:**

5. **`charmbracelet/wish`** — serve the existing TUI over SSH with ~zero new UI
   code. Already noted in the web-companion research; cross-ref epic
   [19-web-companion-apps-over-a-shared-core](../epics/19-web-companion-apps-over-a-shared-core.md).
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
  [evaluate-fang-for-styled-help-errors-and-manpages](6fdtbb4037bs-evaluate-fang-for-styled-help-errors-and-manpages.md); this task assumes its
  dependency graph as the starting point.
- Domain-model / machine-contract changes.

## Related

- Source: `planning/research/2026-06-21-fang-evaluation-spike.md` (the spike that
  surfaced lipgloss v2 + bubbletea v2 availability).
- Epic [20-cli-ux-and-ergonomics](../epics/20-cli-ux-and-ergonomics.md) · TUI epic
  [18-tui-bubble-tea-interactive-planning-browser](../epics/18-tui-bubble-tea-interactive-planning-browser.md) · web epic
  [19-web-companion-apps-over-a-shared-core](../epics/19-web-companion-apps-over-a-shared-core.md).
- Sibling human-face work: [evaluate-fang-for-styled-help-errors-and-manpages](6fdtbb4037bs-evaluate-fang-for-styled-help-errors-and-manpages.md),
  [glamour-render-markdown-bodies-in-show](6fdtbb40306e-glamour-render-markdown-bodies-in-show.md).

## Decision (2026-06-23)

Full writeup: `planning/research/2026-06-23-lipgloss-v2-charm-ecosystem.md`.

- **Greenlight (small, low-risk):** lipgloss/v2 `tree` for `epic show` (recommended FIRST slice — net-new human surface, no goldens to break, v2 already in the graph); `vhs` doc GIFs (tooling, not a runtime dep).
- **Decline/defer:** `table` for list output (machine-contract + golden risk for ~0 value); `list` bullets (minor); `charmbracelet/log` (muddies the --json/stderr contract); `wish` → epic 19.
- **Own epic-18 sprint (not here):** TUI → bubbletea v2 + bubbles v2 + lipgloss v2. Mechanical but large; pays down debt — consolidates the two lipgloss majors onto v2, gets the new renderer, and removes bubbletea v1's init()/OSC-11 ~5s-timeout footgun (bubbletea is in every binary's import graph — verify impact as step 1).

Verified current state: bubbletea v2.0.0 stable (Feb 2026, ~v2.0.6; v1 frozen), bubbles v2.0.0 stable, lipgloss v2 stable (repo on v2.0.4). Proposed follow-ups listed in the doc.
