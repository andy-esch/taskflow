package core

import (
	"fmt"
	"strings"

	"github.com/andy-esch/taskflow/internal/domain"
)

// WorkspaceStore is the secondary-adapter port for resolving and opening one local
// planning workspace. The adapter owns repo discovery and concrete persistence; this
// application layer owns the runtime workspace handed to a primary adapter.
type WorkspaceStore interface {
	OpenWorkspace(start string) (WorkspaceSource, error)
}

// WorkspaceSource is the adapter-neutral result of local discovery. Store and Layout
// are intentionally separate capabilities even when one concrete filesystem value
// implements both: entity use cases must not grow knowledge of watcher paths.
type WorkspaceSource struct {
	Checkout     string
	PlanningRoot string
	PlanningID   string
	Store        Store
	Layout       Layout
}

// WorkspaceRequest identifies an explicit local entry point. SpaceID is presentation
// context supplied by registry navigation; ExpectedPlanningID is the optional durable
// assertion rechecked after discovery. Neither value chooses the tree: the local start
// path remains authoritative and the home registry remains advisory.
type WorkspaceRequest struct {
	Start              string
	SpaceID            string
	ExpectedPlanningID string
}

// Workspace is one opened planning context ready for a primary adapter. Planning owns
// entity use cases while Layout exposes only the paths needed for live reload.
type Workspace struct {
	SpaceID      string
	Checkout     string
	PlanningRoot string
	PlanningID   string
	Planning     *Service
	Layout       Layout
}

// WorkspaceService opens arbitrary local planning trees without exposing discovery or
// filesystem adapters to the TUI (or a future served primary adapter).
type WorkspaceService struct {
	store WorkspaceStore
}

func NewWorkspaceService(store WorkspaceStore) *WorkspaceService {
	return &WorkspaceService{store: store}
}

// Open resolves request.Start and assembles the neutral runtime workspace. An explicit
// non-empty start is required: silently treating an empty registry path as cwd would make
// a failed atlas selection fall back to whichever tree launched the process.
func (s *WorkspaceService) Open(request WorkspaceRequest) (Workspace, error) {
	if strings.TrimSpace(request.Start) == "" {
		return Workspace{}, fmt.Errorf("%w: workspace start path is required", domain.ErrValidation)
	}
	if s == nil || s.store == nil {
		return Workspace{}, fmt.Errorf("%w: workspace opener is unavailable", domain.ErrValidation)
	}
	source, err := s.store.OpenWorkspace(request.Start)
	if err != nil {
		return Workspace{}, err
	}
	if source.Store == nil || source.Layout == nil || source.PlanningRoot == "" {
		return Workspace{}, fmt.Errorf("%w: workspace adapter returned incomplete capabilities", domain.ErrValidation)
	}
	// Registry diagnosis is necessarily a snapshot. Recheck its durable identity after
	// discovery so a checkout replaced or repointed between atlas load and Enter cannot
	// silently open a different planning tree under the selected space label. Legacy
	// id-less trees remain navigable because there is no durable assertion to compare.
	if request.ExpectedPlanningID != "" && source.PlanningID != request.ExpectedPlanningID {
		return Workspace{}, fmt.Errorf(
			"%w: workspace planning id %q no longer matches selected id %q",
			domain.ErrConflict, source.PlanningID, request.ExpectedPlanningID,
		)
	}
	return Workspace{
		SpaceID:      request.SpaceID,
		Checkout:     source.Checkout,
		PlanningRoot: source.PlanningRoot,
		PlanningID:   source.PlanningID,
		Planning:     NewService(source.Store),
		Layout:       source.Layout,
	}, nil
}
