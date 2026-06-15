package domain

import (
	"path"
	"testing"
)

// TestTaskStatusDirs pins that every status maps to "tasks/<status>", in order —
// the single source the store, init, and completion now share.
func TestTaskStatusDirs(t *testing.T) {
	got := TaskStatusDirs()
	if len(got) != len(AllStatuses()) {
		t.Fatalf("want one dir per status (%d), got %d", len(AllStatuses()), len(got))
	}
	for i, st := range AllStatuses() {
		if want := path.Join(TasksDir, st.Dir()); got[i] != want {
			t.Errorf("dir[%d] = %q, want %q", i, got[i], want)
		}
	}
}

// TestAuditBucketDirs is the audit-bucket sibling.
func TestAuditBucketDirs(t *testing.T) {
	got := AuditBucketDirs()
	if len(got) != len(AllAuditBuckets()) {
		t.Fatalf("want one dir per bucket (%d), got %d", len(AllAuditBuckets()), len(got))
	}
	for i, b := range AllAuditBuckets() {
		if want := path.Join(AuditsDir, b.Dir()); got[i] != want {
			t.Errorf("dir[%d] = %q, want %q", i, got[i], want)
		}
	}
}
