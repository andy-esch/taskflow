---
schema: 1
id: 6gppr2kfa1dz
status: ready-to-start
epic: 18-tui-bubble-tea-interactive-planning-browser
description: Split 'unreadable record' from 'broken gate' at every Thread marker site with one helper over core state, and name the anchorable and gate-sound questions separately.
effort: half day
tier: 3
priority: medium
autonomy_level: 3
tags: [tui, threads, design]
created: "2026-10-23"
audit_sources: [planning/audits/6ga93gvffkpc-2026-09-15-adapter-hygiene.md]
updated_at: "2026-10-23"
---
# Stop branding readable broken-gate Thread members as unreadable

## Objective

Thread surfaces draw `theme.MarkerUnreadable`, the red `!` that the dashboard and legend define as
an unreadable file, whenever a member's gate is broken, even when the member's own file is fine and
the broken record is a prerequisite that may not be on screen. A fix was approved during
2026-10-23 audit triage. Choose the marker once from core state so it cannot misdirect diagnosis.

## Acceptance criteria

- [ ] One TUI helper chooses the marker from a member's task and graph state, and every Thread member/node marker site plus the spatial inspector's role/gate color uses it.
- [ ] A readable task with a missing or unreadable prerequisite renders a broken-gate marker distinct from the unreadable marker, covered by a fixture asserting the unreadable glyph is absent.
- [ ] Anchor selection and marker choice call separately named predicates (record readable versus gate sound), with no change to which members are anchorable.
- [ ] The `?` legend describes the broken-gate marker.

## Out of scope

- Publishing a new derived boolean on core's task graph state, unless the helper cannot be written without it.
- The broader Thread row redesign in [6g87qn72901g](6g87qn72901g-make-alternate-tui-views-and-thread-identity-discoverable.md).

## Related

- Epic [18-tui-bubble-tea-interactive-planning-browser](../epics/18-tui-bubble-tea-interactive-planning-browser.md)
- Audit [2026-09-15 adapter hygiene](../audits/6ga93gvffkpc-2026-09-15-adapter-hygiene.md), findings M2 and L2
- Thread [Refine Thread and TUI navigation](../threads/6g90h4pg0q7n-refine-thread-and-tui-navigation.md)
