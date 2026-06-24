package render

import (
	"strings"
	"testing"

	"github.com/andy-esch/taskflow/internal/theme"
)

func TestRenderBody(t *testing.T) {
	const md = "# Title\n\nsome **bold** text\n"
	style := func() string { return theme.MarkdownStyleDark }

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
	if fb := RenderBody(NewStyle(true), md, func() string { return "" }, false); !strings.Contains(fb, "\x1b[") {
		t.Errorf("empty style should fall back and still render: %q", fb)
	}
}

// TestRenderBody_StyleProviderLazy pins the OSC-11 fix: the style provider (which
// queries the terminal background) must run ONLY when styled markdown is actually
// rendered — never on the raw / color-off / empty-body paths where its result is
// discarded.
func TestRenderBody_StyleProviderLazy(t *testing.T) {
	const md = "# Title\n"
	calls := 0
	style := func() string { calls++; return theme.MarkdownStyleDark }

	RenderBody(NewStyle(false), md, style, false) // color off
	RenderBody(NewStyle(true), md, style, true)   // --raw
	RenderBody(NewStyle(true), "", style, false)  // empty body
	if calls != 0 {
		t.Fatalf("style provider must not run when the body isn't rendered (the OSC-11 query would fire), got %d calls", calls)
	}

	RenderBody(NewStyle(true), md, style, false) // renders → provider runs once
	if calls != 1 {
		t.Errorf("style provider should run exactly once when rendering, got %d", calls)
	}
}
