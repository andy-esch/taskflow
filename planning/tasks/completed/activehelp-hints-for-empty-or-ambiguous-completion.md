---
status: completed
epic: 20-cli-ux-and-ergonomics
description: cobra ActiveHelp hints where completion is empty/ambiguous (title/area on new, status on task move); degrades gracefully off bash V2
effort: Unknown
tier: 3
priority: medium
autonomy_level: 3
tags: [cli, dx]
created: "2026-06-19"
started_at: "2026-06-19"
updated_at: "2026-06-19"
completed_at: "2026-06-19"
---
## Objective

Use cobra [ActiveHelp](https://github.com/spf13/cobra/blob/main/site/content/active_help.md)
to print short guidance during shell completion where the normal completion system
has nothing useful to offer — the recommended use case. Small, additive, and it
degrades gracefully (ActiveHelp shows only on bash V2 / 4.4+; elsewhere it's just
silently absent). The agent-safe, non-TTY complement to the planned huh pickers.

## Scope (where the hints help)

- `task new` / `epic new` / `audit new` — the positional arg is a free-form
  title/area; today completion falls back to filenames (useless). Add a
  `ValidArgsFunction` that emits `NoFileComp` + an ActiveHelp line ("provide a
  task title — quote if it has spaces" / "provide an area, e.g. dispatcher").
- `task move <task>... <status>` — when ≥1 task is typed, the trailing arg is a
  status; append an ActiveHelp line noting that (the status values are already
  offered).

Gate each behind `cobra.GetActiveHelpConfig(cmd) != "off"` so users can disable.

## Acceptance criteria

- [ ] `__complete task new ''` (and epic/audit new) returns an `_activeHelp_`
      hint and `NoFileComp` (no filename completion).
- [ ] `__complete task move <task> ''` includes the status hint.
- [ ] Existing completion tests stay green (hints don't disturb the slug/status
      candidate assertions).

## Out of scope

- Rephrasing every command's help; this is completion-time hints only.

## Related

- Epic [[20-cli-ux-and-ergonomics]]
- Complements [[interactive-prompt-layer-gh-style-pickers]] (the TTY picker face).
