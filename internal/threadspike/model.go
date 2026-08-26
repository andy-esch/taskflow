// Package threadspike is an explicitly experimental, non-public vertical slice
// for validating ADR-0006. Its types are intentionally isolated from the public
// domain and wire contracts until the spike recommends whether they deserve to
// become production concepts.
package threadspike

import (
	"sort"

	"github.com/andy-esch/taskflow/internal/domain"
)

const (
	// Dir is the proposed flat entity directory. It lives here rather than in
	// domain.Layout while this remains a decision spike.
	Dir = "threads"

	PlanSchema = 1
)

// ThreadStatus is the proposed persisted Thread lifecycle.
type ThreadStatus string

const (
	ThreadUnstarted  ThreadStatus = "unstarted"
	ThreadInProgress ThreadStatus = "in-progress"
	ThreadCompleted  ThreadStatus = "completed"
	ThreadAbandoned  ThreadStatus = "abandoned"
)

func (s ThreadStatus) Valid() bool {
	switch s {
	case ThreadUnstarted, ThreadInProgress, ThreadCompleted, ThreadAbandoned:
		return true
	default:
		return false
	}
}

// Task is the minimum graph-bearing projection of a production task. Keeping
// DependsOn beside (not yet inside) domain.Task lets the spike parse the proposed
// field without silently publishing it through schema/wire contracts.
type Task struct {
	Record    domain.Task
	DependsOn []string
	Body      string
}

// Thread is the proposed persisted initiative document. Tasks is a semantic set
// serialized in stable-ID order; its position carries no execution meaning.
type Thread struct {
	ID          string       `yaml:"id"`
	Slug        string       `yaml:"-"`
	Status      ThreadStatus `yaml:"status"`
	Description string       `yaml:"description"`
	Goal        string       `yaml:"goal"`
	Created     string       `yaml:"created"`
	StartedAt   string       `yaml:"started_at,omitempty"`
	EndedAt     string       `yaml:"ended_at,omitempty"`
	Tags        []string     `yaml:"tags,omitempty"`
	Tasks       []string     `yaml:"tasks"`
	Path        string       `yaml:"-"`
	Body        string       `yaml:"-"`
}

// Snapshot is one repository-consistent input to graph analysis or apply
// validation. Problems make graph-sensitive mutation fail closed.
type Snapshot struct {
	RepoID   string
	Tasks    map[string]Task
	Threads  map[string]Thread
	Epics    map[string]bool
	Problems []domain.FileProblem
}

func sortedUnique(values []string) []string {
	seen := make(map[string]bool, len(values))
	out := make([]string, 0, len(values))
	for _, value := range values {
		if !seen[value] {
			seen[value] = true
			out = append(out, value)
		}
	}
	sort.Strings(out)
	return out
}

func contains(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}
