---
schema: 1
id: 6g0fz5c2x5t4
status: next-up
epic: 29-multi-space-planning-a-home-registry-and-the-atlas
description: 'A user-scoped config.toml under XDG (not os.UserConfigDir), env-overridable for tests, adding the tier theme/pager always wanted: flag > env > repo > home > default. Useful alone.'
effort: Unknown
tier: 3
priority: medium
autonomy_level: 3
tags: [config, cli]
created: "2026-08-15"
---

# Home-level user config: XDG location, env override, theme/pager precedence

## Objective

`[theme]` and `[pager]` are documented in `internal/config/config.go` as
"local-terminal concerns," yet today they can only be set in a *repo* config — so a
preference about your own terminal must be repeated per project, and a shared
planning repo carries one contributor's taste. Give them a user-level tier.

This is the **independently useful** first slice of epic 29: it commits to nothing
about spaces, but it builds and proves the home-config plumbing everything else in
the epic would sit on. Worth doing on its own merits even if the rest is dropped.

## Notes

- `$XDG_CONFIG_HOME/tskflwctl/config.toml`, falling back to
  `~/.config/tskflwctl/config.toml`. **Not** Go's `os.UserConfigDir()` — it returns
  `~/Library/Application Support` on darwin, wrong for a dotfile-friendly CLI whose
  users commit their config.
- The env override is **load-bearing, not a nicety**: the suite is `t.TempDir()`
  -disciplined and nothing in CI may read or write a real `$HOME`. Design it in from
  the start.
- Precedence becomes flag > env > repo config > **home config** > built-in default —
  a small change to `themeName()` and the pager resolution.
- Writes (when they come, in the registry task) should match
  `setTrackedReposInText`'s discipline: atomic + surgical, preserving comments, key
  order, and unknown keys.
- A missing or unreadable home config must **degrade quietly to today's behavior**,
  never error — it is a preferences file, not a marker.

## Acceptance criteria

- [ ] Home config resolves via `XDG_CONFIG_HOME` → `~/.config/tskflwctl/config.toml`,
      with an env override for the whole location
- [ ] `[theme].name` and `[pager]` in the home config are honored at the documented
      precedence, verified by a table test over all five tiers
- [ ] No test reads or writes a real `$HOME` (assert via the override)
- [ ] Missing / malformed home config degrades to current behavior without erroring
- [ ] `just test` + `just lint` green

## Out of scope

- The `[[space]]` registry itself (separate task) — this lands the file and its
  precedence only
- Any change to repo-level discovery

## Related

- Epic [29-multi-space-planning-a-home-registry-and-the-atlas](../epics/29-multi-space-planning-a-home-registry-and-the-atlas.md)
