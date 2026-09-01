---
schema: 1
id: 6g5rxq1g5mp1
status: next-up
epic: 18-tui-bubble-tea-interactive-planning-browser
description: Discover and watch configured entity directories created or replaced after TUI launch, including a previously absent threads directory.
effort: 1-2 days
tier: 3
priority: medium
autonomy_level: 3
tags: [tui, watcher, threads, correctness]
created: "2026-09-01"
---

# Keep TUI live reload healthy when entity directories appear

## Objective

Keep live reload truthful when a configured entity directory is absent at TUI startup and appears
later, or when a watched directory is atomically replaced. This is now load-bearing for Threads:
launching against a pre-Thread planning tree must not require a TUI restart before newly created
Thread documents become visible.

## Current constraint

The layout correctly returns all configured leaf paths, including `threads/`, but `newWatcher`
silently skips any path that cannot be added. If at least one other entity directory is watchable,
the footer reports live reload as available forever; creating the missing directory later does not
attach a watch, and replacing a watched directory can leave the watcher bound to the old inode.

## Scope

- Track desired normalized watch paths separately from the subset currently attached.
- Observe the nearest appropriate existing parent for unavailable configured leaves and reconcile
  watches when a desired directory is created, renamed, removed, or replaced; avoid teaching the
  TUI the planning layout beyond paths supplied by `core.Layout`.
- Keep ordinary entity-file event bursts on the existing debounce/reload path. Directory-watch
  reconciliation may produce a reload nudge but must not create an event loop.
- Surface partial watch degradation honestly and recover the healthy state when all desired paths
  become available. Manual `r` remains the fallback when no safe watch can be established.
- Close parent and leaf watches during workspace switches and program shutdown so stale sessions do
  not keep delivering events.

## Acceptance criteria

- [ ] Starting with no `threads/` directory and then creating a Thread causes the active workspace
  to reload without restarting the TUI; a second change inside that directory is also observed.
- [ ] Removing and recreating or atomically replacing a configured entity directory reattaches the
  desired watch and yields current data rather than remaining bound to the old inode.
- [ ] Partial attachment failure is represented as degraded live reload, not an unqualified healthy
  watcher, and recovery clears that degradation.
- [ ] Path normalization prevents duplicate watches for symlink/alias spellings while workspace
  switching closes every watch owned by the previous session.
- [ ] Event coalescing remains bounded: directory reconciliation and the ensuing file events do not
  cause reload storms or busy retry loops.

## Stress tests

- Missing leaf at startup, nested parent creation, permission/add failure followed by recovery,
  directory rename/recreate, atomic replacement, rapid create/write/rename bursts, symlinked planning
  roots, workspace switch during reconciliation, and watcher shutdown with pending events.

## Out of scope

- Recursive watching of the repository, watching `.git`, periodic polling, operating-system-specific
  mutation locking, or making live reload mandatory for the TUI to run.

## Sequencing

Independent watcher hardening discovered while scoping Thread reads. The contention-safe Thread
projection task depends on it because that task promises coherent reloads for task and Thread files,
including planning spaces that predate `threads/`.

## Related

- Epic [18-tui-bubble-tea-interactive-planning-browser](../epics/18-tui-bubble-tea-interactive-planning-browser.md)
- Downstream [contention-safe Thread projection loading](6g5rwjqeh6a6-wire-thread-projections-into-the-tui-with-contention-safe-reloads.md)
- Thread [Complete production Threads](../threads/6g503c6pfqeb-complete-production-threads.md)
