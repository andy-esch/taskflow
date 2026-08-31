---
schema: 1
id: 6g503c6pfqeb
status: in-progress
description: Deliver the remaining production Thread capabilities
goal: Correct eligibility, then ship guarded mutations, bulk linking, generated views, and the TUI
created: "2026-08-29"
tags: [threads, dogfood]
tasks: [6g3q4rtv8d0a, 6g3q4rv1w9e2, 6g3q4rv89vzw, 6g4wm2yf6tyj, 6g5075cga2nt, 6g5f1d23jy1b, 6g5fthzwbeq1, 6g5fy1m967ka, 6g5gbk5a5bt0, 6g5m69wpydzw]
updated_at: "2026-08-31"
started_at: "2026-08-30"
---

# Thread: Complete production Threads

**Goal.** Correct eligibility, then ship guarded mutations, bulk linking, generated views, and the TUI

## Context

This Thread is the dogfood surface for the production implementation itself. It should make active
work, dispatchable work, external gates, and the next dependency-ordered slice discoverable through
the same commands being built. Real sequencing belongs in repository-global task dependencies;
membership alone does not manufacture an ordering.

The first natural release boundary is v0.18.0 after deterministic graph/plan views and the small
in-flight frontier presentation follow-up are reviewed, merged, and exercised from a clean `main`
build. That release should describe Threads as a CLI preview whose interface may still evolve. The
TUI and advanced graph calculations remain later usage-informed work, not release gates.
