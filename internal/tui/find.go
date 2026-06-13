package tui

import (
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

// matchPos is one occurrence of the find query: a content line and the byte range
// of the hit within that line's plain (ANSI-stripped) text. The line drives
// scroll; the byte range drives the highlight.
type matchPos struct {
	line   int
	b0, b1 int
}

// finder is the detail-pane "/" text search: vim-like find-in-body with n/N
// navigation over individual occurrences (not whole lines). bubbles/viewport has
// no built-in search, so matches are tracked by position.
type finder struct {
	input   textinput.Model
	typing  bool
	query   string
	matches []matchPos
	cur     int // index into matches — the focused occurrence
}

func newFinder() finder {
	ti := textinput.New()
	ti.Prompt = "/"
	ti.CharLimit = 64
	return finder{input: ti}
}

func (f finder) active() bool { return f.query != "" }

var (
	findMatch   = lipgloss.NewStyle().Background(lipgloss.Color("3")).Foreground(lipgloss.Color("0"))
	findCurrent = lipgloss.NewStyle().Background(lipgloss.Color("11")).Foreground(lipgloss.Color("0")).Bold(true)
)

// foldMatches returns the [start,end) byte ranges in s of every case-insensitive
// match of query. Comparison is rune-by-rune via unicode.ToLower, so a rune whose
// lowercase changes byte length (e.g. U+0130 İ) can't misalign offsets — the
// ranges always index s, never a folded copy. Matches are non-overlapping.
func foldMatches(s, query string) [][2]int {
	if query == "" {
		return nil
	}
	q := []rune(strings.ToLower(query))
	var out [][2]int
	for start := 0; start < len(s); {
		_, sz := utf8.DecodeRuneInString(s[start:])
		i, qi := start, 0
		for qi < len(q) && i < len(s) {
			r, n := utf8.DecodeRuneInString(s[i:])
			if unicode.ToLower(r) != q[qi] {
				break
			}
			i += n
			qi++
		}
		if qi == len(q) {
			out = append(out, [2]int{start, i})
			start = i // non-overlapping
		} else {
			start += sz
		}
	}
	return out
}

// highlightLine rebuilds one styled content line with each occurrence on it
// highlighted *over the styled text* (cutting on display columns with ansi.Cut),
// so the line's original field colors survive everywhere except the matched
// spans. The occurrence at byte offset curB0 gets the brighter "current" style
// (pass -1 for none). occ are this line's [b0,b1) ranges, ascending; plain is the
// ANSI-stripped line.
func highlightLine(styled, plain string, occ [][2]int, curB0 int) string {
	width := ansi.StringWidth(plain)
	var b strings.Builder
	prevCol := 0
	for _, r := range occ {
		c0 := ansi.StringWidth(plain[:r[0]])
		c1 := ansi.StringWidth(plain[:r[1]])
		b.WriteString(ansi.Cut(styled, prevCol, c0)) // unchanged span keeps its styling
		style := findMatch
		if r[0] == curB0 {
			style = findCurrent
		}
		b.WriteString(style.Render(plain[r[0]:r[1]])) // the matched text, restyled
		prevCol = c1
	}
	b.WriteString(ansi.Cut(styled, prevCol, width))
	return b.String()
}
