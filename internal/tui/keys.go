package tui

import "github.com/charmbracelet/bubbles/key"

// keyMap holds the bindings the root model matches itself. List/viewport
// navigation (j/k/g/G, ctrl+d/u) is handled by those sub-components, so it's not
// here — only the keys that change focus or app state.
type keyMap struct {
	Right       key.Binding // l / enter → detail
	Left        key.Binding // h → back to list
	Back        key.Binding // esc → back to list
	Top         key.Binding // g (detail: scroll to top; the list binds it itself)
	Bottom      key.Binding // G (detail: scroll to bottom)
	ToggleFocus key.Binding // tab
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
	ToggleFocus: key.NewBinding(key.WithKeys("tab")),
	Refresh:     key.NewBinding(key.WithKeys("r")),
	Quit:        key.NewBinding(key.WithKeys("q")),
	ForceQuit:   key.NewBinding(key.WithKeys("ctrl+c")),
}
