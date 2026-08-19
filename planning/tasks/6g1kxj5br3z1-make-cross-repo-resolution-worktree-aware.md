---
schema: 1
id: 6g1kxj5br3z1
status: next-up
epic: 23-point-an-impl-repo-at-an-external-planning-repo
description: Worktrees look like untracked impl repos, so 4 dirs warn on every command. Resolve a worktree to its canonical checkout for both linkback and relative planning_repo resolution.
effort: Unknown
tier: 3
priority: high
autonomy_level: 3
tags: [config, multi-repo, git]
created: "2026-08-19"
---
# Make cross-repo resolution worktree-aware

## Objective

Four of the desirelines worktrees emit `⚠ one-sided link` on **every command**, because
`CheckLinks` compares physical directory paths and a worktree is a different directory
but the same repo. Fix the model rather than the symptom, and apply the same fix to
*resolution* so a worktree resolves its `planning_repo` the way its canonical checkout
would.

Full evidence and the rejected alternatives:
[worktrees-and-relative-paths-in-cross-repo-planning](../research/6g1emag02srv-worktrees-and-relative-paths-in-cross-repo-planning.md).

## Two halves, one mechanism

Both need "given a directory, find the canonical checkout":

- **A — linkback (`CheckLinks`).** Compare the *canonical* checkout against
  `tracked_repos`, so a worktree of a tracked repo is considered tracked.
- **B — resolution (`resolvePlanningRepo`).** When the config lives in a worktree,
  resolve a **relative** `planning_repo` against the canonical checkout's directory, so
  `../desirelines-planning` means the same thing from every worktree.

## Notes

- A worktree's `.git` is a **file** containing
  `gitdir: /…/<main>/.git/worktrees/<name>`. The canonical checkout is
  `<main>` — a plain file read plus a path walk; **no git binary, no subprocess**.
  `config.go` already special-cases `.git`-as-a-file for the climb boundary
  (`exists(filepath.Join(dir, ".git"))`), so this extends an existing concession
  rather than introducing a new concept.
- Keep comparisons on **physical** paths (`Abs` + `EvalSymlinks`), as `CheckLinks`
  already does — the canonical checkout is just one more spelling to normalize.
- Submodules also carry a `.git` file. Decide explicitly whether they resolve the same
  way or are left alone; do not let the behavior fall out by accident.
- A bare repo, a corrupted `gitdir:` line, or a `gitdir:` pointing somewhere that no
  longer exists must all degrade to today's behavior, never error. This is a
  convenience layer over a warning path — it must not become a new failure mode.
- **Do NOT** simply suppress the warning for worktrees. That hides the only visible
  symptom of the silent-wrong-tree hazard documented in the research doc.

## Acceptance criteria

- [ ] A worktree of a repo listed in `tracked_repos` produces **no** linkback warning
- [ ] A worktree of an **untracked** repo still warns (the check is not just disabled)
- [ ] A relative `planning_repo` resolves identically from the canonical checkout and
      from a worktree placed in a different parent directory
- [ ] Non-worktree behavior is byte-identical to today (regression test)
- [ ] A malformed/dangling `.git` file degrades to current behavior without erroring
- [ ] Submodule behavior is decided and covered by a test either way
- [ ] `just test` + `just lint` green

## Out of scope

- The durable-identity fix (research doc, layer C) — that is decided with `space.id`
  on the registry task, not here.
- Registry-based recovery (layer D) — falls out of epic 29 slice 2.

## Related

- Research: [worktrees-and-relative-paths-in-cross-repo-planning](../research/6g1emag02srv-worktrees-and-relative-paths-in-cross-repo-planning.md)
- Epic [23-point-an-impl-repo-at-an-external-planning-repo](../epics/23-point-an-impl-repo-at-an-external-planning-repo.md)
