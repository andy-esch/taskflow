package domain

import "fmt"

// AuditBucket is an audit's lifecycle state, authoritative in frontmatter (ADR-0003
// §4). Under the flat, id-led layout there is no bucket directory — the value stands
// on its own in the file; a missing/unrecognized one is lint-flagged. Findings have
// their own per-finding status inside the body; the audit-level state is the bucket.
type AuditBucket string

const (
	AuditOpen     AuditBucket = "open"
	AuditClosed   AuditBucket = "closed"
	AuditDeferred AuditBucket = "deferred"
)

var auditBuckets = []AuditBucket{AuditOpen, AuditClosed, AuditDeferred}

// AllAuditBuckets returns every audit bucket.
func AllAuditBuckets() []AuditBucket { return auditBuckets }

// ParseAuditBucket validates s.
func ParseAuditBucket(s string) (AuditBucket, error) {
	for _, b := range auditBuckets {
		if AuditBucket(s) == b {
			return b, nil
		}
	}
	return "", fmt.Errorf("%w: invalid audit bucket %q (open|closed|deferred)", ErrValidation, s)
}

// Dir is the directory name for this bucket.
func (b AuditBucket) Dir() string { return string(b) }

// Valid reports whether b is a known bucket.
func (b AuditBucket) Valid() bool { _, err := ParseAuditBucket(string(b)); return err == nil }

// Audit is a code-audit document. Its bucket is authoritative in frontmatter
// (ADR-0003 §4, flat layout); finding counts are parsed from the body.
type Audit struct {
	Slug string `yaml:"-"`
	// BucketFellBack is set by the store when the frontmatter bucket is missing or
	// unrecognized — under the flat layout there is no directory to fall back to, so
	// Bucket keeps its raw value; the audit still lists and lint flags it
	// (FrontmatterBucketIssues).
	BucketFellBack bool `yaml:"-"`

	// ID is the stable 12-char identifier (ADR-0003 §3): it leads the flat filename
	// (audits/<id>-<slug>.md) and is the primary resolution key.
	ID string `yaml:"id"`

	// Bucket is the audit's lifecycle state — authoritative, read from frontmatter.
	Bucket AuditBucket `yaml:"bucket"`
	Area   string      `yaml:"area"`
	Date   string      `yaml:"date"`
	// Updated is the audit's own last-edited date (stamped by edit/append). Unlike
	// Date — immutable, part of the slug — this advances on each content edit. A
	// bucket move (close/reopen/defer) rewrites the `bucket:` frontmatter in place but
	// does NOT touch this stamp.
	Updated string `yaml:"updated_at"`

	Findings int `yaml:"-"`
	// UnparsedFindings counts diagnostic-only or repairable finding-like headers,
	// not real findings. It qualifies read completeness without inventing statuses
	// or changing the parsed-finding denominator.
	UnparsedFindings int `yaml:"-"`
	// Per-disposition finding tally (see TallyFindings), the segmented progress
	// bar's source. Open + Active + Done + Dropped ≤ Findings (an unrecognized or
	// missing status, which audit lint flags, counts toward none and falls into the
	// bar's empty track). OpenFindings is kept for the JSON open_findings field, the
	// `-c open` projection, and the "(N open)" detail suffix.
	OpenFindings    int `yaml:"-"` // status: open
	ActiveFindings  int `yaml:"-"` // status: in-progress
	DoneFindings    int `yaml:"-"` // status: fixed, tracked
	DroppedFindings int `yaml:"-"` // status: deferred, superseded, wontfix
}

// Resolved is the audit's SETTLED count — every finding that has reached a terminal
// disposition, whether it was done here (fixed/tracked) or dropped (deferred,
// superseded, wontfix). It is the numerator beside Percent, and the continuous form
// of parsed settlement. Unparsed evidence can still make Settled false even
// when Resolved == Findings.
//
// It used to count only DoneFindings, which made a fully-triaged audit read as
// unfinished — `2026-06-27-consumer-data-flow-architecture` closed with 17 done and 5
// dropped and displayed "77% fixed", contradicting its own "ready to close" state on
// the same line. Deciding not to fix a finding IS resolving it; the segmented bar still
// separates the bands, so nothing about HOW it was settled is lost.
func (a Audit) Resolved() int { return a.DoneFindings + a.DroppedFindings }

// UnsettledFindings includes open, in-progress, missing, and invalid statuses.
// It uses the same terminal bands as progress/readiness without treating an
// unparsed heading as a parsed finding.
func (a Audit) UnsettledFindings() int { return a.Findings - a.Resolved() }

// Percent is the share of findings settled, 0–100 (0 when there are none) — the
// segmented bar's headline number. 100 means every finding has a terminal disposition,
// which for an open audit with complete parsing is exactly ReadyToClose.
func (a Audit) Percent() int {
	if a.Findings == 0 {
		return 0
	}
	return a.Resolved() * 100 / a.Findings
}

// Settled reports whether every finding has reached a terminal disposition — done
// (fixed/tracked) or dropped (deferred/superseded/wontfix) — so an open audit has
// nothing left to work and is a "ready to close" call-to-action. False when any
// finding is still open/in-progress OR carries an unrecognized status (Done +
// Dropped < Findings), when finding-like headers remain unparsed, and for an
// audit with no findings at all.
func (a Audit) Settled() bool {
	return a.UnparsedFindings == 0 && a.Findings > 0 && a.UnsettledFindings() == 0
}

// ReadyToClose is the call-to-action shared by the --json envelope (ready_to_close)
// and the human progress line: an OPEN audit that is Settled has no findings left
// to work and can be closed. A closed/deferred audit is not "ready to close" (it is
// already off the open board), so this is false there regardless of Settled.
func (a Audit) ReadyToClose() bool { return a.Bucket == AuditOpen && a.Settled() }

// ValidateMove is the audit bucket-write policy over adapter-established counts.
// Close and defer require complete parsing, including diagnostic-only ambiguous
// headers, and every parsed finding must have a terminal status. An empty audit is
// allowed even though ReadyToClose deliberately does not advertise it. Reopen
// remains available for repairing any readable audit. Persistence adapters must
// apply this policy to the exact source their write guard protects, before no-op
// or dry-run returns, never to a separate preflight read.
func (a Audit) ValidateMove(to AuditBucket) error {
	if !to.Valid() {
		return fmt.Errorf("%q: %w", to, ErrValidation)
	}
	if to == AuditOpen {
		return nil
	}
	if a.UnparsedFindings > 0 {
		return &AuditIncompleteEvidenceError{Slug: a.Slug, Count: a.UnparsedFindings, Target: to}
	}
	if unsettled := a.UnsettledFindings(); unsettled > 0 {
		return &AuditUnsettledFindingsError{Slug: a.Slug, Count: unsettled, Target: to}
	}
	return nil
}

// AuditIncompleteEvidenceError describes a semantic refusal without inventing a
// record selector from its display slug or declared ID. Persistence adapters add
// their established source identity for actionable diagnostics.
type AuditIncompleteEvidenceError struct {
	Slug   string
	Count  int
	Target AuditBucket
}

func (e *AuditIncompleteEvidenceError) Error() string {
	return fmt.Sprintf("%v: audit %q has %d unparsed finding-like header(s); repair or clarify them before moving to %s",
		ErrValidation, e.Slug, e.Count, e.Target)
}

func (e *AuditIncompleteEvidenceError) Unwrap() error { return ErrValidation }

// AuditUnsettledFindingsError refuses leaving the open bucket while parsed
// findings lack terminal statuses. Adapters supply source-backed lint advice.
type AuditUnsettledFindingsError struct {
	Slug   string
	Count  int
	Target AuditBucket
}

func (e *AuditUnsettledFindingsError) Error() string {
	return fmt.Sprintf("%v: audit %q has %d unsettled parsed finding(s); every finding needs a terminal status before moving to %s",
		ErrValidation, e.Slug, e.Count, e.Target)
}

func (e *AuditUnsettledFindingsError) Unwrap() error { return ErrValidation }
