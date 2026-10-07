package cli

import (
	"encoding/json"
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
			for _, args := range [][]string{{"audit", "list", "--color=never"}, {"audit", "show", "probe", "--frontmatter-only", "--color=never"}} {
				out := runRoot(t, append([]string{"-C", root}, args...)...)
				if tc.unparsed > 0 {
					if !strings.Contains(out, "unparsed") || !strings.Contains(out, "audit lint probe") || strings.Contains(out, "no findings") || strings.Contains(out, "ready to close") {
						t.Fatalf("read misrepresented incomplete findings:\n%s", out)
					}
				} else if strings.Contains(out, "unparsed") || strings.Contains(out, "audit lint probe") {
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
