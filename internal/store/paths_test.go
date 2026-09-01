package store

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/andy-esch/taskflow/internal/domain"
	"github.com/andy-esch/taskflow/internal/testutil"
)

func TestResolveThreadPathRemainsParseFreeForMalformedDocuments(t *testing.T) {
	root := t.TempDir()
	threadID := testutil.TaskID("malformed-thread")
	path := filepath.Join(root, domain.ThreadsDir, threadID+"-malformed-thread.md")
	testutil.Write(t, path, "---\nid: [unterminated\n---\n# Repair me\n")
	fs := NewFS(root)

	for _, ref := range []string{threadID, "malformed-thread"} {
		got, err := fs.ResolveThreadPath(ref)
		if err != nil || got != path {
			t.Fatalf("ResolveThreadPath(%q) = %q, %v; want %q", ref, got, err, path)
		}
	}
	if _, _, err := fs.GetThread(threadID); err == nil {
		t.Fatal("GetThread parsed malformed frontmatter; repair-path test is not exercising the failure case")
	}
}

func TestResolveThreadPathPreservesFilenameIdentityAndAmbiguityRules(t *testing.T) {
	root := t.TempDir()
	firstID := testutil.TaskID("repair-alpha")
	secondID := testutil.TaskID("repair-beta")
	frontmatterID := testutil.TaskID("drifted-frontmatter")
	first := filepath.Join(root, domain.ThreadsDir, firstID+"-repair-alpha.md")
	second := filepath.Join(root, domain.ThreadsDir, secondID+"-repair-beta.md")
	testutil.Write(t, first, "---\nid: "+frontmatterID+"\nstatus: unstarted\n---\n# Drifted\n")
	testutil.Write(t, second, "not valid Thread frontmatter\n")
	fs := NewFS(root)

	got, err := fs.ResolveThreadPath(firstID)
	if err != nil || got != first {
		t.Fatalf("filename id path = %q, %v; want %q", got, err, first)
	}
	if _, err := fs.ResolveThreadPath(frontmatterID); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("frontmatter-only id resolved despite parse-free filename identity: %v", err)
	}
	if _, err := fs.ResolveThreadPath("repair"); !errors.Is(err, domain.ErrAmbiguous) {
		t.Fatalf("ambiguous Thread slug error = %v", err)
	}
}

func TestResolveThreadPathPreservesSymlinkedPlanningRoot(t *testing.T) {
	physical := t.TempDir()
	threadID := testutil.TaskID("symlink-thread")
	physicalPath := filepath.Join(physical, domain.ThreadsDir, threadID+"-symlink-thread.md")
	testutil.Write(t, physicalPath, "malformed but path-resolvable\n")
	parent := t.TempDir()
	linkedRoot := filepath.Join(parent, "planning-link")
	if err := os.Symlink(physical, linkedRoot); err != nil {
		t.Fatal(err)
	}

	got, err := NewFS(linkedRoot).ResolveThreadPath("symlink-thread")
	want := filepath.Join(linkedRoot, domain.ThreadsDir, threadID+"-symlink-thread.md")
	if err != nil || got != want {
		t.Fatalf("symlink path = %q, %v; want %q", got, err, want)
	}
}
