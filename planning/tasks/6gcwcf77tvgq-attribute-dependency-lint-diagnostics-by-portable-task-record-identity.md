---
schema: 1
id: 6gcwcf77tvgq
status: completed
epic: 21-code-quality-architecture-hardening
description: Stop semantic graph lint findings from disappearing or colliding when a task source has no filesystem path.
effort: 1-2 days
tier: 2
priority: medium
autonomy_level: 3
tags: [architecture, lint, diagnostics, ports]
created: "2026-09-23"
depends_on: [6g5vm4efjcdv]
updated_at: "2026-09-24"
started_at: "2026-09-23"
completed_at: "2026-09-24"
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

- [x] Pathless task records receive every applicable dependency and lifecycle-consistency lint
      finding.
- [x] Duplicate canonical IDs remain attributable to each conflicting record without using an ID as
      a falsely unique map key.
- [x] An explicit record identity wins over a contradictory location, and core never parses the
      location for identity.
- [x] Filesystem-backed lint output remains equally actionable and byte-stable except for deliberate
      machine-contract additions.
- [x] Focused tests cover pathless records, duplicate IDs, cycles, legacy declarations, and an
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

## Implementation progress (2026-09-23)

Dependency lint now correlates findings through an opaque, snapshot-local readable-record reference instead of `Task.Path`. The analyzer attaches that reference to record-owned structural problems and legacy diagnostics, and preserves the representative record for cycle and lifecycle-consistency findings. The public graph query remains unchanged: the ephemeral correlation handle is cleared before problems or legacy diagnostics leave the graph.

Core no longer parses `<id>-<slug>.md` locations to manufacture unreadable-task identity. The filesystem scanner remains the owner of that naming convention and supplies recovered identity before adapting local diagnostics into `TaskGraphRead`. Focused coverage proves pathless dependency, lifecycle, cycle, and legacy findings; duplicate IDs with a shared opaque location; explicit identity precedence; and pathless unreadable-record attribution.

Validation: `go test -race ./...`, `golangci-lint run ./...`, repository `lint --json`, and `git diff --check` are clean.

## Adversarial review closeout (2026-09-24)

Antigravity reconstructed the record-attribution flow, exercised pathless and duplicate-record matrices, compared filesystem lint output with `main`, and mutation-tested eight critical seams. The only finding was a missing regression assertion for clearing the private record reference from public legacy diagnostics. The test now proves both halves of that boundary: internal diagnostics retain attribution, while `LegacyDiagnostics()` suppresses it; removing the clearing line fails the focused test. Finding L1 is fixed and the audit is closed.

Post-review validation: the focused mutation-killing test passes; `go test -race -vet=off ./...`, repository lint, audit lint, and `git diff --check` pass. The shared worktree’s concurrent Go 1.26 update currently makes the independent golangci/vet pass stop on three pre-existing `%q` format assertions outside this task; those parallel files were left untouched.
