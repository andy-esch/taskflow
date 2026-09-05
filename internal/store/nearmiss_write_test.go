package store

import (
	"errors"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/andy-esch/taskflow/internal/domain"
	"github.com/andy-esch/taskflow/internal/testutil"
)

const cleanAuditSource = "---\nid: 6fjangd7kvc1\nbucket: open\narea: c\ndate: \"2026-01-01\"\n---\n" +
	"# Audit: c\n\n## Findings\n\n#### H1. a finding · **Status:** open\n\nBody.\n"

const driftedAuditSource = "---\nid: 6fjangd7kvd1\nbucket: open\narea: d\ndate: \"2026-01-01\"\n---\n" +
	"# Audit: d\n\n## Findings\n\n#### H-1. pre-existing drift\n\n**Status:** open\n"

var writeNow = time.Date(2026, 6, 20, 0, 0, 0, 0, time.UTC)

func auditRepo(t *testing.T, name, source string) (*FS, string) {
	t.Helper()
	root := t.TempDir()
	p, content := testutil.AuditFixture(root, "open", name, source)
	testutil.Write(t, p, content)
	return NewFS(root), p
}

// `audit append --help` promises the finding grammar is checked. Until now it was
// deferred to `audit lint`, which only validated findings that already parsed — so a
// drifted header fell between the two and vanished. It is now refused at the write.
func TestAppendAuditBody_RefusesIntroducedNearMissHeader(t *testing.T) {
	fs, p := auditRepo(t, "2026-01-01-c.md", cleanAuditSource)
	before, _ := os.ReadFile(p)

	_, _, err := fs.AppendAuditBody("2026-01-01-c", "#### H-2. a drifted header\n\n**Status:** open", writeNow, false)
	if !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("append must refuse a heading that would parse to nothing, got %v", err)
	}
	if !strings.Contains(err.Error(), `write "#### H2. a drifted header"`) {
		t.Errorf("the refusal must name the canonical replacement, got: %v", err)
	}
	if after, _ := os.ReadFile(p); string(after) != string(before) {
		t.Errorf("a refused append must not write:\n%s", after)
	}
}

// A canonical append still lands — the guard must not make appending findings harder.
func TestAppendAuditBody_AcceptsCanonicalHeader(t *testing.T) {
	fs, p := auditRepo(t, "2026-01-01-c.md", cleanAuditSource)
	if _, _, err := fs.AppendAuditBody("2026-01-01-c", "#### M1. a good header · **Status:** open", writeNow, false); err != nil {
		t.Fatalf("a canonical header must be accepted: %v", err)
	}
	b, _ := os.ReadFile(p)
	if !strings.Contains(string(b), "#### M1. a good header") {
		t.Errorf("the append did not land:\n%s", b)
	}
}

// Pre-existing drift must not freeze the document: only what THIS write introduces
// is refused, or one bad header would block every later append and status stamp.
func TestAppendAuditBody_AllowsUnrelatedAppendToAlreadyDriftedAudit(t *testing.T) {
	fs, p := auditRepo(t, "2026-01-01-d.md", driftedAuditSource)
	if _, _, err := fs.AppendAuditBody("2026-01-01-d", "## Candidate tasks\n\n- none", writeNow, false); err != nil {
		t.Fatalf("an unrelated append to a drifted audit must be allowed: %v", err)
	}
	b, _ := os.ReadFile(p)
	if !strings.Contains(string(b), "Candidate tasks") {
		t.Errorf("the append did not land:\n%s", b)
	}
	if !strings.Contains(string(b), "#### H-1. pre-existing drift") {
		t.Errorf("the pre-existing drift must be left for `lint --fix`, not silently rewritten:\n%s", b)
	}
}

// The repair path must stay permissive — it exists precisely to write drifted files.
func TestTransformAuditBody_RepairPathIsNotBlockedByTheWriteGuard(t *testing.T) {
	fs, p := auditRepo(t, "2026-01-01-d.md", driftedAuditSource)
	_, _, changed, err := fs.TransformAuditBody("2026-01-01-d", writeNow, false, func(current string) (string, error) {
		fixed, _ := domain.CanonicalizeFindingHeaders(current)
		return fixed, nil
	})
	if err != nil {
		t.Fatalf("the repair path must not be blocked by the write guard: %v", err)
	}
	if !changed {
		t.Fatal("the repair should have changed the body")
	}
	b, _ := os.ReadFile(p)
	if !strings.Contains(string(b), "#### H1. pre-existing drift") {
		t.Errorf("the repair did not land:\n%s", b)
	}
}

// A status stamp on an already-drifted audit must still work: EditFinding routes
// through TransformAuditBody, which the guard deliberately does not cover.
func TestEditFinding_WorksOnAnAuditWithUnrelatedDrift(t *testing.T) {
	source := "---\nid: 6fjangd7kve1\nbucket: open\narea: e\ndate: \"2026-01-01\"\n---\n" +
		"# Audit: e\n\n## Findings\n\n#### H1. good · **Status:** open\n\n#### M-2. drifted\n\n**Status:** open\n"
	fs, p := auditRepo(t, "2026-01-01-e.md", source)
	_, _, changed, err := fs.TransformAuditBody("2026-01-01-e", writeNow, false, func(current string) (string, error) {
		return strings.Replace(current, "#### H1. good · **Status:** open", "#### H1. good · **Status:** fixed", 1), nil
	})
	if err != nil || !changed {
		t.Fatalf("stamping a good finding beside drift must work: changed=%v err=%v", changed, err)
	}
	b, _ := os.ReadFile(p)
	if !strings.Contains(string(b), "**Status:** fixed") {
		t.Errorf("the stamp did not land:\n%s", b)
	}
}
