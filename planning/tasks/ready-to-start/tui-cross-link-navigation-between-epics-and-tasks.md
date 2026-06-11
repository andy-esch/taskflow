---
status: ready-to-start
epic: 18-tui-bubble-tea-interactive-planning-browser
description: Follow structured references (epic to its tasks, task to its epic) with a jump + back-stack; defer body wikilinks and peek-overlay
effort: Unknown
tier: 3
priority: medium
autonomy_level: 3
tags: [tui, bubble-tea]
created: "2026-06-11"
---

# TUI: cross-link navigation between epics and tasks

## Objective

Make references navigable: from an epic's detail, jump to one of its tasks; from a
task, jump to its epic — then jump back. Turns the browser from three independent
lists into a linked graph you can walk, k9s/vim-style.

## Approach (start structured + jump/back-stack)

- **Structured references only, to start** — the data is already loaded: an epic's
  detail lists its task slugs; a task has an `epic:` field. No body parsing needed.
- **Jump + back-stack:** a follow key (e.g. `Enter` on a highlighted reference, or
  `gd`) pushes the current `(entity, id)` onto a nav stack and switches the active
  tab + selection to the target. A back key (`Ctrl+o`, vim-style) pops it. Reuses
  the entity registry + `selectByID` + the per-tab cursor that already exist.
- **Making references selectable:** the epic detail currently renders task slugs as
  plain lines. Give the detail pane a notion of "focusable references" (the epic's
  task list; the task's epic field) with a cursor, so the user can pick which to
  follow. Simplest first cut: `Enter` on a task-tab item follows its `epic:`; on an
  epics-tab item, a small selectable sub-list of its tasks.
- Show a breadcrumb/back hint in the footer when the nav stack is non-empty.

## Deferred options (note, don't build yet)

- **Body `[[wikilink]]` following** — parse `[[slug]]` in rendered bodies and make
  them followable. More work (parse + resolve + ambiguous-slug handling); the
  planning repo uses this syntax, so it's valuable later.
- **Window-in-window peek** — open the target in a transient overlay (reuse the
  `overlay()` compositor from the `?` help modal) for a glance without leaving the
  current view, as an alternative/complement to full jump.

## Acceptance criteria

- [ ] From an epic, follow a reference to one of its tasks and land on that task's
      detail (correct tab + selection); from a task, follow `epic:` to its epic.
- [ ] A back key returns to the prior view with the previous cursor restored;
      the stack handles multiple hops.
- [ ] A missing/ambiguous target is handled gracefully (no crash; a clear message).
      Suite + lint green.

## Out of scope

- Body wikilink parsing and the peek-overlay (deferred options above).
- Cross-linking to projects/ADRs/research (those entities don't exist yet).

## Related

- Epic [[18-tui-bubble-tea-interactive-planning-browser]]
- Reuses the `overlay()` helper from the S2b `?` modal (if the peek option lands)
