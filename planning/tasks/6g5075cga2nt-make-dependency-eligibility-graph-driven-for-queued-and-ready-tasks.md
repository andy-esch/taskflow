---
schema: 1
id: 6g5075cga2nt
status: ready-to-start
epic: 30-threads-and-task-dependency-graphs
description: Treat next-up and ready-to-start tasks as equally eligible when their dependency gate is clear, and align every frontier and start path.
effort: 1-2 days
tier: 1
priority: high
autonomy_level: 3
tags: [threads, graph, lifecycle, cli]
created: "2026-08-29"
depends_on: [6g3q4rtmv4ak]
updated_at: "2026-08-29"
---
# Make dependency eligibility graph-driven for queued and ready tasks

## Objective

Correct the shared derived-state contract so graph clearance, rather than a mandatory
`ready-to-start` hop, determines whether pending work is eligible and appears in a frontier.

## Scope

- Broaden the task-graph eligibility derivation to queued (`next-up`) and candidate
  (`ready-to-start`) roles while keeping their lifecycle labels distinct.
- Route `task start`, generic moves into `in-progress`, batch behavior, and TUI actions through the
  corrected guarded authorization without weakening graph-health or source-role checks.
- Align Thread frontier and `task list --unblocked` with the shared eligible predicate; do not
  introduce a second projection-specific readiness calculation.
- Update human and JSON explanations, reflected schema semantics/versioning, generated CLI documentation, architecture guidance, and parity fixtures.

## Acceptance criteria

- [ ] `Eligible` is true for `next-up` and `ready-to-start` tasks only when the authoritative graph
  is healthy and their dependency gate is clear; it remains false for blocked/broken gates and all
  other lifecycle roles.
- [ ] `task start` and every generic or TUI path into `in-progress` accept either pending source
  role, while `--force` bypasses only a blocked dependency gate from either role and never bypasses
  broken graph evidence.
- [ ] Thread frontier and `task list --unblocked` deterministically include clear-gated `next-up`
  and `ready-to-start` tasks, retain queued/candidate role labels, and fail closed with structured
  diagnosis on an unsound graph.
- [ ] `ready-to-start` remains an optional scoping/handoff signal; `task ready` is still supported
  but no longer required before guarded execution.
- [ ] Wire/schema comments, versioning, human output, generated docs, and focused golden/parity tests describe the corrected semantics without changing dependency ownership or persisted task status.
- [ ] The full race-enabled suite, lint, formatting, schema/doc generation checks, planning lint, and diff hygiene pass.

## Stress tests

- Cross `next-up` and `ready-to-start` with clear, blocked, and broken gates; cover direct and
  transitively unsound prerequisites.
- Cover `in-progress`, `deferred`, `completed`, and `deprecated` sources, forced starts from both
  pending roles, same-state operations, and mixed batches.
- Prove Thread frontier and `task list --unblocked` use the same derivation and stable ordering.

## Sequencing

This is a dogfood correction to ADR-0006 after Thread read projections landed. It must precede
guarded Thread membership/lifecycle mutations so later receipts and UX do not build on the narrower
eligibility contract. The existing dogfood Thread cannot add this task through a supported command
yet, so sequence it as an external prerequisite rather than editing Thread membership by hand.
