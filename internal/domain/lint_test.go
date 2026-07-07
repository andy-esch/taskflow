package domain

import (
	"strings"
	"testing"
)

func cleanTask() Task {
	return Task{
		ID: "6fjangd7kvh0", Slug: "good", Status: StatusReadyToStart, Epic: "e1",
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

func TestFrontmatterStatusIssues(t *testing.T) {
	// A fell-back frontmatter status (missing or unrecognized) is flagged.
	if got := FrontmatterStatusIssues(Task{StatusFellBack: true}); len(got) == 0 {
		t.Error("a fell-back frontmatter status must be flagged")
	}
	// A valid frontmatter status (no fallback) is clean.
	if got := FrontmatterStatusIssues(Task{Status: StatusReadyToStart}); len(got) != 0 {
		t.Errorf("a valid frontmatter status must not be flagged: %+v", got)
	}
}

func TestFrontmatterBucketIssues(t *testing.T) {
	if got := FrontmatterBucketIssues(Audit{BucketFellBack: true}); len(got) == 0 {
		t.Error("a fell-back frontmatter bucket must be flagged")
	}
	if got := FrontmatterBucketIssues(Audit{Bucket: AuditOpen}); len(got) != 0 {
		t.Errorf("a valid frontmatter bucket must not be flagged: %+v", got)
	}
}

func TestEpicNameIssue(t *testing.T) {
	// NN-<slug> names (any digit width) are clean.
	for _, ok := range []string{"00-taskflow-v1-core", "01-backend", "27-x", "100-scale"} {
		if got := EpicNameIssue(ok); len(got) != 0 {
			t.Errorf("%q is a valid NN- name, must not be flagged: %+v", ok, got)
		}
	}
	// Non-NN names are flagged (fail-open: flagged, not dropped) — including an empty slug.
	for _, bad := range []string{"taskflow-v1-core", "demo", "e1", "1-unpadded", "01-", "01"} {
		if got := EpicNameIssue(bad); len(got) == 0 {
			t.Errorf("%q is not NN-<slug> and must be flagged", bad)
		}
	}
}

func TestIDDriftIssue(t *testing.T) {
	// The frontmatter id and the filename id disagree — flagged.
	if got := IDDriftIssue("6fjangd7kvzz", "6fjangd7kvaa"); len(got) == 0 {
		t.Error("a frontmatter id that disagrees with the filename id must be flagged")
	}
	// Matching ids (the post-flatten norm) — clean.
	if got := IDDriftIssue("6fjangd7kvaa", "6fjangd7kvaa"); len(got) != 0 {
		t.Errorf("matching ids must not be flagged: %+v", got)
	}
	// An empty or whitespace-only side is left to MissingIDIssue, not double-reported here.
	for _, blank := range []string{"", "   "} {
		if got := IDDriftIssue(blank, "6fjangd7kvaa"); len(got) != 0 {
			t.Errorf("a blank frontmatter id (%q) is MissingIDIssue's job, not drift: %+v", blank, got)
		}
	}
}

func cleanEpic() Epic {
	return Epic{ID: "20-good", Status: "active", Priority: "high", Description: "a goal"}
}

func TestLintEpic_Clean(t *testing.T) {
	if issues := LintEpic(cleanEpic()); len(issues) != 0 {
		t.Errorf("clean epic has issues: %+v", issues)
	}
}

func TestLintEpic_BadStatus(t *testing.T) {
	e := cleanEpic()
	e.Status = "bananas" // outside the closed vocabulary
	found := false
	for _, i := range LintEpic(e) {
		if i.Field == "status" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected an invalid-status issue, got %+v", LintEpic(e))
	}
}

func TestLintEpic_BadPriority(t *testing.T) {
	e := cleanEpic()
	e.Priority = "urgent" // not high|medium|low
	found := false
	for _, i := range LintEpic(e) {
		if i.Field == "priority" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected a priority issue, got %+v", LintEpic(e))
	}
}

func TestLintEpic_MissingDescription(t *testing.T) {
	e := cleanEpic()
	e.Description = ""
	found := false
	for _, i := range LintEpic(e) {
		if i.Field == "description" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected a missing-description issue, got %+v", LintEpic(e))
	}
}

// TestLintEpic_ArchivedMinimal pins the active-vs-archived split: an archived
// epic is withdrawn, so missing priority/description must NOT be nagged — only
// the status vocabulary is enforced (mirroring archived tasks).
func TestLintEpic_ArchivedMinimal(t *testing.T) {
	e := Epic{ID: "20-old", Status: "deprecated"} // no priority, no description
	if issues := LintEpic(e); len(issues) != 0 {
		t.Errorf("archived epic must not be nagged about fields: %+v", issues)
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

// TestDuplicateEpicNNIssues: epics sharing a leading NN are each flagged (naming the
// colliding peer); a unique NN is left alone.
func TestDuplicateEpicNNIssues(t *testing.T) {
	got := DuplicateEpicNNIssues([]string{"01-billing", "01-invoicing", "02-solo"})
	if len(got) != 2 {
		t.Fatalf("both epics sharing NN 01 should be flagged, got %d: %+v", len(got), got)
	}
	if _, ok := got["02-solo"]; ok {
		t.Errorf("a unique NN key must not be flagged: %q", got["02-solo"].Message)
	}
	if !strings.Contains(got["01-billing"].Message, "01-invoicing") {
		t.Errorf("the issue should name the colliding peer, got %q", got["01-billing"].Message)
	}
	if !strings.Contains(got["01-invoicing"].Message, "01-billing") {
		t.Errorf("the issue should name the colliding peer, got %q", got["01-invoicing"].Message)
	}
}
