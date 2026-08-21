package tui

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/andy-esch/taskflow/internal/core"
	"github.com/andy-esch/taskflow/internal/store"
)

type tuiConfigStore struct {
	state   core.ConfigurationState
	changes []core.PreferenceChange
}

func (s *tuiConfigStore) LoadConfiguration(string) (core.ConfigurationState, error) {
	return s.state, nil
}
func (s *tuiConfigStore) MigrateConfiguration(string, bool) (core.ConfigurationMigration, error) {
	return core.ConfigurationMigration{}, nil
}
func (s *tuiConfigStore) DiagnoseConfiguration(string) (core.ConfigurationDiagnosis, error) {
	return core.ConfigurationDiagnosis{Root: s.state.Repository.PlanningRoot}, nil
}
func (s *tuiConfigStore) SetPreference(_ string, change core.PreferenceChange, dryRun bool) (core.PreferenceResult, error) {
	s.changes = append(s.changes, change)
	if change.Field == core.PreferenceTheme && change.Scope == core.ConfigScopeUser {
		if change.Unset {
			s.state.User.ThemeName = ""
		} else {
			s.state.User.ThemeName = change.Value
		}
	}
	return core.PreferenceResult{Change: change, Changed: true, DryRun: dryRun}, nil
}

func configuredModel(t *testing.T) (Model, *tuiConfigStore) {
	t.Helper()
	root := seedRepo(t)
	cfgStore := &tuiConfigStore{state: core.ConfigurationState{
		Repository: core.RepositoryConfiguration{
			Path: root + "/.tskflwctl.toml", PlanningRoot: root, ID: "space-id",
			Mode: core.ConfigModeScaffold, PendingMigration: []core.ConfigurationMigrationKind{},
		},
		User: core.UserConfiguration{Path: "/home/config.toml"},
	}}
	configSvc := core.NewConfigurationService(cfgStore)
	m := New(core.NewService(store.NewFS(root)), WithConfiguration(configSvc, root, core.ConfigurationOverrides{
		DefaultTheme: "neon", KnownThemes: []string{"catppuccin", "neon"},
		DefaultPagerEnable: true, DefaultPager: "less -FRX",
	}))
	tm, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 42})
	m = tm.(Model)
	tm, _ = m.Update(m.Init()())
	return tm.(Model), cfgStore
}

func TestMainTUIConfigCommandMountsSharedEditor(t *testing.T) {
	m, _ := configuredModel(t)
	if !containsWord(m.paletteCommands(), "config") {
		t.Fatal("config should be discoverable in the command palette")
	}
	cmd := m.runPaletteCommand("config")
	if !m.configOpen || cmd == nil {
		t.Fatal(":config should open and load the Config/About surface")
	}
	m = drain(t, m, cmd)
	view := ansi.Strip(m.View().Content)
	for _, want := range []string{"Configuration / About", "user (default)", "space-id", "health"} {
		if !strings.Contains(view, want) {
			t.Errorf("config overlay missing %q:\n%s", want, view)
		}
	}

	// The overlay owns scope keys, then q closes only the overlay rather than the app.
	tm, _ := m.Update(press("s"))
	m = tm.(Model)
	if got := m.configEditor.PlainContent(100, 40); !strings.Contains(got, "repository override") {
		t.Fatalf("scope key did not reach shared editor:\n%s", got)
	}
	tm, cmd = m.Update(press("q"))
	m = tm.(Model)
	if m.configOpen || cmd != nil {
		t.Fatalf("q should close the Config/About layer only: open=%v cmd=%T", m.configOpen, cmd)
	}
}

func TestMainTUIConfigIsDiscoverableFromOverviewAndHelp(t *testing.T) {
	m, _ := configuredModel(t)
	view := ansi.Strip(m.View().Content)
	for _, want := range []string{"workspace", "Configuration / About"} {
		if !strings.Contains(view, want) {
			t.Errorf("configured Overview should show %q:\n%s", want, view)
		}
	}

	configCursor := -1
	for cursor, row := range m.dash.nav {
		if target := m.dash.rows[row].target; target != nil && target.action == dashOpenConfiguration {
			configCursor = cursor
			break
		}
	}
	if configCursor < 0 {
		t.Fatal("configured Overview should have a selectable Configuration / About target")
	}
	m.dash.cursor = configCursor
	tm, cmd := m.Update(press("enter"))
	m = tm.(Model)
	if !m.onDash || !m.configOpen || cmd == nil {
		t.Fatal("enter on the Overview Configuration / About row should open and load the shared editor")
	}
	m = drain(t, m, cmd)
	if got := m.configEditor.PlainContent(100, 40); !strings.Contains(got, "Configuration / About") {
		t.Fatalf("Overview entry did not load the shared editor:\n%s", got)
	}

	help := strings.Join(helpLines(focusList, entityDashboard, 120, &testStyles), "\n")
	for _, want := range []string{":config", "open Configuration / About for this space"} {
		if !strings.Contains(help, want) {
			t.Errorf("? help should document %q:\n%s", want, help)
		}
	}

	// New(svc) is intentionally valid for read-only/embedded consumers. Without
	// the optional capability, Overview must not offer a selectable dead end.
	unconfigured := loadedDash(t, 120, 40)
	if got := ansi.Strip(unconfigured.View().Content); strings.Contains(got, "Configuration / About") {
		t.Errorf("Overview without ConfigurationService should omit the config entry:\n%s", got)
	}
}

func TestMainTUIConfigEditorWritesThroughConfigurationService(t *testing.T) {
	m, cfgStore := configuredModel(t)
	m = drain(t, m, m.runPaletteCommand("config"))
	tm, _ := m.Update(press("l")) // neon → catppuccin, live preview only
	m = tm.(Model)
	if len(cfgStore.changes) != 0 {
		t.Fatal("moving the theme cursor must not write")
	}
	tm, cmd := m.Update(press("enter"))
	m = drain(t, tm.(Model), cmd)
	if len(cfgStore.changes) != 1 || cfgStore.changes[0].Field != core.PreferenceTheme || cfgStore.changes[0].Scope != core.ConfigScopeUser {
		t.Fatalf("configuration change = %+v", cfgStore.changes)
	}
	if !m.configOpen {
		t.Fatal("successful save should keep Config/About open")
	}
}

func containsWord(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}
