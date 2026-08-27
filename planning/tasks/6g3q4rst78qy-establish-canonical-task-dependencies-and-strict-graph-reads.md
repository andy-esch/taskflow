---
schema: 1
id: 6g3q4rst78qy
status: completed
epic: 30-threads-and-task-dependency-graphs
description: Model canonical depends_on data and deterministic strict graph snapshots, derived gate state, legacy diagnostics, and bounded library comparison.
effort: 3-5 days
tier: 1
priority: high
autonomy_level: 3
tags: [threads, graph, storage, migration]
created: "2026-08-25"
updated_at: "2026-08-27"
started_at: "2026-08-26"
completed_at: "2026-08-27"
---
# Establish canonical task dependencies and strict graph reads

## Objective

Introduce the production read foundation for one planning-repository task DAG without exposing graph mutations yet.

## Scope

- Model `depends_on` as a sorted, duplicate-free set of stable task IDs on `domain.Task` and in the task field/schema contracts.
- Define a small taskflow-owned graph analyzer and strict repository snapshot distinct from resilient repair-oriented listing.
- Diagnose—but do not rewrite—the live legacy `blocked_by`, `dependencies`, and `blocks` vocabulary, including slug-to-ID resolution failures and ambiguity.
- Implement deterministic validation, cycle paths, memoized sound completion/gate derivation, lifecycle roles, blocker/downstream traversal, and topological waves.
- Define blocker records with stable reason tokens and one deterministic shortest explanatory path.
- Prevent generic `task set`, including `--force`, and unguarded `task edit` from introducing or changing `depends_on`.
- Run the bounded owned-versus-`dominikbraun/graph` bake-off and record the decision; retain only the implementation surface justified by that decision.

## Acceptance criteria

- [x] Valid task frontmatter round-trips `depends_on` in stable ID order through domain, store, schema, and wire-facing task representations where applicable.
- [x] A strict snapshot identifies malformed/unreadable tasks, ID drift, unknown status, duplicate/self/missing dependencies, and attributable cycles for fail-closed consumers while diagnostic consumers retain the problem list.
- [x] Existing resilient list/lint repair behavior remains available and reports every graph problem deterministically.
- [x] Legacy dependency diagnostics report each resolvable target ID and actionable missing/ambiguous slug failures; no migration write occurs in this task.
- [x] Gate derivation implements broken-over-blocked precedence, treats deferred prerequisites as blocked, and explains unfinished, parked, withdrawn, missing, and unsound-completed blockers.
- [x] Sound completion and derived graph state memoize one result per task per snapshot and are O(V+E).
- [x] Generic set always rejects `depends_on`, including under `--force`; interactive edit cannot land a dependency delta before guarded validation exists.
- [x] The bounded library comparison exercises the same taskflow-owned cases as the owned analyzer; the retained choice and rationale are recorded without preserving a speculative adapter abstraction.
- [x] No public dependency or Thread mutation is introduced by this task.
- [x] Every first-party write path, including lint --fix, task creation, and
  malformed-frontmatter edit, refuses an unvalidated graph-owned field change.
- [x] Cycle analysis uses SCC membership, attributes every cyclic task
  deterministically, and emits one representative path per component without
  duplicate self-edge noise.
- [x] Topological completeness is false on degraded or broken snapshots, and
  resolved legacy edges produce degraded health only when their projected union
  is a legal DAG.
- [x] Unreadable files, invalid references, and duplicate task IDs retain
  path-faithful identity and actionable lint/blocker diagnostics without
  silently dropping a duplicate record's defects.
- [x] Safely resolvable legacy references remain visible as advisory
  ordinary-lint findings with exit zero; unresolved, ambiguous, or unsafe legacy
  references remain validation errors.
- [x] Separate causal and action-frontier blocker projections are explicit;
  eligibility uses derived state, and ordinary lint avoids expanding every
  causal path on deep inconsistent chains.
- [x] The retained analyzer implementation and its tests are reproducible
  in-repository, with non-vacuous exact assertions and adversarial graph shapes.
- [x] Every finding in both 2026-08-26 dependency-foundation audits is marked
  fixed, tracked with a concrete destination, or rejected with recorded
  evidence.

## Stress tests

- Randomized input/map order produces byte-for-byte stable diagnostics and plans.
- Deep chains, wide frontiers, disconnected tasks, duplicate edges, missing IDs, self-edges, and exact cycle paths are covered.
- Reconvergent diamond chains assert bounded visit counts, not merely acceptable wall-clock time.
- Ordinary repository lint reports exactly the six expected degraded legacy-field diagnostics and
  no unaccounted graph defects; the full test, race, formatting, schema, and diff checks pass.

## Sequencing

First production slice. It unlocks guarded dependency writes and supplies the pure role/gate/sound-completion analysis consumed independently by eligibility enforcement and Threads.

## Implementation record (2026-08-26)

- Added canonical `depends_on` persistence and schema/wire projection without adding a public graph
  mutation. Readers retain malformed evidence for lint; valid serialization and outward projection
  use stable ID order.
- Added one immutable strict snapshot with `healthy`, `degraded`, and `broken` health. Deterministic
  problems cover unreadable tasks, missing/drifted/duplicate IDs, invalid status, duplicate/self/
  invalid/missing edges, cycles, and unresolved legacy references. Derived role, gate, sound
  completion, blockers, downstream impact, and topological waves stay behind taskflow-owned types.
- Generic set and interactive edit now reject changes to all graph-owned fields (`depends_on`,
  `blocked_by`, `dependencies`, and `blocks`), including force/unset paths. This prevents supported
  commands from manufacturing invalid graph states; ordinary lint remains the noisy guard for raw
  edits and older binaries.
- Real-repository dogfood loaded 279 tasks in 12–29 ms across two observed runs. Health was
  `degraded`, with zero structural problems and exactly six resolvable legacy `blocked_by` field
  occurrences. Ordinary human and JSON lint now emit those six grouped, stable-ID/edge advisories
  and no unreadable files while exiting zero. Mutation and dispatch remain closed until guarded
  migration in `6g3q4rt7mgjn`.
- The isolated `dominikbraun/graph` v0.23.0 adapter passed the same bounded taskflow-owned cases as
  the owned analyzer. It still required owned deterministic cycle attribution, custom
  wave derivation, and taskflow-controlled shortest-path tie breaking. That is no material code
  reduction, while the upstream API remains explicitly unstable before v1; retain the owned
  O(V+E) analyzer for V1 and revisit libraries only if the analysis surface grows. The temporary
  bake-off module added no project dependency; the repository keeps direct adversarial tests rather
  than a one-implementation adapter interface.
- Stress coverage includes randomized ordering, deep chains, wide/disconnected frontiers,
  duplicate/self/missing/invalid edges, exact and self cycles, all blocker reason tokens,
  deterministic shortest paths, legacy direction/resolution, reopen invalidation, immutable query
  results, and a 120-layer reconvergent diamond with exact visit-count assertions.

## Adversarial hardening record (2026-08-27)

- Replaced back-edge cycle discovery with SCC membership: every cyclic task receives deterministic
  attribution, one edge-following representative path is retained per component, and self-edges do
  not also emit generic cycle noise. Canonical and resolved legacy edges are analyzed as one
  structural union, so a legacy self-edge or cycle is broken rather than degraded.
- Replaced the ambiguous blocker API with `CausalBlockers`, `BlockingFrontier`, and `ExplainGate`.
  Authorization remains `State.Eligible`; lint uses the bounded action frontier. Traversal keeps
  predecessor links and materializes only emitted paths, making the complexity contract explicitly
  output-sensitive.
- Closed the remaining first-party write gaps: task creation rejects dependency-bearing records,
  lint repair skips graph-owned normalization noisily, and an unreadable edit baseline rejects every
  candidate rather than accepting deletion. Safe legacy debt is an advisory in human/JSON lint with
  exit zero; unsafe/unresolved debt remains a validation error.
- Preserved stable-ID identity for unreadable filenames, distinguished missing/unreadable/invalid
  references, and keyed graph lint attribution by source path so both duplicate-ID records retain
  their own defects. The speculative one-implementation analyzer interface and hidden testdata
  contract were removed in favor of direct adversarial tests.
- Both independent audits now have no open findings. Machine-readable field ownership is tracked by
  `6g3q4rt7mgjn`; canonical snapshot-loader consolidation is tracked by `6g3q4rt0wzkq`; the measured
  deep-chain optimization trigger is also explicit on `6g3q4rt7mgjn`.

Validation: full `go test ./...` and `go test -race ./...`; golangci-lint (zero issues); `go vet
./...`; module tidy diff; generated CLI docs; schema-comment freshness; golden machine contracts;
`git diff --check`; both audit lints; and live planning lint all pass. Live planning lint emits
exactly six advisory legacy-field findings, zero unreadable files, and exits zero.
