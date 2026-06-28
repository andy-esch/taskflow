package wire

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/andy-esch/taskflow/internal/domain"
)

// TestJSONSchema_HasFieldDescriptions guards the field-description feature against
// silently going inert (descriptions on the projection DTOs come from jsonschema
// struct tags — if those are dropped the schema reverts to no field descriptions
// and nothing else would catch it).
func TestJSONSchema_HasFieldDescriptions(t *testing.T) {
	b, err := JSONSchema()
	if err != nil {
		t.Fatal(err)
	}
	var doc struct {
		Defs map[string]struct {
			Properties map[string]struct {
				Description string `json:"description"`
			} `json:"properties"`
		} `json:"$defs"`
	}
	if err := json.Unmarshal(b, &doc); err != nil {
		t.Fatal(err)
	}
	// The entity projections agents validate against must carry per-field descriptions.
	want := map[string][]string{
		"TaskJSON":    {"slug", "status", "tier"},
		"FindingJSON": {"code", "status", "effort"},
		"AuditJSON":   {"bucket", "open_findings"},
	}
	for def, fields := range want {
		props := doc.Defs[def].Properties
		for _, f := range fields {
			if props[f].Description == "" {
				t.Errorf("$defs/%s/properties/%s should carry a field description", def, f)
			}
		}
	}
}

// TestEpicStatusDescriptionMatchesVocab guards the one epic-status description that
// can't be derived: EpicMetaJSON.Status's `jsonschema:"description=…"` struct TAG
// (dto.go) is a compile-time literal, so a future epic-vocab change would silently
// leave it stale. Pin it to domain.AllEpicStatuses() so the drift fails loudly here.
func TestEpicStatusDescriptionMatchesVocab(t *testing.T) {
	b, err := JSONSchema()
	if err != nil {
		t.Fatal(err)
	}
	var doc struct {
		Defs map[string]struct {
			Properties map[string]struct {
				Description string `json:"description"`
			} `json:"properties"`
		} `json:"$defs"`
	}
	if err := json.Unmarshal(b, &doc); err != nil {
		t.Fatal(err)
	}
	desc := doc.Defs["EpicMetaJSON"].Properties["status"].Description
	if desc == "" {
		t.Fatal("EpicMetaJSON.status should carry a description tag")
	}
	for _, s := range domain.AllEpicStatuses() {
		if !strings.Contains(desc, s) {
			t.Errorf("EpicMetaJSON.status description %q is missing epic status %q "+
				"(dto.go's jsonschema tag drifted from domain.AllEpicStatuses())", desc, s)
		}
	}
}
