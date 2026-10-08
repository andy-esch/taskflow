---
schema: 1
id: 6g77rn6hvmh8
status: next-up
epic: 20-cli-ux-and-ergonomics
description: audit new --body bypasses the near-miss guard entirely, and EditAudit discards a frontmatter-split error so it can refuse drift it did not introduce.
effort: 2-3 hours
tier: 3
priority: medium
autonomy_level: 3
tags: [audit, store, robustness]
created: "2026-09-05"
---

# Close the remaining audit body write-guard gaps

## Objective

<why / what — one short paragraph>

## Acceptance criteria

- [ ] <observable outcome>

## Out of scope

- <explicitly excluded>

## Related

- Epic [20-cli-ux-and-ergonomics](../epics/20-cli-ux-and-ergonomics.md)

## Adjacent source-identity check for the scope pass (2026-10-08)

When scoping whole-file edit guards, also examine the CLI's post-edit lint route:
`newAuditEditCmd` currently receives a domain-only `EditAudit` result, calls
`LintAudits` with the original selector, and prints `audit lint <display-slug>`.
The source-backed diagnostics principle applies here too; duplicate slugs or an
ID-shadowing slug can make advice ambiguous or address a different audit. Do not
substitute a frontmatter-declared ID or re-resolve a display label after committing
an edit; decide what source evidence the edit receipt must retain.

This is a pre-existing whole-file-edit receipt/diagnostics gap, not part of the
new incomplete-evidence read and lifecycle fix. Its demonstrated selector hazard
is recorded in [the Codex identity review](../audits/6ghhwe5kzj01-2026-10-07-audit-incomplete-evidence-implementation-codex.md).
