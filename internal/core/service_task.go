package core

import (
	"fmt"
	"strconv"
	"strings"
	"time"

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
		if f.RevisitDue && (t.Status != domain.StatusDeferred || !domain.IsRevisitDue(t.RevisitAt, now)) {
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

// ReplaceBody overwrites a task's markdown body in one atomic, validated write —
// the agent face of body editing (`task set --body`), beside the human EditTask.
// Frontmatter is preserved surgically and updated_at is stamped. Returns the
// reloaded task and the resulting body.
func (s *Service) ReplaceBody(slug, body string, dryRun bool) (domain.Task, string, error) {
	return s.store.EditBody(slug, body, false, time.Now(), dryRun)
}

// AppendBody appends a section to a task's markdown body (`task append`),
// separated by a blank line, in one atomic, validated write. Returns the reloaded
// task and the resulting body.
func (s *Service) AppendBody(slug, text string, dryRun bool) (domain.Task, string, error) {
	return s.store.EditBody(slug, text, true, time.Now(), dryRun)
}

// Move transitions a task to the given status (lifecycle engine behind the
// explicit verbs). Moving to the current status is an idempotent no-op.
// dryRun validates everything and returns the would-be task without writing.
func (s *Service) Move(slug string, to domain.Status, dryRun bool) (domain.Task, error) {
	return s.store.Move(slug, to, time.Now(), dryRun)
}

// DeferTask moves a task to deferred and, when until is non-empty, records it as
// the revisit_at ("snooze until") date — the two halves of `task defer --until`.
// The date is set through the same validated, surgical SetFields path `task set`
// uses, so it can't write a form the linter rejects; the caller validates the
// date up front. A bare defer (empty until) is exactly Move(StatusDeferred).
// dryRun previews the move without writing; the field would be set on the real
// run, so the previewed task carries the would-be revisit_at for the report.
func (s *Service) DeferTask(slug, until string, dryRun bool) (domain.Task, error) {
	t, err := s.Move(slug, domain.StatusDeferred, dryRun)
	if err != nil {
		return domain.Task{}, err
	}
	if until == "" {
		return t, nil
	}
	if dryRun {
		// Nothing was written, so the moved file isn't in deferred/ yet; reflect the
		// would-be field in the preview without touching disk.
		t.RevisitAt = until
		return t, nil
	}
	// The Move above already persisted (file relocated into deferred/). This is a
	// SECOND write, so the two halves aren't atomic: if SetFields fails here the
	// task IS deferred but carries no revisit_at. Name that partial state in the
	// error (keeping the sentinel via %w for the exit code) so the report doesn't
	// read as "nothing happened" — a re-run of `task defer <slug> --until <date>`
	// is an idempotent Move no-op then a clean SetFields, so retry recovers it.
	out, err := s.SetFields(slug, map[string]any{"revisit_at": until}, false, false)
	if err != nil {
		return domain.Task{}, fmt.Errorf("%q deferred but revisit date %q not recorded (retry `task defer %s --until %s`): %w", slug, until, slug, until, err)
	}
	return out, nil
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
	switch {
	case p.Start:
		status = domain.StatusInProgress
	case p.Next:
		status = domain.StatusNextUp
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
	// `new` must not scaffold a file its own linter rejects: every active task needs
	// tags, and a next-up/in-progress one needs a description. The same rule the
	// SetFields write path applies, defined once in the domain (decided 2026-06-12).
	if err := domain.ActiveTaskFieldErr(t); err != nil {
		return domain.Task{}, err
	}
	// A task born in-progress (`new --start`) gets the same started_at stamp Move
	// writes, so "every in-progress task has a started_at" holds however it got there.
	if status == domain.StatusInProgress {
		t.StartedAt = t.Created
	}
	body := p.Body
	if body == "" {
		tmpl, err := s.templateBody("task", p.Template)
		if err != nil {
			return domain.Task{}, err
		}
		body = renderTemplate(tmpl, map[string]string{"title": p.Title, "epic": p.Epic})
	}
	return s.store.CreateTask(t, body, p.DryRun)
}
