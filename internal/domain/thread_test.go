package domain

import (
	"errors"
	"testing"
)

func validThread() Thread {
	return Thread{
		ID: "6g3q4rtmv4ak", Status: ThreadStatusUnstarted,
		Description: "Ship Thread documents", Goal: "Dogfood the remaining Thread work",
		Created: "2026-08-29", Tasks: []string{"6g3q4rtmv4ak", "6g4wm2yf6tyj"},
	}
}

func TestValidateThreadDocument(t *testing.T) {
	if err := ValidateThreadDocument(validThread()); err != nil {
		t.Fatalf("valid Thread: %v", err)
	}

	tests := []struct {
		name string
		edit func(*Thread)
	}{
		{"invalid id", func(v *Thread) { v.ID = "bad" }},
		{"invalid status", func(v *Thread) { v.Status = "paused" }},
		{"missing description", func(v *Thread) { v.Description = "" }},
		{"missing goal", func(v *Thread) { v.Goal = "" }},
		{"invalid target", func(v *Thread) { v.TargetDate = "soon" }},
		{"invalid member", func(v *Thread) { v.Tasks = []string{"bad"} }},
		{"duplicate member", func(v *Thread) { v.Tasks = []string{"6g3q4rtmv4ak", "6g3q4rtmv4ak"} }},
		{"unsorted members", func(v *Thread) { v.Tasks = []string{"6g4wm2yf6tyj", "6g3q4rtmv4ak"} }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value := validThread()
			tt.edit(&value)
			if err := ValidateThreadDocument(value); !errors.Is(err, ErrValidation) {
				t.Fatalf("error = %v, want ErrValidation", err)
			}
		})
	}
}

func TestThreadStatusesAreClosedAndCopied(t *testing.T) {
	values := AllThreadStatuses()
	values[0] = "changed"
	if AllThreadStatuses()[0] != ThreadStatusUnstarted {
		t.Fatal("AllThreadStatuses exposed mutable backing storage")
	}
	for _, status := range AllThreadStatuses() {
		if err := ValidateThreadStatus(status); err != nil {
			t.Errorf("status %q: %v", status, err)
		}
	}
}
