package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/andy-esch/taskflow/internal/domain"
)

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
