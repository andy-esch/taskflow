package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func mkdirs(t *testing.T, paths ...string) {
	t.Helper()
	for _, p := range paths {
		if err := os.MkdirAll(p, 0o755); err != nil {
			t.Fatal(err)
		}
	}
}

func writeConfig(t *testing.T, dir, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, ConfigFile), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

// eval resolves symlinks in p (matching Discover), so a Root comparison holds on
// platforms where t.TempDir() is itself symlinked — e.g. macOS, where /var ->
// /private/var makes the resolved Root differ from the raw temp path.
func eval(t *testing.T, p string) string {
	t.Helper()
	r, err := filepath.EvalSymlinks(p)
	if err != nil {
		t.Fatalf("eval %q: %v", p, err)
	}
	return r
}

// TestDiscover_GitFileBoundary pins the worktree/submodule case: .git is a
// FILE there, and the climb must stop at it instead of escaping into a parent
// that happens to have a planning tree.
func TestDiscover_GitFileBoundary(t *testing.T) {
	parent := t.TempDir()
	mkdirs(t, filepath.Join(parent, "tasks")) // a trap above the repo
	repo := filepath.Join(parent, "worktree")
	mkdirs(t, repo)
	if err := os.WriteFile(filepath.Join(repo, ".git"), []byte("gitdir: /elsewhere\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := Discover(repo); err == nil {
		t.Fatal("discovery must stop at a .git FILE boundary, not climb into the parent's tasks/")
	}
}

// TestTaskflowRoot_TOMLValueForms pins the line-scan parser against valid TOML
// the old trim-based extraction mangled: inline comments and quoted values.
func TestTaskflowRoot_TOMLValueForms(t *testing.T) {
	for _, tc := range []struct {
		line, want string
	}{
		{`taskflow_root = "./planning"`, "./planning"},
		{`taskflow_root = "./planning" # planning lives here`, "./planning"},
		{`taskflow_root = './planning'`, "./planning"},
		{`taskflow_root = planning # bare value with comment`, "planning"},
		{`taskflow_root = "."`, "."},
		{`taskflow_root = "unterminated`, "."}, // unset rather than guessed
		{`# taskflow_root = "commented-out"`, "."},
		{`taskflow_root = "a\"b"`, "."},  // basic string w/ escape: refused, not mis-read as `a\`
		{`taskflow_root = "a\\b"`, "."},  // ditto for \\
		{`taskflow_root = 'a\b'`, `a\b`}, // literal string: backslash is intentional, preserved
	} {
		dir := t.TempDir()
		writeConfig(t, dir, tc.line+"\n")
		if got := taskflowRoot(filepath.Join(dir, ConfigFile)); got != tc.want {
			t.Errorf("taskflowRoot(%q) = %q, want %q", tc.line, got, tc.want)
		}
	}
}

// TestDiscover_RejectsBadConfiguredRoots pins the loud-failure contract: an
// escaping or not-a-planning-tree taskflow_root errors instead of presenting a
// clean empty project that `task new` would then fork.
func TestDiscover_RejectsBadConfiguredRoots(t *testing.T) {
	// Escaping root.
	dir := t.TempDir()
	mkdirs(t, filepath.Join(dir, "repo", "tasks"))
	repo := filepath.Join(dir, "repo")
	writeConfig(t, repo, "taskflow_root = \"../outside\"\n")
	if _, err := Discover(repo); err == nil || !strings.Contains(err.Error(), "escapes") {
		t.Errorf("an escaping taskflow_root should error, got %v", err)
	}

	// Root that exists but has no tasks/.
	dir2 := t.TempDir()
	mkdirs(t, filepath.Join(dir2, "notes"))
	writeConfig(t, dir2, "taskflow_root = \"./notes\"\n")
	if _, err := Discover(dir2); err == nil || !strings.Contains(err.Error(), "tasks/") {
		t.Errorf("a non-planning taskflow_root should error mentioning tasks/, got %v", err)
	}

	// Inline comment on a good root still resolves (regression pairing).
	dir3 := t.TempDir()
	mkdirs(t, filepath.Join(dir3, "planning", "tasks"))
	writeConfig(t, dir3, "taskflow_root = \"./planning\" # note\n")
	cfg, err := Discover(dir3)
	if err != nil || cfg.Root != eval(t, filepath.Join(dir3, "planning")) {
		t.Errorf("commented good root should resolve, got %v / %v", cfg, err)
	}
}

// TestDiscover_RejectsSymlinkEscape pins L8: a taskflow_root that stays lexically
// inside the repo but is a SYMLINK pointing outside it must be rejected — the
// containment check resolves symlinks, so `planning -> /outside` can't slip past a
// no-`..` text check.
func TestDiscover_RejectsSymlinkEscape(t *testing.T) {
	outside := t.TempDir()
	mkdirs(t, filepath.Join(outside, "tasks")) // a real planning tree, but outside the repo
	repo := t.TempDir()
	if err := os.Symlink(outside, filepath.Join(repo, "planning")); err != nil {
		t.Skipf("symlinks unsupported here: %v", err)
	}
	writeConfig(t, repo, "taskflow_root = \"planning\"\n")
	if _, err := Discover(repo); err == nil || !strings.Contains(err.Error(), "escapes") {
		t.Errorf("a taskflow_root symlinked outside the repo must be rejected, got %v", err)
	}
}

// TestDiscover_ResolvesSymlinkedWorktree pins L19: discovering through a symlink to a
// planning tree resolves to the real root (physical ancestry), not the logical path.
func TestDiscover_ResolvesSymlinkedWorktree(t *testing.T) {
	real := t.TempDir()
	mkdirs(t, filepath.Join(real, "tasks"))
	link := filepath.Join(t.TempDir(), "link")
	if err := os.Symlink(real, link); err != nil {
		t.Skipf("symlinks unsupported here: %v", err)
	}
	cfg, err := Discover(link)
	if err != nil {
		t.Fatalf("discovery through a symlink should succeed: %v", err)
	}
	if cfg.Root != eval(t, real) {
		t.Errorf("Root = %q, want resolved %q", cfg.Root, eval(t, real))
	}
}
