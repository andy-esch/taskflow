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

// TestThemeName_Precedence pins the theme-selection contract — flag > env > config,
// trimmed, "" when none set — which is the heart of T5. Pin it directly: observing
// it via the resolved Theme is impossible while only one theme is registered (every
// name Lookups to the default).
func TestThemeName_Precedence(t *testing.T) {
	cases := []struct {
		name           string
		flag, env, cfg string
		want           string
	}{
		{"flag wins over env+config", "flagt", "envt", "cfgt", "flagt"},
		{"env over config", "", "envt", "cfgt", "envt"},
		{"config when no flag/env", "", "", "cfgt", "cfgt"},
		{"none → empty (default downstream)", "", "", "", ""},
		{"blank flag falls through to config", "   ", "", "cfgt", "cfgt"},
		{"value is trimmed", " neon ", "", "", "neon"},
	}
	for _, c := range cases {
		if got := themeName(c.flag, c.env, c.cfg); got != c.want {
			t.Errorf("%s: themeName(%q,%q,%q) = %q, want %q", c.name, c.flag, c.env, c.cfg, got, c.want)
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
