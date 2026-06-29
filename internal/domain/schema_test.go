package domain

import (
	"reflect"
	"strings"
	"testing"
)

// TestTaskAuthoringFieldsMatchRegistry is the no-drift guard: every documented
// task authoring field must be a real known field, and its declared type must
// equal the registry's — so the schema can never advertise a type the writer
// won't produce, or a field the tool doesn't know.
func TestTaskAuthoringFieldsMatchRegistry(t *testing.T) {
	fields, err := AuthoringFields("task")
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range fields {
		if !KnownTaskField(f.Name) {
			t.Errorf("authoring field %q is not a known task field", f.Name)
		}
		if got := FieldType(f.Name); got != f.Type {
			t.Errorf("field %q: doc type %q != registry type %q", f.Name, f.Type, got)
		}
		if f.Description == "" || f.Example == "" {
			t.Errorf("field %q: description/example must be non-empty", f.Name)
		}
	}
}

// TestEpicAuthoringFieldsMatchRegistry mirrors the task no-drift guard for epics:
// every documented epic authoring field must be a real known epic field, so the
// schema can never advertise an epic field the tool doesn't know.
func TestEpicAuthoringFieldsMatchRegistry(t *testing.T) {
	fields, err := AuthoringFields("epic")
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range fields {
		if !KnownEpicField(f.Name) {
			t.Errorf("authoring field %q is not a known epic field", f.Name)
		}
		if f.Description == "" || f.Example == "" {
			t.Errorf("field %q: description/example must be non-empty", f.Name)
		}
	}
}

// TestAuditAuthoringFieldsMatchStruct is the audit no-drift guard, mirroring the
// task/epic ones. Audits have no settable-field map (area/date are immutable
// identity — there is no `audit set`), so the registry they drift against is the
// Audit struct's own yaml tags: every documented audit authoring field must be a
// real persisted field, so `schema audit` can't advertise a field the tool won't
// store. (Type isn't compared to FieldType the way the task guard does — FieldType
// is a task-field utility; the epic guard skips it for the same reason.)
func TestAuditAuthoringFieldsMatchStruct(t *testing.T) {
	fields, err := AuthoringFields("audit")
	if err != nil {
		t.Fatal(err)
	}
	persisted := map[string]bool{}
	rt := reflect.TypeOf(Audit{})
	for i := range rt.NumField() {
		name, _, _ := strings.Cut(rt.Field(i).Tag.Get("yaml"), ",")
		if name != "" && name != "-" {
			persisted[name] = true
		}
	}
	for _, f := range fields {
		if !persisted[f.Name] {
			t.Errorf("audit authoring field %q is not a yaml field on domain.Audit", f.Name)
		}
		if f.Description == "" || f.Example == "" {
			t.Errorf("field %q: description/example must be non-empty", f.Name)
		}
	}
}

func TestFieldType(t *testing.T) {
	for name, want := range map[string]string{
		"tier": "int", "autonomy_level": "int",
		"tags": "list", "dependencies": "list",
		"created": "date", "audited": "date",
		"description": "string", "epic": "string", "nonexistent": "string",
	} {
		if got := FieldType(name); got != want {
			t.Errorf("FieldType(%q) = %q, want %q", name, got, want)
		}
	}
}

func TestAuthoringFields_Kinds(t *testing.T) {
	if _, err := AuthoringFields("bogus"); err == nil {
		t.Error("unknown kind should error")
	}
	for _, k := range SchemaKinds() {
		f, err := AuthoringFields(k)
		if err != nil || len(f) == 0 {
			t.Errorf("kind %q: err=%v fields=%d (want resolvable, non-empty)", k, err, len(f))
		}
		if len(Conventions(k)) == 0 {
			t.Errorf("kind %q: conventions should be non-empty", k)
		}
	}
}
