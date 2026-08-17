package core

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/andy-esch/taskflow/internal/domain"
)

// NewResearchParams are the inputs for creating a research doc. Created defaults to
// today when empty. There is no status/bucket to choose (research has no lifecycle)
// and no epic to attach (provenance is a body concern) — so the input set is just
// "what is it, when, and how do I find it again".
type NewResearchParams struct {
	Title       string
	Created     string // YYYY-MM-DD; empty → today
	Description string
	Tags        []string
	Body        string // override the scaffold entirely (mutually exclusive with Template)
	Template    string // name of the body scaffold to use; empty = the kind's default
	DryRun      bool   // validate + report the would-be doc without writing
}

// NewResearch validates and creates a research doc, returning it. The title must
// produce a non-empty slug and created must be YYYY-MM-DD (today when omitted).
//
// The id is minted FROM created (not from "now"), so a doc backdated to when the work
// actually happened sorts into place chronologically — the property that lets the
// corpus be ordered by id (ADR-0003 §3). On invalid input it returns ErrValidation and
// nothing is written.
func (s *Service) NewResearch(p NewResearchParams) (domain.Research, error) {
	if err := templateBodyConflict(p.Body, p.Template); err != nil {
		return domain.Research{}, err
	}
	title := strings.TrimSpace(p.Title)
	if title == "" {
		return domain.Research{}, fmt.Errorf("%w: research title is required", domain.ErrValidation)
	}
	created := p.Created
	if created == "" {
		created = s.now().Format("2006-01-02")
	}
	if err := domain.ValidateDate(created); err != nil {
		return domain.Research{}, err
	}
	if err := domain.ValidateDescription(p.Description); err != nil {
		return domain.Research{}, err
	}
	// Any title is accepted: Slugify derives a filesystem-safe slug while the full
	// original title is preserved as the body H1. An empty slug is the only hard guard.
	slug := domain.Slugify(title)
	if slug == "" {
		return domain.Research{}, fmt.Errorf("%w: title produced an empty slug: %q", domain.ErrValidation, title)
	}
	// Mint from created, not now — a day-precision date, so same-day docs land in the
	// same millisecond slot and their relative order comes from NewAt's random low bits.
	// That intra-day order is deliberately not meaningful (epic 28, 2026-08-14).
	day, err := time.Parse("2006-01-02", created)
	if err != nil {
		return domain.Research{}, fmt.Errorf("%w: unparseable created date %q", domain.ErrValidation, created)
	}
	r := domain.Research{
		Slug:        slug,
		ID:          s.newIDAt(day.UnixMilli()),
		Created:     created,
		Description: strings.TrimSpace(p.Description),
		Tags:        p.Tags,
	}
	body := p.Body
	if body == "" {
		tmpl, err := s.templateBody("research", p.Template)
		if err != nil {
			return domain.Research{}, err
		}
		body = renderTemplate(tmpl, map[string]string{"title": title, "date": created})
	}
	return s.store.CreateResearch(r, body, p.DryRun)
}

// ListResearch returns every research doc, newest first, plus any per-file load
// problems. There is no status/bucket to filter on, so unlike ListTasks/ListAudits
// there is no default-view carve — the whole corpus is the listing. tag filters to one
// topical tag (case-insensitive); empty means no filter.
func (s *Service) ListResearch(tag string) ([]domain.Research, []domain.FileProblem, error) {
	docs, problems, err := s.store.ListResearch()
	if err != nil {
		return nil, nil, err
	}
	out := make([]domain.Research, 0, len(docs))
	for _, r := range docs {
		if tag != "" && !hasTag(r.Tags, tag) {
			continue
		}
		out = append(out, r)
	}
	// Newest first — the useful default for a corpus read chronologically. Ties break on
	// slug so the order is stable (same-day docs are common: dates are day-precision).
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Created != out[j].Created {
			return out[i].Created > out[j].Created
		}
		return out[i].Slug < out[j].Slug
	})
	return out, problems, nil
}

// ShowResearch returns one research doc plus its body.
func (s *Service) ShowResearch(slug string) (domain.Research, string, error) {
	return s.store.GetResearch(slug)
}

// ResearchPath resolves a research doc's file path without reading or parsing it —
// the seam for `research path` (parse-free, like TaskPath).
func (s *Service) ResearchPath(slug string) (string, error) {
	return s.store.ResolveResearchPath(slug)
}
