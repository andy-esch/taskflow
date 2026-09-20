package domain

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var (
	findingSectionHeadingRe = regexp.MustCompile(`(?m)^##[ \t]+Findings[ \t]*\r?$`)
	topSectionHeadingRe     = regexp.MustCompile(`(?m)^#{1,2}[ \t]+`)
	reservedFindingFieldRe  = regexp.MustCompile(`(?i)\*\*(?:Status|File|Component|Effort|Urgency|Recommendation|Resolution):\*\*`)
)

var (
	findingBands     = []string{"H", "M", "L"}
	findingEfforts   = []string{"XS", "S", "M", "L"}
	findingUrgencies = []string{"acute", "soon", "eventually"}
)

// FindingDraft is the semantic input for one tool-authored audit finding. The caller
// supplies prose and metadata, never a finding code or structural Markdown.
type FindingDraft struct {
	Band           string
	Title          string
	File           string
	Component      string
	Effort         string
	Urgency        string
	Body           string
	Recommendation string
}

// FindingBands returns the severity bands accepted for finding identities.
func FindingBands() []string { return append([]string(nil), findingBands...) }

// FindingEfforts returns the canonical optional effort vocabulary.
func FindingEfforts() []string { return append([]string(nil), findingEfforts...) }

// FindingUrgencies returns the canonical optional urgency vocabulary.
func FindingUrgencies() []string { return append([]string(nil), findingUrgencies...) }

// CreateFinding validates and renders one canonical finding, allocates its next
// monotonic audit-local code, and inserts it into the real Findings section. It is pure:
// core recomputes it against every fresh body supplied by the durable CAS transform.
func CreateFinding(body string, draft FindingDraft) (string, Finding, error) {
	normalized, err := normalizeFindingDraft(draft)
	if err != nil {
		return "", Finding{}, err
	}
	if nearMisses := NearMissFindingHeaders(body); len(nearMisses) > 0 {
		return "", Finding{}, fmt.Errorf("%w: audit has %d near-miss finding header(s); run `lint --fix` before allocating a new code", ErrValidation, len(nearMisses))
	}
	findings := ParseFindings(body)
	if issues := LintFindings("", findings); len(issues) > 0 {
		return "", Finding{}, fmt.Errorf("%w: existing finding %s is malformed: %s; run `audit lint` and repair it before creating another finding", ErrValidation, issues[0].Field, issues[0].Message)
	}
	if issues := LintCandidateTasks(body, findings); len(issues) > 0 {
		return "", Finding{}, fmt.Errorf("%w: existing Candidate tasks projection is malformed: %s; inspect it with `audit lint`, then repair the audit with `audit edit` before creating another finding", ErrValidation, issues[0].Message)
	}
	code, err := nextFindingCode(findings, normalized.Band)
	if err != nil {
		return "", Finding{}, err
	}
	block := renderFindingBlock(code, normalized)
	next, err := insertFindingBlock(body, block)
	if err != nil {
		return "", Finding{}, err
	}
	var created []Finding
	for _, finding := range ParseFindings(next) {
		if strings.EqualFold(finding.Code, code) {
			created = append(created, finding)
		}
	}
	if len(created) != 1 {
		return "", Finding{}, fmt.Errorf("%w: created finding %s parsed %d times; no write is safe", ErrValidation, code, len(created))
	}
	return next, created[0], nil
}

func normalizeFindingDraft(draft FindingDraft) (FindingDraft, error) {
	band, err := normalizeFindingChoice("band", draft.Band, findingBands, strings.ToUpper)
	if err != nil {
		return FindingDraft{}, err
	}
	title, err := normalizeFindingLine("title", draft.Title, true, true)
	if err != nil {
		return FindingDraft{}, err
	}
	if reservedFindingFieldRe.MatchString(title) {
		return FindingDraft{}, fmt.Errorf("%w: finding title cannot contain a reserved finding metadata label; pass metadata through its dedicated flag", ErrValidation)
	}
	file, err := normalizeFindingLine("file", draft.File, false, false)
	if err != nil {
		return FindingDraft{}, err
	}
	component, err := normalizeFindingLine("component", draft.Component, false, false)
	if err != nil {
		return FindingDraft{}, err
	}
	effort, err := normalizeFindingChoice("effort", draft.Effort, findingEfforts, strings.ToUpper)
	if err != nil {
		return FindingDraft{}, err
	}
	urgency, err := normalizeFindingChoice("urgency", draft.Urgency, findingUrgencies, strings.ToLower)
	if err != nil {
		return FindingDraft{}, err
	}
	recommendation, err := normalizeFindingLine("recommendation", draft.Recommendation, false, true)
	if err != nil {
		return FindingDraft{}, err
	}
	if reservedFindingFieldRe.MatchString(recommendation) {
		return FindingDraft{}, fmt.Errorf("%w: finding recommendation cannot contain a reserved finding metadata label", ErrValidation)
	}
	details := strings.TrimSpace(normalizeNewlines(draft.Body))
	visibleDetails := blankFences(details)
	if docHeaderRe.MatchString(visibleDetails) {
		return FindingDraft{}, fmt.Errorf("%w: finding body cannot contain an unfenced Markdown heading; use prose or a fenced example so one creation cannot open another section", ErrValidation)
	}
	if reservedFindingFieldRe.MatchString(visibleDetails) {
		return FindingDraft{}, fmt.Errorf("%w: finding body cannot contain a reserved finding metadata label; use the dedicated flag or fence a literal example", ErrValidation)
	}
	return FindingDraft{
		Band: band, Title: title, File: file, Component: component,
		Effort: effort, Urgency: urgency, Body: details, Recommendation: recommendation,
	}, nil
}

func normalizeFindingLine(field, value string, required, allowSeparators bool) (string, error) {
	if strings.ContainsAny(value, "\r\n") {
		return "", fmt.Errorf("%w: finding %s must be one line", ErrValidation, field)
	}
	value = strings.Join(strings.Fields(value), " ")
	if required && value == "" {
		return "", fmt.Errorf("%w: finding %s is required", ErrValidation, field)
	}
	if !allowSeparators && strings.ContainsAny(value, "|·*") {
		return "", fmt.Errorf("%w: finding %s cannot contain `|`, `·`, or `*` because those delimit metadata fields", ErrValidation, field)
	}
	return value, nil
}

func normalizeFindingChoice(field, value string, allowed []string, normalize func(string) string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		if field == "band" {
			return "", fmt.Errorf("%w: finding band is required — expected one of: %s", ErrValidation, strings.Join(allowed, ", "))
		}
		return "", nil
	}
	value = normalize(value)
	for _, candidate := range allowed {
		if value == candidate {
			return value, nil
		}
	}
	return "", fmt.Errorf("%w: unknown finding %s %q — expected one of: %s", ErrValidation, field, value, strings.Join(allowed, ", "))
}

func nextFindingCode(findings []Finding, band string) (string, error) {
	seen := make(map[string]struct{}, len(findings))
	max := 0
	for _, finding := range findings {
		code := strings.ToUpper(finding.Code)
		if _, exists := seen[code]; exists {
			return "", fmt.Errorf("%w: duplicate existing finding code %s makes allocation ambiguous; repair the audit before creating another finding", ErrValidation, code)
		}
		seen[code] = struct{}{}
		if !strings.HasPrefix(code, band) || len(code) == len(band) {
			continue
		}
		suffix := code[len(band):]
		if strings.Trim(suffix, "0123456789") != "" {
			continue
		}
		n, err := strconv.Atoi(suffix)
		if err != nil {
			return "", fmt.Errorf("%w: finding code %s cannot participate in allocation: %v", ErrValidation, code, err)
		}
		if n > max {
			max = n
		}
	}
	return band + strconv.Itoa(max+1), nil
}

func renderFindingBlock(code string, draft FindingDraft) string {
	blocks := []string{fmt.Sprintf("#### %s. %s · **Status:** open", code, draft.Title)}
	var metadata []string
	var location []string
	if draft.File != "" {
		location = append(location, "**File:** "+draft.File)
	}
	if draft.Component != "" {
		location = append(location, "**Component:** "+draft.Component)
	}
	if len(location) > 0 {
		metadata = append(metadata, strings.Join(location, " | "))
	}
	var planning []string
	if draft.Effort != "" {
		planning = append(planning, "**Effort:** "+draft.Effort)
	}
	if draft.Urgency != "" {
		planning = append(planning, "**Urgency:** "+draft.Urgency)
	}
	if len(planning) > 0 {
		metadata = append(metadata, strings.Join(planning, " · "))
	}
	if len(metadata) > 0 {
		blocks = append(blocks, strings.Join(metadata, "\n"))
	}
	if draft.Body != "" {
		blocks = append(blocks, draft.Body)
	}
	if draft.Recommendation != "" {
		blocks = append(blocks, "**Recommendation:** "+draft.Recommendation)
	}
	return strings.Join(blocks, "\n\n")
}

func insertFindingBlock(body, block string) (string, error) {
	prose := blankFences(body)
	headings := findingSectionHeadingRe.FindAllStringIndex(prose, -1)
	if len(headings) > 1 {
		return "", fmt.Errorf("%w: audit has %d real Findings sections; merge them before creating a finding", ErrValidation, len(headings))
	}
	if len(headings) == 1 {
		at := len(body)
		if next := topSectionHeadingRe.FindStringIndex(prose[headings[0][1]:]); next != nil {
			at = headings[0][1] + next[0]
		}
		return insertMarkdownBlock(body, at, block), nil
	}

	at := len(body)
	if candidate := candidateHeadingRe.FindStringIndex(prose); candidate != nil {
		at = candidate[0]
	}
	return insertMarkdownBlock(body, at, "## Findings\n\n"+block), nil
}

func insertMarkdownBlock(body string, at int, block string) string {
	prefix := ""
	if at > 0 {
		switch {
		case strings.HasSuffix(normalizeNewlines(body[:at]), "\n\n"):
		case strings.HasSuffix(body[:at], "\n"):
			prefix = "\n"
		default:
			prefix = "\n\n"
		}
	}
	suffix := "\n"
	if at < len(body) {
		suffix = "\n\n"
	}
	return body[:at] + prefix + block + suffix + body[at:]
}
