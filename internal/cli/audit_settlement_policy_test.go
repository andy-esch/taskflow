package cli

import (
	"encoding/json"
	"fmt"
	"maps"
	"path/filepath"
	"strings"
	"testing"

	"github.com/andy-esch/taskflow/internal/domain"
	"github.com/andy-esch/taskflow/internal/testutil"
	"github.com/andy-esch/taskflow/internal/wire"
)

func TestAuditParsedSettlementCLIRefusalsAndLint(t *testing.T) {
	for _, status := range []string{"in-progress", "", "opne"} {
		for _, bucket := range []string{"closed", "deferred"} {
			t.Run(status+"/"+bucket, func(t *testing.T) {
				root := setupRepo(t)
				path, content := testutil.AuditFixture(root, bucket, "probe.md", "---\narea: probe\n---\n#### H1. Unsettled · **Status:** "+status+"\n")
				testutil.Write(t, path, content)
				before := testutil.SnapshotTree(t, root)
				verb := "close"
				if bucket == "deferred" {
					verb = "defer"
				}
				for _, dry := range []bool{true, false} {
					args := []string{"-C", root, "audit", verb, "probe", "--json"}
					if dry {
						args = append(args, "--dry-run")
					}
					result, err := runRootStreams(t, args...)
					if ExitCode(err) != 11 {
						t.Fatalf("same-bucket call must refuse: %v", err)
					}
					var receipt wire.MovesEnvelope
					if err := json.Unmarshal([]byte(result.Out), &receipt); err != nil || receipt.SchemaVersion != wire.SchemaVersion || receipt.DryRun != dry || len(receipt.Moves) != 1 || !strings.Contains(receipt.Moves[0].Error, "1 unsettled parsed") {
						t.Fatalf("refusal receipt: %v %s", err, result.Out)
					}
				}
				for _, args := range [][]string{{"audit", "lint", "probe", "--json"}, {"lint", "--json"}} {
					result, err := runRootStreams(t, append([]string{"-C", root}, args...)...)
					if ExitCode(err) != 11 || !strings.Contains(result.Out, "unsettled parsed finding") {
						t.Fatalf("lint omitted non-open settlement defect: out=%s err=%v", result.Out, err)
					}
				}
				if !maps.Equal(before, testutil.SnapshotTree(t, root)) {
					t.Fatal("refused command/lint changed planning tree")
				}
			})
		}
	}
}

func TestAuditUnsettledAdviceExecutesAgainstSourceNotDeclarationOrSlug(t *testing.T) {
	const sourceID, siblingID = "6g0000000011", "6g0000000012"
	for _, status := range []string{"in-progress", "opne", ""} {
		t.Run(status, func(t *testing.T) {
			root := setupRepo(t)
			for _, record := range []struct{ id, slug, declared, body string }{
				{sourceID, siblingID, siblingID, "#### H1. Unsettled · **Status:** " + status + "\n"},
				{siblingID, "empty", siblingID, "# Empty audit\n"},
			} {
				path := filepath.Join(root, domain.AuditsDir, record.id+"-"+record.slug+".md")
				testutil.Write(t, path, fmt.Sprintf("---\nschema: 1\nid: %s\nbucket: open\narea: probe\n---\n%s", record.declared, record.body))
			}
			before := testutil.SnapshotTree(t, root)
			for _, verb := range []string{"close", "defer"} {
				result, err := runRootStreams(t, "-C", root, "audit", verb, sourceID, "--json")
				if ExitCode(err) != 11 || !strings.Contains(result.Out, "audit findings "+sourceID) || !strings.Contains(result.Out, "audit lint "+sourceID) || strings.Contains(result.Out, "audit findings "+siblingID) {
					t.Fatalf("advice chose wrong source: out=%s err=%v", result.Out, err)
				}
				rows := runRoot(t, "-C", root, "audit", "findings", sourceID, "--json")
				if !strings.Contains(rows, `"code":"H1"`) || !strings.Contains(rows, `"status":"`+status+`"`) {
					t.Fatalf("recommended findings route missed original status: %s", rows)
				}
				lint, err := runRootStreams(t, "-C", root, "audit", "lint", sourceID, "--json")
				if ExitCode(err) != 11 || !strings.Contains(lint.Out, "disagrees with the filename id") {
					t.Fatalf("recommended lint route missed original declaration drift: %s %v", lint.Out, err)
				}
			}
			if !maps.Equal(before, testutil.SnapshotTree(t, root)) {
				t.Fatal("advice/refusal changed planning tree")
			}
		})
	}
}

func TestAuditFindingCLIRequiresExplicitReopenToReactivate(t *testing.T) {
	for _, bucket := range []string{"closed", "deferred"} {
		for _, dry := range []bool{true, false} {
			t.Run(fmt.Sprintf("%s/dry=%t", bucket, dry), func(t *testing.T) {
				root := setupRepo(t)
				path, content := testutil.AuditFixture(root, bucket, "probe.md", "---\narea: probe\n---\n#### H1. Done · **Status:** fixed\n")
				testutil.Write(t, path, content)
				before := testutil.SnapshotTree(t, root)
				for _, status := range []string{"open", "in-progress"} {
					args := []string{"-C", root, "audit", "finding", "probe", "H1", "--status", status, "--note", "Must not land", "--json"}
					if dry {
						args = append(args, "--dry-run")
					}
					_, err := runRootStreams(t, args...)
					if ExitCode(err) != 11 || !strings.Contains(err.Error(), "reopen the audit first") || !maps.Equal(before, testutil.SnapshotTree(t, root)) {
						t.Fatalf("CLI reactivated non-open audit: %v", err)
					}
				}
				runRoot(t, "-C", root, "audit", "reopen", "probe")
				runRoot(t, "-C", root, "audit", "finding", "probe", "H1", "--status", "in-progress")
				_, err := runRootStreams(t, "-C", root, "audit", "close", "probe")
				if ExitCode(err) != 11 {
					t.Fatalf("active reopened finding must prevent closure: %v", err)
				}
				runRoot(t, "-C", root, "audit", "finding", "probe", "H1", "--status", "fixed")
				runRoot(t, "-C", root, "audit", "close", "probe")
			})
		}
	}
}
