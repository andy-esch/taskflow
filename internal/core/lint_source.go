package core

import (
	"cmp"
	"slices"

	"github.com/andy-esch/taskflow/internal/domain"
)

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
// explicit when the adapter can recover it; Location is optional adapter-neutral
// context and must never be parsed by core to reconstruct identity. Path is
// separate, optional local repair context: an adapter may legitimately provide
// both an opaque Location and a filesystem Path. LocationIsPath declares that
// Location itself is also a path and remains as a compatibility fallback for
// producers that have not populated Path independently.
type LintLoadProblem struct {
	EntityKind     LintEntityKind
	EntityID       string
	EntitySlug     string
	Location       string
	LocationIsPath bool
	Path           string
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
	if problem.Path != "" {
		return problem.Path
	}
	if problem.EntityKind != "" {
		return "unidentified " + string(problem.EntityKind) + " record"
	}
	return "unidentified planning record"
}

// lintLoadProblemsFromFiles is the compatibility boundary for aggregate store
// reads that still expose domain.FileProblem. Adapters recover identity before
// returning the value; core copies it and never parses Location for meaning.
func lintLoadProblemsFromFiles(kind LintEntityKind, problems []domain.FileProblem) []LintLoadProblem {
	out := make([]LintLoadProblem, 0, len(problems))
	for _, problem := range problems {
		out = append(out, LintLoadProblem{
			EntityKind: kind, EntityID: problem.EntityID, EntitySlug: problem.EntitySlug,
			Location: problem.Path, LocationIsPath: problem.Path != "", Path: problem.Path,
			Message: problem.Message,
		})
	}
	return out
}

// canonicalLintLoadProblems makes public aggregate projections independent of
// adapter return order. Kind rank preserves the dashboard's task → epic → audit
// grouping while the remaining explicit fields provide stable within-kind order;
// location is never interpreted to derive identity.
func canonicalLintLoadProblems(problems []LintLoadProblem) []LintLoadProblem {
	out := append([]LintLoadProblem(nil), problems...)
	slices.SortFunc(out, func(left, right LintLoadProblem) int {
		return cmp.Or(
			cmp.Compare(lintEntityKindRank(left.EntityKind), lintEntityKindRank(right.EntityKind)),
			cmp.Compare(string(left.EntityKind), string(right.EntityKind)),
			cmp.Compare(left.EntityID, right.EntityID),
			cmp.Compare(left.EntitySlug, right.EntitySlug),
			cmp.Compare(boolRank(left.LocationIsPath), boolRank(right.LocationIsPath)),
			cmp.Compare(left.Location, right.Location),
			cmp.Compare(left.Path, right.Path),
			cmp.Compare(left.Message, right.Message),
		)
	})
	return out
}

func lintEntityKindRank(kind LintEntityKind) int {
	switch kind {
	case LintEntityTask:
		return 0
	case LintEntityEpic:
		return 1
	case LintEntityAudit:
		return 2
	case LintEntityResearch:
		return 3
	case LintEntityThread:
		return 4
	default:
		return 5
	}
}

func boolRank(value bool) int {
	if value {
		return 1
	}
	return 0
}
