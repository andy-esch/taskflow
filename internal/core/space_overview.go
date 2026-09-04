package core

import (
	"fmt"
	"time"

	"github.com/andy-esch/taskflow/internal/domain"
)

// PlanningSummarySource is the complete read capability for one planning-space
// dashboard. TaskGraphSource is explicit rather than discovered with a type
// assertion: every adapter must provide the same strict task snapshot contract,
// including record identity for unreadable tasks.
type PlanningSummarySource interface {
	SummaryStore
	TaskGraphSource
}

// Pin both halves of the composite capability independently. This keeps a future
// refactor from replacing either compile-time requirement with optional runtime
// discovery while leaving the immediate summarize call buildable.
var (
	_ SummaryStore    = (PlanningSummarySource)(nil)
	_ TaskGraphSource = (PlanningSummarySource)(nil)
)

// SpaceOverviewStore is the narrow secondary-adapter port for opening one planning tree.
// Registry cataloging is deliberately supplied by SpaceRegistryService instead of being
// folded into this filesystem capability.
type SpaceOverviewStore interface {
	OpenPlanningStore(root string) (PlanningSummarySource, error)
}

// SpaceSummary is the result for one logical planning identity. Selected is the healthy
// entry point used to read the summary. Summary is nil when no entry point could be read;
// LoadError explains that group-local failure without failing the complete overview.
type SpaceSummary struct {
	ID         string
	PlanningID string
	Entries    []SpaceEntryPoint
	Selected   *SpaceEntryPoint
	Summary    *Summary
	LoadError  string
}

// SpaceInProgress adds the local space address to a task in the combined working set.
type SpaceInProgress struct {
	SpaceID    string
	PlanningID string
	Task       domain.Task
}

// SpaceOverview is the read-only, cross-space dashboard projection.
type SpaceOverview struct {
	Spaces     []SpaceSummary
	InProgress []SpaceInProgress
}

// SpaceOverviewService owns cross-space orchestration over the registry and planning
// stores. It is deliberately separate from Service: a single-tree Service remains usable
// without any home registry, while this use case works from anywhere on the machine.
type SpaceOverviewService struct {
	registry *SpaceRegistryService
	store    SpaceOverviewStore
	now      func() time.Time
}

func NewSpaceOverviewService(registry *SpaceRegistryService, store SpaceOverviewStore) *SpaceOverviewService {
	return &SpaceOverviewService{registry: registry, store: store, now: time.Now}
}

// Overview returns one summary per logical planning identity. Registry decode/read
// errors are fatal because no trustworthy group list exists. Entry-point and per-tree
// read failures stay inside their group so one dead checkout never hides healthy spaces.
func (s *SpaceOverviewService) Overview() (SpaceOverview, error) {
	catalog, err := s.registry.Catalog()
	if err != nil {
		return SpaceOverview{}, err
	}
	overview := SpaceOverview{
		Spaces:     make([]SpaceSummary, 0, len(catalog.Groups)),
		InProgress: []SpaceInProgress{},
	}
	asOf := s.now()
	for _, group := range catalog.Groups {
		space := summarizeSpaceGroup(group, s.store, asOf)
		overview.Spaces = append(overview.Spaces, space)
		if space.Summary == nil {
			continue
		}
		for _, task := range space.Summary.InProgress {
			overview.InProgress = append(overview.InProgress, SpaceInProgress{
				SpaceID: space.ID, PlanningID: space.PlanningID, Task: task,
			})
		}
	}
	return overview, nil
}

func summarizeSpaceGroup(group SpaceGroup, source SpaceOverviewStore, asOf time.Time) SpaceSummary {
	space := SpaceSummary{
		PlanningID: group.PlanningID,
		Entries:    append([]SpaceEntryPoint{}, group.Entries...),
	}
	selected := preferredHealthyEntry(group.Entries)
	if selected == nil {
		if len(group.Entries) > 0 {
			space.ID = group.Entries[0].ID
		}
		space.LoadError = "no healthy entry point"
		return space
	}
	space.ID = selected.ID
	entry := *selected
	space.Selected = &entry

	planningStore, err := source.OpenPlanningStore(selected.Root)
	if err != nil {
		space.LoadError = fmt.Sprintf("open planning tree: %v", err)
		return space
	}
	summary, err := summarize(planningStore, planningStore, asOf)
	if err != nil {
		space.LoadError = err.Error()
		return space
	}
	space.Summary = &summary
	return space
}

// preferredHealthyEntry implements the shared identity's routing policy: a healthy
// direct planning checkout wins even when it was registered after a pointer; otherwise
// use the first healthy entry in registry order.
func preferredHealthyEntry(entries []SpaceEntryPoint) *SpaceEntryPoint {
	var first *SpaceEntryPoint
	for i := range entries {
		entry := &entries[i]
		if !entry.Healthy() {
			continue
		}
		if first == nil {
			first = entry
		}
		if entry.Role == SpaceRoleDirect {
			return entry
		}
	}
	return first
}
