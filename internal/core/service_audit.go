package core

import (
	"fmt"
	"strings"

	"github.com/andy-esch/taskflow/internal/domain"
)

// NewAuditParams are the inputs for creating an audit. Date defaults to today
// when empty; the audit is always created in the open bucket.
type NewAuditParams struct {
	Area     string
	Date     string // YYYY-MM-DD; empty → today
	Body     string // override the scaffold entirely (mutually exclusive with Template)
	Template string // name of the body scaffold to use; empty = the kind's default
	DryRun   bool   // validate + report the would-be audit without writing
}

// NewAudit validates and creates an audit in the open bucket, returning it. The
// area must produce a non-empty slug and the date must be YYYY-MM-DD (today when
// omitted); the slug is `<date>-<area-slug>`. On invalid input it returns
// ErrValidation and nothing is written.
func (s *Service) NewAudit(p NewAuditParams) (domain.Audit, error) {
	if err := templateBodyConflict(p.Body, p.Template); err != nil {
		return domain.Audit{}, err
	}
	area := strings.TrimSpace(p.Area)
	if area == "" {
		return domain.Audit{}, fmt.Errorf("%w: audit area is required", domain.ErrValidation)
	}
	// Any area is accepted: Slugify derives a filesystem-safe id while the full
	// original area is preserved (frontmatter + body). The empty-slug error below
	// is the only hard guard — an area that slugifies to nothing.
	date := p.Date
	if date == "" {
		date = s.now().Format("2006-01-02")
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
		tmpl, err := s.templateBody("audit", p.Template)
		if err != nil {
			return domain.Audit{}, err
		}
		body = renderTemplate(tmpl, map[string]string{"area": area, "date": date})
	}
	return s.store.CreateAudit(a, body, p.DryRun)
}

// ListAudits returns audits in the requested bucket (default: open), plus any
// per-file load problems. bucket="" + all=false means open only. An unknown
// bucket is validated up front and returns ErrValidation rather than a silently
// empty list, which agents routing on exit codes can't tell apart from an empty
// bucket — mirroring ListTasks' status check.
func (s *Service) ListAudits(bucket string, all bool) ([]domain.Audit, []domain.FileProblem, error) {
	if bucket != "" {
		if _, err := domain.ParseAuditBucket(bucket); err != nil {
			return nil, nil, err
		}
	}
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

// EditAudit opens an audit for whole-file editing — the human face of mutation,
// complementing the agent-facing `audit append` (the audit counterpart to EditTask).
// The store accepts the save only if it still parses as an audit; the caller surfaces
// finding-level lint (status vocab, bucket↔state) on the result. Returns the reloaded
// audit and whether anything changed.
func (s *Service) EditAudit(slug string, edit func(current string, prevErr error) (string, error)) (domain.Audit, bool, error) {
	return s.store.EditAudit(slug, s.now(), edit)
}

// AppendAuditBody appends a section to an audit's markdown body (`audit append`) in
// one atomic, validated write — the agent face of audit body editing, beside the
// human EditAudit. Stamps updated_at (the audit's `date` stays immutable — it's the
// slug). Returns the reloaded audit and the resulting body.
func (s *Service) AppendAuditBody(slug, text string, dryRun bool) (domain.Audit, string, error) {
	return s.store.AppendAuditBody(slug, text, s.now(), dryRun)
}
