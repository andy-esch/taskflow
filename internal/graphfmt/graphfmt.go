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

// Mermaid renders a deterministic top-to-bottom Mermaid flowchart. Synthetic
// node names keep task IDs and labels out of Mermaid identifier syntax.
func Mermaid(projection core.ThreadGraphProjection) (string, error) {
	prepared, err := prepare(projection)
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
	out.WriteString("  classDef member fill:#e8f1ff,stroke:#3267a8,stroke-width:1px\n")
	out.WriteString("  classDef externalGate fill:#fff4d6,stroke:#9a6700,stroke-width:1px,stroke-dasharray:5 3\n")
	return out.String(), nil
}

// DOT renders a deterministic Graphviz directed graph. Custom task_id and role
// attributes make the semantic distinction available to downstream DOT tooling
// without asking it to parse the visible label.
func DOT(projection core.ThreadGraphProjection) (string, error) {
	prepared, err := prepare(projection)
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
		style := "rounded"
		if node.Role == core.ThreadTaskExternalGate {
			style = "rounded,dashed"
		}
		fmt.Fprintf(&out, "  n%d [label=%s, task_id=%s, role=%s, style=%s];\n",
			index, quoteDOT(prepared.labels[index]), quoteDOT(node.TaskID), quoteDOT(string(node.Role)), quoteDOT(style))
	}
	for _, edge := range projection.Edges {
		fmt.Fprintf(&out, "  %s -> %s;\n", prepared.nodeNames[edge.From], prepared.nodeNames[edge.To])
	}
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

func prepare(projection core.ThreadGraphProjection) (preparedProjection, error) {
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
		prepared.labels[index] = nodeLabel(node)
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

func nodeLabel(node core.ThreadGraphNode) string {
	label := node.Label
	if label == "" {
		label = node.TaskID
	}
	parts := []string{label, node.TaskID}
	metadata := string(node.Role)
	if node.Status != "" {
		metadata = string(node.Status) + " · " + metadata
	}
	if metadata != "" {
		parts = append(parts, metadata)
	}
	if node.Description != "" {
		parts = append(parts, node.Description)
	}
	return strings.Join(parts, "\n")
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
