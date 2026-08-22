package tui

import (
	"path/filepath"
	"reflect"

	tea "charm.land/bubbletea/v2"

	"github.com/andy-esch/taskflow/internal/core"
)

// sessionMsg stamps every asynchronous command result with the workspace generation
// that launched it. Entity messages already protect tab/detail ordering inside one
// workspace; this outer stamp prevents a slow read or mutation from the previous space
// from landing after an atlas switch and acting on the new service.
type sessionMsg struct {
	gen uint64
	msg tea.Msg
}

func scopeSession(gen uint64, cmd tea.Cmd) tea.Cmd {
	if cmd == nil {
		return nil
	}
	return func() tea.Msg {
		msg := cmd()
		switch msg := msg.(type) {
		case nil:
			return nil
		case tea.BatchMsg:
			// Batch carries commands, not results: scope the children instead, or every
			// command in the batch escapes the stamp. tea.Sequence has the same shape but
			// an UNEXPORTED payload type, so it cannot be unwrapped here — this package
			// must not use tea.Sequence while session scoping is in force (guarded by
			// TestSessionScopingHasNoSequenceCommandsToLeak).
			wrapped := make([]tea.Cmd, len(msg))
			for i, child := range msg {
				wrapped[i] = scopeSession(gen, child)
			}
			return tea.BatchMsg(wrapped)
		case tea.QuitMsg, tea.SuspendMsg:
			// These are Bubble Tea control messages, not model data. Hiding them inside
			// sessionMsg would stop the runtime from recognizing them.
			return msg
		default:
			// Bubble Tea has additional unexported runtime-control messages (notably
			// ExecProcess and OSC52 clipboard writes). They must remain
			// visible to the runtime; only application/component results are scoped.
			t := reflect.TypeOf(msg)
			// A pointer type is unnamed, so its own PkgPath is "" — dereference first or a
			// pointer-shaped runtime message would be hidden inside sessionMsg.
			for t != nil && t.Kind() == reflect.Pointer {
				t = t.Elem()
			}
			if t != nil && t.PkgPath() == "charm.land/bubbletea/v2" {
				return msg
			}
			return sessionMsg{gen: gen, msg: msg}
		}
	}
}

// spaceSession is the browsing state that should survive atlas round trips. Process-wide
// chrome, modal state, and the watcher are deliberately excluded; only the active space
// owns a watcher, and transient overlays should never reopen after a switch.
type spaceSession struct {
	tabs      []*entityTab
	active    int
	onDash    bool
	dash      dashboard
	detail    detailPane
	focus     focus
	zoom      bool
	navStack  []navLoc
	detailGen int
	movedAway string
}

func workspaceKey(workspace core.Workspace) string {
	if workspace.Checkout != "" {
		return filepath.Clean(workspace.Checkout)
	}
	return filepath.Clean(workspace.PlanningRoot)
}

func (m *Model) saveSession() {
	key := workspaceKey(m.workspace)
	if key == "." || m.svc == nil {
		return
	}
	m.sessions[key] = spaceSession{
		tabs: m.tabs, active: m.active, onDash: m.onDash, dash: m.dash,
		detail: m.detail, focus: m.focus, zoom: m.zoom,
		navStack: append([]navLoc(nil), m.navStack...), detailGen: m.detailGen,
		movedAway: m.movedAway,
	}
}

// activateWorkspace atomically replaces the planning context after Open succeeded. A
// known checkout restores its UI state against the freshly-opened service; a first visit
// receives the ordinary dashboard-first session. Every loaded surface is refreshed so
// preserved cursors/filters do not imply preserved filesystem data.
func (m *Model) activateWorkspace(workspace core.Workspace, nextWatcher *watcher, watchErr error) tea.Cmd {
	m.saveSession()
	// The resume snapshot belongs to the space being left; the incoming space's own
	// focus/zoom come from its cached session (or the first-visit defaults) below.
	m.atlasResume = atlasResume{}
	oldWatcher := m.watch
	m.workspace = workspace
	m.svc = workspace.Planning
	m.configStart = workspace.Checkout
	m.watch = nextWatcher
	m.watchOff = watchErr != nil
	m.sessionGen++
	m.onAtlas = false
	m.atlas.opening = false
	m.atlas.openErr = ""
	m.closeTransientUI()

	key := workspaceKey(workspace)
	saved, restored := m.sessions[key]
	if restored {
		m.tabs, m.active, m.onDash, m.dash = saved.tabs, saved.active, saved.onDash, saved.dash
		m.detail, m.focus, m.zoom = saved.detail, saved.focus, saved.zoom
		m.navStack, m.detailGen, m.movedAway = saved.navStack, saved.detailGen, saved.movedAway
	} else {
		m.tabs = newEntityTabs(m.st)
		m.active = 0
		m.onDash = true
		m.dash = dashboard{}
		m.detail = newDetailPane(m.st)
		m.focus = focusList
		m.zoom = false
		m.navStack = nil
		m.detailGen = 0
		m.movedAway = ""
	}
	m.dirtyGen = 0
	m.recomputeLayout()

	var loads tea.Cmd
	if restored && (!m.onDash || m.dash.loaded) {
		loads = m.reloadAll()
	} else {
		loads = loadDashboard(m.svc)
	}
	return tea.Batch(closeWatcher(oldWatcher), loads, waitForFS(m.watch))
}

func (m *Model) closeTransientUI() {
	m.showHelp = false
	m.helpScroll = 0
	m.configOpen = false
	m.action.close()
	m.follow.close()
	m.edit.close()
	m.palette.close()
	m.cmd.blur()
	m.flash = ""
}

func closeWatcher(w *watcher) tea.Cmd {
	if w == nil {
		return nil
	}
	return func() tea.Msg {
		_ = w.close()
		return nil
	}
}
