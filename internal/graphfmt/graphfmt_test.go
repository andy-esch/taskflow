package graphfmt

import (
	"strings"
	"testing"

	"github.com/andy-esch/taskflow/internal/core"
	"github.com/andy-esch/taskflow/internal/domain"
)

func hostileProjection() core.ThreadGraphProjection {
	return core.ThreadGraphProjection{
		Nodes: []core.ThreadGraphNode{
			{
				TaskID: "6g0000000001", Label: `first \"task\"`, Description: "line <one>\n%%{init:x}%%",
				Status: domain.StatusNextUp, Role: core.ThreadTaskMember,
			},
			{
				TaskID: "6g0000000002", Label: "outside]gate", Description: "gate & context",
				Status: domain.StatusCompleted, Role: core.ThreadTaskExternalGate,
			},
		},
		Edges: []core.ThreadGraphEdge{{From: "6g0000000002", To: "6g0000000001"}},
	}
}

func TestFormattersReplaceVisualSpoofingControls(t *testing.T) {
	projection := hostileProjection()
	projection.Nodes[0].Description = "safe\u202egnirts\u202c and a\u200bb"
	mermaid, err := Mermaid(projection)
	if err != nil {
		t.Fatal(err)
	}
	dot, err := DOT(projection)
	if err != nil {
		t.Fatal(err)
	}
	for _, forbidden := range []string{"&#8238;", "&#8236;", "&#8203;", "\u202e", "\u202c", "\u200b"} {
		if strings.Contains(mermaid, forbidden) || strings.Contains(dot, forbidden) {
			t.Fatalf("format control %q survived:\nMermaid:\n%s\nDOT:\n%s", forbidden, mermaid, dot)
		}
	}
	if strings.Count(mermaid, "&#65533;") != 3 || strings.Count(dot, "�") != 3 {
		t.Fatalf("format controls were not visibly replaced:\nMermaid:\n%s\nDOT:\n%s", mermaid, dot)
	}
}

func TestMermaidEscapesHostileLabelsAndPreservesProjectionOrder(t *testing.T) {
	got, err := Mermaid(hostileProjection())
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"flowchart TD\n",
		`n0["first &#92;&#34;task&#92;&#34;<br/>6g0000000001<br/>next-up &#183; member`,
		`line &#60;one&#62;<br/>&#37;&#37;&#123;init:x&#125;&#37;&#37;`,
		"n1 --> n0\n",
		"class n0 member\n",
		"class n1 externalGate\n",
		"subgraph legend[\"Legend\"]\n",
		"legendMember[\"Thread member<br/>blue &#183; solid border\"]\n",
		"legendExternalGate[\"External prerequisite<br/>not a Thread member &#183; amber &#183; dashed border\"]\n",
		"class legendMember member\n",
		"class legendExternalGate externalGate\n",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("Mermaid output missing %q:\n%s", want, got)
		}
	}
	if strings.Contains(got, "%%{init:x}%%") || strings.Contains(got, "line <one>") {
		t.Fatalf("hostile syntax was not neutralized:\n%s", got)
	}
	if strings.Count(got, " --> ") != len(hostileProjection().Edges) {
		t.Fatalf("legend created or hid a dependency edge:\n%s", got)
	}
}

func TestDOTEscapesHostileLabelsAndPreservesProjectionOrder(t *testing.T) {
	got, err := DOT(hostileProjection())
	if err != nil {
		t.Fatal(err)
	}
	firstNode := strings.Index(got, "n0 [")
	secondNode := strings.Index(got, "n1 [")
	edge := strings.Index(got, "n1 -> n0;")
	if firstNode < 0 || secondNode <= firstNode || edge <= secondNode {
		t.Fatalf("DOT order lost:\n%s", got)
	}
	for _, want := range []string{
		`first \\\"task\\\"\n6g0000000001`,
		`role="external-gate"`,
		`style="rounded,dashed,filled"`,
		`color="#9a6700"`,
		`fillcolor="#fff4d6"`,
		"subgraph cluster_legend {",
		`label="Thread member\nblue fill, solid border"`,
		`label="External prerequisite\nnot a Thread member\namber fill, dashed border"`,
		`role="legend"`,
	} {
		if !strings.Contains(got, want) {
			t.Errorf("DOT output missing %q:\n%s", want, got)
		}
	}
	if strings.Count(got, " -> ") != len(hostileProjection().Edges) {
		t.Fatalf("DOT legend created or hid a dependency edge:\n%s", got)
	}
}

func TestFormattersAcceptEmptyAndRejectMalformedProjection(t *testing.T) {
	wantMermaid := "flowchart TD\n" +
		"  %% graph_health=unknown projection_health=unknown topology_complete=false\n" +
		"  subgraph legend[\"Legend\"]\n" +
		"    direction LR\n" +
		"    legendMember[\"Thread member<br/>blue &#183; solid border\"]\n" +
		"    legendExternalGate[\"External prerequisite<br/>not a Thread member &#183; amber &#183; dashed border\"]\n" +
		"  end\n" +
		"  class legendMember member\n" +
		"  class legendExternalGate externalGate\n" +
		"  classDef member fill:#e8f1ff,stroke:#3267a8,stroke-width:1px\n" +
		"  classDef externalGate fill:#fff4d6,stroke:#9a6700,stroke-width:1px,stroke-dasharray:5 3\n"
	if got, err := Mermaid(core.ThreadGraphProjection{}); err != nil || got != wantMermaid {
		t.Fatalf("empty Mermaid got=%q err=%v", got, err)
	}
	wantDOT := "digraph thread {\n" +
		"  // graph_health=unknown projection_health=unknown topology_complete=false\n" +
		"  rankdir=TB;\n" +
		"  node [shape=box];\n" +
		"  subgraph cluster_legend {\n" +
		"    label=\"Legend\";\n" +
		"    legend_member [label=\"Thread member\\nblue fill, solid border\", role=\"legend\", style=\"rounded,filled\", color=\"#3267a8\", fillcolor=\"#e8f1ff\"];\n" +
		"    legend_external_gate [label=\"External prerequisite\\nnot a Thread member\\namber fill, dashed border\", role=\"legend\", style=\"rounded,dashed,filled\", color=\"#9a6700\", fillcolor=\"#fff4d6\"];\n" +
		"  }\n" +
		"}\n"
	if got, err := DOT(core.ThreadGraphProjection{}); err != nil || got != wantDOT {
		t.Fatalf("empty DOT got=%q err=%v", got, err)
	}
	malformed := hostileProjection()
	malformed.Edges[0].From = "missing"
	if _, err := Mermaid(malformed); err == nil {
		t.Fatal("dangling edge should be rejected")
	}
}

func TestFormattersKeepLegendDeterministicForMemberOnlyDegradedProjection(t *testing.T) {
	projection := core.ThreadGraphProjection{
		View: core.ThreadView{
			GraphHealth:      core.GraphDegraded,
			ProjectionHealth: core.GraphBroken,
		},
		Nodes: []core.ThreadGraphNode{{
			TaskID: "6g0000000001", Label: "member only", Status: domain.StatusInProgress,
			Role: core.ThreadTaskMember,
		}},
	}

	firstMermaid, err := Mermaid(projection)
	if err != nil {
		t.Fatal(err)
	}
	secondMermaid, err := Mermaid(projection)
	if err != nil {
		t.Fatal(err)
	}
	if firstMermaid != secondMermaid {
		t.Fatalf("Mermaid output is not deterministic:\nfirst:\n%s\nsecond:\n%s", firstMermaid, secondMermaid)
	}
	for _, want := range []string{
		"graph_health=degraded projection_health=broken topology_complete=false",
		"class n0 member",
		"subgraph legend[\"Legend\"]",
	} {
		if !strings.Contains(firstMermaid, want) {
			t.Errorf("Mermaid output missing %q:\n%s", want, firstMermaid)
		}
	}
	if strings.Count(firstMermaid, "subgraph legend[") != 1 || strings.Count(firstMermaid, " --> ") != 0 {
		t.Fatalf("Mermaid legend is unbounded or introduced an edge:\n%s", firstMermaid)
	}

	firstDOT, err := DOT(projection)
	if err != nil {
		t.Fatal(err)
	}
	secondDOT, err := DOT(projection)
	if err != nil {
		t.Fatal(err)
	}
	if firstDOT != secondDOT {
		t.Fatalf("DOT output is not deterministic:\nfirst:\n%s\nsecond:\n%s", firstDOT, secondDOT)
	}
	for _, want := range []string{
		"graph_health=degraded projection_health=broken topology_complete=false",
		`role="member"`,
		"subgraph cluster_legend {",
	} {
		if !strings.Contains(firstDOT, want) {
			t.Errorf("DOT output missing %q:\n%s", want, firstDOT)
		}
	}
	if strings.Count(firstDOT, "subgraph cluster_legend {") != 1 || strings.Count(firstDOT, " -> ") != 0 {
		t.Fatalf("DOT legend is unbounded or introduced an edge:\n%s", firstDOT)
	}
}
