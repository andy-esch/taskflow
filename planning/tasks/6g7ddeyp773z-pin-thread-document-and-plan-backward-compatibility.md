---
schema: 1
id: 6g7ddeyp773z
status: next-up
epic: 30-threads-and-task-dependency-graphs
description: Make the persisted Thread and bulk-plan upgrade promise executable with historical compatibility fixtures.
effort: 1-2 days
tier: 2
priority: high
autonomy_level: 4
tags: [threads, compatibility, testing]
created: "2026-09-06"
depends_on: [6g6wdvfjdaaa]
updated_at: "2026-09-06"
---

# Pin Thread document and plan backward compatibility

## Objective

Turn the persisted upgrade promise in the Thread graduation contract into executable coverage.
Prove that current code can consume the artifacts shipped in v0.18.0 and v0.19.0 without promoting
the repository's currently advisory document `schema` marker into a Thread-only policy.

## Scope

- Add provenance-labelled fixtures representing the schema-1 Thread documents, authoring
  manifests, and materialized apply plans emitted by the two preview releases.
- Exercise current list/show/projection and guarded mutation paths against old Thread documents,
  including shared membership, lifecycle timestamps, and an unknown additive field that must
  survive a surgical update.
- Pin the actual historical Thread shape rather than inferring guarantees from `schema: 1`. A
  future incompatible persisted change must add read compatibility and an explicit migration before
  it can ship; adopting a cross-entity schema enforcement policy is tracked separately.
- Pin schema-zero authoring-manifest shorthand (both an omitted key and explicit zero), explicit
  schema-1 authoring, and strict schema-1 apply behavior. An unsupported manifest or plan must fail
  before repository mutation and identify the supported version or remedy.
- Replay retained plans against repository configuration carrying the expected durable ID. Also
  cover a pre-migration repository with no durable ID, which must fail before mutation and name
  `config migrate` as the remedy. The user-scoped spaces registry is not required plan context.
- Add compatibility-focused command-level assertions for the stable Thread read, mutation, update,
  compose, and apply envelopes—including failure receipts—without freezing additive fields or
  human presentation.

## Acceptance criteria

- [ ] Committed fixtures with release provenance cover the persisted Thread and bulk-link artifacts
      users could retain from v0.18.0 and v0.19.0; the current binary reads and projects them.
- [ ] A guarded membership or lifecycle update of an old schema-1 Thread preserves its stable ID,
      body, comments, key order where already promised, and unknown additive frontmatter.
- [ ] The compatibility fixture proves the fields and meanings actually emitted by the preview
      releases, while documenting that the shared document `schema` marker is advisory and not a
      Thread-only mutation gate.
- [ ] Schema-zero (omitted and explicit) and explicit schema-1 authoring manifests, plus strict
      schema-1 apply plans, retain their documented behavior; unsupported versions fail without
      writes and a schema-1 interrupted plan remains retryable.
- [ ] A retained plan replays against migrated repository identity; the same plan against legacy
      configuration without a durable ID fails before mutation with the `config migrate` remedy.
- [ ] Stable Thread JSON fields, meanings, roles, health values, edge direction, lifecycle
      operations, and error classifications have command-level compatibility coverage. The four
      mutation-side envelopes (`thread_mutation`, `thread_update`, `thread_compose`, and
      `thread_apply`) are exercised with non-default values and at least one structured failure.
- [ ] Focused tests, full race tests, lint, generated schema/docs checks, and planning lint pass.

## Stress tests

- Unknown frontmatter, old lifecycle timestamps, shared membership, an external gate, a partial
  apply receipt, malformed Thread evidence during graph repair, and a pathless Thread adapter.

## Out of scope

- Designing schema 2 or implementing a speculative migration with no format change to perform.
- Backward compatibility for arbitrary hand-written files that were never valid schema-1 Threads.
- Enforcing the reserved document `schema` key; that is cross-entity work in
  [`6g7f0tqgftg3`](6g7f0tqgftg3-enforce-reserved-document-schema-versions-across-entity-writers.md).
- Freezing human CLI/TUI layout or Mermaid/DOT formatting.

## Related

- Epic [30-threads-and-task-dependency-graphs](../epics/30-threads-and-task-dependency-graphs.md)
- Contract [Threads compatibility and preview graduation](../../docs/THREADS_COMPATIBILITY.md)
- ADR [0006 — Adopt Threads as task DAGs](../adrs/0006-adopt-threads-as-task-dags.md)
- Preview checkpoints [v0.18.0](6g5m69wpydzw-cut-v0.18.0-as-a-cli-threads-preview.md) and
  [v0.19.0](6g6scc9jgxae-cut-v0.19.0-as-a-tui-threads-preview.md)
