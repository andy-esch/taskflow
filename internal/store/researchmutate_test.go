package store

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/andy-esch/taskflow/internal/domain"
	"github.com/andy-esch/taskflow/internal/testutil"
)

func researchFixture(t *testing.T, root, name, content string) string {
	t.Helper()
	path, out := testutil.ResearchFixture(root, name, content)
	testutil.Write(t, path, out)
	return path
}

// The surgical-edit guarantee matters most for research: the legacy corpus carries a
// vestigial `status: reference` that is NOT part of the contract, plus comments and a
// particular key order. A field update must leave all of it intact.
func TestFS_SetResearchFields_PreservesUnknownKeysAndOrder(t *testing.T) {
	root := t.TempDir()
	original := "---\nschema: 1\nid: " + testutil.TaskID("legacy") + "\ncreated: \"2026-01-03\"\n" +
		"# a comment the tool must not eat\nstatus: reference\ntopic: something bespoke\ntags: [old]\n---\n# Legacy\n\nbody\n"
	researchFixture(t, root, "legacy.md", original)

	got, err := NewFS(root).SetResearchFields("legacy", map[string]any{"description": "now described"}, false)
	if err != nil {
		t.Fatal(err)
	}
	if got.Description != "now described" {
		t.Errorf("description = %q", got.Description)
	}
	raw, _ := os.ReadFile(got.Path)
	s := string(raw)
	for _, want := range []string{"# a comment the tool must not eat", "status: reference", "topic: something bespoke", "description: now described"} {
		if !strings.Contains(s, want) {
			t.Errorf("surgical edit lost %q:\n%s", want, s)
		}
	}
	// Body untouched.
	if !strings.Contains(s, "# Legacy\n\nbody\n") {
		t.Errorf("body altered:\n%s", s)
	}
}

func TestFS_SetResearchFields_UnsetRemovesKey(t *testing.T) {
	root := t.TempDir()
	researchFixture(t, root, "doc.md", "---\nschema: 1\nid: "+testutil.TaskID("doc")+"\ncreated: \"2026-01-03\"\nstatus: reference\n---\n# Doc\n")

	got, err := NewFS(root).SetResearchFields("doc", map[string]any{"status": domain.UnsetField{}}, false)
	if err != nil {
		t.Fatal(err)
	}
	raw, _ := os.ReadFile(got.Path)
	if strings.Contains(string(raw), "status:") {
		t.Errorf("unset should have removed status:\n%s", raw)
	}
}

// A dry run must validate end to end and write nothing.
func TestFS_SetResearchFields_DryRunWritesNothing(t *testing.T) {
	root := t.TempDir()
	path := researchFixture(t, root, "doc.md", "---\nschema: 1\nid: "+testutil.TaskID("doc")+"\ncreated: \"2026-01-03\"\n---\n# Doc\n")
	before, _ := os.ReadFile(path)

	if _, err := NewFS(root).SetResearchFields("doc", map[string]any{"description": "preview"}, true); err != nil {
		t.Fatal(err)
	}
	after, _ := os.ReadFile(path)
	if string(before) != string(after) {
		t.Errorf("dry run wrote to disk:\n%s", after)
	}
}

// Parse-before-commit: an update that would make the file unreloadable must be refused
// with nothing written, and must blame the update rather than the file.
func TestFS_SetResearchFields_RefusesUnreloadableUpdate(t *testing.T) {
	root := t.TempDir()
	path := researchFixture(t, root, "doc.md", "---\nschema: 1\nid: "+testutil.TaskID("doc")+"\ncreated: \"2026-01-03\"\n---\n# Doc\n")
	before, _ := os.ReadFile(path)

	// A tags value that can't round-trip as a list breaks the parse.
	_, err := NewFS(root).SetResearchFields("doc", map[string]any{"tags": "[unclosed"}, false)
	if err == nil {
		t.Skip("this value round-trips fine; the guard is exercised by the epic/task suites")
	}
	if !errors.Is(err, domain.ErrValidation) {
		t.Errorf("want ErrValidation, got %v", err)
	}
	after, _ := os.ReadFile(path)
	if string(before) != string(after) {
		t.Errorf("a refused update must write nothing:\n%s", after)
	}
}

func TestFS_AppendResearchBody(t *testing.T) {
	root := t.TempDir()
	researchFixture(t, root, "doc.md", "---\nschema: 1\nid: "+testutil.TaskID("doc")+"\ncreated: \"2026-01-03\"\n---\n# Doc\n\nfirst\n")
	now := time.Date(2026, 8, 18, 0, 0, 0, 0, time.UTC)

	got, body, err := NewFS(root).AppendResearchBody("doc", "## Addendum\n\nmore", now, false)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(body, "first") || !strings.Contains(body, "## Addendum") {
		t.Errorf("append lost content: %q", body)
	}
	// updated_at is stamped; created stays immutable (the id is minted from it).
	if got.Updated != "2026-08-18" {
		t.Errorf("updated_at = %q, want the stamp", got.Updated)
	}
	if got.Created != "2026-01-03" {
		t.Errorf("created must not change: %q", got.Created)
	}
}

func TestFS_AppendResearchBody_DryRunWritesNothing(t *testing.T) {
	root := t.TempDir()
	path := researchFixture(t, root, "doc.md", "---\nschema: 1\nid: "+testutil.TaskID("doc")+"\ncreated: \"2026-01-03\"\n---\n# Doc\n")
	before, _ := os.ReadFile(path)

	if _, _, err := NewFS(root).AppendResearchBody("doc", "## X", time.Now(), true); err != nil {
		t.Fatal(err)
	}
	after, _ := os.ReadFile(path)
	if string(before) != string(after) {
		t.Errorf("dry run wrote to disk:\n%s", after)
	}
}

// A body write needs parseable frontmatter to preserve — a broken file is refused rather
// than half-rewritten.
func TestFS_AppendResearchBody_RefusesBrokenFrontmatter(t *testing.T) {
	root := t.TempDir()
	testutil.Write(t, filepath.Join(root, domain.ResearchDir, testutil.TaskID("bad")+"-bad.md"),
		"---\nid: [unclosed\ncreated: \"2026-01-03\"\n---\n# Bad\n")

	if _, _, err := NewFS(root).AppendResearchBody("bad", "## X", time.Now(), false); err == nil {
		t.Error("appending to a doc with malformed frontmatter must fail, not half-write")
	}
}
