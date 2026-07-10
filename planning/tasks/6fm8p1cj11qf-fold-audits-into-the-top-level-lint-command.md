---
schema: 1
id: 6fm8p1cj11qf
status: ready-to-start
epic: 21-code-quality-architecture-hardening
description: Top-level `tskflwctl lint` checks tasks + epics but not audits (audit checks live only behind `audit lint`); fold audit lint into the lint roster so the hygiene gate covers all three entities.
effort: Unknown
tier: 3
priority: medium
autonomy_level: 3
tags: [lint, audit, core]
created: "2026-07-09"
---
# Fold audits into the top-level `lint` command

## The gap (researched 2026-07-09)

`tskflwctl lint` — the canonical hygiene gate ("keep `planning/` lint-clean") —
validates **tasks and epics but not audits**. Its success line even says so:
`✔ all active tasks and epics pass lint`. Audit validation exists, but only
behind the *separate* `audit lint [audit]` subcommand, which a plain `lint` (and
any CI/hygiene run of it) never invokes.

Two layers to the gap:

### A. Wiring — audits are absent from the top-level roster

`core.Service.Lint()` (`internal/core/service.go:216`) lists tasks + epics and
returns their issues; it never touches audits. So audit frontmatter problems
(missing/foreign `bucket`, missing/drifted `id`, and finding-level status-vocab /
bucket↔state violations) — all of which the code *already* detects via
`Service.LintAudits()` (`internal/core/finding.go:196`) — don't surface through
the command everyone actually runs.

A sharper symptom of the split: `lint --fix` already walks audits and backfills
their ids (`internal/store/fix.go:16-18,88`), yet read-only `lint` never reports
an audit problem. The repairer sees audits; the reporter doesn't.

### B. Coverage — no field-level `LintAudit`

Even `LintAudits` only checks findings + id + bucket. There is no domain
`LintAudit(a)` analogous to `LintTask`/`LintEpic` that validates an audit's own
content frontmatter: `area`, `date`, `updated_at` (`internal/domain/audit.go`).
Tasks require `created` be present and `YYYY-MM-DD`; the audit `date` field gets
no presence or format check at all. This is exactly the archived-entity asymmetry
epic 26 flags (the desirelines cleanup found a `deprecated_date`-vs-`deprecated_at`
audit and dateless files that active-only lint let sail through).

## Proposed change

Fold the existing audit checks into the top-level `lint` so one `tskflwctl lint`
run reports tasks + epics + audits:

- Run `LintAudits` (or its per-audit `check`) from `runLint` / `Service.Lint`,
  merging audit `LintResult`s and `FileProblem`s into the same render + exit path.
- Update the human footer + success wording ("tasks, epics, and audits") — the
  neutral "item" noun for the mixed roster is already in place.
- Keep `audit lint [audit]` as the focused single-audit subcommand; this task is
  about the *aggregate* gate covering audits too, not replacing it.

## Acceptance criteria

- [ ] `tskflwctl lint` (no subcommand) reports audit issues — bad/missing bucket,
      missing/drifted audit id, and finding-level (status vocab, bucket↔state) —
      in the same pass as tasks and epics.
- [ ] `lint --json` includes audit results/problems in the envelope.
- [ ] Success message no longer claims only "tasks and epics".
- [ ] `audit lint` behavior is unchanged; no double-reporting when both are run.
- [ ] A test covers an audit with a frontmatter defect being caught by top-level `lint`.

## Out of scope / deferred

- **Which audit content fields are *required*** (`area`/`date` presence, `date`
  format, `updated_at`) plus a new `LintAudit(a)` field-nag — that's a strictness
  decision for epic 26's ADR (per-entity/per-status matrix, Q1/Q9). This task is
  the *wiring* fix; add the field nags once the ADR decides them.

## Related

- Epic [26-frontmatter-schema-declared-validation-contract](../epics/26-frontmatter-schema-declared-validation-contract.md) — owns the field-strictness policy (Q1 strictness matrix, Q9 entity coverage) this defers to.
- Prior art: `Service.Lint` (`core/service.go:216`), `Service.LintAudits` (`core/finding.go:196`), `runLint` (`cli/lint.go:33`), `FrontmatterBucketIssues` / `MissingIDIssue` / `IDDriftIssue` (`domain/lint.go`), `FixFrontmatter` (`store/fix.go:16`).
