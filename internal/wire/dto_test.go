package wire

import (
	"encoding/json"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/andy-esch/taskflow/internal/domain"
)

// TestToTaskJSON_CarriesID checks the stable id flows into the wire DTO and that
// omitempty holds — a task with no id (created before id assignment) must not emit
// an empty "id" that a consumer could mistake for a real key.
func TestToTaskJSON_CarriesID(t *testing.T) {
	got := ToTaskJSON(domain.Task{Slug: "x", Status: domain.StatusInProgress, ID: "6fjangd7kvh1"})
	if got.ID != "6fjangd7kvh1" {
		t.Errorf("ToTaskJSON dropped the id: got %q", got.ID)
	}
	b, err := json.Marshal(ToTaskJSON(domain.Task{Slug: "x", Status: domain.StatusNextUp}))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(b), `"id"`) {
		t.Errorf("an id-less task must omit the id field, got %s", b)
	}
}

// TestToAuditJSON_CarriesID mirrors the task check for audits.
func TestToAuditJSON_CarriesID(t *testing.T) {
	got := ToAuditJSON(domain.Audit{Slug: "2026-01-02-x", Bucket: domain.AuditOpen, ID: "6fjangd7kvh3"})
	if got.ID != "6fjangd7kvh3" {
		t.Errorf("ToAuditJSON dropped the id: got %q", got.ID)
	}
}

// TestTaskJSONDescriptionTagMatchesCap guards the one description-cap copy that
// can't derive from domain.MaxDescriptionLen — the jsonschema struct tag (tags are
// static literals). Every other copy (CLI flag help, schema authoring guidance)
// is built from the constant; this test keeps the tag from silently drifting.
func TestTaskJSONDescriptionTagMatchesCap(t *testing.T) {
	f, ok := reflect.TypeOf(TaskJSON{}).FieldByName("Description")
	if !ok {
		t.Fatal("TaskJSON has no Description field")
	}
	tag := f.Tag.Get("jsonschema")
	want := strconv.Itoa(domain.MaxDescriptionLen)
	if !strings.Contains(tag, want) {
		t.Errorf("TaskJSON Description jsonschema tag %q must mention the cap %s — update the tag (and the schema golden) when MaxDescriptionLen changes", tag, want)
	}
}
