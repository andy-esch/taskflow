---
status: completed
epic: 18-tui-bubble-tea-interactive-planning-browser
description: 'Fresh-eyes follow-ups from S2b review: occurrence-level find, preserve field colors on matches, per-entity sort columns, View-purity, help scroll'
effort: Unknown
tier: 3
priority: medium
autonomy_level: 3
tags: [tui, bubble-tea, cleanup]
created: "2026-06-11"
updated_at: "2026-06-12"
started_at: "2026-06-12"
completed_at: "2026-06-12"
id: 6fb7ym4023q0
---

# TUI S2b polish: find occurrences, highlight fidelity, per-entity sort

## Objective

Deferred items from the S2b fresh-eyes review (2026-06-11). The three high-value
fixes (chip on reversed-default, sort-badge suppression under filter, unified
status-view table) landed in S2b; these remaining items are real but want more
thought than a sprint-close patch. Each is independent — pick up à la carte.

## Scope

- [ ] **Find: occurrence-level, not line-level** (`detail.go refreshFind`,
      `find.go`). `find.lines` holds one index per matching *line*, so a line with
      two hits counts once, `[1/N]` undercounts, `n`/`N` step line-to-line, and the
      "current" highlight brightens *every* occurrence on the line. Track matches as
      (line, column) occurrences: footer counts hits, `n`/`N` walk hits, only the
      focused hit gets the bright style.
- [ ] **Preserve field colors on matched lines** (`detail.go:~197`). Matched lines
      are rebuilt from `ansi.Strip`-ped text so a highlight can't split an escape —
      but that drops the line's original styling (e.g. `/in-progress` hitting the
      `status: ● in-progress` field row loses the colored glyph). Highlight over the
      *styled* line using an ANSI-aware splitter (`ansi.Cut`/segment walk) instead of
      rebuilding from plain. Invisible for body text; this only bites field rows.
- [ ] **Per-entity sort columns** (`sort.go`, `item.go`, `entity.go`). Sort columns
      are global, but epics have no tier/updated and audits no priority/tier — so
      cycling to those on those tabs shows a `sort:tier↓` chip while nothing reorders
      (everything ties → slug). Let each entity declare its own ordered sort columns
      (e.g. on the `entityTab`), so `o` only cycles meaningful ones per tab.
- [ ] **Chip sync out of `View()`** (`model.go listPaneContent`). The chip is
      written into `list.Title` from inside `View()` (mutating state via the tab
      pointer) because the `/` filter prompt renders in that same slot. It's
      idempotent but violates Bubble Tea's "View is pure" contract. Move the
      title-sync into `Update` after sort/view/filter transitions (note: the filter
      applies async via `FilterMatchesMsg`, so the forward-to-list path needs to
      re-sync too), or render the chip as its own line without losing the prompt.
- [ ] **Sort arrow semantics + help scroll** (low). The chip arrow means
      "default vs reversed," so `↓` is newest-first for `updated` but A→Z for `slug` —
      the same glyph implies opposite directions. Consider a true asc/desc indicator
      per column. Separately, the `?` help box isn't scrollable: on short terminals
      (≲16 rows) the `MaxHeight` clamp clips the List/Detail sections. Make it a
      scrollable viewport or paginate.

## Acceptance criteria

- [ ] Detail find reports and navigates individual occurrences; only the focused
      hit is brightened; matched field rows keep their colors.
- [ ] `o` cycles only the columns meaningful for the active entity; no chip shows a
      sort that doesn't reorder.
- [ ] The chip no longer requires mutating state inside `View()`.
- [ ] Suite + lint green.

## Out of scope

- Sort-under-filter actually reordering the filtered view — bubbles ranks filtered
  items by fuzzy score; the S2b fix (suppress the sort badge while filtering) is the
  accepted behavior unless we replace the list filter with an order-stable one.

## Related

- Epic [[18-tui-bubble-tea-interactive-planning-browser]]
- Follows [[tui-sprint-2b-search-status-views-and-interactive-sort]]

## Closure (2026-06-12)

Completed as the merged single pass (absorbing
[[tui-review-polish-batch-sort-rank-help-drift-width-audit-scope]]), built
across two agents: the macOS agent implemented (a) the find rewrite —
occurrence-level n/N (`matchPos`), rune-by-rune `foldMatches` that never
indexes a folded copy (U+0130-safe, regression-tested), `highlightLine`
preserving field colors via display-column `ansi.Cut`; (b) `rankOf` sentinel
(unknown statuses sort last); (c) per-entity sort columns
(`taskSortCols`/`epicSortCols`/`auditSortCols`); (d) chip-sync moved into
Update (View pure); (f) display-cell `padRight` for the date column; (g)
`sortArrow` per-column direction semantics. The container agent finished the
pass: updated the stale sort-chip test + `TestSortArrow`; (e) help/keys truth
(d/u pages lists, ctrl+d/u half-pages detail, find keys documented) plus a
scrollable help overlay (j/k scroll, anything else closes;
`TestModel_HelpScrollRevealsTail`); audits open-bucket scope note
(empty-state + help Notes row) per the recorded recommendation; design doc
updated to the ≥90 two-pane threshold (code wins); pagination reserve kept
deliberately (documented at the reserve site). Suite, vet, golangci-lint all
green.
