---
schema: 1
id: 6gcwd78p9r04
status: in-progress
description: Eliminate path-shaped application contracts and primary-to-secondary planning-data bypasses.
goal: Core use cases consume identity-bearing semantic ports; local paths are optional capabilities, and primary adapters cannot bypass the application boundary.
created: "2026-09-23"
tags: [architecture, ports, adapters, hardening]
tasks: [6g5vm4efjcdv, 6g6jqqcdehne, 6gcwcf77tvgq, 6gcwcf7gjayh, 6gcwcf7rgxef, 6gcwcf80v8hg, 6gcwcf88z57p, 6gcwcf8gzn50, 6gcwcf8rxe72]
updated_at: "2026-09-23"
started_at: "2026-09-23"
---

# Thread: Make planning data access adapter neutral

**Goal.** Core use cases consume identity-bearing semantic ports; local paths are optional capabilities, and primary adapters cannot bypass the application boundary.

## Context

Taskflow already enforces inward package dependencies and has proven adapter-neutral patterns for
Thread reads, task-graph reads, and repository lint load diagnostics. The remaining application
surface is uneven: dashboards and ordinary lists collapse identity into `FileProblem`, semantic
lint attribution and finding queries use paths as record handles, task/epic/audit/research reads mix
portable data with local path navigation, and some CLI controllers reach directly into the
filesystem adapter.

This Thread closes those seams in evidence-driven stages:

1. Finish the current lint diagnostic foundation, then independently harden graph-lint attribution,
   Board/status diagnostics, and audit finding queries.
2. Use those concrete ports to design one coherent loaded-record and optional-local-capability
   contract instead of copying Thread-specific shapes mechanically.
3. Migrate ordinary list diagnostics, local path resolution, and CLI planning-data operations in
   parallel behind that design.
4. Isolate composition wiring and make the intended controller boundary executable through the
   standard lint suite.

Markdown-first and git-native behavior remain product constraints. The goal is not to hide Markdown
or replace storage; it is to stop application semantics from requiring a filesystem path when stable
identity, an optional source location, or an explicit local capability is the honest contract.

## Completion signal

- Core semantic read and diagnostic ports do not expose `domain.FileProblem` or require local paths.
- Local path navigation remains available through explicit optional capabilities.
- Primary adapters cannot open planning persistence directly outside the documented composition
  boundary, and a dependency rule catches regressions.
- Filesystem behavior, guarded mutation evidence, TUI editing, and public machine compatibility stay
  intact throughout the migration.
