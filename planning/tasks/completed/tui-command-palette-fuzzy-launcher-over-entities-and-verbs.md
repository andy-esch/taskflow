---
schema: 1
status: completed
epic: 18-tui-bubble-tea-interactive-planning-browser
description: 'Floating fuzzy launcher (⌘P style) over all entities + verbs — type to jump to any task/epic/audit or run a verb. Reuses listfilter + bubbles/list on the existing modal system; complements the : bar.'
effort: Unknown
tier: 3
priority: medium
autonomy_level: 3
tags: [tui, ux]
created: "2026-06-23"
updated_at: "2026-06-24"
started_at: "2026-06-24"
completed_at: "2026-06-24"
---
## Objective

A floating fuzzy launcher (⌘P / Telescope style) over a single index of **all entities + verbs**. One key opens an overlay with a text input and a live-filtered list; typing fuzzy-matches across every task/epic/audit (slug, title, description) and the lifecycle/nav verbs; Enter jumps to the entity or runs the verb; Esc dismisses. It collapses today's structured nav (tab-switch → `/` filter, or `:` command-word jump) into one motion across all three kinds, and makes verbs discoverable by name. Complements `:` (kept as the terse path), not a replacement.

## Acceptance criteria

- [ ] A dedicated key opens the palette; Esc closes it; it captures keys while open and does not fire while a `/` filter or detail-find is being typed.
- [ ] The index spans all loaded tasks, epics, and audits (matched on slug/title/description) plus the lifecycle + nav verbs.
- [ ] Fuzzy matching reuses `listfilter`; the list reuses `bubbles/list` (as the prompt picker does) — slots into the existing modal registry (help/action/follow/edit), no new overlay machinery.
- [ ] Enter on an entity jumps to it (switch tab + select + open detail, via the existing `jumpTo`); Enter on a verb runs it via the existing dispatch.
- [ ] Help overlay (`?`) lists the new key; README TUI keys updated.
- [ ] teatest/unit coverage: open → type → select entity asserts the jump; a verb selection asserts dispatch.

## Out of scope

- The overlay-layers harvest (`tui-overlay-compositing-via-lipgloss-v2-layers-canvas`) — independent; the palette ships on the current overlay system.
- Recents / command history / pinned commands — a later nicety.
- Verbs needing extra input (editor, multi-step) — route to the existing flows (action menu / inline edit) rather than re-implementing here.

## Related

- Design: `planning/research/2026-06-23-tui-v2-feature-wins-decisions.md` (layers/overlays adopt decision).
- Reuses: `internal/listfilter`, `bubbles/list` (see `internal/cli/prompt/picker.go`), the modal registry + `jumpTo`/dispatch in `internal/tui`.
- Epic [[18-tui-bubble-tea-interactive-planning-browser]].
