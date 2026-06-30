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

// styles is the per-Model theming bundle: the active palette plus every chrome
// lipgloss.Style derived from it, and the color render helpers as methods. It
// replaces the former package-global `pal` + chrome `var`s + `applyTheme` mutation
// — theming is now per-Model state (multi-session-safe and test-isolated). The
// root Model holds a *styles and the list delegates share that same pointer, so
// Run can repopulate it once (after background detection) and every surface sees
// the new palette without rebuilding.
type styles struct {
	pal design.Palette

	// Chrome styles, each derived from pal. selected/dim are color-independent
	// (bold/faint only) but live here for uniformity.
	selected     lipgloss.Style // the "› " row cursor (bold)
	dimStyle     lipgloss.Style // faint text (the dim() method renders through this)
	activeTab    lipgloss.Style
	paneActive   lipgloss.Style
	paneInactive lipgloss.Style

	dashHeading lipgloss.Style

	helpBorder  lipgloss.Style
	helpHeading lipgloss.Style
	helpKey     lipgloss.Style

	actionBorder  lipgloss.Style
	dangerBorder  lipgloss.Style
	actionHeading lipgloss.Style

	findMatch   lipgloss.Style
	findCurrent lipgloss.Style

	editAreaBox lipgloss.Style

	// Frame sizes derived from the pane style (not a hardcoded 2) so a future
	// border/padding change can't silently desync sizing. The border STYLE is
	// theme-independent, so these stay valid across a palette swap.
	paneHFrame int
	paneVFrame int
}

// newStyles builds every chrome style (and the derived frame sizes) from palette
// p — the single place chrome color is decided, replacing both the old per-file
// `var` declarations and applyTheme's in-place mutation. Each style keeps its exact
// structure (border, padding, bold) so the swap is purely a recolor.
func newStyles(p design.Palette) styles {
	accent := p.Accent.Color()
	paneActive := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(accent)
	return styles{
		pal: p,

		selected:     lipgloss.NewStyle().Bold(true),
		dimStyle:     lipgloss.NewStyle().Faint(true),
		activeTab:    lipgloss.NewStyle().Bold(true).Foreground(accent),
		paneActive:   paneActive,
		paneInactive: lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(p.BorderIdle.Color()),

		dashHeading: lipgloss.NewStyle().Bold(true).Foreground(p.Heading.Color()),

		helpBorder:  lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(p.BorderActive.Color()).Padding(0, 2),
		helpHeading: lipgloss.NewStyle().Bold(true).Foreground(p.Heading.Color()),
		helpKey:     lipgloss.NewStyle().Bold(true),

		actionBorder:  lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(p.BorderActive.Color()).Padding(0, 2),
		dangerBorder:  lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(p.Danger.Color()).Padding(0, 2),
		actionHeading: lipgloss.NewStyle().Bold(true).Foreground(p.Heading.Color()),

		findMatch:   lipgloss.NewStyle().Background(p.Match.Color()).Foreground(p.MatchFg.Color()),
		findCurrent: lipgloss.NewStyle().Background(p.MatchCurrent.Color()).Foreground(p.MatchFg.Color()).Bold(true),

		editAreaBox: lipgloss.NewStyle().Border(lipgloss.NormalBorder()).BorderForeground(p.BorderIdle.Color()).Padding(0, 1),

		paneHFrame: paneActive.GetHorizontalFrameSize(),
		paneVFrame: paneActive.GetVerticalFrameSize(),
	}
}

// lipColor maps a semantic theme.Color to its concrete palette color — the TUI's
// truecolor rendering of the same status semantics the CLI renders as ANSI.
func (s styles) lipColor(c theme.Color) color.Color { return s.pal.Of(c).Color() }

func (s styles) fg(c theme.Color, str string) string {
	return lipgloss.NewStyle().Foreground(s.lipColor(c)).Render(str)
}

// glyph renders a theme Token (status / bucket / liveness / marker) as its colored
// glyph — the shared shorthand for the fg(tok.Color, tok.Glyph) the rows + dashboard
// repeat, so a marker is drawn from theme rather than a re-typed literal.
func (s styles) glyph(t theme.Token) string { return s.fg(t.Color, t.Glyph) }

func (s styles) dim(str string) string { return s.dimStyle.Render(str) }

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
func (s styles) miniBar(pct, width int) string { return progressbar.Render(pct, width, s.pal) }

// segBar renders an audit's finding breakdown as a stacked bar (done/active/
// dropped over the open/empty track) via the shared progressbar package — the same
// renderer the CLI uses (Style.SegmentBar), so the two surfaces can't drift.
func (s styles) segBar(done, active, dropped, total, width int) string {
	return progressbar.RenderSegments(progressbar.Segments{Done: done, Active: active, Dropped: dropped, Total: total}, width, s.pal)
}

// statusText renders a colored glyph + status label.
func (s styles) statusText(st domain.Status) string {
	tok := theme.Status(st)
	return s.fg(tok.Color, tok.Glyph+" "+string(st))
}

// priorityText colors a priority label (empty stays empty).
func (s styles) priorityText(p string) string {
	if p == "" {
		return ""
	}
	return s.fg(theme.Priority(p), p)
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
