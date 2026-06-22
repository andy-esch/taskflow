package cli

import (
	"encoding/json"
	"strings"
	"testing"
)

// template list/show need no planning repo (built-in templates), so these run
// without -C — proving the schema-style "runs anywhere" wiring.

func TestTemplateList_Human(t *testing.T) {
	out := runRoot(t, "template", "list")
	for _, want := range []string{"task", "epic", "audit", "security"} {
		if !strings.Contains(out, want) {
			t.Errorf("template list missing %q:\n%s", want, out)
		}
	}
}

func TestTemplateList_JSON(t *testing.T) {
	out := runRoot(t, "template", "list", "--json")
	var got struct {
		SchemaVersion string `json:"schema_version"`
		Templates     []struct {
			Kind        string `json:"kind"`
			Name        string `json:"name"`
			Description string `json:"description"`
		} `json:"templates"`
	}
	if err := json.Unmarshal([]byte(out), &got); err != nil {
		t.Fatalf("bad json: %v\n%s", err, out)
	}
	if got.SchemaVersion == "" || len(got.Templates) < 4 {
		t.Errorf("unexpected envelope: %+v", got)
	}
}

func TestTemplateList_KindFilter(t *testing.T) {
	out := runRoot(t, "template", "list", "--kind", "audit", "--json")
	var got struct {
		Templates []struct {
			Kind string `json:"kind"`
		} `json:"templates"`
	}
	if err := json.Unmarshal([]byte(out), &got); err != nil {
		t.Fatal(err)
	}
	if len(got.Templates) != 2 {
		t.Fatalf("--kind audit should list 2 templates, got %d", len(got.Templates))
	}
	for _, tpl := range got.Templates {
		if tpl.Kind != "audit" {
			t.Errorf("--kind audit returned a %q template", tpl.Kind)
		}
	}
}

func TestTemplateList_UnknownKindRejected(t *testing.T) {
	if _, err := runRootRC(t, "template", "list", "--kind", "bogus"); err == nil || ExitCode(err) != 11 {
		t.Fatalf("unknown --kind should exit 11, got %v", err)
	}
}

func TestTemplateShow_SecurityBody(t *testing.T) {
	out := runRoot(t, "template", "show", "audit", "security")
	for _, want := range []string{"Security audit:", "Threat model", "<area>"} {
		if !strings.Contains(out, want) {
			t.Errorf("template show missing %q:\n%s", want, out)
		}
	}
}

func TestTemplateShow_DefaultsToDefault(t *testing.T) {
	out := runRoot(t, "template", "show", "task")
	if !strings.Contains(out, "## Acceptance criteria") {
		t.Errorf("template show task (default) missing its body:\n%s", out)
	}
}

func TestTemplateShow_UnknownNameRejected(t *testing.T) {
	_, err := runRootRC(t, "template", "show", "audit", "bogus")
	if err == nil || ExitCode(err) != 11 {
		t.Fatalf("unknown template name should exit 11, got %v", err)
	}
	if !strings.Contains(err.Error(), "security") {
		t.Errorf("error should list the available templates: %v", err)
	}
}
