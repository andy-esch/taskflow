// Package configui is the reusable Bubble Tea primary adapter for configuration.
// It is embedded by both `config edit` and the full taskflow TUI. Reads, validation,
// and writes all flow through core.ConfigurationService; this package never reads or
// edits TOML itself.
package configui

import (
	"fmt"
	"io"
	"strings"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/andy-esch/taskflow/internal/core"
	"github.com/andy-esch/taskflow/internal/design"
	"github.com/andy-esch/taskflow/internal/themepreview"
)

const (
	rowTheme = iota
	rowPagerEnabled
	rowPagerCommand
	rowCount
)

const (
	pagerInherit = iota
	pagerOn
	pagerOff
)

type loadedMsg struct {
	snapshot  core.ConfigurationSnapshot
	diagnosis core.ConfigurationDiagnosis
	healthErr error
	err       error
}

type savedMsg struct {
	snapshot core.ConfigurationSnapshot
	result   core.PreferenceResult
	err      error
}

// Editor is the shared, value-model configuration form. It exposes safe,
// presentation-only settings and a read-only Config/About summary. Scope defaults to
// user; choosing repository scope always requires an explicit key press.
type Editor struct {
	svc       *core.ConfigurationService
	start     string
	overrides core.ConfigurationOverrides
	dark      bool

	snapshot  core.ConfigurationSnapshot
	diagnosis core.ConfigurationDiagnosis
	healthErr error
	loaded    bool
	loading   bool
	saving    bool
	closed    bool

	scope       core.ConfigScope
	cursor      int
	themeCursor int
	pagerChoice int
	editing     bool
	command     textinput.Model

	width, height int
	notice        string
	err           string
}

// New constructs an editor. Init performs all I/O through the configuration service.
func New(svc *core.ConfigurationService, start string, overrides core.ConfigurationOverrides, dark bool) Editor {
	input := textinput.New()
	input.Prompt = ""
	input.Placeholder = "less -FRX"
	input.CharLimit = 512
	input.SetWidth(44)
	return Editor{
		svc: svc, start: start, overrides: overrides, dark: dark,
		scope: core.ConfigScopeUser, loading: true, command: input,
	}
}

// Init loads the typed projection and diagnosis. It is safe to call again when an
// embedded editor is reopened.
func (e Editor) Init() tea.Cmd { return e.loadCmd() }

func (e Editor) loadCmd() tea.Cmd {
	return func() tea.Msg {
		snapshot, err := e.svc.Snapshot(e.start, e.overrides)
		if err != nil {
			return loadedMsg{err: err}
		}
		diagnosis, healthErr := e.svc.Diagnose(e.start)
		return loadedMsg{snapshot: snapshot, diagnosis: diagnosis, healthErr: healthErr}
	}
}

// Update implements tea.Model. Embedded consumers may route the returned command's
// message back to this method; no command assumes ownership of the outer program.
func (e Editor) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		e.width, e.height = msg.Width, msg.Height
		return e, nil
	case loadedMsg:
		e.loading = false
		if msg.err != nil {
			e.err = msg.err.Error()
			return e, nil
		}
		e.loaded, e.err = true, ""
		e.snapshot, e.diagnosis, e.healthErr = msg.snapshot, msg.diagnosis, msg.healthErr
		e.resetScopedValues()
		return e, nil
	case savedMsg:
		e.saving = false
		if msg.err != nil {
			e.err = msg.err.Error()
			return e, nil
		}
		e.snapshot = msg.snapshot
		e.err = ""
		if msg.result.Changed {
			e.notice = "saved " + string(msg.result.Change.Field) + " in " + string(msg.result.Change.Scope) + " config"
		} else {
			e.notice = "already set"
		}
		e.resetScopedValues()
		return e, nil
	case tea.KeyPressMsg:
		return e.handleKey(msg)
	}
	if e.editing {
		var cmd tea.Cmd
		e.command, cmd = e.command.Update(msg)
		return e, cmd
	}
	return e, nil
}

func (e Editor) handleKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	if e.editing {
		switch msg.String() {
		case "esc":
			e.editing = false
			e.command.Blur()
			e.err = ""
			return e, nil
		case "enter":
			value := strings.TrimSpace(e.command.Value())
			if value == "" {
				e.err = "pager command cannot be blank; press u to inherit"
				return e, nil
			}
			e.editing = false
			e.command.Blur()
			return e.save(core.PreferenceChange{
				Scope: e.scope, Field: core.PreferencePagerCommand, Value: value,
			})
		}
		e.err = ""
		var cmd tea.Cmd
		e.command, cmd = e.command.Update(msg)
		return e, cmd
	}

	if e.saving {
		if msg.String() == "esc" || msg.String() == "q" {
			e.closed = true
		}
		return e, nil
	}
	if !e.loaded {
		if msg.String() == "esc" || msg.String() == "q" {
			e.closed = true
		}
		return e, nil
	}

	switch msg.String() {
	case "q", "esc":
		e.closed = true
		return e, nil
	case "up", "k":
		e.cursor = (e.cursor + rowCount - 1) % rowCount
	case "down", "j":
		e.cursor = (e.cursor + 1) % rowCount
	case "s", "tab":
		if e.scope == core.ConfigScopeUser {
			e.scope = core.ConfigScopeRepository
		} else {
			e.scope = core.ConfigScopeUser
		}
		e.notice, e.err = "", ""
		e.resetScopedValues()
	case "left", "h":
		e.moveValue(-1)
	case "right", "l":
		e.moveValue(1)
	case "u", "backspace":
		return e.unsetCurrent()
	case "enter":
		return e.applyCurrent()
	}
	return e, nil
}

func (e *Editor) moveValue(delta int) {
	e.notice, e.err = "", ""
	switch e.cursor {
	case rowTheme:
		if n := len(design.Names()); n > 0 {
			e.themeCursor = ((e.themeCursor+delta)%n + n) % n
		}
	case rowPagerEnabled:
		e.pagerChoice = ((e.pagerChoice+delta)%3 + 3) % 3
	}
}

func (e Editor) applyCurrent() (tea.Model, tea.Cmd) {
	switch e.cursor {
	case rowTheme:
		names := design.Names()
		if len(names) == 0 {
			return e, nil
		}
		return e.save(core.PreferenceChange{
			Scope: e.scope, Field: core.PreferenceTheme, Value: names[e.themeCursor],
		})
	case rowPagerEnabled:
		if e.pagerChoice == pagerInherit {
			return e.save(core.PreferenceChange{Scope: e.scope, Field: core.PreferencePagerEnabled, Unset: true})
		}
		value := "true"
		if e.pagerChoice == pagerOff {
			value = "false"
		}
		return e.save(core.PreferenceChange{
			Scope: e.scope, Field: core.PreferencePagerEnabled, Value: value,
		})
	case rowPagerCommand:
		e.editing = true
		e.err, e.notice = "", ""
		e.command.SetValue(e.rawPagerCommand())
		e.command.CursorEnd()
		return e, e.command.Focus()
	}
	return e, nil
}

func (e Editor) unsetCurrent() (tea.Model, tea.Cmd) {
	fields := [...]core.PreferenceField{core.PreferenceTheme, core.PreferencePagerEnabled, core.PreferencePagerCommand}
	return e.save(core.PreferenceChange{Scope: e.scope, Field: fields[e.cursor], Unset: true})
}

func (e Editor) save(change core.PreferenceChange) (tea.Model, tea.Cmd) {
	e.saving, e.err, e.notice = true, "", ""
	return e, func() tea.Msg {
		result, err := e.svc.SetPreference(e.start, change, false)
		if err != nil {
			return savedMsg{err: err}
		}
		snapshot, err := e.svc.Snapshot(e.start, e.overrides)
		return savedMsg{snapshot: snapshot, result: result, err: err}
	}
}

func (e *Editor) resetScopedValues() {
	themeName := e.rawTheme()
	if themeName == "" {
		themeName = e.snapshot.Effective.Theme.Value
	}
	e.themeCursor = stringIndex(design.Names(), themeName)
	if e.themeCursor < 0 {
		e.themeCursor = 0
	}
	switch enabled := e.rawPagerEnabled(); {
	case enabled == nil:
		e.pagerChoice = pagerInherit
	case *enabled:
		e.pagerChoice = pagerOn
	default:
		e.pagerChoice = pagerOff
	}
}

func (e Editor) rawTheme() string {
	if e.scope == core.ConfigScopeRepository {
		return e.snapshot.Repository.ThemeName
	}
	return e.snapshot.User.ThemeName
}

func (e Editor) rawPagerEnabled() *bool {
	if e.scope == core.ConfigScopeRepository {
		return e.snapshot.Repository.PagerEnabled
	}
	return e.snapshot.User.PagerEnabled
}

func (e Editor) rawPagerCommand() string {
	if e.scope == core.ConfigScopeRepository {
		return e.snapshot.Repository.PagerCommand
	}
	return e.snapshot.User.PagerCommand
}

func stringIndex(values []string, want string) int {
	for i, value := range values {
		if value == want {
			return i
		}
	}
	return -1
}

// Closed reports whether the user asked the containing surface to close.
func (e Editor) Closed() bool { return e.closed }

// SetSize supplies the available size when the editor is embedded in another TUI.
func (e *Editor) SetSize(width, height int) { e.width, e.height = width, height }

// SetDark updates the background variant used by the live preview.
func (e *Editor) SetDark(dark bool) { e.dark = dark }

// View implements tea.Model for the focused `config edit` program.
func (e Editor) View() tea.View {
	w, h := e.width, e.height
	if w < 1 {
		w = 80
	}
	if h < 1 {
		h = 28
	}
	v := tea.NewView(e.Content(w, h))
	v.AltScreen = true
	v.WindowTitle = "tskflwctl · config"
	return v
}

// Content renders the same editor as a bounded component for a parent TUI.
func (e Editor) Content(maxW, maxH int) string {
	if e.loading {
		return "loading configuration…"
	}
	if !e.loaded {
		return "Configuration\n\n" + e.err + "\n\nq/esc close"
	}
	pal := e.previewPalette()
	border := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(pal.BorderActive.Color()).Padding(0, 2)
	heading := lipgloss.NewStyle().Bold(true).Foreground(pal.Heading.Color())
	dim := lipgloss.NewStyle().Faint(true)
	accent := lipgloss.NewStyle().Bold(true).Foreground(pal.Accent.Color())

	var b strings.Builder
	fmt.Fprintln(&b, heading.Render("Configuration / About"))
	fmt.Fprintf(&b, "%s  %s\n\n", dim.Render("write scope"), accent.Render(scopeLabel(e.scope)))

	rows := []struct{ label, value string }{
		{"Theme", e.themeValue()},
		{"Pager", e.pagerValue()},
		{"Pager command", e.commandValue()},
	}
	for i, row := range rows {
		cursor := "  "
		label := row.label
		if i == e.cursor {
			cursor, label = "› ", accent.Render(label)
		}
		fmt.Fprintf(&b, "%s%-16s %s\n", cursor, label, row.value)
	}

	fmt.Fprintln(&b)
	fmt.Fprintln(&b, heading.Render("Live theme preview"))
	fmt.Fprintln(&b, e.preview(pal))

	fmt.Fprintln(&b, heading.Render("About this planning space"))
	repo := e.snapshot.Repository
	fmt.Fprintf(&b, "  %-14s %s\n", "mode", repo.Mode)
	fmt.Fprintf(&b, "  %-14s %s\n", "planning root", emptyAs(repo.PlanningRoot, "—"))
	fmt.Fprintf(&b, "  %-14s %s\n", "planning id", emptyAs(repo.ID, "not migrated"))
	fmt.Fprintf(&b, "  %-14s %s\n", "repo config", emptyAs(repo.Path, "—"))
	fmt.Fprintf(&b, "  %-14s %s\n", "user config", emptyAs(e.snapshot.User.Path, "—"))
	fmt.Fprintf(&b, "  %-14s %s\n", "health", e.healthLabel())
	if len(repo.PendingMigration) > 0 {
		fmt.Fprintf(&b, "  %-14s %s\n", "migration", "config migrate available")
	}

	status := ""
	if e.saving {
		status = accent.Render("saving…")
	} else if e.err != "" {
		status = lipgloss.NewStyle().Foreground(pal.Danger.Color()).Render(e.err)
	} else if e.notice != "" {
		status = accent.Render(e.notice)
	}
	if status != "" {
		fmt.Fprintln(&b)
		fmt.Fprintln(&b, status)
	}
	fmt.Fprintln(&b)
	fmt.Fprint(&b, dim.Render("↑/↓ field  ←/→ choose  enter edit/save  u inherit  s scope  q close"))

	width := min(max(maxW, 24), 86)
	box := border.MaxWidth(width).Render(strings.TrimRight(b.String(), "\n"))
	return lipgloss.NewStyle().MaxWidth(maxW).MaxHeight(maxH).Render(box)
}

func (e Editor) previewPalette() design.Palette {
	name := e.snapshot.Effective.Theme.Value
	if names := design.Names(); e.themeCursor >= 0 && e.themeCursor < len(names) {
		name = names[e.themeCursor]
	}
	theme, _ := design.Lookup(name)
	return theme.For(e.dark)
}

func (e Editor) preview(pal design.Palette) string {
	names := design.Names()
	name := ""
	if e.themeCursor >= 0 && e.themeCursor < len(names) {
		name = names[e.themeCursor]
	}
	parts := make([]string, 0, len(themepreview.Swatches(pal)))
	for _, swatch := range themepreview.Swatches(pal) {
		chip := lipgloss.NewStyle().Background(pal.Base.Color()).Foreground(swatch.Hue.Color()).Render("██")
		parts = append(parts, chip+" "+swatch.Token)
	}
	return "  " + name + "  " + strings.Join(parts, "  ") + "\n  " + themepreview.Bar(pal, 28)
}

func (e Editor) themeValue() string {
	names := design.Names()
	if e.themeCursor < 0 || e.themeCursor >= len(names) {
		return "—"
	}
	value := names[e.themeCursor]
	if e.rawTheme() == "" {
		return value + "  (inherited; enter to override)"
	}
	return value
}

func (e Editor) pagerValue() string {
	switch e.pagerChoice {
	case pagerOn:
		return "on"
	case pagerOff:
		return "off"
	default:
		state := "off"
		if e.snapshot.Effective.PagerEnabled.Value {
			state = "on"
		}
		return "inherit  (effective " + state + ")"
	}
}

func (e Editor) commandValue() string {
	if e.editing {
		return e.command.View()
	}
	if value := e.rawPagerCommand(); value != "" {
		return value
	}
	return "inherit  (effective " + e.snapshot.Effective.PagerCommand.Value + ")"
}

func (e Editor) healthLabel() string {
	if e.healthErr != nil {
		return "unavailable: " + e.healthErr.Error()
	}
	if n := e.diagnosis.ProblemCount(); n > 0 {
		return fmt.Sprintf("%d problem(s); run config doctor", n)
	}
	return "ok"
}

func scopeLabel(scope core.ConfigScope) string {
	if scope == core.ConfigScopeRepository {
		return "repository override"
	}
	return "user (default)"
}

func emptyAs(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}

type standalone struct{ Editor }

func (m standalone) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	next, cmd := m.Editor.Update(msg)
	m.Editor = next.(Editor)
	if m.Closed() {
		return m, tea.Quit
	}
	return m, cmd
}

// Run launches the focused editor with the supplied terminal streams.
func Run(svc *core.ConfigurationService, start string, overrides core.ConfigurationOverrides, dark bool, in io.Reader, out io.Writer) error {
	m := standalone{Editor: New(svc, start, overrides, dark)}
	_, err := tea.NewProgram(m, tea.WithInput(in), tea.WithOutput(out)).Run()
	return err
}

// PlainContent is a small testing/debugging helper that strips terminal styling from
// the component view without exposing its internal style construction.
func (e Editor) PlainContent(width, height int) string { return ansi.Strip(e.Content(width, height)) }
