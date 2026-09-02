package domain

import (
	"fmt"
	"regexp"
	"strings"
	"unicode/utf8"
)

// ACCount is the acceptance-criteria checkbox tally of a task body: how many of
// the criteria are checked out of the total. A body with no acceptance-criteria
// section (or one with no checkboxes) has a zero tally.
type ACCount struct {
	Checked int
	Total   int
	// Explained is how many of the UNMET criteria say why they are unmet — deferred,
	// wontfix, or n/a. Without it a tally of "1/4" cannot distinguish three criteria still
	// to do from one to do and two deliberately not happening, which is the difference
	// between a task that is stalled and one that is finished with scope removed.
	Explained int
}

// Criterion is one acceptance-criteria checkbox for `task ac --list`: its 1-based
// position, whether it's checked, and the first-line text after the checkbox.
type Criterion struct {
	Index   int
	Checked bool
	Text    string
	// State is the criterion's disposition. The bracket supplies met/not-met; a
	// `· **deferred:** why` suffix refines the not-met case. Every criterion written before
	// this vocabulary existed parses exactly as it always did, which is why there is
	// nothing to migrate.
	State CriterionState
	// Reason is the explanation carried by a non-binary state, required for those and
	// empty otherwise.
	Reason string
	// Suffix is the state as the author WROTE it, empty when none was written. State is the
	// resolved disposition; keeping both lets lint quote what was typed when the two
	// disagree — a checked criterion that also claims to be deferred resolves to met, and a
	// message naming "met" would tell the author nothing about their own edit.
	Suffix CriterionState
}

// acCheckbox is an acceptance-criteria checkbox located in a body: the 0-based line
// index of its marker, the index of its LAST line, and current state/text.
//
// A criterion is not necessarily one line. The corpus wraps them, and treating only the
// marker line as the criterion truncated every wrapped one — `task ac --list`, `task show`
// and the JSON all showed "…rather than introducing a" and silently dropped the rest —
// while the state writer appended its suffix mid-sentence, leaving the remainder stranded
// on the line below. text is the criterion's full logical text with the wrapping collapsed.
type acCheckbox struct {
	line    int // the marker line
	end     int // the criterion's last line (== line when it does not wrap)
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

// UnterminatedFence reports a code fence left open at the end of body: the 1-based
// line it was opened on and the delimiter that opened it. Everything after such a
// fence is invisible to every structure scanner here — headings, sections,
// acceptance criteria, audit findings — so a body that ends inside one silently
// under-reports its own contents.
//
// It is checked on WRITE (store.writeBody) rather than only on read, because the
// write is the last moment the author still has the text and can fix it cheaply.
// By the time a reader notices, the evidence is a finding that cannot be stamped or
// a criterion that never appears in the tally.
func UnterminatedFence(body string) (line int, marker string, ok bool) {
	var fence fenceScanner
	openedAt, openedWith := 0, ""
	for i, ln := range strings.Split(normalizeNewlines(body), "\n") {
		wasOpen := fence.open
		fence.inCode(ln)
		if !wasOpen && fence.open {
			openedAt, openedWith = i+1, strings.Repeat(string(fence.marker), fence.length)
		}
	}
	if fence.open {
		return openedAt, openedWith, true
	}
	return 0, "", false
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
func scanAcceptanceCheckboxes(body string) (lines []string, boxes []acCheckbox, sectionEnd int) {
	lines = strings.Split(normalizeNewlines(body), "\n")
	var (
		fence      fenceScanner
		inSection  bool
		sectionLvl int
		open       = -1 // index of the criterion still accepting continuation lines
	)
	// One past the last non-blank line of the section: where a new criterion is appended.
	// -1 until the section is found, so "no section" is distinguishable from "an empty one".
	sectionEnd = -1
	for i, ln := range lines {
		if fence.inCode(ln) {
			open = -1
			continue
		}
		if m := bodyHeadingRe.FindStringSubmatch(ln); m != nil {
			open = -1
			lvl := len(m[1])
			switch {
			case !inSection:
				if strings.Contains(strings.ToLower(m[2]), "acceptance") {
					inSection, sectionLvl, sectionEnd = true, lvl, i+1
				}
			case lvl <= sectionLvl: // a same-or-higher heading ends the section
				return lines, boxes, sectionEnd
			}
			continue
		}
		if inSection {
			if m := bodyCheckboxRe.FindStringSubmatch(ln); m != nil {
				boxes = append(boxes, acCheckbox{line: i, end: i, checked: m[1] == "x" || m[1] == "X", text: checkboxText(ln)})
				open = len(boxes) - 1
				continue
			}
			// A wrapped criterion continues on the following indented, non-list lines. A
			// blank line, a new list item, or a heading ends it — the same rule a markdown
			// reader applies, kept in this one pass so the fence tracker stays in step.
			if open >= 0 && isCriterionContinuation(ln) {
				boxes[open].end = i
				boxes[open].text += " " + strings.TrimSpace(ln)
				continue
			}
			open = -1
		}
		if inSection && strings.TrimSpace(ln) != "" {
			sectionEnd = i + 1
		}
	}
	return lines, boxes, sectionEnd
}

// isCriterionContinuation reports whether ln continues the criterion above it: indented,
// non-blank, and not itself a list item (a nested bullet is a sub-list, not more sentence).
func isCriterionContinuation(ln string) bool {
	if strings.TrimSpace(ln) == "" || !strings.HasPrefix(ln, " ") && !strings.HasPrefix(ln, "\t") {
		return false
	}
	return !acListItemRe.MatchString(ln)
}

// CountAcceptanceCriteria tallies the acceptance-criteria checkboxes. No such
// section — or none with checkboxes — yields a zero tally.
func CountAcceptanceCriteria(body string) ACCount {
	_, boxes, _ := scanAcceptanceCheckboxes(body)
	var c ACCount
	for _, b := range boxes {
		c.Total++
		if b.checked {
			c.Checked++
			continue
		}
		if _, state, _, _ := splitCriterion(b.text, false); state.NeedsReason() {
			c.Explained++
		}
	}
	return c
}

// proseWrapWidth is where markdown prose this tool WRITES is hard-wrapped: criteria, and a
// finding's resolution note. One number, because the corpus is written at one width by hand
// and two arbitrary margins would show up as ragged diffs between adjacent lines.
const proseWrapWidth = 80

// wrapProse renders text as hard-wrapped markdown lines: the first begins with lead, the
// rest with indent, and no line exceeds proseWrapWidth RUNES.
//
// Runes, not bytes: planning prose is full of `—`, `·`, and `→`, and measuring their UTF-8
// length pulls those lines visibly short of the margin. A word longer than the margin gets
// its own line rather than being broken, since it is likely a path, id, or URL. A word
// starting `**` never begins a continuation line — it would read as a markdown label
// starting a new block, which for a resolution note means lint counting it as a second one.
func wrapProse(text, lead, indent string, width int) []string {
	line := lead
	var out []string
	for _, w := range strings.Fields(text) {
		// Measure the line as it stands, trailing space included, so the test is exactly
		// "would appending this word overflow" rather than a running total to keep in step.
		if utf8.RuneCountInString(line)+utf8.RuneCountInString(w) > width &&
			line != lead && line != indent && !strings.HasPrefix(w, "**") {
			out = append(out, strings.TrimRight(line, " "))
			line = indent
		}
		line += w + " "
	}
	return append(out, strings.TrimRight(line, " "))
}

// wrapCriterion renders a criterion as `- [ ] text`, wrapped, with continuations indented
// two spaces — which is what scanAcceptanceCheckboxes reads back as one criterion.
func wrapCriterion(text string) []string {
	return wrapProse(text, "- [ ] ", "  ", proseWrapWidth)
}

// validCriterionText is the shared guard for text a criterion is given. A newline would
// break out of the list item and can manufacture phantom checkboxes — the same hazard a
// reason carries, for the same reason.
func validCriterionText(text string) (string, error) {
	if strings.ContainsAny(text, "\r\n") {
		return "", fmt.Errorf("%w: a criterion is a single item — a newline here would break out of the list and can manufacture phantom checkboxes", ErrValidation)
	}
	t := strings.Join(strings.Fields(text), " ")
	if t == "" {
		return "", fmt.Errorf("%w: a criterion needs text", ErrValidation)
	}
	return t, nil
}

// AddCriterion appends an unchecked criterion to the acceptance-criteria section. It is the
// write half the vocabulary shipped without: evolving a task's criteria meant dumping the
// body, patching the markdown by hand, and writing it back — the raw-surgery gap M4 of
// 2026-07-24-ai-agent-cli-ergonomics names.
//
// The section must already exist. Creating one would mean guessing where it belongs in a
// body the tool did not write, and a heading in the wrong place is harder to notice than a
// refusal.
func AddCriterion(body, text string) (string, error) {
	t, err := validCriterionText(text)
	if err != nil {
		return "", err
	}
	lines, boxes, end := scanAcceptanceCheckboxes(body)
	if end < 0 {
		return "", fmt.Errorf("%w: task has no `## Acceptance criteria` section to add to — add one with `task append`", ErrValidation)
	}
	at := end
	item := wrapCriterion(t)
	if len(boxes) > 0 {
		at = boxes[len(boxes)-1].end + 1 // straight after the last criterion, not the section
	} else if at > 0 && strings.TrimSpace(lines[at-1]) != "" {
		// The section has a heading and no criteria yet, and no blank line to sit under.
		// Every other block this tool writes is separated from its heading; butting the
		// first criterion against it would make the tool's output the odd one out.
		item = append([]string{""}, item...)
	}
	out := append([]string{}, lines[:at]...)
	out = append(out, item...)
	return strings.Join(append(out, lines[at:]...), "\n"), nil
}

// RemoveCriterion deletes the 1-based nth criterion, all of its lines. Indexes after it
// shift down, which is why the CLI prints the renumbered list afterwards.
func RemoveCriterion(body string, n int) (string, error) {
	lines, boxes, _ := scanAcceptanceCheckboxes(body)
	box, err := nthCriterion(boxes, n, "remove")
	if err != nil {
		return "", err
	}
	from, to := box.line, box.end+1
	// Collapse the gap the criterion leaves behind: with a blank line on each side, removing
	// what was between them would leave two, which no other write here produces.
	if from > 0 && strings.TrimSpace(lines[from-1]) == "" &&
		to < len(lines) && strings.TrimSpace(lines[to]) == "" {
		to++
	}
	out := append([]string{}, lines[:from]...)
	return strings.Join(append(out, lines[to:]...), "\n"), nil
}

// ReplaceCriterionText rewrites the 1-based nth criterion's text, KEEPING its checkbox and
// any state suffix. Rewording a criterion is not the same as changing its disposition, and
// silently dropping a `wontfix` and its reason because the wording changed would lose a
// decision — use `task ac --check`/`--defer`/… to change that deliberately.
//
// A suffix the vocabulary does NOT recognise is part of the criterion's text (splitCriterion
// leaves it there rather than guessing, and lint reports it), so it is replaced along with
// the rest — rewording a criterion whose state was typo'd drops the typo.
func ReplaceCriterionText(body string, n int, text string) (string, error) {
	t, err := validCriterionText(text)
	if err != nil {
		return "", err
	}
	lines, boxes, _ := scanAcceptanceCheckboxes(body)
	box, err := nthCriterion(boxes, n, "replace")
	if err != nil {
		return "", err
	}
	_, state, reason, ok := criterionSuffix(box.text)
	if ok && state.NeedsReason() {
		t += fmt.Sprintf(" · **%s:** %s", state, reason)
	}
	rewritten := wrapCriterion(t)
	if box.checked {
		rewritten[0] = strings.Replace(rewritten[0], "- [ ]", "- [x]", 1)
	}
	out := append([]string{}, lines[:box.line]...)
	out = append(out, rewritten...)
	return strings.Join(append(out, lines[box.end+1:]...), "\n"), nil
}

// nthCriterion resolves a 1-based index against the parsed boxes, naming the verb in the
// error so a bad index reads as what the caller was trying to do.
func nthCriterion(boxes []acCheckbox, n int, verb string) (acCheckbox, error) {
	if len(boxes) == 0 {
		return acCheckbox{}, fmt.Errorf("%w: task has no acceptance criteria to %s", ErrValidation, verb)
	}
	if n < 1 || n > len(boxes) {
		return acCheckbox{}, fmt.Errorf("%w: criterion %d out of range (have %d)", ErrValidation, n, len(boxes))
	}
	return boxes[n-1], nil
}

// UnexplainedCriteria returns the criteria that are unmet AND say nothing about why —
// a bare unticked box. They are the ones that block `task complete`, and the distinction
// is the whole point of the state vocabulary: a criterion marked `wontfix` or `deferred`
// has been DECIDED, and a decision should not stand in the way of finishing a task. Only
// silence should.
func UnexplainedCriteria(body string) []Criterion {
	var out []Criterion
	for _, c := range ListAcceptanceCriteria(body) {
		if c.State == CriterionUnmet {
			out = append(out, c)
		}
	}
	return out
}

// CriterionCount is one state's share of a task's acceptance criteria.
type CriterionCount struct {
	State CriterionState
	N     int
}

// TallyCriteria counts a body's acceptance criteria by state, in the vocabulary's own
// order, omitting states with no members. It is the roll-up's source: a task with a
// criterion that is deferred rather than merely unticked has made a DECISION, and a bare
// "3/8" cannot say so — it reads as five things still to do.
func TallyCriteria(body string) []CriterionCount {
	byState := map[CriterionState]int{}
	for _, c := range ListAcceptanceCriteria(body) {
		byState[c.State]++
	}
	out := make([]CriterionCount, 0, len(byState))
	for _, st := range CriterionStates() {
		if n := byState[st]; n > 0 {
			out = append(out, CriterionCount{State: st, N: n})
		}
	}
	return out
}

// ListAcceptanceCriteria returns the acceptance criteria in body order, 1-based —
// the `task ac --list` view an agent then flips by index.
func ListAcceptanceCriteria(body string) []Criterion {
	_, boxes, _ := scanAcceptanceCheckboxes(body)
	out := make([]Criterion, len(boxes))
	for i, b := range boxes {
		text, state, reason, suffix := splitCriterion(b.text, b.checked)
		out[i] = Criterion{Index: i + 1, Checked: b.checked, Text: text, State: state, Reason: reason, Suffix: suffix}
	}
	return out
}

// SetCriterionState sets the 1-based nth criterion to an explicit state, rewriting only
// that line. This is the write path the vocabulary needed to ship WITH: findings gained
// seven statuses and no verb, so every change since has been a hand edit — the habit that
// let the vocabulary drift from its own documentation. A state that is only reachable by
// hand-editing is a state nobody can be held to.
//
// The bracket and the suffix are kept consistent by construction: met checks the box and
// drops any suffix, and every other state unchecks it, so the two halves can never
// disagree the way lint would otherwise have to report.
func SetCriterionState(body string, n int, state CriterionState, reason string) (string, error) {
	return SetCriteriaState(body, []int{n}, state, reason)
}

// SetCriteriaState sets every 1-based index in ns to one state in a SINGLE pass over the
// body. It exists because setting k criteria one call at a time is k file rewrites, k
// updated_at bumps, and k filesystem events for what is semantically one edit — the shell
// loop callers were writing to work around a single-index API.
//
// Every index is validated BEFORE any line is touched, so a bad index anywhere in the
// list rejects the whole request rather than leaving the body half-applied. Indices are
// deduplicated and order does not matter: each criterion is written at most once, and
// 3,1 means the same as 1,3. Returning the input body unchanged means every requested
// criterion was already in the target state, so the caller can skip the write.
//
// One pass is safe because every edit here replaces a line in place — nothing is
// inserted or removed — so the scanned box positions stay valid throughout.
func SetCriteriaState(body string, ns []int, state CriterionState, reason string) (string, error) {
	if strings.ContainsAny(reason, "\r\n") {
		return "", fmt.Errorf("%w: a reason must be a single line — a newline here would break out of the criterion and can manufacture phantom checkboxes", ErrValidation)
	}
	if state.NeedsReason() && strings.TrimSpace(reason) == "" {
		return "", fmt.Errorf("%w: %s needs a reason — say why, so it cannot be mistaken for an oversight",
			ErrValidation, state)
	}
	if !state.NeedsReason() && strings.TrimSpace(reason) != "" {
		return "", fmt.Errorf("%w: %s takes no reason (only %s do)",
			ErrValidation, state, strings.Join(CriterionSuffixStates(), ", "))
	}
	lines, boxes, _ := scanAcceptanceCheckboxes(body)
	if len(boxes) == 0 {
		return "", fmt.Errorf("%w: task has no acceptance criteria to set", ErrValidation)
	}
	if len(ns) == 0 {
		return "", fmt.Errorf("%w: no criterion index given to set %s", ErrValidation, state)
	}
	// Validate every index first: a partially-applied write would leave the file in a
	// state the caller never asked for and cannot distinguish from success.
	seen := make(map[int]bool, len(ns))
	targets := make([]int, 0, len(ns))
	for _, n := range ns {
		if n < 1 || n > len(boxes) {
			return "", fmt.Errorf("%w: criterion %d out of range (have %d)", ErrValidation, n, len(boxes))
		}
		if seen[n] {
			continue // a repeated index writes one criterion once, not twice
		}
		seen[n] = true
		targets = append(targets, n)
	}
	for _, n := range targets {
		box := boxes[n-1]
		// Strip whatever suffix is there before writing the new one, so repeated calls
		// replace rather than stack. It can sit on any line of a wrapped criterion —
		// including the wrong one, if it was written before the writer knew criteria wrap.
		for j := box.line; j <= box.end; j++ {
			lines[j] = stripCriterionSuffixLine(lines[j])
		}
		lines[box.line] = replaceCheckboxLine(lines[box.line], state.Met(), strings.TrimSpace(checkboxText(lines[box.line])))
		// The suffix belongs at the END of the criterion, which on a wrapped one is not
		// the marker line: appending it there splits the sentence and leaves its tail
		// dangling under the reason.
		if state.NeedsReason() {
			lines[box.end] = strings.TrimRight(lines[box.end], " \t") +
				fmt.Sprintf(" · **%s:** %s", state, strings.TrimSpace(reason))
		}
	}
	return strings.Join(lines, "\n"), nil
}

// stripCriterionSuffixLine removes a trailing disposition suffix from one line, leaving a
// line that carries none untouched. Line-level rather than text-level because a wrapped
// criterion's suffix may be on a continuation line, which has no checkbox marker to parse.
func stripCriterionSuffixLine(line string) string {
	stripped, _, _, ok := criterionSuffix(line)
	if !ok {
		return line
	}
	// criterionSuffix trims the whole text, indentation included. Re-attaching the line's
	// own indent is what keeps a continuation line a continuation: strip it and the next
	// scan no longer sees the line as part of the criterion, so the following write leaves
	// its suffix behind and `met` stops clearing it.
	indent := line[:len(line)-len(strings.TrimLeft(line, " \t"))]
	return indent + strings.TrimRight(stripped, " \t")
}

// replaceCheckboxLine rebuilds one checkbox line, preserving its original indentation and
// bullet so a nested or differently-bulleted list survives a state change untouched.
func replaceCheckboxLine(line string, checked bool, text string) string {
	loc := bodyCheckboxRe.FindStringIndex(line)
	if loc == nil {
		return line
	}
	box := "[ ]"
	if checked {
		box = "[x]"
	}
	prefix := line[:loc[0]]
	marker := line[loc[0]:loc[1]]
	// Rewrite only the bracket inside the matched marker, so the bullet character and any
	// spacing the author used are preserved verbatim.
	marker = strings.Replace(marker, "[ ]", box, 1)
	marker = strings.Replace(marker, "[x]", box, 1)
	marker = strings.Replace(marker, "[X]", box, 1)
	// bodyCheckboxRe stops at the closing bracket, so the single separating space that the
	// canonical form uses is re-added here rather than inherited.
	return prefix + marker + " " + text
}

// SetAcceptanceCriterion flips the 1-based nth acceptance-criteria checkbox to
// checked/unchecked, returning the new body. Only that one checkbox's `[ ]`/`[x]`
// is rewritten — every other byte (frontmatter is handled upstream) is preserved.
// It is idempotent: flipping to the current state returns the body unchanged (the
// caller can skip the write). ErrValidation when there's no acceptance section or n
// is out of range.
func SetAcceptanceCriterion(body string, n int, checked bool) (string, error) {
	lines, boxes, _ := scanAcceptanceCheckboxes(body)
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
	return append(issues, lintCriterionStates(body)...)
}

// lintCriterionStates validates the state suffix on each criterion. Every message names the
// legal set, because an error that only says a value is wrong leaves the reader — human or
// agent — to go and find what right looks like (finding M1 of the finding-status audit).
func lintCriterionStates(body string) []Issue {
	var issues []Issue
	for _, c := range ListAcceptanceCriteria(body) {
		switch {
		case c.Checked && c.Suffix != "":
			// The bracket and the suffix disagree. Met is met; a met criterion that also
			// claims to be deferred is a half-finished edit, not a state.
			issues = append(issues, Issue{Field: "acceptance", Message: fmt.Sprintf(
				"criterion %d is checked but carries a **%s:** suffix — a met criterion has no state suffix; uncheck it or drop the suffix",
				c.Index, c.Suffix)})
		case c.State.NeedsReason() && c.Reason == "":
			// A deferral with no why is indistinguishable from an oversight, which is the
			// defect this vocabulary exists to remove.
			issues = append(issues, Issue{Field: "acceptance", Message: fmt.Sprintf(
				"criterion %d is **%s:** with no reason — say why, e.g. `· **%s:** waiting on the schema ADR`",
				c.Index, c.State, c.State)})
		}
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
// criterionMarkerRe matches ONE `· **label:**` marker on a criterion line. It is
// deliberately un-anchored: the disposition suffix is the LAST such marker, and RE2 matches
// leftmost, so a single anchored pattern silently picked the first one. A criterion reading
// `Tests · **Coverage:** 100% · **deferred:** waiting on harness` matched `**Coverage:**`,
// failed to parse it as a state, and dropped the deferral entirely — the tally lost it and
// the criterion read as a plain oversight.
//
// Between the separator and the bold marker it tolerates a decoration run — anything that
// is not a letter, digit, or `*` — so `· ⏳ **deferred:**` matches. This repo's own
// candidate lists use ✅ ⏳ ⛔, and finding M2 of the finding-status audit records what
// happens when an emoji is left to chance: it gets captured AS the value.
var criterionMarkerRe = regexp.MustCompile(`(?i)[ \t]*·[^*\p{L}\p{N}]*\*\*([^:*]+):\*\*[ \t]*`)

// criterionSuffix finds the trailing disposition suffix: the LAST `· **label:**` on the line
// whose label is a state word. Everything before it is the criterion's own text — which may
// legitimately contain other bold labels — and everything after it is the reason.
//
// ok is false when no marker names a state, in which case the line is ordinary text and is
// left whole rather than guessed at.
func criterionSuffix(text string) (body string, state CriterionState, reason string, ok bool) {
	ms := criterionMarkerRe.FindAllStringSubmatchIndex(text, -1)
	for i := len(ms) - 1; i >= 0; i-- {
		m := ms[i]
		st, valid := ParseCriterionState(text[m[2]:m[3]])
		if !valid {
			continue
		}
		return strings.TrimSpace(text[:m[0]]), st, strings.TrimSpace(text[m[1]:]), true
	}
	return text, "", "", false
}

// splitCriterion separates a criterion's text from its state suffix. An unrecognised
// `· **label:**` is left as ordinary text rather than guessed at: lint reports it, and
// silently swallowing it would hide a typo the author needs to see.
func splitCriterion(text string, checked bool) (body string, state CriterionState, reason string, suffix CriterionState) {
	stripped, written, why, ok := criterionSuffix(text)
	if !ok {
		if checked {
			return text, CriterionMet, "", ""
		}
		return text, CriterionUnmet, "", ""
	}
	if checked {
		// The bracket wins for the resolved state — met is met — while Suffix preserves the
		// contradiction for lint to report against what the author actually typed.
		return stripped, CriterionMet, why, written
	}
	return stripped, written, why, written
}

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
