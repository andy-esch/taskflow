package core

import (
	"cmp"
	"fmt"
	"slices"

	"github.com/andy-esch/taskflow/internal/domain"
	"github.com/andy-esch/taskflow/internal/id"
)

// LintSource is the consumer-owned read port for repository lint. Each method
// returns the decoded records and neutral load failures from one resilient scan;
// keeping the methods separate lets `audit lint` read audits without scanning
// unrelated entity kinds.
type LintSource interface {
	AuditSnapshotSource
	ReadLintTasks() ([]LoadedRecord[TaskWithBody], []LoadProblem, error)
	ReadLintEpics() ([]LoadedRecord[domain.Epic], []LoadProblem, error)
	ReadLintResearch() ([]LoadedRecord[domain.Research], []LoadProblem, error)
}

func crossKindIdentityIssue(stableID, otherKind string) domain.Issue {
	return domain.Issue{Field: "id", Message: fmt.Sprintf(
		"stable id %s is also used by a %s — task and Thread identities must be globally unique", stableID, otherKind)}
}

// Only explicit, well-formed recovered IDs are collision evidence. The adapter
// owns recovery/trust; locations, repair paths, slugs, and messages are never
// parsed here. Preserve existing readable-record policy separately.
func recoveredCrossKindIDIssues(problems []LoadProblem, otherIdentities map[string]bool, otherKind string) map[string]domain.Issue {
	issues := make(map[string]domain.Issue)
	for _, problem := range problems {
		if id.Valid(problem.EntityID) && otherIdentities[problem.EntityID] {
			issues[problem.EntityID] = crossKindIdentityIssue(problem.EntityID, otherKind)
		}
	}
	return issues
}

func loadProblemLabel(problem LoadProblem) string {
	if problem.EntitySlug != "" {
		return problem.EntitySlug
	}
	if problem.EntityID != "" {
		return problem.EntityID
	}
	if problem.Location != "" {
		return problem.Location
	}
	if problem.LocalPath != "" {
		return problem.LocalPath
	}
	if problem.EntityKind != "" {
		return "unidentified " + string(problem.EntityKind) + " record"
	}
	return "unidentified planning record"
}

// canonicalLoadProblems makes public aggregate projections independent of
// adapter return order. Kind rank preserves the dashboard's task → epic → audit
// grouping while the remaining explicit fields provide stable within-kind order;
// location is never interpreted to derive identity.
func canonicalLoadProblems(problems []LoadProblem) []LoadProblem {
	out := append([]LoadProblem(nil), problems...)
	slices.SortFunc(out, func(left, right LoadProblem) int {
		return cmp.Or(
			cmp.Compare(entityKindRank(left.EntityKind), entityKindRank(right.EntityKind)),
			cmp.Compare(string(left.EntityKind), string(right.EntityKind)),
			cmp.Compare(left.EntityID, right.EntityID),
			cmp.Compare(left.EntitySlug, right.EntitySlug),
			cmp.Compare(left.Location, right.Location),
			cmp.Compare(left.LocalPath, right.LocalPath),
			cmp.Compare(left.Message, right.Message),
		)
	})
	return out
}

func entityKindRank(kind EntityKind) int {
	switch kind {
	case EntityTask:
		return 0
	case EntityEpic:
		return 1
	case EntityAudit:
		return 2
	case EntityResearch:
		return 3
	case EntityThread:
		return 4
	default:
		return 5
	}
}
