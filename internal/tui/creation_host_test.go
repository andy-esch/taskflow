package tui

import (
	"fmt"
	"maps"
	"slices"
	"strings"
	"testing"

	"charm.land/bubbles/v2/cursor"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/andy-esch/taskflow/internal/testutil"
)

// This inventory covers mechanical contracts, not domain defaults or receipts.
// Per-kind durable expectations remain in the concrete creation tests.
var creationContractKinds = map[string]entityKind{
	"task": entityTasks, "epic": entityEpics, "audit": entityAudits, "research": entityResearch,
	"thread": entityThreads,
}

func openCreationKind(t *testing.T, m Model, kind entityKind) Model {
	t.Helper()
	m = settleCreate(t, m, m.switchTab(indexOfKind(m.tabs, kind)))
	next, cmd := createKey(m, "c")
	m = settleCreate(t, next, cmd)
	if f := m.activeCreation(); f == nil || f.kind != kind {
		t.Fatal("page-local action did not open its kind")
	}
	return m
}

func creationTitle(m Model) string {
	return creationContractInputs(&m)[0].Value()
}

func TestEmptyCreationHintsOpenNativeForms(t *testing.T) {
	for _, tc := range []struct {
		name string
		kind entityKind
		hint string
	}{
		{"tasks", entityTasks, "Press c to create a task."},
		{"threads", entityThreads, "Press c to create an empty, unstarted Thread."},
	} {
		t.Run(tc.name, func(t *testing.T) {
			m := loadedAt(t, testutil.NewRepo(t).Root, 120, 40)
			m = settleCreate(t, m, m.switchTab(indexOfKind(m.tabs, tc.kind)))
			if hint := m.listPaneContent(); !strings.Contains(hint, tc.hint) || strings.Contains(hint, "tskflwctl") {
				t.Fatalf("empty page does not guide native creation: %q", hint)
			}
			m, cmd := createKey(m, "c")
			m = settleCreate(t, m, cmd)
			if f := m.activeCreation(); f == nil || f.kind != tc.kind {
				t.Fatal("advertised empty-page action did not open its creation form")
			}
		})
	}
}

func TestCreationPageHintsFollowActions(t *testing.T) {
	for name, kind := range creationContractKinds {
		t.Run(name, func(t *testing.T) {
			m := loaded(t, 160, 40)
			m = settleCreate(t, m, m.switchTab(indexOfKind(m.tabs, kind)))
			action, ok := m.pageCreationAction()
			if !ok || action.kind != palAction || action.word != "new-"+name || !action.ref.empty() || !slices.Contains(m.creationActions(), action) {
				t.Fatal("page creation metadata disagrees with its palette action or inherits selected context")
			}
			if !strings.Contains(ansi.Strip(m.footer()), "c new "+name) {
				t.Fatal("page-local creation is missing from the footer")
			}
			m.onDash = true
			if m.pageCreationHint() != "" {
				t.Fatal("dashboard advertises page creation")
			}
			m.onDash, m.onAtlas = false, true
			if _, ok := m.pageCreationAction(); ok || m.pageCreationHint() != "" || len(m.creationActions()) != 0 {
				t.Fatal("atlas advertises page creation")
			}
		})
	}
}

// Test-only mechanics inventory; expectations for parameters, defaults and
// durable records must not derive from these widget arrays or production helpers.
func creationContractInputs(m *Model) []textinput.Model {
	switch m.creation.kind {
	case entityTasks:
		return m.creation.task.inputs[:]
	case entityEpics:
		return m.creation.epic.inputs[:]
	case entityAudits:
		return m.creation.audit.inputs[:]
	case entityResearch:
		return m.creation.research.inputs[:]
	case entityThreads:
		return m.creation.thread.inputs[:]
	}
	panic("missing creation contract input inventory")
}

func creationContractControls(m *Model) creationControls {
	switch m.creation.kind {
	case entityTasks:
		return m.creation.task.controls()
	case entityEpics:
		return m.creation.epic.controls()
	case entityAudits:
		return m.creation.audit.controls()
	case entityResearch:
		return m.creation.research.controls()
	case entityThreads:
		return m.creation.thread.controls()
	}
	panic("missing creation contract controls inventory")
}

func TestCreationHostLifetimeContract(t *testing.T) {
	for name, kind := range creationContractKinds {
		t.Run(name, func(t *testing.T) {
			root := seedRepo(t)
			m := openCreationKind(t, loadedAt(t, root, 120, 40), kind)
			before := testutil.SnapshotTree(t, root)
			m = pasteCreation(m, "Keep my intent")
			prior := m
			gen := m.activeCreation().gen
			// Host entry points refuse replacement independently of shell guards.
			if m.creation.openTask(m.svc, m.sessionGen, "wrong", entityRef{}) != nil ||
				m.creation.openEpic(m.svc, m.sessionGen, "wrong") != nil ||
				m.creation.openAudit(m.svc, m.sessionGen, "wrong") != nil ||
				m.creation.openResearch(m.svc, m.sessionGen, "wrong") != nil ||
				m.creation.openThread(m.svc, m.sessionGen, "wrong") != nil || m.creation.gen != gen {
				t.Fatal("active intent was replaced or consumed a generation")
			}
			m, _ = createKey(m, "esc")
			if m.activeCreation().mode != creationDiscardMode || prior.activeCreation().mode != creationFieldsMode {
				t.Fatal("discard state leaked to the prior Model or skipped confirmation")
			}
			m, _ = createKey(m, "enter")
			if creationTitle(m) != "Keep my intent" || m.activeCreation().mode != creationFieldsMode {
				t.Fatal("default discard action did not keep input")
			}
			m, _ = createKey(m, "esc")
			m, _ = createKey(m, "y")
			if m.activeCreation() != nil || !prior.activeCreation().active {
				t.Fatal("close failed or changed the retained prior Model")
			}
			if m.creation.begin(entityAtlas) || m.creation.gen != gen {
				t.Fatal("unsupported kind consumed a lifetime or opened a form")
			}
			nextKind := entityTasks
			if kind == entityTasks {
				nextKind = entityEpics
			}
			m = openCreationKind(t, m, nextKind)
			if m.activeCreation().gen <= gen || creationTitle(m) != "" {
				t.Fatal("cross-kind reopen reused an attempt or inherited text")
			}
			m.creation.displace()
			if !m.activeCreation().displaced || !m.activeCreation().locked || m.creation.begin(kind) {
				t.Fatal("displaced recovery lost ownership or could be replaced")
			}
			if !maps.Equal(before, testutil.SnapshotTree(t, root)) {
				t.Fatal("mechanical lifetime transitions wrote planning data")
			}
		})
	}
}

// A routing-only probe: receipt policy and real side effects are independently
// pinned by the existing per-kind origin/no-replay and durable adapter tests.
type creationRouteProbe struct {
	kind         entityKind
	gen, session uint64
	calls        *int
}

func (r creationRouteProbe) creationRoute() (entityKind, uint64, uint64) {
	return r.kind, r.gen, r.session
}
func (r creationRouteProbe) applyCreation(*Model) tea.Cmd {
	(*r.calls)++
	return nil
}

func TestCreationHostResultGuards(t *testing.T) {
	for name, kind := range creationContractKinds {
		for _, state := range []string{"matching", "closed", "not pending", "other kind", "stale attempt", "other origin"} {
			t.Run(name+"/"+state, func(t *testing.T) {
				m := openCreationKind(t, loaded(t, 120, 40), kind)
				f := m.activeCreation()
				f.saving = true
				calls := 0
				result := creationRouteProbe{kind: kind, gen: f.gen, session: f.session, calls: &calls}
				switch state {
				case "closed":
					m.creationDraft().close()
				case "not pending":
					f.saving = false
				case "other kind":
					result.kind = entityTasks
					if kind == entityTasks {
						result.kind = entityEpics
					}
				case "stale attempt":
					result.gen++
				case "other origin":
					result.session++
				}
				_, cmd := m.Update(result)
				want := 0
				if state == "matching" {
					want = 1
				}
				if calls != want || cmd != nil {
					t.Fatal("creation routing accepted the wrong lifetime or duplicated delivery")
				}
			})
		}
	}
}

func TestCreationHarnessDeliversFiniteWidgetResults(t *testing.T) {
	for name, kind := range creationContractKinds {
		t.Run(name, func(t *testing.T) {
			m := openCreationKind(t, loaded(t, 120, 40), kind)
			m.sessionScope = true
			finite := scopeSession(m.sessionGen, m.activeCreation().widgetCmd(func() tea.Msg {
				return tea.PasteMsg{Content: "finite delivery"}
			}))
			m = settleCreate(t, m, tea.Batch(nil, finite))
			if creationTitle(m) != "finite delivery" {
				t.Fatal("settling discarded finite widget input")
			}
		})
	}
}

func TestCreationHarnessStopsMaintenanceAndBoundsWork(t *testing.T) {
	m := openCreationKind(t, loaded(t, 120, 40), entityEpics)
	maintenance := m.activeCreation().widgetCmd(func() tea.Msg { return cursor.BlinkMsg{} })
	if _, err := settleCreationCommands(m, maintenance, 1); err != nil {
		t.Fatal(err)
	}
	var cycle tea.Cmd
	cycle = func() tea.Msg { return tea.BatchMsg{cycle} }
	if _, err := settleCreationCommands(m, cycle, 3); err == nil || !strings.Contains(err.Error(), "exceeded budget 3") {
		t.Fatal("unbounded work did not produce a useful budget failure")
	}
}

func TestCreationVisibilityMatchesLayoutBoundary(t *testing.T) {
	for name, kind := range creationContractKinds {
		for _, size := range []struct {
			width, height int
			visible       bool
		}{{45, 18, false}, {46, 17, false}, {46, 18, true}, {47, 19, true}} {
			t.Run(fmt.Sprintf("%s/%dx%d", name, size.width, size.height), func(t *testing.T) {
				m := openCreationKind(t, loaded(t, 120, 40), kind)
				next, _ := m.Update(tea.WindowSizeMsg{Width: size.width, Height: size.height})
				m = next.(Model)
				view := ansi.Strip(m.View().Content)
				if (creationModal{}).inputVisible(&m) != size.visible || strings.Contains(view, "Resize to continue") == size.visible {
					t.Fatal("rendered resize state and input authority disagree")
				}
				m = pasteCreation(m, "X")
				want := ""
				if size.visible {
					want = "X"
					if !strings.Contains(view, "[Create]") {
						t.Fatal("supported boundary hid the explicit submit action")
					}
				}
				if creationTitle(m) != want {
					t.Fatal("input did not follow the rendered visibility contract")
				}
			})
		}
	}
}
