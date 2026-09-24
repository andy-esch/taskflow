---
schema: 1
id: 6gcwcf7rgxef
status: ready-to-start
epic: 21-code-quality-architecture-hardening
description: Define shared record, diagnostic, source-location, and local-path contracts for task, epic, audit, and research adapters.
effort: 1-2 days
tier: 2
priority: medium
autonomy_level: 2
tags: [architecture, ports, design, entities]
created: "2026-09-23"
depends_on: [6g6jqqcdehne, 6gcwcf77tvgq, 6gcwcf7gjayh]
updated_at: "2026-09-23"
---

# Design portable entity reads and optional local source capabilities

## Objective

Define the target application boundary for readable and unreadable task, epic, audit, and research
records before mechanically removing their filesystem-shaped contracts. Thread reads and the new
lint port provide useful precedents, but the aggregate `Store` still returns `FileProblem`, embeds
`Path` in domain entities, and combines semantic reads with local path navigation.

## Design questions

- What shared loaded-record envelope carries semantic entity data, canonical identity, optional
  source location, and opaque revision evidence without turning every use case into a generic entity
  API?
- Which operations require source locations, and which require a specifically local filesystem path?
- Should ordinary entity reads share the lint diagnostic type or a more general application-owned
  diagnostic vocabulary?
- How do TUI editing and `<entity> path` obtain optional local capabilities while web, database, and
  service adapters remain pathless?
- Which domain `Path`, `FilenameID`, and `SourceVersion` fields are genuine semantic evidence versus
  adapter metadata that belongs in read envelopes?

## Constraints

- Keep Markdown and stable IDs as product contracts; adapter-neutral does not mean hiding the
  markdown-first model.
- Preserve resilient partial reads, parse-free repair lookup, guarded snapshot comparison, and
  existing public machine compatibility.
- Keep entity-specific use cases explicit; do not introduce a universal `EntityStore` abstraction.
- Reconcile rather than duplicate the broader shared entity-integrity design task.

## Acceptance criteria

- [ ] A decision table classifies every `domain.FileProblem`, entity `Path`, `Resolve*Path`, and
      source-version crossing as semantic, optional local capability, adapter evidence, or debt.
- [ ] The target port and record shapes support filesystem, database, remote-service, and read-only
      adapters without invented paths or ambiguous identity.
- [ ] The design explains duplicate-ID attribution, malformed-record repair, TUI editing, and
      optimistic-concurrency evidence.
- [ ] Compatibility and migration sequencing are explicit for core, wire, CLI, TUI, fakes, and the
      filesystem adapter.
- [ ] The implementation tasks behind this design are amended if the chosen boundary invalidates
      their current assumptions.

## Out of scope

- Implementing the port migration.
- Choosing a database or changing the authoritative storage format.
- Generalizing entity creation or mutation behind one generic API.

## Related

- Epic [21-code-quality-architecture-hardening](../epics/21-code-quality-architecture-hardening.md)
- Thread [Make planning data access adapter neutral](../threads/6gcwd78p9r04-make-planning-data-access-adapter-neutral.md)
- [Shared entity-integrity foundations](6g7wsr8yvt1a-define-the-next-shared-entity-integrity-foundations.md)
- [Thread path/read split precedent](6g5ryqqx5ab7-split-local-thread-path-resolution-from-portable-thread-reads.md)
- [Thread diagnostic precedent](6g5rxq1ravd3-make-thread-read-diagnostics-adapter-neutral.md)
