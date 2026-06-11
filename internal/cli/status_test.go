package cli

import (
	"encoding/json"
	"strings"
	"testing"
)

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
