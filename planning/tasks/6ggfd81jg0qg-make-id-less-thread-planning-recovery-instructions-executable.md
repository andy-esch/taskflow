---
schema: 1
id: 6ggfd81jg0qg
status: ready-to-start
epic: 20-cli-ux-and-ergonomics
description: Make Thread compose/apply recovery hints executable for marker-backed, markerless, and missing pointer-target identities.
effort: 2-4 hours
tier: 3
priority: medium
autonomy_level: 3
tags: [cli, recovery, threads, testing]
created: "2026-10-04"
audit_sources: [2026-10-04-workspace-thread-apply-identity-parity-implementation-codex]
updated_at: "2026-10-04"
depends_on: [6g1xp8qymz1m]
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

- [ ] ID-less compose/apply guidance distinguishes or clearly explains marker-backed migration versus markerless initialization.
- [ ] Tests execute the emitted recovery steps against direct marker-backed, markerless, and missing pointer-target marker fixtures.
- [ ] Following the instructions preserves existing semantic documents and does not silently rebind an opted-in pointer or adopt a stale plan.
- [ ] Refusal sentinels, durability receipts, and identity guards remain unchanged; normal tests, lint, generated docs, and planning lint pass.

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
