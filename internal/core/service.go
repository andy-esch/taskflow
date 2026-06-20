package core

import (
	"fmt"
	"strconv"
	"strings"
	"time"

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

// TaskFilter narrows a task listing. Zero-valued fields are ignored. When no
// explicit Status is given and All is false, only active tasks are returned.
type TaskFilter struct {
	Status string
	Epic   string
	Tag    string
	All    bool
}

// ListTasks returns tasks matching the filter, plus any per-file load problems.
// Filter values are validated first — an unknown status or epic returns
// ErrValidation rather than a silently empty list, which agents routing on exit
// codes can't tell apart from an empty bucket. (The epic check costs one
// ListEpics call, only when that filter is set.)
func (s *Service) ListTasks(f TaskFilter) ([]domain.Task, []domain.FileProblem, error) {
	if f.Status != "" {
		if _, err := domain.ParseStatus(f.Status); err != nil {
			return nil, nil, err
		}
	}
	if f.Epic != "" {
		epics, _, err := s.store.ListEpics()
		if err != nil {
			return nil, nil, err
		}
		if !epicExists(epics, f.Epic) {
			return nil, nil, fmt.Errorf("%w: unknown epic %q", domain.ErrValidation, f.Epic)
		}
	}
	all, problems, err := s.store.ListTasks()
	if err != nil {
		return nil, nil, err
	}
	activeOnly := f.Status == "" && !f.All
	out := make([]domain.Task, 0, len(all))
	for _, t := range all {
		if activeOnly && !t.Status.IsActive() {
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

// EditTask opens a task for whole-file editing — the human face of mutation,
// complementing the field-level `task set`. edit (run by the cli's $EDITOR layer)
// receives the current file content and returns the new content; the store
// accepts it only if it still parses as a task, reopening the editor on a broken
// edit. Returns the reloaded task and whether anything changed.
func (s *Service) EditTask(slug string, edit func(current string, prevErr error) (string, error)) (domain.Task, bool, error) {
	return s.store.EditTask(slug, edit)
}

// Move transitions a task to the given status (lifecycle engine behind the
// explicit verbs). Moving to the current status is an idempotent no-op.
// dryRun validates everything and returns the would-be task without writing.
func (s *Service) Move(slug string, to domain.Status, dryRun bool) (domain.Task, error) {
	return s.store.Move(slug, to, time.Now(), dryRun)
}

// SetFields validates and applies frontmatter updates to a task (stamping
// updated_at) in a single atomic write. On any invalid value it returns
// ErrValidation and nothing is written.
//
// Values arriving as strings from the `--set key=value` escape hatch are coerced
// to the native type a known typed field needs (per the domain field registry)
// before the store serializes them — otherwise the store would write a
// corrupting !!str (e.g. tier: "4") that the strict loader then can't read back.
// Keys outside the registry are rejected unless force is set — a typo'd field
// name must not silently persist. A domain.UnsetField value removes the key;
// an empty epic detaches the task (both decided 2026-06-12). When `epic` is set
// non-empty it must exist, mirroring NewTask, so set can't orphan a task out of
// its epic's rollup.
func (s *Service) SetFields(slug string, updates map[string]any, force, dryRun bool) (domain.Task, error) {
	if len(updates) == 0 {
		return domain.Task{}, fmt.Errorf("%w: no fields given", domain.ErrValidation)
	}
	withMeta := make(map[string]any, len(updates)+1)
	for field, val := range updates {
		if _, unset := val.(domain.UnsetField); unset {
			switch field {
			case "status":
				return domain.Task{}, fmt.Errorf("%w: status is the directory — use `task <verb>`/`task move`", domain.ErrValidation)
			case "updated_at":
				return domain.Task{}, fmt.Errorf("%w: updated_at is stamped automatically and cannot be unset", domain.ErrValidation)
			}
			// A typo'd field name must not silently persist (or, here, silently
			// no-op) — gate unset on the registry too, mirroring the set path.
			if !force && !domain.KnownTaskField(field) {
				return domain.Task{}, unknownFieldErr(field)
			}
			withMeta[field] = val
			continue
		}
		if !force && !domain.KnownTaskField(field) {
			return domain.Task{}, unknownFieldErr(field)
		}
		coerced, err := coerceField(field, val)
		if err != nil {
			return domain.Task{}, err
		}
		if err := domain.ValidateField(field, stringify(coerced)); err != nil {
			return domain.Task{}, err
		}
		withMeta[field] = coerced
	}
	if epic, ok := withMeta["epic"].(string); ok {
		if epic == "" {
			withMeta["epic"] = domain.UnsetField{} // detach from the epic
		} else {
			epics, _, err := s.store.ListEpics()
			if err != nil {
				return domain.Task{}, err
			}
			if !epicExists(epics, epic) {
				return domain.Task{}, fmt.Errorf("%w: unknown epic %q", domain.ErrValidation, epic)
			}
		}
	}
	withMeta["updated_at"] = time.Now().Format("2006-01-02")
	return s.store.SetFields(slug, withMeta, dryRun)
}

// unknownFieldErr is the shared rejection for a field outside the registry, used
// by both the set and unset paths of SetFields.
func unknownFieldErr(field string) error {
	return fmt.Errorf(
		"%w: unknown field %q (known fields only; use --force for a custom field)", domain.ErrValidation, field)
}

// coerceField converts a string `--set` value into the native type its field
// needs (int / []string). Values already of the right type (from typed flags) and
// genuinely-custom fields pass through unchanged.
func coerceField(field string, val any) (any, error) {
	str, isStr := val.(string)
	if !isStr {
		return val, nil // a typed flag already supplied the native type
	}
	switch {
	case domain.IntFields[field]:
		n, err := strconv.Atoi(strings.TrimSpace(str))
		if err != nil {
			return nil, fmt.Errorf("%w: %s must be an integer, got %q", domain.ErrValidation, field, str)
		}
		return n, nil
	case domain.ListFields[field]:
		return splitList(str), nil
	}
	return str, nil
}

// splitList parses a comma-separated `--set tags=a,b` value into a trimmed,
// empty-free slice.
func splitList(s string) []string {
	out := []string{}
	for _, part := range strings.Split(s, ",") {
		if p := strings.TrimSpace(part); p != "" {
			out = append(out, p)
		}
	}
	return out
}

// NewTaskParams are the inputs for creating a task. Tier/Autonomy default to 3,
// Priority to "medium", Effort to "Unknown" when zero (set by the CLI flags).
type NewTaskParams struct {
	Title       string
	Epic        string
	Description string
	Effort      string
	Priority    string
	Tier        int
	Autonomy    int
	Tags        []string
	Next        bool   // create in next-up instead of ready-to-start
	Start       bool   // create directly in in-progress (mutually exclusive with Next)
	Body        string // override the default handoff scaffold
	DryRun      bool   // validate + report the would-be task without writing
}

const taskBodyTemplate = `
# %s

## Objective

<why / what — one short paragraph>

## Acceptance criteria

- [ ] <observable outcome>

## Out of scope

- <explicitly excluded>

## Related

- Epic [[%s]]
`

// NewTask validates and creates a task, returning the created task. The epic
// must exist; tier/autonomy/priority/description are validated. On any invalid
// input it returns ErrValidation and nothing is written.
func (s *Service) NewTask(p NewTaskParams) (domain.Task, error) {
	epics, _, err := s.store.ListEpics()
	if err != nil {
		return domain.Task{}, err
	}
	if !epicExists(epics, p.Epic) {
		return domain.Task{}, fmt.Errorf("%w: unknown epic %q", domain.ErrValidation, p.Epic)
	}
	if err := domain.ValidatePriority(p.Priority); err != nil {
		return domain.Task{}, err
	}
	if err := domain.ValidateTier(p.Tier); err != nil {
		return domain.Task{}, err
	}
	if err := domain.ValidateAutonomy(p.Autonomy); err != nil {
		return domain.Task{}, err
	}
	if err := domain.ValidateDescription(p.Description); err != nil {
		return domain.Task{}, err
	}
	// Tags are required at creation (decided 2026-06-12): lint demands non-empty
	// tags on active tasks, and `new` must not scaffold a file its own linter
	// rejects.
	if len(p.Tags) == 0 {
		return domain.Task{}, fmt.Errorf("%w: at least one tag is required (--tags)", domain.ErrValidation)
	}
	if err := domain.ValidateTitle(p.Title); err != nil {
		return domain.Task{}, err
	}
	slug := domain.Slugify(p.Title)
	if slug == "" {
		return domain.Task{}, fmt.Errorf("%w: title produced an empty slug: %q", domain.ErrValidation, p.Title)
	}
	status := domain.StatusReadyToStart
	switch {
	case p.Start:
		status = domain.StatusInProgress
	case p.Next:
		status = domain.StatusNextUp
	}
	// A task born next-up/in-progress is active, and lint requires a description
	// there — so `new --next`/`--start` must carry one, for the same reason tags
	// are required: `new` must not scaffold a file its own linter rejects.
	if (status == domain.StatusNextUp || status == domain.StatusInProgress) && strings.TrimSpace(p.Description) == "" {
		return domain.Task{}, fmt.Errorf("%w: --description is required for a next-up/in-progress task (--next/--start)", domain.ErrValidation)
	}
	t := domain.Task{
		Slug:        slug,
		Status:      status,
		Epic:        p.Epic,
		Description: p.Description,
		Effort:      p.Effort,
		Tier:        p.Tier,
		Priority:    p.Priority,
		Autonomy:    p.Autonomy,
		Tags:        p.Tags,
		Created:     time.Now().Format("2006-01-02"),
	}
	// A task born in-progress (`new --start`) gets the same started_at stamp Move
	// writes, so "every in-progress task has a started_at" holds however it got there.
	if status == domain.StatusInProgress {
		t.StartedAt = t.Created
	}
	body := p.Body
	if body == "" {
		body = fmt.Sprintf(taskBodyTemplate, p.Title, p.Epic)
	}
	return s.store.CreateTask(t, body, p.DryRun)
}

// NewEpicParams are the inputs for creating an epic.
type NewEpicParams struct {
	Title       string
	Description string
	Status      string
	Priority    string
	Tags        []string
	Body        string
	DryRun      bool // validate + report the would-be epic without writing
}

const epicBodyTemplate = `
# %s

**Goal.** %s

## Why this is its own epic

<one paragraph: what makes this its own epic vs folding into a sibling?>

## Out of scope

- <explicitly excluded>
`

// NewEpic validates and creates an epic (auto-numbered NN-<slug>). Description
// is required (single line, ≤ the description cap); priority is validated.
func (s *Service) NewEpic(p NewEpicParams) (domain.Epic, error) {
	if strings.TrimSpace(p.Description) == "" {
		return domain.Epic{}, fmt.Errorf("%w: epic description is required", domain.ErrValidation)
	}
	if err := domain.ValidateDescription(p.Description); err != nil {
		return domain.Epic{}, err
	}
	if err := domain.ValidatePriority(p.Priority); err != nil {
		return domain.Epic{}, err
	}
	if err := domain.ValidateEpicStatus(p.Status); err != nil {
		return domain.Epic{}, err
	}
	if err := domain.ValidateTitle(p.Title); err != nil {
		return domain.Epic{}, err
	}
	slug := domain.Slugify(p.Title)
	if slug == "" {
		return domain.Epic{}, fmt.Errorf("%w: title produced an empty slug: %q", domain.ErrValidation, p.Title)
	}
	e := domain.Epic{
		Status:      p.Status,
		Description: p.Description,
		Priority:    p.Priority,
		Tags:        p.Tags,
		Created:     time.Now().Format("2006-01-02"),
	}
	body := p.Body
	if body == "" {
		body = fmt.Sprintf(epicBodyTemplate, p.Title, p.Description)
	}
	return s.store.CreateEpic(slug, e, body, p.DryRun)
}

func epicExists(epics []domain.Epic, id string) bool {
	for _, e := range epics {
		if e.ID == id {
			return true
		}
	}
	return false
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
	return rollupEpics(epics, tasks), append(ep1, ep2...), nil
}

// rollupEpics joins tasks onto their epics (by the tasks' `epic:` field) to
// produce per-epic done/total counts. Shared by ListEpics and Summary.
func rollupEpics(epics []domain.Epic, tasks []domain.Task) []EpicSummary {
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
	return out
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

// NewAuditParams are the inputs for creating an audit. Date defaults to today
// when empty; the audit is always created in the open bucket.
type NewAuditParams struct {
	Area   string
	Date   string // YYYY-MM-DD; empty → today
	Body   string // override the default scaffold
	DryRun bool   // validate + report the would-be audit without writing
}

// auditBodyTemplate is the default audit scaffold. The finding example is fenced
// so a fresh audit counts zero findings until real ones are added (parseAudit
// excludes fenced blocks). It stays generic — no repo-specific conventions-doc
// link (a repo that has one, e.g. desirelines-planning's HOWTO-execute, points
// at it from its own tooling, not from the shared tool's scaffold).
const auditBodyTemplate = "\n# Audit: %s — %s\n\n" +
	"> Edit findings in place and flip each `**Status:**` as you work it.\n\n" +
	"## Findings\n\n" +
	"<!-- One finding per issue, in this shape (un-fence it): -->\n\n" +
	"```\n" +
	"#### H1. <title>  · **Status:** open\n\n" +
	"**File:** <path:line> | **Component:** <component>\n" +
	"**Effort:** <XS|S|M|L> · **Urgency:** <acute|soon|eventually>\n\n" +
	"<what's wrong, why it matters, evidence>\n\n" +
	"**Recommendation:** <minimum fix>\n" +
	"```\n\n" +
	"## Candidate tasks\n\n" +
	"<!-- Mirror each finding: ✅ done · ⚠️ partial · ⏳ open · ⛔ won't do -->\n\n" +
	"- ⏳ `tskflwctl task new \"<title>\" --epic <id> --tags <tag>` — <one line>\n"

// NewAudit validates and creates an audit in the open bucket, returning it. The
// area must produce a non-empty slug and the date must be YYYY-MM-DD (today when
// omitted); the slug is `<date>-<area-slug>`. On invalid input it returns
// ErrValidation and nothing is written.
func (s *Service) NewAudit(p NewAuditParams) (domain.Audit, error) {
	area := strings.TrimSpace(p.Area)
	if area == "" {
		return domain.Audit{}, fmt.Errorf("%w: audit area is required", domain.ErrValidation)
	}
	if err := domain.ValidateTitle(area); err != nil {
		return domain.Audit{}, err
	}
	date := p.Date
	if date == "" {
		date = time.Now().Format("2006-01-02")
	}
	if err := domain.ValidateDate(date); err != nil {
		return domain.Audit{}, err
	}
	areaSlug := domain.Slugify(area)
	if areaSlug == "" {
		return domain.Audit{}, fmt.Errorf("%w: area produced an empty slug: %q", domain.ErrValidation, area)
	}
	a := domain.Audit{
		Slug:   date + "-" + areaSlug,
		Bucket: domain.AuditOpen,
		Area:   area,
		Date:   date,
	}
	body := p.Body
	if body == "" {
		body = fmt.Sprintf(auditBodyTemplate, area, date)
	}
	return s.store.CreateAudit(a, body, p.DryRun)
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
func (s *Service) MoveAudit(slug string, to domain.AuditBucket, dryRun bool) (domain.Audit, error) {
	return s.store.MoveAudit(slug, to, dryRun)
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
