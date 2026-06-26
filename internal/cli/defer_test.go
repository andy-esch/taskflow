package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// readTaskFile returns the raw contents of a task file at tasks/<status>/<name>.
func readTaskFile(t *testing.T, root, status, name string) string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(root, "tasks", status, name))
	if err != nil {
		t.Fatalf("read %s/%s: %v", status, name, err)
	}
	return string(b)
}

// TestTaskDefer_Until pins the snooze path: `task defer <slug> --until <date>`
// moves the task to deferred/ AND records revisit_at; a bare defer (no --until)
// is unchanged (no revisit_at written).
func TestTaskDefer_Until(t *testing.T) {
	root := setupRepo(t)
	runRoot(t, "-C", root, "task", "defer", "alpha", "--until", "2026-09-01")

	// Moved out of ready-to-start, into deferred/.
	notExist(t, filepath.Join(root, "tasks", "ready-to-start", "alpha.md"))
	got := readTaskFile(t, root, "deferred", "alpha.md")
	if !strings.Contains(got, `revisit_at: "2026-09-01"`) {
		t.Errorf("deferred task should carry revisit_at:\n%s", got)
	}
	if !strings.Contains(got, "status: deferred") {
		t.Errorf("deferred task frontmatter status should be deferred:\n%s", got)
	}
}

// TestTaskDefer_NoUntilUnchanged pins the indefinite defer: without --until the
// task moves to deferred/ with no revisit_at field (exactly as before).
func TestTaskDefer_NoUntilUnchanged(t *testing.T) {
	root := setupRepo(t)
	runRoot(t, "-C", root, "task", "defer", "alpha")
	got := readTaskFile(t, root, "deferred", "alpha.md")
	if strings.Contains(got, "revisit_at") {
		t.Errorf("a bare defer must not write revisit_at:\n%s", got)
	}
}

// TestTaskDefer_BadDateExit11 pins the up-front validation: a malformed --until
// is a loud validation error (exit 11) BEFORE anything moves — the task stays put.
func TestTaskDefer_BadDateExit11(t *testing.T) {
	root := setupRepo(t)
	_, err := runRootRC(t, "-C", root, "task", "defer", "alpha", "--until", "next-week")
	if err == nil {
		t.Fatal("a bad --until date should fail")
	}
	if ExitCode(err) != 11 {
		t.Errorf("want exit 11 for a bad date, got %d (%v)", ExitCode(err), err)
	}
	// Nothing moved: alpha is still in ready-to-start, not deferred.
	if _, statErr := os.Stat(filepath.Join(root, "tasks", "ready-to-start", "alpha.md")); statErr != nil {
		t.Errorf("a bad date must not move the task: %v", statErr)
	}
	notExist(t, filepath.Join(root, "tasks", "deferred", "alpha.md"))
}

// TestTaskDefer_UntilDryRun pins the preview: --dry-run previews the snooze
// without writing — the file stays in its original bucket.
func TestTaskDefer_UntilDryRun(t *testing.T) {
	root := setupRepo(t)
	out, err := runRootRC(t, "-C", root, "task", "defer", "alpha", "--until", "2026-09-01", "--dry-run")
	if err != nil {
		t.Fatalf("dry-run defer failed: %v\n%s", err, out)
	}
	if !strings.Contains(out, "would move") {
		t.Errorf("dry-run should preview the move, got:\n%s", out)
	}
	// Nothing written: still in ready-to-start, not deferred.
	if _, statErr := os.Stat(filepath.Join(root, "tasks", "ready-to-start", "alpha.md")); statErr != nil {
		t.Errorf("dry-run must not move the task: %v", statErr)
	}
	notExist(t, filepath.Join(root, "tasks", "deferred", "alpha.md"))
}

// TestTaskSet_RevisitAt pins that revisit_at rides the generic `task set` path for
// free: --set writes the field (date-validated) and --unset clears it.
func TestTaskSet_RevisitAt(t *testing.T) {
	root := setupRepo(t)
	runRoot(t, "-C", root, "task", "set", "alpha", "--set", "revisit_at=2026-09-01")
	got := readTaskFile(t, root, "ready-to-start", "alpha.md")
	if !strings.Contains(got, `revisit_at: "2026-09-01"`) {
		t.Errorf("task set --set revisit_at should write the field:\n%s", got)
	}

	// A bad date is rejected (the field routes through ValidateDate).
	if _, err := runRootRC(t, "-C", root, "task", "set", "alpha", "--set", "revisit_at=someday"); err == nil {
		t.Error("task set revisit_at=<bad date> should fail validation")
	} else if ExitCode(err) != 11 {
		t.Errorf("want exit 11 for a bad revisit_at, got %d", ExitCode(err))
	}

	// --unset clears it.
	runRoot(t, "-C", root, "task", "set", "alpha", "--unset", "revisit_at")
	got = readTaskFile(t, root, "ready-to-start", "alpha.md")
	if strings.Contains(got, "revisit_at") {
		t.Errorf("task set --unset revisit_at should remove the field:\n%s", got)
	}
}
