---
schema: 1
id: 6g2nnkfk1em1
status: ready-to-start
epic: 29-multi-space-planning-a-home-registry-and-the-atlas
description: Make entry-point preference base-over-worktree explicit so a card cannot silently summarize a feature branch, then surface branch and worktree state.
effort: M
tier: 3
priority: medium
autonomy_level: 3
tags: [tui, atlas, worktree, spacehealth]
created: "2026-08-22"
updated_at: "2026-08-23"
---
# Make worktree-aware entry selection correct, then show it

## Objective

Two things, in this order: a **correctness** fix, then the badges this task was originally
scoped as.

A git worktree of a *direct* planning checkout has its own copy of the planning files on
another branch. Two such worktrees registered under one repo id are therefore two
genuinely different planning states in one logical space — and `preferredHealthyEntry`
picks the first healthy **direct** entry *in registry order*. That order is arbitrary, so
the atlas card and `status --all` can silently summarize a feature branch's planning tree
instead of the main one, with nothing on screen saying so.

Only worktrees of a *direct* checkout are exposed. Pointer repos are structurally immune:
`desirelines-deploy` sits on a feature branch but resolves to `desirelines-planning/planning`
regardless, because a branch does not move `planning_repo`.

**Measured 2026-08-23: all six registered spaces are `base` checkouts, so this cannot fire
today.** That is why it is sequenced last — see
[the design doc](../research/6g2qtp0022t7-the-atlas-as-a-dashboard-of-dashboards.md).
Re-measure before starting; if worktrees have since been registered, this moves ahead of
the tile work.

## Acceptance criteria

- [ ] `spacehealth.SpaceProblem` carries branch and linked-worktree state from
  `config.DescribeCheckout`, and `core.SpaceEntryPoint` exposes them as typed fields.
- [ ] Entry-point preference becomes explicit and documented: a direct **base** checkout
  wins over a direct **worktree**, which wins over a pointer — replacing today's implicit
  reliance on registry order. `status --all` and the atlas share the change, since both
  read `preferredHealthyEntry`.
- [ ] When a card summarizes a worktree because no base checkout is registered, it says so
  rather than presenting the summary as the space's canonical state.
- [ ] Entry rows show the branch name and a worktree marker; a base checkout with no branch
  (detached HEAD, non-git planning dir) renders without inventing one.
- [ ] The tile grid's **reserved title-row slot** is filled by this marker, with no
  re-layout of the tile — that slot exists precisely so this task is additive.
- [ ] `space list` / registry JSON either carry the same fields or deliberately do not,
  recorded in the notes rather than left implicit.
- [ ] Tests cover: base vs linked worktree vs detached HEAD vs non-git dir; preference
  ordering with a base and a worktree registered under one id; and the worktree-only case
  that must disclose itself.

## Out of scope

- Summarizing more than one worktree per space, or reconciling planning state across
  branches. One tree per card stands.
- Any git operation beyond reading `HEAD` — `DescribeCheckout` already avoids shelling out.
- A mutation-confirmation policy for acting on a non-default worktree; that needs its own
  product decision if navigation ever demonstrates the hazard.

## Related

- Epic [29-multi-space-planning-a-home-registry-and-the-atlas](../epics/29-multi-space-planning-a-home-registry-and-the-atlas.md)
- Design: [The atlas as a dashboard of dashboards](../research/6g2qtp0022t7-the-atlas-as-a-dashboard-of-dashboards.md)
- Reserved slot comes from: [Re-lay the atlas spaces view as a tile grid](6g2xnn4yes8a-re-lay-the-atlas-spaces-view-as-a-tile-grid.md)
- Audit finding M2: [2026-08-22-multi-workspace-atlas](../audits/6g2k3qye4qma-2026-08-22-multi-workspace-atlas.md)
