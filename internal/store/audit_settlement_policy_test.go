package store

import (
	"bytes"
	"errors"
	"fmt"
	"maps"
	"os"
	"strings"
	"testing"

	"github.com/andy-esch/taskflow/internal/core"
	"github.com/andy-esch/taskflow/internal/domain"
	"github.com/andy-esch/taskflow/internal/testutil"
)

func TestAuditMovesRequireTerminalParsedFindingsBeforeEffects(t *testing.T) {
	for _, status := range []string{"open", "in-progress", "", "opne", "IN-PROGRESS"} {
		for _, to := range []domain.AuditBucket{domain.AuditClosed, domain.AuditDeferred} {
			for _, from := range []domain.AuditBucket{domain.AuditOpen, to} {
				for _, dry := range []bool{true, false} {
					t.Run(fmt.Sprintf("%s/%s-to-%s/dry=%t", status, from, to, dry), func(t *testing.T) {
						root := t.TempDir()
						path, content := testutil.AuditFixture(root, string(from), "probe.md",
							"---\narea: probe\n---\n#### H1. Unsettled · **Status:** "+status+"\n\n#### M1. Done · **Status:** fixed\n")
						testutil.Write(t, path, content)
						fs := testutil.Must(NewFS(root, core.UnrestrictedMutations()))
						before := testutil.SnapshotTree(t, root)
						_, err := fs.MoveAudit("probe", to, dry)
						var refusal *core.AuditMoveError
						var unsettled *domain.AuditUnsettledFindingsError
						if !errors.Is(err, domain.ErrValidation) || !errors.As(err, &refusal) || !errors.As(err, &unsettled) || unsettled.Count != 1 || refusal.Source.ID != testutil.TaskID("probe") {
							t.Fatalf("move lost guarded policy/identity: %v", err)
						}
						if !maps.Equal(before, testutil.SnapshotTree(t, root)) {
							t.Fatal("refused move changed tree bytes, modes, or entries")
						}
						if _, err := fs.MoveAudit("probe", domain.AuditOpen, dry); err != nil {
							t.Fatalf("reopen for repair: %v", err)
						}
					})
				}
			}
		}
	}
}

func TestAuditMoveRetryRevalidatesUnsettledParsedStatuses(t *testing.T) {
	for _, status := range []string{"in-progress", "", "opne"} {
		for _, to := range []domain.AuditBucket{domain.AuditClosed, domain.AuditDeferred} {
			t.Run(status+"/"+string(to), func(t *testing.T) {
				root := t.TempDir()
				path, content := testutil.AuditFixture(root, "open", "probe.md", "---\narea: probe\n---\n#### H1. Done · **Status:** fixed\n")
				testutil.Write(t, path, content)
				concurrent := []byte(strings.Replace(content, "**Status:** fixed", "**Status:** "+status, 1))
				previous := testHookBeforeMoveAuditWrite
				t.Cleanup(func() { testHookBeforeMoveAuditWrite = previous })
				calls := 0
				testHookBeforeMoveAuditWrite = func() {
					calls++
					testHookBeforeMoveAuditWrite = previous
					if err := os.WriteFile(path, concurrent, 0o644); err != nil {
						t.Fatal(err)
					}
				}
				fs := testutil.Must(NewFS(root, core.UnrestrictedMutations()))
				svc := core.MustNewService(fs, core.WithRetry(1, func(int) {}))
				_, err := svc.MoveAudit("probe", to, false)
				var unsettled *domain.AuditUnsettledFindingsError
				if !errors.As(err, &unsettled) || calls != 1 {
					t.Fatalf("retry did not reload current statuses: calls=%d err=%v", calls, err)
				}
				after, err := os.ReadFile(path)
				if err != nil || !bytes.Equal(after, concurrent) {
					t.Fatalf("move overwrote concurrent evidence: %v\n%s", err, after)
				}
			})
		}
	}
}

func TestFindingEditsCannotReactivateNonOpenAudit(t *testing.T) {
	for _, bucket := range []domain.AuditBucket{domain.AuditClosed, domain.AuditDeferred} {
		for _, prior := range []string{"fixed", "open", "in-progress"} {
			for _, status := range []string{"open", "IN-PROGRESS 2026-10-09"} {
				for _, dry := range []bool{true, false} {
					t.Run(fmt.Sprintf("%s/%s-to-%s/dry=%t", bucket, prior, status, dry), func(t *testing.T) {
						root := t.TempDir()
						path, content := testutil.AuditFixture(root, string(bucket), "probe.md", "---\narea: probe\n---\n#### H1. Issue · **Status:** "+prior+"\n\n## Candidate tasks\n\n"+domain.CandidateTasksMarkerComment()+"\n")
						testutil.Write(t, path, content)
						fs := testutil.Must(NewFS(root, core.UnrestrictedMutations()))
						svc := core.MustNewService(fs)
						before := testutil.SnapshotTree(t, root)
						note, candidate := "Must not partially land", "Must not add a row"
						_, changed, err := svc.EditFinding("probe", "H1", core.FindingEdit{Status: status, Note: &note, Candidate: &candidate}, dry)
						if !errors.Is(err, domain.ErrValidation) || !strings.Contains(err.Error(), "reopen the audit first") || changed {
							t.Fatalf("non-open status edit: changed=%t err=%v", changed, err)
						}
						if !maps.Equal(before, testutil.SnapshotTree(t, root)) {
							t.Fatal("refused compound edit changed bytes, modes, or entries")
						}
						// Pin the unchanged-status path separately: no-op does not bypass policy.
						_, _, err = svc.SetFindingStatus("probe", "H1", status, dry)
						if !errors.Is(err, domain.ErrValidation) || !maps.Equal(before, testutil.SnapshotTree(t, root)) {
							t.Fatalf("status-only/no-op bypass: %v", err)
						}
					})
				}
			}
		}
	}
}

func TestFindingEditsAllowTerminalRepairAndMetadataWithoutCertifyingWholeAudit(t *testing.T) {
	for _, bucket := range []domain.AuditBucket{domain.AuditClosed, domain.AuditDeferred} {
		for _, status := range []string{"FIXED 2026-10-09 (PR #1)", "tracked by 6g0000000011", "deferred (later)", "superseded", "wontfix"} {
			t.Run(string(bucket)+"/"+status, func(t *testing.T) {
				root := t.TempDir()
				path, content := testutil.AuditFixture(root, string(bucket), "probe.md", "---\narea: probe\ncustom: retained\n---\n#### H1. Repair · **Status:** opne\n\n#### M1. Still broken · **Status:** in-progress\n\n## Candidate tasks\n\n"+domain.CandidateTasksMarkerComment()+"\n")
				testutil.Write(t, path, content)
				fs := testutil.Must(NewFS(root, core.UnrestrictedMutations()))
				svc := core.MustNewService(fs)
				before := testutil.SnapshotTree(t, root)
				if _, changed, err := svc.SetFindingStatus("probe", "H1", status, true); err != nil || !changed || !maps.Equal(before, testutil.SnapshotTree(t, root)) {
					t.Fatalf("terminal preview: changed=%t err=%v", changed, err)
				}
				if _, changed, err := svc.SetFindingStatus("probe", "H1", status, false); err != nil || !changed {
					t.Fatalf("terminal repair: changed=%t err=%v", changed, err)
				}
				note := "Explaining the remaining repair"
				if _, changed, err := svc.EditFinding("probe", "M1", core.FindingEdit{Note: &note}, false); err != nil || !changed {
					t.Fatalf("metadata-only repair: changed=%t err=%v", changed, err)
				}
				candidate := "Resume after reopening"
				if _, changed, err := svc.EditFinding("probe", "M1", core.FindingEdit{Candidate: &candidate}, false); err != nil || !changed {
					t.Fatalf("candidate-only repair: changed=%t err=%v", changed, err)
				}
				a, body, err := fs.GetAudit("probe")
				if err != nil || a.Bucket != bucket || a.ActiveFindings != 1 || !strings.Contains(body, note) || !strings.Contains(body, candidate) {
					t.Fatalf("repair certified or changed unrelated state: %+v %v\n%s", a, err, body)
				}
				if _, err := svc.MoveAudit("probe", bucket, true); !errors.Is(err, domain.ErrValidation) {
					t.Fatalf("same-bucket move must still report residual defect: %v", err)
				}
			})
		}
	}
}

func TestFindingEditRetryObservesConcurrentNonOpenBucket(t *testing.T) {
	for _, target := range []domain.AuditBucket{domain.AuditClosed, domain.AuditDeferred} {
		t.Run(string(target), func(t *testing.T) {
			fs, path := transformAuditRepo(t)
			settled := strings.Replace(transformAuditSource, "**Status:** open", "**Status:** fixed", 1)
			testutil.Write(t, path, settled)
			previous := testHookBeforeBodyWrite
			t.Cleanup(func() { testHookBeforeBodyWrite = previous })
			calls := 0
			var concurrent []byte
			testHookBeforeBodyWrite = func() {
				calls++
				testHookBeforeBodyWrite = previous
				if _, err := fs.MoveAudit("2026-01-01-a", target, false); err != nil {
					t.Fatal(err)
				}
				var err error
				concurrent, err = os.ReadFile(path)
				if err != nil {
					t.Fatal(err)
				}
			}
			svc := core.MustNewService(fs, core.WithRetry(1, func(int) {}))
			_, changed, err := svc.SetFindingStatus("2026-01-01-a", "H1", "in-progress", false)
			if calls != 1 || changed || !errors.Is(err, domain.ErrValidation) || !strings.Contains(err.Error(), "reopen the audit first") {
				t.Fatalf("edit ignored concurrent bucket: calls=%d changed=%t err=%v", calls, changed, err)
			}
			after, err := os.ReadFile(path)
			if err != nil || !bytes.Equal(after, concurrent) {
				t.Fatalf("retry overwrote concurrent bucket change: %v\n%s", err, after)
			}
		})
	}
}
