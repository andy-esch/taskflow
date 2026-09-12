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
		"Alpha Task<br/>ready-to-start &#183; Thread member<br/>ID 6fjangd7kvh0",
		"Gamma Task<br/>completed &#183; external prerequisite<br/>ID 6fjangd7kvh2",
		"--> ",
		`subgraph legend["Legend &#183; compact labels &#183; add --details for descriptions"]`,
		"Thread member<br/>blue &#183; solid border",
		"External prerequisite<br/>not a Thread member &#183; amber &#183; dashed border",
	} {
		if !strings.Contains(mermaid, want) {
			t.Errorf("Mermaid missing %q:\n%s", want, mermaid)
		}
	}
	if strings.Contains(mermaid, "A fully specified ready-to-start task") {
		t.Fatalf("compact Mermaid leaked a task description:\n%s", mermaid)
	}
	detailed := runRoot(t, "-C", fixtureRepo, "thread", "graph", "fixture-thread", "--details")
	if !strings.Contains(detailed, "A fully specified ready-to-start task for golden snapshots") {
		t.Fatalf("--details omitted the task description:\n%s", detailed)
	}
	if !strings.Contains(detailed, `Legend &#183; detailed labels &#183; descriptions included`) ||
		strings.Contains(detailed, "add --details") {
		t.Fatalf("--details retained compact-view guidance:\n%s", detailed)
	}

	dot := runRoot(t, "-C", fixtureRepo, "thread", "graph", "fixture-thread", "--format", "dot")
	for _, want := range []string{
		"digraph thread {", `role="member"`, `role="external-gate"`, " -> ",
		"subgraph cluster_legend", `Thread member\nblue fill, solid border`,
		`External prerequisite\nnot a Thread member\namber fill, dashed border`,
	} {
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
	tests := []struct {
		args []string
		want string
	}{
		{args: []string{"-C", fixtureRepo, "thread", "graph", "fixture-thread", "--format", "dot", "--json"}, want: "renderer flags"},
		{args: []string{"-C", fixtureRepo, "thread", "graph", "fixture-thread", "--details", "--json"}, want: "renderer flags"},
		{args: []string{"-C", fixtureRepo, "thread", "graph", "fixture-thread", "--format", "ascii"}, want: "format"},
		{args: []string{"-C", fixtureRepo, "thread", "graph", "missing-thread", "--format", "ascii"}, want: "unsupported Thread graph format"},
	}
	for _, test := range tests {
		out, err := runRootRC(t, test.args...)
		if err == nil {
			t.Fatalf("%v should fail, output=%s", test.args, out)
		}
		if !strings.Contains(err.Error(), test.want) {
			t.Fatalf("%v error=%v", test.args, err)
		}
	}
}
