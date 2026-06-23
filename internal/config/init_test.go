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

	for _, d := range []string{"tasks/ready-to-start", "tasks/deferred", "epics", "projects", "audits/open"} {
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

// TestInitScaffoldsEveryStatusAndBucket is the sync guard: `init` must create a
// directory for every domain status and audit bucket, so adding one to the enum
// can't silently ship with init not scaffolding it (while the watcher already
// watches it). Derives expectations from the same enums Init does.
func TestInitScaffoldsEveryStatusAndBucket(t *testing.T) {
	root := t.TempDir()
	if _, err := Init(root, false); err != nil {
		t.Fatal(err)
	}
	for _, st := range domain.AllStatuses() {
		if !isDir(filepath.Join(root, "tasks", st.Dir())) {
			t.Errorf("init did not scaffold tasks/%s", st.Dir())
		}
	}
	for _, b := range domain.AllAuditBuckets() {
		if !isDir(filepath.Join(root, "audits", b.Dir())) {
			t.Errorf("init did not scaffold audits/%s", b.Dir())
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
	for _, d := range append(domain.TaskStatusDirs(),
		append([]string{domain.EpicsDir, domain.ProjectsDir}, domain.AuditBucketDirs()...)...) {
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
	dir := filepath.Join(root, "tasks", "ready-to-start")
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
		if c == "tasks/ready-to-start/.gitkeep" {
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
