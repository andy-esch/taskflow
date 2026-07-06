---
status: completed
epic: 17-pm-go-cli
description: Add an UPDATED relative-date column and terminal-width-aware truncation of the description column, pipe-safe and TTY-driven
effort: Unknown
tier: 3
priority: medium
autonomy_level: 3
tags: [pm-tooling, go, cli, ergonomics]
created: "2026-06-09"
updated_at: "2026-06-09"
completed_at: "2026-06-09"
id: 6fakbec0073w
---

# Relative dates and width-aware truncation in task list

## Objective

The two output-formatting follow-ups that make `task list` feel finished, after
the color/glyph work.

## Done

- **Relative dates:** `domain.Task` gained `Updated` (`updated_at`); `task list`
  has an `UPDATED` column rendering `RelativeDate` (`today`/`3d ago`/`2w ago`/
  `5mo ago`/`1y ago`, falling back to `created`). `task show` shows the absolute
  date + `(relative)`. `taskJSON` now carries `created`/`updated_at` for agents.
- **Width-aware truncation:** `writeTable` takes a `maxWidth`; the description
  (last) column truncates with `…` to fit, never below the header label. Width
  comes from the terminal (`golang.org/x/term`, added) with a `COLUMNS`
  override; **0 = no limit**, so piped/non-TTY output stays full-width.
- `Style` carries the width (`WithWidth`), set in `PersistentPreRunE`.

## Acceptance

- [x] `UPDATED` column with relative dates; `--json` carries the dates.
- [x] Truncation fits `COLUMNS=70`, full width when piped; tests for
      `relativeDateFrom`, truncation-to-width, and no-limit-keeps-full. Full
      suite + lint green; demoed.

## Out of scope

- `glamour`-rendered `show` body (its own task); a `--sort` flag.

## Related

- Epic [17-pm-go-cli](../epics/17-pm-go-cli.md); follows
  [cli-color-glyphs-table-headers-render-styling](6fakbec017kj-cli-color-glyphs-table-headers-render-styling.md); unblocks the recency/stale
  parts of [at-a-glance-dashboard-status-board](6fakbec01wt8-at-a-glance-dashboard-status-board.md).
