# Configuration and planning spaces

[Architecture index](../ARCHITECTURE.md) · [Composition](boundaries.md#composition-and-local-exceptions)

This guide explains local discovery, preferences and workspace relationships.
[Core configuration](../../internal/core/configuration.go),
[registry](../../internal/core/space_registry.go) and
[workspace](../../internal/core/workspace.go) services own application policy;
secondary adapters own TOML/filesystem operations.

## Repository discovery and identity

[Config](../../internal/config/config.go) walks from an explicit start to the nearest
governing marker or Git/root boundary. A missing config is distinct from a malformed
existing one: only an ordinary miss may take a no-repository landing route.

`taskflow_root` is in-tree, anchored to its config directory. Outward
`planning_repo`/`tracked_repos` references anchor to the canonical Git checkout,
so a committed relative reference means the same thing across worktrees.
[Worktree resolution](../../internal/config/worktree.go) owns that distinction.

Init bootstraps topology and durable repository identity. A pointer verifies its
stored expected ID after resolution; missing/mismatched identity refuses rather than
binding whichever tree happens to occupy a path. For one governing marker, config-
directory and configured-data-directory entry points share the physical tree and ID.
Ancestor identity is used only for an exact containment-checked root match, stopping
at the nearest marker/Git boundary and never following pointer ancestors.

Safe config upgrades use explicit, idempotent migration, not another init.
Direct/pointer ID backfills and preference/linkback edits are surgical, atomic and
guarded with a config-directory lock and re-read. A missing-ID data-directory repair
migrates the governing marker instead of minting a second one.
Conflicting overlapping markers remain
[unresolved authority/recovery work](../../planning/tasks/6gjgccnzxj0a-decide-and-enforce-identity-authority-for-overlapping-planning-markers.md);
single-marker alias parity does not choose a winner for them.

Init's typed receipt reports created config, checkout and planning identity.
Optional home registration is a separate best-effort registry use case, not a reason
for config to import home storage. It must not disguise successful bootstrap as
rolled back when registration fails.

## Preferences and home registry

[User config](../../internal/userconfig/userconfig.go) owns home preferences and the
separate advisory spaces registry. Its directory is resolved from
`TSKFLW_CONFIG_HOME`, XDG config home, then the documented config fallback—not
macOS Application Support. Precedence is flag > env > repo > user > default,
merged field-by-field; pointer-valued preferences distinguish unset from false.

Presentation reads are available even when discovery is skipped. Missing home data
is normal; malformed preferences warn/degrade, unlike malformed repository markers
where guessing would fork topology. Cooperating preference/registry writes use
surgical atomic TOML edits under a directory lock; unknown keys and comments survive.
Config never imports userconfig.

[Configstore](../../internal/configstore/fs.go) implements the neutral configuration
port. [Configui](../../internal/configui/editor.go) edits typed preferences through the
application service, defaults to user scope, requires explicit repo scope, and leaves
topology/identity/registry read-only. It is shared by the CLI launcher and full TUI.

## Catalog and summaries

[SpaceRegistryService](../../internal/core/space_registry.go) owns identity grouping, label
policy, selection and mutation receipts through its port. Local
[spacehealth](../../internal/spacehealth/diagnose.go) diagnoses entry points; it
does not repair or group the registry. Spacestore translates these adapter records
once into neutral core values.

[SpaceOverviewService](../../internal/core/space_overview.go) opens a narrow read-only
summary source and chooses a healthy entry point per identity. Metadata summary and
canonical task graph capabilities are explicit, so counts, active work and graph
health share one task snapshot. Health is derived from typed states, not an arbitrary
adapter claim.

Per-space failures retain neutral class/message data. A contended retry returns
only replacements for affected groups, never a complete registry overview. CLI
renders available results before applying its partial-result exit policy; TUI
retains coherent stale groups according to [its refresh rules](tui.md#asynchronous-state-and-refresh).

## Workspace opening

[WorkspaceService](../../internal/core/workspace.go) opens an explicit local start
through [workspacestore](../../internal/workspacestore/fs.go). Registry labels are
presentation context, not discovery input. The resulting service and watcher Layout
are separate capabilities, swapped only on successful opening.

Shared construction retains the initial discovery corpus and a fresh identity reader
anchored to its governing marker, including pointers. Later stores retain the invocation
mutation policy and source-set binding. Neither CLI nor TUI imports persistence to
manufacture a missing service.
[Real opening tests](../../internal/workspacestore/fs_test.go) and
[apply parity tests](../../internal/appwiring/thread_apply_test.go) cover root/ID
replacements, authorization and direct/pointer routes.
