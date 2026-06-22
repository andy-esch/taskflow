package core

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/andy-esch/taskflow/internal/domain"
)

// TestRenderTemplate_LiteralPercentSafe pins the named-placeholder model's core
// safety win over the old Printf approach: {{key}} is filled, a literal '%' is left
// untouched (no format hazard), and an unknown {{token}} is left visible (not
// silently dropped) — exactly what makes author-supplied bodies safe.
func TestRenderTemplate_LiteralPercentSafe(t *testing.T) {
	got := renderTemplate("100% done: {{x}} / {{missing}}", map[string]string{"x": "Y"})
	if want := "100% done: Y / {{missing}}"; got != want {
		t.Errorf("renderTemplate = %q, want %q", got, want)
	}
}

// TestTemplateBody_NamedAndLabels: the core renderer resolves a named template and
// applies the kind's preview labels (registry-driven, no per-kind switch); unknown
// kind/name → ErrValidation.
func TestTemplateBody_NamedAndLabels(t *testing.T) {
	def, err := TemplateBody("audit", "")
	if err != nil {
		t.Fatal(err)
	}
	sec, err := TemplateBody("audit", "security")
	if err != nil {
		t.Fatal(err)
	}
	if def == sec {
		t.Error("default and security templates should differ")
	}
	if !strings.Contains(sec, "Security audit:") || !strings.Contains(sec, "<area>") || strings.Contains(sec, "{{area}}") {
		t.Errorf("security preview labels wrong:\n%s", sec)
	}
	task, err := TemplateBody("task", "")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(task, "<title>") || !strings.Contains(task, "<epic-id>") {
		t.Errorf("task preview labels wrong:\n%s", task)
	}
	if _, err := TemplateBody("audit", "nope"); !errors.Is(err, domain.ErrValidation) {
		t.Errorf("unknown name should be ErrValidation, got %v", err)
	}
}

// fakeTemplates is an in-memory TemplateSource. Injecting it proves the
// list/show/create paths resolve through the port — not domain.Template* directly
// — which is the whole point of the seam (epic 22 swaps the source, not the CLI).
type fakeTemplates struct {
	byKind map[string][]domain.NamedTemplate
}

func (f fakeTemplates) Templates(kind string) ([]domain.NamedTemplate, error) {
	ts, ok := f.byKind[kind]
	if !ok {
		return nil, fmt.Errorf("%w: unknown kind %q", domain.ErrValidation, kind)
	}
	return ts, nil
}

func (f fakeTemplates) Lookup(kind, name string) (domain.NamedTemplate, error) {
	ts, err := f.Templates(kind)
	if err != nil {
		return domain.NamedTemplate{}, err
	}
	if name == "" {
		name = "default"
	}
	for _, t := range ts {
		if t.Name == name {
			return t, nil
		}
	}
	return domain.NamedTemplate{}, fmt.Errorf("%w: unknown %s template %q", domain.ErrValidation, kind, name)
}

// TestService_TemplateSource_Seam pins the epic-22 port move: ShowTemplate,
// ListTemplates, and the create path all read the injected source, so a repo-local
// source will layer on with no change to these methods.
func TestService_TemplateSource_Seam(t *testing.T) {
	src := fakeTemplates{byKind: map[string][]domain.NamedTemplate{
		"task": {{Name: "default", Description: "fake-desc", Body: "FAKE-BODY {{title}}"}},
	}}
	fs := &fakeStore{epics: []domain.Epic{{ID: "e1"}}}
	svc := NewService(fs, WithTemplateSource(src))

	// ShowTemplate reads the injected source (metadata + raw, unfilled body).
	info, body, err := svc.ShowTemplate("task", "")
	if err != nil {
		t.Fatalf("ShowTemplate: %v", err)
	}
	if info.Kind != "task" || info.Name != "default" || info.Description != "fake-desc" || body != "FAKE-BODY {{title}}" {
		t.Errorf("ShowTemplate didn't read the injected source: %+v body=%q", info, body)
	}

	// ListTemplates reads the injected source.
	list, err := svc.ListTemplates("task")
	if err != nil || len(list) != 1 || list[0].Description != "fake-desc" {
		t.Fatalf("ListTemplates didn't read the injected source: %+v (%v)", list, err)
	}

	// The create path fills the injected body's placeholders with real values —
	// proof NewTask resolves through the port, not domain.Template*.
	if _, err := svc.NewTask(NewTaskParams{
		Title: "Hi", Epic: "e1", Tier: 3, Autonomy: 3, Priority: "medium", Tags: []string{"x"},
	}); err != nil {
		t.Fatal(err)
	}
	if len(fs.createdBodies) != 1 || !strings.Contains(fs.createdBodies[0], "FAKE-BODY Hi") {
		t.Errorf("create path didn't fill the injected template body: %q", fs.createdBodies)
	}
}

// TestService_ShowTemplate_UnknownIsValidation pins the exit-11 contract: an
// unknown kind/name surfaces ErrValidation through the service, as the old direct
// domain.LookupTemplate call did.
func TestService_ShowTemplate_UnknownIsValidation(t *testing.T) {
	svc := NewService(nopStore{}) // built-in source
	if _, _, err := svc.ShowTemplate("nope", ""); !errors.Is(err, domain.ErrValidation) {
		t.Errorf("unknown kind should be ErrValidation, got %v", err)
	}
	if _, err := svc.ListTemplates("nope"); !errors.Is(err, domain.ErrValidation) {
		t.Errorf("unknown kind list should be ErrValidation, got %v", err)
	}
}

// TestService_BuiltinTemplateService pins the repo-less fallback: a store-less
// service still answers template queries from the built-in registry.
func TestService_BuiltinTemplateService(t *testing.T) {
	svc := NewBuiltinTemplateService()
	list, err := svc.ListTemplates("")
	if err != nil {
		t.Fatalf("ListTemplates: %v", err)
	}
	if len(list) == 0 {
		t.Error("built-in service should list the compiled-in templates with no repo")
	}
	if _, body, err := svc.ShowTemplate("task", ""); err != nil || body == "" {
		t.Errorf("built-in service should resolve the default task template: body=%q (%v)", body, err)
	}
}
