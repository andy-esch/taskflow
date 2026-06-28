// command_dispatch.go: the `:` command bar + ctrl+p palette half — word dispatch,
// completion options, and the palette index/keys.
package tui

import (
	"strings"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
)

// dispatchCommand resolves the typed `:` word to an entity tab or a task status
// view and applies it; an unknown word reopens the bar with an inline error.
// isDashboardWord reports whether a `:`-command word selects the landing
// dashboard (`:dashboard`, or the `:d` shorthand).
func isDashboardWord(word string) bool { return word == "dashboard" || word == "d" }

// commandHint lists the `:` commands matching what's typed so far (all of them on
// an empty prompt) — inline discovery of the command vocabulary, narrowing as you
// type. Uses the canonical command set (no short t/e/a aliases, to stay readable);
// the footer dims and width-truncates it.
func (m Model) commandHint() string {
	cur := m.cmd.value()
	var matches []string
	for _, w := range m.paletteCommands() {
		if strings.HasPrefix(w, cur) {
			matches = append(matches, w)
		}
	}
	return strings.Join(matches, " · ")
}

func (m Model) dispatchCommand() (tea.Model, tea.Cmd) {
	word := m.cmd.value()
	m.cmd.blur()
	if word == "" {
		return m, nil
	}
	if isDashboardWord(word) {
		return m, m.enterDash()
	}
	for i, t := range m.tabs {
		if t.matches(word) {
			return m, m.switchTab(i)
		}
	}
	if view, i, ok := m.resolveView(word); ok {
		return m, m.applyView(i, view)
	}
	if tr, ok := transitionFor(m.cur().transitions, word); ok {
		id, _, ok := m.selectedLifecycle()
		if !ok {
			cmd := m.cmd.focus()
			m.cmd.err = "select a row first"
			return m, cmd
		}
		return m, m.beginTransition(id, tr) // confirm/revisit/apply per the verb
	}
	cmd := m.cmd.focus()
	m.cmd.err = "unknown: " + word
	return m, cmd
}

func (m Model) entityNames() []string {
	names := make([]string, len(m.tabs))
	for i, t := range m.tabs {
		names[i] = t.name
	}
	return names
}

// commandOptions is the `:` Tab-completion set: every tab's names + aliases +
// view-axis words (so any tab is reachable by name), plus the ACTIVE tab's
// lifecycle verbs. Verbs are context-scoped to match dispatchCommand, which
// resolves them against m.cur().transitions — so completion never offers a verb
// that would be "unknown" on the tab in view. View words are deduped because tasks
// and audits share "deferred"/"all".
func (m Model) commandOptions() []string {
	opts := make([]string, 0, len(m.tabs)*2+len(m.cur().transitions)+12)
	seen := make(map[string]bool)
	add := func(w string) {
		if !seen[w] {
			seen[w] = true
			opts = append(opts, w)
		}
	}
	for _, t := range m.tabs {
		add(t.name)
		for _, a := range t.aliases {
			add(a)
		}
		for _, w := range t.viewWords() {
			add(w)
		}
	}
	for _, tr := range m.cur().transitions {
		add(tr.verb)
	}
	add("dashboard")
	add("d")
	return opts
}

// openPalette builds the candidate index from the loaded tabs, kicks off loads for
// any tab not yet loaded (the index fills in via the reindex hook in
// handleListLoaded while the palette is open), sizes the overlay, and focuses the
// query.
func (m *Model) openPalette() tea.Cmd {
	var loads []tea.Cmd
	for _, t := range m.tabs {
		if !t.loaded {
			loads = append(loads, t.reload(m.svc, ""))
		}
	}
	w := min(max(m.width-8, 28), 64)
	h := min(max(m.paneOuterH-4, 4), 16)
	cmds := append(loads, m.palette.open(m.paletteIndex(), w, h))
	return tea.Batch(cmds...)
}

// paletteIndex is the flat candidate set: every loaded entity (jump to it) plus
// every `:` command word (dispatch it).
func (m Model) paletteIndex() []paletteItem {
	var items []paletteItem
	for _, t := range m.tabs {
		for _, it := range t.list.Items() {
			ei, ok := it.(entityItem)
			if !ok {
				continue
			}
			items = append(items, paletteItem{
				kind: palJump, ek: t.kind, id: ei.id(),
				title: ei.id(), filter: ei.id() + " " + t.name,
			})
		}
	}
	for _, w := range m.paletteCommands() {
		items = append(items, paletteItem{kind: palCommand, word: w, title: ":" + w, filter: w})
	}
	return items
}

// paletteCommands is the canonical command words for the palette: tab names, each
// tab's view words, and the active tab's verbs. Unlike commandOptions (which feeds
// `:` Tab-completion), it omits the short aliases (t/e/a, task/epic/audit) — in a
// fuzzy list they'd surface as near-duplicate rows, the exact clutter the palette
// exists to avoid.
func (m Model) paletteCommands() []string {
	seen := make(map[string]bool)
	var out []string
	add := func(w string) {
		if w != "" && !seen[w] {
			seen[w] = true
			out = append(out, w)
		}
	}
	add("dashboard") // the landing screen (omit the :d shorthand — palette avoids near-dupes)
	for _, t := range m.tabs {
		add(t.name)
		for _, w := range t.viewWords() {
			add(w)
		}
	}
	for _, tr := range m.cur().transitions {
		add(tr.verb)
	}
	return out
}

// handlePaletteKey drives the palette while open: Esc closes, Enter runs the
// selection, ↑/↓ move the cursor, anything else edits the query (re-filtering on
// each keystroke).
func (m *Model) handlePaletteKey(msg tea.KeyPressMsg) tea.Cmd {
	switch {
	case key.Matches(msg, keys.Back):
		m.palette.close()
		return nil
	case msg.String() == "enter":
		it, ok := m.palette.list.SelectedItem().(paletteItem)
		m.palette.close()
		if !ok {
			return nil
		}
		return m.runPaletteItem(it)
	case msg.String() == "up" || msg.String() == "down":
		var cmd tea.Cmd
		m.palette.list, cmd = m.palette.list.Update(msg)
		return cmd
	default:
		var cmd tea.Cmd
		m.palette.input, cmd = m.palette.input.Update(msg)
		m.palette.refilter()
		return cmd
	}
}

// runPaletteItem executes the chosen candidate: jump to an entity, or dispatch a
// `:` command word.
func (m *Model) runPaletteItem(it paletteItem) tea.Cmd {
	switch it.kind {
	case palJump:
		m.pushLoc()
		return m.jumpTo(it.ek, it.id)
	case palCommand:
		return m.runPaletteCommand(it.word)
	}
	return nil
}

// runPaletteCommand resolves a `:` word the same way dispatchCommand does (shared
// predicates: tab match → view → transition), but reports failure with a flash
// since the palette has no command-bar line to re-focus. A verb acts on the
// underlying selected row, exactly like `:`.
func (m *Model) runPaletteCommand(word string) tea.Cmd {
	if isDashboardWord(word) {
		return m.enterDash()
	}
	for i, t := range m.tabs {
		if t.matches(word) {
			return m.switchTab(i)
		}
	}
	if view, i, ok := m.resolveView(word); ok {
		return m.applyView(i, view)
	}
	if tr, ok := transitionFor(m.cur().transitions, word); ok {
		id, _, ok := m.selectedLifecycle()
		if !ok {
			m.flash, m.flashErr = "select a row first to :"+word, true
			return nil
		}
		return m.beginTransition(id, tr) // confirm/revisit/apply per the verb
	}
	return nil
}
