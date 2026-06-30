package tui

import (
	"strings"

	"charm.land/bubbles/v2/key"
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

// he turns a keyMap binding into a help row, sourcing BOTH the displayed key and the
// description from the binding's key.WithHelp — so the `?` overlay can't drift from
// keys.go (the single source).
func he(b key.Binding) helpEntry { return helpEntry{b.Help().Key, b.Help().Desc} }

// helpSections derives the `?` overlay's keybinding rows from the keyMap (keys.go).
// The only literal rows are keys with NO keyMap binding: the list/viewport's own
// j/k + paging, and the list's `/` filter (keys.Find is the DETAIL find). The
// focus-routed handlers in model.go still match the bindings; only the displayed
// vocabulary now follows keys.go automatically.
var helpSections = []helpSection{
	{"Global", []helpEntry{
		he(keys.Palette),
		he(keys.Command),
		{"/", "filter the list (slug, desc, tags)"}, // the list's own filter (no keyMap binding)
		he(keys.FilterMode),
		he(keys.Sort), he(keys.SortRev),
		he(keys.StatusView), he(keys.StatusRev),
		he(keys.Action),
		he(keys.Edit),
		he(keys.OpenEditor),
		he(keys.Follow),
		he(keys.JumpBack),
		he(keys.Yank), he(keys.YankPath),
		he(keys.PrevTab), he(keys.NextTab),
		he(keys.ToggleFocus),
		he(keys.Zoom),
		he(keys.Refresh),
		he(keys.Help),
		he(keys.Quit), he(keys.ForceQuit),
	}},
	{"List", []helpEntry{
		{"j / k", "move down / up"},
		{"g / G", "top / bottom"},
		{"d / u", "page down / up (pgdn/pgup)"},
		he(keys.Right),
		he(keys.Left),
	}},
	{"Detail", []helpEntry{
		{"j / k", "scroll down / up"},
		{"ctrl+d / u", "half-page down / up"},
		he(keys.Top), he(keys.Bottom),
		he(keys.Find),
		he(keys.FindNext), he(keys.FindPrev),
		he(keys.RawToggle),
		he(keys.Back),
	}},
}

// symbolsFor builds the glyph legend for the active screen — what the leading
// status / liveness / bucket glyphs (and the ⚠/↻/✓ markers) in the rows actually
// mean, so a reader can decode the column without leaving the TUI. The glyph AND
// marker rows (tok) are sourced from the SAME theme tokens the row delegates + the
// dashboard draw (theme.Status/Liveness/Bucket/FindingStatus/Marker*), so they can't
// drift from what's on screen; only the percent-band rows (mark) are hand-labeled.
// ok is false on a screen with no glyph vocabulary of its own.
func symbolsFor(kind entityKind, s styles) (helpSection, bool) {
	tok := func(t theme.Token, desc string) helpEntry { return helpEntry{s.fg(t.Color, t.Glyph), desc} }
	mark := func(c theme.Color, label, desc string) helpEntry { return helpEntry{s.fg(c, label), desc} }
	var e []helpEntry
	switch kind {
	case entityTasks:
		for _, st := range domain.AllStatuses() {
			e = append(e, tok(theme.Status(st), string(st)))
		}
		e = append(e,
			tok(theme.MarkerWarn, "misfiled — status ≠ folder"),
			tok(theme.MarkerRevisit, "revisit (snooze) date reached"),
		)
	case entityEpics:
		e = append(e,
			tok(theme.Liveness("working"), "working — live work in progress"),
			tok(theme.Liveness("fresh"), "fresh — new epic, no tasks yet"),
			tok(theme.Liveness("dormant"), "dormant — drained / quiet (id dimmed)"),
			tok(theme.MarkerWarn, "non-conforming status (→ active/retired/deprecated)"),
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
			// ✓ ready-to-close (U+2713) and the ✔ all-clear (U+2714) are deliberately
			// distinct glyphs — both come from theme.Marker* so they can't drift.
			tok(theme.MarkerReadyToClose, "audit ready to close (findings resolved)"),
			tok(theme.MarkerWarn, "needs attention (misfiled / non-conforming)"),
			tok(theme.MarkerRevisit, "revisit (snooze) reached"),
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

// helpHFrame is the help box's horizontal chrome (border + padding), derived from
// the box's STYLE so a padding/border change can't desync the wrap width. The
// frame size is color-independent (it's the rounded border + Padding(0,2)), so it
// stays a package var rather than per-Model state — only the box's COLOR is themed
// (styles.helpBorder).
var helpHFrame = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Padding(0, 2).GetHorizontalFrameSize()

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
func helpSectionsFor(f focus, kind entityKind, s styles) []helpSection {
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
	if sym, ok := symbolsFor(kind, s); ok { // what the row glyphs mean, page-specific
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
func helpLines(f focus, kind entityKind, contentW int, s styles) []string {
	sections := helpSectionsFor(f, kind, s)
	// Widest key column across the shown sections → aligned descriptions.
	keyW := 0
	for _, sec := range sections {
		for _, e := range sec.entries {
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

	lines := []string{fit(s.helpHeading.Render("Keys"))}
	for _, sec := range sections {
		lines = append(lines, fit(""), fit(s.dim(sec.title)))
		for _, e := range sec.entries {
			keyPad := strings.Repeat(" ", max(keyW-lipgloss.Width(e.keys), 0))
			// wrap reflows + right-pads each description line to descW (lipgloss Width),
			// so the column stays rectangular and the rows align.
			desc := strings.Split(wrap(e.desc, descW), "\n")
			lines = append(lines, fit("  "+s.helpKey.Render(e.keys)+keyPad+"  "+desc[0]))
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
func helpBox(maxW, maxH, scroll int, f focus, kind entityKind, s styles) string {
	lines := helpLines(f, kind, helpWidth(maxW)-helpHFrame, s)
	const frameV = 2 // top+bottom border rows
	if innerH := maxH - frameV; innerH > 0 && len(lines) > innerH {
		maxScroll := len(lines) - innerH
		scroll = min(max(scroll, 0), maxScroll)
		lines = lines[scroll : scroll+innerH]
	}
	box := s.helpBorder.Render(strings.Join(lines, "\n"))
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
