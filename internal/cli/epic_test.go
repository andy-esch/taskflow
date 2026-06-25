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

// TestEpicMove pins the `epic move <id> <status>` verb: it surgically rewrites
// the epic's status FIELD (no file moves), through the shared moves report.
func TestEpicMove(t *testing.T) {
	root := setupEpicRepo(t) // epic "demo", status active
	out := runRoot(t, "-C", root, "epic", "move", "demo", "retired")
	if !strings.Contains(out, "demo -> retired") {
		t.Errorf("unexpected output: %q", out)
	}
	b, err := os.ReadFile(filepath.Join(root, "epics", "demo.md"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(b), "status: retired") || strings.Contains(string(b), "status: active") {
		t.Errorf("status field not rewritten:\n%s", b)
	}
	// The file stays at epics/demo.md — epic status is a field, not a directory.
	if _, err := os.Stat(filepath.Join(root, "epics", "demo.md")); err != nil {
		t.Errorf("epic file should stay put after a move: %v", err)
	}
}

// TestEpicMove_DryRun previews the move without writing.
func TestEpicMove_DryRun(t *testing.T) {
	root := setupEpicRepo(t)
	original, err := os.ReadFile(filepath.Join(root, "epics", "demo.md"))
	if err != nil {
		t.Fatal(err)
	}
	out := runRoot(t, "-C", root, "epic", "move", "demo", "retired", "--dry-run")
	if !strings.Contains(out, "would move") {
		t.Errorf("dry-run should say 'would move', got: %q", out)
	}
	b, _ := os.ReadFile(filepath.Join(root, "epics", "demo.md"))
	if string(b) != string(original) {
		t.Errorf("--dry-run must not write:\n%s", b)
	}
}

// TestEpicMove_BadStatus_ExitCode pins exit 11 (validation) for an
// out-of-vocabulary target status.
func TestEpicMove_BadStatus_ExitCode(t *testing.T) {
	root := setupEpicRepo(t)
	_, err := runRootRC(t, "-C", root, "epic", "move", "demo", "bogus")
	if err == nil {
		t.Fatal("a bad target status should error")
	}
	if ExitCode(err) != 11 {
		t.Errorf("want exit 11 (validation), got %d", ExitCode(err))
	}
}

// TestComplete_EpicMove_StatusArg pins position-aware completion: epic ids for the
// leading args, the closed epic-status set on the final (status) arg — never ids.
func TestComplete_EpicMove_StatusArg(t *testing.T) {
	root := setupEpicRepo(t) // epic "demo"
	// Leading arg: epic ids offered, no statuses.
	first := complete(t, "-C", root, "epic", "move", "")
	if !has(first, "demo") {
		t.Errorf("first arg should offer epic ids, got %v", first)
	}
	if has(first, "retired") {
		t.Errorf("first arg must not offer statuses, got %v", first)
	}
	// Final (status) arg: the closed epic-status set is offered.
	last := complete(t, "-C", root, "epic", "move", "demo", "")
	for _, st := range []string{"active", "retired", "deprecated"} {
		if !has(last, st) {
			t.Errorf("status arg should offer %q, got %v", st, last)
		}
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
