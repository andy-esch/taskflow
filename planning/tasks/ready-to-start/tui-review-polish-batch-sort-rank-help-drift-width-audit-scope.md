---
status: ready-to-start
epic: 18-tui-bubble-tea-interactive-planning-browser
description: Working-set sort ranks unknown statuses first, help/keys drift, byte-padded date column, unicode find-highlight misalign, audits open-bucket-only
effort: Unknown
tier: 3
priority: low
autonomy_level: 3
tags: [go, tui, polish]
created: "2026-06-12"
---
# TUI review polish batch (sort rank, help drift, width, audit scope)

> ⚠️ **Externally proposed — needs independent review before implementing.**
> Low-severity items from [[2026-06-12-critical-code-review-multi-lens]]
> (plus still-open A4/A5 from
> [[2026-06-11-critical-review-and-polish-research]]). The implementing agent
> should decide which earn their change. Partially adjacent to
> [[tui-s2b-polish-find-occurrences-highlight-fidelity-per-entity-sort]] —
> coordinate, do not duplicate.

## Objective

1. **A4 — `sortWorkingSet` ranks unknown statuses as in-progress.**
   `statusRank` map-miss yields rank 0 (`internal/tui/commands.go:116-129`);
   give unknowns a sentinel last rank.
2. **Help/keys drift.** `help.go:36` (and the `keys.go:6` header) claim
   `ctrl+d / u` half-page in lists — not bound there in bubbles/list v1
   (`u/d` are full-page; `h/l` are shadowed by focus keys). The Detail help
   section omits `/`, `n/N` which `model.go:258-264` implements.
3. **Unicode find-highlight misalignment.** `highlightOccurrences`
   (`find.go:43-55`) computes offsets on `strings.ToLower(plain)` but slices
   `plain`; runes whose lowercase changes byte length (e.g. U+0130) misalign
   or can index past the end. Fold without changing lengths.
4. **Date column misaligns for non-ASCII slugs.** `item.go:68` pads by bytes
   (`%-*s`) while `truncate` budgets display cells; pad with
   `slugW - lipgloss.Width(slug)` spaces.
5. **Audits are open-bucket-only with no in-TUI escape.**
   `commands.go:90-92` hardcodes `ListAudits("", false)`; add a bucket-view
   axis or an explicit scope note in help.
6. **A5 residue.** Two-pane threshold `>= 90` (`model.go:471`) vs design doc
   ≥100 — pick one, fix the doc; reclaim the unconditionally reserved
   pagination row (`model.go:469`) when `TotalPages <= 1`.

## Acceptance criteria

- [ ] Foreign-status tasks sort last in the working set (test).
- [ ] Help text matches actually-bound keys in both panes.
- [ ] Find-highlight survives a length-changing-fold rune without panic
      (test with U+0130).
- [ ] Accepted items have tests; suite + lint green.

## Related

- Epic [[18-tui-bubble-tea-interactive-planning-browser]]
- Touches `internal/tui/commands.go`, `help.go`, `keys.go`, `find.go`,
  `item.go`, `model.go`.