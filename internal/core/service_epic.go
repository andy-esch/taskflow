package core

import (
	"fmt"
	"sort"
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

// Liveness is an epic's DERIVED activity band — read from its task rollup, never
// stored. It refines the `active` status (a live domain bucket) into how busy that
// bucket is right now, so a surface can foreground live work and recede quiet
// domains without anyone hand-maintaining a status. Only meaningful for active
// epics: a terminal (retired/deprecated) epic also computes a band, but callers
// gate on status first (see dashboardEpics / the epics-tab view filter).
type Liveness string

const (
	LivenessWorking Liveness = "working" // has open (pending/in-progress) tasks
	LivenessFresh   Liveness = "fresh"   // declared bucket with no tasks filed yet
	LivenessDormant Liveness = "dormant" // had tasks, all now done/withdrawn — quiet
)

// Open is the count of an epic's not-yet-done tasks. Total already excludes
// withdrawn/deprecated tasks (see TaskRollup), so Open is the real pending workload;
// 0 means nothing is in flight.
func (e EpicSummary) Open() int { return e.Total - e.Done }

// Liveness derives the activity band from the rollup (see the Liveness type):
// working = open tasks remain; fresh = no tasks filed yet (a new bucket, kept
// visible); dormant = had work, all of it now done/withdrawn. The Total>0 guard is
// what separates a *drained* epic (dormant) from a brand-*new* one (fresh) — both
// have zero open tasks, but only the new one should read as live.
func (e EpicSummary) Liveness() Liveness {
	switch {
	case e.Open() > 0:
		return LivenessWorking
	case e.Total == 0:
		return LivenessFresh
	default:
		return LivenessDormant
	}
}

// Live reports whether the epic should read as foreground (working or fresh) rather
// than a receded dormant bucket. The dashboard and the epics list both lead with
// Live epics and dim the rest.
func (e EpicSummary) Live() bool { return e.Liveness() != LivenessDormant }

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
		out[i] = rollupEpic(e, byEpic[e.ID])
	}
	return out
}

// rollupEpic builds one epic's summary from the tasks that belong to it — the
// single place an EpicSummary is assembled (TaskRollup called once), so the
// done/total/percent rule can't drift across the list/status path (rollupEpics)
// and the show/detail path (ShowEpic, audit M3).
func rollupEpic(e domain.Epic, its []domain.Task) EpicSummary {
	done, total, deprecated := TaskRollup(its)
	return EpicSummary{
		Epic: e, Total: total, Done: done, Deprecated: deprecated,
		LastUpdated: epicLastUpdated(its),
	}
}

// epicsByRecent returns the rollups ordered most-recently-updated first — the
// dashboard's "what moved lately" lens. Stable, so equal dates keep rollupEpics'
// (store) order; undated epics ("" LastUpdated) sink to the bottom. Summary applies
// it so every dashboard surface reads epics in ONE order (audit M2); the order
// lives here, in the aggregate, rather than being re-derived per adapter.
func epicsByRecent(epics []EpicSummary) []EpicSummary {
	out := append([]EpicSummary(nil), epics...)
	sort.SliceStable(out, func(i, j int) bool { return out[i].LastUpdated > out[j].LastUpdated })
	return out
}

// dashboardEpics is the landing-screen epic lens: the live buckets, ordered live
// work first and then by recency, so a cap'd dashboard fills with epics that have
// momentum and dormant domains sink to the bottom. Visibility fails OPEN — it hides
// only the known TERMINAL statuses (retired/deprecated via IsEpicArchived), so an
// unknown/foreign/missing status (e.g. an epic ported from a repo that uses
// planning/in-progress/completed) still shows and gets flagged elsewhere, rather
// than silently vanishing. Layered on epicsByRecent: recency orders within each
// liveness band, then a stable partition floats Live epics above dormant ones. The
// epics TAB keeps the full roster (its own view axis reaches retired/deprecated);
// this narrowing is dashboard-only, so `status` and the TUI dashboard still agree.
func dashboardEpics(epics []EpicSummary) []EpicSummary {
	live := make([]EpicSummary, 0, len(epics))
	for _, e := range epics {
		if !domain.IsEpicArchived(e.Epic.Status) {
			live = append(live, e)
		}
	}
	out := epicsByRecent(live)
	sort.SliceStable(out, func(i, j int) bool { return out[i].Live() && !out[j].Live() })
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

// ShowEpic returns an epic's rollup summary (so show/detail consume
// EpicSummary.Percent()/Done/Total instead of re-deriving the rule — audit M3),
// the tasks that belong to it, and its body.
func (s *Service) ShowEpic(id string) (EpicSummary, []domain.Task, string, error) {
	epic, body, err := s.store.GetEpic(id)
	if err != nil {
		return EpicSummary{}, nil, "", err
	}
	tasks, _, err := s.store.ListTasks()
	if err != nil {
		return EpicSummary{}, nil, "", err
	}
	var its []domain.Task
	for _, t := range tasks {
		if t.Epic == id {
			its = append(its, t)
		}
	}
	return rollupEpic(epic, its), its, body, nil
}
