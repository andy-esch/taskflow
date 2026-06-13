package tui

import (
	"strings"

	"github.com/charmbracelet/glamour"
)

// glamourStyle is the markdown render style — "dark" is the safe default for
// terminal use; per-theme styling is a deferred polish.
const glamourStyle = "dark"

// newGlamourRenderer builds a renderer wrapped to width. Constructing it (goldmark
// + chroma style setup) is the CPU-heavy part, so the detail pane caches one per
// width (see detailPane.prettyBody) rather than rebuilding per selection.
func newGlamourRenderer(width int) (*glamour.TermRenderer, error) {
	return glamour.NewTermRenderer(
		glamour.WithStandardStyle(glamourStyle),
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
func glamourBody(md string, width int) string {
	r, err := newGlamourRenderer(width)
	if err != nil {
		return wrap(md, width)
	}
	out, ok := renderMarkdown(r, md)
	if !ok {
		return wrap(md, width)
	}
	return out
}
