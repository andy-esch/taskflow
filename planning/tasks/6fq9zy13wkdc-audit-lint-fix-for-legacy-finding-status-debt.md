---
schema: 1
id: 6fq9zy13wkdc
status: ready-to-start
epic: 20-cli-ux-and-ergonomics
description: No audit lint --fix to normalize legacy finding statuses (emoji/legacy words); schema audit omits the status vocabulary.
effort: Unknown
tier: 3
priority: medium
autonomy_level: 3
tags: [audit, lint]
created: "2026-07-18"
---
# audit lint --fix for legacy finding-status debt + document the vocabulary

## Objective

`audit lint` enforces a strict finding-status vocabulary
(deferred/fixed/in-progress/landed/open/superseded/wontfix) but there is no
`audit lint --fix` to normalize older audits (emoji ✅/⏳/⛔, legacy words like
`tracked`/`declined`, or pre-`Status:` findings). Separately, `schema audit`
documents the Status-line *format* but not the allowed *vocabulary* (only the
top-level `schema` lists it).

## Acceptance criteria

- [ ] `audit lint --fix` normalizes finding statuses (strip emoji, map declined→wontfix / tracked→superseded, backfill missing)
- [ ] `schema audit` lists the finding-status vocabulary

## Notes

- Was masked in the wild by the P2 abort — whole-tree `audit lint` never
  completed until the invalid-id files were fixed, hiding ~10 audits of debt.
- Confirmed: no `--fix` on `audit lint`; `schema audit` omits the vocab.
- Source: https://github.com/andy-esch/taskflow/issues/105 (P3, Medium)
