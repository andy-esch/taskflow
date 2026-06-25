package core

import (
	"fmt"
	"strings"
	"time"

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
		Created:     time.Now().Format("2006-01-02"),
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
		out[i] = EpicSummary{Epic: e, Total: total, Done: done, Deprecated: deprecated}
	}
	return out
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
