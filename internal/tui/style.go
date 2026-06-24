package tui

import (
	"image/color"
	"strings"

	"charm.land/bubbles/v2/progress"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/andy-esch/taskflow/internal/domain"
	"github.com/andy-esch/taskflow/internal/theme"
)

// accent is the focus/selection accent (cyan), shared by the active pane border
// and the active tab in the strip. lipgloss v2 Color is a func returning a
// color.Color value (not a const string type), so this is a var.
var accent = lipgloss.Color("6")

var (
	selectedStyle = lipgloss.NewStyle().Bold(true)
	dimStyle      = lipgloss.NewStyle().Faint(true)
	activeTab     = lipgloss.NewStyle().Bold(true).Foreground(accent)

	// Two focus signals: an accent border + a bold title on the focused pane.
	paneActive   = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(accent)
	paneInactive = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("8"))

	// Frame sizes derived from the pane style (not a hardcoded 2) so a future
	// border/padding change can't silently desync sizing.
	paneHFrame = paneActive.GetHorizontalFrameSize()
	paneVFrame = paneActive.GetVerticalFrameSize()
)

// lipColor maps a semantic theme.Color to a lipgloss 16-color (the TUI's
// rendering of the same status semantics the CLI renders as ANSI).
func lipColor(c theme.Color) color.Color {
	switch c {
	case theme.ColorRed:
		return lipgloss.Color("1")
	case theme.ColorGreen:
		return lipgloss.Color("2")
	case theme.ColorYellow:
		return lipgloss.Color("3")
	case theme.ColorBlue:
		return lipgloss.Color("4")
	case theme.ColorCyan:
		return lipgloss.Color("6")
	case theme.ColorGray:
		return lipgloss.Color("8")
	default:
		return lipgloss.Color("")
	}
}

func fg(c theme.Color, s string) string {
	return lipgloss.NewStyle().Foreground(lipColor(c)).Render(s)
}

func dim(s string) string { return dimStyle.Render(s) }

// neonBar is the 80s-synthwave fill gradient for the rollup bars: neon purple →
// cyan → pink. Truecolor (downsampled on lesser terminals). Keep in sync with
// render's copy so the CLI status bars match the TUI.
var neonBar = []color.Color{
	lipgloss.Color("#b026ff"), // neon purple
	lipgloss.Color("#00e5ff"), // neon cyan
	lipgloss.Color("#ff2ec4"), // neon pink
}

// miniBar renders a width-cell progress bar for pct (0–100) via the bubbles v2
// progress component. The fill is the 80s-neon gradient (neonBar) anchored to the
// full width, so how far the fill reaches reads as progress; the hue is purely
// decorative — the discrete completion tier (gray/yellow/green) still shows in the
// % text beside the bar. Empty cells stay gray (distinguished from fill by the ░
// vs █ glyph). Mirrors the CLI render.Style.Bar so both surfaces show the same bar.
func miniBar(pct, width int) string {
	if width < 1 {
		width = 1
	}
	p := progress.New(
		progress.WithWidth(width),
		progress.WithoutPercentage(),          // callers render the % separately
		progress.WithFillCharacters('█', '░'), // non-half-block ⇒ one color per cell
		progress.WithColors(neonBar...),
	)
	p.EmptyColor = lipColor(theme.ColorGray)
	return p.ViewAs(float64(pct) / 100)
}

// statusText renders a colored glyph + status label.
func statusText(st domain.Status) string {
	tok := theme.Status(st)
	return fg(tok.Color, tok.Glyph+" "+string(st))
}

// priorityText colors a priority label (empty stays empty).
func priorityText(p string) string {
	if p == "" {
		return ""
	}
	return fg(theme.Priority(p), p)
}

// truncate shortens s to max display cells with a trailing "…". ANSI- and
// width-aware (handles wide runes + embedded escapes), so it can't overflow a
// cell budget — the discipline for anything fed to a Join.
func truncate(s string, max int) string {
	if max < 1 {
		return ""
	}
	return ansi.Truncate(s, max, "…")
}

// padRight pads s with spaces to w *display cells* (not bytes), so a column stays
// aligned even when a value has multi-byte or wide runes. Overlong s is returned
// unchanged (truncate to the budget first).
func padRight(s string, w int) string {
	if gap := w - ansi.StringWidth(s); gap > 0 {
		return s + strings.Repeat(" ", gap)
	}
	return s
}
