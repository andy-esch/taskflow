---
schema: 1
id: 6gjdgpq17sr6
status: next-up
epic: 20-cli-ux-and-ergonomics
description: Align pointer binding and validation for data-directory aliases without minting a second marker or weakening durable-ID checks.
effort: S
tier: 2
priority: high
autonomy_level: 3
tags: [config, identity, recovery]
created: "2026-10-10"
depends_on: [6ggfd81jg0qg]
updated_at: "2026-10-10"
---
# Preserve planning identity for pointers addressing data subdirectories

## Objective

Make a pointer naming the actual data subdirectory of a marker-backed planning repository
resolve the same durable identity as a pointer naming its governing configuration directory.
Do not create a second configuration marker, remove expected-ID checks, or silently rebind.

## Evidence

Found during the ID-less Thread recovery work. `TestDiscover_PlanningRepoHonorsTargetTaskflowRoot`
already documents that pointers at a repository root and its data subdirectory are equivalent,
but its fixture has no durable identity. `InitPointer`/`migrateUnlocked` obtain target identity
through `planningRepoID`/`configForRoot`, which walk to the governing ancestor. The subsequent
`resolvePlanningRepo` opted-in check only reads a marker immediately at the pointer address.

A temporary real-config probe on 2026-10-10 created a planning repo with `taskflow_root = "planning"`,
a legacy pointer targeting `<repo>/planning`, and ran `Migrate`. Migration succeeded and added
`planning_repo_id`; the next `Discover` refused with conflict because the addressed subdirectory
carried no immediate marker. The probe was removed rather than committing a test that blesses
the defect. This mismatch predates the recovery fix.

The recovery change now recognizes this ancestor-marker case and tells callers to point at the
existing governing configuration rather than recommending initialization of a second marker.
That bounds current advice; it does not claim alias identity parity is implemented.

## Scope

- Align target identity lookup/validation across direct discovery, pointer initialization,
  legacy-pointer migration, and opted-in pointer discovery.
- Preserve canonical-worktree anchoring, physical-root checks, pointer-chain refusal,
  mismatch refusal, and markerless compatibility.
- Test flat and nested/custom data directories, correct/missing/changed IDs, symlinks,
  and unknown configuration fields through real configuration/CLI callers.
- Either honor the already-documented address equivalence or explicitly revisit that
  contract before writing bindings; no successful init/migration may create a pointer
  that its own next discovery cannot read.

## Acceptance criteria

- [ ] Pointers at the governing config directory and its data subdirectory have equivalent durable-ID validation.
- [ ] Real pointer init and migration tests prove the written binding is usable on subsequent discovery.
- [ ] Wrong/missing target identity remains fail-closed; recovery preserves marker topology and never silently rebinds.
- [ ] Full tests, lint, planning lint, and documentation checks pass.

## Sequencing

Follow the ID-less recovery task in the CLI-contract Thread. It does not block bounded query
delivery and does not reopen the completed adapter-boundary refactor Thread.
