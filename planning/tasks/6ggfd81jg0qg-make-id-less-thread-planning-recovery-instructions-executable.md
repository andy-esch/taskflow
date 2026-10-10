---
schema: 1
id: 6ggfd81jg0qg
status: completed
epic: 20-cli-ux-and-ergonomics
description: Make Thread compose/apply recovery hints executable for marker-backed, markerless, and missing pointer-target identities.
effort: 2-4 hours
tier: 3
priority: medium
autonomy_level: 3
tags: [cli, recovery, threads, testing]
created: "2026-10-04"
audit_sources: [2026-10-04-workspace-thread-apply-identity-parity-implementation-codex, 2026-10-05-explicit-persistence-authorization-implementation-codex]
updated_at: "2026-10-10"
depends_on: [6g1xp8qymz1m]
started_at: "2026-10-10"
completed_at: "2026-10-10"
---
# Make ID-less Thread planning recovery instructions executable

## Objective

Give callers a recovery instruction that works in the state that prevented Thread compose/apply,
while preserving repository identity binding and all existing refusal/durability behavior.

## Evidence

The [workspace parity Codex review](../audits/6ggee2fx4053-2026-10-04-workspace-thread-apply-identity-parity-implementation-codex.md)
executed the current instruction rather than just checking its spelling. After a direct marker is
removed, Thread apply says to run `config migrate`, but that command refuses the markerless tree
and says to initialize it first. `internal/config/migrate.go:143` requires a governing marker;
the hint in `internal/store/threadapply.go:218` and related core compose/prepare errors cannot
assume one exists. This limitation predates the workspace parity change.

The independent [authorization Codex review](../audits/6ggw69vdfk1z-2026-10-05-explicit-persistence-authorization-implementation-codex.md)
reproduced it again on 2026-10-05 using the fresh binary: markerless `config migrate` exits 11
and asks for initialization. Explicit initialization preserves documents; recomposition succeeds,
while the old plan still conflicts with the regenerated identity. This strengthens the existing
task's recovery evidence; no repair is claimed implemented and no refactor prerequisite is added.

## Scope

- Inspect ID-less compose and apply guidance across direct legacy markers, markerless planning
  trees, and opted-in pointers whose target marker is missing.
- Prefer the smallest honest conditional guidance or existing neutral configuration evidence;
  do not put filesystem discovery or local paths into core/domain just to select a command.
- For an existing ID-less marker, preserve the migration route. For a markerless tree, name explicit
  initialization before recomposition. Pointer rebinding must remain deliberate.
- Execute each suggested command in an isolated fixture, with explicit `init --path` and
  `--no-register` where needed; do not let fixture cwd or home registration mask a bad instruction.
- Keep durable-ID replacement explicit: a new ID invalidates old plans, and recovery must not
  suggest blindly reusing an old materialized plan.

## Acceptance criteria

- [x] ID-less compose/apply guidance distinguishes or clearly explains marker-backed migration versus markerless initialization.
- [x] Tests execute the emitted recovery steps against direct marker-backed, markerless, and missing pointer-target marker fixtures.
- [x] Following the instructions preserves existing semantic documents and does not silently rebind an opted-in pointer or adopt a stale plan.
- [x] Refusal sentinels, durability receipts, and identity guards remain unchanged; normal tests, lint, generated docs, and planning lint pass.

## Implementation

Portable compose, preparation, and the apply writer now share honest conditional guidance:
`config migrate` for an existing marker, or explicit `init --path <planning-repo>
--taskflow-root <existing-tree> --no-register` for a markerless tree. The explicit layout
argument prevents a nested `planning/` tree from being hidden by a newly scaffolded flat tree.
Restore the original configuration where possible; regenerating identity requires recomposition.

Configuration discovery already knows marker/layout evidence, so missing-ID pointer refusals
and legacy-pointer migration give a shell-quoted command for the actual target. Target recovery
does not modify the pointer or adopt its expected ID. An opted-in pointer continues refusing
the regenerated ID; a legacy pointer binds only through a subsequent deliberate migration.

`internal/cli/thread_recovery_test.go` parses emitted commands with a POSIX shell and executes
their argv through the real CLI/composition. Flat/nested, legacy-marker/markerless, and
bound/unbound-pointer cases run from an unrelated cwd with an unrelated ambient space. They
check preserved document bytes, no home registration, surgical marker migration, old-plan
refusal, recomposition success, and continued pointer conflict.

An adjacent pre-existing mismatch is tracked by
[data-directory pointer identity parity](6gjdgpq17sr6-preserve-planning-identity-for-pointers-addressing-data-subdirectories.md):
legacy migration can bind the governing ancestor's ID, but discovery checks only an immediate
marker at the pointer address. Recovery now directs this case to the existing governing marker,
not initialization of a second marker; alias identity semantics remain a deliberate follow-up.

Validation: focused real CLI replay tests and `just test` (race), `just lint`, `just build`,
generated CLI/schema drift tests, and planning lint pass. A compiler-valid mutation forcing
markerless pointer recovery to use `--taskflow-root .` failed the intended nested-target
regressions; restored production behavior passes.

Independent [Codex](../audits/6gje9v0x5g5d-2026-10-10-bounded-planning-queries-and-identity-recovery-implementation-codex.md)
and [Antigravity](../audits/6gje9v15sjjm-2026-10-10-bounded-planning-queries-and-identity-recovery-implementation-antigravity.md)
review briefs cover this fix together with bounded queries. Implementation criteria are met;
these independent reviews have not yet run.

## Sequencing and boundaries

A small follow-up in the [CLI contract Thread](../threads/6gbn4v2jpf2m-make-cli-contracts-self-describing-and-bounded.md),
after the already-shipped configuration lifecycle foundation. It is independent of the remaining
adapter-neutral refactor slices and does not become a new blocker for their final regression pass.

## Related

- [Configuration lifecycle foundation](6g1xp8qymz1m-consolidate-the-configuration-lifecycle-under-one-config-command-hub.md)
- [Workspace identity parity](6gg7e594gcms-preserve-planning-identity-revalidation-in-workspace-opened-thread-apply.md)
- [Adapter-neutral Thread](../threads/6gcwd78p9r04-make-planning-data-access-adapter-neutral.md)

## Out of scope

Changing planning storage, minting identity automatically during compose/apply, loosening pointer
checks, broad retry policy, or implementing a generic remediation framework.
