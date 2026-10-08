package cli

import (
	"encoding/json"
	"maps"
	"strings"
	"testing"

	"github.com/andy-esch/taskflow/internal/testutil"
	"github.com/andy-esch/taskflow/internal/wire"
)

func TestAuditReadSurfacesQualifyUnparsedFindings(t *testing.T) {
	for _, tc := range []struct {
		name, body       string
		parsed, unparsed int
		ready            bool
	}{
		{"empty", "# Audit\n", 0, 0, false},
		{"unparsed", "## Findings\n\n#### H-1. Issue · **Status:** open\n", 0, 1, false},
		{"ambiguous", "## Findings\n\n#### H1 Missing metadata\n", 0, 1, false},
		{"mixed", "#### M1. Real · **Status:** fixed\n\n#### H-1. Lost · **Status:** open\n", 1, 1, false},
		{"settled", "#### M1. Real · **Status:** fixed\n", 1, 0, true},
		{"ordinary", "### S3 Storage Architecture\n\n```md\n#### H-1. Example · **Status:** open\n```\n", 0, 0, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			root := setupRepo(t)
			path, content := testutil.AuditFixture(root, "open", "probe.md",
				"---\narea: probe\ndate: 2026-10-06\n---\n"+tc.body)
			testutil.Write(t, path, content)
			for _, args := range [][]string{{"audit", "list", "--color=never"}, {"audit", "show", "probe", "--frontmatter-only", "--color=never"}, {"audit", "info", "probe", "--color=never"}} {
				out := runRoot(t, append([]string{"-C", root}, args...)...)
				if tc.unparsed > 0 {
					if !strings.Contains(out, "unparsed") || !strings.Contains(out, "audit lint "+testutil.TaskID("probe")) || strings.Contains(out, "no findings") || strings.Contains(out, "ready to close") {
						t.Fatalf("read misrepresented incomplete findings:\n%s", out)
					}
				} else if strings.Contains(out, "unparsed") || strings.Contains(out, "audit lint ") {
					t.Fatalf("clean audit got warning:\n%s", out)
				}
			}
			out := runRoot(t, "-C", root, "audit", "show", "probe", "--frontmatter-only", "--json")
			var shown wire.AuditShowEnvelope
			if err := json.Unmarshal([]byte(out), &shown); err != nil {
				t.Fatal(err)
			}
			out = runRoot(t, "-C", root, "audit", "list", "--json")
			var listed wire.AuditsEnvelope
			if err := json.Unmarshal([]byte(out), &listed); err != nil {
				t.Fatal(err)
			}
			if len(listed.Audits) != 1 {
				t.Fatalf("unexpected list: %+v", listed)
			}
			for _, audit := range []wire.AuditJSON{shown.Audit, listed.Audits[0]} {
				if audit.Findings != tc.parsed || audit.UnparsedFindings != tc.unparsed || audit.ReadyToClose != tc.ready {
					t.Fatalf("wrong body-derived evidence: %+v", audit)
				}
			}
			infoOut := runRoot(t, "-C", root, "audit", "info", "probe", "--json")
			var info wire.AuditInfoEnvelope
			if err := json.Unmarshal([]byte(infoOut), &info); err != nil {
				t.Fatal(err)
			}
			if info.SchemaVersion != wire.SchemaVersion || info.AuditInfo.Findings.Total != tc.parsed || info.AuditInfo.UnparsedFindings != tc.unparsed {
				t.Fatalf("info lost or mixed finding evidence: %+v", info)
			}
			if strings.Contains(infoOut, `"unparsed_findings"`) != (tc.unparsed > 0) {
				t.Fatalf("info optional count presence disagrees with evidence: %s", infoOut)
			}
			for _, selector := range []string{"unparsed", "unparsed_findings"} {
				out = runRoot(t, "-C", root, "audit", "list", "--json", "-c", "slug,"+selector)
				var projected struct {
					Audits []map[string]string `json:"audits"`
				}
				if err := json.Unmarshal([]byte(out), &projected); err != nil {
					t.Fatal(err)
				}
				want := ""
				if tc.unparsed > 0 {
					want = "1"
				}
				if got := projected.Audits[0][selector]; got != want {
					t.Fatalf("projection %s=%q want %q", selector, got, want)
				}
			}
		})
	}
}

func TestAuditMoveRefusalsExposeIncompleteEvidenceAndExitValidation(t *testing.T) {
	for _, body := range []string{"## Findings\n\n#### H-1. Lost · **Status:** fixed\n", "## Findings\n\n#### H1 Missing metadata\n"} {
		for _, verb := range []string{"close", "defer"} {
			for _, dry := range []bool{false, true} {
				root := setupRepo(t)
				path, content := testutil.AuditFixture(root, "open", "probe.md", "---\narea: probe\ndate: 2026-10-07\n---\n"+body)
				testutil.Write(t, path, content)
				before := testutil.SnapshotTree(t, root)
				args := []string{"-C", root, "audit", verb, "probe", "--json"}
				if dry {
					args = append(args, "--dry-run")
				}
				result, err := runRootStreams(t, args...)
				if ExitCode(err) != 11 {
					t.Fatalf("%s dry=%t: err=%v, want exit 11", verb, dry, err)
				}
				var receipt wire.MovesEnvelope
				if err := json.Unmarshal([]byte(result.Out), &receipt); err != nil {
					t.Fatal(err)
				}
				if receipt.SchemaVersion != wire.SchemaVersion || receipt.DryRun != dry || len(receipt.Moves) != 1 || !strings.Contains(receipt.Moves[0].Error, "unparsed") || !strings.Contains(receipt.Moves[0].Error, "audit lint "+testutil.TaskID("probe")) {
					t.Fatalf("refusal receipt omitted diagnosis: %+v", receipt)
				}
				if !maps.Equal(before, testutil.SnapshotTree(t, root)) {
					t.Fatal("refused command changed planning tree")
				}
				// Execute the recommended diagnostic route, not just a string match.
				lint, err := runRootStreams(t, "-C", root, "audit", "lint", "probe", "--json")
				if ExitCode(err) != 11 || !strings.Contains(lint.Out, "findings") {
					t.Fatalf("diagnostic route failed: out=%s err=%v", lint.Out, err)
				}
			}
		}
	}
}
