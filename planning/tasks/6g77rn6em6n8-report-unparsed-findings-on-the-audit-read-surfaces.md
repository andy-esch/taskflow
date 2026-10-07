---
schema: 1
id: 6g77rn6em6n8
status: completed
epic: 20-cli-ux-and-ergonomics
description: audit show and audit list say 'no findings' for an audit whose headings all failed to parse, which is confidently false on the cheapest triage path.
effort: 2-3 hours
tier: 3
priority: high
autonomy_level: 3
tags: [audit, cli, ux]
created: "2026-09-05"
updated_at: "2026-10-07"
started_at: "2026-10-06"
completed_at: "2026-10-07"
---

# Report unparsed findings on the audit read surfaces

## Objective

Distinguish an intentionally empty audit from one whose code-shaped finding headings
could not be parsed. Publish the shared classifier's incomplete-read evidence on audit
list/show, including JSON and projected lists, without treating suspected findings as
real findings or inventing statuses. Keep ordinary reads informational; normal lint
remains the validation surface.

## Acceptance criteria

- [x] Human audit list and show distinguish empty, wholly unparsed, mixed, and
  settled-but-incomplete audits and name audit lint as the diagnostic route.
- [x] Full JSON and the optional unparsed projected-list column expose the
  derived count without inventing findings or changing parsed denominators.
- [x] Portable and filesystem-backed tests prove counts and readiness agree;
  ordinary headings and fenced examples cause no warnings.
- [x] Schema revision/classification, generated command docs, focused tests,
  full race suite, lint, build, and planning/audit lint pass.

## Out of scope

- Changing the canonical finding grammar, guessing statuses, or counting suspected
  headings as parsed findings.
- Rewriting ambiguous prose; high-confidence repair belongs to the paired recognizer task.
- A general structured-diagnostics redesign or audit lifecycle redesign.

## Related

- Epic [20-cli-ux-and-ergonomics](../epics/20-cli-ux-and-ergonomics.md)
- [Shared classifier and safe repair](6g77rn6b9wf8-narrow-the-near-miss-finding-recognizer-and-re-measure-across-every-entity-type.md)

## Design contract (2026-10-06)

Expose an additive, runtime-derived `unparsed_findings` count on the audit DTO and
an optional `unparsed` list column. Zero remains omitted in JSON. Parsed finding
counts and percentages retain their existing denominators; incomplete audits must
not advertise "no findings" or "ready to close" without qualification. Human
list/show output points to `audit lint <audit>` for the exact heading and remedy.
The secondary adapter establishes body-derived counts alongside existing tallies;
portable consumers use the same domain value without reading local paths.

## Implementation evidence (2026-10-06)

Audit reads carry runtime UnparsedFindings separately from parsed findings/tallies.
Human list/show distinguish empty, wholly unparsed, mixed, and settled-but-incomplete
documents and point to audit lint. Parsed percentages retain their denominator;
incomplete audits cannot advertise readiness. Ordinary reads remain informational.

The audit field addition is additive: optional unparsed_findings is omitted at
zero, and the opt-in unparsed list column accepts canonical unparsed_findings.
The combined revision 1.82 is classified NOT ADDITIVE because init now rejects
formerly accepted explicit selector combinations and incomplete audit readiness
is deliberately withheld. Callers must choose one explicit init target.
Projected missing values are blank; existing default table/CSV columns do not change.
A loaded-record column-lift regression now pins optional-column policy so future
adapter decoration cannot silently expand default output.

Portable core and filesystem-backed CLI tests pin counts, readiness, ordinary prose,
fenced examples, non-default typed JSON, and aliases. Generated command docs, schema
comments, contract goldens, race suite, lint, build, and planning/audit lint passed.
Remaining TUI and closure/deferral consumers are explicitly tracked in
[6gh82rm9sf3b](6gh82rm9sf3b-carry-unparsed-finding-evidence-into-audit-tui-and-lifecycle-decisions.md),
sequenced after this task; no general lifecycle redesign was folded into the read fix.

## External review handoff (2026-10-06)

Prepared two independent audits with no findings:
[Codex](../audits/6gh86jxj4sve-2026-10-06-shared-write-and-audit-safety-implementation-codex.md)
and [Antigravity](../audits/6gh86jxtyx5k-2026-10-06-shared-write-and-audit-safety-implementation-antigravity.md).

Codex leads init/audit/machine-contract checks; Antigravity leads YAML preservation.
Both cross-check the other lens and must use independent dirty-state-capturing sandboxes,
bounded compiler-valid mutation evidence, and guarded one-audit transfer. At handoff,
implementation remained in-progress pending owner triage; final dispositions follow.

## Review triage (2026-10-07)

Codex completed real read/projection probes and non-default exact-schema validation.
Its shared-authority H1 finding is fixed with repeated actual lint --fix and JSON
readiness regressions; phantom settled findings cannot be manufactured from the reported
examples. Full race suite, lint, build, generated comparisons, and planning lint pass.
The unchanged compact audit info warning gap joins TUI/lifecycle consumers in the
existing 6gh82rm9sf3b followup, rather than claiming every audit consumer is migrated.
Antigravity independently confirmed qualified list output, absent readiness for
incomplete evidence, and opt-in projections; its shared YAML findings are fixed.
Both audits are closed and this task is completed in
[PR #287](https://github.com/andy-esch/taskflow/pull/287). These are read/readiness
guarantees, not completion of the separate lifecycle/TUI followup.

## Additional baseline audit evidence (2026-10-06)

Cross-referenced by audit 2026-10-06-adapter-hygiene: M1 (partial overlap — that finding stays open). M1 supplies the mechanism behind this task's symptom and a verified reproduction: the store's `parseAuditWithFindings` already computes `NearMisses` on every audit read, but `core.AuditWithBody` carries only `{Audit, Body}`, so `ShowAudit` discards them and both `cli/audit.go:490` and `tui/detail.go:1720` re-run the narrow `domain.ParseFindings` instead. Reproduced in a throwaway repo: with a `#### M-2.` heading in the body, `audit lint` reports "its work is invisible" while `audit show` says nothing on either surface, and `audit show --json` publishes no structured findings at all. M1 goes beyond this task in also proposing the de-duplication of the parse, which is why it is not marked tracked here.

This is pre-implementation evidence. The aggregate CLI warning is addressed here;
structured finding output and duplicate adapter parsing are not claimed fixed.
The broader M1 finding remains open, and remaining consumer warnings are sequenced
in 6gh82rm9sf3b rather than silently widening this task.
