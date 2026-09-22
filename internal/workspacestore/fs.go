// Package workspacestore is the local-filesystem secondary adapter for opening one
// planning workspace. It owns the translation from repo-scoped discovery and the
// concrete Markdown store into the neutral core workspace source.
package workspacestore

import (
	"github.com/andy-esch/taskflow/internal/config"
	"github.com/andy-esch/taskflow/internal/core"
	"github.com/andy-esch/taskflow/internal/store"
)

type FS struct {
	mutationAuthorization func() error
}

type Option func(*FS)

// WithMutationAuthorization carries a primary adapter's mutation policy into
// every planning store opened through this workspace boundary.
func WithMutationAuthorization(authorize func() error) Option {
	return func(store *FS) { store.mutationAuthorization = authorize }
}

func New(opts ...Option) *FS {
	workspaceStore := &FS{}
	for _, opt := range opts {
		opt(workspaceStore)
	}
	return workspaceStore
}

var _ core.WorkspaceStore = (*FS)(nil)

func (f *FS) OpenWorkspace(start string) (core.WorkspaceSource, error) {
	cfg, err := config.Discover(start)
	if err != nil {
		return core.WorkspaceSource{}, err
	}
	fs := store.NewFS(cfg.Root, store.WithMutationAuthorization(f.mutationAuthorization))
	checkout := cfg.Dir
	if checkout == "" {
		checkout = cfg.Root
	}
	return core.WorkspaceSource{
		Checkout: checkout, PlanningRoot: cfg.Root, PlanningID: cfg.ID,
		Store: fs, TaskGraphs: fs, Threads: fs, ThreadPaths: fs, Layout: fs,
	}, nil
}
