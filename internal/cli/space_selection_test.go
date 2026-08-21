package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/andy-esch/taskflow/internal/config"
	"github.com/andy-esch/taskflow/internal/spacehealth"
	"github.com/andy-esch/taskflow/internal/testutil"
	"github.com/andy-esch/taskflow/internal/userconfig"
	"github.com/andy-esch/taskflow/internal/wire"
)

func runSelection(t *testing.T, args ...string) (stdout, stderr string, err error) {
	t.Helper()
	var out, errOut bytes.Buffer
	cmd := NewRootCmd(strings.NewReader(""), &out, &errOut)
	cmd.SetArgs(args)
	cmd.SetOut(&out)
	cmd.SetErr(&errOut)
	err = cmd.Execute()
	return out.String(), errOut.String(), err
}

func selectedPlanningFixture(t *testing.T) (planning, pointer string, cfg *config.Config) {
	t.Helper()
	planning = initializedSpaceRepo(t)
	path, content := testutil.TaskFixture(planning, "ready-to-start", "alpha.md",
		"---\nstatus: ready-to-start\ndescription: selected alpha\ntags: [space]\n---\n# Alpha\n")
	testutil.Write(t, path, content)
	pointer = t.TempDir()
	if _, err := config.InitPointer(pointer, planning, false); err != nil {
		t.Fatalf("init pointer: %v", err)
	}
	var err error
	cfg, err = config.Discover(planning)
	if err != nil {
		t.Fatalf("discover planning fixture: %v", err)
	}
	return planning, pointer, cfg
}

func registerSelectionFixtures(t *testing.T) (planning, pointer string, cfg *config.Config) {
	t.Helper()
	spaceConfigHome(t)
	planning, pointer, cfg = selectedPlanningFixture(t)
	for _, space := range []userconfig.Space{
		{ID: "planning", Path: planning, VerifyID: cfg.ID},
		{ID: "implementation", Path: pointer, VerifyID: cfg.ID},
	} {
		if _, _, err := userconfig.AddSpace(space, false); err != nil {
			t.Fatalf("register %s: %v", space.ID, err)
		}
	}
	return planning, pointer, cfg
}

// A local label addresses one exact registry row even when multiple entry points share
// the same durable planning identity. The selected pointer remains visible on the
// workspace receipt while discovery still routes into the shared planning tree.
func TestGlobalSpace_SelectsExactEntryPointAndCarriesReceipt(t *testing.T) {
	planning, pointer, _ := registerSelectionFixtures(t)

	out, errOut, err := runSelection(t, "--space", "implementation", "workspace", "--json")
	if err != nil {
		t.Fatalf("workspace through pointer: %v\n%s%s", err, out, errOut)
	}
	ws := decodeWorkspace(t, out).Workspace
	if ws.Space != "implementation" {
		t.Errorf("workspace.space = %q, want selected local label", ws.Space)
	}
	if ws.Source != wire.WorkspaceSourcePointer || ws.PlanningRoot != physicalPath(planning) {
		t.Errorf("pointer selection resolved incorrectly: %+v (pointer %s)", ws, pointer)
	}

	out, errOut, err = runSelection(t, "--space", "planning", "workspace", "--json")
	if err != nil {
		t.Fatalf("workspace through direct entry: %v\n%s%s", err, out, errOut)
	}
	direct := decodeWorkspace(t, out).Workspace
	if direct.Space != "planning" || direct.Source != wire.WorkspaceSourceConfig {
		t.Errorf("direct selection lost exact entry point: %+v", direct)
	}

	// Mutation receipts carry the selector too: the safety assertion survives after
	// the write, rather than existing only in argv or shell history.
	out, errOut, err = runSelection(t, "--space", "implementation", "--json", "task", "start", "alpha")
	if err != nil {
		t.Fatalf("mutate through selected pointer: %v\n%s%s", err, out, errOut)
	}
	var mutation struct {
		Workspace wire.WorkspaceJSON `json:"workspace"`
	}
	if err := json.Unmarshal([]byte(out), &mutation); err != nil {
		t.Fatalf("decode mutation receipt: %v\n%s", err, out)
	}
	if mutation.Workspace.Space != "implementation" || mutation.Workspace.PlanningRoot != physicalPath(planning) {
		t.Errorf("mutation receipt hid selected space: %+v", mutation.Workspace)
	}
}

func TestGlobalSpace_EnvironmentAndFlagPrecedence(t *testing.T) {
	registerSelectionFixtures(t)
	t.Setenv("TSKFLW_SPACE", "implementation")

	out, errOut, err := runSelection(t, "workspace", "--json")
	if err != nil {
		t.Fatalf("environment selection: %v\n%s%s", err, out, errOut)
	}
	if got := decodeWorkspace(t, out).Workspace.Space; got != "implementation" {
		t.Errorf("TSKFLW_SPACE selected %q, want implementation", got)
	}

	out, errOut, err = runSelection(t, "--space", "planning", "workspace", "--json")
	if err != nil {
		t.Fatalf("flag selection: %v\n%s%s", err, out, errOut)
	}
	if got := decodeWorkspace(t, out).Workspace.Space; got != "planning" {
		t.Errorf("--space must override TSKFLW_SPACE, got %q", got)
	}
}

func TestGlobalSpace_RejectsChdirCombination(t *testing.T) {
	planning, _, _ := registerSelectionFixtures(t)
	_, _, err := runSelection(t, "-C", planning, "--space", "planning", "workspace")
	if err == nil {
		t.Fatal("-C plus --space must fail")
	}
	if ExitCode(err) != 11 || !strings.Contains(err.Error(), "two answers to one question") {
		t.Fatalf("combination error = %v (exit %d), want clear validation error", err, ExitCode(err))
	}
}

func TestGlobalSpace_UnknownListsKnownLabels(t *testing.T) {
	registerSelectionFixtures(t)
	_, _, err := runSelection(t, "--space", "missing", "task", "list")
	if err == nil {
		t.Fatal("unknown --space must fail rather than discover from cwd")
	}
	if ExitCode(err) != 10 {
		t.Fatalf("unknown --space exit = %d, want 10: %v", ExitCode(err), err)
	}
	for _, want := range []string{`unknown space "missing"`, "implementation, planning"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("unknown-space error missing %q: %v", want, err)
		}
	}

	// Commands with best-effort repo discovery may only fall back when selection was
	// ambient. An explicit bad address remains fatal.
	_, _, err = runSelection(t, "--space", "missing", "template", "list")
	if err == nil || ExitCode(err) != 10 {
		t.Fatalf("template swallowed explicit bad space: %v", err)
	}
}

func TestGlobalSpace_BrokenEntryUsesSharedDiagnosis(t *testing.T) {
	spaceConfigHome(t)
	missing := userconfig.Space{ID: "missing", Path: filepath.Join(t.TempDir(), "gone")}
	if _, _, err := userconfig.AddSpace(missing, false); err != nil {
		t.Fatal(err)
	}
	problem := spacehealth.DiagnoseSpace(missing)
	_, _, err := runSelection(t, "--space", "missing", "workspace")
	if err == nil || ExitCode(err) != 10 {
		t.Fatalf("missing entry error = %v (exit %d), want not-found", err, ExitCode(err))
	}
	for _, want := range []string{problem.Message, problem.Remedy} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("selection error must reuse diagnosis %q: %v", want, err)
		}
	}

	repo := initializedSpaceRepo(t)
	mismatch := userconfig.Space{ID: "mismatch", Path: repo, VerifyID: "6gwrongid000"}
	if _, _, addErr := userconfig.AddSpace(mismatch, false); addErr != nil {
		t.Fatal(addErr)
	}
	problem = spacehealth.DiagnoseSpace(mismatch)
	_, _, err = runSelection(t, "--space", "mismatch", "workspace")
	if err == nil || ExitCode(err) != 14 {
		t.Fatalf("identity mismatch error = %v (exit %d), want conflict", err, ExitCode(err))
	}
	if !strings.Contains(err.Error(), problem.Message) {
		t.Errorf("mismatch selection must reuse health diagnosis %q: %v", problem.Message, err)
	}
}

func TestGlobalSpace_DuplicateLabelIsInvalidNotFirstMatchWins(t *testing.T) {
	home := spaceConfigHome(t)
	first := initializedSpaceRepo(t)
	second := initializedSpaceRepo(t)
	registry := "schema_version = 1\n\n" +
		"[[space]]\nid = \"duplicate\"\npath = " + strconvQuote(first) + "\n\n" +
		"[[space]]\nid = \"duplicate\"\npath = " + strconvQuote(second) + "\n"
	if err := os.WriteFile(filepath.Join(home, userconfig.SpacesFile), []byte(registry), 0o644); err != nil {
		t.Fatal(err)
	}
	_, _, err := runSelection(t, "--space", "duplicate", "workspace")
	if err == nil || ExitCode(err) != 11 || !strings.Contains(err.Error(), "appears more than once") {
		t.Fatalf("duplicate selector must reject the invalid registry, got %v (exit %d)", err, ExitCode(err))
	}
}

func TestComplete_GlobalSpaceAndSelectedEntities(t *testing.T) {
	registerSelectionFixtures(t)
	if got := complete(t, "--space", "pla"); !has(got, "planning") || has(got, "implementation") {
		t.Errorf("global --space completion should prefix-filter labels: %v", got)
	}
	if got := complete(t, "--space", "planning", "task", "show", ""); !has(got, "alpha") {
		t.Errorf("entity completion did not discover through selected space: %v", got)
	}
}
