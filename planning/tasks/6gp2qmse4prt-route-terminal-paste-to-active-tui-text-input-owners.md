---
schema: 1
id: 6gp2qmse4prt
status: ready-to-start
epic: 18-tui-bubble-tea-interactive-planning-browser
description: Repair terminal paste ownership in the inline editor, palette and command bar while preserving selected-source and delayed-input guards.
effort: 0.5-1 day
tier: 2
priority: medium
autonomy_level: 3
tags: [tui, authoring, input]
created: "2026-10-21"
---
# Route terminal paste to active TUI text-input owners

## Objective

Make terminal paste behave like text input in the existing editor, palette and
command bar without leaking to background surfaces or bypassing selected-source
mutation guards. This is adjacent authoring hardening, not additional entity-form
scope.

## Confirmed reproduction

Self-review of epic creation used `Model.Update(tea.PasteMsg{Content: ...})`
after actual key entry into each surface:

- `e`, Enter: the description textarea retains its old value instead of the paste.
- Ctrl+P: the palette query remains empty.
- `:`: the command-bar input remains empty.

No native clipboard executable was invoked. These are reducer/program event-path
reproductions, not claims about physical-terminal decoding. Inspect the embedded
configuration editor too; its outer routing may have the same issue, but that
surface has not yet been independently reproduced.

## Scope

- Inventory immediate `tea.PasteMsg` routing in Model, the ordered modal registry,
  inline editor, palette, command bar and configuration embedding.
- Give the first visible active text owner paste; unsupported modal modes consume
  it rather than leaking text or triggering hotkeys. Keep existing no-modal list
  search semantics where intentional.
- Keep terminal paste distinct from delayed native-clipboard messages. Preserve
  session, selected-record/list generation and widget ownership checks on delayed
  deliveries; do not broadcast messages to every text widget.
- Input only stages text. Editing, command execution and persistence still require
  their existing explicit actions and guarded services in tea.Cmd.
- Reuse a bounded routing mechanism only where the actual consumers justify it.
  Do not turn this into a global widget registry or entity-form schema.

## Acceptance criteria

- [ ] Editor description/text fields, palette query and command bar accept terminal paste through their production opening keys and Model.Update; configuration embedding is checked and fixed or explicitly documented if unaffected.
- [ ] Paste stays with the visible text owner; help, enum, confirmation, hidden/narrow and inactive surfaces neither edit background drafts nor execute pasted hotkeys.
- [ ] Selected-source/session/list guards and delayed clipboard ownership remain intact; staging does not write, and intended submit executes at most once through the existing service route.
- [ ] Focused program/reducer regressions cover each confirmed surface, hostile owner changes and relevant refresh/resize cases; owning input-routing guidance is updated without promising unsupported paste behavior.

## Sequencing

Independent member of the TUI routine-authoring Thread. No dependency on epic
creation: the reproduction exists on the older editing/launcher surfaces.

## Starting points

- `internal/tui/model.go`: immediate paste routing.
- `internal/tui/overlay.go`: modal precedence and optional paste handler.
- `internal/tui/edit.go`, `command_dispatch.go`, `palette.go` and command bar.
- `internal/tui/staged_creation_view.go`: visible-owner policy for the two new
  creation forms; use as a comparison, not permission to bypass other guards.
- `docs/architecture/tui.md`: input ownership and selected-action contracts.
