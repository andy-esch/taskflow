---
schema: 1
id: 6g1xp8qymz1m
status: completed
epic: 20-cli-ux-and-ergonomics
description: Make init bootstrap-only and consolidate config inspection, migration, health, and interactive editing beneath one visible config command.
effort: L
tier: 1
priority: high
autonomy_level: 3
tags: [cli, ux, config, migration, tui]
created: "2026-08-20"
started_at: "2026-08-20"
updated_at: "2026-08-21"
completed_at: "2026-08-21"
---
# Consolidate the configuration lifecycle under one config command hub

## Objective

Make configuration discoverable and maintainable without turning `init` into a
catch-all or adding another unrelated top-level command. Introduce one visible
`config` command family that owns configuration inspection, upgrades, health, and
editing; simplify `init` back to establishing a project's planning topology.

This replaces the earlier premise that `init` should configure every key. Creation and
ongoing maintenance are different lifecycle operations:

- `init` establishes a new local tree or external-planning pointer.
- `config` explains, diagnoses, upgrades, and edits an existing installation.

The command family must also provide the application seam for the planned TUI
"Config / About" surface. The focused CLI editor and the full TUI must not grow
independent config readers, validators, or writers.

## Why now

The legacy external-pointer case exposes the current ownership problem. Given:

```toml
planning_repo = "../desirelines-planning"
```

bare `tskflwctl init` resolves the existing file as a scaffold target, then fails because
the config layer correctly refuses to scaffold over a pointer. An explicit pointer init
can backfill `planning_repo_id`, but that behavior is obscure, and its human dry-run can
describe the current on-disk state after previewing a different planned state.

That is not merely one dispatch bug. Schema repair and backfill have accumulated inside
a command whose intuitive job is bootstrapping. At the same time, effective theme and
pager settings are difficult to discover because values may come from flags,
environment, repository config, user config, or defaults.

## Command contract

| command | responsibility |
| --- | --- |
| `tskflwctl init` | Create a new scaffold or external-planning pointer. On an existing config, report that it is initialized and identify any available migration; do not silently perform schema upgrades. |
| `tskflwctl config` | Deterministic alias for `config show`, regardless of TTY state. |
| `tskflwctl config show` | Print the resolved effective configuration, configuration paths, planning target, and the source/provenance of each effective value. |
| `tskflwctl config migrate` | Plan and apply idempotent, non-destructive configuration upgrades and durable-ID backfills. Support `--dry-run`, `--json`, and non-interactive use. |
| `tskflwctl config doctor` | Audit repository linkback/config integrity and the home space registry, retaining the current human sections, JSON contract, and validation exit behavior. |
| `tskflwctl config edit` | Launch a focused interactive configuration UI with explicit repository/user scopes, typed controls, validation, and previews. |

### Top-level command budget

`config` earns a top-level slot by absorbing an existing cluster rather than adding a
fourth repair surface:

- Make `config doctor` canonical.
- Preserve `tskflwctl doctor` as a hidden compatibility forwarding command for at least
  one minor release. It must behave identically, including JSON and exit codes.
- Keep `space` top-level: registered spaces are a user-facing domain/address-book
  concept, not a generic settings file.
- Keep `workspace` top-level for now: it answers the operational safety question "where
  will this command act?" `config show` may include the same resolution details without
  removing the shortcut.
- Keep `theme` top-level for catalog discovery and previews. Selecting the active theme
  belongs in `config edit`.
- Keep planning-content validation under `lint`; `config doctor` does not absorb it.

The visible root help count therefore stays neutral when `config` replaces `doctor`.

## Configuration model and scopes

Build one typed application-level projection consumed by all config surfaces. It should
expose:

- repository config path and raw repository-scoped values;
- user config path and raw user-scoped values;
- resolved planning root, planning mode, durable planning id, and tracked repositories;
- effective theme and pager values with provenance (`flag`, `environment`, `repo`,
  `user`, or `default`);
- pending migrations and configuration/link health.

Keep repository and user scopes visually and structurally separate. Theme and pager are
personal terminal preferences and should default to user scope; a repository override
must be an explicit choice. `[pager].enabled` remains tri-state (`unset`, `true`,
`false`) so an override can be returned to inheritance.

Structural values are not ordinary toggles:

- `id` and `planning_repo_id` are displayed but not freely editable.
- Repointing `planning_repo` or moving `taskflow_root` requires a purpose-built,
  data-safe migration and is not enabled by a generic editor field.
- The space registry remains tool-managed through `space`; the editor may display its
  relationship and doctor result but must not expose arbitrary `spaces.toml` editing.

## Migration behavior

The first migration set extracts the safe backfills currently hidden in `init`:

1. Mint a missing direct planning-repository `id`.
2. Resolve a legacy external `planning_repo` relative to the pointer config and record
   its durable `planning_repo_id`.
3. Preserve already-current files byte-for-byte and report that no migration is needed.

Migration planning and application must share one result model so human and JSON output
cannot disagree. A dry-run reports only the planned post-migration result; it must not
preview a successful update and then label the unchanged disk state as failed.

Writes are atomic and surgical: preserve comments, key order, relative path spelling,
and unknown keys. A failed validation or write leaves the original file untouched.

## Interactive editor and TUI reuse

`config edit` is a focused Bubble Tea configuration surface, not an editor for raw TOML.
It should initially support the safe presentation settings:

- theme selection, with the existing theme preview repainting as the cursor moves;
- pager enabled as inherit/on/off;
- pager command with typed validation and a clear indication of its write scope.

`huh.Select` has no highlight-change hook, so the live theme preview needs a small
Bubble Tea selection model. Reuse the theme registry and preview renderer rather than
reimplementing either.

The main `tskflwctl ui` Config/About screen should consume the same projection,
validation/write service, and preferably the same reusable editor component. Read-only
identity and health information can land before all fields become editable, but no TUI
code should write TOML directly.

Interactive entry is TTY-gated. `config show`, `config doctor`, and `config migrate`
remain deterministic for agents and pipes; `config edit` must fail clearly rather than
block when no interactive terminal is available.

## Delivery slices

### Slice A — configuration lifecycle and release boundary

- Add `config show`, `config migrate`, and `config doctor`.
- Keep hidden top-level `doctor` compatibility forwarding.
- Simplify existing-config `init` behavior and fix the legacy-pointer failure.
- Add effective-value provenance and align human/JSON/dry-run results.
- Update command help and configuration documentation.

This is the coherent minimum for the next minor release; the richer editor need not
delay it.

### Slice B — one visual editor in two contexts

- Add the focused `config edit` Bubble Tea surface.
- Add or deepen the main TUI Config/About surface using the same seam.
- Provide live theme preview and scope-aware pager controls.

## Acceptance criteria

- [x] Root help exposes one `config` family without increasing the visible top-level
      command count; hidden `doctor` forwarding preserves compatibility
- [x] `config`/`config show` report effective values, paths, planning identity/mode, and
      per-value provenance in stable human and JSON forms
- [x] `config migrate` safely backfills both a missing direct repo `id` and a legacy
      pointer's `planning_repo_id`; repeated runs are no-ops
- [x] `config migrate --dry-run` and `--json` describe the same planned outcome as the
      applied path, without contradictory current-state verification
- [x] Bare `init` in an already-initialized direct or external-pointer repo reports the
      existing topology and points to migration when needed; it does not produce the
      scaffold-over-pointer conflict or silently upgrade the config
- [x] `config doctor` preserves repository-link and home-registry diagnosis, JSON, and
      validation exit semantics; the hidden top-level alias is behaviorally identical
- [x] Repository and user configuration scopes remain distinct in the projection and UI;
      theme/pager default to user scope and repo overrides require explicit selection
- [x] `config edit` supports live-preview theme choice plus scoped pager inherit/on/off
      and command editing, and refuses non-TTY invocation without blocking
- [x] The focused editor and full TUI Config/About surface reuse the same typed read,
      validation, migration, and write seam; neither edits TOML directly
- [x] Config mutations are atomic and preserve comments, key order, relative path
      spelling, and unknown keys
- [x] Help and configuration docs explain the lifecycle split, command ownership,
      precedence/provenance, migration workflow, and compatibility alias
- [x] Focused unit/model/CLI integration tests cover direct, pointer, current, malformed,
      dry-run, JSON, non-TTY, and interactive paths; `just test` and `just lint` pass

## Implementation closeout (2026-08-21)

Configuration is now one application use case behind the consumer-owned
`core.ConfigurationStore` port. `configstore.FS` composes repository config, user config,
link health, and registry health into neutral core projections; Cobra, the focused
Bubble Tea editor, and the main TUI all call `ConfigurationService` rather than reading
or writing TOML. Theme validation receives the design vocabulary through an injected
configuration option, preserving the core's independence from the concrete registry.

`init` is bootstrap-only for bare existing-config invocations. Safe direct-id and
pointer-id backfills moved to an explicit idempotent migration whose dry-run and apply
share one result model. Repository migration, presentation preferences, and linkback
writes take a shared directory lock, re-read while locked, validate the final TOML, and
atomically rename it; user preferences share the equivalent home-config lock with the
space registry. Comments, ordering, unknown keys, and relative pointer spelling survive.

The editor defaults to user scope, makes repository overrides explicit, supports live
theme preview, pager inherit/on/off, pager command validation, and read-only identity and
health context. `config edit` and `ui` reject non-interactive or dry-run invocation; the
main TUI mounts the same editor from a capability-aware Overview row, `:config`, and the
command palette, and documents the named command in `?` help. The top-level doctor
remains hidden and behaviorally identical to canonical `config doctor`.

Post-audit hardening made the repository and user atomic-write boundaries follow
dotfiles-managed symlinks and preserve existing permission bits. Migration and both
preference scopes now update the symlink target without replacing the link or widening a
restrictive mode; focused regressions cover link survival, target updates, and `0600`
preservation.

Validation completed with the full race-enabled suite and golangci-lint (`0 issues`).
Generated CLI reference, schema comments, JSON Schema, and machine-contract goldens were
regenerated for schema version 1.40; `config show --json` has its own byte-stable golden.

## Out of scope

- A generic `config set <dotted-key> <value>` surface. Add a scriptable setter only after
  the typed model and scope rules exist; do not expose storage details as the first API.
- Moving an existing planning tree by changing `taskflow_root`.
- Repointing an implementation repository to a different external planning repository.
- Moving `space`, `workspace`, `theme`, or planning `lint` beneath `config`.
- Arbitrary manual editing of durable IDs or the home space registry.
- Removing the top-level `doctor` compatibility alias in the same release.
- Cutting the release itself.

## Related

- Epic [20-cli-ux-and-ergonomics](../epics/20-cli-ux-and-ergonomics.md)
- TUI direction and the existing Config/About concept:
  [dashboard-extension-ideas](../research/6fgcr2402att-dashboard-extension-ideas.md)
- The theme registry and preview renderer: epic
  [25-design-system-coherent-palette-and-selectable-themes](../epics/25-design-system-coherent-palette-and-selectable-themes.md)
- Home/user config precedence:
  [home-level-user-config-xdg-location-env-override-theme-pager-precedence](6g0fz5c2x5t4-home-level-user-config-xdg-location-env-override-theme-pager-precedence.md)
- Existing doctor/linkback behavior:
  [linkback-integrity-ambient-warnings-and-doctor-command](6fes83r035p9-linkback-integrity-ambient-warnings-and-doctor-command.md)
  and [space-health-diagnosis-and-a-doctor-registry-section](6g0fzhc1235a-space-health-diagnosis-and-a-doctor-registry-section.md)
- Why the home tier and space registry exist:
  [multi-space-home-registry-and-the-atlas](../research/6g0ajre026c6-multi-space-home-registry-and-the-atlas.md)
