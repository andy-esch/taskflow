package cli

import (
	"path/filepath"
	"strings"
	"testing"
)

// TestActiveHelp_NewTitle: `task new` with no positional yet surfaces an
// ActiveHelp hint (and suppresses misleading filename completion).
func TestActiveHelp_NewTitle(t *testing.T) {
	root := setupRepo(t)
	out := runRoot(t, "__complete", "-C", root, "task", "new", "")
	if !strings.Contains(out, "_activeHelp_") || !strings.Contains(out, "title") {
		t.Errorf("expected an active-help title hint for `task new`:\n%s", out)
	}
}

// TestActiveHelp_MoveStatus: once a task is named, `task move` hints that the
// trailing argument is the target status.
func TestActiveHelp_MoveStatus(t *testing.T) {
	root := setupRepo(t)
	out := runRoot(t, "__complete", "-C", root, "task", "move", "alpha", "")
	if !strings.Contains(out, "_activeHelp_") || !strings.Contains(out, "status") {
		t.Errorf("expected an active-help status hint for `task move`:\n%s", out)
	}
}

// TestCreate_LinksPathOnColor pins the OSC 8 contract: the create confirmation
// path is a clickable file:// link on a TTY, plain text off it, and never in the
// JSON envelope.
func TestCreate_LinksPathOnColor(t *testing.T) {
	root := freshRepo(t)
	mustWrite(t, filepath.Join(root, "epics", "e1.md"), "---\nstatus: in-progress\n---\n# E1\n")

	on := runRoot(t, "-C", root, "task", "new", "Linky", "--epic", "e1", "--tags", "a", "--color=always")
	if !strings.Contains(on, "\x1b]8;;file://") {
		t.Errorf("created path should be an OSC 8 file:// link under --color=always:\n%q", on)
	}
	off := runRoot(t, "-C", root, "task", "new", "Linky2", "--epic", "e1", "--tags", "a", "--color=never")
	if strings.Contains(off, "\x1b]8;;") {
		t.Errorf("created path must be plain under --color=never:\n%q", off)
	}
	js := runRoot(t, "-C", root, "task", "new", "Linky3", "--epic", "e1", "--tags", "a", "--json", "--color=always")
	if strings.Contains(js, "\x1b]8;;") {
		t.Errorf("--json envelope must never contain OSC 8:\n%q", js)
	}
}
