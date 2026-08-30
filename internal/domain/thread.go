package domain

import (
	"fmt"
	"slices"
	"sort"
	"strings"

	"github.com/andy-esch/taskflow/internal/id"
)

// ThreadStatus is the persisted initiative lifecycle. Thread lifecycle changes
// are deliberately separate from member-task lifecycle and dependency state.
type ThreadStatus string

const (
	ThreadStatusUnstarted  ThreadStatus = "unstarted"
	ThreadStatusInProgress ThreadStatus = "in-progress"
	ThreadStatusCompleted  ThreadStatus = "completed"
	ThreadStatusCancelled  ThreadStatus = "cancelled"
)

var threadStatuses = []ThreadStatus{
	ThreadStatusUnstarted,
	ThreadStatusInProgress,
	ThreadStatusCompleted,
	ThreadStatusCancelled,
}

// AllThreadStatuses returns the closed lifecycle vocabulary in display order.
func AllThreadStatuses() []ThreadStatus { return append([]ThreadStatus(nil), threadStatuses...) }

// ValidateThreadStatus rejects values outside the persisted Thread lifecycle.
func ValidateThreadStatus(status ThreadStatus) error {
	if slices.Contains(threadStatuses, status) {
		return nil
	}
	if status == "abandoned" {
		return fmt.Errorf("%w: legacy thread status %q was replaced by %q; update the Thread frontmatter", ErrValidation, status, ThreadStatusCancelled)
	}
	values := make([]string, len(threadStatuses))
	for i, value := range threadStatuses {
		values[i] = string(value)
	}
	return fmt.Errorf("%w: invalid thread status %q (valid: %s)", ErrValidation, status, strings.Join(values, ", "))
}

// Thread is a named initiative view over the repository-global task DAG. It
// persists metadata and membership only; task files remain the dependency source
// of truth and all graph-derived state is computed at read time.
type Thread struct {
	Slug          string `yaml:"-"`
	Path          string `yaml:"-"`
	SourceVersion string `yaml:"-"`

	ID         string `yaml:"id"`
	FilenameID string `yaml:"-"`

	Status      ThreadStatus `yaml:"status"`
	Description string       `yaml:"description"`
	Goal        string       `yaml:"goal"`
	TargetDate  string       `yaml:"target_date,omitempty"`
	Created     string       `yaml:"created"`
	Updated     string       `yaml:"updated_at,omitempty"`
	StartedAt   string       `yaml:"started_at,omitempty"`
	EndedAt     string       `yaml:"ended_at,omitempty"`
	Tags        []string     `yaml:"tags,omitempty"`
	Tasks       []string     `yaml:"tasks"`
}

// ValidateThreadDocument enforces invariants shared by guarded creation and
// future guarded Thread mutations. Readers may still retain malformed documents
// as FileProblems so diagnostic commands can explain direct hand edits.
func ValidateThreadDocument(thread Thread) error {
	if !id.Valid(thread.ID) {
		return fmt.Errorf("%w: thread id %q is not a stable id", ErrValidation, thread.ID)
	}
	if err := ValidateThreadStatus(thread.Status); err != nil {
		return err
	}
	if strings.TrimSpace(thread.Description) == "" {
		return fmt.Errorf("%w: thread description is required", ErrValidation)
	}
	if err := ValidateDescription(thread.Description); err != nil {
		return err
	}
	if strings.TrimSpace(thread.Goal) == "" {
		return fmt.Errorf("%w: thread goal is required", ErrValidation)
	}
	if strings.ContainsAny(thread.Goal, "\r\n") {
		return fmt.Errorf("%w: thread goal must be a single line", ErrValidation)
	}
	if err := ValidateDate(thread.Created); err != nil {
		return fmt.Errorf("thread created: %w", err)
	}
	for field, value := range map[string]string{
		"target_date": thread.TargetDate,
		"updated_at":  thread.Updated,
		"started_at":  thread.StartedAt,
		"ended_at":    thread.EndedAt,
	} {
		if value != "" {
			if err := ValidateDate(value); err != nil {
				return fmt.Errorf("thread %s: %w", field, err)
			}
		}
	}
	seen := make(map[string]bool, len(thread.Tasks))
	for _, taskID := range thread.Tasks {
		if !id.Valid(taskID) {
			return fmt.Errorf("%w: thread task id %q is not a stable task id", ErrValidation, taskID)
		}
		if seen[taskID] {
			return fmt.Errorf("%w: thread repeats task id %s", ErrValidation, taskID)
		}
		seen[taskID] = true
	}
	if !sort.StringsAreSorted(thread.Tasks) {
		return fmt.Errorf("%w: thread task ids must be sorted", ErrValidation)
	}
	return nil
}
