package cli

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/andy-esch/taskflow/internal/domain"
)

func setupAuditRepo(t *testing.T) string {
	t.Helper()
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
	write("tasks/ready-to-start/.gitkeep", "") // so Discover anchors here
	write("audits/open/o.md", "---\nid: 6fjangd7kvh1\narea: dispatcher\n---\n#### H1. t  · **Status:** open\n")
	write("audits/closed/c.md", "---\nid: 6fjangd7kvh2\narea: web\n---\n#### M1. t  · **Status:** fixed\n")
	return root
}

// TestAuditAppend_JSON pins the `audit_mutation` --json envelope (the contract the
// schema_version 1.20 bump is for): a parseable envelope with the reloaded audit,
// dry_run=false, and the echoed resulting body.
func TestAuditAppend_JSON(t *testing.T) {
	root := setupAuditRepo(t)
	out := runRoot(t, "-C", root, "--json", "audit", "append", "o", "--body", "#### M9. new  · **Status:** open")
	var env struct {
		SchemaVersion string `json:"schema_version"`
		DryRun        bool   `json:"dry_run"`
		Body          string `json:"body"`
		Audit         struct {
			Slug   string `json:"slug"`
			Bucket string `json:"bucket"`
		} `json:"audit"`
	}
	if err := json.Unmarshal([]byte(out), &env); err != nil {
		t.Fatalf("audit append --json is not a parseable envelope: %v\n%s", err, out)
	}
	if env.SchemaVersion == "" || env.Audit.Slug != "o" || env.Audit.Bucket != "open" {
		t.Errorf("audit append --json envelope wrong:\n%s", out)
	}
	if env.DryRun {
		t.Error("a real append should report dry_run=false")
	}
	if !strings.Contains(env.Body, "#### M9. new") {
		t.Errorf("append --json should echo the resulting body:\n%s", out)
	}
}

// --dry-run previews an audit append without writing.
func TestAuditAppend_DryRun_NoWrite(t *testing.T) {
	root := setupAuditRepo(t)
	p := filepath.Join(root, "audits", "open", "o.md")
	before, _ := os.ReadFile(p)
	runRoot(t, "-C", root, "--dry-run", "audit", "append", "o", "--body", "#### NOPE.  · **Status:** open")
	if after, _ := os.ReadFile(p); !bytes.Equal(before, after) {
		t.Error("--dry-run audit append must not write")
	}
}

// Empty append input is a clean validation error, not an empty write.
func TestAuditAppend_Empty_Errors(t *testing.T) {
	root := setupAuditRepo(t)
	var out bytes.Buffer
	cmd := NewRootCmd(strings.NewReader(""), &out, &out)
	cmd.SetArgs([]string{"-C", root, "audit", "append", "o", "--body", "   "})
	if err := cmd.Execute(); !errors.Is(err, domain.ErrValidation) {
		t.Errorf("empty audit append should wrap ErrValidation (exit 11), got %v", err)
	}
}

// Passing both --body and --body-file to `audit append` is a usage error, not a
// silent precedence pick — mirroring `task append`/`task new`.
func TestAuditAppend_BodyAndBodyFile_Exclusive(t *testing.T) {
	root := setupAuditRepo(t)
	if _, err := runRootRC(t, "-C", root, "audit", "append", "o", "--body", "x", "--body-file", "-"); err == nil {
		t.Fatal("`audit append --body … --body-file -` should be rejected (mutually exclusive)")
	}
}

// `audit edit --dry-run` is rejected (it's interactive, no preview) — a safety flag
// must never be a silent no-op.
func TestAuditEdit_RejectsDryRun(t *testing.T) {
	root := setupAuditRepo(t)
	var out bytes.Buffer
	cmd := NewRootCmd(strings.NewReader(""), &out, &out)
	cmd.SetArgs([]string{"-C", root, "--dry-run", "audit", "edit", "o"})
	if err := cmd.Execute(); !errors.Is(err, domain.ErrValidation) {
		t.Errorf("`audit edit --dry-run` should be rejected with ErrValidation, got %v", err)
	}
}

func TestAuditList_DefaultsToOpen(t *testing.T) {
	root := setupAuditRepo(t)
	out := runRoot(t, "-C", root, "audit", "list", "--json")

	var got struct {
		Audits []struct {
			Slug   string `json:"slug"`
			Bucket string `json:"bucket"`
		} `json:"audits"`
	}
	if err := json.Unmarshal([]byte(out), &got); err != nil {
		t.Fatalf("bad json: %v\n%s", err, out)
	}
	if len(got.Audits) != 1 || got.Audits[0].Slug != "o" || got.Audits[0].Bucket != "open" {
		t.Errorf("default should be open only: %+v", got.Audits)
	}
}

func TestAuditList_All(t *testing.T) {
	root := setupAuditRepo(t)
	out := runRoot(t, "-C", root, "audit", "list", "--all")
	if !strings.Contains(out, "o") || !strings.Contains(out, "c") {
		t.Errorf("--all should show both buckets:\n%s", out)
	}
}

func TestAuditClose_MovesBucket(t *testing.T) {
	root := setupAuditRepo(t)
	// A clean audit (no open findings) closes fine — `o` carries an open finding
	// and is covered by TestAuditClose_RejectsOpenFindings below.
	mustWrite(t, filepath.Join(root, "audits", "open", "clean.md"),
		"---\narea: clean\n---\n#### H1. t  · **Status:** fixed\n")
	out := runRoot(t, "-C", root, "audit", "close", "clean")
	if !strings.Contains(out, "clean -> closed") {
		t.Errorf("unexpected output: %q", out)
	}
	if _, err := os.Stat(filepath.Join(root, "audits", "closed", "clean.md")); err != nil {
		t.Errorf("audit not moved to closed: %v", err)
	}
}

// M4 (2026-06-22 audit): closing/deferring an audit that still has open findings
// must be refused (the bucket↔state invariant `audit lint` enforces), with the
// audit left in its original bucket.
func TestAuditClose_RejectsOpenFindings(t *testing.T) {
	root := setupAuditRepo(t) // `o` has H1 open
	if _, err := runRootRC(t, "-C", root, "audit", "close", "o"); err == nil {
		t.Fatal("closing an audit with open findings must be rejected")
	}
	if _, err := os.Stat(filepath.Join(root, "audits", "open", "o.md")); err != nil {
		t.Errorf("a rejected close must leave the audit in open/: %v", err)
	}
}

// TestAuditList_ConflictingFlagsError pins the mutual exclusion: --closed
// --deferred (or --all with either) must error, not silently prefer one.
func TestAuditList_ConflictingFlagsError(t *testing.T) {
	root := setupRepo(t)
	for _, args := range [][]string{
		{"audit", "list", "--closed", "--deferred"},
		{"audit", "list", "--all", "--closed"},
	} {
		var out bytes.Buffer
		cmd := NewRootCmd(strings.NewReader(""), &out, &out)
		cmd.SetArgs(append([]string{"-C", root}, args...))
		cmd.SetOut(&out)
		cmd.SetErr(&out)
		if err := cmd.Execute(); err == nil {
			t.Errorf("%v should error (mutually exclusive flags)", args)
		}
	}
}
