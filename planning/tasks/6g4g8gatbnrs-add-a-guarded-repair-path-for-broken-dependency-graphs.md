---
schema: 1
id: 6g4g8gatbnrs
status: next-up
epic: 30-threads-and-task-dependency-graphs
description: Repair cycles, dangling edges, and other broken graph-owned state without requiring an unsafe generic mutation path.
effort: 3-5 days
tier: 2
priority: high
autonomy_level: 3
tags: [threads, graph, storage, cli]
created: "2026-08-28"
depends_on: [6g3q4rt7mgjn, 6g697mp8s4tx, 6g6scc9jgxae, 6g721vewvvrz, 6g72ncs4xjdm]
updated_at: "2026-09-05"
started_at: "2026-09-05"
audit_sources: [planning/audits/6g71vzq8wdnj-2026-09-05-add-a-guarded-repair-path-for-broken-dependency-graphs-antigravity.md, planning/audits/6g71yr50md16-2026-09-05-add-a-guarded-repair-path-for-broken-dependency-graphs-claude.md]
---

# Add a guarded repair path for broken dependency graphs

## Objective

Provide an explicit recovery capability for graph-owned frontmatter that is already broken, without making generic setters or ordinary dependency mutations capable of bypassing repository-global validation.

## Scope

- Introduce a dedicated repair mutation port. It alone may accept a broken source graph; ordinary
  add/remove/migrate, lifecycle, generic setters, and lint fixing retain their existing fail-closed
  contracts.
- Make repairs removal-only and source-declaration-targeted. Canonical and legacy declarations,
  including legacy `blocks` whose declaration owner differs from its projected dependent, must be
  named without replacing a task's complete desired dependency state or clearing unrelated fields.
- Permit a successful repair to leave the repository broken when the selected intent strictly
  improves the structural state. Every durable prefix must remain syntactically valid, introduce no
  declaration or edge absent from the source snapshot, remove only authorized declarations, and
  discharge at least one selected repair intent. Graph health and the number or shape of emitted
  diagnostics may remain unchanged or appear worse as SCCs split or previously masked defects
  become visible; those presentation changes are not new structural corruption.
- Prove progress and preservation separately. A deterministic source-level defect measure must be
  componentwise non-increasing with at least one strict reduction for the complete plan; an explicit
  declaration-containment/minimality proof must reject unrelated constraint deletion. Keep the
  mathematical measure internal rather than freezing it as public wire vocabulary.
- Define `task depend repair` as the sole CLI family. With no mutation selector it is diagnostic and
  prints actionable, copyable commands. `--auto` may only deduplicate canonical values, remove
  self-edges, and clear present-but-empty legacy keys. Invalid or dangling raw references require an
  explicit `--drop`; cycle edges and ambiguous legacy references always require explicit selection.
  A manifest is always available for a larger repair but never required by operation count.
- Preserve the repository lock, snapshot-local planning, whole-snapshot and per-file CAS, surgical
  frontmatter updates, dry-run, hard minimality rejection, and typed partial-commit failure. A
  manifest expresses convergent repair intent and is reauthorized against the current snapshot, so
  retry after an interrupted durable prefix does not depend on an obsolete whole-plan fingerprint.
- Diagnose exact source, field, raw value, projected edge, and defect class. Human previews provide
  commands suitable for patching the selected defects; JSON retains raw invalid values and stable
  selectors for agents and other adapters.
- Compute task-state and Thread-projection impacts for readable Threads. Malformed Thread documents
  do not block repair of the underlying task graph; receipts explicitly report incomplete Thread
  evidence alongside the impacts that could be computed.

## Acceptance criteria

- [ ] A broken source graph can enter only the dedicated repair planner; ordinary add/remove/migrate,
      lifecycle, generic task mutation, and `lint --fix` continue to fail closed.
- [ ] Validation runs over the full source-level projection and preserves duplicate-ID shadow
      records, unreadable records/revisions, duplicate declarations, raw invalid values, and legacy
      declaration ownership. It does not reconstruct a prospective repository from representative
      `TaskGraph.Task()` values.
- [ ] Progress and preservation are independent hard checks: the complete plan strictly improves
      the structural defect measure, every prefix is structurally non-worsening and discharges
      selected intent, and no unrelated declaration or valid constraint disappears.
- [ ] Cycle, self-edge, dangling-reference, invalid-ID, duplicate-edge, and each legacy-field fixture
      have an actionable preview and converge to the selected repaired state, whether or not
      unrelated residual problems leave the repository broken.
- [ ] Bare diagnosis and `--dry-run` explain each defect, distinguish auto-safe from explicit
      repairs, show exact copyable selectors, and predict residual problems without writing.
- [ ] `--auto` is limited to duplicate/self/empty-legacy cleanup. Invalid and dangling values remain
      verbatim until explicitly dropped; cycle and ambiguous-legacy choices are never guessed.
- [ ] Concurrent task, dependency, and Thread-evidence edits—including byte
  changes to still-unreadable task or Thread files—produce a typed conflict
  rather than a stale repair. Every injected durable prefix is diagnosable,
  convergent on retry, and never repeats an already-satisfied dedupe or drop
  intent.
- [ ] Human and JSON receipts derive `Changed` from actual materialized writes and report `Committed`,
      initial/final health, selected and removed declarations, addressed and residual defects, raw
      removed values, workspace, applied/remaining files, task-state impacts, readable Thread
      impacts, and incomplete Thread diagnostics.
- [ ] Normal lint and other graph-health surfaces point to defect-specific repair diagnosis once it
      exists, including “repair, then migrate” where legacy migration cannot yet run.

## Out of scope

- Automatic best-guess repair of ambiguous legacy references.
- A generic `--force` escape hatch for arbitrary graph writes.
- Rewriting lifecycle status or Thread membership as part of dependency repair.
- Repair through `lint --fix`, replacement-style desired dependency sets, heuristic cycle breaking,
  automatic slug-to-ID reinterpretation, rollback of an already durable prefix, or adoption of a
  graph library that does not address the source-declaration problem.

## Sequencing

The v0.19.0 preview shipped the diagnostic baseline. Adversarial design and implementation review
found three foundations that must land before the recovery subsystem: opaque revisions for
unreadable task sources, the same protection for unreadable Thread evidence, and a lossless
graph-declaration projection and simulator. This task resumes after those foundations and owns
repair policy, guarded materialization, receipts, CLI/wire contracts, and guidance. The larger
question of whether relational planning data should remain authoritative Markdown is tracked
separately and does not weaken the current repair contract.

## Related

- Epic [30-threads-and-task-dependency-graphs](../epics/30-threads-and-task-dependency-graphs.md)
- Prerequisite [unreadable-source revision tokens](6g721tvf4crh-version-unreadable-task-graph-sources-for-repair-safe-cas.md)
- Prerequisite [unreadable Thread-source revision tokens](6g72ncs4xjdm-version-unreadable-thread-sources-for-repair-safe-cas.md)
- Prerequisite [source-level graph declarations](6g721vewvvrz-model-graph-owned-source-declarations-for-sound-repair.md)
- Follow-up research [Markdown-first storage durability](6g721w07mv1d-reassess-markdown-first-storage-durability-for-relational-planning-data.md)
