---
schema: 1
id: 6g72wf39pyhb
status: in-progress
epic: 20-cli-ux-and-ergonomics
description: Silent-drop of near-miss audit finding headers hides actionable work; add a narrow recognizer, a loud lint rule, and lint --fix auto-repair.
effort: 4-8 hours
tier: 2
priority: high
autonomy_level: 3
tags: [audit, findings, dx, robustness]
created: "2026-09-05"
updated_at: "2026-09-05"
depends_on: [6fm8p1cj11qf, 6g392b0rps7w]
audit_sources: [planning/audits/6g750etcvfqg-2026-09-05-tool-owned-sub-entity-writes-implementation.md]
started_at: "2026-09-05"
---
# Make near-miss audit findings loud and self-repairing

## Objective

A malformed finding header is dropped by `ParseFindings` with no signal anywhere: `audit lint`
prints green, `audit show` renders `0% settled 0/0`, and the finding's actionable work becomes
invisible. Close that gap with a narrow near-miss recognizer, a loud lint rule, and an
auto-repair — so drift is corrected by the tool instead of by asking the author to rewrite.

## Measured diagnosis

Every silent loss is in the **code token** `findingHeaderRe = ^#{2,6}\s+([A-Z]+\d+)\.`.
Nothing about `**Status:**` placement or separators causes one. Verified across 19 shapes
against the built binary:

- **Silent loss (7):** `H1:` · `H-1.` · `h1.` · `H1 ` (no period) · `**H1.**` · `1.` · depth-7
- **Already loud, `audit lint` exit 11 (6):** em-dash separator · pipe separator · bare spaces ·
  `**Status**:` · invalid status word · missing status
- **Already correct (4):** `**Status:**` on its own line · `BTA1.` · `H01.` · `## H1.` (depth 2)

Single-variable attribution of the canonical incident (audit `6g5axb85endz`, eight findings
rendered `0/0` and rewritten by hand):

| Variant | Parsed |
|---|---|
| `#### BTA-01 — title`, `**Status:**` on its own line (as it happened) | 0 |
| **only** the code token fixed to `B1.`, status still on its own line | 1 |
| canonical code, em-dash in title, status on its own line | 1 |

The hyphenated code was solely responsible. Two prior write-ups (this task's predecessor and its
own earlier revision) additionally blamed the em-dash and the newline before `**Status:**`; both
are innocent, and that misattribution has already propagated into reviewer briefs as a false
rule ("do not put status on a separate line"). Correct the briefs as part of this work.

## The gap, stated precisely

`audit append --help` says *"Finding grammar is left to `audit lint`."* `audit lint` validates
only findings that already parsed. Everything that fails to parse falls between the two.

## Design: a narrow near-miss recognizer beside the canonical matcher

This is the pattern `LintAcceptanceCriteria` already uses — `acCheckboxOKRe` (canonical) beside
`acCheckboxyRe` (botched) — including its stated rule that the near-miss class stay
*deliberately narrow so a warning is high-confidence, not noise*. Findings are the one
actionable sub-entity that never got it.

Recognizer (letter-led code is the essential narrowing):

    ^#{2,6}[ \t]+\*{0,2}([A-Za-z]{1,4})[-_ ]?(\d{1,3})[.:)—–-]?\*{0,2}[.:)—–-]?[ \t]+\S

Measured on the real corpus (57 audits, 426 canonical findings): **9 of 9 silent shapes caught,
0 false positives.** For contrast, a code-agnostic near-miss pattern of the shape previously
proposed (`[A-Za-z0-9_-]+`) yields **116 false positives**, because it matches ordinary numbered
section headings (`### 1. Lifecycle and Dependency Semantics`) used in 12 audits.

A recognizer hit that is not canonical becomes a lint error naming the exact replacement, and
counts toward the denominator so the audit cannot read as clean.

## Rejected alternatives, with the evidence

- **Orphan `**Status:**` detection.** 51 unclaimed markers across 37 of 57 audits, every one
  legitimate instructional prose in *inline* backticks (`` `**Status:**` ``), which fenced-block
  masking does not cover. 100% false-positive rate; would turn lint red corpus-wide.
- **Structural lint of "the findings section".** No such concept exists — `findingHeaderRe` is a
  global scan. Audits legitimately nest findings under varying headings (`### Findings` inside
  `## Reviewer report`); such a rule would fail conforming audits.
- **Explicit zero-findings attestation (`**Findings:** none`).** Introduces a new hand-authored
  magic string to solve a hand-authored-magic-string problem. Render `no findings` instead of
  `0% settled 0/0` and lint an open audit that has neither findings nor recognizer hits.
- **Findings in YAML frontmatter.** Findings are multi-paragraph prose; this fights the
  git-native/AI-native premise and is already out of scope.
- **Tolerant parsing (accept `H-1.` as canonical).** Strictly worse than the recognizer:
  tolerating drift lets the corpus diverge instead of converging.

## Scope

- Near-miss recognizer + `audit lint` rule with line number and exact canonical replacement.
- `lint --fix` canonicalizes near-miss finding headers (the auto-repair that removes the
  rewrite-by-hand loop). Codes are renumbered only when the drift is unambiguous.
- Honor `audit append`'s documented promise: reject or auto-fix a recognizer hit at write time,
  so the check cannot be forgotten. Same on `audit edit` save.
- Replace `0% settled 0/0` with an explicit no-findings rendering in `audit show` / `audit list`.
- Regression fixtures for all 7 silent shapes plus the 12 corpus false-positive shapes.
- Correct the "status on a separate line" misstatement in reviewer-brief boilerplate and
  `schema audit`.

## Out of scope

- `audit finding add` / code allocation — worth having, but it protects only the path that goes
  through it, while the recognizer protects append, edit, raw writes, and hand edits alike.
  Track separately.
- Generalizing the writer/recognizer/lint/fix contract across every actionable sub-entity
  (candidate-task `✅/⏳/⛔` markers have no code referencing them at all). Branch from this
  slice once the pattern has its second working instance.
- Converting audits to a validated AST or template-declared structure; migrating historical audits.

## Sequencing

`fold-audits-into-the-top-level-lint-command` (`6fm8p1cj11qf`) is a genuine prerequisite for the
signal to be *seen*: `audit lint` is today a separate command an agent must remember, so a new
rule inside it is easy to miss. Land that first, or land it with this.

## Related

- Supersedes [`6g5erdkd5pk4`](6g5erdkd5pk4-make-audit-findings-unforgeable-a-finding-writer-command-and-near-miss-lint.md) (deprecated)
- Prerequisite: [`6fm8p1cj11qf`](6fm8p1cj11qf-fold-audits-into-the-top-level-lint-command.md)
- Precedent: `LintAcceptanceCriteria` / `acCheckboxyRe` in `internal/domain/body.go`
- Grammar: `internal/domain/finding.go`
- Epic: [20-cli-ux-and-ergonomics](../epics/20-cli-ux-and-ergonomics.md)

## Acceptance criteria

- [ ] A narrow, letter-led near-miss recognizer sits beside findingHeaderRe and
  flags every non-canonical finding header with its line number and exact
  canonical replacement.
- [ ] The recognizer produces zero false positives across the existing audit
  corpus, verified by a fixture built from the 12 audits that use numbered
  section headings.
- [ ] All seven silent-loss shapes (H1:, H-1., h1., H1 without a period, **H1.**,
  bare 1., depth-7) are caught by lint rather than dropped.
- [ ] lint --fix canonicalizes a near-miss finding header in place through the
  audit body-transform CAS (6g392b0rps7w), so a drifted audit is repaired by the
  tool instead of rewritten by its author; FixFrontmatter stays
  frontmatter-only.
- [ ] audit append and audit edit honor the promise in audit append --help by
  routing a recognizer hit through the same body-transform op, so the check
  cannot be forgotten or bypassed by a raw write.
- [ ] audit show and audit list render an explicit no-findings state instead of
  0% settled 0/0, and lint flags an open audit with neither findings nor
  recognizer hits.
- [ ] Reviewer-brief boilerplate and schema audit no longer state that **Status:**
  on its own line breaks parsing, because it does not.
