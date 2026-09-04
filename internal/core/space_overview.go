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

// SpaceLoadFailure is adapter-neutral evidence that one logical planning tree could not
// be summarized. Class is retained separately from Message so a long-lived primary
// adapter can distinguish planner-window contention from durable failure without parsing
// storage-specific prose.
type SpaceLoadFailure struct {
	Class   domain.Class
	Message string
}

// SpaceSummary is the result for one logical planning identity. Selected is the healthy
// entry point used to read the summary. A fresh service result has either Summary or
// Failure. A long-lived caller may retain the last coherent Summary beside a contended
// Failure; Stale makes that deliberate reconciliation explicit.
type SpaceSummary struct {
	ID         string
	PlanningID string
	Entries    []SpaceEntryPoint
	Selected   *SpaceEntryPoint
	Summary    *Summary
	Failure    *SpaceLoadFailure
	Stale      bool
}

// Contended reports whether this summary hit the guarded mutation planner window.
func (s SpaceSummary) Contended() bool {
	return s.Failure != nil && s.Failure.Class == domain.ClassConflict
}

// ReconciliationKey is the stable, in-memory identity shared by overview merging and
// primary-adapter cursor restoration. It is not a persisted identifier: PlanningID is
// authoritative when available, with entry/summary labels only supporting legacy groups.
func (s SpaceSummary) ReconciliationKey() string {
	if s.PlanningID != "" {
		return "planning:" + s.PlanningID
	}
	if len(s.Entries) > 0 {
		return "entry:" + s.Entries[0].ID
	}
	return "summary:" + s.ID
}

// SpaceInProgress adds the local space address to a task in the combined working set.
// Stale is true only when the task came from a summary retained across contention.
type SpaceInProgress struct {
	SpaceID    string
	PlanningID string
	Task       domain.Task
	Stale      bool
}

// SpaceOverview is the read-only, cross-space dashboard projection.
type SpaceOverview struct {
	Spaces     []SpaceSummary
	InProgress []SpaceInProgress
}

// SpaceOverviewRefresh is a partial replacement set returned by RetryContended. It is a
// distinct type so callers cannot mistake the retried subset for a complete registry
// overview and accidentally drop independently healthy spaces.
type SpaceOverviewRefresh struct {
	Spaces []SpaceSummary
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
	spaces := make([]SpaceSummary, 0, len(catalog.Groups))
	asOf := s.now()
	for _, group := range catalog.Groups {
		spaces = append(spaces, summarizeSpaceGroup(group, s.store, asOf))
	}
	return spaceOverviewFromSummaries(spaces), nil
}

// RetryContended re-runs only the groups whose previous read hit the guarded planner
// window. It deliberately does not re-read the registry or unrelated healthy trees: the
// caller already owns one complete registry generation and will supersede this bounded
// retry with a later full refresh when necessary.
func (s *SpaceOverviewService) RetryContended(current SpaceOverview) SpaceOverviewRefresh {
	refresh := SpaceOverviewRefresh{Spaces: []SpaceSummary{}}
	asOf := s.now()
	for _, space := range current.Spaces {
		if !space.Contended() {
			continue
		}
		refresh.Spaces = append(refresh.Spaces, summarizeSpaceGroup(SpaceGroup{
			PlanningID: space.PlanningID,
			Entries:    append([]SpaceEntryPoint(nil), space.Entries...),
		}, s.store, asOf))
	}
	return refresh
}

// RetainContendedSpaceSummaries reconciles a complete fresh registry overview with the
// previous generation. Only planner-window contention may retain an older coherent
// summary; durable failures replace it and remain explicit.
func RetainContendedSpaceSummaries(previous, current SpaceOverview) SpaceOverview {
	prior := make(map[string]SpaceSummary, len(previous.Spaces))
	for _, space := range previous.Spaces {
		prior[space.ReconciliationKey()] = space
	}
	spaces := append([]SpaceSummary(nil), current.Spaces...)
	for i := range spaces {
		if before, ok := prior[spaces[i].ReconciliationKey()]; ok {
			spaces[i] = retainContendedSpaceSummary(before, spaces[i])
		}
	}
	return spaceOverviewFromSummaries(spaces)
}

// ApplySpaceOverviewRefresh applies a RetryContended subset without disturbing spaces
// outside that subset. A repeated conflict keeps the same coherent summary marked stale;
// success or a durable failure replaces it.
func ApplySpaceOverviewRefresh(current SpaceOverview, refresh SpaceOverviewRefresh) SpaceOverview {
	spaces := append([]SpaceSummary(nil), current.Spaces...)
	positions := make(map[string]int, len(spaces))
	for i, space := range spaces {
		positions[space.ReconciliationKey()] = i
	}
	for _, updated := range refresh.Spaces {
		if i, ok := positions[updated.ReconciliationKey()]; ok {
			spaces[i] = retainContendedSpaceSummary(spaces[i], updated)
		}
	}
	return spaceOverviewFromSummaries(spaces)
}

func retainContendedSpaceSummary(previous, current SpaceSummary) SpaceSummary {
	if !current.Contended() || previous.Summary == nil {
		return current
	}
	retained := cloneSpaceSummary(*previous.Summary)
	current.Summary = &retained
	current.Stale = true
	return current
}

// cloneSpaceSummary owns every mutable slice reachable from a retained Summary. Core
// projections are treated as immutable after construction, but making the last-coherent
// snapshot independent prevents a future adapter or reducer from changing old Atlas state
// through a shared slice backing array.
func cloneSpaceSummary(summary Summary) Summary {
	cloned := summary
	cloned.Counts = append([]StatusCount(nil), summary.Counts...)
	cloned.InProgress = make([]domain.Task, len(summary.InProgress))
	for i, task := range summary.InProgress {
		cloned.InProgress[i] = cloneTask(task)
	}
	cloned.Epics = append([]EpicSummary(nil), summary.Epics...)
	for i := range cloned.Epics {
		cloned.Epics[i].Epic.Tags = append([]string(nil), summary.Epics[i].Epic.Tags...)
	}
	cloned.OpenAudits = append([]domain.Audit(nil), summary.OpenAudits...)
	cloned.Findings.ByUrgency = append([]CountBy(nil), summary.Findings.ByUrgency...)
	cloned.Findings.ByComponent = append([]CountBy(nil), summary.Findings.ByComponent...)
	cloned.Findings.Acute = append([]AuditFinding(nil), summary.Findings.Acute...)
	cloned.Problems = append([]domain.FileProblem(nil), summary.Problems...)
	return cloned
}

func spaceOverviewFromSummaries(spaces []SpaceSummary) SpaceOverview {
	overview := SpaceOverview{
		Spaces:     append([]SpaceSummary(nil), spaces...),
		InProgress: []SpaceInProgress{},
	}
	for _, space := range overview.Spaces {
		if space.Summary == nil {
			continue
		}
		for _, task := range space.Summary.InProgress {
			overview.InProgress = append(overview.InProgress, SpaceInProgress{
				SpaceID: space.ID, PlanningID: space.PlanningID, Task: task, Stale: space.Stale,
			})
		}
	}
	return overview
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
		space.Failure = &SpaceLoadFailure{
			Class: domain.ClassNotFound, Message: "no healthy entry point",
		}
		return space
	}
	space.ID = selected.ID
	entry := *selected
	space.Selected = &entry

	planningStore, err := source.OpenPlanningStore(selected.Root)
	if err != nil {
		space.Failure = &SpaceLoadFailure{
			Class: domain.Classify(err), Message: fmt.Sprintf("open planning tree: %v", err),
		}
		return space
	}
	summary, err := summarize(planningStore, planningStore, asOf)
	if err != nil {
		space.Failure = &SpaceLoadFailure{Class: domain.Classify(err), Message: err.Error()}
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
