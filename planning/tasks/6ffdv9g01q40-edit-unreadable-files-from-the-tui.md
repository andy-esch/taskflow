---
schema: 1
status: ready-to-start
epic: 18-tui-bubble-tea-interactive-planning-browser
description: The TUI surfaces malformed files only as a footer count (N unreadable); they are not selectable, so you cannot press E to open one in your editor and fix it. Make them reachable for editing.
effort: Unknown
tier: 3
priority: medium
autonomy_level: 3
tags: [tui]
created: "2026-06-24"
id: 6ffdv9g01q40
audited: "2026-09-13"
updated_at: "2026-09-13"
---
# Edit unreadable files from the TUI

## Objective

The TUI lists only files that parse: `ListTasks/ListEpics/ListAudits` return
`(items, problems, err)`, and the model keeps `problems` separate from the list
items (model.go: `tab.problems = msg.problems`). They surface only as a footer
count — `! N unreadable`. So a file with malformed frontmatter is invisible: you
can't select it, and the `E` → $EDITOR key (which acts on the selected row's path)
can't reach it. That's precisely the moment you most want to jump into your editor
— to *fix* the broken file. Make the unreadable files reachable so `E` can open one.

## Context

Follows the `E` → $EDITOR feature (open the selected entity's whole file in
$EDITOR; see `internal/editor` + model.go `openInEditor`). `E` already works on any
*parsed* entity from either pane; this closes the gap for *unparsed* ones.
`domain.FileProblem` carries the path + parse error. Relates to epic 18 (TUI).

## Acceptance criteria

- [ ] Unreadable files (`problems`) are reachable in the TUI — e.g. a "problems"
      view/list or a footer affordance that jumps to them — without disturbing the
      normal browse flow.
- [ ] `E` opens a selected unreadable file in $EDITOR; on save, live-reload
      promotes it into the normal list if it now parses (and out of `problems`).
- [ ] The parse error is visible (detail pane or row) so you know what to fix
      before opening the editor.
- [ ] go build ./... + go test ./... + golangci-lint run ./... green; README TUI
      section + docs/ARCHITECTURE.md updated if a new view/key is added.

## Implementation sketch

Two plausible shapes — pick one in the design:

1. **A "problems" pseudo-list.** Surface `problems` as selectable rows (their own
   view on the active tab, or a dedicated overlay). Each row exposes a `path()` so
   the existing `selectedPath()`/`E` path just works; the detail pane shows the
   parse error. Smallest reuse of the current machinery.
2. **A footer jump.** From the `! N unreadable` footer, a key cycles to the first
   problem and opens it. Less UI, but problems stay non-browsable — edit-then-recheck.

Either way the reload story already exists: editing the file in place triggers
fsnotify (+ the explicit reload), and a now-parsing file naturally moves from
`problems` into `items`.

## Risks / gotchas

- `problems` are per-tab and per-load (gen-stamped) — a problems view must honor the
  same reload/restore discipline as the entity lists, or a fixed file flickers
  between the problems view and the normal list.
- Keep it TUI-only — don't leak a problems surface into the machine output; the CLI
  already reports unreadable files via `lint`.
- `E` on a file that still won't parse after the edit must stay graceful (it
  reappears in `problems`, no crash) — the same external-edit path `E` already uses.

## Done when

You can see the `! N unreadable` files, select one, press `E`, fix it, and watch it
join the normal list on save — build/test/lint green and the docs updated.

## Sweep audit 2026-09-13

Automated weekly sweep. Citations re-checked against `main` (`1ca31b9`); the premise is
intact and every referenced symbol still exists.

**Verified accurate (2026-09-13):** `ListTasks`/`ListEpics`/`ListAudits` still return
`(items, problems, err)` (`internal/tui/commands.go:39,98,173`); the model still keeps
them apart at `model.go:642` (`tab.problems = msg.problems`) on the
`[]domain.FileProblem` field declared at `entity.go:184`; the footer is still only a
count — `! %d unreadable` at `view.go:526`; and `E` still routes through
`selectedPath()` (`model.go:1449`) into `openInEditor()` (`model.go:1496`), which is
exactly why it cannot reach a file that never became a row. `internal/editor` is
unchanged.

**Scope has GROWN since 2026-06-24 — two things a design should now account for.**

1. **Two unreadable channels, not one.** Research joined the `problems` pattern
   (`commands.go:272`), but Threads did **not**: `thread_projection.go:26,49` carries its
   own `readProblems` field with a separate footer string at `view.go:530`
   (`! %d unreadable Thread record(s)`). A "problems view" built only on `tab.problems`
   would silently skip malformed Thread documents — and those are precisely the ones
   `store`'s filename-only `ThreadPathSource` resolver was written to keep findable for
   repair, so a path to open them does exist.
2. **A third surface already shows the count.** The dashboard prints
   `%d unreadable file(s) (run lint)` (`dashboard.go:209`). Whatever affordance this task
   picks, the dashboard is a natural launch point for it — the dashboard is a *launch*
   surface by design and `dashJump` already exists.

Not a shipped criterion, but worth knowing: `thread_spatial.go:455` and `detail.go:141`
already reason about unreadable *graph nodes* being inspectable and yankable-but-not-
always-safe-to-open. That is a separate axis (graph records, not files that failed to
parse), but the "copy the id even when you cannot open the thing" precedent is directly
reusable for criterion 3.

Acceptance criteria left untouched — none is demonstrably met.

## Progress log

- 2026-09-13: automated weekly sweep — every citation re-verified and still exact; scope grew, since Threads now carry a second, separate unreadable channel (`readProblems`) and the dashboard shows a third count.
