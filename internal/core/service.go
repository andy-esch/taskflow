package core

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/andy-esch/taskflow/internal/domain"
	"github.com/andy-esch/taskflow/internal/id"
)

// Service is the application core: framework-agnostic use cases over the ports.
// It has no fs and no cobra, so it is testable in isolation and reused by both
// primary adapters (the cli and the tui).
type Service struct {
	store      Store
	templates  TemplateSource
	now        func() time.Time  // wall clock, injectable for deterministic snooze/revisit queries
	newID      func() string     // stable-id mint (default id.New), injectable so created-file tests are deterministic
	maxRetries int               // bounded OCC auto-retry for scriptable mutations (see retryOnConflict / WithRetry)
	retrySleep func(attempt int) // backoff+jitter before a retry; injectable so tests run instantly
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

// WithClock overrides the Service wall clock — injected so snooze/revisit queries
// (the Summary `revisit_due` count and `task list --revisit-due`) compute "has the
// revisit date arrived?" against a fixed time in tests. Defaults to time.Now.
func WithClock(now func() time.Time) Option {
	return func(s *Service) {
		if now != nil {
			s.now = now
		}
	}
}

// WithIDGen overrides the stable-id generator the create paths mint from — injected
// so a test that snapshots a *created file* gets a fixed id instead of a random one.
// Defaults to id.New. Mirrors WithClock: the two non-deterministic create-time inputs,
// time and identity, are both injectable.
func WithIDGen(gen func() string) Option {
	return func(s *Service) {
		if gen != nil {
			s.newID = gen
		}
	}
}

// NewService wires the core to its store; templates default to the built-in
// source unless WithTemplateSource overrides it.
func NewService(store Store, opts ...Option) *Service {
	s := &Service{store: store, templates: builtinTemplates{}, now: time.Now, newID: id.New, maxRetries: defaultMaxRetries, retrySleep: defaultRetrySleep}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

// Now is the Service's wall clock (the injected one, default time.Now). Exposed so
// the adapters compute snooze/revisit due-ness against the SAME clock the core
// uses for its date stamps — a WithClock injection then governs everything.
func (s *Service) Now() time.Time { return s.now() }

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
	Counts        []StatusCount        // every status in display order (count may be 0)
	InProgress    []domain.Task        // the in-progress working set
	Epics         []EpicSummary        // epic rollups, most-recently-updated first (the one dashboard order both `status` and the TUI render)
	OpenAudits    []domain.Audit       // audits still in the open bucket (actionable work)
	ReadyToClose  int                  // open audits with every finding resolved/dropped ("ready to close") — the aggregate, computed once here so no surface re-derives it off OpenAudits (audit M9)
	Findings      FindingsRollup       // actionable audit findings (open/in-progress) aggregated by urgency + component
	Misfiled      int                  // tasks whose status disagrees with their folder
	RevisitDue    int                  // deferred tasks whose revisit_at (snooze-until) date has arrived
	BadEpicStatus int                  // epics whose status is outside the canonical vocabulary (a fixable data problem, not dropped)
	Problems      []domain.FileProblem // unreadable files
}

// Summary composes a one-screen overview from a single scan of tasks + epics +
// audits. Only OPEN audits are surfaced — the actionable subset, paralleling the
// in-progress task working set (closed/deferred audits are done/parked).
func (s *Service) Summary() (Summary, error) {
	tasks, p1, err := s.store.ListTasks()
	if err != nil {
		return Summary{}, err
	}
	epics, p2, err := s.store.ListEpics()
	if err != nil {
		return Summary{}, err
	}
	// One scan of every audit body serves BOTH the open-bucket list (audit-level
	// tallies) AND the findings rollup below — the store hands back the findings it
	// already parsed for the tally, so Summary never re-reads a body (the H2 fix).
	audits, p3, err := s.store.ListAuditsWithFindings()
	if err != nil {
		return Summary{}, err
	}
	var openAudits []domain.Audit
	var actionable []AuditFinding
	readyToClose := 0
	for _, a := range audits {
		if a.Audit.Bucket == domain.AuditOpen {
			openAudits = append(openAudits, a.Audit)
			// "Ready to close" = an open audit with nothing left to work (every finding
			// resolved or dropped). Counted once here from the same scan, so the CLI and
			// TUI dashboards read s.ReadyToClose instead of each re-walking OpenAudits.
			if a.Audit.Settled() {
				readyToClose++
			}
		}
		// Actionable audit findings (open / in-progress) across ALL audits — finding
		// status, not audit bucket, is what's actionable (an open-bucket audit can be
		// all-superseded; a closed one shouldn't carry open findings but lint catches
		// that). Same filter + (audit, document) order as QueryFindings would produce.
		for _, fd := range a.Findings {
			if isActionableFinding(fd) {
				actionable = append(actionable, AuditFinding{Finding: fd, Audit: a.Audit.Slug, Bucket: string(a.Audit.Bucket)})
			}
		}
	}
	counts := map[domain.Status]int{}
	var inProgress []domain.Task
	misfiled := 0
	revisitDue := 0
	now := s.now()
	for _, t := range tasks {
		counts[t.Status]++
		if t.Status == domain.StatusInProgress {
			inProgress = append(inProgress, t)
		}
		if t.Misfiled() {
			misfiled++
		}
		// Move clears revisit_at when a task leaves deferred, so a stray date on a
		// non-deferred task is only possible via a manual `task set`/edit; either
		// way the nudge stays scoped to tasks parked in deferred/ whose snooze date
		// has arrived.
		if domain.IsTaskRevisitDue(t, now) {
			revisitDue++
		}
	}
	ordered := make([]StatusCount, 0, len(domain.AllStatuses()))
	for _, st := range domain.AllStatuses() {
		ordered = append(ordered, StatusCount{Status: st, Count: counts[st]})
	}
	// Epics with a status outside the canonical vocabulary are a fixable data problem
	// (lint reports each by id); surface the count so the dashboard can nudge — these
	// epics are NOT dropped (dashboardEpics fails open), just flagged.
	badEpicStatus := 0
	for _, e := range epics {
		if !domain.IsKnownEpicStatus(e.Status) {
			badEpicStatus++
		}
	}
	return Summary{
		Counts:     ordered,
		InProgress: inProgress,
		// Active-only + live-first HERE so the dashboard's "what's live right now"
		// lens is a property of the aggregate, not re-derived per surface — the CLI
		// `status` and the TUI dashboard then agree by construction (audit M2).
		// Retired/deprecated epics and drained-dormant ordering are handled in
		// dashboardEpics; the entity list / `epic list` keep the full roster in store
		// order via rollupEpics directly.
		Epics:         dashboardEpics(rollupEpics(epics, tasks)),
		OpenAudits:    openAudits,
		ReadyToClose:  readyToClose,
		Findings:      rollupFindings(actionable),
		Misfiled:      misfiled,
		RevisitDue:    revisitDue,
		BadEpicStatus: badEpicStatus,
		Problems:      append(append(p1, p2...), p3...),
	}, nil
}

// LintResult is the set of frontmatter issues for one entity (a task by slug, or
// an epic by id — the Slug field carries whichever as the label).
type LintResult struct {
	Slug   string
	Issues []domain.Issue
}

// Lint validates active tasks' frontmatter (joining against known epics for the
// epic-existence check) AND the epics themselves. Returns one LintResult per
// task or epic with issues.
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
			// Archived tasks skip the field nags but still get the universal checks:
			// status/folder drift, a missing/unrecognized frontmatter status, and a
			// missing stable id.
			issues = append(domain.MisfiledIssues(t), domain.FrontmatterStatusIssues(t)...)
			issues = append(issues, domain.MissingIDIssue(t.ID)...)
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
	// Epics get linted too: the same closed status vocabulary plus priority and a
	// present description (a deprecated epic is spared the field nags — see
	// domain.LintEpic). The epic id slots into Slug as the result's label.
	for _, e := range epics {
		if issues := domain.LintEpic(e); len(issues) > 0 {
			results = append(results, LintResult{Slug: e.ID, Issues: issues})
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
		g.statuses = append(g.statuses, string(t.FolderStatus)) // the physical dirs, not the (authoritative) frontmatter status
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
