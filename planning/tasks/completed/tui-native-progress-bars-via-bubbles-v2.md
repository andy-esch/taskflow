---
schema: 1
status: completed
epic: 18-tui-bubble-tea-interactive-planning-browser
description: 'No-brainer harvest: replace hand-rolled miniBar + Style.Bar with the bubbles v2 progress component (net code deletion). Depends on the v2 migration.'
effort: Unknown
tier: 3
priority: medium
autonomy_level: 3
tags: [tui, ux]
created: "2026-06-23"
updated_at: "2026-06-23"
completed_at: "2026-06-23"
---
# TUI native progress bars via bubbles v2

No-brainer harvest after the v2 migration: replace the hand-rolled bars with the
bubbles v2 `progress` component — **net code deletion**, less to maintain.

## Scope
- Retire `miniBar` (`internal/tui/style.go:62`; used in `detail.go` epic progress
  + `item.go` epic-list rows).
- Retire `Style.Bar` (`internal/cli/render/style.go:121`; the `status` dashboard
  epic bars). `progress.ViewAs(pct)` renders a static styled bar, so it works in
  the non-interactive `render` path too.
- Keep the rendered percentage consistent with the corrected rollup (deprecated
  excluded from the denominator).

## Acceptance criteria
- [ ] `miniBar` + `Style.Bar` removed; both call sites use bubbles v2 `progress`.
- [ ] TUI + `status` bars render correctly; teatest/goldens updated as needed.
- [ ] Net LOC down; suite + lint green.

## Depends on
- The v2 migration (bubbles v2 must be in the module graph).

## Related
- Plan: `planning/research/2026-06-23-tui-v2-migration-plan.md`.
- [[18-tui-bubble-tea-interactive-planning-browser]].
