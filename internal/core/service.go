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
	store     Store
	templates TemplateSource
}

// Option configures a Service at construction. Functional options keep the common
// NewService(store) call unchanged while leaving room for injected ports (the
// template source today; repo-local sources in epic 22).
type Option func(*Service)

// WithTemplateSource overrides the built-in template source — epic 22 wires a
// repo-local source layered over the built-ins, and tests inject a fake to prove
// the list/show/create paths read through the port, not domain.Template* directly.
func WithTemplateSource(src TemplateSource) Option {
	return func(s *Service) {
		if src != nil {
			s.templates = src
		}
	}
}

// NewService wires the core to its store; templates default to the built-in
// source unless WithTemplateSource overrides it.
func NewService(store Store, opts ...Option) *Service {
	s := &Service{store: store, templates: builtinTemplates{}}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

// NewBuiltinTemplateService returns a Service backed only by the built-in
// TemplateSource and no store — for the repo-less self-description surfaces
// (`template list/show`, like `schema`). Only the template methods are safe to
// call on it. When a planning repo IS present, the resolved store-backed Service
// is used instead, and epic 22 layers repo-local templates over the built-ins.
func NewBuiltinTemplateService() *Service { return NewService(nil) }

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
	// Duplicate slug across status dirs: a Ctrl-C in Move's write-then-remove
	// window (or a stray hand-copy) leaves the same slug in two dirs, which makes
	// every later resolve(slug) return ErrAmbiguous — the task can't be shown,
	// moved, or set by name. Both copies are listed (different dirs), so group by
	// slug and flag any with >1. Surfaced loudly here because there's no other
	// signal; status==directory means the dirs are always distinct. (Tasks only:
	// MoveAudit is an atomic rename, so audits have no such window.)
	results = append(results, duplicateSlugIssues(tasks)...)
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

// duplicateSlugIssues flags any slug that appears in more than one status dir,
// reporting the buckets it occupies. Deterministic: groups in first-seen order
// (tasks arrive in status-dir order), so the output is stable across runs.
func duplicateSlugIssues(tasks []domain.Task) []LintResult {
	type group struct{ statuses []string }
	groups := map[string]*group{}
	var order []string
	for _, t := range tasks {
		g, ok := groups[t.Slug]
		if !ok {
			g = &group{}
			groups[t.Slug] = g
			order = append(order, t.Slug)
		}
		g.statuses = append(g.statuses, string(t.Status))
	}
	var out []LintResult
	for _, slug := range order {
		g := groups[slug]
		if len(g.statuses) < 2 {
			continue
		}
		out = append(out, LintResult{Slug: slug, Issues: []domain.Issue{{
			Field: "slug",
			Message: fmt.Sprintf("duplicate: same slug in %d dirs (%s); resolve is ambiguous until you remove the wrong copy",
				len(g.statuses), strings.Join(g.statuses, ", ")),
		}}})
	}
	return out
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
