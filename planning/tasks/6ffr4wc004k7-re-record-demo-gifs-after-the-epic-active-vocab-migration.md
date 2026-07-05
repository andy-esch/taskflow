---
schema: 1
status: ready-to-start
epic: 20-cli-ux-and-ergonomics
description: The epic status migration (planning -> active etc.) re-staled the README demo GIFs; re-run just gifs against assets/demo-planning to refresh tui/status/audit-show.gif.
effort: S
tier: 3
priority: low
autonomy_level: 3
tags: [docs]
created: "2026-06-25"
id: 6ffr4wc004k7
---
## Objective

The epic-vocab migration changed epic statuses to active/retired/deprecated, so the committed demo GIFs (which show epic status) are stale again. Re-record them.

## Acceptance criteria

- [ ] `just gifs` re-run; assets/{tui,status,audit-show}.gif reflect the active/retired/deprecated vocab + current output.
- [ ] No README/link changes needed (paths unchanged).

## Notes

Needs the `vhs` tool (host/CI, not in the dev container). Runs against [regenerate-demo-gifs-for-the-new-progress-and-symbology-output](6ffr4wc010sj-regenerate-demo-gifs-for-the-new-progress-and-symbology-output.md)'s assets/demo-planning fixture.