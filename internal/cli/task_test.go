package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// setupRepo builds a temporary planning tree and returns its root.
func setupRepo(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	write := func(status, name, content string) {
		dir := filepath.Join(root, "tasks", status)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write("ready-to-start", "alpha.md", "---\nstatus: ready-to-start\ndescription: alpha\n---\n# Alpha\n")
	write("in-progress", "beta.md", "---\nstatus: in-progress\ndescription: beta\n---\n# Beta\n")
	return root
}

// runRoot executes the root command in-process against args, capturing output.
func runRoot(t *testing.T, args ...string) string {
	t.Helper()
	var out bytes.Buffer
	cmd := NewRootCmd(&out, &out)
	cmd.SetArgs(args)
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute %v: %v\noutput:\n%s", args, err, out.String())
	}
	return out.String()
}

func TestTaskList_JSON(t *testing.T) {
	root := setupRepo(t)
	out := runRoot(t, "-C", root, "task", "list", "--json")

	var got struct {
		SchemaVersion string           `json:"schema_version"`
		Tasks         []map[string]any `json:"tasks"`
	}
	if err := json.Unmarshal([]byte(out), &got); err != nil {
		t.Fatalf("invalid json: %v\n%s", err, out)
	}
	if got.SchemaVersion == "" {
		t.Error("missing schema_version")
	}
	if len(got.Tasks) != 2 {
		t.Fatalf("want 2 tasks, got %d", len(got.Tasks))
	}
}

func TestTaskList_StatusFilter(t *testing.T) {
	root := setupRepo(t)
	out := runRoot(t, "-C", root, "task", "list", "--status", "in-progress")
	if !bytes.Contains([]byte(out), []byte("beta")) {
		t.Errorf("want beta in output:\n%s", out)
	}
	if bytes.Contains([]byte(out), []byte("alpha")) {
		t.Errorf("alpha should be filtered out:\n%s", out)
	}
}

// TestTaskList_InvalidFiltersExit11 pins that a typo'd --status or an unknown
// --epic is a loud validation error (exit 11) naming the problem — not a
// silently empty list indistinguishable from an empty bucket.
func TestTaskList_InvalidFiltersExit11(t *testing.T) {
	root := setupRepo(t)
	for _, tc := range []struct {
		args []string
		want string
	}{
		{[]string{"task", "list", "--status", "bogus"}, "ready-to-start"}, // enumerates valid statuses
		{[]string{"task", "list", "--epic", "nope"}, `unknown epic "nope"`},
	} {
		var out bytes.Buffer
		cmd := NewRootCmd(&out, &out)
		cmd.SetArgs(append([]string{"-C", root}, tc.args...))
		cmd.SetOut(&out)
		cmd.SetErr(&out)
		err := cmd.Execute()
		if err == nil {
			t.Fatalf("%v should fail", tc.args)
		}
		if ExitCode(err) != 11 {
			t.Errorf("%v: want exit 11, got %d (%v)", tc.args, ExitCode(err), err)
		}
		if !bytes.Contains([]byte(err.Error()), []byte(tc.want)) {
			t.Errorf("%v: error should mention %q, got %q", tc.args, tc.want, err.Error())
		}
	}
}

// TestTaskMove_InvalidStatusEnumerates pins that the move error lists the
// valid statuses (it previously just said `invalid status "x"`).
func TestTaskMove_InvalidStatusEnumerates(t *testing.T) {
	root := setupRepo(t)
	var out bytes.Buffer
	cmd := NewRootCmd(&out, &out)
	cmd.SetArgs([]string{"-C", root, "task", "move", "alpha", "limbo"})
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	err := cmd.Execute()
	if err == nil {
		t.Fatal("move to an invalid status should fail")
	}
	if ExitCode(err) != 11 {
		t.Errorf("want exit 11, got %d", ExitCode(err))
	}
	if !bytes.Contains([]byte(err.Error()), []byte("in-progress")) {
		t.Errorf("error should enumerate valid statuses, got %q", err.Error())
	}
}

// TestCreate_ContractValidation pins the D1/D2 decisions at the CLI seam:
// tagless task creation and an off-vocabulary epic status both exit 11.
func TestCreate_ContractValidation(t *testing.T) {
	root := setupRepo(t)
	if err := os.MkdirAll(filepath.Join(root, "epics"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "epics", "e1.md"),
		[]byte("---\nstatus: planning\ndescription: e\n---\n# E\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		args []string
		want string
	}{
		{[]string{"task", "new", "Tagless", "--epic", "e1", "--description", "d"}, "tag"},
		{[]string{"epic", "new", "Weird", "--description", "d", "--status", "bananas"}, "planning"}, // enumerates
	} {
		var out bytes.Buffer
		cmd := NewRootCmd(&out, &out)
		cmd.SetArgs(append([]string{"-C", root}, tc.args...))
		cmd.SetOut(&out)
		cmd.SetErr(&out)
		err := cmd.Execute()
		if err == nil || ExitCode(err) != 11 {
			t.Errorf("%v: want exit 11, got %v", tc.args, err)
		}
		if err != nil && !strings.Contains(err.Error(), tc.want) {
			t.Errorf("%v: error should mention %q, got %q", tc.args, tc.want, err.Error())
		}
	}
}

// TestInit_JSON pins the one command that previously ignored --json.
func TestInit_JSON(t *testing.T) {
	root := t.TempDir()
	decode := func(out string) (string, []string) {
		var got struct {
			SchemaVersion string   `json:"schema_version"`
			Root          string   `json:"root"`
			Created       []string `json:"created"`
		}
		if err := json.Unmarshal([]byte(out), &got); err != nil {
			t.Fatalf("init --json invalid: %v\n%s", err, out)
		}
		if got.SchemaVersion == "" {
			t.Error("init envelope missing schema_version")
		}
		return got.Root, got.Created
	}
	gotRoot, created := decode(runRoot(t, "init", "--path", root, "--json"))
	if gotRoot != root || len(created) == 0 {
		t.Errorf("first init should report created paths at %q: %v", root, created)
	}
	// Idempotent re-run: created is an empty array, never null.
	out := runRoot(t, "init", "--path", root, "--json")
	if _, created := decode(out); created == nil || len(created) != 0 {
		t.Errorf("re-init should report an empty created array:\n%s", out)
	}
}

func TestRoot_NotAPlanningRepo(t *testing.T) {
	// A temp dir with no tasks/ should error clearly, not panic.
	var out bytes.Buffer
	cmd := NewRootCmd(&out, &out)
	cmd.SetArgs([]string{"-C", t.TempDir(), "task", "list"})
	if err := cmd.Execute(); err == nil {
		t.Fatal("expected an error for a non-planning dir")
	}
}
