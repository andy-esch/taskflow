package domain

import (
	"fmt"
	"regexp"
	"strings"
)

// CandidateTasksVersion marks the first tool-owned candidate-task grammar.
// Sections without this marker are legacy prose: readable and intentionally
// ignored by candidate lint and synchronization rather than guessed at.
const CandidateTasksVersion = "candidate-tasks:v1"

var (
	candidateHeadingRe       = regexp.MustCompile(`(?m)^##[ \t]+Candidate tasks[ \t]*\r?$`)
	sectionHeadingRe         = regexp.MustCompile(`(?m)^#{1,6}[ \t]+`)
	candidateRowRe           = regexp.MustCompile(`^- ([^ \t]+)[ \t]+([A-Z]+\d+)[ \t]+·[ \t]+([a-z-]+)[ \t]+—[ \t]+(.+?)[ \t]*$`)
	candidateVersionMarkerRe = regexp.MustCompile(`^<!--[ \t]+(candidate-tasks:[A-Za-z0-9._:-]+)(?:[ \t·]|-->)`)
)

// CandidateTasksMarkerComment is the exact managed-section marker emitted by
// audit templates. The legend is derived from FindingStatusSpecs so a new status
// cannot silently leave the persisted convention behind.
func CandidateTasksMarkerComment() string {
	parts := make([]string, 0, len(findingStatusSpecs))
	for _, spec := range findingStatusSpecs {
		parts = append(parts, spec.Glyph+" "+spec.Status)
	}
	return "<!-- " + CandidateTasksVersion + " · " + strings.Join(parts, " · ") + " -->"
}

func candidateTasksScaffold() string {
	return "## Candidate tasks\n\n" + CandidateTasksMarkerComment() + "\n" +
		"<!-- Add or replace one row with `tskflwctl audit finding <audit> <code> --candidate \"<one line>\"`; an empty value removes it. -->\n"
}

// InsertBeforeTrailingCandidateTasks keeps the managed projection as the final
// audit section when `audit append` adds narrative prose. The default scaffold
// deliberately ends with Candidate tasks; blindly appending after it makes the new
// prose part of that managed grammar and leaves ordinary lint with an error no
// candidate mutation can repair. The caller supplies LF-normalized text. False
// means no trailing Candidate tasks section exists and ordinary append semantics
// should be used.
func InsertBeforeTrailingCandidateTasks(body, addition string) (string, bool) {
	sections := candidateSections(body)
	if len(sections) == 0 {
		return body, false
	}
	section := sections[len(sections)-1]
	if strings.TrimSpace(body[section.End:]) != "" {
		return body, false
	}
	return insertMarkdownBlock(body, section.Start, addition), true
}

// candidateTask is one canonical row in a managed Candidate tasks section.
// Status is stored beside Glyph deliberately: statuses that share a glyph
// (deferred/superseded) remain distinguishable in plain Markdown.
type candidateTask struct {
	Glyph  string
	Code   string
	Status string
	Text   string
	Span   Span
	Line   int
}

type candidateSection struct {
	Start        int
	ContentStart int
	End          int
	Managed      bool
	MarkerCount  int
	MarkerLine   string
	OtherMarkers []string
	Rows         []candidateTask
	Malformed    []candidateMalformedLine
}

type candidateMalformedLine struct {
	Line int
	Text string
}

func candidateSections(body string) []candidateSection {
	prose := blankFences(body)
	headings := candidateHeadingRe.FindAllStringIndex(prose, -1)
	out := make([]candidateSection, 0, len(headings))
	for _, heading := range headings {
		end := len(body)
		if next := sectionHeadingRe.FindStringIndex(prose[heading[1]:]); next != nil {
			end = heading[1] + next[0]
		}
		contentStart := heading[1]
		if contentStart < len(body) && body[contentStart] == '\r' {
			contentStart++
		}
		if contentStart < len(body) && body[contentStart] == '\n' {
			contentStart++
		}
		section := candidateSection{Start: heading[0], ContentStart: contentStart, End: end}
		parseCandidateSection(body, prose, &section)
		out = append(out, section)
	}
	return out
}

func parseCandidateSection(body, prose string, section *candidateSection) {
	lineNo := 1 + strings.Count(body[:section.ContentStart], "\n")
	for pos := section.ContentStart; pos < section.End; {
		lineEnd := pos
		for lineEnd < section.End && body[lineEnd] != '\n' {
			lineEnd++
		}
		line := strings.TrimSuffix(body[pos:lineEnd], "\r")
		visibleLine := strings.TrimSuffix(prose[pos:lineEnd], "\r")
		trimmed := strings.TrimSpace(visibleLine)
		switch {
		case isCandidateTasksV1Marker(trimmed):
			section.Managed = true
			section.MarkerCount++
			section.MarkerLine = trimmed
		case candidateVersionMarker(trimmed) != "":
			section.OtherMarkers = append(section.OtherMarkers, candidateVersionMarker(trimmed))
		case trimmed == "", strings.HasPrefix(trimmed, "<!--") && strings.HasSuffix(trimmed, "-->"):
			// Blank space and explanatory comments are allowed around managed rows.
		case strings.HasPrefix(trimmed, "- "):
			match := candidateRowRe.FindStringSubmatch(trimmed)
			if match == nil {
				section.Malformed = append(section.Malformed, candidateMalformedLine{Line: lineNo, Text: trimmed})
				break
			}
			section.Rows = append(section.Rows, candidateTask{
				Glyph: match[1], Code: match[2], Status: match[3], Text: match[4],
				Span: Span{Start: pos, End: lineEnd}, Line: lineNo,
			})
		default:
			section.Malformed = append(section.Malformed, candidateMalformedLine{Line: lineNo, Text: strings.TrimSpace(line)})
		}
		pos = lineEnd
		if pos < section.End && body[pos] == '\n' {
			pos++
			lineNo++
		}
	}
}

func isCandidateTasksV1Marker(line string) bool {
	prefix := "<!-- " + CandidateTasksVersion
	if !strings.HasPrefix(line, prefix) {
		return false
	}
	rest := line[len(prefix):]
	return strings.HasPrefix(rest, " ") || strings.HasPrefix(rest, "\t") || strings.HasPrefix(rest, "-->")
}

func candidateVersionMarker(line string) string {
	match := candidateVersionMarkerRe.FindStringSubmatch(line)
	if len(match) != 2 || match[1] == CandidateTasksVersion {
		return ""
	}
	return match[1]
}

// LintCandidateTasks validates only explicitly managed v1 sections. Legacy
// Candidate tasks prose has no reliable finding linkage and is intentionally
// tolerated rather than guessed at or bulk-migrated.
func LintCandidateTasks(body string, findings []Finding) []Issue {
	sections := candidateSections(body)
	managed := make([]candidateSection, 0, len(sections))
	var issues []Issue
	for _, section := range sections {
		if section.Managed {
			managed = append(managed, section)
			continue
		}
		if len(section.OtherMarkers) > 0 {
			for _, marker := range section.OtherMarkers {
				issues = append(issues, Issue{Field: "candidate_tasks", Message: fmt.Sprintf(
					"unsupported Candidate tasks marker %q — expected %q", marker, CandidateTasksVersion)})
			}
			continue
		}
		if len(section.Rows) > 0 {
			issues = append(issues, Issue{Field: "candidate_tasks", Message: fmt.Sprintf(
				"unmanaged Candidate tasks section contains %d canonical row(s) — restore the `%s` marker or rewrite the rows as legacy prose", len(section.Rows), CandidateTasksVersion)})
		}
	}
	if len(managed) == 0 {
		return issues
	}
	if len(managed) > 1 {
		issues = append(issues, Issue{Field: "candidate_tasks", Message: fmt.Sprintf(
			"%d managed Candidate tasks sections — expected exactly one `%s` section", len(managed), CandidateTasksVersion)})
		return issues
	}
	section := managed[0]
	if section.MarkerCount != 1 {
		issues = append(issues, Issue{Field: "candidate_tasks", Message: fmt.Sprintf(
			"managed Candidate tasks section has %d version markers — expected exactly one", section.MarkerCount)})
	}
	if section.MarkerLine != CandidateTasksMarkerComment() {
		issues = append(issues, Issue{Field: "candidate_tasks", Message: fmt.Sprintf(
			"managed marker legend drifted — expected %q", CandidateTasksMarkerComment())})
	}
	for _, line := range section.Malformed {
		issues = append(issues, Issue{Field: "candidate_tasks", Message: fmt.Sprintf(
			"line %d is not a canonical candidate row — expected `- <glyph> <CODE> · <status> — <text>`: %q", line.Line, line.Text)})
	}

	byCode := make(map[string]Finding, len(findings))
	for _, finding := range findings {
		byCode[strings.ToUpper(finding.Code)] = finding
	}
	seen := make(map[string]int)
	for _, row := range section.Rows {
		code := strings.ToUpper(row.Code)
		seen[code]++
		finding, ok := byCode[code]
		if !ok {
			issues = append(issues, Issue{Field: row.Code, Message: fmt.Sprintf(
				"candidate row on line %d references no parsed finding", row.Line)})
			continue
		}
		if !ValidFindingStatus(row.Status) {
			issues = append(issues, Issue{Field: row.Code, Message: fmt.Sprintf(
				"candidate row has unknown status %q", row.Status)})
			continue
		}
		if want := FindingStatusGlyph(row.Status); row.Glyph != want {
			issues = append(issues, Issue{Field: row.Code, Message: fmt.Sprintf(
				"candidate glyph %q disagrees with status %s — expected %q", row.Glyph, row.Status, want)})
		}
		if !strings.EqualFold(row.Status, finding.Status) {
			issues = append(issues, Issue{Field: row.Code, Message: fmt.Sprintf(
				"candidate status %s disagrees with finding status %s", row.Status, finding.Status)})
		}
	}
	for code, count := range seen {
		if count > 1 {
			issues = append(issues, Issue{Field: code, Message: fmt.Sprintf(
				"%d candidate rows reference this finding — exactly one is allowed", count)})
		}
	}
	return issues
}

func managedCandidateSection(body string) (candidateSection, error) {
	sections := candidateSections(body)
	var managed []candidateSection
	for _, section := range sections {
		if section.Managed {
			managed = append(managed, section)
		}
	}
	switch len(managed) {
	case 1:
		return managed[0], nil
	case 0:
		if len(sections) == 0 {
			return candidateSection{}, fmt.Errorf("%w: this audit has no Candidate tasks section; add candidates only to a `%s` audit", ErrValidation, CandidateTasksVersion)
		}
		return candidateSection{}, fmt.Errorf("%w: this audit uses a legacy Candidate tasks section; add candidates only to a `%s` audit rather than guessing at legacy prose", ErrValidation, CandidateTasksVersion)
	default:
		return candidateSection{}, fmt.Errorf("%w: this audit has %d managed Candidate tasks sections; run `audit lint` and repair the duplicate before writing", ErrValidation, len(managed))
	}
}

func canonicalCandidateRow(finding Finding, text string) (string, error) {
	if !ValidFindingStatus(finding.Status) {
		return "", fmt.Errorf("%w: finding %s has unknown status %q; repair the finding status before writing its candidate", ErrValidation, finding.Code, finding.Status)
	}
	return fmt.Sprintf("- %s %s · %s — %s", FindingStatusGlyph(finding.Status), finding.Code,
		strings.ToLower(finding.Status), text), nil
}

// SetFindingCandidate adds, replaces, or removes one managed candidate row.
// Empty text removes the row. It never interprets an unversioned legacy section.
func SetFindingCandidate(body, code, text string) (string, error) {
	if strings.ContainsAny(text, "\r\n") {
		return "", fmt.Errorf("%w: a candidate task is one line; a newline would escape the managed row", ErrValidation)
	}
	var finding *Finding
	for _, candidate := range ParseFindings(body) {
		if strings.EqualFold(candidate.Code, code) {
			copy := candidate
			finding = &copy
			break
		}
	}
	if finding == nil {
		return "", fmt.Errorf("%w: no finding %q in this audit", ErrNotFound, code)
	}
	section, err := managedCandidateSection(body)
	if err != nil {
		return "", err
	}
	if section.MarkerCount != 1 || section.MarkerLine != CandidateTasksMarkerComment() || len(section.Malformed) > 0 {
		return "", fmt.Errorf("%w: the managed Candidate tasks section is malformed; run `audit lint` and repair it before writing a candidate", ErrValidation)
	}
	var matches []candidateTask
	for _, row := range section.Rows {
		if strings.EqualFold(row.Code, finding.Code) {
			matches = append(matches, row)
		}
	}
	if len(matches) > 1 {
		return "", fmt.Errorf("%w: finding %s has %d candidate rows; run `audit lint` and remove the duplicate before writing", ErrValidation, finding.Code, len(matches))
	}
	text = strings.TrimSpace(text)
	if text == "" {
		if len(matches) == 0 {
			return body, nil
		}
		return removeCandidateLine(body, section, matches[0].Span), nil
	}
	row, err := canonicalCandidateRow(*finding, text)
	if err != nil {
		return "", err
	}
	if len(matches) == 1 {
		return body[:matches[0].Span.Start] + row + body[matches[0].Span.End:], nil
	}
	return insertCandidateLine(body, section, row), nil
}

func removeCandidateLine(body string, section candidateSection, span Span) string {
	if len(section.Rows) == 1 {
		start, end := span.Start, span.End
		for start > section.ContentStart && strings.ContainsRune(" \t\r\n", rune(body[start-1])) {
			start--
		}
		for end < section.End && strings.ContainsRune(" \t\r\n", rune(body[end])) {
			end++
		}
		separator := "\n"
		if end < len(body) {
			separator = "\n\n"
		}
		return body[:start] + separator + body[end:]
	}
	start, end := span.Start, span.End
	if end < len(body) && body[end] == '\n' {
		end++
	} else if start > 0 && body[start-1] == '\n' {
		start--
	}
	return body[:start] + body[end:]
}

func insertCandidateLine(body string, section candidateSection, row string) string {
	insertAt := section.End
	for insertAt > 0 && strings.ContainsRune(" \t\r\n", rune(body[insertAt-1])) {
		insertAt--
	}
	before := "\n\n"
	if len(section.Rows) > 0 {
		before = "\n"
	}
	after := "\n"
	if section.End < len(body) {
		after = "\n\n"
	}
	return body[:insertAt] + before + row + after + body[section.End:]
}

// syncManagedCandidateStatus refreshes every canonical row for code after the
// finding status changes. Duplicate rows remain a lint error, but none is left
// carrying stale status; malformed rows stay untouched and loudly linted.
func syncManagedCandidateStatus(body, code string) string {
	var finding *Finding
	for _, candidate := range ParseFindings(body) {
		if strings.EqualFold(candidate.Code, code) {
			copy := candidate
			finding = &copy
			break
		}
	}
	if finding == nil || !ValidFindingStatus(finding.Status) {
		return body
	}
	sections := candidateSections(body)
	var replacements []struct {
		span Span
		text string
	}
	for _, section := range sections {
		if !section.Managed {
			continue
		}
		for _, row := range section.Rows {
			if !strings.EqualFold(row.Code, code) {
				continue
			}
			canonical, err := canonicalCandidateRow(*finding, row.Text)
			if err == nil {
				replacements = append(replacements, struct {
					span Span
					text string
				}{row.Span, canonical})
			}
		}
	}
	for i := len(replacements) - 1; i >= 0; i-- {
		r := replacements[i]
		body = body[:r.span.Start] + r.text + body[r.span.End:]
	}
	return body
}
