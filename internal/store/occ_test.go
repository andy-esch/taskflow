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
