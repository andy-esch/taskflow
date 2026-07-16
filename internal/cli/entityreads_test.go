package cli

import (
	"encoding/json"
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/andy-esch/taskflow/internal/domain"
	"github.com/andy-esch/taskflow/internal/testutil"
)

// canonAbs is the canonical absolute path (symlinks resolved, as discovery does).
func canonAbs(t *testing.T, p string) string {
	t.Helper()
	r, err := filepath.EvalSymlinks(p)
	if err != nil {
		t.Fatal(err)
	}
	return r
}

// --- epic ---

func TestEpicPath(t *testing.T) {
	root := setupEpicRepo(t)
	out := runRoot(t, "-C", root, "epic", "path", "demo")
	want := canonAbs(t, filepath.Join(root, "epics", "demo.md"))
	if strings.TrimSpace(out) != want {
		t.Errorf("epic path = %q, want %q", strings.TrimSpace(out), want)
	}
}

func TestEpicShow_FrontmatterOnly(t *testing.T) {
	root := setupEpicRepo(t)
	out := runRoot(t, "-C", root, "epic", "show", "demo", "--frontmatter-only")
	if strings.Contains(out, "# Demo Epic") {
		t.Errorf("--frontmatter-only must drop the epic body:\n%s", out)
	}
	if !strings.Contains(out, "demo") { // metadata + roster still shown
		t.Errorf("--frontmatter-only should still show epic metadata/roster:\n%s", out)
	}
	if strings.HasSuffix(out, "\n\n") { // no stray blank line under the roster
		t.Errorf("epic --frontmatter-only must not leave a trailing blank line:\n%q", out)
	}
}

// audit --frontmatter-only drops the body (keeping the finding tree) with no stray
// trailing blank line.
func TestAuditShow_FrontmatterOnly(t *testing.T) {
	root := setupAuditRepo(t)
	out := runRoot(t, "-C", root, "audit", "show", "o", "--frontmatter-only")
	if strings.HasSuffix(out, "\n\n") {
		t.Errorf("audit --frontmatter-only must not leave a trailing blank line:\n%q", out)
	}
}

func TestEpicShow_SectionNotFound(t *testing.T) {
	root := setupEpicRepo(t)
	if _, err := runRootRC(t, "-C", root, "epic", "show", "demo", "--section", "nope"); !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("a missing epic section should wrap ErrNotFound, got %v", err)
	}
}

// --- audit ---

func TestAuditPath(t *testing.T) {
	root := setupAuditRepo(t)
	out := runRoot(t, "-C", root, "audit", "path", "o")
	want := canonAbs(t, filepath.Join(root, domain.AuditsDir, testutil.TaskID("o")+"-o.md"))
	if strings.TrimSpace(out) != want {
		t.Errorf("audit path = %q, want %q", strings.TrimSpace(out), want)
	}
}

// audit info reports the finding disposition tally (the audit analogue of the AC
// tally) without the body. Fixture audit "o" has one open finding.
func TestAuditInfo_JSON(t *testing.T) {
	root := setupAuditRepo(t)
	out := runRoot(t, "-C", root, "--json", "audit", "info", "o")
	var env struct {
		SchemaVersion string `json:"schema_version"`
		AuditInfo     struct {
			Slug     string `json:"slug"`
			Bucket   string `json:"bucket"`
			Path     string `json:"path"`
			Findings struct {
				Total      int `json:"total"`
				Open       int `json:"open"`
				InProgress int `json:"in_progress"`
				Done       int `json:"done"`
				Dropped    int `json:"dropped"`
			} `json:"findings"`
		} `json:"audit_info"`
	}
	if err := json.Unmarshal([]byte(out), &env); err != nil {
		t.Fatalf("audit info --json not parseable: %v\n%s", err, out)
	}
	ai := env.AuditInfo
	want := canonAbs(t, filepath.Join(root, domain.AuditsDir, testutil.TaskID("o")+"-o.md"))
	if env.SchemaVersion == "" || ai.Slug != "o" || ai.Bucket != "open" || ai.Path != want {
		t.Errorf("audit info metadata wrong:\n%s", out)
	}
	if ai.Findings.Total != 1 || ai.Findings.Open != 1 {
		t.Errorf("findings tally = %+v, want {Total:1 Open:1}\n%s", ai.Findings, out)
	}
}

// audit show --section narrows to one body section; the metadata + finding tree
// still render, but other body sections do not.
func TestAuditShow_Section(t *testing.T) {
	root := setupAuditRepo(t)
	p, out := testutil.AuditFixture(root, "open", "o.md",
		"---\nid: "+testutil.TaskID("o")+"\nbucket: open\narea: dispatcher\n---\n## Threat model\n\nboundaries.\n\n## Findings\n\n#### H1. t  · **Status:** open\n")
	testutil.Write(t, p, out)
	res := runRoot(t, "-C", root, "audit", "show", "o", "--section", "findings", "--raw")
	if !strings.Contains(res, "## Findings") || !strings.Contains(res, "H1") {
		t.Errorf("--section findings should show that section:\n%s", res)
	}
	if strings.Contains(res, "Threat model") || strings.Contains(res, "boundaries.") {
		t.Errorf("--section findings must not show other body sections:\n%s", res)
	}
}

// --- parse-free path (the Q1 follow-up) ---

// task path resolves even a file whose frontmatter won't parse — exactly when you
// need the path to open and repair it. task show, which parses, must still fail.
func TestTaskPath_BrokenFrontmatter(t *testing.T) {
	root := setupRepo(t)
	broken := filepath.Join(root, domain.TasksDir, testutil.TaskID("broken")+"-broken.md")
	testutil.Write(t, broken, "---\nstatus: ready-to-start\nbad: [unclosed\n---\n# Broken\n")

	out := runRoot(t, "-C", root, "task", "path", "broken")
	if strings.TrimSpace(out) != canonAbs(t, broken) {
		t.Errorf("task path must resolve a broken-frontmatter file:\ngot  %q\nwant %q", strings.TrimSpace(out), canonAbs(t, broken))
	}
	if _, err := runRootRC(t, "-C", root, "task", "show", "broken"); err == nil {
		t.Error("task show should fail on broken frontmatter (the contrast that motivates a parse-free path)")
	}
}
