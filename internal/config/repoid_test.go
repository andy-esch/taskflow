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
	if _, err := Init(root, false); err != nil {
		t.Fatal(err)
	}
	if idOf(t, root) == "" {
		t.Error("init must mint a durable id into the scaffolded config")
	}
}

// TestInit_BackfillsMissingID pins the migration path. Re-running init already repairs a
// scaffold repo (it is how a missing .gitkeep is restored), so it is also how every repo
// predating the mint gets an id — no new command surface.
func TestInit_BackfillsMissingID(t *testing.T) {
	root := t.TempDir()
	if _, err := Init(root, false); err != nil {
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

	if _, err := Init(root, false); err != nil {
		t.Fatal(err)
	}
	first := idOf(t, root)
	if first == "" {
		t.Fatal("re-running init must backfill a missing id")
	}
	if _, err := Init(root, false); err != nil {
		t.Fatal(err)
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
	if _, err := Init(plan, false); err != nil {
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
		if _, err := Init(decoy, false); err != nil {
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
	if _, err := Init(decoy, false); err != nil {
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
