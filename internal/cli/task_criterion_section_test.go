package cli

import (
	"errors"
	"maps"
	"path/filepath"
	"strings"
	"testing"

	"github.com/andy-esch/taskflow/internal/domain"
	"github.com/andy-esch/taskflow/internal/testutil"
)

func TestCriterionStructureCLIRefusesBeforeEffects(t *testing.T) {
	for name, body := range map[string]string{
		"duplicate":               "## Acceptance\r\n- [x] first\r\n## Acceptance criteria\r\n- [ ] second\r\n",
		"unknown":                 "## Notes / acceptance\n- [ ] unknown\n",
		"title and unknown child": "# Align acceptance lint\n\n## Objective\n- [ ] ordinary note\n\n## Integrated acceptance notes\n### AcceptanceChecklist\n- [ ] unknown\n",
	} {
		for _, completed := range []bool{false, true} {
			t.Run(name+"/completed="+map[bool]string{false: "false", true: "true"}[completed], func(t *testing.T) {
				root := setupRepo(t)
				if completed {
					runRoot(t, "-C", root, "task", "complete", "alpha")
				}
				// Body replacement remains available to repair invalid structure.
				runRoot(t, "-C", root, "task", "set", "alpha", "--body", body)
				before := testutil.SnapshotTree(t, root)
				commands := [][]string{
					{"task", "info", "alpha", "--json"},
					{"task", "ac", "alpha", "--json"},
					{"lint", "--json"},
					{"lint", "--fix", "--json"},
				}
				for _, mutation := range [][]string{
					{"task", "ac", "alpha", "--check", "1"},
					{"task", "ac", "alpha", "--add", "new"},
					{"task", "ac", "alpha", "--remove", "1"},
					{"task", "ac", "alpha", "--replace", "1", "--text", "new"},
					{"task", "ac", "alpha", "--defer", "1", "--reason", "later"},
					{"task", "ac", "alpha", "--tracked", "1", "--to", "beta", "--reason", "handoff"},
					{"task", "complete", "alpha"},
					{"task", "complete", "alpha", "--force"},
					{"task", "move", "alpha", "completed", "--force"},
				} {
					commands = append(commands, mutation)
					preview := append([]string{}, mutation...)
					commands = append(commands, append(preview, "--dry-run"))
				}
				for _, command := range commands {
					out, err := runRootRC(t, append([]string{"-C", root}, command...)...)
					if !errors.Is(err, domain.ErrValidation) || ExitCode(err) != 11 {
						t.Fatalf("%v bypassed structure: %v\n%s", command, err, out)
					}
					if !strings.Contains(out+err.Error(), "acceptance") || strings.Contains(out, `"task_info":`) || strings.Contains(out, `"acceptance":[]`) {
						t.Fatalf("%v hid the diagnostic behind a success tally: %s, %v", command, out, err)
					}
					if name == "title and unknown child" && !strings.Contains(strings.ReplaceAll(out+err.Error(), `\"`, `"`), `"AcceptanceChecklist" at line 7`) {
						t.Fatalf("%v blamed the title or narrative ancestor instead of the offending heading: %s, %v", command, out, err)
					}
					if !maps.Equal(before, testutil.SnapshotTree(t, root)) {
						t.Fatalf("%v changed bytes or modes before refusing", command)
					}
				}
				// Repair previews describe proposed fixes, not the post-repair lint
				// pass performed by real --fix. They still must not rewrite headings.
				runRoot(t, "-C", root, "lint", "--fix", "--json", "--dry-run")
				if !maps.Equal(before, testutil.SnapshotTree(t, root)) {
					t.Fatal("repair preview rewrote criterion structure")
				}
				if out := runRoot(t, "-C", root, "task", "show", "alpha", "--raw"); !strings.Contains(out, "acceptance") && !strings.Contains(out, "Acceptance") {
					t.Fatal("raw body repair access disappeared")
				}
				if name == "title and unknown child" {
					repaired := strings.Replace(body, "### AcceptanceChecklist", "### Acceptance criteria", 1)
					runRoot(t, "-C", root, "task", "set", "alpha", "--body", repaired)
					if out := runRoot(t, "-C", root, "task", "info", "alpha", "--json"); !strings.Contains(out, `"total":1`) {
						t.Fatalf("heading-only repair absorbed ordinary checklist: %s", out)
					}
					if out := runRoot(t, "-C", root, "task", "show", "alpha", "--raw"); !strings.Contains(out, repaired) {
						t.Fatalf("heading-only repair changed the title or narrative: %s", out)
					}
				}
				runRoot(t, "-C", root, "task", "set", "alpha", "--body", "## Acceptance criteria\n- [ ] real\n")
				out := runRoot(t, "-C", root, "task", "info", "alpha", "--json")
				if !strings.Contains(out, `"total":1`) {
					t.Fatalf("explicit repair did not restore the real tally: %s", out)
				}
				runRoot(t, "-C", root, "task", "complete", "alpha", "--force")
			})
		}
	}
}

func TestCriterionTitlesAndNarrativeChildrenRemainOrdinary(t *testing.T) {
	for _, title := range []string{"Acceptance", "Acceptance criteria"} {
		t.Run(title, func(t *testing.T) {
			root := freshRepo(t)
			mustWrite(t, filepath.Join(root, "epics", "01-quality.md"), "---\nstatus: active\npriority: medium\ndescription: quality\n---\n# Quality\n")
			runRoot(t, "-C", root, "task", "new", title, "--epic", "01-quality", "--tags", "test")
			slug := domain.Slugify(title)
			// Exercise the actual creation scaffold, not just a headless fragment.
			runRoot(t, "-C", root, "lint", "--links")
			runRoot(t, "-C", root, "task", "info", slug, "--json")
			runRoot(t, "-C", root, "task", "ac", slug, "--add", "real criterion")
			raw := runRoot(t, "-C", root, "task", "show", slug, "--raw")
			if !strings.Contains(raw, "# "+title+"\n") {
				t.Fatalf("criterion edit changed task title: %s", raw)
			}
			body := "# " + title + "\n\n## Integrated acceptance notes\nprose\n### Notes\n- [ ] ordinary note\n"
			runRoot(t, "-C", root, "task", "set", slug, "--body", body)
			before := testutil.SnapshotTree(t, root)
			out := runRoot(t, "-C", root, "task", "info", slug, "--json")
			if !strings.Contains(out, `"checked":0,"total":0`) {
				t.Fatalf("ordinary checklist became criteria: %s", out)
			}
			runRoot(t, "-C", root, "task", "ac", slug, "--json")
			runRoot(t, "-C", root, "lint", "--links")
			runRoot(t, "-C", root, "task", "complete", slug, "--dry-run")
			if !maps.Equal(before, testutil.SnapshotTree(t, root)) {
				t.Fatal("read or completion preview modified an ordinary checklist")
			}
			runRoot(t, "-C", root, "task", "complete", slug)
		})
	}
}

func TestCriterionCanonicalCLISelectionDoesNotRewriteNarrative(t *testing.T) {
	root := setupRepo(t)
	body := "### Integrated acceptance notes\n- [ ] narrative\n\n## Acceptance criteria (draft)\n- [ ] real\n\n## Progress on acceptance\n- [ ] outside\n"
	runRoot(t, "-C", root, "task", "set", "alpha", "--body", body)
	out := runRoot(t, "-C", root, "task", "info", "alpha", "--json")
	if !strings.Contains(out, `"checked":0,"total":1`) {
		t.Fatalf("narrative shadowed canonical criteria: %s", out)
	}
	if _, err := runRootRC(t, "-C", root, "task", "complete", "alpha", "--dry-run"); !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("real unmet criterion failed to gate completion: %v", err)
	}
	runRoot(t, "-C", root, "task", "ac", "alpha", "--check", "1")
	actual := runRoot(t, "-C", root, "task", "show", "alpha", "--raw")
	if !strings.Contains(actual, strings.Replace(body, "- [ ] real", "- [x] real", 1)) {
		t.Fatalf("criterion edit changed narrative content: %s", actual)
	}
	runRoot(t, "-C", root, "task", "complete", "alpha")
}
