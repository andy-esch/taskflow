package domain

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
	"unicode/utf8"
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
	Status    string `json:"status"` // open | in-progress | fixed | tracked | deferred | superseded | wontfix (see FindingStatuses)
	File      string `json:"file"`
	Component string `json:"component"`
	Effort    string `json:"effort"`  // XS | S | M | L
	Urgency   string `json:"urgency"` // acute | soon | eventually
	// StatusSpan locates the whole status VALUE in the audit body — the token plus any
	// decoration after it, which the line formats carry (`fixed 2026-01-01 (PR #9)`,
	// `deferred (see ADR-0003)`). Status above is just the token; the span is what a write
	// replaces, so re-stamping a decorated status swaps it rather than appending beside it.
	//
	// Carrying the span rather than re-finding the token on write is what makes a status
	// change surgical: everything else in the file, including prose that happens to contain
	// the same word, is untouched by construction. Zero when there is no status to rewrite.
	StatusSpan Span `json:"-"`
	// StatusDecoration is everything after the status token on the Status line — the date,
	// PR link, `by <task-id>`, or reason the line formats carry. Parsed out because Status
	// is only the vocabulary word: without this the wire drops the date on `fixed
	// 2026-08-24` and the destination on `tracked by <id>` entirely.
	StatusDecoration string `json:"status_decoration,omitempty"`
	// Note is the finding's resolution paragraph — HOW it was resolved, as opposed to the
	// one-word Status saying THAT it was. Empty when the finding carries none.
	Note string `json:"note,omitempty"`
	// NoteSpan covers the note including its label, so re-noting replaces the block rather
	// than nesting a second label inside the first. SectionSpan bounds the finding's own
	// text, which is where a first note gets appended.
	NoteSpan    Span `json:"-"`
	SectionSpan Span `json:"-"`
	// NoteLabels counts `**Resolution:**` labels in the section. Both note-shape lint rules
	// read it: more than one means the extras are silently ignored, and one with an empty
	// Note means the label has no paragraph on its line. Hand edits are how both happen.
	NoteLabels int `json:"-"`
}

// Span is a byte range in the audit body: [Start, End).
type Span struct{ Start, End int }

// Empty reports whether the span locates nothing.
func (s Span) Empty() bool { return s.End <= s.Start }

var (
	// findingHeaderRe matches a finding sub-header ("#### H1." / "### M2."),
	// capturing the code and the rest of the line (the title, possibly with an
	// inline "· **Status:** …").
	findingHeaderRe = regexp.MustCompile(`(?m)^#{2,6}\s+([A-Z]+\d+)\.[ \t]*(.*)$`)
	// fenceRe spans a ```-fenced code block, stripped first so example finding
	// syntax in docs or the scaffold isn't parsed as a real finding.
	fenceRe = regexp.MustCompile("(?s)```.*?```")
	// statusRe captures the status VALUE's leading run after `**Status:**`, but ONLY
	// where the marker is authoritative — at line start (a status line) or right after
	// the header's `· ` separator — so a literal `**Status:**` mentioned in a title or
	// prose can't be mistaken for the status. The token is the first run with no
	// whitespace/·/|, so "fixed 2026-01-01 (PR #9)" yields "fixed" and "open-ish"
	// stays distinct from "open". `*` is excluded too, so an EMPTY status before a
	// following bold label (`**Status:** **Effort:** S`) parses as "" (then lint
	// flags the missing status) instead of grabbing "**Effort:**" as garbage.
	//
	// The leading `(?:…)*` group swallows DECORATION before the token — the ✅/⏳/⛔
	// prefixes consumers write to keep a finding visually in sync with the candidate
	// list below it. That prefix was the single highest-frequency parse failure in the
	// corpus (finding M2 of 2026-08-17-finding-status-surface): it was captured AS the
	// status, so a correctly-worked finding linted red. It is captured here rather than
	// skipped so the span still covers it and a re-stamp REPLACES it — leaving a stale
	// ✅ beside a new `deferred` would restate the same two-lists-disagree bug in the
	// file itself.
	statusRe = regexp.MustCompile(`(?mi)(?:^[ \t]*|·[ \t]*)\*\*Status:\*\*[ \t]*((?:[^\s·|*\p{L}\p{N}]+[ \t]*)*[^\s·|*]+)`)
	// docHeaderRe matches any ATX header line. A finding's section ends at the next finding
	// OR at the next section of the document (`## Candidate tasks`), whichever comes first.
	// Without that second bound the LAST finding's section swallows the rest of the file,
	// and a resolution note appended to it would land under the wrong heading.
	docHeaderRe = regexp.MustCompile(`(?m)^#{1,6}[ \t]`)
	// findingNoteRe locates a resolution note's label at line start. The note's EXTENT is
	// computed from there rather than captured, because a paragraph ends at a blank line —
	// something a character class cannot express, and the reason this is not a `**Field:**`
	// entry in the single-line grammar above (that parser stops at the first `|`/`·`/`**`,
	// which would truncate prose mid-sentence).
	findingNoteRe = regexp.MustCompile(`(?m)^\*\*Resolution:\*\*[ \t]*`)
	fileRe        = fieldValueRe("File")
	componentRe   = fieldValueRe("Component")
	effortRe      = fieldValueRe("Effort")
	urgencyRe     = fieldValueRe("Urgency")
)

// fieldValueRe matches `**Label:** value`, where value runs to the next field
// separator (| or ·), the next **bold**, or end of line.
func fieldValueRe(label string) *regexp.Regexp {
	return regexp.MustCompile(`(?i)\*\*` + label + `:\*\*\s*([^|·*\n]+)`)
}

// blankFences masks fenced code blocks while PRESERVING every byte offset: each character
// inside a fence becomes a space, and newlines are kept so line-anchored patterns still see
// the same line structure. Fenced example syntax therefore cannot match as a real finding,
// and every index into the result is still a valid index into the original body.
//
// Deleting the fences instead — which is what this did — silently desynchronised the spans
// from the file they were computed for. Any fence before a finding shifted every later
// offset by its length, so a status write landed somewhere else entirely: on the scaffold
// `audit new` emits, `## Candidate tasks` became `## Candidafixedasks` while the finding's
// own status went unchanged, and the command still reported success.
func blankFences(body string) string {
	out := []byte(body)
	for _, m := range fenceRe.FindAllStringIndex(body, -1) {
		for i := m[0]; i < m[1]; i++ {
			if out[i] != '\n' {
				out[i] = ' '
			}
		}
	}
	return string(out)
}

// ParseFindings parses every finding in an audit body, in document order.
func ParseFindings(body string) []Finding {
	prose := blankFences(body)
	headers := findingHeaderRe.FindAllStringSubmatchIndex(prose, -1)
	out := make([]Finding, 0, len(headers))
	for i, h := range headers {
		end := len(prose)
		if i+1 < len(headers) {
			end = headers[i+1][0] // section runs to the next finding header
		}
		// …but no further than the next heading of any kind: the last finding must not
		// annex `## Candidate tasks` and everything under it.
		if m := docHeaderRe.FindStringIndex(prose[h[1]:end]); m != nil {
			end = h[1] + m[0]
		}
		section := prose[h[0]:end]
		status, span := fieldSpan(statusRe, section, h[0])
		decoration := ""
		if !span.Empty() {
			decoration = statusDecoration(prose[span.Start:span.End])
		}
		note, noteSpan, labels := findingNote(section, h[0])
		out = append(out, Finding{
			Code:             prose[h[2]:h[3]],
			Title:            stripInlineStatus(prose[h[4]:h[5]]),
			Status:           status,
			StatusSpan:       span,
			StatusDecoration: decoration,
			Note:             note,
			NoteSpan:         noteSpan,
			SectionSpan:      Span{Start: h[0], End: end},
			NoteLabels:       labels,
			File:             field(fileRe, section),
			Component:        field(componentRe, section),
			Effort:           field(effortRe, section),
			Urgency:          field(urgencyRe, section),
		})
	}
	return out
}

// findingStatuses is the legal finding-status vocabulary (the audit HOWTO + the
// `audit new` scaffold). A free-text Status edit can write a typo; `audit lint`
// catches it against this set.
var findingStatuses = map[string]bool{
	"open": true, "in-progress": true, "fixed": true, "tracked": true,
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
			issues = append(issues, Issue{Field: f.Code, Message: fmt.Sprintf(
				"missing **Status:** — expected one of: %s", strings.Join(FindingStatuses(), ", "))})
		case !ValidFindingStatus(f.Status):
			issues = append(issues, Issue{Field: f.Code, Message: fmt.Sprintf(
				"unknown status %q — expected one of: %s", f.Status, strings.Join(FindingStatuses(), ", "))})
		// `audit finding` refuses to WRITE a destination-less `tracked`; a hand edit can
		// still produce one, and a handoff with nowhere to follow is the improvisation the
		// word was introduced to replace.
		case strings.EqualFold(f.Status, ResolutionTracked) && f.StatusDecoration == "":
			issues = append(issues, Issue{Field: f.Code, Message: fmt.Sprintf(
				"`%s` needs a destination — write `%s by <task-id>`", ResolutionTracked, ResolutionTracked)})
		}
		switch {
		case f.NoteLabels > 1:
			issues = append(issues, Issue{Field: f.Code, Message: fmt.Sprintf(
				"more than one `%s` block — only the first is read", FindingNoteLabel)})
		case f.NoteLabels == 1 && f.Note == "":
			issues = append(issues, Issue{Field: f.Code, Message: fmt.Sprintf(
				"empty `%s` label — the paragraph must start on the label's own line", FindingNoteLabel)})
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
	Done    int // fixed, tracked
	Dropped int // deferred, superseded, wontfix
}

// TallyFindings groups findings by disposition for the bar. The mapping is the
// single source of "what each status means for progress": fixed/tracked are done,
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
		case "fixed", ResolutionTracked:
			// `tracked` counts as DONE from the audit's point of view: the finding has been
			// transferred to a task and is no longer the audit's business. That is the
			// distinction it exists to draw — a deferred finding is still owned here, a
			// tracked one is not, and an audit should not stay open waiting on work it
			// handed away.
			t.Done++
		case ResolutionDeferred, "superseded", ResolutionWontFix:
			t.Dropped++
		}
	}
	return t
}

// fieldSpan is field() that also reports where the captured token sits, offset by the
// section's own position in the body. Only the status needs this today; keeping it beside
// field means the two cannot disagree about WHICH token they matched — the drift the
// grammar comment warns about.
func fieldSpan(re *regexp.Regexp, section string, offset int) (string, Span) {
	m := re.FindStringSubmatchIndex(section)
	if m == nil || m[2] < 0 {
		return "", Span{}
	}
	// The span keeps the decoration (so a re-stamp overwrites it); the returned TOKEN
	// drops it, via the same stripper criterion suffixes use — one rule for "ignore the
	// author's glyphs, read the word", spelled once.
	return stripLeadingDecoration(strings.TrimSpace(section[m[2]:m[3]])),
		Span{Start: offset + m[2], End: offset + statusValueEnd(section, m[3])}
}

// findingNote extracts a finding's resolution note: the paragraph introduced by
// `**Resolution:**`. The returned text is the paragraph with its hard wrapping collapsed to
// single spaces, so a reader gets one logical string; the span covers label AND paragraph,
// so re-noting overwrites the block instead of stacking a second label beneath the first.
func findingNote(section string, offset int) (string, Span, int) {
	all := findingNoteRe.FindAllStringIndex(section, -1)
	if len(all) == 0 {
		return "", Span{}, 0
	}
	m := all[0]
	rest := section[m[1]:]
	stop := len(rest)
	if i := strings.Index(rest, "\n\n"); i >= 0 { // a paragraph ends at a blank line
		stop = i
	}
	text := strings.TrimRight(rest[:stop], " \t\n")
	// A label with a BLANK LINE after it terminates at offset 0, so there is no text to
	// read — and reporting a span anyway is how the paragraph below it gets orphaned: the
	// write path would replace the label alone and strand the prose it was introducing.
	// No text, no note, no span; `LintFindings` names the stray label instead.
	if strings.TrimSpace(text) == "" {
		return "", Span{}, len(all)
	}
	return strings.Join(strings.Fields(text), " "),
		Span{Start: offset + m[0], End: offset + m[1] + len(text)}, len(all)
}

// noteWrapWidth is where a resolution note is hard-wrapped. The audit corpus is written at
// this width by hand, and a tool that owns the formatting should produce what a careful
// author would rather than one very long line with an ugly diff.
const noteWrapWidth = 80

// wrapNote renders the label and paragraph as hard-wrapped markdown. Wrapping is safe to do
// here because findingNote reads the paragraph back through strings.Fields, so the wrapped
// form and the logical string round-trip to each other.
func wrapNote(text string) string {
	line := FindingNoteLabel
	width := utf8.RuneCountInString(line)
	var b strings.Builder
	for _, w := range strings.Fields(text) {
		// Runes, not bytes: an audit's prose is full of `—`, `→`, and `✅`, and measuring
		// their UTF-8 length would pull those lines visibly short.
		wl := utf8.RuneCountInString(w)
		// +1 for the space that would join them; an over-long word gets its own line
		// rather than being broken, since it is likely a path, id, or URL.
		// …and never BREAK before a `**` token: a continuation line starting with one reads
		// as a second label, and lint would count it as a duplicate of this very note.
		if width+1+wl > noteWrapWidth && line != FindingNoteLabel && !strings.HasPrefix(w, "**") {
			b.WriteString(line + "\n")
			line, width = w, wl
			continue
		}
		line += " " + w
		width += 1 + wl
	}
	b.WriteString(line)
	return b.String()
}

// statusDecoration splits the trailing decoration off a raw status value: everything after
// the vocabulary word, with the word's own leading glyphs already stripped by the caller.
func statusDecoration(raw string) string {
	rest := stripLeadingDecoration(strings.TrimSpace(raw))
	if i := strings.IndexAny(rest, " \t"); i >= 0 {
		return strings.TrimSpace(rest[i:])
	}
	return ""
}

// isBlank reports whether b is whitespace the block layout can absorb — used to find where
// a note's surrounding blank lines begin and end, so inserting or removing one leaves
// exactly one blank line between blocks rather than a growing gap.
func isBlank(b byte) bool { return b == ' ' || b == '\t' || b == '\n' || b == '\r' }

// statusValueEnd extends a matched status token to the end of its VALUE: the decoration the
// line formats allow (`fixed 2026-01-01 (PR #9)`) belongs to the status, while a following
// `·`/`|`-separated field or bold label belongs to the next field. Without this a re-stamp
// would leave the old decoration stranded after the new value.
func statusValueEnd(section string, tokenEnd int) int {
	rest := section[tokenEnd:]
	end := len(rest)
	if i := strings.IndexByte(rest, '\n'); i >= 0 {
		end = i
	}
	for _, sep := range []string{" · ", " | ", "**"} {
		if i := strings.Index(rest[:end], sep); i >= 0 {
			end = i
		}
	}
	return tokenEnd + len(strings.TrimRight(rest[:end], " \t"))
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

// SetFindingStatus rewrites one finding's status token in an audit body, returning the new
// body. It replaces exactly the span ParseFindings located, so the rest of the file —
// including prose elsewhere that happens to contain the same word, and the finding's own
// title — is byte-identical by construction rather than by careful regex.
//
// This exists because there was no validated write path at all: every status change went
// through a hand edit or a scripted search-and-replace, which is precisely how a vocabulary
// drifts. The value is validated here, so an illegal status cannot be written by any caller.
func SetFindingStatus(body, code, status string) (string, error) {
	// The value may carry decoration the line formats define (`fixed 2026-01-01 (PR #9)`,
	// `superseded by <link>`). Only the leading TOKEN is normalised and vocabulary-checked;
	// the decoration is written verbatim, because it carries dates, links, and document
	// names whose case is meaningful and not ours to flatten.
	if strings.ContainsAny(status, "\r\n") {
		return "", fmt.Errorf("%w: a status must be a single line — a newline here would break the finding header", ErrValidation)
	}
	want := strings.TrimSpace(status)
	token, decoration := want, ""
	if i := strings.IndexAny(want, " \t"); i >= 0 {
		token, decoration = want[:i], strings.TrimSpace(want[i:])
	}
	token = strings.ToLower(token)
	want = token
	if decoration != "" {
		want += " " + decoration
	}
	// `tracked` without a destination is the improvisation this word replaces: it would say
	// the finding left the audit while leaving no way to find where it went. Required here
	// for the same reason a criterion's deferral requires a reason.
	if token == ResolutionTracked && decoration == "" {
		return "", fmt.Errorf("%w: `tracked` needs a destination — write `tracked by <task-id>` so the handoff can be followed", ErrValidation)
	}
	if !ValidFindingStatus(token) {
		return "", fmt.Errorf("%w: unknown finding status %q — expected one of: %s",
			ErrValidation, token, strings.Join(FindingStatuses(), ", "))
	}
	for _, f := range ParseFindings(body) {
		if !strings.EqualFold(f.Code, code) {
			continue
		}
		if f.StatusSpan.Empty() {
			return "", fmt.Errorf("%w: finding %s has no **Status:** line to rewrite — add one, or run `lint --fix`",
				ErrValidation, f.Code)
		}
		return body[:f.StatusSpan.Start] + want + body[f.StatusSpan.End:], nil
	}
	return "", fmt.Errorf("%w: no finding %q in this audit", ErrNotFound, code)
}

// FindingNoteLabel introduces a finding's resolution paragraph.
const FindingNoteLabel = "**Resolution:**"

// SetFindingNote writes a finding's resolution note — the paragraph saying HOW it was
// resolved, beside the one word saying THAT it was. It exists because the alternative is a
// hand-typed block, and hand-typed blocks in this file are what the 2026-08-17 audit was
// about: the label lands in the right place, inside the right finding, or the write fails.
//
// An empty note REMOVES the block. Re-noting replaces it, matching how re-stamping replaces
// a status rather than appending beside it.
func SetFindingNote(body, code, note string) (string, error) {
	// One paragraph, no newlines. A note carrying "\n#### H9." or a fence would open a
	// heading or a code block mid-finding and silently restructure the document — the same
	// class of corruption as the offset desync in H1 of the 2026-08-24 audit, arriving
	// through content instead of arithmetic.
	if strings.ContainsAny(note, "\r\n") {
		return "", fmt.Errorf("%w: a resolution note is a single paragraph — a newline here could open a heading or a fence inside the finding (use `audit edit` for more)", ErrValidation)
	}
	want := strings.Join(strings.Fields(note), " ")
	for _, f := range ParseFindings(body) {
		if !strings.EqualFold(f.Code, code) {
			continue
		}
		if !f.NoteSpan.Empty() {
			if want == "" { // remove the block, and the blank line that set it apart
				cut := f.NoteSpan.Start
				for cut > 0 && isBlank(body[cut-1]) {
					cut--
				}
				tail := f.NoteSpan.End
				for tail < len(body) && isBlank(body[tail]) {
					tail++
				}
				sep := "\n"
				if tail < len(body) {
					sep = "\n\n"
				}
				return body[:cut] + sep + body[tail:], nil
			}
			return body[:f.NoteSpan.Start] + wrapNote(want) + body[f.NoteSpan.End:], nil
		}
		if want == "" {
			return body, nil // nothing to remove
		}
		// First note: append it as the section's last block. Trimming back over the
		// section's trailing blank lines and reusing them as the separator is what keeps
		// the gap to the next finding at exactly one blank line.
		cut := f.SectionSpan.End
		for cut > f.SectionSpan.Start && isBlank(body[cut-1]) {
			cut--
		}
		tail := body[cut:]
		if tail == "" {
			tail = "\n"
		}
		return body[:cut] + "\n\n" + wrapNote(want) + tail, nil
	}
	return "", fmt.Errorf("%w: no finding %q in this audit", ErrNotFound, code)
}
