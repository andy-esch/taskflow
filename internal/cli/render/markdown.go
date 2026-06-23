package render

import (
	"strings"

	"charm.land/glamour/v2"

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
		// Pin the resolved style (not WithAutoStyle, which picks the uncolored
		// "ascii" style off a TTY — so `--color=always` piped + tests would get an
		// unstyled body under a colored header). glamour v2 owns color-profile
		// detection (colorprofile), so v1's WithColorProfile is gone.
		glamour.WithStandardStyle(style),
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
