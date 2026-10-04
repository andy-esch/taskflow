---
schema: 1
id: 6gg7e594gcms
status: ready-to-start
epic: 21-code-quality-architecture-hardening
description: Restore guarded Thread apply parity between ordinary and workspace-opened planning services.
effort: 4-8 hours
tier: 2
priority: medium
autonomy_level: 3
tags: [architecture, ports, threads, safety]
created: "2026-10-03"
updated_at: "2026-10-03"
depends_on: [6gcwcf8rxe72]
---
## Objective

Give a workspace-opened planning service the same guarded Thread-apply identity revalidation as
the ordinary CLI opener, without moving filesystem discovery into a primary adapter.

## Evidence

`internal/workspacestore/fs.go` constructs `store.FS` with mutation authorization but without
`WithPlanningIdentityReader`. `internal/appwiring/wiring.go` supplies both to ordinary CLI opening.
The workspace service consequently exposes Thread apply, but a valid plan fails even in dry-run:
`validation failed: planning repository identity cannot be re-read for Thread apply`.

An isolated `TestWorkspaceIdentityParityProbe` opened the same initialized corpus both ways,
composed the same one-member manifest, and called `ApplyThreadPlan(plan, true)`. Direct opening
passed; workspace opening failed with that message. This predates the composition extraction and
fails closed; no current TUI apply workflow is being declared broken.

## Scope

- Reuse or consolidate local opening/identity-reader wiring at the secondary/composition boundary.
- Preserve per-invocation mutation authorization and checked source-set composition.
- Do not turn optional local discovery into a semantic `core` dependency.

## Acceptance criteria

- [ ] Direct and pointer workspace opening can compose and dry-run a valid Thread apply plan.
- [ ] Repointed roots or replaced planning identity fail before apply, with no persistence.
- [ ] Authorization is preserved for read-only, dry-run, and committing callers.
- [ ] Tests compare the ordinary and workspace opening paths; no permissive identity fallback exists.
- [ ] Shared opening code, if introduced, retains lazy discovery and one observed initial corpus.

## Related

- [Composition isolation](6gcwcf8rxe72-isolate-cli-composition-wiring-and-enforce-controller-boundaries.md)
- [Adapter-neutral planning Thread](../threads/6gcwd78p9r04-make-planning-data-access-adapter-neutral.md)

## Out of scope

- Adding Thread apply to the TUI or changing its preview UX.
- Changing graph or persisted Thread formats.
