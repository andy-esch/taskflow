---
schema: 1
id: 6g0fzhc3m7mc
status: completed
epic: 29-multi-space-planning-a-home-registry-and-the-atlas
description: Per-space summaries plus a combined in-progress list across the registry. The CLI counterpart of the proposed atlas, and the cheapest test of whether the board is worth building.
effort: Unknown
tier: 3
priority: medium
autonomy_level: 3
tags: [cli, multi-repo]
created: "2026-08-15"
updated_at: "2026-08-21"
started_at: "2026-08-21"
completed_at: "2026-08-21"
---
# `status --all`: the cross-space CLI overview

## Objective

Answer the question that actually motivated epic 29 — **"what am I in the middle of,
anywhere?"** — in the CLI, before any TUI work. `status` is already the terminal
counterpart of the TUI's Overview dashboard; `status --all` is the counterpart of the
proposed atlas, and the cheapest possible test of whether a cross-space view is worth
building a screen for.

## Notes

- One `core.Summary` per **logical planning identity**, using the shared
  `spacehealth.Group` projection rather than iterating registry rows. A direct planning
  checkout plus two registered implementation pointers is one block and contributes each
  task once, not three times. Use a healthy direct entry point when available, otherwise
  the first healthy entry in registry order.
- Render each summary as a compact per-space block
  plus a **combined in-progress list** with a space badge on each row. The combined
  list is the hypothesis worth testing: cards orient, but the rail may be the real
  payload.
- Broken entry points render their diagnosis within their logical space (from the health
  projection) rather than failing the command. A group remains loadable through another
  healthy entry point; one dead path must not cost the other spaces.
- `--json` carries a `schema_version` and an array of per-space envelopes; goldens
  regenerated (`go test ./internal/cli -update`).
- Reads only. No writes, no registration side effects.
- Sequential scanning is fine to start; measure before adding concurrency.

## Acceptance criteria

- [x] `status --all` summarizes every logical planning identity plus a combined
  in-progress list, without duplicate tasks from pointer/direct entry points
- [x] A broken entry point shows its diagnosis inline; the command still exits 0 and a
  group with another healthy entry remains summarized
- [x] Unreadable planning files or a selected tree that fails during loading render as
  partial results and make the overall command exit non-zero
- [x] No registry / no spaces → falls back to today's single-repo `status` output
- [x] `--json` envelope with one top-level `schema_version`; nested summaries are
  versionless payloads and goldens are regenerated
- [x] `just test` + `just lint` green

## Out of scope

- The TUI atlas board — this exists partly to decide whether that is worth building
- Caching or persisted stats

## Related

- Epic [29-multi-space-planning-a-home-registry-and-the-atlas](../epics/29-multi-space-planning-a-home-registry-and-the-atlas.md)
- Sketch: [6g0ajre026c6-multi-space-home-registry-and-the-atlas](../research/6g0ajre026c6-multi-space-home-registry-and-the-atlas.md)

## Implementation (2026-08-21)

Implemented `status --all` as the read-only cross-space CLI projection.
`core.SpaceOverviewService` owns grouping orchestration, healthy-direct/first-healthy
selection, one shared as-of clock, per-group failure isolation, and the combined
in-progress set through consumer-owned `SpaceOverviewStore` and `SummaryStore` ports.
`internal/spacestore` is the filesystem adapter and reuses `spacehealth.Group`, so Cobra
does not reinterpret registry identity or diagnosis.

Human output is one compact block per durable planning identity plus a space-badged
in-progress rail. Broken entry points and late summary-load failures remain inline data
and do not fail healthy groups. With an empty registry, `status --all` resolves the
current repo and delegates to the existing single-repo renderer/error contract
byte-for-byte. Explicit `--space` conflicts with `--all`; ambient `TSKFLW_SPACE` cannot
collapse the all-spaces scope.

The JSON contract is schema version 1.42 with `StatusAllEnvelope`, the selected entry
point per space, all entry-point diagnoses, nested summaries, and a combined in-progress
array. The outer envelope owns the one schema version; its nested summaries reuse a
versionless `SummaryJSON` payload. Schema comments, JSON Schema, machine goldens,
generated CLI reference, README, architecture notes, epic state, and the atlas research
progress note were updated.
End-to-end tests cover direct/pointer deduplication, broken alternate diagnosis,
cross-space badges, empty-registry fallback, scope conflict, and the byte golden; core
tests cover selection and group-local load isolation.

The pre-release hardening pass replaced unconstrained role/state strings plus an
independent healthy boolean with typed `core.SpaceRole` / `core.SpaceState` values and a
derived `Healthy()` invariant. It also made the all-spaces exit policy match ordinary
`status`: broken registry entries stay informational, while unreadable entity files and
post-selection tree-load failures return exit 11 only after all available output has been
rendered. Status renderers now live in the feature-scoped `render/status.go` rather than
continuing to grow the package's catch-all file.

Validation: `just test` passes under the race detector; `just lint` reports 0 issues;
`git diff --check` is clean. A read-only smoke run against the real six-entry registry
collapsed it to four logical spaces and five in-progress rows.
