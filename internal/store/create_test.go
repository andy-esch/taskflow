package store

import (
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/andy-esch/taskflow/internal/domain"
)

func TestCreateTask_OrderQuotingClobber(t *testing.T) {
	fs := NewFS(t.TempDir())
	task := domain.Task{
		Slug: "demo", Status: domain.StatusReadyToStart, Epic: "e1",
		Description: "has a colon: yes", Effort: "Unknown", Tier: 3,
		Priority: "medium", Autonomy: 3, Tags: []string{"a", "b"}, Created: "2026-06-08",
	}

	got, err := fs.CreateTask(task, "\n# Demo\n")
	if err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(got.Path)
	if err != nil {
		t.Fatal(err)
	}
	s := string(b)

	// Canonical order: status before epic before description.
	si, ei, di := strings.Index(s, "status:"), strings.Index(s, "epic:"), strings.Index(s, "description:")
	if si < 0 || si >= ei || ei >= di {
		t.Errorf("frontmatter not in canonical order:\n%s", s)
	}
	// A colon in the description must be quoted (the pm non-conformant-YAML trap);
	// yaml.v3 uses single quotes, which round-trip fine.
	if !strings.Contains(s, "'has a colon: yes'") {
		t.Errorf("description with colon not quoted:\n%s", s)
	}
	// And it must re-parse — proving valid YAML was written.
	if _, _, err := fs.GetTask("demo"); err != nil {
		t.Errorf("created task does not re-parse: %v", err)
	}
	// Clobber refused with a conflict (not a generic validation error).
	if _, err := fs.CreateTask(task, "x"); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("clobber should be ErrConflict, got %v", err)
	}
}

func TestCreateEpic_AutoNumber(t *testing.T) {
	fs := NewFS(t.TempDir())
	// First epic → 01; with an existing 04-... the next is 05.
	first, err := fs.CreateEpic("alpha", domain.Epic{Status: "planning", Description: "d", Priority: "medium", Created: "2026-06-08"}, "\n# Alpha\n")
	if err != nil {
		t.Fatal(err)
	}
	if first.ID != "01-alpha" {
		t.Errorf("first epic id = %q, want 01-alpha", first.ID)
	}
	if err := os.WriteFile(fs.epicsDir+"/04-beta.md", []byte("---\nstatus: planning\n---\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	next, err := fs.CreateEpic("gamma", domain.Epic{Status: "planning", Description: "d", Priority: "medium", Created: "2026-06-08"}, "\n# G\n")
	if err != nil {
		t.Fatal(err)
	}
	if next.ID != "05-gamma" {
		t.Errorf("next epic id = %q, want 05-gamma", next.ID)
	}
}
