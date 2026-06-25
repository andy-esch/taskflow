package cli

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/andy-esch/taskflow/internal/cli/prompt"
	"github.com/andy-esch/taskflow/internal/core"
	"github.com/andy-esch/taskflow/internal/domain"
	"github.com/andy-esch/taskflow/internal/store"
)

// TestTaskEdit_NonInteractive pins the agent contract: `task edit` off a TTY
// (buffer stderr → gate closed) is exit-11 validation pointing at `task set` —
// it never opens an editor and never hangs.
func TestTaskEdit_NonInteractive(t *testing.T) {
	root := setupRepo(t) // alpha (ready-to-start), beta (in-progress)
	var out bytes.Buffer
	cmd := NewRootCmd(strings.NewReader(""), &out, &out)
	cmd.SetArgs([]string{"-C", root, "task", "edit", "alpha"})
	err := cmd.Execute()
	if err == nil {
		t.Fatalf("`task edit` non-interactively should error; output:\n%s", out.String())
	}
	if !errors.Is(err, domain.ErrValidation) {
		t.Errorf("`task edit` off a TTY should wrap ErrValidation (exit 11), got %v", err)
	}
}

// TestTaskEditBare_NonInteractive: a bare `task edit` (no slug) off a TTY is
// exit-11 with a task-selection message — NOT a cobra "accepts 1 arg(s)" error.
// On a TTY it opens the picker instead, mirroring the bare transition verbs.
func TestTaskEditBare_NonInteractive(t *testing.T) {
	root := setupRepo(t)
	var out bytes.Buffer
	cmd := NewRootCmd(strings.NewReader(""), &out, &out)
	cmd.SetArgs([]string{"-C", root, "task", "edit"}) // no slug
	err := cmd.Execute()
	if !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("bare `task edit` off a TTY should wrap ErrValidation (exit 11), got %v", err)
	}
	if strings.Contains(err.Error(), "accepts") {
		t.Errorf("bare edit should offer the picker contract, not a cobra arg error: %v", err)
	}
}

// TestTaskEditBare_Picker: bare `task edit` on a TTY resolves a slug through the
// task picker (here a fake prompter), proving taskOptions feeds the chooser.
func TestTaskEditBare_Picker(t *testing.T) {
	root := setupRepo(t) // alpha (ready-to-start), beta (in-progress)
	f := &prompt.Fake{SelectAnswers: []string{"beta"}}
	app := &App{Svc: core.NewService(store.NewFS(root)), Gate: prompt.NewGate(true), Prompt: f}
	slug, err := app.fillSelect("", "specify a task to edit", "no tasks available to edit", "Task to edit", app.taskOptions)
	if err != nil || slug != "beta" {
		t.Fatalf("picker should resolve to beta, got %q %v", slug, err)
	}
	if len(f.Asked) != 1 || f.Asked[0] != "Task to edit" {
		t.Errorf("expected one prompt titled 'Task to edit', got %v", f.Asked)
	}
}

// TestEditViaEditor_RunsEditorOnTemp exercises the real exec glue (the gate test
// can't reach it): a fake editor that appends to the file it's handed proves the
// callback round-trips current content through a temp file and reads the result.
func TestEditViaEditor_RunsEditorOnTemp(t *testing.T) {
	dir := t.TempDir()
	script := filepath.Join(dir, "fake-editor.sh")
	// Appends a line to its argument ($1 is the temp file path we pass).
	if err := os.WriteFile(script, []byte("#!/bin/sh\nprintf 'appended\\n' >> \"$1\"\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	app := &App{In: strings.NewReader(""), Out: &bytes.Buffer{}, ErrOut: &bytes.Buffer{}}
	got, err := app.editViaEditor(script)("original\n", nil)
	if err != nil {
		t.Fatalf("editViaEditor: %v", err)
	}
	if want := "original\nappended\n"; got != want {
		t.Errorf("editor output = %q, want %q", got, want)
	}
}

// A non-existent editor surfaces as ErrValidation (exit 11), not a panic.
func TestEditViaEditor_BadEditor_ErrValidation(t *testing.T) {
	app := &App{In: strings.NewReader(""), Out: &bytes.Buffer{}, ErrOut: &bytes.Buffer{}}
	_, err := app.editViaEditor("definitely-not-a-real-editor-xyz")("x", nil)
	if !errors.Is(err, domain.ErrValidation) {
		t.Errorf("a missing editor should wrap ErrValidation, got %v", err)
	}
}
