package tui

import (
	"image/color"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/andy-esch/taskflow/internal/design"
	"github.com/andy-esch/taskflow/internal/domain"
	"github.com/andy-esch/taskflow/internal/progressbar"
	"github.com/andy-esch/taskflow/internal/theme"
)

// pal is the active palette every chrome style derives from. It defaults to the
// neon theme's dark variant at package init — so the styles built below (and tests
// that render without Run) have a real palette — and Run swaps in the background-
// appropriate palette via applyTheme before the program starts. A package var (not
// a Model field) because the chrome styles below and the free fns (fg/lipColor) are
// already package-global; this routes every color through one palette without
// threading it through every render call.
var pal = design.Default().Dark

// accent is the focus/selection accent, shared by the active pane border and the
// active tab in the strip. lipgloss v2 Color is a func returning a color.Color
// value (not a const string type), so this is a var.
var accent = pal.Accent.Color()

var (
	selectedStyle = lipgloss.NewStyle().Bold(true)
	dimStyle      = lipgloss.NewStyle().Faint(true)
	activeTab     = lipgloss.NewStyle().Bold(true).Foreground(accent)

	// Two focus signals: an accent border + a bold title on the focused pane.
	paneActive   = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(accent)
	paneInactive = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(pal.BorderIdle.Color())

	// Frame sizes derived from the pane style (not a hardcoded 2) so a future
	// border/padding change can't silently desync sizing. The border STYLE is
	// theme-independent, so these stay valid when applyTheme recolors paneActive.
	paneHFrame = paneActive.GetHorizontalFrameSize()
	paneVFrame = paneActive.GetVerticalFrameSize()
)

// applyTheme repoints every package-global chrome style at palette p, recoloring
// in place so each style keeps its structure (border, padding, bold) and the
// frame-size derivations stay valid. Called once from Run with the background-
// appropriate palette; a future live-retheme can call it again on a
// BackgroundColorMsg. The styles assigned here are declared across sibling files
// (dashHeading, helpBorder, …) but are package-scoped, so this is the single place
// chrome color is decided — no stray literals.
func applyTheme(p design.Palette) {
	pal = p
	accent = p.Accent.Color()
	activeTab = activeTab.Foreground(accent)
	paneActive = paneActive.BorderForeground(accent)
	paneInactive = paneInactive.BorderForeground(p.BorderIdle.Color())
	dashHeading = dashHeading.Foreground(p.Heading.Color())
	helpBorder = helpBorder.BorderForeground(p.BorderActive.Color())
	helpHeading = helpHeading.Foreground(p.Heading.Color())
	actionBorder = actionBorder.BorderForeground(p.BorderActive.Color())
	dangerBorder = dangerBorder.BorderForeground(p.Danger.Color())
	actionHeading = actionHeading.Foreground(p.Heading.Color())
	findMatch = findMatch.Background(p.Match.Color()).Foreground(p.MatchFg.Color())
	findCurrent = findCurrent.Background(p.MatchCurrent.Color()).Foreground(p.MatchFg.Color())
	editAreaBox = editAreaBox.BorderForeground(p.BorderIdle.Color())
}

// lipColor maps a semantic theme.Color to its concrete palette color — the TUI's
// truecolor rendering of the same status semantics the CLI renders as ANSI. Reads
// the active palette live, so it follows an applyTheme swap.
func lipColor(c theme.Color) color.Color { return pal.Of(c).Color() }

func fg(c theme.Color, s string) string {
	return lipgloss.NewStyle().Foreground(lipColor(c)).Render(s)
}

// glyph renders a theme Token (status / bucket / liveness / marker) as its colored
// glyph — the shared shorthand for the fg(tok.Color, tok.Glyph) the rows + dashboard
// repeat, so a marker is drawn from theme rather than a re-typed literal.
func glyph(t theme.Token) string { return fg(t.Color, t.Glyph) }

func dim(s string) string { return dimStyle.Render(s) }

// osc8 wraps s in an OSC 8 terminal hyperlink to url, so supporting terminals make
// it click-to-open (they typically underline it, which is the affordance). The TUI
// doesn't capture the mouse, so clicks reach the terminal, not the app.
func osc8(s, url string) string {
	return "\x1b]8;;" + url + "\x1b\\" + s + "\x1b]8;;\x1b\\"
}

// miniBar renders the epic rollup bar (epic-list rows, epic-detail line) via the
// shared progressbar package — same constructor + neon palette the CLI status bar
// uses, so the two surfaces can't drift. The % text beside the bar carries the
// discrete completion tier color (theme.Percent).
func miniBar(pct, width int) string { return progressbar.Render(pct, width) }

// segBar renders an audit's finding breakdown as a stacked bar (done/active/
// dropped over the open/empty track) via the shared progressbar package — the same
// renderer the CLI uses (Style.SegmentBar), so the two surfaces can't drift.
func segBar(done, active, dropped, total, width int) string {
	return progressbar.RenderSegments(progressbar.Segments{Done: done, Active: active, Dropped: dropped, Total: total}, width)
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
