---
schema: 1
id: 6gkt3fq0pv26
status: ready-to-start
epic: 21-code-quality-architecture-hardening
description: Carry selected task identities through queued TUI mutations without restarting fuzzy reference selection in Service.
effort: 0.5-1 day
tier: 2
priority: high
autonomy_level: 2
tags: [correctness, identity, tui, core]
created: "2026-10-14"
depends_on: [6gkmexz4vh23]
updated_at: "2026-10-14"
audit_sources: [2026-10-14-task-correctness-detour-implementation-claude]
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

- [ ] Selected task field/lifecycle actions follow renames but refuse deletion, duplicate IDs and ID-shaped/case-folded impostor slugs without mutating siblings.
- [ ] Queued TUI command regressions exercise the actual action entry points before watcher refresh, including set/unset, move/defer and source/session changes.
- [ ] Initial CLI fuzzy/ID-shaped human references still work, while explicit selected-identity actions never reselect a different record.
- [ ] Existing authorization, lifecycle policy/receipts, stale-message handling, warning/reload and tea.Cmd boundaries remain covered.
- [ ] Owning source comments/TUI guide describe the distinction, and normal validation passes.

## Sequencing and ownership

Depends on [strict task mutation targets](6gkmexz4vh23-bind-task-mutation-retries-to-strict-source-identities.md),
whose narrow internal target type and lifecycle retry rule this service boundary extends.
Belongs to the existing Refine Thread and TUI navigation Thread: selected action targets
are part of its shared navigation contract, not a reason to reopen the completed adapter refactor.
