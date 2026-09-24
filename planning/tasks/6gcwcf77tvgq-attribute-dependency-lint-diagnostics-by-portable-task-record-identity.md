---
schema: 1
id: 6gcwcf77tvgq
status: ready-to-start
epic: 21-code-quality-architecture-hardening
description: Stop semantic graph lint findings from disappearing or colliding when a task source has no filesystem path.
effort: 1-2 days
tier: 2
priority: medium
autonomy_level: 3
tags: [architecture, lint, diagnostics, ports]
created: "2026-09-23"
depends_on: [6g5vm4efjcdv]
updated_at: "2026-09-23"
---

# Attribute dependency lint diagnostics by portable task record identity

## Objective

Make semantic dependency lint attribution independent of filesystem paths. `Service.Lint` currently
builds an adapter-neutral task snapshot, but `dependencyLintIssues` indexes findings by
`GraphProblem.Path`, skips pathless problems, and joins findings back through `task.Path`. A remote
or in-memory adapter can therefore report a broken graph while ordinary lint silently drops some of
the defects.

## Scope

- Introduce a taskflow-owned record reference that can attribute a graph diagnostic to the exact
  readable task record without requiring a local path.
- Preserve correct attribution for duplicate task IDs; do not replace the path map with a naive
  ID-only map that merges distinct records.
- Move `<id>-<slug>.md` identity recovery out of core and into the filesystem adapter. Core must not
  parse an opaque location to manufacture identity.
- Keep local paths as optional repair context and retain current graph-health and lint-severity
  behavior.

## Acceptance criteria

- [ ] Pathless task records receive every applicable dependency and lifecycle-consistency lint
      finding.
- [ ] Duplicate canonical IDs remain attributable to each conflicting record without using an ID as
      a falsely unique map key.
- [ ] An explicit record identity wins over a contradictory location, and core never parses the
      location for identity.
- [ ] Filesystem-backed lint output remains equally actionable and byte-stable except for deliberate
      machine-contract additions.
- [ ] Focused tests cover pathless records, duplicate IDs, cycles, legacy declarations, and an
      unreadable record carrying identity but no location.

## Out of scope

- Changing graph validity, repair authorization, or lint severity policy.
- Migrating non-task list diagnostics; that is tracked separately.
- Introducing a database or HTTP adapter.

## Related

- Epic [21-code-quality-architecture-hardening](../epics/21-code-quality-architecture-hardening.md)
- Thread [Make planning data access adapter neutral](../threads/6gcwd78p9r04-make-planning-data-access-adapter-neutral.md)
- [Adapter-neutral repository lint diagnostics](6g5vm4efjcdv-make-repository-lint-load-diagnostics-adapter-neutral.md)
- [Portable Board/status diagnostics](6g6jqqcdehne-preserve-portable-load-diagnostics-in-board-and-status.md)
