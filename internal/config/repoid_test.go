package config

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/andy-esch/taskflow/internal/domain"
)

// idOf reads the durable id out of a config file.
func idOf(t *testing.T, dir string) string {
	t.Helper()
	cf, err := readConfigFile(filepath.Join(dir, ConfigFile))
	if err != nil {
		t.Fatal(err)
	}
	return cf.ID
}

// TestInit_MintsDurableID: a scaffolded planning repo carries an id from birth.
func TestInit_MintsDurableID(t *testing.T) {
	root := t.TempDir()
	if _, err := Init(root, "", false); err != nil {
		t.Fatal(err)
	}
	if idOf(t, root) == "" {
		t.Error("init must mint a durable id into the scaffolded config")
	}
}

// TestMigrate_BackfillsMissingID pins the lifecycle split: init repairs scaffold
// directories but does not alter an existing config; Migrate owns the id backfill.
func TestMigrate_BackfillsMissingID(t *testing.T) {
	root := t.TempDir()
	if _, err := Init(root, "", false); err != nil {
		t.Fatal(err)
	}
	// strip the id, as a pre-mint repo would be
	path := filepath.Join(root, ConfigFile)
	b, _ := os.ReadFile(path)
	var kept []string
	for _, l := range strings.Split(string(b), "\n") {
		if !strings.HasPrefix(strings.TrimSpace(l), "id = ") {
			kept = append(kept, l)
		}
	}
	if err := os.WriteFile(path, []byte(strings.Join(kept, "\n")), 0o644); err != nil {
		t.Fatal(err)
	}
	if idOf(t, root) != "" {
		t.Fatal("precondition: the id should be gone")
	}

	if _, err := Init(root, "", false); err != nil {
		t.Fatal(err)
	}
	if got := idOf(t, root); got != "" {
		t.Fatalf("re-running init must not migrate config, got id %q", got)
	}
	result, err := migrateWithIDGen(root, false, func() string { return "6g245fixedid" })
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Steps) != 1 || result.Steps[0].Kind != MigrationRepoID {
		t.Fatalf("migration steps = %+v", result.Steps)
	}
	first := idOf(t, root)
	if first != "6g245fixedid" {
		t.Fatalf("migration id = %q", first)
	}
	againResult, err := Migrate(root, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(againResult.Steps) != 0 {
		t.Fatalf("second migration must be a no-op, got %+v", againResult.Steps)
	}
	if again := idOf(t, root); again != first {
		t.Errorf("backfill must be idempotent: %q then %q", first, again)
	}
}

// TestInsertTopLevelKey_NeverLandsInsideATable is the correctness trap of a text-level
// insert: appending at the end of a file that has real tables would silently make the key
// a sub-key of the LAST table (`pager.id`) rather than a top-level one.
func TestInsertTopLevelKey_NeverLandsInsideATable(t *testing.T) {
	withTable := "taskflow_root = \".\"\n\n[pager]\nenabled = true\n"
	got := insertTopLevelKey(withTable, "id = \"x\"")
	if strings.Index(got, "id = \"x\"") > strings.Index(got, "[pager]") {
		t.Errorf("the key must be inserted BEFORE the first table, got:\n%s", got)
	}
	noTable := "taskflow_root = \".\"\n"
	if got := insertTopLevelKey(noTable, "id = \"x\""); !strings.Contains(got, "id = \"x\"") {
		t.Errorf("a table-less file should still receive the key, got:\n%s", got)
	}
}

// planningAndPointer builds a planning repo + an impl repo pointing at it, and returns
// both plus the planning repo's minted id.
func planningAndPointer(t *testing.T) (plan, impl, planID string) {
	t.Helper()
	parent := t.TempDir()
	plan = filepath.Join(parent, "planning")
	if _, err := Init(plan, "", false); err != nil {
		t.Fatal(err)
	}
	impl = filepath.Join(parent, "impl")
	mustMkdir(t, impl)
	if _, err := InitPointer(impl, "../planning", false); err != nil {
		t.Fatal(err)
	}
	return plan, impl, idOf(t, plan)
}

// TestInitPointer_RecordsTargetID: writing a pointer opts it into verification.
func TestInitPointer_RecordsTargetID(t *testing.T) {
	_, impl, planID := planningAndPointer(t)
	cf, err := readConfigFile(filepath.Join(impl, ConfigFile))
	if err != nil {
		t.Fatal(err)
	}
	if cf.PlanningRepoID != planID {
		t.Errorf("pointer recorded %q, want the target's id %q", cf.PlanningRepoID, planID)
	}
	if _, err := Discover(impl); err != nil {
		t.Errorf("a matching id must resolve cleanly: %v", err)
	}
}

// TestVerify_MismatchAndMissingBothFail is the core contract. The MISSING case is the
// load-bearing one: the hazard this closes is an unrelated planning repo sitting where a
// committed relative path lands, and such a repo has no id at all — tolerating that would
// leave the hole open while appearing to close it.
func TestVerify_MismatchAndMissingBothFail(t *testing.T) {
	t.Run("a different repo is refused", func(t *testing.T) {
		plan, impl, _ := planningAndPointer(t)
		decoy := plan + "-decoy"
		if _, err := Init(decoy, "", false); err != nil {
			t.Fatal(err)
		}
		if err := os.Rename(plan, plan+"-moved"); err != nil {
			t.Fatal(err)
		}
		if err := os.Rename(decoy, plan); err != nil {
			t.Fatal(err)
		}
		_, err := Discover(impl)
		if !errors.Is(err, domain.ErrConflict) {
			t.Errorf("a different planning repo must be ErrConflict (exit 14), got %v", err)
		}
	})

	t.Run("a target with no id is refused", func(t *testing.T) {
		plan, impl, _ := planningAndPointer(t)
		if err := os.RemoveAll(plan); err != nil {
			t.Fatal(err)
		}
		mustMkdir(t, filepath.Join(plan, domain.TasksDir)) // bare tasks/, no config, no id
		_, err := Discover(impl)
		if !errors.Is(err, domain.ErrConflict) {
			t.Errorf("a target carrying no id must be ErrConflict, got %v", err)
		}
	})
}

// TestVerify_LegacyPointerIsSilent is what makes this non-breaking: a pointer that never
// opted in resolves exactly as before, even against a repo with a different id — and must
// NOT warn, or it would recreate the always-firing noise this epic just removed.
func TestVerify_LegacyPointerIsSilent(t *testing.T) {
	plan, impl, _ := planningAndPointer(t)
	// strip the expectation, as a pre-mint pointer would be
	path := filepath.Join(impl, ConfigFile)
	b, _ := os.ReadFile(path)
	var kept []string
	for _, l := range strings.Split(string(b), "\n") {
		if !strings.HasPrefix(strings.TrimSpace(l), "planning_repo_id") {
			kept = append(kept, l)
		}
	}
	if err := os.WriteFile(path, []byte(strings.Join(kept, "\n")), 0o644); err != nil {
		t.Fatal(err)
	}
	// ...and swap the target for an entirely different planning repo.
	decoy := plan + "-decoy"
	if _, err := Init(decoy, "", false); err != nil {
		t.Fatal(err)
	}
	if err := os.RemoveAll(plan); err != nil {
		t.Fatal(err)
	}
	if err := os.Rename(decoy, plan); err != nil {
		t.Fatal(err)
	}

	cfg, err := Discover(impl)
	if err != nil {
		t.Fatalf("a legacy pointer must resolve exactly as before, got %v", err)
	}
	if cfg.ID == "" {
		t.Error("Config.ID should still report the resolved tree's id")
	}
}

// TestMigrate_BackfillsExistingPointer closes the migration gap reported from real use
// while keeping bootstrap and maintenance separate.
func TestMigrate_BackfillsExistingPointer(t *testing.T) {
	parent := t.TempDir()
	plan := filepath.Join(parent, "planning")
	if _, err := Init(plan, "", false); err != nil {
		t.Fatal(err)
	}
	impl := filepath.Join(parent, "impl")
	mustMkdir(t, impl)

	// a pointer as written BEFORE durable ids existed: target, no expectation
	writeConfig(t, impl, "planning_repo = \"../planning\"\n")
	if cf, _ := readConfigFile(filepath.Join(impl, ConfigFile)); cf.PlanningRepoID != "" {
		t.Fatal("precondition: the legacy pointer should carry no id")
	}

	created, err := InitPointer(impl, "../planning", false)
	if err != nil {
		t.Fatalf("re-init on an existing pointer must not error: %v", err)
	}
	if len(created.Created) != 0 {
		t.Errorf("re-init must not migrate the config, got %v", created)
	}
	result, err := Migrate(impl, false)
	if err != nil {
		t.Fatalf("migrate existing pointer: %v", err)
	}
	if len(result.Steps) != 1 || result.Steps[0].Kind != MigrationPlanningRepoID {
		t.Fatalf("migration steps = %+v", result.Steps)
	}
	cf, err := readConfigFile(filepath.Join(impl, ConfigFile))
	if err != nil {
		t.Fatal(err)
	}
	if cf.PlanningRepoID != idOf(t, plan) {
		t.Errorf("planning_repo_id = %q, want the target's id %q", cf.PlanningRepoID, idOf(t, plan))
	}
	if cf.PlanningRepo != "../planning" {
		t.Errorf("the existing target must be preserved, got %q", cf.PlanningRepo)
	}

	// idempotent: a second migration changes nothing
	again, err := Migrate(impl, false)
	if err != nil || len(again.Steps) != 0 {
		t.Errorf("a second migration must be a no-op, got %v / %v", again, err)
	}
}

func TestMigrate_DryRunReportsPostStateWithoutWriting(t *testing.T) {
	root := t.TempDir()
	mustMkdir(t, filepath.Join(root, domain.TasksDir))
	writeConfig(t, root, "# keep me\ntaskflow_root = \".\"\n\n[theme]\nname = \"neon\"\n")
	before, err := os.ReadFile(filepath.Join(root, ConfigFile))
	if err != nil {
		t.Fatal(err)
	}

	result, err := migrateWithIDGen(root, true, func() string { return "6g245dryrunid" })
	if err != nil {
		t.Fatal(err)
	}
	if !result.DryRun || len(result.Steps) != 1 || result.Steps[0].Value != "6g245dryrunid" {
		t.Fatalf("dry-run result = %+v", result)
	}
	after, err := os.ReadFile(filepath.Join(root, ConfigFile))
	if err != nil {
		t.Fatal(err)
	}
	if string(after) != string(before) {
		t.Fatalf("dry-run changed the file:\n%s", after)
	}
}

func TestMigrate_LegacyPointerRequiresTargetMigrationFirst(t *testing.T) {
	parent := t.TempDir()
	plan := filepath.Join(parent, "planning")
	mustMkdir(t, filepath.Join(plan, domain.TasksDir))
	writeConfig(t, plan, "taskflow_root = \".\"\n")
	impl := filepath.Join(parent, "impl")
	mustMkdir(t, impl)
	writeConfig(t, impl, "planning_repo = \"../planning\"\n")

	pending, err := PendingMigrations(impl)
	if err != nil || len(pending) != 1 || pending[0] != MigrationPlanningRepoID {
		t.Fatalf("pending=%v err=%v", pending, err)
	}
	before, _ := os.ReadFile(filepath.Join(impl, ConfigFile))
	if _, err := Migrate(impl, true); err == nil || !strings.Contains(err.Error(), "config migrate") {
		t.Fatalf("pointer migration should name the target-first remedy: %v", err)
	}
	if after, _ := os.ReadFile(filepath.Join(impl, ConfigFile)); string(after) != string(before) {
		t.Fatal("failed pointer migration changed the config")
	}
}

func TestMigrate_PreservesExistingTOMLText(t *testing.T) {
	root := t.TempDir()
	mustMkdir(t, filepath.Join(root, domain.TasksDir))
	writeConfig(t, root, "# keep me\nunknown = \"value\"\ntaskflow_root = \".\"\n\n[theme]\nname = \"neon\"\n")

	if _, err := migrateWithIDGen(root, false, func() string { return "6g245fixedid" }); err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(filepath.Join(root, ConfigFile))
	if err != nil {
		t.Fatal(err)
	}
	s := string(b)
	for _, want := range []string{"# keep me", `unknown = "value"`, `[theme]`, `name = "neon"`} {
		if !strings.Contains(s, want) {
			t.Errorf("migration dropped %q:\n%s", want, s)
		}
	}
	if strings.Index(s, `id = "6g245fixedid"`) > strings.Index(s, "[theme]") {
		t.Errorf("id landed inside a table:\n%s", s)
	}
}

// TestInitPointer_BackfillWaitsForTargetID: if the TARGET has no id yet, there is nothing
// to record. Never invent one — the id must be the target's own, minted there.
func TestInitPointer_BackfillWaitsForTargetID(t *testing.T) {
	parent := t.TempDir()
	plan := filepath.Join(parent, "planning")
	mustMkdir(t, filepath.Join(plan, domain.TasksDir)) // bare tasks/, no config, no id
	impl := filepath.Join(parent, "impl")
	mustMkdir(t, impl)
	writeConfig(t, impl, "planning_repo = \"../planning\"\n")

	created, err := InitPointer(impl, "../planning", false)
	if err != nil {
		t.Fatalf("must not error when the target has no id yet: %v", err)
	}
	if len(created.Created) != 0 {
		t.Errorf("nothing to record, so nothing should be reported, got %v", created)
	}
	if cf, _ := readConfigFile(filepath.Join(impl, ConfigFile)); cf.PlanningRepoID != "" {
		t.Errorf("must not invent an id, got %q", cf.PlanningRepoID)
	}
}

// TestDescribe covers the read surface `init` uses to REPORT an existing setup instead of
// re-asking settled questions. It must never error and never guess: a dir with no config
// is simply "not initialized".
func TestDescribe(t *testing.T) {
	t.Run("no config", func(t *testing.T) {
		if _, ok := Describe(t.TempDir()); ok {
			t.Error("a dir with no config must report not-initialized")
		}
	})

	t.Run("scaffold reports its root and id", func(t *testing.T) {
		root := t.TempDir()
		if _, err := Init(root, "planning", false); err != nil {
			t.Fatal(err)
		}
		d, ok := Describe(root)
		if !ok {
			t.Fatal("an initialized repo should describe")
		}
		if d.TaskflowRoot != "./planning" {
			t.Errorf("taskflow_root = %q, want ./planning", d.TaskflowRoot)
		}
		if d.ID == "" {
			t.Error("a scaffolded repo carries a durable id")
		}
		if d.PlanningRepo != "" {
			t.Errorf("a scaffold is not a pointer, got planning_repo %q", d.PlanningRepo)
		}
	})

	t.Run("pointer reports its target and whether it verifies", func(t *testing.T) {
		plan, impl, planID := planningAndPointer(t)
		d, ok := Describe(impl)
		if !ok {
			t.Fatal("a pointer repo should describe")
		}
		if d.PlanningRepo != "../planning" {
			t.Errorf("planning_repo = %q", d.PlanningRepo)
		}
		if d.PlanningRepoID != planID {
			t.Errorf("planning_repo_id = %q, want %q — this is what tells a user the pointer verifies",
				d.PlanningRepoID, planID)
		}
		_ = plan
	})

	t.Run("a malformed config degrades to not-initialized", func(t *testing.T) {
		dir := t.TempDir()
		writeConfig(t, dir, "this is not = valid toml [[[\n")
		if _, ok := Describe(dir); ok {
			t.Error("an unreadable config must not be reported as a described layout")
		}
	})
}
