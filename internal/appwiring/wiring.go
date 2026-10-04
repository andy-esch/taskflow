// Package appwiring is the binary's local-adapter composition boundary.
// CLI controllers depend on their own ports, never on this package. Every
// command tree receives a fresh application bundle and authorization closure.
package appwiring

import (
	"github.com/andy-esch/taskflow/internal/cli/ports"
	"github.com/andy-esch/taskflow/internal/configstore"
	"github.com/andy-esch/taskflow/internal/core"
	"github.com/andy-esch/taskflow/internal/design"
	"github.com/andy-esch/taskflow/internal/spacestore"
	"github.com/andy-esch/taskflow/internal/store"
	"github.com/andy-esch/taskflow/internal/workspacestore"
)

func bindingsFor(reads localSources) ports.Bindings {
	return ports.Bindings{
		Compose:           func(authorize func() error) ports.Services { return compose(reads, authorize) },
		IsMissingPlanning: isMissingPlanning,
		RunBrowser:        runBrowser,
		RunConfiguration:  runConfiguration,
		ReadRepository:    reads.readRepository,
		ReadUser:          reads.readUser,
	}
}

func compose(reads localSources, authorize func() error) ports.Services {
	spaceAdapter := spacestore.New(spacestore.WithMutationAuthorization(authorize))
	spaces := core.NewSpaceRegistryService(spaceAdapter)
	return ports.Services{
		Configuration: core.NewConfigurationService(
			configstore.New(configstore.WithMutationAuthorization(authorize)),
			core.WithConfigurationThemes(design.Names()), core.WithSpaceRegistry(spaces)),
		Spaces:   spaces,
		Overview: core.NewSpaceOverviewService(spaces, spaceAdapter),
		Workspaces: core.NewWorkspaceService(
			workspacestore.New(workspacestore.WithMutationAuthorization(authorize))),
		OpenPlanning: func(start string) (ports.Planning, error) {
			return openPlanning(reads, start, authorize)
		},
	}
}

func openPlanning(reads localSources, start string, authorize func() error) (ports.Planning, error) {
	cfg, err := reads.discover(start)
	if err != nil {
		return ports.Planning{}, err
	}
	discoveryStart := cfg.Dir
	if discoveryStart == "" {
		discoveryStart = cfg.Root
	}
	fs := store.NewFS(cfg.Root, store.WithPlanningIdentityReader(func() (string, string, error) {
		fresh, err := reads.discover(discoveryStart)
		if err != nil {
			return "", "", err
		}
		return fresh.Root, fresh.ID, nil
	}), store.WithMutationAuthorization(authorize))
	planning, err := core.NewService(fs)
	if err != nil {
		return ports.Planning{}, err
	}
	return ports.Planning{Repository: repositorySettings(cfg), Service: planning, Layout: fs}, nil
}
