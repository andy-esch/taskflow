package store

import (
	"bytes"
	"reflect"
	"strings"
	"testing"

	"github.com/andy-esch/taskflow/internal/domain"

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

func TestUpdateFrontmatterAppendsDeterministicallyWithoutSortingExistingKeys(t *testing.T) {
	for _, eol := range []string{"\n", "\r\n"} {
		in := []byte(strings.ReplaceAll("---\n# leading\nzeta: keep # zeta note\ntier: 2 # tier note\nuntouched: | # block note\n  first\n  second\nremove: drop\n---\n# Task\n\nbody bytes\n", "\n", eol))
		updates := map[string]any{"gamma": 3, "alpha": 1, "epsilon": 5, "delta": 4, "beta": 2, "tier": 3, "remove": domain.UnsetField{}}
		wantKeys := []string{"zeta", "tier", "untouched", "alpha", "beta", "delta", "epsilon", "gamma"}
		var baseline []byte
		for range 100 {
			out, err := updateFrontmatter(in, updates)
			if err != nil {
				t.Fatal(err)
			}
			fm, body := splitFrontmatter(out)
			var doc yaml.Node
			if err := yaml.Unmarshal(fm, &doc); err != nil {
				t.Fatal(err)
			}
			var keys []string
			for i := 0; i < len(doc.Content[0].Content); i += 2 {
				keys = append(keys, doc.Content[0].Content[i].Value)
			}
			if !reflect.DeepEqual(keys, wantKeys) {
				t.Fatalf("keys=%v, want %v\n%s", keys, wantKeys, out)
			}
			if string(body) != strings.ReplaceAll("# Task\n\nbody bytes\n", "\n", eol) ||
				!bytes.Contains(out, []byte("# leading")) || !bytes.Contains(out, []byte("# tier note")) ||
				!bytes.Contains(out, []byte(strings.ReplaceAll("untouched: | # block note\n  first\n  second\n", "\n", eol))) {
				t.Fatalf("untouched evidence changed:\n%s", out)
			}
			if baseline != nil && !bytes.Equal(out, baseline) {
				t.Fatal("identical updates produced different bytes")
			}
			baseline = out
		}
	}
}
