---
schema: 1
id: 6g048yn6z6s3
status: ready-to-start
epic: 28-first-class-entities-new-planning-nouns
description: Add a research tab + detail view to the TUI (read-only, through core.Service as tea.Cmds); the entity landed CLI-only.
effort: Unknown
tier: 3
priority: low
autonomy_level: 3
tags: [tui]
created: "2026-08-14"
---

# TUI surface for research docs

## Objective

The research noun landed CLI-only — the TUI is entirely untouched, so the corpus is
invisible to the interactive browser. Give research a tab/list + detail view so the
newest planning noun is reachable the same way tasks, epics, and audits are.

## Why it's separate

The first pass deliberately stopped at the CLI. `domain.Descriptor` collapsed the
*domain* enumeration, but the registry's own doc-comment is explicit that the
**render/TUI delegates are still per-entity** (epic 21 M9/M10) — so a TUI surface is
hand-written per kind today, not a registry freebie. Doing it inside the entity PR would
have doubled its size for a surface nobody had asked for yet.

## Acceptance criteria

- [ ] A research list view, newest-first (matching `ListResearch`'s order), with the
      created date as the organizing column — no status/progress column, since research
      has neither.
- [ ] A detail view rendering the doc body.
- [ ] Reached through `core.Service` as `tea.Cmd`s — **no store access from the TUI, and
      no I/O in `Update`/`View`** (the standing architectural rule).
- [ ] `research/` is already in `WatchPaths`, so confirm the watcher reloads the tab on
      a doc added at runtime.
- [ ] Nav/help/keybinding text updated for the new tab.

## Out of scope

- Mutation from the TUI (see [research-mutation-verbs-research-set-edit-append](6g048wqgamte-research-mutation-verbs-research-set-edit-append.md) for the
  CLI verbs that would have to exist first).
- Any dashboard rollup counting research — research has no lifecycle, so it has no
  "done" to roll up.

## Related

- Epic [28-first-class-entities-new-planning-nouns](../epics/28-first-class-entities-new-planning-nouns.md)
- Epic [18-tui-bubble-tea-interactive-planning-browser](../epics/18-tui-bubble-tea-interactive-planning-browser.md)
