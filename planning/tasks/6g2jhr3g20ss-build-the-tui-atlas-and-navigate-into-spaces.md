---
schema: 1
id: 6g2jhr3g20ss
status: completed
epic: 29-multi-space-planning-a-home-registry-and-the-atlas
description: Render one atlas card per planning identity, select its registered entry point, and enter or return from that space without restarting the TUI.
effort: L
tier: 2
priority: high
autonomy_level: 3
tags: [tui, multi-repo, navigation]
created: "2026-08-22"
updated_at: "2026-08-22"
started_at: "2026-08-22"
completed_at: "2026-08-22"
---

# Build the TUI atlas and navigate into spaces

## Objective

Make the atlas the TUI's cross-space navigation layer. It should show one logical card
per durable planning identity, expose the registered entry points behind that identity,
and let the user enter a healthy checkout and return to the atlas without restarting the
program or losing the context they left.

## Acceptance criteria

- [x] With more than one logical space registered, the TUI exposes an atlas surface built
  asynchronously from `core.SpaceOverviewService`; one dead checkout does not hide
  healthy spaces.
- [x] The atlas renders one card per planning identity, including its summary, problems,
  and registered direct/pointer entry points without duplicating a planning tree.
- [x] Each healthy card defaults to the shared preferred entry point and allows another
  healthy entry point to be selected explicitly; broken entries remain visible but cannot
  be entered.
- [x] Entering a space opens it through the reusable workspace boundary and swaps the
  active entity service/layout only after the open succeeds. A failed or stale async open
  leaves the current session intact and explains the error in the atlas.
- [x] Returning to the atlas and re-entering a space restores that space's tabs, cursors,
  filters, dashboard, and navigation state. Only the active space has an fsnotify watcher.
- [x] The current space/entry point is continuously visible while browsing, and atlas
  navigation is discoverable from the footer, `?` help, and command palette.
- [x] The atlas uses the home/global theme rather than the launch repo's override, shows
  every entry-point directory in subdued text, and supports name/activity/registration
  ordering plus reverse without losing the selected logical space.
- [x] Zero/one-space and registry-unavailable states degrade deliberately: the ordinary
  single-space TUI stays usable, while an explicitly opened atlas reports why it cannot
  navigate.
- [x] Model, rendering, navigation, watcher lifecycle, and CLI integration tests cover
  atlas → space → atlas round trips and failure isolation.

## Out of scope

- Remote/served spaces or a transport abstraction.
- `space scan`, moved-checkout guessing, cached summary persistence, or watching every
  registered space continuously.
- Per-space theme overrides or mutation confirmation policy for selected worktrees; those
  require their own product decision if navigation demonstrates the hazard.

## Related

- Epic [29-multi-space-planning-a-home-registry-and-the-atlas](../epics/29-multi-space-planning-a-home-registry-and-the-atlas.md)
- Blocked by: [Establish reusable workspace opening for atlas navigation](6g2jhr31f14p-establish-reusable-workspace-opening-for-atlas-navigation.md)
- Design sketch: [Multi-space: home registry and the atlas](../research/6g0ajre026c6-multi-space-home-registry-and-the-atlas.md)

## Implementation notes (2026-08-22)

The TUI now loads the shared `SpaceOverviewService` projection asynchronously and renders
one card per planning identity with every direct/pointer entry point and its registry
path. Healthy selections open through `WorkspaceService`; durable identity is rechecked
after discovery so a checkout that drifted since the registry snapshot cannot silently
replace the selected tree. Broken entries and per-space summary failures remain local to
their cards.

Each checkout has a cached `spaceSession` containing its tabs, filters, cursors,
dashboard, detail, focus, zoom, and navigation stack. Workspace-generation envelopes
drop old async results after a switch while leaving Bubble Tea runtime control messages
visible. Watchers are created for the candidate workspace, swapped only after a
successful open, and the previous active watcher is then closed; cached sessions retain
no watcher.

The atlas is reachable through `a`, `:atlas`, and the command palette, and appears in the
top rail/footer/context-specific help. It defaults to name order; `o` cycles through
activity and registration order and `O` reverses. Its chrome uses the home-scoped theme
(with process-global flag/environment overrides), never the launch repo's local theme.

Validation: a real terminal smoke navigated from this repo's four-card atlas into the
registered `desirelines` pointer checkout without restarting. `go test -race ./...`,
`golangci-lint run ./...`, planning lint, generated CLI docs, and `git diff --check` are
clean.
