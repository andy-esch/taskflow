---
schema: 1
status: ready-to-start
epic: 20-cli-ux-and-ergonomics
description: Demo GIFs (status/epic-show/audit) still show pre-symbology output; regenerate the assets/vhs tapes so README matches bars, finding glyphs, and the Open-audits dashboard section.
effort: S
tier: 3
priority: low
autonomy_level: 3
tags: [docs, tui]
created: "2026-06-25"
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
