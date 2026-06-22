package core

import (
	"errors"
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
