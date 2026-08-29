package core

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/andy-esch/taskflow/internal/domain"
)

// TaskFilter narrows a task listing. Zero-valued fields are ignored. When no
// explicit Status is given and All is false, only active tasks are returned.
type TaskFilter struct {
	Status     string
	Epic       string
	Tag        string
	All        bool
	RevisitDue bool // only deferred tasks whose revisit_at (snooze-until) date has arrived
	Unblocked  bool // only tasks whose strict graph state is derived Eligible
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
	var graph *TaskGraph
	if f.Unblocked {
		graph = NewTaskGraph(all, problems)
		if graph.Health() != GraphHealthy {
			return nil, problems, fmt.Errorf("%w: task list --unblocked requires a healthy repository task graph; health=%s: %s",
				domain.ErrValidation, graph.Health(), taskGraphHealthDetail(graph))
		}
	}
	// --revisit-due narrows to deferred tasks, so it opts out of the active-only
	// default (deferred is inactive) just like an explicit --status does.
	activeOnly := f.Status == "" && !f.All && !f.RevisitDue
	now := s.now()
	out := make([]domain.Task, 0, len(all))
	for _, t := range all {
		if activeOnly && !t.Status.IsActive() {
			continue
		}
		if f.Status != "" && string(t.Status) != f.Status {
			continue
		}
		// "Up for revisit" = parked in deferred AND the snooze date has arrived;
		// implies the deferred scope and composes with --epic/--tag below.
		if f.RevisitDue && !domain.IsTaskRevisitDue(t, now) {
			continue
		}
		// Match on the epic NN key, not the raw string — so `--epic 24-data-model` finds a
		// task whose ref is a bare `24` or a stale slug, mirroring the rollup/validation join.
		if f.Epic != "" && domain.EpicRefKey(t.Epic) != domain.EpicRefKey(f.Epic) {
			continue
		}
		if f.Tag != "" && !hasTag(t.Tags, f.Tag) {
			continue
		}
		if f.Unblocked && !graph.State(t.ID).Eligible {
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

// LoadTaskGraph is the one canonical filesystem-agnostic snapshot loader used by
// the guarded mutation boundary. Diagnostic consumers that already own task bodies
// (notably lint) construct the same strict projection with NewTaskGraph rather than
// performing a second repository scan.
type TaskGraphSource interface {
	ListTasks() ([]domain.Task, []domain.FileProblem, error)
}

func LoadTaskGraph(source TaskGraphSource) (*TaskGraph, error) {
	tasks, problems, err := source.ListTasks()
	if err != nil {
		return nil, err
	}
	return NewTaskGraph(tasks, problems), nil
}

// TaskPath resolves a task's file path without reading or parsing it — the seam
// for `task path`, which must work even on a file with broken frontmatter.
func (s *Service) TaskPath(slug string) (string, error) {
	return s.store.ResolveTaskPath(slug)
}

// AcceptanceCriteria lists a task's acceptance criteria (read-only, for `task ac
// --list`). Returns the canonical slug (for the envelope) and the ordered criteria.
func (s *Service) AcceptanceCriteria(slug string) (string, []domain.Criterion, error) {
	t, body, err := s.store.GetTask(slug)
	if err != nil {
		return "", nil, err
	}
	return t.Slug, domain.ListAcceptanceCriteria(body), nil
}

// SetAcceptanceCriterion flips a task's nth (1-based) acceptance-criteria checkbox and
// writes the result through the atomic, frontmatter-preserving body-replace path
// (EditBody). Returns the reloaded task, the resulting body, and whether anything
// changed — false means the criterion was already in the target state, so no write was
// performed (and updated_at is not bumped).
func (s *Service) SetAcceptanceCriterion(slug string, n int, checked, dryRun bool) (domain.Task, string, bool, error) {
	t, body, err := s.store.GetTask(slug)
	if err != nil {
		return domain.Task{}, "", false, err
	}
	newBody, err := domain.SetAcceptanceCriterion(body, n, checked)
	if err != nil {
		return domain.Task{}, "", false, err
	}
	if newBody == body {
		return t, body, false, nil // already in the target state — no write
	}
	rt, rb, err := s.store.EditBody(slug, newBody, false, s.now(), dryRun)
	return rt, rb, true, err
}

// SetCriterionState sets a task's nth (1-based) criterion to an explicit state, through the
// same atomic, frontmatter-preserving body-replace path the checkbox flip uses. It is the
// write path the criterion vocabulary shipped WITH, rather than after: a state reachable
// only by hand-editing is one nobody can be held to, which is how the finding vocabulary
// drifted from its own documentation.
func (s *Service) SetCriterionState(slug string, n int, state domain.CriterionState, reason string, dryRun bool) (domain.Task, string, bool, error) {
	t, body, err := s.store.GetTask(slug)
	if err != nil {
		return domain.Task{}, "", false, err
	}
	newBody, err := domain.SetCriterionState(body, n, state, reason)
	if err != nil {
		return domain.Task{}, "", false, err
	}
	if newBody == body {
		return t, body, false, nil // already in the target state — no write, no updated_at bump
	}
	rt, rb, err := s.store.EditBody(slug, newBody, false, s.now(), dryRun)
	return rt, rb, true, err
}

// EditCriteria adds, removes, or rewords one acceptance criterion through the same atomic,
// frontmatter-preserving body-replace path SetCriterionState uses. edit is the domain
// operation; naming it here rather than adding three near-identical methods keeps the write
// path (read → transform → atomic replace, no write when nothing changed) in one place.
func (s *Service) EditCriteria(slug string, dryRun bool, edit func(body string) (string, error)) (domain.Task, string, bool, error) {
	t, body, err := s.store.GetTask(slug)
	if err != nil {
		return domain.Task{}, "", false, err
	}
	newBody, err := edit(body)
	if err != nil {
		return domain.Task{}, "", false, err
	}
	if newBody == body {
		return t, body, false, nil
	}
	rt, rb, err := s.store.EditBody(slug, newBody, false, s.now(), dryRun)
	return rt, rb, true, err
}

// EditTask opens a task for whole-file editing — the human face of mutation,
// complementing the field-level `task set`. edit (run by the cli's $EDITOR layer)
// receives the current file content and returns the new content; the store
// accepts it only if it still parses as a task, reopening the editor on a broken
// edit. Returns the reloaded task and whether anything changed.
func (s *Service) EditTask(slug string, edit func(current string, prevErr error) (string, error)) (domain.Task, bool, error) {
	return s.store.EditTask(slug, s.now(), edit)
}

// ReplaceBody overwrites a task's markdown body in one atomic, validated write —
// the agent face of body editing (`task set --body`), beside the human EditTask.
// Frontmatter is preserved surgically and updated_at is stamped. Returns the
// reloaded task and the resulting body.
func (s *Service) ReplaceBody(slug, body string, dryRun bool) (domain.Task, string, error) {
	return s.store.EditBody(slug, body, false, s.now(), dryRun)
}

// AppendBody appends a section to a task's markdown body (`task append`),
// separated by a blank line, in one atomic, validated write. Returns the reloaded
// task and the resulting body.
func (s *Service) AppendBody(slug, text string, dryRun bool) (domain.Task, string, error) {
	now := s.now()
	type res struct {
		task domain.Task
		body string
	}
	r, err := retryOnConflict(s, dryRun, func() (res, error) {
		t, b, e := s.store.EditBody(slug, text, true, now, dryRun)
		return res{t, b}, e
	})
	return r.task, r.body, err
}

// Move transitions a task through the guarded lifecycle capability. Override is
// typed so dependency eligibility and completion criteria can never be confused
// behind one internal force boolean.
func (s *Service) Move(ref string, to domain.Status, dryRun bool, override TaskLifecycleOverride) (TaskLifecycleReceipt, error) {
	return s.runTaskLifecycleMutation(dryRun, func(graph *TaskGraph) (TaskLifecyclePlan, error) {
		taskID, err := graph.ResolveTaskID(ref)
		if err != nil {
			return TaskLifecyclePlan{}, err
		}
		return TaskLifecyclePlan{TaskID: taskID, To: to, Override: override}, nil
	})
}

// DeferTask moves a task to deferred and, when until is non-empty, records it as
// the revisit_at ("snooze until") date — the two halves of `task defer --until`,
// written together in ONE atomic store operation (audit M4). A bare defer (empty
// until) is exactly Move(StatusDeferred). dryRun previews the move (the store
// reflects the would-be revisit_at on the returned task) without writing.
//
// The date is validated here so the contract holds for every adapter — the same
// guard the old SetFields path applied, kept now that the write bypasses SetFields.
func (s *Service) DeferTask(ref, until string, dryRun bool) (TaskLifecycleReceipt, error) {
	if until != "" {
		if err := domain.ValidateDate(until); err != nil {
			return TaskLifecycleReceipt{}, err
		}
	}
	return s.runTaskLifecycleMutation(dryRun, func(graph *TaskGraph) (TaskLifecyclePlan, error) {
		taskID, err := graph.ResolveTaskID(ref)
		if err != nil {
			return TaskLifecyclePlan{}, err
		}
		return TaskLifecyclePlan{TaskID: taskID, To: domain.StatusDeferred, RevisitAt: until}, nil
	})
}

func (s *Service) runTaskLifecycleMutation(dryRun bool, planner TaskLifecyclePlanner) (TaskLifecycleReceipt, error) {
	if s.lifecycleMutations == nil {
		return TaskLifecycleReceipt{}, fmt.Errorf("task lifecycle mutations are unavailable from this store")
	}
	now := s.now()
	result, err := s.lifecycleMutations.MutateTaskLifecycle(now, dryRun, planner)
	if !dryRun {
		for attempt := 1; attempt <= s.maxRetries && errors.Is(err, domain.ErrConflict) && !result.Committed; attempt++ {
			s.retrySleep(attempt)
			result, err = s.lifecycleMutations.MutateTaskLifecycle(now, dryRun, planner)
		}
	}
	receipt := taskLifecycleReceipt(result)
	if err != nil && result.Committed {
		return receipt, &TaskLifecycleMutationFailure{Cause: err, Receipt: receipt}
	}
	return receipt, err
}

func taskLifecycleReceipt(result TaskLifecycleMutationResult) TaskLifecycleReceipt {
	return TaskLifecycleReceipt{
		Task: result.Task, From: result.From, To: result.Plan.To,
		Changed: result.Changed, DryRun: result.DryRun, Committed: result.Committed, Override: result.Plan.Override,
		Forced: result.OverrideApplied, Before: result.Before, After: result.After,
		OutstandingBlockers: cloneLifecycleBlockers(result.OutstandingBlockers),
		Impacts:             cloneTaskGraphStateImpacts(result.Impacts),
		Remedy:              taskLifecycleRemedy(result),
	}
}

func taskLifecycleRemedy(result TaskLifecycleMutationResult) string {
	parts := make([]string, 0, 2)
	if result.OverrideApplied && result.Plan.Override == TaskLifecycleOverrideDependencyGate {
		parts = append(parts, "resolve the outstanding blockers; the override did not alter dependency edges")
	}
	if len(result.Impacts) > 0 {
		parts = append(parts, "inspect each affected task with `tskflwctl task blockers <task>` and restore sound prerequisites or update its dependencies")
	}
	return strings.Join(parts, "; ")
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
		// A dependency change is a repository-global graph mutation, not ordinary
		// frontmatter surgery. Reject it before the known/custom-field and --force
		// branches so neither spelling can bypass cycle/referential validation. The
		// guarded commands land in the next production slice.
		if domain.IsGraphOwnedTaskField(field) {
			return domain.Task{}, fmt.Errorf(
				"%w: %s is graph-owned and cannot be changed with `task set` (including --force); use guarded dependency operations%s",
				domain.ErrValidation, field, graphFieldDirection(field))
		}
		// Status is lifecycle-owned. Route it through the guarded lifecycle capability
		// so eligibility, timestamps, completion checks, and impact receipts cannot be
		// bypassed. Rejected on both the set and unset paths.
		if field == "status" {
			return domain.Task{}, fmt.Errorf("%w: status is lifecycle-owned — use `task <verb>`/`task move`, not `set`", domain.ErrValidation)
		}
		if _, unset := val.(domain.UnsetField); unset {
			switch field {
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
			// Store the epic's canonical stem (as NewTask does), so `set --set epic=24` and
			// `new --epic 24` leave the same readable `<NN>-<slug>` ref, not a bare NN.
			withMeta["epic"] = canonicalEpic(epics, epic)
		}
	}
	withMeta["updated_at"] = s.now().Format("2006-01-02")
	return retryOnConflict(s, dryRun, func() (domain.Task, error) {
		return s.store.SetFields(slug, withMeta, dryRun)
	})
}

func graphFieldDirection(field string) string {
	if field == "depends_on" {
		return "; use `tskflwctl task depend add` or `tskflwctl task depend remove`"
	}
	return "; remove legacy dependency fields with `tskflwctl task depend migrate`"
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
	case domain.IsIntField(field):
		n, err := strconv.Atoi(strings.TrimSpace(str))
		if err != nil {
			return nil, fmt.Errorf("%w: %s must be an integer, got %q", domain.ErrValidation, field, str)
		}
		return n, nil
	case domain.IsListField(field):
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
	Body        string // override the scaffold entirely (mutually exclusive with Template)
	Template    string // name of the body scaffold to use; empty = the kind's default
	DryRun      bool   // validate + report the would-be task without writing
}

// NewTask validates and creates a task, returning the created task. The epic
// must exist; tier/autonomy/priority/description are validated. On any invalid
// input it returns ErrValidation and nothing is written.
func (s *Service) NewTask(p NewTaskParams) (domain.Task, error) {
	if err := templateBodyConflict(p.Body, p.Template); err != nil {
		return domain.Task{}, err
	}
	epics, _, err := s.store.ListEpics()
	if err != nil {
		return domain.Task{}, err
	}
	if !epicExists(epics, p.Epic) {
		return domain.Task{}, fmt.Errorf("%w: unknown epic %q", domain.ErrValidation, p.Epic)
	}
	// Store and link the epic's canonical stem (resolved on the NN key), so a bare NN or
	// a stale slug the caller passed becomes the epic's current, readable `<NN>-<slug>`.
	p.Epic = canonicalEpic(epics, p.Epic)
	// Defaults for zero-valued fields live here so EVERY caller — not just the CLI
	// flags — produces a valid, lint-clean task; the CLI flag defaults stay as
	// help-text hints. (A second adapter calling NewTask without replicating them
	// would otherwise get ErrValidation or scaffold a lint-failing file.)
	if p.Priority == "" {
		p.Priority = "medium"
	}
	if p.Tier == 0 {
		p.Tier = 3
	}
	if p.Autonomy == 0 {
		p.Autonomy = 3
	}
	if p.Effort == "" {
		p.Effort = "Unknown"
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
	// Any title is accepted: Slugify derives a filesystem-safe id (it word-breaks
	// path separators, control chars, and the unicode punctuation it can't keep)
	// while the full original title is preserved in the body H1. The empty-slug
	// error below is the only hard guard — a title that slugifies to nothing.
	slug := domain.Slugify(p.Title)
	if slug == "" {
		return domain.Task{}, fmt.Errorf("%w: title produced an empty slug: %q", domain.ErrValidation, p.Title)
	}
	status := domain.StatusReadyToStart
	if p.Next {
		status = domain.StatusNextUp
	}
	t := domain.Task{
		Slug:        slug,
		ID:          s.newID(),
		Status:      status,
		Epic:        p.Epic,
		Description: p.Description,
		Effort:      p.Effort,
		Tier:        p.Tier,
		Priority:    p.Priority,
		Autonomy:    p.Autonomy,
		Tags:        p.Tags,
		Created:     s.now().Format("2006-01-02"),
	}
	// `new` must not scaffold a file its own linter rejects: every active task needs
	// tags, and a next-up/in-progress one needs a description. The same rule the
	// SetFields write path applies, defined once in the domain (decided 2026-06-12).
	if err := domain.ActiveTaskFieldErr(t); err != nil {
		return domain.Task{}, err
	}
	body := p.Body
	if body == "" {
		tmpl, err := s.templateBody("task", p.Template)
		if err != nil {
			return domain.Task{}, err
		}
		body = renderTemplate(tmpl, map[string]string{"title": p.Title, "epic": p.Epic})
	}
	if p.Start {
		receipt, err := s.runTaskLifecycleMutation(p.DryRun, func(*TaskGraph) (TaskLifecyclePlan, error) {
			return TaskLifecyclePlan{
				To:     domain.StatusInProgress,
				Create: &TaskLifecycleCreation{Task: t, Body: body},
			}, nil
		})
		return receipt.Task, err
	}
	return s.store.CreateTask(t, body, p.DryRun)
}

// RenameTask re-titles a task (new slug from newTitle, id kept) and cascades its inbound
// body links across the planning tree — see store.RenameTask. Returns the reloaded task
// and the count of inbound links repointed.
func (s *Service) RenameTask(slug, newTitle string, dryRun bool) (domain.Task, int, error) {
	return s.store.RenameTask(slug, newTitle, dryRun)
}
