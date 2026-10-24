---
schema: 1
id: 6gppr2k7xk5s
status: ready-to-start
epic: 21-code-quality-architecture-hardening
description: ShowAudit drops the findings and near-misses its read parsed, so CLI and TUI re-parse the body and audit show (human and --json) cannot report invisible findings.
effort: half day
tier: 2
priority: medium
autonomy_level: 3
tags: [audit, cli, tui, contracts]
created: "2026-10-23"
audit_sources: [planning/audits/6gh1frgyz2fh-2026-10-06-adapter-hygiene.md]
updated_at: "2026-10-23"
---
# Carry parsed audit findings and near-misses through ShowAudit

## Objective

The store parses findings, tallies, near-misses, and candidate issues on every audit read, but
`core.AuditWithBody` carries only `{Audit, Body}`. Both primary adapters re-run
`domain.ParseFindings`, and neither can surface a near-miss that `audit lint` calls invisible work.
Carry the parse through the show use case and consume it.

## Acceptance criteria

- [ ] `ShowAudit` returns the already-computed findings, near-misses, and candidate issues, and the CLI and TUI no longer call `domain.ParseFindings` on the shown body.
- [ ] Human `audit show` reports a near-miss count when one exists, matching what `audit lint` reports for the same body.
- [ ] `audit show --json` publishes structured findings and near-miss evidence as a classified additive revision.
- [ ] A regression builds an audit with a canonical finding plus a near-miss heading and asserts the three surfaces agree.

## Out of scope

- Near-miss columns on `audit list`, and the write-guard receipts owned by [6g77rn6hvmh8](6g77rn6hvmh8-close-the-remaining-audit-body-write-guard-gaps.md).

## Related

- Epic [21-code-quality-architecture-hardening](../epics/21-code-quality-architecture-hardening.md)
- Audit [2026-10-06 adapter hygiene](../audits/6gh1frgyz2fh-2026-10-06-adapter-hygiene.md), finding M1
- Future consumer: [TUI audit finding dispositions](6gmshzxvgagj-manage-audit-finding-dispositions-from-the-tui.md)
