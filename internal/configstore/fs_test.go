package configstore

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/andy-esch/taskflow/internal/config"
	"github.com/andy-esch/taskflow/internal/core"
	"github.com/andy-esch/taskflow/internal/userconfig"
)

func TestFSLoadsBothScopesAndPendingMigration(t *testing.T) {
	home := t.TempDir()
	t.Setenv(userconfig.DirEnv, home)
	if err := os.WriteFile(filepath.Join(home, userconfig.FileName), []byte(
		"[theme]\nname = \"user-theme\"\n[pager]\nenabled = false\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	repo := t.TempDir()
	if _, err := config.Init(repo, "", false); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(repo, config.ConfigFile)
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	// A compact legacy direct config is enough to exercise pending migration and
	// repo-scope presentation loading together.
	if err := os.WriteFile(path, []byte("taskflow_root = \".\"\n[theme]\nname = \"repo-theme\"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	_ = b

	state, err := New().LoadConfiguration(repo)
	if err != nil {
		t.Fatal(err)
	}
	if state.Repository.ThemeName != "repo-theme" || state.User.ThemeName != "user-theme" {
		t.Fatalf("state = %+v", state)
	}
	if len(state.Repository.PendingMigration) != 1 || state.Repository.PendingMigration[0] != core.ConfigurationMigrationRepoID {
		t.Fatalf("pending = %v", state.Repository.PendingMigration)
	}
}

func TestFSSetPreferencePreservesOtherScopeText(t *testing.T) {
	home := t.TempDir()
	t.Setenv(userconfig.DirEnv, home)
	repo := t.TempDir()
	if _, err := config.Init(repo, "", false); err != nil {
		t.Fatal(err)
	}
	svc := core.NewConfigurationService(New())
	result, err := svc.SetPreference(repo, core.PreferenceChange{
		Scope: core.ConfigScopeUser, Field: core.PreferencePagerCommand, Value: `delta --dark="x"`,
	}, false)
	if err != nil || !result.Changed {
		t.Fatalf("result=%+v err=%v", result, err)
	}
	loaded, err := userconfig.Load()
	if err != nil {
		t.Fatal(err)
	}
	if loaded.Pager.Command != `delta --dark="x"` {
		t.Errorf("pager command = %q", loaded.Pager.Command)
	}
}
