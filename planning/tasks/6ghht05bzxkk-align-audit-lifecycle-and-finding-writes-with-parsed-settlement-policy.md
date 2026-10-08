---
schema: 1
id: 6ghht05bzxkk
status: next-up
epic: 20-cli-ux-and-ergonomics
description: Audit closure blocks only literal open findings, unlike readiness; decide and align in-progress/invalid-status policy across lifecycle, finding writes, and lint.
effort: 2-4 hours
tier: 3
priority: medium
autonomy_level: 2
tags: [audit, core, safety]
created: "2026-10-07"
depends_on: [6gh82rm9sf3b]
updated_at: "2026-10-07"
---

# Align audit lifecycle and finding writes with parsed settlement policy

## Objective

Align the definition of parsed finding settlement across audit lifecycle and finding
writes. `Audit.Settled` excludes in-progress and invalid-status findings, but the
existing bucket gate and `LintFindings` check only literal `open`. A non-open audit
can also receive an in-progress finding status through current finding writes.
Decide the intended close/defer policy before tightening these related contracts;
preserve an explicit reopen path for recovery and the existing empty-audit exception.

## Acceptance criteria

- [ ] Decide whether close/defer require all parsed findings to be terminal,
  including in-progress, missing, and invalid statuses; document empty-audit and
  reopen exceptions.
- [ ] Apply the accepted parsed-status policy consistently to guarded bucket
  moves, finding writes against non-open audits, readiness, and lint without
  changing unparsed-header rules.
- [ ] Pin actual filesystem previews, writes, CAS retries, non-open finding
  transitions, and portable domain projections; classify any machine-contract
  changes.

## Out of scope

- Reclassifying headings or changing the approved unparsed-header refusal policy.
- TUI redesign, body-transform unification, or generic persistence extraction.

## Related

- Epic [20-cli-ux-and-ergonomics](../epics/20-cli-ux-and-ergonomics.md)
- Predecessor: [incomplete finding evidence](6gh82rm9sf3b-carry-unparsed-finding-evidence-into-audit-tui-and-lifecycle-decisions.md)
- Related but separate baseline guards: [audit body-write gaps](6g77rn6hvmh8-close-the-remaining-audit-body-write-guard-gaps.md)

## Evidence and sequencing (2026-10-07)

Discovered during the incomplete-evidence followup: `CountOpenFindings` counts only
the `open` token, while `TallyFindings` tracks `in-progress` separately and `Settled`
requires every parsed finding to be done/dropped. Audit bucket moves previously
reparsed the body for that literal-open count; the new shared domain move policy
retains it deliberately rather than silently changing parsed-status semantics.
`LintFindings` and `Service.EditFinding` must be reviewed together: the latter's
transform callback currently ignores audit metadata, unlike `NewFinding`'s explicit
open-bucket gate. A stricter move guard must not leave a second route to the same state.

Fresh-binary disposable-space reproduction confirmed both routes on 2026-10-07:
`audit close` accepted a canonical `H1` with `Status: in-progress`, and
`audit finding <closed-audit> H1 --status in-progress` changed an already closed,
fixed finding back to active. Both exited 0; `audit info --json` then reported
`bucket: closed` and `findings.in_progress: 1`. These are parsed-status policy gaps,
not incomplete-header classification failures.

Follow the incomplete-evidence task. This is independent audit hardening, not a
reason to reopen the completed adapter-neutral or tool-owned sub-entity Threads.
