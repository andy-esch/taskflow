---
schema: 1
id: 6gmshzxvgagj
status: ready-to-start
epic: 18-tui-bubble-tea-interactive-planning-browser
description: Let a human reviewing an audit record a finding disposition and any required reason or destination task without dropping to CLI or hand-editing managed sections.
effort: 1-2 days
tier: 2
priority: medium
autonomy_level: 3
tags: [tui, authoring, ux]
created: "2026-10-17"
depends_on: [6gmshzw7we0m]
updated_at: "2026-10-17"
---
# Manage audit finding dispositions from the TUI

## Objective

Let a human reviewing an audit record a finding disposition and any required reason or destination task without dropping to CLI or hand-editing managed sections.

## Scope

Expose structured finding selection and the existing canonical status/disposition operations. Keep finding prose authoring in the external editor and preserve the audit's managed candidate rows and settlement rules.

## Acceptance criteria

- [ ] Select a concrete finding and apply an existing valid disposition; show its code/title, current state and selected audit context.
- [ ] Collect required resolution/reason/destination-task inputs and route writes through existing core finding/settlement policy rather than raw Markdown manipulation.
- [ ] Tracked findings retain a verified destination task and managed candidate projection; partial errors and terminal-audit refusals are explicit.
- [ ] A refreshed/reordered or renamed finding/audit cannot redirect a pending action; bind the displayed child intent and originating Service and refuse stale targets.
- [ ] Updated audit progress/selection survives watching, and focused selected-target/validation tests, help/docs and normal gates pass.

## Design attention

Review the structured-child action contract with the design task. A finding code or array index alone may not prove the same displayed finding still exists.

## Out of scope

New finding grammar, audit policy redesign, automatic task creation, bulk settlement, or a full finding-body editor.

## Related

- [Authoring capability boundary and sequencing](../threads/6gmshnmnew7p-enable-routine-planning-and-workspace-management-in-the-tui.md).
- [Owning finding.go](../../internal/core/finding.go).
- [Owning detail.go](../../internal/tui/detail.go).
