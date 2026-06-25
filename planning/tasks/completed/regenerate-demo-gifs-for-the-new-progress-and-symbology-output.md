---
schema: 1
status: completed
epic: 20-cli-ux-and-ergonomics
description: Demo GIFs (status/epic-show/audit) still show pre-symbology output; regenerate the assets/vhs tapes so README matches bars, finding glyphs, and the Open-audits dashboard section.
effort: S
tier: 3
priority: low
autonomy_level: 3
tags: [docs, tui]
created: "2026-06-25"
updated_at: "2026-06-25"
started_at: "2026-06-25"
completed_at: "2026-06-25"
---
# Regenerate demo GIFs for the new progress + symbology output

## Objective

The 2026-06-25 symbology work changed what several commands print — epic rows
gained progress bars, `status` gained an Open-audits section, `audit list` shows
a bar instead of `N/M open`, and `audit findings` is status-glyph-coded. The
committed demo GIFs (`assets/status.gif`, `assets/epic-show.gif`,
`assets/task-list.gif`, `assets/help.gif`) still show the **old** output, so the
README's screenshots no longer match the tool. Regenerate them.

## Context

The tapes live in `assets/vhs/*.tape` (status, epic-show, task-list, help) and
render via [vhs](https://github.com/charmbracelet/vhs). README embeds the GIFs
(table at the top + inline). `status.gif` and `epic-show.gif` are the most stale
(bars / audits section); `task-list.gif` is least affected. Relates to epic 20
(the original `vhs-terminal-gifs-for-readme-and-cli-docs` task).

## Acceptance criteria

- [ ] `assets/*.gif` regenerated from the current binary so they show bars,
      glyphs, and the Open-audits section.
- [ ] Tapes still pin a date-stable fixture (no churny timestamps) and a fixed
      width so diffs stay reviewable.
- [ ] README renders correctly with the new GIFs; no broken links.

## Risks / gotchas

- `vhs` likely isn't in the dev container (same story as terraform/just) — this
  may need to run on the host or in CI with the tool installed.
- Keep the recorded planning fixture one that actually exercises the new output
  (at least one open audit, an epic mid-progress) or the GIFs won't show the
  change.

## Done when

The README's GIFs match what the tool prints today — bars, finding glyphs, and
the Open-audits dashboard section all visible.

## Prep landed 2026-06-25 (recording still owed)

Decision: feature **TUI hero + status + audit show** (not the old four).

Done in-repo:
- Curated [`assets/demo-planning/`](../../../assets/demo-planning/) fixture — 3 epics
  mid-progress (50/33/100%), 8 tasks across every status, one open audit whose
  findings span fixed/landed/in-progress/open/deferred/wontfix (so the segmented
  bar shows all bands) + one closed audit. Generated with the tool; lint-clean.
- New tapes: `assets/vhs/tui.tape` (hero), `assets/vhs/audit-show.tape`; rewrote
  `status.tape`. All `cd` (hidden) into the demo fixture so recorded commands stay clean.
- README `## Demos` rewired: hero TUI gif + status + audit-show, with a note on the fixture.
- `assets/vhs/README.md` updated.

Remaining (needs `vhs`, not in the dev container — host/CI):
- Run `just gifs` to record `tui.gif`, `status.gif`, `audit-show.gif`.
- Optionally prune the now-unreferenced `help`/`epic-show`/`task-list` tapes + gifs.

## Recorded + documented 2026-06-25

- GIFs recorded (`tui.gif`, `status.gif`, `audit-show.gif`) against the demo fixture.
- Added [`assets/README.md`](../../../assets/README.md) — a demos hub embedding the
  three GIFs and pulling together the sub-READMEs ([`vhs/`](../../../assets/vhs/README.md)
  recording + [`demo-planning/`](../../../assets/demo-planning/README.md) data); added
  [`assets/demo-planning/README.md`](../../../assets/demo-planning/README.md); linked
  the hub from the root README's Demos section.
