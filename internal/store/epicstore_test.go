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
