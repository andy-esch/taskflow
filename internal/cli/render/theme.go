package render

import (
	"fmt"
	"io"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/andy-esch/taskflow/internal/design"
	"github.com/andy-esch/taskflow/internal/progressbar"
	"github.com/andy-esch/taskflow/internal/theme"
	"github.com/andy-esch/taskflow/internal/wire"
)

// themeSwatch is one named color in a theme preview.
type themeSwatch struct {
	token string
	hue   design.Hue
}

// themeSwatches is the fixed, ordered set of tokens a preview shows: the accent
// (the theme's signature) and the six semantic colors.
func themeSwatches(pal design.Palette) []themeSwatch {
	return []themeSwatch{
		{"accent", pal.Accent},
		{"red", pal.Of(theme.ColorRed)},
		{"green", pal.Of(theme.ColorGreen)},
		{"yellow", pal.Of(theme.ColorYellow)},
		{"blue", pal.Of(theme.ColorBlue)},
		{"cyan", pal.Of(theme.ColorCyan)},
		{"gray", pal.Of(theme.ColorGray)},
	}
}

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
	for _, sw := range themeSwatches(pal) {
		if st.on {
			block := lipgloss.NewStyle().Background(canvas).Foreground(sw.hue.Color()).Render("  ███  ")
			fmt.Fprintf(w, "  %s  %-7s %s\n", block, sw.token, sw.hue.Hex)
		} else {
			fmt.Fprintf(w, "  %-7s %s\n", sw.token, sw.hue.Hex)
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
	bar := progressbar.Render(60, 24, pal)
	if !st.on {
		bar = ansi.Strip(bar)
	}
	fmt.Fprintf(w, "  bar     %s\n", bar)
}

// ThemePreviewJSON emits a theme's palette swatches as the machine form.
func ThemePreviewJSON(w io.Writer, name, variant string, pal design.Palette) error {
	sw := themeSwatches(pal)
	entries := make([]wire.ThemeSwatch, len(sw))
	for i, s := range sw {
		entries[i] = wire.ThemeSwatch{Token: s.token, Hex: s.hue.Hex, ANSI: s.hue.ANSI}
	}
	return wire.EncodeJSON(w, wire.ToThemePreviewEnvelope(name, variant, entries))
}
