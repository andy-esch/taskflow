package tui

import (
	"fmt"
	"slices"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/andy-esch/taskflow/internal/core"
	"github.com/andy-esch/taskflow/internal/domain"
	"github.com/andy-esch/taskflow/internal/theme"
)

func threadFocusProjection() core.ThreadGraphProjection {
	task := func(id string, status domain.Status, description string) domain.Task {
		return domain.Task{ID: id, Slug: id, Status: status, Description: description}
	}
	state := func(id string, role core.LifecycleRole, gate core.GateState, eligible bool) core.TaskGraphState {
		return core.TaskGraphState{TaskID: id, Role: role, Gate: gate, Eligible: eligible}
	}
	tasks := map[string]domain.Task{
		"g": task("g", domain.StatusCompleted, "external gate"),
		"a": task("a", domain.StatusCompleted, "completed prerequisite"),
		"b": task("b", domain.StatusInProgress, "active focal task"),
		"c": task("c", domain.StatusNextUp, "first dependent"),
		"d": task("d", domain.StatusReadyToStart, "second dependent"),
		"e": task("e", domain.StatusNextUp, "two hops away"),
		"f": task("f", domain.StatusDeferred, "isolated member"),
		"x": task("x", domain.StatusCompleted, "second prerequisite"),
	}
	states := map[string]core.TaskGraphState{
		"g": state("g", core.RoleNominallyComplete, core.GateClear, false),
		"a": state("a", core.RoleNominallyComplete, core.GateClear, false),
		"b": state("b", core.RoleInFlight, core.GateClear, false),
		"c": state("c", core.RoleCandidate, core.GateClear, true),
		"d": state("d", core.RoleCandidate, core.GateClear, true),
		"e": state("e", core.RoleQueued, core.GateBlocked, false),
		"f": state("f", core.RoleParked, core.GateClear, false),
		"x": state("x", core.RoleNominallyComplete, core.GateClear, false),
		"u": state("u", core.RoleUnknown, core.GateBroken, false),
	}
	members := make([]core.ThreadTaskView, 0, 8)
	for _, id := range []string{"a", "b", "c", "d", "e", "f", "x"} {
		members = append(members, core.ThreadTaskView{Role: core.ThreadTaskMember, Task: tasks[id], State: states[id]})
	}
	members = append(members, core.ThreadTaskView{
		Role: core.ThreadTaskMember, Task: domain.Task{ID: "u"}, State: states["u"],
	})
	view := core.ThreadView{
		Thread:  domain.Thread{ID: "thread", Slug: "focus-thread", Status: domain.ThreadStatusInProgress},
		Members: members,
		ExternalGates: []core.ThreadExternalGate{{
			ThreadTaskView: core.ThreadTaskView{Role: core.ThreadTaskExternalGate, Task: tasks["g"], State: states["g"]},
		}},
		GraphHealth: core.GraphHealthy, ProjectionHealth: core.GraphHealthy,
	}
	view.Frontier = []core.ThreadTaskView{members[2], members[3]}
	nodes := make([]core.ThreadGraphNode, 0, 9)
	for _, id := range []string{"a", "b", "c", "d", "e", "f", "g", "x"} {
		role := core.ThreadTaskMember
		if id == "g" {
			role = core.ThreadTaskExternalGate
		}
		nodes = append(nodes, core.ThreadGraphNode{
			TaskID: id, Label: id, Description: tasks[id].Description,
			Status: tasks[id].Status, Role: role, State: states[id],
		})
	}
	nodes = append(nodes, core.ThreadGraphNode{
		TaskID: "u", Label: "unreadable", Role: core.ThreadTaskMember, State: states["u"],
	})
	return core.ThreadGraphProjection{
		View: view, Nodes: nodes,
		Edges: []core.ThreadGraphEdge{
			{From: "a", To: "b"}, {From: "b", To: "c"}, {From: "b", To: "d"},
			{From: "b", To: "u"}, {From: "c", To: "e"}, {From: "g", To: "a"}, {From: "x", To: "b"},
		},
		Waves: []core.ThreadGraphWave{
			{Index: 1, TaskIDs: []string{"a", "f", "x"}}, {Index: 2, TaskIDs: []string{"b"}},
			{Index: 3, TaskIDs: []string{"c", "d"}}, {Index: 4, TaskIDs: []string{"e"}},
		},
		TopologyComplete: true,
	}
}

func TestThreadSpatialFocusUsesPreferredIdentityAndRestoresFullGraph(t *testing.T) {
	projection := threadFocusProjection()
	detail := newThreadDetail(projection, "", "", "").withDetailView(string(threadDetailSpatial)).(threadDetail)
	if got := detail.detailSelectionKey(); got != "b" {
		t.Fatalf("first spatial selection=%q want in-flight b", got)
	}
	fullCache := detail.spatial

	focusedContent, err := detail.toggleDetailLocalFocus()
	if err != nil {
		t.Fatal(err)
	}
	focused := focusedContent.(threadDetail)
	if focused.focus == nil || focused.focus.projection.Scope == nil {
		t.Fatal("focus did not retain a bounded projection and scope")
	}
	scope := focused.focus.projection.Scope
	if scope.FocalTaskID != "b" || scope.Depth != 1 || scope.ShownNodes != 6 || scope.TotalNodes != 9 ||
		scope.HiddenNodes != 3 || len(scope.BoundaryEdges) != 2 {
		t.Fatalf("focus scope=%+v", scope)
	}
	if focused.spatial != fullCache || focused.focus.spatial == fullCache {
		t.Fatal("focus replaced the cached full layout instead of using its own bounded cache")
	}
	fullPrepared := detail.spatialPreparedForViewport(180)
	focusedPrepared := focused.spatialPreparedForViewport(180)
	if fullPrepared.layout == nil || focusedPrepared.layout == nil {
		t.Fatal("focus or full graph did not produce a spatial layout")
	}
	if got, want := focusedPrepared.layout.byID["b"].alias, fullPrepared.layout.byID["b"].alias; got != want {
		t.Fatalf("focused focal alias=%q want stable full-graph alias %q", got, want)
	}
	shown := make(map[string]bool, len(focused.focus.projection.Nodes))
	for _, node := range focused.focus.projection.Nodes {
		shown[node.TaskID] = true
	}
	if len(focusedPrepared.layout.byID) != len(shown) {
		t.Fatalf("focused layout has %d nodes want %d scoped nodes", len(focusedPrepared.layout.byID), len(shown))
	}
	for taskID := range shown {
		if _, ok := focusedPrepared.layout.byID[taskID]; !ok {
			t.Fatalf("focused layout omitted scoped task %q", taskID)
		}
	}

	rendered := ansi.Strip(focused.renderDetail(180, 40, &testStyles))
	for _, want := range []string{
		"ZOOMED · ONE-HOP", "╭─ ZOOMED · ONE-HOP", "6/9 shown", "3 hidden", "2 boundary edge(s)", "z full graph", "graph healthy", "unreadable",
	} {
		if !strings.Contains(rendered, want) {
			t.Errorf("focused render omitted %q:\n%s", want, rendered)
		}
	}
	if fullRendered := ansi.Strip(detail.renderDetail(180, 40, &testStyles)); strings.Contains(fullRendered, "ZOOMED") {
		t.Fatalf("full graph falsely advertised a zoomed subgraph:\n%s", fullRendered)
	}
	cycled := focused.withDetailView(string(threadDetailSummary)).(threadDetail)
	cycled = cycled.withDetailView(string(threadDetailSpatial)).(threadDetail)
	if cycled.detailLocalFocusActive() {
		t.Fatal("leaving and re-entering the spatial view retained stale local focus")
	}
	capacity := focused.activeProjection()
	capacityRendered := ansi.Strip(renderThreadSpatialPrepared(
		capacity,
		threadSpatialPrepared{issue: "review capacity probe", nodeCount: 600, edgeCount: 1, fallbackTaskID: "b"},
		"", "b", 120, 30, &testStyles,
	))
	if !strings.Contains(capacityRendered, "ZOOMED · ONE-HOP") || !strings.Contains(capacityRendered, "z for the full graph") {
		t.Fatalf("focused capacity fallback omitted mode or return guidance:\n%s", capacityRendered)
	}

	focused = focused.withDetailSelection("c").(threadDetail)
	fullContent, err := focused.toggleDetailLocalFocus()
	if err != nil {
		t.Fatal(err)
	}
	full := fullContent.(threadDetail)
	if full.focus != nil || full.detailSelectionKey() != "b" || full.spatial != fullCache {
		t.Fatalf("full graph restore focus=%v selected=%q cache-preserved=%v", full.focus != nil, full.detailSelectionKey(), full.spatial == fullCache)
	}
}

func TestThreadSpatialEntryReanchorsToWorkWhileReloadRestorationRemainsExplicit(t *testing.T) {
	projection := threadFocusProjection()
	detail := newThreadDetail(projection, "", "", "").withDetailView(string(threadDetailSpatial)).(threadDetail)
	detail = detail.withDetailSelection("g").(threadDetail)
	detail = detail.withDetailView(string(threadDetailSummary)).(threadDetail)
	detail = detail.withDetailView(string(threadDetailSpatial)).(threadDetail)
	if got := detail.detailSelectionKey(); got != "b" {
		t.Fatalf("deliberate spatial re-entry selected %q want in-flight b", got)
	}

	// With no work in flight or on the supplied frontier, use the newest
	// portable task activity stamp and never default to a historical gate merely
	// because its stable ID sorts first.
	recent := threadFocusProjection()
	recent.View.Frontier = nil
	for index := range recent.View.Members {
		recent.View.Members[index].State.Role = core.RoleQueued
		recent.View.Members[index].Task.Updated = "2026-09-10"
		if recent.View.Members[index].Task.CanonicalID() == "e" {
			recent.View.Members[index].Task.Updated = "2026-09-12"
		}
	}
	for index := range recent.Nodes {
		if recent.Nodes[index].Role == core.ThreadTaskMember {
			recent.Nodes[index].State.Role = core.RoleQueued
		}
	}
	if got := threadSpatialPreferredTaskID(recent); got != "e" {
		t.Fatalf("recent-activity fallback=%q want member e", got)
	}

	// Malformed tool-managed activity cannot sort above a valid date; a valid
	// created date remains the portable fallback when updated_at is corrupt.
	for index := range recent.View.Members {
		recent.View.Members[index].Task.Updated = "someday"
		recent.View.Members[index].Task.Created = "2026-09-10"
		if recent.View.Members[index].Task.CanonicalID() == "d" {
			recent.View.Members[index].Task.Created = "2026-09-12"
		}
	}
	if got := threadSpatialPreferredTaskID(recent); got != "d" {
		t.Fatalf("validated created-activity fallback=%q want member d", got)
	}

	// Candidate selection scans the supplied view rather than an arbitrary
	// ID-sorted prefix, so capacity fallback still identifies active work.
	large := threadFocusProjection()
	for index := range large.View.Members {
		large.View.Members[index].State.Role = core.RoleQueued
	}
	for index := range large.Nodes {
		if large.Nodes[index].Role == core.ThreadTaskMember {
			large.Nodes[index].State.Role = core.RoleQueued
		}
	}
	for index := 0; index < threadSpatialMaxNodes; index++ {
		id := fmt.Sprintf("bulk-%04d", index)
		task := domain.Task{ID: id, Slug: id, Status: domain.StatusNextUp, Created: "2026-09-01"}
		state := core.TaskGraphState{TaskID: id, Role: core.RoleQueued, Gate: core.GateClear}
		large.View.Members = append(large.View.Members, core.ThreadTaskView{Role: core.ThreadTaskMember, Task: task, State: state})
		large.Nodes = append(large.Nodes, core.ThreadGraphNode{TaskID: id, Label: id, Role: core.ThreadTaskMember, State: state})
	}
	active := domain.Task{ID: "zzzz-active", Slug: "zzzz-active", Status: domain.StatusInProgress, Created: "2026-09-12"}
	activeState := core.TaskGraphState{TaskID: active.ID, Role: core.RoleInFlight, Gate: core.GateClear}
	large.View.Members = append(large.View.Members, core.ThreadTaskView{Role: core.ThreadTaskMember, Task: active, State: activeState})
	large.Nodes = append(large.Nodes, core.ThreadGraphNode{TaskID: active.ID, Label: active.Slug, Role: core.ThreadTaskMember, State: activeState})
	if got := threadSpatialPreferredTaskID(large); got != active.ID {
		t.Fatalf("large-Thread fallback=%q want in-flight %q", got, active.ID)
	}

	onlyGateReadable := threadFocusProjection()
	onlyGateReadable.View.Frontier = nil
	for index := range onlyGateReadable.View.Members {
		onlyGateReadable.View.Members[index].Task.Slug = ""
		onlyGateReadable.View.Members[index].State.Role = core.RoleUnknown
	}
	for index := range onlyGateReadable.Nodes {
		if onlyGateReadable.Nodes[index].Role == core.ThreadTaskMember {
			onlyGateReadable.Nodes[index].State.Role = core.RoleUnknown
		}
	}
	if got := threadSpatialPreferredTaskID(onlyGateReadable); got != "g" {
		t.Fatalf("readable-gate fallback=%q want external gate g", got)
	}
}

func TestThreadSpatialFocusSurvivesCoherentReloadByCanonicalIdentity(t *testing.T) {
	projection := threadFocusProjection()
	detail := newThreadDetail(projection, "", "", "").withDetailView(string(threadDetailSpatial)).(threadDetail)
	focused, err := detail.toggleDetailLocalFocus()
	if err != nil {
		t.Fatal(err)
	}
	focused = focused.(threadDetail).withDetailSelection("c")

	pane := newDetailPane(&testStyles)
	pane.SetSize(160, 35)
	pane.SetContent("thread", focused)
	fresh := threadFocusProjection()
	fresh.Nodes[2].Label = "c-renamed"
	pane.SetContent("thread", newThreadDetail(fresh, "", "", ""))
	reloaded := pane.content.(threadDetail)
	if reloaded.focus == nil || reloaded.focus.focalTaskID != "b" || reloaded.detailSelectionKey() != "c" {
		t.Fatalf("reload focus=%+v selected=%q", reloaded.focus, reloaded.detailSelectionKey())
	}
	if node, ok := spatialPlacement(*reloaded.spatialPrepared().layout, "c"); !ok || node.node.Label != "c-renamed" {
		t.Fatalf("reload retained stale node: %+v present=%v", node.node, ok)
	}
	if got, want := len(reloaded.spatialPrepared().layout.byID), reloaded.focus.projection.Scope.ShownNodes; got != want {
		t.Fatalf("reload focused layout has %d nodes want scoped %d", got, want)
	}
	if pane.selectDetailTask("removed-target") || pane.content.(threadDetail).detailSelectionKey() != "c" {
		t.Fatal("a stale chooser target replaced the preserved canonical selection")
	}

	// Deleting the focal task must fail open even when another readable task's
	// slug happens to contain the old ID and the public selector could resolve it
	// as a fuzzy user reference.
	deleted := threadFocusProjection()
	deleted.Nodes = slices.DeleteFunc(deleted.Nodes, func(node core.ThreadGraphNode) bool { return node.TaskID == "b" })
	deleted.View.Members = slices.DeleteFunc(deleted.View.Members, func(member core.ThreadTaskView) bool {
		return member.State.TaskID == "b"
	})
	deleted.Edges = slices.DeleteFunc(deleted.Edges, func(edge core.ThreadGraphEdge) bool {
		return edge.From == "b" || edge.To == "b"
	})
	deleted.View.Members[0].Task.Slug = "follow-up-to-b"
	pane.SetContent("thread", newThreadDetail(deleted, "", "", ""))
	withoutFocal := pane.content.(threadDetail)
	if withoutFocal.detailLocalFocusActive() || withoutFocal.detailSelectionKey() == "b" {
		t.Fatalf("deleted focal restored fuzzy focus=%v selection=%q", withoutFocal.detailLocalFocusActive(), withoutFocal.detailSelectionKey())
	}
}

func TestThreadSpatialScopeCardUsesOnlyEmptyCanvasSpace(t *testing.T) {
	projection, err := core.SelectThreadGraphNeighborhood(threadFocusProjection(), "b", 1)
	if err != nil {
		t.Fatal(err)
	}
	blank := newThreadSpatialCanvas(80, 20)
	if !annotateThreadSpatialScopeCard(blank, projection, 0, 0, 80, 20) {
		t.Fatal("blank canvas did not accept the scope card")
	}
	rendered := ansi.Strip(blank.renderLine(1, &testStyles) + "\n" + blank.renderLine(2, &testStyles) + "\n" + blank.renderLine(3, &testStyles))
	if !strings.Contains(rendered, "ZOOMED · ONE-HOP") || !strings.Contains(rendered, "shown") ||
		!strings.Contains(rendered, "boundary edge") {
		t.Fatalf("scope card omitted evidence:\n%s", rendered)
	}

	occupied := newThreadSpatialCanvas(80, 20)
	for row := 0; row < 20; row++ {
		occupied.putText(0, row, strings.Repeat("·", 80), theme.ColorGray, false)
	}
	if annotateThreadSpatialScopeCard(occupied, projection, 0, 0, 80, 20) {
		t.Fatal("scope card covered an occupied canvas instead of yielding")
	}
	if line := ansi.Strip(occupied.renderLine(1, &testStyles)); line != strings.Repeat("·", 80) {
		t.Fatalf("yielding scope card altered graph ink: %q", line)
	}
	if annotateThreadSpatialScopeCard(newThreadSpatialCanvas(80, 20), projection, 0, 0, 20, 6) {
		t.Fatal("scope card ignored its hostile-size guard")
	}
}

func TestThreadSpatialFocusDirectionalBranchesUseChooserAndDeadEndsExplain(t *testing.T) {
	m, _ := threadModel(t)
	tm, _ := m.Update(tea.WindowSizeMsg{Width: 160, Height: 36})
	m = openThreads(t, tm.(Model))
	m.detail = newDetailPane(m.st)
	m.detail.SetSize(160-m.st.paneHFrame, 36-m.st.paneVFrame)
	detail := newThreadDetail(threadFocusProjection(), "", "", "").withDetailView(string(threadDetailSpatial))
	m.detail.SetContent(m.selectedKey(), detail)
	m.focus, m.zoom, m.immersiveZoom = focusDetail, true, true

	tm, _ = m.Update(press("z"))
	m = tm.(Model)
	if !selectedThreadDetail(t, m).detailLocalFocusActive() || !strings.Contains(ansi.Strip(m.footer()), "z full graph") {
		t.Fatalf("z did not enter discoverable focus: footer=%q", ansi.Strip(m.footer()))
	}

	tm, _ = m.Update(press("l"))
	m = tm.(Model)
	if !m.direction.active || len(m.direction.tasks) != 2 || m.direction.label != "dependent" {
		t.Fatalf("fan-out did not open chooser: %+v", m.direction)
	}
	tm, _ = m.Update(press("j"))
	m = tm.(Model)
	tm, _ = m.Update(press("enter"))
	m = tm.(Model)
	if m.direction.active || selectedThreadDetail(t, m).detailSelectionKey() != "d" {
		t.Fatalf("chooser did not select d: active=%v selected=%q", m.direction.active, selectedThreadDetail(t, m).detailSelectionKey())
	}

	tm, _ = m.Update(press("h"))
	m = tm.(Model)
	if selectedThreadDetail(t, m).detailSelectionKey() != "b" {
		t.Fatalf("single prerequisite did not move immediately: %q", selectedThreadDetail(t, m).detailSelectionKey())
	}
	tm, _ = m.Update(press("h"))
	m = tm.(Model)
	if !m.direction.active || len(m.direction.tasks) != 2 || m.direction.label != "prerequisite" {
		t.Fatalf("fan-in did not open chooser: %+v", m.direction)
	}
	tm, _ = m.Update(press("enter"))
	m = tm.(Model)
	if selectedThreadDetail(t, m).detailSelectionKey() != "a" {
		t.Fatalf("prerequisite chooser did not select a: %q", selectedThreadDetail(t, m).detailSelectionKey())
	}
	tm, _ = m.Update(press("h"))
	m = tm.(Model)
	if !m.flashErr || !strings.Contains(m.flash, "no readable direct prerequisite") {
		t.Fatalf("boundary dead end was silent: err=%v flash=%q", m.flashErr, m.flash)
	}
}

func TestThreadSpatialDirectionalChooserRevalidatesAgainstCurrentGraph(t *testing.T) {
	m, _ := threadModel(t)
	tm, _ := m.Update(tea.WindowSizeMsg{Width: 160, Height: 36})
	m = openThreads(t, tm.(Model))
	m.detail = newDetailPane(m.st)
	m.detail.SetSize(160-m.st.paneHFrame, 36-m.st.paneVFrame)
	m.detail.SetContent(m.selectedKey(), newThreadDetail(threadFocusProjection(), "", "", "").withDetailView(string(threadDetailSpatial)))
	m.focus, m.zoom, m.immersiveZoom = focusDetail, true, true
	tm, _ = m.Update(press("z"))
	m = tm.(Model)
	tm, _ = m.Update(press("l"))
	m = tm.(Model)
	if !m.direction.active || m.direction.tasks[0].CanonicalID() != "c" {
		t.Fatalf("expected dependent chooser before reload: %+v", m.direction)
	}

	// The selected snapshot target remains readable, but reversing its supplied
	// edge means it no longer represents the direction the user chose.
	fresh := threadFocusProjection()
	for index := range fresh.Edges {
		if fresh.Edges[index] == (core.ThreadGraphEdge{From: "b", To: "c"}) {
			fresh.Edges[index] = core.ThreadGraphEdge{From: "c", To: "b"}
		}
	}
	m.detail.SetContent(m.detail.loadedKey, newThreadDetail(fresh, "", "", ""))
	tm, _ = m.Update(press("enter"))
	m = tm.(Model)
	if selectedThreadDetail(t, m).detailSelectionKey() != "b" || !m.flashErr ||
		!strings.Contains(m.flash, "no longer available") {
		t.Fatalf("stale direction committed: selected=%q err=%v flash=%q",
			selectedThreadDetail(t, m).detailSelectionKey(), m.flashErr, m.flash)
	}
}

func TestThreadSpatialReloadClosesChooserAndZoomRoutesVisibleShellState(t *testing.T) {
	m, _ := threadModel(t)
	tm, _ := m.Update(tea.WindowSizeMsg{Width: 160, Height: 36})
	m = openThreads(t, tm.(Model))
	m.detail = newDetailPane(m.st)
	m.detail.SetSize(160-m.st.paneHFrame, 36-m.st.paneVFrame)
	m.detail.SetContent(m.selectedKey(), newThreadDetail(threadFocusProjection(), "", "", "").withDetailView(string(threadDetailSpatial)))
	m.focus, m.zoom, m.immersiveZoom = focusDetail, true, true
	tm, _ = m.Update(press("z"))
	m = tm.(Model)
	tm, _ = m.Update(press("l"))
	m = tm.(Model)
	if !m.direction.active {
		t.Fatal("expected chooser before coherent reload")
	}
	tm, _ = m.Update(detailMsg{
		kind: entityThreads, id: m.selectedKey(), gen: m.detailGen,
		content: newThreadDetail(threadFocusProjection(), "", "", ""),
	})
	m = tm.(Model)
	if m.direction.active || !strings.Contains(m.flash, "choose a direction again") {
		t.Fatalf("reload retained stale chooser: active=%v flash=%q", m.direction.active, m.flash)
	}

	// Leaving the immersive shell exposes the split view. There z must restore
	// visible full-screen detail, not mutate a hidden graph lens.
	tm, _ = m.Update(press("tab"))
	m = tm.(Model)
	if m.zoom || m.focus != focusList {
		t.Fatalf("tab did not expose split list state: zoom=%v focus=%v", m.zoom, m.focus)
	}
	tm, _ = m.Update(press("z"))
	m = tm.(Model)
	if !m.zoom || m.focus != focusDetail || !selectedThreadDetail(t, m).detailLocalFocusActive() {
		t.Fatalf("split-view z changed hidden lens: zoom=%v focus=%v local=%v",
			m.zoom, m.focus, selectedThreadDetail(t, m).detailLocalFocusActive())
	}

	// An unreadable selected node keeps z in the local graph interaction and
	// explains the failed focus instead of silently exiting full-screen.
	tm, _ = m.Update(press("z"))
	m = tm.(Model)
	m.detail.content = selectedThreadDetail(t, m).withDetailSelection("u")
	m.detail.render()
	tm, _ = m.Update(press("z"))
	m = tm.(Model)
	if !m.zoom || m.focus != focusDetail || !m.flashErr || !strings.Contains(m.flash, "not readable") {
		t.Fatalf("unreadable z route: zoom=%v focus=%v err=%v flash=%q", m.zoom, m.focus, m.flashErr, m.flash)
	}
}

func TestThreadSpatialFocusHandlesExternalIsolatedDegradedAndNarrowCases(t *testing.T) {
	projection := threadFocusProjection()
	for _, tc := range []struct {
		focal string
		shown int
	}{
		{focal: "g", shown: 2},
		{focal: "f", shown: 1},
		{focal: "e", shown: 2},
	} {
		detail := newThreadDetail(projection, "", "", "")
		detail.view, detail.selection = threadDetailSpatial, tc.focal
		focused, err := detail.toggleDetailLocalFocus()
		if err != nil {
			t.Fatalf("focus %s: %v", tc.focal, err)
		}
		if got := focused.(threadDetail).focus.projection.Scope.ShownNodes; got != tc.shown {
			t.Errorf("focus %s shown=%d want %d", tc.focal, got, tc.shown)
		}
	}

	degraded := projection
	degraded.TopologyComplete = false
	degraded.View.GraphHealth = core.GraphBroken
	detail := newThreadDetail(degraded, "", "", "")
	detail.view, detail.selection = threadDetailSpatial, "b"
	focused, err := detail.toggleDetailLocalFocus()
	if err != nil {
		t.Fatal(err)
	}
	plain := ansi.Strip(focused.(threadDetail).renderDetail(120, 30, &testStyles))
	if !strings.Contains(plain, "partial") || !strings.Contains(plain, "graph broken") {
		t.Fatalf("focused graph hid degraded evidence:\n%s", plain)
	}
	narrow := ansi.Strip(focused.(threadDetail).renderDetail(40, 10, &testStyles))
	if !strings.Contains(narrow, "ZOOMED · ONE-HOP") || !strings.Contains(narrow, "give it room") ||
		!strings.Contains(narrow, "z full") || !strings.Contains(narrow, "active focal task") {
		t.Fatalf("narrow focus did not fail open:\n%s", narrow)
	}

	hostile := projection
	for index := range hostile.Nodes {
		if hostile.Nodes[index].TaskID == "b" {
			hostile.Nodes[index].Label = "hostile\n\x1b[31mfocus"
		}
	}
	detail = newThreadDetail(hostile, "", "", "")
	detail.view, detail.selection = threadDetailSpatial, "b"
	focused, err = detail.toggleDetailLocalFocus()
	if err != nil {
		t.Fatal(err)
	}
	hostilePlain := ansi.Strip(focused.(threadDetail).renderDetail(120, 30, &testStyles))
	if !strings.Contains(hostilePlain, "hostile↵�[31mfocus") {
		t.Fatalf("focused graph did not neutralize hostile label:\n%s", hostilePlain)
	}

	unreadable := newThreadDetail(projection, "", "", "")
	unreadable.view, unreadable.selection = threadDetailSpatial, "u"
	if unreadable.detailLocalFocusAvailable() {
		t.Fatal("unreadable node incorrectly offered as a focus target")
	}
}
