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
updated_at: "2026-10-04"
audited: "2026-10-04"
audit_sources: [2026-10-04-weekly-task-sweep]
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

## Sweep verification (2026-10-04)

Every row of the Objective's table re-checked against current code; all four still
accurate, unchanged since the task was written:

- `internal/tui/help.go:37` — still `{":config", "open Configuration / About for this space"}`
- `internal/tui/help.go:77` — still `{":config", "open Configuration / About for the current space"}` (the atlas help section)
- the editor still defaults to user scope — `internal/configui/editor.go:86` initialises
  `scope: core.ConfigScopeUser`, and `editor.go:50` states the intent outright:
  *"choosing repository scope always requires an explicit key press"*; the switch is
  `editor.go:196-199`
- `config edit --help` still lists `--help` as its only command-specific flag

One narrowing for the fourth acceptance criterion, which asks that `config edit --help`
state "the default scope and how to change it": **the default scope is already stated.**
The long help reads *"User scope is the default; repository overrides must be selected
explicitly."* What is missing is only the second half — which key performs that explicit
selection (`s`/`tab`). So that criterion is partially satisfied already and the remaining
work there is one clause, not a rewrite. Left unticked: it is not fully met, and deciding
how to re-word a criterion is a human call.

2026-10-04: automated weekly sweep — all four table rows re-verified unchanged
(`help.go:37`/`:77`, `configui/editor.go:86`, `config edit --help`); noted AC4's
default-scope half is already shipped, leaving only the scope-switch key undocumented.
