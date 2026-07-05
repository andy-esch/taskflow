---
status: completed
epic: 18-tui-bubble-tea-interactive-planning-browser
description: Follow structured references (epic to its tasks, task to its epic) with a jump + back-stack; defer body wikilinks and peek-overlay
effort: Unknown
tier: 3
priority: medium
autonomy_level: 3
tags: [tui, bubble-tea]
created: "2026-06-11"
started_at: "2026-06-13"
updated_at: "2026-06-13"
completed_at: "2026-06-13"
id: 6fb7ym40100g
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

- [x] From an epic, follow a reference to one of its tasks and land on that task's
      detail (correct tab + selection); from a task, follow `epic:` to its epic.
- [x] A back key returns to the prior view with the previous cursor restored;
      the stack handles multiple hops.
- [x] A missing/ambiguous target is handled gracefully (no crash; a clear message).
      Suite + lint green.

## Out of scope

- Body wikilink parsing and the peek-overlay (deferred options above).
- Cross-linking to projects/ADRs/research (those entities don't exist yet).

## Related

- Epic [[18-tui-bubble-tea-interactive-planning-browser]]
- Reuses the `overlay()` helper from the S2b `?` modal (if the peek option lands)

## Closure (2026-06-13)

Shipped as designed (structured references + jump/back-stack; wikilinks and
peek-overlay stay deferred). Implementation notes:
- **`f` follows, `ctrl+o` pops** (vim jumplist). On the tasks tab `f` jumps
  straight to the selected task's `epic:`; on the epics tab it opens a
  reference picker over the epic's tasks (`nav.go: followMenu`, modeled on
  the S4 action menu — modal, j/k/⏎/esc, floated via `overlay()`). Audits
  flash "no linked entities".
- **Back-stack** (`navStack []navLoc` on the Model) handles multiple hops;
  the footer shows a breadcrumb (`↩ ctrl+o <id> (n)`) while non-empty.
- **Hidden targets handled, not failed:** a jump clears any applied filter,
  and a task hidden by the current status view (e.g. an epic's completed
  task vs the working set) escalates the tasks tab to `:all` and reloads
  with the cursor restore pending — with a flash explaining the view change.
  Missing targets flash; nothing crashes.
- **Picker data source:** the epic's task list rides in the already-loaded
  (stale-guarded) `epicDetail` content — no new service calls; an ID
  mismatch flashes "references still loading…".
- Tradeoff recorded in-code: `f` shadows the undocumented f-paging alias in
  list/viewport (d/u and ctrl+d/u remain the documented paging keys).
- Tests: `nav_test.go` — task→epic→back round-trip with breadcrumb, epic→task
  via picker + two-hop unwind, modal q-safety, esc-cancel, no-epic/audit
  dead-ends, and the :all escalation. Seed tasks now carry `epic: 01-test`.
  Suite, vet, golangci-lint, gofmt all green.

### Addendum (same day, post-closure sweep)

A "what else is in scope?" pass added: README + ARCHITECTURE.md updated
(`f`/`ctrl+o` in the key list; `nav.go` in the file map); three more tests —
jump-clears-applied-filter, the picker layout invariant at small sizes, and
**a dangling `epic:` reference** (the historical-B1 data case). The last one
caught a real gap: the not-found flash only fired when the target tab was
already loaded — a dangling jump to an unloaded tab silently left a stale
pending restore. Fixed in `handleListLoaded`: an unsatisfied restore on a
fully-visible (unfiltered) load now flashes "<id> not found" on the active
tab and clears; filtered tabs still keep the restore pending for the async
refilter. This also gives feedback when the selected item is deleted
externally mid-reload.
