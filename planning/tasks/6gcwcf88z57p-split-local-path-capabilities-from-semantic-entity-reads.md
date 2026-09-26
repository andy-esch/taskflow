---
schema: 1
id: 6gcwcf88z57p
status: ready-to-start
epic: 21-code-quality-architecture-hardening
description: Remove local path, filename identity, and revision evidence from semantic entity values and expose local navigation explicitly.
effort: 3-5 days
tier: 2
priority: medium
autonomy_level: 2
tags: [architecture, ports, entities, filesystem]
created: "2026-09-23"
depends_on: [6gcwcf80v8hg, 6gdx7mcqm371, 6gdx7mcqq67d, 6gdx7mcrq8s8]
updated_at: "2026-09-26"
---

# Split local source capabilities from semantic entity reads

## Objective

Complete the separation that the Thread path work began, then apply it uniformly. Semantic adapters
must provide canonical record identity without manufacturing filesystem paths, filename-derived
fields, or ordinary-read revision tokens. Local CLI/TUI navigation must retain parse-free access to
malformed source files when a filesystem adapter supports it, and guarded task/Thread snapshots must
retain their stronger version evidence outside the domain values.

## Scope

- Remove `ResolveTaskPath`, `ResolveEpicPath`, `ResolveAuditPath`, and `ResolveResearchPath` from the
  aggregate semantic `Store` capabilities.
- Introduce narrow entity-specific optional local-path capabilities with the proven Thread
  detach/explicit-override composition rules, typed-nil handling, and the shared source-set identity
  established by the dedicated composition task.
- Move `Path` and `FilenameID` out of task, epic, audit, research, and Thread domain values and into
  the read/source envelopes selected by the design task. Migrate navigation and lint to the
  adapter-supplied canonical source ID rather than `CanonicalID()` fallbacks.
- Move task and Thread `SourceVersion` values into versioned guarded-record/problem wrappers without
  weakening complete-snapshot CAS or leaking tokens into planners and projections.
- Finish the Thread precedent by removing its remaining readable-record `Path`, `FilenameID`, and
  `SourceVersion` leakage; keep the existing `ThreadPathSource` behavior compatible.
- Update CLI path/info commands and TUI edit/open behavior to request the optional capability and
  explain when it is unavailable.
- Replace mutation-returned domain paths deliberately: resolve after success where safe, or carry
  the operation-specific local outcome metadata established by the sequenced receipt task. Do not
  reopen dry-run or durable-prefix output design during this final removal slice.

## Acceptance criteria

- [ ] A pathless adapter implements semantic reads without stubbing any `Resolve*Path` method or
      populating fake entity paths.
- [ ] Semantic task, epic, audit, research, and Thread domain values contain no local path,
      filename-derived identity, or guarded revision token.
- [ ] Local `<entity> path` commands still resolve malformed id-led documents without parsing their
      frontmatter.
- [ ] TUI navigation and editing consume explicit optional source capabilities rather than entity
      `Path` fields.
- [ ] Split-source construction cannot silently pair semantic reads with an unrelated path resolver.
- [ ] The final domain-field removal lands only after ordinary/show/wire projections, TUI identity,
      local mutation receipts, and source-set validation no longer consume those fields.
- [ ] Guarded mutation planners and public wire projections remain free of opaque revision tokens.
- [ ] Task/Thread whole-snapshot comparisons still fail closed for missing or changed readable and
      unreadable source revisions, and duplicate-ID lint remains attributable without path identity.

## Out of scope

- Replacing Markdown storage or implementing a remote adapter.
- Making filesystem path commands portable; their local nature should remain explicit.
- Combining entity-specific stores into one generic interface.

## Related

- Epic [21-code-quality-architecture-hardening](../epics/21-code-quality-architecture-hardening.md)
- Thread [Make planning data access adapter neutral](../threads/6gcwd78p9r04-make-planning-data-access-adapter-neutral.md)
- [Thread path/read split precedent](6g5ryqqx5ab7-split-local-thread-path-resolution-from-portable-thread-reads.md)
- [Portable entity-read design](6gcwcf7rgxef-design-portable-entity-reads-and-optional-local-source-capabilities.md)
