package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/andy-esch/taskflow/internal/userconfig"
)

// homeConfig writes a user config into a fresh dir and points the location env at
// it, overriding the empty dir TestMain pinned for this package.
func homeConfig(t *testing.T, body string) {
	t.Helper()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, userconfig.FileName), []byte(body), 0o644); err != nil {
		t.Fatalf("write user config: %v", err)
	}
	t.Setenv(userconfig.DirEnv, dir)
}

// TestUserConfig_LoadsInSetStyle pins WHERE the home config is read. It must land in
// setStyle, not resolve(): `schema` (like init/doctor/completion) never discovers a
// planning repo, yet must still honor a user's theme. A regression that moved this
// into resolve() would leave a.User nil here.
func TestUserConfig_LoadsInSetStyle(t *testing.T) {
	homeConfig(t, "[theme]\nname = \"neon\"\n[pager]\ncommand = \"delta\"\n")

	var out bytes.Buffer
	app := &App{Out: &out, ErrOut: &out, In: strings.NewReader("")}
	app.setStyle()

	if app.User == nil {
		t.Fatal("setStyle must populate App.User")
	}
	if app.User.Theme.Name != "neon" {
		t.Errorf("theme = %q, want neon", app.User.Theme.Name)
	}
	if got := app.pagerProgram(); got != "delta" && os.Getenv("TSKFLW_PAGER") == "" {
		t.Errorf("pagerProgram = %q, want delta from the home config", got)
	}
}

// TestUserConfig_MalformedWarnsButRuns is the resilience contract: a typo in a
// PREFERENCES file must not break every command on the machine. It warns on stderr
// and the command still succeeds — unlike the repo config, where a bad marker is
// fatal because guessing there would fork the data.
func TestUserConfig_MalformedWarnsButRuns(t *testing.T) {
	homeConfig(t, "[theme\nname = ")

	var out bytes.Buffer
	cmd := NewRootCmd(strings.NewReader(""), &out, &out)
	cmd.SetArgs([]string{"version"})
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("a malformed user config must not fail the command, got %v\noutput:\n%s", err, out.String())
	}
	if !strings.Contains(out.String(), "ignoring user config") {
		t.Errorf("want a warning naming the ignored config, got:\n%s", out.String())
	}
}

// TestUserConfig_AbsentIsSilent: no user config is the normal case for most users, so
// it must produce no warning and no behavior change at all.
func TestUserConfig_AbsentIsSilent(t *testing.T) {
	t.Setenv(userconfig.DirEnv, t.TempDir()) // exists, but holds no config

	var out bytes.Buffer
	app := &App{Out: &out, ErrOut: &out, In: strings.NewReader("")}
	app.setStyle()

	if app.User == nil {
		t.Fatal("App.User must be usable even with no file on disk")
	}
	if out.String() != "" {
		t.Errorf("an absent user config must be silent, got %q", out.String())
	}
}

// TestUserConfig_SilentOnCompletionPath pins the M2 fix. Completion output is
// consumed by the shell, so a stray ⚠ is noise at best and corrupts the display at
// worst — the same rule warnUnknownTheme already followed and the user-config warning
// originally ignored.
func TestUserConfig_SilentOnCompletionPath(t *testing.T) {
	homeConfig(t, "[theme\nname = ") // malformed: would warn on a normal run

	var out, errOut bytes.Buffer
	cmd := NewRootCmd(strings.NewReader(""), &out, &errOut)
	cmd.SetArgs([]string{"__complete", "task", ""})
	cmd.SetOut(&out)
	cmd.SetErr(&errOut)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("completion must not fail: %v", err)
	}
	if strings.Contains(errOut.String(), "ignoring user config") {
		t.Errorf("no warning may reach the completion path, got:\n%s", errOut.String())
	}
}

// TestWarnPresentation_ReachesStyleOnlyCommands pins the L2 fix. version / init /
// schema / doctor / theme each override PersistentPreRunE to run without a planning
// repo; before the shared seam, all but `theme` silently dropped the theme warning,
// so an unrecognized name in a user config went unreported on exactly the commands a
// new user runs first.
func TestWarnPresentation_ReachesStyleOnlyCommands(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, userconfig.FileName),
		[]byte("[theme]\nname = \"nope-not-a-theme\"\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	for _, args := range [][]string{
		{"version"},
		{"schema"},
		{"theme", "list"},
		{"init", "--dry-run", "--path", filepath.Join(dir, "probe")},
	} {
		t.Setenv(userconfig.DirEnv, dir)
		var out, errOut bytes.Buffer
		cmd := NewRootCmd(strings.NewReader(""), &out, &errOut)
		cmd.SetArgs(args)
		cmd.SetOut(&out)
		cmd.SetErr(&errOut)
		if err := cmd.Execute(); err != nil {
			t.Fatalf("%v: %v\n%s", args, err, out.String())
		}
		if !strings.Contains(errOut.String(), "unknown theme") {
			t.Errorf("%v: an unrecognized theme must be reported here too, got stderr:\n%s", args, errOut.String())
		}
	}
}
