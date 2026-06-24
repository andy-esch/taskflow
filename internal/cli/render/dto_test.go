package render

import (
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/andy-esch/taskflow/internal/domain"
)

// TestTaskJSONDescriptionTagMatchesCap guards the one description-cap copy that
// can't derive from domain.MaxDescriptionLen — the jsonschema struct tag (tags are
// static literals). Every other copy (CLI flag help, schema authoring guidance)
// is built from the constant; this test keeps the tag from silently drifting.
func TestTaskJSONDescriptionTagMatchesCap(t *testing.T) {
	f, ok := reflect.TypeOf(taskJSON{}).FieldByName("Description")
	if !ok {
		t.Fatal("taskJSON has no Description field")
	}
	tag := f.Tag.Get("jsonschema")
	want := strconv.Itoa(domain.MaxDescriptionLen)
	if !strings.Contains(tag, want) {
		t.Errorf("taskJSON Description jsonschema tag %q must mention the cap %s — update the tag (and the schema golden) when MaxDescriptionLen changes", tag, want)
	}
}
