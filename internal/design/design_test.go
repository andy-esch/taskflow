package design

import (
	"testing"

	"github.com/andy-esch/taskflow/internal/theme"
)

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
