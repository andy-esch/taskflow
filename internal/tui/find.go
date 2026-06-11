package tui

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/lipgloss"
)

// finder is the detail-pane "/" text search: vim-like find-in-body with n/N
// navigation. bubbles/viewport has no built-in search, so matches are tracked by
// line index in the (already width-wrapped) content — the same index the
// viewport scrolls by. Matches are highlighted; the current one stands out.
type finder struct {
	input  textinput.Model
	typing bool   // the query is being entered
	query  string // applied query ("" = inactive)
	lines  []int  // content line indices that contain a match
	cur    int    // index into lines — the focused match
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

// highlightOccurrences wraps every case-insensitive occurrence of query in plain
// with style. plain must be ANSI-free (we rebuild matched lines from stripped
// text, so a highlight can't land inside an existing escape sequence).
func highlightOccurrences(plain, query string, style lipgloss.Style) string {
	if query == "" {
		return plain
	}
	q := strings.ToLower(query)
	var b strings.Builder
	rest, low := plain, strings.ToLower(plain)
	for {
		i := strings.Index(low, q)
		if i < 0 {
			b.WriteString(rest)
			return b.String()
		}
		b.WriteString(rest[:i])
		b.WriteString(style.Render(rest[i : i+len(q)]))
		rest, low = rest[i+len(q):], low[i+len(q):]
	}
}
