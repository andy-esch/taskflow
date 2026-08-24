package store

import (
	"errors"
	"strings"
	"testing"

	"github.com/andy-esch/taskflow/internal/domain"
)

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

// The id is the canonical key the filename encodes, so no writer may put an illegal one on
// disk. This is the single choke point every field write passes through — including
// `task set --force`, which bypasses the known-field registry and was observed persisting
// an invalid id before this guard existed.
func TestUpdateFrontmatterRejectsAnInvalidID(t *testing.T) {
	content := []byte("---\nschema: 1\nid: 6fbj870001t6\nstatus: ready-to-start\n---\n# T\n")
	for _, bad := range []string{"6fbj87000lt6", "6fbj87000ut6", "short", "6FBJ870001T6"} {
		if _, err := updateFrontmatter(content, map[string]any{"id": bad}); err == nil {
			t.Errorf("updateFrontmatter accepted the invalid id %q", bad)
		} else if !errors.Is(err, domain.ErrValidation) {
			t.Errorf("id %q rejected with %v; want a validation failure", bad, err)
		}
	}
	// A VALID id that disagrees with the filename is deliberately still allowed: that is
	// drift, which lint reports with a rename remedy, and banning it here would break the
	// fix pass that repairs exactly that.
	out, err := updateFrontmatter(content, map[string]any{"id": "6fbj870009t6"})
	if err != nil {
		t.Fatalf("a valid id must still be writable: %v", err)
	}
	if !strings.Contains(string(out), "id: 6fbj870009t6") {
		t.Errorf("the valid id was not written:\n%s", out)
	}
	// Unrelated updates are untouched by the guard.
	if _, err := updateFrontmatter(content, map[string]any{"status": "in-progress"}); err != nil {
		t.Errorf("a write with no id update was rejected: %v", err)
	}
}
