---
schema: 1
id: 6gkt3fphx1ts
status: ready-to-start
epic: 21-code-quality-architecture-hardening
description: Stop audit, research and epic retry loops from reselecting another record after alias drift or deletion.
effort: 1-2 days
tier: 2
priority: high
autonomy_level: 2
tags: [correctness, identity, core, store, concurrency]
created: "2026-10-14"
audit_sources: [2026-10-14-task-correctness-detour-implementation-claude]
updated_at: "2026-10-14"
---
# Pin non-task mutation retries to selected source identities

## Objective and evidence

Extend the strict selected-identity rule to ordinary audit, research and epic
mutations without introducing a universal entity mutation framework.

The owner independently reproduced audit append and research field updates mutating
a pre-existing `alpha-two` after `alpha` was renamed to `omega` in their pre-lock
CAS seams. Both commands returned nil. The source retained its bytes and the sibling
changed. This is deterministic real-store evidence, not measured concurrent frequency.
Probe: `/private/tmp/taskflow-target-neighbors.VNxBO5/neighbor_retry_probe_test.go`
(temporary evidence, not a durable test dependency).

## Scope

- Audit close/defer/reopen, append, finding creation/edit and header canonicalization;
  research field updates/append; epic field/lifecycle updates.
- Resolve initial references once, then use strict adapter-established source identities
  across attempts. A disappeared identity must never fall back to another human label.
- Audit header canonicalization already selected a snapshot occurrence: use its
  `Source.ID`, not the domain slug. Initial duplicate-label selection must remain loud.
- Follow the task mutation target pattern, keeping entity-specific ports and policy.
  Preserve original-reference ergonomics, preview/no-op behavior, authorization,
  timestamp capture, retry bounds and source-preserving CAS.
- Do not silently include interactive editors, creates, graph planners or new wire
  shapes. Record any separately confirmed exposure at those boundaries.

## Acceptance criteria

- [ ] Each ordinary retry family retains its first source identity through rename, alias reuse, deletion and ID-shaped label fallback; unrelated records stay byte-identical.
- [ ] Audit bulk header repair targets snapshot identities even with duplicate display slugs, and retains partial progress and baseline-aware repair policy.
- [ ] Empty identity and duplicate canonical IDs refuse before callbacks, preview/no-op returns or effects, using useful existing sentinels.
- [ ] Portable and real-adapter regressions cover transient initial selection, one retry budget, normal concurrent edits, and the observed audit/research wrong-target cases.
- [ ] Owning comments/guides and full validation accurately separate fixed families from remaining selection boundaries.

## Sequencing

Use [strict task targeting](6gkmexz4vh23-bind-task-mutation-retries-to-strict-source-identities.md)
as the proven pattern, not an artificial code dependency: these entity-specific fixes
can proceed independently. Feed concrete port/retry decisions into the shared-integrity
design rather than waiting for it or reopening completed legacy tasks.
