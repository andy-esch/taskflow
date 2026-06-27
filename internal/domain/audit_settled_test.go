package domain

import "testing"

func TestAuditSettled(t *testing.T) {
	for _, c := range []struct {
		name string
		a    Audit
		want bool
	}{
		{"all done", Audit{Findings: 3, DoneFindings: 3}, true},
		{"all dropped (e.g. superseded)", Audit{Findings: 2, DroppedFindings: 2}, true},
		{"mixed done + dropped", Audit{Findings: 4, DoneFindings: 1, DroppedFindings: 3}, true},
		{"has open", Audit{Findings: 2, OpenFindings: 1, DoneFindings: 1}, false},
		{"has in-progress", Audit{Findings: 2, ActiveFindings: 1, DoneFindings: 1}, false},
		{"unknown-status leftover (Done+Dropped < Findings)", Audit{Findings: 2, DoneFindings: 1}, false},
		{"no findings", Audit{Findings: 0}, false},
	} {
		if got := c.a.Settled(); got != c.want {
			t.Errorf("%s: Settled() = %v, want %v", c.name, got, c.want)
		}
	}
}
