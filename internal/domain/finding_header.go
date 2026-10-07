package domain

import "strings"

// FindingHeaderDisposition separates recognition from permission to repair.
type FindingHeaderDisposition string

const (
	FindingHeaderCanonical  FindingHeaderDisposition = "canonical"
	FindingHeaderRepairable FindingHeaderDisposition = "repairable"
	FindingHeaderAmbiguous  FindingHeaderDisposition = "ambiguous"
	FindingHeaderOrdinary   FindingHeaderDisposition = "ordinary"
)

// FindingHeaderClassification locates one code-shaped heading and records its
// evidence class. Canonical is a proposed spelling, not permission to rewrite.
type FindingHeaderClassification struct {
	Line            int
	Text, Canonical string
	Code            string
	Disposition     FindingHeaderDisposition
}

// ClassifyFindingHeaders classifies unfenced, code-shaped headings. A code token
// alone is not finding evidence: S3, V2, MP3 and Top3 occur in ordinary prose.
// Auto-repair requires a Status marker owned by that heading (before the next
// heading) plus a Findings context, a known severity band, or explicit code-token
// drift. A severity/code-shaped candidate inside Findings without Status evidence
// is ambiguous; an undecorated S3/V2/MP3/Top3 prose heading remains ordinary.
// Defined-code narrative references are ordinary; a duplicate definition with
// Status evidence is ambiguous.
//
// The parser is unchanged: one canonical syntax, no alternate persisted forms.
// Numbered/word-number/seven-hash headings are intentionally outside this grammar.
func ClassifyFindingHeaders(body string) []FindingHeaderClassification {
	body = normalizeNewlines(body)
	lines := strings.Split(blankFences(body), "\n")
	defined := make(map[string]bool)
	for _, finding := range ParseFindings(body) {
		i := len(finding.Code)
		for i > 0 && finding.Code[i-1] >= '0' && finding.Code[i-1] <= '9' {
			i--
		}
		defined[canonicalFindingCode(finding.Code[:i], finding.Code[i:])] = true
	}
	var out []FindingHeaderClassification
	findingsDepth := 0
	for i, line := range lines {
		if !docHeaderRe.MatchString(line) {
			continue
		}
		depth := len(line) - len(strings.TrimLeft(line, "#"))
		title := strings.TrimSpace(strings.TrimRight(strings.TrimSpace(line[depth:]), "#"))
		if strings.EqualFold(title, "Findings") {
			findingsDepth = depth
			continue
		}
		insideFindings := findingsDepth > 0 && depth > findingsDepth
		if findingsDepth > 0 && depth <= findingsDepth {
			findingsDepth = 0
		}
		if findingHeaderRe.MatchString(line) {
			out = append(out, FindingHeaderClassification{Line: i + 1, Text: line, Canonical: line, Disposition: FindingHeaderCanonical})
			continue
		}
		m := nearMissHeaderRe.FindStringSubmatch(line)
		if m == nil {
			continue
		}
		end := i + 1
		for end < len(lines) && !docHeaderRe.MatchString(lines[end]) {
			end++
		}
		hasStatus := findingStatusMarkerRe.MatchString(findingStatusSource(strings.Join(lines[i:end], "\n")))
		code := canonicalFindingCode(m[2], m[3])
		token := strings.Fields(line[depth:])[0]
		decoratedCode := token != m[2]+m[3]
		knownBand := anyEqualFoldFindingBand(m[2])
		class := FindingHeaderOrdinary
		switch {
		case hasStatus && !defined[code] && (insideFindings || decoratedCode || knownBand):
			class = FindingHeaderRepairable
		case hasStatus:
			class = FindingHeaderAmbiguous // do not manufacture duplicate codes
		case defined[code]:
			// A closeout reference, not a second definition.
		case insideFindings && (decoratedCode || knownBand):
			class = FindingHeaderAmbiguous
		}
		out = append(out, FindingHeaderClassification{Line: i + 1, Text: line,
			Canonical: canonicalFindingHeader(m[1], m[2], m[3], m[4]), Code: code, Disposition: class})
	}
	counts := make(map[string]int)
	for _, header := range out {
		if header.Disposition == FindingHeaderRepairable {
			counts[header.Code]++
		}
	}
	for i := range out {
		if out[i].Disposition == FindingHeaderRepairable && counts[out[i].Code] > 1 {
			out[i].Disposition = FindingHeaderAmbiguous
		}
	}
	return out
}

func anyEqualFoldFindingBand(letters string) bool {
	for _, band := range findingBands {
		if strings.EqualFold(letters, band) {
			return true
		}
	}
	return false
}
