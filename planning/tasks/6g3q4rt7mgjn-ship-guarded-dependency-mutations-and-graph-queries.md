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
---
# Ship guarded dependency mutations and graph queries

## Objective

Expose safe repository-global dependency operations and deterministic read queries over the strict graph foundation.

## Scope

- Add `task depend add/remove`, blocker, downstream-impact, and explanatory plan use cases and CLI surfaces.
- Perform exact-ID resolution, duplicate/self/missing checks, cycle validation, dry-run, and write inside the repository mutation guard.
- Return stable human and machine receipts that make the global blast radius explicit.

## Acceptance criteria

- [ ] Add/remove are additive/removal-aware, idempotency behavior is explicit, and invalid unions never write.
- [ ] `--dry-run` performs the same authoritative validation without mutation.
- [ ] Blocker, unblocks, and topology results are deterministic and distinguish direct from transitive information where exposed.
- [ ] Machine output contains stable task IDs and attributable validation/conflict errors without graph-library types.
- [ ] Lint and queries remain useful for hand-edited broken repositories while mutation fails closed.

## Stress tests

- Concurrent opposite edges, duplicate adds, absent removals, stale content, malformed repositories, and deep/wide/disconnected graphs.
- Direct dependency commands racing another direct command and a future bulk-apply-shaped writer.

## Sequencing

Requires the strict read foundation and portable mutation guard. Its production commands become the first dependency dogfood surface for the remaining epic tasks.
