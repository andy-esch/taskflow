---
schema: 1
id: 6g721tvf4crh
status: next-up
epic: 30-threads-and-task-dependency-graphs
description: Make whole-snapshot comparison detect concurrent content changes in task files that remain unreadable.
effort: 1 day
tier: 2
priority: high
autonomy_level: 4
tags: [threads, graph, storage, concurrency]
created: "2026-09-05"
audit_sources: [planning/audits/6g71yr50md16-2026-09-05-add-a-guarded-repair-path-for-broken-dependency-graphs-claude.md]
---

# Version unreadable task graph sources for repair-safe CAS

## Objective

Close the whole-snapshot compare-and-swap blind spot for task files that remain unreadable across
two scans. Broken-graph repair will operate precisely where raw edits and parse failures are likely,
so an unchanged error message cannot stand in for unchanged source bytes.

## Scope

- Add an opaque source revision to the adapter-neutral task-graph load-problem contract. Core may
  compare and copy the token but must not infer filesystem or hashing semantics from it.
- Have the filesystem adapter derive the revision from the exact source bytes before parsing, so
  malformed content still participates in authoritative snapshot comparison.
- Include unreadable-source revisions in `SameSourceSnapshot` while preserving deterministic
  problem ordering and the existing readable-task `SourceVersion` contract.
- Cover readable-to-unreadable, unreadable-to-readable, path/identity drift, and content changes
  that reproduce the same parse error.

## Acceptance criteria

- [ ] Two scans of byte-identical unreadable task content compare as the same source snapshot.
- [ ] Changing an unreadable task's bytes makes the snapshots differ even when its path, recovered
      identity, problem code, and error message are unchanged.
- [ ] Transitions between readable and unreadable representations always invalidate the snapshot.
- [ ] Pathless adapters can supply their own opaque revision without fabricating a local path, and
      core never exposes the token as task-domain or public wire data.
- [ ] Existing ordinary graph reads and mutations retain their behavior; focused race tests prove a
      stale broken-graph repair would receive `ErrConflict` before any write.

## Out of scope

- Implementing graph repair, changing the task document format, or designing a universal revision
  scheme for every entity type.

## Sequencing

Land before the source-declaration projection and guarded repair capability. It is intentionally
small enough to establish the repair transaction's concurrency premise independently.

## Related

- Epic [30-threads-and-task-dependency-graphs](../epics/30-threads-and-task-dependency-graphs.md)
- Audit [guarded graph repair — Claude](../audits/6g71yr50md16-2026-09-05-add-a-guarded-repair-path-for-broken-dependency-graphs-claude.md) (H6)
