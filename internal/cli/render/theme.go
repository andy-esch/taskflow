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
	for _, sw := range themeSwatches(pal) {
		if st.on {
			block := lipgloss.NewStyle().Foreground(sw.hue.Color()).Render("███")
			fmt.Fprintf(w, "  %s  %-7s %s\n", block, sw.token, sw.hue.Hex)
		} else {
			fmt.Fprintf(w, "  %-7s %s\n", sw.token, sw.hue.Hex)
		}
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
