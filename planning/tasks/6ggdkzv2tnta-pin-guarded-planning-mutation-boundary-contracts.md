---
schema: 1
id: 6ggdkzv2tnta
status: ready-to-start
epic: 21-code-quality-architecture-hardening
description: Pin malformed-plan refusal, committed Thread creation recovery, and explicit-space selection before boundary closeout.
effort: 4-8 hours
tier: 2
priority: medium
autonomy_level: 3
tags: [architecture, testing, ports, mutations]
created: "2026-10-04"
updated_at: "2026-10-04"
audit_sources: [2026-10-02-test-rigour]
depends_on: [6gcqz5aefjjf, 6gg7e594gcms, 6gg7e59cyxxh, 6ggdkztshmzz]
---
# Pin guarded planning mutation boundary contracts

## Objective

Finish the adapter-neutral Thread with behavior-level regressions that prove settled planning
identity, mutation authorization, refusal, and recovery boundaries are load-bearing. Close the
specific uncovered guards without turning this into a general conformance framework.

## Evidence

[2026-10-02 test rigour](../audits/6gfrcytd9n9a-2026-10-02-test-rigour.md), M1-M3:

- M1: malformed manifest and stale Thread-apply refusals lack focused negative tests; some
  existing rows assert message substrings without the error class. Re-measure today's reachable
  branches rather than treating the audit's historical 24/37 coverage count as a current target.
- M2: Thread creation's post-commit no-retry test injects a generic error. It cannot detect
  removal of the committed guard because the error does not match `ErrConflict`.
- M3: an explicitly invalid `--space` must not silently fall back to ambient theme output;
  template has the analogous regression, but theme does not.

These are presently test gaps around functioning guards, not a claim that the shipped binary
already accepts malformed plans or retries committed Thread creation.

## Scope

- Extend focused compose/apply refusal tests over invalid manifests, edited plans, missing
  prerequisites/dependents, unavailable authoritative bodies, and other reachable boundary
  conditions. Assert the sentinel, useful diagnostic, and absence of persistence where applicable.
- Pin Thread creation recovery for both generic and conflict-wrapping post-commit errors through
  a portable fake and the real adapter/service composition. Preserve the committed identity and
  receipt, and prove no retry or remint occurs.
- Add the small explicit-space theme regression; inspect sibling guards without redesigning the
  command resolver or silently expanding into unrelated CLI cleanup.
- After the four implementation prerequisites land, run a focused closeout review over planning
  identity, authorization, core-owned impact/recovery semantics, and adapter ownership. Existing
  task-local tests remain the first home for behavior; this pass fills demonstrated gaps.
- Use exact guard-removal probes where useful. A failing probe must reach the target guard, not
  fail because its fixture lacks unrelated required metadata or violates an earlier condition.

## Acceptance criteria

- [ ] Reachable malformed manifest, edited plan, and stale corpus refusals are
  covered with sentinel and useful diagnostic assertions; failed or dry-run
  applies do not persist changes.
- [ ] Both generic and conflict-wrapping post-commit Thread creation failures
  retain committed identity and receipt with no retry/remint, proven through
  portable core tests and real adapter/service composition.
- [ ] An explicitly invalid theme --space selection returns the documented
  failure rather than ambient success; an exact guard-removal probe fails for
  this behavior.
- [ ] Targeted compose/apply and committed-conflict guard-removal probes fail
  for their intended assertion, not unrelated fixture validation; restored
  focused tests pass.
- [ ] The final boundary review covers the four completed prerequisites without
  duplicating the workspace identity matrix; remaining findings are fixed or
  explicitly tracked with destinations and no unresolved data-safety blocker.
- [ ] Full tests, race tests, standard lint, generated-doc checks, and
  planning/audit lint pass; the Thread graph and closeout evidence distinguish
  implemented, tracked, and released work.

## Out of scope

- Duplicating the root/identity-replacement matrix owned by the workspace-parity task.
- Redesigning retry/durability predicates, sharing all manifest/plan validators, or building a
  general adapter-conformance framework or second backend.
- The audit's L1-L3, unrelated atomic-write/locking/schema work, and TUI visual semantics.

## Related

- [Adapter-neutral planning Thread](../threads/6gcwd78p9r04-make-planning-data-access-adapter-neutral.md)
- [Workspace identity parity](6gg7e594gcms-preserve-planning-identity-revalidation-in-workspace-opened-thread-apply.md)
- [Explicit persistence authorization](6gg7e59cyxxh-require-an-explicit-mutation-authorization-policy-at-persistence-composition.md)
- [Cross-kind collision lint](6gcqz5aefjjf-lint-cross-kind-task-and-thread-id-collisions-on-unreadable-records.md)
- [Core-owned impact/recovery semantics](6ggdkztshmzz-make-dependency-impact-and-recovery-semantics-core-owned.md)
