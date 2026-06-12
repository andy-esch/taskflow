---
status: completed
epic: 18-tui-bubble-tea-interactive-planning-browser
description: Per-tab error state, generation guard for list loads, preserve detail scroll on fs reloads, let r recover a failed initial load, watcher-off signal
effort: Unknown
tier: 3
priority: medium
autonomy_level: 3
tags: [go, tui, bug, robustness]
created: "2026-06-12"
updated_at: "2026-06-12"
started_at: "2026-06-12"
completed_at: "2026-06-12"
---
# TUI reload robustness

> ⚠️ **Externally proposed — filed from the 2026-06-12 review**
> ([[2026-06-12-critical-code-review-multi-lens]], findings M12/M13/M14 +
> still-open A2/A5 from [[2026-06-11-critical-review-and-polish-research]]).
> Robustness of the watcher/reload plumbing — real bugs, but none blocks the
> branch the way the filter cluster does.

## Objective

1. **M12 — Every fs event yanks the detail pane to the top.**
   `handleListLoaded` → `refreshDetail` → `SetContent` → `vp.GotoTop()`
   (`internal/tui/model.go:170`, `detail.go:62`). Preserve `YOffset`
   (clamped) when the incoming `(kind,id)` matches the displayed item.
2. **M13 — One failing tab's loader blanks the whole UI; concurrent reloads
   race on `m.err`.** Loaders return a global `errMsg`
   (`commands.go:36-38`); `model.go:135-137` swaps the entire View while any
   tab's success clears `m.err` (`model.go:157`) — nondeterministic flicker
   during `reloadAll`, and the failing tab keeps stale rows silently. Store
   the error per tab; reserve `m.err` for the nothing-loaded case.
3. **M14 — List loads have no generation guard** (detail loads do, via
   `isCurrentSelection`, `model.go:104-105`). Cmds run concurrently; cycling
   status views fast can land an older load last — chip says
   `view:completed`, rows show the previous view (`model.go:150-171`). Stamp
   `listLoadedMsg` with a generation or its view; drop mismatches.
4. **A2 — A failed *initial* load is unrecoverable via `r`.** `reloadAll`
   (`model.go:75-84`) skips tabs where `!t.loaded`; after an initial
   `errMsg` nothing is loaded, so `r` produces zero commands. Always reload
   the active tab.
5. **Low — same-id detail loads are unordered** (stale guard is id-equality
   only, `model.go:103-108`); a monotonic request generation makes it
   airtight while you are in here.
6. **A5 residue — `Run` silently leaves live-reload off** when `newWatcher`
   errors (`tui.go:14`); add a one-time dim footer note.

## Acceptance criteria

- [x] External write while reading a long detail body does not reset scroll.
- [x] One tab failing while another succeeds shows the error in that tab
      only; no full-screen swap, no stale-rows-with-no-signal.
- [x] `r` recovers from a failed initial load (test must go through the `r`
      key path, not `m.Init()()` — see `TestModel_RecoversFromFatalError`).
- [x] Out-of-order list/detail responses cannot regress the visible state.

## Related

- Epic [[18-tui-bubble-tea-interactive-planning-browser]]
- Touches `internal/tui/model.go`, `internal/tui/commands.go`,
  `internal/tui/detail.go`, `internal/tui/tui.go`, `internal/tui/watch.go`.