package cli

import (
	"strings"
	"testing"
)

func TestColor_AlwaysEmitsANSI(t *testing.T) {
	root := setupRepo(t)
	out := runRoot(t, "-C", root, "--color", "always", "task", "list")
	if !strings.Contains(out, "\x1b[") {
		t.Errorf("--color=always should emit ANSI, got:\n%q", out)
	}
}

func TestColor_DefaultIsPlainForNonTTY(t *testing.T) {
	root := setupRepo(t)
	// runRoot writes to a bytes.Buffer (not a TTY) → auto resolves to plain.
	out := runRoot(t, "-C", root, "task", "list")
	if strings.Contains(out, "\x1b[") {
		t.Errorf("non-TTY output must be plain, got ANSI:\n%q", out)
	}
}

func TestColor_NeverIsPlain(t *testing.T) {
	root := setupRepo(t)
	out := runRoot(t, "-C", root, "--color", "never", "task", "list")
	if strings.Contains(out, "\x1b[") {
		t.Errorf("--color=never must be plain, got ANSI:\n%q", out)
	}
}
