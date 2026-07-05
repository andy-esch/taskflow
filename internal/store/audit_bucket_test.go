package store

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/andy-esch/taskflow/internal/domain"
	"github.com/andy-esch/taskflow/internal/testutil"
)

// MoveAudit rewrites the authoritative bucket frontmatter in place — under the flat
// layout (ADR-0003 §4) the file path never changes; only the `bucket:` field moves.
func TestMoveAudit_WritesBucketFrontmatter(t *testing.T) {
	root := t.TempDir()
	path, out := testutil.AuditFixture(root, "open", "2026-01-02-c.md",
		"---\nid: 6fjjt6s9ttz3\nbucket: open\narea: x\ndate: 2026-01-02\n---\n#### H1. t  · **Status:** fixed\n")
	testutil.Write(t, path, out)

	a, err := NewFS(root).MoveAudit("2026-01-02-c", domain.AuditClosed, false)
	if err != nil {
		t.Fatal(err)
	}
	if a.Bucket != domain.AuditClosed {
		t.Errorf("moved audit bucket = %q, want closed", a.Bucket)
	}
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("file should stay at its original flat path: %v", err)
	}
	if !strings.Contains(string(b), "bucket: closed") {
		t.Errorf("MoveAudit must rewrite the bucket frontmatter:\n%s", b)
	}
}

// audit reopen on an audit whose bucket frontmatter is MISSING (it fell back and was
// flagged) HEALS the frontmatter — writing bucket: open even though the bucket isn't
// changing, clearing the flag lint raised. The file stays at its flat path (no
// relocation under the flat layout).
func TestMoveAudit_HealsFellBackBucket(t *testing.T) {
	root := t.TempDir()
	// No bucket frontmatter → falls back and is flagged (BucketFellBack). Written to a
	// raw flat id-led path so the missing bucket is NOT injected by the fixture helper.
	path := filepath.Join(root, "audits", testutil.TaskID("2026-01-02-h")+"-2026-01-02-h.md")
	testutil.Write(t, path,
		"---\nid: 6fjjt6s9ttz6\narea: x\ndate: 2026-01-02\n---\n# a\n")

	a, err := NewFS(root).MoveAudit("2026-01-02-h", domain.AuditOpen, false)
	if err != nil {
		t.Fatal(err)
	}
	if a.BucketFellBack {
		t.Error("after healing, the reloaded audit should no longer report a fell-back bucket")
	}
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(b), "bucket: open") {
		t.Errorf("reopen must write the missing bucket frontmatter:\n%s", b)
	}
}
