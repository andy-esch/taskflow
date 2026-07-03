package store

import (
	"os"
	"strings"
	"testing"
	"time"

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

func TestFixFrontmatter_MovesMisfiledTask(t *testing.T) {
	root := t.TempDir()
	// frontmatter says ready-to-start but the file sits in completed/ → misfiled.
	writeTask(t, root, "completed", "drifted.md", "---\nid: 6fjangd7kvab\nstatus: ready-to-start\nepic: e1\n---\n# x\n")
	// a foreign/legacy status word isn't a valid status → the folder governs, no move.
	writeTask(t, root, "completed", "legacy.md", "---\nid: 6fjangd7kvac\nstatus: superseded\n---\n# x\n")

	res, err := NewFS(root).FixFrontmatter(false)
	if err != nil {
		t.Fatal(err)
	}

	// The misfiled file was RELOCATED to ready-to-start/, its frontmatter untouched.
	if _, err := os.Stat(root + "/tasks/completed/drifted.md"); !os.IsNotExist(err) {
		t.Error("drifted file should have been moved out of completed/")
	}
	b, err := os.ReadFile(root + "/tasks/ready-to-start/drifted.md")
	if err != nil {
		t.Fatalf("drifted file not found in ready-to-start/: %v", err)
	}
	if !strings.Contains(string(b), "status: ready-to-start") {
		t.Errorf("the move must not rewrite the frontmatter status:\n%s", b)
	}
	// The foreign-status file stays put — the folder governs as a fallback.
	if _, err := os.Stat(root + "/tasks/completed/legacy.md"); err != nil {
		t.Errorf("foreign-status file should stay in completed/: %v", err)
	}
	// The move is reported (keyed by the old path); the legacy file is not.
	var movedReported, legacyReported bool
	for _, r := range res {
		if strings.Contains(r.Path, "drifted") {
			movedReported = true
		}
		if strings.Contains(r.Path, "legacy") {
			legacyReported = true
		}
	}
	if !movedReported {
		t.Errorf("drifted move not reported: %+v", res)
	}
	if legacyReported {
		t.Errorf("legacy file should not be in fix results: %+v", res)
	}
}

func TestFixFrontmatter_MisfiledMove_DryRunPreviewsOnly(t *testing.T) {
	root := t.TempDir()
	writeTask(t, root, "completed", "drifted.md", "---\nid: 6fjangd7kvad\nstatus: ready-to-start\nepic: e1\n---\n# x\n")

	res, err := NewFS(root).FixFrontmatter(true) // dry-run
	if err != nil {
		t.Fatal(err)
	}
	if len(res) != 1 || !strings.Contains(res[0].Path, "drifted") {
		t.Fatalf("dry-run should preview the move, got %+v", res)
	}
	// Nothing actually moved.
	if _, err := os.Stat(root + "/tasks/completed/drifted.md"); err != nil {
		t.Errorf("dry-run must not move the file: %v", err)
	}
	if _, err := os.Stat(root + "/tasks/ready-to-start/drifted.md"); !os.IsNotExist(err) {
		t.Error("dry-run must not create the target file")
	}
}

func TestFixFrontmatter_MisfiledMove_SkipsCollision(t *testing.T) {
	root := t.TempDir()
	// dup claims ready-to-start, but that dir already holds a DIFFERENT file with the
	// same slug → moving would clobber it, so the fix leaves the misfiled one put.
	writeTask(t, root, "completed", "dup.md", "---\nid: 6fjangd7kvae\nstatus: ready-to-start\n---\n# a\n")
	writeTask(t, root, "ready-to-start", "dup.md", "---\nid: 6fjangd7kvaf\nstatus: ready-to-start\nepic: e1\n---\n# b\n")

	if _, err := NewFS(root).FixFrontmatter(false); err != nil {
		t.Fatal(err)
	}
	// Both files remain (no clobber).
	for _, p := range []string{"/tasks/completed/dup.md", "/tasks/ready-to-start/dup.md"} {
		if _, err := os.Stat(root + p); err != nil {
			t.Errorf("both dup files should remain (no clobber), missing %s: %v", p, err)
		}
	}
	// The occupant is intact (still body "b").
	b, _ := os.ReadFile(root + "/tasks/ready-to-start/dup.md")
	if !strings.Contains(string(b), "# b") {
		t.Errorf("the target occupant must not be clobbered:\n%s", b)
	}
}

// Moving a misfiled file to the status it already authoritatively has does NOT change
// its status (no re-stamp — `from` is the frontmatter value), but DOES relocate the
// file to match, so a verb never leaves a file it touched stranded in the wrong dir.
func TestMove_RelocatesMisfiledWithoutRestamp(t *testing.T) {
	root := t.TempDir()
	// frontmatter says in-progress, but the file sits in ready-to-start/ (misfiled).
	writeTask(t, root, "ready-to-start", "m.md", "---\nid: 6fjangd7kvag\nstatus: in-progress\nepic: e1\n---\n# x\n")

	task, err := NewFS(root).Move("m", domain.StatusInProgress, time.Date(2026, 6, 20, 0, 0, 0, 0, time.UTC), false)
	if err != nil {
		t.Fatal(err)
	}
	// Relocated to in-progress/ (no longer misfiled)...
	if _, err := os.Stat(root + "/tasks/ready-to-start/m.md"); !os.IsNotExist(err) {
		t.Error("the misfiled file should have been relocated out of ready-to-start/")
	}
	b, err := os.ReadFile(root + "/tasks/in-progress/m.md")
	if err != nil {
		t.Fatalf("file should be relocated to in-progress/: %v", err)
	}
	if task.Misfiled() {
		t.Error("the relocated task should no longer report misfiled")
	}
	// ...but started_at was NOT stamped: `from` (frontmatter) already equalled `to`, so
	// there was no real transition to re-date.
	if strings.Contains(string(b), "started_at") {
		t.Errorf("moving to the already-current status must not re-stamp started_at:\n%s", b)
	}
}
