package core

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/andy-esch/taskflow/internal/domain"
)

// SpaceRegistryStore is the consumer-owned port for the machine-local registry. The
// adapter owns filesystem discovery, diagnosis, and persistence; the application service
// owns grouping, selection, labels, and mutation semantics.
type SpaceRegistryStore interface {
	ListSpaceEntries() ([]SpaceEntryPoint, error)
	PrepareSpace(path string) (SpaceRegistration, error)
	AddSpace(registration SpaceRegistration, dryRun bool) (SpaceEntryPoint, bool, error)
	ForgetSpace(id string, dryRun bool) (SpaceEntryPoint, bool, error)
}

// SpaceRegistryService is the framework-free application boundary shared by every
// registry consumer. It deliberately does not open or read planning trees.
type SpaceRegistryService struct {
	store SpaceRegistryStore
}

func NewSpaceRegistryService(store SpaceRegistryStore) *SpaceRegistryService {
	return &SpaceRegistryService{store: store}
}

// Catalog returns one diagnosed registry snapshot and its logical-space grouping.
func (s *SpaceRegistryService) Catalog() (SpaceCatalog, error) {
	entries, err := s.store.ListSpaceEntries()
	if err != nil {
		return SpaceCatalog{}, err
	}
	if entries == nil {
		entries = []SpaceEntryPoint{}
	}
	return SpaceCatalog{Entries: entries, Groups: groupSpaceEntries(entries)}, nil
}

// Resolve selects exactly one registered entry point by its machine-local label. Broken
// selections fail loudly so an explicit selection can never fall back to another tree.
func (s *SpaceRegistryService) Resolve(id string) (SpaceEntryPoint, error) {
	catalog, err := s.Catalog()
	if err != nil {
		return SpaceEntryPoint{}, err
	}
	var selected *SpaceEntryPoint
	for i := range catalog.Entries {
		entry := &catalog.Entries[i]
		if entry.ID != id {
			continue
		}
		if selected != nil {
			return SpaceEntryPoint{}, fmt.Errorf(
				"%w: invalid space registry: space %q appears more than once",
				domain.ErrValidation, id,
			)
		}
		selected = entry
	}
	if selected != nil {
		if !selected.Healthy() {
			return SpaceEntryPoint{}, selectedSpaceError(*selected)
		}
		return *selected, nil
	}

	known := make([]string, 0, len(catalog.Entries))
	for _, entry := range catalog.Entries {
		known = append(known, entry.ID)
	}
	sort.Strings(known)
	if len(known) == 0 {
		return SpaceEntryPoint{}, fmt.Errorf(
			"%w: unknown space %q — none are registered; run `space add`",
			domain.ErrNotFound, id,
		)
	}
	return SpaceEntryPoint{}, fmt.Errorf(
		"%w: unknown space %q — known: %s",
		domain.ErrNotFound, id, strings.Join(known, ", "),
	)
}

// Add validates a checkout before applying label policy or attempting a write. This
// preserves the public contract that a bad path leaves no registry behind.
func (s *SpaceRegistryService) Add(path, id string, dryRun bool) (SpaceMutation, error) {
	registration, err := s.store.PrepareSpace(path)
	if err != nil {
		return SpaceMutation{}, err
	}
	return s.add(registration, id, dryRun)
}

// RegisterInitialized registers a checkout that a preceding bootstrap use case has
// already validated. This keeps config bootstrapping independent of the home registry
// while still routing registry policy and mutation through this one application boundary.
func (s *SpaceRegistryService) RegisterInitialized(
	checkout, verifyID string, dryRun bool,
) (SpaceRegistrationReceipt, error) {
	mutation, err := s.add(SpaceRegistration{Checkout: checkout, VerifyID: verifyID}, "", dryRun)
	if err != nil {
		return SpaceRegistrationReceipt{}, err
	}
	// A physical-path match normally makes registration idempotent. For a freshly
	// initialized checkout, however, a different stored verification id means the path was
	// reused for a different planning identity. Never turn that safety failure into an
	// "already registered" receipt or silently rewrite the assertion; require the explicit
	// forget/add repair so the old identity cannot be rebound by bootstrap.
	if !mutation.Changed && verifyID != "" && mutation.Entry.VerifyID != verifyID {
		return SpaceRegistrationReceipt{}, fmt.Errorf(
			"%w: checkout %q is already registered as %q with verify_id %q, not %q; "+
				"run `space forget %q`, then `space add %q --id %q`",
			domain.ErrConflict, checkout, mutation.Entry.ID, mutation.Entry.VerifyID, verifyID,
			mutation.Entry.ID, checkout, mutation.Entry.ID,
		)
	}
	return SpaceRegistrationReceipt{
		ID: mutation.Entry.ID, Path: mutation.Entry.Path,
		VerifyID: mutation.Entry.VerifyID, Changed: mutation.Changed, DryRun: mutation.DryRun,
	}, nil
}

func (s *SpaceRegistryService) add(
	registration SpaceRegistration, id string, dryRun bool,
) (SpaceMutation, error) {
	if id == "" {
		id = defaultSpaceID(registration.Checkout)
	}
	if err := validateSpaceID(id); err != nil {
		return SpaceMutation{}, err
	}
	registration.ID = id
	entry, changed, err := s.store.AddSpace(registration, dryRun)
	if err != nil {
		return SpaceMutation{}, err
	}
	return SpaceMutation{Entry: entry, Changed: changed, DryRun: dryRun}, nil
}

// Forget drops only the registry row. A missing id is a not-found use-case error; the
// adapter remains responsible solely for the persistence operation and its receipt.
func (s *SpaceRegistryService) Forget(id string, dryRun bool) (SpaceMutation, error) {
	entry, changed, err := s.store.ForgetSpace(id, dryRun)
	if err != nil {
		return SpaceMutation{}, err
	}
	if !changed {
		return SpaceMutation{}, fmt.Errorf(
			"%w: no space named %q — `space list` shows the registered ones",
			domain.ErrNotFound, id,
		)
	}
	return SpaceMutation{Entry: entry, Changed: true, DryRun: dryRun}, nil
}

// groupSpaceEntries treats a durable planning id as the preferred key. Legacy healthy
// entries fall back to the resolved root; broken id-less entries remain isolated. Both
// group and entry order follow registry insertion order.
func groupSpaceEntries(entries []SpaceEntryPoint) []SpaceGroup {
	groups := make([]SpaceGroup, 0, len(entries))
	byKey := make(map[string]int, len(entries))
	for i, entry := range entries {
		key := spaceGroupKey(entry, i)
		groupIndex, exists := byKey[key]
		if !exists {
			groupIndex = len(groups)
			byKey[key] = groupIndex
			groups = append(groups, SpaceGroup{PlanningID: entry.PlanningID})
		}
		groups[groupIndex].Entries = append(groups[groupIndex].Entries, entry)
	}
	return groups
}

func spaceGroupKey(entry SpaceEntryPoint, index int) string {
	if entry.PlanningID != "" {
		return "id:" + entry.PlanningID
	}
	if entry.Root != "" {
		return "root:" + entry.Root
	}
	return fmt.Sprintf("entry:%d:%s", index, entry.ID)
}

func selectedSpaceError(entry SpaceEntryPoint) error {
	sentinel := domain.ErrNotFound
	if entry.State == SpaceStateMismatch {
		sentinel = domain.ErrConflict
	}
	message := fmt.Sprintf("registered space %q: %s", entry.ID, entry.Detail)
	if entry.Remedy != "" {
		message += "; " + entry.Remedy
	}
	return fmt.Errorf("%w: %s", sentinel, message)
}

func defaultSpaceID(dir string) string {
	return strings.ToLower(filepath.Base(filepath.Clean(dir)))
}

func validateSpaceID(id string) error {
	if id == "" {
		return fmt.Errorf("%w: a space needs a label; pass --id", domain.ErrValidation)
	}
	if strings.ContainsAny(id, " \t/\\\"'") {
		return fmt.Errorf(
			"%w: space label %q may not contain spaces, quotes or path separators",
			domain.ErrValidation, id,
		)
	}
	return nil
}
