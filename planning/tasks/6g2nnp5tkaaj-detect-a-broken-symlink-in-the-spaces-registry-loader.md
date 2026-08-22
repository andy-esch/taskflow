---
schema: 1
id: 6g2nnp5tkaaj
status: ready-to-start
epic: 29-multi-space-planning-a-home-registry-and-the-atlas
description: Port userconfig.go's Lstat probe into spaces.go so an unmounted dotfiles symlink reports an error instead of an empty registry.
effort: XS
tier: 3
priority: medium
autonomy_level: 3
tags: [userconfig, integrity, bug]
created: "2026-08-22"
updated_at: "2026-08-22"
---
# Detect a broken symlink in the spaces registry loader

## Objective

`internal/userconfig/userconfig.go` deliberately separates "no config file" from "a
symlink whose target is gone": on `ErrNotExist` it re-probes with `os.Lstat` and returns
an actionable "broken symlink -> <target>" error, because this audience commits its
config and symlinks it out of a dotfiles repo that may not be mounted.

`internal/userconfig/spaces.go:readRegistry` uses a bare `os.ReadFile` and treats every
`ENOENT` as an un-initialized registry, returning `initialSpacesText, nil, nil`. A
`spaces.toml` symlinked to an unmounted volume therefore reports zero registered spaces
rather than a fault. This predates the atlas, but the atlas raised its cost: the failure
mode is now a TUI that says "No spaces registered. Run `tskflwctl space add <path>`" and
invites the user to re-register spaces they already have — and `space add` would then
write a fresh registry over the dangling link.

## Acceptance criteria

- [ ] `readRegistry` performs the same `os.Lstat` probe on the `ENOENT` path and returns a
  broken-symlink error naming the target.
- [ ] A genuinely absent `spaces.toml` still degrades silently to the empty registry.
- [ ] The error reaches `space list`, `doctor`, `status --all`, and the TUI atlas as a
  failure rather than an empty list.
- [ ] Tests cover absent file, broken symlink, and healthy symlink.

## Out of scope

- Auditing every other `os.ReadFile` in the tree for the same pattern; do that only if a
  second real instance shows up.

## Related

- Epic [29-multi-space-planning-a-home-registry-and-the-atlas](../epics/29-multi-space-planning-a-home-registry-and-the-atlas.md)
- Audit finding H3: [2026-08-22-multi-workspace-atlas](../audits/6g2k3qye4qma-2026-08-22-multi-workspace-atlas.md)
