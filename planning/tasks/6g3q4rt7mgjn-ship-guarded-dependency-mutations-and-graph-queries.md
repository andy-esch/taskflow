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
updated_at: "2026-08-26"
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

## Stress tests

- Concurrent opposite edges, duplicate adds, absent removals, stale content, malformed repositories, and deep/wide/disconnected graphs.
- Direct dependency commands racing another direct command and a future bulk-apply-shaped writer.
- Concurrent-cycle errors retain validation semantics while explaining that repository state changed after the command began when attributable.

## Sequencing

Requires the strict read foundation and portable mutation guard. Its production commands persist the epic's bootstrap edges and become the first dependency dogfood surface for the remaining tasks.
