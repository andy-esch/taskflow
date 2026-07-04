package store

import (
	"errors"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/andy-esch/taskflow/internal/domain"
)

// SetFields now conflicts on a concurrent IN-PLACE content edit (same path, different
// bytes) — the lost-update the old path-CAS silently allowed, since it only checked
// whether the file had relocated. The losing write must not clobber the concurrent edit.
func TestSetFields_ConflictsOnConcurrentContentEdit(t *testing.T) {
	root := t.TempDir()
	writeTask(t, root, "ready-to-start", "c.md",
		"---\nid: 6fjangd7kvc1\nstatus: ready-to-start\nepic: e1\ntier: 2\npriority: high\neffort: 1h\ncreated: 2026-01-01\ntags: [a]\ndescription: d\n---\n# c\n")
	fs := NewFS(root)
	p := root + "/tasks/ready-to-start/c.md"

	orig := testHookBeforeSetFieldsWrite
	defer func() { testHookBeforeSetFieldsWrite = orig }()
	testHookBeforeSetFieldsWrite = func() {
		// A different writer lands an in-place edit between our validation and our write.
		_ = os.WriteFile(p, []byte("---\nid: 6fjangd7kvc1\nstatus: ready-to-start\nepic: e1\ntier: 2\npriority: high\neffort: 1h\ncreated: 2026-01-01\ntags: [a]\ndescription: CHANGED\n---\n# c\n"), 0o644)
		testHookBeforeSetFieldsWrite = orig // fire once
	}

	if _, err := fs.SetFields("c", map[string]any{"priority": "low"}, false); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("a concurrent in-place edit must conflict (exit 14), got %v", err)
	}
	b, _ := os.ReadFile(p)
	if !strings.Contains(string(b), "description: CHANGED") || strings.Contains(string(b), "priority: low") {
		t.Errorf("the losing write must not clobber the concurrent edit:\n%s", b)
	}
}

// Move conflicts on a concurrent in-place edit of the source during the move, leaving the
// concurrent edit intact and no file at the target (nothing moved).
func TestMove_ConflictsOnConcurrentContentEdit(t *testing.T) {
	root := t.TempDir()
	writeTask(t, root, "ready-to-start", "m.md",
		"---\nid: 6fjangd7kvm1\nstatus: ready-to-start\nepic: e1\n---\n# m\n")
	fs := NewFS(root)
	p := root + "/tasks/ready-to-start/m.md"

	orig := testHookBeforeMoveWrite
	defer func() { testHookBeforeMoveWrite = orig }()
	testHookBeforeMoveWrite = func() {
		_ = os.WriteFile(p, []byte("---\nid: 6fjangd7kvm1\nstatus: ready-to-start\nepic: e1\npriority: high\n---\n# m EDITED\n"), 0o644)
		testHookBeforeMoveWrite = orig
	}

	_, err := fs.Move("m", domain.StatusInProgress, time.Date(2026, 6, 20, 0, 0, 0, 0, time.UTC), false)
	if !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("a concurrent in-place edit during a move must conflict, got %v", err)
	}
	if _, err := os.Stat(root + "/tasks/in-progress/m.md"); !os.IsNotExist(err) {
		t.Error("the losing move must not create the target file")
	}
	if b, _ := os.ReadFile(p); !strings.Contains(string(b), "m EDITED") {
		t.Errorf("the concurrent edit must survive at the source:\n%s", b)
	}
}

// MoveAudit conflicts on a concurrent in-place edit of the source during the move.
func TestMoveAudit_ConflictsOnConcurrentContentEdit(t *testing.T) {
	root := t.TempDir()
	writeAudit(t, root, "open", "2026-01-02-x.md",
		"---\nid: 6fjjt6s9ttx1\nbucket: open\narea: x\ndate: 2026-01-02\n---\n#### H1. t  · **Status:** fixed\n")
	fs := NewFS(root)
	p := root + "/audits/open/2026-01-02-x.md"

	orig := testHookBeforeMoveAuditWrite
	defer func() { testHookBeforeMoveAuditWrite = orig }()
	testHookBeforeMoveAuditWrite = func() {
		_ = os.WriteFile(p, []byte("---\nid: 6fjjt6s9ttx1\nbucket: open\narea: x\ndate: 2026-01-02\nnote: CHANGED\n---\n#### H1. t  · **Status:** fixed\n"), 0o644)
		testHookBeforeMoveAuditWrite = orig
	}
	if _, err := fs.MoveAudit("2026-01-02-x", domain.AuditClosed, false); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("a concurrent in-place edit during an audit move must conflict, got %v", err)
	}
	if _, err := os.Stat(root + "/audits/closed/2026-01-02-x.md"); !os.IsNotExist(err) {
		t.Error("the losing move must not create the target file")
	}
	if b, _ := os.ReadFile(p); !strings.Contains(string(b), "note: CHANGED") {
		t.Errorf("the concurrent edit must survive at the source:\n%s", b)
	}
}

// SetEpicFields now conflicts on a concurrent in-place edit — epics had NO write guard
// before (last-write-wins); version-CAS gives them one.
func TestSetEpicFields_ConflictsOnConcurrentContentEdit(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(root+"/epics", 0o755); err != nil {
		t.Fatal(err)
	}
	p := root + "/epics/99-x.md"
	if err := os.WriteFile(p, []byte("---\nstatus: active\npriority: high\ndescription: e\n---\n# E\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	fs := NewFS(root)

	orig := testHookBeforeEpicWrite
	defer func() { testHookBeforeEpicWrite = orig }()
	testHookBeforeEpicWrite = func() {
		_ = os.WriteFile(p, []byte("---\nstatus: active\npriority: high\ndescription: CHANGED\n---\n# E\n"), 0o644)
		testHookBeforeEpicWrite = orig
	}
	if _, err := fs.SetEpicFields("99-x", map[string]any{"priority": "low"}, false); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("a concurrent in-place epic edit must conflict, got %v", err)
	}
	if b, _ := os.ReadFile(p); !strings.Contains(string(b), "description: CHANGED") || strings.Contains(string(b), "priority: low") {
		t.Errorf("the losing epic write must not clobber the concurrent edit:\n%s", b)
	}
}

// EditBody (task append) conflicts on a concurrent edit during the write window
// (testHookBeforeBodyWrite covers both EditBody and AppendAuditBody).
func TestEditBody_ConflictsOnConcurrentContentEdit(t *testing.T) {
	root := t.TempDir()
	writeTask(t, root, "ready-to-start", "b.md",
		"---\nid: 6fjangd7kvb1\nstatus: ready-to-start\nepic: e1\n---\n# b\n")
	fs := NewFS(root)
	p := root + "/tasks/ready-to-start/b.md"

	orig := testHookBeforeBodyWrite
	defer func() { testHookBeforeBodyWrite = orig }()
	testHookBeforeBodyWrite = func() {
		_ = os.WriteFile(p, []byte("---\nid: 6fjangd7kvb1\nstatus: ready-to-start\nepic: e1\npriority: high\n---\n# CONCURRENT\n"), 0o644)
		testHookBeforeBodyWrite = orig
	}
	if _, _, err := fs.EditBody("b", "appended text", true, time.Date(2026, 6, 20, 0, 0, 0, 0, time.UTC), false); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("a concurrent edit during append must conflict, got %v", err)
	}
	if b, _ := os.ReadFile(p); !strings.Contains(string(b), "CONCURRENT") || strings.Contains(string(b), "appended text") {
		t.Errorf("the losing append must not clobber the concurrent edit:\n%s", b)
	}
}

// EditTask (the editor path via editFile) conflicts when a concurrent writer changes the
// file DURING the editor window — surfaced, never silently rebased onto the other change.
func TestEditTask_ConflictsOnConcurrentContentEdit(t *testing.T) {
	root := t.TempDir()
	writeTask(t, root, "ready-to-start", "e.md",
		"---\nid: 6fjangd7kve1\nstatus: ready-to-start\nepic: e1\n---\n# e\n")
	fs := NewFS(root)
	p := root + "/tasks/ready-to-start/e.md"

	edit := func(current string, prevErr error) (string, error) {
		// A concurrent writer lands a change while the human is "in $EDITOR".
		_ = os.WriteFile(p, []byte("---\nid: 6fjangd7kve1\nstatus: ready-to-start\nepic: e1\npriority: high\n---\n# CONCURRENT\n"), 0o644)
		return current + "\nedited by human\n", nil
	}
	if _, _, err := fs.EditTask("e", time.Date(2026, 6, 20, 0, 0, 0, 0, time.UTC), edit); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("a concurrent edit during the editor window must conflict, got %v", err)
	}
	if b, _ := os.ReadFile(p); !strings.Contains(string(b), "CONCURRENT") {
		t.Errorf("the human's edit must be rejected, the concurrent edit survive:\n%s", b)
	}
}
