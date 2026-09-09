package tui

import (
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/x/ansi"

	"github.com/andy-esch/taskflow/internal/core"
	"github.com/andy-esch/taskflow/internal/domain"
	"github.com/andy-esch/taskflow/internal/theme"
)

// The spatial prototype is intentionally a TUI-only adapter over the existing
// projection. These constants describe terminal cells, not repository graph
// semantics, and can be replaced with another renderer without a data migration.
const (
	threadSpatialNodeWidth      = 22
	threadSpatialNodeSlotHeight = 5
	threadSpatialNodeGapX       = 8
	threadSpatialNodeGapY       = 1
	threadSpatialNodeStrideY    = threadSpatialNodeSlotHeight + threadSpatialNodeGapY
	threadSpatialNodeTop        = 2
	threadSpatialMinWidth       = 60
	threadSpatialMinHeight      = 12
	threadSpatialMaxNodes       = 512
	threadSpatialMaxEdges       = 2048
	threadSpatialMaxCanvasCells = 750_000
)

type threadSpatialNode struct {
	node   core.ThreadGraphNode
	alias  string
	column int
	row    int
	x      int
	y      int
	order  int
}

type threadSpatialLayout struct {
	nodes        []threadSpatialNode
	byID         map[string]threadSpatialNode
	columns      [][]string
	columnLabels []string
	width        int
	height       int
}

func buildThreadSpatialLayout(projection core.ThreadGraphProjection) threadSpatialLayout {
	byNode := make(map[string]core.ThreadGraphNode, len(projection.Nodes))
	order := make(map[string]int, len(projection.Nodes))
	allIDs := make([]string, 0, len(projection.Nodes))
	for index, node := range projection.Nodes {
		if _, seen := byNode[node.TaskID]; seen {
			continue
		}
		byNode[node.TaskID] = node
		order[node.TaskID] = index
		allIDs = append(allIDs, node.TaskID)
	}
	for _, wave := range projection.Waves {
		for _, taskID := range wave.TaskIDs {
			if _, exists := byNode[taskID]; exists || taskID == "" {
				continue
			}
			byNode[taskID] = core.ThreadGraphNode{TaskID: taskID, Label: taskID, Role: core.ThreadTaskMember}
			order[taskID] = len(order)
			allIDs = append(allIDs, taskID)
		}
	}

	columns := rankThreadSpatialColumns(allIDs, projection.Edges, order)
	columns = placeThreadSpatialExternalGates(columns, projection.Edges, byNode, order)
	labels := labelThreadSpatialColumns(columns, projection, byNode)
	aliases := threadGraphAliases(projection)

	layout := threadSpatialLayout{
		byID: make(map[string]threadSpatialNode, len(allIDs)), columns: columns, columnLabels: labels,
		width: 1, height: 1,
	}
	strideX := threadSpatialNodeWidth + threadSpatialNodeGapX
	// Every row reserves the expanded focus-card height. Compact nodes occupy
	// the middle three rows, so moving focus can reveal detail without shifting
	// graph geometry or colliding with the node below it.
	for column, ids := range columns {
		for row, taskID := range ids {
			node := byNode[taskID]
			placement := threadSpatialNode{
				node: node, alias: aliases[taskID], column: column, row: row,
				x: 3 + column*strideX, y: threadSpatialNodeTop + row*threadSpatialNodeStrideY,
				order: order[taskID],
			}
			layout.nodes = append(layout.nodes, placement)
			layout.byID[taskID] = placement
			layout.width = max(layout.width, placement.x+threadSpatialNodeWidth+4)
			layout.height = max(layout.height, placement.y+threadSpatialNodeSlotHeight)
		}
	}
	return layout
}

// placeThreadSpatialExternalGates keeps a source-only boundary gate close to
// the work it actually gates. Longest-path ranking otherwise leaves every
// source in column zero, which makes a late external prerequisite draw across
// unrelated early waves. A gate may move only inside the open interval between
// its latest included prerequisite and earliest included dependent, so every
// supplied edge continues to point left-to-right.
func placeThreadSpatialExternalGates(
	columns [][]string,
	edges []core.ThreadGraphEdge,
	byNode map[string]core.ThreadGraphNode,
	order map[string]int,
) [][]string {
	positions := make(map[string]int)
	for column, ids := range columns {
		for _, taskID := range ids {
			positions[taskID] = column
		}
	}
	for _, taskID := range threadSpatialOrderedIDs(order) {
		if byNode[taskID].Role != core.ThreadTaskExternalGate {
			continue
		}
		current := positions[taskID]
		latestIncoming := -1
		earliestOutgoing := len(columns)
		for _, edge := range edges {
			if edge.To == taskID {
				if column, ok := positions[edge.From]; ok {
					latestIncoming = max(latestIncoming, column)
				}
			}
			if edge.From == taskID {
				if column, ok := positions[edge.To]; ok {
					earliestOutgoing = min(earliestOutgoing, column)
				}
			}
		}
		desired := earliestOutgoing - 1
		if earliestOutgoing == len(columns) || desired <= current || desired <= latestIncoming {
			continue
		}
		columns[current] = removeThreadSpatialID(columns[current], taskID)
		columns[desired] = append(columns[desired], taskID)
		positions[taskID] = desired
	}
	less := func(a, b string) bool {
		if order[a] != order[b] {
			return order[a] < order[b]
		}
		return a < b
	}
	compacted := make([][]string, 0, len(columns))
	for _, column := range columns {
		if len(column) == 0 {
			continue
		}
		sort.SliceStable(column, func(i, j int) bool { return less(column[i], column[j]) })
		compacted = append(compacted, column)
	}
	return compacted
}

func threadSpatialOrderedIDs(order map[string]int) []string {
	ids := make([]string, 0, len(order))
	for taskID := range order {
		ids = append(ids, taskID)
	}
	sort.SliceStable(ids, func(i, j int) bool {
		if order[ids[i]] != order[ids[j]] {
			return order[ids[i]] < order[ids[j]]
		}
		return ids[i] < ids[j]
	})
	return ids
}

func removeThreadSpatialID(ids []string, remove string) []string {
	kept := ids[:0]
	for _, taskID := range ids {
		if taskID != remove {
			kept = append(kept, taskID)
		}
	}
	return kept
}

// rankThreadSpatialColumns computes a longest-path layer over only the supplied
// projection nodes and edges. This is presentation geometry: it does not infer
// task eligibility or alter the core member-wave explanation. Cyclic residue is
// placed in one final partial column so broken evidence remains visible.
func rankThreadSpatialColumns(ids []string, edges []core.ThreadGraphEdge, order map[string]int) [][]string {
	included := make(map[string]bool, len(ids))
	indegree := make(map[string]int, len(ids))
	outgoing := make(map[string][]string, len(ids))
	for _, taskID := range ids {
		included[taskID] = true
	}
	seenEdges := make(map[core.ThreadGraphEdge]bool, len(edges))
	for _, edge := range edges {
		if !included[edge.From] || !included[edge.To] || edge.From == edge.To || seenEdges[edge] {
			continue
		}
		seenEdges[edge] = true
		outgoing[edge.From] = append(outgoing[edge.From], edge.To)
		indegree[edge.To]++
	}
	less := func(a, b string) bool {
		if order[a] != order[b] {
			return order[a] < order[b]
		}
		return a < b
	}
	for taskID := range outgoing {
		sort.SliceStable(outgoing[taskID], func(i, j int) bool { return less(outgoing[taskID][i], outgoing[taskID][j]) })
	}
	ready := make([]string, 0)
	for _, taskID := range ids {
		if indegree[taskID] == 0 {
			ready = append(ready, taskID)
		}
	}
	sort.SliceStable(ready, func(i, j int) bool { return less(ready[i], ready[j]) })

	ranks := make(map[string]int, len(ids))
	processed := make(map[string]bool, len(ids))
	maxRank := 0
	for len(ready) > 0 {
		current := ready[0]
		ready = ready[1:]
		if processed[current] {
			continue
		}
		processed[current] = true
		maxRank = max(maxRank, ranks[current])
		for _, dependent := range outgoing[current] {
			ranks[dependent] = max(ranks[dependent], ranks[current]+1)
			indegree[dependent]--
			if indegree[dependent] == 0 {
				ready = append(ready, dependent)
				sort.SliceStable(ready, func(i, j int) bool { return less(ready[i], ready[j]) })
			}
		}
	}
	for _, taskID := range ids {
		if !processed[taskID] {
			ranks[taskID] = maxRank + 1
		}
	}
	for _, rank := range ranks {
		maxRank = max(maxRank, rank)
	}
	columns := make([][]string, maxRank+1)
	for _, taskID := range ids {
		columns[ranks[taskID]] = append(columns[ranks[taskID]], taskID)
	}
	compacted := make([][]string, 0, len(columns))
	for _, column := range columns {
		if len(column) == 0 {
			continue
		}
		sort.SliceStable(column, func(i, j int) bool { return less(column[i], column[j]) })
		compacted = append(compacted, column)
	}
	return compacted
}

func labelThreadSpatialColumns(columns [][]string, projection core.ThreadGraphProjection, byNode map[string]core.ThreadGraphNode) []string {
	waves := make(map[string]int)
	for _, wave := range projection.Waves {
		for _, taskID := range wave.TaskIDs {
			waves[taskID] = wave.Index
		}
	}
	labels := make([]string, 0, len(columns))
	for index, column := range columns {
		waveSet := make(map[int]bool)
		external, unranked := false, false
		for _, taskID := range column {
			node := byNode[taskID]
			if node.Role == core.ThreadTaskExternalGate {
				external = true
			}
			if wave := waves[taskID]; wave > 0 {
				waveSet[wave] = true
			} else if node.Role != core.ThreadTaskExternalGate {
				unranked = true
			}
		}
		label := fmt.Sprintf("layer %d", index+1)
		waveNumbers := make([]int, 0, len(waveSet))
		for wave := range waveSet {
			waveNumbers = append(waveNumbers, wave)
		}
		sort.Ints(waveNumbers)
		switch len(waveNumbers) {
		case 1:
			label = fmt.Sprintf("wave %d", waveNumbers[0])
		case 2:
			label = fmt.Sprintf("waves %d+%d", waveNumbers[0], waveNumbers[1])
		default:
			if len(waveNumbers) > 2 {
				label = fmt.Sprintf("waves %d–%d", waveNumbers[0], waveNumbers[len(waveNumbers)-1])
			}
		}
		switch {
		case unranked:
			label += " + unranked"
		case external && len(waveSet) > 0:
			label += " + gate"
		case external && len(waveSet) == 0:
			label = "external"
		}
		labels = append(labels, label)
	}
	return labels
}

// threadSpatialMove implements layout navigation only. It never traverses past
// the bounded projection or derives readiness. Horizontal movement follows an
// actual edge in the requested direction first, even when that edge skips a
// populated presentation column. If several edges qualify, the nearest
// connected column wins before geometric row ties. Only a node with no edge in
// that direction falls back to the nearest populated column. Vertical movement
// stays within the current column.
func threadSpatialMove(projection core.ThreadGraphProjection, selected string, dx, dy int) string {
	layout := buildThreadSpatialLayout(projection)
	current, ok := spatialPlacement(layout, selected)
	if !ok {
		return firstSpatialSelectable(layout)
	}
	if dy != 0 {
		column := spatialSelectableColumn(layout, current.column)
		for index, candidate := range column {
			if candidate.node.TaskID != current.node.TaskID {
				continue
			}
			next := min(max(index+sign(dy), 0), len(column)-1)
			return column[next].node.TaskID
		}
		return current.node.TaskID
	}
	if dx == 0 {
		return current.node.TaskID
	}
	direct := make([]threadSpatialNode, 0)
	for _, candidate := range layout.nodes {
		if sign(candidate.column-current.column) != sign(dx) {
			continue
		}
		if threadSpatialDirectNeighbor(projection.Edges, current.node.TaskID, candidate.node.TaskID, dx) {
			direct = append(direct, candidate)
		}
	}
	if len(direct) > 0 {
		nearestColumn := direct[0].column
		for _, candidate := range direct[1:] {
			if abs(candidate.column-current.column) < abs(nearestColumn-current.column) {
				nearestColumn = candidate.column
			}
		}
		nearest := direct[:0]
		for _, candidate := range direct {
			if candidate.column == nearestColumn {
				nearest = append(nearest, candidate)
			}
		}
		return bestSpatialCandidate(current, nearest).node.TaskID
	}

	for column := current.column + sign(dx); column >= 0 && column < len(layout.columns); column += sign(dx) {
		candidates := spatialSelectableColumn(layout, column)
		if len(candidates) == 0 {
			continue
		}
		return bestSpatialCandidate(current, candidates).node.TaskID
	}
	return current.node.TaskID
}

func threadSpatialDirectNeighbor(edges []core.ThreadGraphEdge, current, candidate string, dx int) bool {
	for _, edge := range edges {
		if dx < 0 && edge.To == current && edge.From == candidate {
			return true
		}
		if dx > 0 && edge.From == current && edge.To == candidate {
			return true
		}
	}
	return false
}

func threadSpatialSelectedTaskID(projection core.ThreadGraphProjection, selected string) string {
	layout := buildThreadSpatialLayout(projection)
	return threadSpatialSelectedTaskIDInLayout(layout, selected)
}

func threadSpatialSelectedTaskIDInLayout(layout threadSpatialLayout, selected string) string {
	if _, ok := spatialPlacement(layout, selected); ok {
		return selected
	}
	return firstSpatialSelectable(layout)
}

func sign(value int) int {
	if value < 0 {
		return -1
	}
	if value > 0 {
		return 1
	}
	return 0
}

func spatialPlacement(layout threadSpatialLayout, taskID string) (threadSpatialNode, bool) {
	node, ok := layout.byID[taskID]
	return node, ok
}

func firstSpatialSelectable(layout threadSpatialLayout) string {
	if len(layout.nodes) > 0 {
		return layout.nodes[0].node.TaskID
	}
	return ""
}

func spatialSelectableColumn(layout threadSpatialLayout, column int) []threadSpatialNode {
	nodes := make([]threadSpatialNode, 0)
	for _, node := range layout.nodes {
		if node.column == column {
			nodes = append(nodes, node)
		}
	}
	sort.SliceStable(nodes, func(i, j int) bool {
		if nodes[i].row != nodes[j].row {
			return nodes[i].row < nodes[j].row
		}
		if nodes[i].order != nodes[j].order {
			return nodes[i].order < nodes[j].order
		}
		return nodes[i].node.TaskID < nodes[j].node.TaskID
	})
	return nodes
}

func bestSpatialCandidate(current threadSpatialNode, candidates []threadSpatialNode) threadSpatialNode {
	sort.SliceStable(candidates, func(i, j int) bool {
		iDistance := abs(candidates[i].row - current.row)
		jDistance := abs(candidates[j].row - current.row)
		if iDistance != jDistance {
			return iDistance < jDistance
		}
		if candidates[i].order != candidates[j].order {
			return candidates[i].order < candidates[j].order
		}
		return candidates[i].node.TaskID < candidates[j].node.TaskID
	})
	return candidates[0]
}

func abs(value int) int {
	if value < 0 {
		return -value
	}
	return value
}

type threadSpatialCell struct {
	text         string
	continuation bool
	color        theme.Color
	bold         bool
	accent       bool
	connector    threadSpatialConnector
}

type threadSpatialConnector uint8

const (
	threadSpatialLeft threadSpatialConnector = 1 << iota
	threadSpatialRight
	threadSpatialUp
	threadSpatialDown
)

type threadSpatialCanvas struct {
	cells [][]threadSpatialCell
	width int
}

func newThreadSpatialCanvas(width, height int) *threadSpatialCanvas {
	cells := make([][]threadSpatialCell, max(height, 1))
	for row := range cells {
		cells[row] = make([]threadSpatialCell, max(width, 1))
	}
	return &threadSpatialCanvas{cells: cells, width: max(width, 1)}
}

func (c *threadSpatialCanvas) putText(x, y int, value string, color theme.Color, bold bool) {
	if y < 0 || y >= len(c.cells) {
		return
	}
	column := x
	last := -1
	for _, r := range value {
		cellWidth := ansi.StringWidth(string(r))
		if cellWidth == 0 {
			if last >= 0 {
				c.cells[y][last].text += string(r)
			}
			continue
		}
		if column < 0 {
			column += cellWidth
			continue
		}
		if column+cellWidth > c.width {
			return
		}
		c.cells[y][column] = threadSpatialCell{text: string(r), color: color, bold: bold}
		last = column
		for offset := 1; offset < cellWidth; offset++ {
			c.cells[y][column+offset] = threadSpatialCell{continuation: true, color: color, bold: bold}
		}
		column += cellWidth
	}
}

func (c *threadSpatialCanvas) putConnector(x, y int, directions threadSpatialConnector, selected bool) {
	if y < 0 || y >= len(c.cells) || x < 0 || x >= c.width {
		return
	}
	cell := &c.cells[y][x]
	if cell.text != "▶" && cell.text != "◀" {
		cell.connector |= directions
	}
	if selected {
		cell.accent, cell.bold = true, true
	} else if cell.color == theme.ColorNone {
		cell.color = theme.ColorGray
	}
}

func (c *threadSpatialCanvas) putArrow(x, y int, arrow rune, selected bool) {
	if y < 0 || y >= len(c.cells) || x < 0 || x >= c.width {
		return
	}
	cell := &c.cells[y][x]
	cell.text = string(arrow)
	cell.connector = 0
	// Direction markers use the same yellow active-work language as Thread
	// frontier pointers. Selected connection strokes remain accent-colored, so
	// the arrowhead is legible as direction rather than another focus segment.
	cell.accent = false
	cell.color, cell.bold = theme.ColorYellow, selected
}

func (c *threadSpatialCanvas) horizontal(x0, x1, y int, selected bool) {
	if x1 < x0 {
		x0, x1 = x1, x0
	}
	for x := x0; x <= x1; x++ {
		directions := threadSpatialConnector(0)
		if x > x0 {
			directions |= threadSpatialLeft
		}
		if x < x1 {
			directions |= threadSpatialRight
		}
		c.putConnector(x, y, directions, selected)
	}
}

func (c *threadSpatialCanvas) vertical(x, y0, y1 int, selected bool) {
	if y1 < y0 {
		y0, y1 = y1, y0
	}
	for y := y0; y <= y1; y++ {
		directions := threadSpatialConnector(0)
		if y > y0 {
			directions |= threadSpatialUp
		}
		if y < y1 {
			directions |= threadSpatialDown
		}
		c.putConnector(x, y, directions, selected)
	}
}

func threadSpatialConnectorGlyph(directions threadSpatialConnector) string {
	switch directions {
	case threadSpatialLeft, threadSpatialRight, threadSpatialLeft | threadSpatialRight:
		return "─"
	case threadSpatialUp, threadSpatialDown, threadSpatialUp | threadSpatialDown:
		return "│"
	case threadSpatialRight | threadSpatialDown:
		return "┌"
	case threadSpatialLeft | threadSpatialDown:
		return "┐"
	case threadSpatialRight | threadSpatialUp:
		return "└"
	case threadSpatialLeft | threadSpatialUp:
		return "┘"
	case threadSpatialLeft | threadSpatialRight | threadSpatialDown:
		return "┬"
	case threadSpatialLeft | threadSpatialRight | threadSpatialUp:
		return "┴"
	case threadSpatialUp | threadSpatialDown | threadSpatialRight:
		return "├"
	case threadSpatialUp | threadSpatialDown | threadSpatialLeft:
		return "┤"
	case threadSpatialLeft | threadSpatialRight | threadSpatialUp | threadSpatialDown:
		return "┼"
	default:
		return " "
	}
}

func (c *threadSpatialCanvas) renderLine(row int, s *styles) string {
	if row < 0 || row >= len(c.cells) {
		return ""
	}
	var out strings.Builder
	var segment strings.Builder
	currentColor := theme.ColorNone
	currentBold := false
	currentAccent := false
	flush := func() {
		if segment.Len() == 0 {
			return
		}
		value := segment.String()
		if currentAccent {
			value = s.accent(value)
		} else if currentColor != theme.ColorNone {
			value = s.fg(currentColor, value)
		}
		if currentBold {
			value = s.selected.Render(value)
		}
		out.WriteString(value)
		segment.Reset()
	}
	for _, cell := range c.cells[row] {
		if cell.continuation {
			continue
		}
		text := cell.text
		if text == "" && cell.connector != 0 {
			text = threadSpatialConnectorGlyph(cell.connector)
		}
		if text == "" {
			text = " "
		}
		if cell.color != currentColor || cell.bold != currentBold || cell.accent != currentAccent {
			flush()
			currentColor, currentBold, currentAccent = cell.color, cell.bold, cell.accent
		}
		segment.WriteString(text)
	}
	flush()
	return strings.TrimRight(out.String(), " ")
}

func renderThreadSpatial(projection core.ThreadGraphProjection, pathIssue, selectedTaskID string, width, height int, s *styles) string {
	if width <= 0 {
		width = 120
	}
	if height <= 0 {
		height = 28
	}
	layout := buildThreadSpatialLayout(projection)
	selectedTaskID = threadSpatialSelectedTaskIDInLayout(layout, selectedTaskID)
	if width < threadSpatialMinWidth || height < threadSpatialMinHeight {
		return renderThreadSpatialNarrow(projection, layout, selectedTaskID, width, height, s)
	}
	if issue := threadSpatialCapacityIssue(layout, len(projection.Edges)); issue != "" {
		return renderThreadSpatialCapacityFallback(projection, layout, selectedTaskID, issue, width, height, s)
	}

	fixedRows := 8 // header + two legend rows + five-row focus card
	canvasHeight := max(1, height-fixedRows)
	canvas := renderThreadSpatialCanvas(projection, layout, selectedTaskID)
	selected, selectedOK := spatialPlacement(layout, selectedTaskID)
	panX, panY := 0, 0
	if selectedOK {
		panX = min(max(selected.x+threadSpatialNodeWidth/2-width/2, 0), max(layout.width-width, 0))
		panY = min(max(selected.y+threadSpatialNodeSlotHeight/2-canvasHeight/2, 0), max(layout.height-canvasHeight, 0))
	}

	topology := "partial"
	if projection.TopologyComplete {
		topology = "complete"
	}
	health := fmt.Sprintf("graph %s · projection %s", projection.View.GraphHealth, projection.View.ProjectionHealth)
	if projection.View.Inconsistent {
		health += " · inconsistent"
	}
	focus := "empty"
	if selectedOK && selected.column < len(layout.columnLabels) {
		focus = layout.columnLabels[selected.column]
	}
	header := fmt.Sprintf("spatial graph · %s · focus %s · prerequisite ─▶ dependent · %s", topology, focus, health)
	if pathIssue != "" {
		header += " · local path unavailable"
	}
	lines := []string{
		truncate(header, width),
		truncate(threadSpatialStatusLegend(s), width),
		truncate(threadSpatialRoleLegend(s), width),
	}
	for row := 0; row < canvasHeight; row++ {
		line := canvas.renderLine(panY+row, s)
		line = ansi.Cut(line, panX, panX+width)
		lines = append(lines, truncate(line, width))
	}
	lines = append(lines, threadSpatialInspector(projection, layout, selectedTaskID, width, s)...)
	return strings.Join(lines[:min(len(lines), height)], "\n")
}

func threadSpatialCapacityIssue(layout threadSpatialLayout, edgeCount int) string {
	switch {
	case len(layout.nodes) > threadSpatialMaxNodes:
		return fmt.Sprintf("%d nodes exceeds the %d-node prototype limit", len(layout.nodes), threadSpatialMaxNodes)
	case edgeCount > threadSpatialMaxEdges:
		return fmt.Sprintf("%d edges exceeds the %d-edge prototype limit", edgeCount, threadSpatialMaxEdges)
	case layout.width > 0 && layout.height > threadSpatialMaxCanvasCells/layout.width:
		return fmt.Sprintf("%d×%d layout exceeds the %d-cell prototype canvas limit",
			layout.width, layout.height, threadSpatialMaxCanvasCells)
	default:
		return ""
	}
}

func renderThreadSpatialCapacityFallback(
	projection core.ThreadGraphProjection,
	layout threadSpatialLayout,
	selectedTaskID, issue string,
	width, height int,
	s *styles,
) string {
	health := fmt.Sprintf("graph %s · projection %s", projection.View.GraphHealth, projection.View.ProjectionHealth)
	lines := []string{
		truncate("spatial graph · bounded prototype fallback · "+health, width),
		truncate(threadSpatialStatusLegend(s), width),
		truncate(threadSpatialRoleLegend(s), width),
		"",
		truncate(s.fg(theme.ColorYellow, theme.MarkerWarn.Glyph+" capacity guard")+" · "+issue, width),
		truncate(fmt.Sprintf("projection has %d nodes and %d edges; no partial graph was rendered", len(layout.nodes), len(projection.Edges)), width),
		truncate("Use v or Esc for the complete wave reader; f still opens the task picker.", width),
	}
	lines = append(lines, threadSpatialInspector(projection, layout, selectedTaskID, width, s)...)
	return strings.Join(lines[:min(len(lines), height)], "\n")
}

func renderThreadSpatialCanvas(projection core.ThreadGraphProjection, layout threadSpatialLayout, selected string) *threadSpatialCanvas {
	canvas := newThreadSpatialCanvas(layout.width, layout.height)
	strideX := threadSpatialNodeWidth + threadSpatialNodeGapX
	for column, label := range layout.columnLabels {
		x := 3 + column*strideX
		canvas.putText(x, 0, truncate(label, threadSpatialNodeWidth), theme.ColorGray, false)
	}
	for _, edge := range projection.Edges {
		from, fromOK := spatialPlacement(layout, edge.From)
		to, toOK := spatialPlacement(layout, edge.To)
		if !fromOK || !toOK {
			continue
		}
		drawThreadSpatialEdge(canvas, from, to, edge.From == selected || edge.To == selected)
	}
	for _, node := range layout.nodes {
		drawThreadSpatialNode(canvas, node, node.node.TaskID == selected)
	}
	return canvas
}

func drawThreadSpatialEdge(canvas *threadSpatialCanvas, from, to threadSpatialNode, selected bool) {
	fromX, fromY := from.x+threadSpatialNodeWidth, from.y+threadSpatialNodeSlotHeight/2
	toY := to.y + threadSpatialNodeSlotHeight/2
	if to.x > from.x {
		toX := to.x - 1
		if to.column-from.column > 1 {
			drawThreadSpatialLongEdge(canvas, from, to, fromX, fromY, toX, toY, selected)
			return
		}
		middle := fromX + max(1, (toX-fromX)/2)
		canvas.horizontal(fromX, middle, fromY, selected)
		canvas.vertical(middle, fromY, toY, selected)
		canvas.horizontal(middle, toX, toY, selected)
		canvas.putArrow(toX, toY, '▶', selected)
		return
	}

	// Partial/unranked topology can place an edge in the same or an earlier
	// column. Loop it around the right edge instead of pretending it follows the
	// healthy left-to-right wave direction.
	lane := max(from.x, to.x) + threadSpatialNodeWidth + 2
	canvas.horizontal(fromX, lane, fromY, selected)
	canvas.vertical(lane, fromY, toY, selected)
	canvas.horizontal(to.x+threadSpatialNodeWidth, lane, toY, selected)
	canvas.putArrow(to.x+threadSpatialNodeWidth, toY, '◀', selected)
}

// drawThreadSpatialLongEdge moves the cross-column segment onto the blank row
// reserved between node slots. The earlier straight route could pass through an
// intermediate node and borrow that node's arrowhead, visually inventing an
// edge that was not present in the projection.
func drawThreadSpatialLongEdge(
	canvas *threadSpatialCanvas,
	from, to threadSpatialNode,
	fromX, fromY, toX, toY int,
	selected bool,
) {
	trackRow := min(from.row, to.row)
	trackY := threadSpatialNodeTop + trackRow*threadSpatialNodeStrideY + threadSpatialNodeSlotHeight
	fromLane := fromX + 1
	toLane := toX - 1
	canvas.horizontal(fromX, fromLane, fromY, selected)
	canvas.vertical(fromLane, fromY, trackY, selected)
	canvas.horizontal(fromLane, toLane, trackY, selected)
	canvas.vertical(toLane, trackY, toY, selected)
	canvas.horizontal(toLane, toX, toY, selected)
	canvas.putArrow(toX, toY, '▶', selected)
}

func drawThreadSpatialNode(canvas *threadSpatialCanvas, placement threadSpatialNode, selected bool) {
	node := placement.node
	color := theme.Status(node.Status).Color
	if node.State.Role == core.RoleUnknown || node.State.Gate == core.GateBroken {
		color = theme.ColorRed
	}
	left, horizontal, right := "┌", "─", "┐"
	bottomLeft, bottomHorizontal, bottomRight := "└", "─", "┘"
	vertical := "│"
	if node.Role == core.ThreadTaskExternalGate {
		left, horizontal, right = "╔", "═", "╗"
		bottomLeft, bottomHorizontal, bottomRight = "╚", "═", "╝"
		vertical = "║"
	}
	top := left + strings.Repeat(horizontal, threadSpatialNodeWidth-2) + right
	bottom := bottomLeft + strings.Repeat(bottomHorizontal, threadSpatialNodeWidth-2) + bottomRight
	label := terminalText(node.Label)
	if label == "" {
		label = terminalText(node.TaskID)
	}
	marker := theme.Status(node.Status).Glyph
	if node.State.Role == core.RoleUnknown || node.State.Gate == core.GateBroken {
		marker = theme.MarkerUnreadable.Glyph
	}
	insideWidth := threadSpatialNodeWidth - 4
	if !selected {
		inside := truncate(marker+" "+label, insideWidth)
		middle := vertical + " " + padRight(inside, insideWidth) + " " + vertical
		compactY := placement.y + 1
		canvas.putText(placement.x, compactY, top, color, false)
		canvas.putText(placement.x, compactY+1, middle, color, false)
		canvas.putText(placement.x, compactY+2, bottom, color, false)
		return
	}

	// Selection expands inside its pre-reserved five-row slot without changing
	// the box's semantic status color. Focus is conveyed by geometry, the yellow
	// pointer, and accent-colored touching edges rather than repainting state.
	line := func(value string) string {
		return vertical + " " + padRight(truncate(value, insideWidth), insideWidth) + " " + vertical
	}
	statusLabel := string(node.Status)
	if statusLabel == "" {
		statusLabel = "unknown"
	}
	canvas.putText(placement.x, placement.y, top, color, true)
	canvas.putText(placement.x, placement.y+1, line(marker+" "+label), color, true)
	canvas.putText(placement.x, placement.y+2, line("["+placement.alias+"] "+statusLabel), color, false)
	canvas.putText(placement.x, placement.y+3, line(string(node.State.Role)+" / "+string(node.State.Gate)), color, false)
	canvas.putText(placement.x, placement.y+4, bottom, color, true)
	canvas.putText(placement.x-2, placement.y+2, "›", theme.ColorYellow, true)
}

func threadSpatialStatusLegend(s *styles) string {
	entries := []struct {
		status string
		label  string
	}{
		{"in-progress", "active"}, {"next-up", "next"}, {"ready-to-start", "ready"},
		{"completed", "done"}, {"deferred", "deferred"}, {"deprecated", "deprecated"},
	}
	parts := make([]string, 0, len(entries)+1)
	parts = append(parts, s.dim("status"))
	for _, entry := range entries {
		token := theme.Status(domain.Status(entry.status))
		parts = append(parts, s.fg(token.Color, token.Glyph)+" "+entry.label)
	}
	return strings.Join(parts, "  ")
}

func threadSpatialRoleLegend(s *styles) string {
	return s.dim("roles") + "  ┌ member ┐  ╔ external gate ╗  " +
		s.accent("magenta lines") + " touch focus  " + s.fg(theme.ColorYellow, "› focus · ▶ direction")
}

func threadSpatialInspector(projection core.ThreadGraphProjection, layout threadSpatialLayout, selected string, width int, s *styles) []string {
	placement, ok := spatialPlacement(layout, selected)
	if !ok {
		return threadSpatialInspectorBox("focus", []string{"no readable task", "", ""}, width, s)
	}
	node := placement.node
	status := theme.Status(node.Status)
	statusLabel := string(node.Status)
	if statusLabel == "" {
		statusLabel = "unknown"
	}
	name := terminalText(node.Label)
	if name == "" {
		name = terminalText(node.TaskID)
	}
	selectedLine := s.fg(status.Color, status.Glyph+" "+statusLabel) + "  " +
		s.accent(name) + "  " + s.dim("("+terminalText(node.TaskID)+")")
	stateValue := string(node.State.Role) + "/" + string(node.State.Gate)
	switch {
	case node.State.Role == core.RoleUnknown || node.State.Gate == core.GateBroken:
		stateValue = s.fg(theme.ColorRed, stateValue)
	case node.State.Gate == core.GateBlocked:
		stateValue = s.fg(theme.ColorYellow, stateValue)
	case node.State.Gate == core.GateClear:
		stateValue = s.fg(theme.ColorGreen, stateValue)
	}
	stateLine := s.dim("state") + " " + stateValue + "  " + s.dim("role") + " " + string(node.Role)
	needs, unlocks := threadSpatialConnections(projection, layout, selected)
	if needs != "" {
		stateLine += "  " + s.accent("needs") + " " + needs
	}
	if unlocks != "" {
		stateLine += "  " + s.accent("unlocks") + " " + unlocks
	}
	description := terminalText(node.Description)
	if description == "" {
		description = "(no description)"
	}
	return threadSpatialInspectorBox("focus ["+placement.alias+"]", []string{
		selectedLine,
		stateLine,
		s.dim("about") + " " + description,
	}, width, s)
}

func threadSpatialInspectorBox(title string, content []string, width int, s *styles) []string {
	if width < 4 {
		return []string{truncate(strings.Join(content, " "), width)}
	}
	insideWidth := width - 2
	title = terminalText(title)
	topLabel := truncate("─ "+title+" ", insideWidth)
	top := s.accent("╭" + topLabel + strings.Repeat("─", max(insideWidth-ansi.StringWidth(topLabel), 0)) + "╮")
	lines := []string{top}
	for index := 0; index < 3; index++ {
		value := ""
		if index < len(content) {
			value = truncate(content[index], max(insideWidth-2, 0))
		}
		lines = append(lines, s.accent("│")+" "+padRight(value, max(insideWidth-2, 0))+" "+s.accent("│"))
	}
	lines = append(lines, s.accent("╰"+strings.Repeat("─", insideWidth)+"╯"))
	return lines
}

func threadSpatialConnections(projection core.ThreadGraphProjection, layout threadSpatialLayout, selected string) (string, string) {
	needs := make([]string, 0)
	unlocks := make([]string, 0)
	for _, edge := range projection.Edges {
		switch {
		case edge.To == selected:
			needs = append(needs, threadSpatialConnectionLabel(layout, edge.From))
		case edge.From == selected:
			unlocks = append(unlocks, threadSpatialConnectionLabel(layout, edge.To))
		}
	}
	return strings.Join(needs, ", "), strings.Join(unlocks, ", ")
}

func threadSpatialConnectionLabel(layout threadSpatialLayout, taskID string) string {
	placement, ok := spatialPlacement(layout, taskID)
	if !ok {
		return terminalText(taskID)
	}
	label := terminalText(placement.node.Label)
	if label == "" {
		label = terminalText(taskID)
	}
	if placement.alias == "" {
		return label
	}
	return "[" + placement.alias + "] " + label
}

func renderThreadSpatialNarrow(projection core.ThreadGraphProjection, layout threadSpatialLayout, selected string, width, height int, s *styles) string {
	if width <= 0 {
		width = 1
	}
	if height <= 0 {
		height = 1
	}
	lines := []string{
		truncate("spatial graph · prerequisite ─▶ dependent", width),
		truncate(threadSpatialStatusLegend(s), width),
		"",
		truncate(fmt.Sprintf("This prototype needs at least %d×%d terminal cells.", threadSpatialMinWidth, threadSpatialMinHeight), width),
		truncate("Resize to reveal the graph; Esc returns to the wave reader.", width),
		"",
	}
	lines = append(lines, threadSpatialInspector(projection, layout, selected, width, s)...)
	return strings.Join(lines[:min(len(lines), height)], "\n")
}
