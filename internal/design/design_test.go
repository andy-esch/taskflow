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
