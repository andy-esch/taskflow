package userconfig

import (
	"os"
	"path/filepath"
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
