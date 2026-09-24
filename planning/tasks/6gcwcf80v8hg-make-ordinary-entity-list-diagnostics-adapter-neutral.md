---
schema: 1
id: 6gcwcf80v8hg
status: ready-to-start
epic: 21-code-quality-architecture-hardening
description: Preserve identity and optional locations through task, epic, audit, research, and finding list projections.
effort: 1-2 days
tier: 2
priority: medium
autonomy_level: 3
tags: [architecture, diagnostics, ports, json]
created: "2026-09-23"
updated_at: "2026-09-23"
depends_on: [6gcwcf7rgxef]
---

# Make ordinary entity list diagnostics adapter neutral

## Objective

Finish the resilient-read diagnostic migration for ordinary task, epic, audit, research, and finding
lists. Their core and wire contracts still expose `domain.FileProblem{path,message}`, so a pathless
adapter loses entity kind and stable identity even though lint, Threads, and task-graph reads can
already preserve it.

## Scope

- Apply the portable diagnostic contract selected by the preceding design task to ordinary list
  services and their CLI/wire projections.
- Preserve entity kind, stable ID/slug when recoverable, optional opaque location, and message.
- Keep the historical JSON `path` field compatible for local sources while never labelling a URI as
  a filesystem path.
- Update task, epic, audit, research, and finding lists together so adapters do not face two
  competing resilient-read vocabularies.

## Acceptance criteria

- [ ] Every ordinary entity-list port and service result is free of `domain.FileProblem`.
- [ ] Pathless failures retain kind and identity through core, human output, JSON, and generated
      schema.
- [ ] Local failures remain actionable and retain current partial-result and exit behavior.
- [ ] Explicit identity wins over misleading location on every entity kind.
- [ ] Machine-contract fixtures document the additive compatibility mapping and reject schema drift.

## Out of scope

- Changing which malformed documents are accepted or how lint repairs them.
- Board/status diagnostics, which have a dedicated predecessor task.
- Moving local path navigation out of semantic entity ports.

## Related

- Epic [21-code-quality-architecture-hardening](../epics/21-code-quality-architecture-hardening.md)
- Thread [Make planning data access adapter neutral](../threads/6gcwd78p9r04-make-planning-data-access-adapter-neutral.md)
- [Portable Board/status diagnostics](6g6jqqcdehne-preserve-portable-load-diagnostics-in-board-and-status.md)
- [Stable lint diagnostic kinds](6gcqz5b0j0dg-give-lint-diagnostics-stable-kinds-and-repair-applicability.md)
