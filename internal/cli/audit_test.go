package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
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
	write("audits/open/o.md", "---\narea: dispatcher\n---\n#### H1. t  · **Status:** open\n")
	write("audits/closed/c.md", "---\narea: web\n---\n#### M1. t  · **Status:** fixed\n")
	return root
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
	out := runRoot(t, "-C", root, "audit", "close", "o")
	if !strings.Contains(out, "o -> closed") {
		t.Errorf("unexpected output: %q", out)
	}
	if _, err := os.Stat(filepath.Join(root, "audits", "closed", "o.md")); err != nil {
		t.Errorf("audit not moved to closed: %v", err)
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
		cmd := NewRootCmd(&out, &out)
		cmd.SetArgs(append([]string{"-C", root}, args...))
		cmd.SetOut(&out)
		cmd.SetErr(&out)
		if err := cmd.Execute(); err == nil {
			t.Errorf("%v should error (mutually exclusive flags)", args)
		}
	}
}
