---
schema: 1
id: 6gh82rm9sf3b
status: completed
epic: 20-cli-ux-and-ergonomics
description: Carry incomplete finding evidence through TUI metadata and define closure/deferral policy rather than silently ignoring unparsed headings.
effort: 2-4 hours
tier: 3
priority: medium
autonomy_level: 3
tags: [audit, tui, safety]
created: "2026-10-06"
depends_on: [6g77rn6em6n8]
updated_at: "2026-10-08"
started_at: "2026-10-07"
completed_at: "2026-10-08"
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

- [x] TUI audit list/detail distinguish empty, unparsed-only, and mixed audits and offer the lint diagnostic route without local path access.
- [x] Define a consistent closure/deferral policy for repairable versus ambiguous headers; any refusal is actionable and runs before dry-run or writes.
- [x] Tests pin portable TUI evidence and actual filesystem lifecycle behavior, including wholly unparsed and settled-but-incomplete audits.
- [x] Existing empty-audit closure and normal settled audits remain supported; docs, race tests, lint, and planning lint pass.
- [x] Compact audit info human/JSON receipts carry incomplete-read evidence
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
uses CountOpenFindings(ParseFindings(body)). The ambiguity/override policy is now
settled below rather than left as an implementation-time choice.
Codex's 2026-10-06 safety review independently verified audit info's omitted warning
at internal/wire/dto.go (AuditInfoJSON/ToAuditInfoJSON) and internal/cli/render/render.go
(AuditInfoHuman). This is an existing adjacent consumer gap, not a new classifier defect.

## Approved policy and implementation scope (2026-10-07)

Close and defer refuse both evidence-backed repairable headers and diagnostic-only
ambiguity until they are repaired or clarified. No new force bypass. Reopen remains
available for recovery; existing genuinely empty and settled audit moves remain
supported. Validate the counts from the exact source protected by the persistence
guard, including previews and same-bucket calls, not a separate UI preflight read.
Domain policy is shared through the audit persistence port; adapters do not invent
their own interpretation of unparsed evidence.

Audit info exposes optional top-level `unparsed_findings`, separate from the unchanged
parsed disposition tally; zero is omitted. Revision 1.83 is NOT ADDITIVE because
formerly accepted close/defer calls now refuse, even though the receipt field alone
is additive. Existing list projections keep their vocabulary and defaults.

The pre-existing difference between readiness and literal-open-only lifecycle
gating for *parsed* statuses is not silently broadened here. It is tracked in
[parsed settlement policy](6ghht05bzxkk-align-audit-lifecycle-and-finding-writes-with-parsed-settlement-policy.md).

## Implementation and validation (2026-10-07)

- Domain `Audit.ValidateMove` owns the policy. FS calls it on freshly parsed,
  write-guarded content before same-bucket and preview returns; service conflict
  retries reload the count. No extra preflight read or persistence API was added.
- TUI rows show a compact warning before the label; detail supplies the audit-lint
  route. Empty, unparsed-only, mixed, and settled portable reads are distinguished
  at 80/120/240 columns, retaining source identity and no inferred local path.
  Audit info's human and JSON receipts carry the same loaded evidence.
- Full `just test` (race), `just lint`, `just build`, and `just tidy-check` passed.
  Regenerated CLI docs match disposable generator output; machine goldens and
  schema validation cover revision 1.83 and the nonzero/omitted info field.
  `lint --links` and `audit lint` passed, including the new planning files.
- Real-binary disposable-space smoke confirmed exit 11 and diagnostic advice for
  incomplete close/defer, successful empty/settled moves, and repairability:
  `lint --fix` makes the evidenced malformed header canonical and closeable while
  leaving ambiguous text unchanged and blocked. The lint command still reports
  the residual ambiguity, rather than implying the repair solved everything.
- Three compiler-valid mutation checks were killed by their intended regressions:
  bypassing the store move policy, zeroing the info mapper's count, and disabling
  the TUI detail warning branch. All probes were restored and targeted tests passed.
  Filesystem tests also pin bytes/modes/tree preservation and a conflict retry
  after concurrent ambiguity is introduced.

All implementation criteria are met. Both independent reviews have been processed,
their audits closed, and the task completed for PR delivery after owner verification:

- [Codex lifecycle/persistence review](../audits/6ghhwe5kzj01-2026-10-07-audit-incomplete-evidence-implementation-codex.md)
- [Antigravity presentation/wire review](../audits/6ghhwe5wnb7n-2026-10-07-audit-incomplete-evidence-implementation-antigravity.md)

## Review outcomes (2026-10-08)

Codex M1 was reproduced and fixed: a display slug can shadow a sibling's canonical
ID or be duplicated, making apparently actionable lint advice inspect the wrong
audit. Domain policy now returns `AuditIncompleteEvidenceError`; the guarded
adapter supplies its exact `RecordSource` in `core.AuditMoveError`. Human info,
list/show/status, and TUI detail keep loaded source identity separate from display
metadata. No second read or frontmatter-ID fallback was added to construct advice.

`TestAuditIncompleteEvidenceAdviceRetainsSourceIdentity` executes the emitted lint
route across duplicate slugs, ID-shadowing slugs, and declared-ID drift, including
both close/defer preview/apply refusals. Portable CLI/TUI fixtures also disagree
on declared/source ID. A fresh binary reproduced correct advice in a disposable
ID-shadowing space; lint selected the original ambiguous audit and audit hashes
remained unchanged. The same hazard in the pre-existing whole-file edit receipt
is recorded with the [body-write gaps task](6g77rn6hvmh8-close-the-remaining-audit-body-write-guard-gaps.md).

Antigravity reported no findings. Its info-mapper mutation instead exercised the
older list/show mapper; that is useful coverage but not the requested new-branch
proof or independent corroboration of Codex's identity result. Owner re-executed
the exact `ToAuditInfoJSON` count-loss mutation and a slug-substitution mutation
in `AuditMoveError`: both compiled and were killed by the intended regressions.
Both probes were restored and targeted race tests passed. The full race suite,
lint, build, module tidiness, generated-document comparison, and planning lint
are the final gates for the corrected implementation.

## Related

- [CLI incomplete-read evidence](6g77rn6em6n8-report-unparsed-findings-on-the-audit-read-surfaces.md)
- [Shared classifier](6g77rn6b9wf8-narrow-the-near-miss-finding-recognizer-and-re-measure-across-every-entity-type.md)
- [Remaining body-write gaps](6g77rn6hvmh8-close-the-remaining-audit-body-write-guard-gaps.md)
