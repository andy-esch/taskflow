package editor

import (
	"reflect"
	"testing"
)

// TestResolve_Precedence: $VISUAL wins, then $EDITOR, then vi as the fallback.
func TestResolve_Precedence(t *testing.T) {
	t.Setenv("VISUAL", "myvisual")
	t.Setenv("EDITOR", "myeditor")
	if got := Resolve(); got != "myvisual" {
		t.Errorf("with $VISUAL set, Resolve() = %q, want %q", got, "myvisual")
	}

	t.Setenv("VISUAL", "  ") // blank/whitespace is treated as unset
	if got := Resolve(); got != "myeditor" {
		t.Errorf("with $VISUAL blank, Resolve() = %q, want %q", got, "myeditor")
	}

	t.Setenv("EDITOR", "")
	if got := Resolve(); got != "vi" {
		t.Errorf("with neither set, Resolve() = %q, want %q", got, "vi")
	}
}

// TestCommand_SplitsArgs: a multi-word editor keeps its flags, and the path is
// appended last.
func TestCommand_SplitsArgs(t *testing.T) {
	cmd := Command("code -w", "/tmp/x.md")
	want := []string{"code", "-w", "/tmp/x.md"}
	if !reflect.DeepEqual(cmd.Args, want) {
		t.Errorf("Command args = %v, want %v", cmd.Args, want)
	}

	cmd = Command("vi", "/tmp/y.md")
	if want := []string{"vi", "/tmp/y.md"}; !reflect.DeepEqual(cmd.Args, want) {
		t.Errorf("single-word editor args = %v, want %v", cmd.Args, want)
	}
}

// TestCommand_BlankFallsBackToResolve: a blank program can't yield an empty
// command — it falls back to Resolve (vi here, with no env set).
func TestCommand_BlankFallsBackToResolve(t *testing.T) {
	t.Setenv("VISUAL", "")
	t.Setenv("EDITOR", "")
	cmd := Command("   ", "/tmp/z.md")
	if want := []string{"vi", "/tmp/z.md"}; !reflect.DeepEqual(cmd.Args, want) {
		t.Errorf("blank editor args = %v, want %v", cmd.Args, want)
	}
}
