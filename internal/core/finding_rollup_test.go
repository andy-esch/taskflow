package core

import (
	"reflect"
	"testing"

	"github.com/andy-esch/taskflow/internal/domain"
)

func TestRollupFindings(t *testing.T) {
	f := func(status, urgency, component string) AuditFinding {
		return AuditFinding{Finding: domain.Finding{Status: status, Urgency: urgency, Component: component}}
	}
	r := rollupFindings([]AuditFinding{
		f("open", "acute", "stravapipe / write paths"),
		f("in-progress", "soon", "stravapipe / pubsub"),
		f("in-progress", "eventually", "dispatcher"),
		f("open", "soon", "web"),
		f("in-progress", "", "stravapipe"), // unspecified urgency
	})
	if r.Open != 2 || r.InProgress != 3 {
		t.Errorf("counts: open=%d in-progress=%d, want 2 / 3", r.Open, r.InProgress)
	}
	// Canonical order: acute, soon, eventually, then extras (unspecified).
	wantUrg := []CountBy{{"acute", 1}, {"soon", 2}, {"eventually", 1}, {"unspecified", 1}}
	if !reflect.DeepEqual(r.ByUrgency, wantUrg) {
		t.Errorf("by urgency = %+v, want %+v", r.ByUrgency, wantUrg)
	}
	// Top-level component, count desc then key asc: stravapipe(3), dispatcher(1), web(1).
	wantComp := []CountBy{{"stravapipe", 3}, {"dispatcher", 1}, {"web", 1}}
	if !reflect.DeepEqual(r.ByComponent, wantComp) {
		t.Errorf("by component = %+v, want %+v", r.ByComponent, wantComp)
	}
	if len(r.Acute) != 1 || r.Acute[0].Urgency != "acute" {
		t.Errorf("acute list = %+v, want exactly the 1 acute finding", r.Acute)
	}
}

func TestTopComponent(t *testing.T) {
	for _, tc := range []struct{ in, want string }{
		{"stravapipe / write paths", "stravapipe"},
		{"dispatcher", "dispatcher"},
		{"  topology / postgres + bigquery ", "topology"},
		{"", ""},
		{"   ", ""},
	} {
		if got := topComponent(tc.in); got != tc.want {
			t.Errorf("topComponent(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}
