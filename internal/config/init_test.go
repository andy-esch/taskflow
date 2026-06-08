package config

import (
	"path/filepath"
	"testing"
)

func TestInit(t *testing.T) {
	root := t.TempDir()

	created, err := Init(root)
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
	again, err := Init(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(again) != 0 {
		t.Errorf("second Init created %v, want none", again)
	}
}
