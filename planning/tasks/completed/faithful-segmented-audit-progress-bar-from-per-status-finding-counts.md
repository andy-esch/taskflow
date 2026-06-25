---
schema: 1
status: completed
epic: 20-cli-ux-and-ergonomics
description: Replace resolved = total − open (which counts deferred/dropped findings as done) with a per-status finding rollup so the audit bar tells the truth.
effort: M
tier: 3
priority: low
autonomy_level: 3
tags: [audit, render]
created: "2026-06-25"
updated_at: "2026-06-25"
started_at: "2026-06-25"
completed_at: "2026-06-25"
---
# Faithful segmented audit progress bar from per-status finding counts

## Objective

The audit rollup bar (added 2026-06-25) defines "resolved" as
`Findings − OpenFindings`, where only `status: open` counts as open. So every
other finding state — in-progress, fixed, landed, deferred, superseded, wontfix —
counts as "resolved," which means a *deferred* finding inflates the bar and a
fully-deferred audit reads 100%. Make the bar tell the truth: derive progress
from the real finding-status groups instead of a single open/closed split.

## Context

`domain.Audit` carries only `Findings` + `OpenFindings` (audit.go); the per-status
breakdown is thrown away after `CountOpenFindings`. `domain.ParseFindings` already
yields every finding's status, so the counts exist at parse time — they just need
plumbing onto the Audit struct (or a sibling rollup type) the way `EpicSummary`
carries Done/Total/Deprecated. `theme.FindingStatus` already assigns each status a
colour, so a segmented bar has a palette ready. Relates to epic 20 (output/UX) and
touches the audit data model + store load path lightly.

## Acceptance criteria

- [ ] Decide the rollup groups and document them, e.g. done (fixed/landed) vs
      in-progress vs open vs dropped (deferred/superseded/wontfix) — and what each
      contributes to the bar (full / partial / empty).
- [ ] `domain.Audit` (or a rollup struct) carries the per-status counts, sourced
      once from `ParseFindings`; `Resolved()`/`Percent()` redefined against it.
- [ ] The bar renders the segments (or at minimum stops counting parked/dropped
      findings as done); CLI + TUI stay in lockstep via the shared progressbar.
- [ ] `--json` audit envelopes optionally expose the breakdown (additive).
- [ ] go build/test/lint green; new counts unit-tested; snapshots/docs updated.

## Implementation sketch

- Add a `FindingCounts` (by status group) to the audit load path in
  `store/auditstore.go`, where `Findings`/`OpenFindings` are already set from
  `ParseFindings`.
- Either extend `progressbar.Render` to take multiple segment widths, or render a
  single bar against a refined "done" definition first (cheaper) and segment later.
- Keep the loose open-count meaning available for back-compat (`audit list -c open`
  still works).

## Risks / gotchas

- Multi-segment bars are visually noisy at width 8 (TUI rows) — may need a wider
  bar, or a single-fill refined definition there with segments only in detail/show.
- "Done" semantics are a judgment call (does wontfix count as resolved? it's not
  open, but it's not fixed) — settle it explicitly; it changes every percentage.
- Don't break the `bucket↔state` lint invariant (a non-open audit has no open
  findings) — that rule is about `open`, independent of this richer rollup.

## Done when

An audit's bar reflects genuinely-resolved work — a deferred-heavy audit no longer
reads as "done" — with the status groups documented, counts tested, and CLI/TUI in
sync.
