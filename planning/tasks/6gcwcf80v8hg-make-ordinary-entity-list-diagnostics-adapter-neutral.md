---
schema: 1
id: 6gcwcf80v8hg
status: ready-to-start
epic: 21-code-quality-architecture-hardening
description: Preserve identity and optional locations through task, epic, audit, research, and finding list projections.
effort: 3-5 days
tier: 2
priority: medium
autonomy_level: 3
tags: [architecture, diagnostics, ports, json]
created: "2026-09-23"
updated_at: "2026-09-26"
depends_on: [6gcwcf7rgxef]
---

# Make ordinary entity list diagnostics adapter neutral

## Objective

Finish the resilient-read diagnostic migration for ordinary task, epic, audit, research, and finding
list/show projections. Their core and wire contracts still expose `domain.FileProblem{path,message}`
or bare domain values, so a pathless adapter loses entity kind and stable identity even though lint,
Threads, and task-graph reads can already preserve it.

## Scope

- Promote `LintLoadProblem` to the general core `LoadProblem` vocabulary selected by the design
  task, with explicit `Location` and `LocalPath` rather than a path marker or a second wrapper.
- Introduce the shared immutable `RecordSource` / `LoadedRecord` value building blocks and
  entity-specific task, epic, audit, and research read snapshots. Do not introduce a generic
  entity repository or kind-switched read method.
- Make the filesystem adapter produce one authoritative internal loaded/versioned scan per entity
  request and project existing graph/list/show/lint shapes from it during migration. Do not add an
  ordinary task scan beside `TaskGraphRead` or hide rereads behind compatibility adapters.
- Apply those snapshots to ordinary list services and their CLI/wire projections, including the
  transitional dashboard `SummaryStore` seam.
- Migrate single-document Get/Show boundaries and wire converters to loaded source identity before
  later tasks remove `FilenameID` or `Path`; keep any temporary bare-domain projection explicit and
  local to compatibility code.
- Preserve entity kind, stable ID/slug when recoverable, optional opaque location, and message.
- Keep the historical JSON `path` field compatible for local sources while never labelling a URI as
  a filesystem path.
- Update task, epic, audit, research, and finding lists together so adapters do not face two
  competing resilient-read vocabularies.

## Acceptance criteria

- [ ] Every ordinary entity-list port and service result is free of `domain.FileProblem`.
- [ ] Read ports remain entity-specific while their result values share one source and diagnostic
      vocabulary; no universal entity store or kind switch is introduced.
- [ ] Task graph, ordinary list, show, lint, Board, and Summary projections reuse one authoritative
      loaded scan where they request the same snapshot; tests pin adapter call counts.
- [ ] Pathless failures retain kind and identity through core, human output, JSON, and generated
      schema.
- [ ] Local failures remain actionable and retain current partial-result and exit behavior.
- [ ] Explicit identity wins over misleading location on every entity kind.
- [ ] Machine-contract fixtures document the additive compatibility mapping and reject schema drift.
- [ ] `LintSource`, `AuditSnapshotSource`, Board/status, and ordinary list consumers use the promoted
      core diagnostic without a behaviorally identical compatibility wrapper.
- [ ] Single-record show results and wire converters obtain canonical identity from the loaded
      source envelope while public JSON compatibility remains unchanged.

## Out of scope

- Changing which malformed documents are accepted or how lint repairs them.
- Board/status diagnostics, which have a dedicated predecessor task.
- Moving local path navigation out of semantic entity ports.
- Removing domain `Path`, `FilenameID`, or guarded `SourceVersion` fields; the sequenced source/path
  task performs that migration after these snapshots exist.

## Related

- Epic [21-code-quality-architecture-hardening](../epics/21-code-quality-architecture-hardening.md)
- Thread [Make planning data access adapter neutral](../threads/6gcwd78p9r04-make-planning-data-access-adapter-neutral.md)
- [Portable Board/status diagnostics](6g6jqqcdehne-preserve-portable-load-diagnostics-in-board-and-status.md)
- [Stable lint diagnostic kinds](6gcqz5b0j0dg-give-lint-diagnostics-stable-kinds-and-repair-applicability.md)
