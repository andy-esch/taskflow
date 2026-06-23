---
schema: 1
status: completed
epic: 20-cli-ux-and-ergonomics
description: Render epic show's epic→tasks hierarchy as a lipgloss/v2 tree grouped by status (TTY human face; --json untouched). v2 already in graph.
effort: Unknown
tier: 3
priority: medium
autonomy_level: 3
tags: [cli, ux, render]
created: "2026-06-23"
started_at: "2026-06-23"
updated_at: "2026-06-23"
completed_at: "2026-06-23"
---

# epic show lipgloss v2 tree rendering

## Objective

<why / what — one short paragraph>

## Acceptance criteria

- [ ] <observable outcome>

## Out of scope

- <explicitly excluded>

## Related

- Epic [[20-cli-ux-and-ergonomics]]

Done: epic show renders a status-grouped lipgloss/v2 tree (rootless; node text pre-styled by st so --color is honored, connectors plain → ANSI-free under --color=never; --json untouched). First real lipgloss/v2 use in render. Test: TestEpicShowHuman_Tree.
