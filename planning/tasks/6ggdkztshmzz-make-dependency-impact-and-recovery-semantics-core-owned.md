---
schema: 1
id: 6ggdkztshmzz
status: ready-to-start
epic: 21-code-quality-architecture-hardening
description: Publish core-owned dependency impact and recovery intent consistently to human and machine callers.
effort: 4-8 hours
tier: 2
priority: medium
autonomy_level: 3
tags: [architecture, core, recovery, json]
created: "2026-10-04"
updated_at: "2026-10-04"
audit_sources: [2026-09-15-adapter-hygiene]
depends_on: [6gcwcf8rxe72]
---
# Make dependency impact and recovery semantics core-owned

## Objective

Let CLI, TUI, and machine callers consume the same core-owned decision about newly unsafe
dependency impacts and the recovery intent it implies. Primary adapters should format that
decision, not re-derive graph policy or decide which repairs to recommend.

## Evidence

[Adapter hygiene audit](../audits/6ga93gvffkpc-2026-09-15-adapter-hygiene.md), M1 and L1,
identifies one seam rather than two unrelated fixes: dependency mutation receipts omit the
recovery guidance present on lifecycle receipts, while renderers repeat the newly-unsafe
Gate/Inconsistent predicate. The spellings currently agree; this is an ownership and machine
contract gap, not evidence that today's dependency mutations compute the wrong graph.

The 2026-10-04 survey still finds the rule in `core/service_task.go`,
`cli/render/dependency.go`, and `cli/render/render.go`; `DependencyMutationReceipt` and its
wire projection do not publish recovery guidance. Recheck these sites before implementing.

## Scope

- Define the smallest core-owned impact/recovery contract that both dependency and lifecycle
  consumers can use. Choose a predicate, classified impact, or receipt data based on actual
  consumers; do not add a generic receipt framework.
- Remove adapter-owned Gate/Inconsistent policy, including the wire-string comparison in the
  lifecycle renderer. Wire conversion copies semantics rather than recomputing them.
- Give dependency JSON callers the same actionable recovery intent as human callers. Core owns
  the intent; interface-specific command spelling and visual treatment may remain in adapters.
- Keep explanations honest for no-op, dry-run, refused, and committed/partial outcomes. Recovery
  must not imply that an already durable write is safe to retry blindly.
- Follow ADR-0008's additive machine-contract revision policy and update schema, fixtures,
  changelog, and consumer-facing documentation together if public fields change.

## Acceptance criteria

- [ ] Core owns the newly unsafe impact decision and recovery intent; dependency
  and lifecycle renderers no longer re-derive Gate/Inconsistent policy,
  including comparisons against wire strings.
- [ ] Focused tests pin no-op, safe, already unsafe, newly
  blocked/broken/inconsistent, and cleared impacts without changing eligibility
  or authorization behavior.
- [ ] Human dependency receipts and machine projections preserve equivalent
  actionable recovery intent; interface-specific presentation does not determine
  whether recovery is needed.
- [ ] Dry-run, refusal, and committed/partial outcomes retain accurate
  durability evidence and never suggest an unsafe blind retry.
- [ ] Any public JSON additions follow ADR-0008's monotonic contract revision
  policy with schema, golden fixtures, changelog, and documentation updated
  together.
- [ ] Portable service tests demonstrate the contract without filesystem paths;
  normal tests, lint, generated-doc checks, and planning lint pass.

## Out of scope

- Changing graph eligibility, dependency authorization, or repair transaction semantics.
- Requiring every receipt to have a non-empty remedy regardless of its outcome.
- A cross-entity receipt framework, a new persistence backend, or TUI marker redesign.

## Related

- [Adapter-neutral planning Thread](../threads/6gcwd78p9r04-make-planning-data-access-adapter-neutral.md)
- [Merged composition boundary](6gcwcf8rxe72-isolate-cli-composition-wiring-and-enforce-controller-boundaries.md), PR #277
- [ADR-0008: monotonic machine-contract revisions](../adrs/0008-use-monotonic-revisions-for-the-json-machine-contract.md)
- [Final guarded-contract regression pass](6ggdkzv2tnta-pin-guarded-planning-mutation-boundary-contracts.md)
- [Existing shared-integrity design](6g7wsr8yvt1a-define-the-next-shared-entity-integrity-foundations.md)
  remains the home for broader retry/durability policy; this task closes an already observed
  impact/recovery boundary without pre-approving that design's child tasks.
