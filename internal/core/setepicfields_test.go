package core_test

// Like setfields_coercion_test.go, these run the Service against the REAL store so
// the YAML round-trip (a `--set tags=a,b` becoming a native sequence, the surgical
// frontmatter preserving key order/comments) is actually exercised — the in-package
// fakeStore can't corrupt, so it can't prove it doesn't.

import (
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/andy-esch/taskflow/internal/core"
	"github.com/andy-esch/taskflow/internal/domain"
	"github.com/andy-esch/taskflow/internal/store"
	"github.com/andy-esch/taskflow/internal/testutil"
)

func epicFieldsRepo(t *testing.T) *core.Service {
	t.Helper()
	r := testutil.NewRepo(t)
	r.Epic("01-e.md", "---\nschema: 1\nstatus: active\ndescription: e\npriority: medium\ntags: [seed]\ncreated: \"2026-06-01\"\n---\n# e\n")
	return core.NewService(store.NewFS(r.Root))
}

// TestSetEpicFields_RoundTrip is the headline guard: a `--set tags=a,b` (a string
// from the escape hatch) must write a native YAML list and reload cleanly, and a
// typed priority lands — no FileProblem, no dropout.
func TestSetEpicFields_RoundTrip(t *testing.T) {
	svc := epicFieldsRepo(t)
	if _, err := svc.SetEpicFields("01-e", map[string]any{"priority": "high", "tags": "ui, cli"}, false, false); err != nil {
		t.Fatalf("SetEpicFields rejected a valid update: %v", err)
	}
	epic, _, _, err := svc.ShowEpic("01-e")
	if err != nil {
		t.Fatalf("epic no longer reloads after set (corrupted): %v", err)
	}
	if epic.Priority != "high" {
		t.Errorf("priority = %q, want high", epic.Priority)
	}
	if len(epic.Tags) != 2 || epic.Tags[0] != "ui" || epic.Tags[1] != "cli" {
		t.Errorf("tags = %v, want [ui cli] (trimmed)", epic.Tags)
	}
}

// TestSetEpicFields_Surgical pins the surgical contract: unknown keys (schema),
// comments, and key order survive; the body is untouched. A custom field needs
// --force, then round-trips readably.
func TestSetEpicFields_Surgical(t *testing.T) {
	svc := epicFieldsRepo(t)
	epic, err := svc.SetEpicFields("01-e", map[string]any{"description": "nicer goal"}, false, false)
	if err != nil {
		t.Fatalf("SetEpicFields: %v", err)
	}
	raw, err := os.ReadFile(epic.Path)
	if err != nil {
		t.Fatal(err)
	}
	s := string(raw)
	// The schema key (which the Epic struct doesn't model) and the body survive.
	if !strings.Contains(s, "schema: 1") || !strings.Contains(s, "# e") {
		t.Errorf("surgical write dropped an unknown field or the body:\n%s", s)
	}
	if !strings.Contains(s, "description: nicer goal") {
		t.Errorf("description not updated:\n%s", s)
	}
	// No updated_at is stamped — epics have no such field (consistent with MoveEpic).
	if strings.Contains(s, "updated_at") {
		t.Errorf("epic set must not stamp updated_at:\n%s", s)
	}
}

// TestSetEpicFields_BadPriorityNoWrite: a rejected set leaves the file readable and
// unchanged (validation before any disk touch).
func TestSetEpicFields_BadPriorityNoWrite(t *testing.T) {
	svc := epicFieldsRepo(t)
	if _, err := svc.SetEpicFields("01-e", map[string]any{"priority": "urgent"}, false, false); !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("want ErrValidation for a bad priority, got %v", err)
	}
	epic, _, _, err := svc.ShowEpic("01-e")
	if err != nil {
		t.Fatalf("a rejected set must leave the epic readable, got: %v", err)
	}
	if epic.Priority != "medium" {
		t.Errorf("priority changed despite a rejected set: %q", epic.Priority)
	}
}

// TestSetEpicFields_DryRun previews without writing.
func TestSetEpicFields_DryRun(t *testing.T) {
	svc := epicFieldsRepo(t)
	epic, err := svc.SetEpicFields("01-e", map[string]any{"priority": "high"}, false, true)
	if err != nil {
		t.Fatalf("dry-run SetEpicFields: %v", err)
	}
	if epic.Priority != "high" {
		t.Errorf("dry-run should return the would-be epic (priority high), got %q", epic.Priority)
	}
	reloaded, _, _, _ := svc.ShowEpic("01-e")
	if reloaded.Priority != "medium" {
		t.Errorf("--dry-run must not write: priority is now %q", reloaded.Priority)
	}
}

// TestEditEpic_RealStore exercises the editor loop against the real store: a no-op
// edit reports no change; a valid edit lands and reloads.
func TestEditEpic_RealStore(t *testing.T) {
	svc := epicFieldsRepo(t)

	// No-op: return the content unchanged → changed=false, no write.
	_, changed, err := svc.EditEpic("01-e", func(cur string, _ error) (string, error) { return cur, nil })
	if err != nil {
		t.Fatalf("no-op EditEpic: %v", err)
	}
	if changed {
		t.Errorf("an unchanged edit should report changed=false")
	}

	// A real edit: bump the priority in place.
	epic, changed, err := svc.EditEpic("01-e", func(cur string, _ error) (string, error) {
		return strings.Replace(cur, "priority: medium", "priority: high", 1), nil
	})
	if err != nil {
		t.Fatalf("EditEpic: %v", err)
	}
	if !changed || epic.Priority != "high" {
		t.Errorf("edit should land: changed=%v priority=%q", changed, epic.Priority)
	}
}

// TestEditEpic_RejectsBrokenSave: a save that no longer parses (and isn't fixed on
// re-prompt) is ErrValidation, with nothing written.
func TestEditEpic_RejectsBrokenSave(t *testing.T) {
	svc := epicFieldsRepo(t)
	// The edit callback returns broken frontmatter, then (on reopen) the SAME broken
	// content — the "user gave up" path → ErrValidation.
	broken := "---\nstatus: active\ndescription: e\n  bad: : indent\n"
	_, _, err := svc.EditEpic("01-e", func(_ string, _ error) (string, error) { return broken, nil })
	if !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("a broken save kept broken should be ErrValidation, got %v", err)
	}
	if epic, _, _, e := svc.ShowEpic("01-e"); e != nil || epic.Description != "e" {
		t.Errorf("a rejected edit must leave the epic intact, got epic=%+v err=%v", epic, e)
	}
}
