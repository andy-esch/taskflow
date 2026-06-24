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

var keys = keyMap{
	Right:       key.NewBinding(key.WithKeys("l", "enter")),
	Left:        key.NewBinding(key.WithKeys("h")),
	Back:        key.NewBinding(key.WithKeys("esc")),
	Top:         key.NewBinding(key.WithKeys("g")),
	Bottom:      key.NewBinding(key.WithKeys("G")),
	Find:        key.NewBinding(key.WithKeys("/")),
	FindNext:    key.NewBinding(key.WithKeys("n")),
	FindPrev:    key.NewBinding(key.WithKeys("N")),
	ToggleFocus: key.NewBinding(key.WithKeys("tab")),
	Command:     key.NewBinding(key.WithKeys(":")),
	Palette:     key.NewBinding(key.WithKeys("ctrl+p")),
	PrevTab:     key.NewBinding(key.WithKeys("[")),
	NextTab:     key.NewBinding(key.WithKeys("]")),
	Sort:        key.NewBinding(key.WithKeys("o")),
	SortRev:     key.NewBinding(key.WithKeys("O")),
	StatusView:  key.NewBinding(key.WithKeys("s")),
	StatusRev:   key.NewBinding(key.WithKeys("S")),
	FilterMode:  key.NewBinding(key.WithKeys("F"), key.WithHelp("F", "filter mode")),
	Action:      key.NewBinding(key.WithKeys("m")),
	Edit:        key.NewBinding(key.WithKeys("e")),
	RawToggle:   key.NewBinding(key.WithKeys("R")),
	Follow:      key.NewBinding(key.WithKeys("f")),
	JumpBack:    key.NewBinding(key.WithKeys("ctrl+o")),
	Yank:        key.NewBinding(key.WithKeys("y")),
	YankPath:    key.NewBinding(key.WithKeys("Y")),
	Help:        key.NewBinding(key.WithKeys("?")),
	Refresh:     key.NewBinding(key.WithKeys("r")),
	Quit:        key.NewBinding(key.WithKeys("q")),
	ForceQuit:   key.NewBinding(key.WithKeys("ctrl+c")),
}
