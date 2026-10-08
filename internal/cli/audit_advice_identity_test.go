package cli

import (
	"encoding/json"
	"fmt"
	"maps"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/andy-esch/taskflow/internal/domain"
	"github.com/andy-esch/taskflow/internal/testutil"
	"github.com/andy-esch/taskflow/internal/wire"
)

// Execute the advice, not merely its literal text: slugs and declarations can
// shadow another record's ID even though the initiating read/move uses the right ID.
func TestAuditIncompleteEvidenceAdviceRetainsSourceIdentity(t *testing.T) {
	const sourceID, siblingID = "6g0000000011", "6g0000000012"
	for _, tc := range []struct{ name, slug, siblingSlug, declaredID string }{
		{"duplicate-slug", "probe", "probe", sourceID},
		{"id-shadowing-slug", siblingID, "clean", sourceID},
		{"declared-id-drift", siblingID, "clean", siblingID},
	} {
		t.Run(tc.name, func(t *testing.T) {
			root := setupRepo(t)
			for _, record := range []struct{ id, slug, declaredID, body string }{
				{sourceID, tc.slug, tc.declaredID, "## Findings\n\n#### H1 Missing metadata\n"},
				{siblingID, tc.siblingSlug, siblingID, "# Empty audit\n"},
			} {
				path := filepath.Join(root, domain.AuditsDir, record.id+"-"+record.slug+".md")
				testutil.Write(t, path, fmt.Sprintf("---\nschema: 1\nid: %s\nbucket: open\narea: probe\ndate: \"2026-10-07\"\n---\n%s", record.declaredID, record.body))
			}
			before := testutil.SnapshotTree(t, root)
			checkAdvice := func(text string) {
				t.Helper()
				matches := regexp.MustCompile(`audit lint ([a-zA-Z0-9-]+)`).FindAllStringSubmatch(text, -1)
				if len(matches) == 0 {
					t.Fatalf("missing executable advice: %s", text)
				}
				for _, match := range matches {
					if match[1] != sourceID {
						t.Fatalf("advice chose %s instead of source %s: %s", match[1], sourceID, text)
					}
					lint, err := runRootStreams(t, "-C", root, "audit", "lint", match[1], "--json")
					if ExitCode(err) != 11 || !strings.Contains(lint.Out, "H1 Missing metadata") {
						t.Fatalf("emitted route missed original evidence: out=%s err=%v", lint.Out, err)
					}
				}
			}
			for _, args := range [][]string{
				{"audit", "info", sourceID}, {"audit", "show", sourceID, "--frontmatter-only"},
				{"audit", "list"}, {"status"},
			} {
				checkAdvice(runRoot(t, append([]string{"-C", root, "--color=never"}, args...)...))
			}
			for _, verb := range []string{"close", "defer"} {
				for _, dry := range []bool{false, true} {
					args := []string{"-C", root, "audit", verb, sourceID, "--json"}
					if dry {
						args = append(args, "--dry-run")
					}
					result, err := runRootStreams(t, args...)
					if ExitCode(err) != 11 {
						t.Fatalf("%s dry=%t: %v", verb, dry, err)
					}
					var receipt wire.MovesEnvelope
					if err := json.Unmarshal([]byte(result.Out), &receipt); err != nil || len(receipt.Moves) != 1 {
						t.Fatalf("refusal receipt: %s, %v", result.Out, err)
					}
					checkAdvice(receipt.Moves[0].Error)
				}
			}
			if !maps.Equal(before, testutil.SnapshotTree(t, root)) {
				t.Fatal("read, refusal, or emitted diagnostic changed the planning tree")
			}
		})
	}
}
