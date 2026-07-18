---
schema: 1
id: 6fq9t7y8gvy5
status: ready-to-start
epic: 20-cli-ux-and-ergonomics
description: huh.Select's Filtering(true) shows the / filter line instead of the picker title, so epic/audit/task pickers render unlabeled (unlike the tags Input).
effort: Unknown
tier: 3
priority: low
autonomy_level: 3
tags: [cli, prompt]
created: "2026-07-18"
---
# Picker title hidden in filter mode

## Objective

Every interactive picker passes a title — e.g. `fillSelect(..., "Epic for this
task", ...)` at `internal/cli/task.go:122` — but `SelectOne` builds the
`huh.Select` with `Filtering(true)` (`internal/cli/prompt/tty.go`), and a
filtering Select renders its `/` filter-input line *in place of* the title, even
at rest. So the epic / audit / task pickers render unlabeled: the user sees a
list with no statement of what's being chosen. The tags prompt is a plain
`huh.Input` (no filtering), which is why it alone shows its `"Tags
(comma-separated)"` label. Surfaced by the `assets/picker.gif` demo.

## Acceptance criteria

- [ ] A filtering picker shows its title/label alongside (not replaced by) the filter input
- [ ] The epic (`task new`), audit, and task pickers all state what's being picked
- [ ] The tags text prompt is unaffected

## Notes

- Fix locus: `SelectOne` in `internal/cli/prompt/tty.go` — the `huh.Select` config.
- Likely a `Description`/header on the field, or a huh option that keeps the
  `Title` visible in filter mode.

## Related

- Epic [20-cli-ux-and-ergonomics](../epics/20-cli-ux-and-ergonomics.md)
