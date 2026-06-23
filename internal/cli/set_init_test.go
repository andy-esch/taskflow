package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/andy-esch/taskflow/internal/cli/prompt"
)

// TestResolveInitTarget pins init's mode decision (the flag-twin): the flag
// forces pointer mode; off a TTY (gate closed) it defaults to scaffold without
// prompting; on a TTY it asks here-vs-elsewhere and, for elsewhere, the typed path.
func TestResolveInitTarget(t *testing.T) {
	// --planning-repo set → pointer outright, no prompt.
	app := &App{Gate: prompt.NewGate(false), Prompt: &prompt.Fake{}}
	if p, repo, err := app.resolveInitTarget("../planning", true); err != nil || !p || repo != "../planning" {
		t.Errorf("flag should force pointer mode: %v %q %v", p, repo, err)
	}
	// Headless (gate off), no flag → scaffold; empty Fake would error if prompted.
	app = &App{Gate: prompt.NewGate(false), Prompt: &prompt.Fake{}}
	if p, _, err := app.resolveInitTarget("", false); err != nil || p {
		t.Errorf("headless no-flag should be scaffold: pointer=%v err=%v", p, err)
	}
	// TTY, choose "here" → scaffold.
	app = &App{Gate: prompt.NewGate(true), Prompt: &prompt.Fake{SelectAnswers: []string{"here"}}}
	if p, _, err := app.resolveInitTarget("", false); err != nil || p {
		t.Errorf(`"here" should be scaffold: pointer=%v err=%v`, p, err)
	}
	// TTY, choose "elsewhere" then type a path → pointer.
	app = &App{Gate: prompt.NewGate(true), Prompt: &prompt.Fake{
		SelectAnswers: []string{"elsewhere"}, TextAnswers: []string{"../planning"},
	}}
	if p, repo, err := app.resolveInitTarget("", false); err != nil || !p || repo != "../planning" {
		t.Errorf(`"elsewhere" should be pointer with the typed path: %v %q %v`, p, repo, err)
	}
}

// TestInit_Pointer: `init --planning-repo` writes a pointer config (no tree) and
// --json reports mode "pointer".
func TestInit_Pointer(t *testing.T) {
	parent := t.TempDir()
	impl := filepath.Join(parent, "impl")
	if err := os.MkdirAll(impl, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(parent, "planning", "tasks"), 0o755); err != nil {
		t.Fatal(err)
	}
	out := runRoot(t, "init", "--path", impl, "--planning-repo", "../planning", "--json")
	var got struct {
		Mode         string   `json:"mode"`
		PlanningRepo string   `json:"planning_repo"`
		Created      []string `json:"created"`
	}
	if err := json.Unmarshal([]byte(out), &got); err != nil {
		t.Fatalf("invalid json: %v\n%s", err, out)
	}
	if got.Mode != "pointer" || got.PlanningRepo != "../planning" {
		t.Errorf("pointer --json wrong: %+v", got)
	}
	if _, err := os.Stat(filepath.Join(impl, ".tskflwctl.toml")); err != nil {
		t.Errorf("pointer config not written: %v", err)
	}
	if _, err := os.Stat(filepath.Join(impl, "tasks")); !os.IsNotExist(err) {
		t.Error("pointer mode must not scaffold tasks/")
	}
}

// TestInit_Track: `init --track` seeds the planning repo's tracked_repos (and
// dedups repeated flags).
func TestInit_Track(t *testing.T) {
	planning := filepath.Join(t.TempDir(), "planning")
	runRoot(t, "init", "--path", planning, "--track", "../impl-a", "--track", "../impl-a")
	b, err := os.ReadFile(filepath.Join(planning, ".tskflwctl.toml"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(b), `tracked_repos = ["../impl-a"]`) {
		t.Errorf("--track should seed (and dedup) tracked_repos:\n%s", b)
	}
}

// TestInit_LinkBack: `init --planning-repo` records this repo in the planning
// repo's tracked_repos; `--no-link-back` suppresses it.
func TestInit_LinkBack(t *testing.T) {
	parent := t.TempDir()
	impl := filepath.Join(parent, "impl")
	planning := filepath.Join(parent, "planning")
	if err := os.MkdirAll(impl, 0o755); err != nil {
		t.Fatal(err)
	}
	runRoot(t, "init", "--path", planning) // scaffold the planning repo

	runRoot(t, "init", "--path", impl, "--planning-repo", "../planning")
	if b, _ := os.ReadFile(filepath.Join(planning, ".tskflwctl.toml")); !strings.Contains(string(b), `"../impl"`) {
		t.Errorf("auto-link-back should record the impl in planning's tracked_repos:\n%s", b)
	}

	other := filepath.Join(parent, "other")
	if err := os.MkdirAll(other, 0o755); err != nil {
		t.Fatal(err)
	}
	runRoot(t, "init", "--path", other, "--planning-repo", "../planning", "--no-link-back")
	if b, _ := os.ReadFile(filepath.Join(planning, ".tskflwctl.toml")); strings.Contains(string(b), `"../other"`) {
		t.Errorf("--no-link-back must not record the impl:\n%s", b)
	}
}

// TestInit_TrackPointerConflict: --track is meaningless in pointer mode → exit 11.
func TestInit_TrackPointerConflict(t *testing.T) {
	parent := t.TempDir()
	impl := filepath.Join(parent, "impl")
	if err := os.MkdirAll(impl, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(parent, "planning", "tasks"), 0o755); err != nil {
		t.Fatal(err)
	}
	if _, err := runRootRC(t, "init", "--path", impl, "--planning-repo", "../planning", "--track", "../x"); err == nil || ExitCode(err) != 11 {
		t.Fatalf("--track + --planning-repo should exit 11, got %v", err)
	}
}

// TestInit_Pointer_BadTarget: a planning-repo that isn't a planning root exits 11
// and writes nothing.
func TestInit_Pointer_BadTarget(t *testing.T) {
	impl := t.TempDir()
	if _, err := runRootRC(t, "init", "--path", impl, "--planning-repo", "../nope"); err == nil || ExitCode(err) != 11 {
		t.Fatalf("a bad planning-repo target should exit 11, got %v", err)
	}
	if _, err := os.Stat(filepath.Join(impl, ".tskflwctl.toml")); !os.IsNotExist(err) {
		t.Error("a rejected pointer init must leave no config behind")
	}
}

func TestTaskSet(t *testing.T) {
	root := setupRepo(t)
	out := runRoot(t, "-C", root, "task", "set", "alpha", "--priority", "low", "--tags", "x,y")
	if !strings.Contains(out, "updated alpha") {
		t.Errorf("unexpected output: %q", out)
	}
	b, err := os.ReadFile(filepath.Join(root, "tasks", "ready-to-start", "alpha.md"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(b), "priority: low") {
		t.Errorf("priority not set:\n%s", b)
	}
}

func TestTaskSet_ArbitraryKeyValue(t *testing.T) {
	root := setupRepo(t)
	// Unknown keys are rejected without --force (decided 2026-06-12): a typo'd
	// field name must not silently persist.
	{
		var out bytes.Buffer
		cmd := NewRootCmd(strings.NewReader(""), &out, &out)
		cmd.SetArgs([]string{"-C", root, "task", "set", "alpha", "--set", "owner=me"})
		cmd.SetOut(&out)
		cmd.SetErr(&out)
		err := cmd.Execute()
		if err == nil || ExitCode(err) != 11 || !strings.Contains(err.Error(), "--force") {
			t.Fatalf("unknown key without --force should exit 11 mentioning --force, got %v", err)
		}
	}
	out := runRoot(t, "-C", root, "task", "set", "alpha", "--force",
		"--set", "owner=me", "--set", "custom_field=keep me")
	if !strings.Contains(out, "updated alpha") {
		t.Errorf("unexpected output: %q", out)
	}
	b, err := os.ReadFile(filepath.Join(root, "tasks", "ready-to-start", "alpha.md"))
	if err != nil {
		t.Fatal(err)
	}
	s := string(b)
	if !strings.Contains(s, "owner: me") {
		t.Errorf("arbitrary field not written:\n%s", s)
	}
	// A value with a space (the colon-free case) must round-trip readably.
	if !strings.Contains(s, "custom_field:") || !strings.Contains(s, "keep me") {
		t.Errorf("arbitrary field with space not written:\n%s", s)
	}
	// The body and existing fields must survive the surgical write.
	if !strings.Contains(s, "# Alpha") || !strings.Contains(s, "status: ready-to-start") {
		t.Errorf("body or existing field lost:\n%s", s)
	}
}

func TestTaskSet_MalformedSet_Errors(t *testing.T) {
	root := setupRepo(t)
	var out bytes.Buffer
	cmd := NewRootCmd(strings.NewReader(""), &out, &out)
	cmd.SetArgs([]string{"-C", root, "task", "set", "alpha", "--set", "noequals"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected an error for --set without '='")
	}
	if ExitCode(err) != 11 {
		t.Errorf("want exit code 11 (validation), got %d", ExitCode(err))
	}
}

func TestTaskSet_InvalidPriority_Exit11(t *testing.T) {
	root := setupRepo(t)
	var out bytes.Buffer
	cmd := NewRootCmd(strings.NewReader(""), &out, &out)
	cmd.SetArgs([]string{"-C", root, "task", "set", "alpha", "--priority", "urgent"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected validation error")
	}
	if ExitCode(err) != 11 {
		t.Errorf("want exit code 11 (validation), got %d", ExitCode(err))
	}
}

func TestTaskSet_NoFields_Errors(t *testing.T) {
	root := setupRepo(t)
	var out bytes.Buffer
	cmd := NewRootCmd(strings.NewReader(""), &out, &out)
	cmd.SetArgs([]string{"-C", root, "task", "set", "alpha"})
	if err := cmd.Execute(); err == nil {
		t.Fatal("expected an error when no fields are given")
	}
}

func TestInit_ThenList(t *testing.T) {
	root := t.TempDir()
	out := runRoot(t, "init", "--path", root)
	if !strings.Contains(out, "initialized") {
		t.Errorf("unexpected init output: %q", out)
	}
	// list should now work (and be empty) without a "not a planning repo" error.
	if listOut := runRoot(t, "-C", root, "task", "list"); listOut != "" {
		t.Errorf("expected empty list, got %q", listOut)
	}
}
