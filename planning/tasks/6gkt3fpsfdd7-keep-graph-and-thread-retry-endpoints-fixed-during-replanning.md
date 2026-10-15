---
schema: 1
id: 6gkt3fpsfdd7
status: ready-to-start
epic: 21-code-quality-architecture-hardening
description: Retain selected dependency and Thread endpoint IDs while recomputing authorized changes from each fresh guarded snapshot.
effort: 1-2 days
tier: 2
priority: high
autonomy_level: 2
tags: [correctness, identity, graph, threads, concurrency]
created: "2026-10-14"
audit_sources: [2026-10-14-task-correctness-detour-implementation-claude]
updated_at: "2026-10-14"
---
# Keep graph and Thread retry endpoints fixed during replanning

## Objective and evidence

Replan from fresh authoritative graph/Thread snapshots without changing the records
the user selected on an earlier attempt.

Owner real-store probes renamed `alpha` to `omega` in the dependency pre-verify seam,
then observed `AddTaskDependencies("alpha", ["prerequisite"])` adding the edge to the
pre-existing `alpha-two` instead. A Thread lifecycle probe likewise renamed
`alpha-thread` and observed `StartThread("alpha-thread")` start `alpha-thread-two`.
Both returned nil and changed the sibling. The repository guards/CAS correctly detected
the first stale snapshot; retry-time reference resolution changed the intent afterward.
Probe: `/private/tmp/taskflow-target-neighbors.VNxBO5/neighbor_retry_probe_test.go`.

## Scope and constraints

- Dependency add/remove: retain both dependent and prerequisite IDs after selection.
- Thread membership/lifecycle and create-time member refs: distinguish selected IDs
  from first-attempt human refs; inventory bulk apply/repair's explicit-ID paths too.
- Resolve first inside the authoritative guard, never by a preliminary unguarded scan.
  Retain semantic endpoint identity, not stale YAML, complete write sets, authorization
  or impacts. Recompute policy and requested changes from each fresh snapshot.
- Preserve committed/partial-prefix retry stops, warning attribution, no-op/preview
  behavior, source-set evidence and explicit mutation authorization.
- Do not retrofit a universal transaction/retry framework or change public resolution
  semantics. Stop for a design decision if a broader contract change is necessary.

## Acceptance criteria

- [ ] Dependency and Thread retries retain all first-selected endpoint IDs under rename, alias reuse, deletion and ID-shaped labels; no other task or Thread is modified.
- [ ] Fresh eligibility, membership, dependencies and impacts are recomputed rather than replaying a stale write set; new unrelated changes survive.
- [ ] Missing/colliding selected endpoints fail closed, including previews/no-ops; initial human-reference ergonomics stay unchanged.
- [ ] Pre-planner contention does not invent selection, later empty results do not erase binding, and committed/partial outcomes are never blindly retried.
- [ ] Real-store probes and portable planner tests cover both reproduced wrong-target cases, and the inventory explicitly distinguishes bulk/repair paths.
- [ ] Owning port/retry guidance, tests and planning evidence agree; normal validation passes.

## Sequencing

[Task lifecycle binding](6gkmexz4vh23-bind-task-mutation-retries-to-strict-source-identities.md)
provides a small related example. This work remains independently actionable: use its
principle, not a stale whole-plan replay. No dependency on broad integrity research.
