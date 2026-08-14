package cli

import (
	"path/filepath"
	"strings"
	"testing"
)

// Research is part of the top-level `lint` roster: a doc missing its required created
// date is an issue, and a well-formed one is silent. `created` is required because the
// id is minted from it.
func TestLint_CoversResearch(t *testing.T) {
	root := freshRepo(t)
	runRoot(t, "-C", root, "research", "new", "Good doc")
	mustWrite(t, filepath.Join(root, "research", "6dr29v000zzr-no-date.md"),
		"---\nschema: 1\nid: 6dr29v000zzr\n---\n# No date\n")

	out, err := runRootRC(t, "-C", root, "lint")
	if err == nil {
		t.Fatal("a research doc with no created date must fail lint")
	}
	if !strings.Contains(out, "no-date") || !strings.Contains(out, "created") {
		t.Errorf("lint should name the doc and the field, got:\n%s", out)
	}
	if strings.Contains(out, "good-doc") {
		t.Errorf("a well-formed research doc must not be flagged:\n%s", out)
	}
}

// A drifted frontmatter id (disagreeing with the filename id) is flagged for research
// the same way it is for tasks — the filename id is the canonical resolution key.
func TestLint_ResearchIDDrift(t *testing.T) {
	root := freshRepo(t)
	mustWrite(t, filepath.Join(root, "research", "6dr29v000zzr-drift.md"),
		"---\nschema: 1\nid: 6dr29v001111\ncreated: \"2026-01-03\"\n---\n# Drift\n")

	out, err := runRootRC(t, "-C", root, "lint")
	if err == nil {
		t.Fatal("a drifted research id must fail lint")
	}
	if !strings.Contains(out, "disagrees with the filename id") {
		t.Errorf("lint should flag the id drift, got:\n%s", out)
	}
}
