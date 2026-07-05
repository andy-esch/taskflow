package config

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/andy-esch/taskflow/internal/domain"
)

// TestInitPointer covers pointer mode: a validated planning_repo config is
// written, NO tree is scaffolded, and Discover then follows the pointer.
func TestInitPointer(t *testing.T) {
	parent := t.TempDir()
	impl := filepath.Join(parent, "impl")
	planning := filepath.Join(parent, "planning")
	mustMkdir(t, impl)
	mustMkdir(t, filepath.Join(planning, "tasks"))

	created, err := InitPointer(impl, "../planning", false)
	if err != nil {
		t.Fatalf("InitPointer: %v", err)
	}
	if len(created) != 1 || created[0] != ConfigFile {
		t.Errorf("should create only the config, got %v", created)
	}
	b, err := os.ReadFile(filepath.Join(impl, ConfigFile))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(b), `planning_repo = "../planning"`) {
		t.Errorf("config missing planning_repo:\n%s", b)
	}
	if isDir(filepath.Join(impl, "tasks")) {
		t.Error("pointer mode must NOT scaffold a tasks/ tree")
	}
	// Discover from impl now resolves OUT to the external planning repo.
	cfg, err := Discover(impl)
	if err != nil || cfg.Root != evalOr(planning) {
		t.Errorf("Discover should follow the pointer to %q, got %v / %v", planning, cfg, err)
	}
	// Idempotent: a second call writes nothing.
	if again, err := InitPointer(impl, "../planning", false); err != nil || len(again) != 0 {
		t.Errorf("re-init should be a no-op, got %v / %v", again, err)
	}
}

func TestInitPointer_RejectsBadTarget(t *testing.T) {
	impl := t.TempDir()
	// ../planning doesn't exist / has no tasks/ → loud, nothing written.
	if _, err := InitPointer(impl, "../planning", false); err == nil || !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("bad target should wrap ErrValidation, got %v", err)
	}
	if fileExists(filepath.Join(impl, ConfigFile)) {
		t.Error("a rejected pointer init must leave no config behind")
	}
	// A blank path is a validation error too.
	if _, err := InitPointer(impl, "  ", false); err == nil || !errors.Is(err, domain.ErrValidation) {
		t.Errorf("empty planning_repo should error, got %v", err)
	}
}

func TestInitPointer_DryRun(t *testing.T) {
	parent := t.TempDir()
	impl := filepath.Join(parent, "impl")
	mustMkdir(t, impl)
	mustMkdir(t, filepath.Join(parent, "planning", "tasks"))
	created, err := InitPointer(impl, "../planning", true)
	if err != nil || len(created) != 1 {
		t.Fatalf("dry-run should report the config, got %v / %v", created, err)
	}
	if fileExists(filepath.Join(impl, ConfigFile)) {
		t.Error("dry-run must not write the config")
	}
}

// TestInitPointer_CreatesMissingDir: pointer mode creates a missing --path dir
// (parity with scaffold Init), but only AFTER the target validates.
func TestInitPointer_CreatesMissingDir(t *testing.T) {
	parent := t.TempDir()
	mustMkdir(t, filepath.Join(parent, "planning", "tasks"))
	impl := filepath.Join(parent, "newimpl") // does not exist yet
	created, err := InitPointer(impl, "../planning", false)
	if err != nil {
		t.Fatalf("InitPointer should create a missing dir, got %v", err)
	}
	if len(created) != 1 || !fileExists(filepath.Join(impl, ConfigFile)) {
		t.Errorf("pointer config not written into the created dir: %v", created)
	}
	// A bad target must NOT create the dir (validation precedes mkdir).
	bad := filepath.Join(parent, "nope-impl")
	if _, err := InitPointer(bad, "../planning-missing", false); err == nil {
		t.Fatal("bad target should error")
	}
	if isDir(bad) {
		t.Error("a rejected pointer init must not create the target dir")
	}
}

// TestInitPointer_ModeCollision: an existing config with a DIFFERENT target (or a
// scaffold) is refused (ErrConflict), not silently dropped; the same target is an
// idempotent no-op.
func TestInitPointer_ModeCollision(t *testing.T) {
	parent := t.TempDir()
	impl := filepath.Join(parent, "impl")
	mustMkdir(t, impl)
	mustMkdir(t, filepath.Join(parent, "planning-a", "tasks"))
	mustMkdir(t, filepath.Join(parent, "planning-b", "tasks"))

	if _, err := InitPointer(impl, "../planning-a", false); err != nil {
		t.Fatal(err)
	}
	// Same target → idempotent no-op.
	if again, err := InitPointer(impl, "../planning-a", false); err != nil || len(again) != 0 {
		t.Errorf("same-target re-init should be a no-op, got %v / %v", again, err)
	}
	// Different target → ErrConflict, original preserved.
	if _, err := InitPointer(impl, "../planning-b", false); err == nil || !errors.Is(err, domain.ErrConflict) {
		t.Errorf("re-pointing to a new target should be ErrConflict, got %v", err)
	}
	if b, _ := os.ReadFile(filepath.Join(impl, ConfigFile)); !strings.Contains(string(b), "planning-a") {
		t.Errorf("the original target must be preserved on a refused re-point:\n%s", b)
	}
	// A scaffold config + pointer init → ErrConflict (mode switch refused).
	scaf := filepath.Join(parent, "scaf")
	if _, err := Init(scaf, false); err != nil {
		t.Fatal(err)
	}
	if _, err := InitPointer(scaf, "../planning-a", false); err == nil || !errors.Is(err, domain.ErrConflict) {
		t.Errorf("pointer init over a scaffold config should be ErrConflict, got %v", err)
	}
}

// TestInit_RefusesOverPointer: scaffolding over an existing pointer config is
// refused (ErrConflict) — it would orphan a local tree while discovery follows
// the pointer.
func TestInit_RefusesOverPointer(t *testing.T) {
	parent := t.TempDir()
	impl := filepath.Join(parent, "impl")
	mustMkdir(t, impl)
	mustMkdir(t, filepath.Join(parent, "planning", "tasks"))
	if _, err := InitPointer(impl, "../planning", false); err != nil {
		t.Fatal(err)
	}
	if _, err := Init(impl, false); err == nil || !errors.Is(err, domain.ErrConflict) {
		t.Errorf("scaffold over a pointer config should be ErrConflict, got %v", err)
	}
	if isDir(filepath.Join(impl, "tasks")) {
		t.Error("a refused scaffold must not create a tasks/ tree")
	}
}

// TestAddTrackedRepo: surgical append with physical-path dedup, comments preserved.
func TestAddTrackedRepo(t *testing.T) {
	parent := t.TempDir()
	planning := filepath.Join(parent, "planning")
	mustMkdir(t, filepath.Join(planning, "tasks"))
	if _, err := Init(planning, false); err != nil {
		t.Fatal(err)
	}
	if added, err := AddTrackedRepo(planning, "../impl-a", false); err != nil || !added {
		t.Fatalf("first add should succeed, got %v / %v", added, err)
	}
	if b, _ := os.ReadFile(filepath.Join(planning, ConfigFile)); !strings.Contains(string(b), `tracked_repos = ["../impl-a"]`) {
		t.Errorf("tracked_repos not written:\n%s", b)
	}
	// Same path again → no-op; an ABSOLUTE path to the same dir → physical dedup.
	if added, _ := AddTrackedRepo(planning, "../impl-a", false); added {
		t.Error("re-adding the same path should be a no-op")
	}
	if added, _ := AddTrackedRepo(planning, filepath.Join(parent, "impl-a"), false); added {
		t.Error("an absolute path to an already-tracked repo must dedup")
	}
	// A distinct repo appends in order; comments survive the surgical edit.
	if added, _ := AddTrackedRepo(planning, "../impl-b", false); !added {
		t.Error("a distinct repo should be added")
	}
	b, _ := os.ReadFile(filepath.Join(planning, ConfigFile))
	if !strings.Contains(string(b), `tracked_repos = ["../impl-a", "../impl-b"]`) {
		t.Errorf("second repo not appended in order:\n%s", b)
	}
	if !strings.Contains(string(b), "# tskflwctl planning config") {
		t.Errorf("surgical edit must preserve comments:\n%s", b)
	}
}

func TestAddTrackedRepo_MissingConfigIsNoop(t *testing.T) {
	dir := t.TempDir() // no .tskflwctl.toml
	if added, err := AddTrackedRepo(dir, "../impl", false); err != nil || added {
		t.Errorf("missing config should be a silent no-op, got %v / %v", added, err)
	}
	if fileExists(filepath.Join(dir, ConfigFile)) {
		t.Error("must not create a config")
	}
}

// TestLinkBack: the reverse of a planning_repo pointer — the impl is recorded in
// the planning repo's tracked_repos as the planning→impl relative path.
func TestLinkBack(t *testing.T) {
	parent := t.TempDir()
	impl := filepath.Join(parent, "desirelines")
	planning := filepath.Join(parent, "desirelines-planning")
	mustMkdir(t, impl)
	mustMkdir(t, filepath.Join(planning, "tasks"))
	if _, err := Init(planning, false); err != nil {
		t.Fatal(err)
	}
	back, err := LinkBack(impl, "../desirelines-planning", false)
	if err != nil || back != "../desirelines" {
		t.Fatalf("back-link = %q (err %v), want ../desirelines", back, err)
	}
	if b, _ := os.ReadFile(filepath.Join(planning, ConfigFile)); !strings.Contains(string(b), `"../desirelines"`) {
		t.Errorf("planning tracked_repos not updated:\n%s", b)
	}
	if back, err := LinkBack(impl, "../desirelines-planning", false); err != nil || back != "" {
		t.Errorf("re-link should be a no-op, got %q / %v", back, err)
	}
}

func TestLinkBack_MissingPlanningConfigIsNoop(t *testing.T) {
	parent := t.TempDir()
	impl := filepath.Join(parent, "impl")
	mustMkdir(t, impl)
	mustMkdir(t, filepath.Join(parent, "planning", "tasks")) // has tasks/, NO config
	if back, err := LinkBack(impl, "../planning", false); err != nil || back != "" {
		t.Errorf("missing planning config should be a silent no-op, got %q / %v", back, err)
	}
}

// TestSetTrackedReposInText: surgical replace / append-when-absent / empty list.
func TestSetTrackedReposInText(t *testing.T) {
	in := "# header\ntaskflow_root = \".\"\n# note\ntracked_repos = []\n"
	out := setTrackedReposInText(in, []string{"../a", "../b"})
	if !strings.Contains(out, `tracked_repos = ["../a", "../b"]`) ||
		!strings.Contains(out, "# header") || !strings.Contains(out, "# note") ||
		!strings.Contains(out, `taskflow_root = "."`) {
		t.Errorf("surgical replace failed:\n%s", out)
	}
	out2 := setTrackedReposInText("taskflow_root = \".\"\n", []string{"../a"})
	if !strings.Contains(out2, `taskflow_root = "."`) || !strings.Contains(out2, `tracked_repos = ["../a"]`) {
		t.Errorf("append-when-absent failed:\n%s", out2)
	}
	if out3 := setTrackedReposInText("tracked_repos = [\"x\"]\n", nil); !strings.Contains(out3, "tracked_repos = []") {
		t.Errorf("empty should render []:\n%s", out3)
	}
	// A comment containing the literal `tracked_repos = [...]` must NOT be edited —
	// only the real, line-anchored assignment is rewritten (and exactly once).
	commented := "# example: tracked_repos = [\"x\"]\ntracked_repos = []\n"
	co := setTrackedReposInText(commented, []string{"../a"})
	if !strings.Contains(co, `# example: tracked_repos = ["x"]`) {
		t.Errorf("a bracketed comment must be preserved verbatim:\n%s", co)
	}
	if strings.Count(co, `tracked_repos = ["../a"]`) != 1 {
		t.Errorf("only the real assignment should be rewritten once:\n%s", co)
	}
	// A `]` inside a quoted value must not end the array early.
	br := setTrackedReposInText(`tracked_repos = ["../a]b"]`+"\n", []string{"../a]b", "../c"})
	if !strings.Contains(br, `tracked_repos = ["../a]b", "../c"]`) {
		t.Errorf("a bracket inside a value should not truncate the span:\n%s", br)
	}
}

// TestAddTrackedRepo_BracketInPath is the regression for the review BLOCKER: a
// stored path containing `]` must survive a SECOND edit without corrupting the
// TOML (the old `[^\]]*` regex spliced into the middle of such a value).
func TestAddTrackedRepo_BracketInPath(t *testing.T) {
	parent := t.TempDir()
	planning := filepath.Join(parent, "planning")
	mustMkdir(t, filepath.Join(planning, "tasks"))
	if _, err := Init(planning, false); err != nil {
		t.Fatal(err)
	}
	if _, err := AddTrackedRepo(planning, "../imp]l", false); err != nil {
		t.Fatal(err)
	}
	if _, err := AddTrackedRepo(planning, "../impl-b", false); err != nil {
		t.Fatalf("a second edit must not corrupt a ]-bearing config: %v", err)
	}
	cf, err := readConfigFile(filepath.Join(planning, ConfigFile))
	if err != nil {
		t.Fatalf("config corrupted by a ]-path: %v", err)
	}
	if len(cf.TrackedRepos) != 2 || cf.TrackedRepos[0] != "../imp]l" || cf.TrackedRepos[1] != "../impl-b" {
		t.Errorf("entries wrong after a ]-path edit: %v", cf.TrackedRepos)
	}
}

func mustMkdir(t *testing.T, p string) {
	t.Helper()
	if err := os.MkdirAll(p, 0o755); err != nil {
		t.Fatal(err)
	}
}

func TestInit(t *testing.T) {
	root := t.TempDir()

	created, err := Init(root, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(created) == 0 {
		t.Fatal("expected dirs/config to be created")
	}

	for _, d := range []string{"tasks", "epics", "projects", "audits"} {
		if !isDir(filepath.Join(root, filepath.FromSlash(d))) {
			t.Errorf("missing dir %s", d)
		}
	}
	if !fileExists(filepath.Join(root, ConfigFile)) {
		t.Error("config file not written")
	}

	// Discover should now resolve this root.
	cfg, err := Discover(root)
	if err != nil || cfg.Root != eval(t, root) {
		t.Errorf("Discover after init = %+v, %v", cfg, err)
	}

	// Idempotent: a second run creates nothing.
	again, err := Init(root, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(again) != 0 {
		t.Errorf("second Init created %v, want none", again)
	}
}

// TestInitScaffoldsEntityDirs pins the flat layout (ADR-0003 §4): `init` creates the
// entity parents (tasks/epics/audits/projects) and NO per-status or per-bucket subdirs
// (the flat store never reads them; a file dropped in one would be invisible).
func TestInitScaffoldsEntityDirs(t *testing.T) {
	root := t.TempDir()
	if _, err := Init(root, false); err != nil {
		t.Fatal(err)
	}
	for _, d := range []string{domain.TasksDir, domain.EpicsDir, domain.AuditsDir, domain.ProjectsDir} {
		if !isDir(filepath.Join(root, d)) {
			t.Errorf("init did not scaffold %s/", d)
		}
	}
	for _, st := range domain.AllStatuses() {
		if isDir(filepath.Join(root, "tasks", st.Dir())) {
			t.Errorf("init should NOT scaffold a per-status dir tasks/%s under the flat layout", st.Dir())
		}
	}
	for _, b := range domain.AllAuditBuckets() {
		if isDir(filepath.Join(root, "audits", b.Dir())) {
			t.Errorf("init should NOT scaffold a per-bucket dir audits/%s under the flat layout", b.Dir())
		}
	}
}

// TestInitGitkeepsEveryDir pins that init drops a .gitkeep in each scaffolded
// dir, so an empty planning tree is git-committable.
func TestInitGitkeepsEveryDir(t *testing.T) {
	root := t.TempDir()
	if _, err := Init(root, false); err != nil {
		t.Fatal(err)
	}
	for _, d := range []string{domain.TasksDir, domain.EpicsDir, domain.AuditsDir, domain.ProjectsDir} {
		keep := filepath.Join(root, filepath.FromSlash(d), ".gitkeep")
		if !fileExists(keep) {
			t.Errorf("init did not write %s/.gitkeep", d)
		}
	}
}

// TestInitRetrofitsGitkeep pins that re-running init on a tree whose dirs exist
// but lack .gitkeep adds the keep (repairs older trees).
func TestInitRetrofitsGitkeep(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, "tasks")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	created, err := Init(root, false)
	if err != nil {
		t.Fatal(err)
	}
	if !fileExists(filepath.Join(dir, ".gitkeep")) {
		t.Error("init should add a .gitkeep to a pre-existing dir that lacked one")
	}
	// The keep is reported as created even though the dir itself already existed.
	var sawKeep bool
	for _, c := range created {
		if c == "tasks/.gitkeep" {
			sawKeep = true
		}
	}
	if !sawKeep {
		t.Errorf("the retrofitted .gitkeep should be reported in created: %v", created)
	}
}

func TestDiscover_ConfigAnchorsAndHonorsRoot(t *testing.T) {
	repo := t.TempDir()
	// Config at the repo root points taskflow_root at a planning/ subdir that
	// holds the actual tasks/ tree.
	// Use the "./planning" form (with leading ./), matching a real root config.
	if err := os.WriteFile(filepath.Join(repo, ConfigFile),
		[]byte("taskflow_root = \"./planning\"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	planning := filepath.Join(repo, "planning")
	if err := os.MkdirAll(filepath.Join(planning, "tasks"), 0o755); err != nil {
		t.Fatal(err)
	}

	cfg, err := Discover(repo)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Root != eval(t, planning) {
		t.Errorf("config taskflow_root not honored: got %q, want %q", cfg.Root, planning)
	}
}

func TestDiscover_FallsBackToTasksDir(t *testing.T) {
	// A hand-made repo with tasks/ but no config still resolves.
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "tasks"), 0o755); err != nil {
		t.Fatal(err)
	}
	cfg, err := Discover(root)
	if err != nil || cfg.Root != eval(t, root) {
		t.Errorf("fallback discovery = %+v, %v", cfg, err)
	}
}
