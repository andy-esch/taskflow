---
schema: 1
id: 6gbn4g1j40pj
status: ready-to-start
epic: 20-cli-ux-and-ergonomics
description: Add deterministic optional bounds, continuation, sorting, and filtering to planning-entity list queries without changing unbounded defaults.
effort: M
tier: 2
priority: high
autonomy_level: 3
tags: [cli, query, agents, json]
created: "2026-09-19"
---

# Bound and page agent-facing list queries

## Objective

Let agents ask for a deterministic, resumable subset of planning entities instead of ingesting an
unbounded corpus. Define the query and continuation contract in shared application/core terms so a
filesystem adapter can scan locally today without locking future database or web adapters into CLI
or Markdown details.

## Scope

- Inventory task, epic, audit, research, Thread, and audit-finding list surfaces and define which
  common bounds apply truthfully to each.
- Add deterministic optional `--sort`, `--limit`, and continuation behavior plus bounded text and
  updated-since filtering where the underlying entity contract supports them.
- Return enough JSON metadata to resume without guessing: applied ordering, returned count, and an
  opaque or explicitly versioned next cursor. Preserve current unbounded behavior when no bound is
  requested.
- Keep CLI rendering downstream of a reusable query/result contract; do not claim filesystem reads
  avoid a full scan when only their returned projection is bounded.

## Acceptance criteria

- [ ] A written contract defines stable ordering, tie-breaking, cursor invalidation, filter
      composition, limits, and behavior when the corpus changes between pages.
- [ ] Supported list surfaces expose consistent optional bounds and reject unsupported or ambiguous
      combinations with structured validation errors.
- [ ] JSON responses publish deterministic continuation metadata without making human/CSV output
      parse cursors from prose; legacy unbounded calls remain compatible.
- [ ] The reusable query boundary is independent of filesystem paths, Cobra flags, and Markdown
      parsing, with adapter and CLI tests for empty, exact-boundary, final-page, mutation-between-
      pages, and invalid-cursor cases.
- [ ] Contract revision/classification, schema/docs, token-volume evidence, full tests, lint,
      planning lint, and diff checks are recorded.

## Out of scope

- Claiming keyset pagination avoids a filesystem scan in the local adapter.
- A general search index, ranking engine, remote API, or database migration.
- Changing body-mutation receipts; compact receipts remain a separate compatibility decision.

## Related

- Audit [AI-agent CLI ergonomics, M1](../audits/6fsa47r4f7es-2026-07-24-ai-agent-cli-ergonomics.md)
- Task [Make Thread list projectable](6gbn4g1eypzs-make-thread-list-projectable-like-other-planning-entity-lists.md)
- ADR [Monotonic JSON machine-contract revisions](../adrs/0008-use-monotonic-revisions-for-the-json-machine-contract.md)
