---
schema: 1
id: 6gcwcf88z57p
status: ready-to-start
epic: 21-code-quality-architecture-hardening
description: Apply the ThreadPathSource pattern to task, epic, audit, and research so portable adapters never invent paths.
effort: 2-3 days
tier: 2
priority: medium
autonomy_level: 2
tags: [architecture, ports, entities, filesystem]
created: "2026-09-23"
depends_on: [6gcwcf7rgxef]
updated_at: "2026-09-23"
---

# Split local path capabilities from semantic entity reads

## Objective

Apply the proven Thread read/path separation to task, epic, audit, and research. Semantic adapters
must be able to provide portable records without manufacturing filesystem paths, while local CLI/TUI
navigation must retain parse-free access to malformed source files when a filesystem adapter supports
it.

## Scope

- Remove `ResolveTaskPath`, `ResolveEpicPath`, `ResolveAuditPath`, and `ResolveResearchPath` from the
  aggregate semantic `Store` capabilities.
- Introduce narrow optional local-path capabilities with explicit composition and typed-nil handling.
- Move adapter metadata selected by the design task out of domain entities and into read/source
  envelopes without weakening guarded CAS evidence.
- Update CLI path/info commands and TUI edit/open behavior to request the optional capability and
  explain when it is unavailable.

## Acceptance criteria

- [ ] A pathless adapter implements semantic reads without stubbing any `Resolve*Path` method or
      populating fake entity paths.
- [ ] Local `<entity> path` commands still resolve malformed id-led documents without parsing their
      frontmatter.
- [ ] TUI navigation and editing consume explicit optional source capabilities rather than entity
      `Path` fields.
- [ ] Split-source construction cannot silently pair semantic reads with an unrelated path resolver.
- [ ] Guarded mutation planners and public wire projections remain free of opaque revision tokens.

## Out of scope

- Replacing Markdown storage or implementing a remote adapter.
- Making filesystem path commands portable; their local nature should remain explicit.
- Combining entity-specific stores into one generic interface.

## Related

- Epic [21-code-quality-architecture-hardening](../epics/21-code-quality-architecture-hardening.md)
- Thread [Make planning data access adapter neutral](../threads/6gcwd78p9r04-make-planning-data-access-adapter-neutral.md)
- [Thread path/read split precedent](6g5ryqqx5ab7-split-local-thread-path-resolution-from-portable-thread-reads.md)
- [Portable entity-read design](6gcwcf7rgxef-design-portable-entity-reads-and-optional-local-source-capabilities.md)
