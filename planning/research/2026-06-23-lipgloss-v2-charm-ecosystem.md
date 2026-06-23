---
status: reference
created: "2026-06-23"
tags: [cli, ux, tui, lipgloss, bubbletea, research, decision]
---

# lipgloss v2 / charm-ecosystem UI options — findings & decision

Capture-and-decide for [[explore-lipgloss-v2-and-charm-ecosystem-ui-enhancements]]
(epic [[20-cli-ux-and-ergonomics]]). Builds on the
[[2026-06-21-fang-evaluation-spike]] (fang landed → `charm.land/lipgloss/v2`
v2.0.4 is already in the module graph). **No code here — options + a recommended
first slice.** Governing rule (epic 20): never compromise the agent/pipeline
contract — every visual nicety is TTY-gated and raw under `--json`/pipe.

## TL;DR

- **Greenlight now (small, low-risk):** lipgloss/v2 **`tree` for `epic show`** (the
  recommended first slice — net-new human surface, no golden tests to break, v2
  already in the graph) and **`vhs`** doc GIFs (tooling, not a runtime dep).
- **Decline / defer:** `table` for list output (machine-contract + golden risk for
  ~0 value — the porcelain `-o table` already exists); `list` bullets (minor);
  `charmbracelet/log` (would muddy the `--json`/stderr discipline); `wish` →
  belongs to epic [[19-web-companion-apps-over-a-shared-core]], not here.
- **Scope as its own epic-18 sprint (NOT here):** migrate the TUI to bubbletea
  v2 + bubbles v2 + lipgloss v2. Big but mechanical, and it pays down real debt
  (consolidates the two lipgloss majors, gets the new renderer, and removes a
  latency footgun — see below).

## Ecosystem state (verified 2026-06-23)

| lib | status | notes |
| :-- | :-- | :-- |
| **lipgloss v2** | **stable** (repo on `charm.land/lipgloss/v2` v2.0.4) | `table`/`tree`/`list` sub-packages at `charm.land/lipgloss/v2/{table,tree,list}`. `Color` is now a func returning `image/color.Color` (not a string); `AdaptiveColor` removed; `lipgloss.Println` auto-downsamples. |
| **bubbletea v2** | **stable v2.0.0 (Feb 2026), latest ~v2.0.6; v1 frozen** | Declarative `View()` returns a `tea.View` struct; `KeyMsg`→`KeyPressMsg` (`.Type`→`.Code`, `.Runes`→`.Text`); mouse types are interfaces; new ncurses-style "Cursed" renderer (perf). |
| **bubbles v2** | **stable v2.0.0 (Feb 24 2026)** | Requires bubbletea v2 + lipgloss v2 (upgrade all three together). Getter/setters (`SetWidth`/`Width`) replace exported fields across filepicker/help/progress/table/textinput/viewport. |
| **fang** | adopted (v1.0.0) | Already on lipgloss/v2 — the reason v2 is in our graph. |

(The task's seed said "v2.1.0"; the verified current line is the stable v2.0.x
series — immaterial to the calls below.)

Today's repo: TUI + `internal/cli/render` are on **lipgloss v1** (v1.1.1) +
**bubbletea v1.3.10** + **bubbles v1**; **fang pulls lipgloss/v2 alongside** for
its own use. So we already carry **two lipgloss majors** — the migration option
(4) is what retires that.

## Per-option calls

**1. lipgloss/v2 `tree` for `epic show` / planning hierarchy — ✅ ADOPT (first slice).**
Render an epic → its member tasks (grouped by status) as a styled tree; later, a
`status` planning tree. *Value:* genuine net-new visual on the human face.
*Risk:* low — `epic show`'s human output has no byte-stable golden (the
golden-locked surface is `epic show --json`, untouched); TTY-gated, raw under
`--json`/pipe per the rule. *Effort:* small. *Caveat:* it introduces lipgloss
**v2** styles into `render` (which otherwise uses v1 `Style`) — a self-contained
block, fine locally, and a nudge toward eventually moving `render` onto v2.

**2. `table` for `task list`/`epic list` — ❌ DECLINE (machine-contract risk, ~0 value).**
The porcelain `-o table`/`csv`/`json` are deliberately byte-stable, golden-locked,
ANSI-free pipeline contracts — `lipgloss/table` (borders/ANSI) must never touch
them. That leaves only the *default human* `task list` (already a styled
`writeTable`) as a candidate reskin: marginal visual gain, real golden/contract
risk, and a second table engine to maintain. Not worth it now. Revisit only if a
broader human-face polish is greenlit, and even then gate it hard.

**3. `list` for audit findings / candidate lists — 🟡 MINOR, skip for now.**
Styled nested bullets. Low value over the existing findings table; low risk.
Park it; fold into a future human-face polish if ever.

**4. TUI → bubbletea v2 + bubbles v2 + lipgloss v2 — 📦 SCOPE AS ITS OWN EPIC-18 SPRINT.**
Not for this task. *Effort:* large — a full TUI port: `View()`→`tea.View`,
`KeyMsg`→`KeyPressMsg`, mouse interfaces, bubbles getter/setters, `AdaptiveColor`
removal (manual light/dark — note our `theme` pkg already centralizes this).
Charm calls it "mechanical, not conceptual" with an LLM-friendly old→new lookup
table, but it's still the whole `internal/tui`. *Upside beyond polish (why it's
worth a sprint):*
  - **Consolidates the two lipgloss majors** (v1 TUI + v2 fang) onto one v2 —
    retires the debt the fang adoption introduced.
  - **New "Cursed" renderer** — faster, lower bandwidth (matters if `wish` ever
    serves the TUI).
  - **Removes a latency footgun:** bubbletea **v1** ships a package-level `init()`
    that calls `lipgloss.HasDarkBackground()`; v2 removed it because it caused a
    **~5s OSC-11 timeout** on terminals that don't answer. bubbletea is in *every*
    tskflwctl binary's import graph (via the TUI), so this can tax even
    `task list`. ⚠️ **Verify whether tskflwctl is actually affected today**
    (TTY-only; agents/pipes likely short-circuit) — if so, this elevates the
    migration from "polish" to "fixes a real startup-latency bug."

**5. `wish` (TUI over SSH) — ↪️ DEFER to epic 19.** Real (serve the TUI over SSH
with ~zero new UI code; v2's renderer makes it cheaper), but it's a web-companion
direction. Cross-ref [[19-web-companion-apps-over-a-shared-core]]; don't open here.

**6. `vhs` (terminal GIFs for docs) — ✅ GREENLIGHT (small).** Script GIFs of
`--help`/`status`/the TUI for the README + generated CLI docs. Tooling, **not a
runtime dep**; cheap docs win; zero contract surface.

**7. `charmbracelet/log` — ❌ DECLINE.** The tool's contract is the `--json`
envelope + strict stderr discipline (diagnostics only, no chatter). A styled
logger risks muddying both. No.

## Proposed follow-up tasks (for whatever's greenlit)

1. **`epic show` lipgloss/v2 `tree` rendering** (epic 20) — the first slice;
   TTY-gated, `--json` untouched. *(greenlit)*
2. **`vhs` doc GIFs** (epic 20) — README + CLI-docs GIFs of help/status/TUI;
   tooling only. *(greenlit, small)*
3. **TUI v2 migration** (epic 18) — bubbletea v2 + bubbles v2 + lipgloss v2; its
   own sprint, with the lipgloss-consolidation + renderer + OSC-11-init upside in
   the brief, and a first step that **confirms the OSC-11 latency** on v1. *(scoped)*

Declined/parked, recorded so they're not re-litigated: `table` reskin (2), `list`
bullets (3), `charmbracelet/log` (7); `wish` (5) tracked under epic 19.

## Sources

- Bubble Tea v2 upgrade guide: https://github.com/charmbracelet/bubbletea/blob/main/UPGRADE_GUIDE_V2.md
- Bubbles v2 upgrade guide: https://github.com/charmbracelet/bubbles/blob/main/UPGRADE_GUIDE_V2.md
- lipgloss v2 (`charm.land/lipgloss/v2`) + `/tree` `/list` `/table`: https://pkg.go.dev/charm.land/lipgloss/v2
- Bubbles v2.0.0 release (Feb 24 2026): https://github.com/charmbracelet/bubbles/releases/tag/v2.0.0
- Bubble Tea v2 "What's New": https://github.com/charmbracelet/bubbletea/discussions/1374
