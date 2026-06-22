package domain

import (
	"errors"
	"regexp"
	"strings"
	"testing"
)

// TestTemplate_DefaultAndNamed: an empty name selects the kind's default; a named
// one selects that template; the security audit template is distinct and recognizable.
func TestTemplate_DefaultAndNamed(t *testing.T) {
	def, err := Template("audit", "")
	if err != nil {
		t.Fatalf("default audit template: %v", err)
	}
	named, err := Template("audit", "default")
	if err != nil || named != def {
		t.Errorf(`Template("audit","default") should equal the empty-name default`)
	}
	sec, err := Template("audit", "security")
	if err != nil {
		t.Fatalf("security audit template: %v", err)
	}
	if sec == def {
		t.Error("security template should differ from default")
	}
	if !strings.Contains(sec, "Security audit:") || !strings.Contains(sec, "Review checklist") {
		t.Errorf("security template missing its distinctive sections:\n%s", sec)
	}
}

// TestTemplate_UnknownErrors: unknown kind or template name is ErrValidation, and
// the bad-name message lists what's available (so the CLI can guide the user).
func TestTemplate_UnknownErrors(t *testing.T) {
	if _, err := Template("bogus", ""); !errors.Is(err, ErrValidation) {
		t.Errorf("unknown kind should be ErrValidation, got %v", err)
	}
	_, err := Template("audit", "nope")
	if !errors.Is(err, ErrValidation) {
		t.Fatalf("unknown template name should be ErrValidation, got %v", err)
	}
	for _, want := range []string{"default", "security"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error should list available template %q: %v", want, err)
		}
	}
}

// TestTemplateNames: default is first (so it's the obvious choice in completion),
// audit offers security, and an unknown kind yields nil.
func TestTemplateNames(t *testing.T) {
	for _, kind := range SchemaKinds() {
		names := TemplateNames(kind)
		if len(names) == 0 || names[0] != DefaultTemplate {
			t.Errorf("%s: TemplateNames should start with %q, got %v", kind, DefaultTemplate, names)
		}
	}
	if got := TemplateNames("audit"); !contains(got, "security") {
		t.Errorf(`audit should offer "security", got %v`, got)
	}
	if TemplateNames("bogus") != nil {
		t.Error("TemplateNames(bogus) should be nil")
	}
}

// placeholderRe matches a {{key}} fill token.
var placeholderRe = regexp.MustCompile(`{{([a-z_]+)}}`)

// TestTemplates_OnlyDeclaredPlaceholders is the load-bearing robustness guard for
// the named-placeholder model: every {{token}} a template uses must be a declared
// Placeholder for its kind (so the create path fills it — no leftover {{token}} in
// a created file), and no legacy Printf %s/%d remains. This is the invariant an
// author-supplied/repo-local template (step 4) must satisfy, validated on load.
func TestTemplates_OnlyDeclaredPlaceholders(t *testing.T) {
	for _, kind := range SchemaKinds() {
		declared := map[string]bool{}
		for _, p := range Placeholders(kind) {
			declared[p.Key] = true
		}
		for _, name := range TemplateNames(kind) {
			body, err := Template(kind, name)
			if err != nil {
				t.Fatalf("%s/%s: %v", kind, name, err)
			}
			for _, m := range placeholderRe.FindAllStringSubmatch(body, -1) {
				if !declared[m[1]] {
					t.Errorf("%s/%s uses undeclared placeholder %q (declared: %v)", kind, name, m[1], Placeholders(kind))
				}
			}
			if strings.Contains(body, "%s") || strings.Contains(body, "%d") {
				t.Errorf("%s/%s still has a legacy Printf verb; templates use {{key}} now:\n%s", kind, name, body)
			}
		}
	}
}

// TestAuditTemplates_FreshCountZeroFindings: every audit template's finding example
// is fenced, so a freshly created audit has zero open findings and is lint-clean.
// (Placeholders live in the header, not findings, so the raw body is representative.)
func TestAuditTemplates_FreshCountZeroFindings(t *testing.T) {
	for _, name := range TemplateNames("audit") {
		body, err := Template("audit", name)
		if err != nil {
			t.Fatal(err)
		}
		if n := len(ParseFindings(body)); n != 0 {
			t.Errorf("audit/%s: a fresh audit should have 0 parsed findings, got %d", name, n)
		}
	}
}

func contains(ss []string, want string) bool {
	for _, s := range ss {
		if s == want {
			return true
		}
	}
	return false
}

// TestLookupTemplate: metadata access for `template show` — default vs named, and
// ErrValidation for an unknown kind or name.
func TestLookupTemplate(t *testing.T) {
	def, err := LookupTemplate("audit", "")
	if err != nil || def.Name != DefaultTemplate {
		t.Fatalf("default lookup: err=%v name=%q", err, def.Name)
	}
	sec, err := LookupTemplate("audit", "security")
	if err != nil || sec.Name != "security" || sec.Description == "" || sec.Body == "" {
		t.Errorf("security lookup: err=%v info=%+v", err, sec)
	}
	if _, err := LookupTemplate("audit", "nope"); !errors.Is(err, ErrValidation) {
		t.Errorf("unknown name should be ErrValidation, got %v", err)
	}
	if _, err := LookupTemplate("bogus", ""); !errors.Is(err, ErrValidation) {
		t.Errorf("unknown kind should be ErrValidation, got %v", err)
	}
}

// TestTemplatesFor: the listable set per kind, ErrValidation for unknown, and a
// defensive copy (mutating the result must not corrupt the registry).
func TestTemplatesFor(t *testing.T) {
	ts, err := TemplatesFor("audit")
	if err != nil || len(ts) != 2 {
		t.Fatalf("audit templates: err=%v n=%d", err, len(ts))
	}
	if _, err := TemplatesFor("bogus"); !errors.Is(err, ErrValidation) {
		t.Errorf("unknown kind should be ErrValidation, got %v", err)
	}
	ts[0].Name = "MUTATED"
	if again, _ := TemplatesFor("audit"); again[0].Name == "MUTATED" {
		t.Error("TemplatesFor must return a copy, not the registry's slice")
	}
}
