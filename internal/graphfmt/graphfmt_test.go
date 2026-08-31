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
	} {
		if !strings.Contains(got, want) {
			t.Errorf("Mermaid output missing %q:\n%s", want, got)
		}
	}
	if strings.Contains(got, "%%{init:x}%%") || strings.Contains(got, "line <one>") {
		t.Fatalf("hostile syntax was not neutralized:\n%s", got)
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
	for _, want := range []string{`first \\\"task\\\"\n6g0000000001`, `role="external-gate"`, `style="rounded,dashed"`} {
		if !strings.Contains(got, want) {
			t.Errorf("DOT output missing %q:\n%s", want, got)
		}
	}
}

func TestFormattersAcceptEmptyAndRejectMalformedProjection(t *testing.T) {
	if got, err := Mermaid(core.ThreadGraphProjection{}); err != nil || !strings.HasPrefix(got, "flowchart TD\n") {
		t.Fatalf("empty Mermaid got=%q err=%v", got, err)
	}
	if got, err := DOT(core.ThreadGraphProjection{}); err != nil || got != "digraph thread {\n  // graph_health=unknown projection_health=unknown topology_complete=false\n  rankdir=TB;\n  node [shape=box];\n}\n" {
		t.Fatalf("empty DOT got=%q err=%v", got, err)
	}
	malformed := hostileProjection()
	malformed.Edges[0].From = "missing"
	if _, err := Mermaid(malformed); err == nil {
		t.Fatal("dangling edge should be rejected")
	}
}
