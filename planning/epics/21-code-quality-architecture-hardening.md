---
schema: 1
status: active
description: 'Keep the multi-adapter core honest with enforceable dependency direction, consumer-owned ports, and trigger-scoped consolidation instead of speculative rewrites.'
priority: medium
tags: [architecture, quality]
created: "2026-06-22"
updated_at: "2026-08-21"
---

# Code quality & architecture hardening

**Goal.** Keep tskflwctl's CLI, Bubble Tea applications, and future served adapter on
one application core without turning filesystem/configuration details into shared global
machinery. Prefer small consumer-owned ports, explicit composition, and executable
dependency rules; make larger seams earn their cost through a real second consumer.

## Why this is its own epic

This work crosses feature epics because it governs the dependencies *between* their
packages. Entity fan-out, TUI reducer shape, wire contracts, configuration, persistence,
and cross-space orchestration each have their own product home; the rules that keep all of
them reusable belong here. It is also the durable home for adversarial-review findings
whose remedy changes a boundary rather than one feature's behavior.

## Foundation already completed

- `internal/core` owns the application services and the store/configuration/cross-space
  ports they consume; concrete Markdown and TOML adapters point inward.
- The machine JSON contract moved from the CLI renderer into neutral `internal/wire`, and
  domain error classes plus lifecycle transitions are shared across adapters.
- Entity descriptors collapsed schema/scaffold metadata fan-out; the TUI lifecycle registry,
  modal stack, generation-stamped reloads, and concern-based file splits removed the main
  reducer and god-file pressure found in the original audit.
- Planning writes have lock-backed CAS, atomic create/replace, resilient one-pass reads,
  generic editor loops, and compile-time interface assertions for the concrete store.
- Configuration editing and cross-space status now have reusable core services over
  consumer-owned ports. The CLI and both TUI contexts already exercise those seams.

## Live work

- [`codify-package-dependency-fitness-rules-and-refresh-the-architecture-baseline`](../tasks/6g28rv8j9d9p-codify-package-dependency-fitness-rules-and-refresh-the-architecture-baseline.md)
  records the actual package graph and turns its stable inward rules into lint checks.
- [`establish-one-reusable-space-registry-application-boundary`](../tasks/6g28rv8jm1g7-establish-one-reusable-space-registry-application-boundary.md)
  is the one consolidation currently earned by multiple registry consumers. It must stay
  separate from opening a planning workspace.
- The remaining ready tasks are bounded correctness or maintainability findings (BOM and
  block-scalar handling, migration fidelity, error classification, test-stream fidelity,
  and top-level audit lint), not a mandate for a package rewrite.

## Trigger map for deferred seams

| Deferred work | Activate when | Why not before |
| --- | --- | --- |
| [`Resolve() -> Workspace`](../tasks/6fgcr2403sjn-reusable-workspace-discovery-seam-lift-init-doctor-fix-off-the-cli.md) | The atlas or served adapter must open arbitrary planning trees, or a second adapter needs init/doctor/fix | Today Cobra is the only workspace composition root; moving the wiring alone adds indirection. |
| [`context.Context` through core/store`](../tasks/6fgcr24023yr-thread-context.context-through-the-core-and-store-ports.md) | A served HTTP request path needs cancellation, deadlines, or tracing | CLI/TUI calls are local and bounded; the broad mechanical retrofit should follow the real request boundary. |
| [`FindingsRollup` web view model](../tasks/6fgq1n002whs-recast-findingsrollup-as-a-composed-service-view-model-for-web-pagination-sort.md) | A web findings page needs independent pagination/filter/sort | Existing CLI/TUI summary consumers already share the required aggregate. |
| [Worktree-anchor cache](../tasks/6g1sbcbfb4mk-cache-worktree-anchoring-if-it-ever-shows-up-in-a-profile.md) | A profile of `status --all` or another loop implicates it | Current measurements put the repeated stats below noise; caching creates invalidation state. |

## Out of scope

- Repackaging the repository, renaming layers, or introducing a framework to resemble a
  textbook architecture.
- Building speculative web/atlas interfaces before their primary adapter is committed.
- Folding repo-scoped configuration and home-scoped registry state into one abstraction
  merely because both persist TOML.
- Treating every direct filesystem/process import as a violation: the composition root,
  test fixtures, terminal clipboard/editor helpers, and narrow `Fixer`/`Linter`/`Layout`
  ports are intentional exceptions documented in `docs/ARCHITECTURE.md`.
- Absorbing product behavior into this epic when a feature or data-model epic owns it.
