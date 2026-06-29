package render

import (
	"bytes"
	"strings"
	"testing"

	"github.com/andy-esch/taskflow/internal/design"
)

// A plain (color-off) preview is byte-stable: token -> hex lines + bar glyphs, no ANSI.
func TestThemePreviewHuman_Plain(t *testing.T) {
	var b bytes.Buffer
	pal := design.Default().Dark
	ThemePreviewHuman(&b, NewStyle(false), "neon", "dark", pal)
	out := b.String()
	if !strings.Contains(out, "neon (dark)") {
		t.Errorf("missing header:\n%s", out)
	}
	if !strings.Contains(out, "accent") || !strings.Contains(out, pal.Accent.Hex) {
		t.Errorf("missing accent swatch (token + hex):\n%s", out)
	}
	if strings.Contains(out, "\x1b[") {
		t.Errorf("plain preview must emit no ANSI:\n%q", out)
	}
}

func TestThemePreviewJSON(t *testing.T) {
	var b bytes.Buffer
	if err := ThemePreviewJSON(&b, "neon", "dark", design.Default().Dark); err != nil {
		t.Fatal(err)
	}
	out := b.String()
	for _, want := range []string{
		"schema_version", `"name":"neon"`, `"variant":"dark"`, "accent", design.Default().Dark.Accent.Hex,
	} {
		if !strings.Contains(out, want) {
			t.Errorf("JSON missing %q:\n%s", want, out)
		}
	}
}
