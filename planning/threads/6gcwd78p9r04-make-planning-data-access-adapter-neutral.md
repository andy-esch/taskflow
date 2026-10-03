---
schema: 1
id: 6gcwd78p9r04
status: in-progress
description: Eliminate path-shaped application contracts and primary-to-secondary planning-data bypasses.
goal: Core use cases consume identity-bearing semantic ports; local paths are optional capabilities, and primary adapters cannot bypass the application boundary.
created: "2026-09-23"
tags: [architecture, ports, adapters, hardening]
tasks: [6g5vm4efjcdv, 6g6jqqcdehne, 6gcwcf77tvgq, 6gcwcf7gjayh, 6gcwcf7rgxef, 6gcwcf80v8hg, 6gcwcf88z57p, 6gcwcf8gzn50, 6gcwcf8rxe72, 6gdx7mcqm371, 6gdx7mcqq67d, 6gdx7mcrq8s8, 6ge1bacd3bd2, 6ge7qn9ptaxv]
updated_at: "2026-10-03"
started_at: "2026-09-23"
---

# Thread: Make planning data access adapter neutral

**Goal.** Core use cases consume identity-bearing semantic ports; local paths are optional capabilities, and primary adapters cannot bypass the application boundary.

## Context

Taskflow already enforces inward package dependencies and has proven adapter-neutral patterns for
Thread reads and task-graph reads. The Thread's first evidence stage is now complete: repository
lint diagnostics, graph-lint attribution, Board/status diagnostics, and audit finding queries all
preserve stable identity without requiring a filesystem path. Those slices also establish that an
opaque source location and a local repair path may coexist, core never reconstructs identity from
either value, guarded source revisions remain private, and public projections use deterministic
ordering without rescanning records.

The loaded-record and optional-local-capability design is complete. Ordinary task, epic, audit,
research, and finding reads now preserve portable source identity and diagnostics, and split
planning-data capabilities must agree on one source-set witness at construction. Local mutation
receipts and ordinary-create recovery no longer depend on domain paths. TUI identity/refresh and
readable occurrence locations have shipped, and semantic entity values no longer carry local paths,
filename identities, or guarded revisions. CLI repair, body-link checking, and completion are the
current implementation slice; executable enforcement of the composition boundary remains next.

This Thread closes those seams in evidence-driven stages:

1. Establish and stress-test the diagnostic/read precedents through lint, graph attribution,
   Board/status, and audit findings. **Completed.**
2. Design the shared loaded-record and optional-local-capability contract. **Completed.**
3. Promote portable diagnostics and migrate ordinary list/show, Summary, lint, and wire
   projections onto loaded-record evidence. **Completed.**
4. Bind split capabilities to one source set, preserve local create/rename outcome evidence outside
   domain records, and report ordinary-create guard-release failures without suggesting an unsafe
   retry. **Completed.**
5. Preserve readable occurrence locations and move TUI identity/refresh onto portable loaded-record
   evidence while retaining fail-closed duplicate-ID behavior. **Completed.**
6. After those prerequisites, move local paths, filename-derived identity, and guarded revisions
   out of semantic entity values; finish the partial Thread precedent and wire explicit local
   capabilities. **Completed.**
7. Route CLI planning-data operations through those settled application ports. **In progress:**
   named repair/link/completion use cases and portable probes implemented; both reviews reconciled,
   including resolver-compatible completion fixes. Ready for merge.
8. Isolate composition wiring and make the intended controller boundary executable through the
   standard lint suite.

Markdown-first and git-native behavior remain product constraints. The goal is not to hide Markdown
or replace storage; it is to stop application semantics from requiring a filesystem path when stable
identity, an optional source location, or an explicit local capability is the honest contract.

## Completion signal

- Core semantic read and diagnostic ports do not expose `domain.FileProblem` or require local paths.
- Local path navigation remains available through explicit optional capabilities.
- Independently composed readers, path resolvers, and mutators cannot silently address different
  planning corpora, and local dry-run/recovery outcomes no longer depend on domain paths.
- Primary adapters cannot open planning persistence directly outside the documented composition
  boundary, and a dependency rule catches regressions.
- Filesystem behavior, guarded mutation evidence, TUI editing, and public machine compatibility stay
  intact throughout the migration.
