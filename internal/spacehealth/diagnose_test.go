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

func TestGroup_DirectAndPointerEntriesSharePlanningIdentity(t *testing.T) {
	planning := initializedRepo(t)
	firstPointer := t.TempDir()
	secondPointer := t.TempDir()
	if _, err := config.InitPointer(firstPointer, planning, false); err != nil {
		t.Fatal(err)
	}
	if _, err := config.InitPointer(secondPointer, planning, false); err != nil {
		t.Fatal(err)
	}

	resolved, err := config.Discover(planning)
	if err != nil {
		t.Fatal(err)
	}
	problems := []SpaceProblem{
		DiagnoseSpace(userconfig.Space{ID: "impl", Path: firstPointer, VerifyID: resolved.ID}),
		DiagnoseSpace(userconfig.Space{ID: "planning", Path: planning, VerifyID: resolved.ID}),
		DiagnoseSpace(userconfig.Space{ID: "deploy", Path: secondPointer, VerifyID: resolved.ID}),
	}

	groups := Group(problems)
	if len(groups) != 1 {
		t.Fatalf("groups = %+v, want one planning identity", groups)
	}
	if groups[0].PlanningID != resolved.ID || len(groups[0].Entries) != 3 {
		t.Fatalf("group = %+v", groups[0])
	}
	wantRoles := []Role{RolePointer, RoleDirect, RolePointer}
	for i, want := range wantRoles {
		if got := groups[0].Entries[i].Role; got != want {
			t.Errorf("entry %d role = %q, want %q", i, got, want)
		}
		if got := groups[0].Entries[i].PlanningID; got != resolved.ID {
			t.Errorf("entry %d planning id = %q, want %q", i, got, resolved.ID)
		}
	}
}

func TestGroup_PointerOnlyBrokenAndLegacyFallbacks(t *testing.T) {
	planning := initializedRepo(t)
	resolved, err := config.Discover(planning)
	if err != nil {
		t.Fatal(err)
	}
	pointerA := t.TempDir()
	pointerB := t.TempDir()
	if _, err := config.InitPointer(pointerA, planning, false); err != nil {
		t.Fatal(err)
	}
	if _, err := config.InitPointer(pointerB, planning, false); err != nil {
		t.Fatal(err)
	}
	missing := filepath.Join(t.TempDir(), "missing")

	// The two pointers prove a direct checkout need not itself be registered. The
	// missing entry proves verify_id retains group membership after its path rots.
	problems := []SpaceProblem{
		DiagnoseSpace(userconfig.Space{ID: "pointer-a", Path: pointerA}),
		DiagnoseSpace(userconfig.Space{ID: "missing", Path: missing, VerifyID: resolved.ID}),
		DiagnoseSpace(userconfig.Space{ID: "pointer-b", Path: pointerB}),
	}
	groups := Group(problems)
	if len(groups) != 1 || len(groups[0].Entries) != 3 {
		t.Fatalf("pointer-only group with retained broken identity = %+v", groups)
	}
	if problems[0].PlanningID != resolved.ID || problems[0].Role != RolePointer {
		t.Errorf("legacy registry entry did not derive pointer identity: %+v", problems[0])
	}
	if problems[1].Role != RoleUnknown || problems[1].PlanningID != resolved.ID {
		t.Errorf("broken entry lost its honest role or intended identity: %+v", problems[1])
	}

	// A config-less tasks tree has no durable id. Its direct checkout and a legacy
	// pointer still group by the physical resolved root; unrelated broken entries do not.
	idless := t.TempDir()
	if err := os.Mkdir(filepath.Join(idless, domain.TasksDir), 0o755); err != nil {
		t.Fatal(err)
	}
	legacyPointer := t.TempDir()
	if _, err := config.InitPointer(legacyPointer, idless, false); err != nil {
		t.Fatal(err)
	}
	legacy := Group([]SpaceProblem{
		DiagnoseSpace(userconfig.Space{ID: "idless", Path: idless}),
		DiagnoseSpace(userconfig.Space{ID: "idless-pointer", Path: legacyPointer}),
		DiagnoseSpace(userconfig.Space{ID: "gone-a", Path: filepath.Join(t.TempDir(), "gone")}),
		DiagnoseSpace(userconfig.Space{ID: "gone-b", Path: filepath.Join(t.TempDir(), "gone")}),
	})
	if len(legacy) != 3 || len(legacy[0].Entries) != 2 {
		t.Fatalf("legacy fallback groups = %+v, want root pair plus two isolated failures", legacy)
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
