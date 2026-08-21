// Package themepreview projects a design palette into the stable, ordered sample
// set used by every interactive and non-interactive theme preview. Rendering
// adapters remain free to choose their own layout, but they cannot drift on which
// semantic colors or gradient are being reviewed.
package themepreview

import (
	"github.com/andy-esch/taskflow/internal/design"
	"github.com/andy-esch/taskflow/internal/progressbar"
	"github.com/andy-esch/taskflow/internal/theme"
)

// Swatch is one named color in a theme preview.
type Swatch struct {
	Token string
	Hue   design.Hue
}

// Swatches returns the canonical preview tokens: the signature accent followed
// by all six semantic colors.
func Swatches(pal design.Palette) []Swatch {
	return []Swatch{
		{Token: "accent", Hue: pal.Accent},
		{Token: "red", Hue: pal.Of(theme.ColorRed)},
		{Token: "green", Hue: pal.Of(theme.ColorGreen)},
		{Token: "yellow", Hue: pal.Of(theme.ColorYellow)},
		{Token: "blue", Hue: pal.Of(theme.ColorBlue)},
		{Token: "cyan", Hue: pal.Of(theme.ColorCyan)},
		{Token: "gray", Hue: pal.Of(theme.ColorGray)},
	}
}

// Bar renders the canonical sample gradient at the requested cell width.
func Bar(pal design.Palette, width int) string { return progressbar.Render(60, width, pal) }
