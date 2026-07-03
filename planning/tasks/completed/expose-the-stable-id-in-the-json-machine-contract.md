---
schema: 1
id: 6fjan6e76nex
status: completed
epic: 24-data-model-evolution-stable-key-storage-read-model-content-occ
description: Add id to TaskJSON/AuditJSON (+mappers), bump schema_version, regen goldens+schema+docs/cli. The safe half carved from machine-contract-and-docs; status==directory doc-retirement stays with flatten.
effort: Unknown
tier: 3
priority: medium
autonomy_level: 3
tags: [schema, wire]
created: "2026-07-03"
updated_at: "2026-07-03"
started_at: "2026-07-03"
completed_at: "2026-07-03"
---
# Expose the stable id in the JSON machine contract

## Objective

The stable 12-char `id` (ADR-0003) already landed on disk — minted on create,
stored in frontmatter — but no `--json` consumer could read it: `TaskJSON` and
`AuditJSON` had no `id` field, so `task list`/`task show --json`, the new `board
--json`, and any future web read side all silently dropped the key the whole epic
is built around. This exposes it — the safe half of the machine-contract work,
decoupled from the storage flatten.

## Acceptance criteria

- [x] `TaskJSON` and `AuditJSON` carry `id` (leading field, like `EpicMetaJSON`),
  `omitempty` so entities created before id assignment don't emit an empty key;
  mapped in `ToTaskJSON` / `ToAuditJSON`.
- [x] `slug` schema descriptions reworded (no longer "identifier") so `id` vs
  `slug` — stable key vs human handle — reads unambiguously.
- [x] `schema_version` 1.23 → 1.24 with a changelog entry.
- [x] Fixture tasks + audit carry ids so the goldens exercise `id` end-to-end;
  goldens and `schema --json-schema` regenerated (docs/cli unaffected).
- [x] Wire mapper tests: `ToTaskJSON`/`ToAuditJSON` carry the id and omitempty
  holds for an id-less entity.

## Out of scope

- Retiring the `status == directory` invariant in CLAUDE.md / ARCHITECTURE /
  README — stays coupled to the flatten (the other half of
  [[machine-contract-and-docs-for-id-flat-schema-version-bump]]).
- Epics — already expose `id` (their NN-slug identity); no short-id there.

## Related

- Epic [[24-data-model-evolution-stable-key-storage-read-model-content-occ]]
- ADR [[0003-stable-key-id-addressed-storage]]
