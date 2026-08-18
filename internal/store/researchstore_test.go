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

func writeResearch(t *testing.T, root, name, content string) {
	t.Helper()
	path, out := testutil.ResearchFixture(root, name, content)
	testutil.Write(t, path, out)
}

func TestFS_ListResearch(t *testing.T) {
	root := t.TempDir()
	writeResearch(t, root, "theming.md", "---\ncreated: 2026-06-23\ndescription: Weighed three libs\ntags: [tui, color]\n---\n# Theming\n\nbody\n")

	docs, problems, err := NewFS(root).ListResearch()
	if err != nil {
		t.Fatal(err)
	}
	if len(problems) != 0 {
		t.Fatalf("unexpected problems: %+v", problems)
	}
	if len(docs) != 1 {
		t.Fatalf("want 1 doc, got %d", len(docs))
	}
	r := docs[0]
	if r.Slug != "theming" || r.Created != "2026-06-23" || r.Description != "Weighed three libs" {
		t.Errorf("metadata wrong: %+v", r)
	}
	if len(r.Tags) != 2 || r.Tags[0] != "tui" {
		t.Errorf("tags wrong: %+v", r.Tags)
	}
	// FilenameID is the canonical key, parsed from the id-led name.
	if r.FilenameID != testutil.TaskID("theming") {
		t.Errorf("FilenameID = %q, want the filename's leading id", r.FilenameID)
	}
}

// A research doc's vestigial `status:` (18 legacy docs carry `status: reference`) is
// NOT part of the contract, and must ride along as an unknown field rather than
// failing the parse — research has no status vocabulary to validate it against.
func TestFS_ListResearch_UnknownFieldsPreserved(t *testing.T) {
	root := t.TempDir()
	writeResearch(t, root, "legacy.md", "---\nstatus: reference\ncreated: 2026-01-03\n---\n# Legacy\n")

	docs, problems, err := NewFS(root).ListResearch()
	if err != nil {
		t.Fatal(err)
	}
	if len(problems) != 0 {
		t.Fatalf("a vestigial status: must not be a problem, got %+v", problems)
	}
	if len(docs) != 1 || docs[0].Created != "2026-01-03" {
		t.Fatalf("want the doc parsed with its date, got %+v", docs)
	}
}

// A fence-less research file is a loud FileProblem naming the shape, not a silently
// empty doc — the same contract as tasks/audits.
func TestFS_ListResearch_MissingFrontmatterIsLoud(t *testing.T) {
	root := t.TempDir()
	writeResearch(t, root, "notes.md", "# Some notes\n\nno frontmatter\n")

	docs, problems, err := NewFS(root).ListResearch()
	if err != nil {
		t.Fatal(err)
	}
	if len(docs) != 0 {
		t.Errorf("a fence-less file must not parse as research, got %+v", docs)
	}
	if len(problems) != 1 || !strings.Contains(problems[0].Message, "missing frontmatter") ||
		!strings.Contains(problems[0].Message, "schema research") {
		t.Errorf("want one loud, shape-naming problem, got %+v", problems)
	}
}

// A non-id-led .md in research/ is the ADR-0003 carveout gate: a FileProblem telling
// the author to move it to meta/, not a parsed entity. This is exactly what the 28
// pre-migration docs look like, so the message matters.
func TestFS_ListResearch_NonIDLedIsFileProblem(t *testing.T) {
	root := t.TempDir()
	testutil.Write(t, filepath.Join(root, domain.ResearchDir, "2026-06-28-legacy-name.md"),
		"---\ncreated: 2026-06-28\n---\n# Legacy\n")

	docs, problems, err := NewFS(root).ListResearch()
	if err != nil {
		t.Fatal(err)
	}
	if len(docs) != 0 {
		t.Errorf("a non-id-led file must not parse as research, got %+v", docs)
	}
	if len(problems) != 1 || !strings.Contains(problems[0].Message, "no leading id") {
		t.Errorf("want the carveout-gate problem, got %+v", problems)
	}
}

func TestFS_CreateResearch(t *testing.T) {
	root := t.TempDir()
	fs := NewFS(root)
	r := domain.Research{Slug: "storage-model", ID: "6dvxwxg034xm", Created: "2026-01-15", Tags: []string{"core"}}

	got, err := fs.CreateResearch(r, "# Storage model\n", false)
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Join(root, domain.ResearchDir, "6dvxwxg034xm-storage-model.md")
	if got.Path != want {
		t.Errorf("path = %q, want %q", got.Path, want)
	}
	content, err := os.ReadFile(want)
	if err != nil {
		t.Fatal(err)
	}
	// Frontmatter order is the researchFields contract, and there must be no
	// status/bucket key — research has no lifecycle.
	for _, want := range []string{"schema: 1", "id: 6dvxwxg034xm", `created: "2026-01-15"`, "tags: [core]"} {
		if !strings.Contains(string(content), want) {
			t.Errorf("created file missing %q:\n%s", want, content)
		}
	}
	if strings.Contains(string(content), "status:") || strings.Contains(string(content), "bucket:") {
		t.Errorf("research must not carry a status/bucket:\n%s", content)
	}
	// And it round-trips through the parser.
	docs, problems, err := fs.ListResearch()
	if err != nil || len(problems) != 0 {
		t.Fatalf("round-trip failed: err=%v problems=%+v", err, problems)
	}
	if len(docs) != 1 || docs[0].Slug != "storage-model" || docs[0].Created != "2026-01-15" {
		t.Errorf("round-trip mismatch: %+v", docs)
	}
}

func TestFS_CreateResearch_RefusesClobber(t *testing.T) {
	root := t.TempDir()
	fs := NewFS(root)
	r := domain.Research{Slug: "dup", ID: "6dvxwxg034xm", Created: "2026-01-15"}
	if _, err := fs.CreateResearch(r, "# Dup\n", false); err != nil {
		t.Fatal(err)
	}
	_, err := fs.CreateResearch(r, "# Dup again\n", false)
	if !errors.Is(err, domain.ErrConflict) {
		t.Errorf("re-create must be ErrConflict, got %v", err)
	}
}

func TestFS_GetResearch_FuzzyAndByID(t *testing.T) {
	root := t.TempDir()
	writeResearch(t, root, "storage-model-options.md", "---\ncreated: 2026-01-15\n---\n# Storage\n\nprose\n")
	fs := NewFS(root)

	// A unique prefix resolves, as does the exact stable id.
	for _, key := range []string{"storage-model-options", "storage-model", testutil.TaskID("storage-model-options")} {
		r, body, err := fs.GetResearch(key)
		if err != nil {
			t.Fatalf("GetResearch(%q): %v", key, err)
		}
		if r.Slug != "storage-model-options" {
			t.Errorf("GetResearch(%q) resolved to %q", key, r.Slug)
		}
		if !strings.Contains(body, "prose") {
			t.Errorf("GetResearch(%q) body = %q", key, body)
		}
	}
	if _, _, err := fs.GetResearch("nope"); !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("unknown slug must be ErrNotFound, got %v", err)
	}
}

// The research dir must be watched, or a doc added at runtime stays invisible to the TUI.
func TestFS_WatchPaths_IncludesResearch(t *testing.T) {
	root := t.TempDir()
	want := filepath.Join(root, domain.ResearchDir)
	for _, p := range NewFS(root).WatchPaths() {
		if p == want {
			return
		}
	}
	t.Errorf("WatchPaths missing %q: %+v", want, NewFS(root).WatchPaths())
}

// lint's MissingIDMessage promises "`lint --fix` assigns one". That was a dead end for
// research until researchDir joined the fixer's dir list, so assert the repair really
// happens rather than trusting the message.
func TestFS_FixFrontmatter_BackfillsResearchID(t *testing.T) {
	root := t.TempDir()
	stem := testutil.TaskID("no-id") + "-no-id.md"
	testutil.Write(t, filepath.Join(root, domain.ResearchDir, stem), "---\nschema: 1\ncreated: \"2026-01-03\"\n---\n# No id\n")
	fs := NewFS(root)

	results, err := fs.FixFrontmatter(false)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 || !strings.Contains(strings.Join(results[0].Changes, ","), "id") {
		t.Fatalf("want one id-backfill result, got %+v", results)
	}
	// And the backfilled id is the filename's id (the canonical key), so no drift.
	docs, problems, err := fs.ListResearch()
	if err != nil || len(problems) != 0 {
		t.Fatalf("err=%v problems=%+v", err, problems)
	}
	if len(docs) != 1 || docs[0].ID != testutil.TaskID("no-id") {
		t.Errorf("id = %q, want the filename id %q", docs[0].ID, testutil.TaskID("no-id"))
	}
	if len(domain.LintResearch(docs[0])) != 0 {
		t.Errorf("doc should lint clean after --fix, got %+v", domain.LintResearch(docs[0]))
	}
}

// createFileAtomic stages its temp in the TARGET dir, so a crashed `research new` leaves
// an orphan in research/ — which the sweep skipped until researchDir was added.
func TestFS_FixFrontmatter_SweepsResearchTempOrphan(t *testing.T) {
	root := t.TempDir()
	orphan := filepath.Join(root, domain.ResearchDir, ".tskflwctl-crashed.tmp")
	testutil.Write(t, orphan, "partial")
	old := time.Now().Add(-2 * time.Hour)
	if err := os.Chtimes(orphan, old, old); err != nil {
		t.Fatal(err)
	}

	if _, err := NewFS(root).FixFrontmatter(false); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(orphan); !os.IsNotExist(err) {
		t.Errorf("aged temp orphan in research/ should be swept, stat err = %v", err)
	}
}

// A fresh temp (a concurrent write in flight) must NOT be swept.
func TestFS_FixFrontmatter_LeavesFreshResearchTemp(t *testing.T) {
	root := t.TempDir()
	fresh := filepath.Join(root, domain.ResearchDir, ".tskflwctl-inflight.tmp")
	testutil.Write(t, fresh, "in flight")

	if _, err := NewFS(root).FixFrontmatter(false); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(fresh); err != nil {
		t.Errorf("a fresh temp must survive the sweep: %v", err)
	}
}
