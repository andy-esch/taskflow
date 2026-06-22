---
schema: 1
status: ready-to-start
epic: 21-code-quality-architecture-hardening
description: The a-menu, :verbs, transitions and follow are task-only; lift them onto each entity's declared table so audits/epics get lifecycle.
effort: Unknown
tier: 3
priority: medium
autonomy_level: 3
tags: [tui, architecture]
created: "2026-06-22"
---
# Make TUI lifecycle action machinery registry-driven

## Objective

The entity registry covers read/browse, but the action/transition/follow machinery
is task-only (selectedTask(), task-only transitionFor/applyTransition/svc.Move,
followSelected switch). Audits already have CLI close/reopen/defer with zero TUI
path. Lift the action/transition table onto each entity's declared transitions so
the `a` menu and `:` verbs are registry-driven.

## Audit reference

planning/audits/open/2026-06-22-code-quality-architecture.md — **M10**. Either make it registry-driven OR scope the doc's "no new
keybindings" promise to read-only browse. Relates to epic 18 (TUI).

## Acceptance criteria

- [ ] Lifecycle actions work for any entity that declares transitions (no per-entity edits in model.go/action.go/nav.go).
- [ ] ARCHITECTURE.md / entity.go comment corrected.
- [ ] just test + just lint green.

## Implementation plan

**Approach.** Make the `a` menu and `:` verbs registry-driven by hanging a *transition
table + a move function* off each `entityTab`, and actually wire audits up (they have
CLI close/reopen/defer but zero TUI path today). The cleanest fit for this codebase:
add `transitions []transition` and `applyMove func(svc, id, to) error` to `entityTab`,
and generalize the task-only plumbing in model.go to read them off `m.cur()` instead
of calling `selectedTask()`/`svc.Move` directly. Audits use bucket transitions
(open/closed/deferred → `svc.MoveAudit`); epics declare none (no-op, honestly). This
is preferable to the doc-only escape hatch the audit offers as the alternative,
because audit lifecycle in the TUI is a real gap, not hypothetical — and the registry
framing in entity.go already promises it.

**Steps.**
1. **Generalize the transition type (action.go).** Today `transition{verb, to
   domain.Status, destructive}` is task-specific (`to` is a `domain.Status`). Make
   the destination a string the `applyMove` closure interprets, or keep two small
   tables (task statuses vs audit buckets) and store the per-entity `[]transition` on
   the tab. Keep `transitionFor`/`transitionVerbs`/`validTransitions` but make them
   operate on a passed `[]transition` (or the active tab's) rather than the package
   global. The audit transitions are: `close → closed`, `reopen → open`,
   `defer → deferred` (mirror `audit close/reopen/defer`); `close`/`defer` to a
   non-open bucket are the ones the store already guards on open findings (M4) — the
   error will surface as `actionErrMsg`, which is correct.
2. **entityTab fields (entity.go).** Add `transitions []transition` and
   `applyMove func(svc *core.Service, id string, tr transition) tea.Cmd` to the
   struct; populate in `newEntityTabs`: tasks → `transitions` + a closure over
   `svc.Move`; audits → audit transitions + `svc.MoveAudit`; epics → nil. Correct the
   struct/registry comment ("no new keybindings") to scope it honestly: read/browse is
   keybinding-free; lifecycle is declared per entity here.
3. **Model wiring (model.go).** In `handleKey`'s `keys.Action` case, replace
   `m.selectedTask()` with a generic `m.selectedID()` + a guard on
   `len(m.cur().transitions) > 0`; open the menu with the entity's transition set
   (`m.action.open(id, m.cur().transitions, cur)` — pass the current state so
   `validTransitions` drops the no-op row). In `dispatchCommand`, route a `:`verb
   through `m.cur().transitions`/the active tab's `transitionFor` and the tab's
   `applyMove`. Replace `applyTransition` (currently hard-wired to `svc.Move` returning
   a `domain.Status`) with a call to `m.cur().applyMove`. The `movedMsg`/`actionErrMsg`
   flow is already entity-agnostic (`movedMsg{slug, to}` — keep `to` a string).
4. **action.go menu.** `actionMenu.open`/`openConfirm` already take a `[]transition`
   via `validTransitions`; just feed the entity's table. The destructive-confirm path
   (deprecate / audit close-with-findings) is unchanged.
5. **followSelected (nav.go).** The `default: "no linked entities here"` branch is
   fine to leave — audits genuinely have no structured references yet; this task is
   about *lifecycle*, not follow. Note that in the body so a reader isn't surprised.
6. **Doc.** Update `docs/ARCHITECTURE.md` §"The TUI" `sort.go/.../action.go` bullet and
   the M10 framing: state that lifecycle is registry-driven off each entity's declared
   transitions (audits now mutate in-TUI), and that follow-references remain
   per-entity until those entities gain structured links.

**Tests.** Add to `internal/tui/action_test.go` (message-injection style): build the
model, switch to the audits tab, press `a`, assert the menu opens with close/reopen/
defer; select `close`, assert a `movedMsg`/`actionErrMsg` is produced and (against a
real temp-repo service) the audit relocates. Add a `:close` dispatch test on the
audits tab. Keep the existing task-action tests green (they exercise the same path).
Cover the open-findings guard: closing an audit with open findings flashes red
(actionErrMsg), no move.

**Risks / gotchas.** (a) Bubble Tea `Update` is serial — no data races, but make sure
`m.action.slug`/the selected id is captured at menu-open, not re-read at apply (a
reload could move the cursor). (b) `movedMsg.to` is a `domain.Status` today; widen to
`string` (or `fmt.Stringer`) so audit buckets fit, and check the flash/`movedAway`
logic in model.go still reads correctly. (c) The post-move reload + `movedAway` guard
(H5) was written for tasks leaving the active task list; verify it behaves for an
audit leaving the `open` bucket view (same shape — a moved id absent from the reloaded
list is success, not "not found"). (d) `:` verb completion (`commandOptions` →
`transitionVerbs`) must now union every entity's verbs (or be context-scoped to the
active tab) — pick one and keep `dispatchCommand`'s resolution consistent with it.

**Done when.** `a` and `:`verbs drive lifecycle on tasks AND audits with no per-entity
branch left in model.go/action.go (the tables live in entity.go), epics declare none
cleanly, the doc/entity.go comment is corrected, and `go build ./...`,
`go test ./...`, `golangci-lint run ./...` are green.
