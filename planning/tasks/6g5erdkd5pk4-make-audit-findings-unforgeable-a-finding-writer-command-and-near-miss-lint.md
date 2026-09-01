---
schema: 1
id: 6g5erdkd5pk4
status: ready-to-start
epic: 20-cli-ux-and-ergonomics
description: Findings must be hand-authored into exact syntax and a near-miss parses to zero findings while lint stays green; add a writer command and diagnostics.
effort: Unknown
tier: 3
priority: medium
autonomy_level: 3
tags: [cli, agents, robustness, dx]
created: "2026-08-31"
updated_at: "2026-08-31"
---
# Make audit findings unforgeable: a finding writer command and near-miss lint

## Objective

A finding is the one structured sub-entity in the corpus that must still be hand-authored into
exact syntax, and a near-miss is *silently invisible*. The grammar in `internal/domain/finding.go`
requires `#### <CODE>. <title>` where `findingHeaderRe` is `^#{2,6}\s+([A-Z]+\d+)\.` — so a code with
a hyphen, an em-dash instead of the `.`, or a `**Status:**` on its own line yields **zero** parsed
findings. Both `audit lint` and `tskflwctl lint` then report green, because an audit whose findings
all failed to parse is indistinguishable from an audit with no findings.

Minimal reproduction (2026-08-31, fresh scaffold):

```
#### BTA-01 — A finding the parser cannot see

**Status:** open

**File:** `foo.go:1` | **Component:** foo

Body text.

**Recommendation:** fix it.
```

→ `audit lint` prints "✔ all audit findings pass lint" (exit 0), `tskflwctl lint` prints
"✔ all planning entities and dependency links pass lint" (exit 0), and `audit show` renders
`findings: 0% settled 0/0`.

This bit for real while writing audit `6g5axb85endz`: eight complete, evidence-backed findings
rendered as `0/0` and the entire findings section had to be rewritten after the fact. The cost is
not the typo — it is that nothing anywhere reported a problem, so the error survived a full
self-review and was caught only by a human reading the rollup.

The fix direction is already established twice in this repo: `task ac` owns acceptance criteria and
`audit finding` owns `**Status:**`/`**Resolution:**`, both precisely because hand-editing structured
markdown is unreliable. Finding *creation* is the remaining hole in that policy — `audit finding`
can re-stamp a finding it can already see, but nothing can bring one into existence.

## Acceptance criteria

- [ ] A writer command creates a whole finding block atomically — shape along the lines of
  `audit finding new <audit> --severity <band> --title <t> --file <f> --component <c> --effort <e>
  --urgency <u> [--body-file -]` — emitting the canonical header, metadata line, body, and
  `**Status:** open` in one validated write.
- [ ] The code is allocated by the tool, never typed: the severity band picks the letter and the
  command takes the next free number within that audit, so codes cannot collide or skip.
- [ ] `--effort` and `--urgency` are validated against the documented vocabularies at write time,
  matching how `audit finding --status` already validates the finding-status vocabulary.
- [ ] Lint reports a near-miss finding header: an ATX header whose text matches a loose code shape
  (letters, optional separator, digits) but not `findingHeaderRe`.
- [ ] Lint reports an orphan `**Status:**` marker — one in an authoritative position, outside a
  fenced block, that no parsed finding claims.
- [ ] `audit lint` and `audit show` distinguish "this audit has no findings" from "this audit has
  finding-shaped prose that did not parse". The success message must not be reachable for the
  second case.
- [ ] A regression corpus covers the shapes seen in practice — hyphenated code (`BTA-01`), em-dash
  instead of `.`, `**Status:**` on its own line, lowercase code, missing trailing `.` — each
  producing a diagnostic rather than silence.
- [ ] `schema audit` guidance points at the writer command as the authoring path, so the fenced
  example no longer has to be un-fenced and retyped by hand (the step where the `.` and `·` are
  lost).

## Out of scope

- **Widening `findingHeaderRe`** to accept more shapes. One authoritative grammar plus a writer
  command is the established posture; a looser grammar trades a silent failure for an ambiguous one
  and makes surgical `**Status:**` re-stamping harder to keep exact.
- Rewriting or reformatting existing audits — new diagnostics should be clean against the corpus,
  and any that are not are separate triage.
- The task-attached structured review loop in epic
  `27-agent-code-review-on-tasks-structured-review-loop`. That is a different noun with its own
  design questions; this task hardens the audit surface that ships today.

## Related

- Epic [20-cli-ux-and-ergonomics](../epics/20-cli-ux-and-ergonomics.md)
- Grammar and single source of truth: `internal/domain/finding.go` (`findingHeaderRe`, `statusRe`,
  `fieldValueRe`)
- Writer precedent: `audit finding` (status/resolution), `task ac` (acceptance criteria)
- Ancestor: [`6fd5r5c03v5y`](6fd5r5c03v5y-audit-finding-level-operations-query-write-lint-sync.md) —
  finding-level operations (query/write/lint/sync)
- Encountered in: audit `6g5axb85endz`
  ([2026-08-30-bulk-thread-apply-implementation-claude](../audits/6g5axb85endz-2026-08-30-bulk-thread-apply-implementation-claude.md))
