---
schema: 1
id: 6g503c6pfqeb
status: in-progress
description: Deliver production Threads through the CLI preview and a usage-informed TUI
goal: Preserve the production graph foundation while carrying Threads into a faithful TUI
created: "2026-08-29"
tags: [threads, dogfood]
tasks: [6g3q4rtv8d0a, 6g3q4rv1w9e2, 6g3q4rv89vzw, 6g4g8gatbnrs, 6g4wm2yf6tyj, 6g5075cga2nt, 6g5f1d23jy1b, 6g5fthzwbeq1, 6g5fy1m967ka, 6g5gbk5a5bt0, 6g5m69wpydzw, 6g5rwjqeh6a6, 6g5rwjqr7rt8, 6g5rwjr0dz4p, 6g5rxq17px59, 6g5rxq1g5mp1, 6g5rxq1ravd3, 6g5ryqqx5ab7, 6g63db3sdfrh, 6g697mp8s4tx, 6g6dw5js81f3, 6g6jqqcdehne]
updated_at: "2026-09-03"
started_at: "2026-08-30"
---

# Thread: Complete production Threads

**Goal.** Preserve the production graph foundation while carrying Threads into a faithful TUI

## Context

This Thread is the dogfood surface for the production implementation itself. It should make active
work, dispatchable work, external gates, and the next dependency-ordered slice discoverable through
the same commands being built. Real sequencing belongs in repository-global task dependencies;
membership alone does not manufacture an ordering.

The v0.18.0 CLI preview shipped after deterministic graph/plan views and the in-flight frontier
presentation passed a clean-build dogfood run. The remaining work deliberately hardens general TUI
identity and watcher recovery, then separates optional local Thread paths and portable diagnostics
before loading Thread projections into the second primary adapter. Read-only list/detail UX then
ships before any topology view, so real TUI usage—not the availability of a graph projection—decides
the smallest useful terminal presentation.

The deprecated combined TUI member remains as planning history. Its replacement tasks and the
foundation gaps found while scoping them are members of this Thread, while repository-global
dependencies remain the sole source of execution order.
