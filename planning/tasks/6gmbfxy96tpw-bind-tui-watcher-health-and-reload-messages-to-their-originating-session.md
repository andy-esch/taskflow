---
schema: 1
id: 6gmbfxy96tpw
status: ready-to-start
epic: 21-code-quality-architecture-hardening
description: Reject stale watcher and debounce callbacks after Atlas workspace swaps without corrupting current reload health or subscriptions.
effort: half day
tier: 2
priority: medium
autonomy_level: 3
tags: [tui, watcher, concurrency, correctness]
created: "2026-10-16"
---
# Bind TUI watcher health and reload messages to their originating session

## Objective

Keep TUI reload health and subscriptions bound to the current watcher after an
Atlas workspace swap. A delayed event or reconciliation from a replaced watcher
must not change the new workspace's health or re-arm listeners/debounce work there.

## Scope

- Reproduce October 16 concurrency audit L1 deterministically before fixing it.
  Owner inspection confirms both message types currently lack resource identity;
  the particular queued-fsnotify-event race remains unproven until exercised.
- Bind `fsEventMsg` and `watcherReconciledMsg` to their originating watcher/session;
  reject stale results before health, dirty generation, subscriptions or reloads change.
- Audit debounce results too: `dirtyGen` resets on workspace activation and alone
  may not distinguish callbacks from different sessions. Preserve valid current
  watcher recovery and no-watcher startup behavior.
- Keep I/O in commands and use existing session/generation idioms rather than
  introducing a general event bus or watcher redesign.

## Acceptance criteria

- [ ] Stale event/reconciliation messages after workspace activation cannot change health, dirty generation or issue work for the new watcher.
- [ ] Cross-session debounce collisions cannot refresh or overwrite the current workspace's watcher health.
- [ ] Current watcher healthy/degraded/unavailable recovery still works, including manual refresh and repeated swaps.
- [ ] Focused race tests and the normal quality gates pass.

## Evidence and sequencing

[Concurrency audit L1](../audits/6gm8g0f66k80-2026-10-16-concurrency-and-atomicity.md).
Join Refine Thread and TUI navigation as an independent correctness follow-up;
there is no implementation dependency on graph-repair retry selection.
