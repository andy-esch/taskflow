package cli

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"github.com/andy-esch/taskflow/internal/wire"
)

func TestThreadGraphFormatsAndPlanUseSharedProjection(t *testing.T) {
	mermaid := runRoot(t, "-C", fixtureRepo, "thread", "graph", "fixture-thread")
	for _, want := range []string{
		"flowchart TD\n",
		"graph_health=healthy projection_health=healthy topology_complete=true",
		"6fjangd7kvh2",
		"external-gate",
		"--> ",
	} {
		if !strings.Contains(mermaid, want) {
			t.Errorf("Mermaid missing %q:\n%s", want, mermaid)
		}
	}

	dot := runRoot(t, "-C", fixtureRepo, "thread", "graph", "fixture-thread", "--format", "dot")
	for _, want := range []string{"digraph thread {", `role="member"`, `role="external-gate"`, " -> "} {
		if !strings.Contains(dot, want) {
			t.Errorf("DOT missing %q:\n%s", want, dot)
		}
	}

	plan := runRoot(t, "-C", fixtureRepo, "thread", "plan", "fixture-thread")
	for _, want := range []string{"topology complete", "External gates", "gamma-task", "Wave 1", "alpha-task", "beta-task"} {
		if !strings.Contains(plan, want) {
			t.Errorf("plan missing %q:\n%s", want, plan)
		}
	}
}

func TestThreadGraphAndPlanJSONExposeSameNeutralProjection(t *testing.T) {
	decode := func(command string) wire.ThreadGraphProjectionJSON {
		t.Helper()
		out := runRoot(t, "-C", fixtureRepo, "thread", command, "fixture-thread", "--json")
		var envelope struct {
			SchemaVersion string                         `json:"schema_version"`
			Projection    wire.ThreadGraphProjectionJSON `json:"projection"`
		}
		if err := json.Unmarshal([]byte(out), &envelope); err != nil {
			t.Fatalf("%s JSON: %v\n%s", command, err, out)
		}
		if envelope.SchemaVersion != wire.SchemaVersion || len(envelope.Projection.Nodes) != 3 ||
			len(envelope.Projection.Edges) != 1 || len(envelope.Projection.Waves) != 1 ||
			!envelope.Projection.TopologyComplete {
			t.Fatalf("%s projection=%+v", command, envelope.Projection)
		}
		return envelope.Projection
	}
	graph, plan := decode("graph"), decode("plan")
	if !reflect.DeepEqual(graph, plan) {
		t.Fatalf("graph and plan JSON diverged:\ngraph=%+v\nplan=%+v", graph, plan)
	}
}

func TestThreadGraphRejectsRendererSelectionInJSONAndUnknownFormats(t *testing.T) {
	for _, args := range [][]string{
		{"-C", fixtureRepo, "thread", "graph", "fixture-thread", "--format", "dot", "--json"},
		{"-C", fixtureRepo, "thread", "graph", "fixture-thread", "--format", "ascii"},
	} {
		out, err := runRootRC(t, args...)
		if err == nil {
			t.Fatalf("%v should fail, output=%s", args, out)
		}
		if !strings.Contains(err.Error(), "format") {
			t.Fatalf("%v error=%v", args, err)
		}
	}
}
