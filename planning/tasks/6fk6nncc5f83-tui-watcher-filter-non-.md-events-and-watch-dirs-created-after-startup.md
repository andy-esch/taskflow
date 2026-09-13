---
schema: 1
id: 6fk6nncc5f83
status: ready-to-start
epic: 18-tui-bubble-tea-interactive-planning-browser
description: The fsnotify watcher reloads on ANY event in the entity dirs, so editor swap/backup files cause reload churn; events are never filtered to .md.
effort: S
tier: 3
priority: low
autonomy_level: 3
tags: [tui, watcher, robustness]
created: "2026-07-05"
updated_at: "2026-09-13"
audited: "2026-09-13"
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

> **Update 2026-09-13** (automated weekly sweep): the first half of this finding has
> SHIPPED; the second half is unchanged and still reproducible.
>
> **Watching dirs created after startup — done.** `newWatcher` no longer best-effort
> `Add`s once and forgets. `watcher.reconcile` (`internal/tui/watch.go:87`) converges the
> concrete fsnotify set on the current desired set, attaching present leaves plus
> de-duplicated nearest-existing-parent sentinels so a leaf that appears later is
> discovered (`:31`, `:49`). Every event reconciles (`:341`, `:346`) and `debounceTick`
> reconciles once more after the quiet period (`:361-368`) to close the nested-creation
> race. Partial coverage is reported as `degraded` rather than silently missing.
>
> **Filtering non-`.md` events — still open.** `waitForFS` (`:331-349`) reads the channel
> as `case _, ok := <-w.fsw.Events:` — the event value is discarded, so neither the
> filename nor the op is ever inspected, and any event on a watched dir still returns a
> bare `fsEventMsg`. An editor's `.slug.md.swp` still triggers the same reload as a real
> save. The 200ms debounce and cursor-preserve-by-id reload still mitigate it, so `low`
> priority remains right.
>
> Scope is therefore roughly halved; `effort` set to `S` and the `description` narrowed to
> the remaining half. The title still names both halves — renaming was left to a human.

## Progress log

- 2026-09-13: automated weekly sweep — verified against `main`; the watch-dirs-created-after-startup half has shipped (`reconcile` + parent sentinels), the non-`.md` event filter has not; `effort` → S and `description` narrowed to the remaining half.
