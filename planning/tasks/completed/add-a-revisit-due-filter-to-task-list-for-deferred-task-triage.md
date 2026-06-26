---
schema: 1
status: completed
epic: 20-cli-ux-and-ergonomics
description: Focused query for deferred tasks due for revisit — task list --revisit-due (composes with --json/-q/-c/--epic/--tag) — instead of the status dashboard nudge.
effort: S
tier: 3
priority: medium
autonomy_level: 3
tags: [cli]
created: "2026-06-26"
updated_at: "2026-06-26"
started_at: "2026-06-26"
completed_at: "2026-06-26"
---

# Add a --revisit-due filter to task list for deferred-task triage

## Objective

Give "deferred tasks that are up for revisit" a focused, composable query instead
of relying on the `status` dashboard nudge — which is a *count* buried among
misfiled/epics/counts (a human-glance affordance, the wrong shape for an agent or
a script). The terse machine path should be `task list --revisit-due --json`.

This is a filter on `task list`, NOT a new command: `task list --json` already *is*
the API, so a predicate gets us the feature while reusing projection (`-c`), output
modes (`-o table|csv|json`), quiet slugs (`-q`), and composition with other filters
for free. A standalone command would have to re-expose all of that and duplicate the
render/JSON paths the codebase keeps in one place.

## Acceptance criteria

- [ ] `task list --revisit-due` lists only deferred tasks whose `revisit_at` is on
      or before today (reuse `domain.IsRevisitDue`); the flag implies `status=deferred`,
      since the invariant is `revisit_at present ⟺ deferred`.
- [ ] Implemented as a predicate in `core.TaskFilter` / `ListTasks` — no new command,
      no duplicated rendering or JSON.
- [ ] Composes unchanged with `--epic`, `--tag`, `-c`, `-o table|csv|json`, `-q`.
- [ ] Machine path: `task list --revisit-due --json` returns the standard tasks
      envelope; `-q` emits bare slugs, so
      `task list --revisit-due -q | xargs tskflwctl task next` resumes everything due.
- [ ] The "due" cutoff uses an injected/deterministic clock (mirroring `IsRevisitDue`)
      so tests don't depend on the wall clock.
- [ ] `docs/cli` regenerated for the new flag. No `schema_version` bump — the output
      shape is unchanged (it's a filter, not a new field/envelope).
- [ ] The `status` nudge stays the human glance; optionally repoint its hint at the
      new query.

## Out of scope

- A dedicated `deferred`/`revisit` command *group* — deferred is a task *status*, not
  a separate entity (unlike audits); this is one filter on `task list`.
- Auto-resuming due tasks — the snooze only lists/nudges; resume stays manual
  (`task next`/`task ready`).
- Changing how `revisit_at` is set or cleared — already shipped (set on defer, cleared
  on leaving deferred).

## Possible extensions (not required for this task)

- `--revisit-by <date>` (default today) for "what's coming due", not just "due now".
- A thin sugar verb `task revisit` ⇒ `task list --revisit-due` with deferred-friendly
  default columns, implemented as a shell over the same list path — only if a named
  endpoint earns its keep.

## Related

- Epic [[20-cli-ux-and-ergonomics]]
- Follows [[set-a-revisit-date-when-deferring-snooze-and-surface-what-is-due]]
