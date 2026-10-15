---
schema: 1
id: 6gkg89ay3b5b
status: completed
epic: 21-code-quality-architecture-hardening
description: Close criterion identity, ID-diagnostic, field-ordering and create-option gaps from the correctness audit.
effort: 4-6 hours
tier: 2
priority: medium
autonomy_level: 3
tags: [correctness, identity, core, store, determinism]
created: "2026-10-13"
started_at: "2026-10-13"
audit_sources: [2026-10-13-correctness-and-errors]
updated_at: "2026-10-14"
completed_at: "2026-10-14"
---
# Fix task identity guards and deterministic field updates

## Objective

Resolve the four bounded findings in [the correctness audit](../audits/6gk9j6ff3ke9-2026-10-13-correctness-and-errors.md) before resuming documentation work. Keep the fixes in one reviewable PR with logical implementation groups.

## Scope

- M1: compare adapter-established source identities for criterion self-tracking, require usable identity evidence and bind the body mutation to the resolved source ID. Test aliases, declaration/slug drift and dry-run no-effects.
- M2: distinguish uppercase identity spelling from invalid Crockford characters. Preserve strict lowercase validity and identity-preserving manual advice; do not expand automatic repair. Check adjacent create/update/repair diagnostics using the same validation evidence.
- M3: validate task update keys and apply shared frontmatter updates in deterministic lexical order. Preserve existing key order, comments, untouched values and body bytes; do not turn updates into whole-file sorting.
- L1: reject Next plus Start in the application use case before reads, minting or creation, independently of Cobra.

## Acceptance criteria

- [ ] Self-tracking is rejected by authoritative source ID across supported ref shapes; valid cross-task tracking records the destination ID and mutates the resolved source identity, including pathless/drifted records. · **tracked:** by 6gkmexz4vh23 — Self-tracking and rename-bound mutations are fixed; strict missing-ID action selection remains in this follow-up.
- [x] Case-only ID defects recommend the lowercase spelling without claiming an excluded character or a new identity; genuine aliases, u and other invalid characters retain accurate, strict diagnostics without new automatic repair authority.
- [x] Identical multi-key updates append new keys in stable order while preserving old key order, comments and body; invalid multi-field requests report a deterministic first error and write nothing.
- [x] Core rejects Next plus Start with ErrValidation before effects for both ordinary and preview calls; the individual flags remain valid.
- [x] Focused red-before-green regressions, real-store or CLI probes, full race tests, lint, build, documentation and planning checks pass; owning comments/guides accurately describe the changed contracts.
- [x] All source-audit findings have evidence-backed dispositions, and the stable-diagnostics task records its M2 overlap/predecessor without adopting its vocabulary or schema changes here.

## Boundaries

No new transaction/port framework, automatic identity repair, lint-kind vocabulary, wire shape, Thread membership policy or package documentation migration. Existing source-set, guarded persistence, CAS and receipts remain authoritative. This small cross-cutting detour is not a member or artificial prerequisite of the documentation Thread.

## Related

- [Stable diagnostic kinds/applicability](6gcqz5b0j0dg-give-lint-diagnostics-stable-kinds-and-repair-applicability.md): M2 overlaps its diagnostic inventory; correct advice first, retain the design gate.
- [Shared entity-integrity design](6g7wsr8yvt1a-define-the-next-shared-entity-integrity-foundations.md): broader cross-entity identity/transaction design remains separate.
- [Selective package contracts](6g9cz5sk3b0d-adopt-selective-contract-focused-go-package-documentation.md): resume after this detour without imposing an unrelated graph prerequisite.

## Implementation and evidence (2026-10-14)

Owner-authored tests reproduced all four original findings before fixes. Criterion
tracking now uses both adapter-established source IDs and a stable mutation target;
portable tests include declaration/slug disagreement and missing/blank evidence.
The real CLI exercises all 5x5 reference pairs across write/preview (50 cases), with
exact source-byte checks. Conflicting create options refuse before reads or minting.

Shared frontmatter updates keep old keys/comments/body in place and append new keys
lexically; repeated LF/CRLF tests also preserve untouched literal scalars. Core
validation reports a stable first error across unknown/unset/typed/owned fields.
ID diagnostics share strict evidence across filename/create/update/refusal paths;
case-only and mixed alias advice does not broaden automatic repair authority.

Full just test (race), just lint (0 issues), just build, just docs-check, planning
lint --links and whitespace checks passed. In an unregistered throwaway space, ID
and prefix self-tracking refused with exit 11 and unchanged SHA-256; 10/10 invalid
multi-field requests named aaa; valid tracking retained its destination ID after
rename; reversed custom updates appended alpha/beta/zeta. Case-only list/fix JSON
reported the lowercase spelling and manual refusal while preserving bytes and the
actual uppercase filename. Validation failures in hostile fixtures were expected,
not counted as successful lint of damaged input.

All four original findings are fixed and the source audit is closed. An additional
owner-discovered reason mismatch is tracked separately by
[align tracked writes and lint reason semantics](6gkkzvr92waq-align-tracked-criterion-writes-and-lint-reason-semantics.md),
sequenced after this identity correction with an explicit policy gate. Writer/lint
disagreement there is not silently resolved by changing accepted handoff semantics.

At the initial handoff, the six checked criteria reflected owner evidence and the
task stayed in progress for independent review. Review below records the later
strict-action limitation rather than retaining an unconditional completion claim.
No Git branch, staging, commits or pushes had been performed at that handoff.

## Independent review handoff (2026-10-14)

Prepared open, empty reports for [Claude](../audits/6gkm2xz5s9qx-2026-10-14-task-correctness-detour-implementation-claude.md)
and [Antigravity](../audits/6gkm2xzes3zk-2026-10-14-task-correctness-detour-implementation-antigravity.md).
Claude leads identity/action binding and application validation; Antigravity leads
deterministic YAML preservation and diagnostic/repair authority, with shared-contract
cross-checks. Required evidence includes actual producer workflows, a compiler-valid
mutation killed by the named regression, and a distinct systemic second pass.

The preceding import-map review's cache/configuration discovery and overbroad clean
verdict informed the briefs: inherited gates are labelled, baseline probes are not
universal proof, and the known reason-policy follow-up is explicit. At preparation,
review had not run. Both assigned briefs and the dirty/untracked implementation were
frozen for capture; reviewers were assigned independent clones and guarded one-file
transfer.

## Review triage (2026-10-14)

Claude's L1 is valid and fixed here: equal endpoint IDs now receive a canonical
resolution check before self-track refusal. Portable error-sentinel tests and a real
CLI two-file shared-ID fixture failed before the fix, then passed for both writes and
previews with exit 13, both candidate names and unchanged bytes. The 50 ordinary
self-alias cases still refuse with exit 11. A permanent real-store rename/CAS retry
regression also proves the tracked handoff follows the original source and leaves
the pre-existing prefix-match sibling unchanged.

Claude's pre-existing M1 is valid, independently reproduced and tracked by
[strict source-identity retries](6gkmexz4vh23-bind-task-mutation-retries-to-strict-source-identities.md).
Checkbox/state/criterion editing, append and field updates can retry onto a sibling
after a rename. An additional owner probe found that even the tracked ID query can
fall back to a sibling's ID-shaped slug after the source is deleted. Therefore
resolving once is not the complete fix: selected-identity action lookup must remain
strict inside the adapter. This follow-up owns the residual portion of criterion 1,
is next-up after this detour, and does not wait for a universal integrity framework.
The retry comment and planning-data guide now state these limits explicitly.

Antigravity found no additional defects and supplied useful independent YAML,
casing/manual-repair and mutation-kill evidence. Its broad clean verdict does not
rule out the demonstrated retry/selection gaps; owner annotations separate the
tested workflows from universal guarantees and correct its filesystem rationale.
Both audits have terminal dispositions; the original reports remain historical.

The reason-policy follow-up still needs its previously identified human decision.
Neither reason requirements nor automatic repair authority changed during review.

Post-triage validation passed: full just test (race), just lint (0 issues), just build,
just docs-check, a fresh count=1 race run of core/store/CLI/architecture navigation,
planning lint --links, both audit lint calls and tracked/untracked whitespace checks.
The wrong-target reproduction tests are retained only in owner scratch; the permanent
suite contains the passing tracked-rename and collision-diagnostic regressions.
The detour is ready for Git/PR handoff, with criterion 1's residual explicitly tracked.

## PR handoff (2026-10-14)

Completed through the lifecycle verb: five criteria are met and criterion 1's remaining
strict-action selection guarantee is explicitly handed to the next-up identity retry
task. Both reason-policy and strict-identity successors now have clear dependency gates.
The documentation Thread remains healthy and unchanged; this standalone detour has no
Thread neighborhood to export. Delivery branch: `fix/task-identity-and-deterministic-updates`.
