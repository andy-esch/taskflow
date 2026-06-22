---
schema: 1
status: completed
epic: 18-tui-bubble-tea-interactive-planning-browser
description: Inline-edit task fields (description/priority/tags/effort/tier) in the TUI via SetFields — typed widgets, no $EDITOR escape
effort: Unknown
tier: 3
priority: medium
autonomy_level: 3
tags: [tui, ux]
created: "2026-06-21"
updated_at: "2026-06-22"
completed_at: "2026-06-22"
---
## Objective

Let the TUI edit a task's **fields in place**, without leaving for `$EDITOR`.
Today the TUI is read + lifecycle-move only: the `a` action menu applies status
transitions via `core.Service.Move`, but there is **no field editing and no
`$EDITOR` launch at all**. This adds the "agent face" (`task set`) as a
human-driven inline surface — "`task set` with a GUI."

## Scope — typed field edits via `Service.SetFields`

Inline-edit a small set of **known, typed** fields, each with a focused widget,
validated by core on submit (then reload):

- `description` → single-line text input
- `priority` → enum picker (high/medium/low)
- `tags` → multi-value input
- `effort`, `tier` → small pickers

Mechanism: open the editor widget over the detail/list pane, on submit fire a
`tea.Cmd` calling `core.Service.SetFields` (the same call `task set` uses) →
reload. No new core surface; reuses the S4 mutation plumbing.

## Architecture (non-negotiable — same as the rest of the TUI)

- Goes through `core.Service.SetFields` — **never `store`/the fs**. The write is
  a `tea.Cmd` returning a `tea.Msg`; **no I/O in `Update`/`View`**.
- Core re-validates on `SetFields` (enums, key-order, surgical frontmatter), so
  typed widgets + core validation keep the file safe — the human TUI is the
  third mutation face but adds **no** new validation path.
- Human-only surface → zero agent/pipeline-contract risk.

## Out of scope (deliberate)

- **No raw-frontmatter text editor.** Editing fields, not YAML — a raw box would
  bypass the enum/key-order/validation discipline the repo is built on. If true
  freeform is ever wanted, an `$EDITOR` escape (`task edit` equivalent) is the
  right tool for *that*, as a separate power-key — not this task's primary path.
- **Status is not a SetFields field.** Status == directory; it stays in the `a`
  action menu (`Move`), not inline edit.

## Open questions / decisions

- [x] Final editable field set (default: description / priority / tags / effort /
      tier). Trim or extend?
- [x] Keybinding + UX: a dedicated `e` (edit) that opens a field picker, or
      per-field keys? Modal like the `a` menu / `?` help / `:` command bar.
- [x] Widget per field type (text vs enum-select vs tag-multi) and where the
      editor floats (over detail vs inline in the row).
- [x] Optional follow-on: an `$EDITOR` escape for the body (separate key), if the
      inline typed fields prove insufficient for freeform notes.

## Acceptance criteria

- [x] From the TUI, edit description/priority/tags (+ effort/tier) on the focused
      task; the change persists via `Service.SetFields` and the view reloads.
- [x] Invalid input is rejected with the core validation error surfaced in the
      TUI (no partial/corrupt write); cancel (`Esc`) is a no-op.
- [x] All writes run as `tea.Cmd`s — no I/O in `Update`/`View`; store untouched.
- [x] Message-injection unit tests for the edit flow + a `teatest` golden for the
      editor layout; suite + lint green; help footer + key matrix updated.

## Outcome (2026-06-22)

Shipped as a new modal in the M14 overlay registry (`internal/tui/edit.go` +
`editModal` in `overlay.go`). **Decisions:** field set = description / priority /
tags / effort / tier (the default); a dedicated `e` opens a two-phase modal — a
field picker, then a per-field widget (text input for description/tags/effort,
enum cursor for priority/tier) floated over the body like the `a` menu; submit
fires `Service.SetFields` (force=false, dryRun=false) as a `tea.Cmd` → `editedMsg`
→ flash + reload; invalid input comes back as `actionErrMsg` (red flash, no write);
Esc cancels. Tags ride as a comma-list and tier as a string — the SetFields
coercion turns them into a YAML list / int, the same path `task set` uses. Task-only
(SetFields has no epic/audit path); status stays in `a`. The `$EDITOR`-escape
follow-on is left out of scope as planned.

**Testing note:** used the codebase's established message-injection + `View()`
substring assertions (e.g. `TestModel_EditMenuComposites`, and the
`TestModel_ViewFitsTerminal` layout invariant covers the overlay) rather than
adding a `teatest` dependency the repo doesn't otherwise use — same coverage intent,
no new dep.

## Related

- Epic [[18-tui-bubble-tea-interactive-planning-browser]] (extends S4 mutations
  [[tui-sprint-4-mutations-and-actions]]).
- Mutation faces: `task set` (agent/field-level) vs `task edit` (`$EDITOR`/human)
  — this is the TUI's in-place field face over the same `SetFields`.
- `internal/tui/action.go` (the `a` lifecycle menu this sits beside);
  `internal/core/service.go` (`SetFields`).
