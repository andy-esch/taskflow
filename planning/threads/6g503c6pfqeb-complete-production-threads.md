---
schema: 1
id: 6g503c6pfqeb
status: in-progress
description: Deliver production Threads through the CLI preview and a usage-informed TUI
goal: Preserve the production graph foundation while carrying Threads into a faithful TUI
created: "2026-08-29"
tags: [threads, dogfood]
tasks: [6g3q4rtv8d0a, 6g3q4rv1w9e2, 6g3q4rv89vzw, 6g4g8gatbnrs, 6g4wm2yf6tyj, 6g5075cga2nt, 6g5f1d23jy1b, 6g5fthzwbeq1, 6g5fy1m967ka, 6g5gbk5a5bt0, 6g5m69wpydzw, 6g5rwjqeh6a6, 6g5rwjqr7rt8, 6g5rwjr0dz4p, 6g5rxq17px59, 6g5rxq1g5mp1, 6g5rxq1ravd3, 6g5ryqqx5ab7, 6g63db3sdfrh, 6g697mp8s4tx, 6g6dw5js81f3, 6g6jqqcdehne, 6g6scc9jgxae, 6g6wdvfjdaaa, 6g6wdvfp2ksa, 6g721tvf4crh, 6g721vewvvrz, 6g72ncs4xjdm, 6g7ddeyp773z, 6g7ddfhh2jc2, 6g7fhfpmy032, 6g7j2ebatzyt, 6g86016qvk5d, 6g86c7y6hn41, 6g86e10dztpf, 6g86g03zfj2f, 6g87qn72901g, 6g8btt5hcgs9, 6g8by30btznq, 6g8ezj5e51hg, 6g8k47wg8xks, 6g8vxbv3d4xn, 6g8vxcnezktm]
updated_at: "2026-09-10"
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
one shared diagnostic vocabulary precedes portable board/status projections; and frontier output
gains planning context without inventing a scheduler. Repository-global dependencies record the
release boundary without pretending those branches technically depend on each other.

Two additional post-release members close planning gaps exposed by that fan-out. One makes frontier
output carry priority, tier, and effort without redefining eligibility; the other defines objective
compatibility and graduation gates so this Thread has an evidence-based finish line rather than an
indefinite preview label.

That graduation branch is now explicit: the compatibility contract, v0.18.0/v0.19.0 historical
fixtures, and preview-labelled v0.20.0 hardening checkpoint have shipped. That release unlocks the
tier-1 spatial graph design/prototype intended to seed the next feature release. Graduation is now
deliberately deferred while additional installed preview releases accumulate multi-release soak
evidence. The spatial work remains a removable presentation extension over
`ThreadGraphProjection`, not a second graph model or silent core commitment; that architectural
isolation contains flagship risk without demoting its product importance.

The first spatial prototype has now upheld that boundary in live dogfooding: it consumes the shared
projection, preserves canonical selection through add/rename reloads, and navigates actual edges
rather than assuming adjacent columns are connected. The same pass caught a visually false
long-route junction and ambiguous shell actions. Those findings are explicit downstream members:
dense-route truthfulness and one-hop focus stay in the Threads epic, while reusable Back navigation
structured child-action targeting, and alternate-view/Thread identity discoverability stay in the
general TUI epic. Their shared Thread membership keeps the feedback loop visible without inventing
dependencies among parallel concerns.

The subsequent experience review added one more parallel member for responsive spatial layout.
That task owns complete node/layer viewport units, adaptive label identity, and vertical/pane
budgeting; dense-route hardening owns edge truthfulness at the resulting boundaries. Research did
not establish a numeric size cutoff for node-link views: this dogfooded graph is a sparse 40-node,
51-edge DAG, so the full graph remains a flagship surface to harden, waves remain the ordered scan,
and one-hop focus remains a complementary local lens.

Responsive-layout dogfooding then made the next uncertainty concrete: a healthy large graph can be
fast yet still be difficult to comprehend when only a few complete nodes fit in the viewport.
Prominent node/layer/row scope and a deliberate narrow-terminal card remain bounded corrections in
the responsive task. A new design branch now precedes one-hop implementation to compare full-graph,
overview/focus, progressive-disclosure, and web-extension strategies against real navigation jobs.
In parallel, a whole-TUI information-architecture review is allowed to challenge the growing mix of
entity tabs, alternate views, pane focus, zoom, and history before the narrower view-discoverability
task implements more shell chrome.

Two implementation reviews then found a missing deepest-row long-edge track and Atlas losing
presentation-owned immersive-zoom state, plus unpinned ordering, capacity, and manual-zoom
invariants. Those bounded corrections are consolidated into one tier-1 prototype-hardening member
that precedes dense routing, responsive layout, and one-hop focus. Cascaded external-gate placement
remains inside dense-route hardening rather than duplicating its ownership.

The guarded-repair branch was split after adversarial review. Unreadable task sources first gain
opaque revision evidence for whole-snapshot CAS; unreadable Thread evidence then gains the same
protection before a repair receipt may rely on partial Thread impacts. A lossless source-declaration
projection separately preserves raw canonical and legacy ownership during simulation. Only after
those foundations does the user-facing repair planner and mutation path resume. These are real
dependency edges rather than descriptive phases, so the Thread frontier exposes their order while
the other independent post-preview branches remain eligible.

The deprecated combined TUI member remains as planning history. Its replacement tasks and the
foundation gaps found while scoping them are members of this Thread, while repository-global
dependencies remain the sole source of execution order.
