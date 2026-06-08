package store

import "testing"

func TestSplitFrontmatter(t *testing.T) {
	fm, body := splitFrontmatter([]byte("---\nstatus: x\nepic: y\n---\n# Title\nbody\n"))
	if string(fm) != "status: x\nepic: y\n" {
		t.Errorf("fm = %q", fm)
	}
	if string(body) != "# Title\nbody\n" {
		t.Errorf("body = %q", body)
	}
}

func TestSplitFrontmatter_None(t *testing.T) {
	content := []byte("# no frontmatter\nhi\n")
	fm, body := splitFrontmatter(content)
	if fm != nil {
		t.Errorf("expected nil fm, got %q", fm)
	}
	if string(body) != string(content) {
		t.Errorf("body changed: %q", body)
	}
}

func TestSplitFrontmatter_BodyWithFence(t *testing.T) {
	// A `---` inside the body (after the closing fence) must not confuse the split.
	fm, body := splitFrontmatter([]byte("---\nstatus: x\n---\nintro\n\n---\n\noutro\n"))
	if string(fm) != "status: x\n" {
		t.Errorf("fm = %q", fm)
	}
	if string(body) != "intro\n\n---\n\noutro\n" {
		t.Errorf("body = %q", body)
	}
}
