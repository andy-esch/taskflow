package domain

import (
	"errors"
	"testing"
)

func TestParseAuditBucket_Valid(t *testing.T) {
	for _, b := range AllAuditBuckets() {
		got, err := ParseAuditBucket(string(b))
		if err != nil || got != b {
			t.Errorf("ParseAuditBucket(%q) = %v, %v", b, got, err)
		}
	}
}

// TestParseAuditBucket_InvalidWrapsValidation pins the error contract: a bad
// bucket must map to exit 11 (ErrValidation), like ParseStatus — otherwise it
// escapes as a generic exit 1 and breaks code-based routing.
func TestParseAuditBucket_InvalidWrapsValidation(t *testing.T) {
	_, err := ParseAuditBucket("bogus")
	if err == nil {
		t.Fatal("expected an error for an invalid audit bucket")
	}
	if !errors.Is(err, ErrValidation) {
		t.Errorf("invalid audit bucket must wrap ErrValidation, got %v", err)
	}
}

// TestAudit_ResolvedAndPercent pins the finding rollup the bars draw from,
// including the zero-findings guard (no divide-by-zero → 0%) and that the
// percentage truncates like the epic rollup.
func TestAudit_ResolvedAndPercent(t *testing.T) {
	cases := []struct {
		findings, open            int
		wantResolved, wantPercent int
	}{
		{0, 0, 0, 0},   // no findings → 0%, not a panic
		{4, 1, 3, 75},  // 3 of 4 resolved
		{4, 0, 4, 100}, // all resolved
		{4, 4, 0, 0},   // none resolved
		{3, 1, 2, 66},  // integer truncation (66.6 → 66)
	}
	for _, c := range cases {
		a := Audit{Findings: c.findings, OpenFindings: c.open}
		if got := a.Resolved(); got != c.wantResolved {
			t.Errorf("Resolved(findings=%d open=%d) = %d, want %d", c.findings, c.open, got, c.wantResolved)
		}
		if got := a.Percent(); got != c.wantPercent {
			t.Errorf("Percent(findings=%d open=%d) = %d, want %d", c.findings, c.open, got, c.wantPercent)
		}
	}
}
