package spacehealth

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/andy-esch/taskflow/internal/config"
	"github.com/andy-esch/taskflow/internal/domain"
	"github.com/andy-esch/taskflow/internal/userconfig"
)

func TestDiagnoseRegistry_AllSupportedStates(t *testing.T) {
	home := t.TempDir()
	t.Setenv(userconfig.DirEnv, home)

	okRepo := initializedRepo(t)
	if err := os.WriteFile(filepath.Join(okRepo, domain.TasksDir, "6g0000000001-work.md"), []byte("---\nstatus: in-progress\n---\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	emptyRepo := initializedRepo(t)
	missingRepo := filepath.Join(t.TempDir(), "gone")
	notRepo := t.TempDir()
	unreadableRepo := t.TempDir()
	if err := os.WriteFile(filepath.Join(unreadableRepo, config.ConfigFile), []byte("[broken\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	mismatchRepo := initializedRepo(t)

	fixtures := []userconfig.Space{
		{ID: "ok", Path: okRepo},
		{ID: "empty", Path: emptyRepo},
		{ID: "missing", Path: missingRepo},
		{ID: "not-repo", Path: notRepo},
		{ID: "unreadable", Path: unreadableRepo},
		{ID: "mismatch", Path: mismatchRepo, VerifyID: "6gwrongid000"},
	}
	for _, fixture := range fixtures {
		if added, _, err := userconfig.AddSpace(fixture, false); err != nil || !added {
			t.Fatalf("register %s: added=%v err=%v", fixture.ID, added, err)
		}
	}

	got, err := DiagnoseRegistry()
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != len(fixtures) {
		t.Fatalf("got %d diagnoses, want %d: %+v", len(got), len(fixtures), got)
	}
	byID := make(map[string]SpaceProblem, len(got))
	for _, problem := range got {
		byID[problem.Space.ID] = problem
	}
	want := map[string]Kind{
		"ok": KindOK, "empty": KindEmpty, "missing": KindMissing,
		"not-repo": KindNotARepo, "unreadable": KindUnreadable, "mismatch": KindMismatch,
	}
	for id, kind := range want {
		problem, exists := byID[id]
		if !exists {
			t.Errorf("missing diagnosis for %s", id)
			continue
		}
		if problem.Kind != kind {
			t.Errorf("%s kind=%q, want %q: %+v", id, problem.Kind, kind, problem)
		}
		if problem.Broken() != (kind != KindOK && kind != KindEmpty) {
			t.Errorf("%s Broken()=%v for kind %q", id, problem.Broken(), kind)
		}
		if kind != KindOK && problem.Message == "" {
			t.Errorf("%s has no diagnosis message: %+v", id, problem)
		}
		if problem.Broken() && problem.Remedy == "" {
			t.Errorf("%s has no remedy: %+v", id, problem)
		}
	}
	if byID["ok"].Root == "" || byID["empty"].Root == "" || byID["mismatch"].Root == "" {
		t.Errorf("resolved states must retain their planning root: %+v", byID)
	}
	if !strings.Contains(byID["mismatch"].Message, "does not match") {
		t.Errorf("mismatch diagnosis is not specific: %+v", byID["mismatch"])
	}

	spaces, err := userconfig.Spaces()
	if err != nil || len(spaces) != len(fixtures) {
		t.Errorf("diagnosis mutated/forgot the registry: got=%v err=%v", spaces, err)
	}
}

func TestDiagnoseRegistry_MissingFileIsHealthyEmptyRegistry(t *testing.T) {
	t.Setenv(userconfig.DirEnv, t.TempDir())
	got, err := DiagnoseRegistry()
	if err != nil || len(got) != 0 {
		t.Fatalf("missing registry: got=%v err=%v", got, err)
	}
}

func initializedRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	if _, err := config.Init(dir, "", false); err != nil {
		t.Fatalf("init repo: %v", err)
	}
	return dir
}
