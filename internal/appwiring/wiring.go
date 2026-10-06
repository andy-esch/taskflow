// Package appwiring is the binary's local-adapter composition boundary.
// CLI controllers depend on their own ports, never on this package. Every
// command tree receives a fresh application bundle and authorization closure.
// Preserve lazy opening, source-set identity, and policy propagation when adding
// bindings; see docs/ARCHITECTURE.md#planning-data-change-checklist and core port contracts.
package appwiring

import (
	"github.com/andy-esch/taskflow/internal/cli/ports"
	"github.com/andy-esch/taskflow/internal/configstore"
	"github.com/andy-esch/taskflow/internal/core"
	"github.com/andy-esch/taskflow/internal/design"
	"github.com/andy-esch/taskflow/internal/spacestore"
	"github.com/andy-esch/taskflow/internal/workspacestore"
)

func bindingsFor(reads localSources) ports.Bindings {
	return ports.Bindings{
		Compose:           func(authorize func() error) (ports.Services, error) { return compose(reads, authorize) },
		IsMissingPlanning: isMissingPlanning,
		RunBrowser:        runBrowser,
		RunConfiguration:  runConfiguration,
		ReadRepository:    reads.readRepository,
		ReadUser:          reads.readUser,
	}
}

func compose(reads localSources, authorize func() error) (ports.Services, error) {
	policy := core.GuardedMutations(authorize)
	spaceAdapter, err := spacestore.New(policy)
	if err != nil {
		return ports.Services{}, err
	}
	configuration, err := configstore.New(policy)
	if err != nil {
		return ports.Services{}, err
	}
	workspaces, err := workspacestore.New(policy)
	if err != nil {
		return ports.Services{}, err
	}
	spaces := core.NewSpaceRegistryService(spaceAdapter)
	return ports.Services{
		Configuration: core.NewConfigurationService(
			configuration,
			core.WithConfigurationThemes(design.Names()), core.WithSpaceRegistry(spaces)),
		Spaces:   spaces,
		Overview: core.NewSpaceOverviewService(spaces, spaceAdapter),
		Workspaces: core.NewWorkspaceService(
			workspaces),
		OpenPlanning: func(start string) (ports.Planning, error) {
			return openPlanning(reads, start, policy)
		},
	}, nil
}

func openPlanning(reads localSources, start string, policy core.MutationPolicy) (ports.Planning, error) {
	if err := policy.Validate(); err != nil {
		return ports.Planning{}, err
	}
	cfg, err := reads.discover(start)
	if err != nil {
		return ports.Planning{}, err
	}
	fs, err := workspacestore.NewPlanningStore(cfg, reads.discover, policy)
	if err != nil {
		return ports.Planning{}, err
	}
	planning, err := core.NewService(fs)
	if err != nil {
		return ports.Planning{}, err
	}
	return ports.Planning{Repository: repositorySettings(cfg), Service: planning, Layout: fs}, nil
}
