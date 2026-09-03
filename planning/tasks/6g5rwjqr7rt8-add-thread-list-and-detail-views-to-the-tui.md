---
schema: 1
id: 6g5rwjqr7rt8
status: in-progress
epic: 30-threads-and-task-dependency-graphs
description: Expose Thread navigation, lifecycle, progress, frontier, gates, and inconsistency in responsive read-only TUI list and detail views.
effort: 2-3 days
tier: 3
priority: medium
autonomy_level: 3
tags: [threads, tui, ux, graph]
created: "2026-09-01"
depends_on: [6g5rwjqeh6a6, 6g5rxq17px59]
updated_at: "2026-09-03"
started_at: "2026-09-02"
---

# Add Thread list and detail views to the TUI

## Objective

Make Threads a first-class, read-only TUI destination after their shared projections and reload
behavior are proven. The screen should answer which initiatives exist, what is active, what can
start next, what is gated, and whether completion is graph-sound without hiding those distinctions
behind one optimistic progress number.

## Scope

- Add a Threads tab/list and detail route through the existing entity registry, command palette,
  tab navigation, selection restore, filtering, and sorting conventions. File opening/path copy is
  capability-aware: provide it for local workspaces and do not imply a path for portable readers.
- Make that registry route the single owner of Thread list/detail generations, loaded/error state,
  selection identity, and reloads. Delete or collapse the temporary parallel fields in
  `threadProjectionState`; retain only repository-level Thread diagnostics that genuinely do not
  belong to one row. Fold the temporary `readSurface` distinction into the registry's entity/screen
  identity where that removes duplicate discriminators.
- Entering the Threads route must issue its first lazy list request, after which task and Thread
  watcher events reload that same state owner. A background or restored route must not need a
  second cache to remain live.
- Render persisted lifecycle separately from derived health: status, nominal and drained progress,
  inconsistency, in-flight members, dispatchable frontier members, missing/broken members, and
  immediate external gates must keep their core meanings.
- Show the persisted Thread body and member ordering without implying that membership creates
  dependency edges or that list position authorizes dispatch.
- Start read-only. Thread lifecycle and membership actions remain CLI operations until TUI usage
  demonstrates a coherent mutation workflow worth designing.
- Define deliberate normal, narrow, and very small terminal layouts. Degradation may collapse
  secondary columns, but must retain identity, lifecycle, graph health, and the distinction between
  in-flight, dispatchable pending, and pending work that current graph evidence cannot dispatch.
- Build parity fixtures from the same repository snapshots used by core/CLI projection tests rather
  than recreating graph cases in UI-only form.

## Acceptance criteria

- [x] Threads are reachable as a first-class TUI tab and through command-palette navigation, with
  stable filtering, sorting, selection restoration, and help text consistent with the other
  read-only entities; path copy/open works when locally available and degrades explicitly when not.
- [x] One production integration test enters the registry route and proves one list request, one
  generation/error owner, one canonical detail selection, and one reload path; no parallel Thread
  list/detail state machine remains.
- [x] List rows distinguish lifecycle, sound progress, graph/projection health, and active/eligible
  work without recomputing any of those values in the TUI.
- [x] Detail output faithfully presents members, immediate external gates, in-flight work,
  dispatchable frontier, missing/broken evidence, inconsistency, and the Thread body from the shared
  core projection.
- [x] Completed-but-unsound, cancelled, empty, shared-member, externally gated, and healthy active
  Threads remain visually distinguishable in parity fixtures.
- [x] Normal, narrow, and minimum supported terminal tests prove that essential state remains
  readable and navigation/help do not overflow or become unreachable.
- [x] No TUI-local graph traversal, eligibility rule, direct filesystem read/write, membership
  mutation, or lifecycle mutation is introduced.

## Stress tests

- Large Thread lists, long titles, many members and gates, completed projection drift, missing
  members, shared membership, filter/reload races, background-tab reload, and repeated transitions
  between normal and narrow terminal widths.

## Out of scope

- Node-edge or wave visualization, Thread mutation actions, critical-path/forecasting features, and
  changes to CLI or wire semantics solely to simplify presentation.

## Sequencing

Depends on contention-safe TUI projection loading. Ship and dogfood this list/detail slice before
choosing the smallest useful topology presentation; the next task must not assume a diagram is the
best terminal interface.

## Related

- Epic [30-threads-and-task-dependency-graphs](../epics/30-threads-and-task-dependency-graphs.md)
- ADR [0006 — Adopt Threads as task DAGs](../adrs/0006-adopt-threads-as-task-dags.md)
- Thread [Complete production Threads](../threads/6g503c6pfqeb-complete-production-threads.md)
- Foundation [contention-safe Thread projection loading](6g5rwjqeh6a6-wire-thread-projections-into-the-tui-with-contention-safe-reloads.md)
- Supersedes part of [the original combined TUI slice](6g3q4rv89vzw-add-usage-informed-thread-views-to-the-tui.md)

## Implementation progress (2026-09-02)

Threads now participate as a read-only registry entity with lazy list/detail loading, canonical
selection identity, filtering, sort/restore behavior, palette and command navigation, per-space
session state, and the ordinary contention-safe reload path. The temporary `threadProjectionState`
and its dedicated list/detail messages and read surfaces are gone. Rows retain each shared
`core.ThreadView`; only repository-wide graph and unreadable-record diagnostics remain as typed
metadata on the registry tab.

Responsive rows keep lifecycle, graph/projection health, nominal and sound rollups, and compact
in-flight/frontier/not-dispatchable counts. Detail renders the core projection's members, active work,
dispatchable frontier, immediate external gates, inconsistency evidence, diagnostics, and persisted
body. Thread mutation affordances remain absent. Optional `ThreadPathSource` is resolved beside the
detail read: local workspaces retain path copy, clickable titles, and `$EDITOR`, while portable
readers keep semantic browsing and explicitly report that a local path is unavailable.

Coverage includes the registry ownership path, task-only and Thread-document reloads, split/pathless
adapters, canonical state preservation across filter/sort/reload, normal/narrow/minimum rows, help
layout, and completed-unsound, cancelled, empty, shared-member, externally gated, and healthy active
projections produced by the core projector. The full race-enabled suite and golangci-lint pass;
generated CLI docs and `git diff --check` are clean. Repository planning and audit-finding lint pass.

## Adversarial review closeout (2026-09-03)

Two independent implementation audits—[Claude](../audits/6g6db7yt62r4-2026-09-03-tui-thread-list-detail-implementation-claude.md)
and [Antigravity](../audits/6g6db7z2qfze-2026-09-03-tui-thread-list-detail-implementation-antigravity.md)—confirmed
the single-registry-owner and portable/local capability boundaries, then exposed responsive and
coverage gaps. All findings were fixed in this task and both audits are closed.

Rows now reserve lifecycle, health, progress where space permits, and graph-derived work before an
explicitly budgeted identity tail. Grapheme-aware middle elision keeps long distinct slugs and
duplicate canonical hints distinguishable; tiny layouts compact only counts over 99. Pending work
is partitioned against the authoritative core frontier, so clear-gated members suppressed by
unhealthy repository evidence are reported as pending but not dispatchable rather than disappearing.

Regression coverage directly injects stale list/detail messages, pins the read-only registry/action
contract, distinguishes nominal from sound progress, exercises degraded-graph frontier semantics,
and checks path-loading/pathless inline-edit feedback plus Unicode terminal-cell budgets. No finding
required an out-of-scope follow-up.

Final validation: uncached `go test -count=1 -race ./...`, golangci-lint, generated-doc drift,
module tidiness, planning lint, audit-finding lint, and `git diff --check main` all pass.
