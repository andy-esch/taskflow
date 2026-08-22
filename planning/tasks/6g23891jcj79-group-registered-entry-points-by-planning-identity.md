---
schema: 1
id: 6g23891jcj79
status: completed
epic: 29-multi-space-planning-a-home-registry-and-the-atlas
description: Treat one durable planning identity as one logical space and its registered direct/pointer paths as selectable entry points.
effort: M
tier: 3
priority: high
autonomy_level: 3
tags: [cli, multi-repo, tui, config]
created: "2026-08-20"
started_at: "2026-08-20"
updated_at: "2026-08-22"
completed_at: "2026-08-20"
---

# Group registered entry points by planning identity

## Objective

Make the registry reflect the relationship already encoded by modern repo configs:
the direct planning checkout carries `id`, while each implementation pointer carries
that same value as `planning_repo_id`. The resolved durable id defines **one logical
space**; the independently registered paths are **entry points** into it, addressed by
their local registry labels.

This is a derived read model, not new registry state. `spaces.toml` keeps one row per
path so `--space <label>` can still select an exact checkout. List, cross-space status,
and the proposed atlas group those rows by resolved planning identity so the same task
tree is not presented or summarized multiple times.

The TUI concept already settled on “one planning repo, one card” for worktrees. This
extends the same rule to direct and pointer registrations: one card per planning
identity; direct/pointer checkouts and later-discovered worktrees are selectable entry
points or variants on that card, never duplicate cards.

## Model and display contract

- Prefer the registry entry's `verify_id` as the group identity. It is the durable id
  asserted when the entry was added and remains useful when its path is now missing or
  mismatched. For a legacy entry without `verify_id`, use a successfully discovered id,
  then the physical resolved planning root. A broken id-less entry remains its own group.
- Derive the entry role from repo discovery: `direct` when the registered config owns the
  planning tree, `pointer` when it has `planning_repo`, and `unknown` when a broken entry
  cannot be inspected. Do not add role, parent, or child fields to `spaces.toml`.
- Preserve registry order for groups and entries. For human presentation only, prefer the
  first direct entry as a group's anchor; indentation means “same planning space,” not
  filesystem ownership.
- Keep JSON's `spaces` array flat and backward-compatible. Add derived `role` and
  `planning_id` fields so callers can reproduce the projection without reading repo
  configs. The registry label (`id`) remains the address for a specific entry point.
- `status --all` and the atlas summarize/load once per group. The global `--space` flag
  still selects an entry point, so two labels in one group may intentionally choose
  different working directories while reaching the same planning data.

## Acceptance criteria

- [x] A shared, read-only health projection groups direct and pointer registrations with
  the same durable planning id, with deterministic order and a safe legacy fallback.
- [x] Each diagnosed entry exposes `direct`, `pointer`, or `unknown`; no relationship
  metadata is persisted to `spaces.toml`.
- [x] `space list` renders a multi-entry planning identity as one compact tree while
  keeping every label, state, path, diagnosis, and remedy visible.
- [x] `space list --json` stays a flat array and adds derived `role` and `planning_id`;
  schema/docs/goldens are regenerated.
- [x] Tests cover one direct planning checkout plus two pointers, pointer-only and
  singleton groups, broken entries with a retained `verify_id`, and legacy id-less repos.
- [x] The `status --all`, global `--space`, atlas research, ADR, and architecture docs use
  the logical-space/entry-point distinction consistently.
- [x] `just test`, `just lint`, and documentation checks pass.

## Completion

Implemented the derived grouping projection and entry roles without changing the
registry schema. (The projection now lives in `core.SpaceRegistryService.Catalog`.) Human
`space list` promotes a direct checkout to the top of a
compact tree and nests every other registered entry point that shares its planning id;
machine output remains a flat, insertion-ordered array with `planning_id` and `role` under
schema version 1.39.

Validated against the real six-entry registry: `desirelines-planning`, `desirelines`, and
`desirelines-deploy` render as one group, while JSON retains all three exact checkout
labels. Legacy rows derive a newly available config id, id-less healthy repos fall back to
physical planning root, and broken entries retain grouping through `verify_id` without
inventing a role. Full race tests pass, golangci-lint reports zero issues, planning lint is
clean, CLI docs/schema comments/goldens are regenerated, and `git diff --check` is clean.

Architecture follow-up (2026-08-22): grouping policy now lives in
`core.SpaceRegistryService.Catalog`; `spacehealth` remains the filesystem diagnosis helper
behind `spacestore`. The behavior and wire contract described above are unchanged.

## Out of scope

- Implementing `status --all`, global `--space`, or the atlas itself.
- Registering or persisting discovered worktrees; their live variant discovery remains a
  later TUI concern.
- Adding parent/child, role, or group tables to the registry schema.
- Choosing among multiple direct clones/worktrees beyond the deterministic first-direct
  display preference; labels continue to address exact registered paths.

## Related

- Epic [29-multi-space-planning-a-home-registry-and-the-atlas](../epics/29-multi-space-planning-a-home-registry-and-the-atlas.md)
- [Multi-space: home registry and the atlas](../research/6g0ajre026c6-multi-space-home-registry-and-the-atlas.md)
- [`status --all`: the cross-space CLI overview](6g0fzhc3m7mc-status-all-the-cross-space-cli-overview.md)
- [Global `--space` flag](6g0fzk8mazrc-global-space-flag-run-any-command-against-a-registered-space.md)
