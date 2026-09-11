---
schema: 1
id: 6g90h4pg0q7n
status: in-progress
description: Make large Thread graphs comprehensible and shared TUI navigation predictable, visible, and reversible.
goal: Ship an evidence-backed large-graph experience and a coherent shared TUI navigation model.
created: "2026-09-11"
tags: [threads, tui, ux, dogfood]
tasks: [6g6wdvfp2ksa, 6g86016qvk5d, 6g86c7y6hn41, 6g86g03zfj2f, 6g87qn72901g, 6g8vxbv3d4xn, 6g8vxcnezktm, 6g914ybs9t87, 6g9150nrt4p9]
updated_at: "2026-09-11"
started_at: "2026-09-11"
---
# Thread: Refine Thread and TUI navigation

## Context

The production Threads foundation, compatibility contract, three preview releases, and first usable spatial graph shipped through [Complete production Threads](6g503c6pfqeb-complete-production-threads.md). Dogfooding has moved the primary uncertainty from graph correctness to comprehension: large DAGs need better overview and focus, while the TUI as a whole needs a teachable model for views, action targets, history, and retreat.

This successor keeps those concerns together because the graph is the sharpest real-world test of the shared navigation shell. Repository-global dependencies remain the source of execution order; membership records the product outcome and preserves the dogfooding loop.

## Finish line

Close this Thread when large Thread graphs have an evidence-backed presentation and navigation model, the selected implementation slices are usable, and shared TUI navigation makes views, targets, and return paths discoverable without memorized hidden modes.

## Sequencing

Begin with the large-graph presentation investigation and whole-TUI information-architecture review. Those design passes may rescope the one-hop focus and alternate-view discoverability tasks before implementation. Reusable Back and structured action-target contracts may proceed where they do not prejudge that design. Frontier planning metadata is an independent, bounded CLI/core usability slice.

Graph-export legibility can proceed independently by making current Mermaid and DOT role styling
self-explaining. Bounded one- and two-hop exports follow the large-graph design pass so their shared
projection seam and omitted-scope treatment are deliberate; once shipped, external PR-curation
tooling can use those review-sized diagrams without coupling GitHub policy into taskflow.

Preview graduation remains a separate deferred release decision, and portable board/status diagnostics remains independent adapter-foundation work. Neither should keep this usability Thread open by membership alone.
