package render

import (
	"fmt"
	"io"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/andy-esch/taskflow/internal/design"
	"github.com/andy-esch/taskflow/internal/themepreview"
	"github.com/andy-esch/taskflow/internal/wire"
)

// ThemePreviewHuman renders a palette as a swatch grid: a truecolor block + hex per
// token, plus a sample gradient bar. When styling is off (piped / --color=never) it
// degrades to a plain token→hex list + the stripped bar glyphs, so it stays
// byte-stable like every other surface.
func ThemePreviewHuman(w io.Writer, st Style, name, variant string, pal design.Palette) {
	fmt.Fprintf(w, "%s (%s)\n", st.Bold(name), variant)
	// Paint the colored samples on the palette's INTENDED background (pal.Base), not
	// the reviewer's terminal bg — so a palette's colors are judged against the bg
	// they were tuned for. This is what makes `--variant light` faithful from a dark
	// terminal (and vice-versa).
	canvas := pal.Base.Color()
	for _, sw := range themepreview.Swatches(pal) {
		if st.on {
			block := lipgloss.NewStyle().Background(canvas).Foreground(sw.Hue.Color()).Render("  ███  ")
			fmt.Fprintf(w, "  %s  %-7s %s\n", block, sw.Token, sw.Hue.Hex)
		} else {
			fmt.Fprintf(w, "  %-7s %s\n", sw.Token, sw.Hue.Hex)
		}
	}
	// Chrome the foreground swatches can't show: the find highlight is a bg+fg PAIR
	// (legibility depends on both), and borders are a frame color. Human-only — the
	// JSON preview stays the semantic-swatch machine contract — this is the surface
	// for eyeballing highlight/border legibility (notably the light palette) from any
	// terminal, paired with `--variant`.
	if st.on {
		chip := func(label string, bg, fg design.Hue) string {
			return lipgloss.NewStyle().Background(bg.Color()).Foreground(fg.Color()).Render(" " + label + " ")
		}
		fmt.Fprintf(w, "  find    %s %s\n", chip("match", pal.Match, pal.MatchFg), chip("current", pal.MatchCurrent, pal.MatchFg))
		rule := func(h design.Hue) string {
			return lipgloss.NewStyle().Background(canvas).Foreground(h.Color()).Render(" ────── ")
		}
		fmt.Fprintf(w, "  border  %s active  %s idle\n", rule(pal.BorderActive), rule(pal.BorderIdle))
	}
	bar := themepreview.Bar(pal, 24)
	if !st.on {
		bar = ansi.Strip(bar)
	}
	fmt.Fprintf(w, "  bar     %s\n", bar)
}

// ThemePreviewJSON emits a theme's palette swatches as the machine form.
func ThemePreviewJSON(w io.Writer, name, variant string, pal design.Palette) error {
	sw := themepreview.Swatches(pal)
	entries := make([]wire.ThemeSwatch, len(sw))
	for i, s := range sw {
		entries[i] = wire.ThemeSwatch{Token: s.Token, Hex: s.Hue.Hex, ANSI: s.Hue.ANSI}
	}
	return wire.EncodeJSON(w, wire.ToThemePreviewEnvelope(name, variant, entries))
}
