---
schema: 1
status: completed
epic: 20-cli-ux-and-ergonomics
description: Deferring records no revisit date — items just sit in deferred/ indefinitely. Add task defer --until <date> (a snooze/revisit field) and surface deferred items whose date has arrived in status.
effort: M
tier: 3
priority: medium
autonomy_level: 3
tags: [cli]
created: "2026-06-26"
updated_at: "2026-06-26"
started_at: "2026-06-26"
completed_at: "2026-06-26"
id: 6fg2ef801c1m
---
# Set a revisit date when deferring (snooze) and surface what's due

## Objective

Deferring parks an item with **no signal for when to return to it** — `task defer
<slug>` just drops it into `deferred/` indefinitely. Make `defer` mean "snooze until
X": capture an optional **revisit date** at defer time, and surface deferred items
whose date has arrived, so deferred work resurfaces instead of being forgotten.

## Context

- `task defer <slug>` moves a task to `deferred/` (status == directory); nothing
  records *when* to revisit. Tasks carry `created` / `updated_at` but no
  revisit/snooze field.
- Audits also have a `deferred` bucket (`audit defer`); epics have `deprecated`.
  Scope tasks first; consider audits (and whether epics ever "snooze").
- The `status` dashboard already shows active counts + in-progress + epics + open
  audits — a natural place to nudge "N deferred due to revisit."
- Surfaced 2026-06-26 right after deferring
  [[make-tui-clickable-links-work-in-tmux-ghostty-plus-document-recipe]].

## Acceptance criteria

- [ ] `task defer <slug> [--until <date>]` records a revisit-date frontmatter field
      (name TBD — see Decisions); the date is validated. Defer WITHOUT the flag stays
      indefinite, exactly as today (the field is optional).
- [ ] The date is surfaced: `task list --status deferred` shows it, AND deferred
      tasks whose revisit date is ≤ today appear as a nudge in `status` (e.g.
      "2 deferred due to revisit") — decide the exact surface.
- [ ] `task set` / `edit` can change or clear it (un-snooze); `--json` carries it
      (schema bump if a task-envelope field is added).
- [ ] go build/test/lint green; docs/cli regenerated (new flag); README note.

## Decisions to settle

- **Field name:** `revisit_at` vs `snooze_until` vs `defer_until`.
- **Date input:** absolute `YYYY-MM-DD` to start, or also relative (`--until 2w` /
  `+30d`) like other date conveniences? Absolute first is simplest.
- **Surface:** a `status` nudge, a dedicated `task list --due` / `--overdue` filter,
  or both?
- **Arrived date behavior:** AUTO-promote back to ready-to-start, or just NUDGE?
  (Recommend nudge — don't move the file out from under the user; they decide.)
- **Scope creep:** audits (`audit defer --until`) and epics — tasks first.

## Risks / gotchas

- The tool is otherwise time-agnostic in tests — thread the clock the way `updated_at`
  is set so "is it due?" comparisons against today and the tests stay deterministic.
- Keep it a DISTINCT intent field — don't overload `created`/`updated_at`.
- Lint: a revisit date on a non-deferred task is probably stale — decide whether lint
  flags / strips it (and whether promoting/starting a task clears it).

## Related

- Epic [[20-cli-ux-and-ergonomics]].

## Done when

`task defer <slug> --until 2026-09-01` snoozes it with a recorded date, and when that
date arrives the dashboard nudges you to revisit it — defer becomes snooze.
