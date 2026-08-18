package core

import (
	"errors"
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
	// ValidateMintableDate, not ValidateDate: the id is minted from this date, so it must
	// also be inside the range an id can encode — otherwise the timestamp wraps and the
	// doc sorts wrongly forever, with no other symptom.
	if err := domain.ValidateMintableDate(created); err != nil {
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
	// Mint, then REGENERATE on a collision with an id already on disk. Minting is keyed on
	// a day, so same-day docs all draw from one 2^17 random tail — id.NewAt's own doc
	// requires callers to dedupe, and the store reports a clash as ErrConflict. Bounded:
	// with a fixed injected generator (WithIDGen) every attempt collides, so give up and
	// surface the conflict rather than spin.
	var lastErr error
	for attempt := 0; attempt < maxIDMintAttempts; attempt++ {
		r.ID = s.newIDAt(day.UnixMilli())
		got, err := s.store.CreateResearch(r, body, p.DryRun)
		if err == nil {
			return got, nil
		}
		if !errors.Is(err, domain.ErrConflict) {
			return domain.Research{}, err
		}
		lastErr = err
	}
	return domain.Research{}, lastErr
}

// maxIDMintAttempts bounds the regenerate-on-collision loop in NewResearch. A real
// collision has probability 2^-17 per same-day pair, so a second attempt effectively
// always succeeds; the bound exists for the degenerate case of a fixed injected id
// generator, where every attempt collides and spinning would hang.
const maxIDMintAttempts = 8

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

// SetResearchFields updates frontmatter fields on a research doc (`research set`) — the
// agent face of mutation, beside the human EditResearch. updated_at is stamped here so
// every adapter gets it. force accepts a key the tool doesn't know (mirroring task/epic
// set), which is how the legacy corpus's vestigial `status: reference` stays reachable.
//
// Protected fields are rejected with the reason from the domain, so the CLI, a future
// TUI action, and any other adapter all explain the refusal identically.
func (s *Service) SetResearchFields(slug string, updates map[string]any, force, dryRun bool) (domain.Research, error) {
	if len(updates) == 0 {
		return domain.Research{}, fmt.Errorf("%w: no fields given", domain.ErrValidation)
	}
	clean := make(map[string]any, len(updates)+1)
	for field, val := range updates {
		if reason, protected := domain.ProtectedResearchField(field); protected {
			return domain.Research{}, fmt.Errorf("%w: %s cannot be set — %s", domain.ErrValidation, field, reason)
		}
		if _, unset := val.(domain.UnsetField); unset {
			// A typo'd key must not silently no-op — gate unset on the registry too,
			// mirroring the set path (and the task/epic contract).
			if !force && !domain.KnownResearchField(field) {
				return domain.Research{}, unknownResearchFieldErr(field)
			}
			clean[field] = val
			continue
		}
		if !force && !domain.KnownResearchField(field) {
			return domain.Research{}, unknownResearchFieldErr(field)
		}
		coerced := coerceResearchField(field, val)
		if err := domain.ValidateResearchField(field, stringify(coerced)); err != nil {
			return domain.Research{}, err
		}
		clean[field] = coerced
	}
	clean["updated_at"] = s.now().Format("2006-01-02")
	return retryOnConflict(s, dryRun, func() (domain.Research, error) {
		return s.store.SetResearchFields(slug, clean, dryRun)
	})
}

// coerceResearchField turns a string `--set key=value` into the field's native YAML
// type. Only `tags` is a list; everything else stays a scalar. A value that already
// arrived typed (from a typed flag) passes through.
func coerceResearchField(field string, val any) any {
	str, isStr := val.(string)
	if !isStr {
		return val
	}
	if domain.IsResearchListField(field) {
		return splitList(str)
	}
	return str
}

func unknownResearchFieldErr(field string) error {
	return fmt.Errorf("%w: unknown research field %q (known: %s) — use --force to set it anyway",
		domain.ErrValidation, field, strings.Join(domain.KnownResearchFieldNames(), ", "))
}

// EditResearch opens a research doc for whole-file editing — the human face of mutation,
// complementing the agent-facing `research set`/`research append`. The store accepts the
// save only if it still parses. Returns the reloaded doc and whether anything changed.
func (s *Service) EditResearch(slug string, edit func(current string, prevErr error) (string, error)) (domain.Research, bool, error) {
	return s.store.EditResearch(slug, s.now(), edit)
}

// AppendResearchBody appends a section to a research doc's body (`research append`) in
// one atomic, validated write — the agent face of body editing. Stamps updated_at;
// `created` stays immutable because the id is minted from it.
func (s *Service) AppendResearchBody(slug, text string, dryRun bool) (domain.Research, string, error) {
	now := s.now()
	type res struct {
		doc  domain.Research
		body string
	}
	r, err := retryOnConflict(s, dryRun, func() (res, error) {
		d, b, e := s.store.AppendResearchBody(slug, text, now, dryRun)
		return res{d, b}, e
	})
	return r.doc, r.body, err
}
