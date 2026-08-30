package cli

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/andy-esch/taskflow/internal/cli/render"
	"github.com/andy-esch/taskflow/internal/design"
	"github.com/andy-esch/taskflow/internal/domain"
)

// TestResolveVariant pins the `theme preview --variant` contract: dark/light are
// explicit, "auto" detects only when allowed (else deterministic dark), and an
// unknown value is an ErrValidation (exit 11). detectDark is stubbed so the test
// never depends on a real terminal.
func TestResolveVariant(t *testing.T) {
	detectLight := func() bool { return false } // stand-in for a light terminal
	cases := []struct {
		name        string
		flag        string
		allowDetect bool
		wantDark    bool
		wantErr     bool
	}{
		{"explicit dark", "dark", true, true, false},
		{"explicit light", "light", true, false, false},
		{"auto detects when allowed", "auto", true, false, false}, // detect → light
		{"auto is deterministic dark when not allowed", "auto", false, true, false},
		{"unknown value errors", "lite", true, false, true},
	}
	for _, c := range cases {
		dark, err := resolveVariant(c.flag, c.allowDetect, detectLight)
		if (err != nil) != c.wantErr {
			t.Errorf("%s: err=%v, wantErr=%v", c.name, err, c.wantErr)
		}
		if err == nil && dark != c.wantDark {
			t.Errorf("%s: dark=%v, want %v", c.name, dark, c.wantDark)
		}
	}
	if _, err := resolveVariant("nope", true, detectLight); !errors.Is(err, domain.ErrValidation) {
		t.Errorf("unknown variant should wrap ErrValidation (exit 11), got %v", err)
	}
}

// TestThemeName_Precedence pins the theme-selection contract — flag > env > repo
// config > home config, trimmed, "" when none set — which is the heart of T5. Pin it
// directly: observing it via the resolved Theme is impossible while only one theme is
// registered (every name Lookups to the default). All four tiers are covered,
// including the repo-beats-home rule (a project can pin a theme for everyone working
// in it; the home tier is what a person sets once for their own terminal).
func TestThemeName_Precedence(t *testing.T) {
	cases := []struct {
		name                 string
		flag, env, cfg, user string
		want                 string
	}{
		{"flag wins over all", "flagt", "envt", "cfgt", "usert", "flagt"},
		{"env over repo+home", "", "envt", "cfgt", "usert", "envt"},
		{"repo config over home", "", "", "cfgt", "usert", "cfgt"},
		{"home when nothing above it", "", "", "", "usert", "usert"},
		{"none → empty (default downstream)", "", "", "", "", ""},
		{"blank flag falls through to repo config", "   ", "", "cfgt", "", "cfgt"},
		{"blank repo config falls through to home", "", "", "   ", "usert", "usert"},
		{"value is trimmed", " neon ", "", "", "", "neon"},
		{"home value is trimmed", "", "", "", " neon ", "neon"},
	}
	for _, c := range cases {
		if got := themeName(c.flag, c.env, c.cfg, c.user); got != c.want {
			t.Errorf("%s: themeName(%q,%q,%q,%q) = %q, want %q", c.name, c.flag, c.env, c.cfg, c.user, got, c.want)
		}
	}
}

// TestWarnUnknownTheme: an explicitly-set unrecognized name warns to stderr (so a
// "none"/typo isn't a silent neon fall-back); empty / "auto" / a real theme don't.
func TestWarnUnknownTheme(t *testing.T) {
	t.Setenv("TSKFLW_THEME", "") // isolate from the ambient env
	warn := func(flag string) string {
		var buf bytes.Buffer
		a := &App{ErrOut: &buf, Theme: flag, Style: render.NewStyle(false), Th: design.Default()}
		a.warnUnknownTheme()
		return buf.String()
	}
	if out := warn("none"); !strings.Contains(out, "unknown theme") || !strings.Contains(out, "none") {
		t.Errorf("name=none: want a warning naming it, got %q", out)
	}
	for _, name := range []string{"", "auto", "neon"} {
		if out := warn(name); out != "" {
			t.Errorf("name=%q: want no warning, got %q", name, out)
		}
	}
}

// TestThemeEntries: `theme list`'s rows — every registered theme, sorted, with the
// default and the active one flagged.
func TestThemeEntries(t *testing.T) {
	got := themeEntries("catppuccin")
	if len(got) != 2 || got[0].Name != "catppuccin" || got[1].Name != "neon" {
		t.Fatalf("themeEntries = %+v, want [catppuccin, neon] (sorted)", got)
	}
	if !got[0].Active || got[1].Active {
		t.Errorf("active flags wrong: catppuccin should be active, neon not: %+v", got)
	}
	if got[0].Default || !got[1].Default {
		t.Errorf("default flags wrong: neon is the default, catppuccin is not: %+v", got)
	}
}

// TestThemeFlagFrom covers the raw-arg scan chrome uses before cobra parses
// anything. fang's help renders from cobra's help path, which returns before
// PersistentPreRunE, so --theme has to be read straight off os.Args or styled
// help silently ignores it.
func TestThemeFlagFrom(t *testing.T) {
	for _, tc := range []struct {
		name string
		args []string
		want string
	}{
		{"absent", []string{"task", "--help"}, ""},
		{"space form", []string{"--theme", "catppuccin", "task"}, "catppuccin"},
		{"equals form", []string{"--theme=catppuccin", "task"}, "catppuccin"},
		{"after a subcommand", []string{"task", "list", "--theme", "neon"}, "neon"},
		{"trailing --theme with no value", []string{"task", "--theme"}, ""},
		{"bare --theme followed by another flag", []string{"--theme", "--json"}, ""},
		{"bare --theme followed by a short flag", []string{"--theme", "-C", "/tmp"}, ""},
		{"literal --theme after -- is not the flag", []string{"--", "--theme", "neon"}, ""},
		{"empty args", nil, ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := themeFlagFrom(tc.args); got != tc.want {
				t.Errorf("themeFlagFrom(%q) = %q, want %q", tc.args, got, tc.want)
			}
		})
	}
}

// TestChromeThemeHonorsFlagAndEnv pins the two precedence tiers chrome can
// resolve without discovering a planning repo, and that it degrades to the
// default rather than failing when nothing selects a theme.
func TestChromeThemeHonorsFlagAndEnv(t *testing.T) {
	t.Setenv("TSKFLW_THEME", "")
	// A directory with no config and no parent config, so repo discovery is a
	// miss and only the flag/env tiers are in play.
	t.Chdir(t.TempDir())

	if got := ChromeTheme([]string{"--theme", "catppuccin"}).Name; got != "catppuccin" {
		t.Errorf("flag tier: got %q, want catppuccin", got)
	}
	t.Setenv("TSKFLW_THEME", "catppuccin")
	if got := ChromeTheme(nil).Name; got != "catppuccin" {
		t.Errorf("env tier: got %q, want catppuccin", got)
	}
	if got := ChromeTheme([]string{"--theme", "neon"}).Name; got != "neon" {
		t.Errorf("flag must beat env: got %q, want neon", got)
	}
	t.Setenv("TSKFLW_THEME", "")
	if got := ChromeTheme(nil).Name; got != design.Default().Name {
		t.Errorf("no selection should fall back to the default, got %q", got)
	}
	// An unregistered name must not leave chrome unpainted.
	if got := ChromeTheme([]string{"--theme", "no-such-theme"}).Name; got != design.Default().Name {
		t.Errorf("unknown theme should degrade to the default, got %q", got)
	}
}
