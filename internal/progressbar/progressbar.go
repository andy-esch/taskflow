// Package progressbar renders the rollup completion bar shared by both surfaces —
// the CLI `status` dashboard (via render.Style.Bar) and the TUI epic rows/detail
// (via tui.miniBar). One constructor + one palette so the two can't drift (the
// concern the removed theme.BarFill seam used to cover).
package progressbar

import (
	"image/color"

	"charm.land/bubbles/v2/progress"
	"charm.land/lipgloss/v2"
)

// neon is the 80s-synthwave fill gradient: purple → cyan → pink.
var neon = []color.Color{
	lipgloss.Color("#b026ff"), // neon purple
	lipgloss.Color("#00e5ff"), // neon cyan
	lipgloss.Color("#ff2ec4"), // neon pink
}

// Render returns a width-cell progress bar for pct (0–100): the neon gradient fill
// anchored to the full width (so the fill's reach reads as progress) over a dim
// gray empty track, '█'/'░' glyphs. Always emits lipgloss ANSI; callers that owe a
// byte-stable plain output (the CLI machine path) strip it when styling is off.
func Render(pct, width int) string {
	if width < 1 {
		width = 1
	}
	p := progress.New(
		progress.WithWidth(width),
		progress.WithoutPercentage(),          // callers render the % separately
		progress.WithFillCharacters('█', '░'), // non-half-block ⇒ one color per cell
		progress.WithColors(neon...),
	)
	p.EmptyColor = lipgloss.Color("8") // dim gray unfilled track
	return p.ViewAs(float64(pct) / 100)
}
