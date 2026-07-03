---
status: completed
epic: 18-tui-bubble-tea-interactive-planning-browser
description: Clear the stale tui spike, extract a shared status theme, add the ui command, and stand up the Bubble Tea model skeleton with a test harness
effort: Unknown
tier: 3
priority: medium
autonomy_level: 3
tags: [tui, bubble-tea]
created: "2026-06-10"
updated_at: "2026-06-10"
started_at: "2026-06-10"
completed_at: "2026-06-10"
id: 6faxn1802tph
---

# TUI sprint 0 foundation, shared theme, test harness

## Objective

De-risk everything: prove the TUI-as-second-adapter wiring end-to-end with a
minimal model, extract the shared theme, and stand up the test harness — before
any real UI. See [[18-tui-bubble-tea-interactive-planning-browser]] for the
locked architecture/decisions.

## Scope

- [ ] Replace the 53-line `internal/tui/model.go` stub.
- [ ] **Extract `theme`** (dependency-free, imports only `domain`): move the
      `statusGlyph`/`Bucket`/priority switches out of `render/style.go` into
      `theme.Status/Bucket/Priority → {Glyph, Color hex}`. CLI `render` maps
      `Color`→nearest ANSI (keep byte-stable output); confirm CLI tests still
      pass. This is the one cross-cutting refactor; do it first.
- [ ] `go get` `charmbracelet/bubbles`, `charmbracelet/x/teatest`,
      `fsnotify/fsnotify`.
- [ ] `tskflwctl ui` cobra command in `root.go` → `tui.Run(app.Svc, app.Cfg.Root)`.
- [ ] Root `Model{svc, root, width, height, focus, ...}`; `Init` fires an
      async `loadTasks` Cmd; handle `tea.WindowSizeMsg`; render a single
      (bubbles/list) task list; `q`/`Ctrl+c` quit. No detail pane yet.
- [ ] `messages.go`/`commands.go`: `loadTasks` Cmd → `tasksLoadedMsg`; **no
      service I/O in `Update`/`View`**.
- [ ] Test harness: one message-injection unit test (send `WindowSizeMsg` +
      `tasksLoadedMsg`, assert `View()` shows a task) over a `fakeStore`-backed
      `Service`; one `teatest` smoke test (inject a no-op watcher so it exits).

## Acceptance

- [x] `tskflwctl ui` launches, loads tasks via `Service`, lists them (colored
      status glyphs via the shared theme), quits cleanly, resizes without panic.

## Done (2026-06-10)

- **`internal/theme`** extracted (imports only `domain`): `Status`/`Bucket`/
  `Priority`/`Percent` → `{Glyph, Color}`. CLI `render` maps `Color`→ANSI via
  `ansiCode` (**byte-stable** — all CLI tests pass unchanged); TUI maps via
  `lipColor`→lipgloss. One place decides "in-progress is yellow ●".
- **`internal/tui`** (stub replaced): `tui.Run(svc, root)`, root `Model` (loads
  tasks via a `loadTasks` **Cmd** — no service I/O in `Update`/`View`),
  `bubbles/list` with a custom `taskDelegate` (themed glyph row), `WindowSizeMsg`
  sizing, `q`/`Ctrl+c` quit. Production TUI imports **only `core.Service`** (no
  store/fs) — the second-adapter invariant holds.
- **`ui` cobra command** wired to `app.Svc`/`app.Cfg.Root`.
- **Deps:** `bubbles`, `x/exp/teatest`. **Harness:** a message-injection unit
  test (size + load → `View()` shows the task) over a real `Service` on a temp
  repo, + a `teatest` smoke test (launch + `q` + finished). Full suite + lint
  green.

## Out of scope

- Detail pane, navigation, tabs, search (S1+). Just prove the wiring + harness.

## Related

- Epic [[18-tui-bubble-tea-interactive-planning-browser]]
