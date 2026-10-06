// Package workspacestore is the local-filesystem secondary adapter for opening one
// planning workspace. It owns the translation from repo-scoped discovery and the
// concrete Markdown store into the neutral core workspace source.
package workspacestore

import (
	"github.com/andy-esch/taskflow/internal/config"
	"github.com/andy-esch/taskflow/internal/core"
)

type FS struct {
	mutationPolicy core.MutationPolicy
}

// New validates the policy carried into every later-opened planning store.
// It performs no discovery or invocation authorization.
func New(policy core.MutationPolicy) (*FS, error) {
	if err := policy.Validate(); err != nil {
		return nil, err
	}
	return &FS{mutationPolicy: policy}, nil
}

var _ core.WorkspaceStore = (*FS)(nil)

func (f *FS) OpenWorkspace(start string) (core.WorkspaceSource, error) {
	if err := f.mutationPolicy.Validate(); err != nil {
		return core.WorkspaceSource{}, err
	}
	cfg, err := config.Discover(start)
	if err != nil {
		return core.WorkspaceSource{}, err
	}
	fs, err := NewPlanningStore(cfg, config.Discover, f.mutationPolicy)
	if err != nil {
		return core.WorkspaceSource{}, err
	}
	checkout := cfg.Dir
	if checkout == "" {
		checkout = cfg.Root
	}
	return core.WorkspaceSource{
		Checkout: checkout, PlanningRoot: cfg.Root, PlanningID: cfg.ID,
		Store: fs, TaskGraphs: fs, TaskPaths: fs, EpicPaths: fs, AuditPaths: fs,
		ResearchPaths: fs, Threads: fs, ThreadPaths: fs, Layout: fs,
	}, nil
}
