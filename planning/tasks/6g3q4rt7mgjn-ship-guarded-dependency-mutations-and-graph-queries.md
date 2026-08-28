---
schema: 1
id: 6g3q4rt7mgjn
status: completed
epic: 30-threads-and-task-dependency-graphs
description: Expose safe repository-global dependency changes, migration, blockers, downstream impact, and deterministic explanatory graph queries.
effort: 5-8 days
tier: 1
priority: high
autonomy_level: 3
tags: [threads, graph, cli, storage]
created: "2026-08-25"
updated_at: "2026-08-28"
depends_on: [6g3q4rst78qy, 6g3q4rt0wzkq]
started_at: "2026-08-27"
completed_at: "2026-08-28"
---
# Ship guarded dependency mutations and graph queries

## Objective

Expose safe repository-global dependency operations and deterministic explanatory queries over the strict graph foundation.

## Scope

- Add guarded `task depend add/remove` operations and the repository-wide `task depend migrate` convergence command.
- Resolve every user task reference inside the authoritative immutable graph snapshot, returning canonical stable IDs while preserving Taskflow's ordinary exact, prefix, and substring matching policy.
- Add `task blockers` with action-frontier and causal-closure projections, `task unblocks` as a downstream-impact query, and `task list --unblocked` as a fail-closed eligibility selector.
- Reuse the existing taskflow-owned topological analysis for later Thread projections; do not add a repository-global `task plan` command in this slice.
- Keep dependency deltas out of generic set/edit paths and return stable human and machine receipts that expose graph health, workspace identity, no-ops, and resumable partial application.

## Command contracts

- `task depend add <task> --on <prerequisite>...` and `task depend remove <task> --on <prerequisite>...` resolve all operands from the guarded snapshot. Adds and removes are set operations: already-present adds and absent removes succeed as explicit idempotent skips.
- `task depend migrate` converts every exactly resolvable legacy `blocked_by`, `dependencies`, and `blocks` occurrence into the canonical `depends_on` union and clears the legacy fields. Present-but-empty legacy keys are occurrences too: lint reports them and migration removes and receipts them. The command is repository-wide in V1, has no per-task selector, and inherits global `--dry-run` and `--json` behavior.
- `task blockers <task>` defaults to the action-oriented frontier. `--causal` selects the full forensic closure, and machine output names the chosen projection.
- `task unblocks <task>` returns every transitive downstream dependent with current lifecycle role, gate, eligibility, and direct/transitive attribution. It reports impact, not a counterfactual promise that completing the source immediately makes every result eligible.
- `task list --unblocked` selects only derived `Eligible` tasks. An unsound relevant graph returns no eligible work together with an explicit diagnosis.
- `task set`, including `--force`, and `task edit` reject canonical or legacy dependency deltas and direct users to `task depend add/remove`. V1 does not reinterpret arbitrary editor-produced graph changes.

## Reference-resolution contract

The guarded planner receives only `*TaskGraph`, so resolution is a pure snapshot operation. It returns one canonical stable ID using the same tiers as ordinary task commands: exact ID or slug, then unique case-insensitive prefix, then unique case-insensitive substring. Missing and ambiguous inputs retain their typed errors and deterministic candidate details. Share or extract the existing matching policy, or pin exact parity with common contract tests; Store resolution before or during the callback is not acceptable.

## Mutation and recovery contract

- The Store owns the canonical-root guard, strict snapshot load, pure planner invocation, source/plan/prefix/final validation, surgical materialization, whole-snapshot and per-target CAS, and atomic per-file replacement.
- Planners use only supplied immutable graph values, emit deterministic prefix-safe write order, and never call Store or Service. Every durable prefix and the final state must remain sound.
- A semantic change receives the Service clock and stamps `updated_at` exactly once. Idempotent skips remain byte-identical. Dry-run performs the same authoritative planning and validation without writing.
- Migration is serialized but is not an all-files rollback transaction. A later apply or CAS failure may leave the sound durable prefix reported by the guard; retry rebuilds from the current snapshot and converges.
- Edge receipts identify canonical dependent/prerequisite IDs and per-edge applied or skipped outcomes, plus `changed`, `dry_run`, and workspace identity. Migration receipts additionally identify planned, applied, skipped, and remaining work.
- A failure after a durable prefix exposes applied task IDs and remaining resumable work through typed errors and structured JSON diagnostics. The existing prose-only generic error payload is insufficient for this case.

## Query and wire contract

- Blocker entries distinguish direct and transitive constraints and carry a stable reason plus one deterministic shortest path.
- Diagnostic reads project canonical and exactly resolved legacy edges, so migration preserves the blocker/downstream meaning of an unchanged repository. They return deterministic results accompanied by graph health and taskflow-owned structured problems; mutations and eligibility selectors still fail closed unless the graph is healthy.
- Machine output contains stable task IDs, the queried task's derived role/gate/eligibility, current derived state for returned tasks, projection names, workspace identity on mutations, and attributable validation/conflict/recovery details. No graph-library type crosses the core, CLI, or wire boundary.
- New envelopes participate in the reflected JSON Schema, schema comments, golden output, and human/JSON parity tests. Exact Go DTO names are implementation details rather than ADR contracts.
- Machine-readable authoring schema continues to mark every graph-owned dependency field unavailable to generic set/unset and directs callers to guarded dependency commands.

## Acceptance criteria

- [x] Snapshot-local reference resolution matches ordinary task resolution for exact IDs/slugs, case-insensitive unique prefixes, unique substrings, missing values, and ambiguity without calling Store.
- [x] Add/remove validate exact canonical endpoints, duplicate/self/missing edges, the proposed union, cycles, every durable prefix, and final graph health before writing.
- [x] Already-present adds and absent removals are successful receipt-bearing no-ops with no byte or timestamp change.
- [x] `--dry-run` executes the same authoritative resolution, planning, and validation with zero replacements.
- [x] `task depend migrate` converts the six live resolvable legacy occurrences with frontmatter/body preservation; unsafe resolution or projected graph state writes nothing.
- [x] Migration remains graph-sound after every injected write failure and retry converges from every durable prefix.
- [x] Human and machine mutation receipts distinguish applied and skipped edges; partial-failure diagnostics preserve applied and remaining migration work.
- [x] `task blockers` defaults to action frontier, `--causal` returns full closure, and both expose deterministic reason/path/directness data with explicit graph health.
- [x] `task unblocks` deterministically reports all transitive downstream dependents and current derived state without claiming counterfactual eligibility.
- [x] `task list --unblocked` filters on derived eligibility and returns no dispatchable work on an unsound relevant graph.
- [x] `task set --force` and `task edit` cannot bypass graph ownership or guarded dependency validation.
- [x] Query and mutation JSON envelopes enter the reflected schema and preserve human/machine semantic parity without graph-library types.
- [x] Semantic changes use the Service clock exactly once; no-ops do not advance `updated_at`.
- [x] Deep-chain stress establishes the supported 4,096-edge structural/frontier envelope; a separate 512-edge test records the quadratic full-path output amplification for causal/downstream queries so any future cap is evidence-driven.
- [x] Production commands persist this epic's bootstrap dependency edges and exercise blockers, unblocks, and unblocked selection against the real planning graph.

## Stress tests

- Concurrent opposite edges, duplicate adds, absent removals, stale source content, malformed repositories, and direct dependency commands racing another command or a future bulk-shaped writer.
- Exact resolver parity across ordinary and guarded paths, including duplicate slugs, case, prefix/substring tiers, and deterministic ambiguity.
- Legacy migration failure after every write prefix, idempotent retry, body/custom-frontmatter preservation, and structured partial-failure output.
- Healthy, degraded, broken, deep, wide, disconnected, reconvergent, and cyclic query fixtures; authorization must never be inferred from an empty blocker result.
- Human, JSON, schema, error-classification, and `task edit` rejection coverage.

## Sequencing and dogfood gate

Requires the strict read foundation and portable mutation guard. Once released, use the production migration command for the six legacy occurrences, persist the epic bootstrap edges through `task depend add`, and use the production query commands while implementing every remaining Threads task. This is the first dependency-DAG dogfood surface; public Thread planning remains in the later Thread entity slice.

## Readiness review disposition (2026-08-27)

The focused Gemini audit is accepted in substance. Its resolver, migration-command, query-default, unblocked-selector, and receipt concerns are owned here rather than split into separate tasks. Its illustrative DTO field lists are not frozen, and its all-files atomicity language is narrowed to the guard's actual sound-prefix and resumable-retry contract. The expanded command, migration, wire, and recovery surface raises the estimate from 3-5 to 5-8 days.

## Implementation review gate (2026-08-28)

The guarded add/remove/migrate commands, explanatory blocker/downstream queries, fail-closed unblocked selector, typed recovery receipts, wire envelopes, and generated command documentation are implemented on `feat/guarded-dependency-operations`. The real planning tree was migrated from all six legacy field occurrences to eight canonical edges, then a second migration returned a byte-preserving no-op. Production commands also persisted this epic's bootstrap sequence: the current slice depends on both completed foundations and the five later slices form one downstream chain.

Tests cover resolver parity, idempotent byte preservation, cycle refusal, every durable prefix of a multi-file migration, convergent retry, raw content preservation, healthy/degraded/broken reads, human/JSON parity, structured failure recovery, a complete 4,096-edge structural/frontier chain, and a measured 512-edge full-path amplification envelope. `go test ./...`, `go test -race ./...`, golangci-lint, module tidiness, reflected JSON Schema validation, generated CLI docs, planning lint, and diff hygiene were clean before independent review; the review amendments below must pass the same gates before closeout.

## Adversarial implementation review disposition (2026-08-28)

The Claude audit found the meaningful gaps. Resolved legacy edges now participate in blocker, downstream, sound-completion, and gate projections instead of only cycle analysis; queried-task state is explicit in human and JSON output; a `blocks`-only interruption test mutation-proofs the dependent-first migration order; duplicate IDs remain ambiguous inside the snapshot resolver; empty legacy keys are diagnosed, migrated, and receipted; human cycle diagnostics are deduplicated; and stale recovery/command guidance names the exact supported commands. Full-path query amplification remains intentionally uncapped, but its quadratic shape is now recorded by a bounded test because the live graph depth is six.

Repairing an already-broken dependency graph is intentionally not smuggled into this task's ordinary mutation guard; it is tracked by `6g4g8gatbnrs`. Post-dependency-mutation state impact is tracked in the lifecycle/eligibility follow-up `6g3q4rte8kc1`, where the shared before/after graph-state receipt shape belongs. Terse `-q` query modes are deferred until dogfooding demonstrates a concrete workflow rather than added speculatively.

Both independent implementation audits are closed with every finding fixed, tracked, or explicitly deferred. Post-review validation is clean: all package tests and the full race suite pass, golangci-lint reports zero issues, module tidiness and diff hygiene are clean, reflected schema/goldens/generated CLI docs are synchronized, and ordinary plus audit planning lint report no issues.
