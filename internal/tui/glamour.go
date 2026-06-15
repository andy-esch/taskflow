package tui

import (
	"strings"

	"github.com/charmbracelet/glamour"
)

// glamourStyleFor picks the glamour standard-style for the terminal background:
// "light" on a light terminal, "dark" (the safe default) otherwise. The
// background is resolved ONCE at startup (Run) and threaded through — a
// mid-program OSC background query would race Bubble Tea's input reader.
func glamourStyleFor(darkBG bool) string {
	if darkBG {
		return "dark"
	}
	return "light"
}

// newGlamourRenderer builds a renderer wrapped to width, in the given standard
// style (empty falls back to "dark"). Constructing it (goldmark + chroma style
// setup) is the CPU-heavy part, so the detail pane caches one per width+style
// (see detailPane.prettyBody) rather than rebuilding per selection.
func newGlamourRenderer(width int, style string) (*glamour.TermRenderer, error) {
	if style == "" {
		style = "dark"
	}
	return glamour.NewTermRenderer(
		glamour.WithStandardStyle(style),
		glamour.WithWordWrap(max(width-2, 20)), // leave room for glamour's left margin
	)
}

// renderMarkdown renders md with r, trimming the blank lines glamour pads with so
// the body sits flush under the field block. An empty body renders empty.
func renderMarkdown(r *glamour.TermRenderer, md string) (string, bool) {
	if strings.TrimSpace(md) == "" {
		return "", true
	}
	out, err := r.Render(md)
	if err != nil {
		return "", false
	}
	return strings.Trim(out, "\n"), true
}

// glamourBody renders markdown with a fresh renderer — the uncached path used by
// tests and as a fallback. It is called from Update, NEVER View. On any error
// (e.g. a renderer that won't build) it returns plain wrapped text.
func glamourBody(md string, width int, style string) string {
	r, err := newGlamourRenderer(width, style)
	if err != nil {
		return wrap(md, width)
	}
	out, ok := renderMarkdown(r, md)
	if !ok {
		return wrap(md, width)
	}
	return out
}
