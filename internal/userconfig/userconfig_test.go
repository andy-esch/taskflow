package userconfig

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// write drops a config.toml into a fresh dir and points DirEnv at it.
func write(t *testing.T, body string) string {
	t.Helper()
	dir := t.TempDir()
	if body != "" {
		if err := os.WriteFile(filepath.Join(dir, FileName), []byte(body), 0o644); err != nil {
			t.Fatalf("write config: %v", err)
		}
	}
	t.Setenv(DirEnv, dir)
	return dir
}

// TestDir_Precedence pins the location contract: our env override first, then XDG,
// then ~/.config — and never os.UserConfigDir(), which would put this under
// ~/Library/Application Support on darwin.
func TestDir_Precedence(t *testing.T) {
	t.Run("DirEnv wins", func(t *testing.T) {
		t.Setenv(DirEnv, "/explicit/dir")
		t.Setenv("XDG_CONFIG_HOME", "/xdg")
		got, err := Dir()
		if err != nil || got != "/explicit/dir" {
			t.Errorf("Dir() = %q, %v; want /explicit/dir", got, err)
		}
	})
	t.Run("XDG_CONFIG_HOME when DirEnv unset", func(t *testing.T) {
		t.Setenv(DirEnv, "")
		t.Setenv("XDG_CONFIG_HOME", "/xdg")
		got, err := Dir()
		want := filepath.Join("/xdg", AppDir)
		if err != nil || got != want {
			t.Errorf("Dir() = %q, %v; want %q", got, err, want)
		}
	})
	t.Run("relative DirEnv is made absolute", func(t *testing.T) {
		t.Setenv(DirEnv, "relcfg")
		got, err := Dir()
		if err != nil {
			t.Fatalf("Dir: %v", err)
		}
		if !filepath.IsAbs(got) {
			t.Errorf("Dir() = %q, want an absolute path — a cwd-relative config dir makes the same environment behave differently per directory", got)
		}
		wd, _ := os.Getwd()
		if want := filepath.Join(wd, "relcfg"); got != want {
			t.Errorf("Dir() = %q, want %q", got, want)
		}
	})
	t.Run("relative XDG_CONFIG_HOME is IGNORED per the XDG spec", func(t *testing.T) {
		home := t.TempDir()
		t.Setenv(DirEnv, "")
		t.Setenv("XDG_CONFIG_HOME", "relxdg")
		t.Setenv("HOME", home)
		got, err := Dir()
		want := filepath.Join(home, ".config", AppDir)
		if err != nil || got != want {
			t.Errorf("Dir() = %q, %v; want the ~/.config fallback %q (the spec says a non-absolute XDG_CONFIG_HOME must be ignored, not repaired)", got, err, want)
		}
	})
	t.Run("~/.config fallback", func(t *testing.T) {
		home := t.TempDir()
		t.Setenv(DirEnv, "")
		t.Setenv("XDG_CONFIG_HOME", "")
		t.Setenv("HOME", home)
		got, err := Dir()
		want := filepath.Join(home, ".config", AppDir)
		if err != nil || got != want {
			t.Errorf("Dir() = %q, %v; want %q", got, err, want)
		}
	})
}

// TestLoad_Missing: no user config is the NORMAL case, not an error — and the
// returned config must still be usable, since callers dereference it unconditionally.
func TestLoad_Missing(t *testing.T) {
	write(t, "")
	cfg, err := Load()
	if err != nil {
		t.Fatalf("missing file should not error, got %v", err)
	}
	if cfg == nil {
		t.Fatal("Load must never return a nil config")
	}
	if cfg.Path != "" || cfg.Theme.Name != "" || cfg.Pager.Enabled != nil || cfg.Pager.Command != "" {
		t.Errorf("missing file should yield the zero config, got %+v", cfg)
	}
}

// TestLoad_Values reads every field the home tier owns today.
func TestLoad_Values(t *testing.T) {
	dir := write(t, "[theme]\nname = \"neon\"\n\n[pager]\nenabled = true\ncommand = \"delta\"\n")
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.Theme.Name != "neon" {
		t.Errorf("theme = %q, want neon", cfg.Theme.Name)
	}
	if cfg.Pager.Enabled == nil || !*cfg.Pager.Enabled {
		t.Errorf("pager.enabled = %v, want true", cfg.Pager.Enabled)
	}
	if cfg.Pager.Command != "delta" {
		t.Errorf("pager.command = %q, want delta", cfg.Pager.Command)
	}
	if want := filepath.Join(dir, FileName); cfg.Path != want {
		t.Errorf("Path = %q, want %q", cfg.Path, want)
	}
}

// TestLoad_EnabledFalseIsDistinctFromUnset is the whole reason Enabled is a *bool:
// the tiers merge field-by-field, so "unset here" (nil, defer down) must never be
// confused with an explicit "off".
func TestLoad_EnabledFalseIsDistinctFromUnset(t *testing.T) {
	write(t, "[pager]\nenabled = false\n")
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.Pager.Enabled == nil {
		t.Fatal("explicit `enabled = false` must not read as unset")
	}
	if *cfg.Pager.Enabled {
		t.Error("enabled = false should decode as false")
	}

	write(t, "[pager]\ncommand = \"less\"\n")
	cfg, err = Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.Pager.Enabled != nil {
		t.Error("an absent `enabled` must stay nil so it defers to the tier below")
	}
}

// TestLoad_Malformed: a typo in a PREFERENCES file must not lock you out of every
// command on the machine — the deliberate opposite of the repo config, where a bad
// marker is fatal because guessing there would fork the data. Errors, but still
// hands back a usable empty config.
func TestLoad_Malformed(t *testing.T) {
	write(t, "[theme\nname = ")
	cfg, err := Load()
	if err == nil {
		t.Error("malformed TOML should report an error for the caller to warn with")
	}
	if cfg == nil {
		t.Fatal("Load must never return a nil config, even on error")
	}
	if cfg.Theme.Name != "" || cfg.Pager.Enabled != nil {
		t.Errorf("malformed file should degrade to the zero config, got %+v", cfg)
	}
}

// TestLoad_NoHomeDirIsSilent pins the M1 contract: an unresolvable home directory
// (minimal containers, daemons, some CI) is not a problem worth reporting — there is
// no preferences file to miss, so the empty config is the complete answer. Reporting
// it would put a ⚠ on every command in those environments, which the user cannot act
// on. A file that EXISTS but is broken still errors (TestLoad_Malformed).
func TestLoad_NoHomeDirIsSilent(t *testing.T) {
	t.Setenv(DirEnv, "")
	t.Setenv("XDG_CONFIG_HOME", "")
	t.Setenv("HOME", "")
	if _, err := Dir(); err == nil {
		t.Skip("this platform still resolves a home dir without $HOME; nothing to assert")
	}
	cfg, err := Load()
	if err != nil {
		t.Errorf("an unresolvable home dir must degrade silently, got error: %v", err)
	}
	if cfg == nil {
		t.Fatal("Load must never return a nil config")
	}
	if cfg.Path != "" || cfg.Theme.Name != "" {
		t.Errorf("want the zero config, got %+v", cfg)
	}
}

// TestLoad_BrokenSymlinkIsReported separates the two things ENOENT can mean. A
// missing file is silence; a symlink whose target is gone is a real problem for this
// audience specifically — people who commit their config typically symlink it out of
// a dotfiles repo that may not be mounted, and silence there looks like "my settings
// stopped working for no reason".
func TestLoad_BrokenSymlinkIsReported(t *testing.T) {
	dir := t.TempDir()
	link := filepath.Join(dir, FileName)
	if err := os.Symlink(filepath.Join(dir, "nowhere", "config.toml"), link); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}
	t.Setenv(DirEnv, dir)

	cfg, err := Load()
	if err == nil {
		t.Fatal("a dangling config symlink must be reported, not treated as an absent file")
	}
	if !strings.Contains(err.Error(), "broken symlink") {
		t.Errorf("error should name the cause, got %v", err)
	}
	if cfg == nil {
		t.Fatal("Load must never return a nil config, even on error")
	}
}

// TestGuardTestIsolation_FiresWithoutIsolation pins the M5 fix: pinning DirEnv in one
// package's TestMain protects only that package, so the real-home fallback refuses to
// run under `go test` at all. Without this, a future package that reaches userconfig
// would silently read the developer's own ~/.config and pass locally while behaving
// differently elsewhere — a failure with no symptom.
func TestGuardTestIsolation_FiresWithoutIsolation(t *testing.T) {
	t.Setenv(DirEnv, "")
	t.Setenv("XDG_CONFIG_HOME", "")
	// HOME deliberately NOT touched: this is exactly what an unisolated package looks
	// like from in here.
	if os.Getenv("HOME") != startupHome {
		t.Skip("HOME already differs from process start; the guard cannot be exercised")
	}
	if _, err := Dir(); err == nil {
		t.Fatal("Dir() must refuse the real-home fallback under test")
	} else if !strings.Contains(err.Error(), DirEnv) {
		t.Errorf("the error must name the env var that fixes it, got %v", err)
	}
	// Load still degrades quietly — a guard failure is a Dir() error like any other,
	// and Load's contract is that an unresolvable dir yields the empty config.
	cfg, err := Load()
	if err != nil || cfg == nil || cfg.Theme.Name != "" {
		t.Errorf("Load should stay silent and empty, got cfg=%+v err=%v", cfg, err)
	}
}
