---
schema: 1
id: 6g721tvf4crh
status: completed
epic: 30-threads-and-task-dependency-graphs
description: Make whole-snapshot comparison detect concurrent content changes in task files that remain unreadable.
effort: 1 day
tier: 2
priority: high
autonomy_level: 4
tags: [threads, graph, storage, concurrency]
created: "2026-09-05"
audit_sources: [planning/audits/6g71yr50md16-2026-09-05-add-a-guarded-repair-path-for-broken-dependency-graphs-claude.md, planning/audits/6g725zn6r92x-2026-09-05-version-unreadable-task-graph-sources-for-repair-safe-cas-implementation-claude.md, planning/audits/6g726063yv11-2026-09-05-version-unreadable-task-graph-sources-for-repair-safe-cas-implementation.md]
updated_at: "2026-09-05"
started_at: "2026-09-05"
completed_at: "2026-09-05"
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

- [x] Two scans of byte-identical unreadable task content compare as the same source snapshot.
- [x] Changing an unreadable task's bytes makes the snapshots differ even when its path, recovered
      identity, problem code, and error message are unchanged.
- [x] Transitions between readable and unreadable representations always invalidate the snapshot.
- [x] Pathless adapters can supply their own opaque revision without fabricating a local path, and
      core never exposes the token as task-domain or public wire data.
- [x] Existing ordinary graph reads and mutations retain their behavior; focused race tests prove a
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

## Implementation progress (2026-09-05)

`TaskGraphLoadProblem` now carries an opaque, non-wire `SourceVersion`. The filesystem graph reader hashes exact unreadable bytes during the same resilient scan used by ordinary task listing, while ordinary `FileProblem`, planner-facing tasks, and machine contracts remain unchanged. `TaskGraph` privately normalizes and compares complete load-problem evidence; missing revisions fail closed for CAS equality.

The graph mutation pre-write check is now a named helper valid for both healthy mutations and the forthcoming broken-graph repair path. Regression coverage proves deterministic pathless revisions, adapter-order independence, missing-token refusal, readable/unreadable transitions, exact-byte restoration, and `ErrConflict` when changed unreadable bytes reproduce the same diagnostic.

Validation: focused core/store tests pass; the full `go test ./...` suite passes; `just test` passes under the race detector; `just lint`, `just tidy-check`, planning lint, audit lint, and `git diff --check` are clean.

## Adversarial review closeout (2026-09-05)

Claude and Antigravity independently found the same four issues. The implementation now directly
tests the both-empty missing-revision case, routes all five graph-aware pre-write guards through the
shared snapshot verifier, and uses one file-diagnostic conversion for unreadable identity. Their
Thread-side CAS finding is valid but outside this task: it is tracked by `6g72ncs4xjdm`, sequenced
after this slice and before guarded repair, and dogfooded in `complete-production-threads`.
