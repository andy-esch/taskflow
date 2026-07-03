package store

import (
	"os"
	"strings"
	"testing"

	"github.com/andy-esch/taskflow/internal/domain"
)

func TestParseTask_FrontmatterIsAuthoritative(t *testing.T) {
	root := t.TempDir()
	writeTask(t, root, "completed", "drifted.md", "---\nstatus: ready-to-start\n---\n# x\n") // misfiled
	writeTask(t, root, "completed", "legacy.md", "---\nstatus: superseded\n---\n# x\n")      // foreign vocab
	writeTask(t, root, "completed", "clean.md", "---\nstatus: completed\n---\n# x\n")        // ok

	tasks, _, err := NewFS(root).ListTasks()
	if err != nil {
		t.Fatal(err)
	}
	by := map[string]domain.Task{}
	for _, tk := range tasks {
		by[tk.Slug] = tk
	}

	// Frontmatter wins: the file in completed/ declaring ready-to-start reads as
	// ready-to-start, with the folder captured as the (stale) mirror → misfiled.
	if m := by["drifted"]; m.Status != domain.StatusReadyToStart || m.FolderStatus != domain.StatusCompleted || !m.Misfiled() {
		t.Errorf("drifted: status=%q folder=%q misfiled=%v (want ready-to-start/completed/true)",
			m.Status, m.FolderStatus, m.Misfiled())
	}
	// A foreign/legacy word isn't a valid status, so the folder governs as a fallback
	// and the file is not misfiled.
	if l := by["legacy"]; l.Status != domain.StatusCompleted || l.Misfiled() {
		t.Errorf("legacy foreign vocab should fall back to the folder: status=%q misfiled=%v", l.Status, l.Misfiled())
	}
	if c := by["clean"]; c.Misfiled() {
		t.Errorf("clean task wrongly flagged misfiled")
	}
}

func TestFixFrontmatter_RealignsMisfiledStatus(t *testing.T) {
	root := t.TempDir()
	writeTask(t, root, "completed", "drifted.md", "---\nstatus: ready-to-start\nepic: e1\n---\n# x\n")
	writeTask(t, root, "completed", "legacy.md", "---\nstatus: superseded\n---\n# x\n")

	res, err := NewFS(root).FixFrontmatter(false)
	if err != nil {
		t.Fatal(err)
	}

	read := func(p string) string {
		b, err := os.ReadFile(p)
		if err != nil {
			t.Fatal(err)
		}
		return string(b)
	}
	if s := read(root + "/tasks/completed/drifted.md"); !strings.Contains(s, "status: completed") {
		t.Errorf("drifted status not realigned to folder:\n%s", s)
	}
	if s := read(root + "/tasks/completed/legacy.md"); !strings.Contains(s, "status: superseded") {
		t.Errorf("foreign status word should be left untouched:\n%s", s)
	}
	var fixedDrifted bool
	for _, r := range res {
		if strings.Contains(r.Path, "drifted") {
			fixedDrifted = true
		}
		if strings.Contains(r.Path, "legacy") {
			t.Errorf("legacy file should not be in fix results: %+v", r)
		}
	}
	if !fixedDrifted {
		t.Errorf("drifted file not reported as fixed: %+v", res)
	}
}
