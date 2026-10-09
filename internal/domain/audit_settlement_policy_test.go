package domain

import (
	"errors"
	"slices"
	"strings"
	"testing"
)

// Portable counts must agree with the parsed status policy without body I/O.
func TestParsedSettlementPolicyAgreesAcrossReadinessMovesAndLint(t *testing.T) {
	for _, status := range append(FindingStatuses(), "", "opne", "IN-PROGRESS", "FIXED", " TRACKED ") {
		t.Run(status, func(t *testing.T) {
			findings := []Finding{{Code: "H1", Status: status, StatusDecoration: "by task"}}
			tally := TallyFindings(findings)
			a := Audit{Slug: "portable", Bucket: AuditOpen, Findings: 1,
				OpenFindings: tally.Open, ActiveFindings: tally.Active,
				DoneFindings: tally.Done, DroppedFindings: tally.Dropped}
			terminal := slices.Contains([]string{"fixed", "tracked", "deferred", "superseded", "wontfix"}, strings.ToLower(strings.TrimSpace(status)))
			if TerminalFindingStatus(status) != terminal {
				t.Fatalf("terminal vocabulary drift for %q", status)
			}
			if a.Settled() != terminal || a.ReadyToClose() != terminal || (a.UnsettledFindings() == 0) != terminal {
				t.Fatalf("read policy disagrees for %q: %+v", status, a)
			}
			for _, bucket := range []AuditBucket{AuditClosed, AuditDeferred} {
				err := a.ValidateMove(bucket)
				if terminal {
					if err != nil {
						t.Fatalf("terminal move: %v", err)
					}
				} else {
					var unsettled *AuditUnsettledFindingsError
					if !errors.Is(err, ErrValidation) || !errors.As(err, &unsettled) || unsettled.Count != 1 || unsettled.Target != bucket {
						t.Fatalf("nonterminal move lost typed refusal: %v", err)
					}
				}
				bucketIssue := false
				for _, issue := range LintFindings(string(bucket), findings) {
					if issue.Field == "bucket" {
						bucketIssue = true
						if !strings.Contains(issue.Message, "1 unsettled") {
							t.Fatalf("wrong bucket diagnostic: %+v", issue)
						}
					}
				}
				if bucketIssue == terminal {
					t.Fatalf("lint disagrees for status %q, bucket %s: %v", status, bucket, LintFindings(string(bucket), findings))
				}
			}
			if err := a.ValidateMove(AuditOpen); err != nil {
				t.Fatalf("reopen must remain available: %v", err)
			}
		})
	}
}

func TestParsedSettlementDoesNotSubsumeMetadataLint(t *testing.T) {
	findings := []Finding{{Code: "H1", Status: "tracked"}}
	if !TerminalFindingStatus(findings[0].Status) || len(LintFindings("closed", findings)) != 1 {
		t.Fatal("tracked is a terminal token, but a missing destination must still lint")
	}
}
