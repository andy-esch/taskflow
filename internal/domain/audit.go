package domain

import "fmt"

// AuditBucket is an audit's lifecycle state, identical to the directory it
// lives in (audits/<bucket>/). Findings have their own per-finding status
// inside the body; the audit-level state is just the bucket.
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

// Audit is a code-audit document. Its bucket is the directory; finding counts
// are parsed from the body.
type Audit struct {
	Slug   string      `yaml:"-"`
	Path   string      `yaml:"-"`
	Bucket AuditBucket `yaml:"-"`
	Area   string      `yaml:"area"`
	Date   string      `yaml:"date"`

	Findings     int `yaml:"-"`
	OpenFindings int `yaml:"-"`
}

// Resolved is the number of findings no longer open (Findings − OpenFindings) —
// the audit analog of an epic's Done count. "Open" is the single not-yet-handled
// state (see CountOpenFindings); everything else (in-progress, fixed, landed,
// deferred, superseded, wontfix) counts as resolved for the rollup.
func (a Audit) Resolved() int { return a.Findings - a.OpenFindings }

// Percent is the share of findings resolved, 0–100 (0 when there are none) —
// mirroring EpicSummary.Percent so both rollups (and their bars) read the same.
func (a Audit) Percent() int {
	if a.Findings == 0 {
		return 0
	}
	return a.Resolved() * 100 / a.Findings
}
