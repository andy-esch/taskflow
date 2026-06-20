package cli

import (
	"bytes"
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/andy-esch/taskflow/internal/cli/prompt"
	"github.com/andy-esch/taskflow/internal/domain"
)

// failOpts asserts the option provider is NOT called (the non-prompt paths must
// never pay to gather options).
func failOpts(t *testing.T) func() ([]prompt.Option, error) {
	return func() ([]prompt.Option, error) {
		t.Helper()
		t.Fatal("optsFn must not be called on the flag/error path")
		return nil, nil
	}
}

func TestFillSelect_FlagValueWins(t *testing.T) {
	// Even with the gate open, a provided value short-circuits — no prompt, no opts.
	app := &App{Gate: prompt.NewGate(true), Prompt: &prompt.Fake{}}
	got, err := app.fillSelect("already", "--epic is required", "no epics", "Pick", failOpts(t))
	if err != nil || got != "already" {
		t.Fatalf("fillSelect(value) = %q, %v; want already", got, err)
	}
}

func TestFillSelect_GateClosed_RequiredError(t *testing.T) {
	// Agent/pipeline path: missing value, gate closed → exit-11 validation error,
	// and the options provider is never invoked.
	app := &App{Gate: prompt.NewGate(false), Prompt: &prompt.Fake{}}
	_, err := app.fillSelect("", "--epic is required", "no epics", "Pick", failOpts(t))
	if !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("closed gate should wrap ErrValidation (exit 11), got %v", err)
	}
	if !bytes.Contains([]byte(err.Error()), []byte("--epic")) {
		t.Errorf("error should name the flag: %v", err)
	}
}

func TestFillSelect_GateOpen_Prompts(t *testing.T) {
	// Human path: missing value, gate open → prompt fires and returns the choice.
	f := &prompt.Fake{SelectAnswers: []string{"picked"}}
	app := &App{Gate: prompt.NewGate(true), Prompt: f}
	got, err := app.fillSelect("", "--epic is required", "no epics", "Epic for this task", func() ([]prompt.Option, error) {
		return []prompt.Option{{Label: "E1", Value: "picked"}}, nil
	})
	if err != nil || got != "picked" {
		t.Fatalf("fillSelect(prompt) = %q, %v; want picked", got, err)
	}
	if len(f.Asked) != 1 || f.Asked[0] != "Epic for this task" {
		t.Errorf("expected one prompt with the title, got %v", f.Asked)
	}
}

// TestFillSelect_EmptyOptions: gate open but nothing to pick → a helpful error,
// never a dead-end picker (the prompter is not even called).
func TestFillSelect_EmptyOptions_Errors(t *testing.T) {
	app := &App{Gate: prompt.NewGate(true), Prompt: &prompt.Fake{}} // empty fake would error if called
	_, err := app.fillSelect("", "--epic is required", "no epics exist yet", "Pick",
		func() ([]prompt.Option, error) { return nil, nil })
	if !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("empty options should wrap ErrValidation, got %v", err)
	}
	if !strings.Contains(err.Error(), "no epics exist yet") {
		t.Errorf("empty error should use emptyMsg, got %v", err)
	}
}

// TestGateOpen pins the gate contract: prompting requires ALL of !--json,
// !--no-input, stdin-TTY, and stderr-TTY — any one false closes it.
func TestGateOpen(t *testing.T) {
	if !gateOpen(false, false, true, true) {
		t.Error("all clear should open the gate")
	}
	for name, args := range map[string][4]bool{
		"--json":      {true, false, true, true},
		"--no-input":  {false, true, true, true},
		"stdin pipe":  {false, false, false, true},
		"stderr pipe": {false, false, true, false},
	} {
		if gateOpen(args[0], args[1], args[2], args[3]) {
			t.Errorf("%s should close the gate", name)
		}
	}
}

// TestExitCode_Abort pins the ctrl-c contract: a prompt abort is exit 130 (the
// SIGINT convention) and renders as a quiet "aborted", not a scary error.
func TestExitCode_Abort(t *testing.T) {
	if got := ExitCode(prompt.ErrAborted); got != 130 {
		t.Errorf("aborted prompt should be exit 130, got %d", got)
	}
	var b bytes.Buffer
	WriteError(&b, prompt.ErrAborted, false)
	if got := strings.TrimSpace(b.String()); got != "aborted" {
		t.Errorf("aborted prompt should render as a quiet line, got %q", got)
	}
}

func TestFillText_FlagValueWins(t *testing.T) {
	app := &App{Gate: prompt.NewGate(true), Prompt: &prompt.Fake{}} // empty fake errors if called
	got, err := app.fillText("already", "--description is required", "Desc", "ph")
	if err != nil || got != "already" {
		t.Fatalf("fillText(value) = %q, %v; want already", got, err)
	}
}

func TestFillText_GateClosed_RequiredError(t *testing.T) {
	app := &App{Gate: prompt.NewGate(false), Prompt: &prompt.Fake{}}
	if _, err := app.fillText("", "--description is required", "Desc", "ph"); !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("closed gate should wrap ErrValidation (exit 11), got %v", err)
	}
}

func TestFillText_GateOpen_Prompts(t *testing.T) {
	f := &prompt.Fake{TextAnswers: []string{"typed it"}}
	app := &App{Gate: prompt.NewGate(true), Prompt: f}
	got, err := app.fillText("", "req", "Description", "ph")
	if err != nil || got != "typed it" {
		t.Fatalf("fillText(prompt) = %q, %v; want 'typed it'", got, err)
	}
	if len(f.Asked) != 1 || f.Asked[0] != "Description" {
		t.Errorf("expected one text prompt with the title, got %v", f.Asked)
	}
}

// TestTaskNew_StartMissingDescription_NonInteractive: a --start task without
// --description, run non-interactively, is exit-11 — the prompt is TTY-only.
func TestTaskNew_StartMissingDescription_NonInteractive(t *testing.T) {
	root := freshRepo(t)
	mustWrite(t, filepath.Join(root, "epics", "e1.md"), "---\nstatus: in-progress\n---\n")

	var out bytes.Buffer
	cmd := NewRootCmd(&out, &out)
	cmd.SetArgs([]string{"-C", root, "task", "new", "T", "--epic", "e1", "--tags", "a", "--start"})
	if err := cmd.Execute(); !errors.Is(err, domain.ErrValidation) {
		t.Errorf("--start without --description non-interactively should be ErrValidation (exit 11), got %v", err)
	}
}

// TestTaskNew_MissingEpic_NonInteractive pins the agent contract end-to-end:
// `task new` without --epic, run non-interactively (buffer stderr → gate closed),
// is exit-11 validation — never a prompt, never a hang.
func TestTaskNew_MissingEpic_NonInteractive(t *testing.T) {
	root := freshRepo(t)
	mustWrite(t, filepath.Join(root, "epics", "e1.md"), "---\nstatus: in-progress\n---\n")

	var out bytes.Buffer
	cmd := NewRootCmd(&out, &out)
	cmd.SetArgs([]string{"-C", root, "task", "new", "Some title", "--tags", "a"})
	err := cmd.Execute()
	if err == nil {
		t.Fatalf("missing --epic non-interactively should error; output:\n%s", out.String())
	}
	if !errors.Is(err, domain.ErrValidation) {
		t.Errorf("missing required input should wrap ErrValidation (exit 11), got %v", err)
	}
}

// TestBareTransition_NonInteractive pins that a bare transition verb (no task
// args) is exit-11 validation off a TTY — the picker is TTY-only, never a hang.
func TestBareTransition_NonInteractive(t *testing.T) {
	root := setupRepo(t) // alpha (ready-to-start), beta (in-progress)
	var out bytes.Buffer
	cmd := NewRootCmd(&out, &out)
	cmd.SetArgs([]string{"-C", root, "task", "start"}) // no task arg
	err := cmd.Execute()
	if err == nil {
		t.Fatalf("bare `task start` non-interactively should error; output:\n%s", out.String())
	}
	if !errors.Is(err, domain.ErrValidation) {
		t.Errorf("bare transition should wrap ErrValidation (exit 11), got %v", err)
	}
}
