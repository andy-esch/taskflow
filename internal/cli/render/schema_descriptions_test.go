package render

import (
	"encoding/json"
	"testing"
)

// TestJSONSchema_HasFieldDescriptions guards the field-description feature against
// silently going inert (invopop's AddGoComments skips unexported projection types,
// so descriptions come from jsonschema struct tags — if those are dropped the
// schema reverts to no field descriptions and nothing else would catch it).
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
		"taskJSON":    {"slug", "status", "tier"},
		"findingJSON": {"code", "status", "effort"},
		"auditJSON":   {"bucket", "open_findings"},
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
