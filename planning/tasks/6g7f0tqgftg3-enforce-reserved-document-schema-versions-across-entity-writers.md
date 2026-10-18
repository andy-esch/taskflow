---
schema: 1
id: 6g7f0tqgftg3
status: ready-to-start
epic: 26-frontmatter-schema-declared-validation-contract
description: Turn the reserved on-disk schema marker into one cross-entity read diagnostic and guarded-write compatibility boundary.
effort: 2-3 days
tier: 3
priority: medium
autonomy_level: 3
tags: [schema, frontmatter, validation, compatibility]
created: "2026-09-06"
depends_on: [6fkkz41cax80]
updated_at: "2026-10-18"
audit_sources: [2026-09-07-arch-data-model-and-storage, 2026-10-18-weekly-task-sweep]
audited: "2026-10-18"
---

# Enforce reserved document schema versions across entity writers

## Objective

Decide and implement one coherent meaning for the coarse `schema` key already stamped into planning
documents. Avoid both failure modes: treating an ignored marker as a real safety boundary, and
adding a Thread-only check that leaves every sibling entity with different upgrade behavior.

## Scope

- Follow the policy decisions in `adr-close-frontmatter-schema-policy-questions`, especially its
  entity coverage, unknown-field, rollout, and version-evolution questions.
- Define shared read behavior for missing, current, older, malformed, and newer schema values across
  tasks, epics, audits, research, Threads, and future registered entity kinds.
- Define separately whether diagnostic reads may retain usable data and when any tool-owned writer
  must fail closed rather than rewrite a document from an unsupported format.
- Route parsing/lint diagnostics and guarded/generic writers through one declared policy rather than
  per-entity checks. Preserve adapter-neutral evidence; filesystem paths are optional context.
- Provide an explicit migration contract before the first real schema bump, including dry-run,
  idempotency, partial-failure recovery, and Git-native rollback expectations.
- Inventory existing schema-less fixtures and real planning data so rollout is deliberate rather
  than an accidental mass failure.

## Acceptance criteria

- [ ] One accepted policy defines the meaning of missing, old, current, malformed, and newer
      document schema values for both reads and writes.
- [ ] A shared implementation covers every first-class document kind without teaching individual
      parsers or writers divergent version rules.
- [ ] Normal lint reports schema incompatibility with stable adapter-neutral diagnostics, and every
      tool-owned writer observes the same fail-open/fail-closed decision before mutation.
- [ ] Current-schema and intentionally supported legacy documents retain unknown fields, bodies,
      comments, identity, and existing surgical-write guarantees.
- [ ] Unsupported-format tests prove no write occurs through generic edits, lifecycle verbs,
      dependency/Thread compound mutations, TUI actions, or future primary adapters.
- [ ] Migration behavior is implemented or explicitly deferred until a real bump, but the first
      bump cannot ship without a named executable path and compatibility fixtures.
- [ ] Corpus, demo, cross-platform, race, lint, and machine-diagnostic coverage pass.

## Stress tests

- Schema-less pre-marker fixtures, schema 1 with unknown fields, malformed scalar/list/map schema
  values, a future version, archived entities, unreadable records, external planning repositories,
  pathless adapters, and concurrent compound graph/Thread writes.

## Out of scope

- Adopting schema enforcement as part of Thread preview graduation.
- Designing a runtime/plugin schema language or conflating document `schema` with JSON
  `schema_version`.
- Inventing schema 2 without a real incompatible persisted change.

## Related

- Epic [26-frontmatter-schema-declared-validation-contract](../epics/26-frontmatter-schema-declared-validation-contract.md)
- Policy task [ADR — declared frontmatter-schema contract](6fkkz41cax80-adr-close-frontmatter-schema-policy-questions.md)
- Thread graduation contract [task](6g6wdvfjdaaa-define-the-thread-preview-graduation-and-compatibility-contract.md)
- Compatibility guide [Threads compatibility and preview graduation](../../docs/THREADS_COMPATIBILITY.md)

Reinforced by audit 2026-09-07-arch-data-model-and-storage: M2. Corpus inventory as of 2026-09-07 — 93 documents carry no `schema:` key (86/334 tasks, 5/15 epics, 2/69 audits; research and Threads are complete). `lint --fix` does not backfill it, so absence must be decided (backfill now, or declare absence == 1 permanently) before the first real bump.

## Sweep refresh (2026-10-18)

2026-10-18: automated weekly sweep — the design is unchanged and unstarted, but
the corpus inventory at the foot of this task has moved and a new document kind
has appeared that the rollout decision now has to cover.

**Corpus inventory refreshed (2026-10-18).** The original count was taken on
2026-09-07. Re-counting `grep -L '^schema:'` across `planning/`:

| Kind | Missing `schema:` 2026-09-07 | Missing `schema:` 2026-10-18 |
| --- | --- | --- |
| tasks | 86 / 334 | 86 / 426 |
| epics | 5 / 15 | 5 / 15 |
| audits | 2 / 69 | 2 / 193 |
| research | 0 / — | 0 / 33 |
| threads | 0 / — | 0 / 10 |
| **ADRs** | not counted | **7 / 8** |

The three original absolute counts are *identical* six weeks later while the
denominators grew by 92 tasks and 124 audits. That is the useful signal: every
document written by the tool since the marker shipped carries `schema:`, so the
93 (now 94, excluding ADRs) schema-less documents are a closed, pre-marker
backlog that is not growing. Backfilling is a bounded one-time migration, not a
moving target — which makes "backfill now" materially cheaper than the
2026-09-07 framing suggested.

**New information — ADRs are a document kind this task does not yet cover.**
`planning/adrs/` holds 8 documents and 7 carry no `schema:` key. They are not yet
a registered entity: `tskflwctl schema` lists doc kinds as
`task, epic, audit, research, thread`, so ADRs are neither parsed nor linted as
entities today. But ADRs are an explicit deliverable of
epic 28-first-class-entities-new-planning-nouns (currently 63%, "routines,
projects, ADRs"). This task's Scope line says "tasks, epics, audits, research,
Threads, and future registered entity kinds" — ADRs are exactly that case, and
they will arrive *already carrying* a schema-less legacy corpus. Whichever way
the policy decides absence, the ADR noun should not be registered before the
decision lands, or it inherits the same ambiguity on day one.

**Verified accurate (2026-10-18).** `lint --fix` still does not backfill
`schema:` — its help text scopes repairs to "ordinary frontmatter and missing
ids" with graph-owned fields skipped, and nothing in `internal/store/fix.go`
touches the schema key. The policy dependency and both cross-links resolve.

**Dependency state (2026-10-18).** `gate: blocked`, `eligible: false`,
`role: candidate`. The single blocker is `6fkkz41cax80`
(`adr-close-frontmatter-schema-policy-questions`, `next-up`, reason
`not-started`), which matches this task's Scope line exactly — no prose/graph
disagreement. That blocker is itself `gate: clear` and eligible, so this task is
one task away from startable.
