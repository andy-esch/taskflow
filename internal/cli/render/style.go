package render

import (
	"fmt"
	"io"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/andy-esch/taskflow/internal/domain"
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

// Enabled reports whether styling is active.
func (s Style) Enabled() bool { return s.on }

func (s Style) wrap(code, text string) string {
	if !s.on || text == "" {
		return text
	}
	return code + text + ansiReset
}

// Bold / Dim style arbitrary text.
func (s Style) Bold(t string) string { return s.wrap(ansiBold, t) }
func (s Style) Dim(t string) string  { return s.wrap(ansiDim, t) }

// Status renders a status: a colored glyph + label when styled, the plain label
// otherwise (so non-color output stays byte-stable).
func (s Style) Status(st domain.Status) string {
	if !s.on {
		return string(st)
	}
	glyph, code := statusGlyph(st)
	return code + glyph + " " + string(st) + ansiReset
}

func statusGlyph(st domain.Status) (glyph, code string) {
	switch st {
	case domain.StatusInProgress:
		return "●", ansiYellow
	case domain.StatusNextUp:
		return "●", ansiBlue
	case domain.StatusReadyToStart:
		return "○", ansiCyan
	case domain.StatusCompleted:
		return "✔", ansiGreen
	case domain.StatusDeprecated:
		return "✘", ansiRed
	case domain.StatusDeferred:
		return "◌", ansiGray
	default:
		return "•", ansiGray
	}
}

// Bucket colors an audit bucket name.
func (s Style) Bucket(b string) string {
	switch b {
	case "open":
		return s.wrap(ansiYellow, b)
	case "closed":
		return s.wrap(ansiGreen, b)
	case "deferred":
		return s.wrap(ansiGray, b)
	default:
		return b
	}
}

// Priority colors a priority label.
func (s Style) Priority(p string) string {
	switch p {
	case "high":
		return s.wrap(ansiRed, p)
	case "medium":
		return s.wrap(ansiYellow, p)
	case "low":
		return s.wrap(ansiGray, p)
	default:
		return p
	}
}

// Percent colors a completion percentage: gray <34, yellow <100, green at 100.
func (s Style) Percent(pct int) string {
	txt := fmt.Sprintf("%d%%", pct)
	switch {
	case pct >= 100:
		return s.wrap(ansiGreen, txt)
	case pct >= 34:
		return s.wrap(ansiYellow, txt)
	default:
		return s.wrap(ansiGray, txt)
	}
}

// Green / Red / Warn style status glyphs and success/error/warning text.
func (s Style) Green(t string) string { return s.wrap(ansiGreen, t) }
func (s Style) Red(t string) string   { return s.wrap(ansiRed, t) }
func (s Style) Warn(t string) string  { return s.wrap(ansiYellow, t) }

// RelativeDate renders a YYYY-MM-DD date as a compact "today" / "3d ago" /
// "2w ago" / "5mo ago" / "1y ago". Empty or unparseable input yields "".
func RelativeDate(date string) string { return relativeDateFrom(date, time.Now()) }

func relativeDateFrom(date string, now time.Time) string {
	t, err := time.Parse(time.DateOnly, date)
	if err != nil {
		return ""
	}
	days := int(now.Sub(t).Hours() / 24)
	switch {
	case days < 0:
		return date // future date — show it verbatim rather than "−3d"
	case days == 0:
		return "today"
	case days == 1:
		return "yesterday"
	case days < 7:
		return fmt.Sprintf("%dd ago", days)
	case days < 30:
		return fmt.Sprintf("%dw ago", days/7)
	case days < 365:
		return fmt.Sprintf("%dmo ago", days/30)
	default:
		return fmt.Sprintf("%dy ago", days/365)
	}
}

var ansiRe = regexp.MustCompile("\x1b\\[[0-9;]*m")

// visibleWidth is the rune width of s ignoring ANSI escapes — so colored cells
// align correctly (text/tabwriter counts escape bytes, which breaks alignment).
func visibleWidth(s string) int {
	return utf8.RuneCountInString(ansiRe.ReplaceAllString(s, ""))
}

// writeTable prints a left-aligned, ANSI-aware table. Cells are already styled
// by the caller; columns pad to their max visible width with a 2-space gutter,
// and the last column isn't padded. Nothing is written for an empty body.
// truncate shortens a plain string to max visible runes with a trailing "…".
// Cells containing ANSI are returned unchanged (truncating them risks cutting an
// escape); the last column is plain in practice.
func truncate(s string, max int) string {
	if max <= 1 || strings.Contains(s, "\x1b") {
		return s
	}
	r := []rune(s)
	if len(r) <= max {
		return s
	}
	return string(r[:max-1]) + "…"
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
		fmt.Fprintln(w, strings.TrimRight(b.String(), " "))
	}
	if len(header) > 0 {
		write(header)
	}
	for _, r := range rows {
		write(r)
	}
}
