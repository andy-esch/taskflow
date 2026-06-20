package prompt

import (
	"errors"
	"io"

	"github.com/charmbracelet/huh"
)

// ttyPrompter is the real-terminal Prompter: list pickers use bubbles/list (the
// same component the TUI browses with — see picker.go) where its filter/viewport
// are battle-tested, and text inputs use huh (where there's no list and so no
// filter quirk). Everything renders to out (stderr), so stdout stays a clean data
// stream.
type ttyPrompter struct {
	in  io.Reader
	out io.Writer
}

// NewTTY returns the interactive Prompter over the given input and output (stderr).
func NewTTY(in io.Reader, out io.Writer) Prompter {
	return ttyPrompter{in: in, out: out}
}

// SelectOne picks one option from a fuzzy-filterable list (bubbles/list).
func (p ttyPrompter) SelectOne(title string, opts []Option) (string, error) {
	return runPicker(p.in, p.out, title, opts)
}

// Text reads a single line via huh (where it has no list-filter quirks), themed
// to match the rest of the CLI.
func (p ttyPrompter) Text(title, placeholder string) (string, error) {
	var v string
	field := huh.NewInput().Title(title).Placeholder(placeholder).Value(&v)
	err := huh.NewForm(huh.NewGroup(field)).
		WithInput(p.in).
		WithOutput(p.out).
		WithTheme(huh.ThemeDracula()).
		Run()
	if errors.Is(err, huh.ErrUserAborted) {
		return "", ErrAborted
	}
	if err != nil {
		return "", err
	}
	return v, nil
}
