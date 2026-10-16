---
schema: 1
id: 6gm9ny98zyf2
status: next-up
epic: 21-code-quality-architecture-hardening
description: Prevent a successful explicit repair selection from switching physical task sources after alias or path drift during conflict retries.
effort: 1 day
tier: 2
priority: high
autonomy_level: 2
tags: [correctness, identity, repair, concurrency]
created: "2026-10-16"
---
# Pin explicit graph-repair source selections across retries

## Objective and evidence

An explicit repair selection must not become another physical task source after a
pre-commit conflict. Preserve fresh defect authorization and convergent repair, not
a cached write set. This is a bounded follow-up to the three-task safety PR, not a
universal mutation framework or an implicit extension of `--auto` semantics.

The [Claude safety review](../audits/6gm8vm7aq7zt-2026-10-16-selected-identity-safety-implementation-claude.md)
raised H-a as unproven. Owner validation confirmed it on 2026-10-16:

1. Seed distinct task IDs A/alpha and B/bravo, both with `depends_on: [invalid-token]`.
2. Request `RepairTaskGraph` dropping that declaration using `Source.TaskSlug=alpha`.
3. At `testHookBeforeGraphRepairVerify`, rename A to A-renamed and B to B-alpha,
   keeping the IDs and contents unchanged. The first attempt detects the source CAS
   conflict before any write.
4. On retry, `normalizeRepairEditSource` reselects B by the old slug. The call returns
   nil and `AppliedFiles` names B-alpha; A retains its defect. Strict improvement and
   removal-only policy do not establish that B was the requested source.

The compiler-valid throwaway real-store probe passed its bug-reproduction assertion
under `go test -race -run '^TestOwnerProbeGraphRepairRebindsExactSlug$' -count=1 -v`.
Fixture IDs were `ekpd8vnydg2h` (A) and `h46gx48dzz67` (B). The probe was removed rather
than retained as passing safety coverage; replace it with an opposite, red-first
regression when implementing. No real planning artifacts were used as fixtures.

## Scope and constraints

- Inventory CLI drop/clear selectors, portable `TaskGraphSourceEdit` intent and
  `RepairTaskGraph` retry behavior. Resolve explicit selectors first inside the guard;
  after selection retain enough source evidence to refuse alias/path/ID replacement.
- Broken graphs may contain duplicate IDs and readable physical occurrences. Do not
  collapse source identity to task ID alone, use opaque Location as authority or lose
  the existing exact local-path disambiguation. Ask before choosing any new identity
  or absence policy that changes accepted repair contracts.
- Recompute current defects, requested declaration occurrence, strict improvement,
  residual problems and task/Thread impacts. Preserve unrelated concurrent changes,
  malformed-Thread tolerance and committed-prefix retry stops.
- Treat auto-cleanup separately: it deliberately rediagnoses the current graph and
  does not mean a user's prior explicit source selection may be replaced.

## Acceptance criteria

- [ ] A real-store red-first regression reproduces the confirmed slug-swap failure;
  fixed behavior never repairs B when A was first selected.
- [ ] Explicit ID/slug/path selectors retain physical source selection across rename,
  deletion, alias/path reuse and duplicate-ID occurrences, with documented refusal
  behavior when exact source continuity cannot be proved.
- [ ] Portable pre-planner/empty-result conflicts and real adapter tests preserve
  fresh authorization, unrelated changes, preview/no-op behavior and durable receipts.
- [ ] Auto-cleanup semantics remain deliberate; source identity policy is documented
  beside owning core/adapter code and the focused planning-data/persistence guides.
- [ ] Normal build, race tests, lint and independent hostile review pass.

## Sequencing

Add to Finish selected identity safety as the next bounded repair follow-up. The
current three fixes need not wait to ship; they are precedent, not an artificial
dependency. Keep broad entity-integrity/storage redesign in its existing design task.
