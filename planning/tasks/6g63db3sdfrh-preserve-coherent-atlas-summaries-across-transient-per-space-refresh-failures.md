---
schema: 1
id: 6g63db3sdfrh
status: next-up
epic: 18-tui-bubble-tea-interactive-planning-browser
description: Retain coherent per-space Atlas summaries across planner-window contention using structured partial-failure evidence and bounded retries.
effort: 1-2 days
tier: 3
priority: medium
autonomy_level: 3
tags: [tui, atlas, concurrency, architecture]
created: "2026-09-02"
depends_on: [6g5rwjqeh6a6]
updated_at: "2026-09-02"
---

# Preserve coherent Atlas summaries across transient per-space refresh failures

## Objective

Keep the Atlas's last coherent per-space summary visible when a refresh overlaps a transient guarded
mutation, without hiding durable failures or re-reading unrelated spaces indefinitely.

## Context

`SpaceOverviewService.Overview` intentionally contains planning-tree failures in each
`SpaceSummary` so one broken checkout cannot fail the whole Atlas. Today that boundary flattens the
error to `LoadError string`. A planner-window `domain.ErrConflict` is therefore no longer
classifiable by the TUI's shared top-level read retry policy, and simply wrapping `atlasLoadedMsg`
would not solve the actual partial-failure case.

## Scope

- Preserve structured, adapter-neutral per-space failure evidence through the core Atlas projection
  so callers can distinguish transient contention from durable unreadable/broken state without
  parsing text.
- On an Atlas refresh, retain the last coherent summary for a space whose current read hits
  planner-window contention, mark that space visibly stale, and retry the bounded failed work after
  an appropriate quiet period.
- Keep independently successful spaces current; one failed group must not blank or repeatedly
  reload the entire Atlas without a bound.
- Drop obsolete retries across newer Atlas generations and workspace sessions.
- Keep first-load failures and durable per-space failures explicit, and preserve the Atlas rule that
  one bad checkout cannot fail the complete overview.

## Acceptance criteria

- [ ] A real planner-window test with at least two registered spaces proves the contended space
  retains its last coherent summary while the other space remains readable and current.
- [ ] Per-space failures remain structured through core and are rendered without string parsing or
  filesystem-specific policy in the TUI.
- [ ] Transient work is retried with a documented bound and superseded generations/sessions cannot
  land stale results.
- [ ] Repeated contention and non-conflict failures become visible without a retry loop; a later
  manual or watcher refresh can recover.
- [ ] Single-space dashboard/entity/Thread retry behavior remains unchanged.

## Stress tests

- One contended space among several healthy spaces, first Atlas load, repeated contention, a durable
  broken tree, stale retry after a newer refresh, and workspace switching during the quiet period.

## Out of scope

- Eagerly refreshing every registered space on each filesystem event, changing registry routing,
  hiding durable checkout failures, or coupling core projections to Bubble Tea.

## Sequencing

This is the next Thread frontier item and the final bounded hardening gate for the v0.19.0 TUI
preview. It follows the now-completed contention, stable-identity, and Thread view foundations and
does not redefine Thread projection semantics. Land it before the post-release portable-diagnostic
migration so the two tasks do not evolve `SpaceSummary` failure evidence concurrently: this task
owns transient whole-space refresh state; the later work owns record-level load diagnostics.

## Related

- Epic [18-tui-bubble-tea-interactive-planning-browser](../epics/18-tui-bubble-tea-interactive-planning-browser.md)
- Thread [Complete production Threads](../threads/6g503c6pfqeb-complete-production-threads.md)
- Foundation [Thread projections and contention-safe reloads](6g5rwjqeh6a6-wire-thread-projections-into-the-tui-with-contention-safe-reloads.md)
