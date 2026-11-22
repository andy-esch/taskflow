---
schema: 1
id: 6gmshzxksjcn
status: ready-to-start
epic: 18-tui-bubble-tea-interactive-planning-browser
description: Add or remove task prerequisites while inspecting a task or Thread, with explicit direction and the existing repository-global graph guards.
effort: 1-2 days
tier: 2
priority: medium
autonomy_level: 3
tags: [tui, authoring, ux]
created: "2026-10-17"
depends_on: [6gmshzw7we0m]
updated_at: "2026-11-22"
audit_sources: [2026-11-22-weekly-task-sweep]
audited: "2026-11-22"
---
# Edit task dependencies from the TUI with guarded graph feedback

## Objective

Make small sequencing changes possible while inspecting a task or Thread, using explicit dependency direction and existing repository-global graph guards.

## Scope

Provide searchable add/remove dependency actions with clear prerequisite/dependent labels. Resolve selections by adapter-established identity and reuse core graph validation, impact and recovery projections.

## Acceptance criteria

- [ ] From a selected task, add or remove an explicit prerequisite with readable endpoint labels and identity disambiguation; direction is never inferred from visual screen position.
- [ ] Cycles, dangling/ambiguous endpoints and broken-graph constraints are enforced by core; dependency editing is not raw SetFields.
- [ ] Keep both selected endpoint intents bound across refresh, rename, deletion and workspace changes rather than starting a fresh human-reference lookup at submission.
- [ ] Show guard refusals, proposed effects where supported, applied/remaining progress and core remedies; do not replay partial work.
- [ ] Thread membership is unchanged unless a separate explicit action requests it, and focused graph/identity/UI tests plus normal gates and help/docs pass.

## Design attention

A task outside the displayed Thread may still be a valid prerequisite in the same planning space. Avoid hiding global effects behind a local graph view.

## Out of scope

Cross-planning-space edges, graph repair, bulk manifests, force overrides and drag-and-drop graph editing.

## Related

- [Authoring capability boundary and sequencing](../threads/6gmshnmnew7p-enable-routine-planning-and-workspace-management-in-the-tui.md).
- [Owning dependency_operations.go](../../internal/core/dependency_operations.go).
- [Owning thread_projection.go](../../internal/tui/thread_projection.go).
- [Owning THREADS_COMPATIBILITY.md](../../docs/THREADS_COMPATIBILITY.md).

## Sweep verification (2026-11-22)

First automated weekly sweep for this task (no prior `audited:`), against `main`
(`6abe9de`). `internal/` last changed 2026-10-25. **Nothing in the body is stale.
Two things are worth recording: the dependency gate has flipped, and AC 3's core
seam does not exist yet.**

**The gate flipped from blocked to clear, and the body never claimed otherwise.**
`depends_on: [6gmshzw7we0m]` points at
[define-the-tui-authoring-and-workspace-action-contract](6gmshzw7we0m-define-the-tui-authoring-and-workspace-action-contract.md),
which is now `completed`. `task blockers` reports `role: candidate`, `gate: clear`,
`eligible: true` with an empty blocker frontier, and the task appears in
`task list --unblocked`. The owning Thread
(`6gmshnmnew7p`) reports `graph: healthy · projection healthy · 0 in flight ·
12 eligible member(s)`, this task among them. So this task became startable without
any prose needing correction — the edge is still real and still correctly recorded,
it is simply satisfied. Four sibling tasks on the same prerequisite became eligible
at the same time.

**Verified accurate (2026-11-22):** all four `## Related` targets resolve —
`internal/core/dependency_operations.go`, `internal/tui/thread_projection.go`,
`docs/THREADS_COMPATIBILITY.md`, and the authoring Thread file.

**AC 2's premise holds and is stronger than written.** `task set` refuses
graph-owned fields outright: its help states `depends_on` and the legacy
`blocked_by`/`dependencies`/`blocks` fields "cannot be changed or removed here,
including with `--force`". So "dependency editing is not raw SetFields" is enforced
by the existing guard, not merely a design intention — the TUI could not bypass core
here even by accident. Core supplies `AddTaskDependencies` /
`RemoveTaskDependencies` (`core/dependency_operations.go:104`, `:110`) with cycle,
dangling-endpoint and broken-graph validation behind
`runDependencyMutation` (`:139`).

**AC 1 is unstarted: the TUI has read-only dependency surfaces only.** `internal/tui`
renders dependency *information* — the Thread topology view with its explicit
`[prerequisite] ─▶ [dependent]` direction legend (`detail.go:1417`), member
dependency ranks (`detail.go:1387`, `:1482`), incoming-edge grouping
(`detail.go:1660`), and the graph key in help (`help.go:129-130`) — and has no
add/remove action anywhere. The readable endpoint labels and explicit-direction
vocabulary AC 1 asks for therefore already exist as rendering; what is missing is
the action.

**AC 3 needs a core entry point that is not exported yet — the one real finding.**
AC 3 requires both selected endpoint intents to stay bound "rather than starting a
fresh human-reference lookup at submission". Core already has the planner for
exactly this: `planSelectedDependencyEdges`
(`core/dependency_operations.go:253`), which takes a `dependentID` plus resolved
prerequisite ids. But it is unexported and reached only from inside
`runDependencyMutation`'s retry path (`:177`). Both **public** service entry points
still take human references (`AddTaskDependencies(taskRef string, prerequisiteRefs
[]string, ...)`), so a TUI caller today has no way to submit a selected identity —
it would have to hand back a string and re-resolve, which is precisely what AC 3
forbids. The shape needed is a selected-identity sibling in the style of
`Service.SetSelectedEpicFields` (`core/service_epic.go:323`). Recorded as a finding
rather than a scope change: adding a core method is implementation work this task
authorizes, not something a sweep decides.

No acceptance criterion was ticked: none is met, and no implementation exists.

## Progress Log

- 2026-11-22: automated weekly sweep — first audit; all four cited paths verified; the `6gmshzw7we0m` prerequisite is completed so the gate is now `clear` and this task is eligible (with four siblings); TUI confirmed read-only on dependencies; flagged that AC 3 needs an exported selected-identity entry point since `planSelectedDependencyEdges` is unexported and both public methods still take human refs.
