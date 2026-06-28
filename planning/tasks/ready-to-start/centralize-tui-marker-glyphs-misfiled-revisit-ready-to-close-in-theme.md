---
schema: 1
status: ready-to-start
epic: 21-code-quality-architecture-hardening
description: The misfiled/revisit/ready-to-close markers (⚠/↻/✓) are hand-typed in item.go, dashboard.go, and the help legend; promote them to theme tokens so all surfaces share one source.
effort: S
tier: 3
priority: low
autonomy_level: 3
tags: [tui, theme, maintainability]
created: "2026-06-28"
---
## Why

From the 2026-06-28 adversarial review of the `?` symbol-legend work. The legend glyph ROWS are correctly drift-proof (sourced from `theme.Status/Liveness/Bucket/FindingStatus` + `domain.AllStatuses()`/`AllAuditBuckets()`). But the MARKER glyphs are still hand-typed string literals in multiple surfaces:

- `internal/tui/item.go` — `⚠` (misfiled), `↻` (revisit-due) on task rows.
- `internal/tui/dashboard.go` — `⚠` (misfiled/non-conforming), `↻` (revisit), `✓` (ready to close), `✔` (all clear).
- `internal/tui/help.go` `symbolsFor` — the legend re-types `⚠ ↻ ✓` to describe them.

Concrete latent drift already present: the legend uses `✓` U+2713 for "ready to close" (matching `dashboard.go`) but the word "done" elsewhere is `✔` U+2714 — two checkmarks reconciled only by eye. The review scoped the legend comment to admit the `mark` rows are hand-labeled; this task removes the underlying duplication.

## Goal

Promote the cross-surface markers to `theme` tokens (e.g. `theme.MarkerMisfiled`, `theme.MarkerRevisit`, `theme.MarkerReadyToClose` — glyph + Color, like `theme.Status`), and have item.go, dashboard.go, and the help legend all read them. Then the legend `mark` rows can source the glyph from theme too, restoring the "can’t drift" property for markers.

## Notes / acceptance

- Low priority / small: the current state is correct, just duplicated; this is a DRY + drift-proofing pass.
- Natural sibling of the presentation-seam work (`shared-presentation-seam-for-the-bar-percent-counts-composite`, H5/M1/M10) — keeps semantic glyph+color decisions in `theme`.
- Reconcile the `✓`/`✔` usage while here.