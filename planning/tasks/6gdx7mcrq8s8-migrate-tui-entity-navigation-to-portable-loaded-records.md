---
schema: 1
id: 6gdx7mcrq8s8
status: completed
epic: 21-code-quality-architecture-hardening
description: Move TUI identity, refresh, and local-action state onto adapter-neutral read evidence while preserving fail-closed ambiguity.
effort: 5-8 days
tier: 2
priority: medium
autonomy_level: 2
tags: [architecture, tui, ports]
audit_sources: [planning/audits/6gefxf155mea-2026-09-28-portable-tui-loaded-record-navigation-implementation-codex.md, planning/audits/6gefxfparh2z-2026-09-28-portable-tui-loaded-record-navigation-implementation-antigravity.md]
created: "2026-09-26"
depends_on: [6gcwcf80v8hg]
updated_at: "2026-09-28"
started_at: "2026-09-27"
completed_at: "2026-09-28"
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
- Extend the core Thread and in-progress dashboard/Atlas projections to carry source identity from
  the same adapter snapshot as their semantic values. Keep release-facing CLI/wire shapes stable
  while their remaining bare-domain consumers migrate; the TUI must not infer an ID from those
  compatibility values.
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

- [x] TUI list/detail keys for every entity family come from portable source identity and no longer
      depend on a domain path or filename-derived fallback.
- [x] Missing and duplicate canonical source IDs visibly fail/degrade the loaded registry before any
      show, edit, copy-path, or mutation action can address an occurrence.
- [x] Snapshot-local slice positions and source locations are never accepted as durable or writable
      entity identity.
- [x] Selection restoration remains stable across reorder/refresh for a unique source ID and clears
      safely when that identity becomes ambiguous.
- [x] Delayed detail and local-action results cannot act on a different selection after file-watch or
      manual refresh changes the list generation.
- [x] Pathless semantic browsing remains viable and tests pin the handoff to the later optional-path
      capability slice.

## Out of scope

- Rendering two duplicate-ID records as independently actionable TUI rows.
- Introducing snapshot-local row handles for durable navigation or mutation.
- Implementing the local-path capability split owned by the downstream source/path task.

## Scope expansion (2026-09-27)

The original entity-tab scope could not make Thread, dashboard, or Atlas navigation portable:
their projections exposed bare domain values without source identity. This task therefore also
changed those core read contracts, preserving CLI/wire compatibility and taking source evidence
from the same adapter snapshot rather than a second scan or domain fallback.

## Implementation evidence (2026-09-28)

All five entity tabs now use `RecordSource.ID` for list/detail keys, cursor restoration, structured
task navigation, and action targets. Invalid or duplicate identities quarantine the loaded list;
generation- and source-stamped detail/action results cannot update a newer selection. The dashboard
and Atlas consume identity-bearing in-progress records from the same Summary snapshot. Thread
list projections carry source identity from one versioned `ThreadRead`, while selected show reads
carry their own source identity. Guarded ordinary writers reject readable sources with missing
identity, duplicate IDs, or declared-ID drift; graph repair retains its intentional
incomplete-Thread exception. Source revisions stay outside semantic Thread values and public wire
output. Pathless Thread detail remains browsable while local actions explain their unavailable
capability. Full Go tests, focused race tests, vet, and planning lint pass.

The downstream [source/path split](6gcwcf88z57p-split-local-path-capabilities-from-semantic-entity-reads.md)
still owns removing compatibility `Path`, `FilenameID`, and `SourceVersion` fields and switching
remaining TUI local actions to explicit optional path capabilities. This task leaves those fields
in place without treating them as navigation identity.

## Codex review checkpoint (2026-09-28)

The Codex implementation review found three real gaps, now fixed and marked in the audit: a
status/bucket-filtered tab could hide a duplicate source ID; acute audit findings could link to an
audit whose duplicate had no acute findings; and Thread JSON replaced the historical declared
`thread.id` with its source ID. Task, epic, and audit tabs now validate full-family identities
before filtering; complete task/audit source IDs from the existing Summary reads also guard
Overview and Atlas jumps. Hidden duplicates cannot become navigation targets.

Antigravity's independent review reported no additional findings. Its sandbox captured the
pre-Codex-fix implementation, however, so its clean verdict is not an endorsement of the final
patch. In particular, its wire probe observed and accepted the changed public `thread.id` that
Codex identified as a compatibility regression. Its duplicate probes also did not exercise
cross-status or cross-bucket collisions. Those cases now have explicit regression tests; the
current patch passes the full Go suite, focused race tests, planning lint, and audit lint.

## Related

- Epic [21-code-quality-architecture-hardening](../epics/21-code-quality-architecture-hardening.md)
- Thread [Make planning data access adapter neutral](../threads/6gcwd78p9r04-make-planning-data-access-adapter-neutral.md)
- [Portable entity-read design](6gcwcf7rgxef-design-portable-entity-reads-and-optional-local-source-capabilities.md)
- [Codex design finding M3](../audits/6gdx1bvdf16p-2026-09-26-portable-entity-read-local-source-capability-design-codex.md)
- [Antigravity design finding M2](../audits/6gdx1bvz683y-2026-09-26-portable-entity-read-local-source-capability-design-antigravity.md)
- [Codex implementation review](../audits/6gefxf155mea-2026-09-28-portable-tui-loaded-record-navigation-implementation-codex.md)
- [Antigravity implementation review](../audits/6gefxfparh2z-2026-09-28-portable-tui-loaded-record-navigation-implementation-antigravity.md)
