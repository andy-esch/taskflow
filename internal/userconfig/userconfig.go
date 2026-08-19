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
	"testing"

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
// $XDG_CONFIG_HOME/tskflwctl, else ~/.config/tskflwctl. The result is always
// ABSOLUTE — a cwd-relative config dir would make the same environment behave
// differently depending on where you happened to run from, and would put a useless
// relative path in any diagnostic.
//
// The two env vars degrade differently on a relative value, on purpose. DirEnv is
// OURS, so a relative value is resolved against the cwd (a convenience for
// `TSKFLW_CONFIG_HOME=./fixtures` in a script). XDG_CONFIG_HOME is governed by the
// XDG Base Directory spec, which says a value that is not absolute "must be ignored"
// — so it is, falling through to ~/.config rather than being silently repaired into
// something the spec says we should not have used.
//
// Explicitly NOT os.UserConfigDir(), which returns ~/Library/Application Support on
// darwin — wrong for a dotfile-friendly CLI whose users commit their config.
func Dir() (string, error) {
	if v := strings.TrimSpace(os.Getenv(DirEnv)); v != "" {
		abs, err := filepath.Abs(v)
		if err != nil {
			return "", fmt.Errorf("resolve %s=%q: %w", DirEnv, v, err)
		}
		return abs, nil
	}
	if v := strings.TrimSpace(os.Getenv("XDG_CONFIG_HOME")); filepath.IsAbs(v) {
		return filepath.Join(v, AppDir), nil
	}
	if err := guardTestIsolation(); err != nil {
		return "", err
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("locate home dir: %w", err)
	}
	return filepath.Join(home, ".config", AppDir), nil
}

// startupHome is $HOME as it was when the process began, captured before any test
// can rewrite it. Not DI state — a one-word snapshot, read only by
// guardTestIsolation.
var startupHome = os.Getenv("HOME")

// guardTestIsolation refuses the REAL-home fallback inside a test binary.
//
// Pinning $TSKFLW_CONFIG_HOME in a package's TestMain only protects that package, so
// a new package that reaches this code would silently read the developer's own
// ~/.config/tskflwctl/config.toml and pass locally while behaving differently on
// another machine — a failure with no symptom, which is the worst kind
// (audit 2026-08-18-multi-space-config-foundation, M5).
//
// A test that has isolated itself has necessarily changed something: it set DirEnv or
// XDG_CONFIG_HOME (handled by the callers above), or it pointed HOME at a t.TempDir.
// Reaching here with HOME untouched means no isolation at all, so it errors instead —
// loudly, and only ever under `go test`.
func guardTestIsolation() error {
	if !testing.Testing() || os.Getenv("HOME") != startupHome {
		return nil
	}
	return fmt.Errorf(
		"userconfig: refusing to read the real home dir under test — set %s (see TestMain in internal/cli) or override HOME",
		DirEnv)
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
		// No resolvable home directory (a minimal container, a daemon, some CI) is
		// NOT a problem to report: there is no preferences file to miss, so the empty
		// config is the correct and complete answer. Reporting it would put a ⚠ on
		// every single command in those environments — noise the user cannot act on.
		// Contrast a file that EXISTS but is broken, below, which is worth saying.
		return &Config{}, nil
	}
	path := filepath.Join(dir, FileName)
	var cf configFileTOML
	if _, err := toml.DecodeFile(path, &cf); err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			// ENOENT is ambiguous: no file at all (the normal case, silent), or a
			// SYMLINK whose target is gone — which is a real, actionable problem for
			// exactly this audience, since people who commit their config typically
			// symlink it out of a dotfiles repo that may not be mounted. Lstat sees the
			// link itself, so it separates the two. Done only on the error path, so the
			// common case still costs no extra syscall.
			if fi, lerr := os.Lstat(path); lerr == nil && fi.Mode()&os.ModeSymlink != 0 {
				target, _ := os.Readlink(path)
				return &Config{}, fmt.Errorf("read %s: broken symlink -> %s (target does not exist)", path, target)
			}
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
