package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/andy-esch/taskflow/internal/config"
	"github.com/andy-esch/taskflow/internal/userconfig"
	"github.com/andy-esch/taskflow/internal/wire"
)

func TestConfigShowReportsScopesAndEffectiveProvenance(t *testing.T) {
	home := t.TempDir()
	t.Setenv(userconfig.DirEnv, home)
	if err := os.WriteFile(filepath.Join(home, userconfig.FileName), []byte(
		"[theme]\nname = \"catppuccin\"\n[pager]\nenabled = false\ncommand = \"delta\"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	repo := t.TempDir()
	if _, err := config.Init(repo, "planning", false); err != nil {
		t.Fatal(err)
	}
	// A repository command override should win field-by-field while pager enabled
	// continues to inherit from the user scope.
	path := filepath.Join(repo, config.ConfigFile)
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, append(b, []byte("\n[pager]\ncommand = \"less -R\"\n")...), 0o644); err != nil {
		t.Fatal(err)
	}

	out := runRoot(t, "-C", repo, "config", "show", "--json")
	var env wire.ConfigEnvelope
	if err := json.Unmarshal([]byte(out), &env); err != nil {
		t.Fatalf("decode config JSON: %v\n%s", err, out)
	}
	if env.Repository.Mode != string("scaffold") || env.Repository.TaskflowRoot != "./planning" {
		t.Errorf("repository = %+v", env.Repository)
	}
	if env.Effective.Theme.Value != "catppuccin" || env.Effective.Theme.Source != "user" {
		t.Errorf("theme = %+v", env.Effective.Theme)
	}
	if env.Effective.PagerEnabled.Value || env.Effective.PagerEnabled.Source != "user" {
		t.Errorf("pager enabled = %+v", env.Effective.PagerEnabled)
	}
	if env.Effective.PagerCommand.Value != "less -R" || env.Effective.PagerCommand.Source != "repository" {
		t.Errorf("pager command = %+v", env.Effective.PagerCommand)
	}

	bare := runRoot(t, "-C", repo, "config", "--json")
	if bare != out {
		t.Errorf("bare config must be identical to config show\nconfig: %s\nshow: %s", bare, out)
	}
}

func TestConfigMigrateDirectDryRunApplyAndNoop(t *testing.T) {
	repo := t.TempDir()
	if err := os.MkdirAll(filepath.Join(repo, "tasks"), 0o755); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(repo, config.ConfigFile)
	legacy := "# keep\ntaskflow_root = \".\"\nunknown = \"yes\"\n"
	if err := os.WriteFile(path, []byte(legacy), 0o644); err != nil {
		t.Fatal(err)
	}

	dry := runRoot(t, "-C", repo, "config", "migrate", "--dry-run", "--json")
	var preview wire.ConfigMigrationEnvelope
	if err := json.Unmarshal([]byte(dry), &preview); err != nil {
		t.Fatalf("decode preview: %v\n%s", err, dry)
	}
	if !preview.DryRun || !preview.Changed || len(preview.Steps) != 1 || preview.Steps[0].Key != "id" {
		t.Fatalf("preview = %+v", preview)
	}
	if preview.Workspace.RepoID != preview.Steps[0].Value {
		t.Errorf("preview workspace must carry the would-be repo id: %+v", preview.Workspace)
	}
	if b, _ := os.ReadFile(path); string(b) != legacy {
		t.Fatalf("dry-run wrote the file:\n%s", b)
	}

	applied := runRoot(t, "-C", repo, "config", "migrate", "--json")
	var receipt wire.ConfigMigrationEnvelope
	if err := json.Unmarshal([]byte(applied), &receipt); err != nil {
		t.Fatal(err)
	}
	if receipt.DryRun || !receipt.Changed || receipt.Workspace.RepoID == "" {
		t.Fatalf("receipt = %+v", receipt)
	}
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(b), "# keep") || !strings.Contains(string(b), `unknown = "yes"`) {
		t.Fatalf("migration did not preserve unrelated text:\n%s", b)
	}

	current := runRoot(t, "-C", repo, "config", "migrate", "--json")
	var noop wire.ConfigMigrationEnvelope
	if err := json.Unmarshal([]byte(current), &noop); err != nil {
		t.Fatal(err)
	}
	if noop.Changed || len(noop.Steps) != 0 {
		t.Errorf("second migration = %+v", noop)
	}
}

func TestInitExistingLegacyPointerHandsOffToConfigMigrate(t *testing.T) {
	parent := t.TempDir()
	planning := filepath.Join(parent, "planning")
	impl := filepath.Join(parent, "impl")
	if _, err := config.Init(planning, "", false); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(impl, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(impl, config.ConfigFile), []byte(
		"# old pointer\nplanning_repo = \"../planning\"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := config.AddTrackedRepo(planning, "../impl", false); err != nil {
		t.Fatal(err)
	}

	out := runRoot(t, "init", "--path", impl, "--color=never", "--no-register")
	if !strings.Contains(out, "already initialized") || !strings.Contains(out, "config migrate") {
		t.Fatalf("init should hand off to migration:\n%s", out)
	}
	if b, _ := os.ReadFile(filepath.Join(impl, config.ConfigFile)); strings.Contains(string(b), "planning_repo_id") {
		t.Fatalf("init silently migrated the pointer:\n%s", b)
	}

	migrated := runRoot(t, "-C", impl, "config", "migrate", "--json")
	var receipt wire.ConfigMigrationEnvelope
	if err := json.Unmarshal([]byte(migrated), &receipt); err != nil {
		t.Fatalf("decode migration: %v\n%s", err, migrated)
	}
	if len(receipt.Steps) != 1 || receipt.Steps[0].Key != "planning_repo_id" {
		t.Fatalf("pointer migration = %+v", receipt)
	}
	if b, _ := os.ReadFile(filepath.Join(impl, config.ConfigFile)); !strings.Contains(string(b), "planning_repo_id") || !strings.Contains(string(b), "# old pointer") {
		t.Fatalf("pointer was not migrated surgically:\n%s", b)
	}
}

func TestInitExistingDirectReportsWithoutMigrating(t *testing.T) {
	repo := t.TempDir()
	if err := os.MkdirAll(filepath.Join(repo, "tasks"), 0o755); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(repo, config.ConfigFile)
	legacy := "# legacy direct\ntaskflow_root = \".\"\n"
	if err := os.WriteFile(path, []byte(legacy), 0o644); err != nil {
		t.Fatal(err)
	}
	out := runRoot(t, "init", "--path", repo, "--color=never", "--no-register")
	if !strings.Contains(out, "already initialized") || !strings.Contains(out, "repo-id") || !strings.Contains(out, "config migrate") {
		t.Fatalf("legacy direct init should report the migration handoff:\n%s", out)
	}
	if after, _ := os.ReadFile(path); string(after) != legacy {
		t.Fatal("bare init migrated a direct config")
	}
	runRoot(t, "-C", repo, "config", "migrate")
	out = runRoot(t, "init", "--path", repo, "--color=never", "--no-register")
	if strings.Contains(out, "configuration update available") {
		t.Fatalf("current direct init reported a migration:\n%s", out)
	}
}

func TestConfigDoctorCanonicalAndTopLevelCompatibilityAlias(t *testing.T) {
	_, impl := linkedPair(t, true)
	canonical := runRoot(t, "-C", impl, "config", "doctor", "--json")
	compat := runRoot(t, "-C", impl, "doctor", "--json")
	if canonical != compat {
		t.Errorf("doctor alias drifted\nconfig doctor: %s\ndoctor: %s", canonical, compat)
	}
	help := runRoot(t, "--help")
	if !strings.Contains(help, "  config") {
		t.Errorf("root help does not expose config:\n%s", help)
	}
	if strings.Contains(help, "  doctor") {
		t.Errorf("root help should hide the compatibility alias:\n%s", help)
	}
}

func TestConfigEditRefusesNonInteractiveAndDryRunInvocation(t *testing.T) {
	repo := t.TempDir()
	if _, err := config.Init(repo, "", false); err != nil {
		t.Fatal(err)
	}
	_, err := runRootRC(t, "-C", repo, "config", "edit")
	if err == nil || ExitCode(err) != 11 || !strings.Contains(err.Error(), "interactive terminal") {
		t.Fatalf("non-interactive config edit should fail clearly with validation: %v", err)
	}
	_, err = runRootRC(t, "-C", repo, "config", "edit", "--dry-run")
	if err == nil || ExitCode(err) != 11 || !strings.Contains(err.Error(), "dry-run") {
		t.Fatalf("interactive dry-run should be rejected explicitly: %v", err)
	}
}

func TestConfigCommandsRejectMalformedRepositoryConfigWithoutWriting(t *testing.T) {
	repo := t.TempDir()
	if err := os.MkdirAll(filepath.Join(repo, "tasks"), 0o755); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(repo, config.ConfigFile)
	malformed := "[theme\nname = \"neon\"\n"
	if err := os.WriteFile(path, []byte(malformed), 0o644); err != nil {
		t.Fatal(err)
	}
	for _, args := range [][]string{
		{"-C", repo, "config", "show"},
		{"-C", repo, "config", "migrate", "--dry-run"},
	} {
		_, err := runRootRC(t, args...)
		if err == nil || ExitCode(err) != 11 {
			t.Fatalf("%v should reject malformed config with validation: %v", args, err)
		}
	}
	if after, _ := os.ReadFile(path); string(after) != malformed {
		t.Fatal("failed config command changed malformed TOML")
	}
}
