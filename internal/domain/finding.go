package domain

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
)

// Finding is one parsed audit finding. The grammar is fixed by the `audit new`
// scaffold (domain's auditBodyTemplate) and audits/HOWTO-execute.md: a `#### CODE.`
// sub-header carrying a title, a `**Status:**`, and optional `**File:**` /
// `**Component:**` / `**Effort:**` / `**Urgency:**` metadata. Fields absent in
// the prose are "". ParseFindings is the SINGLE definition of the grammar — the
// finding counts and (future) per-finding queries both derive from it, so they
// can't drift from each other or from what the tool writes.
type Finding struct {
	Code      string `json:"code"` // H1, M2, S3 …
	Title     string `json:"title"`
	Status    string `json:"status"` // open | in-progress | fixed | landed | deferred | superseded | wontfix (see FindingStatuses)
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
	// statusRe captures the status TOKEN after `**Status:**`, but ONLY where the
	// marker is authoritative — at line start (a status line) or right after the
	// header's `· ` separator — so a literal `**Status:**` mentioned in a title or
	// prose can't be mistaken for the status. The token is the first run with no
	// whitespace/·/|, so "fixed 2026-01-01 (PR #9)" yields "fixed" and "open-ish"
	// stays distinct from "open". `*` is excluded too, so an EMPTY status before a
	// following bold label (`**Status:** **Effort:** S`) parses as "" (then lint
	// flags the missing status) instead of grabbing "**Effort:**" as garbage.
	statusRe    = regexp.MustCompile(`(?mi)(?:^[ \t]*|·[ \t]*)\*\*Status:\*\*[ \t]*([^\s·|*]+)`)
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

// findingStatuses is the legal finding-status vocabulary (the audit HOWTO + the
// `audit new` scaffold). A free-text Status edit can write a typo; `audit lint`
// catches it against this set.
var findingStatuses = map[string]bool{
	"open": true, "in-progress": true, "fixed": true, "landed": true,
	"deferred": true, "superseded": true, "wontfix": true,
}

// FindingStatuses returns the legal finding statuses, sorted (for help/schema).
func FindingStatuses() []string {
	out := make([]string, 0, len(findingStatuses))
	for s := range findingStatuses {
		out = append(out, s)
	}
	sort.Strings(out)
	return out
}

// ValidFindingStatus reports whether s is a legal finding status (case-insensitive).
func ValidFindingStatus(s string) bool {
	return findingStatuses[strings.ToLower(strings.TrimSpace(s))]
}

// LintFindings validates an audit's parsed findings plus the bucket↔state
// invariant, returning one Issue per problem (Field = the finding code, or
// "bucket" for the audit-level check). It checks what the in-repo grammar makes
// knowable: every finding carries a legal **Status:**, and a non-open audit has no
// still-open findings. (The closeout-block nuance + candidate-list drift live with
// the finding-write/sync surface, which parses the candidate list.)
func LintFindings(bucket string, fs []Finding) []Issue {
	var issues []Issue
	for _, f := range fs {
		switch {
		case f.Status == "":
			issues = append(issues, Issue{Field: f.Code, Message: "missing **Status:**"})
		case !ValidFindingStatus(f.Status):
			issues = append(issues, Issue{Field: f.Code, Message: fmt.Sprintf("unknown status %q", f.Status)})
		}
	}
	if bucket != "" && bucket != string(AuditOpen) {
		if open := CountOpenFindings(fs); open > 0 {
			issues = append(issues, Issue{Field: "bucket", Message: fmt.Sprintf("%s audit still has %d open finding(s)", bucket, open)})
		}
	}
	return issues
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

// FindingTally is the per-disposition finding breakdown the segmented progress
// bar bands by. Open + Active + Done + Dropped ≤ len(findings): a finding with an
// unrecognized or missing status (audit lint flags those) counts toward none, so
// the bar's empty track absorbs it — still, correctly, "not done".
type FindingTally struct {
	Open    int // open
	Active  int // in-progress
	Done    int // fixed, landed
	Dropped int // deferred, superseded, wontfix
}

// TallyFindings groups findings by disposition for the bar. The mapping is the
// single source of "what each status means for progress": fixed/landed are done,
// in-progress is active, deferred/superseded/wontfix are dropped (decided or
// parked, not fixed), open is outstanding.
func TallyFindings(fs []Finding) FindingTally {
	var t FindingTally
	for _, f := range fs {
		switch strings.ToLower(strings.TrimSpace(f.Status)) {
		case "open":
			t.Open++
		case "in-progress":
			t.Active++
		case "fixed", "landed":
			t.Done++
		case "deferred", "superseded", "wontfix":
			t.Dropped++
		}
	}
	return t
}

func field(re *regexp.Regexp, section string) string {
	if m := re.FindStringSubmatch(section); m != nil {
		return strings.TrimSpace(m[1])
	}
	return ""
}

// stripInlineStatus drops the header's trailing "· **Status:** …" (keyed on the
// `· ` separator, so a literal `**Status:**` inside the title survives), leaving
// just the title.
func stripInlineStatus(title string) string {
	if i := strings.Index(title, "· **Status:**"); i >= 0 {
		title = title[:i]
	}
	return strings.TrimRight(strings.TrimSpace(title), " ·\t")
}
