package tui

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"

	"github.com/andy-esch/taskflow/internal/domain"
	"github.com/andy-esch/taskflow/internal/theme"
)

var (
	selectedStyle = lipgloss.NewStyle().Bold(true)
	dimStyle      = lipgloss.NewStyle().Faint(true)

	// Two focus signals: an accent border + a bold title on the focused pane.
	paneActive   = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("6"))
	paneInactive = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("8"))

	// Frame sizes derived from the pane style (not a hardcoded 2) so a future
	// border/padding change can't silently desync sizing.
	paneHFrame = paneActive.GetHorizontalFrameSize()
	paneVFrame = paneActive.GetVerticalFrameSize()
)

// lipColor maps a semantic theme.Color to a lipgloss 16-color (the TUI's
// rendering of the same status semantics the CLI renders as ANSI).
func lipColor(c theme.Color) lipgloss.Color {
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
