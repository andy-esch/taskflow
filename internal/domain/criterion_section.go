package domain

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// Criterion section selection is shared by reads, edits, lint and completion.
// A missing section is valid; an ambiguous or unrecognized checklist is not an
// empty section. Callers must preserve this distinction instead of showing 0/0.
// See docs/architecture/planning-data.md#criterion-section-selection.
type criterionSection struct {
	start, end, level int
	title             string
	checklist         bool
}

var criterionHeadingRe = regexp.MustCompile(`(?i)^acceptance criteria(?:[ \t]+\([^\r\n]*\))?$`)

func isAcceptanceSectionHeading(title string) bool {
	title = strings.TrimSpace(bodyATXClosingRe.ReplaceAllString(title, ""))
	return strings.EqualFold(title, "acceptance") || criterionHeadingRe.MatchString(title)
}

// selectCriterionSection scans all headings before choosing: a later canonical
// section cannot be shadowed by prose, and two genuine sections never become
// first-wins. Only unfenced checklist evidence makes a nonstandard heading a
// repair problem, and only when there is no recognized section. H1s are document
// titles, not criterion sections. Unknown headings own only their direct content;
// recognized criterion sections also include their deeper subsections.
func selectCriterionSection(lines []string) (criterionSection, error) {
	var spans []criterionSection
	var open, recognized []int
	var fence fenceScanner
	direct := -1
	for i, line := range lines {
		if fence.inCode(line) {
			continue
		}
		if h := bodyHeadingRe.FindStringSubmatch(line); h != nil {
			level := len(h[1])
			direct = -1
			for len(open) > 0 && spans[open[len(open)-1]].level >= level {
				spans[open[len(open)-1]].end = i
				open = open[:len(open)-1]
			}
			// Keep the old broad match as diagnostic evidence, never selection:
			// formerly counted forms such as AcceptanceCriteria must not vanish.
			if level >= 2 && strings.Contains(strings.ToLower(h[2]), "acceptance") {
				index := len(spans)
				spans = append(spans, criterionSection{start: i, end: len(lines), level: level, title: h[2]})
				open = append(open, index)
				direct = index
				if isAcceptanceSectionHeading(h[2]) {
					recognized = append(recognized, index)
				}
			}
			continue
		}
		_, malformed := malformedCheckbox(line)
		if direct >= 0 && (bodyCheckboxRe.MatchString(line) || malformed) {
			spans[direct].checklist = true
		}
	}
	if len(recognized) > 1 {
		var locations []string
		for _, index := range recognized {
			locations = append(locations, strconv.Itoa(spans[index].start+1))
		}
		return criterionSection{}, fmt.Errorf("%w: %d acceptance-criteria sections at lines %s; merge them into one `Acceptance criteria` section before reading, editing or completing",
			ErrValidation, len(recognized), strings.Join(locations, ", "))
	}
	if len(recognized) == 1 {
		return spans[recognized[0]], nil
	}
	for _, span := range spans {
		if span.checklist {
			return criterionSection{}, fmt.Errorf("%w: unrecognized acceptance checklist heading %q at line %d; rename it to `Acceptance criteria` (or `Acceptance`) before reading, editing or completing",
				ErrValidation, span.title, span.start+1)
		}
	}
	return criterionSection{start: -1, end: -1}, nil
}
