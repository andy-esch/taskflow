package core

// AuditSnapshot is one resilient, body-aware audit read. Readable audit
// records retain every finding-derived projection from the same source bytes;
// unreadable records retain portable identity and optional repair context.
// Keeping both outcomes in one value makes it impossible for a consumer to
// accidentally combine records and diagnostics from different scans.
type AuditSnapshot struct {
	Audits   []AuditWithFindings
	Problems []LintLoadProblem
}

// AuditSnapshotSource is the consumer-owned read port shared by finding
// queries and audit lint. An empty selector reads the resilient repository
// snapshot; a non-empty selector resolves one audit using the ordinary
// exact/prefix/substring contract and returns a one-record snapshot. It
// deliberately carries no path-keyed lookup: an adapter may read local files,
// remote objects, or an in-memory corpus and still expose the same application
// snapshot.
type AuditSnapshotSource interface {
	ReadAuditSnapshot(selector string) (AuditSnapshot, error)
}
