---
schema: 1
id: 6gcwd78p9r04
status: completed
description: Eliminate path-shaped application contracts and primary-to-secondary planning-data bypasses.
goal: Core use cases consume identity-bearing semantic ports; local paths are optional capabilities, and primary adapters cannot bypass the application boundary.
created: "2026-09-23"
tags: [architecture, ports, adapters, hardening]
tasks: [6g5vm4efjcdv, 6g6jqqcdehne, 6gcqz5aefjjf, 6gcwcf77tvgq, 6gcwcf7gjayh, 6gcwcf7rgxef, 6gcwcf80v8hg, 6gcwcf88z57p, 6gcwcf8gzn50, 6gcwcf8rxe72, 6gdx7mcqm371, 6gdx7mcqq67d, 6gdx7mcrq8s8, 6ge1bacd3bd2, 6ge7qn9ptaxv, 6gg7e594gcms, 6gg7e59cyxxh, 6ggdkztshmzz, 6ggdkzv2tnta]
updated_at: "2026-10-05"
started_at: "2026-09-23"
ended_at: "2026-10-05"
---

# Thread: Make planning data access adapter neutral

**Goal.** Core use cases consume identity-bearing semantic ports; local paths are optional capabilities, and primary adapters cannot bypass the application boundary.

## Context

Taskflow already enforces inward package dependencies and has proven adapter-neutral patterns for
Thread reads and task-graph reads. The Thread's first evidence stage is now complete: repository
lint diagnostics, graph-lint attribution, Board/status diagnostics, and audit finding queries all
preserve stable identity without requiring a filesystem path. Those slices also establish that an
opaque source location and a local repair path may coexist, core never reconstructs identity from
either value, guarded source revisions remain private, and public projections use deterministic
ordering without rescanning records.

The loaded-record and optional-local-capability design is complete. Ordinary task, epic, audit,
research, and finding reads now preserve portable source identity and diagnostics, and split
planning-data capabilities must agree on one source-set witness at construction. Local mutation
receipts and ordinary-create recovery no longer depend on domain paths. TUI identity/refresh and
readable occurrence locations have merged, and semantic entity values no longer carry local paths,
filename identities, or guarded revisions. CLI repair, body-link checking, and completion have
merged in PR #275. Concrete composition isolation and executable controller enforcement merged
in PR #277; workspace identity parity merged in PR #279. Cross-kind collision lint merged in
PR #280 and core-owned impact/recovery semantics merged in PR #282. Explicit persistence
authorization merged in PR #283. The final guarded-contract regression pass merged in
[PR #284](https://github.com/andy-esch/taskflow/pull/284); this Thread is completed.

This Thread closes those seams in evidence-driven stages:

1. Establish and stress-test the diagnostic/read precedents through lint, graph attribution,
   Board/status, and audit findings. **Completed.**
2. Design the shared loaded-record and optional-local-capability contract. **Completed.**
3. Promote portable diagnostics and migrate ordinary list/show, Summary, lint, and wire
   projections onto loaded-record evidence. **Completed.**
4. Bind split capabilities to one source set, preserve local create/rename outcome evidence outside
   domain records, and report ordinary-create guard-release failures without suggesting an unsafe
   retry. **Completed.**
5. Preserve readable occurrence locations and move TUI identity/refresh onto portable loaded-record
   evidence while retaining fail-closed duplicate-ID behavior. **Completed.**
6. After those prerequisites, move local paths, filename-derived identity, and guarded revisions
   out of semantic entity values; finish the partial Thread precedent and wire explicit local
   capabilities. **Completed.**
7. Route CLI planning-data operations through those settled application ports. **Completed:**
   named repair/link/completion use cases, portable probes, and resolver-compatible completion fixes
   merged in PR #275; both reviews reconciled.
8. Isolate composition wiring and make the intended controller boundary executable through the
   standard lint suite. **Completed, merged in PR #277:** explicit binary wiring and launch contracts;
   default-deny/recursive lint rules and actual forbidden-import probes verified. Codex and Antigravity
   findings are fixed and both audits closed; real-reader/corpus mutation tests now pin startup
   laziness and no fallback. Release inclusion is a separate milestone.
9. Finish four independent implementation slices over the settled boundary:

   - [Workspace Thread-apply identity parity](../tasks/6gg7e594gcms-preserve-planning-identity-revalidation-in-workspace-opened-thread-apply.md),
     including direct/pointer roots and identity replacement. **Completed, merged in PR #279:**
     both reviews reconciled. The Codex test-gap finding now has hostile initial-repoint
     coverage plus an executable direct-discovery guard. The pre-existing
     markerless recovery hint is tracked separately, not added as a refactor closure blocker.
   - [Explicit persistence authorization](../tasks/6gg7e59cyxxh-require-an-explicit-mutation-authorization-policy-at-persistence-composition.md),
     **Completed, merged in PR #283:** required policy arguments and deliberate unrestricted
     fixture mode were user-approved before implementation. All adapters fail closed on omission;
     Atlas readers are read-only and late opening preserves guards. Both independent reviews are
     reconciled and closed; their useful populated-record/full-bundle probes are now permanent
     regressions. Validation is green; no new data-safety blocker was demonstrated.
   - [Cross-kind collision lint on unreadable records](../tasks/6gcqz5aefjjf-lint-cross-kind-task-and-thread-id-collisions-on-unreadable-records.md),
     **Completed, merged in PR #280:** portable recovered identity now participates in lint,
     including both unreadable owners. Membership validity and creation policy are unchanged;
     core/CLI regressions and four restored owner mutation probes pass.
   - [Core-owned dependency impact and recovery semantics](../tasks/6ggdkztshmzz-make-dependency-impact-and-recovery-semantics-core-owned.md),
     **Completed, merged in PR #282:** adapter-hygiene M1/L1 are fixed without changing graph
     eligibility. Both reviews are reconciled; the reproduced Thread-only TUI feedback gap is
     tracked in the TUI refinement Thread rather than added to this closure path.

   These slices can progress independently; graph edges impose no artificial serial order.
10. [Pin guarded planning mutation boundary contracts](../tasks/6ggdkzv2tnta-pin-guarded-planning-mutation-boundary-contracts.md)
    after all four slices. **Completed, merged in PR #284:** malformed/stale-plan, committed-conflict
    retry, and explicit-space theme regressions now pin the functioning guards. The bounded final
    review covered identity, authorization, recovery, and adapter ownership; no production guard
    needed changing. Task-local tests were not postponed to this pass.

Markdown-first and git-native behavior remain product constraints. The goal is not to hide Markdown
or replace storage; it is to stop application semantics from requiring a filesystem path when stable
identity, an optional source location, or an explicit local capability is the honest contract.

## Completion signal

- Core semantic read and diagnostic ports do not expose `domain.FileProblem` or require local paths.
- Local path navigation remains available through explicit optional capabilities.
- Independently composed readers, path resolvers, and mutators cannot silently address different
  planning corpora, and local dry-run/recovery outcomes no longer depend on domain paths.
- Primary adapters cannot open planning persistence directly outside the documented composition
  boundary, and a dependency rule catches regressions.
- Filesystem behavior, guarded mutation evidence, TUI editing, and public machine compatibility stay
  intact throughout the migration.
- Workspace and ordinary opening both revalidate planning identity; omission of mutation policy is
  handled explicitly rather than silently authorizing an unintended caller.
- Lint retains safely recovered cross-kind collision identity, and dependency impact/recovery policy
  is core-owned and preserved for human and machine consumers.
- Closure requires the bounded regression/closeout task to complete with no unresolved boundary
  data-safety blocker.
  Review findings handed to followups are marked **tracked**, never represented as implemented fixes.

## Scope boundary and bookkeeping (2026-10-05)

This Thread has a finish line, not a mandate to absorb every architecture audit. The two new tasks
above and the existing collision-lint task are members of this graph. Composition and workspace
parity, collision lint, and core recovery semantics are completed in PRs #277, #279, #280,
and #282; explicit persistence authorization merged in PR #283.
**19 of 19 members are completed and merged**; PR #284 landed and guarded Thread completion
succeeded. Older loaded-record TUI reviews are
closed because their recorded findings were fixed (or absent), not because a new review was
performed today.

The proposed dependency-policy ADR remains in the
[documentation Thread](6g9czp7g9pt3-make-documentation-layered-executable-and-agent-navigable.md),
and still needs user acceptance. Broader atomic-write/locking/schema work and a general conformance
framework remain in their existing homes. TUI broken-gate versus unreadable-marker semantics stay
outside this Thread; the related audit findings remain open. The
[`cli.App` breadth decision](../tasks/6g63hjme8czk-decide-whether-cli.apps-breadth-has-a-real-trigger-yet.md)
has refreshed assumptions, not a pre-approved refactor.

## Closeout and retrospective (2026-10-05)

PR #284 is merged in `f50deee`. The guarded `thread complete` preview and commit succeeded:
19/19 members are completed and soundly drained, all 107 member ACs are checked, all three external
gates are soundly complete, and graph/projection health is healthy with no in-flight or frontier
work. The [final task's boundary matrix](../tasks/6ggdkzv2tnta-pin-guarded-planning-mutation-boundary-contracts.md#pre-closeout-boundary-matrix)
records the behavior-level evidence for every completion signal above.

The original goal is met: semantic records no longer require filesystem evidence; diagnostic/source
identity is explicit; local navigation and guarded versions are separate capabilities; composition
checks paired corpora and explicit authorization; and controllers use application-owned policy and
recovery rather than constructing persistence. The CLI and TUI are real consumers of that boundary.
Future adapters have an honest contract, not a claim that a remote/database backend has been built.
Named topology, checkout, watcher, and process exceptions remain deliberate local integration.

Lessons worth retaining:

- **Optional fields were not enough.** Removing source metadata from domain values, after supplying
  loaded records and operation-specific receipts, prevented consumers from quietly rebuilding a
  filesystem dependency. Keep meaning, source context, and authority separate.
- **Ports alone do not make composition safe.** Source-set witnesses, late identity revalidation,
  explicit policies, and no-fallback opening were necessary complements. Witness equality does not
  replace snapshot consistency, CAS, or adapter tests.
- **Recovery is part of the application contract.** An error can accompany durable success;
  preserving the original receipt and suppressing blind retries matters more than a generic error
  label. Proposed impacts are not proof that writes committed.
- **Negative tests need reachability and evidence.** Real populated adapters, whole-tree no-effects
  checks, and compiler-valid guard-removal probes caught weaknesses happy-path coverage missed.
  The final Codex review is accepted; Antigravity is only qualified corroboration because its
  inaccurate inventory and missing attestation are recorded, not papered over.

Reviewed the 32 audits linked by member task bodies: 30 are closed; the mixed adapter-hygiene and
test-rigour source audits remain open for five unrelated TUI/test findings. The final contract
findings now say merged in PR #284, and stale workspace review dispositions now link PR #279.
Tracked historical handoffs are not rewritten as proof of a new independent review.

Fresh merged-main race tests, standard lint, and build passed. The closeout documentation adds
the short [architecture checklist](../../docs/ARCHITECTURE.md#planning-data-change-checklist), routes README/agent/package
guidance to it, and repairs three historical body links. No runtime or wire behavior changes.
This is bounded refactor completion, not universal conformance, ADR acceptance, or release readiness.

## Remaining work has separate homes

- **Before the next release:** [untouched YAML scalar fidelity](../tasks/6g1dhhk6721x-a-surgical-frontmatter-write-re-folds-multi-line-block-scalars-onto-one-line.md)
  remains a reproduced, high-priority storage defect. Closure does not fix or de-prioritize it.
- **Documentation Thread:** [source ownership/agent routing](../tasks/6g9cz5saenme-establish-documentation-ownership-and-agent-routing.md),
  [architecture restructuring](../tasks/6g6x7e2ef37r-restructure-the-architecture-documentation-into-focused-guides.md),
  [package contracts](../tasks/6g9cz5sk3b0d-adopt-selective-contract-focused-go-package-documentation.md),
  [import-map verification](../tasks/6g63hjm7cp6w-test-the-documented-import-graph-instead-of-date-stamping-a-manual-review.md),
  and the [dependency-policy ADR](../tasks/6gg7e59mm68g-record-the-hexagonal-dependency-policy-and-composition-exceptions-in-an-adr.md)
  remain open. The checklist is integrated into the existing architecture reference as an
  immediate router, not a new document layout or completion of those larger tasks.
- **New enforcement followup:** [direct-I/O/framework import guards](../tasks/6ggxymjbf86r-guard-core-and-domain-against-direct-i-o-and-framework-imports.md)
  follows the ADR and precedes documentation reconciliation. Existing lint blocks internal edges,
  not all future effect/framework imports; current core/domain imports comply. This is future
  regression prevention, not an unresolved runtime bypass or a reason to reopen the migration.
- **CLI/TUI Threads:** [ID-less recovery instructions](../tasks/6ggfd81jg0qg-make-id-less-thread-planning-recovery-instructions-executable.md)
  and [Thread-only TUI recovery](../tasks/6ggkdbg0816h-preserve-thread-only-lifecycle-recovery-guidance-in-the-tui.md)
  retain their existing destinations. Other TUI/test findings stay visible in their source audits.

No remaining task is required to satisfy this Thread's bounded completion signal. Do not add
unrelated storage, rendering, or documentation rewrites to a completed graph merely because they
were noticed during the refactor.
