---
schema: 1
id: 6gmshzwfkj62
status: ready-to-start
epic: 18-tui-bubble-tea-interactive-planning-browser
description: Maintain the Atlas registry from the TUI while preserving the distinction between a planning space and its registered entry points.
effort: 1-2 days
tier: 2
priority: high
autonomy_level: 3
tags: [tui, authoring, ux]
created: "2026-10-17"
depends_on: [6gmshzw7we0m, 6gmy4c15aj30]
updated_at: "2026-10-25"
audited: "2026-10-25"
---
# Register and forget existing space entry points from Atlas

## Objective

Let a user maintain the Atlas registry without leaving the TUI while preserving the distinction between a planning space and its registered entry points.

## Scope

Add an existing checkout or implementation pointer by path with an optional registry label; forget one explicitly selected registration. Reuse core.SpaceRegistryService and binary-bound registry capabilities, not CLI subprocesses or direct registry writes.

## Acceptance criteria

- [ ] Atlas exposes discoverable register-existing and forget-registration actions with the selected label, physical target context and planning identity visible.
- [ ] Add validates before mutation, handles duplicate registration and broken paths through existing core policy, preserves input on error, and refreshes the correct Atlas selection.
- [ ] Forget confirmation identifies the exact entry point and states that no planning files, checkout or other registrations are deleted.
- [ ] Active-entry removal, multiple entries for one planning identity, stale selection, asynchronous results and workspace/session switching follow the reviewed action contract and have regression coverage.
- [ ] Inject the registry application capability through composition, perform I/O only in tea.Cmd, retain committed/partial evidence and pass normal code/docs gates.
- [ ] Conditional forget binds the confirmed registration under the owning
  guarded persistence seam, refusing reused-label or changed-target drift;
  core/port contracts remain registry-format-neutral and have portable plus
  real-adapter tests. · **tracked:** by 6gmy4c15aj30 — Conditional confirmed-entry comparison and portable/real-adapter evidence belong to the prerequisite; this task proves Atlas selection, confirmation and injected dispatch.

## Design attention

Owner confirmed registration-only removal from the machine-wide registry; keep an
already-open checkout usable and visibly unregistered. Do not delete planning
files or edit repo configuration. A grouped Atlas row is not authority to forget
every checkout. A registry mutation must not retarget an active planning write.

Follow the [adapter-safe gate](../research/6gmrkq00348a-tui-authoring-interaction-checkpoint.md#adapter-safe-implementation-gate):
the TUI receives an application capability, not the concrete registry adapter or
its file format. Any conditional selected-entry extension belongs at the owning
core/port/persistence seam and requires portable plus guarded-adapter evidence.
The [confirmed-registration foundation](6gmy4c15aj30-guard-forgetting-a-confirmed-space-registration-against-label-reuse.md)
owns that comparison/removal contract. This task consumes it through injected
application services and owns Atlas confirmation, current-session presentation
and refresh behavior, not registry-file surgery.

## Out of scope

Initializing a planning repo, linking implementation repos, renaming labels, deleting directories, or a new registry format.

## Related

- [Authoring capability boundary and sequencing](../threads/6gmshnmnew7p-enable-routine-planning-and-workspace-management-in-the-tui.md).
- [Owning space_registry.go](../../internal/core/space_registry.go).
- [Owning atlas.go](../../internal/tui/atlas.go).
- [Owning model.go](../../internal/tui/model.go).

## Sweep refresh (2026-10-25)

2026-10-25: automated weekly sweep — all cited references verified accurate; the
task is still entirely unshipped and its framing is unchanged.

- Every cited path exists: `internal/core/space_registry.go`,
  `internal/tui/atlas.go`, `internal/tui/model.go`, the authoring Thread and the
  interaction-checkpoint research doc.
- The reuse premise holds. `core.SpaceRegistryService` exposes `Add`,
  `RegisterInitialized` and `Forget` (`space_registry.go:113-184`) over the
  consumer-owned `SpaceRegistryStore` port, so the Scope line's "reuse
  core.SpaceRegistryService, not CLI subprocesses" is still the available route.
- No register/forget action exists in the Atlas yet. Grepping
  `internal/tui/atlas.go` for register/forget finds only presentation strings
  ("registered", "no registered entry point for this space") — no action,
  command or confirmation path. Scope is unstarted, not partially shipped.
- Dependency state matches the prose. `6gmshzw7we0m` (the action contract) is
  `completed`, so the only remaining blocker is `6gmy4c15aj30`, exactly as the
  Design attention section describes. `task blockers` reports
  `role: candidate`, `gate: blocked`, `eligible: false` with that one edge.
- `description` tightened to 131 chars (was 153, over the 150-char sweep
  threshold); wording unchanged in meaning. `## Objective` keeps the original
  full sentence.
