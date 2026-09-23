---
schema: 1
id: 6g7wsr8yvt1a
status: ready-to-start
epic: 21-code-quality-architecture-hardening
description: 'Design pass: reconcile the identity, schema, locking, atomic-write and diagnostic foundations into one sequenced roadmap before spinning out work.'
effort: 1-2 days
tier: 2
priority: medium
autonomy_level: 2
tags: [architecture, integrity, store, design]
created: "2026-09-07"
updated_at: "2026-09-22"
audited: "2026-09-20"
audit_sources: [2026-09-20-weekly-task-sweep, planning/audits/6gc7jd9aq1q9-2026-09-21-arch-failure-and-recovery.md]
---
# Define the next shared entity-integrity foundations

## Objective

Turn the creation-guard survey into one evidence-based foundation roadmap before adding more entity
kinds or independently implementing more invariants. Reconcile existing identity, schema, locking,
atomic-write, and adapter-neutral diagnostic work; investigate the newly exposed gaps; then recommend
a deliberate sequence. This task owns design and scoping, not the implementations it may authorize.

## Existing foundation work to reconcile

- [Frontmatter-schema policy ADR](6fkkz41cax80-adr-close-frontmatter-schema-policy-questions.md),
  especially its unresolved planning-space-wide versus per-kind stable-ID namespace decision.
- [Reserved document-schema enforcement](6g7f0tqgftg3-enforce-reserved-document-schema-versions-across-entity-writers.md)
  after that policy is accepted.
- [Bounded and observable repository-lock acquisition](6g6yjm16dkgt-design-bounded-and-observable-repository-lock-acquisition.md)
  for CLI, TUI, and eventual served adapters.
- [Unified atomic-file-write guarantees](6g63jj1dh0sb-unify-the-three-divergent-writefileatomic-implementations.md)
  across store, config, and userconfig.
- [Adapter-neutral repository lint diagnostics](6g5vm4efjcdv-make-repository-lint-load-diagnostics-adapter-neutral.md)
  before remote or served adapters consume the aggregate lint boundary.
- [Markdown-first storage durability reassessment](6g721w07mv1d-reassess-markdown-first-storage-durability-for-relational-planning-data.md)
  as a later evidence-driven decision, not an immediate rewrite.

## Newly observed questions

1. Ordinary creates and legacy mutations use `writeLock`, which discards repository-guard release
   failures and cannot distinguish pre-commit failure from “write committed, cleanup failed.” Decide
   whether ordinary mutations need typed committed outcomes, a narrower cleanup policy, or an
   explicitly accepted limitation without inducing unsafe conflict retries.
2. The process-wide canonical-root guard registry retains every root for the process lifetime.
   Determine whether reference-counted eviction belongs in the bounded-lock design for long-lived
   multi-space TUI/server consumers, or whether measured usage makes the retained entries acceptable.
3. `createFileAtomic` provides race-free `O_EXCL` creation but can leave a partial new file if the
   process dies after opening the destination. Decide whether lint/repair is the intentional recovery
   contract or whether complete-content exclusive creation needs a separate portable primitive.
4. Decide when a reusable secondary-adapter conformance suite becomes valuable. It should verify
   behavioral identity/atomicity contracts without requiring every adapter to implement filesystem
   locks; avoid building it before a second real adapter makes the shared contract concrete.
5. Guarded mutation services currently spell “did anything become durable?” four ways across seven
   retry loops, while task rename does not retry a pre-commit conflict at all. Decide one shared
   durability predicate and retry contract, including whether rename participates or receives an
   explicit policy carveout, before another mutation family copies an existing pattern.

## Design constraints

- Do not collapse semantically distinct task, Thread, epic, audit, and research creation into a
  universal `CreateEntity` core port. Share persistence mechanics and behavior-level contracts while
  keeping entity authorization explicit.
- Filesystem locking remains secondary-adapter policy. A database or service adapter may satisfy the
  same behavior with native constraints and transactions.
- Graph-aware Thread, dependency, repair, and create-and-start paths retain their pure core planners,
  authoritative snapshots, CAS, and committed/partial-durability receipts.
- Do not create child taskfiles during this design pass. Record proposed ownership and sequencing in
  this task for human review first; spin out only the approved gaps afterward.

## Acceptance criteria

- [ ] Inventory the existing foundation tasks, their decisions, dependencies, and overlapping scope;
      correct stale assumptions without duplicating their implementation work.
- [ ] Produce one decision table for stable-ID ownership, schema compatibility, lock acquisition and
      lifecycle, atomic create/replace guarantees, guarded-mutation durability/retry semantics, and
      adapter-neutral diagnostic evidence.
- [ ] Reproduce or otherwise substantiate each newly observed question and distinguish current CLI/TUI
      risk from future long-lived or remote-adapter risk.
- [ ] Recommend which new concerns should amend existing tasks, become separate tasks, remain measured
      follow-ups, or be explicitly accepted—with reasons and a proposed sequence.
- [ ] Explain why shared adapter mechanics do not imply one generic core entity-creation API, and name
      the behavioral contract a future non-filesystem adapter must satisfy.
- [ ] Present the proposed roadmap for human approval before creating any follow-up taskfiles or
      changing production code.

## Out of scope

- Implementing any of the linked foundation tasks or newly observed changes.
- Choosing a new authoritative storage model before guarded repair and real usage provide evidence.
- Introducing a generic entity framework, plugin schema language, database, or served adapter.

## Related

- [Shared ordinary-create guard](6g7s6hr3qnfq-serialize-research-id-collision-checks-with-creation.md)
- [Architecture](../../docs/ARCHITECTURE.md)
- [Epic 21](../epics/21-code-quality-architecture-hardening.md)
- Audit [2026-09-21 architecture: failure and recovery](../audits/6gc7jd9aq1q9-2026-09-21-arch-failure-and-recovery.md), findings M1 and L1

## Sweep verification (2026-09-20)

Automated weekly sweep re-read every reference in this task against
`internal/store` and `internal/config` at `934e1cf`. No section is obsolete; the
annotations below narrow two of the four newly observed questions.

**Foundation links — all six still exist and are still open** (`ready-to-start`
or `next-up`): the frontmatter-schema ADR, reserved document-schema enforcement,
bounded lock acquisition, atomic-write unification, adapter-neutral lint
diagnostics, and the markdown-durability reassessment. Nothing in this task's
reconcile list has shipped underneath it.

**Related link is stale in one direction:** `6g7s6hr3qnfq-serialize-research-id-collision-checks-with-creation`
is now `completed`. The shared ordinary-create guard it names is landed, so the
roadmap should treat it as settled input rather than as parallel work.

**Q1 — verified accurate.** `FS.writeLock` (`internal/store/lock.go:103`) still
wraps `checkedWriteLock` and discards the release error (`func() { _ = release() }`),
while `checkedWriteLock` (`lock.go:111`) returns it via `errors.Join` for the
control-inverted boundary. The two-tier split the question describes is exactly
what is on disk; the decision is still open.

**Q2 — verified accurate, and the code now states the same question.**
`repositoryGuards.byRoot` (`internal/store/lock.go:28`) is still a process-wide
map with no eviction, and its doc comment already records the intended answer:
"CLI processes normally retain one entry; a future long-lived multi-space
adapter should reference-count and evict idle entries." That is a lean, not a
decision — the roadmap should either ratify it or measure it.

**Q3 — narrower than written.** `createFileAtomicWithMode`
(`internal/store/atomic.go:96`) now removes the destination on every in-process
failure path, including a failed `Close`. The residual exposure is therefore
only *process death* between `O_EXCL` open and `Close`, not ordinary write/sync
errors. `sweepStaleTemps` (`atomic.go`) covers `.tskflwctl-*.tmp` orphans but not
a half-written destination created by `createFileAtomic`, so the lint/repair
question the criterion poses is still live — just scoped to crash recovery.

**Q4 — unchanged.** No second secondary adapter exists; the conformance-suite
trigger has not fired.

**Atomic-write divergence still three-way:** `internal/store/atomic.go`,
`internal/config/config.go`, and `internal/userconfig/paths.go` each carry their
own implementation, so the linked unification task's premise holds.

## Progress Log

- 2026-09-22: architecture audit M1/L1 added one design question: name the shared durable-outcome
  predicate and retry rule, then either include task rename or document a deliberate carveout. No
  implementation child was created before this design task's required human review.
- 2026-09-20: automated weekly sweep — all six foundation links and all four newly observed questions re-verified against current code; Q3 narrowed to process-death recovery only; noted the completed shared-create-guard related link.
