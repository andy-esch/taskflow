// Package spacestore is the composite filesystem secondary adapter for the
// cross-space overview use case. It translates the shared spacehealth projection into
// neutral core values and opens the existing Markdown Store for a selected planning root.
package spacestore

import (
	"github.com/andy-esch/taskflow/internal/core"
	"github.com/andy-esch/taskflow/internal/spacehealth"
	"github.com/andy-esch/taskflow/internal/store"
)

type FS struct{}

func New() *FS { return &FS{} }

var _ core.SpaceOverviewStore = (*FS)(nil)

func (f *FS) ListSpaceGroups() ([]core.SpaceGroup, error) {
	diagnoses, err := spacehealth.DiagnoseRegistry()
	if err != nil {
		return nil, err
	}
	groups := spacehealth.Group(diagnoses)
	out := make([]core.SpaceGroup, 0, len(groups))
	for _, group := range groups {
		converted := core.SpaceGroup{
			PlanningID: group.PlanningID,
			Entries:    make([]core.SpaceEntryPoint, 0, len(group.Entries)),
		}
		for _, diagnosis := range group.Entries {
			s := diagnosis.Space
			converted.Entries = append(converted.Entries, core.SpaceEntryPoint{
				ID: s.ID, Path: s.Path, VerifyID: s.VerifyID,
				PlanningID: diagnosis.PlanningID, Role: core.SpaceRole(diagnosis.Role), Label: s.Label,
				Added: s.Added,
				State: core.SpaceState(diagnosis.Kind), Root: diagnosis.Root,
				Detail: diagnosis.Message, Remedy: diagnosis.Remedy,
			})
		}
		out = append(out, converted)
	}
	return out, nil
}

func (f *FS) OpenPlanningStore(root string) (core.SummaryStore, error) {
	return store.NewFS(root), nil
}
