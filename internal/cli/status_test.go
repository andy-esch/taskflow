package cli

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/andy-esch/taskflow/internal/testutil"
)

// M3 (2026-06-22 audit): status must exit non-zero when files are unreadable,
// matching the list/lint contract — an agent gating on `status` must not get a
// success code on a broken tree. The dashboard still renders first.
func TestStatus_ExitsNonZeroOnUnreadableFiles(t *testing.T) {
	root := setupRepo(t)
	// tier as a quoted string fails the strict decode → a FileProblem.
	path, out := testutil.TaskFixture(root, "ready-to-start", "broken.md",
		"---\nstatus: ready-to-start\ntier: \"4\"\n---\n# Broken\n")
	mustWrite(t, path, out)
	out, err := runRootRC(t, "-C", root, "status")
	if err == nil {
		t.Fatal("status must exit non-zero when a file is unreadable")
	}
	if !strings.Contains(out, "Tasks") {
		t.Errorf("the dashboard should still render before the non-zero exit:\n%s", out)
	}
}

func TestStatus_Smoke(t *testing.T) {
	root := setupRepo(t) // alpha (ready-to-start), beta (in-progress)
	out := runRoot(t, "-C", root, "status")
	for _, want := range []string{"Tasks", "In progress", "beta", "ready-to-start"} {
		if !strings.Contains(out, want) {
			t.Errorf("status output missing %q:\n%s", want, out)
		}
	}
}

func TestStatus_JSON(t *testing.T) {
	root := setupRepo(t)
	out := runRoot(t, "-C", root, "status", "--json")
	var got struct {
		SchemaVersion string           `json:"schema_version"`
		Counts        []map[string]any `json:"counts"`
		InProgress    []map[string]any `json:"in_progress"`
	}
	if err := json.Unmarshal([]byte(out), &got); err != nil {
		t.Fatalf("bad json: %v\n%s", err, out)
	}
	if got.SchemaVersion == "" || len(got.Counts) == 0 {
		t.Errorf("incomplete summary json: %+v", got)
	}
	if len(got.InProgress) != 1 {
		t.Errorf("expected 1 in-progress task (beta), got %+v", got.InProgress)
	}
}
