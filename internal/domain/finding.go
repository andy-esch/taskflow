package domain

import (
	"regexp"
	"strings"
)

// Finding is one parsed audit finding. The grammar is fixed by the `audit new`
// scaffold (core's auditBodyTemplate) and audits/HOWTO-execute.md: a `#### CODE.`
// sub-header carrying a title, a `**Status:**`, and optional `**File:**` /
// `**Component:**` / `**Effort:**` / `**Urgency:**` metadata. Fields absent in
// the prose are "". ParseFindings is the SINGLE definition of the grammar — the
// finding counts and (future) per-finding queries both derive from it, so they
// can't drift from each other or from what the tool writes.
type Finding struct {
	Code      string `json:"code"` // H1, M2, S3 …
	Title     string `json:"title"`
	Status    string `json:"status"` // open | in-progress | fixed | deferred | superseded | wontfix | …
	File      string `json:"file"`
	Component string `json:"component"`
	Effort    string `json:"effort"`  // XS | S | M | L
	Urgency   string `json:"urgency"` // acute | soon | eventually
}

var (
	// findingHeaderRe matches a finding sub-header ("#### H1." / "### M2."),
	// capturing the code and the rest of the line (the title, possibly with an
	// inline "· **Status:** …").
	findingHeaderRe = regexp.MustCompile(`(?m)^#{2,6}\s+([A-Z]+\d+)\.[ \t]*(.*)$`)
	// fenceRe spans a ```-fenced code block, stripped first so example finding
	// syntax in docs or the scaffold isn't parsed as a real finding.
	fenceRe = regexp.MustCompile("(?s)```.*?```")
	// statusRe captures the status TOKEN after `**Status:**` — the first run with
	// no whitespace/·/| — so "fixed 2026-01-01 (PR #9)" yields "fixed" and
	// "open-ish" stays distinct from "open".
	statusRe    = regexp.MustCompile(`(?i)\*\*Status:\*\*\s*([^\s·|]+)`)
	fileRe      = fieldValueRe("File")
	componentRe = fieldValueRe("Component")
	effortRe    = fieldValueRe("Effort")
	urgencyRe   = fieldValueRe("Urgency")
)

// fieldValueRe matches `**Label:** value`, where value runs to the next field
// separator (| or ·), the next **bold**, or end of line.
func fieldValueRe(label string) *regexp.Regexp {
	return regexp.MustCompile(`(?i)\*\*` + label + `:\*\*\s*([^|·*\n]+)`)
}

// ParseFindings parses every finding in an audit body, in document order.
func ParseFindings(body string) []Finding {
	prose := fenceRe.ReplaceAllString(body, "")
	headers := findingHeaderRe.FindAllStringSubmatchIndex(prose, -1)
	out := make([]Finding, 0, len(headers))
	for i, h := range headers {
		end := len(prose)
		if i+1 < len(headers) {
			end = headers[i+1][0] // section runs to the next finding header
		}
		section := prose[h[0]:end]
		out = append(out, Finding{
			Code:      prose[h[2]:h[3]],
			Title:     stripInlineStatus(prose[h[4]:h[5]]),
			Status:    field(statusRe, section),
			File:      field(fileRe, section),
			Component: field(componentRe, section),
			Effort:    field(effortRe, section),
			Urgency:   field(urgencyRe, section),
		})
	}
	return out
}

// CountOpenFindings reports how many findings are open (case-insensitive). The
// "what counts as open" rule lives here, with the rest of the grammar.
func CountOpenFindings(fs []Finding) int {
	n := 0
	for _, f := range fs {
		if strings.EqualFold(f.Status, "open") {
			n++
		}
	}
	return n
}

func field(re *regexp.Regexp, section string) string {
	if m := re.FindStringSubmatch(section); m != nil {
		return strings.TrimSpace(m[1])
	}
	return ""
}

// stripInlineStatus drops a trailing inline "· **Status:** …" the scaffold puts
// on the header line, leaving just the title.
func stripInlineStatus(title string) string {
	if i := strings.Index(title, "**Status:**"); i >= 0 {
		title = title[:i]
	}
	return strings.TrimRight(strings.TrimSpace(title), " ·\t")
}
