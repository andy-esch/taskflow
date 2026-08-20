---
schema: 1
id: 6g1m8mc8p46h
status: ready-to-start
epic: 29-multi-space-planning-a-home-registry-and-the-atlas
description: Settle what identifies a planning repo before the registry schema is written. Gates --space's write guard, linkback verification, the workspace receipt, and moved detection.
effort: Unknown
tier: 3
priority: high
autonomy_level: 3
tags: [design, config, multi-repo]
created: "2026-08-19"
updated_at: "2026-08-19"
---
# Decide `space.id` identity: durable minted id vs path-keyed

## Objective

Settle what identifies a planning repo, **before** any registry schema is written. This
started as one line of the registry task and has since acquired four consumers and a
safety property, which is why it is now its own decision.

## Why this is not a detail

| consumer | what it needs identity for | what breaks if identity is a path |
| --- | --- | --- |
| `--space <id>` selection | the wrong-repo write guard that replaced `--expect-root` | the guarantee "naming a tree cannot resolve to the wrong one" stops holding when a path drifts |
| linkback verification (epic 23) | proving a `planning_repo` pointer reached the intended repo | a committed relative path can **silently** resolve to an unrelated planning repo — demonstrated |
| `wire.WorkspaceJSON` on receipts | proving which tree a mutation changed | the receipt names a path that may since have moved |
| `moved` detection (fork 6) | recognising a repo that changed location | reduces to a filename heuristic |

Evidence for row 2:
[worktrees-and-relative-paths-in-cross-repo-planning](../research/6g1emag02srv-worktrees-and-relative-paths-in-cross-repo-planning.md)
— a worktree outside the sibling layout resolved to a *different* planning repo with no
warning and no error.

## The options

- **Durable minted id.** A stable id written into the planning repo's own **committed**
  `.tskflwctl.toml` (`id = "6g1…"`, minted by `init`). Registry entries and pointers
  reference it; path is a hint, id is the assertion. Follows
  [ADR-0003](../adrs/0003-stable-key-id-addressed-storage.md)'s philosophy of addressing
  by a durable key rather than a location. Costs a backfill for existing repos.
- **Path-keyed label.** `space.id` is a local nickname for a path; nothing is written to
  the target repo. Zero migration, no coordination — but every row above stays fragile,
  and the wrong-tree hazard is permanent.

## What to settle

- Durable or path-keyed? If path-keyed, **record explicitly that the wrong-tree hazard
  is accepted**, so it reads as a decision rather than an omission.
- If durable: where is it minted (`init` only, or `lint --fix`/`doctor` backfill for the
  two existing repos)? Is a repo without one an error, a warning, or silently tolerated?
- Does `wire.WorkspaceJSON` carry it, so the receipt and the registry agree? (Adding a
  field there is cheap now, awkward once consumers pin the shape.)
- Is the id per **planning repo** only, or does an impl repo get one too? Only the
  planning repo needs identity for the four consumers above — resist widening.

## Acceptance criteria

- [ ] The choice is recorded here with its rejected alternative and the accepted cost
- [ ] If durable: minting, backfill, and missing-id behavior are specified
- [ ] The registry task's schema section is updated to match
- [ ] Whether `WorkspaceJSON` carries it is decided either way

## Out of scope

- Implementing the registry — this is the decision that unblocks it.
- Layers A/B of the worktree fix, which are independent of this
  ([make-cross-repo-resolution-worktree-aware](6g1kxj5br3z1-make-cross-repo-resolution-worktree-aware.md)).

## Related

- Epic [29-multi-space-planning-a-home-registry-and-the-atlas](../epics/29-multi-space-planning-a-home-registry-and-the-atlas.md)
- The evidence: [worktrees-and-relative-paths-in-cross-repo-planning](../research/6g1emag02srv-worktrees-and-relative-paths-in-cross-repo-planning.md)
- Consumer 1: [global-space-flag-run-any-command-against-a-registered-space](6g0fzk8mazrc-global-space-flag-run-any-command-against-a-registered-space.md)
- The schema this gates: [space-registry-model-and-the-space-verbs-list-add-forget](6g0fzhbtk0gy-space-registry-model-and-the-space-verbs-list-add-forget.md)

### 2026-08-19 — "durable vs path-keyed" is a false dichotomy

Prompted by asking whether a user should be able to select *which worktree* of a planning
repo to view. Tested: three worktrees of one planning repo, three branches, and the
**same committed id** in all three.

```
CHECKOUT         COMMITTED id     BRANCH
planning         6g1durable0aa    main
planning-wt-a    6g1durable0aa    branch-a
planning-wt-b    6g1durable0aa    branch-b
```

A committed id necessarily identifies the **repo**, not the **checkout**. So a single
durable id cannot serve as the registry key the moment worktrees exist — `--space <id>`
would be ambiguous across them.

**The two identities answer different questions, and this task should decide both:**

| identity | answers | properties | consumers |
| --- | --- | --- | --- |
| **repo identity** (durable, committed) | "is this the planning repo I meant?" | survives moves; **not** unique per checkout | linkback verification, the wrong-tree guard, `moved` detection |
| **checkout identity** (path, local label) | "which working tree am I looking at?" | unique per checkout; **breaks when the path moves** | the registry key, `--space` selection |

Framed as either/or, each option loses something real. Held as two fields, they compose:
the registry keys on a local label + path, and *verifies* against the committed repo id
before writing. Path is the address, id is the assertion — the same split
[worktrees-and-relative-paths-in-cross-repo-planning](../research/6g1emag02srv-worktrees-and-relative-paths-in-cross-repo-planning.md)
layer C proposes for pointers.

**On registering worktrees as spaces — leaning against.** Worktrees are ephemeral:
`desirelines-planning` already carries a **prunable** worktree registration
(`/workspace/dl-sweepport`, a directory that no longer exists). A registry that admits
them accumulates corpses and amplifies the `missing`/`moved` diagnosis problem. Selecting
a checkout stays `-C <path>`, which already works — a planning-repo worktree correctly
resolves to itself.

**But note taskflow cannot follow ADR-0004's trunk-only guidance.** It self-hosts
(`taskflow_root = "./planning"`), so *every* branch and worktree of this repo carries its
own planning tree — the divergence ADR-0004 set out to escape is this repo's default
state, not an exception. Whatever is decided here should be honest about that rather than
assuming one checkout per planning repo.
