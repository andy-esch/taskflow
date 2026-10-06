package workspacestore

import (
	"fmt"

	"github.com/andy-esch/taskflow/internal/config"
	"github.com/andy-esch/taskflow/internal/core"
	"github.com/andy-esch/taskflow/internal/domain"
	"github.com/andy-esch/taskflow/internal/store"
)

// NewPlanningStore binds an already-discovered corpus to its identity re-reader
// and invocation authorization. Construction performs no discovery: ordinary CLI
// and workspace opening must publish capabilities from the same initial observation.
// The re-reader stays anchored at the discovered marker (or the legacy root), not
// cwd or the resolved target of a pointer, so later repointing remains observable.
func NewPlanningStore(cfg *config.Config, discover func(string) (*config.Config, error), policy core.MutationPolicy) (*store.FS, error) {
	if err := policy.Validate(); err != nil {
		return nil, err
	}
	if cfg == nil || cfg.Root == "" {
		return nil, fmt.Errorf("%w: discovered planning root is required", domain.ErrValidation)
	}
	if discover == nil {
		return nil, fmt.Errorf("%w: planning identity discovery is required", domain.ErrValidation)
	}
	discoveryStart := cfg.Dir
	if discoveryStart == "" {
		discoveryStart = cfg.Root
	}
	return store.NewFS(cfg.Root, policy, store.WithPlanningIdentityReader(func() (string, string, error) {
		fresh, err := discover(discoveryStart)
		if err != nil {
			return "", "", err
		}
		if fresh == nil || fresh.Root == "" {
			return "", "", fmt.Errorf("%w: planning identity discovery returned no root", domain.ErrValidation)
		}
		return fresh.Root, fresh.ID, nil
	}))
}
