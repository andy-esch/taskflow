package design

import (
	"fmt"
	"math"
	"testing"

	"github.com/andy-esch/taskflow/internal/theme"
)

// relLuminance is the WCAG 2.1 relative luminance of a #rrggbb hex.
func relLuminance(t *testing.T, hex string) float64 {
	t.Helper()
	var r, g, b int
	if _, err := fmt.Sscanf(hex, "#%02x%02x%02x", &r, &g, &b); err != nil {
		t.Fatalf("bad hex %q: %v", hex, err)
	}
	lin := func(c int) float64 {
		s := float64(c) / 255
		if s <= 0.03928 {
			return s / 12.92
		}
		return math.Pow((s+0.055)/1.055, 2.4)
	}
	return 0.2126*lin(r) + 0.7152*lin(g) + 0.0722*lin(b)
}

// contrastRatio is the WCAG contrast ratio (1–21) between two luminances.
func contrastRatio(a, b float64) float64 {
	if a < b {
		a, b = b, a
	}
	return (a + 0.05) / (b + 0.05)
}

// TestFindHighlightContrastAA guards the find-highlight legibility both palettes
// depend on: the highlighted text (MatchFg) must clear WCAG AA (4.5:1) over BOTH the
// regular-match and current-match backgrounds. The light palette shares one MatchFg
// across two backgrounds, so this is where a white-on-amber regression (~2.3:1, the
// bug this replaced) would be caught.
func TestFindHighlightContrastAA(t *testing.T) {
	for _, bg := range []string{"dark", "light"} {
		p := Default().For(bg == "dark")
		for _, pr := range []struct {
			name   string
			bg, fg string
		}{
			{"match", p.Match.Hex, p.MatchFg.Hex},
			{"current", p.MatchCurrent.Hex, p.MatchFg.Hex},
		} {
			if r := contrastRatio(relLuminance(t, pr.bg), relLuminance(t, pr.fg)); r < 4.5 {
				t.Errorf("%s find-%s contrast %.2f:1 (bg %s / fg %s) < 4.5:1 AA",
					bg, pr.name, r, pr.bg, pr.fg)
			}
		}
	}
}

// The neon-night semantic slots are the contract the CLI's 16-color path and the
// TUI's truecolor path both depend on. Pin hex + ANSI slot so a palette edit is a
// deliberate, reviewed change (and the "danger-red is the legible Outrun swap"
// decision can't silently regress).
func TestNeonDarkSemanticSlots(t *testing.T) {
	p := Default().Dark
	cases := []struct {
		name string
		c    theme.Color
		hex  string
		ansi int
	}{
		{"none", theme.ColorNone, "", NoANSI},
		{"red", theme.ColorRed, "#FF4242", 1},
		{"green", theme.ColorGreen, "#06ea61", 2},
		{"yellow", theme.ColorYellow, "#c9d364", 3},
		{"blue", theme.ColorBlue, "#03aeff", 4},
		{"cyan", theme.ColorCyan, "#42fff9", 6},
		{"gray", theme.ColorGray, "#a3a5a6", 8},
	}
	for _, tc := range cases {
		got := p.Of(tc.c)
		if got.Hex != tc.hex || got.ANSI != tc.ansi {
			t.Errorf("Of(%s) = {%q, %d}, want {%q, %d}", tc.name, got.Hex, got.ANSI, tc.hex, tc.ansi)
		}
	}
}

// The neon-DAY (light) semantic slots are the light-background path T5 exercises;
// pin them so the AA-darkened accents (green/yellow/teal/blue chosen to clear WCAG
// 4.5:1 on the Latte bg) can't silently drift back to Latte's failing defaults.
func TestNeonLightSemanticSlots(t *testing.T) {
	p := Default().Light
	cases := []struct {
		name string
		c    theme.Color
		hex  string
		ansi int
	}{
		{"none", theme.ColorNone, "", NoANSI},
		{"red", theme.ColorRed, "#d20f39", 1},
		{"green", theme.ColorGreen, "#2e7d1f", 2},
		{"yellow", theme.ColorYellow, "#8a6000", 3},
		{"blue", theme.ColorBlue, "#2258cc", 4},
		{"cyan", theme.ColorCyan, "#0e6e74", 6},
		{"gray", theme.ColorGray, "#6c6f85", 8},
	}
	for _, tc := range cases {
		got := p.Of(tc.c)
		if got.Hex != tc.hex || got.ANSI != tc.ansi {
			t.Errorf("light Of(%s) = {%q, %d}, want {%q, %d}", tc.name, got.Hex, got.ANSI, tc.hex, tc.ansi)
		}
	}
}

// The light find-highlight is a tuned, human-validated pair: dark text on an amber
// match / lightened-mauve current, replacing an unreadable white-on-amber (~2.3:1).
// Pin the exact hexes so the approved choice can't drift to a different-but-still-AA
// color unreviewed — the contrast property itself is guarded by
// TestFindHighlightContrastAA.
func TestNeonLightHighlight(t *testing.T) {
	p := Default().Light
	cases := []struct {
		name string
		got  Hue
		hex  string
		ansi int
	}{
		{"match", p.Match, "#df8e1d", 3},
		{"current", p.MatchCurrent, "#c9a6f8", 13},
		{"matchFg", p.MatchFg, "#1e1e2e", 0},
	}
	for _, tc := range cases {
		if tc.got.Hex != tc.hex || tc.got.ANSI != tc.ansi {
			t.Errorf("light %s = {%q, %d}, want {%q, %d}", tc.name, tc.got.Hex, tc.got.ANSI, tc.hex, tc.ansi)
		}
	}
}

// Of must DEGRADE (not panic) on a theme.Color the palette never filled — the
// reason Semantic is a map, not a fixed array. An unmapped slot renders plain.
func TestOfUnknownColorDegrades(t *testing.T) {
	// A value past the defined enum: a future theme.Color the literals haven't filled.
	if got := Default().Dark.Of(theme.Color(99)); got.Hex != "" || got.ANSI != NoANSI {
		t.Errorf("Of(unknown) = {%q, %d}, want plain {\"\", NoANSI}", got.Hex, got.ANSI)
	}
}

// The accent is the neon signature (bright magenta) and must degrade to a chosen
// ANSI slot, not a runtime guess.
func TestNeonAccent(t *testing.T) {
	if a := Default().Dark.Accent; a.Hex != "#ea5ce2" || a.ANSI != 13 {
		t.Errorf("dark accent = {%q, %d}, want {#ea5ce2, 13}", a.Hex, a.ANSI)
	}
}

// The rollup bar gradient is the deliberate truecolor exception: purple -> cyan ->
// pink, three stops, anchored to the existing neon values.
func TestNeonGradient(t *testing.T) {
	g := Default().Dark.Gradient
	want := []string{"#b026ff", "#00e5ff", "#ff2ec4"}
	if len(g) != len(want) {
		t.Fatalf("gradient has %d stops, want %d", len(g), len(want))
	}
	for i, w := range want {
		if g[i].Hex != w {
			t.Errorf("gradient[%d] = %q, want %q", i, g[i].Hex, w)
		}
	}
}

// For picks the background-appropriate palette; Markdown follows the background.
func TestThemeFor(t *testing.T) {
	tm := Default()
	if tm.For(true).Markdown != theme.MarkdownStyleDark {
		t.Errorf("For(dark).Markdown = %q, want %q", tm.For(true).Markdown, theme.MarkdownStyleDark)
	}
	if tm.For(false).Markdown != theme.MarkdownStyleLight {
		t.Errorf("For(light).Markdown = %q, want %q", tm.For(false).Markdown, theme.MarkdownStyleLight)
	}
}

// Lookup degrades unknown/empty names to the default rather than erroring.
func TestLookupDegrades(t *testing.T) {
	if tm, ok := Lookup("neon"); !ok || tm.Name != "neon" {
		t.Errorf("Lookup(neon) = {%q, %v}, want {neon, true}", tm.Name, ok)
	}
	if tm, ok := Lookup("nope"); ok || tm.Name != "neon" {
		t.Errorf("Lookup(nope) = {%q, %v}, want default {neon, false}", tm.Name, ok)
	}
}

// Names is the enumeration `theme list` relies on: every registered theme, SORTED
// (so the listing + its --json are byte-stable).
func TestNames(t *testing.T) {
	got := Names()
	want := []string{"catppuccin", "neon"}
	if len(got) != len(want) {
		t.Fatalf("Names() = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("Names()[%d] = %q, want %q (sorted)", i, got[i], want[i])
		}
	}
}

// The Catppuccin Mocha (dark) semantic slots — the second theme's contract — pinned
// so a palette edit is a reviewed change.
func TestCatppuccinDarkSemanticSlots(t *testing.T) {
	tm, ok := Lookup("catppuccin")
	if !ok {
		t.Fatal("Lookup(catppuccin) not registered")
	}
	p := tm.Dark
	cases := []struct {
		name string
		c    theme.Color
		hex  string
		ansi int
	}{
		{"red", theme.ColorRed, "#f38ba8", 1},
		{"green", theme.ColorGreen, "#a6e3a1", 2},
		{"yellow", theme.ColorYellow, "#f9e2af", 3},
		{"blue", theme.ColorBlue, "#89b4fa", 4},
		{"cyan", theme.ColorCyan, "#89dceb", 6},
		{"gray", theme.ColorGray, "#9399b2", 8},
	}
	for _, tc := range cases {
		if got := p.Of(tc.c); got.Hex != tc.hex || got.ANSI != tc.ansi {
			t.Errorf("Of(%s) = {%q, %d}, want {%q, %d}", tc.name, got.Hex, got.ANSI, tc.hex, tc.ansi)
		}
	}
	if a := p.Accent; a.Hex != "#cba6f7" { // mauve
		t.Errorf("catppuccin accent = %q, want #cba6f7 (mauve)", a.Hex)
	}
	// The theme owns its glamour markdown style (the Theme.Markdown wiring); catppuccin
	// ships tokyo-night, distinct from neon's dracula.
	if p.Markdown != "tokyo-night" {
		t.Errorf("catppuccin dark Markdown = %q, want tokyo-night", p.Markdown)
	}
}
