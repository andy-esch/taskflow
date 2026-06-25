package tui

import (
	"strings"

	"charm.land/lipgloss/v2"
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
		{"ctrl+p", "command palette — fuzzy jump to anything / run a command"},
		{": ", "command / jump (entity, status, or verb)"},
		{"/", "filter the list (slug, desc, tags)"},
		{"F", "filter mode: fuzzy ⇄ substring (default fuzzy)"},
		{"o / O", "cycle sort column / reverse"},
		{"s / S", "cycle view: task status / audit bucket"},
		{"m", "move — lifecycle (start/complete/defer/…); audits: close/reopen/defer; epics: activate/retire/deprecate"},
		{"e", "edit fields in place — tasks: desc/priority/tags/effort/tier · epics: desc/priority/tags"},
		{"E", "open the whole file in $EDITOR (any entity; re-read on save)"},
		{"f", "follow reference (task ⇄ epic)"},
		{"ctrl+o", "jump back (follow history)"},
		{"y / Y", "copy slug / file path to clipboard"},
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

// helpSectionsFor orders the panel by relevance to where you are: the active
// pane's keys first (List on the list, Detail in the detail pane), general Notes
// next, and Global LAST — global keys work everywhere, so they're the best-known
// and least in need of surfacing in a context panel. The inactive pane's section
// is hidden, so `?` shows what actually works right now.
func helpSectionsFor(f focus) []helpSection {
	byTitle := make(map[string]helpSection, len(helpSections))
	for _, s := range helpSections {
		byTitle[s.title] = s
	}
	out := make([]helpSection, 0, 3)
	add := func(title string) {
		if s, ok := byTitle[title]; ok {
			out = append(out, s)
		}
	}
	if f == focusDetail {
		add("Detail")
	} else {
		add("List")
	}
	add("Notes")
	add("Global")
	return out
}

// helpLines builds the overlay's content lines (heading, sections, aligned
// key→desc rows) for the current focus. Shared by helpBox (render) and the model's
// scroll clamp, so both window the SAME content.
func helpLines(f focus) []string {
	sections := helpSectionsFor(f)
	// Widest key column across the shown sections → aligned descriptions.
	keyW := 0
	for _, s := range sections {
		for _, e := range s.entries {
			if w := lipgloss.Width(e.keys); w > keyW {
				keyW = w
			}
		}
	}
	lines := []string{helpHeading.Render("Keys")}
	for _, s := range sections {
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
func helpBox(maxW, maxH, scroll int, f focus) string {
	lines := helpLines(f)
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
//
// It uses lipgloss v2 layer compositing: a fixed width×height canvas with the bg
// layer at the origin and the fg layer centered on top (higher z). The canvas is
// fixed-size, so fg is clipped to it and the output is always exactly
// width×height — the contract the bodyView layout depends on.
func overlay(bg, fg string, width, height int) string {
	x := max((width-lipgloss.Width(fg))/2, 0)
	y := max((height-lipgloss.Height(fg))/2, 0)
	comp := lipgloss.NewCompositor(
		lipgloss.NewLayer(bg),
		lipgloss.NewLayer(fg).X(x).Y(y).Z(1),
	)
	return lipgloss.NewCanvas(width, height).Compose(comp).Render()
}
