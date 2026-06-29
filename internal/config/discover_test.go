package config

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/andy-esch/taskflow/internal/domain"
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

// TestReadConfigFile_TaskflowRootForms pins taskflow_root parsing now that a real
// TOML decoder backs it. Valid TOML the old trim-based scanner mangled (inline
// comments, quoted/literal values) still parses cleanly; the escape cases that the
// hand-scanner deliberately REFUSED (returning ".") now decode correctly, because a
// real decoder handles \" and \\ losslessly. (Empty taskflow_root is normalized to
// "." downstream in configuredRoot, not here.)
func TestReadConfigFile_TaskflowRootForms(t *testing.T) {
	for _, tc := range []struct {
		line, want string
	}{
		{`taskflow_root = "./planning"`, "./planning"},
		{`taskflow_root = "./planning" # planning lives here`, "./planning"},
		{`taskflow_root = './planning'`, "./planning"},
		{`taskflow_root = "."`, "."},
		{`# taskflow_root = "commented-out"`, ""}, // commented out: absent -> "" (defaulted later)
		{`taskflow_root = "a\"b"`, `a"b`},         // basic string escape now decoded, not refused
		{`taskflow_root = "a\\b"`, `a\b`},         // ditto for \\
		{`taskflow_root = 'a\b'`, `a\b`},          // literal string: backslash preserved
	} {
		dir := t.TempDir()
		writeConfig(t, dir, tc.line+"\n")
		cf, err := readConfigFile(filepath.Join(dir, ConfigFile))
		if err != nil {
			t.Errorf("readConfigFile(%q) errored: %v", tc.line, err)
			continue
		}
		if cf.TaskflowRoot != tc.want {
			t.Errorf("readConfigFile(%q).TaskflowRoot = %q, want %q", tc.line, cf.TaskflowRoot, tc.want)
		}
	}
}

// TestReadConfigFile_MalformedIsLoud pins the loud-failure contract for the real
// parser: TOML the old hand-scanner silently tolerated (a bare unquoted value, an
// unterminated quote) now wraps ErrValidation instead of falling back to a default.
func TestReadConfigFile_MalformedIsLoud(t *testing.T) {
	for _, line := range []string{
		`taskflow_root = planning # bare unquoted value`,
		`taskflow_root = "unterminated`,
		`taskflow_root = = "twoeq"`,
	} {
		dir := t.TempDir()
		writeConfig(t, dir, line+"\n")
		_, err := readConfigFile(filepath.Join(dir, ConfigFile))
		if err == nil {
			t.Errorf("readConfigFile(%q) should error on malformed TOML", line)
			continue
		}
		if !errors.Is(err, domain.ErrValidation) {
			t.Errorf("readConfigFile(%q) error should wrap ErrValidation, got %v", line, err)
		}
	}
}

// TestReadConfigFile_NewKeys pins the two newly-exposed keys: planning_repo (a
// string) and tracked_repos (an array) parse RAW, and absent keys default to the
// zero value ("" / nil) — no out-of-tree resolution happens here.
func TestReadConfigFile_NewKeys(t *testing.T) {
	t.Run("both present", func(t *testing.T) {
		dir := t.TempDir()
		writeConfig(t, dir, `taskflow_root = "."
planning_repo = "../planning-repo"
tracked_repos = ["../impl-a", "../impl-b"]
`)
		cf, err := readConfigFile(filepath.Join(dir, ConfigFile))
		if err != nil {
			t.Fatal(err)
		}
		if cf.PlanningRepo != "../planning-repo" {
			t.Errorf("PlanningRepo = %q, want %q", cf.PlanningRepo, "../planning-repo")
		}
		want := []string{"../impl-a", "../impl-b"}
		if len(cf.TrackedRepos) != len(want) {
			t.Fatalf("TrackedRepos = %v, want %v", cf.TrackedRepos, want)
		}
		for i := range want {
			if cf.TrackedRepos[i] != want[i] {
				t.Errorf("TrackedRepos[%d] = %q, want %q", i, cf.TrackedRepos[i], want[i])
			}
		}
	})

	t.Run("absent keys default", func(t *testing.T) {
		dir := t.TempDir()
		writeConfig(t, dir, `taskflow_root = "."`+"\n")
		cf, err := readConfigFile(filepath.Join(dir, ConfigFile))
		if err != nil {
			t.Fatal(err)
		}
		if cf.PlanningRepo != "" {
			t.Errorf("absent planning_repo should be \"\", got %q", cf.PlanningRepo)
		}
		if cf.TrackedRepos != nil {
			t.Errorf("absent tracked_repos should be nil, got %v", cf.TrackedRepos)
		}
	})

	t.Run("empty array", func(t *testing.T) {
		dir := t.TempDir()
		writeConfig(t, dir, `taskflow_root = "."`+"\ntracked_repos = []\n")
		cf, err := readConfigFile(filepath.Join(dir, ConfigFile))
		if err != nil {
			t.Fatal(err)
		}
		if len(cf.TrackedRepos) != 0 {
			t.Errorf("empty tracked_repos should be empty, got %v", cf.TrackedRepos)
		}
	})
}

// TestDiscover_ExposesNewKeysRaw pins that Discover carries planning_repo and
// tracked_repos through to Config RAW and UNRESOLVED — no validation, no
// out-of-tree discovery — while Root resolution is unchanged.
// TestDiscover_ExposesTrackedReposRaw pins that tracked_repos rides through
// Discover RAW and unresolved (a downstream task reads it). planning_repo's raw
// carry — alongside its resolution into Root — is covered by
// TestDiscover_FollowsPlanningRepo; here there's no planning_repo, so Root stays
// the in-tree taskflow_root.
func TestDiscover_ExposesTrackedReposRaw(t *testing.T) {
	repo := t.TempDir()
	mkdirs(t, filepath.Join(repo, "tasks"))
	writeConfig(t, repo, `taskflow_root = "."
tracked_repos = ["../impl-a", "../impl-b"]
`)
	cfg, err := Discover(repo)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Root != eval(t, repo) {
		t.Errorf("Root = %q, want %q", cfg.Root, eval(t, repo))
	}
	want := []string{"../impl-a", "../impl-b"}
	if len(cfg.TrackedRepos) != len(want) {
		t.Fatalf("TrackedRepos = %v, want %v", cfg.TrackedRepos, want)
	}
	for i := range want {
		if cfg.TrackedRepos[i] != want[i] {
			t.Errorf("TrackedRepos[%d] = %q, want %q", i, cfg.TrackedRepos[i], want[i])
		}
	}
}

// TestDiscover_ParsesPager pins the [pager] table: command verbatim and enabled as
// a tri-state (an explicit false is distinct from unset/nil → default on).
func TestDiscover_ParsesPager(t *testing.T) {
	t.Run("explicit", func(t *testing.T) {
		repo := t.TempDir()
		mkdirs(t, filepath.Join(repo, "tasks"))
		writeConfig(t, repo, `taskflow_root = "."
[pager]
enabled = false
command = "bat -p"
`)
		cfg, err := Discover(repo)
		if err != nil {
			t.Fatal(err)
		}
		if cfg.Pager.Enabled == nil || *cfg.Pager.Enabled != false {
			t.Errorf("Pager.Enabled = %v, want explicit false", cfg.Pager.Enabled)
		}
		if cfg.Pager.Command != "bat -p" {
			t.Errorf("Pager.Command = %q, want %q", cfg.Pager.Command, "bat -p")
		}
	})
	t.Run("absent table → nil/empty (default on downstream)", func(t *testing.T) {
		repo := t.TempDir()
		mkdirs(t, filepath.Join(repo, "tasks"))
		writeConfig(t, repo, `taskflow_root = "."`+"\n")
		cfg, err := Discover(repo)
		if err != nil {
			t.Fatal(err)
		}
		if cfg.Pager.Enabled != nil {
			t.Errorf("Pager.Enabled = %v, want nil when [pager] is absent", *cfg.Pager.Enabled)
		}
		if cfg.Pager.Command != "" {
			t.Errorf("Pager.Command = %q, want empty", cfg.Pager.Command)
		}
	})
}

// TestDiscover_ParsesTheme pins the [theme] table: the name rides through Discover
// onto Config.Theme (empty when the table is absent → the default theme downstream).
func TestDiscover_ParsesTheme(t *testing.T) {
	t.Run("explicit name", func(t *testing.T) {
		repo := t.TempDir()
		mkdirs(t, filepath.Join(repo, "tasks"))
		writeConfig(t, repo, `taskflow_root = "."
[theme]
name = "neon"
`)
		cfg, err := Discover(repo)
		if err != nil {
			t.Fatal(err)
		}
		if cfg.Theme.Name != "neon" {
			t.Errorf("Theme.Name = %q, want %q", cfg.Theme.Name, "neon")
		}
	})
	t.Run("absent table → empty name", func(t *testing.T) {
		repo := t.TempDir()
		mkdirs(t, filepath.Join(repo, "tasks"))
		writeConfig(t, repo, `taskflow_root = "."`+"\n")
		cfg, err := Discover(repo)
		if err != nil {
			t.Fatal(err)
		}
		if cfg.Theme.Name != "" {
			t.Errorf("Theme.Name = %q, want empty when [theme] is absent", cfg.Theme.Name)
		}
	})
}

// TestDiscover_FollowsPlanningRepo pins the sanctioned out-of-tree escape:
// planning_repo points discovery at an EXTERNAL planning repo (here a sibling),
// which taskflow_root may not. The raw value still rides along for linkbacks.
func TestDiscover_FollowsPlanningRepo(t *testing.T) {
	parent := t.TempDir()
	impl := filepath.Join(parent, "impl")
	planning := filepath.Join(parent, "planning")
	mkdirs(t, impl, filepath.Join(planning, "tasks"))
	writeConfig(t, impl, "planning_repo = \"../planning\"\n")

	cfg, err := Discover(impl)
	if err != nil {
		t.Fatalf("planning_repo should resolve to the sibling, got %v", err)
	}
	if cfg.Root != eval(t, planning) {
		t.Errorf("Root = %q, want the external planning repo %q", cfg.Root, eval(t, planning))
	}
	if cfg.PlanningRepo != "../planning" {
		t.Errorf("raw PlanningRepo = %q, want %q", cfg.PlanningRepo, "../planning")
	}
}

// TestDiscover_PlanningRepoMustBeValid pins the require+validate contract: a
// planning_repo whose target has no tasks/ (or is missing) is a loud error, not
// a clean empty tree.
func TestDiscover_PlanningRepoMustBeValid(t *testing.T) {
	// Sibling exists but has NO tasks/.
	parent := t.TempDir()
	impl := filepath.Join(parent, "impl")
	mkdirs(t, impl, filepath.Join(parent, "planning"))
	writeConfig(t, impl, "planning_repo = \"../planning\"\n")
	if _, err := Discover(impl); err == nil || !strings.Contains(err.Error(), "tasks/") {
		t.Errorf("planning_repo without tasks/ must error mentioning tasks/, got %v", err)
	}

	// A wholly missing target is equally loud.
	impl2 := t.TempDir()
	writeConfig(t, impl2, "planning_repo = \"../nope\"\n")
	if _, err := Discover(impl2); err == nil || !strings.Contains(err.Error(), "tasks/") {
		t.Errorf("missing planning_repo target must error, got %v", err)
	}
}

// TestDiscover_PlanningRepoVsTaskflowRoot pins precedence: planning_repo wins
// over a default taskflow_root, but a NON-default taskflow_root alongside it is a
// loud conflict (two roots), not a silent override.
func TestDiscover_PlanningRepoVsTaskflowRoot(t *testing.T) {
	parent := t.TempDir()
	impl := filepath.Join(parent, "impl")
	planning := filepath.Join(parent, "planning")
	mkdirs(t, impl, filepath.Join(planning, "tasks"))

	// Explicit-but-default taskflow_root = "." yields to planning_repo.
	writeConfig(t, impl, "taskflow_root = \".\"\nplanning_repo = \"../planning\"\n")
	cfg, err := Discover(impl)
	if err != nil || cfg.Root != eval(t, planning) {
		t.Errorf("default taskflow_root should yield to planning_repo, got %v / %v", cfg, err)
	}

	// "./" is the SAME in-tree root as "." — it must coexist too, not falsely
	// trip the "keep one" conflict on a literal string compare.
	writeConfig(t, impl, "taskflow_root = \"./\"\nplanning_repo = \"../planning\"\n")
	if cfg, err := Discover(impl); err != nil || cfg.Root != eval(t, planning) {
		t.Errorf(`taskflow_root "./" should coexist with planning_repo, got %v / %v`, cfg, err)
	}

	// A non-default taskflow_root next to planning_repo is a conflict.
	impl2 := filepath.Join(t.TempDir(), "impl")
	mkdirs(t, filepath.Join(impl2, "local", "tasks"))
	writeConfig(t, impl2, "taskflow_root = \"./local\"\nplanning_repo = \"../planning\"\n")
	if _, err := Discover(impl2); err == nil || !strings.Contains(err.Error(), "keep one") {
		t.Errorf("planning_repo + a non-default taskflow_root should conflict, got %v", err)
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
