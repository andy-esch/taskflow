---
schema: 1
id: 6fq9t7y8gvy5
status: ready-to-start
epic: 20-cli-ux-and-ergonomics
description: Pickers of 8+ options hide their title behind huh's / filter line, so epic and task pickers render unlabeled; short menus are already fixed.
effort: Unknown
tier: 3
priority: low
autonomy_level: 3
tags: [cli, prompt]
created: "2026-07-18"
updated_at: "2026-09-20"
audited: "2026-09-20"
audit_sources: [2026-09-20-weekly-task-sweep]
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

## Sweep verification (2026-09-20)

Automated weekly sweep re-read this task against `internal/cli/prompt` and
`internal/cli` at `934e1cf`. **The bug has partially shipped a fix** and the two
code references in the Objective are both out of date. The Objective is left
intact as the original report; read it with the corrections below.

> **Update 2026-09-20 — `Filtering(true)` is no longer unconditional.**
> `SelectOne` (now the method `(p ttyPrompter) SelectOne`,
> `internal/cli/prompt/tty.go:80`) builds its `huh.Select` with
> `Filtering(filtering)` (`tty.go:96`), where `filtering` comes from
> `selectLayout(n)` (`tty.go:72`) and is `n >= selectFilterMin` with
> `selectFilterMin = 8` (`tty.go:56`). That change landed via
> `490d6a6` (2026-09-04) and was motivated by filter noise on short menus, not by
> this task — but it fixes this task's symptom for every picker of fewer than 8
> options as a side effect: with filtering off, huh renders the `Title` normally.

**What is left of the bug.** Only pickers that cross the 8-option threshold
still render unlabeled. In this repo that is the epic picker (16 epics) and the
task picker (68 active tasks); the audit picker is over or under depending on the
corpus. So the defect is real but now scope-limited to long lists — which are
exactly the lists where a filter input is present and the user most needs to know
what is being filtered.

**Reference corrections:**

- `fillSelect(..., "Epic for this task", ...)` is at `internal/cli/task.go:137`,
  not `:122`; `:122` now lands in `newTaskNewCmd`'s declaration block. The
  `fillSelect` helper itself is `internal/cli/fill.go:35`, with call sites in
  `task.go:137`, `edit.go:50`, `audit.go:473`, `audit.go:516`, and
  `research.go:115` — note `research.go` is a picker this task predates and does
  not mention.
- `SelectOne` is a method on `ttyPrompter`, not a package-level function; the
  package exports only `NewTTY` and `NewGate`.
- The tags prompt premise is verified accurate: `Text`
  (`internal/cli/prompt/tty.go:115`) uses a plain `huh.NewInput()` with no
  filtering, so AC #3 is unaffected by any fix here.

Acceptance criteria left untouched: the fix has not been made, only narrowed —
AC #2's "epic, audit and task pickers" is still the right target set (plus
research), and none of the three boxes is demonstrably met.

## Progress Log

- 2026-09-20: automated weekly sweep — bug partially fixed as a side effect of 490d6a6 (filtering now only at 8+ options), so it survives only for long pickers; corrected the task.go line reference and noted SelectOne is now a ttyPrompter method.
