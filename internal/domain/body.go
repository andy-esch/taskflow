package domain

import (
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

// CountAcceptanceCriteria tallies the task-list checkboxes inside the body's
// acceptance-criteria section (the first heading whose title contains
// "acceptance"). No such section — or none with checkboxes — yields a zero tally.
// Checkboxes in fenced code blocks are ignored.
func CountAcceptanceCriteria(body string) ACCount {
	sec, ok := Section(body, "acceptance") // Section already normalizes newlines
	if !ok {
		return ACCount{}
	}
	var c ACCount
	var fence fenceScanner
	for _, ln := range strings.Split(sec, "\n") {
		if fence.inCode(ln) {
			continue
		}
		m := bodyCheckboxRe.FindStringSubmatch(ln)
		if m == nil {
			continue
		}
		c.Total++
		if m[1] == "x" || m[1] == "X" {
			c.Checked++
		}
	}
	return c
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
