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

func TestAuditMovesRefuseIncompleteEvidenceBeforeAnyEffects(t *testing.T) {
	for _, tc := range []struct{ name, body string }{
		{"repairable", "## Findings\n\n#### H-1. Lost · **Status:** fixed\n"},
		{"ambiguous", "## Findings\n\n#### H1 Missing metadata\n"},
		{"mixed", "#### M1. Settled · **Status:** fixed\n\n#### H-1. Lost · **Status:** fixed\n"},
	} {
		for _, to := range []domain.AuditBucket{domain.AuditClosed, domain.AuditDeferred} {
			for _, from := range []domain.AuditBucket{domain.AuditOpen, to} {
				for _, dry := range []bool{true, false} {
					t.Run(fmt.Sprintf("%s/%s-to-%s/dry=%t", tc.name, from, to, dry), func(t *testing.T) {
						root := t.TempDir()
						path, content := testutil.AuditFixture(root, string(from), "probe.md", "---\narea: probe\n---\n"+tc.body)
						testutil.Write(t, path, content)
						fs := testutil.Must(NewFS(root, core.UnrestrictedMutations()))
						before := testutil.SnapshotTree(t, root)
						_, err := core.MustNewService(fs).MoveAudit("probe", to, dry)
						var refusal *core.AuditMoveError
						if !errors.Is(err, domain.ErrValidation) || !errors.As(err, &refusal) || !strings.Contains(err.Error(), "unparsed") || !strings.Contains(err.Error(), "audit lint "+testutil.TaskID("probe")) || refusal.Source.ID != testutil.TaskID("probe") {
							t.Fatalf("incomplete evidence must refuse with a diagnostic route: %v", err)
						}
						if !maps.Equal(before, testutil.SnapshotTree(t, root)) {
							t.Fatal("refused move changed tree bytes, modes, or entries")
						}
					})
				}
			}
		}
	}
}

func TestAuditMovesKeepCleanEmptyAndOrdinaryAuditsSupported(t *testing.T) {
	for _, body := range []string{
		"# Empty\n",
		"#### H1. Settled · **Status:** fixed\n\n#### M1. Parked · **Status:** deferred\n",
		"### S3 Storage Architecture\n\n```md\n#### H-1. Example · **Status:** open\n```\n",
	} {
		for _, to := range []domain.AuditBucket{domain.AuditClosed, domain.AuditDeferred} {
			root := t.TempDir()
			path, content := testutil.AuditFixture(root, "open", "probe.md", "---\narea: probe\n---\n"+body)
			testutil.Write(t, path, content)
			fs := testutil.Must(NewFS(root, core.UnrestrictedMutations()))
			before := testutil.SnapshotTree(t, root)
			preview, err := fs.MoveAudit("probe", to, true)
			if err != nil || preview.Bucket != to || !maps.Equal(before, testutil.SnapshotTree(t, root)) {
				t.Fatalf("clean preview to %s: audit=%+v err=%v", to, preview, err)
			}
			after, err := fs.MoveAudit("probe", to, false)
			if err != nil || after.Bucket != to {
				t.Fatalf("clean move to %s: audit=%+v err=%v", to, after, err)
			}
		}
	}
}

func TestAuditReopenAllowsIncompleteEvidenceForRepair(t *testing.T) {
	root := t.TempDir()
	path, content := testutil.AuditFixture(root, "closed", "probe.md", "---\narea: probe\n---\n## Findings\n\n#### H1 Missing metadata\n")
	testutil.Write(t, path, content)
	fs := testutil.Must(NewFS(root, core.UnrestrictedMutations()))
	for _, dry := range []bool{true, false} {
		a, err := fs.MoveAudit("probe", domain.AuditOpen, dry)
		if err != nil || a.Bucket != domain.AuditOpen || a.UnparsedFindings != 1 {
			t.Fatalf("reopen dry=%t: audit=%+v err=%v", dry, a, err)
		}
	}
}

func TestAuditMoveRetryRevalidatesNewUnparsedEvidence(t *testing.T) {
	for _, to := range []domain.AuditBucket{domain.AuditClosed, domain.AuditDeferred} {
		t.Run(string(to), func(t *testing.T) {
			root := t.TempDir()
			path, content := testutil.AuditFixture(root, "open", "probe.md", "---\narea: probe\n---\n#### M1. Settled · **Status:** fixed\n")
			testutil.Write(t, path, content)
			concurrent := []byte(content + "\n## Findings\n\n#### H1 Missing metadata\n")
			previous := testHookBeforeMoveAuditWrite
			t.Cleanup(func() { testHookBeforeMoveAuditWrite = previous })
			testHookBeforeMoveAuditWrite = func() {
				testHookBeforeMoveAuditWrite = previous
				if err := os.WriteFile(path, concurrent, 0o644); err != nil {
					t.Fatal(err)
				}
			}
			fs := testutil.Must(NewFS(root, core.UnrestrictedMutations()))
			svc := core.MustNewService(fs, core.WithRetry(1, func(int) {}))
			if _, err := svc.MoveAudit("probe", to, false); !errors.Is(err, domain.ErrValidation) || !strings.Contains(err.Error(), "unparsed") {
				t.Fatalf("CAS retry failed to revalidate current evidence: %v", err)
			}
			after, err := os.ReadFile(path)
			if err != nil || !bytes.Equal(after, concurrent) {
				t.Fatalf("move overwrote concurrent evidence: err=%v\n%s", err, after)
			}
		})
	}
}
