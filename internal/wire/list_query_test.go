package wire

import (
	"bytes"
	"encoding/json"
	"reflect"
	"testing"

	"github.com/santhosh-tekuri/jsonschema/v6"

	"github.com/andy-esch/taskflow/internal/core"
)

func TestListPageMetadataIsOptionalAndValidatesOnAllTypedLists(t *testing.T) {
	page := &core.ListPage{Order: []string{"-updated", "source_id", "value_digest"}, Limit: 7, Returned: 7,
		Matched: 19, Search: "bike", UpdatedSince: "2026-10-10", NextCursor: "opaque-continuation"}
	metadata := ToListPageJSON(page)
	if ToListPageJSON(nil) != nil || metadata.Limit != 7 || metadata.Returned != 7 || metadata.Matched != 19 || metadata.NextCursor != page.NextCursor ||
		metadata.Search != page.Search || metadata.UpdatedSince != page.UpdatedSince || !reflect.DeepEqual(metadata.Order, page.Order) {
		t.Fatalf("metadata=%+v", metadata)
	}
	tasks := ToTasksEnvelope(nil, nil)
	tasks.Page = metadata
	epics := ToEpicsEnvelope(nil, nil)
	epics.Page = metadata
	audits := ToAuditsEnvelope(nil, nil)
	audits.Page = metadata
	research := ToResearchListEnvelope(nil, nil)
	research.Page = metadata
	findings := ToFindingsEnvelope(nil, nil)
	findings.Page = metadata
	threads := ToThreadsEnvelope(core.ThreadListView{GraphHealth: core.GraphHealthy}, nil)
	threads.Page = metadata
	b, err := JSONSchema()
	if err != nil {
		t.Fatal(err)
	}
	doc, err := jsonschema.UnmarshalJSON(bytes.NewReader(b))
	if err != nil {
		t.Fatal(err)
	}
	id := doc.(map[string]any)["$id"].(string)
	compiler := jsonschema.NewCompiler()
	if err := compiler.AddResource(id, doc); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name  string
		value any
	}{
		{"TasksEnvelope", tasks}, {"EpicsEnvelope", epics}, {"AuditsEnvelope", audits},
		{"ResearchListEnvelope", research}, {"FindingsEnvelope", findings}, {"ThreadsEnvelope", threads},
	} {
		t.Run(tc.name, func(t *testing.T) {
			schema, err := compiler.Compile(id + "#/$defs/" + tc.name)
			if err != nil {
				t.Fatal(err)
			}
			encoded, err := json.Marshal(tc.value)
			if err != nil {
				t.Fatal(err)
			}
			var instance map[string]any
			if err := json.Unmarshal(encoded, &instance); err != nil {
				t.Fatal(err)
			}
			if err := schema.Validate(instance); err != nil {
				t.Fatalf("nondefault page: %v", err)
			}
			p := instance["page"].(map[string]any)
			if p["next_cursor"] != page.NextCursor || p["returned"] != float64(7) {
				t.Fatalf("lost values: %+v", p)
			}
			p["limit"] = float64(1001)
			if err := schema.Validate(instance); err == nil {
				t.Fatal("schema accepted out-of-range limit")
			}
			delete(instance, "page")
			if err := schema.Validate(instance); err != nil {
				t.Fatalf("legacy without page: %v", err)
			}
		})
	}
}
