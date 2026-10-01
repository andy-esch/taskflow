---
schema: 1
id: 6ge1bacd3bd2
status: completed
epic: 21-code-quality-architecture-hardening
description: Attribute duplicate and corrupt readable records using optional opaque source locations without treating location as identity or local path.
effort: 2-3 days
tier: 2
priority: medium
autonomy_level: 3
tags: [architecture, diagnostics, ports, json]
created: "2026-09-26"
depends_on: [6gcwcf80v8hg]
updated_at: "2026-09-30"
started_at: "2026-09-28"
completed_at: "2026-09-30"
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

- [x] Duplicate readable records with one canonical ID but distinct available locations are
      individually attributable in graph/list/lint diagnostics, even with equal semantic values.
- [x] Ordinary full and projected JSON, human output, and generated schema document the additive
      location contract for all four kinds; no URI is emitted as a filesystem `path`.
- [x] Graph corruption stays fail-closed, source ID remains canonical entity identity, an explicit
      local path is the only separate physical repair selector, and guarded mutation/CAS tests
      retain their existing guarantees.
- [x] Fixtures exercise a pathless adapter, mixed local/opaque context, absent location, and
      contradictory declared ID/slug/filename metadata.

## Out of scope

- Using location as a stable identity, lookup selector, or local file handle.
- Removing domain `Path`, `FilenameID`, or `SourceVersion`; the sequenced source/path task owns that.

## Implementation and review

Readable source context now reaches duplicate graph diagnostics, ordinary list/lint DTOs,
opt-in `location` projections, human output, and schema revisions 1.79–1.80. Equal-value records with
distinct opaque locations, absent/redundant locations, contradictory declared IDs, and CAS
path/version checks have focused tests. The full Go suite, core/store race tests, vet, static
lint, and planning lint pass. Both adversarial reviews are dispositioned and closed; all four
acceptance criteria are checked.

The repair seam is now split: `TaskGraphSourceRef.Location` is diagnostic context and
`LocalPath` is the optional physical repair handle. Pathless readable sources are diagnosed
but not offered for local repair; explicit URI selectors are rejected. Review exposed a
pathless-defect receipt identity bug, now fixed by preserving a non-replayable declaration
and matching each raw occurrence separately. Healthy-graph snapshot, stale-context,
legacy attribution, and unreadable-duplicate tests now distinguish location from local
path. The source/path separation task still owns retiring the transitional domain `Path`
carrier and explicitly supplying the local repair capability.

## Related

- Epic [21-code-quality-architecture-hardening](../epics/21-code-quality-architecture-hardening.md)
- Thread [Make planning data access adapter neutral](../threads/6gcwd78p9r04-make-planning-data-access-adapter-neutral.md)
- [Ordinary entity read snapshots](6gcwcf80v8hg-make-ordinary-entity-list-diagnostics-adapter-neutral.md)
- [Codex implementation audit](../audits/6ge0q80cc01r-2026-09-26-ordinary-entity-read-snapshots-implementation-codex.md), finding M2
- [Source/path separation](6gcwcf88z57p-split-local-path-capabilities-from-semantic-entity-reads.md)
- [Claude implementation review](../audits/6gf9nsb0d8ca-2026-09-30-readable-source-location-implementation-claude.md)
- [Antigravity implementation review](../audits/6gf9nsbe7vps-2026-09-30-readable-source-location-implementation-antigravity.md)
