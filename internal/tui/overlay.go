package tui

import tea "github.com/charmbracelet/bubbletea"

// modal is a floating overlay layer (the `?` help panel, the `a` action menu, the
// `f` follow picker): while active it owns every key and floats a box over the
// body. The reducer loops one ordered registry of these (Model.modals) instead of
// an if-chain of `m.X.active` guards plus a parallel bodyView switch, so a NEW
// overlay is one struct + one entry in defaultModals — no new handleKey guard
// block, no new bodyView case, and no re-implemented ForceQuit (that's handled
// once in the handleKey preamble, ahead of the modal loop).
//
// Overlay STATE lives in the Model (value-copied each Update), so these markers are
// stateless and act on the model through *Model — preserving the "value model,
// mutate-the-copy" idiom rather than introducing shared-pointer overlay state.
// handleKey returns handled=false to decline a key (fall through to base routing).
type modal interface {
	active(m *Model) bool
	handleKey(m *Model, msg tea.KeyMsg) (handled bool, cmd tea.Cmd)
	view(m *Model, w, h int) string
}

// defaultModals is the overlay registry in precedence order — help, then the
// action menu, then the follow picker (the order the old guard chain used). The
// first active modal owns the key and the floated box.
func defaultModals() []modal {
	return []modal{helpModal{}, actionModal{}, followModal{}}
}

// helpModal is the `?` keybinding overlay: j/k scroll it (the content can outgrow
// a short terminal); any other key dismisses it.
type helpModal struct{}

func (helpModal) active(m *Model) bool { return m.showHelp }

func (helpModal) handleKey(m *Model, msg tea.KeyMsg) (bool, tea.Cmd) {
	switch msg.String() {
	case "j", "down":
		if m.helpScroll < m.helpMaxScroll() {
			m.helpScroll++
		}
	case "k", "up":
		if m.helpScroll > 0 {
			m.helpScroll--
		}
	default:
		m.showHelp = false
		m.helpScroll = 0
	}
	return true, nil
}

func (helpModal) view(m *Model, w, h int) string { return helpBox(w, h, m.helpScroll) }

// actionModal is the `a` lifecycle action menu: vim-select a transition, Enter
// applies it, a destructive choice gates on y/n.
type actionModal struct{}

func (actionModal) active(m *Model) bool { return m.action.active }

func (actionModal) handleKey(m *Model, msg tea.KeyMsg) (bool, tea.Cmd) {
	return true, m.handleActionKey(msg)
}

func (actionModal) view(m *Model, w, h int) string { return m.action.view(w, h) }

// followModal is the `f` reference picker (an epic → its tasks).
type followModal struct{}

func (followModal) active(m *Model) bool { return m.follow.active }

func (followModal) handleKey(m *Model, msg tea.KeyMsg) (bool, tea.Cmd) {
	return true, m.handleFollowKey(msg)
}

func (followModal) view(m *Model, w, h int) string { return m.follow.view(w, h) }
