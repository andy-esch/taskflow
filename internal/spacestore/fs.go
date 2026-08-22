// Package spacestore is the filesystem secondary adapter for the home registry and
// cross-space planning-tree opener. It is the one translation boundary from
// userconfig/spacehealth values into the application core.
package spacestore

import (
	"errors"
	"fmt"
	"path/filepath"

	"github.com/andy-esch/taskflow/internal/config"
	"github.com/andy-esch/taskflow/internal/core"
	"github.com/andy-esch/taskflow/internal/domain"
	"github.com/andy-esch/taskflow/internal/spacehealth"
	"github.com/andy-esch/taskflow/internal/store"
	"github.com/andy-esch/taskflow/internal/userconfig"
)

type FS struct{}

func New() *FS { return &FS{} }

var (
	_ core.SpaceRegistryStore = (*FS)(nil)
	_ core.SpaceOverviewStore = (*FS)(nil)
)

func (f *FS) ListSpaceEntries() ([]core.SpaceEntryPoint, error) {
	diagnoses, err := spacehealth.DiagnoseRegistry()
	if err != nil {
		return nil, classifyRegistryError(err)
	}
	out := make([]core.SpaceEntryPoint, 0, len(diagnoses))
	for _, diagnosis := range diagnoses {
		out = append(out, toCoreEntry(diagnosis))
	}
	return out, nil
}

func (f *FS) PrepareSpace(path string) (core.SpaceRegistration, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return core.SpaceRegistration{}, err
	}
	cfg, err := config.Discover(abs)
	if err != nil {
		return core.SpaceRegistration{}, err
	}
	// Preserve the entry point rather than the resolved planning root: pointer repos
	// remain independently addressable and discovery performs their routing later.
	repoDir := cfg.Dir
	if repoDir == "" {
		repoDir = cfg.Root
	}
	return core.SpaceRegistration{
		Path: userconfig.TildePath(repoDir), VerifyID: cfg.ID,
	}, nil
}

func (f *FS) AddSpace(registration core.SpaceRegistration, dryRun bool) (core.SpaceEntryPoint, bool, error) {
	added, existing, err := userconfig.AddSpace(userconfig.Space{
		ID: registration.ID, Path: registration.Path, VerifyID: registration.VerifyID,
	}, dryRun)
	if err != nil {
		return core.SpaceEntryPoint{}, false, classifyRegistryError(err)
	}
	return toCoreEntry(spacehealth.DiagnoseSpace(existing)), added, nil
}

func (f *FS) ForgetSpace(id string, dryRun bool) (core.SpaceEntryPoint, bool, error) {
	removed, existing, err := userconfig.ForgetSpace(id, dryRun)
	if err != nil {
		return core.SpaceEntryPoint{}, false, classifyRegistryError(err)
	}
	if !removed {
		return core.SpaceEntryPoint{}, false, nil
	}
	return toCoreEntry(spacehealth.DiagnoseSpace(existing)), true, nil
}

func (f *FS) OpenPlanningStore(root string) (core.SummaryStore, error) {
	return store.NewFS(root), nil
}

func toCoreEntry(diagnosis spacehealth.SpaceProblem) core.SpaceEntryPoint {
	s := diagnosis.Space
	return core.SpaceEntryPoint{
		ID: s.ID, Path: s.Path, Checkout: userconfig.ExpandTilde(s.Path),
		VerifyID: s.VerifyID, PlanningID: diagnosis.PlanningID,
		Role: core.SpaceRole(diagnosis.Role), Label: s.Label, Accent: s.Accent,
		Added: s.Added, State: core.SpaceState(diagnosis.Kind), Root: diagnosis.Root,
		Detail: diagnosis.Message, Remedy: diagnosis.Remedy,
	}
}

func classifyRegistryError(err error) error {
	switch {
	case errors.Is(err, userconfig.ErrSpaceIDConflict):
		return fmt.Errorf("%w: %s", domain.ErrConflict, err.Error())
	case errors.Is(err, userconfig.ErrInvalidRegistry):
		return fmt.Errorf("%w: %s", domain.ErrValidation, err.Error())
	default:
		return err
	}
}
