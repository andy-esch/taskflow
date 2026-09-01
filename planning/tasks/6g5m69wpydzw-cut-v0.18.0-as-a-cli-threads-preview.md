---
schema: 1
id: 6g5m69wpydzw
status: completed
epic: 30-threads-and-task-dependency-graphs
description: Dogfood the complete CLI Thread workflow from clean main and release it as an explicitly evolving preview.
effort: S
tier: 3
priority: medium
autonomy_level: 4
tags: [threads, release, dogfood]
created: "2026-08-31"
depends_on: [6g3q4rv1w9e2, 6g5fthzwbeq1]
updated_at: "2026-09-01"
started_at: "2026-08-31"
completed_at: "2026-09-01"
---

# Cut v0.18.0 as a CLI Threads preview

## Objective

Take a clean-main snapshot after the first complete CLI Thread workflow has survived implementation
review and real planning use. Publish that snapshot as v0.18.0 without implying that the preview
interface, TUI, or advanced graph features are frozen.

## Scope

- Build and install from clean `main`, then exercise Thread creation, dependencies, membership,
  lifecycle, bulk apply, show, frontier, plan, and Mermaid/DOT graph export in a throwaway planning
  space.
- Run the normal release snapshot and validation workflow, including generated docs, schema,
  checksums, and platform artifacts.
- Summarize the shipped Thread capabilities and known preview boundaries in the release notes.
- Tag v0.18.0 only after the dogfood pass and release checks are clean.

## Acceptance criteria

- [x] The graph-view slice and in-flight frontier presentation follow-up are merged to `main`, and
  the release is built from a clean checkout of that commit.
- [x] A throwaway planning space passes the end-to-end CLI Thread workflow, including guarded
  failures and at least one external gate.
- [x] `just release-snapshot` and the repository's full release validation complete successfully;
  generated CLI docs and machine schema match the binary.
- [x] Release notes identify Threads as a CLI preview, enumerate the available commands, and state
  that interface details may still evolve from dogfooding.
- [x] The v0.18.0 tag and published artifacts identify the same commit and expected checksums.

## Stress tests

- Clean install rather than a stale local binary, a blocked dependency transition, shared task
  membership, an active-only frontier, YAML compose/apply retry, and both Mermaid and DOT export.

## Out of scope

- TUI Thread views, declaring the Thread API stable, critical-path or forecasting features, and
  unrelated release-scope expansion.

## Sequencing

Depends on deterministic graph views and the in-flight frontier presentation follow-up. This is the
release checkpoint before the usage-informed TUI slice, not a prerequisite for fixing defects found
during the release dogfood pass.

## Related

- Epic [30-threads-and-task-dependency-graphs](../epics/30-threads-and-task-dependency-graphs.md)
- ADR [0006 — Adopt Threads as task DAGs](../adrs/0006-adopt-threads-as-task-dags.md)
- Thread [Complete production Threads](../threads/6g503c6pfqeb-complete-production-threads.md)

## Clean-main dogfood (2026-08-31)

Built tskflwctl v0.17.1-95-gf33dc00b from a clean detached worktree at merged main commit
f33dc00b6b22422928547890a2a3290e161353f6. The retained lab is
`/tmp/taskflow-v018-dogfood.xq7mv2`: `source` contains the clean build, `space` contains the
throwaway planning repository, and `core.apply.yaml` plus `docs.apply.yaml` are the mode-0600
materialized plans.

The run initialized without registry writes; created two epics and six tasks; composed and applied
a linear three-member Thread behind an external gate; created a second fan-in Thread sharing two
tasks; and proved dry-run, first apply, idempotent retry, and no-clobber behavior. It exercised list,
show, frontier, plan, renderer-neutral JSON, Mermaid, DOT, guarded blocked task start, guarded
premature Thread completion, external-gate release, active-only frontier output, explicit member
add/remove, Thread start/complete/cancel/reopen, and affected-Thread task lifecycle receipts.
Reopening a cancelled Thread was correctly rejected because cancelled is terminal in V1; reopening
a completed Thread succeeded.

Final state: repository graph healthy, all three Thread projections healthy, core-delivery and
cross-cutting-guide completed, lifecycle-probe intentionally cancelled, machine schema 1.58, and
planning lint clean. The README now labels Threads as a v0.18.0 CLI preview. Release snapshot,
release-note publication wiring, and tag/artifact verification remain open.
