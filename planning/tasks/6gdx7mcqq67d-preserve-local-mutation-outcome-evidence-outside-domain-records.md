---
schema: 1
id: 6gdx7mcqq67d
status: ready-to-start
epic: 21-code-quality-architecture-hardening
description: Keep planned, committed, and partially durable local paths in operation-specific receipts before domain paths are removed.
effort: 2-3 days
tier: 2
priority: medium
autonomy_level: 2
tags: [architecture, mutations, filesystem]
created: "2026-09-26"
depends_on: [6gcwcf7rgxef]
updated_at: "2026-09-26"
---

# Preserve local mutation outcome evidence outside domain records

## Objective

Move local destination and recovery evidence into operation-specific mutation receipts before
semantic domain records lose their `Path` fields. This preserves dry-run output and partial
durability diagnostics without making portable records or pathless adapters pretend that a local
filesystem path is semantic data.

## Scope

- Inventory task, epic, audit, research, and Thread create flows plus task rename and any other
  operation whose current human or JSON contract obtains a path from a returned domain record.
- Give local create receipts planned and, when committed, resulting local paths. Dry runs must remain
  useful even though no record exists for a later path lookup.
- Give rename or relocation receipts exact source and destination paths plus the durable stage needed
  to recover after destination-write, inbound-link, or source-cleanup failure.
- Preserve operation-specific local evidence through success, partial-error, human, and JSON
  projections. Keep it optional so remote or database adapters are not required to invent a path.
- Use post-success path resolution only for committed, non-moving operations when the resolver is
  bound to the same source set and no recovery contract needs the original location.
- Remove receipt dependence on domain `Path` while retaining current command wording, safety
  classification, exit behavior, and machine compatibility.

## Acceptance criteria

- [ ] Every create dry run that currently reports a path still reports its planned local destination
      after domain paths are removed.
- [ ] Committed creates report the resulting local path when their adapter supports one, and pathless
      creates express absence without a fabricated location.
- [ ] Task rename partial failures retain exact source/destination paths and durable progress in both
      human and machine-readable recovery output.
- [ ] Mutation receipts use operation-specific values rather than one generic entity mutation
      envelope or a path-bearing domain record.
- [ ] Tests cover dry run, success, pathless success, and each existing durable-prefix failure stage.

## Out of scope

- Adding local-path output to operations that do not currently promise it.
- Treating frontmatter lifecycle changes as file moves when the flat layout does not relocate them.
- Generalizing all entity mutations behind one repository interface.

## Related

- Epic [21-code-quality-architecture-hardening](../epics/21-code-quality-architecture-hardening.md)
- Thread [Make planning data access adapter neutral](../threads/6gcwd78p9r04-make-planning-data-access-adapter-neutral.md)
- [Portable entity-read design](6gcwcf7rgxef-design-portable-entity-reads-and-optional-local-source-capabilities.md)
- [Codex design finding M2](../audits/6gdx1bvdf16p-2026-09-26-portable-entity-read-local-source-capability-design-codex.md)
- [Antigravity design finding L2](../audits/6gdx1bvz683y-2026-09-26-portable-entity-read-local-source-capability-design-antigravity.md)
