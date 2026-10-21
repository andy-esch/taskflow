package tui

import (
	"bytes"
	"errors"
	"maps"
	"os"
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

func createKey(m Model, key string) (Model, tea.Cmd) {
	msg := press(key)
	switch key {
	case "up":
		msg = tea.KeyPressMsg{Code: tea.KeyUp}
	case "down":
		msg = tea.KeyPressMsg{Code: tea.KeyDown}
	case "shift+tab":
		msg = tea.KeyPressMsg{Code: tea.KeyTab, Mod: tea.ModShift}
	case "ctrl+p":
		msg = tea.KeyPressMsg{Code: 'p', Mod: tea.ModCtrl}
	case "ctrl+r":
		msg = tea.KeyPressMsg{Code: 'r', Mod: tea.ModCtrl}
	}
	next, cmd := m.Update(msg)
	return next.(Model), cmd
}

// Settle application messages, not perpetual cursor-blink timers. All actual
// reads/writes and their following reloads still execute through real tea.Cmds.
func settleCreate(t *testing.T, m Model, cmd tea.Cmd) Model {
	t.Helper()
	if cmd == nil {
		return m
	}
	msg := cmd()
	if batch, ok := msg.(tea.BatchMsg); ok {
		for _, child := range batch {
			m = settleCreate(t, m, child)
		}
		return m
	}
	inner := msg
	if scoped, ok := inner.(sessionMsg); ok {
		inner = scoped.msg
	}
	if _, ok := inner.(taskCreateWidgetMsg); ok {
		return m
	}
	next, follow := m.Update(msg)
	return settleCreate(t, next.(Model), follow)
}

func openCreate(t *testing.T, m Model) Model {
	t.Helper()
	m, cmd := createKey(m, "c")
	m = settleCreate(t, m, cmd)
	if !m.createTask.active || !m.createTask.catalogLoaded {
		t.Fatal("task form/catalog did not open")
	}
	return m
}

func fillCreate(m *Model) {
	m.createTask.inputs[taskCreateTitle].SetValue("Ship a task")
	m.createTask.inputs[taskCreateTags].SetValue("tui, authoring")
	m.createTask.inputs[taskCreateDescription].SetValue("A useful task")
	// Model a user choosing the actual adapter-established picker record.
	ep := m.createTask.epics[0]
	m.createTask.epic = entityRef{key: ep.Source.ID, label: ep.Epic.ID}
}

func chooseCreateOption(t *testing.T, m Model, row int, value string) Model {
	t.Helper()
	m.createTask.focus(row)
	m, _ = createKey(m, "enter")
	options := taskCreateEnumOptions(row)
	want := slices.Index(options, value)
	if want < 0 || m.createTask.mode != taskCreateEnumPicker {
		t.Fatalf("no selector option %q on row %d", value, row)
	}
	for range len(options) {
		if m.createTask.enumCursor == want {
			break
		}
		m, _ = createKey(m, "j")
	}
	m, _ = createKey(m, "enter")
	return m
}

func TestTaskCreateFromTasksAndEpicContext(t *testing.T) {
	for _, context := range []string{"tasks", "epic"} {
		t.Run(context, func(t *testing.T) {
			root := seedRepo(t)
			m := loadedAt(t, root, 120, 40)
			if context == "epic" {
				m = settleCreate(t, m, m.switchTab(indexOfKind(m.tabs, entityEpics)))
				m.setFocus(focusDetail)
				m, _ = createKey(m, "c")
				if m.createTask.active {
					t.Fatal("epic-page c must not create a task")
				}
				var cmd tea.Cmd
				m, cmd = createKey(m, "ctrl+p")
				// Named actions lead an empty palette; choose it through real routing.
				m = settleCreate(t, m, cmd)
				item, ok := m.palette.list.SelectedItem().(paletteItem)
				if !ok || item.kind != palAction || item.word != "new-task-in-epic" {
					t.Fatalf("missing contextual action: %+v", item)
				}
				m, cmd = createKey(m, "enter")
				m = settleCreate(t, m, cmd)
				if m.createTask.epic.key != "01-test" {
					t.Fatal("epic was not preselected")
				}
			} else {
				m = openCreate(t, m)
			}
			assertCreateFromForm(t, m, root)
		})
	}
}

func assertCreateFromForm(t *testing.T, m Model, root string) {
	t.Helper()
	before := testutil.SnapshotTree(t, root)
	fillCreate(&m)
	m.createTask.more = true
	m = chooseCreateOption(t, m, taskCreatePriority, "high")
	m = chooseCreateOption(t, m, taskCreateTier, "2")
	m = chooseCreateOption(t, m, taskCreateAutonomy, "4")
	m.createTask.inputs[taskCreateEffort].SetValue("one day")
	// Enter in title advances, never saves; draft manipulation is memory-only.
	m.createTask.focus(taskCreateTitle)
	m, _ = createKey(m, "enter")
	if m.createTask.saving || !maps.Equal(before, testutil.SnapshotTree(t, root)) {
		t.Fatal("editing/advancing wrote a partial record")
	}
	m.createTask.focus(taskCreateSubmit)
	m, write := createKey(m, "enter")
	if !m.createTask.saving || !maps.Equal(before, testutil.SnapshotTree(t, root)) {
		t.Fatal("command construction performed I/O or did not mark submission pending")
	}
	var duplicate tea.Cmd
	m, duplicate = createKey(m, "enter")
	if duplicate != nil {
		t.Fatal("duplicate Create issued a second write")
	}
	m = settleCreate(t, m, write)
	if m.createTask.active || m.cur().kind != entityTasks {
		t.Fatal("success did not return to Tasks")
	}
	selected, ok := m.selectedTask()
	if !ok || selected.Slug != "ship-a-task" || selected.Epic != "01-test" || selected.Status != domain.StatusReadyToStart ||
		selected.Priority != "high" || selected.Tier != 2 || selected.Autonomy != 4 || selected.Effort != "one day" {
		t.Fatalf("wrong created task/focus: %+v", selected)
	}
	record, err := m.svc.ShowTask(selected.ID)
	if err != nil || !strings.Contains(record.Value.Body, "# Ship a task") {
		t.Fatalf("default scaffold lost: %+v %v", record, err)
	}
	ref := m.selectedRef()
	// A watcher-triggered reload racing the creation landing retains its key.
	m = settleCreate(t, m, m.reloadAll())
	if m.selectedRef() != ref {
		t.Fatal("watcher refresh stole created-task focus")
	}
	after := testutil.SnapshotTree(t, root)
	delete(after, filepath.Join(domain.TasksDir, selected.ID+"-ship-a-task.md"))
	if !maps.Equal(before, after) {
		t.Fatal("creation changed another planning record")
	}
}

func TestTaskCreateKeyboardAndDirtyCancel(t *testing.T) {
	root := seedRepo(t)
	m := openCreate(t, loadedAt(t, root, 100, 34))
	before := testutil.SnapshotTree(t, root)
	for _, key := range []string{"q", "h", "j", "k", "n"} {
		m, _ = createKey(m, key)
	}
	if m.createTask.inputs[taskCreateTitle].Value() != "qhjkn" || m.createTask.cursor != taskCreateTitle {
		t.Fatal("text keys escaped into global navigation")
	}
	m, _ = createKey(m, "esc")
	if m.createTask.mode != taskCreateDiscard {
		t.Fatal("dirty cancel did not confirm")
	}
	m, _ = createKey(m, "enter")
	if !m.createTask.active || m.createTask.mode == taskCreateDiscard || m.createTask.inputs[taskCreateTitle].Value() != "qhjkn" {
		t.Fatal("Enter did not default to keeping input")
	}
	m, _ = createKey(m, "esc")
	m, _ = createKey(m, "y")
	if m.createTask.active || !maps.Equal(before, testutil.SnapshotTree(t, root)) {
		t.Fatal("discard persisted input")
	}
	m = openCreate(t, m)
	m, _ = createKey(m, "esc")
	if m.createTask.active {
		t.Fatal("clean form should close without confirmation")
	}
}

func TestTaskCreateValidationPreservesDraftAndRetriesExplicitly(t *testing.T) {
	m := openCreate(t, loaded(t, 120, 40))
	fillCreate(&m)
	m.createTask.inputs[taskCreateTags].SetValue("")
	m.createTask.focus(taskCreateSubmit)
	m, write := createKey(m, "enter")
	m = settleCreate(t, m, write)
	if !m.createTask.active || m.createTask.locked || m.createTask.err == "" || m.createTask.inputs[taskCreateTitle].Value() != "Ship a task" {
		t.Fatal("validation refusal lost or locked the draft")
	}
	m.createTask.inputs[taskCreateTags].SetValue("tui")
	m.createTask.inputs[taskCreateTier].SetValue("not a number")
	m, _ = createKey(m, "enter")
	if m.createTask.saving || m.createTask.cursor != taskCreateTier || !m.createTask.more {
		t.Fatal("unparseable number was not attributed before dispatch")
	}
	m.createTask.inputs[taskCreateTier].SetValue("")
	m.createTask.focus(taskCreateSubmit)
	m, write = createKey(m, "enter")
	m = settleCreate(t, m, write)
	task, ok := m.selectedTask()
	if !ok || m.createTask.active || task.Priority != "medium" || task.Tier != 3 || task.Autonomy != 3 || task.Effort != "Unknown" {
		t.Fatalf("core defaults were not retained: %+v", task)
	}
}

func TestTaskCreatePickerRetainsExactSelectionOnParentDrift(t *testing.T) {
	root := seedRepo(t)
	m := openCreate(t, loadedAt(t, root, 120, 40))
	fillCreate(&m)
	m.createTask.focus(taskCreateEpic)
	m, _ = createKey(m, "enter")
	for _, char := range "test" {
		m, _ = createKey(m, string(char))
	}
	m, _ = createKey(m, "enter")
	if m.createTask.mode == taskCreateEpicPicker || m.createTask.epic.key != "01-test" {
		t.Fatal("picker did not retain source identity")
	}
	m.createTask.focus(taskCreateSubmit)
	m, write := createKey(m, "enter")
	if err := os.Rename(filepath.Join(root, "epics", "01-test.md"), filepath.Join(root, "epics", "01-renamed.md")); err != nil {
		t.Fatal(err)
	}
	before := testutil.SnapshotTree(t, root)
	m = settleCreate(t, m, write)
	if !m.createTask.active || m.createTask.epic.key != "01-test" || m.createTask.inputs[taskCreateTitle].Value() != "Ship a task" || m.createTask.err == "" ||
		!maps.Equal(before, testutil.SnapshotTree(t, root)) {
		t.Fatal("stale selection fell back to renamed parent or lost input")
	}
}

type createReceiptPort struct {
	setID     core.SourceSetID
	calls     int
	target    core.EpicMutationTarget
	task      domain.Task
	body      string
	receipt   core.TaskCreationReceipt
	err       error
	committer core.SelectedEpicTaskCreationStore
}

func (p *createReceiptPort) SourceSetID() core.SourceSetID { return p.setID }
func (p *createReceiptPort) CreateTaskForSelectedEpic(target core.EpicMutationTarget, task domain.Task, body string, dry bool) (core.TaskCreationReceipt, error) {
	p.calls++
	p.target = target
	p.task, p.body = task, body
	if p.committer != nil {
		receipt, err := p.committer.CreateTaskForSelectedEpic(target, task, body, dry)
		if err != nil {
			return receipt, err
		}
		return receipt, p.err
	}
	return p.receipt, p.err
}

// The aggregate only supplies portable semantic reads. Unexpected ordinary
// creation/reference/path calls hit the nil embedded port instead of a fallback.
type createPortableReads struct {
	core.Store
	setID core.SourceSetID
}

func (r *createPortableReads) SourceSetID() core.SourceSetID { return r.setID }
func (*createPortableReads) ReadEpics() (core.EpicRead, error) {
	return core.EpicRead{Records: []core.LoadedRecord[domain.Epic]{{
		Source: core.RecordSource{ID: "urn:epic:selected", Location: "/not-a-local-capability"},
		Value:  domain.Epic{ID: "Friendly display epic", Status: domain.EpicStatusActive},
	}}}, nil
}
func (*createPortableReads) ReadTasks() (core.TaskRead, error) { return core.TaskRead{}, nil }

func TestTaskCreateUsesPortableCapabilityAndOpaquePickerIdentity(t *testing.T) {
	for _, available := range []bool{false, true} {
		t.Run(map[bool]string{false: "absent capability", true: "pathless capability"}[available], func(t *testing.T) {
			reads := &createPortableReads{setID: core.NewSourceSetID()}
			port := &createReceiptPort{setID: reads.setID, receipt: core.TaskCreationReceipt{
				Committed: true, Task: domain.Task{ID: "6g4ym5y2dkp8", Slug: "ship-a-task"},
			}}
			var options []core.Option
			if available {
				options = append(options, core.WithSelectedEpicTaskCreationStore(port))
			}
			svc := core.MustNewService(reads, options...)
			if svc.HasLocalPath(core.EntityTask) || svc.HasLocalPath(core.EntityEpic) {
				t.Fatal("portable creation unexpectedly requires local paths")
			}
			m := New(svc)
			m.onDash = false
			next, _ := m.Update(tea.WindowSizeMsg{Width: 100, Height: 32})
			m = openCreate(t, next.(Model)) // creation also works on an empty task list
			fillCreate(&m)
			m.createTask.focus(taskCreateSubmit)
			m, write := createKey(m, "enter")
			// Execute the real command, then inspect delivery without running the
			// subsequent list/detail loads (not part of this narrow fake).
			result := write()
			next, _ = m.Update(result)
			m = next.(Model)
			if available {
				if port.calls != 1 || port.target.ID != "urn:epic:selected" || port.task.Epic != "urn:epic:selected" ||
					port.task.Slug != "ship-a-task" || port.task.Status != domain.StatusReadyToStart || !strings.Contains(port.body, "# Ship a task") || m.createTask.active {
					t.Fatalf("portable submit fell back/lost identity: port=%+v form=%+v", port, m.createTask)
				}
				if strings.Contains(m.flash, "editor") {
					t.Fatal("pathless creation promised an unavailable local editor")
				}
			} else if !m.createTask.active || m.createTask.locked || !strings.Contains(m.createTask.err, "unavailable") {
				t.Fatal("missing capability was not an explicit input-preserving refusal")
			}
		})
	}
}

func TestTaskCreateCatalogQuarantinesInvalidIdentityAndRejectsStaleLoads(t *testing.T) {
	m := openCreate(t, loaded(t, 120, 40))
	fillCreate(&m)
	f := &m.createTask
	gen, catalog := f.gen, f.catalogGen
	selected := f.epic
	invalid := core.EpicSummary{Source: core.RecordSource{ID: "same"}, Epic: domain.Epic{ID: "Visible"}}
	m.applyTaskCreateEpics(taskCreateEpicsMsg{gen: gen, catalog: catalog, epics: []core.EpicSummary{invalid, invalid}})
	if m.createTask.catalogErr == "" || len(m.createTask.matches()) != 0 || m.createTask.epic != selected {
		t.Fatal("duplicate identity catalog remained selectable or changed selected intent")
	}
	first, second := m.createTask.loadEpics(), m.createTask.loadEpics()
	newer := second().(taskCreateEpicsMsg)
	m.applyTaskCreateEpics(newer)
	m.applyTaskCreateEpics(taskCreateEpicsMsg{gen: gen, catalog: catalog, err: errors.New("stale catalog failure")})
	if m.createTask.catalogErr != "" || !m.createTask.catalogLoaded || len(m.createTask.epics) == 0 {
		t.Fatal("old catalog result replaced a newer coherent load")
	}
	m.createTask.close()
	m = openCreate(t, m)
	m.applyTaskCreateEpics(first().(taskCreateEpicsMsg))
	if m.createTask.gen == gen || m.createTask.catalogErr != "" || m.createTask.inputs[taskCreateTitle].Value() != "" {
		t.Fatal("departed form's load affected a reopened form")
	}
}

func TestTaskCreateWatcherCannotSupersedeCreatedLanding(t *testing.T) {
	m := openCreate(t, loaded(t, 120, 40))
	fillCreate(&m)
	// Creating from a status/filter that excludes a fresh task still lands on it.
	m.cur().statusView = string(domain.StatusInProgress)
	m.cur().list.SetFilterText("alpha")
	m.createTask.focus(taskCreateSubmit)
	m, write := createKey(m, "enter")
	result := write()
	next, landing := m.Update(result)
	m = next.(Model)
	// Do not deliver the landing load yet. A watcher starts a newer generation
	// while the old cursor still points to a different task.
	watch := m.reloadAll()
	m = settleCreate(t, m, watch)
	m = settleCreate(t, m, landing)
	if task, ok := m.selectedTask(); !ok || task.Slug != "ship-a-task" {
		t.Fatal("concurrent watcher load stole the pending created-task selection")
	}
}

func TestTaskCreateInRealProgram(t *testing.T) {
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
	tm.Send(press("c"))
	wait("Nothing saved until Create")
	for _, char := range "Real creation" {
		tm.Send(tea.KeyPressMsg{Code: char, Text: string(char)})
	}
	tm.Send(press("enter")) // title advances to epic, not submit
	tm.Send(press("enter")) // open the epic picker
	wait("Search epics:")
	tm.Send(press("enter")) // choose the existing epic, advance to tags
	for _, char := range "tui, authoring" {
		tm.Send(tea.KeyPressMsg{Code: char, Text: string(char)})
	}
	for range 4 {
		tm.Send(press("tab")) // tags → description → more → cancel → Create
	}
	tm.Send(press("enter"))
	// The flash precedes the async list/detail landing. The scaffold heading
	// proves the created task's detail has loaded, not just its success message.
	wait("created real-creation", "# Real creation")
	tm.Send(press("q"))
	final := tm.FinalModel(t, teatest.WithFinalTimeout(5*time.Second)).(Model)
	if task, ok := final.selectedTask(); !ok || task.Slug != "real-creation" || task.Epic != "01-test" || final.createTask.active {
		t.Fatal("runtime command/widget routing did not complete staged creation")
	}
}

func TestTaskCreateTerminalPasteInRealProgram(t *testing.T) {
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
	tm.Send(press("c"))
	wait("Nothing saved until Create")
	tm.Send(tea.PasteMsg{Content: "Pasted task"})
	tm.Send(press("enter"))
	tm.Send(press("enter"))
	wait("Search epics:")
	tm.Send(tea.PasteMsg{Content: "01-test"})
	tm.Send(press("enter"))
	tm.Send(tea.PasteMsg{Content: "pasted, authoring"})
	tm.Send(press("tab"))
	tm.Send(tea.PasteMsg{Content: "Description 世界"})
	tm.Send(press("tab"))
	tm.Send(press("enter")) // expand More options
	tm.Send(press("tab"))   // priority
	tm.Send(press("tab"))   // effort
	tm.Send(tea.PasteMsg{Content: "2h / α"})
	for range 4 { // effort → tier → autonomy → cancel → Create
		tm.Send(press("tab"))
	}
	tm.Send(press("enter"))
	wait("created pasted-task", "# Pasted task") // wait for loaded detail, not just the flash
	tm.Send(press("q"))
	final := tm.FinalModel(t, teatest.WithFinalTimeout(5*time.Second)).(Model)
	task, ok := final.selectedTask()
	if !ok || task.Slug != "pasted-task" || task.Epic != "01-test" || task.Description != "Description 世界" ||
		task.Effort != "2h / α" || !slices.Equal(task.Tags, []string{"pasted", "authoring"}) {
		t.Fatalf("terminal paste lost or misrouted: %+v", task)
	}
}

func TestTaskCreateTerminalPasteOwnership(t *testing.T) {
	for _, state := range []string{"fields", "picker", "enum", "discard", "saving", "locked", "displaced", "tiny", "help"} {
		t.Run(state, func(t *testing.T) {
			root := seedRepo(t)
			m := openCreate(t, loadedAt(t, root, 120, 40))
			before := testutil.SnapshotTree(t, root)
			switch state {
			case "picker":
				m.createTask.focus(taskCreateEpic)
				m, _ = createKey(m, "enter")
			case "enum":
				m.createTask.more = true
				m.createTask.focus(taskCreatePriority)
				m, _ = createKey(m, "enter")
			case "discard":
				m.createTask.inputs[taskCreateTitle].SetValue("Keep this")
				m, _ = createKey(m, "esc")
			case "saving":
				fillCreate(&m)
				m.createTask.focus(taskCreateSubmit)
				m, _ = createKey(m, "enter") // hold the actual write command
			case "locked":
				m.createTask.locked = true
			case "displaced":
				m.createTask.displace()
			case "tiny":
				next, _ := m.Update(tea.WindowSizeMsg{Width: 30, Height: 10})
				m = next.(Model)
			case "help":
				m.showHelp = true // higher-priority modal must consume, not leak input
			}
			title, query := m.createTask.inputs[taskCreateTitle].Value(), m.createTask.query.Value()
			next, cmd := m.Update(tea.PasteMsg{Content: "qhjkn 世界"})
			m = settleCreate(t, next.(Model), cmd)
			switch state {
			case "fields":
				title = "qhjkn 世界"
			case "picker":
				query = "qhjkn 世界"
			}
			if m.createTask.inputs[taskCreateTitle].Value() != title || m.createTask.query.Value() != query ||
				m.createTask.inputs[taskCreatePriority].Value() != "" || !m.createTask.active ||
				!maps.Equal(before, testutil.SnapshotTree(t, root)) {
				t.Fatal("terminal paste escaped its visible editable owner or wrote a record")
			}
		})
	}
}

func TestTaskCreateReadOnlyAndEmptyEpicCatalog(t *testing.T) {
	root := seedRepo(t)
	m := loadedAt(t, root, 120, 40)
	m.svc = core.MustNewService(testutil.Must(store.NewFS(root, core.ReadOnlyMutations())))
	m = openCreate(t, m)
	fillCreate(&m)
	before := testutil.SnapshotTree(t, root)
	m.createTask.focus(taskCreateSubmit)
	m, cmd := createKey(m, "enter")
	m = settleCreate(t, m, cmd)
	if m.createTask.locked || !m.createTask.active || m.createTask.err == "" || !maps.Equal(before, testutil.SnapshotTree(t, root)) {
		t.Fatal("read-only refusal did not preserve the unwritten draft")
	}
	root = testutil.NewRepo(t).Root
	m = openCreate(t, loadedAt(t, root, 120, 40))
	if !strings.Contains(m.createTask.catalogStatus(), "No readable epics") {
		t.Fatal("empty catalog hides the prerequisite")
	}
	m.createTask.inputs[taskCreateTitle].SetValue("Retained draft")
	m.createTask.focus(taskCreateSubmit)
	m, _ = createKey(m, "enter")
	if m.createTask.saving || m.createTask.cursor != taskCreateEpic || m.createTask.inputs[taskCreateTitle].Value() != "Retained draft" {
		t.Fatal("empty epic submission lost draft or dispatched")
	}
	testutil.Write(t, filepath.Join(root, domain.EpicsDir, "01-new.md"), "---\nstatus: active\n---\n# New epic\n")
	m, _ = createKey(m, "enter") // choose the Epic field, not submit
	m, cmd = createKey(m, "ctrl+r")
	m = settleCreate(t, m, cmd)
	m, _ = createKey(m, "enter")
	if m.createTask.epic.key != "01-new" || m.createTask.inputs[taskCreateTitle].Value() != "Retained draft" {
		t.Fatal("explicit catalog reload could not recover the preserved draft")
	}
}

func TestTaskCreateDelayedWidgetResultsCannotEditAnotherField(t *testing.T) {
	m := openCreate(t, loaded(t, 120, 40))
	late := m.createTask.widgetCmd(func() tea.Msg { return press("x") })
	m.createTask.focus(taskCreateTags)
	next, _ := m.Update(late())
	m = next.(Model)
	if m.createTask.inputs[taskCreateTitle].Value() != "" || m.createTask.inputs[taskCreateTags].Value() != "" {
		t.Fatal("departed field's asynchronous result edited the newly focused field")
	}
	m.createTask.focus(taskCreateEpic)
	m, _ = createKey(m, "enter")
	late = m.createTask.widgetCmd(func() tea.Msg { return press("q") })
	m, _ = createKey(m, "esc")
	m, _ = createKey(m, "enter") // reopen the same picker with a new input generation
	next, _ = m.Update(late())
	m = next.(Model)
	if m.createTask.query.Value() != "" {
		t.Fatal("old picker input reached a new picker activation")
	}
}

func TestTaskCreateConstrainedOptionsStageCancelAndReset(t *testing.T) {
	for _, tc := range []struct {
		row   int
		value string
		want  []string
	}{
		{taskCreatePriority, "high", []string{"default", "high", "medium", "low"}},
		{taskCreateTier, "2", []string{"default", "1", "2", "3", "4", "5"}},
		{taskCreateAutonomy, "4", []string{"default", "1", "2", "3", "4", "5"}},
	} {
		t.Run(taskCreateLabels[tc.row], func(t *testing.T) {
			root := seedRepo(t)
			m := openCreate(t, loadedAt(t, root, 120, 40))
			before := testutil.SnapshotTree(t, root)
			m.createTask.more = true
			m.createTask.focus(tc.row)
			m, _ = createKey(m, "enter")
			if m.createTask.mode != taskCreateEnumPicker || taskCreateInput(tc.row) || !slices.Equal(taskCreateEnumOptions(tc.row), tc.want) {
				t.Fatal("bounded field did not use the expected enum selector")
			}
			if !strings.Contains(ansi.Strip(m.View().Content), "‹default›") {
				t.Fatal("selector did not show its current default choice")
			}
			for _, arbitrary := range []string{"9", "x", "q"} {
				m, _ = createKey(m, arbitrary)
			}
			if m.createTask.inputs[tc.row].Value() != "" || m.createTask.enumCursor != 0 {
				t.Fatal("arbitrary text entered a constrained field")
			}
			for range slices.Index(tc.want, tc.value) {
				m, _ = createKey(m, "j")
			}
			m, _ = createKey(m, "enter")
			if m.createTask.mode == taskCreateEnumPicker || m.createTask.saving || m.createTask.inputs[tc.row].Value() != tc.value || !m.createTask.dirty() {
				t.Fatal("accept did not stage the value without submitting")
			}
			m.createTask.focus(tc.row)
			m, _ = createKey(m, "enter")
			m, _ = createKey(m, "l")
			m, _ = createKey(m, "esc")
			if !m.createTask.active || m.createTask.mode == taskCreateDiscard || m.createTask.inputs[tc.row].Value() != tc.value {
				t.Fatal("Esc did not cancel just the unaccepted selector choice")
			}
			m, _ = createKey(m, "esc")
			if m.createTask.mode != taskCreateDiscard {
				t.Fatal("a staged enum override was not treated as dirty input")
			}
			m, _ = createKey(m, "enter") // keep editing
			m, _ = createKey(m, "enter") // reopen enum
			for m.createTask.enumCursor != 0 {
				m, _ = createKey(m, "k")
			}
			m, _ = createKey(m, "tab")
			if m.createTask.inputs[tc.row].Value() != "" || m.createTask.dirty() || !maps.Equal(before, testutil.SnapshotTree(t, root)) {
				t.Fatal("default reset pinned UI policy or changed persistent data")
			}
		})
	}
}

func TestTaskCreateReceiptRecoveryNeverReplays(t *testing.T) {
	for _, outcome := range []string{"committed cleanup", "unknown", "unconfirmed success", "uncommitted finalization"} {
		t.Run(outcome, func(t *testing.T) {
			root := seedRepo(t)
			fs := testutil.Must(store.NewFS(root, core.UnrestrictedMutations()))
			port := &createReceiptPort{setID: fs.SourceSetID()}
			switch outcome {
			case "committed cleanup":
				port.committer, port.err = fs, &core.CreateFinalizationError{Cause: errors.New("release failed")}
			case "unknown":
				port.err = errors.New("transport disconnected")
			case "uncommitted finalization":
				port.err = &core.CreateFinalizationError{Cause: domain.ErrConflict}
			}
			m := loadedAt(t, root, 120, 40)
			m.svc = core.MustNewService(fs, core.WithSelectedEpicTaskCreationStore(port))
			m = openCreate(t, m)
			fillCreate(&m)
			m.createTask.focus(taskCreateSubmit)
			m, write := createKey(m, "enter")
			m = settleCreate(t, m, write)
			if !m.createTask.locked || m.createTask.err == "" || port.calls != 1 || m.createTask.inputs[taskCreateTitle].Value() != "Ship a task" {
				t.Fatal("uncertain/committed attempt did not retain non-replayable recovery")
			}
			m, cmd := createKey(m, "enter")
			m = settleCreate(t, m, cmd)
			if port.calls != 1 {
				t.Fatal("recovery replayed Create")
			}
			if outcome == "committed cleanup" {
				if task, ok := m.selectedTask(); !ok || task.Slug != "ship-a-task" {
					t.Fatal("inspect did not select the committed task")
				}
			}
		})
	}
}

func TestTaskCreateOriginBindingAndDisplacedRecovery(t *testing.T) {
	for _, issued := range []bool{false, true} {
		t.Run(map[bool]string{false: "draft", true: "issued"}[issued], func(t *testing.T) {
			alpha, beta := seedRepo(t), seedRepo(t)
			m := openCreate(t, loadedAt(t, alpha, 120, 40))
			m.sessionScope = true
			fillCreate(&m)
			var write tea.Cmd
			if issued {
				m.createTask.focus(taskCreateSubmit)
				m, write = createKey(m, "enter")
			}
			oldGen := m.sessionGen
			beforeBeta := testutil.SnapshotTree(t, beta)
			incoming := core.Workspace{Checkout: beta, PlanningRoot: beta,
				Planning: core.MustNewService(testutil.Must(store.NewFS(beta, core.UnrestrictedMutations())))}
			m.activateWorkspace(incoming, nil, nil)
			if !m.createTask.active || !m.createTask.displaced || !m.createTask.locked || m.createTask.inputs[taskCreateTitle].Value() != "Ship a task" {
				t.Fatal("workspace activation lost/rebound the draft")
			}
			if issued {
				m = settleCreate(t, m, write)
				if !m.createTask.receipt.Committed || m.createTask.saving {
					t.Fatal("old-session result lost its origin receipt")
				}
				record, err := m.createTask.origin.ShowTask(m.createTask.receipt.Task.ID)
				if err != nil || record.Value.Task.Slug != "ship-a-task" {
					t.Fatal("issued command did not write exclusively to its originating Service")
				}
			}
			if m.sessionGen != oldGen+1 || !maps.Equal(beforeBeta, testutil.SnapshotTree(t, beta)) {
				t.Fatal("departed form changed the incoming workspace")
			}
			m, cmd := createKey(m, "enter")
			if cmd != nil || !maps.Equal(beforeBeta, testutil.SnapshotTree(t, beta)) {
				t.Fatal("displaced recovery submitted into the new workspace")
			}
		})
	}
}

func TestTaskCreateBlocksPendingWorkspaceOpenAndFitsSmallTerminals(t *testing.T) {
	m := loaded(t, 120, 40)
	m.atlas.opening = true
	m, _ = createKey(m, "c")
	if m.createTask.active || !strings.Contains(m.flash, "pending workspace") {
		t.Fatal("creation opened under an outstanding workspace open")
	}
	m.atlas.opening = false
	m = openCreate(t, m)
	fillCreate(&m)
	m.createTask.more = true
	for _, size := range [][2]int{{120, 40}, {72, 24}, {48, 20}, {30, 10}, {12, 5}} {
		for _, row := range m.createTask.rows() {
			m.createTask.focus(row)
			next, _ := m.Update(tea.WindowSizeMsg{Width: size[0], Height: size[1]})
			m = next.(Model)
			v := m.View().Content
			if lipgloss.Width(v) > size[0] || lipgloss.Height(v) > size[1] {
				t.Fatalf("creation overflows %v: %dx%d", size, lipgloss.Width(v), lipgloss.Height(v))
			}
			if size[0] >= 48 && size[1] >= 20 && !strings.Contains(ansi.Strip(v), "[Create]") {
				t.Fatalf("explicit action hidden at %v:\n%s", size, v)
			}
		}
	}
	if m.createTask.inputs[taskCreateTitle].Value() != "Ship a task" {
		t.Fatal("resizing changed input")
	}
}
