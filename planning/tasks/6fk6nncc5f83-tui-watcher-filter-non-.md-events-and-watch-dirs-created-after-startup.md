---
schema: 1
id: 6fk6nncc5f83
status: ready-to-start
epic: 18-tui-bubble-tea-interactive-planning-browser
description: The fsnotify watcher reloads on any event in the entity dirs (editor swap files churn) and skips dirs absent at startup (new audits/epics unwatched until restart).
effort: Unknown
tier: 3
priority: low
autonomy_level: 3
tags: [tui, watcher, robustness]
created: "2026-07-05"
updated_at: "2026-07-05"
---

# TUI watcher: filter non-.md events and watch dirs created after startup

## Objective

<why / what — one short paragraph>

## Acceptance criteria

- [ ] <observable outcome>

## Out of scope

- <explicitly excluded>

## Related

- Epic [18-tui-bubble-tea-interactive-planning-browser](../epics/18-tui-bubble-tea-interactive-planning-browser.md)

## Finding (adversarial review, 2026-07-05)

`internal/tui/watch.go`:

- `newWatcher` best-effort `fsw.Add`s each dir and skips ones absent at startup; a dir
  created later (a fresh repo's `audits/`/`epics/` created while the TUI runs) is never
  watched, so live-reload silently misses it until restart.
- `waitForFS` returns `fsEventMsg{}` for ANY event on a watched dir — an editor's swap/temp/
  backup files (`.slug.md.swp`) trigger reload churn, screen flashes, cursor stutter.

Fix: filter events to `.md` (ignore dotfiles/swap); re-evaluate missing watched dirs (or
watch-on-create). Pre-existing (not flatten-specific) and mitigated by the 200ms debounce +
cursor-preserved-by-id reload — hence low priority.
