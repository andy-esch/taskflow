---
schema: 1
id: 6g77rn6em6n8
status: next-up
epic: 20-cli-ux-and-ergonomics
description: audit show and audit list say 'no findings' for an audit whose headings all failed to parse, which is confidently false on the cheapest triage path.
effort: 2-3 hours
tier: 3
priority: high
autonomy_level: 3
tags: [audit, cli, ux]
created: "2026-09-05"
updated_at: "2026-10-06"
---

# Report unparsed findings on the audit read surfaces

## Objective

<why / what — one short paragraph>

## Acceptance criteria

- [ ] <observable outcome>

## Out of scope

- <explicitly excluded>

## Related

- Epic [20-cli-ux-and-ergonomics](../epics/20-cli-ux-and-ergonomics.md)

Cross-referenced by audit 2026-10-06-adapter-hygiene: M1 (partial overlap — that finding stays open). M1 supplies the mechanism behind this task's symptom and a verified reproduction: the store's `parseAuditWithFindings` already computes `NearMisses` on every audit read, but `core.AuditWithBody` carries only `{Audit, Body}`, so `ShowAudit` discards them and both `cli/audit.go:490` and `tui/detail.go:1720` re-run the narrow `domain.ParseFindings` instead. Reproduced in a throwaway repo: with a `#### M-2.` heading in the body, `audit lint` reports "its work is invisible" while `audit show` says nothing on either surface, and `audit show --json` publishes no structured findings at all. M1 goes beyond this task in also proposing the de-duplication of the parse, which is why it is not marked tracked here.
