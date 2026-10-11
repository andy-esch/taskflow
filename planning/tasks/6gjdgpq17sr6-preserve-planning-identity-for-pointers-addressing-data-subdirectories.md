---
schema: 1
id: 6gjdgpq17sr6
status: completed
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
started_at: "2026-10-10"
completed_at: "2026-10-10"
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

The preceding recovery change recognized this ancestor-marker case and told callers to point
at the existing governing configuration rather than initialize a second marker. This task
replaces that temporary workaround with single-governing-marker alias identity parity.

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

- [x] Pointers at the governing config directory and its data subdirectory have equivalent durable-ID validation.
- [x] Real pointer init and migration tests prove the written binding is usable on subsequent discovery.
- [x] Wrong/missing target identity remains fail-closed; recovery preserves marker topology and never silently rebinds.
- [x] Full tests, lint, planning lint, and documentation checks pass.

## Sequencing

Follow the ID-less recovery task in the CLI-contract Thread. It does not block bounded query
delivery and does not reopen the completed adapter-boundary refactor Thread.

## Implementation and validation (2026-10-10)

- Discovery, pointer initialization and legacy migration share a captured local target
  resolution: physical data root, governing scaffold configuration, durable ID and
  recovery location. Direct data-directory discovery retains governing preferences.
- Ancestor authority requires an exact, containment-checked `taskflow_root` match.
  The nearest marker and physical Git boundaries stop inheritance; pointer ancestors
  are not followed and malformed configuration is loud. Existing bound-pointer init
  also verifies the expectation instead of reporting an unchecked idempotent no-op.
- `TestPlanningTarget_BindingRoundTrip` first reproduced the defect, then passed for
  flat/nested/custom layouts, repo/data/symlink addresses, init and migration. Negative
  tests cover wrong/missing IDs, unrelated/nearest markers, Git files/directories,
  malformed TOML and escaping roots; refused operations preserve target and pointer
  bytes. Worktree tests pin canonical anchoring against a decoy planning tree.
- Real CLI tests compose via an initialized/migrated alias and apply via the repo
  address. Executed missing-ID recovery migrates the existing governing marker,
  preserves documents and leaves the old expected ID in conflict until deliberate
  rebinding. No second marker is minted by alias binding or recovery.
- `just build`, `just test` (race), `just lint`, `just docs-check`, planning lint and
  `git diff --check` passed. Machine revision remains 1.85: this restores documented
  single-marker address equivalence, without new JSON shapes/selectors/vocabularies
  or changed documented meanings/error codes.

### Residual boundary

[Overlapping-marker identity authority](6gjgccnzxj0a-decide-and-enforce-identity-authority-for-overlapping-planning-markers.md)
tracks a separately reproduced topology problem: explicit initialization at both a
repo and its configured data directory can assign two IDs to the same physical tree.
This task does not decide a winning marker, delete configuration or retrofit an
overlapping-marker migration policy. Independent reviewers should challenge the
scope boundary if they can demonstrate a failure for a single-governing-marker alias.
Independent reviews found no new in-scope product defects; their evidence and limits
are recorded below. Passing owner/reviewer checks is not evidence of branch merge.

## Independent review closeout

- [Codex — authority/identity and shared consumers](../audits/6gjgcy198g1m-2026-10-10-planning-pointer-alias-identity-implementation-codex.md).
- [Antigravity — executable operator/recovery workflows](../audits/6gjgcy1hyt10-2026-10-10-planning-pointer-alias-identity-implementation-antigravity.md).

Both reviewers exercised hostile alias/recovery workflows and compiler-valid mutations
in independent clones, with no new in-scope findings. Codex additionally demonstrated
same-ID root redirection and replacement-ID refusal through real startup/configuration
and replay consumers. Its guarded-transfer log was verified. Antigravity corroborates
the operator path; its blanket race claim and pending transfer attestation carry the
explicit owner qualifications in that audit.

The replay seam is now pinned by
`TestPlanningDataAliasRevalidatesGoverningMarkerBeforeThreadApply`: data-alias opening
through ordinary and workspace paths retains governing-marker revalidation; previews
and apply reject redirected roots or replaced IDs without writes. No production fix
was needed after review, and no global configuration transaction is claimed. The
overlapping-marker follow-up remains open and separate from this completed member.

After adding the replay regression, build, full race tests, lint (zero issues),
generated-documentation checks, both audit lints and planning lint passed again.
Both zero-finding audits are closed with their owner qualifications retained.
