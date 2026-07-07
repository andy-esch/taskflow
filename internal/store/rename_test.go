package store

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/andy-esch/taskflow/internal/domain"
)

// renameRepo seeds a scratch tree: an epic, task A (id-old.md) with an H1, and task B that
// links to A via a relative-path markdown link (display text == A's slug).
func renameRepo(t *testing.T) (root, aPath, bPath string) {
	t.Helper()
	root = t.TempDir()
	write := func(rel, content string) {
		p := filepath.Join(root, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write("epics/01-e.md", "---\nstatus: active\npriority: high\ndescription: e\n---\n# E\n")
	aPath = filepath.Join(root, "tasks", "6fjangd7kva1-old.md")
	write("tasks/6fjangd7kva1-old.md", "---\nid: 6fjangd7kva1\nstatus: ready-to-start\nepic: 01-e\ntier: 2\npriority: high\neffort: 1h\ncreated: 2026-01-01\ntags: [a]\n---\n# Old Title\n\nbody.\n")
	bPath = filepath.Join(root, "tasks", "6fjangd7kvb2-b.md")
	write("tasks/6fjangd7kvb2-b.md", "---\nid: 6fjangd7kvb2\nstatus: ready-to-start\nepic: 01-e\ntier: 2\npriority: high\neffort: 1h\ncreated: 2026-01-01\ntags: [a]\n---\n# B\n\nSee [old](6fjangd7kva1-old.md#h) for context.\n")
	return root, aPath, bPath
}

func TestRenameTask_RenamesAndCascades(t *testing.T) {
	root, aPath, bPath := renameRepo(t)
	newPath := filepath.Join(root, "tasks", "6fjangd7kva1-shiny-new-title.md")

	task, cascade, err := NewFS(root).RenameTask("old", "Shiny New Title", false)
	if err != nil {
		t.Fatal(err)
	}
	if task.Slug != "shiny-new-title" || task.ID != "6fjangd7kva1" {
		t.Errorf("renamed task = %q id=%q, want shiny-new-title / same id", task.Slug, task.ID)
	}
	if cascade != 1 {
		t.Errorf("cascade count = %d, want 1 (task B's inbound link)", cascade)
	}
	// The file is renamed (id kept), the old name is gone.
	if _, err := os.Stat(aPath); !os.IsNotExist(err) {
		t.Error("old file should be removed")
	}
	got, _ := os.ReadFile(newPath)
	if !contains(got, "# Shiny New Title") {
		t.Errorf("body H1 not re-titled:\n%s", got)
	}
	// Task B's inbound link is repointed — filename, display text, AND the #anchor preserved.
	b, _ := os.ReadFile(bPath)
	if !contains(b, "[shiny-new-title](6fjangd7kva1-shiny-new-title.md#h)") {
		t.Errorf("inbound link not cascaded (filename/display/anchor):\n%s", b)
	}
}

// TestRenameTask_CrossDirSameNameLeftAlone: a link is repointed only if it RESOLVES to the
// renamed file — a same-basename file in a different directory (and links to it) are left
// untouched (the cross-dir-collision fix).
func TestRenameTask_CrossDirSameNameLeftAlone(t *testing.T) {
	root, _, _ := renameRepo(t)
	write := func(rel, content string) {
		p := filepath.Join(root, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	// A DIFFERENT file sharing task A's basename, and a task C linking to THAT one.
	write("research/6fjangd7kva1-old.md", "# a research doc\n")
	cPath := "tasks/6fjangd7kvc3-c.md"
	write(cPath, "---\nid: 6fjangd7kvc3\nstatus: ready-to-start\nepic: 01-e\ntier: 2\npriority: high\neffort: 1h\ncreated: 2026-01-01\ntags: [a]\n---\n# C\n\nElsewhere: [old](../research/6fjangd7kva1-old.md).\n")

	_, cascade, err := NewFS(root).RenameTask("old", "New Title", false)
	if err != nil {
		t.Fatal(err)
	}
	// Only task B's link (to tasks/…old.md) cascades; task C's link to research/…old.md doesn't.
	if cascade != 1 {
		t.Errorf("only a link to the ACTUAL renamed file should cascade, got %d", cascade)
	}
	c, _ := os.ReadFile(filepath.Join(root, filepath.FromSlash(cPath)))
	if !contains(c, "../research/6fjangd7kva1-old.md") {
		t.Errorf("a same-named file in another dir must be left untouched:\n%s", c)
	}
	if _, err := os.Stat(filepath.Join(root, "research", "6fjangd7kva1-old.md")); err != nil {
		t.Error("the cross-dir same-name file must not be renamed/removed")
	}
}

func TestRenameTask_DryRunTouchesNothing(t *testing.T) {
	root, aPath, bPath := renameRepo(t)
	aBefore, _ := os.ReadFile(aPath)
	bBefore, _ := os.ReadFile(bPath)

	task, cascade, err := NewFS(root).RenameTask("old", "New", true)
	if err != nil {
		t.Fatal(err)
	}
	if task.Slug != "new" || cascade != 1 {
		t.Errorf("dry run should still report the would-be result: slug=%q cascade=%d", task.Slug, cascade)
	}
	if _, err := os.Stat(aPath); err != nil {
		t.Error("dry run must not remove the old file")
	}
	if a, _ := os.ReadFile(aPath); !equal(a, aBefore) {
		t.Error("dry run modified the target file")
	}
	if b, _ := os.ReadFile(bPath); !equal(b, bBefore) {
		t.Error("dry run modified an inbound-link file")
	}
}

func TestRenameTask_EmptyTitleRejected(t *testing.T) {
	root, _, _ := renameRepo(t)
	if _, _, err := NewFS(root).RenameTask("old", "…", false); !errors.Is(err, domain.ErrValidation) {
		t.Errorf("a title that slugifies to empty must be ErrValidation, got %v", err)
	}
}

// TestRenameTask_TargetCollisionRefused: renaming onto a filename that already exists must
// fail loud (ErrConflict) rather than silently clobber it — the write loop would otherwise
// overwrite the pre-existing same-id file. Nothing on disk changes.
func TestRenameTask_TargetCollisionRefused(t *testing.T) {
	root, aPath, _ := renameRepo(t)
	// A stray file occupying task A's would-be target name (same id, different slug).
	takenPath := filepath.Join(root, "tasks", "6fjangd7kva1-taken.md")
	if err := os.WriteFile(takenPath, []byte("# do not clobber\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	aBefore, _ := os.ReadFile(aPath)

	if _, _, err := NewFS(root).RenameTask("old", "Taken", false); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("rename onto an existing target must be ErrConflict, got %v", err)
	}
	// The source is untouched and the pre-existing target is intact.
	if a, _ := os.ReadFile(aPath); !equal(a, aBefore) {
		t.Error("source file must be unchanged on a refused rename")
	}
	if got, _ := os.ReadFile(takenPath); !contains(got, "do not clobber") {
		t.Error("pre-existing target file must not be overwritten")
	}
}

// TestRenameTask_RefStyleAndFencedExamples: the cascade repoints a reference-style link to
// the renamed file but leaves an inline example link inside a fenced code block untouched.
func TestRenameTask_RefStyleAndFencedExamples(t *testing.T) {
	root := t.TempDir()
	write := func(rel, content string) {
		p := filepath.Join(root, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write("epics/01-e.md", "---\nstatus: active\npriority: high\ndescription: e\n---\n# E\n")
	write("tasks/6fjangd7kva1-old.md", "---\nid: 6fjangd7kva1\nstatus: ready-to-start\nepic: 01-e\ntier: 2\npriority: high\neffort: 1h\ncreated: 2026-01-01\ntags: [a]\n---\n# Old Title\n\nbody.\n")
	bPath := filepath.Join(root, "tasks", "6fjangd7kvb2-b.md")
	write("tasks/6fjangd7kvb2-b.md", "---\nid: 6fjangd7kvb2\nstatus: ready-to-start\nepic: 01-e\ntier: 2\npriority: high\neffort: 1h\ncreated: 2026-01-01\ntags: [a]\n---\n# B\n\n"+
		"Ref: see the [old task][a].\n\n[a]: 6fjangd7kva1-old.md\n\nExample:\n\n```\n[old](6fjangd7kva1-old.md)\n```\n")

	_, cascade, err := NewFS(root).RenameTask("old", "Renamed", false)
	if err != nil {
		t.Fatal(err)
	}
	if cascade != 1 {
		t.Errorf("only the reference-style link should cascade (fenced example skipped), got %d", cascade)
	}
	b, _ := os.ReadFile(bPath)
	if !contains(b, "[a]: 6fjangd7kva1-renamed.md") {
		t.Errorf("reference-style link target not repointed:\n%s", b)
	}
	if !contains(b, "[old](6fjangd7kva1-old.md)") {
		t.Errorf("an inline example inside a code fence must be left untouched:\n%s", b)
	}
}

func contains(b []byte, sub string) bool { return strings.Contains(string(b), sub) }
func equal(a, b []byte) bool             { return bytes.Equal(a, b) }
