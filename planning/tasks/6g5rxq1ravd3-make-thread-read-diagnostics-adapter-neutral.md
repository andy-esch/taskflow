---
schema: 1
id: 6g5rxq1ravd3
status: in-progress
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
started_at: "2026-09-01"
---

# Make Thread read diagnostics adapter-neutral

## Objective

Finish the portability boundary for Thread reads before a second primary adapter consumes it.
`ThreadStore` should report unreadable records with taskflow-owned identity and optional repair
context; a database, remote service, or cache must not synthesize a Markdown path to satisfy core or
the machine contract.

## Starting constraint

At task start, the task-DAG port already returned neutral `TaskGraphRead` diagnostics, but
`ThreadStore.ListThreads` still returned `[]domain.FileProblem` and `thread list --json` published
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

- [x] A pathless fake `ThreadStore` can attribute an unreadable Thread by stable ID and slug, and
  `ListThreadViews` plus machine output retain that attribution without fabricated filesystem data.
- [x] The local FS adapter preserves current readable Threads, exact actionable diagnostics, and
  optional repair paths with no duplicate directory scan.
- [x] Core Thread projections and the TUI-facing service boundary depend only on
  the neutral read contract; no path parsing or filesystem type crosses that
  boundary.
- [x] Missing, invalid, duplicate, drifted, and unreadable Thread identities have deterministic
  ordering and do not collide with readable Threads or task IDs.
- [x] Lint, guarded Thread creation/membership/lifecycle/apply validation, CLI human output, and
  updated JSON fixtures retain their fail-closed behavior.
- [x] Schema comments, generated schema, architecture documentation, and compatibility notes name
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
- Follow-up [adapter-neutral repository lint load diagnostics](6g5vm4efjcdv-make-repository-lint-load-diagnostics-adapter-neutral.md)
- Downstream [contention-safe Thread projection loading](6g5rwjqeh6a6-wire-thread-projections-into-the-tui-with-contention-safe-reloads.md)
- ADR [0006 — Adopt Threads as task DAGs](../adrs/0006-adopt-threads-as-task-dags.md)
- Thread [Complete production Threads](../threads/6g503c6pfqeb-complete-production-threads.md)

## Implementation progress (2026-09-01)

ThreadStore now returns one adapter-neutral ThreadRead snapshot whose unreadable records carry optional stable ID, slug, location, and message. The filesystem adapter converts its native FileProblem scan once at the boundary while guarded creation, membership, lifecycle, and apply paths retain exact file snapshots for under-lock comparison. Core list, compose, lint, and validation flows consume the neutral contract without recovering identity from locations.

Thread projections now fail closed on frontmatter/filename drift, duplicate readable Thread IDs, and task/Thread ID collisions; missing or invalid IDs remain distinct document failures, duplicate slugs remain legal, and completed duplicates are explicitly inconsistent. Portable diagnostics and readable records use deterministic ordering, and pathless or misleading-location tests prove explicit identity remains authoritative.

Thread list human and JSON output map the neutral problem deliberately. Machine schema 1.59 replaces the preview path/message unreadable shape with optional thread_id, thread_slug, location, and required message. Generated schema comments and goldens, the README preview compatibility notice, ADR-0006, and architecture guidance are updated.

Validation is green: go test ./..., golangci-lint with 0 issues, documentation drift, module tidiness, git diff checks, generated schema checks, and planning lint. Focused tests cover pathless and mixed reads, single filesystem scans, invalid filename recovery, missing/invalid/drifted/duplicate/cross-kind identities, guarded validation attribution, CLI partial-failure behavior, and wire mapping.

Closeout also identified the broader Service.Lint FileProblem result as a separate multi-entity portability seam. Task 6g5vm4efjcdv tracks that work behind this task rather than widening the Thread read contract or changing local fix and mutation snapshots here.

## Adversarial review closeout (2026-09-01)

Claude's three findings were fixed in this slice: guarded creation, Thread mutation, task lifecycle, and bulk apply now have explicit unreadable-Thread no-write regressions; the concrete FS portable-read wrapper is pinned to exactly one native scan; and completed duplicate-ID projections no longer repeat completed-thread-unhealthy-evidence. Antigravity's lint-result observation is tracked by 6g5vm4efjcdv, while its request to constrain Location was rejected because optional opaque repair context is the adapter-neutral contract and core never parses it.

Both audits are closed. Validation after remediation is green: focused core/store tests, go test ./..., the full just test race suite, golangci-lint with zero issues, docs drift, module tidiness, planning and audit lint, and git diff --check.
