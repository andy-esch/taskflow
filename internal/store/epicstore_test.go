package store

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/andy-esch/taskflow/internal/domain"
	"github.com/andy-esch/taskflow/internal/testutil"
)

func writeEpic(t *testing.T, root, name, content string) {
	t.Helper()
	testutil.Write(t, filepath.Join(root, domain.EpicsDir, name), content)
}

// TestFS_ListEpics_MissingFrontmatterIsLoud: a fence-less epic file is surfaced
// as a loud FileProblem naming the valid shape, not silently read as an empty
// epic (which would only trip a vaguer "invalid status" downstream).
func TestFS_ListEpics_MissingFrontmatterIsLoud(t *testing.T) {
	root := t.TempDir()
	writeEpic(t, root, "01-x.md", "# Just a heading\n\nno frontmatter here\n")

	epics, problems, err := NewFS(root).ListEpics()
	if err != nil {
		t.Fatal(err)
	}
	if len(epics) != 0 {
		t.Errorf("a fence-less epic must not parse as an epic, got %+v", epics)
	}
	if len(problems) != 1 || !strings.Contains(problems[0].Message, "missing frontmatter") || !strings.Contains(problems[0].Message, "schema epic") {
		t.Errorf("want one loud, shape-naming problem, got %+v", problems)
	}
}

// TestFS_ListEpics_NumericOrder pins that epics come back ordered by their NN-
// number, not lexically — so 100 sorts after 99 (and 10 after 9), which a string
// compare of the zero-padded id gets wrong.
func TestFS_ListEpics_NumericOrder(t *testing.T) {
	root := t.TempDir()
	for _, id := range []string{"09-i", "10-j", "100-k", "02-b"} {
		writeEpic(t, root, id+".md", "---\nstatus: active\n---\n# "+id+"\n")
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

// TestFS_WatchPaths covers the directory set exposed for the TUI watcher. Tasks are
// flat now (ADR-0003 §4), so the tasks dir itself is the only task watch path — no
// per-status subdirs; audits are still bucketed, so their bucket subdirs remain.
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
		if got[filepath.Join(root, "tasks", st.Dir())] {
			t.Errorf("WatchPaths should NOT include a per-status task dir %q under the flat layout", st.Dir())
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
	writeEpic(t, root, "17-x.md", "---\nstatus: active\ndescription: x epic\ntags: [a]\n---\n# Epic X\nbody\n")

	fs := NewFS(root)
	epics, _, err := fs.ListEpics()
	if err != nil {
		t.Fatal(err)
	}
	if len(epics) != 1 || epics[0].ID != "17-x" || epics[0].Status != "active" || epics[0].Description != "x epic" {
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

// TestFS_MoveEpic pins the surgical status-field rewrite: only `status` changes,
// the file stays put (status is a FIELD, not a directory), and unknown fields,
// key order, and the body all survive.
func TestFS_MoveEpic(t *testing.T) {
	root := t.TempDir()
	writeEpic(t, root, "18-tui.md",
		"---\nstatus: active\ndescription: tui epic\ncustom: keep\n---\n# TUI Epic\nbody\n")

	ep, err := NewFS(root).MoveEpic("18-tui", "retired", bodyNow, false)
	if err != nil {
		t.Fatal(err)
	}
	if ep.ID != "18-tui" || ep.Status != "retired" {
		t.Errorf("returned epic wrong: %+v", ep)
	}

	b, err := os.ReadFile(filepath.Join(root, "epics", "18-tui.md"))
	if err != nil {
		t.Fatal(err)
	}
	s := string(b)
	if !strings.Contains(s, "status: retired") {
		t.Errorf("status not rewritten:\n%s", s)
	}
	if strings.Contains(s, "status: active") {
		t.Errorf("old status lingered:\n%s", s)
	}
	if !strings.Contains(s, "custom: keep") || !strings.Contains(s, "description: tui epic") {
		t.Errorf("unknown/other fields lost (surgical write failed):\n%s", s)
	}
	if !strings.Contains(s, "# TUI Epic\nbody\n") {
		t.Errorf("body not preserved:\n%s", s)
	}
	// A real status change stamps updated_at (the seed had none) to now (2026-06-20).
	if ep.Updated != "2026-06-20" || !strings.Contains(s, `updated_at: "2026-06-20"`) {
		t.Errorf("a status move should stamp updated_at; ep.Updated=%q\n%s", ep.Updated, s)
	}
}

// TestFS_MoveEpic_NoOp_NoStamp: moving to the CURRENT status isn't an edit — no
// updated_at stamp (mirrors the "no bump on no-op" rule).
func TestFS_MoveEpic_NoOp_NoStamp(t *testing.T) {
	root := t.TempDir()
	writeEpic(t, root, "18-tui.md", "---\nstatus: active\ndescription: x\n---\n# X\n")
	ep, err := NewFS(root).MoveEpic("18-tui", "active", bodyNow, false)
	if err != nil {
		t.Fatal(err)
	}
	s, _ := os.ReadFile(filepath.Join(root, "epics", "18-tui.md"))
	if ep.Updated != "" || strings.Contains(string(s), "updated_at") {
		t.Errorf("a no-op move (active→active) must not stamp updated_at:\n%s", s)
	}
}

// TestFS_MoveEpic_InvalidStatus rejects an out-of-vocabulary status with the file
// untouched.
func TestFS_MoveEpic_InvalidStatus(t *testing.T) {
	root := t.TempDir()
	const original = "---\nstatus: active\ndescription: x\n---\n# X\n"
	writeEpic(t, root, "18-tui.md", original)

	if _, err := NewFS(root).MoveEpic("18-tui", "bogus", bodyNow, false); !errors.Is(err, domain.ErrValidation) {
		t.Errorf("invalid status should be ErrValidation, got %v", err)
	}
	b, _ := os.ReadFile(filepath.Join(root, "epics", "18-tui.md"))
	if string(b) != original {
		t.Errorf("file must be untouched on a rejected move:\n%s", b)
	}
}

// TestFS_MoveEpic_DryRun validates end-to-end but skips the write.
func TestFS_MoveEpic_DryRun(t *testing.T) {
	root := t.TempDir()
	const original = "---\nstatus: active\n---\n# X\n"
	writeEpic(t, root, "18-tui.md", original)

	ep, err := NewFS(root).MoveEpic("18-tui", "retired", bodyNow, true)
	if err != nil {
		t.Fatal(err)
	}
	if ep.Status != "retired" {
		t.Errorf("dry-run should report the would-be status, got %q", ep.Status)
	}
	b, _ := os.ReadFile(filepath.Join(root, "epics", "18-tui.md"))
	if string(b) != original {
		t.Errorf("dry-run must not write:\n%s", b)
	}
}

func TestFS_MoveEpic_NotFound(t *testing.T) {
	if _, err := NewFS(t.TempDir()).MoveEpic("ghost", "retired", bodyNow, false); !errors.Is(err, domain.ErrNotFound) {
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
