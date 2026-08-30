---
schema: 1
id: 6g4ym5y2dkp8
status: ready-to-start
epic: 18-tui-bubble-tea-interactive-planning-browser
description: Add a create-entity flow to each TUI tab, routed through the existing core New* use cases
effort: 3-5 days
tier: 3
priority: medium
autonomy_level: 3
tags: [entity]
created: "2026-08-29"
updated_at: "2026-08-29"
---
# create entities from TUI

## Objective

The TUI browses and mutates existing entities but cannot create one — every `new` lives on
the CLI. Add a create flow to each entity tab, with the epic page pre-filling the epic on a
new task (the single biggest ergonomic win, because `epic` is a *required* field with no
default).

## What already exists

`core.Service` already owns all five create use cases — `NewTask`, `NewEpic`, `NewAudit`,
`NewResearch`, `NewThread` — each validated, atomic, and adapter-neutral, and each is what
the CLI calls. **No core, store, or domain work is needed.** This is a pure primary-adapter
task: a form overlay plus a `tea.Cmd` per kind.

## What does NOT already exist (the real cost)

`editMenu` (`internal/tui/edit.go`) looks reusable but is not, in its current shape. It is a
**per-field setter against an entity that already exists**: `fieldSetter` is
`func(svc, slug, key, value) tea.Cmd`, and each `enter` fires one `SetFields` write for one
field. Creation is the opposite shape — collect every required field first, validate
nothing until the end, then perform exactly one atomic create.

So the form needs a second mode (collect-then-submit) alongside the existing
edit-in-place mode, or a sibling overlay. Deciding that is the first piece of work, not an
implementation detail. Reusing the widgets (`textinput`, `textarea`, enum cursor), the
overlay/clamp helpers, and the error-on-field rendering is straightforward either way.

## Minimum required input per kind (verified empirically, not from the descriptor)

| kind | user must supply | defaulted by the use case |
| :--- | :--- | :--- |
| audit | area | date → today, bucket → open, template → default |
| research | title | created → today (and mints the id) |
| epic | title, description | status → active, NN prefix |
| thread | title, description, goal | status → always `unstarted` |
| task | title, **epic**, **≥1 tag** | status → ready-to-start, priority, tier, autonomy |

**Do not drive the form off `domain.FieldDoc.Required`.** That flag means "required in the
persisted document", not "the user must supply it at creation", and the two disagree for
three of the five kinds: it marks `status` (epic), `date` (audit), and `created` (research)
required even though creation defaults all three, and it does not mark `title` at all
because the title is a positional argument that mints the slug and id rather than a
frontmatter field. If a registry-driven generic form is wanted, the descriptor needs a
distinct required-at-creation notion first; otherwise write per-kind field sets the way
`editableFields` / `editableEpicFields` already do.

## Open decisions

- **One generic create form or per-kind forms?** Per-kind is honest to the table above and
  matches the existing edit code. Generic needs the descriptor change above. Recommend
  per-kind now, and extract a generic form only if a fifth kind makes it pay.
- **Keybinding.** `n` is taken (find-next) and `N` (find-prev). `c` is free and reads as
  "create". Whatever is chosen must be added to `keys.go` so the `?` overlay picks it up
  automatically (`helpSections` derives from the keyMap).
- **Body and template.** `NewTaskParams` takes `Body` and `Template`. Simplest v1: always
  use the kind's default scaffold and let the user press `E` (`$EDITOR`) afterwards.
  A template picker is a plausible follow-up, not v1.
- **Post-create behaviour.** Reload the tab's list and move the cursor to the new entity.
  Note the watcher (`WatchPaths` covers every entity dir) will *also* fire a reload for the
  same write, so the explicit reload and the debounced fs reload must not fight over the
  cursor — the same hazard the existing `editedMsg` → flash → reload path already handles.
- **Does create need a confirm step?** Lifecycle uses `dangerBorder` + y/n for destructive
  moves. Creation is additive and reversible by deleting a file; recommend no confirm, just
  the flash.

## Acceptance criteria

- [ ] Create a new task from the tasks tab.
- [ ] Create a new task from an epic's page, with that epic pre-filled and not re-asked.
- [ ] Create a new audit from the audits tab.
- [ ] Create a new epic from the epics tab.
- [ ] Create a new research doc from the research tab.
- [ ] Every create routes through the existing `core.Service` `New*` use case — no new validation path, no store access from the TUI, all I/O in a `tea.Cmd`.
- [ ] A core validation error (missing epic, no tags, description too long) is shown on the offending field and the form stays open with the entered values intact.
- [ ] After a successful create the list reloads, the cursor lands on the new entity, and the fs-watcher reload does not move it again.
- [ ] The create key is declared in `keys.go` so the `?` overlay documents it without a separate edit.

## Out of scope

- **Threads.** `entityKind` has exactly four values (tasks, epics, audits, research); there
  is no Threads tab, so "create a Thread from the TUI" is blocked on the TUI gaining one.
  Track it with the Threads TUI work, not here.
- Creating a task directly into `next-up` / `in-progress` (`--next` / `--start`). Create in
  the default state, then use `m`.
- A dry-run preview. The CLI's `--dry-run` has no TUI analogue and creation is cheap to undo.
- Body/template selection at create time (see Open decisions).

## Related

- Epic [18-tui-bubble-tea-interactive-planning-browser](../epics/18-tui-bubble-tea-interactive-planning-browser.md)
- Reuses the overlay/form machinery from `internal/tui/edit.go` and `internal/tui/action.go`.
