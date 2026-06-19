---
status: completed
epic: 20-cli-ux-and-ergonomics
description: Wrap the create-confirmation path in an OSC 8 file:// hyperlink (click-to-open), gated on color+TTY; raw under --json/pipe
effort: Unknown
tier: 3
priority: medium
autonomy_level: 3
tags: [cli, ux]
created: "2026-06-19"
started_at: "2026-06-19"
updated_at: "2026-06-19"
completed_at: "2026-06-19"
---
## Objective

Make the path in the create confirmation click-to-open: wrap it in an
[OSC 8](https://gist.github.com/egmontkob/eb114294efbcd5adb1944c9f3cb5feda)
terminal hyperlink to a `file://` URL, so on supporting terminals you can
cmd/ctrl-click the just-created file to open it. Strictly a human-face nicety —
gated on color being on (TTY), so piped / `--color=never` / `--json` output is
byte-identical plain text. gh does this.

## Design

- `Style.Link(text, url) string` in `render` — wraps `text` in OSC 8 when
  `Style.on`, else returns `text` verbatim (same gate as the other styling).
- `App.linkPath(absPath) string` — renders the absolute path as a relative
  display string linked to `file://<abs>` (relative for readability, absolute in
  the URL so the terminal can resolve it).
- Apply at the three `task/epic/audit new` confirmations (the `created`/`would
  create` line). `--json` keeps the plain relative path in the envelope.
- `file://` scheme only — no `:line` anchor (OS openers don't parse line
  suffixes reliably; editor schemes are terminal/editor-specific).

## Acceptance criteria

- [ ] `Style.Link` emits a valid OSC 8 sequence when on, plain text when off
      (unit test).
- [ ] `task new … --color=always` output contains the OSC 8 link; `--color=never`
      output is the plain relative path (byte-stable).
- [ ] `--json` create envelope is unchanged (plain relative path).

## Out of scope

- Linking lint/fix/problems paths and `show` (possible follow-on; this is the
  create confirmation only).
- Editor-scheme / line-anchored links.

## Related

- Epic [[20-cli-ux-and-ergonomics]]
