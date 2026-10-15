---
schema: 1
id: 6gkkzvr92waq
status: ready-to-start
epic: 21-code-quality-architecture-hardening
description: Resolve whether destination-only criterion handoffs need prose so normal writes and lint agree.
effort: 2-4 hours
tier: 2
priority: medium
autonomy_level: 2
tags: [criteria, validation, lint, contracts]
created: "2026-10-14"
depends_on: [6gkg89ay3b5b]
updated_at: "2026-10-14"
---
# Align tracked criterion writes and lint reason semantics

## Objective

Make the normal criterion handoff verb and lint agree about an ID-only handoff. The correctness detour's real binary probe accepted task ac --tracked 1 --to destination with no reason (exit 0), then lint --links rejected that same criterion as tracked with no reason (exit 11). Supplying explanatory prose made lint green.

## Design gate

Ask the user whether a resolved destination is sufficient handoff explanation or whether separate prose is mandatory. Recommendation: require meaningful prose, consistent with CriterionState.NeedsReason and current lint; verify the published CLI guidance and compatibility consequences before choosing. Do not change the persisted/wire grammar merely to resolve this disagreement.

## Evidence and scope

SetCriteriaTracked formats an ID prefix into the raw reason. SetCriteriaState validates that combined string, but parsing separates TrackedBy from Reason and lint checks only the latter. The historic criterion decisions require explanations for nonbinary states and separately require a tracked destination; they do not clearly settle whether the new structured destination satisfies the explanation requirement.

- Reconcile core/domain write validation, parsed lint, CLI help and source-adjacent comments under the chosen policy.
- Test missing, blank and meaningful explanation through both the application use case and a real CLI space; ensure writer output is lint-clean or refuses before effects.
- Retain stable source/destination identities, self-track refusal, dry-run behavior, body CAS and tracked pointers across renames.
- Decide whether existing ID-only records need only diagnostics or any explicit migration; no silent bulk rewrites.

## Acceptance criteria

- [ ] A user-approved policy establishes whether a destination alone suffices or separate explanatory prose is required, with compatibility implications recorded.
- [ ] Core and CLI handoff writes agree with parsed lint for absent, whitespace-only and meaningful explanations; every refused request leaves bytes unchanged.
- [ ] Existing tracked pointers stay readable and identity-stable; any legacy-record treatment is explicit and tested rather than silently migrated.
- [ ] Owning comments/help, focused tests, generated reference and planning evidence agree with the chosen behavior.

## Related

- [Correctness detour](6gkg89ay3b5b-fix-task-identity-guards-and-deterministic-field-updates.md): sequence after its canonical identity guard rather than repeat the slug/raw-ref bug.
- [Criterion vocabulary decisions](6g31g9f8x4cv-let-an-acceptance-criterion-say-more-than-done-or-not-done.md).
- [Correctness audit](../audits/6gk9j6ff3ke9-2026-10-13-correctness-and-errors.md): owner-discovered during follow-up probes, not an original external-review finding.
