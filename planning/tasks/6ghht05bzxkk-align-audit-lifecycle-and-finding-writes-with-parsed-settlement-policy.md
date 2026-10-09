---
schema: 1
id: 6ghht05bzxkk
status: completed
epic: 20-cli-ux-and-ergonomics
description: Require terminal parsed findings for audit close/defer and non-open finding status writes, aligned with readiness and lint.
effort: 2-4 hours
tier: 3
priority: medium
autonomy_level: 2
tags: [audit, core, safety]
created: "2026-10-07"
depends_on: [6gh82rm9sf3b]
updated_at: "2026-10-09"
started_at: "2026-10-09"
completed_at: "2026-10-09"
---

# Align audit lifecycle and finding writes with parsed settlement policy

## Objective

Align the definition of parsed finding settlement across audit lifecycle and finding
writes. `Audit.Settled` excludes in-progress and invalid-status findings, but the
existing bucket gate and `LintFindings` check only literal `open`. A non-open audit
can also receive an in-progress finding status through current finding writes.
Require terminal parsed statuses for close/defer, preserving an explicit reopen
path for recovery and the existing empty-audit exception.

## Acceptance criteria

- [x] Decide whether close/defer require all parsed findings to be terminal,
  including in-progress, missing, and invalid statuses; document empty-audit and
  reopen exceptions.
- [x] Apply the accepted parsed-status policy consistently to guarded bucket
  moves, finding writes against non-open audits, readiness, and lint without
  changing unparsed-header rules.
- [x] Pin actual filesystem previews, writes, CAS retries, non-open finding
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

## Approved policy (2026-10-09)

Close/defer require complete parsing and every parsed finding to have a terminal
status: `fixed`, `tracked`, `deferred`, `superseded`, or `wontfix`. `open`,
`in-progress`, missing, and invalid statuses block, including same-bucket calls
and previews. No force bypass. Empty audits may close/defer, but do not advertise
ready-to-close; reopening remains available for any readable audit.

Finding status writes in a closed/deferred audit may correct a finding to a
terminal status but cannot reactivate it, even as a no-op or preview. Note and
candidate-only edits remain possible without certifying unrelated pre-existing
defects. Reopen first to resume work. Both lifecycle and finding-write decisions
must use the exact audit metadata/body protected by the write guard and be
recomputed on conflict retries. Lint reports non-open audits with any unsettled
parsed findings. Terminal-token settlement does not replace independent lint for
tracked destinations, resolution notes, or managed candidate rows.

## Implementation and evidence (2026-10-09)

- Domain `UnsettledFindings` uses the same terminal tally as readiness;
  `ValidateMove` refuses with a structured `AuditUnsettledFindingsError`. The
  guarded store retains source identity in `AuditMoveError`, with executable
  `audit findings <source-ID>` and `audit lint <source-ID>` advice. The findings
  route matters: valid in-progress work in an open audit need not fail lint.
- `EditFinding` uses the CAS-protected callback's bucket and recomputes the edit
  on retries. It inspects the normalized parsed status, not decorated input text.
  Terminal corrections and metadata-only edits tolerate unrelated existing
  defects; bucket moves still refuse residual unsettled evidence.
- `LintFindings` reports the non-open bucket invariant for all unsettled parsed
  statuses, alongside missing/invalid-status diagnostics. Open/active counts and
  parsed denominators remain unchanged. CLI/TUI portable projections and real
  store/CLI paths are covered, including same-bucket previews, compound-edit
  refusal without partial status/note/candidate writes, and both CAS races.
- Machine revision **1.84, NOT ADDITIVE** classifies newly refused lifecycle and
  finding writes. Generated schemas, comments, CLI references, and machine goldens
  are updated. Architecture and contributor guidance document the policy.
- Full `just test` (race), `just lint` (zero issues), `just build`, `just tidy-check`,
  planning/link lint, and audit lint passed. A separate disposable docgen output
  exactly matched `docs/cli`; `just docs-check`'s clean-git assertion is deferred
  until commits. The build's module stat-cache permission warning did not prevent
  compilation. Disposable CLI smoke covered active refusal, terminal closure,
  closed-audit reactivation refusal, explicit reopening, and empty closure.
- Compiler-valid mutations reverting the move guard to literal-open counts and
  disabling explicit-status validation each failed their intended real-store
  regression tests; both were restored and rerun green before the final race pass.

The arbitrary-body authoring gap is recorded and sequenced in
[the existing body-write-guard task](6g77rn6hvmh8-close-the-remaining-audit-body-write-guard-gaps.md),
not silently treated as covered here. Implementation ACs are met and independent
review is reconciled; the task is completed for the PR handoff, with merge and
release inclusion remaining separate delivery milestones.

Independent reviews and owner triage:
[Codex](../audits/6gj5ct38xq4t-2026-10-09-audit-parsed-settlement-implementation-codex.md)
and [Antigravity](../audits/6gj5ct3hyzjs-2026-10-09-audit-parsed-settlement-implementation-antigravity.md).

## Review follow-through (2026-10-09)

Codex M1 exposed a pre-existing but in-scope structured-write escape: hard-wrapping
a single-line resolution note could manufacture a new active finding in a closed
or deferred audit. Generated continuations now remain prose, including heading,
label, and fence examples; no global clean-audit requirement blocks unrelated
repairs. Permanent real-store/CLI regressions cover the reported preview/write
matrix and damaged neighboring findings, with structural round-trip and no-effect
assertions. Reverting the fix compiled and failed the intended regressions.

The portable status oracle no longer calls the production terminal helper. New
wire tests validate nondefault readiness against exact emitted-envelope schemas;
both now kill an in-progress-to-done mapping mutation. Antigravity reported no
findings; its useful status/lint evidence is retained with explicit qualifications
for unsupported readiness/schema claims and the missed note path.

Full race tests, lint, build, tidy, generated CLI docs freshness, and planning/link
checks passed after these changes. Both audits are closed with no unresolved
findings; Codex M1 is marked fixed using the finding lifecycle verb. A fresh
binary reproduced M1's input safely in a disposable unregistered space. The
pre-existing finding-edit preview-body/slug-reread receipt gap is added to the
existing [body-write guard followup](6g77rn6hvmh8-close-the-remaining-audit-body-write-guard-gaps.md),
not treated as settled here. The current task is completed for delivery; its
successor is now eligible but still needs a scope pass before implementation.
