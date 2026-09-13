---
schema: 1
id: 6g9mz2shwmb0
status: in-progress
epic: 30-threads-and-task-dependency-graphs
description: Qualify and publish the spatial graph and bounded-focus work as another explicitly preview-labelled installed release.
effort: 1 day
tier: 1
priority: high
autonomy_level: 2
tags: [threads, release, tui, graph, dogfood]
created: "2026-09-13"
started_at: "2026-09-13"
depends_on: [6g7r20ffjf2w, 6g9g37c50qp1]
updated_at: "2026-09-13"
---
## Objective

Qualify and publish v0.21.0 as the next installed Threads preview, centered on the flagship spatial
graph, bounded graph neighborhoods, and one-hop TUI focus. Use the release to accumulate real
compatibility and usability evidence without implying that Threads have graduated from preview.

## Scope

- Cut from one clean, immutable `main` candidate containing the completed spatial-focus and graph
  input-limit work plus the reusable host/container release gate.
- Run `just release-validate` and `just release-validate-container` on that exact candidate and
  record the commit and outcomes here.
- Dogfood a freshly built candidate through full and focused Thread graphs, directional navigation,
  task opening/return, watcher refresh, frontier output, and one-/two-hop Mermaid and JSON exports.
- Curate concise release notes around the spatial Threads story and link the implementing tasks,
  reviews, compatibility contract, and bounded planning neighborhood.
- Tag the recorded candidate, verify the GitHub workflow, four archives, checksums, embedded
  version, and an installed binary, then record immutable evidence.

## Acceptance criteria

- [ ] The one-hop TUI focus, graph input-limit regression, and reusable release-validation tasks
  are completed and merged into the clean candidate.
- [ ] Host and pinned-container release validation pass against the same recorded commit with no
  tracked worktree changes.
- [ ] A fresh candidate binary passes the bounded CLI/TUI dogfood covering full graph, one-hop
  focus and return, directional navigation, task jump/back, live refresh, frontier, and bounded
  Mermaid/JSON exports; findings are fixed or tracked.
- [ ] README and release notes retain the Threads preview classification and use planning tasks,
  reviews, and a bounded Thread graph as evidence without overstating compatibility.
- [ ] The v0.21.0 tag, release workflow, four archives, checksums, extracted version, and installed
  binary all identify the recorded candidate.
- [ ] Preview graduation remains a separate deferred decision, and unfinished navigation design or
  polish work is not made an artificial release prerequisite.

## Out of scope

- Graduating Threads from preview or removing the compatibility notice.
- Completing the whole TUI navigation Thread, its information-architecture redesign, or its demo
  refresh before this checkpoint.
- Treating the container preflight as publication authority or claiming reproducible binaries.

## Related

- Thread [Refine Thread and TUI navigation](../threads/6g90h4pg0q7n-refine-thread-and-tui-navigation.md)
- Contract [Threads compatibility and preview graduation](../../docs/THREADS_COMPATIBILITY.md)
- Previous checkpoint [v0.20.0 compatibility-hardened preview](6g7fhfpmy032-cut-v0.20.0-as-a-compatibility-hardened-threads-preview.md)
- Release gate [Make release validation reproducible](6g7r20ffjf2w-make-release-validation-a-reproducible-one-command-gate.md)
