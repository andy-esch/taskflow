package cli

import (
	"encoding/json"
	"strings"
	"testing"
)

// These close the genuine render-coverage gaps (2026-06-17): the --json variants
// of show/transition and the schema human-mode tables were untested while their
// siblings were. They exercise TaskShowJSON / EpicShowJSON / AuditShowJSON /
// MovesJSON / SchemaHuman / SchemaKindHuman.

func TestTaskShow_JSON(t *testing.T) {
	root := setupRepo(t)
	out := runRoot(t, "-C", root, "task", "show", "alpha", "--json")
	var got struct {
		SchemaVersion string `json:"schema_version"`
		Task          struct {
			Slug, Status string
		} `json:"task"`
		Body string `json:"body"`
	}
	if err := json.Unmarshal([]byte(out), &got); err != nil {
		t.Fatalf("task show --json invalid: %v\n%s", err, out)
	}
	if got.SchemaVersion == "" || got.Task.Slug != "alpha" || got.Task.Status != "ready-to-start" ||
		!strings.Contains(got.Body, "Alpha") {
		t.Errorf("task show --json wrong: %+v", got)
	}
}

func TestEpicShow_JSON(t *testing.T) {
	root := setupEpicRepo(t)
	out := runRoot(t, "-C", root, "epic", "show", "demo", "--json")
	var got struct {
		Epic struct {
			ID, Description string
		} `json:"epic"`
		Tasks []struct{ Slug string } `json:"tasks"`
		Body  string                  `json:"body"`
	}
	if err := json.Unmarshal([]byte(out), &got); err != nil {
		t.Fatalf("epic show --json invalid: %v\n%s", err, out)
	}
	if got.Epic.Description == "" || len(got.Tasks) < 1 || !strings.Contains(got.Body, "Demo") {
		t.Errorf("epic show --json wrong: %+v", got)
	}
}

func TestAuditShow_JSON(t *testing.T) {
	root := setupAuditRepo(t)
	out := runRoot(t, "-C", root, "audit", "show", "o", "--json")
	var got struct {
		Audit struct {
			Slug, Bucket string
		} `json:"audit"`
	}
	if err := json.Unmarshal([]byte(out), &got); err != nil {
		t.Fatalf("audit show --json invalid: %v\n%s", err, out)
	}
	if got.Audit.Slug != "o" || got.Audit.Bucket != "open" {
		t.Errorf("audit show --json wrong: %+v", got)
	}
}

func TestTaskTransition_JSON(t *testing.T) {
	root := setupRepo(t)
	out := runRoot(t, "-C", root, "task", "start", "alpha", "--json")
	var got struct {
		DryRun bool `json:"dry_run"`
		Moves  []struct {
			Slug, To string
		} `json:"moves"`
	}
	if err := json.Unmarshal([]byte(out), &got); err != nil {
		t.Fatalf("task start --json invalid: %v\n%s", err, out)
	}
	if got.DryRun || len(got.Moves) != 1 || got.Moves[0].Slug != "alpha" || got.Moves[0].To != "in-progress" {
		t.Errorf("move --json wrong: %+v", got)
	}
}

func TestSchema_Human(t *testing.T) {
	out := runRoot(t, "-C", t.TempDir(), "schema")
	for _, want := range []string{"Task statuses", "Epic statuses", "Task fields", "Exit codes"} {
		if !strings.Contains(out, want) {
			t.Errorf("schema human missing %q:\n%s", want, out)
		}
	}
}

func TestSchemaKind_Human(t *testing.T) {
	out := runRoot(t, "-C", t.TempDir(), "schema", "task")
	for _, want := range []string{"Sections", "Frontmatter", "Conventions", "Body template"} {
		if !strings.Contains(out, want) {
			t.Errorf("schema task human missing %q:\n%s", want, out)
		}
	}
}
