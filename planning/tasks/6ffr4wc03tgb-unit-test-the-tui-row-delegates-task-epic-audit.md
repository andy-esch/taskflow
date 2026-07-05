---
schema: 1
status: completed
epic: 18-tui-bubble-tea-interactive-planning-browser
description: The task/epic/audit list-row delegates have no direct render test; the symbology pass made the rows richer (bars, glyphs) so a row regression would slip past CI. Add focused render tests.
effort: S
tier: 3
priority: low
autonomy_level: 3
tags: [tui]
created: "2026-06-25"
updated_at: "2026-06-28"
completed_at: "2026-06-28"
id: 6ffr4wc03tgb
---
# Unit-test the TUI row delegates (task / epic / audit)

## Objective

The TUI list rows are rendered by `taskDelegate`, `epicDelegate`, and
`auditDelegate` (`internal/tui/item.go`) — the most-seen UI in the tool — and
none of them has a direct render test. The 2026-06-25 symbology pass made the
audit row materially richer (bucket glyph + progress bar + resolved/total,
replacing a single `■` + `N/M open`), and the epic row carries the bar too, yet a
regression in either would slip past CI. Add focused render tests.

## Context

`*.Render(w, m, index, item)` takes a `bubbles/v2/list.Model`; the audit *detail*
is tested (`TestAuditDetailFindingIndex`, `model_test.go`) but the list *rows* are
not. The glyph/colour vocabulary is pinned in `theme` tests and the CLI render
tests, so these only need to assert the row composition (glyph present, bar
present, counts in `resolved/total` shape, slug + dim area) — ANSI-stripped.
Relates to epic 18 (TUI). Pre-existing gap for all three delegates, not just audit.

## Acceptance criteria

- [ ] A render test per delegate builds a minimal `list.Model` (one item, a set
      width), calls `Render` into a buffer, and asserts the ANSI-stripped row:
      - task: status glyph + slug + relative date; ⚠ when misfiled.
      - epic: bar + `%` + `done/total` + id + dim description.
      - audit: bucket glyph + bar + `%` + `resolved/total` + slug + dim area.
- [ ] Tests are width-stable (no terminal-size flakiness) and ANSI-tolerant.
- [ ] go build/test/lint green.

## Risks / gotchas

- Constructing a `list.Model` in a test is a little fiddly — check whether
  `model_test.go` already has a list/model helper to reuse before hand-rolling one.
- Don't re-assert the glyph→colour mapping (that's `theme`'s job) — assert the row
  *layout*, so the test doesn't duplicate the theme decision table.

## Done when

Each row delegate has a render test that would fail if the bar, glyph, or column
order regressed — closing the coverage gap the symbology work highlighted.

## Completed 2026-06-28

Added `internal/tui/item_test.go` with a render test per delegate (`TestTaskDelegateRow`
+ a `_Misfiled` variant, `TestEpicDelegateRow`, `TestAuditDelegateRow`). Each builds a
one-item `list.New(...)`, renders the row via `delegate.Render` (at index 1 so the
shared "› " cursor is absent), strips ANSI, and asserts via an `assertColumns` helper
that every column is present AND in left-to-right order — so a dropped glyph/bar OR a
reordered column fails. The expected pieces are computed through the production
helpers (`theme.Status/Bucket`, `epicGlyph`, `miniBar`/`segBar`, `theme.Counts`,
`PercentLabelPadded`, `RelativeDate`), so the tests pin row LAYOUT without duplicating
the glyph→colour decision table (theme's own tests cover that). build/vet/test/lint green.
