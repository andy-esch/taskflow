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
depends_on: [6ghht05bzxkk]
updated_at: "2026-10-09"
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

## Parsed-settlement boundary for the scope pass (2026-10-09)

[The settlement-policy task](6ghht05bzxkk-align-audit-lifecycle-and-finding-writes-with-parsed-settlement-policy.md)
guards lifecycle verbs and structured finding status writes, not arbitrary body
authoring. `AppendAuditBody` and whole-file `EditAudit` can still introduce parsed
nonterminal findings in a non-open audit; ordinary lint now reports those states.
The fresh CLI confirmed the append route in a disposable space: close an empty
audit, append a canonical in-progress finding, then `audit lint` reports the
closed-bucket defect (exit 11). Review supplied creation bodies separately for the
existing near-miss gap: CLI audit creation always chooses the open bucket.
Include these routes in this task's scope pass rather
than claiming the structured guard is universal. Preserve repairs to existing
defects, explicit reopen, surgical unknown-field preservation, and CAS revalidation;
decide a baseline-aware body-write policy before blocking every edit of a damaged
audit. This work follows the settlement-policy task, so it reuses its vocabulary
instead of implementing a second disposition rule.

## Structured finding receipt scope check (2026-10-09)

The same receipt/source boundary also affects `newAuditFindingCmd`: after the
guarded `EditFinding` returns, it calls `ShowAudit(a.Slug)` to obtain the response
body. This is a second read by display label, not the source/proposed body that
was protected by the transform. A fresh-binary disposable probe confirmed that
`audit finding <source-ID> H1 --note ... --dry-run --json` returns the old stored
body, while its audit metadata describes the proposed edit. Slug/ID collisions
can also redirect that second read; examine the mutation receipt's declared ID
versus authoritative source ID alongside the whole-file-edit route above.

Scope coherent preview/applied receipts and source-backed followup diagnostics
together, without granting a new read authority after a successful guarded write.
Pin preview body/count agreement, source drift and hostile selectors, plus
concurrent changes between commit and rendering. This is a pre-existing receipt
gap, not a bypass of the new settlement decision. Safe note rendering itself is
fixed in [the settlement-policy review](../audits/6gj5ct38xq4t-2026-10-09-audit-parsed-settlement-implementation-codex.md).
