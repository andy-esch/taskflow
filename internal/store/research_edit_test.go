package store

import (
	"errors"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/andy-esch/taskflow/internal/domain"
	"github.com/andy-esch/taskflow/internal/testutil"
)

// EditResearch shipped with NO test at all — the editor loop, its parse-before-accept, the
// reopen-on-error path, the no-change branch, and the success path were all uncovered.
// Mirrors audit_edit_test.go.

var researchEditNow = time.Date(2026, 8, 18, 0, 0, 0, 0, time.UTC)

func researchEditSeed(t *testing.T) string {
	t.Helper()
	return "---\nschema: 1\nid: " + testutil.TaskID("doc") + "\ncreated: \"2026-01-03\"\nstatus: reference\n---\n# Doc\n\nbody\n"
}

func researchEditRepo(t *testing.T) (*FS, string) {
	t.Helper()
	root := t.TempDir()
	path := researchFixture(t, root, "doc.md", researchEditSeed(t))
	return NewFS(root), path
}

func TestEditResearch_ValidEdit_Writes(t *testing.T) {
	fs, path := researchEditRepo(t)
	want := strings.Replace(researchEditSeed(t), "body", "edited body", 1)

	r, changed, err := fs.EditResearch("doc", researchEditNow, func(cur string, prevErr error) (string, error) {
		if prevErr != nil {
			t.Errorf("unexpected prior error: %v", prevErr)
		}
		if !strings.Contains(cur, "# Doc") {
			t.Errorf("editor got unexpected content: %q", cur)
		}
		return want, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if !changed {
		t.Error("a content change must report changed=true")
	}
	got, _ := os.ReadFile(path)
	if !strings.Contains(string(got), "edited body") {
		t.Errorf("edit not persisted:\n%s", got)
	}
	// A changed save is stamped; created stays immutable, and the unknown key survives.
	if r.Updated != "2026-08-18" {
		t.Errorf("updated_at = %q, want the stamp", r.Updated)
	}
	if r.Created != "2026-01-03" {
		t.Errorf("created must not change: %q", r.Created)
	}
	if !strings.Contains(string(got), "status: reference") {
		t.Errorf("unknown key lost:\n%s", got)
	}
}

// No change => no write, no stamp.
func TestEditResearch_NoChange_ReportsUnchanged(t *testing.T) {
	fs, path := researchEditRepo(t)
	before, _ := os.ReadFile(path)

	_, changed, err := fs.EditResearch("doc", researchEditNow, func(cur string, _ error) (string, error) {
		return cur, nil // saved without editing
	})
	if err != nil {
		t.Fatal(err)
	}
	if changed {
		t.Error("an unchanged save must report changed=false")
	}
	after, _ := os.ReadFile(path)
	if string(before) != string(after) {
		t.Errorf("an unchanged save must not rewrite the file (no updated_at stamp):\n%s", after)
	}
}

// Parse-before-accept: a broken save must NOT land on disk; the editor is reopened with
// the error, and giving up is ErrValidation.
func TestEditResearch_BrokenSaveIsRefused(t *testing.T) {
	fs, path := researchEditRepo(t)
	before, _ := os.ReadFile(path)
	broken := "---\nid: [unclosed\n  bad: : indent\n---\n# x\n"

	_, changed, err := fs.EditResearch("doc", researchEditNow, func(string, error) (string, error) {
		return broken, nil // keep returning the broken content -> caller gives up
	})
	if !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("a broken edit must be ErrValidation, got %v", err)
	}
	if changed {
		t.Error("a refused edit must not report changed")
	}
	after, _ := os.ReadFile(path)
	if string(before) != string(after) {
		t.Errorf("a broken edit must never land on disk:\n%s", after)
	}
}

// The reopen path: the editor is handed the previous error, and a fixed second attempt is
// accepted.
func TestEditResearch_BrokenThenFixed_Reopens(t *testing.T) {
	fs, path := researchEditRepo(t)
	broken := "---\nid: [unclosed\n  bad: : indent\n---\n# x\n"
	fixed := strings.Replace(researchEditSeed(t), "body", "recovered", 1)

	attempts := 0
	_, changed, err := fs.EditResearch("doc", researchEditNow, func(_ string, prevErr error) (string, error) {
		attempts++
		if attempts == 1 {
			if prevErr != nil {
				t.Errorf("first attempt should have no prior error, got %v", prevErr)
			}
			return broken, nil
		}
		if prevErr == nil {
			t.Error("the reopened editor must be told why the previous save was refused")
		}
		return fixed, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if !changed || attempts != 2 {
		t.Errorf("want a reopen then success, got changed=%v attempts=%d", changed, attempts)
	}
	got, _ := os.ReadFile(path)
	if !strings.Contains(string(got), "recovered") {
		t.Errorf("the fixed edit should have landed:\n%s", got)
	}
}

func TestEditResearch_UnknownSlug_NotFound(t *testing.T) {
	fs, _ := researchEditRepo(t)
	ran := false
	_, _, err := fs.EditResearch("nope", researchEditNow, func(string, error) (string, error) {
		ran = true
		return "", nil
	})
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("want ErrNotFound, got %v", err)
	}
	if ran {
		t.Error("the editor must not open for an unresolvable slug")
	}
}
