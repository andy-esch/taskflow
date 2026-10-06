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

type FS struct {
	mutationPolicy core.MutationPolicy
}

// New validates the explicit policy without invoking its authorizer or doing I/O.
func New(policy core.MutationPolicy) (*FS, error) {
	if err := policy.Validate(); err != nil {
		return nil, err
	}
	return &FS{mutationPolicy: policy}, nil
}

func (f *FS) authorizeMutation() error {
	return f.mutationPolicy.Authorize()
}

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
		Checkout: repoDir, VerifyID: cfg.ID,
	}, nil
}

func (f *FS) AddSpace(registration core.SpaceRegistration, dryRun bool) (core.SpaceEntryPoint, bool, error) {
	if err := f.authorizeMutation(); err != nil {
		return core.SpaceEntryPoint{}, false, err
	}
	added, existing, err := userconfig.AddSpace(userconfig.Space{
		ID: registration.ID, Path: userconfig.TildePath(registration.Checkout), VerifyID: registration.VerifyID,
	}, dryRun)
	if err != nil {
		return core.SpaceEntryPoint{}, false, classifyRegistryError(err)
	}
	return toCoreEntry(spacehealth.DiagnoseSpace(existing)), added, nil
}

func (f *FS) ForgetSpace(id string, dryRun bool) (core.SpaceEntryPoint, bool, error) {
	if err := f.authorizeMutation(); err != nil {
		return core.SpaceEntryPoint{}, false, err
	}
	removed, existing, err := userconfig.ForgetSpace(id, dryRun)
	if err != nil {
		return core.SpaceEntryPoint{}, false, classifyRegistryError(err)
	}
	if !removed {
		return core.SpaceEntryPoint{}, false, nil
	}
	return toCoreEntry(spacehealth.DiagnoseSpace(existing)), true, nil
}

func (f *FS) OpenPlanningStore(root string) (core.PlanningSummarySource, error) {
	if err := f.mutationPolicy.Validate(); err != nil {
		return nil, err
	}
	// Atlas summaries cannot inherit the registry's write privilege, even if a
	// caller widens the returned capability through a type assertion.
	fs, err := store.NewFS(root, core.ReadOnlyMutations())
	if err != nil {
		return nil, err
	}
	return fs, nil
}

func toCoreEntry(diagnosis spacehealth.SpaceProblem) core.SpaceEntryPoint {
	s := diagnosis.Space
	return core.SpaceEntryPoint{
		ID: s.ID, Path: s.Path, Checkout: entryCheckout(diagnosis),
		VerifyID: s.VerifyID, PlanningID: diagnosis.PlanningID,
		Role: core.SpaceRole(diagnosis.Role), Label: s.Label, Accent: s.Accent,
		Added: s.Added, State: core.SpaceState(diagnosis.Kind), Root: diagnosis.Root,
		Detail: diagnosis.Message, Remedy: diagnosis.Remedy,
	}
}

// entryCheckout is the path a consumer opens (or compares) this entry by. Discovery's
// resolved directory wins because it is symlink-evaluated exactly like the checkout on an
// opened core.Workspace, so the two compare equal; the mirrored Dir/Root fallback matches
// what workspacestore and the CLI already do. An entry too broken to discover keeps its
// tilde-expanded registry spelling, which is all that is known about it.
func entryCheckout(diagnosis spacehealth.SpaceProblem) string {
	if diagnosis.Dir != "" {
		return diagnosis.Dir
	}
	if diagnosis.Root != "" {
		return diagnosis.Root
	}
	return userconfig.ExpandTilde(diagnosis.Space.Path)
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
