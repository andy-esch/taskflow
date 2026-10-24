---
schema: 1
id: 6g3cp32y9149
status: ready-to-start
epic: 20-cli-ux-and-ergonomics
description: Edit selected criterion states and core-required reason or handoff inputs from task detail, with explicit authoring scope and stale-child protection.
effort: 1-2 days
tier: 2
priority: medium
autonomy_level: 3
tags: [tui, domain, ux]
created: "2026-08-24"
updated_at: "2026-10-24"
audited: "2026-10-04"
depends_on: [6gkkzvr92waq, 6gmshzw7we0m, 6gpxh6ncaqpw]
---
# Let the TUI flip an acceptance criterion

## Objective

Make the criterion roll-up actionable: select a task criterion, inspect its text
and current disposition, then change state without leaving task detail.

## Scope and design checkpoint

State changes are the baseline. Before implementation, settle whether simple
add/reword/remove authoring belongs in this slice; if not, explicitly track that
work. The owner approved routine TUI authoring generally, not this finer boundary.
On 2026-10-18 the owner explicitly left this decision for later. Revisit it at
this task's kickoff; it does not hold the shared creation checkpoint open.
The existing [tracked write/lint policy alignment](6gkkzvr92waq-align-tracked-criterion-writes-and-lint-reason-semantics.md)
is a separate prerequisite. Reuse its owner-approved policy rather than choosing
different reason requirements in the form.
The [criterion-section shadowing repair](6gpxh6ncaqpw-prevent-narrative-headings-from-hiding-task-criteria.md)
is also a prerequisite: narrative headings must not hide the real criterion list
before this UI starts enumerating and acting on it. It is independent of the
tracked-reason policy decision.

Use the core criterion policy and its required reason or destination inputs rather
than copying a state/reason matrix into the TUI. Long prose remains in $EDITOR.
Coordinate with the [authoring interaction contract](6gmshzw7we0m-define-the-tui-authoring-and-workspace-action-contract.md)
for selection, state choice and input placement.

## Implementation constraints

`SetCriterionState` and `EditCriteria` are existing core write paths, not proof that
a selected-child mutation is already safe. Bind the displayed task and criterion
intent across body edits and reordered criteria; a stale list index or fresh
reference lookup must not target a different item. Add a bounded core seam if
needed, and perform I/O only in tea.Cmd.

Use the existing criterion marks and refresh the roll-up after writes. Preserve
input on refusal and present committed/partial recovery truthfully. Do not change
criterion policy merely to simplify the form.

## Acceptance criteria

- [ ] Record the owner's decision on state-only v1 versus simple
  add/reword/remove authoring before implementation; explicitly track excluded
  authoring scope rather than silently dropping it.
- [ ] A criterion's state can be changed from the task detail view without leaving the TUI,
      including the states that require a reason.
- [ ] The write goes through `core.Service`, as a `tea.Cmd` — no store access and no I/O in
      `Update`/`View`.
- [ ] The roll-up in the detail header updates after the write, without a manual refresh.
- [ ] A refused write (a missing reason, an out-of-range index) surfaces as an error flash
      rather than a silent no-op, the way the `task complete` gate's refusal already does.
- [ ] The help legend covers whatever keys this adds.
- [ ] Suite + lint green; no new glyph or colour vocabulary — `theme.CriterionState` already
      owns the marks.
- [ ] Bind the selected task and displayed criterion intent across body edits,
  reorder, reload and workspace changes; do not apply a stale criterion index or
  restart reference lookup against a different source.

## Out of scope

A general Markdown editor, cross-space batch changes, or a new criterion policy.
Simple criterion authoring remains an explicit design decision, not a silent
promise or exclusion.

## Related

- [Authoring Thread](../threads/6gmshnmnew7p-enable-routine-planning-and-workspace-management-in-the-tui.md).
- [Core criterion operations](../../internal/core/service_task.go).
- [Criterion parsing and tally](../../internal/domain/body.go).
- [TUI detail and roll-up](../../internal/tui/detail.go).
- [TUI interaction and refresh contracts](../../docs/architecture/tui.md).
