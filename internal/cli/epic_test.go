package cli

import (
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
	write("epics/demo.md", "---\nstatus: in-progress\ndescription: demo epic\n---\n# Demo Epic\n")
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
