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
