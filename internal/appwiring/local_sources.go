package appwiring

import (
	"errors"
	"path/filepath"

	"github.com/andy-esch/taskflow/internal/cli/ports"
	"github.com/andy-esch/taskflow/internal/config"
	"github.com/andy-esch/taskflow/internal/core"
	"github.com/andy-esch/taskflow/internal/userconfig"
)

// localSources carries startup reads per bindings value, not through mutable
// package-global hooks. Tests wrap these same real readers to observe laziness
// and discovery counts without replacing the actual adapter composition.
// Configuration/registry use cases retain their separate secondary adapters.
type localSources struct {
	discover func(string) (*config.Config, error)
	user     func() (*userconfig.Config, error)
}

// LocalBindings selects the local Markdown/TOML adapters without reading or
// writing anything. Composition and discovery happen at their explicit hooks.
func LocalBindings() ports.Bindings {
	return bindingsFor(localSources{discover: config.Discover, user: userconfig.Load})
}

func isMissingPlanning(err error) bool { return errors.Is(err, config.ErrNoConfig) }

func (reads localSources) readRepository(start string) (core.RepositoryConfiguration, error) {
	cfg, err := reads.discover(start)
	if err != nil {
		return core.RepositoryConfiguration{}, err
	}
	return repositorySettings(cfg), nil
}

func (reads localSources) readUser() (core.UserConfiguration, error) {
	uc, err := reads.user()
	return core.UserConfiguration{
		Path: uc.Path, Exists: uc.Path != "", ThemeName: uc.Theme.Name,
		PagerEnabled: cloneBool(uc.Pager.Enabled), PagerCommand: uc.Pager.Command,
	}, err
}

// Startup needs observed identity and presentation settings only. Full config
// inspection/migration remains ConfigurationService's separate use case; do not
// add another Describe/PendingMigrations/home-registry scan to ordinary startup.
func repositorySettings(cfg *config.Config) core.RepositoryConfiguration {
	repo := core.RepositoryConfiguration{
		Dir: cfg.Dir, PlanningRoot: cfg.Root, ID: cfg.ID, PlanningRepo: cfg.PlanningRepo,
		TrackedRepos: append([]string(nil), cfg.TrackedRepos...), ThemeName: cfg.Theme.Name,
		PagerEnabled: cloneBool(cfg.Pager.Enabled), PagerCommand: cfg.Pager.Command,
		Mode: core.ConfigModeDiscovered,
	}
	if cfg.Dir != "" {
		repo.Path = filepath.Join(cfg.Dir, config.ConfigFile)
		repo.Mode = core.ConfigModeScaffold
		if cfg.PlanningRepo != "" {
			repo.Mode = core.ConfigModePointer
		}
	}
	return repo
}

func cloneBool(value *bool) *bool {
	if value == nil {
		return nil
	}
	copy := *value
	return &copy
}
