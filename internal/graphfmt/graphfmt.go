// Package graphfmt contains pure textual output adapters for the neutral core
// Thread graph projection. It deliberately owns format-specific escaping and has
// no CLI, TUI, HTTP, filesystem, styling, or graph-library dependencies.
package graphfmt

import (
	"fmt"
	"sort"
	"strings"
	"unicode"

	"github.com/andy-esch/taskflow/internal/core"
)

type preparedProjection struct {
	nodeNames map[string]string
	labels    []string
	boundary  []boundaryMarker
}

type boundaryMarker struct {
	visibleTaskID string
	incoming      bool
	count         int
}

type boundaryKey struct {
	visibleTaskID string
	incoming      bool
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
	if projection.Scope != nil {
		fmt.Fprintf(&out, "  %%%% scope=bounded Thread neighborhood focus=%s depth=%d shown_nodes=%d total_nodes=%d hidden_nodes=%d shown_edges=%d total_edges=%d hidden_edges=%d boundary_edges=%d\n",
			escapeMermaid(projection.Scope.FocalTaskID), projection.Scope.Depth,
			projection.Scope.ShownNodes, projection.Scope.TotalNodes, projection.Scope.HiddenNodes,
			projection.Scope.ShownEdges, projection.Scope.TotalEdges, projection.Scope.HiddenEdges, len(projection.Scope.BoundaryEdges))
	}
	for index := range projection.Nodes {
		fmt.Fprintf(&out, "  n%d[\"%s\"]\n", index, escapeMermaid(prepared.labels[index]))
	}
	for _, edge := range projection.Edges {
		fmt.Fprintf(&out, "  %s --> %s\n", prepared.nodeNames[edge.From], prepared.nodeNames[edge.To])
	}
	for index, marker := range prepared.boundary {
		name := fmt.Sprintf("boundary%d", index)
		label := boundaryLabel(marker)
		fmt.Fprintf(&out, "  %s[\"%s\"]\n", name, escapeMermaid(label))
		if marker.incoming {
			fmt.Fprintf(&out, "  %s -.-> %s\n", name, prepared.nodeNames[marker.visibleTaskID])
		} else {
			fmt.Fprintf(&out, "  %s -.-> %s\n", prepared.nodeNames[marker.visibleTaskID], name)
		}
		fmt.Fprintf(&out, "  class %s boundary\n", name)
	}
	for index, node := range projection.Nodes {
		className := "member"
		if node.Role == core.ThreadTaskExternalGate {
			className = "externalGate"
		}
		if projection.Scope != nil && node.TaskID == projection.Scope.FocalTaskID {
			if node.Role == core.ThreadTaskExternalGate {
				className = "externalGateFocus"
			} else {
				className = "memberFocus"
			}
		}
		fmt.Fprintf(&out, "  class n%d %s\n", index, className)
	}
	fmt.Fprintf(&out, "  subgraph legend[\"%s\"]\n", mermaidLegendTitle(options, projection.Scope))
	out.WriteString("    direction LR\n")
	out.WriteString("    legendMember[\"Thread member<br/>blue &#183; solid border\"]\n")
	out.WriteString("    legendExternalGate[\"External prerequisite<br/>not a Thread member &#183; amber &#183; dashed border\"]\n")
	if projection.Scope != nil {
		out.WriteString("    legendFocus[\"Focal task<br/>magenta outline &#183; role fill retained\"]\n")
	}
	if len(prepared.boundary) > 0 {
		out.WriteString("    legendBoundary[\"Omitted continuation<br/>summary marker &#183; dotted edge\"]\n")
	}
	out.WriteString("  end\n")
	out.WriteString("  class legendMember member\n")
	out.WriteString("  class legendExternalGate externalGate\n")
	if projection.Scope != nil {
		out.WriteString("  class legendFocus memberFocus\n")
	}
	if len(prepared.boundary) > 0 {
		out.WriteString("  class legendBoundary boundary\n")
	}
	// Explicit foregrounds keep pale semantic cards legible when a Mermaid
	// host (notably GitHub dark mode) otherwise inherits its page text color.
	out.WriteString("  classDef member fill:#e8f1ff,color:#172033,stroke:#3267a8,stroke-width:1px\n")
	out.WriteString("  classDef externalGate fill:#fff4d6,color:#3d2c00,stroke:#9a6700,stroke-width:1px,stroke-dasharray:5 3\n")
	if projection.Scope != nil {
		out.WriteString("  classDef memberFocus fill:#e8f1ff,color:#172033,stroke:#c026d3,stroke-width:3px\n")
		out.WriteString("  classDef externalGateFocus fill:#fff4d6,color:#3d2c00,stroke:#c026d3,stroke-width:3px,stroke-dasharray:5 3\n")
		if len(prepared.boundary) > 0 {
			out.WriteString("  classDef boundary fill:#f4f4f5,color:#3f3f46,stroke:#71717a,stroke-width:1px,stroke-dasharray:3 3\n")
		}
	}
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
	if projection.Scope != nil {
		fmt.Fprintf(&out, "  // scope=bounded Thread neighborhood focus=%s depth=%d shown_nodes=%d total_nodes=%d hidden_nodes=%d shown_edges=%d total_edges=%d hidden_edges=%d boundary_edges=%d\n",
			projection.Scope.FocalTaskID, projection.Scope.Depth,
			projection.Scope.ShownNodes, projection.Scope.TotalNodes, projection.Scope.HiddenNodes,
			projection.Scope.ShownEdges, projection.Scope.TotalEdges, projection.Scope.HiddenEdges, len(projection.Scope.BoundaryEdges))
		fmt.Fprintf(&out, "  label=%s;\n", quoteDOT(scopeTitle(projection.Scope)))
		out.WriteString("  labelloc=\"t\";\n")
	}
	out.WriteString("  rankdir=TB;\n")
	out.WriteString("  node [shape=box];\n")
	for index, node := range projection.Nodes {
		style, color, fillColor := "rounded,filled", "#3267a8", "#e8f1ff"
		if node.Role == core.ThreadTaskExternalGate {
			style, color, fillColor = "rounded,dashed,filled", "#9a6700", "#fff4d6"
		}
		focal := false
		if projection.Scope != nil && node.TaskID == projection.Scope.FocalTaskID {
			focal, color = true, "#c026d3"
		}
		if focal {
			fmt.Fprintf(&out, "  n%d [label=%s, task_id=%s, role=%s, focal=\"true\", style=%s, color=%s, fillcolor=%s, penwidth=3];\n",
				index, quoteDOT(prepared.labels[index]), quoteDOT(node.TaskID), quoteDOT(string(node.Role)),
				quoteDOT(style), quoteDOT(color), quoteDOT(fillColor))
		} else {
			fmt.Fprintf(&out, "  n%d [label=%s, task_id=%s, role=%s, style=%s, color=%s, fillcolor=%s];\n",
				index, quoteDOT(prepared.labels[index]), quoteDOT(node.TaskID), quoteDOT(string(node.Role)),
				quoteDOT(style), quoteDOT(color), quoteDOT(fillColor))
		}
	}
	for _, edge := range projection.Edges {
		fmt.Fprintf(&out, "  %s -> %s;\n", prepared.nodeNames[edge.From], prepared.nodeNames[edge.To])
	}
	for index, marker := range prepared.boundary {
		name := fmt.Sprintf("boundary%d", index)
		fmt.Fprintf(&out, "  %s [label=%s, role=\"boundary-summary\", shape=box, style=\"rounded,dashed,filled\", color=\"#71717a\", fillcolor=\"#f4f4f5\", fontcolor=\"#3f3f46\"];\n",
			name, quoteDOT(boundaryLabel(marker)))
		if marker.incoming {
			fmt.Fprintf(&out, "  %s -> %s [style=\"dotted\"];\n", name, prepared.nodeNames[marker.visibleTaskID])
		} else {
			fmt.Fprintf(&out, "  %s -> %s [style=\"dotted\"];\n", prepared.nodeNames[marker.visibleTaskID], name)
		}
	}
	out.WriteString("  subgraph cluster_legend {\n")
	fmt.Fprintf(&out, "    label=%s;\n", quoteDOT(dotLegendTitle(options, projection.Scope)))
	out.WriteString("    legend_member [label=\"Thread member\\nblue fill, solid border\", role=\"legend\", style=\"rounded,filled\", color=\"#3267a8\", fillcolor=\"#e8f1ff\"];\n")
	out.WriteString("    legend_external_gate [label=\"External prerequisite\\nnot a Thread member\\namber fill, dashed border\", role=\"legend\", style=\"rounded,dashed,filled\", color=\"#9a6700\", fillcolor=\"#fff4d6\"];\n")
	if projection.Scope != nil {
		out.WriteString("    legend_focus [label=\"Focal task\\nmagenta outline, role fill retained\", role=\"legend\", style=\"rounded,filled\", color=\"#c026d3\", fillcolor=\"#e8f1ff\", penwidth=3];\n")
	}
	if len(prepared.boundary) > 0 {
		out.WriteString("    legend_boundary [label=\"Omitted continuation\\nsummary marker, dotted edge\", role=\"legend\", style=\"rounded,dashed,filled\", color=\"#71717a\", fillcolor=\"#f4f4f5\"];\n")
	}
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
	if projection.Scope != nil {
		scope := projection.Scope
		if scope.Kind != core.ThreadGraphScopeNeighborhood {
			return preparedProjection{}, fmt.Errorf("thread graph has unknown scope kind %q", scope.Kind)
		}
		if scope.Depth != 1 && scope.Depth != 2 {
			return preparedProjection{}, fmt.Errorf("thread graph neighborhood has invalid depth %d", scope.Depth)
		}
		if prepared.nodeNames[scope.FocalTaskID] == "" {
			return preparedProjection{}, fmt.Errorf("thread graph neighborhood focal task %s is not shown", scope.FocalTaskID)
		}
		if scope.ShownNodes != len(projection.Nodes) || scope.TotalNodes < scope.ShownNodes ||
			scope.HiddenNodes != scope.TotalNodes-scope.ShownNodes || scope.ShownEdges != len(projection.Edges) ||
			scope.TotalEdges < scope.ShownEdges || scope.HiddenEdges != scope.TotalEdges-scope.ShownEdges ||
			len(scope.BoundaryEdges) > scope.HiddenEdges {
			return preparedProjection{}, fmt.Errorf("thread graph neighborhood scope counts do not match its projection")
		}
		boundarySeen := make(map[core.ThreadGraphEdge]bool, len(scope.BoundaryEdges))
		groupCounts := make(map[boundaryKey]int)
		for _, edge := range scope.BoundaryEdges {
			if edge.From == "" || edge.To == "" {
				return preparedProjection{}, fmt.Errorf("thread graph neighborhood has a boundary edge with an empty endpoint")
			}
			if boundarySeen[edge] {
				return preparedProjection{}, fmt.Errorf("thread graph neighborhood repeats boundary edge %s -> %s", edge.From, edge.To)
			}
			boundarySeen[edge] = true
			_, fromShown := prepared.nodeNames[edge.From]
			_, toShown := prepared.nodeNames[edge.To]
			if fromShown == toShown {
				return preparedProjection{}, fmt.Errorf("thread graph boundary edge %s -> %s must have exactly one shown endpoint", edge.From, edge.To)
			}
			marker := boundaryMarker{visibleTaskID: edge.From}
			if toShown {
				marker.visibleTaskID, marker.incoming = edge.To, true
			}
			key := boundaryKey{visibleTaskID: marker.visibleTaskID, incoming: marker.incoming}
			groupCounts[key]++
		}
		keys := make([]boundaryKey, 0, len(groupCounts))
		for key := range groupCounts {
			keys = append(keys, key)
		}
		sort.Slice(keys, func(i, j int) bool {
			if keys[i].visibleTaskID != keys[j].visibleTaskID {
				return keys[i].visibleTaskID < keys[j].visibleTaskID
			}
			return keys[i].incoming && !keys[j].incoming
		})
		for _, key := range keys {
			marker := boundaryMarker{visibleTaskID: key.visibleTaskID, incoming: key.incoming}
			marker.count = groupCounts[key]
			prepared.boundary = append(prepared.boundary, marker)
		}
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

func mermaidLegendTitle(options RenderOptions, scope *core.ThreadGraphScope) string {
	if scope != nil {
		detail := "compact labels"
		if options.IncludeDescriptions {
			detail = "detailed labels"
		}
		return escapeMermaid(scopeTitle(scope) + " · " + detail)
	}
	if options.IncludeDescriptions {
		return "Legend &#183; detailed labels &#183; descriptions included"
	}
	return "Legend &#183; compact labels &#183; add --details for descriptions"
}

func dotLegendTitle(options RenderOptions, scope *core.ThreadGraphScope) string {
	if scope != nil {
		detail := "Compact labels"
		if options.IncludeDescriptions {
			detail = "Detailed labels"
		}
		return scopeTitle(scope) + "\n" + detail
	}
	if options.IncludeDescriptions {
		return "Legend\nDetailed labels; descriptions included"
	}
	return "Legend\nCompact labels; add --details for descriptions"
}

func scopeTitle(scope *core.ThreadGraphScope) string {
	hop := "hops"
	if scope.Depth == 1 {
		hop = "hop"
	}
	return fmt.Sprintf("Bounded Thread neighborhood · %d %s · %d/%d nodes shown · %d hidden",
		scope.Depth, hop, scope.ShownNodes, scope.TotalNodes, scope.HiddenNodes)
}

func boundaryLabel(marker boundaryMarker) string {
	direction := "dependent"
	if marker.incoming {
		direction = "prerequisite"
	}
	noun := "edge"
	if marker.count != 1 {
		noun = "edges"
	}
	return fmt.Sprintf("… %d omitted %s %s", marker.count, direction, noun)
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
