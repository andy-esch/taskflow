---
schema: 1
id: 6g28rv8jm1g7
status: ready-to-start
epic: 21-code-quality-architecture-hardening
description: Space list, selection, overview, add, and forget currently orchestrate registry behavior through separate adapter paths that can drift as TUI/web consumers arrive.
effort: L
tier: 3
priority: medium
autonomy_level: 3
tags: [architecture, multi-repo, core]
created: "2026-08-21"
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

- [ ] Record the chosen service/port shape and dependency direction before implementation;
  distinguish the registry/catalog capability from opening a planning tree, rather than
  growing one catch-all filesystem interface.
- [ ] `space list`, explicit `--space` resolution, and `status --all` consume one typed
  core entry/group projection; delete the duplicate `spacehealth -> wire` and
  `spacehealth -> core` field mappings.
- [ ] Route `space add` and `space forget` through the reusable application service while
  preserving validation-before-write, physical-path deduplication, atomic surgical TOML,
  dry-run behavior, mutation receipts, and existing error classification.
- [ ] Keep configuration doctor on the same registry diagnosis vocabulary without making
  repo-scoped `internal/config` import home-scoped registry state.
- [ ] A non-Cobra consumer can list, group, resolve, add, and forget entries using core
  values without importing `userconfig`, `spacehealth`, TOML, or filesystem packages.
- [ ] Existing human output, JSON fields/order, selection precedence, completion behavior,
  and exit codes remain byte-compatible; focused core/adapter tests plus `just test` and
  `just lint` pass.

## Out of scope

- The reusable `Resolve() -> Workspace` seam, init/doctor/fix promotion, or request context;
  those remain in their existing deferred architecture tasks until atlas/web work begins.
- Building the TUI atlas, adding a served adapter, remote registries, filesystem scanning,
  or changing the `spaces.toml` schema.
- Folding repository configuration and the home registry into one bounded context merely
  because both use TOML.

## Related

- Epic [21-code-quality-architecture-hardening](../epics/21-code-quality-architecture-hardening.md)
- Epic [29-multi-space-planning-a-home-registry-and-the-atlas](../epics/29-multi-space-planning-a-home-registry-and-the-atlas.md)
- [`status --all`: the cross-space CLI overview](6g0fzhc3m7mc-status-all-the-cross-space-cli-overview.md)
- [Reusable workspace discovery seam](6fgcr2403sjn-reusable-workspace-discovery-seam-lift-init-doctor-fix-off-the-cli.md)
