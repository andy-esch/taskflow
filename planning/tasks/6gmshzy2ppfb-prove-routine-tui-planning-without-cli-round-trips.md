---
schema: 1
id: 6gmshzy2ppfb
status: ready-to-start
epic: 18-tui-bubble-tea-interactive-planning-browser
description: Close the authoring Thread with a bounded, repeatable human workflow and accurate capability documentation rather than a collection of individually green forms.
effort: half day
tier: 2
priority: medium
autonomy_level: 3
tags: [tui, authoring, ux]
created: "2026-10-17"
depends_on: [6ffdv9g01q40, 6g321mybqvfz, 6g3cp32y9149, 6g4ym5y2dkp8, 6gmbfxy96tpw, 6gmshzwfkj62, 6gmshzwphadw, 6gmshzwxxce3, 6gmshzx55n1y, 6gmshzxchdzc, 6gmshzxksjcn, 6gmshzxvgagj]
updated_at: "2026-10-18"
---
# Prove routine TUI planning without CLI round trips

## Objective

Close the authoring Thread with a bounded, repeatable human workflow and accurate capability documentation rather than a collection of individually green forms.

## Scope

Use a disposable pre-initialized planning setup with multiple entry points and unrelated data. Exercise the finished native actions plus external-editor handoff; CLI may verify results but must not be needed to perform the routine workflow.

## Acceptance criteria

- [ ] Create an epic and tasks, edit metadata/title, record criteria, create other entity kinds, manage Thread membership/dependencies/lifecycle and dispose findings through the TUI.
- [ ] Register an existing entry point and forget one registration in Atlas; prove planning files, other registrations and unrelated checkouts remain intact.
- [ ] Exercise long-body and unreadable-record editor handoff and visibly scoped preferences, including refresh, cancellation, narrow terminal and workspace-switch cases.
- [ ] Document a concise repeatable playbook and the native/editor-assisted/CLI-first boundary in owning help/architecture guidance; remove obsolete limitations.
- [ ] Obtain owner feedback on a usable build, record accepted limitations honestly, and close only after all scoped capabilities and normal validation pass.

## Design attention

Initialize/seed fixtures before the walkthrough; bootstrap is intentionally CLI-first. Verification commands are evidence, not hidden steps required to achieve a TUI action.
The existing [watcher/session hardening](6gmbfxy96tpw-bind-tui-watcher-health-and-reload-messages-to-their-originating-session.md)
is a closeout prerequisite so the multi-workspace run does not knowingly inherit
stale callback/reload hazards. Core selected-epic and registry foundations arrive
transitively through their UI consumers.

## Out of scope

A new release, fixture/GIF restyling, full CLI parity, benchmarks or universal end-to-end UI automation.

## Related

- [Authoring capability boundary and sequencing](../threads/6gmshnmnew7p-enable-routine-planning-and-workspace-management-in-the-tui.md).
- [Owning tui.md](../../docs/architecture/tui.md).
- [Owning README.md](../../assets/README.md).
