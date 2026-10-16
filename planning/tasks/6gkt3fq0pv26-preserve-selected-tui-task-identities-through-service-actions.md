---
schema: 1
id: 6gkt3fq0pv26
status: completed
epic: 21-code-quality-architecture-hardening
description: Carry selected task identities through queued TUI mutations without restarting fuzzy reference selection in Service.
effort: 0.5-1 day
tier: 2
priority: high
autonomy_level: 2
tags: [correctness, identity, tui, core]
created: "2026-10-14"
depends_on: [6gkmexz4vh23]
updated_at: "2026-10-15"
audit_sources: [2026-10-14-task-correctness-detour-implementation-claude]
started_at: "2026-10-15"
completed_at: "2026-10-15"
---
# Preserve selected TUI task identities through service actions

## Objective and evidence

Carry an already-selected task ID across the primary-adapter/application boundary
without treating it as a new human reference. Preserve normal fuzzy CLI selection.

Owner probe: load the TUI, delete one selected task before its queued command runs,
and create another task whose human slug equals the missing task's ID. Calling the
actual `setFieldCmd` returned `editedMsg` and changed the impostor. Its stored source
ID differs from the original selected row. Passing the ID as a string to Service
restarted fuzzy selection; the strict persistence port cannot recover that lost intent.
Probe: `/private/tmp/taskflow-target-neighbors.VNxBO5/tui_selected_action_probe_test.go`.
This is a deterministic queued-action case, not an event-loop timing measurement.

## Scope

- Add narrow, explicit selected-identity application entry points or typed action
  inputs for task field set/unset and lifecycle move/defer. Reuse TaskMutationTarget
  at the task write boundary; do not parse paths or frontmatter declarations.
- Audit detail/edit/local-navigation actions for the same boundary; fix task mutation
  actions here and explicitly track different confirmed families.
- Keep I/O in tea.Cmd and retain generation, workspace/session and selection guards,
  human labels, error flashes, committed-warning handling and reload behavior.
- Do not globally interpret ID-shaped user slugs as strict IDs: initial CLI references
  are intentionally still ergonomic. Do not expand this into a universal entity API.

## Acceptance criteria

- [x] Selected task field/lifecycle actions follow renames but refuse deletion, duplicate IDs and ID-shaped/case-folded impostor slugs without mutating siblings.
- [x] Queued TUI command regressions exercise the actual action entry points before watcher refresh, including set/unset, move/defer and source/session changes.
- [x] Initial CLI fuzzy/ID-shaped human references still work, while explicit selected-identity actions never reselect a different record.
- [x] Existing authorization, lifecycle policy/receipts, stale-message handling, warning/reload and tea.Cmd boundaries remain covered.
- [x] Owning source comments/TUI guide describe the distinction, and normal validation passes.

## Sequencing and ownership

Depends on [strict task mutation targets](6gkmexz4vh23-bind-task-mutation-retries-to-strict-source-identities.md),
whose narrow internal target type and lifecycle retry rule this service boundary extends.
Belongs to the existing Refine Thread and TUI navigation Thread: selected action targets
are part of its shared navigation contract, not a reason to reopen the completed adapter refactor.

## Implementation and owner evidence (2026-10-15)

Branch: `fix/selected-tui-task-actions`. `SetSelectedTaskFields`, `MoveSelectedTask`
and `DeferSelectedTask` accept `TaskMutationTarget` without restarting reference
lookup. Field writes share one coercion/validation helper with ordinary `SetFields`;
lifecycle still authorizes exact IDs inside the guarded graph. CLI matching, optional
capabilities, wire/schema contracts and existing override policy are unchanged.

Actual set/unset/move/defer command closures now use those entry points. Commands
capture their originating Service; existing selection/load/session envelopes still
guard results. Already-issued commands may finish in their original space—this
does not add cancellation or authorize using a target with a different Service.

- [Queued action tests](../../internal/tui/selected_task_action_test.go) were red
  before correction: all four actions reported success against ordinary and
  case-folded ID-shaped impostors (eight failures). They now refuse with no tree
  changes. Rename/alias reuse, missing IDs, collisions and valid writes are controls.
- Real edit forms and command-bar start/defer submissions retain row/generation and
  workspace envelopes before watcher refresh. Tests cover stale results after
  selection/reload, identical IDs in different spaces, read-only refusal and committed
  cleanup-conflict receipts without retry. A port double proves receipt propagation;
  it is not claimed as a real filesystem cleanup failure.
- [Portable Service tests](../../internal/core/selected_task_action_test.go) cover
  skipped resolution, opaque field-target identity, one clock and retry budget,
  shared field policy/coercion/unset, empty identity and invalid-date refusal.
- Several legacy lifecycle tests used a display slug as their selected key. They
  now use fixture-established IDs so completion/eligibility/impact policy remains
  exercised instead of failing early at identity validation.

Full `just test` (race), `just lint`, `just build`, `just docs-check`, planning lint
with links and diff hygiene pass. Owning type/service comments and both focused
architecture guides describe the implemented boundary. Criteria are checked from
owner evidence; independent review and follow-through are recorded below.

## Independent review

Prepared one [Claude implementation audit](../audits/6gm337aba0bh-2026-10-15-selected-tui-task-actions-implementation-claude.md)
against the dirty implementation snapshot. Both passes are complete; the source
helper confirmed its transfer receipt before owner triage. Claude reproduced no
product defect, killed reference-based reversions of all four commands, and
independently exercised UI submissions, real-store contention/committed receipts
and fresh-binary CLI repair compatibility.

Accepted L1 as a coverage gap, not a currently failing activation policy. Added
[real-route activation tests](../../internal/tui/selected_task_session_test.go):
an Atlas open completes after Esc and a task edit, action menu, confirmation or
defer prompt opens in the old checkout. Both checkouts deliberately contain the
same selected ID with matching list generations. Activation must close the old
intent and neither checkout may change, including after delayed confirmation keys.
Removing `m.edit.close()` kills the edit case; separately removing
`m.action.close()` kills menu/confirm/defer. Both safeguards are restored and the
four cases pass five race-enabled repetitions. Owning comments and focused guides
now distinguish unsubmitted forms from commands that already captured a Service.

Full race tests, lint, build and docs-check pass after follow-through. The residual
question about preserving committed warnings from stale results is folded into the
existing [navigation/feedback design task](6g8vxcnezktm-reassess-tui-navigation-and-information-architecture-at-current-scale.md),
without changing the current stale-result policy here.

## Neighboring exposures and disposition

- Confirmed real E/Y actions can accept an impostor through both fuzzy path lookups.
  Effects were constructed but never executed. Filed
  [exact selected path actions](6gm30jat0xsy-keep-selected-tui-path-actions-bound-to-exact-identities.md)
  and added it to the existing navigation Thread without an artificial dependency.
- Confirmed selected `moveAudit` closes an impostor after deletion. Folded that
  earlier-caller boundary into [non-task pinning](6gkt3fphx1ts-pin-non-task-mutation-retries-to-selected-source-identities.md)
  with a dedicated criterion; audit/epic mutation inventory stays there.
- Ordinary detail results already carry source identity and reject mismatches;
  do not claim a reproduced wrong-detail acceptance. Path resolution accompanying
  detail reads belongs in the local-action follow-up's inventory.

Temporary bug-expecting probes are retained at
`/private/tmp/taskflow-selected-action-neighbors.uYLhPU/selected_identity_neighbors_probe_test.go`,
not committed as passing safety coverage. No editor or clipboard side effects ran.
