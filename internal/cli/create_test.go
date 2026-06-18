package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// freshRepo inits an empty planning tree and returns its root.
func freshRepo(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	runRoot(t, "init", "--path", root)
	return root
}

func TestTaskNew_HappyPath(t *testing.T) {
	root := freshRepo(t)
	mustWrite(t, filepath.Join(root, "epics", "e1.md"), "---\nstatus: in-progress\n---\n# E1\n")

	out := runRoot(t, "-C", root, "task", "new", "My Brand New Task", "--epic", "e1", "--tags", "a,b")
	if !strings.Contains(out, "created") {
		t.Errorf("unexpected output: %q", out)
	}
	b, err := os.ReadFile(filepath.Join(root, "tasks", "ready-to-start", "my-brand-new-task.md"))
	if err != nil {
		t.Fatalf("task file not created: %v", err)
	}
	s := string(b)
	for _, want := range []string{
		"status: ready-to-start", "epic: e1", "tier: 3", "priority: medium",
		"effort: Unknown", "## Acceptance criteria", "Epic [[e1]]",
	} {
		if !strings.Contains(s, want) {
			t.Errorf("created task missing %q:\n%s", want, s)
		}
	}
	// Round-trips through show, and the repo is lint-clean (only this task).
	if show := runRoot(t, "-C", root, "task", "show", "my-brand-new-task"); !strings.Contains(show, "e1") {
		t.Errorf("show failed: %q", show)
	}
	runRoot(t, "-C", root, "lint") // would Fatalf if exit != 0
}

func TestTaskNew_Next(t *testing.T) {
	root := freshRepo(t)
	mustWrite(t, filepath.Join(root, "epics", "e1.md"), "---\nstatus: in-progress\n---\n")
	runRoot(t, "-C", root, "task", "new", "Soon", "--epic", "e1", "--tags", "x", "--description", "soon work", "--next")
	if _, err := os.Stat(filepath.Join(root, "tasks", "next-up", "soon.md")); err != nil {
		t.Errorf("--next should land in next-up/: %v", err)
	}
}

// TestTaskNew_ActiveRequiresDescription pins L4: a task born next-up/in-progress
// is active, so `new --next`/`--start` must carry a --description (else it would
// scaffold a file lint immediately rejects). Exit 11.
func TestTaskNew_ActiveRequiresDescription(t *testing.T) {
	root := freshRepo(t)
	mustWrite(t, filepath.Join(root, "epics", "e1.md"), "---\nstatus: in-progress\n---\n")
	for _, flag := range []string{"--next", "--start"} {
		var out bytes.Buffer
		cmd := NewRootCmd(&out, &out)
		cmd.SetArgs([]string{"-C", root, "task", "new", "X", "--epic", "e1", "--tags", "t", flag})
		if err := cmd.Execute(); err == nil || ExitCode(err) != 11 {
			t.Errorf("%s without --description should exit 11, got %v", flag, err)
		}
	}
}

func TestTaskNew_Start(t *testing.T) {
	root := freshRepo(t)
	mustWrite(t, filepath.Join(root, "epics", "e1.md"), "---\nstatus: in-progress\n---\n")
	runRoot(t, "-C", root, "task", "new", "Start Me", "--epic", "e1", "--tags", "x", "--description", "start it", "--start")
	path := filepath.Join(root, "tasks", "in-progress", "start-me.md")
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("--start should land in in-progress/: %v", err)
	}
	// A task born in-progress carries started_at, like one moved there.
	if !strings.Contains(string(b), "started_at:") {
		t.Errorf("--start task should stamp started_at:\n%s", b)
	}
}

func TestTaskNew_BodyFile(t *testing.T) {
	root := freshRepo(t)
	mustWrite(t, filepath.Join(root, "epics", "e1.md"), "---\nstatus: in-progress\n---\n")
	bf := filepath.Join(t.TempDir(), "body.md")
	mustWrite(t, bf, "\n# Custom\n\nfrom a file\n")
	runRoot(t, "-C", root, "task", "new", "File Body", "--epic", "e1", "--tags", "x", "--body-file", bf)
	b, err := os.ReadFile(filepath.Join(root, "tasks", "ready-to-start", "file-body.md"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(b), "from a file") || strings.Contains(string(b), "## Acceptance criteria") {
		t.Errorf("--body-file should replace the scaffold:\n%s", b)
	}
}

func TestTaskNew_BodyFileStdin(t *testing.T) {
	root := freshRepo(t)
	mustWrite(t, filepath.Join(root, "epics", "e1.md"), "---\nstatus: in-progress\n---\n")
	var out bytes.Buffer
	cmd := NewRootCmd(&out, &out)
	cmd.SetIn(strings.NewReader("\n# Piped\n\nfrom stdin\n"))
	cmd.SetArgs([]string{"-C", root, "task", "new", "Piped Body", "--epic", "e1", "--tags", "x", "--body-file", "-"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(filepath.Join(root, "tasks", "ready-to-start", "piped-body.md"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(b), "from stdin") {
		t.Errorf("body should come from stdin:\n%s", b)
	}
}

func TestTaskNew_MutuallyExclusiveFlags(t *testing.T) {
	root := freshRepo(t)
	mustWrite(t, filepath.Join(root, "epics", "e1.md"), "---\nstatus: in-progress\n---\n")
	for _, extra := range [][]string{
		{"--next", "--start"},
		{"--body", "x", "--body-file", "-"},
	} {
		var out bytes.Buffer
		cmd := NewRootCmd(&out, &out)
		cmd.SetArgs(append([]string{"-C", root, "task", "new", "X", "--epic", "e1", "--tags", "x"}, extra...))
		if err := cmd.Execute(); err == nil {
			t.Errorf("expected a flag-conflict error for %v", extra)
		}
	}
}

func TestTaskNew_UnknownEpic_Exit11(t *testing.T) {
	root := freshRepo(t)
	var out bytes.Buffer
	cmd := NewRootCmd(&out, &out)
	cmd.SetArgs([]string{"-C", root, "task", "new", "X", "--epic", "nope"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for unknown epic")
	}
	if ExitCode(err) != 11 {
		t.Errorf("want exit 11, got %d", ExitCode(err))
	}
}

func TestTaskNew_RefusesClobber(t *testing.T) {
	root := freshRepo(t)
	mustWrite(t, filepath.Join(root, "epics", "e1.md"), "---\nstatus: in-progress\n---\n")
	runRoot(t, "-C", root, "task", "new", "Dup", "--epic", "e1", "--tags", "x")
	var out bytes.Buffer
	cmd := NewRootCmd(&out, &out)
	cmd.SetArgs([]string{"-C", root, "task", "new", "Dup", "--epic", "e1", "--tags", "x"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected refusal to clobber an existing task")
	}
	if ExitCode(err) != 14 {
		t.Errorf("clobber should exit 14 (conflict), got %d", ExitCode(err))
	}
}

func TestEpicNew(t *testing.T) {
	root := freshRepo(t)
	out := runRoot(t, "-C", root, "epic", "new", "Payments Revamp", "--description", "Overhaul payments")
	if !strings.Contains(out, "created") {
		t.Errorf("unexpected output: %q", out)
	}
	b, err := os.ReadFile(filepath.Join(root, "epics", "01-payments-revamp.md"))
	if err != nil {
		t.Fatalf("epic not created (auto-number): %v", err)
	}
	s := string(b)
	for _, want := range []string{"status: planning", "description: Overhaul payments", "priority: medium", "**Goal.**"} {
		if !strings.Contains(s, want) {
			t.Errorf("epic missing %q:\n%s", want, s)
		}
	}
}

func TestAuditNew(t *testing.T) {
	root := freshRepo(t)
	out := runRoot(t, "-C", root, "audit", "new", "dispatcher", "--date", "2026-06-16")
	if !strings.Contains(out, "created") {
		t.Errorf("unexpected output: %q", out)
	}
	b, err := os.ReadFile(filepath.Join(root, "audits", "open", "2026-06-16-dispatcher.md"))
	if err != nil {
		t.Fatalf("audit file not created in open/: %v", err)
	}
	s := string(b)
	for _, want := range []string{
		"area: dispatcher", "date: \"2026-06-16\"", "## Findings", "## Candidate tasks",
	} {
		if !strings.Contains(s, want) {
			t.Errorf("created audit missing %q:\n%s", want, s)
		}
	}
	// The scaffold is generic — no repo-specific conventions-doc link.
	if strings.Contains(s, "HOWTO-execute") {
		t.Errorf("scaffold should not hardcode a repo-specific HOWTO link:\n%s", s)
	}
	// Round-trips through show.
	if show := runRoot(t, "-C", root, "audit", "show", "2026-06-16-dispatcher"); !strings.Contains(show, "dispatcher") {
		t.Errorf("show failed: %q", show)
	}
	// The fenced example finding must not inflate the count: a fresh audit is empty.
	var lst struct {
		Audits []struct {
			Slug     string `json:"slug"`
			Findings int    `json:"findings"`
		} `json:"audits"`
	}
	if err := json.Unmarshal([]byte(runRoot(t, "-C", root, "audit", "list", "--json")), &lst); err != nil {
		t.Fatalf("audit list --json invalid: %v", err)
	}
	if len(lst.Audits) != 1 || lst.Audits[0].Findings != 0 {
		t.Errorf("fresh audit should list once with 0 findings, got %+v", lst.Audits)
	}
	// Lifecycle round-trips through the CLI: close moves it to closed/.
	runRoot(t, "-C", root, "audit", "close", "2026-06-16-dispatcher")
	if _, err := os.Stat(filepath.Join(root, "audits", "closed", "2026-06-16-dispatcher.md")); err != nil {
		t.Errorf("close should move the audit to closed/: %v", err)
	}
}

func TestTaskNew_RejectsSlugInAnotherBucket(t *testing.T) {
	root := freshRepo(t)
	mustWrite(t, filepath.Join(root, "epics", "e1.md"), "---\nstatus: in-progress\n---\n")
	runRoot(t, "-C", root, "task", "new", "Dup Task", "--epic", "e1", "--tags", "x")
	runRoot(t, "-C", root, "task", "complete", "dup-task") // → completed/
	// Re-creating the same title (slug now in completed/) refuses with exit 14.
	var out bytes.Buffer
	cmd := NewRootCmd(&out, &out)
	cmd.SetArgs([]string{"-C", root, "task", "new", "Dup Task", "--epic", "e1", "--tags", "x"})
	if err := cmd.Execute(); err == nil || ExitCode(err) != 14 {
		t.Errorf("cross-bucket slug collision should exit 14, got %v", err)
	}
	// The slug still resolves — no second file was created.
	if show := runRoot(t, "-C", root, "task", "show", "dup-task"); !strings.Contains(show, "dup-task") {
		t.Errorf("dup-task should still resolve to a single file: %q", show)
	}
}

func TestAuditNew_BodyOverride(t *testing.T) {
	root := freshRepo(t)
	runRoot(t, "-C", root, "audit", "new", "dispatcher", "--date", "2026-06-17",
		"--body", "\n# Custom\n\nhand-written body\n")
	b, err := os.ReadFile(filepath.Join(root, "audits", "open", "2026-06-17-dispatcher.md"))
	if err != nil {
		t.Fatal(err)
	}
	s := string(b)
	if !strings.Contains(s, "hand-written body") || strings.Contains(s, "## Findings") {
		t.Errorf("--body should replace the scaffold, got:\n%s", s)
	}
}

func TestAuditNew_BodyFile(t *testing.T) {
	root := freshRepo(t)
	bf := filepath.Join(t.TempDir(), "body.md")
	mustWrite(t, bf, "\n# Custom\n\naudit body from a file\n")
	runRoot(t, "-C", root, "audit", "new", "dispatcher", "--date", "2026-06-17", "--body-file", bf)
	b, err := os.ReadFile(filepath.Join(root, "audits", "open", "2026-06-17-dispatcher.md"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(b), "audit body from a file") || strings.Contains(string(b), "## Findings") {
		t.Errorf("audit --body-file should replace the scaffold:\n%s", b)
	}
}

func TestEpicNew_Body(t *testing.T) {
	root := freshRepo(t)
	runRoot(t, "-C", root, "epic", "new", "Payments", "--description", "d", "--body", "\n# Custom\n\nepic body here\n")
	b, err := os.ReadFile(filepath.Join(root, "epics", "01-payments.md"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(b), "epic body here") || strings.Contains(string(b), "**Goal.**") {
		t.Errorf("epic --body should replace the scaffold:\n%s", b)
	}
}

func TestEpicNew_BodyFileStdin(t *testing.T) {
	root := freshRepo(t)
	var out bytes.Buffer
	cmd := NewRootCmd(&out, &out)
	cmd.SetIn(strings.NewReader("\n# Piped\n\nepic from stdin\n"))
	cmd.SetArgs([]string{"-C", root, "epic", "new", "Streamed", "--description", "d", "--body-file", "-"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(filepath.Join(root, "epics", "01-streamed.md"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(b), "epic from stdin") {
		t.Errorf("epic --body-file - should read stdin:\n%s", b)
	}
}

func TestAuditNew_JSONEnvelope(t *testing.T) {
	root := freshRepo(t)
	js := runRoot(t, "-C", root, "audit", "new", "arch-data-flow", "--date", "2026-06-16", "--json")
	var env struct {
		DryRun  bool `json:"dry_run"`
		Created struct {
			Kind, ID, Status, Path string
		} `json:"created"`
	}
	if err := json.Unmarshal([]byte(js), &env); err != nil {
		t.Fatalf("audit new --json invalid: %v\n%s", err, js)
	}
	if env.DryRun || env.Created.Kind != "audit" || env.Created.ID != "2026-06-16-arch-data-flow" {
		t.Errorf("envelope wrong: %+v", env)
	}
	// status = the audit bucket; path is relative to the planning root (not absolute).
	if env.Created.Status != "open" || env.Created.Path != "audits/open/2026-06-16-arch-data-flow.md" {
		t.Errorf("envelope status/path wrong: %+v", env.Created)
	}
}

func TestAuditNew_BadDate_Exit11(t *testing.T) {
	root := freshRepo(t)
	var out bytes.Buffer
	cmd := NewRootCmd(&out, &out)
	cmd.SetArgs([]string{"-C", root, "audit", "new", "x", "--date", "06-16-2026"})
	if err := cmd.Execute(); err == nil || ExitCode(err) != 11 {
		t.Errorf("a malformed date should exit 11 (validation), got %v", err)
	}
}

func TestAuditNew_RefusesClobber(t *testing.T) {
	root := freshRepo(t)
	runRoot(t, "-C", root, "audit", "new", "dispatcher", "--date", "2026-06-16")
	var out bytes.Buffer
	cmd := NewRootCmd(&out, &out)
	cmd.SetArgs([]string{"-C", root, "audit", "new", "dispatcher", "--date", "2026-06-16"})
	if err := cmd.Execute(); err == nil || ExitCode(err) != 14 {
		t.Errorf("clobber should exit 14 (conflict), got %v", err)
	}
}

func TestList_MisfiledMarker(t *testing.T) {
	root := freshRepo(t)
	// A file in ready-to-start/ (active, so it shows) whose frontmatter claims a
	// different recognized status → misfiled, marked with ⚠.
	mustWrite(t, filepath.Join(root, "tasks", "ready-to-start", "drift.md"),
		"---\nstatus: completed\nepic: e\ndescription: d\ntier: 3\npriority: low\neffort: x\ncreated: 2026-06-09\ntags: [a]\n---\n# x\n")
	out := runRoot(t, "-C", root, "task", "list")
	if !strings.Contains(out, "⚠") {
		t.Errorf("expected a ⚠ misfiled marker:\n%q", out)
	}
	if !strings.Contains(out, "misfiled") {
		t.Errorf("expected a misfiled footer note:\n%q", out)
	}
}

func TestLint_FlagsMisfiledArchivedTask(t *testing.T) {
	root := freshRepo(t)
	// A completed/ file with a stale active status — archived, so field lint is
	// skipped, but the drift must still surface.
	mustWrite(t, filepath.Join(root, "tasks", "completed", "old.md"),
		"---\nstatus: in-progress\n---\n# x\n")
	var out bytes.Buffer
	cmd := NewRootCmd(&out, &out)
	cmd.SetArgs([]string{"-C", root, "lint"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected lint to fail on a misfiled archived task")
	}
	if !strings.Contains(out.String(), "old") || !strings.Contains(out.String(), "folder") {
		t.Errorf("expected a misfiled report for 'old':\n%s", out.String())
	}
}

func TestEpicNew_RequiresDescription(t *testing.T) {
	root := freshRepo(t)
	var out bytes.Buffer
	cmd := NewRootCmd(&out, &out)
	cmd.SetArgs([]string{"-C", root, "epic", "new", "X"})
	if err := cmd.Execute(); err == nil {
		t.Fatal("expected error when --description is missing")
	}
}
