---
schema: 1
id: 6g4wm2yf6tyj
status: completed
epic: 30-threads-and-task-dependency-graphs
description: Add authoritative Thread membership and lifecycle writes with committed receipts and concurrency guarantees.
effort: 4-7 days
tier: 1
priority: high
autonomy_level: 3
tags: [threads, graph, lifecycle, cli]
created: "2026-08-29"
depends_on: [6g3q4rtmv4ak, 6g5075cga2nt]
updated_at: "2026-08-30"
started_at: "2026-08-30"
completed_at: "2026-08-30"
---
# Ship guarded Thread membership and lifecycle mutations

## Objective

Make Thread membership and lifecycle authoritative guarded operations over the persisted Thread documents and repository-global task graph.

## Scope

- Add a narrow guarded Thread-mutation capability for member add/remove and start, complete, cancel, and reopen.
- Reuse the canonical-root guard pattern established by the Thread document slice and add a lock-free surgical Thread-update materializer that preserves existing timestamps, unknown fields, key order, and comment content (while the shared YAML editor may normalize inline-comment spacing); the creation-only builder is not an existing-document serializer. Never nest task-graph, task-lifecycle, or Thread guarded capabilities.
- Enforce ADR-0006 membership/lifecycle rules from one authoritative task/Thread snapshot and return typed committed-outcome receipts.
- Augment task lifecycle receipts with affected Thread IDs now that Thread documents can be loaded inside the same guarded task transition.
- Ship human/JSON CLI mutation surfaces and explanatory failures; the usage-informed TUI remains in its later dedicated task.

## Acceptance criteria

- [x] Multi-member add/remove is atomic per command: validate every task and current Thread mutability inside one canonical-root guard, reject the whole request before writing on any invalid intent, and report add-existing/remove-absent entries as idempotent no-ops; membership never changes task-owned dependency edges.
- [x] Start and complete require at least one non-withdrawn member; complete requires every live member to be soundly drained and no member/external path to be broken or inconsistent.
- [x] Cancel may enter from unstarted or in-progress, is terminal and membership-immutable, and never mutates member tasks; completed Threads require explicit reopen before membership changes, and V1 cannot reopen cancelled Threads.
- [x] Every mutation returns stable member/Thread IDs, before/after lifecycle or projection state, changed/committed outcome, graph consequences, and any applicable explanatory remedy in human and JSON output.
- [x] A cleanup failure after a durable Thread write returns a typed committed receipt and is never auto-retried, including when the cleanup error wraps conflict.
- [x] Task lifecycle transitions compare before/after projections and name every affected Thread—including shared membership, downstream-member, and external-gate effects—and completed-Thread consistency changes without writing Thread documents or making them own task status or dependency truth.
- [x] Replace the provisional pre-writer `abandoned` domain/wire/schema token with `cancelled`, bump and regenerate machine contracts, and make lint diagnose any hand-authored legacy value with an actionable repair rather than preserving two canonical synonyms.
- [x] Same-state lifecycle no-ops still validate destination-specific intent and preserve byte-identical Thread content.
- [x] Real two-store races cover membership versus dependency mutation, Thread complete versus task lifecycle change, and task lifecycle impact versus concurrent Thread mutation under default and failure paths.
- [x] Raw-file races are caught by whole-snapshot and immediate-target CAS where possible, with the advisory-lock boundary documented accurately.

## Stress tests

- Empty/all-withdrawn membership, shared members, external gates, deferred and deprecated tasks, unsound completed members, completed drift after upstream reopen, cancelled/completed immutability, atomic and idempotent membership batches, post-commit cleanup failure, and projection-impact attribution.
- Prove cooperating writers wait and re-authorize from fresh state; keep raw-editor CAS tests distinct from that guarantee.

## Sequencing

Requires Thread documents, guarded unstarted creation, read projections, the lock-free creation
materializer, and task `6g5075cga2nt`'s graph-driven eligibility correction. Existing-task bulk
linking depends on this task so compound apply composes the new surgical membership/lifecycle
materializer rather than rebuilding existing Thread files.

## Design decisions (2026-08-30)

The first production Thread made the lifecycle vocabulary concrete. A Thread is a sequenced project:
`complete` is its successful outcome and remains entity-qualified as `thread complete`; `cancel` is
the explicit unsuccessful terminal outcome and may stop either scoped-but-unstarted or active work.
Deferred member tasks remain parked, not terminal, so they prevent successful Thread completion.

Membership commands accept multiple task IDs but mutate the one owning Thread document atomically.
Task lifecycle impact is broader than direct membership: the receipt reports every Thread whose
derived projection changes across the guarded before/after snapshots, while leaving all Thread
documents untouched.

## Implementation progress (2026-08-30)

Implemented the pure membership/lifecycle policy, guarded `ThreadMutationStore`, surgical
existing-document materializer, service and CLI verbs, human/JSON receipts, structured policy and
post-commit failures, and task-lifecycle Thread projection impacts. Machine schema 1.56 replaces
the provisional `abandoned` token with `cancelled`; ordinary lint explains how to repair a legacy
value. README, architecture guidance, generated CLI reference, JSON schema comments, and golden
contracts now describe the shipped boundary.

Stress coverage includes atomic/idempotent batches, empty/all-deprecated/deferred membership,
terminal immutability, same-state revalidation, external gates and completed drift, shared/direct/
downstream projection attribution, post-commit cleanup wrapping conflict, planner re-entry,
whole-snapshot and immediate-target raw races, and three real independent-store serialization
races. Dogfooding used the new verb to start `complete-production-threads`; its Thread lifecycle is
now `in-progress` without changing any member task.

Validation: `go test ./...`; race-enabled core/store/wire/CLI tests; `go vet ./...`;
`go mod tidy -diff`; `golangci-lint run ./...`; generated schema comments and CLI docs; planning
lint; and `git diff --check` are clean. The implementation is ready for adversarial review before
task completion.

## Adversarial review closeout (2026-08-30)

Two independent implementation audits are closed with finding-level resolutions. Claude found one
real completion-policy defect: a deprecated member with a broken prerequisite could block an
otherwise drained Thread. Completion now excludes withdrawn members from that final evidence gate,
with regression coverage for the real projection and focused coverage for the retained defensive
live-member/external-gate checks. Documentation now promises comment-content preservation while
acknowledging the shared YAML editor's inline-spacing normalization.

Antigravity's isolated-task-creation observation is tracked by `6g3q4rtv8d0a`: existing-task bulk
linking does not invent creation impacts, while any future compound `new_task` flow must derive a
single before/after projection across creation, dependencies, and membership. Its CAS observation
was affirmative evidence rather than a defect and was closed without code churn.

Post-review validation: `go test ./...`; race-enabled core/store/CLI/wire tests; `go vet ./...`;
`go mod tidy -diff`; `golangci-lint run ./...`; planning and audit lint; and `git diff --check` all
pass.
