package core

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/andy-esch/taskflow/internal/domain"
)

// Service is the application core: framework-agnostic use cases over the ports.
// It has no fs and no cobra, so it is testable in isolation and reusable by a
// future TUI primary adapter.
type Service struct {
	store Store
}

// NewService wires the core to its store.
func NewService(store Store) *Service { return &Service{store: store} }

// TaskFilter narrows a task listing. Zero-valued fields are ignored. When no
// explicit Status is given and All is false, only active tasks are returned.
type TaskFilter struct {
	Status string
	Epic   string
	Tag    string
	All    bool
}

var activeStatuses = map[domain.Status]bool{
	domain.StatusNextUp:       true,
	domain.StatusReadyToStart: true,
	domain.StatusInProgress:   true,
}

// ListTasks returns tasks matching the filter, plus any per-file load problems.
// Filtering is pure (no I/O).
func (s *Service) ListTasks(f TaskFilter) ([]domain.Task, []domain.FileProblem, error) {
	all, problems, err := s.store.ListTasks()
	if err != nil {
		return nil, nil, err
	}
	activeOnly := f.Status == "" && !f.All
	out := make([]domain.Task, 0, len(all))
	for _, t := range all {
		if activeOnly && !activeStatuses[t.Status] {
			continue
		}
		if f.Status != "" && string(t.Status) != f.Status {
			continue
		}
		if f.Epic != "" && t.Epic != f.Epic {
			continue
		}
		if f.Tag != "" && !hasTag(t.Tags, f.Tag) {
			continue
		}
		out = append(out, t)
	}
	return out, problems, nil
}

// ShowTask returns one task plus its markdown body.
func (s *Service) ShowTask(slug string) (domain.Task, string, error) {
	return s.store.GetTask(slug)
}

// Move transitions a task to the given status (lifecycle engine behind the
// explicit verbs). Moving to the current status is an idempotent no-op.
func (s *Service) Move(slug string, to domain.Status) (domain.Task, error) {
	return s.store.Move(slug, to, time.Now())
}

// SetFields validates and applies frontmatter updates to a task (stamping
// updated_at) in a single atomic write. On any invalid value it returns
// ErrValidation and nothing is written.
func (s *Service) SetFields(slug string, updates map[string]any) (domain.Task, error) {
	if len(updates) == 0 {
		return domain.Task{}, fmt.Errorf("%w: no fields given", domain.ErrValidation)
	}
	for field, val := range updates {
		if err := domain.ValidateField(field, stringify(val)); err != nil {
			return domain.Task{}, err
		}
	}
	withMeta := make(map[string]any, len(updates)+1)
	for k, v := range updates {
		withMeta[k] = v
	}
	withMeta["updated_at"] = time.Now().Format("2006-01-02")
	return s.store.SetFields(slug, withMeta)
}

// EpicSummary is an epic plus its task rollup.
type EpicSummary struct {
	Epic  domain.Epic
	Total int
	Done  int
}

// Percent is the completed share of the epic's tasks.
func (e EpicSummary) Percent() int {
	if e.Total == 0 {
		return 0
	}
	return e.Done * 100 / e.Total
}

// ListEpics returns every epic with its task rollup (joined on the tasks'
// `epic:` field), plus any per-file load problems from either scan.
func (s *Service) ListEpics() ([]EpicSummary, []domain.FileProblem, error) {
	epics, ep1, err := s.store.ListEpics()
	if err != nil {
		return nil, nil, err
	}
	tasks, ep2, err := s.store.ListTasks()
	if err != nil {
		return nil, nil, err
	}
	problems := append(ep1, ep2...)
	idx := make(map[string]*EpicSummary, len(epics))
	out := make([]EpicSummary, len(epics))
	for i := range epics {
		out[i] = EpicSummary{Epic: epics[i]}
		idx[epics[i].ID] = &out[i]
	}
	for _, t := range tasks {
		es, ok := idx[t.Epic]
		if !ok {
			continue
		}
		es.Total++
		if t.Status == domain.StatusCompleted {
			es.Done++
		}
	}
	return out, problems, nil
}

// ShowEpic returns an epic, the tasks that belong to it, and its body.
func (s *Service) ShowEpic(id string) (domain.Epic, []domain.Task, string, error) {
	epic, body, err := s.store.GetEpic(id)
	if err != nil {
		return domain.Epic{}, nil, "", err
	}
	tasks, _, err := s.store.ListTasks()
	if err != nil {
		return domain.Epic{}, nil, "", err
	}
	var its []domain.Task
	for _, t := range tasks {
		if t.Epic == id {
			its = append(its, t)
		}
	}
	return epic, its, body, nil
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
		if !activeStatuses[t.Status] {
			continue
		}
		if issues := domain.LintTask(t, validEpic); len(issues) > 0 {
			results = append(results, LintResult{Slug: t.Slug, Issues: issues})
		}
	}
	return results, problems, nil
}

// LintFix applies safe text-level frontmatter repairs across the planning
// tree (or previews them when dryRun is true), returning the files changed.
func (s *Service) LintFix(dryRun bool) ([]domain.FixResult, error) {
	return s.store.FixFrontmatter(dryRun)
}

// ListAudits returns audits in the requested bucket (default: open), plus any
// per-file load problems. bucket="" + all=false means open only.
func (s *Service) ListAudits(bucket string, all bool) ([]domain.Audit, []domain.FileProblem, error) {
	audits, problems, err := s.store.ListAudits()
	if err != nil {
		return nil, nil, err
	}
	out := make([]domain.Audit, 0, len(audits))
	for _, a := range audits {
		switch {
		case bucket != "":
			if string(a.Bucket) != bucket {
				continue
			}
		case !all && a.Bucket != domain.AuditOpen:
			continue
		}
		out = append(out, a)
	}
	return out, problems, nil
}

// ShowAudit returns one audit plus its body.
func (s *Service) ShowAudit(slug string) (domain.Audit, string, error) {
	return s.store.GetAudit(slug)
}

// MoveAudit relocates an audit to another bucket (close/reopen/defer).
func (s *Service) MoveAudit(slug string, to domain.AuditBucket) (domain.Audit, error) {
	return s.store.MoveAudit(slug, to)
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
