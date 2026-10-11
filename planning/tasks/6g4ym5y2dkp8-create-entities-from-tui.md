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
updated_at: "2026-10-11"
audited: "2026-10-11"
audit_sources: [2026-10-11-weekly-task-sweep]
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

## Sweep refresh (2026-10-11)

2026-10-11: automated weekly sweep — re-verified every citation; two have drifted, the
Threads one removes an "Out of scope" premise. Scope deliberately left unchanged.

**Drift found — the Threads exclusion no longer holds.** The "Out of scope" note above
says `entityKind` has exactly four values and there is no Threads tab. Both are now false:
`entityKind` has five values (`internal/tui/entity.go:24-28`, `entityThreads` third) and a
real Threads tab is registered in `newEntityTabs` with `name: "threads"` and aliases
`th`/`thread` (`internal/tui/entity.go:463`). The stated blocker for "create a Thread from
the TUI" is therefore gone, and `core.Service.NewThread` exists
(`internal/core/service_thread.go:26`) with the title/description/goal input the table
above already records. Whether a Threads create flow belongs in THIS task or stays with
the Threads TUI work is a scoping call left to the human — the sweep does not rescope.

**Drift found — `fieldSetter`'s signature moved.** The body cites it as
`func(svc, slug, key, value) tea.Cmd`. It is now
`func(svc *core.Service, ref entityRef, key, value string) tea.Cmd`
(`internal/tui/edit.go:136`): `slug` became an `entityRef` carrying durable `key` plus
display `label`. This does not change the task's argument — `editMenu` is still a
per-field setter against an existing entity and still fires one `SetFields` write per
`enter` — but any create form reusing the seam takes a ref, not a slug.

**Verified accurate (2026-10-11):**
- All five `core.Service` `New*` use cases exist and no store access from the TUI would be
  needed: `NewTask` (`service_task.go:785`), `NewEpic` (`service_epic.go:25`), `NewAudit`
  (`service_audit.go:24`), `NewResearch` (`service_research.go:34`), `NewThread`
  (`service_thread.go:26`).
- `editableFields` / `editableEpicFields` are still the per-kind field sets
  (`internal/tui/edit.go:57,87`), as the body recommends imitating.
- `c` is still unbound. The bound keys today are
  `/ : ? E F G N O R S Y [ ] a ctrl+c ctrl+o ctrl+p e enter esc f g h l m n o q r s tab v y z`
  (`internal/tui/keys.go`), and `n`/`N` are still find-next/find-prev (`keys.go:57-58`).
- `helpSections` still derives the `?` overlay from the keyMap (`internal/tui/help.go:27`),
  so the last acceptance criterion still holds as written.
- `domain.FieldDoc.Required` still exists (`internal/domain/schema.go:47`), so the
  "do not drive the form off it" warning still has a live target.
- `WatchPaths` still covers every entity dir (`internal/store/fsstore.go:122`, consumed at
  `internal/tui/tui.go:47`), so the cursor-fight hazard in Open decisions is still real.
- `dangerBorder` is still the destructive-confirm idiom (`internal/tui/action.go:210`).
- `NewTaskParams` still takes `Body` and `Template`, mutually exclusive
  (`internal/core/service_task.go:766-786`).

**Acceptance criteria untouched.** No create flow exists anywhere in `internal/tui` — no
create overlay, and no `core.Service` `New*` call from the TUI at all — so no criterion is
met and no box was ticked.
