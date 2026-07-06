---
schema: 1
status: completed
epic: 24-data-model-evolution-stable-key-storage-read-model-content-occ
description: Add the 12-char time-sortable Crockford-base32 id (ms-time + random, collision-checked) and an id frontmatter field to task/epic/audit, minted on create. Additive; no migration yet. Per ADR-0003.
effort: Unknown
tier: 3
priority: high
autonomy_level: 3
tags: [core, storage]
created: "2026-07-01"
updated_at: "2026-07-02"
started_at: "2026-07-01"
completed_at: "2026-07-02"
id: 6fhnydm02bgf
---

# Stable 12-char id: generator and id field (additive foundation)

## Objective

<why / what — one short paragraph>

## Acceptance criteria

- [ ] <observable outcome>

## Out of scope

- <explicitly excluded>

## Related

- Epic [24-data-model-evolution-stable-key-storage-read-model-content-occ](../epics/24-data-model-evolution-stable-key-storage-read-model-content-occ.md)

Implementation complete (pending review/merge).

Generator (internal/id, merged earlier): 12-char lowercase Crockford, 43-bit ms-time + 17-bit random; monotonic New() (unique+sorted within a process); stateless NewAt(unixMilli) for the migration's historical backfill. Adversarially reviewed + subagent cross-checked.

Field wiring (this change): id on domain.Task and domain.Audit (yaml:id); minted in core.NewTask/NewAudit; serialized in taskFields/auditFields right after schema; parsed via the yaml tag. Kept ADDITIVE: id is NOT in the wire DTO / knownTaskFields / schema, so the --json contract is byte-stable (no schema_version bump); LintTask only checks known fields so it tolerates id; immutable for now via the unknown-field guard on task set. schema_comments.json regenerated (2 inert domain-comment entries; --json-schema output unchanged).

Tests: internal/id suite (round-trip, golden vectors, concurrent dedup, nextValue edge cases); store CreateTask id round-trip incl an all-digit id (YAML string preservation); core NewTask/NewAudit mint-valid-id. Full suite + vet + lint green; verified end-to-end via CLI.

Deferred to later tasks: epics get the id in the flatten task (their ID is the NN-slug today); the store id-collision-check (glob <id>-*) lands with flatten when id enters the filename; contract exposure + schema_version bump is the machine-contract task.
