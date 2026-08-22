---
schema: 1
id: 6g2jhr31f14p
status: completed
epic: 29-multi-space-planning-a-home-registry-and-the-atlas
description: Open a registered entry point as a neutral planning workspace so the TUI can switch spaces without importing discovery or filesystem adapters.
effort: M
tier: 2
priority: high
autonomy_level: 3
tags: [architecture, tui, multi-repo]
created: "2026-08-22"
started_at: "2026-08-22"
updated_at: "2026-08-22"
completed_at: "2026-08-22"
---

# Establish reusable workspace opening for atlas navigation

## Objective

Give the atlas a framework-free way to open any healthy registered entry point as a
planning workspace. Discovery and concrete filesystem construction must remain behind a
secondary adapter; the TUI receives a core application capability and neutral workspace
data, not imports of `config`, `store`, or `spacestore`. This is the narrow portion of
[the deferred architecture task](6fgcr2403sjn-reusable-workspace-discovery-seam-lift-init-doctor-fix-off-the-cli.md)
whose trigger has now fired.

## Acceptance criteria

- [x] A consumer-owned core port and service open an explicit start directory into a
  neutral workspace containing the planning identity, resolved root, entity service, and
  watcher layout needed by a primary adapter.
- [x] The filesystem implementation owns `config.Discover` and concrete `store.NewFS`
  construction; `internal/tui` does not import repo discovery or filesystem adapters.
- [x] The TUI composition root can receive the opener alongside its initial workspace;
  existing `tskflwctl ui`, `-C`, and `--space` startup behavior remains unchanged.
- [x] Opening a pointer entry reaches the external planning tree and retains the selected
  entry-point metadata needed to tell the user what they entered.
- [x] Missing, malformed, and identity-mismatched targets preserve the discovery error
  (including typed conflicts where discovery supplies one) and cannot fall back to the
  current workspace.
- [x] Focused core/adapter/CLI tests cover direct and pointer workspaces, error
  propagation, and compatibility of the existing one-space launch.
- [x] Architecture documentation names the new boundary and keeps home registry,
  repo-scoped discovery, and planning storage responsibilities distinct.

## Out of scope

- Atlas rendering, keybindings, cards, and switcher behavior.
- Refactoring the Bubble Tea model into cached per-space sessions.
- Moving `doctor`, `lint --fix`, initialization, or every Cobra discovery call merely
  because the earlier audit task grouped them together.
- Remote/served workspaces, filesystem scanning, or persisted per-space state.

## Related

- Epic [29-multi-space-planning-a-home-registry-and-the-atlas](../epics/29-multi-space-planning-a-home-registry-and-the-atlas.md)
- Next: [Build the TUI atlas and navigate into spaces](6g2jhr3g20ss-build-the-tui-atlas-and-navigate-into-spaces.md)
- Architecture source: [Reusable workspace discovery seam](6fgcr2403sjn-reusable-workspace-discovery-seam-lift-init-doctor-fix-off-the-cli.md)

## Implementation notes (2026-08-22)

`core.WorkspaceService` now opens an explicit local start through the consumer-owned
`WorkspaceStore` port and returns the planning identity/root, selected entry-point label,
entity service, and watcher layout as one neutral runtime workspace.
`internal/workspacestore` is the sole translation point from `config.Discover` and the
concrete Markdown `store.FS`; the CLI composition root injects the service into the TUI.
An empty start is rejected before adapter access so atlas navigation can never degrade to
cwd. Missing/malformed discovery errors pass through unchanged, and pointer identity
mismatches retain `domain.ErrConflict`. The request also carries an optional expected
planning id, rechecked after discovery to reject registry-snapshot drift before a primary
adapter swaps its active context.

Validation: the full `go test -race ./...` suite passed; `golangci-lint run ./...`,
planning lint, and `git diff --check` are clean.
