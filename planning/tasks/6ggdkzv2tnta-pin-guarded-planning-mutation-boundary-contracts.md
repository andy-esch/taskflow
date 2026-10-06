---
schema: 1
id: 6ggdkzv2tnta
status: completed
epic: 21-code-quality-architecture-hardening
description: Pin malformed-plan refusal, committed Thread creation recovery, and explicit-space selection before boundary closeout.
effort: 4-8 hours
tier: 2
priority: medium
autonomy_level: 3
tags: [architecture, testing, ports, mutations]
created: "2026-10-04"
updated_at: "2026-10-05"
audit_sources: [2026-10-02-test-rigour]
depends_on: [6gcqz5aefjjf, 6gg7e594gcms, 6gg7e59cyxxh, 6ggdkztshmzz]
started_at: "2026-10-05"
completed_at: "2026-10-05"
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

- [x] Reachable malformed manifest, edited plan, and stale corpus refusals are
  covered with sentinel and useful diagnostic assertions; failed or dry-run
  applies do not persist changes.
- [x] Both generic and conflict-wrapping post-commit Thread creation failures
  retain committed identity and receipt with no retry/remint, proven through
  portable core tests and real adapter/service composition.
- [x] An explicitly invalid theme --space selection returns the documented
  failure rather than ambient success; an exact guard-removal probe fails for
  this behavior.
- [x] Targeted compose/apply and committed-conflict guard-removal probes fail
  for their intended assertion, not unrelated fixture validation; restored
  focused tests pass.
- [x] The final boundary review covers the four completed prerequisites without
  duplicating the workspace identity matrix; remaining findings are fixed or
  explicitly tracked with destinations and no unresolved data-safety blocker.
- [x] Full tests, race tests, standard lint, generated-doc checks, and
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
- Final independent review briefs: [Codex](../audits/6ggxdjh7vpcr-2026-10-05-guarded-planning-boundary-closeout-codex.md)
  and [Antigravity](../audits/6ggxdjhhafdv-2026-10-05-guarded-planning-boundary-closeout-antigravity.md).
- Implementation and review bookkeeping: [PR #284](https://github.com/andy-esch/taskflow/pull/284).

## Implementation and owner closeout check (2026-10-05)

No production guard needed changing. The demonstrated gaps were in the assertions and
reachability of the regressions, not currently permissive compose/apply behavior.

- Compose tests now cover empty/malformed nodes, keys, task references, membership, edges,
  input clock/generator, and exhausted collisions, asserting the error class and diagnostic.
- Prepare tests establish a valid baseline before editing schema, repository identity,
  creation metadata, edge shape/endpoints, or authoritative body availability. Refusals publish
  no decision and preserve the caller's plan.
- `TestThreadApplyMalformedAndStalePlansDoNotPersist` composes through the real service/store,
  proves a valid preview has no effects, then exercises 11 edited-plan/stale-corpus cases in
  both preview and commit mode. It checks the typed failure and retained full plan/receipt,
  plus entries, bytes, modes, and symlink targets via the shared tree-snapshot oracle.
- Creation recovery runs generic and conflict-wrapping cleanup failures through a portable
  fake and the real service/store. It checks the original committed identity, local outcome,
  typed receipt, one mint, one mutation/release, no retry, and exactly one readable Thread.
- Explicit-space tests pin refusal for template list, theme list, and noninteractive preview.

### Executed guard-removal probes

Independent disposable clone `/private/tmp/isolated-review.L9MpKt`, baseline
`d94dd176b49e2d0d53e6e2fdbc8fc81e829a4be4`, captured the dirty implementation without
changing the owner checkout. Six compiler-valid probe groups removed seven guards:

| Removed behavior | Intended regression result |
| --- | --- |
| Compose duplicate local-key refusal | Specific compose row received success instead of `ErrValidation`. |
| Prepare missing prerequisite/dependent refusals | Core and real-store rows failed on lost endpoint-specific diagnostics; later validation still refused writes. |
| Existing Thread authoritative-body check | Core row received `ErrConflict` for an empty/different body instead of missing-evidence `ErrValidation`. |
| Planned Thread lifecycle check | Core and real-store rows lost the early plan-specific diagnostic; creation validation still refused. |
| Creation's `!result.Committed` retry guard | Conflict rows failed: fake made five calls; real store retried and lost the committed receipt. Generic controls passed. |
| Theme's explicit-space refusal | Both theme rows returned ambient output and success for an unknown space. |

All probes were restored with an empty sandbox diff; the focused suite and standard lint
passed afterward. A separate nested CLI import probe was rejected by depguard, proving the
recursive controller boundary is executable rather than only documented. No compile failures
count as mutation evidence. Redundant guards are classified as diagnostic protection, not
misrepresented as the only protection against unsafe persistence.

### Pre-closeout boundary matrix

| Completion claim | Current evidence |
| --- | --- |
| Semantic records and optional local navigation remain separate | `TestSemanticEntitiesDoNotCarrySourceEvidence`, pathless read/lint fixtures, `TestEntityPathCapabilitiesIgnoreTypedNilAndRejectForeignSourceSets`; source/version evidence lives beside domain values. |
| Split capabilities share one corpus | `TestNewServiceRejectsEveryMismatchedSplitCapability` and missing/empty/unstable witness tests; constructor refuses before use. |
| Ordinary/workspace opening retains initial identity and fresh authorization | PR #279 tests in `appwiring/thread_apply_test.go`, including pointer/direct repointing; PR #283 policy propagation and late-opening tests. Existing matrices were rerun, not copied here. |
| Policy omission, previews, callbacks, and no-ops fail closed | Policy tests plus all-entry/populated-store denial tests and real application composition; read-only Atlas wiring stays explicit. |
| Unreadable cross-kind owners remain lint-visible | PR #280 portable core and human/JSON CLI regressions. |
| Impact/recovery intent is core-owned and survives consumers | PR #282 service impact/recovery tests, wire and CLI assertions; the known Thread-only TUI presentation gap has a separate destination. |
| Primary controllers cannot select persistence | Current ports/composition inventory, normal depguard lint, and the restored nested forbidden-import probe. Named init/workspace topology exceptions remain narrow. |

Fresh boundary-focused tests passed three consecutive runs; the new focused regressions
also passed five consecutive race-enabled runs. Full normal and race suites,
`just lint`, `just build`, generated CLI/schema-comment comparisons in disposable outputs,
planning lint, and audit lint passed. Machine schema remains 1.81; no wire behavior changed.
All 18 completed member tasks have checked ACs, all three external gates are soundly complete,
and the graph/projection are healthy with this task as the only in-flight member.

Owner verdict at the review handoff: ready for independent final review, with no demonstrated unresolved safety
blocker in this migration. The four prerequisites are merged (#279, #280, #282, #283); this
test slice is still local, not merged or released. Keep the task/Thread in progress until
review is reconciled. The source test-rigour audit's unrelated L1-L3 remain open.

The closeout sweep independently reproduced the separate folded-YAML value corruption
(41 to 46 decoded bytes after five unrelated writes). Its existing task
[6g1dhhk6721x](6g1dhhk6721x-a-surgical-frontmatter-write-re-folds-multi-line-block-scalars-onto-one-line.md)
now has high priority, a decoded-value AC, and the storage audit H1 handoff. This is a
near-term storage safety followup, not a fixed defect or reason to restart the port migration.
Other destinations remain [ID-less recovery](6ggfd81jg0qg-make-id-less-thread-planning-recovery-instructions-executable.md),
[Thread-only TUI feedback](6ggkdbg0816h-preserve-thread-only-lifecycle-recovery-guidance-in-the-tui.md),
and [dependency-policy ADR acceptance](6gg7e59mm68g-record-the-hexagonal-dependency-policy-and-composition-exceptions-in-an-adr.md).

### Codex review reconciliation

Codex independently corroborated the bounded closeout with 13 compiler-valid behavior probes,
the recursive lint probe, real durable recovery/empty-body/convergence experiments, and the
18-member/101-checked-AC inventory. No new finding was asserted. Owner spot-checks verified
the retained logs, restoration/transfer evidence, and identical five-file implementation
snapshot; the focused race suite passed again. The Codex audit is reconciled and closed.

No implementation change or new task was needed. Reviewer-only probes remain independent
evidence, not represented as permanent tests. The implementation snapshot stayed untouched
while Antigravity reviewed it. The existing YAML safety followup remains high priority and unfixed.

### Final reconciliation and completion

Antigravity asserted no new production defect. Owner triage corrected its nonexistent service
symbol, fictional gate names, broad purity/command claims, and local-only links. Its independent
baseline, source hashes, restored five-file implementation snapshot, and delivered audit copy
were verified; raw probe logs and transfer-result attestation were not retained. The audit is
closed as qualified corroboration, not credited as an evidence-complete independent clean review.
Codex's accepted review and the owner executed checks remain the closeout basis.

No implementation change was required. This task is complete locally in PR #284, not merged
or released. The Thread now has 19/19 completed members and healthy graph/projection; keep the
Thread itself in progress until this PR merges, then use guarded Thread completion. Existing
followups remain in their stated homes and are not represented as fixes.
