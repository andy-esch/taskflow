package store

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/andy-esch/taskflow/internal/domain"
	"github.com/andy-esch/taskflow/internal/testutil"
)

// SetFields now conflicts on a concurrent IN-PLACE content edit (same path, different
// bytes) — the lost-update the old path-CAS silently allowed, since it only checked
// whether the file had relocated. The losing write must not clobber the concurrent edit.
func TestSetFields_ConflictsOnConcurrentContentEdit(t *testing.T) {
	root := t.TempDir()
	writeTask(t, root, "ready-to-start", "c.md",
		"---\nid: 6fjangd7kvc1\nstatus: ready-to-start\nepic: e1\ntier: 2\npriority: high\neffort: 1h\ncreated: 2026-01-01\ntags: [a]\ndescription: d\n---\n# c\n")
	fs := NewFS(root)
	p := filepath.Join(root, "tasks", testutil.TaskID("c")+"-c.md")

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
	p := filepath.Join(root, "tasks", testutil.TaskID("m")+"-m.md")

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
	// The move is an in-place frontmatter edit now — the file never relocates. The losing
	// move must leave the concurrent edit intact and NOT flip status to in-progress.
	b, _ := os.ReadFile(p)
	if !strings.Contains(string(b), "m EDITED") {
		t.Errorf("the concurrent edit must survive at the source:\n%s", b)
	}
	if !strings.Contains(string(b), "status: ready-to-start") || strings.Contains(string(b), "status: in-progress") {
		t.Errorf("the losing move must not change the status:\n%s", b)
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
	p := filepath.Join(root, "tasks", testutil.TaskID("b")+"-b.md")

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
	p := filepath.Join(root, "tasks", testutil.TaskID("e")+"-e.md")

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

// Regression guard: a concurrent creation of an UNRELATED same-prefix file must NOT
// spuriously conflict. verifyUnchanged re-resolves by the canonical slug (exact), not the
// caller's fuzzy query, so a new "billing-*" task can't make "billing" ambiguous and
// reject an edit to a file that never changed. (Independent review finding #1.)
func TestSetFields_FuzzyQueryDoesNotSpuriouslyConflict(t *testing.T) {
	root := t.TempDir()
	writeTask(t, root, "ready-to-start", "billing-system.md",
		"---\nid: 6fjangd7kvf1\nstatus: ready-to-start\nepic: e1\ntier: 2\npriority: high\neffort: 1h\ncreated: 2026-01-01\ntags: [a]\ndescription: d\n---\n# b\n")
	fs := NewFS(root)

	orig := testHookBeforeSetFieldsWrite
	defer func() { testHookBeforeSetFieldsWrite = orig }()
	testHookBeforeSetFieldsWrite = func() {
		// A cron agent creates an unrelated same-prefix task mid-write — this would make the
		// FUZZY query "billing" ambiguous, but the canonical re-resolve is unaffected.
		writeTask(t, root, "ready-to-start", "billing-gateway.md",
			"---\nid: 6fjangd7kvf2\nstatus: ready-to-start\nepic: e1\n---\n# g\n")
		testHookBeforeSetFieldsWrite = orig
	}
	if _, err := fs.SetFields("billing", map[string]any{"priority": "low"}, false); err != nil {
		t.Fatalf("an unrelated concurrent same-prefix creation must NOT conflict; got %v", err)
	}
	if b, _ := os.ReadFile(filepath.Join(root, "tasks", testutil.TaskID("billing-system")+"-billing-system.md")); !strings.Contains(string(b), "priority: low") {
		t.Errorf("the edit should have landed:\n%s", b)
	}
}

// MoveEpic conflict coverage (resolveEpicPath + the move path; distinct from SetEpicFields).
func TestMoveEpic_ConflictsOnConcurrentContentEdit(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(root+"/epics", 0o755); err != nil {
		t.Fatal(err)
	}
	p := root + "/epics/98-y.md"
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
	if _, err := fs.MoveEpic("98-y", "deprecated", time.Date(2026, 6, 20, 0, 0, 0, 0, time.UTC), false); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("a concurrent edit during an epic move must conflict, got %v", err)
	}
	if b, _ := os.ReadFile(p); !strings.Contains(string(b), "description: CHANGED") || strings.Contains(string(b), "status: deprecated") {
		t.Errorf("the losing epic move must not clobber the concurrent edit:\n%s", b)
	}
}

// AppendAuditBody conflict coverage (resolveAuditPath + writeBody; distinct from EditBody).
func TestAppendAuditBody_ConflictsOnConcurrentContentEdit(t *testing.T) {
	root := t.TempDir()
	writeAudit(t, root, "open", "2026-01-02-ab.md",
		"---\nid: 6fjjt6s9ttab\nbucket: open\narea: x\ndate: 2026-01-02\n---\n#### H1. t  · **Status:** open\n")
	fs := NewFS(root)
	p := root + "/audits/open/2026-01-02-ab.md"

	orig := testHookBeforeBodyWrite
	defer func() { testHookBeforeBodyWrite = orig }()
	testHookBeforeBodyWrite = func() {
		_ = os.WriteFile(p, []byte("---\nid: 6fjjt6s9ttab\nbucket: open\narea: x\ndate: 2026-01-02\nnote: CHANGED\n---\n#### H1. t  · **Status:** open\n"), 0o644)
		testHookBeforeBodyWrite = orig
	}
	if _, _, err := fs.AppendAuditBody("2026-01-02-ab", "#### M9. new  · **Status:** open", time.Date(2026, 6, 20, 0, 0, 0, 0, time.UTC), false); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("a concurrent edit during an audit append must conflict, got %v", err)
	}
	if b, _ := os.ReadFile(p); !strings.Contains(string(b), "note: CHANGED") || strings.Contains(string(b), "M9. new") {
		t.Errorf("the losing append must not clobber the concurrent edit:\n%s", b)
	}
}

// A dry-run must not enter the write critical section — it takes neither the lock nor the
// version-CAS (both write-time concerns), consistent with the movers. Pinned by asserting the
// pre-write hook never fires on a dry-run (normalize-dry-run-vs-version-cas-ordering).
func TestSetFields_DryRunSkipsWriteCriticalSection(t *testing.T) {
	root := t.TempDir()
	writeTask(t, root, "ready-to-start", "d.md",
		"---\nid: 6fjangd7kvd1\nstatus: ready-to-start\nepic: e1\ntier: 2\npriority: high\neffort: 1h\ncreated: 2026-01-01\ntags: [a]\ndescription: d\n---\n# d\n")
	fs := NewFS(root)

	fired := false
	orig := testHookBeforeSetFieldsWrite
	defer func() { testHookBeforeSetFieldsWrite = orig }()
	testHookBeforeSetFieldsWrite = func() { fired = true }

	if _, err := fs.SetFields("d", map[string]any{"priority": "low"}, true); err != nil { // dryRun=true
		t.Fatalf("dry-run should validate + preview without error: %v", err)
	}
	if fired {
		t.Error("a dry-run must not enter the write critical section (the pre-write hook fired)")
	}
}

// TestConcurrentAppends_NoLostUpdates is the regression test for the flock write-lock: real
// concurrent goroutines append to the SAME file, and every write that reports success must
// land exactly once. Without the lock the verify→write window (widened by the temp-file
// fsync) lets writers silently clobber each other — landed < success. With it, a race is a
// DETECTED conflict the writer retries, so nothing is lost. (The hook-based tests can't
// exercise this: the hook bypasses the lock; only a genuine race does.)
func TestConcurrentAppends_NoLostUpdates(t *testing.T) {
	root := t.TempDir()
	writeTask(t, root, "ready-to-start", "race.md",
		"---\nid: 6fjangd7kvrc\nstatus: ready-to-start\nepic: e1\n---\n# body\n")
	fs := NewFS(root)

	const N = 8
	var wg sync.WaitGroup
	var succ int64
	for i := 0; i < N; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			// Retry on conflict (mirrors core.Service's auto-retry), with a small backoff so
			// N goroutines don't livelock. The store itself doesn't retry — that's core's job.
			for attempt := 0; attempt < 300; attempt++ {
				if _, _, err := fs.EditBody("race", fmt.Sprintf("marker-%d", i), true, time.Unix(0, 0), false); err == nil {
					atomic.AddInt64(&succ, 1)
					return
				} else if !errors.Is(err, domain.ErrConflict) {
					t.Errorf("writer %d: unexpected error: %v", i, err)
					return
				}
				time.Sleep(time.Millisecond)
			}
			t.Errorf("writer %d exhausted retries (livelock?)", i)
		}(i)
	}
	wg.Wait()

	b, err := os.ReadFile(filepath.Join(root, "tasks", testutil.TaskID("race")+"-race.md"))
	if err != nil {
		t.Fatal(err)
	}
	landed := int64(strings.Count(string(b), "marker-"))
	if landed != succ {
		t.Errorf("lost/duplicated updates under concurrency: %d markers landed but %d writes reported success", landed, succ)
	}
	if succ != N {
		t.Errorf("want all %d concurrent writers to land, got %d", N, succ)
	}
}
