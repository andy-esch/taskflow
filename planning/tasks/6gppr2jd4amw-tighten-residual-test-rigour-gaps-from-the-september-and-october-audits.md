---
schema: 1
id: 6gppr2jd4amw
status: ready-to-start
epic: 21-code-quality-architecture-hardening
description: Make the smoke test's filesystem probe honest as root, stop pinning template catalog size, add the revisit_at case, and settle two redundant-guard asymmetries in source.
effort: half day
tier: 3
priority: medium
autonomy_level: 3
tags: [testing, core, ci]
created: "2026-10-23"
audit_sources: [planning/audits/6g8zy840z6gj-2026-09-11-test-rigour.md, planning/audits/6gfrcytd9n9a-2026-10-02-test-rigour.md]
updated_at: "2026-10-23"
---
# Tighten residual test-rigour gaps from the September and October audits

## Objective

Two test-rigour audits left five small, independent gaps. None is a live defect; each lets a
regression or an environment difference pass unnoticed. Settle them in one pass, preferring a
source fix where a test would only pin an unreachable or redundant branch.

## Acceptance criteria

- [ ] `TestSmoke_PublishedExitTaxonomyMatchesProcessBehavior` injects a filesystem failure that holds for every uid, or skips with a reason when the chmod does not bite, and its later assertions still run as root.
- [ ] `TestTemplateList_KindFilter` asserts a floor plus the per-row kind invariant; exact catalog contents are pinned only where the catalog is the subject.
- [ ] `ValidateTaskLifecyclePlan` has a negative case for `RevisitAt` with a non-deferred target.
- [ ] `planDependencyEdges` checks the dependent and its prerequisites symmetrically, or drops the unreachable dependent check with a comment naming `ValidateTaskGraphMutationSource` as the real gate; no test is added for an unreachable branch.
- [ ] `thread new`'s four input refusals are asserted once at the layer that enforces them, or one test makes the core spelling deliberately load-bearing, and the core/domain message case divergence is resolved.

## Out of scope

- Removing the `NewThread` validation preamble unless the chosen fix requires it.
- Coverage campaigns beyond these five findings.

## Related

- Epic [21-code-quality-architecture-hardening](../epics/21-code-quality-architecture-hardening.md)
- Audit [2026-09-11 test rigour](../audits/6g8zy840z6gj-2026-09-11-test-rigour.md), findings L1 and L3
- Audit [2026-10-02 test rigour](../audits/6gfrcytd9n9a-2026-10-02-test-rigour.md), findings L1, L2, L3
