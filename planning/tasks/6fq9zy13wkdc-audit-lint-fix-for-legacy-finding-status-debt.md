---
schema: 1
id: 6fq9zy13wkdc
status: next-up
epic: 20-cli-ux-and-ergonomics
description: No audit lint --fix to normalize legacy finding statuses (emoji/legacy words); schema audit omits the status vocabulary.
effort: Unknown
tier: 3
priority: medium
autonomy_level: 3
tags: [audit, lint]
created: "2026-07-18"
updated_at: "2026-08-24"
---
## Objective

`audit lint` enforces a strict finding-status vocabulary, and `audit finding` now
writes it, but there is still no `audit lint --fix` to repair audits written before
either existed. A file with a malformed status is flagged forever and repaired by hand,
which is the practice this whole surface has been retiring.

**This task was rewritten on 2026-08-24.** Its original scope was authored against a
vocabulary that has since changed, and two of its three premises are now wrong:

- It called `tracked` a legacy word to be mapped to `superseded`. `tracked` is now a
  first-class status meaning "handed to a task", and that mapping would DESTROY the
  handoff — the opposite of a repair.
- It listed `landed` as legal. `landed` was dropped (M3 of
  `2026-08-17-finding-status-surface`) after the corpus showed zero uses of it.
- Emoji-stripping was its largest item. M2 of the same audit made the parser
  decoration-tolerant, so `**Status:** ✅ fixed` now reads correctly and lints clean.
  A `--fix` that rewrites those lines is now cosmetic rather than corrective.

What is genuinely left is narrower, and worth checking against a real corpus before
building: there may be very little actual debt remaining.

## Acceptance criteria

- [x] `schema audit` names the finding-status vocabulary.  **met:** the conventions line is
      built from `FindingStatuses()` rather than transcribed, so it cannot fall behind.
- [ ] The remaining debt is MEASURED before it is repaired — how many audits in a real
      corpus still fail `audit lint`, and for what. If the answer is "none", this task
      closes as `wontfix` rather than growing a repair nobody needs.
- [ ] `audit lint --fix` repairs what that measurement actually finds, honouring
      `--dry-run` and reporting per-file changes the way `lint --fix` already does.
- [ ] No repair silently changes a status's MEANING. A backfill of a missing status is
      safe; a mapping between two legal words is not, and must be refused rather than
      guessed — the `tracked → superseded` rule this task used to carry is the cautionary
      example.
- [ ] Errors wrap the domain sentinels; suite + lint green; docs updated.

## Notes

- Was masked in the wild by the P2 abort — whole-tree `audit lint` never completed until
  the invalid-id files were fixed, hiding ~10 audits of debt. That abort is fixed, so the
  measurement above is now possible where it was not before.
- H2 of `2026-08-17-finding-status-surface` is tracked here.
- Source: https://github.com/andy-esch/taskflow/issues/105 (P3, Medium)
