package core

import "github.com/andy-esch/taskflow/internal/domain"

// LintEntityKind identifies the planning record whose source could not be
// decoded. It is application vocabulary, not a filesystem directory name: a
// remote adapter reports the same kind without manufacturing a local path.
type LintEntityKind string

const (
	LintEntityTask     LintEntityKind = "task"
	LintEntityEpic     LintEntityKind = "epic"
	LintEntityAudit    LintEntityKind = "audit"
	LintEntityResearch LintEntityKind = "research"
	LintEntityThread   LintEntityKind = "thread"
)

// LintLoadProblem is one non-fatal failed-record diagnostic. Stable identity is
// explicit when the adapter can recover it; Location is optional repair context
// and must never be parsed by core to reconstruct identity. LocationIsPath lets
// the wire adapter preserve the historical `path` compatibility field for local
// repositories without labelling a remote URI as a filesystem path.
type LintLoadProblem struct {
	EntityKind     LintEntityKind
	EntityID       string
	EntitySlug     string
	Location       string
	LocationIsPath bool
	Message        string
}

// LintSource is the consumer-owned read port for repository lint. Each method
// returns the decoded records and neutral load failures from one resilient scan;
// keeping the methods separate lets `audit lint` read audits without scanning
// unrelated entity kinds.
type LintSource interface {
	AuditSnapshotSource
	ReadLintTasks() ([]TaskWithBody, []LintLoadProblem, error)
	ReadLintEpics() ([]domain.Epic, []LintLoadProblem, error)
	ReadLintResearch() ([]domain.Research, []LintLoadProblem, error)
}

func lintLoadProblemLabel(problem LintLoadProblem) string {
	if problem.EntitySlug != "" {
		return problem.EntitySlug
	}
	if problem.EntityID != "" {
		return problem.EntityID
	}
	if problem.Location != "" {
		return problem.Location
	}
	if problem.EntityKind != "" {
		return "unidentified " + string(problem.EntityKind) + " record"
	}
	return "unidentified planning record"
}
