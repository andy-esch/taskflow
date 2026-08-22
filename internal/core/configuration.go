package core

import (
	"fmt"
	"strings"

	"github.com/andy-esch/taskflow/internal/domain"
)

// ConfigurationStore is the secondary-adapter port for configuration use cases.
// Interfaces live at the consumer: filesystem TOML is one implementation, while tests
// and a future served adapter can supply another without teaching primary adapters how
// repository and user files are stored.
type ConfigurationStore interface {
	LoadConfiguration(start string) (ConfigurationState, error)
	MigrateConfiguration(start string, dryRun bool) (ConfigurationMigration, error)
	DiagnoseConfiguration(start string) (ConfigurationDiagnosis, error)
	SetPreference(start string, change PreferenceChange, dryRun bool) (PreferenceResult, error)
}

// ConfigurationService owns the configuration lifecycle shared by CLI, TUI, and future
// adapters. It contains precedence and validation; it knows no Cobra, Bubble Tea, TOML,
// environment variables, or filesystem paths beyond the values returned by its port.
type ConfigurationService struct {
	store         ConfigurationStore
	spaceRegistry *SpaceRegistryService
	knownThemes   map[string]struct{}
}

// ConfigurationOption supplies application policy without coupling core to a
// concrete design registry.
type ConfigurationOption func(*ConfigurationService)

// WithConfigurationThemes injects the registered theme vocabulary. When present,
// every primary adapter receives the same validation; "auto" remains the explicit
// spelling for the built-in default.
func WithConfigurationThemes(names []string) ConfigurationOption {
	return func(s *ConfigurationService) {
		s.knownThemes = make(map[string]struct{}, len(names))
		for _, name := range names {
			if name = strings.TrimSpace(name); name != "" {
				s.knownThemes[name] = struct{}{}
			}
		}
	}
}

// WithSpaceRegistry composes home-registry diagnosis into configuration doctor without
// coupling the repo-configuration store to home-scoped persistence.
func WithSpaceRegistry(registry *SpaceRegistryService) ConfigurationOption {
	return func(s *ConfigurationService) {
		s.spaceRegistry = registry
	}
}

func NewConfigurationService(store ConfigurationStore, opts ...ConfigurationOption) *ConfigurationService {
	svc := &ConfigurationService{store: store}
	for _, opt := range opts {
		opt(svc)
	}
	return svc
}

type ConfigMode string

const (
	ConfigModeScaffold   ConfigMode = "scaffold"
	ConfigModePointer    ConfigMode = "pointer"
	ConfigModeDiscovered ConfigMode = "discovered"
)

type ConfigSource string

const (
	ConfigSourceFlag        ConfigSource = "flag"
	ConfigSourceEnvironment ConfigSource = "environment"
	ConfigSourceRepository  ConfigSource = "repository"
	ConfigSourceUser        ConfigSource = "user"
	ConfigSourceDefault     ConfigSource = "default"
)

type RepositoryConfiguration struct {
	Path             string
	Dir              string
	PlanningRoot     string
	Mode             ConfigMode
	TaskflowRoot     string
	ID               string
	PlanningRepo     string
	PlanningRepoID   string
	TrackedRepos     []string
	ThemeName        string
	PagerEnabled     *bool
	PagerCommand     string
	PendingMigration []ConfigurationMigrationKind
	MigrationWarning string
}

type UserConfiguration struct {
	Path         string
	Exists       bool
	ThemeName    string
	PagerEnabled *bool
	PagerCommand string
	Warning      string
	RegistryPath string
}

// ConfigurationState is raw data loaded through the port, before invocation-level
// overrides are folded in.
type ConfigurationState struct {
	Repository RepositoryConfiguration
	User       UserConfiguration
}

type EffectiveString struct {
	Value  string
	Source ConfigSource
}

type EffectiveBool struct {
	Value  bool
	Source ConfigSource
}

type EffectiveConfiguration struct {
	Theme        EffectiveString
	PagerEnabled EffectiveBool
	PagerCommand EffectiveString
}

type ConfigurationSnapshot struct {
	Repository RepositoryConfiguration
	User       UserConfiguration
	Effective  EffectiveConfiguration
}

// ConfigurationOverrides are invocation inputs supplied by the primary adapter. Empty
// values mean absent. Environment access stays outside the core so a web adapter and
// tests can resolve the same snapshot deterministically.
type ConfigurationOverrides struct {
	ThemeFlag          string
	ThemeEnvironment   string
	PagerEnabledFlag   *bool
	PagerEnvironment   string
	GenericPagerEnv    string
	DefaultTheme       string
	KnownThemes        []string
	DefaultPagerEnable bool
	DefaultPager       string
}

func (s *ConfigurationService) Snapshot(start string, overrides ConfigurationOverrides) (ConfigurationSnapshot, error) {
	state, err := s.store.LoadConfiguration(start)
	if err != nil {
		return ConfigurationSnapshot{}, err
	}
	return ConfigurationSnapshot{
		Repository: state.Repository,
		User:       state.User,
		Effective:  resolveEffectiveConfiguration(state, overrides),
	}, nil
}

func resolveEffectiveConfiguration(state ConfigurationState, o ConfigurationOverrides) EffectiveConfiguration {
	theme := firstString(
		settingString{o.ThemeFlag, ConfigSourceFlag},
		settingString{o.ThemeEnvironment, ConfigSourceEnvironment},
		settingString{state.Repository.ThemeName, ConfigSourceRepository},
		settingString{state.User.ThemeName, ConfigSourceUser},
		settingString{o.DefaultTheme, ConfigSourceDefault},
	)
	if strings.EqualFold(theme.Value, "auto") {
		theme.Value = o.DefaultTheme
	} else if !containsString(o.KnownThemes, theme.Value) {
		theme = EffectiveString{Value: o.DefaultTheme, Source: ConfigSourceDefault}
	}

	pagerEnabled := EffectiveBool{Value: o.DefaultPagerEnable, Source: ConfigSourceDefault}
	switch {
	case o.PagerEnabledFlag != nil:
		pagerEnabled = EffectiveBool{Value: *o.PagerEnabledFlag, Source: ConfigSourceFlag}
	case state.Repository.PagerEnabled != nil:
		pagerEnabled = EffectiveBool{Value: *state.Repository.PagerEnabled, Source: ConfigSourceRepository}
	case state.User.PagerEnabled != nil:
		pagerEnabled = EffectiveBool{Value: *state.User.PagerEnabled, Source: ConfigSourceUser}
	}

	pagerCommand := firstString(
		settingString{o.PagerEnvironment, ConfigSourceEnvironment},
		settingString{state.Repository.PagerCommand, ConfigSourceRepository},
		settingString{state.User.PagerCommand, ConfigSourceUser},
		settingString{o.GenericPagerEnv, ConfigSourceEnvironment},
		settingString{o.DefaultPager, ConfigSourceDefault},
	)

	return EffectiveConfiguration{Theme: theme, PagerEnabled: pagerEnabled, PagerCommand: pagerCommand}
}

func containsString(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return len(values) == 0
}

type settingString struct {
	value  string
	source ConfigSource
}

func firstString(values ...settingString) EffectiveString {
	for _, v := range values {
		if value := strings.TrimSpace(v.value); value != "" {
			return EffectiveString{Value: value, Source: v.source}
		}
	}
	return EffectiveString{Source: ConfigSourceDefault}
}

type ConfigurationMigrationKind string

const (
	ConfigurationMigrationRepoID         ConfigurationMigrationKind = "repo-id"
	ConfigurationMigrationPlanningRepoID ConfigurationMigrationKind = "planning-repo-id"
)

type ConfigurationMigrationStep struct {
	Kind  ConfigurationMigrationKind
	Key   string
	Value string
}

type ConfigurationMigration struct {
	ConfigPath string
	Mode       ConfigMode
	DryRun     bool
	Steps      []ConfigurationMigrationStep
}

func (m ConfigurationMigration) Changed() bool { return len(m.Steps) > 0 }

func (s *ConfigurationService) Migrate(start string, dryRun bool) (ConfigurationMigration, error) {
	return s.store.MigrateConfiguration(start, dryRun)
}

type ConfigurationProblem struct {
	Repo    string
	Message string
}

type RegistryProblem struct {
	ID      string
	Path    string
	Kind    SpaceState
	Message string
	Remedy  string
}

type RegistryDiagnosis struct {
	Checked  int
	Problems []RegistryProblem
}

type ConfigurationDiagnosis struct {
	Root         string
	Problems     []ConfigurationProblem
	Registry     RegistryDiagnosis
	ConfigPath   string
	UserPath     string
	RegistryPath string
}

func (d ConfigurationDiagnosis) ProblemCount() int {
	return len(d.Problems) + len(d.Registry.Problems)
}

func (s *ConfigurationService) Diagnose(start string) (ConfigurationDiagnosis, error) {
	diagnosis, err := s.store.DiagnoseConfiguration(start)
	if err != nil {
		return ConfigurationDiagnosis{}, err
	}
	if diagnosis.Problems == nil {
		diagnosis.Problems = []ConfigurationProblem{}
	}
	if s.spaceRegistry == nil {
		if diagnosis.Registry.Problems == nil {
			diagnosis.Registry.Problems = []RegistryProblem{}
		}
		return diagnosis, nil
	}
	catalog, err := s.spaceRegistry.Catalog()
	if err != nil {
		return ConfigurationDiagnosis{}, err
	}
	diagnosis.Registry = RegistryDiagnosis{
		Checked: len(catalog.Entries), Problems: []RegistryProblem{},
	}
	for _, entry := range catalog.Entries {
		if entry.Healthy() {
			continue
		}
		diagnosis.Registry.Problems = append(diagnosis.Registry.Problems, RegistryProblem{
			ID: entry.ID, Path: entry.Path, Kind: entry.State,
			Message: entry.Detail, Remedy: entry.Remedy,
		})
	}
	return diagnosis, nil
}

type ConfigScope string

const (
	ConfigScopeUser       ConfigScope = "user"
	ConfigScopeRepository ConfigScope = "repository"
)

type PreferenceField string

const (
	PreferenceTheme        PreferenceField = "theme.name"
	PreferencePagerEnabled PreferenceField = "pager.enabled"
	PreferencePagerCommand PreferenceField = "pager.command"
)

// PreferenceChange is one typed edit. Unset removes the scoped override and restores
// inheritance. Value is validated by the service before the store sees it.
type PreferenceChange struct {
	Scope ConfigScope
	Field PreferenceField
	Value string
	Unset bool
}

type PreferenceResult struct {
	Path    string
	Change  PreferenceChange
	Changed bool
	DryRun  bool
}

func (s *ConfigurationService) SetPreference(start string, change PreferenceChange, dryRun bool) (PreferenceResult, error) {
	change.Value = strings.TrimSpace(change.Value)
	if err := s.validatePreferenceChange(change); err != nil {
		return PreferenceResult{}, err
	}
	return s.store.SetPreference(start, change, dryRun)
}

func (s *ConfigurationService) validatePreferenceChange(change PreferenceChange) error {
	if change.Scope != ConfigScopeUser && change.Scope != ConfigScopeRepository {
		return fmt.Errorf("%w: config scope must be user or repository", domain.ErrValidation)
	}
	switch change.Field {
	case PreferenceTheme:
		if !change.Unset && strings.TrimSpace(change.Value) == "" {
			return fmt.Errorf("%w: %s cannot be blank (unset it to inherit)", domain.ErrValidation, change.Field)
		}
		if !change.Unset && !strings.EqualFold(change.Value, "auto") && len(s.knownThemes) > 0 {
			if _, ok := s.knownThemes[change.Value]; !ok {
				return fmt.Errorf("%w: unknown theme %q", domain.ErrValidation, change.Value)
			}
		}
	case PreferencePagerCommand:
		if !change.Unset && strings.TrimSpace(change.Value) == "" {
			return fmt.Errorf("%w: %s cannot be blank (unset it to inherit)", domain.ErrValidation, change.Field)
		}
	case PreferencePagerEnabled:
		if !change.Unset && change.Value != "true" && change.Value != "false" {
			return fmt.Errorf("%w: pager.enabled must be true, false, or unset", domain.ErrValidation)
		}
	default:
		return fmt.Errorf("%w: unsupported config field %q", domain.ErrValidation, change.Field)
	}
	return nil
}
