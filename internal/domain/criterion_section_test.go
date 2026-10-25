package domain

import (
	"errors"
	"slices"
	"strings"
	"testing"
)

func TestCriterionSectionSelection(t *testing.T) {
	actual := []Criterion{{Index: 1, Text: "actual", State: CriterionUnmet}}
	met := []Criterion{{Index: 1, Text: "actual", Checked: true, State: CriterionMet}}
	for _, tc := range []struct {
		name, body, problem string
		criteria            []Criterion
	}{
		{"canonical after narrative", "# Acceptance work\n### Integrated acceptance notes\n- [ ] narrative\n## Acceptance criteria\n- [ ] actual\n## Progress on acceptance\n- [ ] also narrative\n", "", actual},
		{"annotated case and closing hashes", "## aCcEpTaNcE cRiTeRiA (v1 — draft) ###\n- [x] actual\n", "", met},
		{"legacy alias", "## ACCEPTANCE ##\n- [ ] actual\n", "", actual},
		{"nested headings", "## Delivery\n### Acceptance criteria\n#### Runtime\n- [ ] actual\n### Notes\n- [ ] outside\n", "", actual},
		{"fenced headings", "~~~markdown\n## Acceptance\n- [ ] example\n~~~\n## Acceptance criteria\n```\n## Acceptance\n- [ ] example\n```\n- [ ] actual\n", "", actual},
		{"CRLF", "### Integrated acceptance notes\r\nprose\r\n## Acceptance criteria\r\n- [ ] actual\r\n", "", actual},
		{"empty canonical", "## Acceptance criteria\n\n## Notes\n- [ ] outside\n", "", nil},
		{"narrative without checklist", "## Integrated acceptance notes\nprose\n", "", nil},
		{"ordinary checklist", "## Notes\n- [ ] not criteria\n", "", nil},
		{"title mentioning acceptance", "# Improve acceptance test runner\n## Notes\n- [ ] ordinary note\n", "", nil},
		{"title matching canonical", "# Acceptance criteria\n## Notes\n- [ ] ordinary note\n", "", nil},
		{"title matching alias with real section", "# Acceptance\n## Acceptance criteria\n- [ ] actual\n", "", actual},
		{"canonical title with real section", "# Acceptance criteria\n## Acceptance criteria\n- [ ] actual\n", "", actual},
		{"ordinary child of narrative", "# Normal title\n## Integrated acceptance notes\nprose\n### Notes\n- [ ] ordinary note\n", "", nil},
		{"ordinary grandchild of narrative", "## Acceptance notes\n### Notes\n#### Runtime\n- [ ] ordinary note\n", "", nil},
		{"title closes real section", "## Acceptance criteria\n- [ ] actual\n# Acceptance\n- [ ] ordinary title note\n", "", actual},
		{"duplicate canonical", "## Acceptance criteria\n- [x] first\n## Acceptance criteria\n- [ ] second\n", "2 acceptance-criteria sections at lines 1, 3", nil},
		{"canonical and alias", "## Acceptance\n\n## Acceptance criteria (draft)\n- [ ] actual\n", "2 acceptance-criteria sections at lines 1, 3", nil},
		{"nested duplicate", "## Acceptance criteria\n### Acceptance\n- [ ] actual\n", "2 acceptance-criteria sections at lines 1, 2", nil},
		{"nonstandard alone", "## Notes / acceptance\n- [ ] actual\n", "unrecognized acceptance checklist heading \"Notes / acceptance\" at line 1", nil},
		{"nonstandard suffix", "## Acceptance criteria discussion\n- [ ] actual\n", "unrecognized acceptance checklist heading \"Acceptance criteria discussion\" at line 1", nil},
		{"nonstandard smashed words", "## AcceptanceCriteria\n- [ ] actual\n", "unrecognized acceptance checklist heading \"AcceptanceCriteria\" at line 1", nil},
		{"nonstandard malformed checklist", "## Notes / acceptance\n- [] actual\n", "unrecognized acceptance checklist heading \"Notes / acceptance\" at line 1", nil},
		{"title does not own repair", "# Align acceptance lint\n\n## Objective\n- [ ] ordinary note\n\n## AcceptanceCriteria\n- [ ] actual\n", "unrecognized acceptance checklist heading \"AcceptanceCriteria\" at line 6", nil},
		{"innermost heading owns repair", "## Integrated acceptance notes\n### Acceptance checklist\n- [ ] actual\n", "unrecognized acceptance checklist heading \"Acceptance checklist\" at line 2", nil},
		{"direct narrative checklist still refuses", "## Acceptance notes\n- [ ] actual\n### Notes\nprose\n", "unrecognized acceptance checklist heading \"Acceptance notes\" at line 1", nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			count, err := CountAcceptanceCriteria(tc.body)
			criteria, listErr := ListAcceptanceCriteria(tc.body)
			_, tallyErr := TallyCriteria(tc.body)
			_, unexplainedErr := UnexplainedCriteria(tc.body)
			for _, readErr := range []error{err, listErr, tallyErr, unexplainedErr} {
				if tc.problem != "" {
					if !errors.Is(readErr, ErrValidation) || !strings.Contains(readErr.Error(), tc.problem) {
						t.Fatalf("read error = %v, want validation containing %q", readErr, tc.problem)
					}
				} else if readErr != nil {
					t.Fatal(readErr)
				}
			}
			if tc.problem != "" {
				issues := LintAcceptanceCriteria(tc.body)
				if count != (ACCount{}) || criteria != nil || len(issues) != 1 || !strings.Contains(issues[0].Message, tc.problem) {
					t.Fatal("invalid structure fabricated a selected list or disappeared from lint")
				}
			} else {
				wantCount := ACCount{Total: len(tc.criteria)}
				for _, criterion := range tc.criteria {
					if criterion.Checked {
						wantCount.Checked++
					}
				}
				if count != wantCount || !slices.Equal(criteria, tc.criteria) || len(LintAcceptanceCriteria(tc.body)) != 0 {
					t.Fatalf("wrong selection: count=%+v criteria=%+v, want count=%+v criteria=%+v; lint=%+v", count, criteria, wantCount, tc.criteria, LintAcceptanceCriteria(tc.body))
				}
			}
		})
	}
}

func TestCriterionStructureRefusesEveryEdit(t *testing.T) {
	for _, body := range []string{
		"## Acceptance\r\n- [x] first\r\n## Acceptance criteria\r\n- [ ] second\r\n",
		"## Notes / acceptance\n- [ ] unknown\n",
	} {
		for name, edit := range map[string]func(string) (string, error){
			"add":     func(b string) (string, error) { return AddCriterion(b, "new") },
			"remove":  func(b string) (string, error) { return RemoveCriterion(b, 1) },
			"replace": func(b string) (string, error) { return ReplaceCriterionText(b, 1, "new") },
			"check":   func(b string) (string, error) { return SetAcceptanceCriterion(b, 1, true) },
			"state":   func(b string) (string, error) { return SetCriterionState(b, 1, CriterionMet, "") },
			"multi":   func(b string) (string, error) { return SetCriteriaState(b, []int{1}, CriterionMet, "") },
		} {
			t.Run(name, func(t *testing.T) {
				out, err := edit(body)
				if !errors.Is(err, ErrValidation) || out != "" {
					t.Fatalf("invalid structure produced a writable body: %q, %v", out, err)
				}
			})
		}
	}
}

func TestCriterionEditsSelectCanonicalWithoutChangingProse(t *testing.T) {
	body := "### Integrated acceptance notes\n- [ ] narrative\n\n## Acceptance criteria (draft)\n- [ ] actual\n\n## Notes\n- [ ] outside\n"
	out, err := SetAcceptanceCriterion(body, 1, true)
	if err != nil || out != strings.Replace(body, "- [ ] actual", "- [x] actual", 1) {
		t.Fatalf("edit reached narrative content: %q, %v", out, err)
	}
	section, ok := Section(body, "acceptance")
	if !ok || !strings.HasPrefix(section, "### Integrated acceptance notes") {
		t.Fatal("criterion policy changed the general Section helper")
	}
}
