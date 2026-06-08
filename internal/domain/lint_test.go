package domain

import "testing"

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
