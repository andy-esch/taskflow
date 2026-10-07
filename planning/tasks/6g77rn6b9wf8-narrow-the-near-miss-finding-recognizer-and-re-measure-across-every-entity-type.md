---
schema: 1
id: 6g77rn6b9wf8
status: in-progress
epic: 20-cli-ux-and-ergonomics
description: The recognizer claims ordinary prose headings and lint --fix rewrites them; the clean measurement was taken only on the tool-scaffolded audit corpus.
effort: 2-4 hours
tier: 2
priority: high
autonomy_level: 3
tags: [audit, findings, lint, robustness]
created: "2026-09-05"
updated_at: "2026-10-06"
started_at: "2026-10-06"
---

# Narrow the near-miss finding recognizer and re-measure across every entity type

## Objective

Make near-miss finding detection safe enough for every consequence attached to it. The same
recognizer supplies lint errors, automatic body rewrites under `lint --fix`, and hard refusals in
audit write paths, so an ambiguous match can corrupt prose or block a valid edit. Re-derive the
recognition contract from the canonical finding grammar plus evidence available in the audit body,
preserve detection of genuine dropped findings, and measure the result across every planning entity
type rather than only the tool-scaffolded audit corpus.

Issue #205 and PR #206 close two demonstrated classes: space-separated word/number headings and
narrative headings that repeat an already parsed finding code. They do not settle uninterrupted
ordinary headings such as `S3 Storage Architecture`, `V2 Migration Guide`, `MP3 Audio Support`, or
`Top3 Recommendations`, nor the mismatch between the parser's unbounded code length and the
near-miss recognizer's four-letter/three-digit bounds. This task owns the complete contract rather
than another one-off regex exclusion.

## Acceptance criteria

- [x] A checked-in fixture matrix classifies each representative heading as canonical,
  auto-repairable drift, diagnostic-only ambiguity, or ordinary prose; the documented classifier
  is shared by lint, `lint --fix`, and audit write guards rather than re-derived by each consumer.
- [x] Ordinary headings including `Top 3`, `Edge 1`, `Step 1`, `Tier 1`, `Path 0`, `Why 3`, `S3
  Storage Architecture`, `V2 Migration Guide`, `MP3 Audio Support`, and `Top3 Recommendations`
  produce no finding-header rewrite and do not make `audit append` or `audit edit` refuse the body.
- [x] A narrative heading whose normalized code already belongs to a parsed finding in the same
  body is preserved byte-for-byte, including repeated closeout references such as `## H1 fixed`
  and `### M2 — tracked`; a genuinely drifted, previously undefined finding remains detectable.
- [x] Genuine code-token drift seen in practice remains loud and repairable when the body supplies
  the evidence required by the classifier: colon/no-period separators, hyphenated and underscored
  codes, lowercase and bold code tokens, and `BTA-01`/`WR-01`-style multi-letter codes.
- [x] The recognizer's supported code length is deliberately reconciled with
  `findingHeaderRe`'s `[A-Z]+[0-9]+` grammar: drift involving `AUDIT-1`, `CRITICAL-1`, or `H-1234`
  is either detected under the same evidence rule or explicitly rejected by a documented canonical
  code bound enforced by both parser and writer.
- [x] Corpus coverage scans Markdown bodies for tasks, epics, research, Threads, and audits—not
  only `planning/audits`—and reports the number and disposition of every classifier hit. The
  auto-repair/write-refusal class has zero unreviewed false positives.
- [x] End-to-end tests prove that `lint --fix` cannot turn the known prose fixtures into phantom
  findings or follow-on `missing **Status:**` errors, while a high-confidence malformed finding is
  still repaired and an equivalent newly introduced malformed finding is still rejected at write
  time.
- [x] The contradictory historical claim about bare numeric headings and seven-hash headings is
  resolved: tests and the completed task's amendment state whether each form is intentionally
  outside the grammar or safely detectable, without treating ordinary numbered sections as
  findings.
- [x] Canonicalization remains byte-idempotent on a clean body, and the full race-enabled test,
  lint, and module-tidiness checks pass.

## Out of scope

- Making the canonical parser tolerant of multiple finding syntaxes; one exact stored form remains
  the convergence target.
- Changing how audit list/show distinguish an empty audit from one with unparsed findings; that is
  owned by `6g77rn6em6n8`.
- Closing audit creation/edit baseline gaps, candidate-task marker ownership, or refactoring the
  task/audit body-transform implementations; those have separate tasks.
- Rewriting historical audit prose merely to avoid a classifier match.

## Related

- Epic [20-cli-ux-and-ergonomics](../epics/20-cli-ux-and-ergonomics.md)
- GitHub [issue #205](https://github.com/andy-esch/taskflow/issues/205) and its bounded first repair,
  [PR #206](https://github.com/andy-esch/taskflow/pull/206)
- Original implementation task
  [`6g72wf39pyhb`](6g72wf39pyhb-make-near-miss-audit-findings-loud-and-self-repairing.md)
- Independent implementation audits:
  [Claude](../audits/6g7731f5kjkw-2026-09-05-tool-owned-sub-entity-writes-implementation-claude.md)
  and
  [Antigravity](../audits/6g7731f8zzjq-2026-09-05-tool-owned-sub-entity-writes-implementation-antigravity.md)
- Adjacent task: [report unparsed findings on audit read surfaces](6g77rn6em6n8-report-unparsed-findings-on-the-audit-read-surfaces.md)

## Classifier contract and evidence (2026-10-06)

ClassifyFindingHeaders is shared by lint, safe repair, and introduced-heading write
guards. Its four classes are canonical, evidence-backed repairable drift,
diagnostic-only ambiguity, and ordinary prose. Repair requires local authoritative
Status metadata before the next heading plus Findings context, a known severity band,
or code-token decoration. Defined-code narrative references are ordinary; duplicate
definitions remain diagnostic-only. Repeated malformed candidates cannot create
duplicate canonical codes. Adding Status evidence to a previously ambiguous heading
does not evade the write guard.

Bare S3/V2/MP3/Top3 headings, ordinary word-number headings, and fenced examples are
not rewritten or refused. The code grammar has no arbitrary length bounds, matching
the canonical parser. Repair preserves CRLF and unrelated bytes. Numeric codes and
seven-hash headings intentionally remain outside the grammar, with explicit fixtures.

Live measurement before this batch's followup: 406 task bodies (17 ordinary hits),
15 epics (zero hits), 32 research bodies (8 canonical/28 ordinary), 7 Threads (zero),
151 audits (737 canonical/28 ordinary). Zero repairable or ambiguous live hits;
the checked-in hostile matrix exercises those branches independently. Filesystem
append/edit/repair tests pin prose safety and genuine malformed-finding refusal.
Race suite, lint, build, module tidiness, and planning/audit lint passed.

Amends the original completed task's historical heuristic: raw code-shaped matches
are not repair permission and suspected headings never enter parsed denominators.
Its narrow audit-only measurement did not establish safety for all entity kinds.

## External review handoff (2026-10-06)

Prepared two independent audits with no findings:
[Codex](../audits/6gh86jxj4sve-2026-10-06-shared-write-and-audit-safety-implementation-codex.md)
and [Antigravity](../audits/6gh86jxtyx5k-2026-10-06-shared-write-and-audit-safety-implementation-antigravity.md).

Codex leads init/audit/machine-contract checks; Antigravity leads YAML preservation.
Both cross-check the other lens and must use independent dirty-state-capturing sandboxes,
bounded compiler-valid mutation evidence, and guarded one-audit transfer. Review has
not run; implementation remains in-progress pending owner triage.

## Review triage (2026-10-07)

Codex H1 was confirmed and fixed. Recognition and canonical parsing now share the same
byte-offset-preserving authority boundary: visible heading metadata or a standalone
Status line. Inline code, indented examples, and body prose/quotes cannot authorize a
repair or supply a terminal status. Real lint --fix regressions preserve example bytes,
invent no finding/readiness, and permit narrative append. Domain tests additionally
pin real status spans, literal title code, code-formatted decoration replacement, and
CRLF preservation. The full race suite and lint pass. Codex findings are settled;
Antigravity remains pending, so task lifecycle completion is not yet claimed.
