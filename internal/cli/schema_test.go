package cli

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/andy-esch/taskflow/internal/core"
	"github.com/andy-esch/taskflow/internal/domain"
)

// TestSchema_RunsWithoutPlanningRepo pins the feature's reason to exist: an agent
// must be able to learn the contract in any repo. A bare dir has no tasks/, so
// the root's resolve() would fail — schema overrides it, so this must succeed.
func TestSchema_RunsWithoutPlanningRepo(t *testing.T) {
	bare := t.TempDir()
	if js := runRoot(t, "-C", bare, "schema", "--json"); !strings.Contains(js, "schema_version") {
		t.Errorf("schema should run without a planning repo, got: %q", js)
	}
}

func TestSchemaContract_JSON(t *testing.T) {
	root := freshRepo(t)
	js := runRoot(t, "-C", root, "schema", "--json")
	var c struct {
		SchemaVersion string `json:"schema_version"`
		Statuses      []struct {
			Value  string `json:"value"`
			Active bool   `json:"active"`
		} `json:"statuses"`
		TaskFields []struct{ Name, Type string } `json:"task_fields"`
		ExitCodes  []struct {
			Code int    `json:"code"`
			Name string `json:"name"`
		} `json:"exit_codes"`
		Kinds []string `json:"kinds"`
	}
	if err := json.Unmarshal([]byte(js), &c); err != nil {
		t.Fatalf("schema --json invalid: %v\n%s", err, js)
	}
	// Statuses are derived from the domain set, carrying the active flag.
	if len(c.Statuses) != len(domain.AllStatuses()) {
		t.Errorf("statuses: got %d, want %d (the domain set)", len(c.Statuses), len(domain.AllStatuses()))
	}
	// The field registry is the domain registry, with real (derived) types —
	// including `audited`, proving the type map is the single source.
	types := map[string]string{}
	for _, f := range c.TaskFields {
		types[f.Name] = f.Type
	}
	if types["audited"] != "date" || types["tier"] != "int" || types["tags"] != "list" {
		t.Errorf("task_fields types not derived from the registry: %+v", types)
	}
	if len(c.TaskFields) != len(domain.KnownTaskFieldNames()) {
		t.Errorf("task_fields should cover the whole registry: got %d want %d",
			len(c.TaskFields), len(domain.KnownTaskFieldNames()))
	}
	if len(c.ExitCodes) == 0 || len(c.Kinds) != 3 {
		t.Errorf("contract missing exit codes / kinds: %+v %+v", c.ExitCodes, c.Kinds)
	}
}

func TestSchemaKind_DerivedFromScaffold(t *testing.T) {
	root := freshRepo(t)
	js := runRoot(t, "-C", root, "schema", "task", "--json")
	var ks struct {
		Kind         string                        `json:"kind"`
		Sections     []string                      `json:"sections"`
		BodyTemplate string                        `json:"body_template"`
		Fields       []struct{ Name, Type string } `json:"fields"`
	}
	if err := json.Unmarshal([]byte(js), &ks); err != nil {
		t.Fatalf("schema task --json invalid: %v\n%s", err, js)
	}
	// body_template IS the scaffold — never a second, drift-prone copy.
	want, err := core.ScaffoldBody("task")
	if err != nil {
		t.Fatal(err)
	}
	if ks.BodyTemplate != want {
		t.Errorf("body_template should equal the live scaffold")
	}
	// Sections are the scaffold's own `##` headers, in order.
	gotSections := map[string]bool{}
	for _, s := range ks.Sections {
		gotSections[s] = true
	}
	for _, sec := range []string{"Objective", "Acceptance criteria", "Out of scope", "Related"} {
		if !gotSections[sec] {
			t.Errorf("missing derived section %q in %v", sec, ks.Sections)
		}
	}
}

func TestSchema_UnknownKind_Exit11(t *testing.T) {
	root := freshRepo(t)
	var out bytes.Buffer
	cmd := NewRootCmd(&out, &out)
	cmd.SetArgs([]string{"-C", root, "schema", "bogus"})
	if err := cmd.Execute(); err == nil || ExitCode(err) != 11 {
		t.Errorf("unknown kind should exit 11 (validation), got %v", err)
	}
}
