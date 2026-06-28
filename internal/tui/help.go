package tui

import (
	"strings"

	"charm.land/lipgloss/v2"

	"github.com/andy-esch/taskflow/internal/domain"
	"github.com/andy-esch/taskflow/internal/theme"
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
		{"e", "edit fields in place — tasks: desc/priority/tags/effort/tier (+revisit when deferred) · epics: desc/priority/tags"},
		{"E", "open the whole file in $EDITOR (any entity; re-read on save)"},
		{"f", "follow reference (task ⇄ epic)"},
		{"ctrl+o", "jump back (follow history)"},
		{"y / Y", "copy slug / file path to clipboard"},
		{"[ / ]", "previous / next tab"},
		{"tab", "switch focus (list ⇄ detail)"},
		{"z", "full-screen the detail pane (z/esc to exit)"},
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
}

// symbolsFor builds the glyph legend for the active screen — what the leading
// status / liveness / bucket glyphs (and the ⚠/↻ markers) in the rows actually
// mean, so a reader can decode the column without leaving the TUI. The glyph rows
// (tok) are sourced from the SAME theme tokens the row delegates draw
// (theme.Status/Liveness/Bucket/FindingStatus), so THEY can't drift from what's on
// screen; the marker/percent rows (mark) are hand-labeled and stay in sync by eye.
// ok is false on a screen with no glyph vocabulary of its own.
func symbolsFor(kind entityKind) (helpSection, bool) {
	tok := func(t theme.Token, desc string) helpEntry { return helpEntry{fg(t.Color, t.Glyph), desc} }
	mark := func(c theme.Color, glyph, desc string) helpEntry { return helpEntry{fg(c, glyph), desc} }
	var e []helpEntry
	switch kind {
	case entityTasks:
		for _, st := range domain.AllStatuses() {
			e = append(e, tok(theme.Status(st), string(st)))
		}
		e = append(e,
			mark(theme.ColorYellow, "⚠", "misfiled — status ≠ folder"),
			mark(theme.ColorYellow, "↻", "revisit (snooze) date reached"),
		)
	case entityEpics:
		e = append(e,
			tok(theme.Liveness("working"), "working — live work in progress"),
			tok(theme.Liveness("fresh"), "fresh — new epic, no tasks yet"),
			tok(theme.Liveness("dormant"), "dormant — drained / quiet (id dimmed)"),
			mark(theme.ColorYellow, "⚠", "non-conforming status (→ active/retired/deprecated)"),
		)
	case entityAudits:
		for _, b := range domain.AllAuditBuckets() {
			e = append(e, tok(theme.Bucket(b), string(b)+" bucket"))
		}
		e = append(e,
			tok(theme.FindingStatus("open"), "finding: open"),
			tok(theme.FindingStatus("in-progress"), "finding: in-progress"),
			tok(theme.FindingStatus("fixed"), "finding: fixed / landed"),
			tok(theme.FindingStatus("deferred"), "finding: deferred / superseded"),
			tok(theme.FindingStatus("wontfix"), "finding: wontfix"),
		)
	default: // dashboard — a cross-entity screen, so the essentials of each widget
		e = append(e,
			tok(theme.Status(domain.StatusInProgress), "active / working (task in-progress · epic working)"),
			tok(theme.Liveness("fresh"), "fresh epic (no tasks yet)"),
			tok(theme.Liveness("dormant"), "dormant epic (drained / quiet)"),
			tok(theme.Bucket(domain.AuditOpen), "open audit"),
			// ✓ (U+2713) is the dashboard's "ready to close" badge, distinct from the ✔
			// (U+2714) that means done/closed elsewhere — keep the glyph + label specific.
			mark(theme.ColorGreen, "✓", "audit ready to close (findings resolved)"),
			mark(theme.ColorYellow, "⚠", "needs attention (misfiled / non-conforming)"),
			mark(theme.ColorYellow, "↻", "revisit (snooze) reached"),
		)
	}
	// The completion-percent figure's color band — labels mirror theme.Percent
	// (gray <34, yellow <100, green 100); keep them in sync. The rollup BAR uses a
	// separate gradient (progressbar), so this describes the percent number's color.
	e = append(e,
		mark(theme.ColorGreen, "100%", "complete"),
		mark(theme.ColorYellow, "34–99%", "in progress"),
		mark(theme.ColorGray, "0–33%", "barely started"),
	)
	return helpSection{"Symbols", e}, true
}

// notesFor builds the context Notes for the ACTIVE entity — only that tab's view
// vocabulary (tasks and audits each have a view axis; epics don't), plus the
// generic find note. So the epics/tasks tabs never advertise audit views, and the
// audits tab never shows task views (the `?` panel is page-specific, like the keys).
func notesFor(kind entityKind) helpSection {
	var entries []helpEntry
	switch kind {
	case entityDashboard:
		entries = append(entries, helpEntry{"dashboard", "⏎ open the selected item · ] / [ to the tabs · r refresh"})
	case entityTasks:
		entries = append(entries, helpEntry{"views", "s/S or :working / :deferred / :revisit / :all"})
	case entityAudits:
		entries = append(entries, helpEntry{"views", "s/S or :open / :closed / :deferred / :all to switch bucket"})
	}
	entries = append(entries, helpEntry{"find", "matches the rendered text on screen — R for the raw source"})
	return helpSection{"Notes", entries}
}

var (
	helpBorder   = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("6")).Padding(0, 2)
	helpHeading  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("6"))
	helpKeyStyle = lipgloss.NewStyle().Bold(true)
	// helpHFrame is the box's horizontal chrome (border + padding), derived from the
	// style so a padding/border change can't desync the wrap width.
	helpHFrame = helpBorder.GetHorizontalFrameSize()
)

// helpPrefWidth is the overlay's preferred outer width. The content WRAPS to this
// (descriptions reflow within their column), so the box is a constant width as you
// scroll — rather than resizing to whatever the widest currently-visible line is. It
// narrows to fit a smaller terminal.
const helpPrefWidth = 62

// helpWidth is the overlay's outer width for a given available width: the preferred
// width, clamped down to what fits. A pure function of maxW (not the scroll
// position), which is what keeps the box from resizing as it scrolls.
func helpWidth(maxW int) int {
	if maxW > 0 && maxW < helpPrefWidth {
		return maxW
	}
	return helpPrefWidth
}

// helpSectionsFor orders the panel by relevance to where you are: the active
// pane's keys first (List on the list, Detail in the detail pane), general Notes
// next, and Global LAST — global keys work everywhere, so they're the best-known
// and least in need of surfacing in a context panel. The inactive pane's section
// is hidden, so `?` shows what actually works right now.
func helpSectionsFor(f focus, kind entityKind) []helpSection {
	byTitle := make(map[string]helpSection, len(helpSections))
	for _, s := range helpSections {
		byTitle[s.title] = s
	}
	out := make([]helpSection, 0, 4) // active-pane keys + Symbols + Notes + Global
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
	if sym, ok := symbolsFor(kind); ok { // what the row glyphs mean, page-specific
		out = append(out, sym)
	}
	out = append(out, notesFor(kind)) // page-specific: only the active tab's views
	add("Global")
	return out
}

// helpLines builds the overlay's content lines for the current focus, laid out in a
// fixed content width: each row is a 2-space indent + the key column (aligned) + a
// 2-space gap + a description that WORD-WRAPS within the remaining column, its
// continuation lines indented to sit under the description (not the key). Every line
// is forced to EXACTLY contentW (truncate as the backstop, pad as the common case)
// so the box is a constant width across scroll AND on a terminal too narrow to honor
// the column layout. Shared by helpBox (render) and the model's scroll clamp, so both
// window the SAME content — callers MUST pass the same contentW
// (helpWidth(maxW)-helpHFrame).
func helpLines(f focus, kind entityKind, contentW int) []string {
	sections := helpSectionsFor(f, kind)
	// Widest key column across the shown sections → aligned descriptions.
	keyW := 0
	for _, s := range sections {
		for _, e := range s.entries {
			if w := lipgloss.Width(e.keys); w > keyW {
				keyW = w
			}
		}
	}
	const indent, gap = 2, 2
	descW := max(contentW-indent-keyW-gap, 8) // floor so wrapping stays sane on a narrow box
	contIndent := strings.Repeat(" ", indent+keyW+gap)
	// fit forces a line to EXACTLY contentW. padRight handles the common (too-short)
	// case; truncate is the backstop for when contentW can't fit indent+keyW+gap+descW
	// (a terminal narrower than ~30 cols) — without it those rows would stay over-wide,
	// the box would lose its uniform width, and the outer clamp would shear the border.
	fit := func(s string) string { return padRight(truncate(s, contentW), contentW) }

	lines := []string{fit(helpHeading.Render("Keys"))}
	for _, s := range sections {
		lines = append(lines, fit(""), fit(dim(s.title)))
		for _, e := range s.entries {
			keyPad := strings.Repeat(" ", max(keyW-lipgloss.Width(e.keys), 0))
			// wrap reflows + right-pads each description line to descW (lipgloss Width),
			// so the column stays rectangular and the rows align.
			desc := strings.Split(wrap(e.desc, descW), "\n")
			lines = append(lines, fit("  "+helpKeyStyle.Render(e.keys)+keyPad+"  "+desc[0]))
			for _, cont := range desc[1:] {
				lines = append(lines, fit(contIndent+cont))
			}
		}
	}
	return lines
}

// helpBox renders the keybinding panel at a FIXED width (helpWidth(maxW)) so it
// doesn't resize while scrolling, clamped to fit within (maxW, maxH). When the
// content is taller than the box, scroll (clamped here, not in the model — only
// render knows the box height) picks the visible window; j/k scroll while open.
func helpBox(maxW, maxH, scroll int, f focus, kind entityKind) string {
	lines := helpLines(f, kind, helpWidth(maxW)-helpHFrame)
	const frameV = 2 // top+bottom border rows
	if innerH := maxH - frameV; innerH > 0 && len(lines) > innerH {
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
