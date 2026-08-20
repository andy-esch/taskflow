---
schema: 1
id: 6g1s9jpzr1yk
bucket: closed
area: worktree-aware-resolution
date: "2026-08-19"
---

# Architectural audit — worktree-aware cross-repo resolution

Scope: Adversarial review of the worktree-aware cross-repo resolution changes in `internal/config/worktree.go` and `internal/config/config.go` (`anchorDir`, `canonicalCheckout`, `readGitDirFile`, and modified callers in `resolvePlanningRepo` and `resolveRepoPath`).

Method: Four independent reviews focusing on Git internal edge cases, path handling across symlinks/case-insensitivity, cross-repo linkback consistency, and resolution semantics. All findings are grounded in `file:line` and verified with synthetic test suites.

Overall health is solid. The core design principle — distinguishing in-tree paths (`taskflow_root`, which must resolve locally to prevent cross-checkout write pollution) from out-of-tree escapes (`planning_repo` and `tracked_repos`, which anchor to the canonical checkout) — is sound and properly implemented. The non-negotiable invariant that `taskflow_root` is never redirected from a worktree is fully preserved. However, the Git internals discriminator contains a brittle literal `".git"` base check that breaks `--separate-git-dir` repositories; linkback verification exhibits an asymmetry when worktrees are directly tracked; and worktree branch-level config mutations can collide with main checkout relative path anchors.

## High

#### H1. Worktrees of non-standard or `--separate-git-dir` checkouts fail canonical anchoring due to literal `".git"` base assertion  · **Status:** wontfix

**File:** `internal/config/worktree.go:82-87` | **Component:** config / worktree
**Effort:** S · **Urgency:** soon

`canonicalCheckout` extracts the repository's git directory and asserts `if filepath.Base(repoGitDir) != ".git" { return "", false }`. While intended to filter out bare repositories (which have no working tree), this check prematurely rejects any non-bare repository initialized with `--separate-git-dir` (e.g. `git init --separate-git-dir=/path/main.git /path/main`), submodules with worktrees, or checkouts with custom git directory names.

For these repositories, `canonicalCheckout` returns `("", false)` and degrades to `dir`, leaving the original failure modes (linkback warnings and incorrect relative path resolution) active.

**Evidence:**
A synthetic test creating a repository with `git init --separate-git-dir=main.git main` and adding a worktree `main-wt` yields:
```
anchorDir(main-wt) = "/.../main-wt", want "/.../main"
```
Because `filepath.Base("main.git")` is `"main.git"` rather than `".git"`, `canonicalCheckout` fails to find the canonical working tree.

**Recommendation:** Instead of asserting `filepath.Base(repoGitDir) == ".git"`, inspect whether the candidate directory contains a `.git` file or directory pointing back to `repoGitDir`, or verify whether `repoGitDir` has a `commondir` / `config` with `core.bare = false`.

## Medium

**Resolution (2026-08-19, wontfix — a limitation, and the recommendation is not
implementable).** The gap is real: a `--separate-git-dir` repo with worktrees does not anchor.
But the suggested fixes do not exist on disk. Tested: such a repo's config carries
`bare = false` and **no `core.worktree`**, and `git worktree list` run from the worktree reports
the **git directory** (`main.git`) as the main worktree — not the working tree. Nothing records
where the main working tree is; **git cannot find it either**. Guessing would be strictly worse
than degrading, which is what the code already does.

What WAS wrong is the justification: the comment explained the check as "a bare repo's git dir is
not named `.git`", conflating bare with `--separate-git-dir`. The reviewer was right that the
reasoning is unsound even though the remedy was not available. The comment now states both
layouts and why each is unresolvable, and `TestAnchorDir_SeparateGitDirWorktreeDegrades` pins the
degrade so nobody later "fixes" it into a wrong guess. Severity was overstated — this is a
documented limitation, not an active defect.

#### M1. Asymmetric Linkback verification when `tracked_repos` contains a direct worktree path  · **Status:** fixed

**File:** `internal/config/config.go:568-583`, `internal/config/config.go:531-536` | **Component:** config
**Effort:** S · **Urgency:** soon

When a planning repo's `tracked_repos` contains a worktree path directly (e.g., `tracked_repos = ["../impl-wt"]`), `CheckLinks` running from the planning repo reports 0 errors (`checkTrackedRepo` passes because `implDir` resolves to `../impl-wt` and its pointer resolves back to planning).

However, when `CheckLinks` runs from the worktree itself (`../impl-wt`), `checkBackLink` sets `me := resolveRepoPath(cfg.Dir, ".")` which anchors `cfg.Dir` to the canonical checkout (`/path/to/impl`). When iterating `pcf.TrackedRepos`, `resolveRepoPath(pdir, tr)` evaluates against `pdir` (the planning repo) and resolves to `/path/to/impl-wt`. Because `/path/to/impl-wt != /path/to/impl`, `checkBackLink` reports a false-positive:
`⚠ one-sided link: planning repo "../planning" does not track this repo back`

**Evidence:**
In `checkBackLink`:
```go
me := resolveRepoPath(cfg.Dir, ".") // returns /path/to/impl (anchored)
for _, tr := range pcf.TrackedRepos {
    if resolveRepoPath(pdir, tr) == me { // resolveRepoPath returns /path/to/impl-wt (unanchored)
        return nil
    }
}
```

**Recommendation:** In `checkBackLink`, anchor both sides of the comparison (i.e. canonicalize both candidate paths via `anchorDir(resolveRepoPath(pdir, tr)) == me`).

**Resolution (2026-08-19, fixed — and it was worse than filed).** Confirmed via the
realistic path rather than a hand-edited config: running `init --planning-repo` **from a
worktree** recorded `../impl-wt` in `tracked_repos`, and the very next command from that same
worktree warned it was not tracked. The tool wrote the bad state and then complained about it —
reintroducing the exact symptom this branch exists to fix. Arguably High, not Medium.

Fixed on both axes:

1. **Recording** — `LinkBack` now resolves through `anchorDir`, so it records the canonical
   checkout. This also stops entries rotting: worktrees are ephemeral, so a worktree entry dies
   the moment the worktree is removed.
2. **Tolerance** — `resolveRepoPath` now anchors its RESULT as well as its base, so a config that
   already names a worktree (written by the old behavior, or by hand) still counts as tracked.

Deliberately **not** applied to `resolvePlanningRepo`: pointing at a planning worktree is
legitimate, and a planning worktree resolves to its own tree by design. Tests:
`TestWorktree_LinkBackRecordsCanonicalCheckout`, `TestWorktree_TolerantOfWorktreePathAlreadyRecorded`.

#### M2. Semantic collision when worktree branch modifies relative `planning_repo`  · **Status:** wontfix

**File:** `internal/config/config.go:193-197` | **Component:** config
**Effort:** M · **Urgency:** eventually

When a developer on a feature branch in a worktree modifies `.tskflwctl.toml` to point to a different relative planning repo (e.g. `planning_repo = "../feature-planning"`), `resolvePlanningRepo` reads the value from the worktree's config, but unconditionally joins it against `anchorDir(dir)` (the canonical checkout's directory).

If `feature-planning` was created relative to the worktree directory, resolution fails with a confusing error naming the main checkout's parent directory, or worse, silently binds to an unrelated directory of the same name adjacent to the main checkout.

**Evidence:**
```go
root = filepath.Join(anchorDir(dir), root)
```
The value is taken from the worktree branch config, but the base path is substituted with the canonical checkout directory.

**Recommendation:** Document this boundary semantic clearly, or validate whether relative `planning_repo` paths starting with `./` (in-tree or relative-descendant) should resolve against `dir` while `../` escapes anchor to canonical checkouts.

**Deferred (2026-08-19) — tracked, and deliberately coupled.** Filed as
[decide-how-a-relative-planning-repo-resolves-from-a-worktree-branch](../tasks/6g1sb9keb5c9-decide-how-a-relative-planning-repo-resolves-from-a-worktree-branch.md).
Rare (needs a deliberate per-branch pointer edit) but the bad outcome is a silent wrong-tree bind.
Sequenced **with** the `space.id` identity decision: if pointers verify a durable id after
resolving, the dangerous outcome becomes a loud error regardless of which base was used, which
dissolves most of this without adding a `./` vs `../` rule. Deciding it first risks shipping a
rule that identity verification then makes redundant.

**Closed wontfix (2026-08-20).** Identity verification shipped, which was the coupling this
was deferred on. The dangerous outcome — a silent bind to an unrelated planning repo — is now
`ErrConflict` for any opted-in pointer, on a mismatched id *and* on a target with no id. What
remains is a clearer error, not a wrong result, so the `./` vs `../` prefix rule is
unnecessary. The anchoring is instead documented in the pointer config the tool writes, where
a reader of `.tskflwctl.toml` will find it.

#### M3. `appendTrackedRepo` deduplication fails for worktree paths against canonical checkouts  · **Status:** fixed

**File:** `internal/config/config.go:509-514` | **Component:** config
**Effort:** XS · **Urgency:** soon

When `AddTrackedRepo(planningDir, "../impl-wt")` is called, `target := resolveRepoPath(dir, entry)` evaluates `"../impl-wt"` relative to `dir` (the planning repo). Because `dir` is not a worktree, `resolveRepoPath` returns `/path/to/impl-wt`.

If `tracked_repos` already contains `"../impl"`, `resolveRepoPath(dir, e)` returns `/path/to/impl`. The equality check fails, so `appendTrackedRepo` adds duplicate entries representing checkouts of the same repository.

**Recommendation:** Canonicalize the resolved target in `appendTrackedRepo` (`target := anchorDir(resolveRepoPath(dir, entry))`) so duplicate checkouts of the same repository collapse to a single tracked entry.

## Low

**Resolution (2026-08-19, fixed).** Same root cause as M1 — the resolved *value* was not
anchored, only the base — and fixed by the same one-line change to `resolveRepoPath`. Verified:
`init --track ../impl-wt` against a config already listing `../impl` now collapses to a single
entry instead of duplicating the repository.

#### L1. Redundant un-cached filesystem reads and stats during `CheckLinks` iteration  · **Status:** deferred

**File:** `internal/config/worktree.go:60-93`, `internal/config/config.go:590-615` | **Component:** config / perf
**Effort:** XS · **Urgency:** eventually

`canonicalCheckout` performs multiple disk I/O operations (`os.Lstat`, `os.ReadFile`, `os.Stat`, `filepath.EvalSymlinks`) without caching results. In `CheckLinks`, `resolveRepoPath` is called multiple times per tracked repository in a loop (e.g. 4 times per entry). For a planning repo tracking 10 repositories, this triggers 40+ stats and reads on every command during `warnLinks`.

**Recommendation:** Memoize `canonicalCheckout` within a package-level run or structure `CheckLinks` to resolve each directory once.

**Deferred (2026-08-19) — measured, not assumed.** Filed as
[cache-worktree-anchoring-if-it-ever-shows-up-in-a-profile](../tasks/6g1sbcbfb4mk-cache-worktree-anchoring-if-it-ever-shows-up-in-a-profile.md).
Timed after the change (10 runs each): `task list` 45.3 ms, `doctor` 8.1 ms — and `doctor` is the
linkback-heavy path AND the cheapest command measured, so the extra syscalls are well below noise.
Caching now would add invalidation questions (worktrees appear and vanish mid-session) for no
measured gain. Triggers that would change the call are recorded on the task.

## What is solid (checked, deliberately not findings)

- **In-tree `taskflow_root` protection:** `configuredRoot` strictly avoids `anchorDir`, ensuring that worktrees with in-tree planning (such as `taskflow` itself) always edit their own worktree files and never redirect mutations to the main checkout.
- **Gitdir file parsing:** `readGitDirFile` cleanly handles CRLF line endings, leading/trailing whitespace, and relative `gitdir:` paths (as written by submodules and relative worktrees).
- **Graceful error degradation:** Malformed `.git` files, dangling gitdirs, bare repositories, and submodules all degrade safely to returning the original directory without panics or unexpected errors.
- **Physical path comparison:** Preserving `evalOr` containment and normalization prevents symlink traversal confusion.

## Candidate tasks

- ⏳ `tskflwctl task new "Fix canonical checkout resolution for separate-git-dir worktrees" --epic 23-point-an-impl-repo-at-an-external-planning-repo --tags config,git` — Relax filepath.Base == .git requirement in canonicalCheckout.
- ⏳ `tskflwctl task new "Ensure symmetric linkback verification for tracked worktrees" --epic 23-point-an-impl-repo-at-an-external-planning-repo --tags config` — Canonicalize both sides of the comparison in checkBackLink.
- ⏳ `tskflwctl task new "Deduplicate worktrees in appendTrackedRepo" --epic 23-point-an-impl-repo-at-an-external-planning-repo --tags config` — Apply anchorDir to entry paths in appendTrackedRepo.
