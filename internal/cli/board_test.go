package cli

import (
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"

	"github.com/andy-esch/taskflow/internal/testutil"
)

func TestBoard_Smoke(t *testing.T) {
	root := setupRepo(t) // alpha (ready-to-start), beta (in-progress); no next-up task
	out := runRoot(t, "-C", root, "board")
	for _, want := range []string{"next-up", "ready-to-start", "in-progress", "alpha", "beta"} {
		if !strings.Contains(out, want) {
			t.Errorf("board output missing %q:\n%s", want, out)
		}
	}
	// An empty active status renders a "(none)" column, not a gap (next-up is empty here).
	if !strings.Contains(out, "(none)") {
		t.Errorf("empty next-up column should render (none):\n%s", out)
	}
	// The board is the active pipeline only — terminal/parked statuses are not sections.
	if strings.Contains(out, "completed") || strings.Contains(out, "deferred") {
		t.Errorf("board should exclude completed/deferred sections:\n%s", out)
	}
}

func TestBoard_ExitsNonZeroOnUnreadableFiles(t *testing.T) {
	root := setupRepo(t)
	// A quoted-string tier fails the strict decode → a FileProblem.
	mustWrite(t, filepath.Join(root, "tasks", testutil.TaskID("broken")+"-broken.md"),
		"---\nstatus: ready-to-start\ntier: \"4\"\n---\n# Broken\n")
	out, err := runRootRC(t, "-C", root, "board")
	if err == nil {
		t.Fatal("board must exit non-zero when a file is unreadable (the status/list agent-gating contract)")
	}
	// The board still renders before the non-zero exit.
	if !strings.Contains(out, "ready-to-start") {
		t.Errorf("the board should still render before the non-zero exit:\n%s", out)
	}
}

func TestBoard_JSON(t *testing.T) {
	root := setupRepo(t)
	out := runRoot(t, "-C", root, "board", "--json")
	var got struct {
		SchemaVersion string `json:"schema_version"`
		Columns       []struct {
			Status string           `json:"status"`
			Tasks  []map[string]any `json:"tasks"`
		} `json:"columns"`
	}
	if err := json.Unmarshal([]byte(out), &got); err != nil {
		t.Fatalf("board --json is not valid JSON: %v\n%s", err, out)
	}
	if got.SchemaVersion == "" {
		t.Error("board --json missing schema_version")
	}
	want := []string{"next-up", "ready-to-start", "in-progress"}
	if len(got.Columns) != len(want) {
		t.Fatalf("got %d columns, want the 3-status active pipeline: %+v", len(got.Columns), got.Columns)
	}
	for i, c := range got.Columns {
		if c.Status != want[i] {
			t.Errorf("column %d = %q, want %q", i, c.Status, want[i])
		}
	}
}
