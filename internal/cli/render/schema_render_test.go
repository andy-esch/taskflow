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

func TestSchemaHumanSortsExitCodesAndShowsReservations(t *testing.T) {
	c := SchemaContract{ExitCodes: []wire.SchemaExitCode{
		{Code: 10, Name: "not-found", State: wire.ExitCodeStateActive, Meaning: "missing entity"},
		{Code: 130, Name: "aborted", State: wire.ExitCodeStateActive, Meaning: "prompt aborted"},
		{Code: 0, Name: "ok", State: wire.ExitCodeStateActive, Meaning: "success"},
		{Code: 12, Name: "invalid-transition", State: wire.ExitCodeStateReserved, Meaning: "retired code"},
	}}
	var out bytes.Buffer
	if err := SchemaHuman(&out, NewStyle(false), c); err != nil {
		t.Fatalf("SchemaHuman: %v", err)
	}
	rendered := out.String()
	ordered := []string{"  0   ok", "  10  not-found", "  12  invalid-transition", "  130 aborted"}
	last := -1
	for _, want := range ordered {
		at := strings.Index(rendered, want)
		if at < 0 {
			t.Fatalf("human schema missing exit row %q:\n%s", want, rendered)
		}
		if at <= last {
			t.Fatalf("human exit rows are not numerically sorted:\n%s", rendered)
		}
		last = at
	}
	if !strings.Contains(rendered, "invalid-transition  retired code [reserved]") {
		t.Fatalf("human schema must identify the reserved row:\n%s", rendered)
	}
}
