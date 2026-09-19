package render

import (
	"bytes"
	"strings"
	"testing"

	"github.com/andy-esch/taskflow/internal/wire"
)

func TestSchemaHumanNormalizesRevisionPolicy(t *testing.T) {
	for _, policy := range []wire.SchemaRevisionPolicy{
		{},
		{Scheme: "stale", Scope: "contradictory"},
	} {
		var out bytes.Buffer
		if err := SchemaHuman(&out, NewStyle(false), SchemaContract{RevisionPolicy: policy}); err != nil {
			t.Fatalf("SchemaHuman: %v", err)
		}
		for _, want := range []string{
			"JSON revision: monotonic-revision · additive · classified since 1.68",
			"Scope: all-json-output · generated schema: typed-envelopes",
		} {
			if !strings.Contains(out.String(), want) {
				t.Errorf("human schema missing normalized policy %q:\n%s", want, out.String())
			}
		}
		if strings.Contains(out.String(), "stale") || strings.Contains(out.String(), "contradictory") {
			t.Fatalf("human schema rendered caller-supplied policy:\n%s", out.String())
		}
	}
}
