// Package progressbar renders the rollup completion bar shared by both surfaces —
// the CLI `status` dashboard (via render.Style.Bar) and the TUI epic rows/detail
// (via tui.miniBar). One constructor + one palette so the two can't drift (the
// concern the removed theme.BarFill seam used to cover).
package progressbar

import (
	"image/color"
	"strings"

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

// Segments is an audit's finding breakdown for a stacked bar: counts of done
// (fixed/landed), active (in-progress), and dropped (deferred/superseded/wontfix)
// findings out of Total. The remainder (Total − Done − Active − Dropped) — open
// plus any unrecognized status — is the empty track.
type Segments struct {
	Done    int
	Active  int
	Dropped int
	Total   int
}

// segment bands render in fixed order with DISTINCT glyphs (not just colors) so
// the stacking survives --color=never / a mono terminal: █ done (green), ▓ active
// (yellow), ▒ dropped (gray), ░ the open/empty track (dim). The headline % and
// done/total are rendered separately by callers (the green band's reach == %done).
var segBands = []struct {
	glyph rune
	col   color.Color
}{
	{'█', lipgloss.Color("2")}, // done — green
	{'▓', lipgloss.Color("3")}, // active — yellow
	{'▒', lipgloss.Color("8")}, // dropped — gray
	{'░', lipgloss.Color("8")}, // open / empty track — dim gray (glyph distinguishes it from dropped)
}

// RenderSegments draws s as a width-cell stacked bar. Cells are apportioned by
// largest-remainder so the bands sum to exactly width; a band with a non-zero
// count can still round to zero cells at small widths (the counts beside the bar
// carry the exact numbers). Always emits lipgloss ANSI; the CLI strips it when
// styling is off, leaving the distinct glyphs intact.
func RenderSegments(s Segments, width int) string {
	if width < 1 {
		width = 1
	}
	open := s.Total - s.Done - s.Active - s.Dropped
	if open < 0 {
		open = 0
	}
	cells := apportion([]int{s.Done, s.Active, s.Dropped, open}, width)
	var b strings.Builder
	for i, n := range cells {
		if n <= 0 {
			continue
		}
		band := strings.Repeat(string(segBands[i].glyph), n)
		b.WriteString(lipgloss.NewStyle().Foreground(segBands[i].col).Render(band))
	}
	return b.String()
}

// apportion splits width cells across counts proportionally, distributing the
// rounding leftover to the largest fractional remainders so the result sums to
// exactly width. With every count zero (no findings) the whole width is the last
// band (the empty track).
func apportion(counts []int, width int) []int {
	out := make([]int, len(counts))
	total := 0
	for _, c := range counts {
		total += c
	}
	if total <= 0 {
		out[len(out)-1] = width
		return out
	}
	rem := make([]float64, len(counts))
	used := 0
	for i, c := range counts {
		exact := float64(c) * float64(width) / float64(total)
		out[i] = int(exact)
		rem[i] = exact - float64(out[i])
		used += out[i]
	}
	for k := 0; k < width-used; k++ {
		best, bestRem := -1, -1.0
		for i, r := range rem {
			if r > bestRem {
				best, bestRem = i, r
			}
		}
		if best < 0 {
			break
		}
		out[best]++
		rem[best] = -1 // don't reselect this band
	}
	return out
}
