---
schema: 1
id: 6gdx7mcrq8s8
status: ready-to-start
epic: 21-code-quality-architecture-hardening
description: Move TUI identity, refresh, and local-action state onto adapter-neutral read evidence while preserving fail-closed ambiguity.
effort: 3-5 days
tier: 2
priority: medium
autonomy_level: 2
tags: [architecture, tui, ports]
created: "2026-09-26"
depends_on: [6gcwcf80v8hg]
updated_at: "2026-09-26"
---

# Migrate TUI entity navigation to portable loaded records

## Objective

Move task, epic, audit, research, and Thread TUI identity and refresh state onto the portable loaded
record contract before `CanonicalID()`, filename identity, and local paths leave domain values. Keep
the existing deliberate fail-closed behavior for corrupt identity instead of creating a temporary
row identity that could make an ambiguous record appear safe to mutate.

## Scope

- Build entity list/detail projections from adapter-supplied canonical source IDs and semantic
  values, not domain `FilenameID`, `CanonicalID()` fallbacks, or paths.
- Preserve the entity registry's fail-closed response to missing or duplicate canonical keys. CLI
  list/lint may display every corrupt occurrence with source context; the interactive TUI must not
  turn snapshot position or location into a writable identity.
- Keep durable selection restoration keyed by canonical source ID and clear/degrade it when the new
  snapshot makes that ID missing or ambiguous.
- Make refresh and lazy detail/local-action responses carry the originating generation and source
  ID, then discard them if either no longer matches the selected record.
- Introduce adapter-neutral TUI item projections where useful rather than spreading generic
  `LoadedRecord[T]` mechanics throughout rendering code.
- Retain current local actions during this preparatory slice; the sequenced source/path task switches
  them to explicit optional path capabilities after the identity model is stable.

## Acceptance criteria

- [ ] TUI list/detail keys for every entity family come from portable source identity and no longer
      depend on a domain path or filename-derived fallback.
- [ ] Missing and duplicate canonical source IDs visibly fail/degrade the loaded registry before any
      show, edit, copy-path, or mutation action can address an occurrence.
- [ ] Snapshot-local slice positions and source locations are never accepted as durable or writable
      entity identity.
- [ ] Selection restoration remains stable across reorder/refresh for a unique source ID and clears
      safely when that identity becomes ambiguous.
- [ ] Delayed detail and local-action results cannot act on a different selection after file-watch or
      manual refresh changes the list generation.
- [ ] Pathless semantic browsing remains viable and tests pin the handoff to the later optional-path
      capability slice.

## Out of scope

- Rendering two duplicate-ID records as independently actionable TUI rows.
- Introducing snapshot-local row handles for durable navigation or mutation.
- Implementing the local-path capability split owned by the downstream source/path task.

## Related

- Epic [21-code-quality-architecture-hardening](../epics/21-code-quality-architecture-hardening.md)
- Thread [Make planning data access adapter neutral](../threads/6gcwd78p9r04-make-planning-data-access-adapter-neutral.md)
- [Portable entity-read design](6gcwcf7rgxef-design-portable-entity-reads-and-optional-local-source-capabilities.md)
- [Codex design finding M3](../audits/6gdx1bvdf16p-2026-09-26-portable-entity-read-local-source-capability-design-codex.md)
- [Antigravity design finding M2](../audits/6gdx1bvz683y-2026-09-26-portable-entity-read-local-source-capability-design-antigravity.md)
