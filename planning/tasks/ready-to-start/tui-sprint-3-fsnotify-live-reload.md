---
status: ready-to-start
epic: 18-tui-bubble-tea-interactive-planning-browser
description: Watch the planning dirs and refresh the TUI on external edits, preserving the cursor by slug
effort: Unknown
tier: 3
priority: medium
autonomy_level: 3
tags: [tui, bubble-tea]
created: "2026-06-10"
---

# TUI sprint 3 fsnotify live reload

## Objective

Auto-refresh the TUI when files change on disk (CLI/agent edits, git pulls) —
the `reloadMsg` path was plumbed in S1; this wires the source. (~½–1 day; the
hard parts are debounce + cursor preservation, not fsnotify itself.) See
[[18-tui-bubble-tea-interactive-planning-browser]].

## Scope

- [ ] `watch.go`: a long-lived watcher (outside the event loop) adding watches
      for the ~9 task/epic/audit subdirs (fsnotify is non-recursive); emits one
      **debounced** `reloadMsg` per ~200ms quiet period (coalesce editor
      temp-file-rename storms via `tea.Tick`, not `time.Sleep` in a Cmd).
- [ ] On `reloadMsg`: re-fire the entity loaders **and re-arm** the watcher.
- [ ] **Preserve cursor by slug** (not index): capture `SelectedSlug()` before,
      re-`Select()` it after the reload; clamp to nearest if it vanished.
- [ ] Tests inject a no-op/synthetic watcher (don't depend on real fs events);
      assert a `reloadMsg` re-loads and keeps the cursor on the same slug.

## Acceptance

- [ ] Editing/moving a task file (or a CLI `task move`) updates the TUI within
      ~200ms with the cursor still on the same task. No event-storm thrash.
      Suite + lint green.

## Out of scope

- Watching newly-created dirs at runtime (status dirs are fixed by `init`).

## Follow-up folded in from S2a review (2026-06-11)

- **Reload currently refreshes only the active tab.** `r` (and the `reloadMsg`
  path) reloads `m.cur()` and restores its cursor by id; the other entity tabs
  keep their already-loaded data until next visited. With fsnotify this should
  refresh **all loaded tabs** (or mark inactive ones stale → reload on switch),
  so a `task move` from another process is reflected on whichever tab you land on.
  Preserve each tab's cursor by id, as 2a does for the active one.

## Related

- Epic [[18-tui-bubble-tea-interactive-planning-browser]]
