---
schema: 1
id: 6g1s5m46g1bf
status: completed
epic: 23-point-an-impl-repo-at-an-external-planning-repo
description: workspace cannot distinguish two worktrees of one repo, which differ only by branch. Report the branch and base-vs-worktree, surfacing the non-default case.
effort: Unknown
tier: 3
priority: medium
autonomy_level: 3
tags: [cli, git, multi-repo]
created: "2026-08-19"
updated_at: "2026-08-20"
started_at: "2026-08-20"
completed_at: "2026-08-20"
---
# Report the checkout's branch in `workspace` and receipts

## Objective

`workspace` reports the resolved root, config path, and source — but not **which branch**
the checkout is on. Two worktrees of one repo differ *only* by branch, and their directory
names are often near-identical (`-wt-go-mod`, `-wt-tf-cleanup`), so path alone cannot tell
them apart in a header, a receipt, or a bug report.

Split out of
[make-cross-repo-resolution-worktree-aware](6g1kxj5br3z1-make-cross-repo-resolution-worktree-aware.md)
deliberately: that change closes a live bug and is self-contained, while this is additive
and serves a board that does not exist yet.

## Scope

- **Branch name**, read from the checkout's `HEAD` (a detached HEAD reports the short sha).
- **Whether this checkout is the base one or a linked worktree.** `canonicalCheckout` in
  `internal/config/worktree.go` already distinguishes them; this exposes it.
- **Emphasize the non-default case.** A base checkout on its default branch is the boring
  case and should stay visually quiet; a worktree, or any non-default branch, is what the
  reader needs to notice. Prefer surfacing the exception over labelling everything.

## Notes

- No git binary. `HEAD` is a file: either `ref: refs/heads/<branch>` or a raw sha. For a
  linked worktree, `HEAD` lives in that worktree's own git dir
  (`<common>/worktrees/<name>/HEAD`), not the shared one — the same `gitdir:` read
  `worktree.go` already performs gets you there.
- A planning root that is not in a git repo at all is the normal case for many users:
  report nothing, never error.
- **Decide whether `wire.WorkspaceJSON` carries these.** Cheap now; awkward once
  consumers pin the shape. If yes it is a schema bump (currently 1.33) — and note
  `workspace` is marked EXPERIMENTAL precisely so its shape can still move.
- Where does "default branch" come from? There is no reliable local answer (no remote
  HEAD guarantee). Consider basing the highlight on **base-vs-worktree** instead, which
  is knowable offline, and treating branch purely as a label.

## Acceptance criteria

- [x] `workspace` reports the branch (short sha when detached) and whether the checkout
      is base or a linked worktree
- [x] A non-git planning root reports neither and does not error
- [x] Whether `WorkspaceJSON` carries them is decided; if yes, schema bumped + goldens
- [x] No git binary is invoked
- [x] `just test` + `just lint` green

## Related

- Split from [make-cross-repo-resolution-worktree-aware](6g1kxj5br3z1-make-cross-repo-resolution-worktree-aware.md)
- The consumer: the atlas card header — board section of
  [multi-space-home-registry-and-the-atlas](../research/6g0ajre026c6-multi-space-home-registry-and-the-atlas.md)
