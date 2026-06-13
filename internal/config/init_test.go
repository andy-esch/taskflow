package config

import (
	"os"
	"path/filepath"
	"testing"
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
	if err != nil || cfg.Root != root {
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
	if cfg.Root != planning {
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
	if err != nil || cfg.Root != root {
		t.Errorf("fallback discovery = %+v, %v", cfg, err)
	}
}
