package store

import (
	"strings"
	"testing"
	"time"

	"github.com/andy-esch/taskflow/internal/testutil"
)

var bodyNow = time.Date(2026, 6, 20, 12, 0, 0, 0, time.UTC)

// Replace swaps the whole body, keeps the frontmatter, and stamps updated_at.
func TestEditBody_Replace(t *testing.T) {
	fs, path := editRepo(t)
	task, _, err := fs.EditBody("edit-me", "# New Title\n\nfresh content", false, bodyNow, false)
	if err != nil {
		t.Fatalf("EditBody: %v", err)
	}
	got := readFile(t, path)
	if !strings.Contains(got, "# New Title") || !strings.Contains(got, "fresh content") {
		t.Errorf("new body missing:\n%s", got)
	}
	if strings.Contains(got, "# Title") || strings.Contains(got, "\nbody\n") {
		t.Errorf("old body should be gone:\n%s", got)
	}
	if !strings.Contains(got, "description: original") || !strings.Contains(got, "tier: 2") {
		t.Errorf("frontmatter must be preserved:\n%s", got)
	}
	if !strings.Contains(got, "updated_at:") || !strings.Contains(got, "2026-06-20") {
		t.Errorf("updated_at should be stamped:\n%s", got)
	}
	if task.Updated != "2026-06-20" {
		t.Errorf("returned task.Updated = %q, want 2026-06-20", task.Updated)
	}
}

// Append keeps the existing body and adds the section after one blank line.
func TestEditBody_Append(t *testing.T) {
	fs, path := editRepo(t)
	if _, _, err := fs.EditBody("edit-me", "## Review\n- looks good", true, bodyNow, false); err != nil {
		t.Fatalf("EditBody append: %v", err)
	}
	got := readFile(t, path)
	if !strings.Contains(got, "# Title") || !strings.Contains(got, "## Review") {
		t.Errorf("append should keep old and add new:\n%s", got)
	}
	if !strings.Contains(got, "body\n\n## Review") {
		t.Errorf("expected a single blank-line separator before the appended section:\n%s", got)
	}
	if !strings.HasSuffix(got, "- looks good\n") {
		t.Errorf("appended body should end with a single newline:\n%q", got)
	}
}

// A body edit must not disturb unknown frontmatter keys (surgical preservation).
func TestEditBody_PreservesUnknownKeys(t *testing.T) {
	root := t.TempDir()
	seed := "---\nstatus: ready-to-start\ncustom_field: keep-me\ndescription: d\n---\n# B\n\nold\n"
	writeTask(t, root, "ready-to-start", "u.md", seed)
	path, _ := testutil.TaskFixture(root, "ready-to-start", "u.md", seed)
	fs := NewFS(root)
	if _, _, err := fs.EditBody("u", "# B\n\nnew", false, bodyNow, false); err != nil {
		t.Fatalf("EditBody: %v", err)
	}
	got := readFile(t, path)
	if !strings.Contains(got, "custom_field: keep-me") {
		t.Errorf("unknown key must survive a body edit:\n%s", got)
	}
}

// Dry-run returns the would-be task but writes nothing.
func TestEditBody_DryRun_NoWrite(t *testing.T) {
	fs, path := editRepo(t)
	before := readFile(t, path)
	task, _, err := fs.EditBody("edit-me", "# replaced", false, bodyNow, true)
	if err != nil {
		t.Fatalf("EditBody dryRun: %v", err)
	}
	if task.Updated != "2026-06-20" {
		t.Errorf("dry-run should still return the would-be task, Updated=%q", task.Updated)
	}
	if readFile(t, path) != before {
		t.Error("dry-run must not write")
	}
}

// A CRLF file must round-trip through both replace and append with a single,
// consistent line ending — no mixed CRLF-frontmatter / LF-body diff (the invariant
// pinned for SetFields by TestFS_SetFields_CRLFRoundTrip).
func TestEditBody_CRLFRoundTrip(t *testing.T) {
	for _, tc := range []struct {
		name string
		edit func(*FS) error
	}{
		{"replace", func(fs *FS) error { _, _, e := fs.EditBody("alpha", "# New\n\nfresh", false, bodyNow, false); return e }},
		{"append", func(fs *FS) error { _, _, e := fs.EditBody("alpha", "## More\n- x", true, bodyNow, false); return e }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			root := t.TempDir()
			crlf := strings.ReplaceAll("---\nstatus: ready-to-start\ndescription: old\n---\n# Alpha\n\nbody\n", "\n", "\r\n")
			writeTask(t, root, "ready-to-start", "alpha.md", crlf)
			path, _ := testutil.TaskFixture(root, "ready-to-start", "alpha.md", crlf)
			fs := NewFS(root)
			if err := tc.edit(fs); err != nil {
				t.Fatal(err)
			}
			b := readFile(t, path)
			if lone := strings.Count(b, "\n") - strings.Count(b, "\r\n"); lone != 0 {
				t.Errorf("CRLF file came back with %d bare-LF endings (mixed):\n%q", lone, b)
			}
			if _, _, err := fs.GetTask("alpha"); err != nil {
				t.Errorf("edited CRLF file should still load: %v", err)
			}
		})
	}
}

// appendSection's seam handling: blank-line separator, empty body, empty addition,
// and a CRLF body folded to LF (replaceBodyStamped restores the file's ending).
func TestAppendSection_Edges(t *testing.T) {
	cases := []struct{ old, add, want string }{
		{"# B\n\nbody\n", "## New\n- x", "# B\n\nbody\n\n## New\n- x\n"},
		{"", "## New", "## New\n"},
		{"# B\n", "", "# B\n"},
		{"# B\r\n\r\nbody\r\n", "## New", "# B\n\nbody\n\n## New\n"},
	}
	for i, c := range cases {
		if got := appendSection(c.old, c.add); got != c.want {
			t.Errorf("case %d: appendSection(%q, %q) = %q, want %q", i, c.old, c.add, got, c.want)
		}
	}
}

// A file whose frontmatter won't parse can't be body-edited (it would corrupt the
// source of truth) — the error surfaces, nothing is written.
func TestEditBody_BrokenFrontmatter_Errors(t *testing.T) {
	root := t.TempDir()
	writeTask(t, root, "ready-to-start", "bad.md", "---\nstatus: ready-to-start\nno closing fence\n")
	fs := NewFS(root)
	if _, _, err := fs.EditBody("bad", "x", false, bodyNow, false); err == nil {
		t.Fatal("editing a file with unterminated frontmatter should error")
	}
}

func TestEditBody_EchoesOnDiskBody(t *testing.T) {
	root := t.TempDir()
	crlf := strings.ReplaceAll("---\nstatus: ready-to-start\ndescription: d\n---\n# T\n\nbody\n", "\n", "\r\n")
	writeTask(t, root, "ready-to-start", "alpha.md", crlf)
	_, gotBody, err := NewFS(root).EditBody("alpha", "## New", true, bodyNow, false)
	if err != nil {
		t.Fatal(err)
	}
	// The echoed body matches the file's CRLF ending (so it equals what task show
	// returns), not the LF intermediate.
	if lone := strings.Count(gotBody, "\n") - strings.Count(gotBody, "\r\n"); lone != 0 || !strings.Contains(gotBody, "\r\n") {
		t.Errorf("echoed body should use the file's CRLF ending (no lone LF), got %q", gotBody)
	}
}
