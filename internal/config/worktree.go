package config

import (
	"os"
	"path/filepath"
	"strings"
)

// Worktree awareness.
//
// A git worktree is a SEPARATE DIRECTORY that shares one repository. Nothing in the
// planning model noticed: linkback compares physical directory paths, so a worktree of
// a tracked impl repo looked like an untracked one and warned on every command; and a
// committed relative `planning_repo` resolved against the worktree's own location,
// which is not where its author meant (and, if some other planning repo happened to sit
// at that relative path, resolved SILENTLY to the wrong tree).
//
// The rule this implements is not "worktrees are special" — it is about which side of
// the repo boundary a path points at:
//
//	taskflow_root  → points INSIDE the repo (configuredRoot rejects escapes)
//	                 ⇒ resolve against the config file's ACTUAL dir, always.
//	                   A worktree's planning/ subdir is its own; redirecting it would
//	                   write to a different checkout than the one you are standing in.
//	planning_repo  → points OUTSIDE (the sanctioned escape)
//	tracked_repos  → points OUTSIDE
//	                 ⇒ resolve against the CANONICAL CHECKOUT, so the committed value
//	                   means the same thing from every worktree.
//
// Everything here degrades to the input on any doubt: this is a convenience layer over
// a warning path and a resolution hint, and must never become a new failure mode.

// gitDirPrefix is the marker a `.git` FILE uses to name the real git directory.
const gitDirPrefix = "gitdir:"

// worktreesSegment is the path component git inserts for a linked worktree:
// <repo-git-dir>/worktrees/<name>. Its presence is what distinguishes a worktree from
// the other two things a `.git` FILE can mean — `git init --separate-git-dir` (which IS
// the main checkout, just with its git dir moved) and a submodule (its own repo, whose
// gitdir is `../.git/modules/<name>` and, notably, RELATIVE).
const worktreesSegment = "worktrees"

// anchorDir returns the directory that OUT-OF-TREE relative paths in dir's config
// should resolve against: dir's canonical checkout when dir is a linked worktree, and
// dir itself in every other case.
//
// Returns dir unchanged for: a normal checkout, a `--separate-git-dir` repo, a
// submodule, a bare-repo worktree (whose "canonical checkout" has no working tree), a
// malformed or dangling `.git`, and any read error.
func anchorDir(dir string) string {
	if canonical, ok := canonicalCheckout(dir); ok {
		return canonical
	}
	return dir
}

// canonicalCheckout resolves a linked worktree to the working tree of the repository
// it belongs to. ok is false when dir is not a linked worktree, or when the repository
// has no working tree to point at (the bare + worktrees-only layout).
func canonicalCheckout(dir string) (string, bool) {
	gitPath := filepath.Join(dir, ".git")
	info, err := os.Lstat(gitPath)
	if err != nil || info.IsDir() {
		return "", false // no .git, or a normal checkout — already canonical
	}
	gitDir, ok := readGitDirFile(gitPath, dir)
	if !ok {
		return "", false
	}
	// <repo-git-dir>/worktrees/<name> — and ONLY that shape. The segment must be the
	// second-to-last component; a repo that merely lives under a directory called
	// "worktrees" must not be mistaken for one.
	parent, name := filepath.Split(filepath.Clean(gitDir))
	if name == "" {
		return "", false
	}
	repoGitDir := filepath.Clean(parent)
	if filepath.Base(repoGitDir) != worktreesSegment {
		return "", false // --separate-git-dir, or a submodule
	}
	repoGitDir = filepath.Dir(repoGitDir)
	// Accept only the conventional layout, where the working tree is the parent of a
	// `.git` DIRECTORY. Two other layouts reach here and BOTH are genuinely
	// unresolvable, not merely unhandled:
	//
	//   - a bare repo (`repo.git/worktrees/<name>`) has no working tree at all;
	//   - `git init --separate-git-dir` HAS a working tree, but nothing on disk records
	//     where it is. Its config carries `bare = false` and no `core.worktree`, and
	//     `git worktree list` itself reports the GIT DIRECTORY as the main worktree.
	//     Git cannot find it either.
	//
	// So degrading — the caller keeps the worktree's own directory — is the correct
	// answer for both, not a gap to close later
	// (audit 2026-08-19-worktree-aware-resolution, H1).
	if filepath.Base(repoGitDir) != ".git" {
		return "", false
	}
	checkout := filepath.Dir(repoGitDir)
	if !isDir(checkout) {
		return "", false // the main checkout was moved or deleted
	}
	return evalOr(checkout), true
}

// readGitDirFile reads the `gitdir:` line from a `.git` FILE. A relative value is
// resolved against base (the directory holding the file) — submodules write relative
// gitdirs, so this is not hypothetical.
func readGitDirFile(gitPath, base string) (string, bool) {
	b, err := os.ReadFile(gitPath)
	if err != nil {
		return "", false
	}
	for line := range strings.SplitSeq(string(b), "\n") {
		rest, found := strings.CutPrefix(strings.TrimSpace(line), gitDirPrefix)
		if !found {
			continue
		}
		p := strings.TrimSpace(rest)
		if p == "" {
			return "", false
		}
		if !filepath.IsAbs(p) {
			p = filepath.Join(base, p)
		}
		return filepath.Clean(p), true
	}
	return "", false
}
