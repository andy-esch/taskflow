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
updated_at: "2026-11-15"
audited: "2026-11-15"
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

## Sweep verification 2026-11-15

First automated weekly sweep of this task (no prior `audited:`). Every citation
re-checked against `main` (`5513c3d`).

**Verified accurate (2026-11-15) — the premise fully holds.** All three cited files
exist: [registry service/port](../../internal/core/space_registry.go),
[registry adapter](../../internal/spacestore/fs.go) and
[guarded local registry writes](../../internal/userconfig/spaces.go). The removal path
is still label-only at every layer:
`SpaceRegistryService.Forget(id string, dryRun bool)` (`internal/core/space_registry.go:169`)
→ `FS.ForgetSpace(id string, dryRun bool)` (`internal/spacestore/fs.go:85`)
→ `userconfig.ForgetSpace(id string, dryRun bool)` (`internal/userconfig/spaces.go:194`).
No parameter carries expected-registration evidence, and the `SpaceRegistryStore` port
(`space_registry.go:15-20`) declares no selected-entry variant — a repository-wide search
for `Expected`/`Selected` in the three cited files returns nothing. `Resolve` still
re-looks-up by machine-local label (`space_registry.go:46`), which is exactly the
re-resolution AC 1 wants replaced.

**Dependency state.** `depends_on: [6gmshzw7we0m]`
([define-the-tui-authoring-and-workspace-action-contract](6gmshzw7we0m-define-the-tui-authoring-and-workspace-action-contract.md))
is now `completed` with 5/5 criteria met, so `task blockers` reports
`role=candidate gate=clear eligible=True` and this task is startable. The body never
claimed to be blocked, so there is no prose/graph contradiction to correct.

**Nothing ticked.** No acceptance criterion is met; the guard does not exist yet.

## Progress log

- 2026-11-15: automated weekly sweep — first audit; premise fully verified against `5513c3d` (label-only `Forget` at all three layers, no expected-registration evidence anywhere in the port), and the prerequisite contract task is now completed so the gate is clear.
