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
	options := RenderOptions{IncludeDescriptions: true}
	mermaid, err := MermaidWithOptions(projection, options)
	if err != nil {
		t.Fatal(err)
	}
	dot, err := DOTWithOptions(projection, options)
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
	got, err := MermaidWithOptions(hostileProjection(), RenderOptions{IncludeDescriptions: true})
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"flowchart TD\n",
		`n0["First &#92;&#34;task&#92;&#34;<br/>next-up &#183; Thread member<br/>line &#60;one&#62; &#37;&#37;&#123;init:x&#125;&#37;&#37;<br/>ID 6g0000000001`,
		"n1 --> n0\n",
		"class n0 member\n",
		"class n1 externalGate\n",
		"subgraph legend[\"Legend &#183; detailed labels &#183; descriptions included\"]\n",
		"legendMember[\"Thread member<br/>blue &#183; solid border\"]\n",
		"legendExternalGate[\"External prerequisite<br/>not a Thread member &#183; amber &#183; dashed border\"]\n",
		"class legendMember member\n",
		"class legendExternalGate externalGate\n",
		"classDef member fill:#e8f1ff,color:#172033,stroke:#3267a8",
		"classDef externalGate fill:#fff4d6,color:#3d2c00,stroke:#9a6700",
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
	got, err := DOTWithOptions(hostileProjection(), RenderOptions{IncludeDescriptions: true})
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
		`First \\\"task\\\"\nnext-up · Thread member\nline <one> %%{init:x}%%\nID 6g0000000001`,
		`role="external-gate"`,
		`style="rounded,dashed,filled"`,
		`color="#9a6700"`,
		`fillcolor="#fff4d6"`,
		"subgraph cluster_legend {",
		`label="Legend\nDetailed labels; descriptions included"`,
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

func TestNodeLabelsPrioritizeReadableIdentityAndBoundOptionalDetails(t *testing.T) {
	node := core.ThreadGraphNode{
		TaskID: "6g0000000001", Label: "make-thread-graph-nodes-human-readable-at-review-scale",
		Description: strings.Repeat("detail ", 40), Status: domain.StatusCompleted,
		Role: core.ThreadTaskExternalGate,
	}
	compact := nodeLabel(node, RenderOptions{})
	wantCompact := "Make thread graph nodes human readable at review scale\n" +
		"completed · external prerequisite\nID 6g0000000001"
	if compact != wantCompact {
		t.Fatalf("compact label:\n%s\nwant:\n%s", compact, wantCompact)
	}
	if strings.Contains(compact, "detail") || strings.Contains(compact, "external-gate") {
		t.Fatalf("compact label leaked verbose or misleading text: %q", compact)
	}
	node.Title = "Make TUI Thread graphs human-readable"
	if got := strings.Split(nodeLabel(node, RenderOptions{}), "\n")[0]; got != node.Title {
		t.Fatalf("body-derived title %q was not preferred, got %q", node.Title, got)
	}
	node.Title = "iOS and eBPF support"
	if got := strings.Split(nodeLabel(node, RenderOptions{}), "\n")[0]; got != node.Title {
		t.Fatalf("authored title casing changed from %q to %q", node.Title, got)
	}
	node.Title = ""

	detailed := nodeLabel(node, RenderOptions{IncludeDescriptions: true})
	lines := strings.Split(detailed, "\n")
	if len(lines) != 4 {
		t.Fatalf("detailed label has %d lines, want 4: %q", len(lines), detailed)
	}
	descriptionRunes := len([]rune(lines[2]))
	if descriptionRunes > maxNodeDescriptionRunes ||
		descriptionRunes < maxNodeDescriptionRunes-8 || !strings.HasSuffix(lines[2], "…") {
		t.Fatalf("detailed description was not visibly and deterministically bounded: %q", detailed)
	}
	if lines[3] != "ID 6g0000000001" {
		t.Fatalf("stable ID is not the final label line: %q", detailed)
	}

	longNode := node
	longNode.Label = strings.Repeat("four-", 30)
	firstLine := strings.Split(nodeLabel(longNode, RenderOptions{}), "\n")[0]
	if len([]rune(firstLine)) > maxNodeLabelRunes || len([]rune(firstLine)) < maxNodeLabelRunes-8 ||
		!strings.HasSuffix(firstLine, "four…") {
		t.Fatalf("long label was not visibly and deterministically bounded: %q", firstLine)
	}
	longNode.Label = ""
	longNode.Title = strings.Repeat("界", maxNodeLabelRunes+10)
	firstLine = strings.Split(nodeLabel(longNode, RenderOptions{}), "\n")[0]
	if len([]rune(firstLine)) != maxNodeLabelRunes || !strings.HasSuffix(firstLine, "…") {
		t.Fatalf("Unicode title was not bounded on a rune boundary: %q", firstLine)
	}
	longNode.Title = strings.Repeat("界", 20) + " " + strings.Repeat("界", 45)
	firstLine = strings.Split(nodeLabel(longNode, RenderOptions{}), "\n")[0]
	if len([]rune(firstLine)) != maxNodeLabelRunes || strings.HasPrefix(firstLine, strings.Repeat("界", 20)+"…") {
		t.Fatalf("Unicode word-boundary truncation compared byte and rune offsets: %q", firstLine)
	}
}

func TestCompactFormattersDisambiguateDuplicateLabelsWithoutDescriptions(t *testing.T) {
	projection := hostileProjection()
	projection.Nodes[0].Label = "same-task"
	projection.Nodes[1].Label = "same-task"

	mermaid, err := Mermaid(projection)
	if err != nil {
		t.Fatal(err)
	}
	dot, err := DOT(projection)
	if err != nil {
		t.Fatal(err)
	}
	for _, output := range []string{mermaid, dot} {
		for _, taskID := range []string{"6g0000000001", "6g0000000002"} {
			if !strings.Contains(output, taskID) {
				t.Errorf("duplicate label lost stable identity %s:\n%s", taskID, output)
			}
		}
		for _, description := range []string{"line <one>", "gate & context"} {
			if strings.Contains(output, description) {
				t.Errorf("compact output leaked description %q:\n%s", description, output)
			}
		}
	}
}

func TestFormattersAcceptEmptyAndRejectMalformedProjection(t *testing.T) {
	wantMermaid := "flowchart TD\n" +
		"  %% graph_health=unknown projection_health=unknown topology_complete=false\n" +
		"  subgraph legend[\"Legend &#183; compact labels &#183; add --details for descriptions\"]\n" +
		"    direction LR\n" +
		"    legendMember[\"Thread member<br/>blue &#183; solid border\"]\n" +
		"    legendExternalGate[\"External prerequisite<br/>not a Thread member &#183; amber &#183; dashed border\"]\n" +
		"  end\n" +
		"  class legendMember member\n" +
		"  class legendExternalGate externalGate\n" +
		"  classDef member fill:#e8f1ff,color:#172033,stroke:#3267a8,stroke-width:1px\n" +
		"  classDef externalGate fill:#fff4d6,color:#3d2c00,stroke:#9a6700,stroke-width:1px,stroke-dasharray:5 3\n"
	if got, err := Mermaid(core.ThreadGraphProjection{}); err != nil || got != wantMermaid {
		t.Fatalf("empty Mermaid got=%q err=%v", got, err)
	}
	wantDOT := "digraph thread {\n" +
		"  // graph_health=unknown projection_health=unknown topology_complete=false\n" +
		"  rankdir=TB;\n" +
		"  node [shape=box];\n" +
		"  subgraph cluster_legend {\n" +
		"    label=\"Legend\\nCompact labels; add --details for descriptions\";\n" +
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
		"subgraph legend[\"Legend &#183; compact labels &#183; add --details for descriptions\"]",
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

func TestBoundedFormattersDiscloseScopeFocusAndBoundaryContinuations(t *testing.T) {
	projection := core.ThreadGraphProjection{
		View: core.ThreadView{GraphHealth: core.GraphDegraded, ProjectionHealth: core.GraphBroken},
		Nodes: []core.ThreadGraphNode{
			{TaskID: "6g0000000002", Label: "focus <task>", Status: domain.StatusInProgress, Role: core.ThreadTaskMember},
			{TaskID: "6g0000000003", Label: "shown-child", Status: domain.StatusNextUp, Role: core.ThreadTaskMember},
		},
		Edges: []core.ThreadGraphEdge{{From: "6g0000000002", To: "6g0000000003"}},
		Scope: &core.ThreadGraphScope{
			Kind: core.ThreadGraphScopeNeighborhood, FocalTaskID: "6g0000000002", Depth: 1,
			TotalNodes: 4, ShownNodes: 2, HiddenNodes: 2, TotalEdges: 3, ShownEdges: 1, HiddenEdges: 2,
			BoundaryEdges: []core.ThreadGraphEdge{
				{From: "6g0000000001", To: "6g0000000002"},
				{From: "6g0000000003", To: "6g0000000004"},
			},
		},
	}

	mermaid, err := Mermaid(projection)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"scope=bounded Thread neighborhood focus=6g0000000002 depth=1 shown_nodes=2 total_nodes=4 hidden_nodes=2",
		"Bounded Thread neighborhood &#183; 1 hop &#183; 2/4 nodes shown &#183; 2 hidden",
		`n0["Focus &#60;task&#62;`, "class n0 memberFocus", "stroke:#c026d3",
		`boundary0["&#8230; 1 omitted prerequisite edge"]`, "boundary0 -.-> n0",
		`boundary1["&#8230; 1 omitted dependent edge"]`, "n1 -.-> boundary1",
		"Omitted continuation<br/>summary marker &#183; dotted edge",
	} {
		if !strings.Contains(mermaid, want) {
			t.Errorf("bounded Mermaid missing %q:\n%s", want, mermaid)
		}
	}
	for _, hiddenID := range []string{"6g0000000001", "6g0000000004"} {
		if strings.Contains(mermaid, hiddenID) {
			t.Errorf("bounded Mermaid leaked omitted task identity %s:\n%s", hiddenID, mermaid)
		}
	}

	dot, err := DOT(projection)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"scope=bounded Thread neighborhood focus=6g0000000002 depth=1",
		`label="Bounded Thread neighborhood · 1 hop · 2/4 nodes shown · 2 hidden"`,
		`focal="true"`, `color="#c026d3"`, `penwidth=3`,
		`label="… 1 omitted prerequisite edge"`, `boundary0 -> n0 [style="dotted"]`,
		`label="… 1 omitted dependent edge"`, `n1 -> boundary1 [style="dotted"]`,
	} {
		if !strings.Contains(dot, want) {
			t.Errorf("bounded DOT missing %q:\n%s", want, dot)
		}
	}
}

func TestBoundedFormattersRejectDishonestScopeMetadata(t *testing.T) {
	projection := core.ThreadGraphProjection{
		Nodes: []core.ThreadGraphNode{{TaskID: "6g0000000001", Role: core.ThreadTaskMember}},
		Scope: &core.ThreadGraphScope{
			Kind: core.ThreadGraphScopeNeighborhood, FocalTaskID: "6g0000000001", Depth: 1,
			TotalNodes: 2, ShownNodes: 1, HiddenNodes: 1, TotalEdges: 1, HiddenEdges: 1,
			BoundaryEdges: []core.ThreadGraphEdge{{From: "", To: "6g0000000001"}},
		},
	}
	if _, err := Mermaid(projection); err == nil || !strings.Contains(err.Error(), "empty endpoint") {
		t.Fatalf("invalid boundary error=%v", err)
	}
	projection.Scope.BoundaryEdges = nil
	projection.Scope.ShownNodes = 2
	if _, err := DOT(projection); err == nil || !strings.Contains(err.Error(), "scope counts") {
		t.Fatalf("invalid count error=%v", err)
	}
}
