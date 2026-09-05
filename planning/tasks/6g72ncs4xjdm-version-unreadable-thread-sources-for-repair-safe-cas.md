---
schema: 1
id: 6g72ncs4xjdm
status: completed
epic: 30-threads-and-task-dependency-graphs
description: Make Thread snapshot comparison detect concurrent content changes in Thread files that remain unreadable.
effort: 1 day
tier: 2
priority: high
autonomy_level: 4
tags: [threads, storage, concurrency]
created: "2026-09-05"
depends_on: [6g721tvf4crh]
updated_at: "2026-09-05"
started_at: "2026-09-05"
audit_sources: [planning/audits/6g73ymfrj67y-2026-09-05-thread-source-revision-cas-implementation-claude.md, planning/audits/6g73ymg15x2e-2026-09-05-thread-source-revision-cas-implementation-antigravity.md]
completed_at: "2026-09-05"
---

# Version unreadable Thread sources for repair-safe CAS

## Objective

Give unreadable Thread documents the same opaque source-revision evidence as unreadable task
documents. Graph repair and any later mutation that deliberately tolerates malformed Threads must
not authorize a write from stale `{location, message}` diagnostics when the underlying bytes changed.

## Scope

- Extend the adapter-neutral Thread read-problem contract with an opaque, non-wire source revision.
- Have the filesystem Thread reader hash the exact bytes it already parsed, without a second scan
  or a token leak through ordinary list, lint, CLI, TUI, schema, or wire projections.
- Replace `sameThreadSourceSnapshot`'s `FileProblem` equality with deterministic comparison of
  stable identity, optional location, diagnostic, and a non-empty equal source revision.
- Reuse the shared graph-aware pre-write CAS convention and preserve current rejection of malformed
  Threads in ordinary creation, membership, lifecycle, and bulk-apply planners.

## Acceptance criteria

- [x] Byte-identical unreadable Thread sources compare as the same snapshot, while changed bytes
      that reproduce the same diagnostic compare as different snapshots.
- [x] Missing revisions fail closed, pathless adapters can supply opaque revisions, and adapter
      ordering does not affect comparison.
- [x] Readable/unreadable transitions plus add, remove, rename, and identity drift invalidate the
      Thread snapshot.
- [x] Ordinary Thread diagnostics and all public or machine-readable projections remain compatible
      and cannot expose the revision token.
- [x] Focused store tests prove the shared Thread snapshot verifier returns
  ErrConflict when changed unreadable bytes reproduce the same diagnostic.
- [ ] The first guarded writer allowed to proceed with malformed Thread evidence
  exercises the unreadable-revision CAS before any write. · **tracked:** by 6g4g8gatbnrs — the dedicated repair path is the first writer allowed to plan from malformed Thread evidence and owns the end-to-end proof

## Out of scope

- Implementing dependency repair, allowing ordinary Thread mutations over malformed documents, or
  introducing a universal persisted revision field for planning entities.

## Sequencing

Follow the task-source revision slice so both use the same opaque-token and fail-closed conventions.
Land before guarded graph repair because its readable-Thread impact receipt must either be based on
stable Thread evidence or report that evidence as incomplete.

## Related

- Epic [30-threads-and-task-dependency-graphs](../epics/30-threads-and-task-dependency-graphs.md)
- Predecessor [unreadable task-source revisions](6g721tvf4crh-version-unreadable-task-graph-sources-for-repair-safe-cas.md)
- Follow-on [guarded graph repair](6g4g8gatbnrs-add-a-guarded-repair-path-for-broken-dependency-graphs.md)
- Review evidence [Claude](../audits/6g725zn6r92x-2026-09-05-version-unreadable-task-graph-sources-for-repair-safe-cas-implementation-claude.md) (M2)
- Review evidence [Antigravity](../audits/6g726063yv11-2026-09-05-version-unreadable-task-graph-sources-for-repair-safe-cas-implementation.md) (M2)

## Implementation progress (2026-09-05)

`ThreadReadProblem` now carries an opaque, non-wire `SourceVersion`, and the filesystem derives it from the exact unreadable bytes during the same resilient scan. `ReadThreads` is the filesystem's sole general Thread-listing path, while bulk apply's specialized body-aware read carries the same revisions; guarded writers therefore cannot select a parallel tokenless read. Service, TUI, CLI, schema, and wire projections strip or explicitly omit the token.

Thread creation, membership/lifecycle mutation, bulk apply, and task lifecycle now authorize against one deterministic `ThreadRead` source snapshot through the shared CAS helper. Comparison is order-independent, includes stable identity, optional location, diagnostics, and readable/unreadable revisions, and fails closed when revision evidence is absent. Tests cover same-diagnostic byte changes, exact restoration, pathless adapters, ordering, missing tokens, add/remove/rename/identity and representation transitions, resilient list diagnostics, and token non-disclosure.

Validation: focused core/store/wire tests and focused race tests pass; the full `just test` race suite, `just lint`, `just tidy-check`, `just docs-check`, planning lint, audit lint, and `git diff --check` are clean.

## Adversarial review closeout (2026-09-05)

Claude and Antigravity found no production correctness defect in revision generation, comparison, CAS placement, or token non-disclosure. Their shared mutation probe did expose an architectural regression route: the unused tokenless `ListThreads` API could replace `ReadThreads` without a current test failing. That method, its conversion helpers, and its scan-count-only test seam are now removed; all general and body-aware Thread reads used for mutation carry revisions.

The original writer-level criterion overstated what ordinary mutations can exercise because they deliberately reject malformed Thread evidence before planning. The focused verifier proof remains complete, while the first end-to-end unreadable-to-unreadable writer proof is tracked on `6g4g8gatbnrs`; its concurrency criterion now names still-unreadable Thread bytes explicitly. Eager pre-parse hashing remains intentional: the prerequisite review measured the shared scanner overhead at roughly five percent for 2000 8 KB files, which does not justify weakening the exact-byte invariant. Both audits are closed with every finding fixed, tracked, or dispositioned.

Post-review validation: the full race suite, lint, tidy check, generated-doc check, planning lint, audit lint, and `git diff --check` pass.
