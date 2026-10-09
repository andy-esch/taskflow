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
func (s *Service) NewAudit(p NewAuditParams) (AuditCreationReceipt, error) {
	if err := templateBodyConflict(p.Body, p.Template); err != nil {
		return AuditCreationReceipt{}, err
	}
	area := strings.TrimSpace(p.Area)
	if area == "" {
		return AuditCreationReceipt{}, fmt.Errorf("%w: audit area is required", domain.ErrValidation)
	}
	// Any area is accepted: Slugify derives a filesystem-safe id while the full
	// original area is preserved (frontmatter + body). The empty-slug error below
	// is the only hard guard — an area that slugifies to nothing.
	date := p.Date
	if date == "" {
		date = s.now().Format("2006-01-02")
	}
	if err := domain.ValidateDate(date); err != nil {
		return AuditCreationReceipt{}, err
	}
	areaSlug := domain.Slugify(area)
	if areaSlug == "" {
		return AuditCreationReceipt{}, fmt.Errorf("%w: area produced an empty slug: %q", domain.ErrValidation, area)
	}
	a := domain.Audit{
		Slug:   date + "-" + areaSlug,
		ID:     s.newID(),
		Bucket: domain.AuditOpen,
		Area:   area,
		Date:   date,
	}
	body := p.Body
	if body == "" {
		tmpl, err := s.templateBody("audit", p.Template)
		if err != nil {
			return AuditCreationReceipt{}, err
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
func (s *Service) ListAudits(bucket string, all bool) ([]LoadedRecord[domain.Audit], []LoadProblem, error) {
	if bucket != "" {
		if _, err := domain.ParseAuditBucket(bucket); err != nil {
			return nil, nil, err
		}
	}
	read, err := s.store.ReadAudits()
	if err != nil {
		return nil, nil, err
	}
	read.Records, read.Problems = loadedRecordsWithIDs(EntityAudit, read.Records, read.Problems,
		func(audit domain.Audit) string { return audit.Slug })
	out := make([]LoadedRecord[domain.Audit], 0, len(read.Records))
	for _, record := range read.Records {
		a := record.Value
		switch {
		case bucket != "":
			if string(a.Bucket) != bucket {
				continue
			}
		case !all && a.Bucket != domain.AuditOpen:
			continue
		}
		out = append(out, record)
	}
	return out, read.Problems, nil
}

// ShowAudit returns one audit plus its body.
func (s *Service) ShowAudit(slug string) (LoadedRecord[AuditWithBody], error) {
	record, err := s.store.ReadAudit(slug)
	if err != nil {
		return LoadedRecord[AuditWithBody]{}, err
	}
	if err := requireSourceID(EntityAudit, record.Source); err != nil {
		return LoadedRecord[AuditWithBody]{}, err
	}
	return record, nil
}

// AuditPath resolves an audit's file path without reading or parsing it — the seam
// for `audit path` (parse-free, like TaskPath).
func (s *Service) AuditPath(slug string) (string, error) {
	if s.auditPaths == nil {
		return "", fmt.Errorf("%w: audit path resolution is unavailable from this service", domain.ErrValidation)
	}
	path, err := s.auditPaths.ResolveAuditPath(slug)
	return requireResolvedLocalPath(EntityAudit, path, err)
}

// MoveAudit changes an audit's authoritative bucket (close/reopen/defer). The
// persistence port validates domain policy against its guarded source; a retry
// must reload counts so newly unparsed or unsettled evidence cannot be overwritten.
func (s *Service) MoveAudit(slug string, to domain.AuditBucket, dryRun bool) (domain.Audit, error) {
	return retryOnConflict(s, dryRun, func() (domain.Audit, error) {
		return s.store.MoveAudit(slug, to, dryRun)
	})
}

// EditAudit opens an audit for whole-file editing — the human face of mutation,
// complementing the agent-facing `audit append` (the audit counterpart to EditTask).
// The store accepts the save only if it still parses as an audit; the caller surfaces
// finding-level lint (status vocab, bucket↔state) on the result. Returns the reloaded
// audit and whether anything changed.
func (s *Service) EditAudit(slug string, edit func(current string, prevErr error) (string, error)) (domain.Audit, bool, error) {
	return s.store.EditAudit(slug, s.now(), edit)
}

// AppendAuditBody adds a section to an audit's narrative (`audit append`) in one
// atomic, validated write, preserving a trailing managed Candidate tasks section
// as the final projection. It is the agent face of audit body editing beside the
// human EditAudit. Stamps updated_at (the audit's date stays immutable) and returns
// the reloaded audit and resulting body.
func (s *Service) AppendAuditBody(slug, text string, dryRun bool) (domain.Audit, string, error) {
	now := s.now()
	type res struct {
		audit domain.Audit
		body  string
	}
	r, err := retryOnConflict(s, dryRun, func() (res, error) {
		a, b, e := s.store.AppendAuditBody(slug, text, now, dryRun)
		return res{a, b}, e
	})
	return r.audit, r.body, err
}
