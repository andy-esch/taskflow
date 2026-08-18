---
schema: 1
id: 6g0fzhbtk0gy
status: next-up
epic: 29-multi-space-planning-a-home-registry-and-the-atlas
description: 'Add the [[space]] array to the home config plus space list|add|forget. Registry stays advisory: nothing in it may change what Discover resolves from a cwd. Schema is provisional.'
effort: Unknown
tier: 3
priority: medium
autonomy_level: 3
tags: [config, cli, multi-repo]
created: "2026-08-15"
---
# Space registry model and the `space` verbs (list/add/forget)

## Objective

Add the `[[space]]` array to the home config plus the CLI surface to manage it: the
user-level record of which planning repos exist on this machine. Names and schema
are **provisional** — see the epic's sketch; this task is where they get tested
against reality.

## Notes

Sketch schema (expect it to move):

```toml
[[space]]
id     = "taskflow"                    # stable key; how you address it
path   = "~/git/andy-esch/taskflow"    # the repo dir, NOT the resolved planning root
label  = "tskflwctl"                   # optional display name
accent = "cyan"                        # optional; else derived from id
added  = "2026-08-15"
```

- **Anchor `path` at the repo** (the `.tskflwctl.toml` dir), not the resolved root:
  registration then means one thing in both config modes, and a pointer repo resolves
  through its `planning_repo` to the real root via ordinary discovery. Epic 23
  compatibility falls out for free.
- **`id` is a stable key** (ADR-0003's philosophy), defaulted from the dir basename.
  Collisions resolved at registration, never silently.
- **Dedup on PHYSICAL paths**, reusing the helper `tracked_repos` already dedups with,
  so `../x`, an absolute path, and a symlinked checkout collapse to one entry.
- **Store `~` unexpanded** so the file stays portable and committable.
- **Atomic + surgical writes** — `setTrackedReposInText` is the model.
- `space add` validates through discovery **before** writing anything (the
  "require + error, leave nothing behind" contract `InitPointer` already follows).
- `space forget` never touches the repo on disk — it only drops the entry.

The invariant worth protecting: **the registry is advisory.** Nothing in it may
change what `Discover` resolves from a given cwd. No home config ⇒ today's behavior
exactly; deleting the file costs convenience, never data or addressability.

## Acceptance criteria

- [ ] `space list` (human + `--json` with `schema_version`) shows id, path, state
- [ ] `space add [path]` validates via discovery before writing; a bad path errors
      with nothing left behind
- [ ] `space forget <id>` removes only the entry; the repo is untouched
- [ ] Physical-path dedup: `../x`, an absolute path, and a symlink resolve to one entry
- [ ] Surgical write: comments, key order, and unknown keys survive a round trip
- [ ] Local discovery behavior is provably unchanged with and without a registry
- [ ] `just test` + `just lint` green

## Out of scope

- `--space` (separate task), `init` auto-registration (separate task)
- `space scan` — walking the filesystem for planning repos
- Any TUI surface

## Related

- Epic [29-multi-space-planning-a-home-registry-and-the-atlas](../epics/29-multi-space-planning-a-home-registry-and-the-atlas.md)
- Sketch: [2026-08-15-multi-space-home-registry-and-the-atlas](../research/2026-08-15-multi-space-home-registry-and-the-atlas.md)
