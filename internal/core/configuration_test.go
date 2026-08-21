package core

import "testing"

type fakeConfigurationStore struct {
	state     ConfigurationState
	migration ConfigurationMigration
	diagnosis ConfigurationDiagnosis
	change    PreferenceChange
}

func (f *fakeConfigurationStore) LoadConfiguration(string) (ConfigurationState, error) {
	return f.state, nil
}
func (f *fakeConfigurationStore) MigrateConfiguration(string, bool) (ConfigurationMigration, error) {
	return f.migration, nil
}
func (f *fakeConfigurationStore) DiagnoseConfiguration(string) (ConfigurationDiagnosis, error) {
	return f.diagnosis, nil
}
func (f *fakeConfigurationStore) SetPreference(_ string, change PreferenceChange, dryRun bool) (PreferenceResult, error) {
	f.change = change
	return PreferenceResult{Change: change, DryRun: dryRun}, nil
}

func boolPtr(v bool) *bool { return &v }

func TestConfigurationSnapshotResolvesFieldLevelPrecedence(t *testing.T) {
	store := &fakeConfigurationStore{state: ConfigurationState{
		Repository: RepositoryConfiguration{
			ThemeName: "repo-theme", PagerCommand: "repo-pager",
		},
		User: UserConfiguration{
			ThemeName: "user-theme", PagerEnabled: boolPtr(false), PagerCommand: "user-pager",
		},
	}}
	svc := NewConfigurationService(store)
	snapshot, err := svc.Snapshot(".", ConfigurationOverrides{
		ThemeEnvironment: "env-theme", DefaultTheme: "default-theme",
		DefaultPagerEnable: true, DefaultPager: "default-pager",
	})
	if err != nil {
		t.Fatal(err)
	}
	if got := snapshot.Effective.Theme; got.Value != "env-theme" || got.Source != ConfigSourceEnvironment {
		t.Errorf("theme = %+v", got)
	}
	if got := snapshot.Effective.PagerEnabled; got.Value || got.Source != ConfigSourceUser {
		t.Errorf("pager enabled = %+v", got)
	}
	if got := snapshot.Effective.PagerCommand; got.Value != "repo-pager" || got.Source != ConfigSourceRepository {
		t.Errorf("pager command = %+v", got)
	}
}

func TestConfigurationSnapshotFlagOverrides(t *testing.T) {
	store := &fakeConfigurationStore{state: ConfigurationState{Repository: RepositoryConfiguration{
		ThemeName: "repo", PagerEnabled: boolPtr(true),
	}}}
	svc := NewConfigurationService(store)
	snapshot, err := svc.Snapshot(".", ConfigurationOverrides{
		ThemeFlag: "flag-theme", PagerEnabledFlag: boolPtr(false),
		PagerEnvironment: "env-pager", DefaultPagerEnable: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if snapshot.Effective.Theme.Source != ConfigSourceFlag || snapshot.Effective.Theme.Value != "flag-theme" {
		t.Errorf("theme = %+v", snapshot.Effective.Theme)
	}
	if snapshot.Effective.PagerEnabled.Source != ConfigSourceFlag || snapshot.Effective.PagerEnabled.Value {
		t.Errorf("pager enabled = %+v", snapshot.Effective.PagerEnabled)
	}
	if snapshot.Effective.PagerCommand.Source != ConfigSourceEnvironment || snapshot.Effective.PagerCommand.Value != "env-pager" {
		t.Errorf("pager command = %+v", snapshot.Effective.PagerCommand)
	}
}

func TestConfigurationServiceValidatesPreferenceBeforeStore(t *testing.T) {
	store := &fakeConfigurationStore{}
	svc := NewConfigurationService(store, WithConfigurationThemes([]string{"neon"}))
	if _, err := svc.SetPreference(".", PreferenceChange{
		Scope: ConfigScopeUser, Field: PreferencePagerEnabled, Value: "maybe",
	}, false); err == nil {
		t.Fatal("invalid pager value should fail")
	}
	if store.change.Field != "" {
		t.Fatal("invalid change reached the store")
	}
	if _, err := svc.SetPreference(".", PreferenceChange{
		Scope: ConfigScopeUser, Field: PreferenceTheme, Value: "not-a-theme",
	}, false); err == nil {
		t.Fatal("unknown theme should fail before reaching the store")
	}
	change := PreferenceChange{Scope: ConfigScopeUser, Field: PreferencePagerEnabled, Unset: true}
	if _, err := svc.SetPreference(".", change, false); err != nil {
		t.Fatal(err)
	}
	if store.change != change {
		t.Errorf("stored change = %+v", store.change)
	}
}

func TestConfigurationSnapshotAutoKeepsItsProvenance(t *testing.T) {
	store := &fakeConfigurationStore{state: ConfigurationState{Repository: RepositoryConfiguration{ThemeName: "auto"}}}
	snapshot, err := NewConfigurationService(store).Snapshot(".", ConfigurationOverrides{
		DefaultTheme: "neon", KnownThemes: []string{"neon"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if got := snapshot.Effective.Theme; got.Value != "neon" || got.Source != ConfigSourceRepository {
		t.Fatalf("auto theme = %+v, want resolved default with repository provenance", got)
	}
}
