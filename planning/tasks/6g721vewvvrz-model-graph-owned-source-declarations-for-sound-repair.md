---
schema: 1
id: 6g721vewvvrz
status: next-up
epic: 30-threads-and-task-dependency-graphs
description: Preserve raw canonical and legacy dependency declarations so broken-graph repair can simulate and target source edits without losing evidence.
effort: 2-3 days
tier: 2
priority: high
autonomy_level: 3
tags: [threads, graph, core, storage]
created: "2026-09-05"
depends_on: [6g721tvf4crh]
updated_at: "2026-09-05"
audit_sources: [planning/audits/6g71yr50md16-2026-09-05-add-a-guarded-repair-path-for-broken-dependency-graphs-claude.md]
---

# Model graph-owned source declarations for sound repair

## Objective

Introduce an adapter-neutral source projection beneath `TaskGraph` so recovery code can reason
about every physical graph-owned declaration without losing duplicates, malformed values, legacy
field ownership, duplicate-ID records, or unreadable-source evidence. Keep the semantic graph as
the authoritative query projection; repair needs the lossless source view as its editing proof.

## Scope

- Model each readable task source and its raw graph-owned declarations with source ownership,
  field vocabulary, raw value, and enough stable attribution to name an exact repair intent.
- Distinguish the raw canonical declaration list, the deduplicated canonical edge set, and the
  canonical-plus-legacy projected graph. Callers must not infer one from another.
- Represent legacy `blocks` in its declaring source even though its projected edge points from that
  source to a dependent; do not force every repair through dependent-owned `depends_on` writes.
- Preserve every physical record when IDs collide and preserve unreadable records and revisions
  during prospective simulation. A representative `TaskGraph.Task` map is insufficient input.
- Provide pure projection/simulation support for declaration removal and idempotent deduplication.
  Untouched declarations and records must be demonstrably byte-semantic equivalents in the
  prospective source model.
- Define a narrow materialization vocabulary suitable for a future dedicated repair port. Do not
  widen `TaskDependencyWrite.ClearLegacy`, whose all-fields behavior remains migration-specific.

## Acceptance criteria

- [ ] Duplicate canonical values remain individually diagnosable while `dedupe` means “retain
      exactly one,” independent of occurrence renumbering after an interrupted retry.
- [ ] Invalid and dangling raw values survive verbatim for diagnosis and receipts; the model never
      guesses a replacement ID from a slug-shaped value.
- [ ] A legacy `blocked_by`, `dependencies`, or `blocks` occurrence can be selected without clearing
      unrelated fields or values from the same source.
- [ ] Prospective simulation retains duplicate-ID shadow records, unreadable records and their
      revisions, and all untouched source declarations; tests demonstrate the former representative
      task-map simulation's false-accept and false-reject cases cannot recur.
- [ ] Canonical-only, projected-union, and raw-declaration queries have explicit names, immutable
      results, deterministic ordering, and no filesystem or renderer types in core.
- [ ] Contract tests cover duplicate values, duplicate IDs, invalid tokens, dangling references,
      all legacy directions, legacy-induced cycles, and mixed readable/unreadable repositories.

## Out of scope

- Choosing repair operations for a user, writing files, changing persisted frontmatter, or replacing
  the small owned DAG algorithms with a third-party graph library.

## Sequencing

Follows unreadable-source revisioning, then directly gates the guarded repair task. This task owns
the lossless representation and simulator; the downstream task owns policy, authorization,
materialization, receipts, and CLI behavior.

## Related

- Epic [30-threads-and-task-dependency-graphs](../epics/30-threads-and-task-dependency-graphs.md)
- Audit [guarded graph repair — Claude](../audits/6g71yr50md16-2026-09-05-add-a-guarded-repair-path-for-broken-dependency-graphs-claude.md) (H2, H3, H5, M6, L1)
