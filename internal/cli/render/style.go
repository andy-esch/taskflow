package render

import (
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/x/ansi"

	"github.com/andy-esch/taskflow/internal/domain"
	"github.com/andy-esch/taskflow/internal/progressbar"
	"github.com/andy-esch/taskflow/internal/theme"
)

// ANSI SGR codes. Kept minimal and 16-color so they degrade well everywhere.
const (
	ansiReset  = "\x1b[0m"
	ansiBold   = "\x1b[1m"
	ansiDim    = "\x1b[2m"
	ansiRed    = "\x1b[31m"
	ansiGreen  = "\x1b[32m"
	ansiYellow = "\x1b[33m"
	ansiBlue   = "\x1b[34m"
	ansiCyan   = "\x1b[36m"
	ansiGray   = "\x1b[90m"
)

// ansiCode maps a semantic theme.Color to its ANSI SGR code ("" for none).
func ansiCode(c theme.Color) string {
	switch c {
	case theme.ColorRed:
		return ansiRed
	case theme.ColorGreen:
		return ansiGreen
	case theme.ColorYellow:
		return ansiYellow
	case theme.ColorBlue:
		return ansiBlue
	case theme.ColorCyan:
		return ansiCyan
	case theme.ColorGray:
		return ansiGray
	default:
		return ""
	}
}

// Style renders optional ANSI styling and carries the output width (0 = no
// limit). The zero value is disabled (plain) with no width cap — so tests,
// piped output, and the JSON path are never colored or truncated unless asked.
type Style struct {
	on    bool
	width int
}

// NewStyle returns a Style; enabled controls whether ANSI is emitted.
func NewStyle(enabled bool) Style { return Style{on: enabled} }

// WithWidth returns a copy capped to a terminal width (0 leaves it uncapped, so
// piped output keeps full-width rows).
func (s Style) WithWidth(w int) Style { s.width = w; return s }

func (s Style) wrap(code, text string) string {
	if !s.on || text == "" || code == "" {
		return text
	}
	return code + text + ansiReset
}

// Bold / Dim style arbitrary text.
func (s Style) Bold(t string) string { return s.wrap(ansiBold, t) }
func (s Style) Dim(t string) string  { return s.wrap(ansiDim, t) }

// Link wraps text in an OSC 8 terminal hyperlink to url when styling is on (a
// TTY), so supporting terminals make it click-to-open; otherwise it returns text
// unchanged — keeping piped / --color=never / JSON output byte-stable. url should
// be absolute (e.g. file:///abs/path). Gated on the same `on` flag as color, so
// it never reaches a pipe or an agent.
func (s Style) Link(text, url string) string {
	if !s.on || text == "" || url == "" {
		return text
	}
	const (
		osc8 = "\x1b]8;;"
		st   = "\x1b\\" // String Terminator
	)
	return osc8 + url + st + text + osc8 + st
}

// Status renders a status: a colored glyph + label when styled (semantics from
// the shared theme), the plain label otherwise (so output stays byte-stable).
func (s Style) Status(st domain.Status) string {
	if !s.on {
		return string(st)
	}
	tok := theme.Status(st)
	return ansiCode(tok.Color) + tok.Glyph + " " + string(st) + ansiReset
}

// Bucket renders an audit bucket: a colored glyph + name when styled (the shared
// theme's shape carries the state), the plain name otherwise — mirroring Status,
// so the porcelain path stays byte-stable.
func (s Style) Bucket(b string) string {
	if !s.on {
		return b
	}
	tok := theme.Bucket(domain.AuditBucket(b))
	return ansiCode(tok.Color) + tok.Glyph + " " + b + ansiReset
}

// FindingStatus renders an audit finding's status the way Status renders a task
// status — colored glyph + label when styled, plain label otherwise. An empty
// status (a finding missing its **Status:**) renders as-is so the table cell
// stays blank rather than showing a lone glyph.
func (s Style) FindingStatus(status string) string {
	if !s.on || status == "" {
		return status
	}
	tok := theme.FindingStatus(status)
	return ansiCode(tok.Color) + tok.Glyph + " " + status + ansiReset
}

// Priority colors a priority label.
func (s Style) Priority(p string) string {
	return s.wrap(ansiCode(theme.Priority(p)), p)
}

// Percent colors a completion percentage: gray <34, yellow <100, green at 100.
func (s Style) Percent(pct int) string {
	return s.wrap(ansiCode(theme.Percent(pct)), fmt.Sprintf("%d%%", pct))
}

// Green / Red / Warn style status glyphs and success/error/warning text.
func (s Style) Green(t string) string { return s.wrap(ansiGreen, t) }
func (s Style) Red(t string) string   { return s.wrap(ansiRed, t) }
func (s Style) Warn(t string) string  { return s.wrap(ansiYellow, t) }

// Bar renders a width-char progress bar for pct (0–100) via the shared progressbar
// package (same constructor + neon gradient the TUI's miniBar uses, so the two
// surfaces can't drift). The bar always emits lipgloss ANSI, so when styling is off
// (piped / --json / tests) we strip it back to plain glyphs — keeping the porcelain
// contract byte-stable.
func (s Style) Bar(pct, width int) string {
	if width <= 0 {
		width = 10
	}
	out := progressbar.Render(pct, width)
	if !s.on {
		return ansi.Strip(out)
	}
	return out
}

// SegmentBar renders an audit's finding breakdown as a stacked bar (done/active/
// dropped bands over an open/empty track) via the shared progressbar package — the
// same renderer the TUI uses, so the surfaces can't drift. Like Bar, it strips the
// ANSI back to plain (distinct) glyphs when styling is off, keeping porcelain
// byte-stable.
func (s Style) SegmentBar(done, active, dropped, total, width int) string {
	if width <= 0 {
		width = 10
	}
	out := progressbar.RenderSegments(progressbar.Segments{Done: done, Active: active, Dropped: dropped, Total: total}, width)
	if !s.on {
		return ansi.Strip(out)
	}
	return out
}

// visibleWidth is the DISPLAY-CELL width of s ignoring ANSI escapes — so
// colored cells align (tabwriter counts escape bytes) and wide runes
// (CJK/emoji occupy two cells) don't shift columns the way a rune count did.
func visibleWidth(s string) int {
	return ansi.StringWidth(s)
}

// truncate shortens a plain string to max display cells with a trailing "…".
// Cells containing ANSI are returned unchanged (truncating them risks cutting
// an escape); the last column is plain in practice.
func truncate(s string, max int) string {
	if max <= 1 || strings.Contains(s, "\x1b") {
		return s
	}
	if ansi.StringWidth(s) <= max {
		return s
	}
	return ansi.Truncate(s, max, "…")
}

// writeTable prints a left-aligned, ANSI-aware table. When maxWidth > 0 and the
// natural layout would exceed it, the last column is truncated to fit (never
// below its header width). maxWidth <= 0 means no limit (piped output).
func writeTable(w io.Writer, maxWidth int, header []string, rows [][]string) {
	if len(rows) == 0 {
		return
	}
	cols := len(header)
	for _, r := range rows {
		if len(r) > cols {
			cols = len(r)
		}
	}
	width := make([]int, cols)
	measure := func(cells []string) {
		for c := 0; c < len(cells); c++ {
			if vw := visibleWidth(cells[c]); vw > width[c] {
				width[c] = vw
			}
		}
	}
	measure(header)
	for _, r := range rows {
		measure(r)
	}

	// Fit to maxWidth by shrinking the last column (descriptions/areas).
	if maxWidth > 0 && cols > 0 {
		last := cols - 1
		used := 0
		for c := 0; c < last; c++ {
			used += width[c] + 2 // column + gutter
		}
		avail := maxWidth - used
		headerMin := 0
		if last < len(header) {
			headerMin = visibleWidth(header[last]) // never truncate below the header label
		}
		if avail < headerMin {
			avail = headerMin
		}
		if avail > 0 && avail < width[last] {
			width[last] = avail
			for _, r := range rows {
				if last < len(r) {
					r[last] = truncate(r[last], avail)
				}
			}
		}
	}

	write := func(cells []string) {
		var b strings.Builder
		for c := 0; c < cols; c++ {
			cell := ""
			if c < len(cells) {
				cell = cells[c]
			}
			b.WriteString(cell)
			if c < cols-1 {
				b.WriteString(strings.Repeat(" ", width[c]-visibleWidth(cell)+2))
			}
		}
		line := strings.TrimRight(b.String(), " ")
		// Backstop the last-column shrink above: a wide *non-final* cell (a long
		// slug/component/id) can still push the row past maxWidth, so clamp the whole
		// composed line. ANSI-aware, so a colored cell isn't severed mid-escape;
		// piped output (maxWidth <= 0) stays full-width.
		if maxWidth > 0 {
			line = ansi.Truncate(line, maxWidth, "…")
		}
		fmt.Fprintln(w, line)
	}
	if len(header) > 0 {
		write(header)
	}
	for _, r := range rows {
		write(r)
	}
}
