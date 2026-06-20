package prompt

import (
	"errors"
	"io"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/andy-esch/taskflow/internal/listfilter"
)

// listItem adapts an Option to bubbles/list.Item. FilterValue is the label, so
// the fuzzy `/` filter matches on the visible text — the same component and
// behavior the TUI list uses, which is why filtering/scrolling work correctly
// here where huh.Select's own filter did not.
type listItem struct{ label, value string }

func (i listItem) Title() string       { return i.label }
func (i listItem) Description() string { return "" }
func (i listItem) FilterValue() string { return i.label }

// pickerModel is a one-shot list picker built on bubbles/list — run to completion
// like a huh form, returning the chosen value (or an abort).
type pickerModel struct {
	list    list.Model
	choice  string
	aborted bool
}

func newPicker(title string, opts []Option) pickerModel {
	items := make([]list.Item, len(opts))
	for i, o := range opts {
		items[i] = listItem{label: o.Label, value: o.Value}
	}
	d := list.NewDefaultDelegate()
	d.ShowDescription = false // one compact line per option (bubbles truncates to width)
	d.SetSpacing(0)
	l := list.New(items, d, 0, 0)
	l.Title = title
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(true)
	// Substring, not the default fuzzy — predictable for slug identifiers; the
	// shared matcher the TUI also offers via its filter-mode toggle.
	l.Filter = listfilter.Substring
	return pickerModel{list: l}
}

func (m pickerModel) Init() tea.Cmd { return nil }

func (m pickerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.list.SetSize(msg.Width, msg.Height)
		return m, nil
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			m.aborted = true
			return m, tea.Quit
		}
		// While the filter input is focused, the list owns every key.
		if m.list.FilterState() != list.Filtering {
			switch msg.String() {
			case "enter":
				if it, ok := m.list.SelectedItem().(listItem); ok {
					m.choice = it.value
				}
				return m, tea.Quit
			case "esc":
				// esc clears an applied filter (the list handles that); only abort
				// when there is nothing to clear.
				if m.list.FilterState() == list.Unfiltered {
					m.aborted = true
					return m, tea.Quit
				}
			}
		}
	}
	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m pickerModel) View() string { return m.list.View() }

// runPicker runs the picker to completion on the given TTY (in / out=stderr) and
// returns the chosen value, or ErrAborted if the user cancelled. The alt screen
// keeps the picker from scrolling the surrounding output and restores the terminal
// cleanly on exit (matching the TUI's full-screen feel).
func runPicker(in io.Reader, out io.Writer, title string, opts []Option) (string, error) {
	final, err := tea.NewProgram(
		newPicker(title, opts),
		tea.WithInput(in),
		tea.WithOutput(out),
		tea.WithAltScreen(),
	).Run()
	if err := pickerErr(err); err != nil {
		return "", err
	}
	// Reached only when err == nil, so final is always our model; the comma-ok is
	// defensive against a future bubbletea change.
	if m, ok := final.(pickerModel); ok && !m.aborted {
		return m.choice, nil
	}
	return "", ErrAborted
}

// pickerErr normalizes bubbletea's terminal sentinels: a SIGINT *signal* (vs the
// ctrl-c key, which the model handles) surfaces as ErrInterrupted — map it to
// ErrAborted so it still exits 130 with a quiet "aborted" rather than a raw exit
// 1. Other errors (incl. ErrProgramKilled from a real panic) pass through so they
// stay visible.
func pickerErr(err error) error {
	if errors.Is(err, tea.ErrInterrupted) {
		return ErrAborted
	}
	return err
}
