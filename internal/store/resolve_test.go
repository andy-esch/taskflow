package store

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/andy-esch/taskflow/internal/domain"
)

// fuzzyRepo seeds slugs chosen to exercise every resolution tier.
func fuzzyRepo(t *testing.T) *FS {
	t.Helper()
	root := t.TempDir()
	writeTask(t, root, "ready-to-start", "add-retry-backoff.md",
		"---\nstatus: ready-to-start\ndescription: x\n---\n# t\n")
	writeTask(t, root, "in-progress", "add-retry-jitter.md",
		"---\nstatus: in-progress\ndescription: x\n---\n# t\n")
	writeTask(t, root, "in-progress", "polish.md",
		"---\nstatus: in-progress\ndescription: x\n---\n# t\n")
	// "polish" is also a strict prefix of this one — exact must still win.
	writeTask(t, root, "completed", "polish-batch.md",
		"---\nstatus: completed\ndescription: x\n---\n# t\n")
	return NewFS(root)
}

func TestResolve_FuzzyTiers(t *testing.T) {
	fs := fuzzyRepo(t)
	get := func(q string) (domain.Task, error) {
		task, _, err := fs.GetTask(q)
		return task, err
	}

	// Exact always wins, even when it prefixes another slug.
	if task, err := get("polish"); err != nil || task.Slug != "polish" {
		t.Errorf("exact should win over prefix: %v %q", err, task.Slug)
	}
	// Unique prefix resolves.
	if task, err := get("polish-b"); err != nil || task.Slug != "polish-batch" {
		t.Errorf("unique prefix should resolve: %v %q", err, task.Slug)
	}
	// Unique substring resolves when no prefix matches.
	if task, err := get("jitter"); err != nil || task.Slug != "add-retry-jitter" {
		t.Errorf("unique substring should resolve: %v %q", err, task.Slug)
	}
	// Case-insensitive.
	if task, err := get("JITTER"); err != nil || task.Slug != "add-retry-jitter" {
		t.Errorf("matching should be case-insensitive: %v %q", err, task.Slug)
	}
	// Ambiguous prefix → ErrAmbiguous listing the candidates (sorted, located).
	_, err := get("add-retry")
	if !errors.Is(err, domain.ErrAmbiguous) {
		t.Fatalf("ambiguous prefix should be ErrAmbiguous, got %v", err)
	}
	msg := err.Error()
	if !strings.Contains(msg, "add-retry-backoff (ready-to-start)") ||
		!strings.Contains(msg, "add-retry-jitter (in-progress)") {
		t.Errorf("the error should list located candidates, got %q", msg)
	}
	if strings.Index(msg, "add-retry-backoff") > strings.Index(msg, "add-retry-jitter") {
		t.Errorf("candidates should be sorted, got %q", msg)
	}
	// No match at any tier.
	if _, err := get("zzz"); !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("no match should be ErrNotFound, got %v", err)
	}
	// Path separators and dot-dot are rejected, not joined into paths.
	for _, q := range []string{"a/b", `a\b`, "../escape", ".."} {
		if _, err := get(q); !errors.Is(err, domain.ErrValidation) {
			t.Errorf("query %q should be ErrValidation, got %v", q, err)
		}
	}
}

// TestMove_FuzzyKeepsCanonicalSlug pins the rename trap: moving by an
// abbreviation must keep the file's full slug, not rename it to the query.
func TestMove_FuzzyKeepsCanonicalSlug(t *testing.T) {
	fs := fuzzyRepo(t)
	task, err := fs.Move("backoff", domain.StatusInProgress, time.Now(), false)
	if err != nil {
		t.Fatal(err)
	}
	if task.Slug != "add-retry-backoff" {
		t.Errorf("the returned task should carry the canonical slug, got %q", task.Slug)
	}
	if _, statErr := os.Stat(filepath.Join(fs.tasksDir, "in-progress", "add-retry-backoff.md")); statErr != nil {
		t.Error("the moved file must keep its canonical filename")
	}
	if _, statErr := os.Stat(filepath.Join(fs.tasksDir, "in-progress", "backoff.md")); statErr == nil {
		t.Error("the file must NOT be renamed to the abbreviation")
	}
}

func TestResolveAuditAndEpic_Fuzzy(t *testing.T) {
	root := t.TempDir()
	write := func(rel, content string) {
		p := filepath.Join(root, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write("audits/open/2026-06-01-store-review.md", "---\narea: store\n---\n# A\n")
	write("epics/17-pm-go-cli.md", "---\nstatus: planning\ndescription: e\n---\n# E\n")
	write("epics/18-tui-browser.md", "---\nstatus: planning\ndescription: e\n---\n# E\n")
	fs := NewFS(root)

	if a, _, err := fs.GetAudit("store-review"); err != nil || a.Slug != "2026-06-01-store-review" {
		t.Errorf("audit substring should resolve: %v %q", err, a.Slug)
	}
	if e, _, err := fs.GetEpic("tui"); err != nil || e.ID != "18-tui-browser" {
		t.Errorf("epic substring should resolve: %v %q", err, e.ID)
	}
	if _, _, err := fs.GetEpic("1"); !errors.Is(err, domain.ErrAmbiguous) {
		t.Errorf("ambiguous epic prefix should be ErrAmbiguous, got %v", err)
	}
}

// TestParseAudit_StatusCaseInsensitive pins the (?i) fix: a hand-edited
// "**Status:** Open" counts as open (the guard still blocks "opened").
func TestParseAudit_StatusCaseInsensitive(t *testing.T) {
	body := "---\narea: x\n---\n# A\n\n#### H1. Finding\n**Status:** Open\n\n#### H2. Other\n**Status:** opened-question\n"
	a, err := parseAudit([]byte(body), "audits/open/a.md", domain.AuditOpen)
	if err != nil {
		t.Fatal(err)
	}
	if a.Findings != 2 || a.OpenFindings != 1 {
		t.Errorf("want 2 findings / 1 open (capital Open counts, 'opened-' doesn't), got %d/%d", a.Findings, a.OpenFindings)
	}
}
