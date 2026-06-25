package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func setupEpicRepo(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	write := func(rel, content string) {
		p := filepath.Join(root, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write("epics/demo.md", "---\nstatus: active\ndescription: demo epic\n---\n# Demo Epic\n")
	write("tasks/ready-to-start/a.md", "---\nstatus: ready-to-start\nepic: demo\n---\n# A\n")
	write("tasks/completed/b.md", "---\nstatus: completed\nepic: demo\n---\n# B\n")
	return root
}

func TestEpicList_JSONRollup(t *testing.T) {
	root := setupEpicRepo(t)
	out := runRoot(t, "-C", root, "epic", "list", "--json")

	var got struct {
		Epics []struct {
			ID      string `json:"id"`
			Total   int    `json:"total"`
			Done    int    `json:"done"`
			Percent int    `json:"percent"`
		} `json:"epics"`
	}
	if err := json.Unmarshal([]byte(out), &got); err != nil {
		t.Fatalf("invalid json: %v\n%s", err, out)
	}
	if len(got.Epics) != 1 {
		t.Fatalf("want 1 epic, got %d", len(got.Epics))
	}
	e := got.Epics[0]
	if e.ID != "demo" || e.Total != 2 || e.Done != 1 || e.Percent != 50 {
		t.Errorf("rollup wrong: %+v", e)
	}
}

// TestEpicList_StatusFilter pins the triage filter: --status narrows the set
// (so an agent need not pay for every epic), and an out-of-vocabulary status is
// a loud error rather than a silently-empty list.
func TestEpicList_StatusFilter(t *testing.T) {
	root := setupEpicRepo(t) // one active epic "demo"
	if err := os.WriteFile(filepath.Join(root, "epics", "deprecated-one.md"),
		[]byte("---\nstatus: deprecated\ndescription: old epic\n---\n# Deprecated\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if all := runRoot(t, "-C", root, "epic", "list", "-q"); !strings.Contains(all, "demo") || !strings.Contains(all, "deprecated-one") {
		t.Fatalf("-q should list both epics:\n%s", all)
	}
	active := runRoot(t, "-C", root, "epic", "list", "-q", "--status", "active")
	if !strings.Contains(active, "demo") || strings.Contains(active, "deprecated-one") {
		t.Errorf("--status active should keep only demo:\n%s", active)
	}
	var buf bytes.Buffer
	cmd := NewRootCmd(strings.NewReader(""), &buf, &buf)
	cmd.SetArgs([]string{"-C", root, "epic", "list", "--status", "bogus"})
	if err := cmd.Execute(); err == nil || !strings.Contains(err.Error(), "bogus") {
		t.Errorf("invalid --status should error naming the value, got %v", err)
	}
}

// TestEpicList_PercentColumn pins the projectable rollup % the feedback asked for
// (`-c id,status,percent,description`).
func TestEpicList_PercentColumn(t *testing.T) {
	root := setupEpicRepo(t) // demo: 1 of 2 tasks done = 50%
	out := runRoot(t, "-C", root, "epic", "list", "-o", "table", "-c", "id,percent")
	if h := strings.SplitN(out, "\n", 2)[0]; h != "id\tpercent" {
		t.Errorf("projected header wrong: %q", h)
	}
	if !strings.Contains(out, "demo\t50") {
		t.Errorf("percent column should project the rollup %%, got:\n%s", out)
	}
}

// TestEpicList_JSONProjection covers the projected --json path for a NON-task
// entity (the other CLI tests only exercise task list), confirming compactness,
// the schema_version envelope, -c narrowing, and string-valued numeric columns.
func TestEpicList_JSONProjection(t *testing.T) {
	root := setupEpicRepo(t) // demo: 1 of 2 tasks done = 50%
	out := runRoot(t, "-C", root, "epic", "list", "--json", "-c", "id,percent")
	if strings.Count(out, "\n") != 1 {
		t.Errorf("projected epic --json should be compact (one trailing newline):\n%q", out)
	}
	var got struct {
		SchemaVersion string           `json:"schema_version"`
		Epics         []map[string]any `json:"epics"`
	}
	if err := json.Unmarshal([]byte(out), &got); err != nil {
		t.Fatalf("invalid json: %v\n%s", err, out)
	}
	if got.SchemaVersion == "" {
		t.Errorf("projected envelope must carry schema_version:\n%s", out)
	}
	if len(got.Epics) != 1 {
		t.Fatalf("want 1 epic row, got %d", len(got.Epics))
	}
	row := got.Epics[0]
	if len(row) != 2 || row["id"] != "demo" || row["percent"] != "50" {
		t.Errorf(`row should be exactly {id:"demo", percent:"50"} (percent string-valued): %v`, row)
	}
}

func TestEpicShow(t *testing.T) {
	root := setupEpicRepo(t)
	out := runRoot(t, "-C", root, "epic", "show", "demo")
	if !strings.Contains(out, "demo epic") {
		t.Errorf("missing description:\n%s", out)
	}
	if !strings.Contains(out, "a") || !strings.Contains(out, "b") {
		t.Errorf("should list both tasks:\n%s", out)
	}
	if !strings.Contains(out, "# Demo Epic") {
		t.Errorf("missing body:\n%s", out)
	}
}
