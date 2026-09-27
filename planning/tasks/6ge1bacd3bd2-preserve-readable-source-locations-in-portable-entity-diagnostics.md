---
schema: 1
id: 6ge1bacd3bd2
status: ready-to-start
epic: 21-code-quality-architecture-hardening
description: Attribute duplicate and corrupt readable records using optional opaque source locations without treating location as identity or local path.
effort: 2-3 days
tier: 2
priority: medium
autonomy_level: 3
tags: [architecture, diagnostics, ports, json]
created: "2026-09-26"
depends_on: [6gcwcf80v8hg]
updated_at: "2026-09-26"
---

# Preserve readable source locations in portable entity diagnostics

## Objective

Make distinct physical occurrences of a readable record identifiable in diagnostics when an
adapter supplies only opaque locations. A canonical source ID remains the sole entity identity;
location is explanatory context, never a selector or filesystem path.

## Scope

- Preserve `RecordSource.Location` through task-graph duplicate attribution and any graph source
  references used for diagnostics, without weakening guarded snapshot comparison or repair safety.
- Expose optional readable location context for task, epic, audit, and research list rows and
  relevant lint diagnostics. Decide the smallest additive wire shape and revise generated schema,
  golden fixtures, projected columns, and human output together.
- Keep `LocalPath` independent: URI-like locations must not become a JSON `path`, path command
  result, repair target, or identity fallback.
- Test same-ID/same-slug records at different opaque locations for all four entity kinds, including
  task graph health and duplicate messages. Cover empty locations and local path plus opaque
  location, and show that no location can make a source-less record addressable.

## Acceptance criteria

- [ ] Duplicate readable records with one canonical ID but distinct available locations are
      individually attributable in graph/list/lint diagnostics, even with equal semantic values.
- [ ] Ordinary full and projected JSON, human output, and generated schema document the additive
      location contract for all four kinds; no URI is emitted as a filesystem `path`.
- [ ] Graph corruption stays fail-closed, source ID stays the only selectable identity, and guarded
      mutation/CAS tests retain their existing guarantees.
- [ ] Fixtures exercise a pathless adapter, mixed local/opaque context, absent location, and
      contradictory declared ID/slug/filename metadata.

## Out of scope

- Using location as a stable identity, lookup selector, or local file handle.
- Removing domain `Path`, `FilenameID`, or `SourceVersion`; the sequenced source/path task owns that.

## Related

- Epic [21-code-quality-architecture-hardening](../epics/21-code-quality-architecture-hardening.md)
- Thread [Make planning data access adapter neutral](../threads/6gcwd78p9r04-make-planning-data-access-adapter-neutral.md)
- [Ordinary entity read snapshots](6gcwcf80v8hg-make-ordinary-entity-list-diagnostics-adapter-neutral.md)
- [Codex implementation audit](../audits/6ge0q80cc01r-2026-09-26-ordinary-entity-read-snapshots-implementation-codex.md), finding M2
- [Source/path separation](6gcwcf88z57p-split-local-path-capabilities-from-semantic-entity-reads.md)
