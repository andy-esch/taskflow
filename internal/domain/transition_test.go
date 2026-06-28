package domain

import "testing"

// TestTaskTransitions pins the task lifecycle registry: the exact verbs, their
// destination statuses, and which is destructive. This is the single source both
// the CLI command tree and the TUI action menu build from, so a drift here is a
// drift across every surface — the test makes the mapping explicit.
func TestTaskTransitions(t *testing.T) {
	want := []Transition{
		{Verb: "start", To: string(StatusInProgress), Destructive: false},
		{Verb: "next", To: string(StatusNextUp), Destructive: false},
		{Verb: "ready", To: string(StatusReadyToStart), Destructive: false},
		{Verb: "complete", To: string(StatusCompleted), Destructive: false},
		{Verb: "defer", To: string(StatusDeferred), Destructive: false},
		{Verb: "deprecate", To: string(StatusDeprecated), Destructive: true},
	}
	got := TaskTransitions()
	if len(got) != len(want) {
		t.Fatalf("TaskTransitions() has %d entries, want %d", len(got), len(want))
	}
	for i, w := range want {
		if got[i] != w {
			t.Errorf("TaskTransitions()[%d] = %+v, want %+v", i, got[i], w)
		}
	}
	// Every destination must be a real status so a consumer's Status(tr.To) cast is
	// never a runtime surprise.
	for _, tr := range got {
		if !Status(tr.To).Valid() {
			t.Errorf("task transition %q -> %q is not a valid status", tr.Verb, tr.To)
		}
	}
}

// TestAuditTransitions pins the audit lifecycle registry: verbs, destination
// buckets, and that none is flagged destructive (audit moves are guarded by the
// store on still-open findings, not by an interactive confirm).
func TestAuditTransitions(t *testing.T) {
	want := []Transition{
		{Verb: "close", To: string(AuditClosed), Destructive: false},
		{Verb: "reopen", To: string(AuditOpen), Destructive: false},
		{Verb: "defer", To: string(AuditDeferred), Destructive: false},
	}
	got := AuditTransitions()
	if len(got) != len(want) {
		t.Fatalf("AuditTransitions() has %d entries, want %d", len(got), len(want))
	}
	for i, w := range want {
		if got[i] != w {
			t.Errorf("AuditTransitions()[%d] = %+v, want %+v", i, got[i], w)
		}
	}
	for _, tr := range got {
		if !AuditBucket(tr.To).Valid() {
			t.Errorf("audit transition %q -> %q is not a valid bucket", tr.Verb, tr.To)
		}
	}
}
