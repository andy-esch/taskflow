package tui

import (
	"bytes"
	"errors"
	"fmt"
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

func fillThreadCreate(m Model) Model {
	// Literal keyboard sequence, independent of production row helpers.
	m = pasteCreation(m, "Empty / qhjkn 世界")
	m, _ = createKey(m, "enter")
	m = pasteCreation(m, "A bounded initiative α")
	m, _ = createKey(m, "enter")
	m = pasteCreation(m, "Ship the useful thing")
	return m
}

func submitThreadCreate(m Model) (Model, tea.Cmd) {
	m.creation.thread.controls().focus(threadCreateSubmit)
	return createKey(m, "enter")
}

func TestThreadCreateDurableEmptyUnstarted(t *testing.T) {
	for _, optional := range []bool{false, true} {
		t.Run(fmt.Sprint(optional), func(t *testing.T) {
			root := seedRepo(t)
			fs := testutil.Must(store.NewFS(root, core.UnrestrictedMutations()))
			// Include a real dependency, not merely two unrelated task fixtures.
			if _, err := core.MustNewService(fs).AddTaskDependencies("beta", []string{"alpha"}, false); err != nil {
				t.Fatal(err)
			}
			m := loadedAt(t, root, 120, 40)
			m.svc = core.MustNewService(fs, core.WithClock(func() time.Time { return time.Date(2026, 10, 24, 0, 0, 0, 0, time.UTC) }))
			m = openCreationKind(t, m, entityThreads)
			if m.creation.thread.dirty() {
				t.Fatal("untouched empty Thread form was dirty")
			}
			m = fillThreadCreate(m)
			var wantTags []string
			wantDate := ""
			if optional {
				m, _ = createKey(m, "enter") // goal -> target date
				m = pasteCreation(m, "2026-12-01")
				m, _ = createKey(m, "enter")
				m = pasteCreation(m, "tui, dogfood")
				wantDate, wantTags = "2026-12-01", []string{"tui", "dogfood"}
			}
			before := testutil.SnapshotTree(t, root)
			m, write := submitThreadCreate(m)
			m, duplicate := createKey(m, "enter")
			if write == nil || duplicate != nil || !maps.Equal(before, testutil.SnapshotTree(t, root)) {
				t.Fatal("issued twice or wrote before command execution")
			}
			m = settleCreate(t, m, write)
			ref := m.selectedRef()
			view, body, err := m.svc.ShowThread(ref.key)
			if err != nil {
				t.Fatal(err)
			}
			thread := view.Thread
			if m.activeCreation() != nil || ref.key == "" || view.Source.ID != ref.key || thread.ID != ref.key || thread.Slug != "empty-qhjkn-世界" ||
				thread.Status != domain.ThreadStatusUnstarted || len(thread.Tasks) != 0 || thread.Created != "2026-10-24" ||
				thread.StartedAt != "" || thread.EndedAt != "" || thread.Updated != "" || thread.Description != "A bounded initiative α" ||
				thread.Goal != "Ship the useful thing" || thread.TargetDate != wantDate || !slices.Equal(thread.Tags, wantTags) ||
				!strings.Contains(body, "# Thread: Empty / qhjkn 世界") || !strings.Contains(body, "Ship the useful thing") {
				t.Fatalf("wrong durable Thread or scaffold: %+v\n%s", thread, body)
			}
			after := testutil.SnapshotTree(t, root)
			delete(after, filepath.Join(domain.ThreadsDir, ref.key+"-empty-qhjkn-世界.md"))
			if _, exists := before[domain.ThreadsDir]; !exists {
				delete(after, domain.ThreadsDir)
			}
			if !maps.Equal(before, after) {
				t.Fatal("Thread creation changed a task, edge, existing document, mode or unrelated directory")
			}
			m = settleCreate(t, m, m.reloadAll())
			if m.selectedRef() != ref || selectedThreadDetail(t, m).projection.View.Source.ID != ref.key {
				t.Fatal("reload lost the exact new Thread")
			}
		})
	}
}

func TestThreadCreateValidationAndGraphRefusals(t *testing.T) {
	for _, refusal := range []string{"title", "description", "goal", "target date", "broken graph", "unreadable Thread", "read-only"} {
		t.Run(refusal, func(t *testing.T) {
			root := seedRepo(t)
			m := loadedAt(t, root, 120, 40)
			if refusal == "read-only" {
				m.svc = core.MustNewService(testutil.Must(store.NewFS(root, core.ReadOnlyMutations())))
			}
			m = fillThreadCreate(openCreationKind(t, m, entityThreads))
			row := -1
			switch refusal {
			case "title":
				row = threadCreateTitle
			case "description":
				row = threadCreateDescription
			case "goal":
				row = threadCreateGoal
			case "target date":
				m.creation.thread.controls().focus(threadCreateTargetDate)
				m = pasteCreation(m, "2026-02-30")
			case "broken graph":
				testutil.Write(t, filepath.Join(root, domain.TasksDir, testutil.TaskID("broken")+"-broken.md"),
					"---\nid: "+testutil.TaskID("broken")+"\nstatus: ready-to-start\ndepends_on: [000000000001]\n---\n# Broken\n")
			case "unreadable Thread":
				testutil.Write(t, filepath.Join(root, domain.ThreadsDir, testutil.TaskID("broken-thread")+"-broken.md"), "---\nstatus: [\n---\n# Broken\n")
			}
			if row >= 0 {
				m.creation.thread.inputs[row].SetValue("")
			}
			before := testutil.SnapshotTree(t, root)
			m, cmd := submitThreadCreate(m)
			m = settleCreate(t, m, cmd)
			if f := m.activeCreation(); f == nil || f.locked || f.saving || f.err == "" || !maps.Equal(before, testutil.SnapshotTree(t, root)) {
				t.Fatal("refusal lost input, wrote or disabled deliberate correction")
			}
			if row >= 0 || refusal == "target date" {
				if row >= 0 {
					m.creation.thread.inputs[row].SetValue("Repaired")
				} else {
					m.creation.thread.inputs[threadCreateTargetDate].SetValue("2026-12-01")
				}
				m, cmd = submitThreadCreate(m)
				m = settleCreate(t, m, cmd)
				if m.activeCreation() != nil || m.selectedRef().empty() {
					t.Fatal("deliberate correction/retry did not create")
				}
			}
		})
	}
}

// Read-only portable facade with no path or mutation interfaces. A separate
// creation capability may be explicitly composed over the same source witness.
type threadCreationReads struct {
	core.Store
	core.ThreadStore
	setID     core.SourceSetID
	readCalls int
	memory    *core.ThreadWithBody
}

func (p *threadCreationReads) SourceSetID() core.SourceSetID { return p.setID }
func (p *threadCreationReads) ReadThreads() (core.ThreadRead, error) {
	p.readCalls++
	if p.memory != nil {
		return core.ThreadRead{Records: []core.VersionedRecord[domain.Thread]{{Record: core.LoadedRecord[domain.Thread]{
			Value: p.memory.Thread, Source: core.RecordSource{ID: p.memory.Thread.ID, Location: "opaque:thread"}}}}}, nil
	}
	return p.ThreadStore.ReadThreads()
}
func (p *threadCreationReads) ReadThread(ref string) (core.LoadedRecord[core.ThreadWithBody], error) {
	p.readCalls++
	if p.memory != nil && ref == p.memory.Thread.ID {
		return core.LoadedRecord[core.ThreadWithBody]{Value: *p.memory, Source: core.RecordSource{ID: ref, Location: "opaque:thread"}}, nil
	}
	return p.ThreadStore.ReadThread(ref)
}

type threadCreatePort struct {
	reads   *threadCreationReads
	fs      *store.FS
	outcome string
	calls   int
	plan    core.ThreadCreationPlan
}

func (p *threadCreatePort) SourceSetID() core.SourceSetID { return p.reads.setID }
func (p *threadCreatePort) MutateThreadCreation(now time.Time, dry bool, planner core.ThreadCreationPlanner) (core.ThreadCreationMutationResult, error) {
	p.calls++
	if p.outcome == "cleanup" || p.outcome == "success" {
		result, err := p.fs.MutateThreadCreation(now, dry, planner)
		p.plan = result.Plan
		if err == nil && p.outcome == "cleanup" {
			err = domain.ErrConflict
		} // core wraps real committed evidence; no retry
		return result, err
	}
	snapshot := core.ThreadCreationSnapshot{Graph: core.NewTaskGraph(nil, nil)}
	plan, err := planner(snapshot)
	if err != nil {
		return core.ThreadCreationMutationResult{}, err
	}
	plan, err = core.ValidateThreadCreationPlan(snapshot, plan)
	if err != nil {
		return core.ThreadCreationMutationResult{}, err
	}
	p.plan = plan
	result := core.ThreadCreationMutationResult{Plan: plan, Thread: plan.Thread, Changed: true}
	switch p.outcome {
	case "pathless":
		p.reads.memory = &core.ThreadWithBody{Thread: plan.Thread, Body: plan.Body}
		result.Committed = true
		result.Local.CommittedPath = "opaque:not-navigation-authority"
		return result, nil
	case "refusal":
		return core.ThreadCreationMutationResult{}, domain.ErrValidation
	case "planned refusal":
		result.Local.PlannedPath = "/only-a-plan"
		return result, domain.ErrValidation
	case "conflict":
		return core.ThreadCreationMutationResult{}, domain.ErrConflict
	case "unknown":
		return core.ThreadCreationMutationResult{}, errors.New("unknown transport result")
	case "marked failure":
		return core.ThreadCreationMutationResult{}, &core.ThreadCreationMutationFailure{Cause: domain.ErrValidation, Receipt: core.ThreadCreationReceipt{Committed: true}}
	case "path evidence":
		result.Local.CommittedPath = "/durable-evidence-without-identity"
		return result, domain.ErrValidation
	case "preview":
		result.DryRun = true
		return result, nil
	case "unconfirmed":
		return result, nil
	default:
		panic("unsupported Thread creation test outcome: " + p.outcome)
	}
}

func threadCreationService(fs *store.FS, outcome string) (*core.Service, *threadCreatePort) {
	reads := &threadCreationReads{Store: fs, ThreadStore: fs, setID: fs.SourceSetID()}
	port := &threadCreatePort{reads: reads, fs: fs, outcome: outcome}
	return core.MustNewService(reads, core.WithThreadCreationMutationStore(port), core.WithRetry(1, func(int) {})), port
}

func TestThreadCreatePortableReceiptsAndCoreRetry(t *testing.T) {
	for _, outcome := range []string{"pathless", "refusal", "planned refusal", "conflict", "unknown", "unconfirmed", "marked failure", "path evidence", "preview", "cleanup"} {
		t.Run(outcome, func(t *testing.T) {
			root := seedRepo(t)
			fs := testutil.Must(store.NewFS(root, core.UnrestrictedMutations()))
			m := loadedAt(t, root, 120, 40)
			svc, port := threadCreationService(fs, outcome)
			m.svc = svc
			m = fillThreadCreate(openCreationKind(t, m, entityThreads))
			before := testutil.SnapshotTree(t, root)
			m, write := submitThreadCreate(m)
			m = settleCreate(t, m, write)
			wantCalls := 1
			if outcome == "conflict" {
				wantCalls = 2
			} // core's one retry, never a UI replay
			if port.calls != wantCalls || len(port.plan.Thread.Tasks) != 0 || port.plan.Thread.Status != domain.ThreadStatusUnstarted ||
				port.plan.Thread.Description != "A bounded initiative α" || port.plan.Thread.Goal != "Ship the useful thing" || !strings.Contains(port.plan.Body, "# Thread: Empty / qhjkn 世界") {
				t.Fatalf("typed guarded request was misrouted: calls=%d plan=%+v", port.calls, port.plan)
			}
			if outcome == "pathless" {
				if m.activeCreation() != nil || m.selectedRef().key != port.plan.Thread.ID || strings.Contains(m.flash, "editor") || selectedThreadDetail(t, m).projection.View.Source.ID != port.plan.Thread.ID {
					t.Fatal("pathless landing lost exact identity or promised editor access")
				}
			} else {
				known := outcome == "refusal" || outcome == "planned refusal" || outcome == "conflict"
				if f := m.activeCreation(); f == nil || f.saving || f.locked == known || creationTitle(m) != "Empty / qhjkn 世界" {
					t.Fatal("proposal/durability classification lost draft or chose wrong retry policy")
				}
				if known {
					port.outcome = "pathless"
					m, write = createKey(m, "enter") // deliberate retry only after refusal
					m = settleCreate(t, m, write)
					if m.activeCreation() != nil || port.calls != wantCalls+1 {
						t.Fatal("known refusal could not deliberately retry")
					}
				} else {
					m, follow := createKey(m, "enter")
					m = settleCreate(t, m, follow)
					if port.calls != wantCalls {
						t.Fatal("recovery repeated guarded creation")
					}
					if outcome == "cleanup" && m.selectedRef().key != port.plan.Thread.ID {
						t.Fatal("committed cleanup inspection missed exact Thread")
					}
				}
			}
			if outcome != "cleanup" && !maps.Equal(before, testutil.SnapshotTree(t, root)) {
				t.Fatal("portable/proposed outcome changed real filesystem bytes or modes")
			}
		})
	}
}

func TestThreadCreateMissingCapabilityAndSourceBinding(t *testing.T) {
	root := seedRepo(t)
	fs := testutil.Must(store.NewFS(root, core.UnrestrictedMutations()))
	reads := &threadCreationReads{Store: fs, ThreadStore: fs, setID: fs.SourceSetID()}
	m := loadedAt(t, root, 120, 40)
	m.svc = core.MustNewService(reads) // no creation capability, but full browsing still works
	m = fillThreadCreate(openCreationKind(t, m, entityThreads))
	before := testutil.SnapshotTree(t, root)
	m, cmd := submitThreadCreate(m)
	m = settleCreate(t, m, cmd)
	if f := m.activeCreation(); f == nil || !f.locked || f.saving || !strings.Contains(f.err, "unavailable") || !maps.Equal(before, testutil.SnapshotTree(t, root)) {
		t.Fatal("missing capability lost draft or issued a local fallback write")
	}
	// Missing capability is an unclassified application error: retain safe
	// recovery rather than infer retry permission by parsing the message.
	_, foreign := threadCreationService(testutil.Must(store.NewFS(seedRepo(t), core.UnrestrictedMutations())), "success")
	if _, err := core.NewService(reads, core.WithThreadCreationMutationStore(foreign)); !errors.Is(err, domain.ErrValidation) {
		t.Fatal("foreign creation source witness was accepted")
	}
}

func TestThreadCreateRevalidatesAfterDispatch(t *testing.T) {
	for _, drift := range []string{"cross-kind identity", "authorization"} {
		t.Run(drift, func(t *testing.T) {
			root := seedRepo(t)
			allowed := true
			fs := testutil.Must(store.NewFS(root, core.GuardedMutations(func() error {
				if !allowed {
					return core.ErrReadOnlyPersistence
				}
				return nil
			})))
			reservedID := testutil.TaskID("late-collision")
			m := loadedAt(t, root, 120, 40)
			m.svc = core.MustNewService(fs, core.WithIDGen(func() string { return reservedID }), core.WithRetry(1, func(int) {}))
			m = fillThreadCreate(openCreationKind(t, m, entityThreads))
			m, write := submitThreadCreate(m) // hold the actual origin command
			if drift == "cross-kind identity" {
				testutil.Write(t, filepath.Join(root, domain.TasksDir, reservedID+"-new-owner.md"),
					"---\nid: "+reservedID+"\nstatus: ready-to-start\n---\n# New task owner\n")
			} else {
				allowed = false
			}
			before := testutil.SnapshotTree(t, root)
			m = settleCreate(t, m, write)
			if f := m.activeCreation(); f == nil || f.locked || f.saving || f.err == "" || creationTitle(m) != "Empty / qhjkn 世界" || !maps.Equal(before, testutil.SnapshotTree(t, root)) {
				t.Fatal("guarded execution borrowed stale graph identity or authorization")
			}
		})
	}
}

func TestThreadCreatePaletteLandingAndActualResultGuards(t *testing.T) {
	root := seedRepo(t)
	m := loadedAt(t, root, 120, 40)
	m = settleCreate(t, m, m.switchTab(indexOfKind(m.tabs, entityThreads)))
	m.setFocus(focusDetail)
	m, cmd := createKey(m, "ctrl+p")
	m = settleCreate(t, m, cmd)
	found := false
	for i, raw := range m.palette.list.Items() {
		if it := raw.(paletteItem); it.kind == palAction && it.word == "new-thread" {
			m.palette.list.Select(i)
			found = true
			break
		}
	}
	if !found {
		t.Fatal("New Thread action missing")
	}
	m, cmd = createKey(m, "enter")
	m = fillThreadCreate(settleCreate(t, m, cmd))
	m.cur().list.SetFilterText("no match")
	m, write := submitThreadCreate(m)
	result := write().(threadCreatedMsg)
	stale := result
	stale.gen++
	next, follow := m.Update(stale)
	m = settleCreate(t, next.(Model), follow)
	if !m.creation.thread.saving || follow != nil {
		t.Fatal("actual stale result altered a pending attempt")
	}
	next, landing := m.Update(result)
	m = next.(Model)
	m = settleCreate(t, m, m.reloadAll()) // newer watcher reload before original landing
	m = settleCreate(t, m, landing)
	ref := m.selectedRef()
	if ref.key != result.receipt.Thread.ID || m.activeCreation() != nil {
		t.Fatal("filter/reload lost the created Thread")
	}
	before := testutil.SnapshotTree(t, root)
	next, follow = m.Update(result)
	m = settleCreate(t, next.(Model), follow)
	if follow != nil || m.selectedRef() != ref || !maps.Equal(before, testutil.SnapshotTree(t, root)) {
		t.Fatal("duplicate outcome replayed or stole selection")
	}
	m.atlas.opening = true
	m, _ = createKey(m, "c")
	if m.activeCreation() != nil || !strings.Contains(m.flash, "pending workspace") {
		t.Fatal("opened underneath pending workspace activation")
	}
}

func TestThreadCreateOriginAndDisplacedRecovery(t *testing.T) {
	for _, issued := range []bool{false, true} {
		for _, outcome := range []string{"success", "cleanup", "unknown"} {
			t.Run(fmt.Sprintf("issued=%t/%s", issued, outcome), func(t *testing.T) {
				alpha, beta := seedRepo(t), seedRepo(t)
				fs := testutil.Must(store.NewFS(alpha, core.UnrestrictedMutations()))
				origin, port := threadCreationService(fs, outcome)
				m := loadedAt(t, alpha, 120, 40)
				m.svc, m.sessionScope = origin, true
				m = fillThreadCreate(openCreationKind(t, m, entityThreads))
				var write tea.Cmd
				if issued {
					m, write = submitThreadCreate(m)
				}
				incomingFS := testutil.Must(store.NewFS(beta, core.UnrestrictedMutations()))
				incoming, incomingPort := threadCreationService(incomingFS, "success")
				m.activateWorkspace(core.Workspace{Checkout: beta, PlanningRoot: beta, Planning: incoming}, nil, nil)
				before := testutil.SnapshotTree(t, beta)
				reads := incomingPort.reads.readCalls
				if issued {
					result := write()
					payload := result.(sessionMsg).msg.(threadCreatedMsg)
					if payload.receipt.Committed {
						thread := payload.receipt.Thread
						thread.Description = "Incoming same-ID document"
						if _, err := incomingFS.MutateThreadCreation(time.Now(), false, func(core.ThreadCreationSnapshot) (core.ThreadCreationPlan, error) {
							return core.ThreadCreationPlan{Thread: thread, Body: "# Incoming same ID\n"}, nil
						}); err != nil {
							t.Fatal(err)
						}
						view, _, err := origin.ShowThread(thread.ID)
						if err != nil || view.Thread.Description != "A bounded initiative α" {
							t.Fatal("origin was not durably created")
						}
						before = testutil.SnapshotTree(t, beta)
					}
					next, cmd := m.Update(result)
					m = settleCreate(t, next.(Model), cmd)
					next, cmd = m.Update(result)
					m = settleCreate(t, next.(Model), cmd)
					if f := m.activeCreation(); f == nil || f.saving || !f.displaced || !f.locked || port.calls != 1 {
						t.Fatal("departed outcome lost recovery or replayed")
					}
				}
				prior := m
				m, follow := createKey(m, "enter") // close only, even with a committed same-ID incoming record
				m = settleCreate(t, m, follow)
				if incomingPort.calls != 0 || incomingPort.reads.readCalls != reads || !maps.Equal(before, testutil.SnapshotTree(t, beta)) || !prior.creation.thread.active || creationTitle(prior) != "Empty / qhjkn 世界" {
					t.Fatal("recovery reached incoming source or aliased its retained prior copy")
				}
			})
		}
	}
}

func TestThreadCreateSubmitOriginAndSessionGuards(t *testing.T) {
	for _, change := range []string{"service", "session"} {
		t.Run(change, func(t *testing.T) {
			root := seedRepo(t)
			m := fillThreadCreate(openCreationKind(t, loadedAt(t, root, 120, 40), entityThreads))
			if change == "service" {
				m.svc = core.MustNewService(testutil.Must(store.NewFS(root, core.UnrestrictedMutations())))
			} else {
				m.sessionGen++
			}
			before := testutil.SnapshotTree(t, root)
			m, cmd := submitThreadCreate(m)
			if cmd != nil || !m.creation.thread.displaced || !m.creation.thread.locked || !maps.Equal(before, testutil.SnapshotTree(t, root)) {
				t.Fatal("stale submission authority issued work")
			}
		})
	}
}

func TestThreadCreateNarrowModes(t *testing.T) {
	for _, mode := range []string{"fields", "discard", "pending", "recovery"} {
		t.Run(mode, func(t *testing.T) {
			m := fillThreadCreate(openCreationKind(t, loaded(t, 120, 40), entityThreads))
			switch mode {
			case "discard":
				m, _ = createKey(m, "esc")
			case "pending":
				m, _ = submitThreadCreate(m)
			case "recovery":
				m.creation.thread.displace()
			}
			for _, size := range [][2]int{{120, 40}, {46, 18}, {45, 18}, {46, 17}, {12, 5}, {120, 40}} {
				next, _ := m.Update(tea.WindowSizeMsg{Width: size[0], Height: size[1]})
				m = next.(Model)
				view := m.View().Content
				if lipgloss.Width(view) > size[0] || lipgloss.Height(view) > size[1] {
					t.Fatal("Thread modal overflow")
				}
				if mode == "fields" && size[0] >= 46 && size[1] >= 18 && !strings.Contains(ansi.Strip(view), "[Create]") {
					t.Fatal("supported form hid Create")
				}
			}
		})
	}
}

func TestThreadCreateInRealProgram(t *testing.T) {
	m := loaded(t, 120, 40)
	m = settleCreate(t, m, m.switchTab(indexOfKind(m.tabs, entityThreads)))
	tm := teatest.NewTestModel(t, m, teatest.WithInitialTermSize(120, 40))
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
	tm.Send(press("c"))
	wait("New Thread")
	// Exercise dirty Keep/Discard before doing the actual keyboard-only create.
	tm.Send(tea.PasteMsg{Content: "Discard this"})
	tm.Send(press("esc"))
	wait("Discard unsaved input?")
	tm.Send(press("enter"))
	wait("Discard this")
	tm.Send(press("esc"))
	tm.Send(press("y"))
	wait("c new thread") // let the running renderer settle the closed state
	tm.Send(press("c"))
	wait("New Thread")
	for _, value := range []string{"Runtime Thread 世界", "Useful initiative", "Ship the outcome", "2026-12-01", "tui, runtime"} {
		tm.Send(tea.PasteMsg{Content: value})
		tm.Send(press("enter"))
	}
	tm.Send(press("tab")) // Cancel -> Create
	tm.Send(press("enter"))
	// The success flash precedes async selection/detail loading. The empty
	// rollup belongs to the loaded Thread, never the form or its success message.
	wait("created runtime-thread-世界", "0/0 nominally done")
	tm.Send(press("q"))
	final := tm.FinalModel(t, teatest.WithFinalTimeout(5*time.Second)).(Model)
	view, _, err := final.svc.ShowThread(final.selectedRef().key)
	if err != nil || final.activeCreation() != nil || view.Thread.Status != domain.ThreadStatusUnstarted || len(view.Thread.Tasks) != 0 ||
		view.Thread.Description != "Useful initiative" || view.Thread.Goal != "Ship the outcome" || view.Thread.TargetDate != "2026-12-01" || !slices.Equal(view.Thread.Tags, []string{"tui", "runtime"}) {
		t.Fatalf("runtime did not create exact empty Thread: %+v %v", view.Thread, err)
	}
}
