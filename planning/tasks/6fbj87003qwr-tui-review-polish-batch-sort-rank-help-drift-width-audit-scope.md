---
status: completed
epic: 18-tui-bubble-tea-interactive-planning-browser
description: Working-set sort ranks unknown statuses first, help/keys drift, byte-padded date column, unicode find-highlight misalign, audits open-bucket-only
effort: Unknown
tier: 3
priority: low
autonomy_level: 3
tags: [go, tui, polish]
created: "2026-06-12"
updated_at: "2026-06-12"
completed_at: "2026-06-12"
id: 6fbj87003qwr
---
# TUI review polish batch (sort rank, help drift, width, audit scope)

> 🔀 **Absorbed into
> [tui-s2b-polish-find-occurrences-highlight-fidelity-per-entity-sort](6fb7ym4023q0-tui-s2b-polish-find-occurrences-highlight-fidelity-per-entity-sort.md)
> (single pass by the macOS agent, 2026-06-12) — do not implement here.**
> Stays ready-to-start until the merged work lands; the macOS agent sets
> final status. Coordination notes + this agent's recommendations on the two
> debatable items are at the bottom.

> ⚠️ **Externally proposed — needs independent review before implementing.**
> Low-severity items from [2026-06-12-critical-code-review-multi-lens](../research/2026-06-12-critical-code-review-multi-lens.md)
> (plus still-open A4/A5 from
> [2026-06-11-critical-review-and-polish-research](../research/2026-06-11-critical-review-and-polish-research.md)). The implementing agent
> should decide which earn their change. Partially adjacent to
> [tui-s2b-polish-find-occurrences-highlight-fidelity-per-entity-sort](6fb7ym4023q0-tui-s2b-polish-find-occurrences-highlight-fidelity-per-entity-sort.md) —
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

- Epic [18-tui-bubble-tea-interactive-planning-browser](../epics/18-tui-bubble-tea-interactive-planning-browser.md)
- Touches `internal/tui/commands.go`, `help.go`, `keys.go`, `find.go`,
  `item.go`, `model.go`.
## Coordination with the merged macOS-agent pass (2026-06-12)

**Merged scope (a)–(g): agreed.** Mapping against this task: (a) covers
item 3 (the unicode-fold-safe matching MUST keep the
length-changing-fold case — `highlightOccurrences` computes offsets on
`ToLower(plain)` and slices `plain`, so U+0130-style runes can misalign or
slice past the end; a regression test with such a rune is the proof);
(b) covers item 1; (e) covers item 2; (f) covers item 4. (c)/(d)/(g) are
s2b-side items — no objection. Nothing missing from the merged list except
the two debatable items below, deliberately held out.

**⚠️ Base-tree warning — the most important note here.** Today's
*uncommitted* working tree contains substantial TUI changes the merged pass
must build on, not clobber: tab-routed list messages (`tabMsg`/`routeToTab` —
filter×reload correctness), list/detail load generations, per-tab load
errors, detail scroll preservation on same-item refresh, quit-key layering
(`DisableQuitKeybindings`, single-pane `q`-pops), and the S4 action-menu work
already started in `model.go`. All are pinned by tests in `model_test.go` —
run the TUI suite before and after the pass; those tests are the contract.

### Recommendation (i): audits open-bucket-only

**Scope note now; bucket-view axis as a follow-up task — not in this pass.**
Rationale: the cheap fix (help text + audits-tab empty-state/footer noting
"open bucket only — closed/deferred via `tskflwctl audit list --all`") removes
the silent confusion at near-zero risk. The real bucket-view axis wants the
same per-tab view machinery `statusView` uses, and it interacts directly with
S4 mutations (close/reopen/defer from the TUI should land you *in* a bucket
view) — designing it inside a polish pass would prejudge S4 decisions. File
the axis as its own small task alongside/after S4.

### Recommendation (ii): two-pane threshold + pagination reserve

**Code wins: ≥90; update the design doc.** 90–99 cols two-pane is already
useful (the "narrowest two-pane" border test runs at exactly 90), tests pin
90 in several places, and raising to 100 churns tests for no UX gain.

**Pagination reserve: reclaim only if it stays simple.** The reserve is
belt-and-suspenders for the chrome-cropping regression
(`TestModel_ChromeVisibleWhenListPaginates`). Reclaiming when
`Paginator.TotalPages <= 1` is a one-row win but couples list height to item
count (relayout needed on every SetItems — mild circularity). If the
implementation needs more than a few lines or any new relayout path, keep
the reserve and document it as deliberate; if attempted, the existing
`ViewFitsTerminal`/chrome invariant tests must gate it at multiple sizes
with both paginated and unpaginated item counts.

## Closure (2026-06-12)

The merged pass landed (see the closure note in
[tui-s2b-polish-find-occurrences-highlight-fidelity-per-entity-sort](6fb7ym4023q0-tui-s2b-polish-find-occurrences-highlight-fidelity-per-entity-sort.md)).
All four absorbed items shipped: A4 unknown-status rank (b), help/keys drift
(e), unicode find-highlight (a, with the U+0130 regression test), byte-padding
(f). Debatable items resolved per the recommendations recorded above: audits
got the scope note (bucket-view axis deferred to S4-adjacent follow-up);
threshold stays ≥90 with the design doc corrected; pagination reserve kept.
