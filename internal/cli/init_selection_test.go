package cli

import (
	"encoding/json"
	"maps"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/andy-esch/taskflow/internal/config"
	"github.com/andy-esch/taskflow/internal/testutil"
	"github.com/andy-esch/taskflow/internal/wire"
)

// runInitIn changes the real caller cwd, rather than injecting an extra -C. The
// bootstrap selector tests must distinguish those two locations explicitly.
func runInitIn(t *testing.T, caller string, args ...string) (string, string, error) {
	t.Helper()
	t.Chdir(caller)
	return runSelection(t, args...)
}

func TestInitChdirTargetsFreshDirectoryAndRepairWithoutTouchingCaller(t *testing.T) {
	home := spaceConfigHome(t)
	caller := initializedSpaceRepo(t)
	target := filepath.Join(t.TempDir(), "new-space")
	callerBefore, homeBefore := testutil.SnapshotTree(t, caller), testutil.SnapshotTree(t, home)
	for _, repair := range []bool{false, true} {
		if repair {
			if err := os.Remove(filepath.Join(target, "planning", "threads")); err != nil {
				// Init leaves a .gitkeep, so remove only this fixture placeholder first.
				if err := os.Remove(filepath.Join(target, "planning", "threads", ".gitkeep")); err != nil {
					t.Fatal(err)
				}
				if err := os.Remove(filepath.Join(target, "planning", "threads")); err != nil {
					t.Fatal(err)
				}
			}
		}
		out, errOut, err := runInitIn(t, caller, "-C", target, "init", "--taskflow-root", "planning", "--no-register", "--json")
		if err != nil {
			t.Fatalf("repair=%v: %v\n%s%s", repair, err, out, errOut)
		}
		var receipt wire.InitEnvelope
		if err := json.Unmarshal([]byte(out), &receipt); err != nil {
			t.Fatal(err)
		}
		if receipt.Root != target || receipt.Mode != "scaffold" {
			t.Fatalf("wrong init target: %+v", receipt)
		}
		if _, err := os.Stat(filepath.Join(target, "planning", "threads")); err != nil {
			t.Fatalf("selected target was not scaffolded/repaired: %v", err)
		}
		if !maps.Equal(callerBefore, testutil.SnapshotTree(t, caller)) || !maps.Equal(homeBefore, testutil.SnapshotTree(t, home)) {
			t.Fatal("init changed caller or registry instead of only the selected target")
		}
	}
}

func TestInitSelectorConflictsAreWriteFree(t *testing.T) {
	home := spaceConfigHome(t)
	caller, target := t.TempDir(), t.TempDir()
	for _, selectors := range [][]string{
		{"--path", target, "-C", target},
		{"--path", target, "--space", "missing"},
		{"-C", target, "--space", "missing"},
		{"--path", ""}, {"-C", ""}, {"--space", ""},
	} {
		t.Run(strings.Join(selectors, "/"), func(t *testing.T) {
			beforeCaller, beforeTarget, beforeHome := testutil.SnapshotTree(t, caller), testutil.SnapshotTree(t, target), testutil.SnapshotTree(t, home)
			args := append([]string{"init", "--no-register", "--json"}, selectors...)
			out, _, err := runInitIn(t, caller, args...)
			if err == nil || ExitCode(err) != 11 || out != "" {
				t.Fatalf("invalid selectors accepted: out=%q err=%v", out, err)
			}
			if !maps.Equal(beforeCaller, testutil.SnapshotTree(t, caller)) || !maps.Equal(beforeTarget, testutil.SnapshotTree(t, target)) ||
				!maps.Equal(beforeHome, testutil.SnapshotTree(t, home)) {
				t.Fatal("invalid selectors caused writes")
			}
		})
	}
}

func TestInitRegistrySelectionAndPathPrecedence(t *testing.T) {
	planning, pointer, _ := registerSelectionFixtures(t)
	caller := t.TempDir()
	t.Setenv("TSKFLW_SPACE", "implementation")
	for _, tc := range []struct {
		name, target, mode string
		args               []string
	}{
		{"environment", pointer, "pointer", nil},
		{"space flag", planning, "scaffold", []string{"--space", "planning"}},
		{"path overrides environment", t.TempDir(), "scaffold", []string{"--path"}},
		{"chdir overrides environment", t.TempDir(), "scaffold", []string{"-C"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			args := append([]string{"init", "--json", "--no-register"}, tc.args...)
			if tc.name == "path overrides environment" || tc.name == "chdir overrides environment" {
				args = append(args, tc.target)
			}
			out, errOut, err := runInitIn(t, caller, args...)
			if err != nil {
				t.Fatalf("%v\n%s%s", err, out, errOut)
			}
			var receipt wire.InitEnvelope
			if err := json.Unmarshal([]byte(out), &receipt); err != nil {
				t.Fatal(err)
			}
			if physicalPath(receipt.Root) != physicalPath(tc.target) || receipt.Mode != tc.mode {
				t.Fatalf("wrong selected entry point: %+v", receipt)
			}
		})
	}
	before := testutil.SnapshotTree(t, caller)
	for _, explicit := range []bool{false, true} {
		t.Setenv("TSKFLW_SPACE", "missing")
		args := []string{"init", "--no-register"}
		if explicit {
			args = append(args, "--space", "missing")
		}
		if _, _, err := runInitIn(t, caller, args...); err == nil || ExitCode(err) != 10 {
			t.Fatalf("bad space must refuse, not initialize cwd: %v", err)
		}
	}
	if !maps.Equal(before, testutil.SnapshotTree(t, caller)) {
		t.Fatal("bad registry selection changed caller")
	}
}

func TestInitChdirPointerResolvesPlanningRepoRelativeToSelectedDirectory(t *testing.T) {
	spaceConfigHome(t)
	parent := t.TempDir()
	planning := filepath.Join(parent, "planning")
	runRoot(t, "init", "--path", planning, "--no-register")
	pointer, caller := filepath.Join(parent, "implementation"), t.TempDir()
	out, errOut, err := runInitIn(t, caller, "-C", pointer, "init", "--planning-repo", "../planning", "--no-link-back", "--no-register", "--json")
	if err != nil {
		t.Fatalf("%v\n%s%s", err, out, errOut)
	}
	if _, err := os.Stat(filepath.Join(pointer, config.ConfigFile)); err != nil {
		t.Fatal(err)
	}
	if entries, err := os.ReadDir(caller); err != nil || len(entries) != 0 {
		t.Fatalf("pointer init touched caller: entries=%v err=%v", entries, err)
	}
}
