package store

import (
	"errors"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/andy-esch/taskflow/internal/core"
	"github.com/andy-esch/taskflow/internal/domain"
	"github.com/andy-esch/taskflow/internal/testutil"
)

const transformAuditSource = "---\nid: 6fjangd7kva1\nbucket: open\narea: a\ndate: \"2026-01-01\"\n---\n" +
	"# Audit: a\n\n## Findings\n\n#### H1. a finding · **Status:** open\n\nBody.\n"

func transformAuditRepo(t *testing.T) (*FS, string) {
	t.Helper()
	root := t.TempDir()
	p, content := testutil.AuditFixture(root, "open", "2026-01-01-a.md", transformAuditSource)
	testutil.Write(t, p, content)
	return NewFS(root), p
}

var transformAuditNow = time.Date(2026, 6, 20, 0, 0, 0, 0, time.UTC)

// An error from the transform callback surfaces unchanged and writes nothing —
// the validation-before-write contract the finding stamp depends on.
func TestTransformAuditBody_CallbackErrorWritesNothing(t *testing.T) {
	fs, p := transformAuditRepo(t)
	before, _ := os.ReadFile(p)
	sentinel := errors.New("no such finding")

	_, _, changed, err := fs.TransformAuditBody("2026-01-01-a", transformAuditNow, false,
		func(string) (string, error) { return "", sentinel })
	if !errors.Is(err, sentinel) {
		t.Fatalf("callback error should surface unchanged, got %v", err)
	}
	if changed {
		t.Error("a failed transform must not report a change")
	}
	if after, _ := os.ReadFile(p); string(after) != string(before) {
		t.Errorf("a failed transform must leave the file byte-identical:\n%s", after)
	}
}

// An already-applied edit is a no-op: changed=false and updated_at is NOT stamped.
// Without this, re-running `audit finding --status open` would churn the file date.
func TestTransformAuditBody_UnchangedBodyIsNoOpAndDoesNotStamp(t *testing.T) {
	fs, p := transformAuditRepo(t)
	before, _ := os.ReadFile(p)

	_, _, changed, err := fs.TransformAuditBody("2026-01-01-a", transformAuditNow, false,
		func(current string) (string, error) { return current, nil })
	if err != nil {
		t.Fatal(err)
	}
	if changed {
		t.Error("an identical body must report changed=false")
	}
	after, _ := os.ReadFile(p)
	if string(after) != string(before) {
		t.Errorf("a no-op must not rewrite the file (updated_at must not be stamped):\n%s", after)
	}
	if strings.Contains(string(after), "updated_at") {
		t.Errorf("a no-op stamped updated_at:\n%s", after)
	}
}

// A dry run computes and validates without writing or locking.
func TestTransformAuditBody_DryRunWritesNothing(t *testing.T) {
	fs, p := transformAuditRepo(t)
	before, _ := os.ReadFile(p)

	_, _, changed, err := fs.TransformAuditBody("2026-01-01-a", transformAuditNow, true,
		func(current string) (string, error) { return current + "\nappended by dry run\n", nil })
	if err != nil {
		t.Fatal(err)
	}
	if !changed {
		t.Error("a dry run over a real change should still report changed=true")
	}
	if after, _ := os.ReadFile(p); string(after) != string(before) {
		t.Errorf("a dry run must not write:\n%s", after)
	}
}

// A concurrent write during the read→write window is rejected by the content CAS,
// and the concurrent change survives intact.
func TestTransformAuditBody_ConflictsOnConcurrentContentEdit(t *testing.T) {
	fs, p := transformAuditRepo(t)
	concurrent := strings.Replace(transformAuditSource, "Body.", "CONCURRENT.", 1)

	orig := testHookBeforeBodyWrite
	defer func() { testHookBeforeBodyWrite = orig }()
	testHookBeforeBodyWrite = func() {
		_ = os.WriteFile(p, []byte(concurrent), 0o644)
		testHookBeforeBodyWrite = orig
	}

	_, _, _, err := fs.TransformAuditBody("2026-01-01-a", transformAuditNow, false,
		func(current string) (string, error) {
			return strings.Replace(current, "**Status:** open", "**Status:** fixed", 1), nil
		})
	if !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("a concurrent edit during the write window must conflict, got %v", err)
	}
	after, _ := os.ReadFile(p)
	if !strings.Contains(string(after), "CONCURRENT.") || strings.Contains(string(after), "**Status:** fixed") {
		t.Errorf("the losing transform must not clobber the concurrent edit:\n%s", after)
	}
}

// The transform sees the CURRENT body on every invocation, which is what lets core
// retry a finding stamp around a concurrent append instead of replaying stale text.
func TestTransformAuditBody_TransformSeesFreshBodyAfterConcurrentAppend(t *testing.T) {
	fs, p := transformAuditRepo(t)
	appended := strings.Replace(transformAuditSource, "Body.", "Body.\n\n## Appended section\n\nnew prose.", 1)
	if err := os.WriteFile(p, []byte(appended), 0o644); err != nil {
		t.Fatal(err)
	}

	var seen string
	_, _, changed, err := fs.TransformAuditBody("2026-01-01-a", transformAuditNow, false,
		func(current string) (string, error) {
			seen = current
			return strings.Replace(current, "**Status:** open", "**Status:** fixed", 1), nil
		})
	if err != nil {
		t.Fatal(err)
	}
	if !changed {
		t.Fatal("the status stamp should have changed the body")
	}
	if !strings.Contains(seen, "Appended section") {
		t.Error("transform must receive the freshly-read body, not a stale snapshot")
	}
	after, _ := os.ReadFile(p)
	if !strings.Contains(string(after), "Appended section") {
		t.Errorf("the concurrent append must survive the finding stamp:\n%s", after)
	}
	if !strings.Contains(string(after), "**Status:** fixed") {
		t.Errorf("the finding stamp must land:\n%s", after)
	}
}

// AC: a race between `audit append` and `audit finding` preserves BOTH changes.
// The append lands inside the finding write's CAS window; the finding stamp loses,
// retries through core, recomputes against the appended body, and both survive.
func TestEditFinding_RetriesAroundConcurrentAppendPreservingBoth(t *testing.T) {
	fs, p := transformAuditRepo(t)
	svc := core.NewService(fs, core.WithRetry(4, func(int) {}))

	orig := testHookBeforeBodyWrite
	defer func() { testHookBeforeBodyWrite = orig }()
	testHookBeforeBodyWrite = func() {
		// Restore first: the append below goes through the same write path.
		testHookBeforeBodyWrite = orig
		if _, _, err := fs.AppendAuditBody("2026-01-01-a", "## Appended section\n\nnew prose.", transformAuditNow, false); err != nil {
			t.Errorf("concurrent append failed: %v", err)
		}
	}

	_, changed, err := svc.EditFinding("2026-01-01-a", "H1", core.FindingEdit{Status: "fixed"}, false)
	if err != nil {
		t.Fatalf("the finding stamp should retry around a concurrent append, got %v", err)
	}
	if !changed {
		t.Error("the retried stamp should report a change")
	}
	after, _ := os.ReadFile(p)
	if !strings.Contains(string(after), "Appended section") {
		t.Errorf("the concurrent append must survive:\n%s", after)
	}
	if !strings.Contains(string(after), "**Status:** fixed") {
		t.Errorf("the finding stamp must survive:\n%s", after)
	}
}

// The frontmatter is preserved surgically and the body write stamps updated_at.
func TestTransformAuditBody_StampsUpdatedAtAndPreservesFrontmatter(t *testing.T) {
	fs, p := transformAuditRepo(t)
	_, _, changed, err := fs.TransformAuditBody("2026-01-01-a", transformAuditNow, false,
		func(current string) (string, error) {
			return strings.Replace(current, "**Status:** open", "**Status:** fixed", 1), nil
		})
	if err != nil || !changed {
		t.Fatalf("expected a change, got changed=%v err=%v", changed, err)
	}
	after, _ := os.ReadFile(p)
	for _, want := range []string{"id: 6fjangd7kva1", "bucket: open", "area: a", `date: "2026-01-01"`, `updated_at: "2026-06-20"`, "**Status:** fixed"} {
		if !strings.Contains(string(after), want) {
			t.Errorf("expected %q in:\n%s", want, after)
		}
	}
}
