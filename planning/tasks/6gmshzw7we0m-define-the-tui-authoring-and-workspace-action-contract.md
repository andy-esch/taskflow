---
schema: 1
id: 6gmshzw7we0m
status: completed
epic: 18-tui-bubble-tea-interactive-planning-browser
description: Turn the agreed capability boundary into a small, implementable interaction contract and feedback prototype, not a whole-TUI redesign.
effort: half day
tier: 2
priority: high
autonomy_level: 2
tags: [tui, authoring, ux]
created: "2026-10-17"
started_at: "2026-10-18"
updated_at: "2026-10-18"
completed_at: "2026-10-18"
---
# Define the TUI authoring and workspace action contract

## Objective

Turn the agreed capability boundary into a small, implementable interaction contract and feedback prototype, not a whole-TUI redesign.

## Scope

Inspect existing action/help, per-field edit, overlay, selected-identity and workspace-session seams. Specify staged creation versus existing field-save semantics, contextual action discovery, target labels, validation, cancel/dirty input, narrow terminals, async recovery and watcher focus. Reuse the existing configuration editor and application services.

## Acceptance criteria

- [x] Record the native/editor-assisted/CLI-first boundary in the owning Thread and reconcile existing TUI behavior without claiming planned capabilities are shipped.
- [x] Produce a compact interaction sketch or disposable prototype for task creation and Atlas register/forget; obtain owner feedback before freezing affected interactions.
- [x] Decide action discovery, form submission/cancellation and workspace-switch behavior, including forgetting one entry in a grouped space or the active entry.
- [x] Record owner decisions or explicit scoped deferrals, and identify
  selected-child/selected-record API gaps requiring bounded core extensions; the
  owner-deferred criterion authoring choice belongs to the criteria task before
  its implementation.
- [x] Give implementation slices shared conventions and focused test scenarios without requiring a generic form schema or completion of the broader navigation redesign.

## Design attention

Ask about unsettled product choices at this checkpoint. Persisted descriptor Required flags do not define required user input. Do not turn open layout details into hard-coded keybinding promises.

## Out of scope

Production CRUD implementation, a universal mutation framework, redesigning every tab, and researching/replacing the TUI toolkit.

## Related

- [Interaction proposal, alternatives and owner decisions](../research/6gmrkq00348a-tui-authoring-interaction-checkpoint.md).
- [Authoring capability boundary and sequencing](../threads/6gmshnmnew7p-enable-routine-planning-and-workspace-management-in-the-tui.md).
- [Owning edit.go](../../internal/tui/edit.go).
- [Owning help.go](../../internal/tui/help.go).
- [Owning model.go](../../internal/tui/model.go).
- [Owning tui.md](../../docs/architecture/tui.md).

## Checkpoint progress (2026-10-18)

Drafted a direct staged form versus field-list sketch, Atlas register/forget flows,
shared interaction conventions, selected-target API gaps and focused regressions.
Owner reviewed the plain-text fallback sketches and approved page-local staged
creation with `c`, preserving `n`/`N` search, and machine-wide registration-only
Atlas removal. Shared UI mechanics do not imply a universal cross-entity wizard or
a new generic mutation API. “New task in this epic” remains a separately named action.

At the owner's request, defer criterion add/reword/remove scope to the
[criteria task](6g3cp32y9149-let-the-tui-flip-an-acceptance-criterion.md) before its
implementation; do not hold unrelated creation work behind that choice.
The research records the adapter-safe gate and portable/real-adapter regression
requirements. No production implementation is claimed by this checkpoint.
