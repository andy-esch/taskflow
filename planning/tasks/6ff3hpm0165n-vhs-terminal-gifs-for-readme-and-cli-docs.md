---
schema: 1
status: completed
epic: 20-cli-ux-and-ergonomics
description: Author vhs .tape scripts for help/status/TUI + a generation recipe + README wiring; tooling, not a runtime dep.
effort: Unknown
tier: 3
priority: low
autonomy_level: 3
tags: [cli, docs, tooling]
created: "2026-06-23"
updated_at: "2026-06-23"
completed_at: "2026-06-23"
id: 6ff3hpm0165n
---

# vhs terminal GIFs for README and CLI docs

## Objective

<why / what — one short paragraph>

## Acceptance criteria

- [ ] <observable outcome>

## Out of scope

- <explicitly excluded>

## Related

- Epic [20-cli-ux-and-ergonomics](../epics/20-cli-ux-and-ergonomics.md)

Authored: assets/vhs/{help,status,epic-show,task-list}.tape + assets/vhs/README.md, a 'just gifs' recipe (build → run each tape with ./bin on PATH), and a README Demos section. vhs isn't in the container (not a build/runtime dep), so the GIFs themselves are generated on a vhs-equipped host/CI via 'just gifs' — README image links populate then.
