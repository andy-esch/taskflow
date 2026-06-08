package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
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

func TestRoot_NotAPlanningRepo(t *testing.T) {
	// A temp dir with no tasks/ should error clearly, not panic.
	var out bytes.Buffer
	cmd := NewRootCmd(&out, &out)
	cmd.SetArgs([]string{"-C", t.TempDir(), "task", "list"})
	if err := cmd.Execute(); err == nil {
		t.Fatal("expected an error for a non-planning dir")
	}
}
