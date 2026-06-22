package core

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/andy-esch/taskflow/internal/domain"
)

// Service is the application core: framework-agnostic use cases over the ports.
// It has no fs and no cobra, so it is testable in isolation and reused by both
// primary adapters (the cli and the tui).
type Service struct {
	store Store
}

// NewService wires the core to its store.
func NewService(store Store) *Service { return &Service{store: store} }

// WatchPaths exposes the store's watchable directory set to the TUI, so the
// fs-layout knowledge stays behind the port instead of being rebuilt in the
// watcher (the TUI never reconstructs the planning tree's shape itself).
func (s *Service) WatchPaths() []string { return s.store.WatchPaths() }

// templateBodyConflict rejects supplying both an explicit body and a --template:
// they're mutually exclusive (override the scaffold OR pick one). Enforced in core
// so the declared contract holds for every adapter, not just cobra's flag check.
func templateBodyConflict(body, template string) error {
	if body != "" && template != "" {
		return fmt.Errorf("%w: --body/--body-file and --template are mutually exclusive", domain.ErrValidation)
	}
	return nil
}

// StatusCount is the number of tasks in a status (for the dashboard).
type StatusCount struct {
	Status domain.Status
	Count  int
}

// Summary is the at-a-glance project state for the dashboard.
type Summary struct {
	Counts     []StatusCount        // every status in display order (count may be 0)
	InProgress []domain.Task        // the in-progress working set
	Epics      []EpicSummary        // epic rollups
	Misfiled   int                  // tasks whose status disagrees with their folder
	Problems   []domain.FileProblem // unreadable files
}

// Summary composes a one-screen overview from a single scan of tasks + epics.
func (s *Service) Summary() (Summary, error) {
	tasks, p1, err := s.store.ListTasks()
	if err != nil {
		return Summary{}, err
	}
	epics, p2, err := s.store.ListEpics()
	if err != nil {
		return Summary{}, err
	}
	counts := map[domain.Status]int{}
	var inProgress []domain.Task
	misfiled := 0
	for _, t := range tasks {
		counts[t.Status]++
		if t.Status == domain.StatusInProgress {
			inProgress = append(inProgress, t)
		}
		if t.Misfiled() {
			misfiled++
		}
	}
	ordered := make([]StatusCount, 0, len(domain.AllStatuses()))
	for _, st := range domain.AllStatuses() {
		ordered = append(ordered, StatusCount{Status: st, Count: counts[st]})
	}
	return Summary{
		Counts:     ordered,
		InProgress: inProgress,
		Epics:      rollupEpics(epics, tasks),
		Misfiled:   misfiled,
		Problems:   append(p1, p2...),
	}, nil
}

// LintResult is the set of frontmatter issues for one task.
type LintResult struct {
	Slug   string
	Issues []domain.Issue
}

// Lint validates active tasks' frontmatter, joining against known epics for the
// epic-existence check. Returns one LintResult per task with issues.
func (s *Service) Lint() ([]LintResult, []domain.FileProblem, error) {
	tasks, problems, err := s.store.ListTasks()
	if err != nil {
		return nil, nil, err
	}
	epics, ep2, err := s.store.ListEpics()
	if err != nil {
		return nil, nil, err
	}
	problems = append(problems, ep2...)
	valid := make(map[string]bool, len(epics))
	for _, e := range epics {
		valid[e.ID] = true
	}
	validEpic := func(id string) bool { return valid[id] }

	var results []LintResult
	for _, t := range tasks {
		// Active tasks get the full field lint; archived tasks are only checked
		// for status/folder drift (no point nagging about missing fields on a
		// completed item, but a misfiled one should still surface).
		var issues []domain.Issue
		if t.Status.IsActive() {
			issues = domain.LintTask(t, validEpic)
		} else {
			issues = domain.MisfiledIssues(t)
		}
		if len(issues) > 0 {
			results = append(results, LintResult{Slug: t.Slug, Issues: issues})
		}
	}
	// Epic statuses are a closed vocabulary (see domain.ValidateEpicStatus);
	// files predating the enum (or hand-edited ones) surface here.
	for _, e := range epics {
		if err := domain.ValidateEpicStatus(e.Status); err != nil {
			results = append(results, LintResult{Slug: e.ID, Issues: []domain.Issue{
				{Field: "status", Message: err.Error()},
			}})
		}
	}
	return results, problems, nil
}

// LintFix applies safe text-level frontmatter repairs across the planning
// tree (or previews them when dryRun is true), returning the files changed.
func (s *Service) LintFix(dryRun bool) ([]domain.FixResult, error) {
	return s.store.FixFrontmatter(dryRun)
}

func hasTag(tags []string, want string) bool {
	for _, t := range tags {
		if strings.EqualFold(t, want) {
			return true
		}
	}
	return false
}

func stringify(v any) string {
	switch x := v.(type) {
	case string:
		return x
	case int:
		return strconv.Itoa(x)
	case []string:
		return strings.Join(x, ",")
	default:
		return fmt.Sprintf("%v", x)
	}
}
