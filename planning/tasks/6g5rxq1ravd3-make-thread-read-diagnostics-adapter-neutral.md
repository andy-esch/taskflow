---
schema: 1
id: 6g5rxq1ravd3
status: next-up
epic: 30-threads-and-task-dependency-graphs
description: Replace filesystem-shaped ThreadStore list diagnostics with an identity-aware neutral contract before TUI and future served adapters cement it.
effort: 1-2 days
tier: 3
priority: medium
autonomy_level: 3
tags: [threads, architecture, diagnostics, ports]
created: "2026-09-01"
depends_on: [6g5gbk5a5bt0, 6g5ryqqx5ab7]
updated_at: "2026-09-01"
---

# Make Thread read diagnostics adapter-neutral

## Objective

Finish the portability boundary for Thread reads before a second primary adapter consumes it.
`ThreadStore` should report unreadable records with taskflow-owned identity and optional repair
context; a database, remote service, or cache must not synthesize a Markdown path to satisfy core or
the machine contract.

## Current constraint

The task-DAG port already returns neutral `TaskGraphRead` diagnostics, but
`ThreadStore.ListThreads` still returns `[]domain.FileProblem` and `thread list --json` publishes
those path-shaped values directly. This was explicitly left outside the earlier graph-diagnostic
slice. The TUI can consume the local filesystem form, but doing so would cement an inconsistency in
the otherwise independently injectable `ThreadStore` capability before future served adapters use
it.

## Scope

- Define one neutral Thread-list result/problem contract carrying optional stable Thread ID/slug,
  optional source/location repair context, and an actionable message.
- Change the narrow `ThreadStore` list method and core `ListThreadViews` flow to use that contract;
  build on the preceding split that already moved explicit local path lookup out of the portable
  document-read capability.
- Translate filesystem scan problems once at the FS adapter boundary, recovering filename identity
  there when safe. Core must never parse a path for Thread identity.
- Map neutral diagnostics deliberately in wire output and advance the machine schema. Preserve a
  useful local path when one exists without making it mandatory for remote adapters.
- Keep lint and guarded mutation diagnosis authoritative. If those local maintenance paths still
  need `FileProblem`, adapt internally rather than leaking it back into the read port.

## Acceptance criteria

- [ ] A pathless fake `ThreadStore` can attribute an unreadable Thread by stable ID and slug, and
  `ListThreadViews` plus machine output retain that attribution without fabricated filesystem data.
- [ ] The local FS adapter preserves current readable Threads, exact actionable diagnostics, and
  optional repair paths with no duplicate directory scan.
- [ ] Core Thread projections and the TUI-facing loader depend only on the neutral read contract;
  no path parsing or filesystem type crosses that boundary.
- [ ] Missing, invalid, duplicate, drifted, and unreadable Thread identities have deterministic
  ordering and do not collide with readable Threads or task IDs.
- [ ] Lint, guarded Thread creation/membership/lifecycle/apply validation, CLI human output, and
  updated JSON fixtures retain their fail-closed behavior.
- [ ] Schema comments, generated schema, architecture documentation, and compatibility notes name
  the wire change explicitly.

## Stress tests

- Pathless remote-style problem, misleading location with explicit identity, invalid filename ID,
  frontmatter drift, duplicate stable IDs, duplicate slugs, mixed readable/unreadable lists, local
  scan-count assertion, and split task-graph/Thread adapters.

## Out of scope

- Database or HTTP implementations, redesigning task/epic/audit/research list diagnostics, changing
  guarded mutation policy, or requiring paths in portable machine output.

## Sequencing

Runs after local path resolution leaves the portable Thread read port and before contention-safe TUI
Thread loading, so the second primary adapter is the first consumer of the corrected contract rather
than another reason to preserve the filesystem leak.

## Related

- Epic [30-threads-and-task-dependency-graphs](../epics/30-threads-and-task-dependency-graphs.md)
- Prior task [adapter-neutral task-graph diagnostics](6g5gbk5a5bt0-make-task-graph-load-diagnostics-adapter-neutral.md)
- Predecessor [split local Thread paths from portable reads](6g5ryqqx5ab7-split-local-thread-path-resolution-from-portable-thread-reads.md)
- Downstream [contention-safe Thread projection loading](6g5rwjqeh6a6-wire-thread-projections-into-the-tui-with-contention-safe-reloads.md)
- ADR [0006 — Adopt Threads as task DAGs](../adrs/0006-adopt-threads-as-task-dags.md)
- Thread [Complete production Threads](../threads/6g503c6pfqeb-complete-production-threads.md)
