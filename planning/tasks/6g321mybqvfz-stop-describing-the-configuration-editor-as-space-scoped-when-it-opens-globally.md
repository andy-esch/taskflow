---
schema: 1
id: 6g321mybqvfz
status: ready-to-start
epic: 20-cli-ux-and-ergonomics
description: The config editor defaults to USER scope, but the TUI help says it edits the current space and neither help mentions the scope-switch key.
effort: XS
tier: 3
priority: medium
autonomy_level: 3
tags: [cli, tui, config, discoverability]
created: "2026-08-23"
updated_at: "2026-08-23"
---
# Stop describing the configuration editor as space-scoped when it opens globally

## Objective

Global configuration editing **already exists** and works. `config edit` says so plainly —
"User scope is the default; repository overrides must be selected explicitly" — the TUI's
`:config` opens the same `configui` editor with the same user-scope default, and `s`/`tab`
switches scope inside it. `config show` renders both scopes with provenance.

Nobody can tell, because the surfaces describe it wrongly:

| Where | Says | Actually |
| --- | --- | --- |
| `internal/tui/help.go:37` | ":config — open Configuration / About **for this space**" | opens in **user** scope |
| `internal/tui/help.go:77` | ":config — … **for the current space**" | opens in **user** scope |
| dashboard row | "Configuration / About →" | gives no hint scope exists |
| `config edit --help` | lists only `--help` | never mentions `s`/`tab` switches scope |

So the TUI actively tells you the editor is space-scoped when it is the opposite, and the
one key that reveals the other scope is undocumented everywhere. Reported 2026-08-23 as
"we don't have a global configuration TUI editing screen or even CLI option" — by the
person who built it, which is about as strong a discoverability signal as exists.

## Acceptance criteria

- [ ] The TUI help and the dashboard row describe the editor by what it does: preferences,
  user scope by default, repository override available — not "for this space".
- [ ] The scope-switch key is documented where a reader will meet it: the `?` help, the
  editor's own chrome, and `config edit --help`.
- [ ] The editor shows which scope is active prominently enough that it cannot be mistaken
  after a switch, including in the notice it prints on save.
- [ ] `config edit --help` states the default scope and how to change it, so the CLI is
  self-describing without launching the editor.
- [ ] A test asserts the help text names user scope, so the two cannot drift apart again.

## Out of scope

- Any change to scope precedence, the editable field set, or the editor's layout.
- A separate top-level `:preferences` screen. The editor exists; this is about naming it
  honestly, not adding a second door.

## Related

- Epic [20-cli-ux-and-ergonomics](../epics/20-cli-ux-and-ergonomics.md)
- Surfaces: `internal/configui/editor.go`, `internal/tui/help.go`, `internal/cli/config.go`
