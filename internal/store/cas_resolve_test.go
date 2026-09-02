package store

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/andy-esch/taskflow/internal/domain"
)

// writeDuplicateIDTasks lays down two task files that both claim ONE stable id,
// with distinct slugs so the caller's own query still resolves uniquely and the
// write actually reaches the version-CAS guard.
func writeDuplicateIDTasks(t *testing.T, dup string) string {
	t.Helper()
	root := t.TempDir()
	tasks := filepath.Join(root, "tasks")
	if err := os.MkdirAll(tasks, 0o755); err != nil {
		t.Fatal(err)
	}
	front := "---\nid: " + dup + "\nstatus: ready-to-start\nepic: e1\ndescription: d\ntags: [a]\n---\n"
	for name, body := range map[string]string{dup + "-alpha.md": "# A\n", dup + "-beta.md": "# B\n"} {
		if err := os.WriteFile(filepath.Join(tasks, name), []byte(front+body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return root
}

// TestVerifyUnchanged_DuplicateIDSurfacesAsAmbiguous is the counterpart to
// TestVerifyUnchanged_SlugEqualsAnotherID: when two files genuinely claim one
// stable id, the version-CAS re-resolve must SAY so rather than reporting a
// conflict. A conflict invites a retry, and no retry can make two files into one
// — the caller would spend the whole OCC budget to be told "changed on disk;
// retry", which is both false and unactionable. The ambiguity names the colliding
// files, which is the only thing that makes the state fixable.
func TestVerifyUnchanged_DuplicateIDSurfacesAsAmbiguous(t *testing.T) {
	const dup = "6fjangd7kvc1"
	root := writeDuplicateIDTasks(t, dup)

	// "alpha" resolves uniquely by slug, so the write starts; the guard's exact-id
	// re-resolve is where the duplicate is discovered.
	_, err := NewFS(root).SetFields("alpha", map[string]any{"priority": "low"}, false)
	if err == nil {
		t.Fatal("a duplicate stable id must not be written through")
	}
	if errors.Is(err, domain.ErrConflict) {
		t.Errorf("a duplicate id is not a retryable conflict: %v", err)
	}
	if !errors.Is(err, domain.ErrAmbiguous) {
		t.Fatalf("want ErrAmbiguous (exit 13), got %v", err)
	}
	for _, want := range []string{dup, "alpha", "beta"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("the error should name the colliding files (missing %q): %v", want, err)
		}
	}
}

// TestVerifyUnchanged_BodyWriteAlsoReportsTheAmbiguity exercises a second write
// path, so the split is pinned in the shared guard rather than in one caller.
func TestVerifyUnchanged_BodyWriteAlsoReportsTheAmbiguity(t *testing.T) {
	root := writeDuplicateIDTasks(t, "6fjangd7kvc2")

	_, _, err := NewFS(root).EditBody("alpha", "## Notes\n", true, time.Now(), false)
	if errors.Is(err, domain.ErrConflict) || !errors.Is(err, domain.ErrAmbiguous) {
		t.Fatalf("a body write should report the duplicate id as an ambiguity, got %v", err)
	}
}

// TestVerifyUnchanged_VanishedFileIsStillAConflict pins the other half of the
// split: a file no longer resolvable under its own id really did change under us,
// so it stays a conflict — and stays retryable, since a fresh read can succeed.
func TestVerifyUnchanged_VanishedFileIsStillAConflict(t *testing.T) {
	root := t.TempDir()
	tasks := filepath.Join(root, "tasks")
	if err := os.MkdirAll(tasks, 0o755); err != nil {
		t.Fatal(err)
	}
	const id = "6fjangd7kvc3"
	path := filepath.Join(tasks, id+"-gone.md")
	content := "---\nid: " + id + "\nstatus: ready-to-start\nepic: e1\ndescription: d\ntags: [a]\n---\n# G\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	// A different writer removes the file between our validation and our write.
	orig := testHookBeforeSetFieldsWrite
	t.Cleanup(func() { testHookBeforeSetFieldsWrite = orig })
	testHookBeforeSetFieldsWrite = func() {
		_ = os.Remove(path)
		testHookBeforeSetFieldsWrite = orig // fire once
	}

	_, err := NewFS(root).SetFields("gone", map[string]any{"priority": "low"}, false)
	if !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("a file that vanished under us is a genuine conflict, got %v", err)
	}
}
