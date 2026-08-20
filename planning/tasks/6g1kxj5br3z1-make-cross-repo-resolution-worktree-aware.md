---
schema: 1
id: 6g1kxj5br3z1
status: completed
epic: 23-point-an-impl-repo-at-an-external-planning-repo
description: Worktrees look like untracked impl repos, so 4 dirs warn on every command. Resolve a worktree to its canonical checkout for both linkback and relative planning_repo resolution.
effort: Unknown
tier: 3
priority: high
autonomy_level: 3
tags: [config, multi-repo, git]
created: "2026-08-19"
updated_at: "2026-08-19"
started_at: "2026-08-19"
completed_at: "2026-08-19"
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

- [x] A worktree of a repo listed in `tracked_repos` produces **no** linkback warning
- [x] A worktree of an **untracked** repo still warns (the check is not just disabled)
- [x] A relative `planning_repo` resolves identically from the canonical checkout and
      from a worktree placed in a different parent directory
- [x] Non-worktree behavior is byte-identical to today (regression test)
- [x] A malformed/dangling `.git` file degrades to current behavior without erroring
- [x] Submodule behavior is decided and covered by a test either way
- [x] `just test` + `just lint` green

## Out of scope

- The durable-identity fix (research doc, layer C) — that is decided with `space.id`
  on the registry task, not here.
- Registry-based recovery (layer D) — falls out of epic 29 slice 2.

## Related

- Research: [worktrees-and-relative-paths-in-cross-repo-planning](../research/6g1emag02srv-worktrees-and-relative-paths-in-cross-repo-planning.md)
- Epic [23-point-an-impl-repo-at-an-external-planning-repo](../epics/23-point-an-impl-repo-at-an-external-planning-repo.md)

### 2026-08-19 — pre-implementation findings (tested, not assumed)

Four things settled empirically before starting. The third would have caused a
wrong-checkout write if implemented from the original description.

#### 1. A `.git` FILE has three causes, and only one is a worktree

| cause | `gitdir:` value | canonical checkout |
| --- | --- | --- |
| worktree | `<repo-gitdir>/worktrees/<name>` (absolute) | derivable |
| `git init --separate-git-dir` | a plain dir, no `/worktrees/` segment | **itself** — it IS the main checkout |
| submodule | `../.git/modules/<name>` — **relative** | itself; a submodule is its own repo |

So the discriminator is the **`/worktrees/<name>` suffix on the gitdir**, not merely
"`.git` is a file". Treating a `--separate-git-dir` repo as a worktree would walk up to
an unrelated parent. And the parser must accept a **relative** `gitdir:` (resolved
against the dir holding the `.git` file), which submodules use.

#### 2. Bare-repo worktrees have NO canonical checkout

The bare + worktrees-only workflow yields `gitdir: /…/repo.git/worktrees/wt` — note
`repo.git/worktrees/`, not `.git/worktrees/`. Stripping `/worktrees/<name>` gives the
**bare repo**, which has no working tree and therefore no `.tskflwctl.toml`.

Two consequences: the suffix match must not assume a literal `.git` path segment, and
when the derived checkout has no working tree, the code must **degrade to today's
behavior**, not error. This is a convenience layer over a warning path.

#### 3. A worktree of the PLANNING repo must NOT be redirected

Tested: `tskflwctl` from a worktree of a planning repo resolves to **that worktree's own
tree** — which is correct. The worktree has its own checked-out copy of the planning
files, and that branch's version is what you are editing. Redirecting to the canonical
checkout would write to a different checkout than the one you are standing in.

**So this task is not "redirect worktrees to the canonical checkout."** It is narrower:
redirect only the resolution of paths that point *outside* the repo.

#### 4. The rule that falls out: in-tree vs out-of-tree paths

| key | points | resolve against |
| --- | --- | --- |
| `taskflow_root` | **inside** the repo (guaranteed — `configuredRoot` rejects escapes) | the config file's **actual** dir, always |
| `planning_repo` | **outside** (the sanctioned escape) | the **canonical checkout** |
| `tracked_repos` | **outside** | the **canonical checkout** |

That is a statable rule rather than a worktree special case, and it derives from an
invariant the code already enforces. It also means the planning side needs the same
treatment: if the *planning* repo is a worktree, `checkTrackedRepo` resolving
`../desirelines` from the worktree's dir is wrong for the same reason.

#### Revised acceptance criteria

Supersedes the list above where they conflict:

- [ ] Only a gitdir with a `/worktrees/<name>` suffix is treated as a worktree;
      `--separate-git-dir` and submodule `.git` files are left alone (test each)
- [ ] A relative `gitdir:` is resolved against the dir holding the `.git` file
- [ ] A bare-repo worktree (no canonical working tree) degrades to today's behavior
- [ ] `taskflow_root` resolution is UNCHANGED from a worktree (regression test — this is
      the one that writes to the wrong checkout if broken)
- [ ] `planning_repo` and `tracked_repos` resolve against the canonical checkout
- [ ] A worktree of a repo in `tracked_repos` produces no linkback warning; a worktree of
      an untracked repo still warns
- [ ] Non-worktree behavior byte-identical (regression test)
- [ ] `just test` + `just lint` green

### 2026-08-19 — `workspace` should report the branch (new consumer)

The atlas card design settled on **one card per planning repo, with worktrees as
selectable variants**, and the card header carries the checkout path
(`desirelines (~/git/andy-esch/desirelines)`). See the board section of
[multi-space-home-registry-and-the-atlas](../research/6g0ajre026c6-multi-space-home-registry-and-the-atlas.md).

That makes a gap load-bearing: **`workspace` reports the resolved root, config path and
source, but not the branch.** Two worktrees of one repo differ *only* by branch, and their
directory names are often near-identical (`-wt-go-mod`, `-wt-tf-cleanup`), so path alone
cannot distinguish them in a header or in a receipt.

Worth folding in here since this task is already reading `.git`:

- Report the current branch (read `<gitdir>/HEAD`; a detached HEAD reports the short sha).
- Report whether this checkout is the **base** one (its `.git` is a directory) or a
  worktree — the board needs a "you are on the default" marker.
- Decide whether `wire.WorkspaceJSON` carries both, so the receipt and the board agree.
  Cheap now; awkward once consumers pin the shape.

Grouping itself needs **no committed id** — every checkout of a repo derives the same
**common gitdir** (strip `/worktrees/<name>` from a worktree's `gitdir:`), which is the
natural grouping key and is free. That is the same file read this task already performs,
so exposing it is near-zero extra cost.
