// Package graphfmt contains pure textual output adapters for the neutral core
// Thread graph projection. It deliberately owns format-specific escaping and has
// no CLI, TUI, HTTP, filesystem, styling, or graph-library dependencies.
package graphfmt

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/andy-esch/taskflow/internal/core"
)

type preparedProjection struct {
	nodeNames map[string]string
	labels    []string
}

// RenderOptions controls presentation-only graph detail. It deliberately does
// not alter the adapter-neutral projection or its JSON representation.
type RenderOptions struct {
	IncludeDescriptions bool
}

const (
	maxNodeLabelRunes       = 64
	maxNodeDescriptionRunes = 120
)

// Mermaid renders a deterministic top-to-bottom Mermaid flowchart. Synthetic
// node names keep task IDs and labels out of Mermaid identifier syntax.
func Mermaid(projection core.ThreadGraphProjection) (string, error) {
	return MermaidWithOptions(projection, RenderOptions{})
}

// MermaidWithOptions renders Mermaid with explicit presentation detail.
func MermaidWithOptions(projection core.ThreadGraphProjection, options RenderOptions) (string, error) {
	prepared, err := prepare(projection, options)
	if err != nil {
		return "", err
	}
	var out strings.Builder
	out.WriteString("flowchart TD\n")
	fmt.Fprintf(&out, "  %%%% graph_health=%s projection_health=%s topology_complete=%t\n",
		healthToken(projection.View.GraphHealth), healthToken(projection.View.ProjectionHealth), projection.TopologyComplete)
	for index := range projection.Nodes {
		fmt.Fprintf(&out, "  n%d[\"%s\"]\n", index, escapeMermaid(prepared.labels[index]))
	}
	for _, edge := range projection.Edges {
		fmt.Fprintf(&out, "  %s --> %s\n", prepared.nodeNames[edge.From], prepared.nodeNames[edge.To])
	}
	for index, node := range projection.Nodes {
		className := "member"
		if node.Role == core.ThreadTaskExternalGate {
			className = "externalGate"
		}
		fmt.Fprintf(&out, "  class n%d %s\n", index, className)
	}
	fmt.Fprintf(&out, "  subgraph legend[\"%s\"]\n", mermaidLegendTitle(options))
	out.WriteString("    direction LR\n")
	out.WriteString("    legendMember[\"Thread member<br/>blue &#183; solid border\"]\n")
	out.WriteString("    legendExternalGate[\"External prerequisite<br/>not a Thread member &#183; amber &#183; dashed border\"]\n")
	out.WriteString("  end\n")
	out.WriteString("  class legendMember member\n")
	out.WriteString("  class legendExternalGate externalGate\n")
	out.WriteString("  classDef member fill:#e8f1ff,stroke:#3267a8,stroke-width:1px\n")
	out.WriteString("  classDef externalGate fill:#fff4d6,stroke:#9a6700,stroke-width:1px,stroke-dasharray:5 3\n")
	return out.String(), nil
}

// DOT renders a deterministic Graphviz directed graph. Custom task_id and role
// attributes make the semantic distinction available to downstream DOT tooling
// without asking it to parse the visible label.
func DOT(projection core.ThreadGraphProjection) (string, error) {
	return DOTWithOptions(projection, RenderOptions{})
}

// DOTWithOptions renders Graphviz DOT with explicit presentation detail.
func DOTWithOptions(projection core.ThreadGraphProjection, options RenderOptions) (string, error) {
	prepared, err := prepare(projection, options)
	if err != nil {
		return "", err
	}
	var out strings.Builder
	out.WriteString("digraph thread {\n")
	fmt.Fprintf(&out, "  // graph_health=%s projection_health=%s topology_complete=%t\n",
		healthToken(projection.View.GraphHealth), healthToken(projection.View.ProjectionHealth), projection.TopologyComplete)
	out.WriteString("  rankdir=TB;\n")
	out.WriteString("  node [shape=box];\n")
	for index, node := range projection.Nodes {
		style, color, fillColor := "rounded,filled", "#3267a8", "#e8f1ff"
		if node.Role == core.ThreadTaskExternalGate {
			style, color, fillColor = "rounded,dashed,filled", "#9a6700", "#fff4d6"
		}
		fmt.Fprintf(&out, "  n%d [label=%s, task_id=%s, role=%s, style=%s, color=%s, fillcolor=%s];\n",
			index, quoteDOT(prepared.labels[index]), quoteDOT(node.TaskID), quoteDOT(string(node.Role)),
			quoteDOT(style), quoteDOT(color), quoteDOT(fillColor))
	}
	for _, edge := range projection.Edges {
		fmt.Fprintf(&out, "  %s -> %s;\n", prepared.nodeNames[edge.From], prepared.nodeNames[edge.To])
	}
	out.WriteString("  subgraph cluster_legend {\n")
	fmt.Fprintf(&out, "    label=%s;\n", quoteDOT(dotLegendTitle(options)))
	out.WriteString("    legend_member [label=\"Thread member\\nblue fill, solid border\", role=\"legend\", style=\"rounded,filled\", color=\"#3267a8\", fillcolor=\"#e8f1ff\"];\n")
	out.WriteString("    legend_external_gate [label=\"External prerequisite\\nnot a Thread member\\namber fill, dashed border\", role=\"legend\", style=\"rounded,dashed,filled\", color=\"#9a6700\", fillcolor=\"#fff4d6\"];\n")
	out.WriteString("  }\n")
	out.WriteString("}\n")
	return out.String(), nil
}

func healthToken(health core.GraphHealth) string {
	switch health {
	case core.GraphHealthy, core.GraphDegraded, core.GraphBroken:
		return string(health)
	default:
		return "unknown"
	}
}

func prepare(projection core.ThreadGraphProjection, options RenderOptions) (preparedProjection, error) {
	prepared := preparedProjection{
		nodeNames: make(map[string]string, len(projection.Nodes)),
		labels:    make([]string, len(projection.Nodes)),
	}
	for index, node := range projection.Nodes {
		if node.TaskID == "" {
			return preparedProjection{}, fmt.Errorf("thread graph node %d has an empty task ID", index)
		}
		if _, duplicate := prepared.nodeNames[node.TaskID]; duplicate {
			return preparedProjection{}, fmt.Errorf("thread graph repeats task ID %s", node.TaskID)
		}
		if node.Role != core.ThreadTaskMember && node.Role != core.ThreadTaskExternalGate {
			return preparedProjection{}, fmt.Errorf("thread graph task %s has unknown role %q", node.TaskID, node.Role)
		}
		prepared.nodeNames[node.TaskID] = fmt.Sprintf("n%d", index)
		prepared.labels[index] = nodeLabel(node, options)
	}
	seenEdges := make(map[core.ThreadGraphEdge]bool, len(projection.Edges))
	for _, edge := range projection.Edges {
		if prepared.nodeNames[edge.From] == "" || prepared.nodeNames[edge.To] == "" {
			return preparedProjection{}, fmt.Errorf("thread graph edge %s -> %s has an unknown endpoint", edge.From, edge.To)
		}
		if seenEdges[edge] {
			return preparedProjection{}, fmt.Errorf("thread graph repeats edge %s -> %s", edge.From, edge.To)
		}
		seenEdges[edge] = true
	}
	return prepared, nil
}

func nodeLabel(node core.ThreadGraphNode, options RenderOptions) string {
	label := compactText(node.Title, maxNodeLabelRunes)
	if label == "" {
		label = compactText(strings.ReplaceAll(node.Label, "-", " "), maxNodeLabelRunes)
		if label == "" {
			label = "Task"
		}
		labelRunes := []rune(label)
		labelRunes[0] = unicode.ToUpper(labelRunes[0])
		label = string(labelRunes)
	}
	parts := []string{label}

	metadata := roleLabel(node.Role)
	if node.Status != "" {
		metadata = string(node.Status) + " · " + metadata
	}
	if metadata != "" {
		parts = append(parts, metadata)
	}
	if options.IncludeDescriptions {
		if description := compactText(node.Description, maxNodeDescriptionRunes); description != "" {
			parts = append(parts, description)
		}
	}
	parts = append(parts, "ID "+node.TaskID)
	return strings.Join(parts, "\n")
}

func roleLabel(role core.ThreadTaskRole) string {
	if role == core.ThreadTaskExternalGate {
		return "external prerequisite"
	}
	if role == core.ThreadTaskMember {
		return "Thread member"
	}
	return string(role)
}

func compactText(value string, limit int) string {
	value = strings.Join(strings.Fields(value), " ")
	runes := []rune(value)
	if len(runes) <= limit {
		return value
	}
	cutRunes := runes[:limit-1]
	lastSpace := -1
	for index := len(cutRunes) - 1; index >= 0; index-- {
		if cutRunes[index] == ' ' {
			lastSpace = index
			break
		}
	}
	if lastSpace >= (limit-1)*2/3 {
		cutRunes = cutRunes[:lastSpace]
	}
	return strings.TrimSpace(string(cutRunes)) + "…"
}

func mermaidLegendTitle(options RenderOptions) string {
	if options.IncludeDescriptions {
		return "Legend &#183; detailed labels &#183; descriptions included"
	}
	return "Legend &#183; compact labels &#183; add --details for descriptions"
}

func dotLegendTitle(options RenderOptions) string {
	if options.IncludeDescriptions {
		return "Legend\nDetailed labels; descriptions included"
	}
	return "Legend\nCompact labels; add --details for descriptions"
}

// Mermaid quoted labels still interpret HTML and flowchart punctuation. Keep a
// small readable set and encode every other rune as a numeric entity; newlines
// become renderer-owned line breaks. This also neutralizes directive-like input.
func escapeMermaid(value string) string {
	var out strings.Builder
	for _, r := range value {
		switch {
		case r == '\n':
			out.WriteString("<br/>")
		case unicode.IsLetter(r), unicode.IsDigit(r), strings.ContainsRune(" -_./:()", r):
			out.WriteRune(r)
		case unsafeFormatRune(r):
			out.WriteString("&#65533;")
		default:
			fmt.Fprintf(&out, "&#%d;", r)
		}
	}
	return out.String()
}

func quoteDOT(value string) string {
	var out strings.Builder
	out.WriteByte('"')
	for _, r := range value {
		switch r {
		case '\\':
			out.WriteString(`\\`)
		case '"':
			out.WriteString(`\"`)
		case '\n':
			out.WriteString(`\n`)
		case '\r':
			out.WriteString(`\r`)
		case '\t':
			out.WriteString(`\t`)
		default:
			if unsafeFormatRune(r) {
				out.WriteRune('\uFFFD')
			} else {
				out.WriteRune(r)
			}
		}
	}
	out.WriteByte('"')
	return out.String()
}

// Unicode format controls include bidi overrides and zero-width characters.
// Encoding them as entities would preserve their visual-spoofing behavior after
// an HTML-capable renderer decodes the label, so replace them like C0 controls.
func unsafeFormatRune(r rune) bool {
	return unicode.IsControl(r) || unicode.Is(unicode.Cf, r)
}
