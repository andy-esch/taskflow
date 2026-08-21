// Package configstore is the filesystem secondary adapter for core's configuration
// service. It composes repo config, user config, link health, and registry health without
// allowing the repo-scoped discovery package to depend on home-scoped state.
package configstore

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"

	"github.com/andy-esch/taskflow/internal/config"
	"github.com/andy-esch/taskflow/internal/core"
	"github.com/andy-esch/taskflow/internal/spacehealth"
	"github.com/andy-esch/taskflow/internal/userconfig"
)

type FS struct{}

func New() *FS { return &FS{} }

var _ core.ConfigurationStore = (*FS)(nil)

func (f *FS) LoadConfiguration(start string) (core.ConfigurationState, error) {
	cfg, err := config.Discover(start)
	if err != nil {
		return core.ConfigurationState{}, err
	}
	repo := core.RepositoryConfiguration{
		Dir:          cfg.Dir,
		PlanningRoot: cfg.Root,
		ID:           cfg.ID,
		TrackedRepos: append([]string{}, cfg.TrackedRepos...),
		ThemeName:    cfg.Theme.Name,
		PagerEnabled: cloneBool(cfg.Pager.Enabled),
		PagerCommand: cfg.Pager.Command,
		Mode:         core.ConfigModeDiscovered,
	}
	if cfg.Dir != "" {
		repo.Path = filepath.Join(cfg.Dir, config.ConfigFile)
		description, _ := config.Describe(cfg.Dir)
		repo.TaskflowRoot = description.TaskflowRoot
		repo.PlanningRepo = description.PlanningRepo
		repo.PlanningRepoID = description.PlanningRepoID
		repo.Mode = core.ConfigModeScaffold
		if description.PlanningRepo != "" {
			repo.Mode = core.ConfigModePointer
		}
		pending, pendingErr := config.PendingMigrations(cfg.Dir)
		if pendingErr != nil {
			repo.MigrationWarning = pendingErr.Error()
		} else {
			for _, kind := range pending {
				repo.PendingMigration = append(repo.PendingMigration, toCoreMigrationKind(kind))
			}
		}
	}
	if repo.TrackedRepos == nil {
		repo.TrackedRepos = []string{}
	}
	if repo.PendingMigration == nil {
		repo.PendingMigration = []core.ConfigurationMigrationKind{}
	}

	userPath := ""
	registryPath := ""
	if dir, dirErr := userconfig.Dir(); dirErr == nil {
		userPath = filepath.Join(dir, userconfig.FileName)
		registryPath = filepath.Join(dir, userconfig.SpacesFile)
	}
	uc, userErr := userconfig.Load()
	user := core.UserConfiguration{
		Path: userPath, ThemeName: uc.Theme.Name,
		PagerEnabled: cloneBool(uc.Pager.Enabled), PagerCommand: uc.Pager.Command,
		RegistryPath: registryPath,
	}
	if uc.Path != "" {
		user.Path = uc.Path
		user.Exists = true
	}
	if userErr != nil {
		user.Warning = userErr.Error()
	}
	return core.ConfigurationState{Repository: repo, User: user}, nil
}

func (f *FS) MigrateConfiguration(start string, dryRun bool) (core.ConfigurationMigration, error) {
	result, err := config.Migrate(start, dryRun)
	if err != nil {
		return core.ConfigurationMigration{}, err
	}
	out := core.ConfigurationMigration{
		ConfigPath: result.ConfigPath,
		Mode:       core.ConfigMode(result.Mode),
		DryRun:     result.DryRun,
		Steps:      make([]core.ConfigurationMigrationStep, 0, len(result.Steps)),
	}
	for _, step := range result.Steps {
		out.Steps = append(out.Steps, core.ConfigurationMigrationStep{
			Kind: toCoreMigrationKind(step.Kind), Key: step.Key, Value: step.Value,
		})
	}
	return out, nil
}

func (f *FS) DiagnoseConfiguration(start string) (core.ConfigurationDiagnosis, error) {
	cfg, err := config.Discover(start)
	if err != nil {
		return core.ConfigurationDiagnosis{}, err
	}
	diagnosis := core.ConfigurationDiagnosis{Root: cfg.Root}
	if cfg.Dir != "" {
		diagnosis.ConfigPath = filepath.Join(cfg.Dir, config.ConfigFile)
	}
	if dir, dirErr := userconfig.Dir(); dirErr == nil {
		diagnosis.UserPath = filepath.Join(dir, userconfig.FileName)
		diagnosis.RegistryPath = filepath.Join(dir, userconfig.SpacesFile)
	}
	for _, problem := range config.CheckLinks(cfg) {
		diagnosis.Problems = append(diagnosis.Problems, core.ConfigurationProblem{
			Repo: problem.Repo, Message: problem.Message,
		})
	}
	spaces, err := spacehealth.DiagnoseRegistry()
	if err != nil {
		return core.ConfigurationDiagnosis{}, err
	}
	diagnosis.Registry.Checked = len(spaces)
	for _, space := range spaces {
		if !space.Broken() {
			continue
		}
		diagnosis.Registry.Problems = append(diagnosis.Registry.Problems, core.RegistryProblem{
			ID: space.Space.ID, Path: space.Space.Path, Kind: string(space.Kind),
			Message: space.Message, Remedy: space.Remedy,
		})
	}
	if diagnosis.Problems == nil {
		diagnosis.Problems = []core.ConfigurationProblem{}
	}
	if diagnosis.Registry.Problems == nil {
		diagnosis.Registry.Problems = []core.RegistryProblem{}
	}
	return diagnosis, nil
}

func (f *FS) SetPreference(start string, change core.PreferenceChange, dryRun bool) (core.PreferenceResult, error) {
	encoded, err := encodePreference(change)
	if err != nil {
		return core.PreferenceResult{}, err
	}
	var path string
	var changed bool
	if change.Scope == core.ConfigScopeRepository {
		path, changed, err = config.SetPresentation(start, repoPreferenceField(change.Field), encoded, dryRun)
	} else {
		path, changed, err = userconfig.SetPreference(userPreferenceField(change.Field), encoded, dryRun)
	}
	if err != nil {
		return core.PreferenceResult{}, err
	}
	return core.PreferenceResult{Path: path, Change: change, Changed: changed, DryRun: dryRun}, nil
}

func toCoreMigrationKind(kind config.MigrationKind) core.ConfigurationMigrationKind {
	switch kind {
	case config.MigrationRepoID:
		return core.ConfigurationMigrationRepoID
	case config.MigrationPlanningRepoID:
		return core.ConfigurationMigrationPlanningRepoID
	default:
		return core.ConfigurationMigrationKind(kind)
	}
}

func repoPreferenceField(field core.PreferenceField) config.PresentationField {
	switch field {
	case core.PreferenceTheme:
		return config.PresentationThemeName
	case core.PreferencePagerEnabled:
		return config.PresentationPagerEnabled
	case core.PreferencePagerCommand:
		return config.PresentationPagerCommand
	default:
		return config.PresentationField(field)
	}
}

func userPreferenceField(field core.PreferenceField) userconfig.PreferenceField {
	switch field {
	case core.PreferenceTheme:
		return userconfig.PreferenceThemeName
	case core.PreferencePagerEnabled:
		return userconfig.PreferencePagerEnabled
	case core.PreferencePagerCommand:
		return userconfig.PreferencePagerCommand
	default:
		return userconfig.PreferenceField(field)
	}
}

func encodePreference(change core.PreferenceChange) (*string, error) {
	if change.Unset {
		return nil, nil
	}
	if change.Field == core.PreferencePagerEnabled {
		value := change.Value
		return &value, nil
	}
	var b strings.Builder
	if err := toml.NewEncoder(&b).Encode(struct {
		Value string `toml:"value"`
	}{Value: change.Value}); err != nil {
		return nil, fmt.Errorf("encode %s: %w", change.Field, err)
	}
	line := strings.TrimSpace(b.String())
	_, value, ok := strings.Cut(line, "=")
	if !ok {
		return nil, fmt.Errorf("encode %s: missing value", change.Field)
	}
	value = strings.TrimSpace(value)
	return &value, nil
}

func cloneBool(v *bool) *bool {
	if v == nil {
		return nil
	}
	out := *v
	return &out
}
