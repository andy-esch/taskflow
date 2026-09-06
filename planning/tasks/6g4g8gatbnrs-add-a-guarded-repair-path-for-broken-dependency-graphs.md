---
schema: 1
id: 6g4g8gatbnrs
status: completed
epic: 30-threads-and-task-dependency-graphs
description: Repair cycles, dangling edges, and other broken graph-owned state without requiring an unsafe generic mutation path.
effort: 3-5 days
tier: 2
priority: high
autonomy_level: 3
tags: [threads, graph, storage, cli]
created: "2026-08-28"
depends_on: [6g3q4rt7mgjn, 6g697mp8s4tx, 6g6scc9jgxae, 6g721vewvvrz, 6g72ncs4xjdm]
updated_at: "2026-09-06"
started_at: "2026-09-05"
audit_sources: [planning/audits/6g71vzq8wdnj-2026-09-05-add-a-guarded-repair-path-for-broken-dependency-graphs-antigravity.md, planning/audits/6g71yr50md16-2026-09-05-add-a-guarded-repair-path-for-broken-dependency-graphs-claude.md, planning/audits/6g7cr4psd1nk-2026-09-06-guarded-broken-graph-repair-implementation-claude.md, planning/audits/6g7cr4q1vhms-2026-09-06-guarded-broken-graph-repair-implementation-antigravity.md]
completed_at: "2026-09-06"
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

- [x] A broken source graph can enter only the dedicated repair planner; ordinary add/remove/migrate,
      lifecycle, generic task mutation, and `lint --fix` continue to fail closed.
- [x] Validation runs over the full source-level projection and preserves duplicate-ID shadow
      records, unreadable records/revisions, duplicate declarations, raw invalid values, and legacy
      declaration ownership. It does not reconstruct a prospective repository from representative
      `TaskGraph.Task()` values.
- [x] Progress and preservation are independent hard checks: the complete plan strictly improves
      the structural defect measure, every prefix is structurally non-worsening and discharges
      selected intent, and no unrelated declaration or valid constraint disappears.
- [x] Cycle, self-edge, dangling-reference, invalid-ID, duplicate-edge, and each legacy-field fixture
      have an actionable preview and converge to the selected repaired state, whether or not
      unrelated residual problems leave the repository broken.
- [x] Bare diagnosis and `--dry-run` explain each defect, distinguish auto-safe from explicit
      repairs, show exact copyable selectors, and predict residual problems without writing.
- [x] `--auto` is limited to duplicate/self/empty-legacy cleanup. Invalid and dangling values remain
      verbatim until explicitly dropped; cycle and ambiguous-legacy choices are never guessed.
- [x] Concurrent task, dependency, and Thread-evidence edits—including byte
  changes to still-unreadable task or Thread files—produce a typed conflict
  rather than a stale repair. Every injected durable prefix is diagnosable,
  convergent on retry, and never repeats an already-satisfied dedupe or drop
  intent.
- [x] Human and JSON receipts derive `Changed` from actual materialized writes and report `Committed`,
      initial/final health, selected and removed declarations, addressed and residual defects, raw
      removed values, workspace, applied/remaining files, task-state impacts, readable Thread
      impacts, and incomplete Thread diagnostics.
- [x] Normal lint and other graph-health surfaces point to defect-specific repair diagnosis once it
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

## Implementation progress (2026-09-06)

The dedicated core/store port, source-level diagnosis and removal planner, independent progress and
preservation proofs, surgical filesystem materializer, task-and-Thread evidence CAS, convergent
partial receipts, CLI/manifest surface, human/JSON rendering, and lint/status guidance are now
implemented. The validator also rejects mixed dedupe/exact-drop effects and reauthorizes selected
receipt intent so metadata cannot outrun the guarded operations. Copyable selectors retain raw
colon-bearing and terminal-`#<digits>` values.

Focused fixtures cover every canonical and legacy defect class, residual-broken repairs,
duplicate-ID shadows, YAML aliases, unreadable evidence, exact YAML-node preservation, readable
Thread impacts, late readable/unreadable task and Thread edits, injected durable prefixes, and
convergent retries. The Claude implementation audit found and drove fixes for shadow-owned
self-declarations, the independent containment proof, alias materialization, mutation-killing guard
coverage, truthful pre-write receipts, documentation, and growing-prefix validation cost. Per-write
validation now proves and composes one source group while retaining fresh task and Thread evidence
reads for out-of-band editor safety; the retained benchmark measured 100 repaired files at 1.76s
and 200 at 5.46s, versus the audit baseline of 6.33s and 37.54s. All seven findings are fixed and
both implementation audits are closed.

Validation is clean under the full uncached `go test -count=1 -race ./...`, golangci-lint, module
tidiness, planning and audit lint, repeat generation of CLI/schema artifacts, and
`git diff --check`.

## Related

- Epic [30-threads-and-task-dependency-graphs](../epics/30-threads-and-task-dependency-graphs.md)
- Prerequisite [unreadable-source revision tokens](6g721tvf4crh-version-unreadable-task-graph-sources-for-repair-safe-cas.md)
- Prerequisite [unreadable Thread-source revision tokens](6g72ncs4xjdm-version-unreadable-thread-sources-for-repair-safe-cas.md)
- Prerequisite [source-level graph declarations](6g721vewvvrz-model-graph-owned-source-declarations-for-sound-repair.md)
- Follow-up research [Markdown-first storage durability](6g721w07mv1d-reassess-markdown-first-storage-durability-for-relational-planning-data.md)
