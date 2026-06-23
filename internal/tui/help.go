package tui

import (
	"strings"

	"charm.land/lipgloss/v2"
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
		{"F", "filter mode: fuzzy ⇄ substring (default fuzzy)"},
		{"o / O", "cycle sort column / reverse"},
		{"s / S", "cycle view: task status / audit bucket"},
		{"a", "task actions (start/complete/defer/…)"},
		{"e", "edit task fields (description/priority/tags/effort/tier)"},
		{"f", "follow reference (task ⇄ epic)"},
		{"ctrl+o", "jump back (follow history)"},
		{"[ / ]", "previous / next tab"},
		{"tab", "switch focus (list ⇄ detail)"},
		{"r", "refresh from disk"},
		{"? / esc", "toggle this help"},
		{"q / ctrl+c", "quit / force-quit"},
	}},
	{"List", []helpEntry{
		{"j / k", "move down / up"},
		{"g / G", "top / bottom"},
		{"d / u", "page down / up (pgdn/pgup)"},
		{"enter / l", "open detail"},
		{"h", "back"},
	}},
	{"Detail", []helpEntry{
		{"j / k", "scroll down / up"},
		{"ctrl+d / u", "half-page down / up"},
		{"g / G", "top / bottom"},
		{"/", "find in body"},
		{"n / N", "next / previous match"},
		{"R", "raw ⇄ pretty markdown"},
		{"h / esc", "back to list (esc clears a find first)"},
	}},
	{"Notes", []helpEntry{
		{"audits", "s/S or :open/:closed/:deferred/:all to switch bucket"},
		{"find", "matches the rendered text on screen — R for the raw source"},
	}},
}

var (
	helpBorder   = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("6")).Padding(0, 2)
	helpHeading  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("6"))
	helpKeyStyle = lipgloss.NewStyle().Bold(true)
)

// helpLines builds the overlay's content lines (heading, sections, aligned
// key→desc rows). Shared by helpBox (render) and the model's scroll clamp.
func helpLines() []string {
	// Widest key column across all sections → aligned descriptions.
	keyW := 0
	for _, s := range helpSections {
		for _, e := range s.entries {
			if w := lipgloss.Width(e.keys); w > keyW {
				keyW = w
			}
		}
	}
	lines := []string{helpHeading.Render("Keys")}
	for _, s := range helpSections {
		lines = append(lines, "", dim(s.title))
		for _, e := range s.entries {
			pad := strings.Repeat(" ", max(keyW-lipgloss.Width(e.keys), 0))
			lines = append(lines, "  "+helpKeyStyle.Render(e.keys)+pad+"  "+e.desc)
		}
	}
	return lines
}

// helpBox renders the keybinding panel, clamped to fit within (maxW, maxH).
// When the content is taller than the box, scroll (clamped here, not in the
// model — only render knows the box height) picks the visible window; j/k
// scroll while the overlay is open.
func helpBox(maxW, maxH, scroll int) string {
	lines := helpLines()
	const frameH = 2 // top+bottom border rows
	if innerH := maxH - frameH; innerH > 0 && len(lines) > innerH {
		maxScroll := len(lines) - innerH
		scroll = min(max(scroll, 0), maxScroll)
		lines = lines[scroll : scroll+innerH]
	}
	box := helpBorder.Render(strings.Join(lines, "\n"))
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
