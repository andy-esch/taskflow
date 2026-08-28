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
depends_on: [6g3q4rt7mgjn]
updated_at: "2026-08-28"
---

# Add a guarded repair path for broken dependency graphs

## Objective

Provide an explicit recovery capability for graph-owned frontmatter that is already broken, without making generic setters or ordinary dependency mutations capable of bypassing repository-global validation.

## Scope

- Repair canonical cycles, self-edges, duplicate edges, and missing/invalid `depends_on` references through one guarded, narrowly typed operation.
- Admit only plans that strictly reduce a deterministic graph-problem measure at every durable prefix and never introduce a new problem class or affected task.
- Preserve the existing repository lock, snapshot-local planning, whole-snapshot/per-file CAS, surgical frontmatter updates, dry-run, and typed recovery receipts.
- Provide exact file/field/problem diagnoses and an idempotent retry path when a multi-file repair stops after a durable prefix.
- Decide whether the user surface belongs under `task depend repair`, `lint --fix --graph`, or a smaller family of explicit removal verbs before freezing CLI syntax.

## Acceptance criteria

- [ ] A broken source graph can enter only the dedicated repair planner; ordinary add/remove/migrate and generic task mutation continue to fail closed.
- [ ] The repair proof uses a documented deterministic measure and rejects any plan whose prefix fails to improve or preserve safety while moving monotonically toward a healthy graph.
- [ ] Cycle, self-edge, dangling-reference, invalid-ID, and duplicate-edge fixtures each have an actionable preview and converge to the intended healthy state.
- [ ] A repair cannot silently discard an unrelated valid constraint, widen the affected component, or reinterpret legacy slug references.
- [ ] Concurrent task/dependency edits produce a typed conflict rather than a stale repair, and every injected durable prefix is diagnosable and resumable.
- [ ] Human and JSON output name every removed/replaced declaration, the original problems addressed, remaining problems, workspace, and applied/remaining files.
- [ ] Normal lint points to the repair command once it exists; until then it remains explicit that direct frontmatter editing is the only recovery path.

## Out of scope

- Automatic best-guess repair of ambiguous legacy references.
- A generic `--force` escape hatch for arbitrary graph writes.
- Rewriting lifecycle status or Thread membership as part of dependency repair.

## Sequencing

This is a parallel recovery-hardening task, not a prerequisite for dependency eligibility or the first Thread entity slice. Promote it immediately if dogfooding encounters broken graph-owned state that the normal guarded operations cannot repair. Otherwise reassess it before bulk linking, when wider graph use and longer apply plans make an explicit recovery path more valuable.

## Related

- Epic [30-threads-and-task-dependency-graphs](../epics/30-threads-and-task-dependency-graphs.md)
