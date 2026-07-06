---
schema: 1
status: completed
epic: 24-data-model-evolution-stable-key-storage-read-model-content-occ
description: Retire the status-equals-directory invariant in CLAUDE.md, ARCHITECTURE.md, README once frontmatter is truth (the id-in-envelopes half was carved to its own task, done first). Per ADR-0003.
effort: Unknown
tier: 3
priority: medium
autonomy_level: 3
tags: [schema, docs]
created: "2026-07-01"
updated_at: "2026-07-03"
id: 6fhnydm01hdg
completed_at: "2026-07-03"
---
# Machine contract & docs: retire the status/bucket == directory invariant

## Objective

The doc half carved from the id-in-envelopes work: once frontmatter became the authority
for `status` (tasks) / `bucket` (audits) — ADR-0003 Phase A — retire the now-false
"status/bucket == directory" invariant from the canonical docs.

## Done

- [x] `docs/ARCHITECTURE.md` — frontmatter is authoritative, the directory is a lock-step
  mirror; `Task.Declared` → `FolderStatus`; `lint --fix` MOVES a misfiled file (updated
  during the Phase-A review).
- [x] `CLAUDE.md` — the non-negotiable + the "tasks live in `<status>/`" line now describe
  the mirror model (change status via the lifecycle verbs, `lint --fix` relocates drift).
- [x] `README.md` — "a task's `status:` **is** its directory" → "authoritative in
  frontmatter; the dir is a lock-step mirror"; the `lint --fix` description updated.
- [x] `schema_version` bump (→ 1.25) + the `schema task` convention / jsonschema tags
  landed with the Phase-A task/audit work (steps 4–6).

## Related

- Epic [24-data-model-evolution-stable-key-storage-read-model-content-occ](../epics/24-data-model-evolution-stable-key-storage-read-model-content-occ.md)
- ADR [0003-stable-key-id-addressed-storage](../adrs/0003-stable-key-id-addressed-storage.md)
- The id-in-envelopes half: [expose-the-stable-id-in-the-json-machine-contract](6fjan6e76nex-expose-the-stable-id-in-the-json-machine-contract.md) (done).
