---
schema: 1
id: 6gmshzxksjcn
status: ready-to-start
epic: 18-tui-bubble-tea-interactive-planning-browser
description: Make small sequencing changes while inspecting a task or Thread, using explicit dependency direction and existing repository-global graph guards.
effort: 1-2 days
tier: 2
priority: medium
autonomy_level: 3
tags: [tui, authoring, ux]
created: "2026-10-17"
depends_on: [6gmshzw7we0m]
updated_at: "2026-10-25"
audit_sources: [2026-10-25-weekly-task-sweep]
audited: "2026-10-25"
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

## Sweep refresh (2026-10-25)

2026-10-25: automated weekly sweep — references verified accurate and the task
is unshipped; one finding recorded against AC 3.

- All cited paths exist: `internal/core/dependency_operations.go` (499 lines),
  `internal/tui/thread_projection.go` (80 lines) and
  `docs/THREADS_COMPATIBILITY.md`.
- The reuse premise holds for AC 2. `Service.AddTaskDependencies` and
  `Service.RemoveTaskDependencies` (`dependency_operations.go:104-112`) route
  through one guarded mutation path that already produces cycle/dangling
  refusals, `TaskGraphStateImpact` projections, a `DependencyMutationReceipt`
  and a `dependencyMutationRemedy` string — so "enforced by core, not raw
  SetFields" and AC 4's refusal/progress/remedy surface are available today.
- No dependency editing exists in the TUI yet: no `AddDependency` /
  `RemoveDependency` call site under `internal/tui/`.

**Finding — AC 3 has no public core seam today.** AC 3 asks to "keep both
selected endpoint intents bound across refresh, rename, deletion and workspace
changes rather than starting a fresh human-reference lookup at submission", and
adds "add bounded core support where a fresh-reference API cannot guarantee
that". That is precisely the current situation, and it is worth stating plainly
before implementation starts:

- The exact machinery AC 3 describes already exists as
  `planSelectedDependencyEdges` (`dependency_operations.go:253`), whose comment
  reads: "Recompute effects for exact selected endpoints. Calling ResolveTaskID
  again, even with an ID-shaped string, could select an impostor after
  disappearance."
- But it is unexported, and the only exported entry points take human refs and
  call `graph.ResolveTaskID` on both endpoints first
  (`planDependencyEdges`, `dependency_operations.go:224-248`).
- So an adapter cannot satisfy AC 3 through today's public surface. The "add
  bounded core support" clause is not a contingency here — it is required work,
  and it is a small exported wrapper over code that already exists rather than
  new graph logic.

Core already names this pattern elsewhere: `SetSelectedTaskFields`,
`MoveSelectedTask`, `DeferSelectedTask`, `SetSelectedEpicFields`,
`MoveSelectedEpic`, `MoveSelectedAudit` and `NewTaskForSelectedEpic` all take a
`*MutationTarget` rather than a reference. An `AddSelectedTaskDependencies` /
`RemoveSelectedTaskDependencies` pair would sit in that family. Flagged, not
rescoped — the AC already covers the work and no criterion was ticked.

`description` tightened to 145 chars (was 154, over the 150-char sweep
threshold); meaning unchanged, and `## Objective` keeps the original sentence.
