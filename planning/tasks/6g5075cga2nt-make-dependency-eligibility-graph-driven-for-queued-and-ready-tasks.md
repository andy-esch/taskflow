---
schema: 1
id: 6g5075cga2nt
status: completed
epic: 30-threads-and-task-dependency-graphs
description: Treat next-up and ready-to-start tasks as equally eligible when their dependency gate is clear, and align every frontier and start path.
effort: 1-2 days
tier: 1
priority: high
autonomy_level: 3
tags: [threads, graph, lifecycle, cli]
created: "2026-08-29"
depends_on: [6g3q4rtmv4ak]
updated_at: "2026-08-30"
started_at: "2026-08-30"
completed_at: "2026-08-30"
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

- [x] `Eligible` is true for `next-up` and `ready-to-start` tasks only when the authoritative graph
  is healthy and their dependency gate is clear; it remains false for blocked/broken gates and all
  other lifecycle roles.
- [x] `task start` and every generic or TUI path into `in-progress` accept either pending source
  role, while `--force` bypasses only a blocked dependency gate from either role and never bypasses
  broken graph evidence.
- [x] Thread frontier and `task list --unblocked` deterministically include clear-gated `next-up`
  and `ready-to-start` tasks, retain queued/candidate role labels, and fail closed with structured
  diagnosis on an unsound graph.
- [x] `ready-to-start` remains an optional scoping/handoff signal; `task ready` is still supported
  but no longer required before guarded execution.
- [x] Wire/schema comments, versioning, human output, generated docs, and focused golden/parity tests describe the corrected semantics without changing dependency ownership or persisted task status.
- [x] The full race-enabled suite, lint, formatting, schema/doc generation checks, planning lint, and diff hygiene pass.

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

## Implementation progress (2026-08-30)

The shared `TaskGraphState.Eligible` derivation now admits both queued (`next-up`) and candidate (`ready-to-start`) roles only when the authoritative graph is healthy and the task gate is clear. The roles remain distinct planning metadata. Guarded `task start`, generic moves into `in-progress`, batch receipts, and TUI actions consume that same derivation; Thread frontier and `task list --unblocked` inherit it rather than adding local readiness rules.

Eligibility refusal recovery now asks non-pending work to move to either pending status. Dependency-gate override eligibility is owned by core and projected by wire, eliminating the adapter-side candidate-only check. The status/gate matrix exposed and fixed a related safety mismatch: `--force` previously accepted any non-clear gate after a healthy repository scan, which could include a locally broken gate such as a withdrawn prerequisite. It now bypasses only `blocked`; broken evidence remains non-overrideable with an explanatory remedy.

Focused coverage crosses queued and candidate work with clear, blocked, broken, and forced paths; excludes in-flight, parked, completed, and withdrawn roles; and proves CLI, TUI, Thread frontier, and `--unblocked` parity. Machine schema 1.55 and reflected descriptions publish the semantic change, and generated CLI/README/architecture guidance is current. Validation passes the full race-enabled suite, golangci-lint, `go vet`, module tidy diff, deterministic CLI-doc regeneration, schema/comment and golden drift tests, planning lint, and diff hygiene.

## Adversarial review closeout (2026-08-30)

Two independent implementation audits were reviewed against the code rather than accepted at face value. The Antigravity audit contained only three intentional-behavior observations (canonical create-and-start seed shape, exclusion of in-flight work from dispatch frontiers, and established per-item batch atomicity); those dispositions were independently confirmed. Claude identified five actionable gaps now fixed in this slice: Thread frontier schema wording names both graph and projection health, broken Thread-local evidence has regression coverage, pending-role blocker classification has one owner plus a totality test over every persisted status, the force-policy rationale and override_allowed schema are explicit, and thread list says eligible rather than ready. Its create-and-start observation was rejected as a public lifecycle defect because ready-to-start is only the canonical ephemeral seed shape and is never persisted; comments and validation wording now make that boundary explicit. Both audits are closed with finding-level resolutions written through audit finding.
