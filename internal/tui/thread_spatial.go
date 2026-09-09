package tui

import (
	"fmt"
	"sort"
	"strconv"
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
	threadSpatialNodeTop        = 2
	threadSpatialMinWidth       = 60
	threadSpatialMinHeight      = 12
	threadSpatialMaxNodes       = 512
	threadSpatialMaxEdges       = 2048
	threadSpatialMaxCanvasCells = 500_000
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

type threadSpatialPoint struct {
	x int
	y int
}

type threadSpatialRouteSegment struct {
	from     threadSpatialPoint
	to       threadSpatialPoint
	corridor bool
}

type threadSpatialRoute struct {
	id        int32
	edge      core.ThreadGraphEdge
	segments  []threadSpatialRouteSegment
	arrow     threadSpatialPoint
	arrowRune rune
}

type threadSpatialRouteSeed struct {
	id                          int32
	edge                        core.ThreadGraphEdge
	fromColumn, fromRow         int
	toColumn, toRow             int
	needsTrack, forwardAdjacent bool
}

type threadSpatialRouteLanes struct {
	source int
	target int
}

type threadSpatialLayout struct {
	nodes        []threadSpatialNode
	byID         map[string]threadSpatialNode
	columns      [][]string
	columnX      []int
	columnLabels []string
	routes       []threadSpatialRoute
	width        int
	height       int
}

func buildThreadSpatialLayout(projection core.ThreadGraphProjection) threadSpatialLayout {
	byNode := make(map[string]core.ThreadGraphNode, len(projection.Nodes))
	order := make(map[string]int, len(projection.Nodes))
	allIDs := make([]string, 0, len(projection.Nodes))
	for _, node := range orderedThreadGraphNodes(projection.Nodes) {
		if _, seen := byNode[node.TaskID]; seen {
			continue
		}
		byNode[node.TaskID] = node
		order[node.TaskID] = len(order)
		allIDs = append(allIDs, node.TaskID)
	}
	missing := make(map[string]bool)
	for _, wave := range projection.Waves {
		for _, taskID := range wave.TaskIDs {
			if _, exists := byNode[taskID]; exists || taskID == "" {
				continue
			}
			missing[taskID] = true
		}
	}
	missingIDs := make([]string, 0, len(missing))
	for taskID := range missing {
		missingIDs = append(missingIDs, taskID)
	}
	sort.Strings(missingIDs)
	for _, taskID := range missingIDs {
		byNode[taskID] = core.ThreadGraphNode{TaskID: taskID, Label: taskID, Role: core.ThreadTaskMember}
		order[taskID] = len(order)
		allIDs = append(allIDs, taskID)
	}

	columns := rankThreadSpatialColumns(allIDs, projection.Edges, order)
	columns = placeThreadSpatialExternalGates(columns, projection.Edges, byNode, order)
	labels := labelThreadSpatialColumns(columns, projection, byNode)
	aliases := threadGraphAliases(projection)
	seeds := buildThreadSpatialRouteSeeds(columns, projection.Edges)
	columnX, boundaryLanes, layoutWidth := threadSpatialColumnGeometry(len(columns), seeds)
	rowCount := 0
	for _, ids := range columns {
		rowCount = max(rowCount, len(ids))
	}
	rowY, trackY, layoutHeight := threadSpatialRowGeometry(rowCount, seeds)

	layout := threadSpatialLayout{
		byID: make(map[string]threadSpatialNode, len(allIDs)), columns: columns, columnX: columnX, columnLabels: labels,
		width: max(layoutWidth, 1), height: max(layoutHeight, 1),
	}
	// Every row reserves the expanded focus-card height. Compact nodes occupy
	// the middle three rows, so moving focus can reveal detail without shifting
	// graph geometry or colliding with the node below it. Inter-row and
	// inter-column gaps expand only when distinct edge lanes require the space.
	for column, ids := range columns {
		for row, taskID := range ids {
			node := byNode[taskID]
			placement := threadSpatialNode{
				node: node, alias: aliases[taskID], column: column, row: row,
				x: columnX[column], y: rowY[row],
				order: order[taskID],
			}
			layout.nodes = append(layout.nodes, placement)
			layout.byID[taskID] = placement
		}
	}
	layout.routes = materializeThreadSpatialRoutes(seeds, layout.byID, boundaryLanes, trackY)
	return layout
}

func buildThreadSpatialRouteSeeds(columns [][]string, edges []core.ThreadGraphEdge) []threadSpatialRouteSeed {
	type position struct{ column, row int }
	positions := make(map[string]position)
	for column, ids := range columns {
		for row, taskID := range ids {
			positions[taskID] = position{column: column, row: row}
		}
	}
	seeds := make([]threadSpatialRouteSeed, 0, len(edges))
	for index, edge := range orderedThreadGraphEdges(edges) {
		from, fromOK := positions[edge.From]
		to, toOK := positions[edge.To]
		if !fromOK || !toOK {
			continue
		}
		self := edge.From == edge.To
		needsTrack := self || from.row != to.row || to.column > from.column+1 || to.column < from.column
		seeds = append(seeds, threadSpatialRouteSeed{
			id: int32(index + 1), edge: edge,
			fromColumn: from.column, fromRow: from.row, toColumn: to.column, toRow: to.row,
			forwardAdjacent: to.column == from.column+1,
			needsTrack:      needsTrack,
		})
	}
	return seeds
}

func (seed threadSpatialRouteSeed) targetBoundary() int {
	if seed.toColumn > seed.fromColumn {
		return seed.toColumn - 1
	}
	return seed.toColumn
}

func threadSpatialColumnGeometry(
	columnCount int,
	seeds []threadSpatialRouteSeed,
) ([]int, map[int32]threadSpatialRouteLanes, int) {
	if columnCount == 0 {
		return nil, nil, 1
	}
	type laneUse struct {
		routeID int32
		target  bool
	}
	sourceUses := make(map[int][]laneUse)
	targetUses := make(map[int][]laneUse)
	for _, seed := range seeds {
		if !seed.needsTrack {
			continue
		}
		for _, use := range []struct {
			boundary int
			target   bool
		}{{boundary: seed.fromColumn}, {boundary: seed.targetBoundary(), target: true}} {
			if use.boundary >= 0 && use.boundary < columnCount {
				lane := laneUse{routeID: seed.id, target: use.target}
				if use.target {
					targetUses[use.boundary] = append(targetUses[use.boundary], lane)
				} else {
					sourceUses[use.boundary] = append(sourceUses[use.boundary], lane)
				}
			}
		}
	}
	uses := make(map[int][]laneUse, columnCount)
	for boundary := 0; boundary < columnCount; boundary++ {
		// Source lanes stay beside the prerequisite column and target lanes
		// beside the dependent column. Their endpoint stubs therefore cannot
		// cut across the other side's turning lane.
		uses[boundary] = append(uses[boundary], sourceUses[boundary]...)
		uses[boundary] = append(uses[boundary], targetUses[boundary]...)
	}
	gaps := make([]int, columnCount)
	for boundary := range gaps {
		// Five reserved cells separate route lanes from node endpoint arrows
		// and side-aware fan counts on both sides of the next node.
		gaps[boundary] = max(threadSpatialNodeGapX, len(uses[boundary])+5)
	}
	columnX := make([]int, columnCount)
	columnX[0] = 3
	for column := 1; column < columnCount; column++ {
		columnX[column] = columnX[column-1] + threadSpatialNodeWidth + gaps[column-1]
	}
	lanes := make(map[int32]threadSpatialRouteLanes, len(seeds))
	for boundary, laneUses := range uses {
		for index, use := range laneUses {
			lane := columnX[boundary] + threadSpatialNodeWidth + 3 + index
			assigned := lanes[use.routeID]
			if use.target {
				assigned.target = lane
			} else {
				assigned.source = lane
			}
			lanes[use.routeID] = assigned
		}
	}
	last := columnCount - 1
	width := columnX[last] + threadSpatialNodeWidth + gaps[last] + 1
	return columnX, lanes, width
}

func threadSpatialRowGeometry(
	rowCount int,
	seeds []threadSpatialRouteSeed,
) ([]int, map[int32]int, int) {
	if rowCount == 0 {
		return nil, nil, 1
	}
	tracks := make(map[int][]int32)
	for _, seed := range seeds {
		if seed.needsTrack {
			tracks[min(seed.fromRow, seed.toRow)] = append(tracks[min(seed.fromRow, seed.toRow)], seed.id)
		}
	}
	gaps := make([]int, rowCount)
	for row := range gaps {
		gaps[row] = max(threadSpatialNodeGapY, len(tracks[row]))
	}
	rowY := make([]int, rowCount)
	rowY[0] = threadSpatialNodeTop
	for row := 1; row < rowCount; row++ {
		rowY[row] = rowY[row-1] + threadSpatialNodeSlotHeight + gaps[row-1]
	}
	trackY := make(map[int32]int)
	for row, routeIDs := range tracks {
		for index, routeID := range routeIDs {
			trackY[routeID] = rowY[row] + threadSpatialNodeSlotHeight + index
		}
	}
	last := rowCount - 1
	height := rowY[last] + threadSpatialNodeSlotHeight + gaps[last]
	return rowY, trackY, height
}

func materializeThreadSpatialRoutes(
	seeds []threadSpatialRouteSeed,
	placements map[string]threadSpatialNode,
	lanes map[int32]threadSpatialRouteLanes,
	trackYs map[int32]int,
) []threadSpatialRoute {
	routes := make([]threadSpatialRoute, 0, len(seeds))
	for _, seed := range seeds {
		from, fromOK := placements[seed.edge.From]
		to, toOK := placements[seed.edge.To]
		if !fromOK || !toOK {
			continue
		}
		fromPoint := threadSpatialPoint{x: from.x + threadSpatialNodeWidth, y: from.y + threadSpatialNodeSlotHeight/2}
		toLeft := threadSpatialPoint{x: to.x - 1, y: to.y + threadSpatialNodeSlotHeight/2}
		toRight := threadSpatialPoint{x: to.x + threadSpatialNodeWidth, y: to.y + threadSpatialNodeSlotHeight/2}
		var points []threadSpatialPoint
		arrow, arrowRune := toLeft, '▶'
		switch {
		case seed.needsTrack:
			assigned, trackY := lanes[seed.id], trackYs[seed.id]
			target := toLeft
			if seed.toColumn <= seed.fromColumn {
				target = toRight
				arrow, arrowRune = toRight, '◀'
			}
			points = []threadSpatialPoint{
				fromPoint, {x: assigned.source, y: fromPoint.y}, {x: assigned.source, y: trackY},
				{x: assigned.target, y: trackY}, {x: assigned.target, y: target.y}, target,
			}
		case seed.forwardAdjacent:
			points = []threadSpatialPoint{fromPoint, toLeft}
		}
		routes = append(routes, newThreadSpatialRoute(seed, points, arrow, arrowRune, trackYs[seed.id]))
	}
	return routes
}

func newThreadSpatialRoute(
	seed threadSpatialRouteSeed,
	points []threadSpatialPoint,
	arrow threadSpatialPoint,
	arrowRune rune,
	trackY int,
) threadSpatialRoute {
	route := threadSpatialRoute{id: seed.id, edge: seed.edge, arrow: arrow, arrowRune: arrowRune}
	for index := 1; index < len(points); index++ {
		if points[index-1] == points[index] {
			continue
		}
		route.segments = append(route.segments, threadSpatialRouteSegment{
			from: points[index-1], to: points[index],
			corridor: seed.needsTrack && points[index-1].y == trackY && points[index].y == trackY,
		})
	}
	return route
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
	incoming := make(map[string][]string)
	outgoing := make(map[string][]string)
	for _, edge := range edges {
		if _, fromOK := positions[edge.From]; !fromOK {
			continue
		}
		if _, toOK := positions[edge.To]; !toOK {
			continue
		}
		outgoing[edge.From] = append(outgoing[edge.From], edge.To)
		incoming[edge.To] = append(incoming[edge.To], edge.From)
	}
	ordered := threadSpatialOrderedIDs(order)
	// A downstream gate can move after its upstream gate was already visited.
	// Repeat stable passes until no gate advances. Each move jumps directly to
	// the current valid boundary; a new pass propagates that move through at
	// least one earlier-visited gate in a chain, so node count bounds the fixed
	// point. Pre-indexed incident lists keep each pass linear in supplied edges.
	for pass := 0; pass < max(1, len(ordered)+1); pass++ {
		moved := false
		for _, taskID := range ordered {
			if byNode[taskID].Role != core.ThreadTaskExternalGate {
				continue
			}
			current := positions[taskID]
			latestIncoming := -1
			earliestOutgoing := len(columns)
			for _, prerequisiteID := range incoming[taskID] {
				latestIncoming = max(latestIncoming, positions[prerequisiteID])
			}
			for _, dependentID := range outgoing[taskID] {
				earliestOutgoing = min(earliestOutgoing, positions[dependentID])
			}
			desired := earliestOutgoing - 1
			if earliestOutgoing == len(columns) || desired <= current || desired <= latestIncoming {
				continue
			}
			columns[current] = removeThreadSpatialID(columns[current], taskID)
			columns[desired] = append(columns[desired], taskID)
			positions[taskID] = desired
			moved = true
		}
		if !moved {
			break
		}
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
			label += fmt.Sprintf(" · wave %d", waveNumbers[0])
		case 2:
			label += fmt.Sprintf(" · waves %d+%d", waveNumbers[0], waveNumbers[1])
		default:
			if len(waveNumbers) > 2 {
				parts := make([]string, 0, len(waveNumbers))
				for _, wave := range waveNumbers {
					parts = append(parts, strconv.Itoa(wave))
				}
				label += " · waves " + strings.Join(parts, "+")
			}
		}
		switch {
		case unranked:
			label += " · unranked"
		case external && len(waveSet) > 0:
			label += " · gate"
		case external && len(waveSet) == 0:
			label += " · external"
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
	routeID      int32
	shared       bool
	crossing     bool
	overlap      bool
	conflict     bool
	routeCount   bool
}

type threadSpatialRouteStyle struct {
	id       int32
	from     string
	to       string
	selected bool
}

type threadSpatialConnector uint8

const (
	threadSpatialLeft threadSpatialConnector = 1 << iota
	threadSpatialRight
	threadSpatialUp
	threadSpatialDown
)

type threadSpatialCanvas struct {
	cells       [][]threadSpatialCell
	routeStyles map[int32]threadSpatialRouteStyle
	width       int
}

func newThreadSpatialCanvas(width, height int) *threadSpatialCanvas {
	cells := make([][]threadSpatialCell, max(height, 1))
	for row := range cells {
		cells[row] = make([]threadSpatialCell, max(width, 1))
	}
	return &threadSpatialCanvas{
		cells: cells, routeStyles: make(map[int32]threadSpatialRouteStyle), width: max(width, 1),
	}
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

func (c *threadSpatialCanvas) putAccentText(x, y int, value string, bold bool) {
	c.putText(x, y, value, theme.ColorNone, bold)
	for column := x; column < x+ansi.StringWidth(value) && column < c.width; column++ {
		if y >= 0 && y < len(c.cells) && column >= 0 && !c.cells[y][column].continuation {
			c.cells[y][column].accent = true
		}
	}
}

func (c *threadSpatialCanvas) putRouteConnector(
	x, y int,
	directions threadSpatialConnector,
	style threadSpatialRouteStyle,
) {
	if y < 0 || y >= len(c.cells) || x < 0 || x >= c.width {
		return
	}
	c.routeStyles[style.id] = style
	cell := &c.cells[y][x]
	switch {
	case cell.crossing || cell.overlap || cell.conflict:
		// Dedicated lanes and tracks make each crossing a straight horizontal
		// route over a straight vertical route. A further coincident stroke
		// cannot turn the collision into a junction.
	case cell.routeID == 0 || cell.routeID == style.id:
		cell.connector |= directions
		if cell.routeID == 0 {
			cell.routeID = style.id
		}
	default:
		resident := c.routeStyles[cell.routeID]
		existingHorizontal, existingVertical := threadSpatialConnectorAxes(cell.connector)
		incomingHorizontal, incomingVertical := threadSpatialConnectorAxes(directions)
		sharesEndpoint := resident.from == style.from || resident.from == style.to ||
			resident.to == style.from || resident.to == style.to
		perpendicular := (existingHorizontal && incomingVertical) || (existingVertical && incomingHorizontal)
		existingTurns := existingHorizontal && existingVertical
		incomingTurns := incomingHorizontal && incomingVertical
		switch {
		case perpendicular && (existingTurns || incomingTurns) && sharesEndpoint:
			// A turn that joins a straight stub leading to the same endpoint is a
			// real fan-in/fan-out bundle. Preserve every arm; unlike a crossing,
			// the routes intentionally share the remaining endpoint segment.
			cell.connector |= directions
			cell.shared = true
		case perpendicular && (existingTurns || incomingTurns):
			// This is a routing invariant failure, not a legitimate crossing:
			// flattening it would sever the turning route and invent endpoints.
			cell.connector, cell.shared, cell.crossing = 0, false, false
			cell.conflict = true
		case perpendicular:
			// Physical intersection does not imply connectivity, even when the
			// two routes eventually share an endpoint elsewhere.
			cell.connector, cell.shared = 0, false
			cell.crossing = true
		case sharesEndpoint:
			cell.connector |= directions
			cell.shared = true
		default:
			// Collinear overlap without a common endpoint is an independent
			// bundle. Keep it visibly distinct from a true shared endpoint.
			cell.connector, cell.shared = 0, false
			cell.overlap = true
		}
	}
	if cell.crossing || cell.overlap || cell.conflict {
		// Route collisions describe topology or renderer state, not focus. Keep
		// them neutral so a selected edge cannot make an unrelated crossing look
		// connected to the selected task.
		cell.accent = false
		cell.color = theme.ColorGray
		cell.bold = cell.conflict
	} else if style.selected || (cell.routeID != 0 && c.routeStyles[cell.routeID].selected) {
		cell.accent, cell.bold = true, true
	} else if cell.color == theme.ColorNone {
		cell.color = theme.ColorGray
	}
}

func threadSpatialConnectorAxes(connector threadSpatialConnector) (horizontal, vertical bool) {
	horizontal = connector&(threadSpatialLeft|threadSpatialRight) != 0
	vertical = connector&(threadSpatialUp|threadSpatialDown) != 0
	return horizontal, vertical
}

func (c *threadSpatialCanvas) putRouteArrow(x, y int, arrow rune, style threadSpatialRouteStyle) {
	if y < 0 || y >= len(c.cells) || x < 0 || x >= c.width {
		return
	}
	cell := &c.cells[y][x]
	if cell.text == string(arrow) && cell.routeID != 0 && cell.routeID != style.id {
		cell.shared = true
	}
	cell.text = string(arrow)
	cell.connector = 0
	cell.routeID = style.id
	cell.crossing, cell.overlap, cell.conflict = false, false, false
	// Direction markers use the same yellow active-work language as Thread
	// frontier pointers. Selected connection strokes remain accent-colored, so
	// the arrowhead is legible as direction rather than another focus segment.
	cell.accent = false
	cell.color, cell.bold = theme.ColorYellow, cell.bold || style.selected
}

func (c *threadSpatialCanvas) putRouteCount(point threadSpatialPoint, count int, selected bool) bool {
	if count < 2 || point.y < 0 || point.y >= len(c.cells) || point.x < 0 || point.x >= c.width {
		return false
	}
	label := "+"
	if count < 10 {
		label = strconv.Itoa(count)
	}
	cells := &c.cells[point.y][point.x]
	if cells.crossing || cells.overlap || cells.conflict || cells.text != "" || cells.continuation {
		return false
	}
	cells.text, cells.connector, cells.routeID = label, 0, 0
	cells.shared, cells.crossing = false, false
	cells.color, cells.bold, cells.accent = theme.ColorYellow, selected, false
	cells.routeCount = true
	return true
}

func (c *threadSpatialCanvas) putRouteCountAlong(
	point threadSpatialPoint,
	step, maxSteps, count int,
	selected bool,
) (threadSpatialPoint, bool) {
	for attempt := 0; attempt <= maxSteps; attempt++ {
		candidate := point
		candidate.x += step * attempt
		if c.putRouteCount(candidate, count, selected) {
			return candidate, true
		}
	}
	return threadSpatialPoint{}, false
}

func (c *threadSpatialCanvas) putRouteCountFallback(point threadSpatialPoint, count int, selected bool) bool {
	if count < 2 || point.y < 0 || point.y >= len(c.cells) || point.x < 0 || point.x >= c.width {
		return false
	}
	label := "+"
	if count < 10 {
		label = strconv.Itoa(count)
	}
	c.cells[point.y][point.x] = threadSpatialCell{
		text: label, color: theme.ColorYellow, bold: selected, routeCount: true,
	}
	return true
}

func (c *threadSpatialCanvas) routeHorizontal(x0, x1, y int, style threadSpatialRouteStyle) {
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
		c.putRouteConnector(x, y, directions, style)
	}
}

func (c *threadSpatialCanvas) routeVertical(x, y0, y1 int, style threadSpatialRouteStyle) {
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
		c.putRouteConnector(x, y, directions, style)
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

func threadSpatialSharedConnectorGlyph(directions threadSpatialConnector) string {
	switch directions {
	case threadSpatialLeft, threadSpatialRight, threadSpatialLeft | threadSpatialRight:
		return "═"
	case threadSpatialUp, threadSpatialDown, threadSpatialUp | threadSpatialDown:
		return "║"
	case threadSpatialRight | threadSpatialDown:
		return "╔"
	case threadSpatialLeft | threadSpatialDown:
		return "╗"
	case threadSpatialRight | threadSpatialUp:
		return "╚"
	case threadSpatialLeft | threadSpatialUp:
		return "╝"
	case threadSpatialLeft | threadSpatialRight | threadSpatialDown:
		return "╦"
	case threadSpatialLeft | threadSpatialRight | threadSpatialUp:
		return "╩"
	case threadSpatialUp | threadSpatialDown | threadSpatialRight:
		return "╠"
	case threadSpatialUp | threadSpatialDown | threadSpatialLeft:
		return "╣"
	case threadSpatialLeft | threadSpatialRight | threadSpatialUp | threadSpatialDown:
		return "╬"
	default:
		return " "
	}
}

func threadSpatialFocusConnectorGlyph(directions threadSpatialConnector) string {
	switch directions {
	case threadSpatialLeft, threadSpatialRight, threadSpatialLeft | threadSpatialRight:
		return "━"
	case threadSpatialUp, threadSpatialDown, threadSpatialUp | threadSpatialDown:
		return "┃"
	case threadSpatialRight | threadSpatialDown:
		return "┏"
	case threadSpatialLeft | threadSpatialDown:
		return "┓"
	case threadSpatialRight | threadSpatialUp:
		return "┗"
	case threadSpatialLeft | threadSpatialUp:
		return "┛"
	case threadSpatialLeft | threadSpatialRight | threadSpatialDown:
		return "┳"
	case threadSpatialLeft | threadSpatialRight | threadSpatialUp:
		return "┻"
	case threadSpatialUp | threadSpatialDown | threadSpatialRight:
		return "┣"
	case threadSpatialUp | threadSpatialDown | threadSpatialLeft:
		return "┫"
	case threadSpatialLeft | threadSpatialRight | threadSpatialUp | threadSpatialDown:
		return "╋"
	default:
		return " "
	}
}

func (c *threadSpatialCanvas) routeConflictCount() int {
	count := 0
	for row := range c.cells {
		for column := range c.cells[row] {
			if c.cells[row][column].conflict {
				count++
			}
		}
	}
	return count
}

func threadSpatialRouteConflictSummary(canvas *threadSpatialCanvas) string {
	conflicts := canvas.routeConflictCount()
	if conflicts == 0 {
		return ""
	}
	label := fmt.Sprintf("%d routing conflict", conflicts)
	if conflicts != 1 {
		label += "s"
	}
	return label
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
		if text == "" {
			switch {
			case cell.conflict:
				text = "!"
			case cell.overlap:
				text = "≋"
			case cell.crossing:
				text = "╳"
			case cell.shared:
				text = threadSpatialSharedConnectorGlyph(cell.connector)
			case cell.accent && cell.connector != 0:
				text = threadSpatialFocusConnectorGlyph(cell.connector)
			case cell.connector != 0:
				text = threadSpatialConnectorGlyph(cell.connector)
			}
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
	// The narrow explanation allocates no graph canvas, so it intentionally
	// precedes prototype capacity checks. Small terminals can always retreat to
	// the wave reader even when the spatial projection itself is oversized.
	if width < threadSpatialMinWidth || height < threadSpatialMinHeight {
		return renderThreadSpatialNarrow(projection, layout, selectedTaskID, width, height, s)
	}
	if issue := threadSpatialCapacityIssue(layout, len(projection.Edges)); issue != "" {
		return renderThreadSpatialCapacityFallback(projection, layout, selectedTaskID, issue, width, height, s)
	}

	fixedRows := 8 // header + two legend rows + five-row focus card
	canvasHeight := max(1, height-fixedRows)
	selected, selectedOK := spatialPlacement(layout, selectedTaskID)
	panX, panY := 0, 0
	if selectedOK {
		panX = min(max(selected.x+threadSpatialNodeWidth/2-width/2, 0), max(layout.width-width, 0))
		panY = min(max(selected.y+threadSpatialNodeSlotHeight/2-canvasHeight/2, 0), max(layout.height-canvasHeight, 0))
	}
	canvas := renderThreadSpatialCanvasWindow(
		projection, layout, selectedTaskID,
		panX, panY, width, canvasHeight,
	)
	annotateThreadSpatialRouteBoundaries(canvas, layout, selectedTaskID, panX, panY, width, canvasHeight)

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
	if routeIssue := threadSpatialRouteConflictSummary(canvas); routeIssue != "" {
		header += " · " + routeIssue
	}
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

func annotateThreadSpatialRouteBoundaries(
	canvas *threadSpatialCanvas,
	layout threadSpatialLayout,
	selected string,
	panX, panY, width, height int,
) {
	if width <= 0 || height <= 0 {
		return
	}
	visible := func(point threadSpatialPoint) bool {
		return point.x >= panX && point.x < panX+width && point.y >= panY && point.y < panY+height
	}
	type boundaryKey struct {
		side  byte
		row   int
		kind  byte
		arrow rune
	}
	type boundaryGroup struct {
		anchorX int
		aliases []string
	}
	groups := make(map[boundaryKey]boundaryGroup)
	add := func(endpoint, anchor threadSpatialPoint, kind byte, arrow rune, alias string) {
		side := threadSpatialBoundarySide(endpoint, anchor, panX, panY, width, height)
		if side == 0 {
			return
		}
		row := anchor.y
		switch side {
		case 't':
			row = panY
		case 'b':
			row = panY + height - 1
		}
		key := boundaryKey{side: side, row: row, kind: kind, arrow: arrow}
		group := groups[key]
		if len(group.aliases) == 0 || anchor.x < group.anchorX {
			group.anchorX = anchor.x
		}
		group.aliases = append(group.aliases, alias)
		groups[key] = group
	}
	for _, route := range layout.routes {
		if (route.edge.From != selected && route.edge.To != selected) || len(route.segments) == 0 {
			continue
		}
		first, last, found := threadSpatialVisibleRouteExtent(route, visible)
		if !found {
			continue
		}
		if source := route.segments[0].from; !visible(source) {
			add(source, first, 's', 0, threadSpatialRouteAlias(layout, route.edge.From))
		}
		if !visible(route.arrow) {
			add(route.arrow, last, 't', route.arrowRune, threadSpatialRouteAlias(layout, route.edge.To))
		}
	}
	keys := make([]boundaryKey, 0, len(groups))
	for key := range groups {
		keys = append(keys, key)
	}
	sort.SliceStable(keys, func(i, j int) bool {
		if keys[i].side != keys[j].side {
			return keys[i].side < keys[j].side
		}
		if keys[i].row != keys[j].row {
			return keys[i].row < keys[j].row
		}
		if keys[i].kind != keys[j].kind {
			return keys[i].kind < keys[j].kind
		}
		return keys[i].arrow < keys[j].arrow
	})
	for _, key := range keys {
		group := groups[key]
		label := threadSpatialBoundaryLabel(key.kind, key.arrow, group.aliases)
		putThreadSpatialBoundaryLabel(canvas, key.side, group.anchorX, key.row, label, panX, width)
	}
}

func threadSpatialBoundarySide(
	endpoint, anchor threadSpatialPoint,
	panX, panY, width, height int,
) byte {
	switch {
	case anchor.x == panX && endpoint.x < panX:
		return 'l'
	case anchor.x == panX+width-1 && endpoint.x >= panX+width:
		return 'r'
	case anchor.y == panY && endpoint.y < panY:
		return 't'
	case anchor.y == panY+height-1 && endpoint.y >= panY+height:
		return 'b'
	case endpoint.x < panX:
		return 'l'
	case endpoint.x >= panX+width:
		return 'r'
	case endpoint.y < panY:
		return 't'
	case endpoint.y >= panY+height:
		return 'b'
	default:
		return 0
	}
}

func threadSpatialBoundaryLabel(kind byte, arrow rune, aliases []string) string {
	seen := make(map[string]bool, len(aliases))
	unique := make([]string, 0, len(aliases))
	for _, alias := range aliases {
		if alias == "" || seen[alias] {
			continue
		}
		seen[alias] = true
		unique = append(unique, alias)
	}
	displayed := unique
	if len(unique) > 3 {
		displayed = append(append([]string(nil), unique[:3]...), fmt.Sprintf("+%d", len(unique)-3))
	}
	body := "[" + strings.Join(displayed, ",") + "]"
	switch {
	case kind == 's':
		return body + "…"
	case arrow == '◀':
		return body + "◀…"
	default:
		return "…▶" + body
	}
}

func threadSpatialVisibleRouteExtent(
	route threadSpatialRoute,
	visible func(threadSpatialPoint) bool,
) (threadSpatialPoint, threadSpatialPoint, bool) {
	var first, last threadSpatialPoint
	found := false
	for _, segment := range route.segments {
		dx, dy := sign(segment.to.x-segment.from.x), sign(segment.to.y-segment.from.y)
		point := segment.from
		for {
			if visible(point) {
				if !found {
					first = point
					found = true
				}
				last = point
			}
			if point == segment.to {
				break
			}
			point.x += dx
			point.y += dy
		}
	}
	return first, last, found
}

func threadSpatialRouteAlias(layout threadSpatialLayout, taskID string) string {
	if node, ok := spatialPlacement(layout, taskID); ok && node.alias != "" {
		return node.alias
	}
	return "?"
}

func putThreadSpatialBoundaryLabel(
	canvas *threadSpatialCanvas,
	side byte,
	anchorX, row int,
	label string,
	panX, width int,
) {
	if row < 0 || row >= len(canvas.cells) || label == "" {
		return
	}
	labelWidth := ansi.StringWidth(label)
	viewportWidth := min(width, canvas.width-panX)
	if viewportWidth <= 0 || labelWidth > viewportWidth {
		return
	}
	left, right := panX, panX+viewportWidth-labelWidth
	desired := min(max(anchorX-labelWidth/2, left), right)
	switch side {
	case 'l':
		desired = left
	case 'r':
		desired = right
	}
	available := func(x int) bool {
		if x < left || x > right {
			return false
		}
		for column := x; column < x+labelWidth; column++ {
			cell := canvas.cells[row][column]
			if cell.text != "" || cell.continuation {
				return false
			}
		}
		return true
	}
	for distance := 0; distance <= width; distance++ {
		for _, x := range []int{desired - distance, desired + distance} {
			if available(x) {
				canvas.putAccentText(x, row, label, true)
				return
			}
		}
	}
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

func renderThreadSpatialCanvas(_ core.ThreadGraphProjection, layout threadSpatialLayout, selected string) *threadSpatialCanvas {
	return renderThreadSpatialCanvasWindow(
		core.ThreadGraphProjection{}, layout, selected,
		0, 0, layout.width, layout.height,
	)
}

func renderThreadSpatialCanvasWindow(
	_ core.ThreadGraphProjection,
	layout threadSpatialLayout,
	selected string,
	panX, panY, width, height int,
) *threadSpatialCanvas {
	canvas := newThreadSpatialCanvas(layout.width, layout.height)
	for column, label := range layout.columnLabels {
		x := layout.columnX[column]
		canvas.putText(x, 0, truncate(label, threadSpatialNodeWidth), theme.ColorGray, false)
	}
	routes := threadSpatialRoutesForWindow(layout, selected, panX, panY, width, height)
	for _, route := range routes {
		drawThreadSpatialRouteSegments(canvas, route, route.edge.From == selected || route.edge.To == selected)
	}
	for _, route := range routes {
		drawThreadSpatialRouteAdornments(canvas, route, route.edge.From == selected || route.edge.To == selected)
	}
	for _, node := range layout.nodes {
		drawThreadSpatialNode(canvas, node, node.node.TaskID == selected)
	}
	// Counts normally replace one cell of a shared endpoint stub. Drawing them
	// after nodes also permits the explicit node-border fallback used when every
	// safe stub cell is congested; endpoint multiplicity is never silently lost.
	drawThreadSpatialRouteCounts(canvas, layout, routes, selected)
	return canvas
}

func threadSpatialRoutesForWindow(
	layout threadSpatialLayout,
	selected string,
	panX, panY, width, height int,
) []threadSpatialRoute {
	visibleNode := func(taskID string) bool {
		node, ok := layout.byID[taskID]
		return ok && node.x < panX+width && node.x+threadSpatialNodeWidth > panX &&
			node.y < panY+height && node.y+threadSpatialNodeSlotHeight > panY
	}
	routes := make([]threadSpatialRoute, 0, len(layout.routes))
	for _, route := range layout.routes {
		if route.edge.From == selected || route.edge.To == selected ||
			visibleNode(route.edge.From) || visibleNode(route.edge.To) {
			routes = append(routes, route)
		}
	}
	return routes
}

func drawThreadSpatialRouteSegments(canvas *threadSpatialCanvas, route threadSpatialRoute, selected bool) {
	for _, segment := range route.segments {
		style := threadSpatialRouteStyle{
			id: route.id, from: route.edge.From, to: route.edge.To,
			selected: selected,
		}
		switch {
		case segment.from.y == segment.to.y:
			canvas.routeHorizontal(segment.from.x, segment.to.x, segment.from.y, style)
		case segment.from.x == segment.to.x:
			canvas.routeVertical(segment.from.x, segment.from.y, segment.to.y, style)
		}
	}
}

func drawThreadSpatialRouteAdornments(canvas *threadSpatialCanvas, route threadSpatialRoute, selected bool) {
	style := threadSpatialRouteStyle{id: route.id, from: route.edge.From, to: route.edge.To, selected: selected}
	canvas.putRouteArrow(route.arrow.x, route.arrow.y, route.arrowRune, style)
}

func drawThreadSpatialRouteCounts(
	canvas *threadSpatialCanvas,
	layout threadSpatialLayout,
	routes []threadSpatialRoute,
	selected string,
) {
	type routeEndKey struct {
		taskID string
		kind   byte
		side   rune
	}
	type routeEnd struct {
		point    threadSpatialPoint
		fallback threadSpatialPoint
		step     int
		maxSteps int
		count    int
	}
	ends := make(map[routeEndKey]routeEnd)
	order := make([]routeEndKey, 0)
	add := func(key routeEndKey, point, fallback threadSpatialPoint, step, maxSteps int) {
		end := ends[key]
		if end.count == 0 {
			order = append(order, key)
			end.point, end.fallback, end.step, end.maxSteps = point, fallback, step, maxSteps
		} else {
			end.maxSteps = min(end.maxSteps, maxSteps)
		}
		end.count++
		ends[key] = end
	}
	for _, route := range routes {
		if len(route.segments) == 0 {
			continue
		}
		first := route.segments[0]
		sourcePoint := first.from
		sourcePoint.x += 2
		sourceFallback := threadSpatialCountFallback(layout, route.edge.From, selected, 's', '▶')
		sourceStep := sign(first.to.x - first.from.x)
		add(
			routeEndKey{taskID: route.edge.From, kind: 's', side: '▶'},
			sourcePoint, sourceFallback, sourceStep, max(0, abs(first.to.x-sourcePoint.x)-1),
		)

		last := route.segments[len(route.segments)-1]
		targetPoint := route.arrow
		if route.arrowRune == '▶' {
			targetPoint.x--
		} else {
			targetPoint.x++
		}
		targetFallback := threadSpatialCountFallback(layout, route.edge.To, selected, 't', route.arrowRune)
		targetStep := sign(last.from.x - route.arrow.x)
		add(
			routeEndKey{taskID: route.edge.To, kind: 't', side: route.arrowRune},
			targetPoint, targetFallback, targetStep, max(0, abs(last.from.x-targetPoint.x)-1),
		)
	}
	for _, key := range order {
		end := ends[key]
		if _, ok := canvas.putRouteCountAlong(end.point, end.step, end.maxSteps, end.count, key.taskID == selected); !ok {
			canvas.putRouteCountFallback(end.fallback, end.count, key.taskID == selected)
		}
	}
}

func threadSpatialCountFallback(
	layout threadSpatialLayout,
	taskID, selected string,
	kind byte,
	side rune,
) threadSpatialPoint {
	placement, ok := layout.byID[taskID]
	if !ok {
		return threadSpatialPoint{x: -1, y: -1}
	}
	topY := placement.y + 1
	if taskID == selected {
		topY = placement.y
	}
	x := placement.x + threadSpatialNodeWidth/2
	if kind == 't' {
		if side == '▶' {
			x = placement.x + 2
		} else {
			x = placement.x + threadSpatialNodeWidth - 3
		}
	}
	return threadSpatialPoint{x: x, y: topY}
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
	// the box's semantic status color. Focus is conveyed by the expanded geometry
	// and accent-colored touching edges rather than repainting task state.
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
	return s.dim("roles") + "  ┌ member  ╔ gate  " +
		s.accent("━ focus route") + "  " +
		s.fg(theme.ColorYellow, "▶ direction  2 fan")
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
	for _, edge := range orderedThreadGraphEdges(projection.Edges) {
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
