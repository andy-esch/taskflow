package tui

import (
	"strings"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"

	"github.com/andy-esch/taskflow/internal/theme"
)

// commandBar is the `:` command-jump input (k9s-style). It captures keys while
// active (the model gates global hotkeys on c.active, like the list's filter),
// completes entity names on Tab, and reports its value to the model on Enter.
type commandBar struct {
	ti     textinput.Model
	active bool
	err    string // transient message shown after an unknown command
}

func newCommandBar() commandBar {
	ti := textinput.New()
	ti.Prompt = ":"
	ti.CharLimit = 32
	ti.Placeholder = "tasks·epics·audits (t/e/a)"
	return commandBar{ti: ti}
}

// focus opens the command bar with an empty value.
func (c *commandBar) focus() tea.Cmd {
	c.active = true
	c.err = ""
	c.ti.SetValue("")
	return c.ti.Focus()
}

func (c *commandBar) blur() {
	c.active = false
	c.ti.Blur()
}

func (c *commandBar) update(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	c.ti, cmd = c.ti.Update(msg)
	return cmd
}

// value is the trimmed command word (without the leading ':').
func (c commandBar) value() string { return strings.TrimSpace(c.ti.Value()) }

// complete advances the input to the longest unique prefix among options that
// start with the current value (Tab-completion). With a single match it fills the
// whole name.
func (c *commandBar) complete(options []string) {
	cur := c.value()
	var matches []string
	for _, o := range options {
		if strings.HasPrefix(o, cur) {
			matches = append(matches, o)
		}
	}
	if len(matches) == 0 {
		return
	}
	lcp := matches[0]
	for _, m := range matches[1:] {
		lcp = commonPrefix(lcp, m)
	}
	c.ti.SetValue(lcp)
	c.ti.CursorEnd()
}

func (c commandBar) view() string {
	if c.err != "" {
		return c.ti.View() + "  " + fg(theme.ColorRed, c.err)
	}
	return c.ti.View()
}

func commonPrefix(a, b string) string {
	n := min(len(a), len(b))
	i := 0
	for i < n && a[i] == b[i] {
		i++
	}
	return a[:i]
}
