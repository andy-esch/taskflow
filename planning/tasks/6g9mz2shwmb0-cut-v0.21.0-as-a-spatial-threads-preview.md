---
schema: 1
id: 6g9mz2shwmb0
status: completed
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
completed_at: "2026-09-13"
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

## Release execution playbook

After the release-planning PR lands, freeze one clean `main` commit and run both forms of the same
automated gate. Record the commit printed by each command; they must match.

```bash
git status --short
git rev-parse HEAD
just release-validate
just release-validate-container
```

Use a copy of the touring-bike fixture for the bounded manual pass:

```bash
V021_LAB=$(mktemp -d /tmp/taskflow-v021-dogfood.XXXXXX)
V021_CLI="$PWD/bin/tskflwctl"
cp -R assets/demo-threads "$V021_LAB/space"

just build
"$V021_CLI" -C "$V021_LAB/space" lint
"$V021_CLI" -C "$V021_LAB/space" thread frontier touring-bike-departure
"$V021_CLI" -C "$V021_LAB/space" thread graph touring-bike-departure
"$V021_CLI" -C "$V021_LAB/space" thread graph touring-bike-departure \
  --around service-bottom-bracket-and-crankset --depth 1 --format mermaid
"$V021_CLI" -C "$V021_LAB/space" thread graph touring-bike-departure \
  --around service-bottom-bracket-and-crankset --depth 2 --json
"$V021_CLI" -C "$V021_LAB/space" ui
```

In the TUI, open Threads, cycle to the spatial graph, and confirm initial focus lands on the
in-flight bottom-bracket task. Exercise `h/j/k/l`, ambiguous-neighbor selection, `z` focus and
return, `Enter` task navigation, and `ctrl+o` return. With that view open, use another terminal to
run:

```bash
"$V021_CLI" -C "$V021_LAB/space" task complete service-bottom-bracket-and-crankset
"$V021_CLI" -C "$V021_LAB/space" task start true-and-tension-both-wheels
```

Confirm the graph refreshes without losing coherent selection. Record outcomes and file every
non-trivial finding before checking the dogfood criterion.

## Acceptance criteria

- [x] The one-hop TUI focus, graph input-limit regression, and reusable release-validation tasks
  are completed and merged into the clean candidate.
- [x] Host and pinned-container release validation pass against the same recorded commit with no
  tracked worktree changes.
- [x] A fresh candidate binary passes the bounded CLI/TUI dogfood covering full graph, one-hop
  focus and return, directional navigation, task jump/back, live refresh, frontier, and bounded
  Mermaid/JSON exports; findings are fixed or tracked.
- [x] README and release notes retain the Threads preview classification and use planning tasks,
  reviews, and a bounded Thread graph as evidence without overstating compatibility.
- [x] The v0.21.0 tag, release workflow, four archives, checksums, extracted version, and installed
  binary all identify the recorded candidate.
- [x] Preview graduation remains a separate deferred decision, and unfinished navigation design or
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

## Pre-merge automated evidence (2026-09-13)

Host and pinned-container validation both passed against clean commit
`48844c634be7f8bc9020fd73fbed5be5f59b3556`. Each run completed the focused and full race suites,
format/tidy and generated-artifact checks, lint with zero issues, the package vulnerability scan,
planning integrity, GoReleaser configuration, and an isolated four-platform snapshot. The final
release candidate must repeat both commands after this planning evidence lands on `main`.

## Merged prerequisite evidence (2026-09-13)

[PR #236](https://github.com/andy-esch/taskflow/pull/236) merged the one-hop closeout, graph input-limit
regressions, and reusable release gate to `main` at
`a1789c64b952f92678d7dc9a319b92045074534a`. The release checkpoint remains in progress until that
merged candidate passes the recorded automated and manual qualification steps and publication is
verified.

Both `just release-validate` and `just release-validate-container` subsequently passed against a
clean detached clone of that exact merge commit. Each completed the full shared gate and four-target
snapshot without changing the candidate. This satisfies the automated qualification criterion;
manual TUI dogfood, release notes, and publication verification remain open.

## Final candidate qualification (2026-09-13)

Candidate `e6a9c8078ada5798f3f618e64cacf614fbea38e2` passed both
`just release-validate` and `just release-validate-container`. Each gate completed focused and full
race tests, format/tidy and generated-artifact checks, lint, vulnerability scanning, planning
integrity, GoReleaser validation, and four-target snapshot creation without changing the candidate.

A fresh `tskflwctl v0.20.0-104-ge6a9c80` then exercised a copied touring-bike space. Full Mermaid,
one-hop Mermaid, and two-hop JSON graphs rendered with healthy projections and truthful bounded
scope. The lifecycle guard correctly refused premature completion; after satisfying the fixture's
remaining criteria through `task ac`, completion and start receipts updated the Thread to the wheel
task in flight with no additional frontier. The refreshed one-hop JSON carried that exact status and
final planning lint passed. The maintainer separately ran the current Threads and spatial views in a
real terminal and reported the smoke test looked great. No release-blocking finding was discovered.

## Publication verification in progress (2026-09-13)

Tag `v0.21.0` resolves to the qualified candidate
`e6a9c8078ada5798f3f618e64cacf614fbea38e2`. The
[tag-triggered workflow](https://github.com/andy-esch/taskflow/actions/runs/34779224449) completed
successfully, and the [GitHub Release](https://github.com/andy-esch/taskflow/releases/tag/v0.21.0)
publishes `checksums.txt` plus Darwin/Linux archives for amd64 and arm64. Fresh downloads of all four
archives passed every published SHA-256 check, and the extracted Darwin arm64 binary reports
`tskflwctl 0.21.0`.

The generated release body is still an uncurated commit dump, so the preview/release-notes criterion
remains open. The installed `/Users/andyeschbacher/go/bin/tskflwctl` also still reports the candidate
development version rather than `0.21.0`; installation verification remains open with it.

## Release closeout (2026-09-13)

The published release body now replaces the generated commit dump with concise spatial-Threads
highlights, tag-stable planning and design-review links, the exact bounded Thread neighborhood, and
an explicit preview boundary. The checksum-verified Darwin arm64 artifact was installed at
`/Users/andyeschbacher/go/bin/tskflwctl`; its SHA-256 is
`354a7ae08e313cf13daac5ac819c16cea4e428e17104e599e8310484f766669a`, byte-compares equal to the
downloaded artifact, and the installed command reports `tskflwctl 0.21.0`.
