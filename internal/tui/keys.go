package tui

import "charm.land/bubbles/v2/key"

// keyMap holds the bindings the root model matches itself. List/viewport
// navigation (j/k plus paging: d/u pages the list, ctrl+d/u half-pages the
// detail viewport) is handled by those sub-components, so it's not here — only
// the keys that change focus, switch entity, or change app state.
type keyMap struct {
	Right       key.Binding // l / enter → detail
	Left        key.Binding // h → back to list
	Back        key.Binding // esc → back to list
	Top         key.Binding // g (detail: scroll to top; the list binds it itself)
	Bottom      key.Binding // G (detail: scroll to bottom)
	Find        key.Binding // / (detail: find in body)
	FindNext    key.Binding // n (detail: next match)
	FindPrev    key.Binding // N (detail: previous match)
	ToggleFocus key.Binding // tab
	Zoom        key.Binding // z → full-screen the detail pane (toggle)
	Command     key.Binding // : → entity command-jump
	Palette     key.Binding // ctrl+p → fuzzy command palette (jump to anything / run a command)
	PrevTab     key.Binding // [ → previous entity tab
	NextTab     key.Binding // ] → next entity tab
	Sort        key.Binding // o → cycle sort column
	SortRev     key.Binding // O → toggle sort direction
	StatusView  key.Binding // s → cycle view (task status / audit bucket)
	StatusRev   key.Binding // S → cycle view backward
	FilterMode  key.Binding // F → toggle list filter: fuzzy ⇄ substring
	Action      key.Binding // m → lifecycle move menu (mirrors the CLI `move`/start/complete/…)
	Edit        key.Binding // e → inline field editor (tasks)
	OpenEditor  key.Binding // E → open the selection's whole file in $EDITOR (any entity)
	RawToggle   key.Binding // R → raw ⇄ pretty markdown in the detail body
	Follow      key.Binding // f → follow the selection's reference (task ⇄ epic)
	JumpBack    key.Binding // ctrl+o → pop the follow back-stack (vim jumplist)
	Yank        key.Binding // y → copy the selection's slug/id to the clipboard
	YankPath    key.Binding // Y → copy the selection's file path to the clipboard
	Help        key.Binding // ? → toggle the keybinding overlay
	Refresh     key.Binding // r
	Quit        key.Binding // q (context)
	ForceQuit   key.Binding // ctrl+c
}

// Every binding carries WithHelp(displayKey, desc): the keyMap is the SINGLE source
// of each key's display string AND its `?`-overlay description, so the help overlay
// (help.go) and the footer hints (model.go) derive from here and can't drift. The
// match keys (WithKeys) can differ from the display key — e.g. Right matches l/enter
// but shows "l / ⏎".
var keys = keyMap{
	Right:       key.NewBinding(key.WithKeys("l", "enter"), key.WithHelp("l / ⏎", "open detail")),
	Left:        key.NewBinding(key.WithKeys("h"), key.WithHelp("h", "back to list")),
	Back:        key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "back (clears a find first)")),
	Top:         key.NewBinding(key.WithKeys("g"), key.WithHelp("g", "scroll to top")),
	Bottom:      key.NewBinding(key.WithKeys("G"), key.WithHelp("G", "scroll to bottom")),
	Find:        key.NewBinding(key.WithKeys("/"), key.WithHelp("/", "find in body")),
	FindNext:    key.NewBinding(key.WithKeys("n"), key.WithHelp("n", "next match")),
	FindPrev:    key.NewBinding(key.WithKeys("N"), key.WithHelp("N", "previous match")),
	ToggleFocus: key.NewBinding(key.WithKeys("tab"), key.WithHelp("tab", "switch focus (list ⇄ detail)")),
	Zoom:        key.NewBinding(key.WithKeys("z"), key.WithHelp("z", "full-screen the detail pane (z/esc to exit)")),
	Command:     key.NewBinding(key.WithKeys(":"), key.WithHelp(":", "command / jump (entity, status, or verb)")),
	Palette:     key.NewBinding(key.WithKeys("ctrl+p"), key.WithHelp("ctrl+p", "command palette — fuzzy jump to anything / run a command")),
	PrevTab:     key.NewBinding(key.WithKeys("["), key.WithHelp("[", "previous tab")),
	NextTab:     key.NewBinding(key.WithKeys("]"), key.WithHelp("]", "next tab")),
	Sort:        key.NewBinding(key.WithKeys("o"), key.WithHelp("o", "cycle sort column")),
	SortRev:     key.NewBinding(key.WithKeys("O"), key.WithHelp("O", "reverse sort direction")),
	StatusView:  key.NewBinding(key.WithKeys("s"), key.WithHelp("s", "cycle view (task status / audit bucket)")),
	StatusRev:   key.NewBinding(key.WithKeys("S"), key.WithHelp("S", "cycle view backward")),
	FilterMode:  key.NewBinding(key.WithKeys("F"), key.WithHelp("F", "filter mode: fuzzy ⇄ substring (default fuzzy)")),
	Action:      key.NewBinding(key.WithKeys("m"), key.WithHelp("m", "move — lifecycle (start/complete/defer/…); audits: close/reopen/defer; epics: activate/retire/deprecate")),
	Edit:        key.NewBinding(key.WithKeys("e"), key.WithHelp("e", "edit fields in place — tasks: desc/priority/tags/effort/tier (+revisit when deferred) · epics: desc/priority/tags")),
	OpenEditor:  key.NewBinding(key.WithKeys("E"), key.WithHelp("E", "open the whole file in $EDITOR (any entity; re-read on save)")),
	RawToggle:   key.NewBinding(key.WithKeys("R"), key.WithHelp("R", "raw ⇄ pretty markdown")),
	Follow:      key.NewBinding(key.WithKeys("f"), key.WithHelp("f", "follow reference (task ⇄ epic)")),
	JumpBack:    key.NewBinding(key.WithKeys("ctrl+o"), key.WithHelp("ctrl+o", "jump back (follow history)")),
	Yank:        key.NewBinding(key.WithKeys("y"), key.WithHelp("y", "copy slug to clipboard")),
	YankPath:    key.NewBinding(key.WithKeys("Y"), key.WithHelp("Y", "copy file path to clipboard")),
	Help:        key.NewBinding(key.WithKeys("?"), key.WithHelp("?", "toggle this help (esc to close)")),
	Refresh:     key.NewBinding(key.WithKeys("r"), key.WithHelp("r", "refresh from disk")),
	Quit:        key.NewBinding(key.WithKeys("q"), key.WithHelp("q", "quit")),
	ForceQuit:   key.NewBinding(key.WithKeys("ctrl+c"), key.WithHelp("ctrl+c", "force-quit")),
}
