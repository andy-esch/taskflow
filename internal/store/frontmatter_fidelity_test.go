package store

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/andy-esch/taskflow/internal/core"
	"github.com/andy-esch/taskflow/internal/domain"
	"github.com/andy-esch/taskflow/internal/testutil"
	yaml "go.yaml.in/yaml/v3"
)

// A yaml.Node round trip can change decoded folded values, not just their wrapping.
// Assert both semantic and byte fidelity through every existing-document encoder.
func TestFrontmatterEditsPreserveUntouchedBlockScalars(t *testing.T) {
	for _, indicator := range []string{">", ">-", ">+", "|", "|-", "|+", ">2", "|2-"} {
		for _, eol := range []string{"\n", "\r\n"} {
			t.Run(indicator+"/"+map[string]string{"\n": "LF", "\r\n": "CRLF"}[eol], func(t *testing.T) {
				untouched := "# keep this explanation\nnotes: " + indicator + " # keep this comment\n" +
					"  first wrapped line\n  second wrapped line\n\n    more indented\n  last line\n\n" +
					"custom:\n    nested: |-\n      nested text\n        nested indentation\n    alias: &memo >-\n      anchored text\n      still wrapped\n    reference: *memo\n"
				untouched = strings.ReplaceAll(untouched, "\n", eol)
				content := []byte("---" + eol + untouched + "tier: 2" + eol +
					"depends_on: [6fbj870001t6, 6fbj870001t6]" + eol + "---" + eol + "# Body" + eol)
				wantValues := fidelityValues(t, content)
				for i := 0; i < 5; i++ {
					var err error
					content, err = updateFrontmatter(content, map[string]any{"tier": i + 3})
					if err != nil {
						t.Fatal(err)
					}
					assertFidelity(t, content, untouched, wantValues)
				}
				content, err := replaceBodyStamped(content, "# Changed body\n", "2026-10-06")
				if err != nil {
					t.Fatal(err)
				}
				assertFidelity(t, content, untouched, wantValues)
				content, changed, err := updateDependencySourceEdits(content, []core.TaskGraphSourceEdit{{
					Field: core.TaskDependencyDependsOn, Action: core.TaskGraphSourceDedupe, Value: "6fbj870001t6",
				}}, "2026-10-06")
				if err != nil || !changed {
					t.Fatalf("repair: changed=%v err=%v", changed, err)
				}
				assertFidelity(t, content, untouched, wantValues)
			})
		}
	}
}

func TestBlockScalarEditsWithSiblingChanges(t *testing.T) {
	for _, indent := range []string{"", "  "} {
		t.Run("indent="+indent, func(t *testing.T) {
			content := "---\n# leading comment\n" + indent + "obsolete: old\n" +
				"# keep scalar comment\n" + indent + "notes: |+\n" + indent + "  keep\n\n\n" +
				"# keep following comment\n" + indent + "tier: 2 # keep inline\n---\nbody\n"
			want := fidelityValues(t, []byte(content))["notes"]
			out, err := updateFrontmatter([]byte(content), map[string]any{
				"obsolete": domain.UnsetField{}, "tier": 3, "updated_at": "2026-10-06",
			})
			if err != nil {
				t.Fatal(err)
			}
			if got := fidelityValues(t, out)["notes"]; got != want {
				t.Fatalf("keep chomping changed: got %q want %q", got, want)
			}
			for _, comment := range []string{"# leading comment", "# keep scalar comment", "# keep following comment", "# keep inline"} {
				if bytes.Count(out, []byte(comment)) != 1 {
					t.Fatalf("comment %q lost or duplicated:\n%s", comment, out)
				}
			}
		})
	}
}

func TestBlockScalarEditRefusesBrokenAliasBeforeWriting(t *testing.T) {
	content := []byte("---\nnotes: &memo >-\n  wrapped text\n  keep text\nreference: *memo\ntier: 2\n---\nbody\n")
	if out, err := updateFrontmatter(content, map[string]any{"notes": "new value"}); err == nil || out != nil {
		t.Fatalf("anchor replacement must fail closed: out=%q err=%v", out, err)
	}
}

func TestBlockScalarFirstRemainingEntryRetainsItsCommentWhenChanged(t *testing.T) {
	content := []byte("---\n# document preamble\nobsolete: old\n# notes explanation\nnotes: >-\n  old wrapped\n  text\nother: |-\n  untouched\n---\nbody\n")
	out, err := updateFrontmatter(content, map[string]any{
		"obsolete": domain.UnsetField{}, "notes": "replacement",
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, comment := range []string{"# document preamble", "# notes explanation"} {
		if bytes.Count(out, []byte(comment)) != 1 {
			t.Fatalf("comment %q lost or duplicated:\n%s", comment, out)
		}
	}
}

func TestBlockScalarFieldWritesPreserveValuesAndRefusalLeavesFileUntouched(t *testing.T) {
	root := t.TempDir()
	untouched := "notes: &memo >-\n  wrapped text\n  keep text\n\n    indented text\n  last line\nreference: *memo\n"
	path, content := testutil.TaskFixture(root, "not-started", "probe.md",
		"---\nschema: 1\nepic: 01-x\ntags: [probe]\ntier: 2\n"+untouched+"---\nbody\n")
	testutil.Write(t, path, content)
	fs := testutil.Must(NewFS(root, core.UnrestrictedMutations()))
	want := fidelityValues(t, []byte(content))
	for i := range 5 {
		if _, err := fs.SetFields("probe", map[string]any{"tier": 2 + i%3}, false); err != nil {
			t.Fatal(err)
		}
		after, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		assertFidelity(t, after, untouched, want)
	}
	before, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := fs.SetFields("probe", map[string]any{"notes": "would break the alias"}, false); !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("unsafe field write must fail closed: %v", err)
	}
	after, err := os.ReadFile(path)
	if err != nil || !bytes.Equal(before, after) {
		t.Fatalf("refusal changed the file: err=%v\n%s", err, after)
	}
}

func TestBlockScalarExplicitReplacement(t *testing.T) {
	content := []byte("---\nnotes: >-\n  old text\n  wrapped text\nother: |-\n  untouched\n---\nbody\n")
	out, err := updateFrontmatter(content, map[string]any{"notes": "first\n  indented\nlast\n"})
	if err != nil {
		t.Fatal(err)
	}
	if got := fidelityValues(t, out)["notes"]; got != "first\n  indented\nlast\n" {
		t.Fatalf("replacement decoded as %q", got)
	}
	if !bytes.Contains(out, []byte("other: |-\n  untouched\n")) {
		t.Fatalf("untouched entry changed:\n%s", out)
	}
}

func TestBlockScalarParagraphCommentsAllowSurgicalWrites(t *testing.T) {
	for _, eol := range []string{"\n", "\r\n"} {
		for _, indent := range []string{"", "  "} {
			for _, blanks := range []int{0, 1, 2, 3, 5} {
				t.Run(fmt.Sprintf("eol=%q/indent=%d/blanks=%d", eol, len(indent), blanks), func(t *testing.T) {
					scalar := indent + "notes: |+\n" + indent + "  keep\n\n\n"
					comment := indent + "# paragraph 1\n" + strings.Repeat(indent+"\n", blanks) + indent + "# paragraph 2\n"
					original := []byte(strings.ReplaceAll("---\n"+scalar+comment+indent+"tier: 2\n"+
						indent+"depends_on: [6fbj870001t6, 6fbj870001t6]\n---\nbody\n", "\n", eol))
					wantNotes := fidelityValues(t, original)["notes"]
					for _, operation := range []string{"unrelated field", "commented field", "body", "dependency", "remove commented field"} {
						var out []byte
						var err error
						switch operation {
						case "unrelated field":
							out, err = updateFrontmatter(original, map[string]any{"updated_at": "2026-10-07"})
						case "commented field":
							out, err = updateFrontmatter(original, map[string]any{"tier": 3})
						case "body":
							out, err = replaceBodyStamped(original, "new body\n", "2026-10-07")
						case "dependency":
							out, _, err = updateDependencySourceEdits(original, []core.TaskGraphSourceEdit{{
								Field: core.TaskDependencyDependsOn, Action: core.TaskGraphSourceDedupe, Value: "6fbj870001t6",
							}}, "2026-10-07")
						case "remove commented field":
							out, err = updateFrontmatter(original, map[string]any{"tier": domain.UnsetField{}})
						}
						if err != nil {
							t.Fatalf("%s: valid comment paragraphs froze a write: %v", operation, err)
						}
						if !bytes.Contains(out, []byte(strings.ReplaceAll(scalar, "\n", eol))) || fidelityValues(t, out)["notes"] != wantNotes {
							t.Fatalf("%s: comment backtracking consumed keep-chomp scalar bytes/value:\n%s", operation, out)
						}
						if operation != "commented field" && operation != "remove commented field" && !bytes.Contains(out, []byte(strings.ReplaceAll(comment, "\n", eol))) {
							t.Fatalf("%s: untouched comment paragraph spacing changed:\n%s", operation, out)
						}
					}
				})
			}
		}
	}
}

func TestBlockScalarParagraphCommentsAllowFieldWrites(t *testing.T) {
	for _, eol := range []string{"\n", "\r\n"} {
		t.Run(fmt.Sprintf("eol=%q", eol), func(t *testing.T) {
			root := t.TempDir()
			scalar := strings.ReplaceAll("notes: |+\n  keep\n\n\n", "\n", eol)
			path, content := testutil.TaskFixture(root, "not-started", "paragraph-comments.md",
				strings.ReplaceAll("---\nschema: 1\nepic: 01-x\ntags: [probe]\nnotes: |+\n  keep\n\n\n"+
					"# paragraph 1\n\n\n\n# paragraph 2\ntier: 2\n---\nbody\n", "\n", eol))
			before := []byte(content)
			testutil.Write(t, path, content)
			fs := testutil.Must(NewFS(root, core.UnrestrictedMutations()))
			if _, err := fs.SetFields("paragraph-comments", map[string]any{"tier": 3}, false); err != nil {
				t.Fatal(err)
			}
			after, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			assertFidelity(t, after, scalar, fidelityValues(t, before))
			if fidelityValues(t, after)["tier"] != 3 {
				t.Fatalf("field write did not apply:\n%s", after)
			}
			for _, comment := range []string{"# paragraph 1", "# paragraph 2"} {
				if bytes.Count(after, []byte(comment)) != 1 {
					t.Fatalf("comment %q lost or duplicated:\n%s", comment, after)
				}
			}
		})
	}
}

func TestBlockScalarDependencyRepairRefusesDecodedDriftWithValidYAML(t *testing.T) {
	root := t.TempDir()
	prerequisiteID := testutil.TaskID("fidelity-prerequisite")
	writeGraphMutationTask(t, root, "fidelity-prerequisite", domain.StatusCompleted, nil, "")
	// A hand-authored invalid declaration may remain while --auto removes a
	// different duplicate. Editing that sequence makes yaml.v3 refold the retained
	// scalar into valid YAML with a different decoded value. No fabricated nodes.
	dependencies := "depends_on:\n  - >\n    first line\n      indented line\n  - " + prerequisiteID + "\n  - " + prerequisiteID + "\n"
	path, content := testutil.TaskFixture(root, "not-started", "fidelity-owner.md",
		"---\nschema: 1\nepic: 01-x\ntags: [probe]\ntier: 2\n"+dependencies+"---\nbody\n")
	before := []byte(content)
	testutil.Write(t, path, content)
	edit := core.TaskGraphSourceEdit{Field: core.TaskDependencyDependsOn, Action: core.TaskGraphSourceDedupe, Value: prerequisiteID}
	if out, _, err := updateDependencySourceEdits(before, []core.TaskGraphSourceEdit{edit}, "2026-10-07"); out != nil || !errors.Is(err, domain.ErrValidation) || !strings.Contains(err.Error(), "decoded values differ from the requested edit") {
		t.Fatalf("expected decoded-value refusal rather than a syntax error: out=%q err=%v", out, err)
	}
	fs := testutil.Must(NewFS(root, core.UnrestrictedMutations()))
	service := core.MustNewService(fs)
	for _, dryRun := range []bool{true, false} {
		receipt, err := service.RepairTaskGraph(core.TaskGraphRepairRequest{Auto: true}, dryRun)
		if !errors.Is(err, domain.ErrValidation) || !strings.Contains(err.Error(), "decoded values differ from the requested edit") || receipt.Committed || len(receipt.AppliedFiles) != 0 {
			t.Fatalf("dry-run=%v: unsafe serialization was not refused before effects: receipt=%+v err=%v", dryRun, receipt, err)
		}
		after, err := os.ReadFile(path)
		if err != nil || !bytes.Equal(after, before) {
			t.Fatalf("dry-run=%v: refused graph repair changed file bytes: err=%v\n%s", dryRun, err, after)
		}
	}
}

func fidelityValues(t *testing.T, content []byte) map[string]any {
	t.Helper()
	fm, _ := splitFrontmatter(content)
	var values map[string]any
	if err := yaml.Unmarshal(fm, &values); err != nil {
		t.Fatal(err)
	}
	return values
}

func assertFidelity(t *testing.T, content []byte, untouched string, want map[string]any) {
	t.Helper()
	if !bytes.Contains(content, []byte(untouched)) {
		t.Fatalf("untouched scalar source changed:\n%s", content)
	}
	got := fidelityValues(t, content)
	for _, key := range []string{"notes", "custom"} {
		if !reflect.DeepEqual(got[key], want[key]) {
			t.Fatalf("decoded %s changed: got %#v want %#v", key, got[key], want[key])
		}
	}
}
