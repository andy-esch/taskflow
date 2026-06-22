package core

import "github.com/andy-esch/taskflow/internal/domain"

// TemplateSource resolves body templates by kind. The built-in source reads the
// domain registry; epic 22's repo-local source layers user templates over it.
// It's a port so template DATA resolution moves behind core.Service like every
// other read/create surface — `template list/show` and the create paths consume
// it instead of reaching into domain.Template* directly, which makes step 4
// (repo-local templates) a source swap here rather than a CLI refactor.
type TemplateSource interface {
	// Templates lists the named templates a kind offers (default first);
	// ErrValidation for an unknown kind.
	Templates(kind string) ([]domain.NamedTemplate, error)
	// Lookup resolves one named template (empty name = the kind's default) to its
	// metadata + raw {{placeholder}} body; ErrValidation for an unknown kind/name.
	Lookup(kind, name string) (domain.NamedTemplate, error)
}

// builtinTemplates is the default TemplateSource: the domain's compiled-in
// registry — the same data `schema` self-describes. It needs no planning repo, so
// `template list/show` run anywhere; a repo-local source layers over it later.
type builtinTemplates struct{}

func (builtinTemplates) Templates(kind string) ([]domain.NamedTemplate, error) {
	return domain.TemplatesFor(kind)
}

func (builtinTemplates) Lookup(kind, name string) (domain.NamedTemplate, error) {
	return domain.LookupTemplate(kind, name)
}

// TemplateInfo is one template's listable metadata (kind + name + description) —
// the core read-result `template list`/`show` map into their render DTO.
type TemplateInfo struct {
	Kind        string
	Name        string
	Description string
}

// ListTemplates returns the templates a kind offers, or — when kind=="" — every
// kind's templates in schema (registry) order. An unknown kind is ErrValidation.
// The slice is always non-nil so a `--json` caller emits [] rather than null.
func (s *Service) ListTemplates(kind string) ([]TemplateInfo, error) {
	kinds := domain.SchemaKinds()
	if kind != "" {
		if _, err := s.templates.Templates(kind); err != nil {
			return nil, err // validate the one requested kind → ErrValidation (exit 11)
		}
		kinds = []string{kind}
	}
	out := []TemplateInfo{}
	for _, k := range kinds {
		ts, _ := s.templates.Templates(k) // k is known (from SchemaKinds or validated above)
		for _, t := range ts {
			out = append(out, TemplateInfo{Kind: k, Name: t.Name, Description: t.Description})
		}
	}
	return out, nil
}

// ShowTemplate resolves one named template (empty name = the kind's default) to
// its metadata + raw {{placeholder}} body. An unknown kind/name is ErrValidation.
// The caller renders the body raw (for forking) or with preview labels
// (RenderLabels) — that's a presentation choice, not a resolution one.
func (s *Service) ShowTemplate(kind, name string) (TemplateInfo, string, error) {
	nt, err := s.templates.Lookup(kind, name)
	if err != nil {
		return TemplateInfo{}, "", err
	}
	return TemplateInfo{Kind: kind, Name: nt.Name, Description: nt.Description}, nt.Body, nil
}

// templateBody returns a kind's raw template body via the source (empty name =
// the kind's default). The create paths fill its placeholders with real values.
func (s *Service) templateBody(kind, name string) (string, error) {
	nt, err := s.templates.Lookup(kind, name)
	return nt.Body, err
}
