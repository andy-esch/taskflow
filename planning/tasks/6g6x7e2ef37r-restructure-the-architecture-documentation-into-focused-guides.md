---
schema: 1
id: 6g6x7e2ef37r
status: ready-to-start
epic: 21-code-quality-architecture-hardening
description: Revise the oversized architecture document into a smaller navigable structure with clear ownership and durable cross-links.
effort: 1-2 days
tier: 3
priority: medium
autonomy_level: 3
tags: [docs, architecture, maintainability]
created: "2026-09-04"
depends_on: [6g9cz5saenme]
updated_at: "2026-10-11"
---

# Restructure the architecture documentation into focused guides

## Objective

Make the architecture documentation easier to navigate and maintain by replacing the oversized
single-file structure with a concise entry point and smaller, clearly owned guides organized around
stable architectural concerns. Use the source-of-truth matrix from the documentation-ownership task
to decide what belongs in the index, a focused guide, generated output, package documentation, an
ADR, or planning history.

## Acceptance criteria

- [ ] Define a simple target information architecture before moving content.
- [ ] Break the current architecture document into focused guides with clear scopes and useful
  navigation from a central architecture entry point.
- [ ] Preserve meaningful architectural decisions and context while removing obsolete or duplicate
  prose deliberately rather than accidentally.
- [ ] Repair repository links and contributor guidance that reference the current document layout.
- [ ] Documentation generation/checks, link validation where available, and planning lint pass.

## Out of scope

- Changing product or code architecture merely to match the new document layout.
- Re-litigating accepted ADR decisions unless the restructuring exposes a concrete contradiction
  that should be tracked separately.
- A comprehensive rewrite of unrelated user or CLI documentation.

## Research handoff (2026-10-11)

The [ownership research](../research/6gjggjw01raw-documentation-ownership-and-portable-agent-routing.md)
measured the current reference at 1,087 lines / 11,466 words (baseline 09974fb), larger than the
Thread's original estimate. Inventory its dated `Status (2026-08-25)` and review-history sections
separately from current contracts; do not leave historical feature snapshots presented as current
guidance. Preserve accepted decision history in ADRs and useful review provenance in planning.
Keep the planning-data checklist's owner and old incoming links explicit during extraction.

## Related

- Epic [21-code-quality-architecture-hardening](../epics/21-code-quality-architecture-hardening.md)
- Thread [Make documentation layered, executable, and agent-navigable](../threads/6g9czp7g9pt3-make-documentation-layered-executable-and-agent-navigable.md)
- Follows [Establish documentation ownership and agent routing](6g9cz5saenme-establish-documentation-ownership-and-agent-routing.md)

## Archived architecture snapshot

The following is the old `Status (2026-08-25)` section as found at baseline 4fdf6fcc, not a verified snapshot of that date; later edits had accumulated beneath its dated heading. Preserved as history rather than current implementation guidance.

```markdown
## Status (2026-08-25)

The layered shape now has two substantial primary consumers and several reusable
application seams; it is no longer architecture held in reserve for a hypothetical UI:

- The CLI exposes full task lifecycle and body mutation, epic/audit/research authoring,
  schema and wire discovery, configuration lifecycle, planning-space registry verbs,
  global `--space` selection, and the cross-space `status --all` overview.
- The Bubble Tea app provides its dashboard plus task/epic/audit/research tabs,
  a cross-space atlas with reversible entry-point navigation, filtering/sorting/find,
  structured navigation, task and audit lifecycle actions, task field editing,
  active-space filesystem reloads, and the embedded configuration editor. Reads and
  writes remain asynchronous commands over core services; production TUI code does not
  open planning/config stores.
- The configuration and cross-space features have dedicated core services and composite
  filesystem adapters. `SpaceRegistryService` owns catalog/selection/mutation policy and
  `SpaceOverviewService` owns cross-space summaries; Cobra is one primary adapter over
  both rather than the owner of either use case.
- Writes use atomic replacement/exclusive create, a repo-wide lock, content-version CAS,
  bounded retry for agent mutations, and parse-before-accept editor loops. Machine output
  is a versioned `internal/wire` contract with generated JSON Schema and golden coverage.
- The planning model's closed vocabularies now have write verbs rather than only linters.
  An acceptance criterion carries a state and a reason (`task ac`), an audit finding
  carries a status, resolution paragraph, and optional versioned candidate projection
  (`audit finding`), and the words the two
  share are declared once in `domain/resolution.go`. Two invariants are enforced at write
  time rather than reported after: `task complete` refuses a task with an unexplained
  unmet criterion, mirroring `audit close` requiring terminal findings and complete parsing.

The atlas decision has now activated the narrow reusable workspace-opening boundary:
`core.WorkspaceService` and `internal/workspacestore` can open an explicit local entry
point for a primary adapter without exposing discovery or concrete persistence. The
broader `doctor`/fix promotion remains deferred because the atlas does not consume it.
Thread `context.Context` when an HTTP request path exists; reshape findings pagination
when a web findings view is scoped. Until those triggers occur, package movement or
speculative interfaces are out of scope.
```
