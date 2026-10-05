---
schema: 1
id: 6ggdkztshmzz
status: in-progress
epic: 21-code-quality-architecture-hardening
description: Publish core-owned dependency impact and recovery intent consistently to human and machine callers.
effort: 4-8 hours
tier: 2
priority: medium
autonomy_level: 3
tags: [architecture, core, recovery, json]
created: "2026-10-04"
updated_at: "2026-10-05"
audit_sources: [2026-09-15-adapter-hygiene]
depends_on: [6gcwcf8rxe72]
started_at: "2026-10-04"
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

The pre-implementation 2026-10-04 survey found the rule in `core/service_task.go`,
`cli/render/dependency.go`, and `cli/render/render.go`; `DependencyMutationReceipt` and its
wire projection did not publish recovery guidance.

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

- [x] Core owns the newly unsafe impact decision and recovery intent; dependency
  and lifecycle renderers no longer re-derive Gate/Inconsistent policy,
  including comparisons against wire strings.
- [x] Focused tests pin no-op, safe, already unsafe, newly
  blocked/broken/inconsistent, and cleared impacts without changing eligibility
  or authorization behavior.
- [x] Human dependency receipts and machine projections preserve equivalent
  actionable recovery intent; interface-specific presentation does not determine
  whether recovery is needed.
- [x] Dry-run, refusal, and committed/partial outcomes retain accurate
  durability evidence and never suggest an unsafe blind retry.
- [x] Any public JSON additions follow ADR-0008's monotonic contract revision
  policy with schema, golden fixtures, changelog, and documentation updated
  together.
- [x] Portable service tests demonstrate the contract without filesystem paths;
  normal tests, lint, generated-doc checks, and planning lint pass.

## Implementation and review handoff (2026-10-04)

Implemented locally; remains **in progress** pending integration in [PR #282](https://github.com/andy-esch/taskflow/pull/282).
Source audit M1/L1 are fixed locally, not represented as merged; unrelated findings remain open.
Review handoff is frozen for the independent
[Codex](../audits/6ggjnz40teda-2026-10-04-core-owned-dependency-impact-recovery-implementation-codex.md)
and [Antigravity](../audits/6ggjnz4989cm-2026-10-04-core-owned-dependency-impact-recovery-implementation-antigravity.md)
audits. Both contain the same contract and hostile evidence requirements. Codex is
reviewed and closed with its sole finding tracked; Antigravity's test-coverage
finding is fixed and its audit is closed. Neither closure claims integration.

- `TaskGraphStateImpact.NewlyUnsafe()` and `ThreadProjectionImpact.NewlyInconsistent()`
  own the existing warning predicates. A different non-clear gate still warns, including
  broken → blocked: this is an inspection signal, not a new severity ordering.
- Dependency and lifecycle receipts share task-impact recovery guidance with runnable stable-ID
  blocker commands. Dependency failures distinguish no writes, a durable prefix, and all planned
  writes landing before an error; they require inspection before resuming. Retry mechanics,
  eligibility, overrides, and graph guards are unchanged.
- Impacts describe the complete proposed plan, not proof of persistence. Preview guidance is
  explicitly prospective; applied/remaining IDs and lifecycle `committed` retain durability.
  Pre-write CLI failures still omit a structured dependency mutation payload; their error text
  carries core guidance without representing planned edges as applied.
- Wire revision **1.81, additive** copies `newly_unsafe`/`newly_inconsistent` into the shared
  impact DTOs (including repair impacts) and optional dependency `remedy`. The classified wire
  changelog, schema/goldens, README, architecture, and generated dependency reference agree.
  No persisted frontmatter or closed vocabulary changes.

### Validation

`go test ./...`, `go test -race ./...`, `golangci-lint run ./...` (zero issues), and
`just build` pass. Generated CLI docs and schema comments match fresh temporary output;
`tskflwctl lint` and the adapter-neutral Thread frontier are healthy.

Portable tests cover safe/no-op/already-unsafe/cleared impacts, changed blocked/broken gates,
inconsistency, preview/refusal, partial/all-applied conflicts, and unchanged Thread/override
advice. Schema validation uses populated dependency, lifecycle/Thread, and error recovery branches.
File-backed migration tests exercise every durable prefix, including failure after the final write.

Five compiler-valid mutation probes ran in an independent non-Git temporary copy. Each failed its
named tests; all edits were restored byte-for-byte and the focused suite passed again:

| Broken invariant | Regression that rejected it |
| --- | --- |
| Core task warning always false | `TestTaskGraphStateImpactNewlyUnsafe` |
| Wire drops the task warning flag | `TestImpactWireCopiesCoreDecisionsAndDependencyRemedy`, `TestDependencyMutationPublishesCoreRecoveryIntent` |
| Lifecycle renderer reintroduces its gate-string rule | `TestMovesHumanUsesProjectedImpactFlagsNotWireStatePolicy` |
| Dependency receipt drops core guidance | `TestServiceDependencyAndLifecycleShareRecoveryIntent` |
| Conflict retry continues after durable writes | `TestServiceDependencyRefusalAndCommittedFailureRecovery`, `TestServiceDependencyPartialRecoveryRequiresInspectionAndNeverAutoRetries` |

Fresh-binary throwaway-space dogfood verified JSON preview with no applied IDs, a human newly-blocked
warning with an executable blocker command, a no-op without recovery advice, removal clearing the
warning, and clean lint. An unrelated init selector ambiguity is tracked in
[its own followup](6ggjmtmdd54w-prevent-silently-ignored-target-selectors-during-init.md), sequenced in
the CLI contract Thread rather than added as a refactor closure blocker.

### Codex review disposition (2026-10-04)

Codex found no new core/JSON/retry defect after independent real-store, consumer,
schema, and coordinated mutation probes. Its sole finding is a pre-existing TUI
consumer bug: a completed Thread's sole member can be reopened with zero task
impacts, so TUI feedback discards the nonempty Thread recovery remedy. The owner
reproduced it and tracked
[6ggkdbg0816h](6ggkdbg0816h-preserve-thread-only-lifecycle-recovery-guidance-in-the-tui.md)
after this task in the TUI refinement Thread. It is not represented as fixed or
added to this refactor's closure path.

Review evidence also strengthened permanent tests, without production changes:

- The real migration prefix test injects `ErrConflict` after each write with four
  retries available, asserts one Service attempt and byte-level applied-ID parity,
  then resumes from live operator-edited remaining declarations. It also covers
  failure after the last planned write, which must remain visible rather than turn
  into a successful no-op.
- `TestMutationImpactSchemaBranches` validates 20 populated cases over dependency,
  lifecycle, repair, and their nested error envelopes. True/false warning fields
  remain required, optional remedy omission is deliberate, and structurally deleting
  a required flag is rejected by the schema, not merely by an invalid-JSON parser.
- Removing the durable-prefix retry stop in the independent non-Git probe copy
  makes all three real prefix tests fail; restored tests pass. These permanent
  regressions supplement, not replace, the review's sandbox-local evidence.

Codex audit closure means M1 has an explicit task destination, not that its followup
or integration is done. Antigravity reviewed the original captured implementation;
the post-review delta is tests and planning only, with no production-policy changes.

### Antigravity review disposition (2026-10-05)

Accepted and independently reproduced M1: replacing the wire task warning with
`After.Gate != GateClear` passed the original focused test and full wire/CLI suites.
The production mapper was already correct; this was a regression-test blind spot,
not a shipped warning defect or a new design decision.

`TestImpactWireCopiesCoreDecisionsAndDependencyRemedy` now exercises all **36**
before/after gate/inconsistency pairs through dependency and lifecycle conversion,
plus all **four** Thread consistency transitions. Role-only changes remain quiet,
unchanged unsafe states do not become new warnings, and owner guidance survives
even when the warning flag is false. Core's separate literal predicate table
remains the policy oracle; the wire test checks copying, not a second policy.

Three compiler-valid mapper probes ran only in an independent temporary copy:
the reported shallow gate check, a changed-gate-only rule dropping new
inconsistency, and a Thread rule using only `After.Inconsistent`. The expanded
test rejects each for the intended semantic mismatch; each restored test passes,
and the restored mapper matches the source byte-for-byte. Full race tests,
lint, build, generated-doc parity, and planning/audit lint pass.

Antigravity M1 is fixed locally in PR #282 and the audit is closed. No additional
task or artificial graph edge is needed for this bounded test hardening. Codex's
pre-existing TUI followup remains tracked, not fixed by this review.

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
