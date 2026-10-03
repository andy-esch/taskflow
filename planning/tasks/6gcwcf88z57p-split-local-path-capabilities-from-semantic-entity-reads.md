---
schema: 1
id: 6gcwcf88z57p
status: in-progress
epic: 21-code-quality-architecture-hardening
description: Remove local path, filename identity, and revision evidence from semantic entity values and expose local navigation explicitly.
effort: 3-5 days
tier: 2
priority: medium
autonomy_level: 2
tags: [architecture, ports, entities, filesystem]
created: "2026-09-23"
depends_on: [6gcwcf80v8hg, 6gdx7mcqm371, 6gdx7mcqq67d, 6gdx7mcrq8s8, 6ge1bacd3bd2]
updated_at: "2026-10-02"
started_at: "2026-10-01"
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
  detach/explicit-override composition rules, typed-nil handling, and the `SourceSetProvider`
  witness and checked constructor established by the dedicated composition task. Add every new
  entity path port to that one validation matrix; do not introduce per-entity source-set checks.
- Move `Path` and `FilenameID` out of task, epic, audit, research, and Thread domain values and into
  the read/source envelopes selected by the design task. Migrate navigation and lint to the
  adapter-supplied canonical source ID rather than `CanonicalID()` fallbacks.
- Move task and Thread `SourceVersion` values into versioned guarded-record/problem wrappers without
  weakening complete-snapshot CAS or leaking tokens into planners and projections.
- Supply graph-repair `LocalPath` as an explicit adapter-owned local capability, never infer it from
  a semantic entity path or decide locality with a URI-string heuristic. Keep opaque readable
  locations as diagnostic/stale-context evidence only; core must not advertise a local repair for
  an adapter that has not supplied that capability.
- Finish the Thread precedent by removing its remaining readable-record `Path`, `FilenameID`, and
  `SourceVersion` leakage; keep the existing `ThreadPathSource` behavior compatible.
- Update CLI path/info commands and TUI edit/open behavior to request the optional capability and
  explain when it is unavailable.
- Replace mutation-returned domain paths deliberately: resolve after success where safe, or carry
  the operation-specific local outcome metadata established by the sequenced receipt task. Do not
  reopen dry-run or durable-prefix output design during this final removal slice.

## Acceptance criteria

- [x] A pathless adapter implements semantic reads without stubbing any `Resolve*Path` method or
      populating fake entity paths.
- [ ] Semantic task, epic, audit, research, and Thread domain values contain no local path,
      filename-derived identity, or guarded revision token.
- [x] Local `<entity> path` commands still resolve malformed id-led documents without parsing their
      frontmatter.
- [x] TUI navigation and editing consume explicit optional source capabilities rather than entity
      `Path` fields.
- [x] Split-source construction cannot silently pair semantic reads with an unrelated path resolver.
- [ ] The final domain-field removal lands only after ordinary/show/wire projections, TUI identity,
      local mutation receipts, and source-set validation no longer consume those fields.
- [x] Guarded mutation planners and public wire projections remain free of opaque revision tokens.
- [x] Pathless and non-filesystem adapters cannot accidentally gain a local graph-repair target
      from a domain `Path` value such as `urn:…` or another URI-like key.
- [ ] Task/Thread whole-snapshot comparisons still fail closed for missing or changed readable and
      unreadable source revisions, and duplicate-ID lint remains attributable without path identity.

## Out of scope

- Replacing Markdown storage or implementing a remote adapter.
- Making filesystem path commands portable; their local nature should remain explicit.
- Combining entity-specific stores into one generic interface.

## Progress

- The Task filename-identity removal review exposed a repair reload check that rebuilt source
  identity from the declared ID. It now uses an explicit filesystem source record and verifies the
  source as well as the repaired fields. Regression tests cover missing and drifting declarations
  in dry-run and committed repairs, retaining residual identity diagnostics and unrelated content.
  Both checkpoint reviews are reconciled; Task local-path and Thread mutation migration remain.
- Checkpoint review found that the compatibility `NewTaskGraph`/`TaskGraphRead.Tasks` conversions
  promoted `Task.Path` to a local repair handle. They are now read-only; local repair tests and
  filesystem adapters supply `VersionedRecord.LocalPath` explicitly. Hostile URI and path-shaped
  compatibility inputs remain non-repairable. The TUI edit/yank path now confirms the selected
  stable ID after its initial asynchronous lookup, following a rename to its new path or reporting
  that the ID disappeared. This narrows the action window but cannot make an external editor's
  pathname atomic against a subsequent rename.
- Thread lint now uses readable source IDs and opaque locations; `status --all` projects its
  combined working set from the same source record as per-space summaries. Regression tests cover
  those projections, path-shaped opaque repair locations, and coherent-list stale TUI results.

- Optional task, epic, audit, and research path ports are split from the aggregate semantic
  `Store`. Complete local adapters still supply them; explicit read replacements detach implicit
  paths, and source-set validation rejects foreign resolvers before use.
- `task info` and `audit info` now resolve local paths through those ports by canonical source ID.
  Pathless reads retain the semantic metadata and report the unavailable path honestly.
- Ordinary location projections now use an adapter-supplied `LocationIsPath` presentation hint
  instead of comparing a location to a domain `Path`; the hint never authorizes opening a file.
  Local output remains unchanged while opaque locations remain visible even if path-shaped.
- The filesystem Thread scan now attaches successful-record source revisions directly to guarded
  read wrappers (including the bulk-apply scan). Task and Thread domain types no longer contain
  source-revision fields; guarded wrappers/problems are the sole owners of that evidence.
- The filesystem task scan now carries revisions and local repair paths in guarded record wrappers.
  Ordinary task reads have no revision token; graph CAS detects even body-only source edits. Graph
  diagnostics use explicit source IDs/paths, repair locality no longer uses a URI heuristic, and
  mutation materialization checks the graph's source handle rather than `Task.Path`.
- Epic, audit, and research domain values no longer carry local paths; audit and research no longer
  carry filename IDs. Their filesystem reads retain canonical source identity and location in
  loaded records, including lint, selected reads, and open-audit summaries. The board and in-progress
  summary now retain loaded task records instead of writing source identity back into task values.
  Task lint compares frontmatter IDs with adapter source IDs for both active and archived records.
- Task graph, lint, and selected filesystem reads now attach source IDs and locations at the scan
  boundary. Regression tests cover frontmatter ID drift and parse-free path resolution of malformed
  task, epic, audit, and research documents.
- Thread list/show/graph read projections now use loaded source identity for ID drift and duplicate
  detection; opaque source locations are not promoted to local diagnostic paths. Guarded Thread
  mutation planning still needs this same source-aware migration before the transitional fields can
  leave `domain.Thread`.
- TUI editor/copy-path actions request the optional local path asynchronously by canonical ID,
  confirm that ID before acting, and reject results after a selection, list-generation, or workspace
  change. Detail hyperlinks use a separately resolved path; pathless semantic records cannot supply
  editor paths by accident.
- Task filename identity now lives only in loaded source records, not `domain.Task`; graph tests
  supply distinct source IDs explicitly when frontmatter IDs drift or are missing. The remaining
  domain-field removal is `Task.Path` and Thread `Path`/`FilenameID`, concentrated in compatibility
  constructors, Thread mutation planning/materialization, and their tests. Those callers still need
  explicit source envelopes or local receipts.

## Related

- Epic [21-code-quality-architecture-hardening](../epics/21-code-quality-architecture-hardening.md)
- Thread [Make planning data access adapter neutral](../threads/6gcwd78p9r04-make-planning-data-access-adapter-neutral.md)
- [Thread path/read split precedent](6g5ryqqx5ab7-split-local-thread-path-resolution-from-portable-thread-reads.md)
- [Portable entity-read design](6gcwcf7rgxef-design-portable-entity-reads-and-optional-local-source-capabilities.md)
