package tui

import (
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/andy-esch/taskflow/internal/core"
	"github.com/andy-esch/taskflow/internal/domain"
	"github.com/andy-esch/taskflow/internal/store"
	"github.com/andy-esch/taskflow/internal/testutil"
)

var errPlannerWindow = fmt.Errorf("repository mutation planner is active: %w", domain.ErrConflict)

// countingRead is a stand-in for one asynchronous read. It records how many
// times the policy actually ran it — the difference between "retried once" and
// "spinning" is a count, not a shape.
type countingRead struct {
	calls int
	msgs  []tea.Msg
}

func (r *countingRead) cmd() tea.Cmd {
	return func() tea.Msg {
		msg := r.msgs[min(r.calls, len(r.msgs)-1)]
		r.calls++
		return msg
	}
}

// TestConflictRetriesTheSameRequestAfterOneQuietPeriod pins the whole bounded
// path with the real timer: a conflict is swallowed, the identical read is
// re-run one filesystem quiet period later, and its result lands normally.
func TestConflictRetriesTheSameRequestAfterOneQuietPeriod(t *testing.T) {
	m, _ := threadModel(t)
	list, _, err := m.svc.ListThreadViews()
	if err != nil {
		t.Fatal(err)
	}
	read := &countingRead{msgs: []tea.Msg{
		threadListLoadedMsg{gen: 1, err: errPlannerWindow},
		threadListLoadedMsg{gen: 1, list: list},
	}}
	request := readRequest{surface: readThreadList, gen: 1}
	m.threads.listGen = 1

	msg := withReadConflictRetry(request, read.cmd())()
	conflict, ok := msg.(readConflictMsg)
	if !ok {
		t.Fatalf("a conflicting read = %T, want it held as a readConflictMsg", msg)
	}
	if conflict.request != request {
		t.Fatalf("conflict carried request %+v, want %+v", conflict.request, request)
	}
	start := time.Now()
	tm, cmd := m.Update(conflict)
	m = tm.(Model)
	if m.threads.loadErr != nil || m.threads.loaded {
		t.Fatalf("a first conflict must not become a visible error or a false load: err=%v loaded=%v",
			m.threads.loadErr, m.threads.loaded)
	}
	if cmd == nil {
		t.Fatal("a current request's conflict should schedule a retry")
	}

	retry, ok := cmd().(readRetryMsg)
	if !ok {
		t.Fatalf("scheduled retry produced %T, want readRetryMsg", retry)
	}
	// tea.Tick starts its timer when Update creates the command, so measure from
	// before that reduction. Keep a little scheduling tolerance while still
	// pinning substantially the configured quiet period.
	if waited := time.Since(start); waited < fsDebounce*3/4 {
		t.Errorf("retry fired after %s, want it deferred by the quiet period (%s)", waited, fsDebounce)
	}
	if retry.request != request {
		t.Fatalf("retry carried request %+v, want %+v", retry.request, request)
	}

	tm, cmd = m.Update(retry)
	m = tm.(Model)
	m = drain(t, m, cmd)
	if !m.threads.loaded || m.threads.loadErr != nil {
		t.Fatalf("the retry's success should land: loaded=%v err=%v", m.threads.loaded, m.threads.loadErr)
	}
	if read.calls != 2 {
		t.Errorf("the read ran %d times, want exactly one retry", read.calls)
	}
}

// TestRepeatedConflictBecomesVisibleWithoutSpinning pins the bound: the retry is
// the UNWRAPPED read, so a second conflict is an ordinary durable error and
// schedules nothing further.
func TestRepeatedConflictBecomesVisibleWithoutSpinning(t *testing.T) {
	m, _ := threadModel(t)
	read := &countingRead{msgs: []tea.Msg{threadListLoadedMsg{gen: 1, err: errPlannerWindow}}}
	request := readRequest{surface: readThreadList, gen: 1}
	m.threads.listGen = 1

	conflict := withReadConflictRetry(request, read.cmd())().(readConflictMsg)
	tm, _ := m.Update(conflict)
	m = tm.(Model)
	// Skip the timer, not the reducer: this is the message the tick delivers.
	tm, cmd := m.Update(readRetryMsg(conflict))
	m = tm.(Model)
	if cmd == nil {
		t.Fatal("the scheduled retry should re-run the read")
	}
	again := cmd()
	if _, held := again.(readConflictMsg); held {
		t.Fatal("a retry's own conflict must not be held for another retry")
	}
	tm, cmd = m.Update(again)
	m = tm.(Model)
	if !errors.Is(m.threads.loadErr, domain.ErrConflict) {
		t.Errorf("a repeated conflict should be visible, got err=%v", m.threads.loadErr)
	}
	if cmd != nil {
		t.Error("a repeated conflict must not schedule more work")
	}
	if read.calls != 2 {
		t.Errorf("the read ran %d times, want exactly one retry", read.calls)
	}
}

func TestRepeatedEntityDetailConflictRetainsOnlyTheSameLoadedRecord(t *testing.T) {
	m := loaded(t, 120, 40)
	id := m.selectedID()
	wantBody := "last coherent detail"
	m.detail.SetContent(id, taskDetail{t: domain.Task{Slug: "display-slug"}, body: wantBody})
	m.detailGen++
	request := readRequest{
		surface: readEntityDetail,
		kind:    m.cur().kind,
		id:      id,
		gen:     m.detailGen,
	}
	read := &countingRead{msgs: []tea.Msg{
		detailErrMsg{kind: request.kind, id: id, gen: request.gen, err: errPlannerWindow},
		detailErrMsg{kind: request.kind, id: id, gen: request.gen, err: errPlannerWindow},
	}}
	conflict := withReadConflictRetry(request, read.cmd())().(readConflictMsg)
	tm, _ := m.Update(conflict)
	m = tm.(Model)
	tm, retry := m.Update(readRetryMsg(conflict))
	m = tm.(Model)
	if retry == nil {
		t.Fatal("the current detail request should run its one retry")
	}
	tm, _ = m.Update(retry())
	m = tm.(Model)
	if m.detail.content == nil || m.detail.content.rawBody() != wantBody {
		t.Fatal("a repeated conflict blanked the same record's coherent detail")
	}
	if !m.detail.refreshFailed() || !strings.Contains(ansi.Strip(m.footer()), "detail refresh failed") {
		t.Error("the retained detail's failed refresh is not visible in the footer")
	}

	// A display title can match the selected slug while the canonical loaded
	// record does not. That is not safe to retain after the current read fails.
	m.detail.SetContent("another-stable-id", taskDetail{t: domain.Task{Slug: id}, body: "another record"})
	m.detailGen++
	tm, _ = m.Update(detailErrMsg{
		kind: m.cur().kind,
		id:   id,
		gen:  m.detailGen,
		err:  errPlannerWindow,
	})
	m = tm.(Model)
	if m.detail.content != nil || m.detail.loadedID != "" {
		t.Error("a conflict for a different canonical record retained unrelated detail")
	}
}

// TestNewerGenerationSupersedesAScheduledRetry pins that a retry is only ever
// valid for the request that earned it: any newer load or selection drops it
// before it can touch the model.
func TestNewerGenerationSupersedesAScheduledRetry(t *testing.T) {
	m, _ := threadModel(t)
	read := &countingRead{msgs: []tea.Msg{threadListLoadedMsg{gen: 1, err: errPlannerWindow}}}
	request := readRequest{surface: readThreadList, gen: 1}
	m.threads.listGen = 1
	conflict := withReadConflictRetry(request, read.cmd())().(readConflictMsg)

	// A newer request (a watcher reload, say) starts while the conflict is in the
	// queue behind it.
	m.threads.listGen++

	tm, cmd := m.Update(conflict)
	m = tm.(Model)
	if cmd != nil {
		t.Error("a superseded conflict must not schedule a retry")
	}
	tm, cmd = m.Update(readRetryMsg{request: request, retry: conflict.retry})
	m = tm.(Model)
	if cmd != nil {
		t.Error("a superseded retry must not run its read")
	}
	if read.calls != 1 {
		t.Errorf("the superseded read ran %d times, want 1 (the original attempt)", read.calls)
	}

	// The same rule holds for a detail whose selection moved on.
	m.threads.detailGen, m.threads.detailRef = 3, "delivery"
	moved := readRequest{surface: readThreadDetail, id: "other", gen: 3}
	if m.readRequestCurrent(moved) {
		t.Error("a retry for a Thread that is no longer selected must be dropped")
	}
	if !m.readRequestCurrent(readRequest{surface: readThreadDetail, id: "delivery", gen: 3}) {
		t.Error("the selected Thread's own retry should still be current")
	}
}

// TestReadRequestCurrentEnumeratesEverySurface pins the shared stale-request
// chokepoint. The count sentinel makes a newly-added surface fail this test until
// its current and superseded identities are both defined.
func TestReadRequestCurrentEnumeratesEverySurface(t *testing.T) {
	m := loaded(t, 120, 40)
	m.threads.listGen = 7
	m.threads.detailGen, m.threads.detailRef = 9, "delivery"
	id := m.selectedID()

	tests := []struct {
		name    string
		current readRequest
		stale   readRequest
	}{
		{"entity list",
			readRequest{surface: readEntityList, kind: m.cur().kind, gen: m.cur().loadGen},
			readRequest{surface: readEntityList, kind: m.cur().kind, gen: m.cur().loadGen + 1}},
		{"entity detail",
			readRequest{surface: readEntityDetail, kind: m.cur().kind, id: id, gen: m.detailGen},
			readRequest{surface: readEntityDetail, kind: m.cur().kind, id: "another-selection", gen: m.detailGen}},
		{"dashboard",
			readRequest{surface: readDashboard, gen: m.dash.loadGen},
			readRequest{surface: readDashboard, gen: m.dash.loadGen + 1}},
		{"Thread list",
			readRequest{surface: readThreadList, gen: m.threads.listGen},
			readRequest{surface: readThreadList, gen: m.threads.listGen + 1}},
		{"Thread detail",
			readRequest{surface: readThreadDetail, id: m.threads.detailRef, gen: m.threads.detailGen},
			readRequest{surface: readThreadDetail, id: "another-thread", gen: m.threads.detailGen}},
	}
	if len(tests) != int(readSurfaceCount) {
		t.Fatalf("tested %d read surfaces, want %d", len(tests), readSurfaceCount)
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !m.readRequestCurrent(tt.current) {
				t.Errorf("current request was rejected: %+v", tt.current)
			}
			if m.readRequestCurrent(tt.stale) {
				t.Errorf("superseded request was accepted: %+v", tt.stale)
			}
		})
	}
	if m.readRequestCurrent(readRequest{surface: readSurfaceCount}) {
		t.Error("an unknown read surface must fail closed")
	}
}

func TestReadPolicyResultTypesOptInBesideTheirDeclarations(t *testing.T) {
	for name, msg := range map[string]tea.Msg{
		"entity list":   errMsg{err: errPlannerWindow},
		"entity detail": detailErrMsg{err: errPlannerWindow},
		"dashboard":     dashLoadedMsg{err: errPlannerWindow},
		"Thread list":   threadListLoadedMsg{err: errPlannerWindow},
		"Thread detail": threadDetailLoadedMsg{err: errPlannerWindow},
	} {
		t.Run(name, func(t *testing.T) {
			if !errors.Is(readMessageError(msg), domain.ErrConflict) {
				t.Errorf("%T did not opt into the shared read policy", msg)
			}
		})
	}
	if err := readMessageError(listLoadedMsg{}); err != nil {
		t.Errorf("a success-only result unexpectedly carries an error: %v", err)
	}
}

// TestNonConflictFailuresSkipTheRetryPolicy pins AC6's other half: only the
// documented contention class is transient. Everything else is a durable error
// on its first appearance, and nothing re-runs behind the user's back.
func TestNonConflictFailuresSkipTheRetryPolicy(t *testing.T) {
	m, _ := threadModel(t)
	m.threads.listGen = 1
	for _, err := range []error{domain.ErrNotFound, domain.ErrValidation, errors.New("disk fell over")} {
		read := &countingRead{msgs: []tea.Msg{threadListLoadedMsg{gen: 1, err: err}}}
		msg := withReadConflictRetry(readRequest{surface: readThreadList, gen: 1}, read.cmd())()
		if _, held := msg.(readConflictMsg); held {
			t.Fatalf("%v was treated as transient contention", err)
		}
		tm, cmd := m.Update(msg)
		m = tm.(Model)
		if !errors.Is(m.threads.loadErr, err) {
			t.Errorf("durable error %v was not stored, got %v", err, m.threads.loadErr)
		}
		if cmd != nil {
			t.Errorf("durable error %v scheduled follow-up work", err)
		}
		if read.calls != 1 {
			t.Errorf("durable error %v was re-read %d times", err, read.calls)
		}
	}
	// Recovery is the ordinary path: the next successful read clears it.
	m = drain(t, m, m.requestThreadList())
	if m.threads.loadErr != nil || !m.threads.loaded {
		t.Errorf("a later successful reload should clear the error: err=%v loaded=%v", m.threads.loadErr, m.threads.loaded)
	}
}

// holdPlannerWindow opens the real repository mutation planner on root and
// blocks inside its callback, so every Store read against that tree returns the
// genuine domain.ErrConflict this policy exists for.
func startPlannerWindow(root string) (release func() error, err error) {
	inside, resume, done := make(chan struct{}), make(chan struct{}), make(chan struct{})
	var mutateErr error
	go func() {
		defer close(done)
		_, mutateErr = store.NewFS(root).MutateTaskGraph(time.Now(), true,
			func(*core.TaskGraph) (core.TaskGraphMutationPlan, error) {
				close(inside)
				<-resume
				return core.TaskGraphMutationPlan{}, nil
			})
	}()
	select {
	case <-inside:
	case <-done:
		return nil, mutateErr
	}
	var released bool
	release = func() error {
		if released {
			return mutateErr
		}
		released = true
		close(resume)
		<-done
		return mutateErr
	}
	return release, nil
}

func holdPlannerWindow(t *testing.T, root string) (release func()) {
	t.Helper()
	releaseWindow, err := startPlannerWindow(root)
	if err != nil {
		t.Fatalf("planner window never opened: %v", err)
	}
	release = func() {
		if err := releaseWindow(); err != nil {
			t.Errorf("planner window mutation failed: %v", err)
		}
	}
	t.Cleanup(release)
	return release
}

func TestPlannerWindowSetupFailureReturnsInsteadOfHanging(t *testing.T) {
	r := testutil.NewRepo(t)
	r.Task("next-up", "broken.md", fmt.Sprintf(
		"---\nid: %s\nstatus: next-up\ndescription: broken graph\ndepends_on: [%s]\n---\n# broken\n",
		testutil.TaskID("broken"), testutil.TaskID("missing")))
	if release, err := startPlannerWindow(r.Root); err == nil {
		if release != nil {
			_ = release()
		}
		t.Fatal("a broken graph unexpectedly opened the planner callback")
	}
}

// TestPlannerWindowRetainsEveryLoadedSurface is the shared-policy test: with a
// real guarded mutation in flight, one reload conflicts on EVERY surface at
// once. No surface may blank, become a false empty state, or spin — and each
// converges on its own retry once the window closes.
func TestPlannerWindowRetainsEveryLoadedSurface(t *testing.T) {
	root := threadRepo(t)
	m := New(core.NewService(store.NewFS(root)))
	tm, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	m = tm.(Model)
	m = drainNested(t, m, m.Init()) // dashboard
	tm, cmd := m.Update(press("]")) // tasks tab + its detail
	m = drainNested(t, tm.(Model), cmd)
	m = drainNested(t, m, m.requestThreadList())
	m = drainNested(t, m, m.requestThreadDetail("delivery"))
	if !m.dash.loaded || !m.cur().loaded || !m.threads.loaded || !m.threads.detailOK || !m.detail.hasContent {
		t.Fatalf("setup: every surface should be loaded: dash=%v tab=%v threads=%v detail=%v pane=%v",
			m.dash.loaded, m.cur().loaded, m.threads.loaded, m.threads.detailOK, m.detail.hasContent)
	}
	rows := len(m.cur().list.Items())
	body := m.detail.content.rawBody()

	release := holdPlannerWindow(t, root)
	before := readGenerations(m)

	// One watcher reload while the planner holds the tree.
	var retries []readRetryMsg
	m, retries = drainConflicts(t, m, m.reloadAll())
	if len(retries) != 4 {
		t.Fatalf("expected every loaded surface to hold its conflict, got %d retries", len(retries))
	}
	if !m.dash.loaded || m.dash.loadErr != nil {
		t.Errorf("the dashboard blanked or surfaced a first conflict: loaded=%v err=%v", m.dash.loaded, m.dash.loadErr)
	}
	if len(m.cur().list.Items()) != rows || m.cur().loadErr != nil {
		t.Errorf("the tab lost its rows: %d of %d, err=%v", len(m.cur().list.Items()), rows, m.cur().loadErr)
	}
	if m.detail.content == nil || m.detail.content.rawBody() != body {
		t.Error("the detail pane lost the body it was showing")
	}
	if !m.threads.loaded || m.threads.loadErr != nil || len(m.threads.list.Threads) != 1 {
		t.Errorf("the Thread list blanked: loaded=%v err=%v", m.threads.loaded, m.threads.loadErr)
	}
	if !m.threads.detailOK || m.threads.detailErr != nil {
		t.Errorf("the Thread detail blanked: ok=%v err=%v", m.threads.detailOK, m.threads.detailErr)
	}

	// Each surface reloaded once for the one event, and no surface's failure
	// dragged another into a second read.
	during := readGenerations(m)
	for i, gen := range during {
		if gen != before[i]+1 {
			t.Errorf("surface %d ran %d loads for one reload, want exactly 1", i, gen-before[i])
		}
	}

	// The entity detail is the one surface a list reload does not re-read on its
	// own, so drive it directly: a selection re-read inside the window must keep
	// the body it is showing rather than replacing it with an error pane.
	m, detailRetries := drainConflicts(t, m, m.loadDetail(m.selectedID()))
	if len(detailRetries) != 1 {
		t.Fatalf("a contended detail read should hold one conflict, got %d", len(detailRetries))
	}
	if m.detail.content == nil || m.detail.content.rawBody() != body || m.detail.errMsg != "" {
		t.Errorf("a contended detail re-read blanked the pane: err=%q", m.detail.errMsg)
	}
	retries = append(retries, detailRetries...)

	release()
	for _, retry := range retries {
		tm, cmd := m.Update(retry)
		m = drainNested(t, tm.(Model), cmd)
	}
	// A retry re-runs its own request rather than opening a new one, so no
	// independent surface may have advanced again while the others recovered.
	if after := readGenerations(m); after != during {
		t.Errorf("retries opened new reads on other surfaces: %v, want %v", after, during)
	}
	if m.cur().loadErr != nil || m.dash.loadErr != nil || m.threads.loadErr != nil || m.threads.detailErr != nil {
		t.Errorf("retries after the window closed should all succeed: tab=%v dash=%v threads=%v/%v",
			m.cur().loadErr, m.dash.loadErr, m.threads.loadErr, m.threads.detailErr)
	}
	if len(m.cur().list.Items()) != rows || m.detail.errMsg != "" {
		t.Errorf("the converged model = %d rows, detail err %q", len(m.cur().list.Items()), m.detail.errMsg)
	}
}

// TestFirstLoadContentionIsNotAFalseEmptyState pins the landing case: a
// contended FIRST load has nothing to retain, so it must stay visibly loading
// rather than resolving into an empty dashboard — and the screen you are looking
// at must still be recoverable by the next watcher reload.
func TestFirstLoadContentionIsNotAFalseEmptyState(t *testing.T) {
	root := threadRepo(t)
	m := New(core.NewService(store.NewFS(root)))
	tm, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	m = tm.(Model)

	release := holdPlannerWindow(t, root)
	m, retries := drainConflicts(t, m, m.Init())
	if len(retries) != 1 {
		t.Fatalf("the landing load should hold exactly one conflict, got %d", len(retries))
	}
	if m.dash.loaded || m.dash.loadErr != nil {
		t.Fatalf("a first conflict must not resolve the dashboard: loaded=%v err=%v", m.dash.loaded, m.dash.loadErr)
	}
	if body := ansi.Strip(m.View().Content); !strings.Contains(body, "loading") {
		t.Errorf("a contended first load should still read as loading:\n%s", body)
	}

	// A second conflict at the retry bound is durable, not a spin — and it is
	// still not a false empty state, because nothing was ever loaded.
	tm, cmd := m.Update(retries[0])
	m = drainNested(t, tm.(Model), cmd)
	if m.dash.loaded || !errors.Is(m.dash.loadErr, domain.ErrConflict) {
		t.Fatalf("a repeated first-load conflict should be visible: loaded=%v err=%v", m.dash.loaded, m.dash.loadErr)
	}

	// Recovery: the window closes and the next watcher reload settles the screen
	// the user is looking at, even though it never completed a load.
	release()
	m = drainNested(t, m, m.reloadAll())
	if !m.dash.loaded || m.dash.loadErr != nil {
		t.Fatalf("a watcher reload should recover the landing screen: loaded=%v err=%v", m.dash.loaded, m.dash.loadErr)
	}
}

// readGenerations snapshots every INDEPENDENTLY reloadable surface's request
// counter: the tasks tab, the dashboard, and both Thread caches. The entity
// detail is deliberately absent — it is a derived read, chained off its list's
// result, and is exercised on its own below.
func readGenerations(m Model) [4]int {
	return [4]int{m.cur().loadGen, m.dash.loadGen, m.threads.listGen, m.threads.detailGen}
}

// drainConflicts resolves a reload's command tree, collecting the retry each
// held conflict schedules instead of waiting out its timer.
func drainConflicts(t *testing.T, m Model, cmd tea.Cmd) (Model, []readRetryMsg) {
	t.Helper()
	var retries []readRetryMsg
	if cmd == nil {
		return m, nil
	}
	msg := cmd()
	if msg == nil {
		return m, nil
	}
	if batch, ok := msg.(tea.BatchMsg); ok {
		for _, child := range batch {
			var found []readRetryMsg
			m, found = drainConflicts(t, m, child)
			retries = append(retries, found...)
		}
		return m, retries
	}
	if conflict, ok := msg.(readConflictMsg); ok {
		tm, scheduled := m.Update(conflict)
		if scheduled == nil {
			t.Errorf("a conflict on a current request scheduled no retry: %+v", conflict.request)
		}
		return tm.(Model), append(retries, readRetryMsg(conflict))
	}
	tm, next := m.Update(msg)
	m = tm.(Model)
	if next != nil {
		var found []readRetryMsg
		m, found = drainConflicts(t, m, next)
		retries = append(retries, found...)
	}
	return m, retries
}
