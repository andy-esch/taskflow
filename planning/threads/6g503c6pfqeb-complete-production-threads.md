---
schema: 1
id: 6g503c6pfqeb
status: in-progress
description: Deliver production Threads through the CLI preview and a usage-informed TUI
goal: Preserve the production graph foundation while carrying Threads into a faithful TUI
created: "2026-08-29"
tags: [threads, dogfood]
tasks: [6g3q4rtv8d0a, 6g3q4rv1w9e2, 6g3q4rv89vzw, 6g4g8gatbnrs, 6g4wm2yf6tyj, 6g5075cga2nt, 6g5f1d23jy1b, 6g5fthzwbeq1, 6g5fy1m967ka, 6g5gbk5a5bt0, 6g5m69wpydzw, 6g5rwjqeh6a6, 6g5rwjqr7rt8, 6g5rwjr0dz4p, 6g5rxq17px59, 6g5rxq1g5mp1, 6g5rxq1ravd3, 6g5ryqqx5ab7, 6g63db3sdfrh, 6g697mp8s4tx, 6g6dw5js81f3, 6g6jqqcdehne, 6g6scc9jgxae, 6g6wdvfjdaaa, 6g6wdvfp2ksa, 6g721tvf4crh, 6g721vewvvrz]
updated_at: "2026-09-05"
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
presentation passed a clean-build dogfood run. Stable identity, watcher recovery, portable Thread
reads, contention-safe projection loading, list/detail navigation, and a compact wave view have now
landed. Graph-health reporting and coherent Atlas refresh recovery form the bounded v0.19.0 TUI
preview checkpoint.

After that release, the graph deliberately fans out: guarded repair turns diagnosis into recovery;
one shared diagnostic vocabulary precedes portable board/status projections; and a low-priority
spatial graph prototype tests whether two-dimensional navigation earns a production slice. Those
are independent branches. Repository-global dependencies record the release boundary without
pretending the branches technically depend on each other.

Two additional post-release members close planning gaps exposed by that fan-out. One makes frontier
output carry priority, tier, and effort without redefining eligibility; the other defines objective
compatibility and graduation gates so this Thread has an evidence-based finish line rather than an
indefinite preview label.

The guarded-repair branch was split after adversarial design review. Unreadable task sources first
gain opaque revision evidence for whole-snapshot CAS; a lossless source-declaration projection then
preserves raw canonical and legacy ownership during simulation; only then does the user-facing
repair planner and mutation path resume. These are real dependency edges rather than descriptive
phases, so the Thread frontier now exposes the source-revision task as the next repair slice while
the other independent post-preview branches remain eligible.

The deprecated combined TUI member remains as planning history. Its replacement tasks and the
foundation gaps found while scoping them are members of this Thread, while repository-global
dependencies remain the sole source of execution order.
