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
