---
schema: 1
id: 6g4wm2yf6tyj
status: next-up
epic: 30-threads-and-task-dependency-graphs
description: Add authoritative Thread membership and lifecycle writes with committed receipts and concurrency guarantees.
effort: 4-7 days
tier: 1
priority: high
autonomy_level: 3
tags: [threads, graph, lifecycle, cli]
created: "2026-08-29"
depends_on: [6g3q4rtmv4ak]
updated_at: "2026-08-29"
---
# Ship guarded Thread membership and lifecycle mutations

## Objective

Make Thread membership and lifecycle authoritative guarded operations over the persisted Thread documents and repository-global task graph.

## Scope

- Add a narrow guarded Thread-mutation capability for member add/remove and start, complete, abandon, and reopen.
- Reuse the canonical-root guard pattern established by the Thread document slice and add a lock-free surgical Thread-update materializer that preserves existing timestamps, unknown fields, comments, and key order; the creation-only builder is not an existing-document serializer. Never nest task-graph, task-lifecycle, or Thread guarded capabilities.
- Enforce ADR-0006 membership/lifecycle rules from one authoritative task/Thread snapshot and return typed committed-outcome receipts.
- Augment task lifecycle receipts with affected Thread IDs now that Thread documents can be loaded inside the same guarded task transition.
- Ship human/JSON CLI mutation surfaces and explanatory failures; the usage-informed TUI remains in its later dedicated task.

## Acceptance criteria

- [ ] Member add/remove validates task existence, duplicate/idempotent intent, and current Thread mutability inside one canonical-root guard; membership never changes task-owned dependency edges.
- [ ] Start and complete require at least one non-withdrawn member; complete requires every live member to be soundly drained and no member/external path to be broken or inconsistent.
- [ ] Abandon is terminal and membership-immutable; completed Threads require explicit reopen before membership changes; V1 cannot reopen abandoned Threads.
- [ ] Every mutation returns stable member/Thread IDs, before/after lifecycle or projection state, changed/committed outcome, graph consequences, and an explanatory remedy in human and JSON output.
- [ ] A cleanup failure after a durable Thread write returns a typed committed receipt and is never auto-retried, including when the cleanup error wraps conflict.
- [ ] Task lifecycle transitions name affected Thread IDs and completed-Thread consistency changes without making Thread documents own task status or dependency truth.
- [ ] Same-state lifecycle no-ops still validate destination-specific intent and preserve byte-identical Thread content.
- [ ] Real two-store races cover membership versus dependency mutation, Thread complete versus task lifecycle change, and task lifecycle impact versus concurrent Thread mutation under default and failure paths.
- [ ] Raw-file races are caught by whole-snapshot and immediate-target CAS where possible, with the advisory-lock boundary documented accurately.

## Stress tests

- Empty/all-withdrawn membership, shared members, external gates, deferred and deprecated tasks, unsound completed members, completed drift after upstream reopen, terminal immutability, idempotent membership, post-commit cleanup failure, and mixed-batch attribution.
- Prove cooperating writers wait and re-authorize from fresh state; keep raw-editor CAS tests distinct from that guarantee.

## Sequencing

Requires Thread documents, guarded unstarted creation, read projections, and the lock-free creation materializer. Existing-task bulk linking depends on this task so compound apply composes the new surgical membership/lifecycle materializer rather than rebuilding existing Thread files.
