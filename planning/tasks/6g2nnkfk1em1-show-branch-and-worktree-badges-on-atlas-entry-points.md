---
schema: 1
id: 6g2nnkfk1em1
status: ready-to-start
epic: 29-multi-space-planning-a-home-registry-and-the-atlas
description: Carry DescribeCheckout's branch and IsWorktree into core.SpaceEntryPoint so sibling worktrees of one repo are distinguishable in the atlas.
effort: S
tier: 3
priority: medium
autonomy_level: 3
tags: [tui, atlas, worktree, spacehealth]
created: "2026-08-22"
updated_at: "2026-08-22"
---
# Show branch and worktree badges on atlas entry points

## Objective

`config.DescribeCheckout(dir)` already reports a checkout's branch and whether it is a
linked worktree; `internal/cli/workspace.go` feeds both into `wire.WorkspaceJSON`.
`spacehealth.DiagnoseSpace` and `renderAtlasCard` ignore it, so two registered worktrees
of the same repo render identically (`● id  direct  ~/path`) with nothing but the path to
tell them apart — exactly the case that produced multiple entry points per identity.

## Acceptance criteria

- [ ] `spacehealth.SpaceProblem` carries the branch and worktree flag from
  `DescribeCheckout`, and `core.SpaceEntryPoint` exposes them as typed fields.
- [ ] The atlas entry-point rows render the branch name and a worktree marker; a base
  checkout with no branch (detached, non-git) renders without inventing one.
- [ ] `space list` / registry JSON either carry the same fields or deliberately do not,
  recorded in the task notes rather than left implicit.
- [ ] Tests cover a base checkout, a linked worktree, a detached HEAD, and a
  non-git planning directory.

## Out of scope

- A mutation-confirmation policy for acting on a non-default worktree (deferred by the
  atlas task on purpose; it needs its own product decision).
- Any git operation beyond reading `HEAD` — `DescribeCheckout` already avoids shelling out.

## Related

- Epic [29-multi-space-planning-a-home-registry-and-the-atlas](../epics/29-multi-space-planning-a-home-registry-and-the-atlas.md)
- Audit finding M2: [2026-08-22-multi-workspace-atlas](../audits/6g2k3qye4qma-2026-08-22-multi-workspace-atlas.md)
