package domain

import "testing"

func TestCanonicalIDPrefersAdapterResolutionIdentity(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		semantic string
		got      func() string
	}{
		{"task", "file-task", "semantic-task", func() string { return (Task{FilenameID: "file-task", ID: "semantic-task"}).CanonicalID() }},
		{"thread", "file-thread", "semantic-thread", func() string { return (Thread{FilenameID: "file-thread", ID: "semantic-thread"}).CanonicalID() }},
	}
	for _, tt := range tests {
		t.Run(tt.name+" filename", func(t *testing.T) {
			if got := tt.got(); got != tt.filename {
				t.Fatalf("CanonicalID() = %q, want adapter resolution identity %q", got, tt.filename)
			}
		})
	}

	if got := (Task{ID: "portable-task"}).CanonicalID(); got != "portable-task" {
		t.Errorf("portable Task CanonicalID() = %q", got)
	}
	if got := (Thread{ID: "portable-thread"}).CanonicalID(); got != "portable-thread" {
		t.Errorf("portable Thread CanonicalID() = %q", got)
	}
}
