---
schema: 1
status: deprecated
epic: 18-tui-bubble-tea-interactive-planning-browser
description: 'Lower-priority outcomes of the dashboard adversarial review: chrome-color sweep, deferred onDash/widget-registry decisions, and remaining test gaps (Tier 1+2 already shipped on feat/dashboard-in-tui).'
effort: M
tier: 3
priority: low
autonomy_level: 3
tags: [tui, review]
created: "2026-06-27"
updated_at: "2026-06-28"
deprecated_at: "2026-06-28"
---
# Dashboard review follow-ups (Tier 3 polish + deferred decisions)

Captures the lower-priority outcomes of the adversarial review of the TUI landing
dashboard (`feat/dashboard-in-tui`). The **Tier 1 bugs + Tier 2 cleanups + their
regression tests were applied on that branch**; this task holds what was
deliberately deferred so it isn't lost before the next polish pass.

## Already done on feat/dashboard-in-tui (for reference, not this task)

- Dashboard refreshes on `reloadAll` (no longer goes stale on `r`/fsnotify).
- Cursor-following scroll window (`scrollTo`) — selection can't be clipped off a
  short terminal.
- Durable dashboard load errors (`dashboard.loadErr`): error pane when never
  loaded, last-good rows + footer "refresh failed" on a failed refresh.
- `exitDashboard(i)` helper — the dashboard→tab transition is centralized
  (`switchTab` / `leaveDashTo` / `applyView` / `jumpTo`); `dashJump` is now pure
  routing; `Init()` no longer fires a wasted tab reload.
- `docs/ARCHITECTURE.md` notes why the dashboard is **not** an `entityTab`.
- Tests: mutation refresh, short-terminal cursor visibility, `+N more` overflow,
  durable error path, refresh-failure-keeps-rows, `move()` wrap, empty-nav safety,
  `scrollTo` unit.

## Tier 3 — polish (low value; do opportunistically)

- **Chrome colors bypass `theme`.** `dashHeading = lipgloss.Color("6")` is hardcoded
  — but so are `style.go` (`accent`), `help.go` (`helpHeading`), `action.go`
  (`actionHeading`). This is pre-existing debt, not new. Worth a *single* sweep to
  route UI-chrome colors through a `theme.ColorAccent`-style constant (likely an
  epic [[21-code-quality-architecture-hardening]] item), not a one-file patch.
- **`dashJump` silently no-ops on a missing tab** (`indexOfKind < 0` → `return nil`).
  Harmless today (only the three hardcoded entity kinds are targeted). Add a
  defensive flash once Projects/ADRs tabs land so a registry/tab mismatch is visible.
- **Health widget nav asymmetry.** misfiled/open-audits are navigable; "unreadable
  files (run lint)" is not — there's no lint view to jump to. Decide: leave as-is
  (informational), drop the row, or add an affordance signalling which rows are
  selectable. Note: the reviewer's "visual gap in cursor movement" claim was
  **refuted** — rows render contiguously; only the selectability differs.
- **`entityDashboard = -1` sentinel** — safe and already commented. Optional: define
  it after the `iota` block. Skip unless touching `entity.go` anyway.

## Deferred decisions (recorded so they aren't re-litigated)

- **`onDash bool` vs a screen enum — DEFERRED, keep the bool.** Reviewers
  downgraded this to clarity-not-correctness. The anticipated trigger (a second
  non-list landing surface) is ruled out: the planned Projects entity is a *tab*,
  so `onDash` won't multiply. Documented in ARCHITECTURE.md instead. Revisit only
  if a genuine second landing screen appears.
- **Hand-rolled widget building in `setSummary`** — fine at 4 widgets. Refactor to
  per-widget render funcs (or a small widget registry like `entityTab`) when adding
  the 5th–6th widget (pulse, then Projects-gated goals) or if the function passes
  ~250 lines.
- **`core.Summary` field coupling** — accepted. A data-driven widget/metadata
  approach was explicitly rejected (ARCHITECTURE: clarity over machinery for a few
  heterogeneous widgets). Adding `Pulse` will need a `setSummary` edit; that's
  normal, and mirrors the CLI's `SummaryHuman`.

## Remaining test gaps worth closing next iteration

- Explicit health-widget branch tests: misfiled → `(tasks, view=all)`, open audits
  → `audits`, unreadable → non-navigable info row, all-clear → only "✔ all clear".
- Cursor **preservation by `dashTarget` across refresh.** Today a refresh that
  shrinks the nav set resets the cursor to 0 (`setSummary` clamp). Pair a test with
  a small fix to restore the cursor by target (kind/id/view), like tabs restore by id.
- De-brittle `TestModel_DashboardEnterJumpsToItem`: assert `selectedTarget().id ==
  "alpha"` *before* pressing enter, rather than assuming `nav[0]` is the in-progress
  row (it fails loudly today, but the assumption should be explicit).

## Discard — do NOT re-raise

The review flagged `dashboard.go` `nav` closure (`t := tgt; …&t`) as a
"use-after-free of stack memory" / "all pointers alias the last `t`". **False
positive** — reasoned from a C/C++ model. In Go each closure call gets a fresh `t`
that escape-analysis heap-promotes; the code is correct and idiomatic. No change.

**Deprecated 2026-06-28 (pruned).** Low-value holding task. Its substantive dashboard-review outcomes (Tier 1 bugs + Tier 2 cleanups + regression tests) already shipped on feat/dashboard-in-tui, and the dashboard has since gained M2 epic-order agreement, liveness, ready-to-close, and the M1 column helpers. What remained was "do opportunistically" polish, recorded decisions (preserved in docs/ARCHITECTURE.md), and a couple of minor dashboard test-coverage gaps. Disposition: the chrome-colors-bypass-theme item is subsumed by color-and-design-overhaul; the remaining test gaps + recorded decisions are consciously dropped as not worth a standing task. Nothing actionable is lost.
