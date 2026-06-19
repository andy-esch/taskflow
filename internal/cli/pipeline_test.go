package cli

import (
	"bytes"
	"strings"
	"testing"
)

func TestTaskList_Quiet(t *testing.T) {
	root := setupRepo(t)
	out := runRoot(t, "-C", root, "task", "list", "-q")
	if !strings.Contains(out, "alpha") || !strings.Contains(out, "beta") {
		t.Errorf("-q should list both slugs:\n%q", out)
	}
	// Every line is a bare id — no header, no whitespace/decoration (xargs-safe).
	for _, l := range strings.Split(strings.TrimSpace(out), "\n") {
		if l == "" || strings.ContainsAny(l, " \t") {
			t.Errorf("-q line should be a bare id, got %q", l)
		}
	}
}

func TestTaskList_Plain(t *testing.T) {
	root := setupRepo(t)
	out := runRoot(t, "-C", root, "task", "list", "--plain")
	lines := strings.Split(strings.TrimSpace(out), "\n")
	if len(lines) < 2 {
		t.Fatalf("--plain needs a header + ≥1 row:\n%q", out)
	}
	if lines[0] != "slug\tstatus\ttier\tpriority\tepic\tupdated\tdescription" {
		t.Errorf("--plain header wrong: %q", lines[0])
	}
	if cols := strings.Split(lines[1], "\t"); len(cols) != 7 {
		t.Errorf("--plain row should have 7 tab-separated columns, got %d: %q", len(cols), lines[1])
	}
}

// TestPlain_ByteStableUnderColor pins the --plain contract: it ignores styling
// entirely, so it's byte-identical whether color is forced on or off (a script
// can rely on it). The default human table, by contrast, would emit ANSI here.
func TestPlain_ByteStableUnderColor(t *testing.T) {
	root := setupRepo(t)
	on := runRoot(t, "-C", root, "task", "list", "--plain", "--color=always")
	off := runRoot(t, "-C", root, "task", "list", "--plain", "--color=never")
	if strings.Contains(on, "\x1b[") {
		t.Errorf("--plain must carry no ANSI even with --color=always:\n%q", on)
	}
	if on != off {
		t.Errorf("--plain must be byte-stable across color settings:\n on:  %q\n off: %q", on, off)
	}
}

func TestAuditList_Plain(t *testing.T) {
	root := setupAuditRepo(t)
	out := runRoot(t, "-C", root, "audit", "list", "--plain")
	if h := strings.SplitN(out, "\n", 2)[0]; h != "slug\tbucket\tarea\tdate\tfindings\topen" {
		t.Errorf("audit --plain header wrong: %q", h)
	}
}

func TestTransition_FailureToStderr(t *testing.T) {
	root := setupRepo(t)
	var out, errOut bytes.Buffer
	cmd := NewRootCmd(&out, &errOut)
	// alpha (ready-to-start) starts ok; ghost is not found → a partial failure.
	cmd.SetArgs([]string{"-C", root, "task", "start", "alpha", "ghost"})
	_ = cmd.Execute()
	if !strings.Contains(out.String(), "alpha -> in-progress") {
		t.Errorf("success confirmation should be on stdout:\n%q", out.String())
	}
	if strings.Contains(out.String(), "ghost") {
		t.Errorf("failure must not be on stdout (the data stream):\n%q", out.String())
	}
	if !strings.Contains(errOut.String(), "ghost") {
		t.Errorf("failure should be on stderr:\n%q", errOut.String())
	}
}

func TestList_ModeConflicts(t *testing.T) {
	root := setupRepo(t)
	// --json with -q/--plain is rejected (validation); -q with --plain is a cobra
	// mutual-exclusion error. None may produce data.
	for _, args := range [][]string{
		{"task", "list", "--json", "--plain"},
		{"task", "list", "--json", "-q"},
		{"task", "list", "-q", "--plain"},
	} {
		var out bytes.Buffer
		cmd := NewRootCmd(&out, &out)
		cmd.SetArgs(append([]string{"-C", root}, args...))
		if err := cmd.Execute(); err == nil {
			t.Errorf("expected an error for %v", args)
		}
	}
}
