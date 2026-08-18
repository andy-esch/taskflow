---
schema: 1
id: 6g0fzhc1235a
status: next-up
epic: 29-multi-space-planning-a-home-registry-and-the-atlas
description: 'Widen LinkProblem into a SpaceProblem shared by doctor and the TUI: missing/moved/not-a-repo/unreadable/empty each get a diagnosis and a remedy. Never fatal, never auto-forgets.'
effort: Unknown
tier: 3
priority: medium
autonomy_level: 3
tags: [cli, multi-repo]
created: "2026-08-15"
---
# Space health diagnosis and a `doctor` registry section

## Objective

Registered paths rot — repos move, get deleted, get un-inited. Make that a **kind
diagnosis with a remedy**, never a bad error, and put the logic in one place both the
CLI and (later) the TUI read, so the two can never disagree about what is wrong.

This is the "delightful missing config" requirement of epic 29, and it is a
generalization of machinery that already exists: epic 23's `LinkProblem` →
`CheckLinks` → `doctor` is the same pattern one scope down.

## Notes

Widen `LinkProblem` into a `SpaceProblem{ID, Path, Kind, Message, Remedy}`, resolved
lazily per entry:

| state | detection | message | remedy |
| --- | --- | --- | --- |
| `ok` | discovery succeeds | — | — |
| `missing` | path gone | "not found at ~/…" | forget · repoint |
| `moved` | gone, but a same-named planning repo nearby | "moved to …?" | accept |
| `not-a-repo` | exists, no marker | "no `.tskflwctl.toml` — init here?" | init · forget |
| `unreadable` | TOML / permission error | verbatim parse error | `$EDITOR` · forget |
| `empty` | valid, zero entities | "no tasks yet" | — |

Three principles separate "delightful" from "nagging":

- **A broken entry is never fatal** — it is reported alongside the healthy ones, never
  an error screen and never a blocked command. Same philosophy `lint` already applies
  one scope down: a task with an unrecognized status is *listed and flagged*, never
  moved or dropped.
- **Never auto-forget.** Removal is always explicit.
- **One diagnosis function** behind every surface.

`moved` detection is the questionable one — guessing where a repo went is the kind of
cleverness that misfires. Ship without it if it doesn't feel obviously right; "missing
+ repoint" may be honest enough.

## Acceptance criteria

- [ ] One exported diagnosis function returns typed `SpaceProblem`s for a registry
- [ ] `doctor` gains a registry section: human + `--json` + the existing nonzero-exit
      CI contract, without duplicating the linkback logic
- [ ] Every state in the table has a test with a real temp-dir fixture
- [ ] A broken entry never blocks `space list` or any other command
- [ ] Nothing auto-removes an entry
- [ ] `just test` + `just lint` green

## Out of scope

- TUI rendering of the states (later, if the board proceeds)
- Automatic repair of any kind

## Related

- Epic [29-multi-space-planning-a-home-registry-and-the-atlas](../epics/29-multi-space-planning-a-home-registry-and-the-atlas.md)
- Sketch: [2026-08-15-multi-space-home-registry-and-the-atlas](../research/2026-08-15-multi-space-home-registry-and-the-atlas.md)
- The pattern this generalizes: epic [23-point-an-impl-repo-at-an-external-planning-repo](../epics/23-point-an-impl-repo-at-an-external-planning-repo.md)
