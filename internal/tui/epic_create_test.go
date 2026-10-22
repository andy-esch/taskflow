package tui

import (
	"bytes"
	"errors"
	"maps"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/charmbracelet/x/exp/teatest/v2"

	"github.com/andy-esch/taskflow/internal/core"
	"github.com/andy-esch/taskflow/internal/domain"
	"github.com/andy-esch/taskflow/internal/store"
	"github.com/andy-esch/taskflow/internal/testutil"
)

func openEpicForm(t *testing.T, m Model) Model {
	t.Helper()
	m = settleCreate(t, m, m.switchTab(indexOfKind(m.tabs, entityEpics)))
	m, cmd := createKey(m, "c")
	m = settleCreate(t, m, cmd)
	if !m.creation.epic.active || m.creation.task.active {
		t.Fatal("Epics c did not open only an epic form")
	}
	return m
}

func pasteCreation(m Model, value string) Model {
	next, _ := m.Update(tea.PasteMsg{Content: value})
	return next.(Model)
}

func fillEpic(m Model) Model {
	m = pasteCreation(m, "Organize / qhjkn 世界")
	m, _ = createKey(m, "enter")
	m = pasteCreation(m, "A domain worth organizing α")
	m, _ = createKey(m, "enter")
	m = pasteCreation(m, "planning, tui")
	return m
}

func TestEpicCreateDurableDefaultsAndOverrides(t *testing.T) {
	for _, override := range []bool{false, true} {
		t.Run(map[bool]string{false: "defaults", true: "overrides"}[override], func(t *testing.T) {
			root := seedRepo(t)
			m := openEpicForm(t, loadedAt(t, root, 120, 40))
			before := testutil.SnapshotTree(t, root)
			m = fillEpic(m)
			m, _ = createKey(m, "enter")           // tags -> status, not a write
			status, priority := "active", "medium" // independent expected values
			if override {
				m, _ = createKey(m, "enter") // status selector
				m, _ = createKey(m, "j")     // retired
				m, _ = createKey(m, "esc")   // unaccepted candidate is not staged
				if m.creation.epic.inputs[epicCreateStatus].Value() != "active" {
					t.Fatal("selector cancel staged its candidate")
				}
				m, _ = createKey(m, "enter")
				m, _ = createKey(m, "j")
				m, _ = createKey(m, "enter") // stage retired; advance to priority
				m, _ = createKey(m, "enter")
				m, _ = createKey(m, "j")
				m, _ = createKey(m, "enter") // stage low; advance to Cancel
				status, priority = "retired", "low"
			}
			m.creation.epic.controls().focus(epicCreateSubmit)
			if !maps.Equal(before, testutil.SnapshotTree(t, root)) {
				t.Fatal("staging created a partial record")
			}
			m, write := createKey(m, "enter")
			m, duplicate := createKey(m, "enter")
			if write == nil || duplicate != nil || !maps.Equal(before, testutil.SnapshotTree(t, root)) {
				t.Fatal("command construction wrote early or allowed duplicate submission")
			}
			m = settleCreate(t, m, write)
			ref := m.selectedRef()
			if m.creation.epic.active || ref.key != "02-organize-qhjkn-世界" {
				t.Fatalf("did not land on adapter-numbered epic: %+v", ref)
			}
			detail, err := m.svc.ShowEpic(ref.key)
			if err != nil {
				t.Fatal(err)
			}
			epic := detail.Summary.Epic
			if epic.ID != ref.key || epic.Status != status || epic.Priority != priority ||
				epic.Description != "A domain worth organizing α" || !slices.Equal(epic.Tags, []string{"planning", "tui"}) ||
				!strings.Contains(detail.Body, "# Organize / qhjkn 世界") {
				t.Fatalf("wrong durable epic/scaffold: %+v\n%s", epic, detail.Body)
			}
			after := testutil.SnapshotTree(t, root)
			delete(after, filepath.Join(domain.EpicsDir, ref.key+".md"))
			if !maps.Equal(before, after) {
				t.Fatal("creation changed existing records, modes, or directory structure")
			}
			m = settleCreate(t, m, m.reloadAll())
			if m.selectedRef() != ref {
				t.Fatal("watcher refresh stole created-epic focus")
			}
		})
	}
}

func TestEpicCreatePaletteAndPendingLanding(t *testing.T) {
	root := seedRepo(t)
	m := loadedAt(t, root, 120, 40)
	m = settleCreate(t, m, m.switchTab(indexOfKind(m.tabs, entityEpics)))
	m.setFocus(focusDetail)
	m, cmd := createKey(m, "ctrl+p")
	m = settleCreate(t, m, cmd)
	// Find by action identity, not the incidental ordering of contextual actions.
	found := false
	for i, raw := range m.palette.list.Items() {
		candidate := raw.(paletteItem)
		if candidate.kind == palAction && candidate.word == "new-epic" {
			m.palette.list.Select(i)
			found = true
			break
		}
	}
	if !found {
		t.Fatal("missing New epic palette action")
	}
	m, cmd = createKey(m, "enter")
	m = settleCreate(t, m, cmd)
	if !m.creation.epic.active || m.creation.task.active {
		t.Fatal("New epic action did not open the epic form")
	}
	m = fillEpic(m)
	m.creation.epic.controls().focus(epicCreateSubmit)
	m, cmd = createKey(m, "enter")
	// Widen a pre-existing status restriction and preserve the new target when a
	// watcher reload races the initial landing, not only a later stable refresh.
	m.tabs[indexOfKind(m.tabs, entityEpics)].statusView = "retired"
	next, landing := m.Update(cmd())
	m = next.(Model)
	m = settleCreate(t, m, m.reloadAll())
	m = settleCreate(t, m, landing)
	if m.selectedRef().key != "02-organize-qhjkn-世界" {
		t.Fatal("pending landing lost created epic to status restriction/watcher")
	}
	m.atlas.opening = true
	m, _ = createKey(m, "c")
	if m.creation.epic.active || !strings.Contains(m.flash, "pending workspace") {
		t.Fatal("epic form opened under a pending workspace activation")
	}
}

func TestEpicCreateReadOnlyAndPasteOwnership(t *testing.T) {
	root := seedRepo(t)
	m := loadedAt(t, root, 120, 40)
	m.svc = core.MustNewService(testutil.Must(store.NewFS(root, core.ReadOnlyMutations())))
	m = fillEpic(openEpicForm(t, m))
	before := testutil.SnapshotTree(t, root)
	m.creation.epic.controls().focus(epicCreateSubmit)
	m, cmd := createKey(m, "enter")
	m = settleCreate(t, m, cmd)
	if !m.creation.epic.active || m.creation.epic.locked || m.creation.epic.err == "" || !maps.Equal(before, testutil.SnapshotTree(t, root)) {
		t.Fatal("guarded read-only refusal lost draft or performed a write")
	}
	for _, mode := range []string{"help", "enum", "discard", "saving", "locked", "displaced"} {
		m := fillEpic(openEpicForm(t, loadedAt(t, root, 120, 40)))
		switch mode {
		case "help":
			m.showHelp = true
		case "enum":
			m.creation.epic.controls().focus(epicCreateStatus)
			m, _ = createKey(m, "enter")
		case "discard":
			m, _ = createKey(m, "esc")
		case "saving":
			m.creation.epic.controls().focus(epicCreateSubmit)
			m, _ = createKey(m, "enter")
		case "locked":
			m.creation.epic.locked = true
		case "displaced":
			m.creation.epic.displace()
		}
		before := m.creation.epic.recoveryText()
		m = pasteCreation(m, "invisible")
		if m.creation.epic.recoveryText() != before {
			t.Fatalf("paste escaped %s input ownership", mode)
		}
	}
}

func TestEpicCreateValidationCancelAndEmptyList(t *testing.T) {
	root := testutil.NewRepo(t).Root
	m := openEpicForm(t, loadedAt(t, root, 120, 40))
	if m.creation.epic.dirty() {
		t.Fatal("visible initial enum choices made an untouched draft dirty")
	}
	m, _ = createKey(m, "esc")
	if m.creation.epic.active {
		t.Fatal("clean draft did not cancel")
	}
	m = openEpicForm(t, m)
	m = pasteCreation(m, "Retain title")
	m, _ = createKey(m, "esc")
	m, _ = createKey(m, "enter") // keep editing is the default
	if !m.creation.epic.active || m.creation.epic.inputs[epicCreateTitle].Value() != "Retain title" {
		t.Fatal("dirty confirmation discarded title")
	}
	before := testutil.SnapshotTree(t, root)
	m.creation.epic.controls().focus(epicCreateSubmit)
	m, cmd := createKey(m, "enter")
	m = settleCreate(t, m, cmd)
	if !m.creation.epic.active || m.creation.epic.locked || m.creation.epic.err == "" || !maps.Equal(before, testutil.SnapshotTree(t, root)) {
		t.Fatal("required-description refusal lost draft, locked retry or changed files")
	}
	m.creation.epic.controls().focus(epicCreateDescription)
	m = pasteCreation(m, "A description")
	m.creation.epic.controls().focus(epicCreateSubmit)
	m, cmd = createKey(m, "enter")
	m = settleCreate(t, m, cmd)
	if m.creation.epic.active || m.selectedRef().key != "01-retain-title" {
		t.Fatal("explicit retry did not create first epic with optional empty tags")
	}
	// Complete the advertised no-CLI sequence: use this newly created epic as
	// the exact parent for a task through the existing contextual palette action.
	m, cmd = createKey(m, "ctrl+p")
	m = settleCreate(t, m, cmd)
	m, cmd = createKey(m, "enter")
	m = settleCreate(t, m, cmd)
	if !m.creation.task.active || m.creation.task.epic.key != "01-retain-title" {
		t.Fatal("new epic was not retained as the selected task parent")
	}
	m = pasteCreation(m, "First task")
	m, _ = createKey(m, "enter") // advance to the pre-filled Epic field
	m, _ = createKey(m, "tab")   // accept that existing selection, move to tags
	m = pasteCreation(m, "authoring")
	for range 4 {
		m, _ = createKey(m, "tab")
	}
	m, cmd = createKey(m, "enter")
	m = settleCreate(t, m, cmd)
	if task, ok := m.selectedTask(); !ok || task.Slug != "first-task" || task.Epic != "01-retain-title" {
		t.Fatal("epic-to-task creation did not finish with the selected parent")
	}
}

// This portable aggregate has no local-path interface. Typed creation is the
// only write; unexpected reference/local actions hit the nil embedded Store.
type epicCreatePort struct {
	core.Store
	setID     core.SourceSetID
	calls     int
	epic      domain.Epic
	body      string
	receipt   core.EpicCreationReceipt
	err       error
	committer core.Store
}

func (p *epicCreatePort) SourceSetID() core.SourceSetID { return p.setID }

func (p *epicCreatePort) CreateEpic(slug string, epic domain.Epic, body string, dry bool) (core.EpicCreationReceipt, error) {
	p.calls++
	p.epic, p.body = epic, body
	if p.committer != nil {
		receipt, err := p.committer.CreateEpic(slug, epic, body, dry)
		if err != nil {
			return receipt, err
		}
		return receipt, p.err
	}
	return p.receipt, p.err
}

func TestEpicCreatePortableAndNonReplayableReceipts(t *testing.T) {
	for _, outcome := range []string{"pathless", "read-only", "unknown", "unconfirmed", "finalization", "committed cleanup", "preview"} {
		t.Run(outcome, func(t *testing.T) {
			root := seedRepo(t)
			fs := testutil.Must(store.NewFS(root, core.UnrestrictedMutations()))
			port := &epicCreatePort{setID: fs.SourceSetID()}
			switch outcome {
			case "pathless":
				port.receipt = core.EpicCreationReceipt{Committed: true, Epic: domain.Epic{ID: "urn:epic:opaque"}}
			case "read-only":
				port.err = core.ErrReadOnlyPersistence
			case "unknown":
				port.err = errors.New("connection dropped")
			case "finalization":
				port.err = &core.CreateFinalizationError{Cause: domain.ErrConflict}
			case "committed cleanup":
				port.Store = fs // inspection/reload is real; the pathless fake stays narrow
				port.committer, port.err = fs, &core.CreateFinalizationError{Cause: errors.New("cleanup failed")}
			case "preview":
				port.receipt = core.EpicCreationReceipt{Committed: true, DryRun: true, Epic: domain.Epic{ID: "preview-only"}}
			}
			m := loadedAt(t, root, 120, 40)
			m = settleCreate(t, m, m.switchTab(indexOfKind(m.tabs, entityEpics)))
			m.svc = core.MustNewService(port)
			m, _ = createKey(m, "c")
			m = fillEpic(m)
			before := testutil.SnapshotTree(t, root)
			m.creation.epic.controls().focus(epicCreateSubmit)
			m, write := createKey(m, "enter")
			next, _ := m.Update(write()) // inspect delivery without unrelated fake reads
			m = next.(Model)
			if port.calls != 1 || port.epic.Status != "active" || port.epic.Priority != "medium" ||
				port.epic.Description != "A domain worth organizing α" || !strings.Contains(port.body, "# Organize / qhjkn 世界") {
				t.Fatal("per-kind request/scaffold did not reach the portable creation port")
			}
			if outcome == "pathless" {
				if m.creation.epic.active || m.tabs[indexOfKind(m.tabs, entityEpics)].restore.key != "urn:epic:opaque" || strings.Contains(m.flash, "editor") {
					t.Fatal("pathless receipt lost exact landing identity or promised a local editor")
				}
				return
			}
			if !m.creation.epic.active || m.creation.epic.locked != (outcome != "read-only") {
				t.Fatal("outcome did not preserve the correct retry/recovery state")
			}
			if outcome != "committed cleanup" && !maps.Equal(before, testutil.SnapshotTree(t, root)) {
				t.Fatal("portable refusal/recovery changed files")
			}
			if outcome != "read-only" {
				if !strings.Contains(m.creation.epic.recoveryText(), "Organize / qhjkn 世界") {
					t.Fatal("recovery lost typed input")
				}
				var follow tea.Cmd
				m, follow = createKey(m, "enter") // inspect/close only; never issue another create
				if outcome == "committed cleanup" {
					m = settleCreate(t, m, follow)
					if m.selectedRef().key != "02-organize-qhjkn-世界" {
						t.Fatal("inspection did not load the exact committed epic")
					}
				} else if follow != nil {
					// Execute a wrongly returned create command, rather than checking
					// a counter before the deferred I/O could possibly run.
					follow()
					t.Fatalf("non-inspectable recovery returned a command; create calls=%d", port.calls)
				}
				if port.calls != 1 {
					t.Fatal("recovery replayed Create")
				}
				if outcome == "preview" && m.tabs[indexOfKind(m.tabs, entityEpics)].restore.key == "preview-only" {
					t.Fatal("preview receipt was offered as a durable inspection target")
				}
			}
		})
	}
}

func TestEpicCreateOriginAndWidgetBinding(t *testing.T) {
	for _, issued := range []bool{false, true} {
		t.Run(map[bool]string{false: "draft", true: "issued"}[issued], func(t *testing.T) {
			alpha, beta := seedRepo(t), seedRepo(t)
			m := fillEpic(openEpicForm(t, loadedAt(t, alpha, 120, 40)))
			m.sessionScope = true
			var write tea.Cmd
			if issued {
				m.creation.epic.controls().focus(epicCreateSubmit)
				m, write = createKey(m, "enter")
			}
			before := testutil.SnapshotTree(t, beta)
			m.activateWorkspace(core.Workspace{Checkout: beta, PlanningRoot: beta,
				Planning: core.MustNewService(testutil.Must(store.NewFS(beta, core.UnrestrictedMutations())))}, nil, nil)
			if !m.creation.epic.displaced || !m.creation.epic.locked || m.creation.epic.inputs[epicCreateTitle].Value() != "Organize / qhjkn 世界" {
				t.Fatal("workspace activation discarded or rebound the epic draft")
			}
			if issued {
				next, cmd := m.Update(write())
				m = next.(Model)
				if cmd != nil || !m.creation.epic.receipt.Committed || m.creation.epic.saving {
					t.Fatal("departed result lost receipt or issued incoming-space work")
				}
			}
			m, cmd := createKey(m, "enter")
			if cmd != nil || !maps.Equal(before, testutil.SnapshotTree(t, beta)) {
				t.Fatal("displaced intent mutated or navigated incoming workspace")
			}
		})
	}
	m := openEpicForm(t, loaded(t, 120, 40))
	late := m.creation.epic.widgetCmd(func() tea.Msg { return tea.PasteMsg{Content: "late"} })
	m, _ = createKey(m, "tab")
	next, _ := m.Update(late())
	m = next.(Model)
	if m.creation.epic.inputs[epicCreateDescription].Value() != "" {
		t.Fatal("departed field received delayed input")
	}
	// Deliberately collide form/input generations: the entity kind must still
	// prevent a task widget result from editing an epic field.
	next, _ = m.Update(creationWidgetMsg{kind: entityTasks, gen: m.creation.epic.gen, inputGen: m.creation.epic.inputGen,
		msg: tea.PasteMsg{Content: "wrong kind"}})
	if next.(Model).creation.epic.inputs[epicCreateDescription].Value() != "" {
		t.Fatal("widget delivery crossed entity kinds")
	}
}

func TestEpicCreateRuntimePaste(t *testing.T) {
	tm := teatest.NewTestModel(t, newModel(t), teatest.WithInitialTermSize(120, 40))
	wait := func(texts ...string) {
		t.Helper()
		teatest.WaitFor(t, tm.Output(), func(output []byte) bool {
			for _, text := range texts {
				if !bytes.Contains(output, []byte(text)) {
					return false
				}
			}
			return true
		}, teatest.WithDuration(5*time.Second))
	}
	tm.Send(press("]"))
	wait("beta")
	tm.Send(press("]"))
	wait("01-test")
	tm.Send(press("c"))
	wait("New epic")
	tm.Send(tea.PasteMsg{Content: "Runtime domain"})
	tm.Send(press("enter"))
	tm.Send(tea.PasteMsg{Content: "Useful description 世界"})
	for range 5 { // description -> tags -> status -> priority -> Cancel -> Create
		tm.Send(press("tab"))
	}
	tm.Send(press("enter"))
	wait("created 02-runtime-domain", "# Runtime domain")
	tm.Send(press("q"))
	final := tm.FinalModel(t, teatest.WithFinalTimeout(5*time.Second)).(Model)
	if final.creation.epic.active || final.selectedRef().key != "02-runtime-domain" {
		t.Fatal("running program failed to create and focus epic")
	}
}

func TestEpicCreateNarrowModes(t *testing.T) {
	for _, mode := range []string{"fields", "enum", "discard", "pending", "recovery"} {
		t.Run(mode, func(t *testing.T) {
			m := fillEpic(openEpicForm(t, loaded(t, 120, 40)))
			switch mode {
			case "enum":
				m.creation.epic.controls().focus(epicCreateStatus)
				m, _ = createKey(m, "enter")
			case "discard":
				m, _ = createKey(m, "esc")
			case "pending":
				m.creation.epic.controls().focus(epicCreateSubmit)
				m, _ = createKey(m, "enter")
			case "recovery":
				m.creation.epic.displace()
			}
			for _, size := range [][2]int{{120, 40}, {48, 20}, {30, 10}} {
				next, _ := m.Update(tea.WindowSizeMsg{Width: size[0], Height: size[1]})
				m = next.(Model)
				view := m.View().Content
				if lipgloss.Width(view) > size[0] || lipgloss.Height(view) > size[1] {
					t.Fatalf("%s overflows %v", mode, size)
				}
				if size[0] >= 48 && mode == "fields" && !strings.Contains(ansi.Strip(view), "[Create]") {
					t.Fatal("explicit Create disappeared at supported size")
				}
			}
			before := m.creation.epic.recoveryText()
			next, _ := m.Update(tea.PasteMsg{Content: "invisible"})
			if next.(Model).creation.epic.recoveryText() != before {
				t.Fatal("tiny terminal accepted invisible input")
			}
		})
	}
}

func TestEpicCreateVerboseRecoveryKeepsIdentityAndActionsVisible(t *testing.T) {
	root := seedRepo(t)
	fs := testutil.Must(store.NewFS(root, core.UnrestrictedMutations()))
	diagnostic := strings.Repeat("cleanup diagnostic 世界 ", 90) + "final evidence"
	port := &epicCreatePort{Store: fs, setID: fs.SourceSetID(), committer: fs,
		err: &core.CreateFinalizationError{Cause: errors.New(diagnostic)}}
	m := loadedAt(t, root, 120, 40)
	m = settleCreate(t, m, m.switchTab(indexOfKind(m.tabs, entityEpics)))
	m.svc = core.MustNewService(port)
	m, _ = createKey(m, "c")
	m = fillEpic(m)
	m.creation.epic.controls().focus(epicCreateSubmit)
	m, write := createKey(m, "enter")
	m = settleCreate(t, m, write)
	if !m.creation.epic.locked || !m.creation.epic.receipt.Committed {
		t.Fatal("fixture did not reach committed cleanup recovery")
	}
	for _, size := range [][2]int{{120, 40}, {48, 20}, {46, 18}} {
		next, _ := m.Update(tea.WindowSizeMsg{Width: size[0], Height: size[1]})
		m = next.(Model)
		view := ansi.Strip(m.View().Content)
		for _, evidence := range []string{"Save to:", "Record: 02-organize-qhjkn-世界", "Draft:",
			"Create disabled", "inspect created record", "y copy", "esc close"} {
			if !strings.Contains(view, evidence) {
				t.Fatalf("%v: recovery hides %q:\n%s", size, evidence, view)
			}
		}
		if lipgloss.Width(m.View().Content) > size[0] || lipgloss.Height(m.View().Content) > size[1] {
			t.Fatalf("recovery overflow at %v", size)
		}
		if !strings.Contains(m.creation.epic.recoveryText(), diagnostic) {
			t.Fatal("bounded display discarded full copyable diagnostics")
		}
	}
}
