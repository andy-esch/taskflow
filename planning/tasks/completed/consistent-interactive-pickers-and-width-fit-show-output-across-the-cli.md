---
schema: 1
status: completed
epic: 20-cli-ux-and-ergonomics
description: Bare show/audit-verb/set/append pick a target on a TTY (like task edit); show meta + trees truncate to terminal width.
effort: M
tier: 3
priority: medium
autonomy_level: 3
tags: [cli, tui]
created: "2026-06-25"
updated_at: "2026-06-25"
started_at: "2026-06-25"
completed_at: "2026-06-25"
---
# Consistent interactive pickers + width-fit show output across the CLI

## Objective

Two cross-command inconsistencies surfaced by a 2026-06-25 ecosystem audit:

1. **Pickers on a missing target.** The bubbles/list picker (`prompt.Prompter` +
   `fillSelect`) backs `task edit`, the `task` lifecycle verbs, `task new`
   (`--epic`/`--tags`), and `init` — but NOT `task/epic/audit show`, the `audit`
   lifecycle verbs (`close/reopen/defer`), or `task set/append`. Those error on a
   bare invocation instead of offering a picker on a TTY.
2. **Width-fit `show` output.** `Style.width` (from `terminalWidth`) is honored by
   list tables and the `show` markdown body, but NOT by `show`'s metadata fields
   or its epic/audit trees — so a long `description:` or finding title overflows a
   narrow terminal.

## Acceptance criteria

- [ ] Bare `task/epic/audit show`, `audit close/reopen/defer`, and `task set/append`
      pick a target on a TTY (mirroring `task edit`/the task verbs); non-interactive
      (piped / --json / --no-input) still errors with exit 11 + a clear message.
- [ ] `show` metadata values and the epic/audit trees truncate to the terminal
      width on a TTY; piped output stays full (lossless), matching `writeTable`.
- [ ] go build/test/lint green; docs/cli regenerated if examples change.

## Notes

- The picker is a bubbles/list one (huh.Select was swapped out for filter reasons).
- `tree.Tree` has `.Width(n)`; the meta `field` helper truncates the value to
  `width − labelWidth − 1` via the existing `truncate`/`visibleWidth` helpers.
- Non-interactive behavior is the gate's job (`needPrompt`), unchanged.
