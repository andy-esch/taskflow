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

// TestChromeSurfaceContrastAA is the generalization of TestFindHighlightContrastAA
// to the OTHER foreground-over-colored-background pairing in the project: fang's
// help codeblock, whose Codeblock role is a background and whose every token is
// drawn over it. It exists because the original mapping handed that background
// role a foreground hue (theme.ColorGray), which painted the USAGE box light grey
// and rendered DimmedArgument at 1.00:1 against its own background.
//
// It runs over EVERY registered theme and BOTH backgrounds, and asserts the two
// properties a surface must hold:
//
//	a) every colored role fang draws inside the box clears WCAG AA (4.5:1);
//	b) the surface stays close to the palette Base, because the box's prose is
//	   drawn in the TERMINAL's default foreground (fang shares one Base between
//	   the help body and the codeblock, so no token here can compensate for it) —
//	   while staying far enough to read as a distinct block.
func TestChromeSurfaceContrastAA(t *testing.T) {
	for _, name := range Names() {
		th, ok := Lookup(name)
		if !ok {
			t.Fatalf("%s should be registered", name)
		}
		for _, bg := range []string{"dark", "light"} {
			p := th.For(bg == "dark")
			surface := relLuminance(t, p.Surface.Hex)
			// The roles fang paints over the codeblock background. Title and the
			// error badge are deliberately absent: they render outside the box.
			for role, hue := range map[string]Hue{
				"Program(blue)":      p.Of(theme.ColorBlue),
				"Command(yellow)":    p.Of(theme.ColorYellow),
				"Flag(green)":        p.Of(theme.ColorGreen),
				"QuotedString(cyan)": p.Of(theme.ColorCyan),
				"Dimmed/Comment":     p.Of(theme.ColorGray),
			} {
				if r := contrastRatio(surface, relLuminance(t, hue.Hex)); r < 4.5 {
					t.Errorf("%s/%s: %s over surface %s is %.2f:1 < 4.5:1 AA",
						name, bg, role, p.Surface.Hex, r)
				}
			}
			lift := contrastRatio(surface, relLuminance(t, p.Base.Hex))
			if lift < 1.03 {
				t.Errorf("%s/%s: surface %s is indistinguishable from base %s (%.2f:1)",
					name, bg, p.Surface.Hex, p.Base.Hex, lift)
			}
			// The box's prose is the TERMINAL's default foreground, which is tuned for
			// the base and which no token here can change, so the surface has to stay on
			// the base's side of the light/dark divide. This is the structural form of
			// the original defect — a light-grey box on a dark palette — and it catches
			// it independently of the AA loop above, which only sees the colored roles.
			if darkSurface(t, p.Surface.Hex) != darkSurface(t, p.Base.Hex) {
				t.Errorf("%s/%s: surface %s sits on the opposite side of the light/dark divide from base %s; "+
					"the terminal's default foreground is tuned for the base and would be unreadable on it",
					name, bg, p.Surface.Hex, p.Base.Hex)
			}
		}
	}
}

// darkSurface reports whether light text reads better than dark text on hex — the
// operational form of "this is a dark background". Used to keep a palette's surface
// on the same side of the divide as its base.
func darkSurface(t *testing.T, hex string) bool {
	l := relLuminance(t, hex)
	return contrastRatio(l, relLuminance(t, "#ffffff")) > contrastRatio(l, relLuminance(t, "#000000"))
}

// TestChromeSurfaceSlots pins the surface hexes the way TestNeonAccent pins the
// accent: the contrast properties above accept any value that happens to pass, but
// these three were each CHOSEN from their scheme, and two of them are deliberately
// counter-intuitive. Pinning them makes a palette edit a reviewed decision rather
// than a silent drift away from that provenance.
func TestChromeSurfaceSlots(t *testing.T) {
	for _, tc := range []struct {
		name, hex, why string
		ansi           int
		pal            Palette
	}{
		{"neon dark", "#1a1b1c", "base16 Synth Midnight base01 — the scheme's own lighter-background slot", 0, neonDark},
		{"catppuccin dark", "#11111b", "Mocha crust — RECESSED, because lifting to surface0 drops overlay2 to 4.45:1", 0, mochaDark},
		{"shared light", "#ffffff", "raised white — Latte layers downward, and its AA-darkened accents lose contrast on every darker layer", 15, latteAA},
	} {
		if got := tc.pal.Surface; got.Hex != tc.hex || got.ANSI != tc.ansi {
			t.Errorf("%s surface = {%q, %d}, want {%q, %d} (%s)", tc.name, got.Hex, got.ANSI, tc.hex, tc.ansi, tc.why)
		}
	}
}
