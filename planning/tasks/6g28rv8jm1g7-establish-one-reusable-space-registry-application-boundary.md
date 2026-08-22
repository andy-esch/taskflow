---
schema: 1
id: 6g28rv8jm1g7
status: completed
epic: 21-code-quality-architecture-hardening
description: Space list, selection, overview, add, and forget currently orchestrate registry behavior through separate adapter paths that can drift as TUI/web consumers arrive.
effort: L
tier: 3
priority: medium
autonomy_level: 3
tags: [architecture, multi-repo, core]
created: "2026-08-21"
started_at: "2026-08-21"
updated_at: "2026-08-22"
completed_at: "2026-08-22"
---

# Establish one reusable space-registry application boundary

## Objective

Make the home space registry one reusable application capability rather than a set of
agreeing-by-discipline CLI paths. Today `space list`, `space add|forget`, global
`--space` selection, configuration diagnosis, and `status --all` reach
`userconfig`/`spacehealth` through different adapters and map the same entry fields more
than once. Establish a consumer-owned core service/port that owns the registry use-case
contracts for every primary adapter while retaining filesystem/TOML/discovery details in
secondary adapters.

## Acceptance criteria

- [x] Record the chosen service/port shape and dependency direction before implementation;
  distinguish the registry/catalog capability from opening a planning tree, rather than
  growing one catch-all filesystem interface.
- [x] `space list`, explicit `--space` resolution, and `status --all` consume one typed
  core entry/group projection; delete the duplicate `spacehealth -> wire` and
  `spacehealth -> core` field mappings.
- [x] Route `space add` and `space forget` through the reusable application service while
  preserving validation-before-write, physical-path deduplication, atomic surgical TOML,
  dry-run behavior, mutation receipts, and existing error classification.
- [x] Keep configuration doctor on the same registry diagnosis vocabulary without making
  repo-scoped `internal/config` import home-scoped registry state.
- [x] A non-Cobra consumer can list, group, resolve, add, and forget entries using core
  values without importing `userconfig`, `spacehealth`, TOML, or filesystem packages.
- [x] Existing human output, JSON fields/order, selection precedence, completion behavior,
  and exit codes remain byte-compatible; focused core/adapter tests plus `just test` and
  `just lint` pass.

## Out of scope

- The reusable `Resolve() -> Workspace` seam, init/doctor/fix promotion, or request context;
  those remain in their existing deferred architecture tasks until atlas/web work begins.
- Building the TUI atlas, adding a served adapter, remote registries, filesystem scanning,
  or changing the `spaces.toml` schema.
- Folding repository configuration and the home registry into one bounded context merely
  because both use TOML.

## Design decision (2026-08-21)

The application boundary is a `SpaceRegistryService` in `internal/core`, backed by a
consumer-owned `SpaceRegistryStore`. The store exposes diagnosed entry points plus the
prepare/add/forget persistence operations; the service owns catalog grouping, explicit
id resolution, default-id and id validation policy, and mutation receipts. Its values are
core-owned `SpaceEntryPoint`, `SpaceGroup`, `SpaceCatalog`, and `SpaceMutation` values, so
CLI, completion, TUI, and a future served adapter do not need home-config, diagnosis, or
TOML types.

Opening the planning tree behind a healthy entry remains a separate `SpaceOverviewStore`
port. `SpaceOverviewService` composes the registry service with that opener for
`status --all`, rather than growing a registry interface that also knows every planning
operation. Configuration diagnosis composes the same catalog vocabulary at the core
service layer; repo-scoped configuration discovery remains independent of home registry
state.

The filesystem adapter owns the one translation from `spacehealth`/`userconfig` into core
entries and classifies storage validation/conflict errors at that boundary. Renderers use
the existing core-to-wire translation. This preserves the wire schema and human ordering
while removing presentation-layer knowledge of registry storage and diagnosis details.

## Completion (2026-08-22)

Added the framework-free `core.SpaceRegistryService` and its consumer-owned
`SpaceRegistryStore`. The service now owns catalog grouping, explicit label resolution,
default-label and validation policy, and add/forget receipts. `spacestore.FS` implements
the port, retains all planning discovery and home-registry persistence details, and owns
the sole `spacehealth`/`userconfig` to core entry translation plus domain error
classification.

The root composition creates one registry service for `space list`, `space add|forget`,
global `--space`/`TSKFLW_SPACE`, label completion, configuration doctor, and
`status --all`. `SpaceOverviewService` now composes that catalog with a separate planning
tree opener, while `ConfigurationService` adds registry problems to repo/link diagnosis
without making `configstore` diagnose home registry entries. CLI renderers receive core
values and use the existing neutral core-to-wire mapper, so the JSON schema and output
order did not change.

Focused core tests cover grouping fallbacks, registry order, selection errors, label
policy, validation-before-mutation, receipts, and adapter error propagation. Focused
filesystem tests cover direct/pointer preparation, diagnosis, add/dedup/forget, dry-run,
and validation/conflict classification. The existing CLI integration suite preserves the
human/JSON/selection/completion/exit contracts. Full race tests, golangci-lint, planning
lint, generated-doc drift checks, and `git diff --check` pass.

## Related

- Epic [21-code-quality-architecture-hardening](../epics/21-code-quality-architecture-hardening.md)
- Epic [29-multi-space-planning-a-home-registry-and-the-atlas](../epics/29-multi-space-planning-a-home-registry-and-the-atlas.md)
- [`status --all`: the cross-space CLI overview](6g0fzhc3m7mc-status-all-the-cross-space-cli-overview.md)
- [Reusable workspace discovery seam](6fgcr2403sjn-reusable-workspace-discovery-seam-lift-init-doctor-fix-off-the-cli.md)
