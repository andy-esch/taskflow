package domain

import (
	"strings"
	"testing"
)

func cleanTask() Task {
	return Task{
		Slug: "good", Status: StatusReadyToStart, Epic: "e1",
		Tier: 2, Priority: "high", Effort: "2h", Created: "2026-01-01",
		Tags: []string{"a"},
	}
}

func TestLintTask_Clean(t *testing.T) {
	if issues := LintTask(cleanTask(), func(string) bool { return true }); len(issues) != 0 {
		t.Errorf("clean task has issues: %+v", issues)
	}
}

func TestLintTask_Issues(t *testing.T) {
	bad := Task{Status: StatusInProgress} // missing nearly everything; needs description
	issues := LintTask(bad, func(string) bool { return false })

	got := map[string]bool{}
	for _, i := range issues {
		got[i.Field] = true
	}
	for _, field := range []string{"epic", "tier", "priority", "effort", "created", "tags", "description"} {
		if !got[field] {
			t.Errorf("expected an issue for %q; got %+v", field, issues)
		}
	}
}

func TestLintTask_UnknownEpic(t *testing.T) {
	task := cleanTask()
	issues := LintTask(task, func(string) bool { return false }) // epic not valid
	found := false
	for _, i := range issues {
		if i.Field == "epic" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected unknown-epic issue, got %+v", issues)
	}
}

func TestLintTask_BadDate(t *testing.T) {
	task := cleanTask()
	task.Created = "yesterday" // present but not YYYY-MM-DD
	issues := LintTask(task, func(string) bool { return true })
	found := false
	for _, i := range issues {
		if i.Field == "created" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected a malformed-date 'created' issue, got %+v", issues)
	}
}

func TestMisfiledIssues(t *testing.T) {
	// A recognized status that disagrees with the folder is flagged.
	if got := MisfiledIssues(Task{Status: StatusCompleted, Declared: StatusReadyToStart}); len(got) == 0 {
		t.Error("expected a misfiled issue for ready-to-start in completed/")
	}
	// A foreign/legacy status word is tolerated (folder governs).
	if got := MisfiledIssues(Task{Status: StatusCompleted, Declared: Status("superseded")}); len(got) != 0 {
		t.Errorf("foreign status should not be flagged: %+v", got)
	}
	// Agreement is clean.
	if got := MisfiledIssues(Task{Status: StatusCompleted, Declared: StatusCompleted}); len(got) != 0 {
		t.Errorf("matching status should not be flagged: %+v", got)
	}
}

func TestLintTask_BadConstraints(t *testing.T) {
	task := cleanTask()
	task.Tier = 9
	task.Priority = "urgent"
	task.Description = "x" // fine length, but make it too long instead:
	long := make([]byte, MaxDescriptionLen+1)
	for i := range long {
		long[i] = 'x'
	}
	task.Description = string(long)

	issues := LintTask(task, func(string) bool { return true })
	got := map[string]bool{}
	for _, i := range issues {
		got[i.Field] = true
	}
	if !got["tier"] || !got["priority"] || !got["description"] {
		t.Errorf("expected tier/priority/description issues, got %+v", issues)
	}
}

// TestLintTask_DescriptionLengthInRunes guards that lint counts characters, not
// bytes, matching ValidateDescription — a multibyte description at the cap must
// pass lint just as it passes creation validation.
func TestLintTask_DescriptionLengthInRunes(t *testing.T) {
	task := cleanTask()
	// MaxDescriptionLen multibyte runes: under the rune cap, but far over it in
	// bytes (each 'é' is 2 bytes), so a byte-based check would wrongly flag it.
	task.Description = strings.Repeat("é", MaxDescriptionLen)
	for _, i := range LintTask(task, func(string) bool { return true }) {
		if i.Field == "description" {
			t.Errorf("a %d-rune description (at the cap) must not be flagged too long: %q",
				MaxDescriptionLen, i.Message)
		}
	}

	// One rune over the cap must still be flagged.
	task.Description = strings.Repeat("é", MaxDescriptionLen+1)
	flagged := false
	for _, i := range LintTask(task, func(string) bool { return true }) {
		if i.Field == "description" {
			flagged = true
		}
	}
	if !flagged {
		t.Errorf("a %d-rune description (over the cap) should be flagged too long", MaxDescriptionLen+1)
	}
}
