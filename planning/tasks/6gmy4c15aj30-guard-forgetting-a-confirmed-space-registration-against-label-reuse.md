---
schema: 1
id: 6gmy4c15aj30
status: ready-to-start
epic: 21-code-quality-architecture-hardening
description: Bind an already-confirmed registry entry through the guarded removal so label reuse cannot forget a different checkout.
effort: 1 day
tier: 2
priority: high
autonomy_level: 3
tags: [tui, authoring, architecture, identity]
created: "2026-10-18"
depends_on: [6gmshzw7we0m]
updated_at: "2026-10-25"
audit_sources: [2026-10-25-weekly-task-sweep]
audited: "2026-10-25"
---
# Guard forgetting a confirmed space registration against label reuse

## Objective

Provide a reusable selected-registration removal operation for an adapter that has
already displayed an entry and obtained confirmation. A fresh label lookup must
not remove a different entry after concurrent edits.

## Evidence and scope

SpaceRegistryService.Forget and the underlying registry write currently select by
machine-local label. The local writer locks and reads one snapshot, which protects
its mutation but does not compare that row with the entry the UI confirmed earlier.
A user or another process can forget/re-register that label before submission.

Keep the ordinary CLI label-based verb compatible. Add a bounded adapter-neutral
selected-entry expectation and enforce it against the same authoritative registry
snapshot used for removal. Define the relevant address/target evidence without
treating a diagnostic planning path as permission, or requiring a broken planning
tree to load. Binary composition supplies the registry capability and policy.

## Acceptance criteria

- [ ] An already-confirmed entry can be forgotten using explicit expected registration evidence rather than label-only re-resolution; the owning port defines the comparison contract.
- [ ] A missing entry, duplicate label, reused label or changed target fails closed with no unrelated registration removed, including previews and no-op cases.
- [ ] Comparison and removal use the same guarded persistence snapshot; a UI-side Catalog recheck alone is not accepted as the fix.
- [ ] Forget still edits only the machine-wide registry row, retaining comments, unknown keys/order and siblings; it does not delete planning files, checkouts or repository configuration.
- [ ] Broken entries remain removable without opening their planning tree; new portable contracts do not expose userconfig/TOML or TUI types, and absent capabilities/read-only policy fail explicitly.
- [ ] Preserve existing domain error/receipt policy and ordinary CLI behavior; portable plus real-adapter tests cover concurrent label reuse, unrelated edits, refused writes, dry-run and no filesystem deletion.

## Out of scope

Atlas UI, registration initialization/linking, deleting files, registry label rename,
new registry schema or persistent row IDs, and a generic confirmation/transaction framework.

## Related

- [Authoring Thread](../threads/6gmshnmnew7p-enable-routine-planning-and-workspace-management-in-the-tui.md).
- [Confirmed interaction and adapter-safe gate](../research/6gmrkq00348a-tui-authoring-interaction-checkpoint.md#adapter-safe-implementation-gate).
- [Registry service/port](../../internal/core/space_registry.go),
  [registry adapter](../../internal/spacestore/fs.go) and
  [guarded local registry writes](../../internal/userconfig/spaces.go).
- [Atlas consumer](6gmshzwfkj62-register-and-forget-existing-space-entry-points-from-atlas.md).

## Sweep refresh (2026-10-25)

2026-10-25: automated weekly sweep — the Evidence and scope section re-verified
line by line against current code and found still exactly true; one
cross-cutting observation added.

- `SpaceRegistryService.Forget` (`internal/core/space_registry.go:169`) still
  passes only the machine-local label down to `store.ForgetSpace(id, dryRun)`.
  No expected-entry evidence crosses the port.
- `userconfig.ForgetSpace` (`internal/userconfig/spaces.go:194`) still takes the
  write lock, reads one snapshot, and selects purely by `e.ID == id`. Its doc
  comment already promises "the exact entry selected from the same snapshot used
  for the edit", which protects the *receipt* from racing a second read — but,
  as the task says, it compares nothing against what the UI confirmed earlier.
  The guarded-snapshot half of AC 3 is therefore already in place; the missing
  half is the comparison input.
- `internal/spacestore/fs.go:85` (`FS.ForgetSpace`) is a thin pass-through to
  that function, so the new expectation has one obvious seam to thread through.

**Cross-cutting (see PR Findings):** core already has an established
selected-identity mutation family — `SetSelectedTaskFields`, `MoveSelectedTask`,
`DeferSelectedTask`, `SetSelectedEpicFields`, `MoveSelectedEpic`,
`MoveSelectedAudit`, `NewTaskForSelectedEpic` — each taking a `*MutationTarget`
instead of a human reference. The registry is one of two mutation surfaces with
no member of that family. Worth deciding deliberately whether this task's
"explicit expected registration evidence" should be named and shaped as the
registry's entry in that family rather than as a one-off parameter, so the
pattern stays recognisable. Not rescoped here — the AC wording already permits
either, and the choice is the implementer's.
