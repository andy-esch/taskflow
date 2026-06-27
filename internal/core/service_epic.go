package core

import (
	"fmt"
	"strings"

	"github.com/andy-esch/taskflow/internal/domain"
)

// NewEpicParams are the inputs for creating an epic.
type NewEpicParams struct {
	Title       string
	Description string
	Status      string
	Priority    string
	Tags        []string
	Body        string // override the scaffold entirely (mutually exclusive with Template)
	Template    string // name of the body scaffold to use; empty = the kind's default
	DryRun      bool   // validate + report the would-be epic without writing
}

// NewEpic validates and creates an epic (auto-numbered NN-<slug>). Description
// is required (single line, ≤ the description cap); priority is validated.
func (s *Service) NewEpic(p NewEpicParams) (domain.Epic, error) {
	if err := templateBodyConflict(p.Body, p.Template); err != nil {
		return domain.Epic{}, err
	}
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
	// Any title is accepted: Slugify derives a filesystem-safe id while the full
	// original title is preserved in the body H1. The empty-slug error below is the
	// only hard guard — a title that slugifies to nothing.
	slug := domain.Slugify(p.Title)
	if slug == "" {
		return domain.Epic{}, fmt.Errorf("%w: title produced an empty slug: %q", domain.ErrValidation, p.Title)
	}
	e := domain.Epic{
		Status:      p.Status,
		Description: p.Description,
		Priority:    p.Priority,
		Tags:        p.Tags,
		Created:     s.now().Format("2006-01-02"),
	}
	body := p.Body
	if body == "" {
		tmpl, err := s.templateBody("epic", p.Template)
		if err != nil {
			return domain.Epic{}, err
		}
		body = renderTemplate(tmpl, map[string]string{"title": p.Title, "description": p.Description})
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

// EpicSummary is an epic plus its task rollup. Total/Done/Percent exclude
// deprecated (withdrawn) tasks; Deprecated counts them separately so a consumer
// can still surface "N withdrawn" without dragging the percentage.
type EpicSummary struct {
	Epic       domain.Epic
	Total      int
	Done       int
	Deprecated int
	// LastUpdated is the most recent task activity in the epic (max updated_at,
	// falling back to created) — epics carry no timestamp of their own, so "last
	// touched" is derived from the work that belongs to them. "" when undated.
	LastUpdated string
}

// Percent is the completed share of the epic's NON-deprecated tasks.
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
	byEpic := make(map[string][]domain.Task, len(epics))
	for _, t := range tasks {
		byEpic[t.Epic] = append(byEpic[t.Epic], t)
	}
	out := make([]EpicSummary, len(epics))
	for i, e := range epics {
		done, total, deprecated := TaskRollup(byEpic[e.ID])
		out[i] = EpicSummary{
			Epic: e, Total: total, Done: done, Deprecated: deprecated,
			LastUpdated: epicLastUpdated(byEpic[e.ID]),
		}
	}
	return out
}

// epicLastUpdated is the most recent task activity in an epic — the max of its
// tasks' updated_at (falling back to created when a task was never updated).
// Dates are YYYY-MM-DD, which sorts correctly as a string. "" when the epic has
// no dated tasks.
func epicLastUpdated(tasks []domain.Task) string {
	last := ""
	for _, t := range tasks {
		d := t.Updated
		if d == "" {
			d = t.Created
		}
		if d > last {
			last = d
		}
	}
	return last
}

// TaskRollup counts a task set for an epic-style progress rollup, in ONE place so
// the rule can't drift across the surfaces that draw it (epic list/status via
// rollupEpics, plus epic show and the TUI epic detail). Deprecated tasks are
// WITHDRAWN work — neither done nor pending — so they leave the denominator
// entirely (tracked separately); deferred ("not now") stays in total as real,
// eventually-do work; completed is done.
func TaskRollup(tasks []domain.Task) (done, total, deprecated int) {
	for _, t := range tasks {
		if t.Status == domain.StatusDeprecated {
			deprecated++
			continue
		}
		total++
		if t.Status == domain.StatusCompleted {
			done++
		}
	}
	return done, total, deprecated
}

// MoveEpic transitions an epic to another status (active/retired/deprecated).
// Epic status is a frontmatter field, not a directory, so this rewrites the field
// in place — the file is never relocated. The status is validated against the
// closed epic-status vocabulary; an unknown id is ErrNotFound, a bad status
// ErrValidation. Mirrors MoveAudit (resolve → validate target → surgical write).
func (s *Service) MoveEpic(id, status string, dryRun bool) (domain.Epic, error) {
	if err := domain.ValidateEpicStatus(status); err != nil {
		return domain.Epic{}, err
	}
	return s.store.MoveEpic(id, status, dryRun)
}

// SetEpicFields validates and applies non-status frontmatter updates to an epic
// in a single atomic write — the epic analog of SetFields. Status is moved via
// MoveEpic (`epic move`), not here; an attempt to set or unset it is rejected.
//
// Values arriving as strings from the `--set key=value` escape hatch are coerced
// to the native type a known typed field needs (only `tags` is a list for epics)
// before the store serializes them — otherwise the store would write a corrupting
// !!str that the strict loader can't read back. Keys outside the epic registry are
// rejected unless force is set (a typo'd field name must not silently persist). A
// domain.UnsetField value removes the key. Unlike SetFields, no updated_at is
// stamped — the Epic schema has no updated_at field and MoveEpic doesn't stamp
// one either, so the two epic write paths stay consistent.
func (s *Service) SetEpicFields(id string, updates map[string]any, force, dryRun bool) (domain.Epic, error) {
	if len(updates) == 0 {
		return domain.Epic{}, fmt.Errorf("%w: no fields given", domain.ErrValidation)
	}
	clean := make(map[string]any, len(updates))
	for field, val := range updates {
		if field == "status" {
			return domain.Epic{}, fmt.Errorf("%w: epic status moves via `epic move`, not `set`", domain.ErrValidation)
		}
		if _, unset := val.(domain.UnsetField); unset {
			// A typo'd field name must not silently no-op — gate unset on the registry
			// too, mirroring the set path (and the task SetFields contract).
			if !force && !domain.KnownEpicField(field) {
				return domain.Epic{}, unknownEpicFieldErr(field)
			}
			clean[field] = val
			continue
		}
		if !force && !domain.KnownEpicField(field) {
			return domain.Epic{}, unknownEpicFieldErr(field)
		}
		coerced, err := coerceEpicField(field, val)
		if err != nil {
			return domain.Epic{}, err
		}
		if err := domain.ValidateEpicField(field, stringify(coerced)); err != nil {
			return domain.Epic{}, err
		}
		clean[field] = coerced
	}
	return s.store.SetEpicFields(id, clean, dryRun)
}

// unknownEpicFieldErr is the shared rejection for an epic field outside the
// registry, used by both the set and unset paths of SetEpicFields (the epic
// counterpart to unknownFieldErr).
func unknownEpicFieldErr(field string) error {
	return fmt.Errorf(
		"%w: unknown epic field %q (known fields only; use --force for a custom field)", domain.ErrValidation, field)
}

// coerceEpicField converts a string `--set` value into the native type its epic
// field needs (only `tags` is a list). Values already of the right type (from
// typed flags) and genuinely-custom fields pass through unchanged.
func coerceEpicField(field string, val any) (any, error) {
	str, isStr := val.(string)
	if !isStr {
		return val, nil // a typed flag already supplied the native type
	}
	if domain.IsEpicListField(field) {
		return splitList(str), nil
	}
	return str, nil
}

// EditEpic opens an epic for whole-file editing — the human face of mutation,
// complementing the field-level `epic set` (the epic counterpart to EditTask).
// edit (run by the cli's $EDITOR layer) receives the current file content and
// returns the new content; the store accepts it only if it still parses as an
// epic, reopening the editor on a broken edit. Returns the reloaded epic and
// whether anything changed.
func (s *Service) EditEpic(id string, edit func(current string, prevErr error) (string, error)) (domain.Epic, bool, error) {
	return s.store.EditEpic(id, edit)
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
