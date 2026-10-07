---
schema: 1
id: 6gh82rm9sf3b
status: ready-to-start
epic: 20-cli-ux-and-ergonomics
description: Carry incomplete finding evidence through TUI metadata and define closure/deferral policy rather than silently ignoring unparsed headings.
effort: 2-4 hours
tier: 3
priority: medium
autonomy_level: 3
tags: [audit, tui, safety]
created: "2026-10-06"
depends_on: [6g77rn6em6n8]
updated_at: "2026-10-07"
---
# Carry unparsed finding evidence into audit TUI and lifecycle decisions

## Objective

Close the remaining consumers of incomplete finding evidence after the CLI read fix.
The shared domain count exists, but TUI metadata still renders 0/0 for a wholly
unparsed audit. Store MoveAudit also gates closure/deferral only on parsed open
findings, so it can close an audit whose read model deliberately withholds readiness.
The compact audit info receipt is another remaining consumer: its human and JSON
finding tally omits unparsed evidence. Include it without changing parsed counts.

## Acceptance criteria

- [ ] TUI audit list/detail distinguish empty, unparsed-only, and mixed audits and offer the lint diagnostic route without local path access.
- [ ] Define a consistent closure/deferral policy for repairable versus ambiguous headers; any refusal is actionable and runs before dry-run or writes.
- [ ] Tests pin portable TUI evidence and actual filesystem lifecycle behavior, including wholly unparsed and settled-but-incomplete audits.
- [ ] Existing empty-audit closure and normal settled audits remain supported; docs, race tests, lint, and planning lint pass.
- [ ] Compact audit info human/JSON receipts carry incomplete-read evidence
  without adding suspected headings to parsed counts; optional fields and
  projections follow the machine-contract revision policy.

## Out of scope

- Changing the canonical parser or automatically rewriting ambiguous prose.
- Audit creation/edit baseline gaps, owned by 6g77rn6hvmh8.
- General TUI redesign or body-transform extraction.

## Sequencing and evidence

Follow 6g77rn6em6n8. This is an independent audit-hardening followup in epic 20,
not a reason to reopen either completed architectural or sub-entity Thread.
As observed on 2026-10-06: internal/tui/detail.go renderAuditMeta uses the parsed
denominator without an unparsed warning; internal/store/auditstore.go MoveAudit
uses CountOpenFindings(ParseFindings(body)). Settle whether ambiguous prose needs
an explicit override before tightening lifecycle policy.
Codex's 2026-10-06 safety review independently verified audit info's omitted warning
at internal/wire/dto.go (AuditInfoJSON/ToAuditInfo) and internal/cli/render/render.go
(AuditInfo). This is an existing adjacent consumer gap, not a new classifier defect.

## Related

- [CLI incomplete-read evidence](6g77rn6em6n8-report-unparsed-findings-on-the-audit-read-surfaces.md)
- [Shared classifier](6g77rn6b9wf8-narrow-the-near-miss-finding-recognizer-and-re-measure-across-every-entity-type.md)
- [Remaining body-write gaps](6g77rn6hvmh8-close-the-remaining-audit-body-write-guard-gaps.md)
