---
schema: 1
status: completed
epic: 20-cli-ux-and-ergonomics
description: 'Bring epic list to parity with the other list commands: -o/--output, -c/--columns, -q ids-only, and a filter so triage isn''t all epics at once.'
effort: Unknown
tier: 3
priority: medium
autonomy_level: 3
tags: [cli, output, epic]
created: "2026-06-22"
updated_at: "2026-06-22"
started_at: "2026-06-22"
completed_at: "2026-06-22"
id: 6fes83r00zxx
---
# `epic list` filter + projection parity

**Source.** Product feedback (2026-06-22) from an AI agent driving `tskflwctl`.
The other list commands got `-o/--output`, `-c/--columns`, and `-q` (ids-only)
via the completed [consolidate-output-flags-into-output-and-columns](6fdtbb401htr-consolidate-output-flags-into-output-and-columns.md) and
[pipeline-output-modes-q-plain-stderr-discipline](6fbj870023xj-pipeline-output-modes-q-plain-stderr-discipline.md) — but `epic list` was left
behind. So an agent that reaches for `epic list` (instead of the right
`epic show <id>`) pays for **all** epics' full rollup with no way to project or
filter. The agent's note: "`epic show` was exactly right, but if an agent does
reach for `epic list` there's no `--id`/`--columns`/`-q`."

## Scope

1. Bring `epic list` to parity with the other list commands:
   - `-o/--output` format axis (table/csv/json),
   - `-c/--columns` projection over an epic column registry
     (`id,status,percent,description,…`),
   - `-q` ids-only for piping.
2. A filter to narrow the set (e.g. `--status`, and/or an `--id` selector) so
   triage doesn't always materialize every epic.

## Acceptance criteria

- [ ] `epic list -o table -c id,status,percent,description` projects those columns.
- [ ] `epic list -q` emits ids only, one per line (pipe-friendly).
- [ ] A status/id filter narrows the result set.
- [ ] Shares the output-mode plumbing with the other list commands (no new axis).
- [ ] Suite + lint green; docs/cli regenerated (docs-check gate).

## Related

- Parity with [consolidate-output-flags-into-output-and-columns](6fdtbb401htr-consolidate-output-flags-into-output-and-columns.md).
- Epic [20-cli-ux-and-ergonomics](../epics/20-cli-ux-and-ergonomics.md).
