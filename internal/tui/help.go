package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

// helpEntry is one key→description row in the help overlay.
type helpEntry struct{ keys, desc string }

// helpSection groups related bindings under a heading.
type helpSection struct {
	title   string
	entries []helpEntry
}

// helpSections is the single source of truth for the `?` overlay. Keep it in sync
// with keys.go and the focus-routed handlers in model.go.
var helpSections = []helpSection{
	{"Global", []helpEntry{
		{": ", "command / jump (entity, status, or verb)"},
		{"/", "filter the list (slug, desc, tags)"},
		{"o / O", "cycle sort column / reverse"},
		{"s / S", "cycle status view / backward"},
		{"a", "task actions (start/complete/defer/…)"},
		{"[ / ]", "previous / next tab"},
		{"tab", "switch focus (list ⇄ detail)"},
		{"r", "refresh from disk"},
		{"? / esc", "toggle this help"},
		{"q / ctrl+c", "quit / force-quit"},
	}},
	{"List", []helpEntry{
		{"j / k", "move down / up"},
		{"g / G", "top / bottom"},
		{"ctrl+d / u", "half-page down / up"},
		{"enter / l", "open detail"},
		{"h", "back"},
	}},
	{"Detail", []helpEntry{
		{"j / k", "scroll down / up"},
		{"g / G", "top / bottom"},
		{"h / esc", "back to list"},
	}},
}

var (
	helpBorder   = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("6")).Padding(0, 2)
	helpHeading  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("6"))
	helpKeyStyle = lipgloss.NewStyle().Bold(true)
)

// helpBox renders the keybinding panel, clamped to fit within (maxW, maxH).
func helpBox(maxW, maxH int) string {
	// Widest key column across all sections → aligned descriptions.
	keyW := 0
	for _, s := range helpSections {
		for _, e := range s.entries {
			if w := lipgloss.Width(e.keys); w > keyW {
				keyW = w
			}
		}
	}
	var b strings.Builder
	b.WriteString(helpHeading.Render("Keys"))
	b.WriteString("\n")
	for _, s := range helpSections {
		b.WriteString("\n")
		b.WriteString(dim(s.title))
		b.WriteString("\n")
		for _, e := range s.entries {
			pad := strings.Repeat(" ", max(keyW-lipgloss.Width(e.keys), 0))
			b.WriteString("  " + helpKeyStyle.Render(e.keys) + pad + "  " + e.desc + "\n")
		}
	}
	box := helpBorder.Render(strings.TrimRight(b.String(), "\n"))
	// Last-resort clamp so a tiny terminal can't make the box overflow the body.
	return lipgloss.NewStyle().MaxWidth(maxW).MaxHeight(maxH).Render(box)
}

// overlay composites fg centered over bg, leaving the surrounding bg visible (a
// floating modal, not a blank replacement). bg must already be exactly
// width×height cells (use lipgloss.Place to normalize before calling).
func overlay(bg, fg string, width, height int) string {
	bgLines := strings.Split(bg, "\n")
	fgLines := strings.Split(fg, "\n")
	fgW := 0
	for _, l := range fgLines {
		if w := ansi.StringWidth(l); w > fgW {
			fgW = w
		}
	}
	top := max((height-len(fgLines))/2, 0)
	left := max((width-fgW)/2, 0)
	for i, fl := range fgLines {
		row := top + i
		if row < 0 || row >= len(bgLines) {
			continue
		}
		fl += strings.Repeat(" ", max(fgW-ansi.StringWidth(fl), 0)) // pad to box width
		leftPart := ansi.Cut(bgLines[row], 0, left)
		rightPart := ansi.Cut(bgLines[row], left+fgW, width)
		bgLines[row] = leftPart + fl + rightPart
	}
	return strings.Join(bgLines, "\n")
}
