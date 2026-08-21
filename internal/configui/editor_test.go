package configui

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/andy-esch/taskflow/internal/core"
)

type editorStore struct {
	state     core.ConfigurationState
	diagnosis core.ConfigurationDiagnosis
	changes   []core.PreferenceChange
}

func (s *editorStore) LoadConfiguration(string) (core.ConfigurationState, error) {
	return s.state, nil
}
func (s *editorStore) MigrateConfiguration(string, bool) (core.ConfigurationMigration, error) {
	return core.ConfigurationMigration{}, nil
}
func (s *editorStore) DiagnoseConfiguration(string) (core.ConfigurationDiagnosis, error) {
	return s.diagnosis, nil
}
func (s *editorStore) SetPreference(_ string, change core.PreferenceChange, dryRun bool) (core.PreferenceResult, error) {
	s.changes = append(s.changes, change)
	s.apply(change)
	return core.PreferenceResult{Path: "/config.toml", Change: change, Changed: true, DryRun: dryRun}, nil
}

func (s *editorStore) apply(change core.PreferenceChange) {
	var theme *string
	var enabled **bool
	var command *string
	if change.Scope == core.ConfigScopeRepository {
		theme = &s.state.Repository.ThemeName
		enabled = &s.state.Repository.PagerEnabled
		command = &s.state.Repository.PagerCommand
	} else {
		theme = &s.state.User.ThemeName
		enabled = &s.state.User.PagerEnabled
		command = &s.state.User.PagerCommand
	}
	switch change.Field {
	case core.PreferenceTheme:
		if change.Unset {
			*theme = ""
		} else {
			*theme = change.Value
		}
	case core.PreferencePagerEnabled:
		if change.Unset {
			*enabled = nil
		} else {
			value := change.Value == "true"
			*enabled = &value
		}
	case core.PreferencePagerCommand:
		if change.Unset {
			*command = ""
		} else {
			*command = change.Value
		}
	}
}

func testEditor(t *testing.T) (Editor, *editorStore) {
	t.Helper()
	store := &editorStore{
		state: core.ConfigurationState{
			Repository: core.RepositoryConfiguration{
				Path: "/repo/.tskflwctl.toml", PlanningRoot: "/planning", ID: "planning-id",
				Mode: core.ConfigModePointer, PendingMigration: []core.ConfigurationMigrationKind{},
			},
			User: core.UserConfiguration{Path: "/home/config.toml"},
		},
		diagnosis: core.ConfigurationDiagnosis{Root: "/planning"},
	}
	e := New(core.NewConfigurationService(store), "/repo", core.ConfigurationOverrides{
		DefaultTheme: "neon", KnownThemes: []string{"catppuccin", "neon"},
		DefaultPagerEnable: true, DefaultPager: "less -FRX",
	}, true)
	next, _ := e.Update(e.Init()())
	return next.(Editor), store
}

func configKey(s string) tea.KeyPressMsg {
	switch s {
	case "enter":
		return tea.KeyPressMsg{Code: tea.KeyEnter}
	case "esc":
		return tea.KeyPressMsg{Code: tea.KeyEsc}
	case "right":
		return tea.KeyPressMsg{Code: tea.KeyRight}
	case "down":
		return tea.KeyPressMsg{Code: tea.KeyDown}
	case "backspace":
		return tea.KeyPressMsg{Code: tea.KeyBackspace}
	default:
		r := []rune(s)[0]
		return tea.KeyPressMsg{Code: r, Text: s}
	}
}

func updateEditor(t *testing.T, e Editor, key string) (Editor, tea.Cmd) {
	t.Helper()
	next, cmd := e.Update(configKey(key))
	return next.(Editor), cmd
}

func settleEditor(t *testing.T, e Editor, cmd tea.Cmd) Editor {
	t.Helper()
	if cmd == nil {
		t.Fatal("expected a command")
	}
	next, _ := e.Update(cmd())
	return next.(Editor)
}

func TestEditorDefaultsToUserScopeAndShowsReadOnlyAbout(t *testing.T) {
	e, _ := testEditor(t)
	view := e.PlainContent(100, 40)
	for _, want := range []string{
		"Configuration / About", "user (default)", "planning-id", "/planning", "health", "ok",
		"inherit  (effective on)", "s scope",
	} {
		if !strings.Contains(view, want) {
			t.Errorf("view missing %q:\n%s", want, view)
		}
	}
}

func TestEditorThemePreviewMovesBeforeScopedSave(t *testing.T) {
	e, store := testEditor(t)
	if got := e.PlainContent(100, 40); !strings.Contains(got, "neon") {
		t.Fatalf("initial preview should use effective theme:\n%s", got)
	}
	e, _ = updateEditor(t, e, "right") // neon wraps to catppuccin
	if got := e.PlainContent(100, 40); !strings.Contains(got, "catppuccin") {
		t.Fatalf("preview did not repaint on cursor movement:\n%s", got)
	}
	if len(store.changes) != 0 {
		t.Fatal("moving a preview cursor must not write")
	}
	e, cmd := updateEditor(t, e, "enter")
	e = settleEditor(t, e, cmd)
	if len(store.changes) != 1 || store.changes[0] != (core.PreferenceChange{
		Scope: core.ConfigScopeUser, Field: core.PreferenceTheme, Value: "catppuccin",
	}) {
		t.Fatalf("change = %+v", store.changes)
	}
	if e.snapshot.User.ThemeName != "catppuccin" {
		t.Fatalf("snapshot was not refreshed after save: %+v", e.snapshot.User)
	}
}

func TestEditorRepositoryScopeIsExplicitAndPagerRemainsTriState(t *testing.T) {
	e, store := testEditor(t)
	e, _ = updateEditor(t, e, "s")
	e, _ = updateEditor(t, e, "down")
	e, _ = updateEditor(t, e, "right") // inherit → on
	e, cmd := updateEditor(t, e, "enter")
	e = settleEditor(t, e, cmd)
	if got := store.changes[0]; got.Scope != core.ConfigScopeRepository || got.Field != core.PreferencePagerEnabled || got.Value != "true" || got.Unset {
		t.Fatalf("repository pager change = %+v", got)
	}
	e, cmd = updateEditor(t, e, "u")
	_ = settleEditor(t, e, cmd)
	if got := store.changes[1]; !got.Unset || got.Scope != core.ConfigScopeRepository || got.Field != core.PreferencePagerEnabled {
		t.Fatalf("inherit change = %+v", got)
	}
}

func TestEditorPagerCommandRejectsBlankAndSavesTypedValue(t *testing.T) {
	e, store := testEditor(t)
	e, _ = updateEditor(t, e, "down")
	e, _ = updateEditor(t, e, "down")
	e, _ = updateEditor(t, e, "enter")
	e.command.SetValue("")
	e, cmd := updateEditor(t, e, "enter")
	if cmd != nil || !e.editing || !strings.Contains(e.err, "cannot be blank") {
		t.Fatalf("blank command should remain in editor: editing=%v err=%q", e.editing, e.err)
	}
	e.command.SetValue("delta --dark")
	e, cmd = updateEditor(t, e, "enter")
	e = settleEditor(t, e, cmd)
	if got := store.changes[0]; got.Scope != core.ConfigScopeUser || got.Field != core.PreferencePagerCommand || got.Value != "delta --dark" {
		t.Fatalf("pager command change = %+v", got)
	}
}
