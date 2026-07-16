package domain

import (
	"fmt"
	"regexp"
	"strings"
)

// ACCount is the acceptance-criteria checkbox tally of a task body: how many of
// the criteria are checked out of the total. A body with no acceptance-criteria
// section (or one with no checkboxes) has a zero tally.
type ACCount struct {
	Checked int
	Total   int
}

// Criterion is one acceptance-criteria checkbox for `task ac --list`: its 1-based
// position, whether it's checked, and the first-line text after the checkbox.
type Criterion struct {
	Index   int
	Checked bool
	Text    string
}

// acCheckbox is an acceptance-criteria checkbox located in a body: its 0-based line
// index (so a flip can rewrite exactly that line) and current state/text.
type acCheckbox struct {
	line    int
	checked bool
	text    string
}

// The body-structure model is line-oriented and code-fence aware: a `##` heading
// or a `- [ ]` checkbox inside a fenced block is example prose, not structure, so
// the scanners skip fenced lines — the same "don't treat code as structure"
// discipline scanLinks uses for links.
var (
	bodyHeadingRe  = regexp.MustCompile(`^(#{1,6})[ \t]+(.*\S)[ \t]*$`)
	bodyCheckboxRe = regexp.MustCompile(`^[ \t]*[-*+][ \t]+\[([ xX])\]`)
)

// fenceAt reports whether line is a fenced-code delimiter: a run of >=3 backticks
// or tildes after optional indentation. It returns the fence character, the run
// length, and the text after the run — an info string on an OPENING fence, which
// must be blank for a valid CLOSING fence. ok is false for any non-fence line.
func fenceAt(line string) (marker byte, length int, rest string, ok bool) {
	i := 0
	for i < len(line) && (line[i] == ' ' || line[i] == '\t') {
		i++
	}
	if i >= len(line) || (line[i] != '`' && line[i] != '~') {
		return 0, 0, "", false
	}
	marker = line[i]
	j := i
	for j < len(line) && line[j] == marker {
		j++
	}
	if j-i < 3 {
		return 0, 0, "", false
	}
	return marker, j - i, line[j:], true
}

// fenceScanner tracks fenced-code state across a body's lines. It honors the fence
// CHARACTER and LENGTH rather than a naive toggle: a block closes only on a line of
// the same marker, at least as long, with no trailing info string — so a
// shorter/different inner fence, or an info-stringed line like ```go, stays INSIDE
// the block (CommonMark §4.5). A naive `inFence = !inFence` toggle would wrongly
// close on either and leak the nested content into structure scanning.
type fenceScanner struct {
	open   bool
	marker byte
	length int
}

// inCode advances the scanner by one line and reports whether that line is code (a
// fence delimiter or content inside a fence) and so must be skipped by structure
// scanning.
func (f *fenceScanner) inCode(line string) bool {
	m, l, rest, ok := fenceAt(line)
	if !ok {
		return f.open
	}
	if !f.open {
		f.open, f.marker, f.length = true, m, l
		return true
	}
	if m == f.marker && l >= f.length && strings.TrimSpace(rest) == "" {
		f.open = false
	}
	return true
}

// Section returns the markdown block for the FIRST heading whose title contains
// name (case-insensitive), from that heading through the line before the next
// heading of the same or higher level — nested deeper headings stay inside.
// Trailing blank lines are trimmed. ok is false when no heading matches. Headings
// inside fenced code blocks are ignored so an example `##` never matches. CRLF line
// endings are tolerated (normalized to LF).
func Section(body, name string) (text string, ok bool) {
	lines := strings.Split(normalizeNewlines(body), "\n")
	q := strings.ToLower(strings.TrimSpace(name))
	start, level := -1, 0
	var fence fenceScanner
	for i, ln := range lines {
		if fence.inCode(ln) {
			continue
		}
		m := bodyHeadingRe.FindStringSubmatch(ln)
		if m == nil {
			continue
		}
		lvl := len(m[1])
		if start == -1 {
			if strings.Contains(strings.ToLower(m[2]), q) {
				start, level = i, lvl
			}
			continue
		}
		if lvl <= level { // a same-or-higher heading closes the section
			return trimTrailingBlankLines(lines[start:i]), true
		}
	}
	if start == -1 {
		return "", false
	}
	return trimTrailingBlankLines(lines[start:]), true
}

// scanAcceptanceCheckboxes returns the body split on "\n" (newline-normalized) plus
// the task-list checkboxes inside its acceptance-criteria section — the first heading
// whose title contains "acceptance", up to the next heading of the same or higher
// level. Fence-aware in a single pass (a `##`/`- [ ]` inside a code fence is example
// prose, not structure). One scanner backs the tally, the list, and the flip.
func scanAcceptanceCheckboxes(body string) (lines []string, boxes []acCheckbox) {
	lines = strings.Split(normalizeNewlines(body), "\n")
	var (
		fence      fenceScanner
		inSection  bool
		sectionLvl int
	)
	for i, ln := range lines {
		if fence.inCode(ln) {
			continue
		}
		if m := bodyHeadingRe.FindStringSubmatch(ln); m != nil {
			lvl := len(m[1])
			switch {
			case !inSection:
				if strings.Contains(strings.ToLower(m[2]), "acceptance") {
					inSection, sectionLvl = true, lvl
				}
			case lvl <= sectionLvl: // a same-or-higher heading ends the section
				return lines, boxes
			}
			continue
		}
		if inSection {
			if m := bodyCheckboxRe.FindStringSubmatch(ln); m != nil {
				boxes = append(boxes, acCheckbox{line: i, checked: m[1] == "x" || m[1] == "X", text: checkboxText(ln)})
			}
		}
	}
	return lines, boxes
}

// CountAcceptanceCriteria tallies the acceptance-criteria checkboxes. No such
// section — or none with checkboxes — yields a zero tally.
func CountAcceptanceCriteria(body string) ACCount {
	_, boxes := scanAcceptanceCheckboxes(body)
	var c ACCount
	for _, b := range boxes {
		c.Total++
		if b.checked {
			c.Checked++
		}
	}
	return c
}

// ListAcceptanceCriteria returns the acceptance criteria in body order, 1-based —
// the `task ac --list` view an agent then flips by index.
func ListAcceptanceCriteria(body string) []Criterion {
	_, boxes := scanAcceptanceCheckboxes(body)
	out := make([]Criterion, len(boxes))
	for i, b := range boxes {
		out[i] = Criterion{Index: i + 1, Checked: b.checked, Text: b.text}
	}
	return out
}

// SetAcceptanceCriterion flips the 1-based nth acceptance-criteria checkbox to
// checked/unchecked, returning the new body. Only that one checkbox's `[ ]`/`[x]`
// is rewritten — every other byte (frontmatter is handled upstream) is preserved.
// It is idempotent: flipping to the current state returns the body unchanged (the
// caller can skip the write). ErrValidation when there's no acceptance section or n
// is out of range.
func SetAcceptanceCriterion(body string, n int, checked bool) (string, error) {
	lines, boxes := scanAcceptanceCheckboxes(body)
	if len(boxes) == 0 {
		return "", fmt.Errorf("%w: task has no acceptance criteria to %s", ErrValidation, checkWord(checked))
	}
	if n < 1 || n > len(boxes) {
		return "", fmt.Errorf("%w: criterion %d out of range (have %d)", ErrValidation, n, len(boxes))
	}
	box := boxes[n-1]
	if box.checked == checked {
		return body, nil // already in the target state — no-op
	}
	lines[box.line] = flipCheckbox(lines[box.line], checked)
	return strings.Join(lines, "\n"), nil
}

// Misconfiguration guards for `task ac` / the `ac:` tally, which key off the first
// heading containing "acceptance" and count only well-formed checkboxes. A list item
// whose bracket holds only spaces/tabs/x/X but ISN'T the canonical `[ ]`/`[x]`/`[X]`
// (e.g. `[]`, `[ x]`, `[  ]`) is a botched checkbox that the tally silently drops.
// The class is deliberately narrow — `[1]`, `[-]`, and `[text](url)` links are NOT
// flagged — so a lint warning here is high-confidence, not noise.
var (
	acListItemRe   = regexp.MustCompile(`^[ \t]*[-*+][ \t]+(.*)$`)
	acCheckboxOKRe = regexp.MustCompile(`^\[[ xX]\]`)    // the canonical, valid marker
	acCheckboxyRe  = regexp.MustCompile(`^\[[ \txX]*\]`) // bracket of only blanks/x/X (botched)
)

// LintAcceptanceCriteria reports misconfigurations that would make the acceptance
// tally / `task ac` list lie: a botched checkbox in the (first) acceptance section
// that the scanner silently skips, and more than one acceptance section (only the
// first is used). Empty when the body's acceptance criteria are well-formed. The
// checks are fence-aware, matching the scanner they guard.
func LintAcceptanceCriteria(body string) []Issue {
	lines := strings.Split(normalizeNewlines(body), "\n")
	var (
		issues     []Issue
		fence      fenceScanner
		acSections int
		inFirst    bool
		firstLvl   int
		firstDone  bool
	)
	for _, ln := range lines {
		if fence.inCode(ln) {
			continue
		}
		if m := bodyHeadingRe.FindStringSubmatch(ln); m != nil {
			lvl := len(m[1])
			isAcc := isAcceptanceSectionHeading(m[2])
			if isAcc {
				acSections++
			}
			switch {
			case isAcc && !inFirst && !firstDone:
				inFirst, firstLvl = true, lvl
			case inFirst && lvl <= firstLvl:
				inFirst, firstDone = false, true
			}
			continue
		}
		if inFirst {
			if bad, ok := malformedCheckbox(ln); ok {
				issues = append(issues, Issue{Field: "acceptance", Message: fmt.Sprintf("malformed acceptance checkbox %q — use `- [ ]` or `- [x]` (it is not counted as written)", bad)})
			}
		}
	}
	if acSections > 1 {
		issues = append(issues, Issue{Field: "acceptance", Message: fmt.Sprintf("%d acceptance-criteria sections — the tally and `task ac` use the first and ignore the rest; merge them", acSections)})
	}
	return issues
}

// isAcceptanceSectionHeading is the PRECISE test the lint guard uses to identify an
// acceptance-criteria section — the canonical "Acceptance criteria" name, not merely
// any heading that mentions "acceptance". This is deliberately stricter than the
// tally scanner's substring match: without it, a "## Progress — notes on acceptance
// criteria" heading would be miscounted as a second acceptance section (a false
// positive the guard itself must not raise).
func isAcceptanceSectionHeading(title string) bool {
	t := strings.ToLower(strings.TrimSpace(title))
	return t == "acceptance" || strings.HasPrefix(t, "acceptance criteria") || strings.HasPrefix(t, "acceptance:")
}

// malformedCheckbox reports whether line is a list item whose leading bracket is a
// botched checkbox (blanks/x/X only, but not the canonical form), returning that
// `[…]` token for the message.
func malformedCheckbox(line string) (string, bool) {
	m := acListItemRe.FindStringSubmatch(line)
	if m == nil || acCheckboxOKRe.MatchString(m[1]) {
		return "", false
	}
	if tok := acCheckboxyRe.FindString(m[1]); tok != "" {
		return tok, true
	}
	return "", false
}

func checkWord(checked bool) string {
	if checked {
		return "check"
	}
	return "uncheck"
}

// checkboxText is the criterion text: everything after the `- [x]` marker on the
// checkbox line, trimmed (continuation lines aren't separate criteria).
func checkboxText(line string) string {
	if loc := bodyCheckboxRe.FindStringIndex(line); loc != nil {
		return strings.TrimSpace(line[loc[1]:])
	}
	return strings.TrimSpace(line)
}

// flipCheckbox rewrites just the single character inside a checkbox line's brackets
// to "x" (checked) or " " (unchecked), leaving indentation, marker, and text intact.
func flipCheckbox(line string, checked bool) string {
	loc := bodyCheckboxRe.FindStringSubmatchIndex(line)
	if loc == nil {
		return line
	}
	mark := " "
	if checked {
		mark = "x"
	}
	// loc[2]:loc[3] is capture group 1 — the single char between the brackets.
	return line[:loc[2]] + mark + line[loc[3]:]
}

// normalizeNewlines folds CRLF (and lone CR) to LF so the line-oriented scanners
// don't miss `\r`-terminated headings — files touched on Windows or under a
// core.autocrlf checkout, or a CRLF body piped through --body-file.
func normalizeNewlines(s string) string {
	if !strings.ContainsRune(s, '\r') {
		return s
	}
	return strings.ReplaceAll(strings.ReplaceAll(s, "\r\n", "\n"), "\r", "\n")
}

func trimTrailingBlankLines(lines []string) string {
	end := len(lines)
	for end > 0 && strings.TrimSpace(lines[end-1]) == "" {
		end--
	}
	return strings.Join(lines[:end], "\n")
}
