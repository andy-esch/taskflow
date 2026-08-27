---
schema: 1
id: 6g3q4rt7mgjn
status: next-up
epic: 30-threads-and-task-dependency-graphs
description: Expose safe repository-global dependency changes, blockers, downstream impact, dry-run, and deterministic explanatory plans.
effort: 3-5 days
tier: 1
priority: high
autonomy_level: 3
tags: [threads, graph, cli, storage]
created: "2026-08-25"
updated_at: "2026-08-27"
---
# Ship guarded dependency mutations and graph queries

## Objective

Expose safe repository-global dependency operations and deterministic read queries over the strict graph foundation.

## Scope

- Add `task depend add/remove`, blocker, downstream-impact, and explanatory plan use cases and CLI surfaces.
- Perform exact-ID resolution, duplicate/self/missing checks, cycle validation, dry-run, and write inside the repository mutation guard.
- Migrate diagnosed legacy dependency fields by resolving slugs to stable IDs and applying the validated union inside the same guard.
- Treat `task edit` as a guarded product path for dependency deltas, or reject the delta with direction to the specialized commands.
- Return stable human and machine receipts that make the global blast radius explicit.

## Acceptance criteria

- [ ] Add/remove are additive/removal-aware; an already-present add is an idempotent skip with a receipt entry, and invalid unions never write.
- [ ] `--dry-run` performs the same authoritative validation without mutation.
- [ ] Blocker, unblocks, and topology results are deterministic; blocker entries distinguish direct/transitive work and carry a reason plus one deterministic shortest path.
- [ ] Machine output contains stable task IDs and attributable validation/conflict errors without graph-library types.
- [ ] Diagnostic queries degrade with explicit problems and graph health; frontier/unblocked selectors return no eligible work on an unsound relevant graph; mutation fails closed.
- [ ] The six live legacy `blocked_by` values migrate from resolvable slugs to stable `depends_on` IDs with atomic-frontmatter/body preservation; missing or ambiguous values write nothing.
- [ ] `task set` cannot mutate `depends_on` even with `--force`, and `task edit` cannot bypass guarded graph validation.
- [ ] Public blocker commands expose separately named causal-closure and
  action-frontier projections, while every authorization path uses derived
  eligibility rather than blocker-list emptiness.
- [ ] Machine-readable schema marks graph-owned dependency fields as unavailable
  to generic set/unset and directs callers to guarded dependency operations.
- [ ] Deep-chain stress establishes a supported graph-depth envelope; replace
  recursive sound derivation if measured repository shapes approach unsafe stack
  or latency bounds.
- [ ] Planner callbacks resolve task IDs and legacy references only from
  supplied immutable graph values, emit deterministic prefix-safe write order,
  and never call a Store method.
- [ ] Every semantic dependency change receives the Service clock through the
  mutation port, stamps updated_at exactly once, and an idempotent skip does not
  advance it.
- [ ] task edit keeps the human editor session outside the repository guard and
  either rejects a dependency delta with direction or reapplies it through the
  guarded dependency use case after fresh validation.

## Stress tests

- Concurrent opposite edges, duplicate adds, absent removals, stale content, malformed repositories, and deep/wide/disconnected graphs.
- Direct dependency commands racing another direct command and a future bulk-apply-shaped writer.
- Concurrent-cycle errors retain validation semantics while explaining that repository state changed after the command began when attributable.

## Sequencing

Requires the strict read foundation and portable mutation guard. Its production commands persist the epic's bootstrap edges and become the first dependency dogfood surface for the remaining tasks.

## Mutation-guard integration amendment (2026-08-27)\n\nThe production mutation callback is snapshot-only. Exact-ID lookup, duplicate/self/missing checks, legacy resolution, and write ordering must be derived from the supplied immutable TaskGraph and pure core validators; callback code cannot reach back into Store or Service. Use the named LoadTaskGraph source for query loading and do not restore the unused ReadTaskGraph service seam.\n\nThe Service passes its clock into MutateTaskGraph. A real semantic change is stamped by store materialization, while an already-satisfied add or remove stays byte-identical. The six-task legacy migration must emit one deterministic prefix-safe sequence even though stable-ID order happens to be safe for its projected-edge replacement shape.\n\nHuman editing remains outside the repository guard. After the editor returns, task edit compares the proposed dependency set and either rejects it with exact task depend guidance or invokes the same guarded dependency use case against a fresh snapshot; the editor process is never held inside callback exclusion.
