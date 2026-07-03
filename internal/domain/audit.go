package domain

import "fmt"

// AuditBucket is an audit's lifecycle state; its string value is also the directory
// an audit mirrors into (audits/<bucket>/). The authoritative value now lives in
// frontmatter (ADR-0003 Phase A) — the directory is a lock-step mirror. Findings have
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

// Audit is a code-audit document. Its bucket is authoritative in frontmatter (the
// directory is a lock-step mirror); finding counts are parsed from the body.
type Audit struct {
	Slug string `yaml:"-"`
	Path string `yaml:"-"`
	// FolderBucket is the bucket the file's directory implies (the mirror). Bucket is
	// the authoritative one (frontmatter, ADR-0003 Phase A); the two diverge only when
	// an audit is misfiled — see Misfiled.
	FolderBucket AuditBucket `yaml:"-"`

	// ID is the stable 12-char identifier (ADR-0003), minted on create. Additive —
	// see domain.Task.ID; empty on pre-rollout files.
	ID string `yaml:"id"`

	// Bucket is the audit's lifecycle state — authoritative, read from frontmatter.
	Bucket AuditBucket `yaml:"bucket"`
	Area   string      `yaml:"area"`
	Date   string      `yaml:"date"`
	// Updated is the audit's own last-edited date (stamped by edit/append). Unlike
	// Date — immutable, part of the slug — this advances on each content edit. A
	// bucket move (close/reopen/defer) rewrites the `bucket:` frontmatter and relocates
	// the file, but does NOT touch this stamp.
	Updated string `yaml:"updated_at"`

	Findings int `yaml:"-"`
	// Per-disposition finding tally (see TallyFindings), the segmented progress
	// bar's source. Open + Active + Done + Dropped ≤ Findings (an unrecognized or
	// missing status, which audit lint flags, counts toward none and falls into the
	// bar's empty track). OpenFindings is kept for the JSON open_findings field, the
	// `-c open` projection, and the "(N open)" detail suffix.
	OpenFindings    int `yaml:"-"` // status: open
	ActiveFindings  int `yaml:"-"` // status: in-progress
	DoneFindings    int `yaml:"-"` // status: fixed, landed
	DroppedFindings int `yaml:"-"` // status: deferred, superseded, wontfix
}

// Resolved is the bar's "done" count — findings fixed or landed (DoneFindings),
// the audit analog of an epic's Done. Findings that are merely parked/dropped
// (deferred, superseded, wontfix) or in-progress are NOT counted here; the
// segmented bar shows those as their own bands.
func (a Audit) Resolved() int { return a.DoneFindings }

// Percent is the share of findings done (fixed/landed), 0–100 (0 when there are
// none) — the segmented bar's headline number and its green band's reach.
func (a Audit) Percent() int {
	if a.Findings == 0 {
		return 0
	}
	return a.DoneFindings * 100 / a.Findings
}

// Settled reports whether every finding has reached a terminal disposition — done
// (fixed/landed) or dropped (deferred/superseded/wontfix) — so an open audit has
// nothing left to work and is a "ready to close" call-to-action. False when any
// finding is still open/in-progress OR carries an unrecognized status (Done +
// Dropped < Findings), and for an audit with no findings at all.
func (a Audit) Settled() bool {
	return a.Findings > 0 && a.DoneFindings+a.DroppedFindings == a.Findings
}

// Misfiled reports whether the file's directory disagrees with its authoritative
// (frontmatter) bucket — the mirror is stale and the file should be moved to match.
// Guarded on a valid FolderBucket so an Audit with no folder context is never
// misfiled; a missing/foreign frontmatter bucket falls back to the folder in
// parseAudit, so those arrive here equal and aren't misfiled either.
func (a Audit) Misfiled() bool {
	return a.FolderBucket.Valid() && a.Bucket != a.FolderBucket
}
