---
schema: 1
id: 6gkt3fpsfdd7
status: completed
epic: 21-code-quality-architecture-hardening
description: Retain selected dependency and Thread endpoint IDs while recomputing authorized changes from each fresh guarded snapshot.
effort: 1-2 days
tier: 2
priority: high
autonomy_level: 2
tags: [correctness, identity, graph, threads, concurrency]
created: "2026-10-14"
audit_sources: [2026-10-14-task-correctness-detour-implementation-claude]
updated_at: "2026-10-16"
started_at: "2026-10-15"
completed_at: "2026-10-16"
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

- [x] Dependency and Thread retries retain all first-selected endpoint IDs under rename, alias reuse, deletion and ID-shaped labels; no other task or Thread is modified.
- [x] Fresh eligibility, membership, dependencies and impacts are recomputed rather than replaying a stale write set; new unrelated changes survive.
- [x] Missing/colliding selected endpoints fail closed, including previews/no-ops; initial human-reference ergonomics stay unchanged.
- [x] Pre-planner contention does not invent selection, later empty results do not erase binding, and committed/partial outcomes are never blindly retried.
- [x] Real-store probes and portable planner tests cover both reproduced wrong-target cases, and the inventory explicitly distinguishes bulk/repair paths.
- [x] Owning port/retry guidance, tests and planning evidence agree; normal validation passes.

## Sequencing

[Task lifecycle binding](6gkmexz4vh23-bind-task-mutation-retries-to-strict-source-identities.md)
provides a small related example. This work remains independently actionable: use its
principle, not a stale whole-plan replay. No dependency on broad integrity research.

## Implementation and evidence

Dependency add/remove retain canonical dependent/prerequisite IDs from the first
successful selection inside the graph guard. Each retry verifies those exact tasks
and recomputes changes from current dependencies. Thread creation/membership/lifecycle
retain only endpoint/operation intent inside the successful planner callback, before
adapter validation: even a creation-ID collision returning an empty plan cannot
restart member-reference matching. Fresh adapter validation still owns policy,
membership, impacts and persistence; committed/durable-prefix retry stops remain.

- [Portable retry tests](../../internal/core/graph_retry_target_test.go) cover
  contention before planning, successful selection followed by an empty conflict
  result, another pre-planner conflict, and alias drift before the final attempt.
- [Real-store tests](../../internal/store/graph_retry_target_test.go) cover both
  dependency endpoints, Thread/member rename and disappearance, old-alias and
  ID-shaped impostors, concurrent edge/member additions and the pre-validation
  creation-ID collision case. Snapshots exclude unintended sibling writes.
- Owner compiler-valid probes disabling endpoint binding reproduced wrong-target
  failures; restored safeguards pass focused race suites. The empty-result probe
  also caught the first implementation binding Thread intent too late.

Inventory: [bulk apply](../../internal/core/service_thread_apply.go) already reuses
an explicit repository-bound `ThreadApplyPlan`; [graph repair](../../internal/core/dependency_repair.go)
uses explicit source-declaration selectors (or recomputes deliberate auto-cleanup),
not task/Thread human aliases. Global legacy migration has no selected-reference
intent. These paths retain their existing recovery and prefix tests, not a new
universal retry abstraction.

Member of [Finish selected identity safety](../threads/6gm4h71hm1yt-finish-selected-identity-safety.md).
Owner build/full race tests/lint/docs checks pass. Claude's review verified the retry
binding and mutation discrimination. Its repair hypothesis exposed a separate exact
slug-selector retargeting bug, now owner-reproduced and tracked in
[the graph-repair follow-up](6gm9ny98zyf2-pin-explicit-graph-repair-source-selections-across-retries.md).
The earlier inventory exclusion was too broad: repair is not fuzzy, but exact human
slug selectors can still change source across attempts. That implementation is not
included here. Antigravity's report is deferred for inaccurate evidence. This task
is completed for PR delivery with all criteria met, not a claimed pair of independent
approvals.
