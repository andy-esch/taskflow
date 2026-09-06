---
schema: 1
id: 6g7fhfpmy032
status: next-up
epic: 30-threads-and-task-dependency-graphs
description: Ship the compatibility and repair hardening as another explicit preview checkpoint before reconsidering graduation.
effort: 1 day
tier: 2
priority: medium
autonomy_level: 4
tags: [threads, release, compatibility, dogfood]
created: "2026-09-06"
depends_on: [6g7ddeyp773z, 6g7j2ebatzyt]
updated_at: "2026-09-06"
---

# Cut v0.20.0 as a compatibility-hardened Threads preview

## Objective

Ship the historical-compatibility and guarded-repair foundation as a third deliberately preview-
labelled Thread release. Let real installed binaries exercise the hardened contracts before an
explicit graduation decision; passing the graduation gates makes graduation possible, not automatic.

## Scope

- Start from clean `main` after `pin-thread-document-and-plan-backward-compatibility` lands and
  record the exact candidate commit and binary version.
- Run the compatibility fixtures, focused hostile graph/repair/mutation tests, full race tests,
  lint, generated docs/schema checks, planning lint, and `just release-snapshot`.
- Repeat a bounded CLI/TUI dogfood pass with the production Thread, stable task navigation, live
  refresh, a retained apply plan, and graph diagnosis/repair. Record any behavioral or compatibility
  finding as a named task rather than silently accepting it for graduation.
- Curate concise release notes around compatibility-managed preview data, resumable apply, guarded
  recovery, shared core projections, and the current CLI/TUI capabilities.
- Keep the README preview notice in the tagged candidate and verify the immutable tag, release
  workflow, archives, checksums, and installed binary after publication.

## Acceptance criteria

- [ ] `pin-thread-document-and-plan-backward-compatibility` is complete and the v0.18.0/v0.19.0
      compatibility fixtures pass from the recorded clean candidate.
- [ ] Focused hostile tests, full `go test -race ./...`, lint, generated docs/schema checks,
      planning lint, and `just release-snapshot` pass on that candidate.
- [ ] A fresh candidate binary completes the bounded CLI/TUI dogfood pass; commands and outcomes,
      including any filed follow-ups, are recorded here.
- [ ] README and release notes explicitly retain the Threads preview classification and explain
      the compatibility/recovery value of this release without implying graduation.
- [ ] The published v0.20.0 tag, release workflow, archives, checksums, and installed binary all
      identify the recorded candidate.
- [ ] The graduation task remains a separate explicit decision after this checkpoint; the spatial
      graph experiment is neither silently promoted nor made a release gate.

## Out of scope

- Removing the Threads preview notice or declaring all graduation evidence current.
- Requiring the two-dimensional graph experiment, frontier ranking, or portable convenience-view
  diagnostics merely to cut this checkpoint.
- Folding an experimental spatial renderer into core Thread semantics.

## Related

- Epic [30-threads-and-task-dependency-graphs](../epics/30-threads-and-task-dependency-graphs.md)
- Contract [Threads compatibility and preview graduation](../../docs/THREADS_COMPATIBILITY.md)
- Compatibility fixtures
  [`6g7ddeyp773z`](6g7ddeyp773z-pin-thread-document-and-plan-backward-compatibility.md)
- Graduation decision [`6g7ddfhh2jc2`](6g7ddfhh2jc2-graduate-threads-from-preview.md)
- Thread [Complete production Threads](../threads/6g503c6pfqeb-complete-production-threads.md)
