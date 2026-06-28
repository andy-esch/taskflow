---
schema: 1
status: deferred
epic: 21-code-quality-architecture-hardening
description: Make FindingsRollup a composed Service.FindingsRollup() view-model that Summary composes, so a web findings page can roll up with its own filter/sort/pagination (audit L1, deferred pending web).
effort: S
tier: 3
priority: low
autonomy_level: 3
tags: [architecture, core, web]
created: "2026-06-28"
updated_at: "2026-06-28"
deferred_at: "2026-06-28"
---
## Why

From the 2026-06-27 consumer-data-flow audit, finding L1. `FindingsRollup` is a presentation-shaped aggregate (display-driven ordering: `ByUrgency` in triage order, `ByComponent` most-first, a hand-picked `Acute` call-out) living as a FIXED field on `core.Summary`. Defensible today (the tally is worth more than purity, and triage order is arguably domain logic), but `Summary` risks becoming "whatever the current dashboards need," and a web findings page wanting pagination or another sort would either reuse this fixed shape or call the composable `QueryFindings` and re-roll.

## Goal

Make it a `Service.FindingsRollup()` view-model that `Summary` *composes*, so a web findings page can roll up with its OWN filter/sort/pagination instead of the fixed dashboard shape. Keep the existing `Summary.Findings` field working (the CLI/TUI dashboards consume it unchanged).

## Status: DEFERRED (web-gated)

This only earns its keep once the web adapter (epic 19, `tskflwctl serve`) exists and wants its own findings view. Re-homed here 2026-06-28 because its original tracking task (`let-core-own-the-dashboard-aggregates-adapters-re-derive`, the M2/M3/M9 aggregates cluster) completed. Sibling of the other deferred web-prep findings: H6 (`thread-context...`) and H7 (`reusable-workspace...`). Revisit when `serve` is scoped.