package store

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/andy-esch/taskflow/internal/domain"
)

func writeEpic(t *testing.T, root, name, content string) {
	t.Helper()
	dir := filepath.Join(root, "epics")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

// TestFS_ListEpics_NumericOrder pins that epics come back ordered by their NN-
// number, not lexically — so 100 sorts after 99 (and 10 after 9), which a string
// compare of the zero-padded id gets wrong.
func TestFS_ListEpics_NumericOrder(t *testing.T) {
	root := t.TempDir()
	for _, id := range []string{"09-i", "10-j", "100-k", "02-b"} {
		writeEpic(t, root, id+".md", "---\nstatus: planning\n---\n# "+id+"\n")
	}
	epics, _, err := NewFS(root).ListEpics()
	if err != nil {
		t.Fatal(err)
	}
	var got []string
	for _, e := range epics {
		got = append(got, e.ID)
	}
	want := []string{"02-b", "09-i", "10-j", "100-k"}
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Errorf("epics out of numeric order:\n got %v\nwant %v", got, want)
	}
}

// TestFS_WatchPaths covers the directory set exposed for the TUI watcher: it must
// be derived from the domain enums (one source of layout truth), so every status
// and bucket subdir is present under the right parent.
func TestFS_WatchPaths(t *testing.T) {
	root := filepath.Join("x", "plan")
	got := map[string]bool{}
	for _, d := range NewFS(root).WatchPaths() {
		got[d] = true
	}
	for _, parent := range []string{"epics", "tasks", "audits"} {
		if !got[filepath.Join(root, parent)] {
			t.Errorf("WatchPaths missing entity parent %q", parent)
		}
	}
	for _, st := range domain.AllStatuses() {
		if !got[filepath.Join(root, "tasks", st.Dir())] {
			t.Errorf("WatchPaths missing status dir %q", st.Dir())
		}
	}
	for _, b := range domain.AllAuditBuckets() {
		if !got[filepath.Join(root, "audits", b.Dir())] {
			t.Errorf("WatchPaths missing audit bucket %q", b.Dir())
		}
	}
}

func TestFS_ListEpics_And_GetEpic(t *testing.T) {
	root := t.TempDir()
	writeEpic(t, root, "17-x.md", "---\nstatus: in-progress\ndescription: x epic\ntags: [a]\n---\n# Epic X\nbody\n")

	fs := NewFS(root)
	epics, _, err := fs.ListEpics()
	if err != nil {
		t.Fatal(err)
	}
	if len(epics) != 1 || epics[0].ID != "17-x" || epics[0].Status != "in-progress" || epics[0].Description != "x epic" {
		t.Fatalf("bad epics: %+v", epics)
	}

	ep, body, err := fs.GetEpic("17-x")
	if err != nil {
		t.Fatal(err)
	}
	if ep.ID != "17-x" || !strings.Contains(body, "# Epic X") {
		t.Errorf("bad GetEpic: %+v body=%q", ep, body)
	}

	if _, _, err := fs.GetEpic("nope"); !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("want ErrNotFound, got %v", err)
	}
}

func TestFS_ListEpics_NoDir(t *testing.T) {
	epics, _, err := NewFS(t.TempDir()).ListEpics()
	if err != nil {
		t.Fatalf("missing epics dir should not error: %v", err)
	}
	if len(epics) != 0 {
		t.Errorf("want 0 epics, got %d", len(epics))
	}
}
