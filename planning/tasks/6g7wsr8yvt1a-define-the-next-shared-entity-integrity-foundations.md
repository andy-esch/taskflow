---
schema: 1
id: 6g7wsr8yvt1a
status: ready-to-start
epic: 21-code-quality-architecture-hardening
description: Consolidate identity, schema, locking, atomic-write, and diagnostic foundations; assess mutation-outcome and guard-lifecycle gaps before spinning out work.
effort: 1-2 days
tier: 2
priority: medium
autonomy_level: 2
tags: [architecture, integrity, store, design]
created: "2026-09-07"
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
      lifecycle, atomic create/replace guarantees, and adapter-neutral diagnostic evidence.
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
