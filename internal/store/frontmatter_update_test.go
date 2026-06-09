package store

import (
	"strings"
	"testing"

	yaml "go.yaml.in/yaml/v3"
)

func TestUpdateFrontmatter_PreservesUnknownAndBody(t *testing.T) {
	in := []byte("---\nstatus: ready-to-start\nepic: 01-x\ncustom_field: keep-me\n---\n# Body\ncontent\n")

	out, err := updateFrontmatter(in, map[string]any{
		"status":     "in-progress",
		"started_at": "2026-06-07",
	})
	if err != nil {
		t.Fatal(err)
	}
	s := string(out)

	if !strings.Contains(s, "status: in-progress") {
		t.Errorf("status not updated:\n%s", s)
	}
	if !strings.Contains(s, "custom_field: keep-me") {
		t.Errorf("unknown field dropped (surgical write failed):\n%s", s)
	}
	if !strings.Contains(s, "started_at:") {
		t.Errorf("appended key missing:\n%s", s)
	}
	if !strings.Contains(s, "# Body\ncontent\n") {
		t.Errorf("body not preserved verbatim:\n%s", s)
	}

	// Output must be valid YAML (the whole point of dropping the fallback parser).
	fm, _ := splitFrontmatter(out)
	var check map[string]any
	if err := yaml.Unmarshal(fm, &check); err != nil {
		t.Fatalf("output frontmatter is not valid YAML: %v\n%s", err, fm)
	}
	if check["epic"] != "01-x" {
		t.Errorf("epic changed: %v", check["epic"])
	}
}

func TestUpdateFrontmatter_PreservesCommentsAndOrder(t *testing.T) {
	in := []byte("---\n" +
		"# leading comment\n" +
		"status: ready-to-start # inline note\n" +
		"epic: 01-x\n" +
		"tier: 2\n" +
		"---\nbody\n")

	out, err := updateFrontmatter(in, map[string]any{"status": "in-progress"})
	if err != nil {
		t.Fatal(err)
	}
	s := string(out)

	if !strings.Contains(s, "# leading comment") {
		t.Errorf("head comment dropped:\n%s", s)
	}
	if !strings.Contains(s, "# inline note") {
		t.Errorf("inline comment on the updated value dropped:\n%s", s)
	}
	// Existing keys must keep their relative order (surgical, in-place update).
	si, ei, ti := strings.Index(s, "status:"), strings.Index(s, "epic:"), strings.Index(s, "tier:")
	if si < 0 || si >= ei || ei >= ti {
		t.Errorf("key order changed (want status<epic<tier):\n%s", s)
	}
}

func TestUpdateFrontmatter_RejectsNonMapping(t *testing.T) {
	// Valid YAML, but a sequence rather than a key/value mapping. Mutating it
	// must error, not silently overwrite (which would lose the original data).
	in := []byte("---\n- a\n- b\n---\nbody\n")
	if _, err := updateFrontmatter(in, map[string]any{"status": "in-progress"}); err == nil {
		t.Fatal("expected error for non-mapping frontmatter, got nil")
	}
}
