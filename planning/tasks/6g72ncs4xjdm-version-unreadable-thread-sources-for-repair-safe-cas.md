---
schema: 1
id: 6g72ncs4xjdm
status: next-up
epic: 30-threads-and-task-dependency-graphs
description: Make Thread snapshot comparison detect concurrent content changes in Thread files that remain unreadable.
effort: 1 day
tier: 2
priority: high
autonomy_level: 4
tags: [threads, storage, concurrency]
created: "2026-09-05"
depends_on: [6g721tvf4crh]
updated_at: "2026-09-05"
---

# Version unreadable Thread sources for repair-safe CAS

## Objective

Give unreadable Thread documents the same opaque source-revision evidence as unreadable task
documents. Graph repair and any later mutation that deliberately tolerates malformed Threads must
not authorize a write from stale `{location, message}` diagnostics when the underlying bytes changed.

## Scope

- Extend the adapter-neutral Thread read-problem contract with an opaque, non-wire source revision.
- Have the filesystem Thread reader hash the exact bytes it already parsed, without a second scan
  or a token leak through ordinary list, lint, CLI, TUI, schema, or wire projections.
- Replace `sameThreadSourceSnapshot`'s `FileProblem` equality with deterministic comparison of
  stable identity, optional location, diagnostic, and a non-empty equal source revision.
- Reuse the shared graph-aware pre-write CAS convention and preserve current rejection of malformed
  Threads in ordinary creation, membership, lifecycle, and bulk-apply planners.

## Acceptance criteria

- [ ] Byte-identical unreadable Thread sources compare as the same snapshot, while changed bytes
      that reproduce the same diagnostic compare as different snapshots.
- [ ] Missing revisions fail closed, pathless adapters can supply opaque revisions, and adapter
      ordering does not affect comparison.
- [ ] Readable/unreadable transitions plus add, remove, rename, and identity drift invalidate the
      Thread snapshot.
- [ ] Ordinary Thread diagnostics and all public or machine-readable projections remain compatible
      and cannot expose the revision token.
- [ ] Focused store tests prove graph-aware write guards return `ErrConflict` before any write when
      an unreadable Thread source changes concurrently.

## Out of scope

- Implementing dependency repair, allowing ordinary Thread mutations over malformed documents, or
  introducing a universal persisted revision field for planning entities.

## Sequencing

Follow the task-source revision slice so both use the same opaque-token and fail-closed conventions.
Land before guarded graph repair because its readable-Thread impact receipt must either be based on
stable Thread evidence or report that evidence as incomplete.

## Related

- Epic [30-threads-and-task-dependency-graphs](../epics/30-threads-and-task-dependency-graphs.md)
- Predecessor [unreadable task-source revisions](6g721tvf4crh-version-unreadable-task-graph-sources-for-repair-safe-cas.md)
- Follow-on [guarded graph repair](6g4g8gatbnrs-add-a-guarded-repair-path-for-broken-dependency-graphs.md)
- Review evidence [Claude](../audits/6g725zn6r92x-2026-09-05-version-unreadable-task-graph-sources-for-repair-safe-cas-implementation-claude.md) (M2)
- Review evidence [Antigravity](../audits/6g726063yv11-2026-09-05-version-unreadable-task-graph-sources-for-repair-safe-cas-implementation.md) (M2)
