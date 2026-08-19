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
	// Byte-exact on the whole frontmatter block. A `strings.Contains` per key passes even if
	// every key is REORDERED, which is half of what this test's name promises — verified by
	// mutation: reversing the key order in updateFrontmatter slipped past the old assertion.
	wantFM := "---\nschema: 1\nid: " + testutil.TaskID("legacy") + "\ncreated: \"2026-01-03\"\n" +
		"# a comment the tool must not eat\nstatus: reference\ntopic: something bespoke\ntags: [old]\n" +
		"description: now described\n---\n"
	if !strings.HasPrefix(s, wantFM) {
		t.Errorf("frontmatter not preserved byte-for-byte (comments, unknown keys, and ORDER).\ngot:\n%s\nwant prefix:\n%s", s, wantFM)
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

	// A scalar written into the list-typed `tags` cannot unmarshal, so the reload fails.
	// No t.Skip: a skip here would silently retire the guard behind a green suite.
	_, err := NewFS(root).SetResearchFields("doc", map[string]any{"tags": "not-a-list"}, false)
	if err == nil {
		t.Fatal("an update that makes the file unreloadable must be refused")
	}
	if !errors.Is(err, domain.ErrValidation) {
		t.Errorf("want ErrValidation, got %v", err)
	}
	// The ATTRIBUTION is the point: blame the update, not the (valid) file. Both branches
	// wrap ErrValidation, so only the message distinguishes them.
	if !strings.Contains(err.Error(), "would not reload") {
		t.Errorf("must blame the update, not the file: %v", err)
	}
	if strings.Contains(err.Error(), "not caused by this update") {
		t.Errorf("wrongly blamed the file for a valid original: %v", err)
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

// A duplicate id on a DIFFERENT slug is a different filename, so writeNewFile's O_EXCL
// never sees it — the store must check the id explicitly. Two docs on one id are
// unresolvable by id AND both become unwritable (the write CAS goes ambiguous), so this
// has to be refused at create.
func TestFS_CreateResearch_RefusesDuplicateID(t *testing.T) {
	root := t.TempDir()
	fs := NewFS(root)
	const shared = "6dr29v000zzr"
	if _, err := fs.CreateResearch(domain.Research{Slug: "alpha", ID: shared, Created: "2026-01-03"}, "# A\n", false); err != nil {
		t.Fatal(err)
	}

	// Same id, DIFFERENT slug -> different path, so only an explicit id check catches it.
	_, err := fs.CreateResearch(domain.Research{Slug: "beta", ID: shared, Created: "2026-01-03"}, "# B\n", false)
	if !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("a duplicate id must be ErrConflict, got %v", err)
	}
	if !strings.Contains(err.Error(), "alpha") {
		t.Errorf("the error should name the doc already holding the id: %v", err)
	}
	// And nothing was written.
	if m, _ := filepath.Glob(filepath.Join(root, domain.ResearchDir, "*beta*")); len(m) != 0 {
		t.Errorf("a refused create must write nothing, found %v", m)
	}
}

// The dry-run path must apply the same check — a preview that would fail must fail.
func TestFS_CreateResearch_DuplicateIDRefusedOnDryRun(t *testing.T) {
	root := t.TempDir()
	fs := NewFS(root)
	const shared = "6dr29v000zzr"
	if _, err := fs.CreateResearch(domain.Research{Slug: "alpha", ID: shared, Created: "2026-01-03"}, "# A\n", false); err != nil {
		t.Fatal(err)
	}
	if _, err := fs.CreateResearch(domain.Research{Slug: "beta", ID: shared, Created: "2026-01-03"}, "# B\n", true); !errors.Is(err, domain.ErrConflict) {
		t.Errorf("dry run must refuse a duplicate id too, got %v", err)
	}
}

// A write must not be more permissive than a read. splitFrontmatterStrict returns a nil
// block (not an error) for a fence-less file, and updateFrontmatter/documentMapping would
// CREATE one — so without a guard these succeed and leave the doc with no id and no
// `created`. This is exactly the shape the migration reported for 10 legacy docs.
func TestFS_ResearchWritePaths_RefuseFrontmatterlessDoc(t *testing.T) {
	body := "# Legacy doc\n\n**Created**: 2026-01-03\n\nprose\n"
	cases := []struct {
		name  string
		write func(*FS) error
	}{
		{"SetResearchFields", func(fs *FS) error {
			_, err := fs.SetResearchFields("legacy", map[string]any{"description": "d"}, false)
			return err
		}},
		{"AppendResearchBody", func(fs *FS) error {
			_, _, err := fs.AppendResearchBody("legacy", "## Added", time.Now(), false)
			return err
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			root := t.TempDir()
			path := researchFixture(t, root, "legacy.md", body)

			err := tc.write(NewFS(root))
			if err == nil {
				t.Fatal("writing to a frontmatter-less doc must be refused, not fabricate a block")
			}
			if !errors.Is(err, domain.ErrValidation) {
				t.Errorf("want ErrValidation (exit 11), got %v", err)
			}
			if !strings.Contains(err.Error(), "missing frontmatter") {
				t.Errorf("error should name the real problem: %v", err)
			}
			got, _ := os.ReadFile(path)
			if string(got) != body {
				t.Errorf("a refused write must leave the file byte-identical:\n%s", got)
			}
		})
	}
}

// The version-CAS: a concurrent in-place edit landing between our validation and our write
// must lose, not clobber. This is the guarantee the code comments promise and that nothing
// covered — deleting the verifyUnchanged call left the entire suite green.
func TestFS_SetResearchFields_ConcurrentEditConflicts(t *testing.T) {
	root := t.TempDir()
	fmBase := "---\nschema: 1\nid: " + testutil.TaskID("doc") + "\ncreated: \"2026-01-03\"\n"
	path := researchFixture(t, root, "doc.md", fmBase+"description: original\n---\n# Doc\n")
	fs := NewFS(root)

	orig := testHookBeforeResearchWrite
	defer func() { testHookBeforeResearchWrite = orig }()
	testHookBeforeResearchWrite = func() {
		// A different writer lands an in-place edit in the read->write window.
		_ = os.WriteFile(path, []byte(fmBase+"description: CHANGED BY OTHER\n---\n# Doc\n"), 0o644)
		testHookBeforeResearchWrite = orig // fire once
	}

	_, err := fs.SetResearchFields("doc", map[string]any{"description": "mine"}, false)
	if !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("a concurrent in-place edit must conflict (exit 14), got %v", err)
	}
	got, _ := os.ReadFile(path)
	if !strings.Contains(string(got), "CHANGED BY OTHER") || strings.Contains(string(got), "mine") {
		t.Errorf("the losing write must not clobber the concurrent edit:\n%s", got)
	}
}

// The same guarantee on the body path (AppendResearchBody goes through writeBody, whose
// recheck closure is the research CAS).
func TestFS_AppendResearchBody_ConcurrentEditConflicts(t *testing.T) {
	root := t.TempDir()
	fmBase := "---\nschema: 1\nid: " + testutil.TaskID("doc") + "\ncreated: \"2026-01-03\"\n---\n"
	path := researchFixture(t, root, "doc.md", fmBase+"# Doc\n\noriginal\n")
	fs := NewFS(root)

	orig := testHookBeforeBodyWrite
	defer func() { testHookBeforeBodyWrite = orig }()
	testHookBeforeBodyWrite = func() {
		_ = os.WriteFile(path, []byte(fmBase+"# Doc\n\nCHANGED BY OTHER\n"), 0o644)
		testHookBeforeBodyWrite = orig
	}

	_, _, err := fs.AppendResearchBody("doc", "## Mine", time.Now(), false)
	if !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("a concurrent edit during append must conflict, got %v", err)
	}
	got, _ := os.ReadFile(path)
	if !strings.Contains(string(got), "CHANGED BY OTHER") || strings.Contains(string(got), "## Mine") {
		t.Errorf("the losing append must not clobber:\n%s", got)
	}
}
