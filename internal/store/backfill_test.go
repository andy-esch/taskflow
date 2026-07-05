package store

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	yaml "go.yaml.in/yaml/v3"

	"github.com/andy-esch/taskflow/internal/testutil"
)

func seedFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func frontmatterID(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	fm, _ := splitFrontmatter(b)
	var m struct {
		ID string `yaml:"id"`
	}
	if err := yaml.Unmarshal(fm, &m); err != nil {
		t.Fatalf("parse %s: %v", path, err)
	}
	return m.ID
}

// filenameID is the id-prefix of a flat <id>-<slug>.md file — the canonical id a
// backfill copies into the frontmatter.
func filenameID(t *testing.T, path string) string {
	t.Helper()
	fnID, _, ok := splitFlatName(strings.TrimSuffix(filepath.Base(path), ".md"))
	if !ok {
		t.Fatalf("fixture %s is not id-led", path)
	}
	return fnID
}

func changesHave(changes []string, want string) bool {
	for _, c := range changes {
		if c == want {
			return true
		}
	}
	return false
}

const idlessTask = "---\nstatus: ready-to-start\nepic: e1\ntier: 2\npriority: high\neffort: 2h\ncreated: 2026-01-05\ntags: [a]\n---\n# T\n"

// TestFixFrontmatter_BackfillsMissingTaskID: a task with no frontmatter id gets the
// id that already leads its flat filename — no fresh mint, so the two agree.
func TestFixFrontmatter_BackfillsMissingTaskID(t *testing.T) {
	root := t.TempDir()
	p, out := testutil.TaskFixture(root, "ready-to-start", "t.md", idlessTask)
	seedFile(t, p, out)

	results, err := NewFS(root).FixFrontmatter(false)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 || !changesHave(results[0].Changes, "id: assigned (was missing)") {
		t.Fatalf("expected one id-assigned change, got %+v", results)
	}
	if got := frontmatterID(t, p); got != filenameID(t, p) {
		t.Errorf("backfilled id %q must equal the filename id %q", got, filenameID(t, p))
	}
}

// TestFixFrontmatter_BackfillsMissingAuditID: audits get the same filename-sourced id.
func TestFixFrontmatter_BackfillsMissingAuditID(t *testing.T) {
	root := t.TempDir()
	p, out := testutil.AuditFixture(root, "open", "2026-01-02-x.md", "---\narea: x\ndate: 2026-01-02\n---\n#### H1. t  · **Status:** open\n")
	seedFile(t, p, out)

	if _, err := NewFS(root).FixFrontmatter(false); err != nil {
		t.Fatal(err)
	}
	if got := frontmatterID(t, p); got != filenameID(t, p) {
		t.Errorf("audit backfilled id %q must equal the filename id %q", got, filenameID(t, p))
	}
}

// TestFixFrontmatter_KeepsExistingID: a file that already has an id is untouched.
func TestFixFrontmatter_KeepsExistingID(t *testing.T) {
	root := t.TempDir()
	existing := testutil.TaskID("t") // matches the flat filename, so no drift
	p, out := testutil.TaskFixture(root, "ready-to-start", "t.md", "---\nid: "+existing+"\nstatus: ready-to-start\nepic: e1\ntier: 2\npriority: high\neffort: 2h\ncreated: 2026-01-05\ntags: [a]\n---\n# T\n")
	seedFile(t, p, out)

	results, err := NewFS(root).FixFrontmatter(false)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 0 {
		t.Errorf("a task that already has an id needs no fix, got %+v", results)
	}
	if got := frontmatterID(t, p); got != existing {
		t.Errorf("existing id must be preserved, got %q want %q", got, existing)
	}
}

// TestFixFrontmatter_BackfillsWithoutDate: the id comes from the filename, so a task
// carrying no date field at all is still backfilled (the old date-mint path needed
// one; the flat filename always has the id).
func TestFixFrontmatter_BackfillsWithoutDate(t *testing.T) {
	root := t.TempDir()
	p, out := testutil.TaskFixture(root, "ready-to-start", "t.md", "---\nstatus: ready-to-start\nepic: e1\ntier: 2\npriority: high\neffort: 2h\ntags: [a]\n---\n# T\n")
	seedFile(t, p, out)

	if _, err := NewFS(root).FixFrontmatter(false); err != nil {
		t.Fatal(err)
	}
	if got := frontmatterID(t, p); got != filenameID(t, p) {
		t.Errorf("a dateless task is still backfilled from the filename: got %q, want %q", got, filenameID(t, p))
	}
}

// TestFixFrontmatter_SkipsStrayWithoutIDLedName: a non-id-led .md (a stray the scan
// gate flags) is left alone — minting an id into its frontmatter wouldn't make it an
// entity, so backfill no-ops and the file stays a FileProblem for the operator.
func TestFixFrontmatter_SkipsStrayWithoutIDLedName(t *testing.T) {
	root := t.TempDir()
	p := filepath.Join(root, "tasks", "not-an-entity.md")
	seedFile(t, p, idlessTask)

	results, err := NewFS(root).FixFrontmatter(false)
	if err != nil {
		t.Fatal(err)
	}
	for _, r := range results {
		if changesHave(r.Changes, "id: assigned (was missing)") {
			t.Errorf("a non-id-led stray must not be backfilled: %+v", r)
		}
	}
	if got := frontmatterID(t, p); got != "" {
		t.Errorf("expected no id assigned to a stray, got %q", got)
	}
}
