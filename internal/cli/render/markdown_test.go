package render

import (
	"strings"
	"testing"

	"github.com/andy-esch/taskflow/internal/theme"
)

func TestRenderBody(t *testing.T) {
	const md = "# Title\n\nsome **bold** text\n"
	style := theme.MarkdownStyleDark

	// Disabled style (pipe / --color=never): raw, byte-for-byte.
	if got := RenderBody(NewStyle(false), md, style, false); got != md {
		t.Errorf("disabled style should return the raw body, got %q", got)
	}
	// raw=true forces the source even when styling is on.
	if got := RenderBody(NewStyle(true), md, style, true); got != md {
		t.Errorf("--raw should return the raw body, got %q", got)
	}
	// Empty body stays empty (no glamour padding).
	if got := RenderBody(NewStyle(true), "", style, false); got != "" {
		t.Errorf("empty body should stay empty, got %q", got)
	}
	// Enabled + not raw: rendered — ANSI present, the literal markdown gone.
	got := RenderBody(NewStyle(true), md, style, false)
	if !strings.Contains(got, "\x1b[") {
		t.Errorf("rendered body should contain ANSI: %q", got)
	}
	if strings.Contains(got, "# Title") {
		t.Errorf("rendered body should not contain the literal '# Title': %q", got)
	}
	// An empty style falls back to the dark default rather than erroring.
	if fb := RenderBody(NewStyle(true), md, "", false); !strings.Contains(fb, "\x1b[") {
		t.Errorf("empty style should fall back and still render: %q", fb)
	}
}
