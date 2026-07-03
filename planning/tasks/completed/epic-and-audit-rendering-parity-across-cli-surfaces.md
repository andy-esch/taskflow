---
schema: 1
status: completed
epic: 20-cli-ux-and-ergonomics
description: Bars in epic list/show and audits in the status dashboard, so every progress view reads from one symbology (follow-up to the 2026-06-25 audit-symbology pass).
effort: S
tier: 3
priority: medium
autonomy_level: 3
tags: [cli, render]
created: "2026-06-25"
updated_at: "2026-06-25"
started_at: "2026-06-25"
completed_at: "2026-06-25"
id: 6ffr4wc00bfr
---
# Epic↔audit rendering parity across CLI surfaces

## Objective

The 2026-06-25 audit-symbology pass gave audits the epic-style treatment
(bucket glyph + rollup bar + resolved/total) in `audit list`/`show` and the TUI.
That exposed three cross-entity inconsistencies worth closing so the whole tool
reads from one symbology:

1. **`status` dashboard omits audits.** `SummaryHuman` (render.go) renders only
   Tasks + Epics; audits never appear, even though they now have a `Percent()`
   and a bar. They should slot in next to epics (bar + % + resolved/total).
2. **CLI `epic list` has no bar.** `EpicsHuman` shows `3/4 (75%)` while the TUI
   epic rows and the dashboard both draw a bar. Add the bar for parity.
3. **CLI `epic show` has no bar** while `audit show` now does (the asymmetry this
   pass introduced). Either add a progress line to `epic show` or accept the
   difference — decide and make it deliberate.

## Context

Follows the audit-symbology work (theme.Bucket → Token, theme.FindingStatus,
domain.Audit.Resolved()/Percent(), progressbar reuse). All the machinery exists:
`render.Style.Bar`, `st.Percent`, `progressbar.Render`, and the epic rollup
(`core.EpicSummary.Percent`). This is a consistency sweep, not new infra.
Relates to epic 20 (CLI UX); the dashboard + epic-list/show are all CLI render.

## Acceptance criteria

- [ ] `status` dashboard lists audits with the same bar + %/counts treatment as
      epics (open audits at least; decide whether closed/deferred show).
- [ ] `epic list` (CLI) shows a rollup bar consistent with audits and the TUI.
- [ ] `epic show` (CLI) progress rendering is a deliberate, documented choice
      (bar added, or the asymmetry justified in a comment).
- [ ] No `--json` envelope changes (or additive only); porcelain stays
      byte-stable on the no-color path.
- [ ] go build ./... + go test ./... + golangci-lint run ./... green; golden
      snapshots / unit assertions updated; docs/README refreshed if output shown.

## Implementation sketch

- Dashboard: add the audit rollup to `core.Summary` and a block in `SummaryHuman`
  mirroring the Epics block (render.go ~229).
- `EpicsHuman`: swap `%d/%d (%s)` for `st.Bar(pct, …) + st.Percent + done/total`.
- Keep the bar width/format aligned with the audit list so columns read the same.

## Risks / gotchas

- The `status` summary has a JSON envelope (`SummaryJSON`) — if audits join the
  human view, decide whether they also join the envelope (additive, schema-safe)
  or stay human-only.
- Don't double-encode state: epics already show a STATUS column; a bar is progress,
  not status — keep both meaningful, as the audit list does with its bucket glyph.

## Done when

`status`, `epic list`, `epic show`, `audit list`, and `audit show` all render
progress from one visual vocabulary — build/test/lint green, snapshots + docs
updated.
