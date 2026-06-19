package render

import (
	"strings"

	"github.com/charmbracelet/glamour"
	"github.com/muesli/termenv"

	"github.com/andy-esch/taskflow/internal/theme"
)

// RenderBody returns body rendered as styled markdown (in the given glamour
// style) when styling is on and raw is false; otherwise it returns the raw source
// verbatim. That gate keeps the agent/porcelain contract intact: piped output,
// `--color=never`, `--json`, and an explicit `--raw` all get the unrendered body,
// byte-for-byte. A render error never fails a `show` — it falls back to the raw
// body.
func RenderBody(st Style, body, style string, raw bool) string {
	if raw || !st.on || strings.TrimSpace(body) == "" {
		return body
	}
	if style == "" {
		style = theme.MarkdownStyleDark
	}
	opts := []glamour.TermRendererOption{
		// We've already decided color is on (st.on). WithAutoStyle would pick the
		// uncolored "ascii" style off a TTY (so `--color=always` piped, and tests,
		// would get an unstyled body while the header is colored), so pin the
		// resolved style + a color profile for consistent, deterministic rendering.
		glamour.WithStandardStyle(style),
		glamour.WithColorProfile(termenv.ANSI256),
	}
	if st.width > 0 {
		opts = append(opts, glamour.WithWordWrap(st.width))
	}
	r, err := glamour.NewTermRenderer(opts...)
	if err != nil {
		return body
	}
	out, err := r.Render(body)
	if err != nil {
		return body
	}
	// glamour pads with blank lines; normalize to a single trailing newline so
	// the show layout (metadata header + a blank line + body) stays tight.
	return strings.TrimRight(out, "\n") + "\n"
}
