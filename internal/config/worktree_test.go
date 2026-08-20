package config

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// git runs a git command in dir, skipping the test if git is unavailable.
func git(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(),
		"GIT_AUTHOR_NAME=t", "GIT_AUTHOR_EMAIL=t@t", "GIT_COMMITTER_NAME=t", "GIT_COMMITTER_EMAIL=t@t")
	if out, err := cmd.CombinedOutput(); err != nil {
		if _, lookErr := exec.LookPath("git"); lookErr != nil {
			t.Skip("git not available")
		}
		t.Fatalf("git %s: %v\n%s", strings.Join(args, " "), err, out)
	}
}

// implWithWorktree builds an impl repo pointing at a sibling planning repo, plus a
// linked worktree of that impl. Returns (planning, impl, worktree).
func implWithWorktree(t *testing.T, parent string) (string, string, string) {
	t.Helper()
	plan := filepath.Join(parent, "planning")
	mustMkdir(t, filepath.Join(plan, "tasks"))
	writeConfig(t, plan, "taskflow_root = \".\"\n")

	impl := filepath.Join(parent, "impl")
	mustMkdir(t, impl)
	git(t, impl, "init", "-q", ".")
	if _, err := InitPointer(impl, "../planning", false); err != nil {
		t.Fatal(err)
	}
	git(t, impl, "add", "-A")
	git(t, impl, "commit", "-q", "-m", "init")

	wt := filepath.Join(parent, "impl-wt")
	git(t, impl, "worktree", "add", "-q", wt, "-b", "wtb")
	return plan, impl, wt
}

// TestWorktree_LinkbackCreditsCanonicalCheckout is the reported symptom: four
// directories warning on every command because a worktree of a TRACKED repo looked
// untracked. The planning repo lists only the canonical checkout — as it should, since
// worktrees are created and deleted constantly — so the worktree must resolve to it.
func TestWorktree_LinkbackCreditsCanonicalCheckout(t *testing.T) {
	parent := t.TempDir()
	_, impl, wt := implWithWorktree(t, parent)

	if _, err := LinkBack(impl, "../planning", false); err != nil {
		t.Fatal(err)
	}
	if p := linksAt(t, impl); len(p) != 0 {
		t.Fatalf("the canonical checkout must be clean, got %v", p)
	}
	if p := linksAt(t, wt); len(p) != 0 {
		t.Errorf("a worktree of a TRACKED repo must not warn, got %v", p)
	}
}

// TestWorktree_UntrackedStillWarns is the other half, and the one that proves the fix
// is not simply "stop checking worktrees": with no back-link recorded, the warning must
// still fire from the worktree exactly as it does from the canonical checkout.
func TestWorktree_UntrackedStillWarns(t *testing.T) {
	parent := t.TempDir()
	_, impl, wt := implWithWorktree(t, parent) // note: no LinkBack

	if p := linksAt(t, impl); len(p) == 0 {
		t.Fatal("an untracked impl must warn (precondition)")
	}
	if p := linksAt(t, wt); len(p) == 0 {
		t.Error("a worktree of an UNTRACKED repo must still warn — the check must not be disabled")
	}
}

// TestWorktree_RelativePlanningRepoResolvesTheSame pins the silent-wrong-tree fix: a
// committed `../planning` must name the SAME tree from a worktree as from the canonical
// checkout, even when the worktree sits in a different parent directory — and
// especially when some other planning repo happens to sit where the naive relative path
// would land.
func TestWorktree_RelativePlanningRepoResolvesTheSame(t *testing.T) {
	parent := t.TempDir()
	home := filepath.Join(parent, "home")
	mustMkdir(t, home)
	plan, impl, _ := implWithWorktree(t, home)

	// the worktree lives somewhere else entirely...
	elsewhere := filepath.Join(parent, "elsewhere")
	mustMkdir(t, elsewhere)
	wt := filepath.Join(elsewhere, "impl-wt")
	git(t, impl, "worktree", "add", "-q", wt, "-b", "faraway")

	// ...and a DECOY planning repo sits exactly where "../planning" would land from it.
	decoy := filepath.Join(elsewhere, "planning")
	mustMkdir(t, filepath.Join(decoy, "tasks"))
	writeConfig(t, decoy, "taskflow_root = \".\"\n")

	cfg, err := Discover(wt)
	if err != nil {
		t.Fatalf("discover from the worktree: %v", err)
	}
	if want := evalOr(plan); cfg.Root != want {
		t.Errorf("worktree resolved to %q, want the real planning repo %q — a decoy at the\n"+
			"naive relative path must never win", cfg.Root, want)
	}
}

// TestWorktree_TaskflowRootIsNotRedirected is the regression that matters most: a
// worktree of a repo whose planning is IN-TREE must resolve to its OWN planning
// directory. taskflow itself is laid out this way, so redirecting would write to a
// different checkout than the one the user is standing in.
func TestWorktree_TaskflowRootIsNotRedirected(t *testing.T) {
	parent := t.TempDir()
	repo := filepath.Join(parent, "selfhosted")
	mustMkdir(t, filepath.Join(repo, "planning", "tasks"))
	writeConfig(t, repo, "taskflow_root = \"./planning\"\n")
	git(t, repo, "init", "-q", ".")
	git(t, repo, "add", "-A")
	git(t, repo, "commit", "-q", "-m", "init")

	wt := filepath.Join(parent, "selfhosted-wt")
	git(t, repo, "worktree", "add", "-q", wt, "-b", "wtb")
	mustMkdir(t, filepath.Join(wt, "planning", "tasks"))

	cfg, err := Discover(wt)
	if err != nil {
		t.Fatalf("discover from the worktree: %v", err)
	}
	if want := evalOr(filepath.Join(wt, "planning")); cfg.Root != want {
		t.Errorf("in-tree planning resolved to %q, want the WORKTREE's own %q — "+
			"redirecting taskflow_root would edit another checkout's files", cfg.Root, want)
	}
}

// TestAnchorDir_OnlyLinkedWorktreesAnchor pins the discriminator. A `.git` FILE has
// three causes and only one of them is a worktree; the other two are their own
// canonical checkouts and must be left exactly as they are.
func TestAnchorDir_OnlyLinkedWorktreesAnchor(t *testing.T) {
	parent := t.TempDir()

	t.Run("normal checkout anchors to itself", func(t *testing.T) {
		d := filepath.Join(parent, "normal")
		mustMkdir(t, filepath.Join(d, ".git"))
		if got := anchorDir(d); got != d {
			t.Errorf("anchorDir = %q, want %q unchanged", got, d)
		}
	})

	t.Run("no .git at all anchors to itself", func(t *testing.T) {
		// The common case in this suite and for a planning repo that is not a git repo:
		// anchoring must be a pure no-op, which is what keeps non-worktree behavior
		// byte-identical to before.
		d := filepath.Join(parent, "nogit")
		mustMkdir(t, d)
		if got := anchorDir(d); got != d {
			t.Errorf("anchorDir = %q, want %q unchanged", got, d)
		}
	})

	t.Run("separate-git-dir is the main checkout, not a worktree", func(t *testing.T) {
		d := filepath.Join(parent, "sepdir")
		mustMkdir(t, d)
		gd := filepath.Join(parent, "detached-gitdir")
		mustMkdir(t, gd)
		writeFile(t, filepath.Join(d, ".git"), "gitdir: "+gd+"\n")
		if got := anchorDir(d); got != d {
			t.Errorf("anchorDir = %q, want %q — a --separate-git-dir repo IS the main checkout", got, d)
		}
	})

	t.Run("submodule (relative gitdir) is its own repo", func(t *testing.T) {
		super := filepath.Join(parent, "super")
		mustMkdir(t, filepath.Join(super, ".git", "modules", "sub"))
		sub := filepath.Join(super, "sub")
		mustMkdir(t, sub)
		writeFile(t, filepath.Join(sub, ".git"), "gitdir: ../.git/modules/sub\n")
		if got := anchorDir(sub); got != sub {
			t.Errorf("anchorDir = %q, want %q — a submodule is its own repo", got, sub)
		}
	})

	t.Run("bare-repo worktree has no canonical checkout", func(t *testing.T) {
		d := filepath.Join(parent, "bare-wt")
		mustMkdir(t, d)
		// bare layout: <bare>/repo.git/worktrees/<name> — note the git dir is NOT
		// named .git, so there is no working tree to anchor to.
		gd := filepath.Join(parent, "repo.git", "worktrees", "bare-wt")
		mustMkdir(t, gd)
		writeFile(t, filepath.Join(d, ".git"), "gitdir: "+gd+"\n")
		if got := anchorDir(d); got != d {
			t.Errorf("anchorDir = %q, want %q — a bare repo has no working tree", got, d)
		}
	})

	t.Run("dangling and malformed .git degrade to the input", func(t *testing.T) {
		for name, body := range map[string]string{
			"dangling":  "gitdir: " + filepath.Join(parent, "gone", ".git", "worktrees", "x") + "\n",
			"malformed": "this is not a gitdir line\n",
			"empty":     "gitdir:\n",
		} {
			d := filepath.Join(parent, "bad-"+name)
			mustMkdir(t, d)
			writeFile(t, filepath.Join(d, ".git"), body)
			if got := anchorDir(d); got != d {
				t.Errorf("%s: anchorDir = %q, want %q — must never become a failure mode", name, got, d)
			}
		}
	})
}

// writeFile writes content to path, creating parents.
func writeFile(t *testing.T, path, content string) {
	t.Helper()
	mustMkdir(t, filepath.Dir(path))
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}
