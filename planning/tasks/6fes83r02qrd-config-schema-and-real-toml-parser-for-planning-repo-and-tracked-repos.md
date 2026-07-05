---
schema: 1
status: completed
epic: 23-point-an-impl-repo-at-an-external-planning-repo
description: Swap the hand-rolled one-key TOML scanner for a real parser; widen Config with PlanningRepo + TrackedRepos. Foundation for the epic.
effort: Unknown
tier: 3
priority: high
autonomy_level: 3
tags: [config, discovery]
created: "2026-06-22"
updated_at: "2026-06-22"
started_at: "2026-06-22"
completed_at: "2026-06-22"
id: 6fes83r02qrd
---
# Config schema + real TOML parser

The foundation for the whole epic. Today `internal/config/config.go` hand-rolls a
one-key line scanner (`taskflowRoot`/`tomlStringValue`) and deliberately refuses
anything it can't read losslessly — fine for a single string, but it can't read a
string **array**. This epic adds two new keys (`planning_repo` scalar +
`tracked_repos` array), so swap the scanner for a real parser and widen `Config`.

## Scope

1. Add a TOML dependency (`BurntSushi/toml` or `pelletier/go-toml/v2`) to
   `go.mod` — neither is present today.
2. Replace `taskflowRoot`/`tomlStringValue` with a real decode of
   `.tskflwctl.toml` into a struct: `taskflow_root` (string, default "."),
   `planning_repo` (string, optional), `tracked_repos` ([]string).
3. Widen `config.Config` from `{Root}` to also carry `PlanningRepo` and
   `TrackedRepos` (raw, unresolved — discovery resolves them).
4. Keep the default config template (`defaultConfigTOML`) honest: `tracked_repos`
   stops being "reserved/not read."

## Acceptance criteria

- [ ] A real TOML parser reads all three keys; malformed TOML is a loud error,
      not a silent default.
- [ ] `Config` exposes `PlanningRepo` + `TrackedRepos`.
- [ ] Existing in-tree `taskflow_root` behavior is byte-for-byte unchanged.
- [ ] Suite + lint green.

## Related

- Foundation for the rest of [[23-point-an-impl-repo-at-an-external-planning-repo]].
- Unblocks discovery (next task).
