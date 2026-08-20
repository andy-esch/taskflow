package prompt

import (
	"errors"
	"io"
	"os"

	"charm.land/bubbles/v2/key"
	"charm.land/huh/v2"
	"github.com/charmbracelet/x/ansi"
	"golang.org/x/term"

	"github.com/andy-esch/taskflow/internal/design"
)

// ttyPrompter is the real-terminal Prompter: both list pickers (huh.Select, rendered
// inline below the prompt with a type-to-filter input) and text inputs (huh.Input) go
// through huh, themed to match. Everything renders to out (stderr), so stdout stays a
// clean data stream.
type ttyPrompter struct {
	in    io.Reader
	out   io.Writer
	theme design.Theme
}

// NewTTY returns the interactive Prompter over the given input and output (stderr),
// skinned to theme th (the picker caret + current row use its accent).
func NewTTY(in io.Reader, out io.Writer, th design.Theme) Prompter {
	return ttyPrompter{in: in, out: out, theme: th}
}

// SelectOne picks one option from a filterable list rendered INLINE below the prompt
// (huh.Select) — a "> " caret marks the current row, type to filter. Each label is
// truncated to the terminal width so a long description stays on ONE line rather than
// wrapping. Renders to out (stderr).
// abortKeyMap is huh's default bindings with ESC added to Quit.
//
// huh binds Quit to ctrl+c ALONE, so a prompt could only be escaped with an interrupt —
// which reads as "something went wrong" for what is just changing your mind. The downstream
// handling was already correct (ErrAborted → a quiet "aborted", exit 130); only the binding
// was missing, and this package's own doc comment already promised esc.
//
// ctrl+c is kept so muscle memory still works. Where filtering is on, huh's field consumes
// esc first to clear an active filter, so esc quits once the filter is empty — the
// conventional layering, not a conflict.
func abortKeyMap() *huh.KeyMap {
	km := huh.NewDefaultKeyMap()
	km.Quit = key.NewBinding(key.WithKeys("esc", "ctrl+c"), key.WithHelp("esc", "cancel"))
	return km
}

// selectMenuMax is the longest list shown without scrolling; selectFilterMin is the
// shortest list a filter input earns its place on.
const (
	selectMenuMax   = 10
	selectFilterMin = 8
)

// selectLayout decides a Select's window height (0 = let huh size it) and whether to show
// a filter input, from the number of options.
//
// Both defaults were wrong before, and visibly so on a two-option menu:
//
//   - HEIGHT. huh's Height counts the TITLE, not just the rows (updateViewportSize
//     subtracts the title's height from it), so passing Height(len(options)) left room for
//     len(options)-1 rows and silently hid the last choice. Its unset branch already sizes
//     the viewport to the options, so the fix is to set nothing unless the list is long
//     enough to actually need capping.
//   - FILTERING. A filter input on a two-item menu is noise that reads as a free-text
//     field — you cannot type your way to an answer that is right there. Reserve it for
//     lists long enough that scanning is the slow part.
func selectLayout(n int) (height int, filtering bool) {
	if n > selectMenuMax {
		// +1 for the title line huh charges against Height.
		height = selectMenuMax + 1
	}
	return height, n >= selectFilterMin
}

func (p ttyPrompter) SelectOne(title string, opts []Option) (string, error) {
	width := promptWidth(p.out)
	options := make([]huh.Option[string], len(opts))
	for i, o := range opts {
		label := o.Label
		if width > 6 {
			// Keep it one line: reserve the "> " caret + a small right margin.
			label = ansi.Truncate(label, width-4, "…")
		}
		options[i] = huh.NewOption(label, o.Value)
	}
	height, filtering := selectLayout(len(options))
	var v string
	field := huh.NewSelect[string]().
		Title(title).
		Options(options...).
		Filtering(filtering).
		Value(&v)
	if height > 0 {
		field = field.Height(height)
	}
	err := huh.NewForm(huh.NewGroup(field)).
		WithInput(p.in).
		WithOutput(p.out).
		WithTheme(p.pickerTheme()).
		WithKeyMap(abortKeyMap()).
		Run()
	if errors.Is(err, huh.ErrUserAborted) {
		return "", ErrAborted
	}
	if err != nil {
		return "", err
	}
	return v, nil
}

// Text reads a single line via huh, themed to match the picker.
func (p ttyPrompter) Text(title, placeholder string) (string, error) {
	var v string
	field := huh.NewInput().Title(title).Placeholder(placeholder).Value(&v)
	err := huh.NewForm(huh.NewGroup(field)).
		WithInput(p.in).
		WithOutput(p.out).
		WithTheme(p.pickerTheme()).
		WithKeyMap(abortKeyMap()).
		Run()
	if errors.Is(err, huh.ErrUserAborted) {
		return "", ErrAborted
	}
	if err != nil {
		return "", err
	}
	return v, nil
}

// pickerTheme is the shared huh theme — Dracula as the base, isDark-parameterized
// so huh supplies it from its own (bubbletea v2) background detection — with the
// selection caret + current row drawn in the ACTIVE theme's accent, so the picker
// matches the config-selected CLI/TUI theme instead of a local literal.
func (p ttyPrompter) pickerTheme() huh.Theme {
	return huh.ThemeFunc(func(isDark bool) *huh.Styles {
		s := huh.ThemeDracula(isDark)
		accent := p.theme.For(isDark).Accent.Color()
		s.Focused.SelectSelector = s.Focused.SelectSelector.SetString("> ").Foreground(accent)
		s.Focused.SelectedOption = s.Focused.SelectedOption.Foreground(accent).Bold(true)
		return s
	})
}

// promptWidth is the terminal width of out (the prompts render to stderr), or 80 when
// it can't be detected (piped/redirected) — used to keep each option label on one line.
func promptWidth(out io.Writer) int {
	if f, ok := out.(*os.File); ok {
		if w, _, err := term.GetSize(int(f.Fd())); err == nil && w > 0 {
			return w
		}
	}
	return 80
}
