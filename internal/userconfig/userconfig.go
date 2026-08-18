// Package userconfig loads the USER-scoped (home) configuration — the tier above
// the per-repo .tskflwctl.toml.
//
// It exists because [theme] and [pager] are documented in internal/config as
// "local-terminal concerns" yet could only ever be set in a REPO config, so a
// preference about your own terminal had to be repeated per project and a shared
// planning repo carried one contributor's taste. Those settings belong to a person,
// not a repo; this is where they live.
//
// This package deliberately knows NOTHING about planning repos, and internal/config
// deliberately does not import it. Home-scope data can therefore never influence
// where the planning root is discovered — an invariant the compiler enforces rather
// than one the reader has to remember.
package userconfig

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
)

const (
	// AppDir is the per-application subdirectory under the XDG config home.
	AppDir = "tskflwctl"

	// FileName is the hand-edited user config. It is READ but never written by the
	// tool, so comments and key order are the user's to keep. (Tool-owned home state
	// gets its own file rather than editing this one.)
	FileName = "config.toml"

	// DirEnv overrides the whole config directory. Not a nicety: the test suite is
	// t.TempDir()-disciplined and nothing in CI may read or write a real $HOME, so
	// every test binary that can reach this package pins it.
	DirEnv = "TSKFLW_CONFIG_HOME"
)

// Config is the resolved home-scope config. The zero value is meaningful and means
// "nothing set here" — every field falls through to the tier below it.
//
// Path records the file this was loaded from, and is empty when no file existed.
type Config struct {
	Path  string
	Pager PagerConfig
	Theme ThemeConfig
}

// PagerConfig is the `[pager]` table. Enabled is a pointer for the same reason it is
// one in internal/config: "unset" must stay distinct from an explicit
// `enabled = false`, which is what lets the tiers merge FIELD-BY-FIELD (a nil here
// defers to the next tier down) instead of a whole-table override.
type PagerConfig struct {
	Enabled *bool
	Command string
}

// ThemeConfig is the `[theme]` table: the color theme by name.
type ThemeConfig struct {
	Name string
}

// configFileTOML mirrors the on-disk schema for a real TOML decode.
type configFileTOML struct {
	Pager pagerFileTOML `toml:"pager"`
	Theme themeFileTOML `toml:"theme"`
}

type pagerFileTOML struct {
	Enabled *bool  `toml:"enabled"`
	Command string `toml:"command"`
}

type themeFileTOML struct {
	Name string `toml:"name"`
}

// Dir resolves the config directory: $TSKFLW_CONFIG_HOME, else
// $XDG_CONFIG_HOME/tskflwctl, else ~/.config/tskflwctl.
//
// Explicitly NOT os.UserConfigDir(), which returns ~/Library/Application Support on
// darwin — wrong for a dotfile-friendly CLI whose users commit their config.
func Dir() (string, error) {
	if v := strings.TrimSpace(os.Getenv(DirEnv)); v != "" {
		return v, nil
	}
	if v := strings.TrimSpace(os.Getenv("XDG_CONFIG_HOME")); v != "" {
		return filepath.Join(v, AppDir), nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("locate home dir: %w", err)
	}
	return filepath.Join(home, ".config", AppDir), nil
}

// Load reads the user config. It ALWAYS returns a usable *Config — never nil — so
// callers can use the result unconditionally.
//
// A missing file is not an error: the empty config means "nothing set here". A
// malformed file returns the empty config AND an error, which the caller is expected
// to surface as a warning and continue past. This is the deliberate opposite of the
// repo config, where malformed TOML is fatal: that file is the MARKER that anchors
// discovery, so guessing there would fork the data — this one is preferences, and a
// stray typo must not lock you out of every command on the machine.
func Load() (*Config, error) {
	dir, err := Dir()
	if err != nil {
		return &Config{}, err
	}
	path := filepath.Join(dir, FileName)
	var cf configFileTOML
	if _, err := toml.DecodeFile(path, &cf); err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return &Config{}, nil // no user config is the normal case
		}
		return &Config{}, fmt.Errorf("read %s: %w", path, err)
	}
	return &Config{
		Path:  path,
		Pager: PagerConfig{Enabled: cf.Pager.Enabled, Command: cf.Pager.Command},
		Theme: ThemeConfig{Name: cf.Theme.Name},
	}, nil
}
