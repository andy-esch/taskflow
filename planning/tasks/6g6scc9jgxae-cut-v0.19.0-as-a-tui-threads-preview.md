---
schema: 1
id: 6g6scc9jgxae
status: completed
epic: 30-threads-and-task-dependency-graphs
description: Release the first TUI Thread workflow after graph-health reporting and Atlas refresh hardening pass clean-main dogfood.
effort: S
tier: 3
priority: medium
autonomy_level: 4
tags: [threads, release, tui, dogfood]
created: "2026-09-04"
updated_at: "2026-09-05"
depends_on: [6g63db3sdfrh, 6g697mp8s4tx]
started_at: "2026-09-04"
completed_at: "2026-09-05"
---

# Cut v0.19.0 as a TUI Threads preview

## Objective

Cut the first release in which Threads are useful from both primary local interfaces. Treat v0.19.0
as a second explicit preview checkpoint: the CLI workflow remains available, while the TUI adds
stable Thread list/detail navigation and a compact wave view backed by the same core projections.

## Scope

- Merge graph-health reporting and coherent Atlas refresh recovery before taking the release
  snapshot; these are the bounded integrity fixes discovered while dogfooding the TUI slice.
- Build from clean `main` and exercise the CLI and TUI against a throwaway planning space with
  shared members, an external gate, fan-out/fan-in topology, lifecycle changes, and live reload.
- Verify that status/lint and the TUI expose non-healthy graph evidence without changing frontier
  eligibility or inventing adapter-specific graph semantics.
- Run the normal release snapshot and validation workflow, including generated CLI docs, machine
  schema, checksums, and platform artifacts.
- Curate release notes around the complete Thread workflow and state the remaining preview
  boundaries plainly.

## Acceptance criteria

- [x] `report-graph-degradation-in-status-and-lint` and
      `preserve-coherent-atlas-summaries-across-transient-per-space-refresh-failures` are completed
      and merged to clean `main`.
- [x] A clean-build dogfood pass covers Thread list/detail, wave presentation, task navigation and
      return, external gates, shared membership, lifecycle updates, watcher reload, and a visible
      non-healthy graph diagnostic.
- [x] `just release-snapshot`, the full release validation, generated-doc/schema checks, and
      planning lint pass from the release commit.
- [x] Release notes explain the CLI and TUI capabilities as one workflow and identify guarded graph
      repair, portable multi-entity diagnostics, and a spatial graph view as post-release work—not
      silently shipped or promised stable interfaces.
- [x] The v0.19.0 tag and published artifacts identify the same clean-main commit and expected
      checksums.

## Stress tests

- Narrow and wide terminals, a Thread with multiple waves, a shared task, an external gate, a
  watcher-driven task transition, an Atlas refresh overlapping a guarded mutation, and degraded
  versus broken graph evidence.

## Out of scope

- Declaring the Thread CLI, wire contract, or TUI stable.
- A guarded graph-repair command, remote/database adapters, critical-path calculations, or a
  production two-dimensional graph renderer.

## Sequencing

This is the checkpoint after the first useful TUI topology slice. Atlas recovery is a release gate
because a transient mutation must not make a previously coherent space disappear from the TUI.
The larger repair subsystem, shared portable diagnostic migration, and spatial graph prototype
follow the release so they do not blur this snapshot or force another wire-contract change into it.

## Related

- Epic [30-threads-and-task-dependency-graphs](../epics/30-threads-and-task-dependency-graphs.md)
- Thread [Complete production Threads](../threads/6g503c6pfqeb-complete-production-threads.md)
- Previous release checkpoint [v0.18.0 CLI preview](6g5m69wpydzw-cut-v0.18.0-as-a-cli-threads-preview.md)
