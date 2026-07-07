package store

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFenceSpans(t *testing.T) {
	content := []byte("a\n```\nfenced\n```\nb\n~~~\nalso\n~~~\nc\n")
	spans := fenceSpans(content)
	if len(spans) != 2 {
		t.Fatalf("want 2 fence spans, got %d: %v", len(spans), spans)
	}
	if !inSpans(spans, bytes.Index(content, []byte("fenced"))) {
		t.Error("content inside a ``` fence should be reported inside a span")
	}
	if !inSpans(spans, bytes.Index(content, []byte("also"))) {
		t.Error("content inside a ~~~ fence should be reported inside a span")
	}
	if inSpans(spans, bytes.Index(content, []byte("b\n"))) {
		t.Error("prose between fences should not be inside a span")
	}
}

// TestScanLinks_InlineRefAndCode: scanLinks surfaces inline and reference-style links but
// skips any link inside a fenced code block OR an inline code span (`[..](x.md)` as prose).
func TestScanLinks_InlineRefAndCode(t *testing.T) {
	content := []byte("[a](real.md)\n\n```\n[fenced](z.md)\n```\n\nExample: `[inline](code.md)`\n\n[lbl]: y.md\n")
	var targets []string
	for _, r := range scanLinks(content) {
		targets = append(targets, r.target)
	}
	got := strings.Join(targets, ",")
	if strings.Contains(got, "z.md") {
		t.Errorf("a link inside a fence must be skipped, got %q", got)
	}
	if strings.Contains(got, "code.md") {
		t.Errorf("a link inside an inline code span must be skipped, got %q", got)
	}
	if !strings.Contains(got, "real.md") || !strings.Contains(got, "y.md") {
		t.Errorf("a real inline link and a reference-style def must be found, got %q", got)
	}
}

// TestDanglingLinks_RefStyleAndFences: a broken reference-style link is flagged; a broken
// inline link shown as an example inside a fenced code block is not (it is not a real ref).
func TestDanglingLinks_RefStyleAndFences(t *testing.T) {
	root := t.TempDir()
	body := "# doc\n\nSee [the thing][t].\n\n[t]: 6fjangd7kvzz-missing.md\n\n" +
		"```\n[example](6fjangd7kvyy-alsogone.md)\n```\n"
	p := filepath.Join(root, "tasks", "6fjangd7kva1-a.md")
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}

	probs, err := NewFS(root).DanglingLinks()
	if err != nil {
		t.Fatal(err)
	}
	if len(probs) != 1 {
		t.Fatalf("want exactly 1 dangler (ref-style link; fenced example ignored), got %d: %+v", len(probs), probs)
	}
	if !strings.Contains(probs[0].Message, "6fjangd7kvzz-missing.md") {
		t.Errorf("the reference-style dangler should be the one flagged, got %q", probs[0].Message)
	}
}
