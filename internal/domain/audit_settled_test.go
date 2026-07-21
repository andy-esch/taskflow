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

func TestAuditReadyToClose(t *testing.T) {
	for _, c := range []struct {
		name string
		a    Audit
		want bool
	}{
		{"open + all done", Audit{Bucket: AuditOpen, Findings: 3, DoneFindings: 3}, true},
		{"open + all dropped (triaged)", Audit{Bucket: AuditOpen, Findings: 2, DroppedFindings: 2}, true},
		{"open + still has open", Audit{Bucket: AuditOpen, Findings: 2, OpenFindings: 1, DoneFindings: 1}, false},
		{"open + no findings", Audit{Bucket: AuditOpen, Findings: 0}, false},
		{"closed but settled (already off the board)", Audit{Bucket: AuditClosed, Findings: 1, DoneFindings: 1}, false},
		{"deferred but settled", Audit{Bucket: AuditDeferred, Findings: 2, DroppedFindings: 2}, false},
	} {
		if got := c.a.ReadyToClose(); got != c.want {
			t.Errorf("%s: ReadyToClose() = %v, want %v", c.name, got, c.want)
		}
	}
}
