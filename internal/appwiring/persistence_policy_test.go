package appwiring

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/andy-esch/taskflow/internal/config"
	"github.com/andy-esch/taskflow/internal/configstore"
	"github.com/andy-esch/taskflow/internal/core"
	"github.com/andy-esch/taskflow/internal/spacestore"
	"github.com/andy-esch/taskflow/internal/store"
	"github.com/andy-esch/taskflow/internal/workspacestore"
)

func TestPersistenceConstructorsRequireValidPoliciesWithoutIOOrAuthorization(t *testing.T) {
	root := filepath.Join(t.TempDir(), "not-created")
	discoveries, authorizations := 0, 0
	constructors := map[string]func(core.MutationPolicy) (bool, error){
		"planning":      func(p core.MutationPolicy) (bool, error) { s, err := store.NewFS(root, p); return s == nil, err },
		"configuration": func(p core.MutationPolicy) (bool, error) { s, err := configstore.New(p); return s == nil, err },
		"registry":      func(p core.MutationPolicy) (bool, error) { s, err := spacestore.New(p); return s == nil, err },
		"workspace":     func(p core.MutationPolicy) (bool, error) { s, err := workspacestore.New(p); return s == nil, err },
		"observed planning": func(p core.MutationPolicy) (bool, error) {
			s, err := workspacestore.NewPlanningStore(&config.Config{Root: root}, func(string) (*config.Config, error) {
				discoveries++
				return nil, errors.New("construction must not discover")
			}, p)
			return s == nil, err
		},
	}
	for name, construct := range constructors {
		t.Run(name, func(t *testing.T) {
			for _, p := range []core.MutationPolicy{{}, core.GuardedMutations(nil)} {
				if absent, err := construct(p); !absent || !errors.Is(err, core.ErrInvalidMutationPolicy) {
					t.Fatalf("invalid policy published adapter: absent=%v err=%v", absent, err)
				}
			}
			for _, p := range []core.MutationPolicy{core.ReadOnlyMutations(), core.UnrestrictedMutations(), core.GuardedMutations(func() error {
				authorizations++
				return errors.New("not bound yet")
			})} {
				if absent, err := construct(p); absent || err != nil {
					t.Fatalf("valid policy refused: absent=%v err=%v", absent, err)
				}
			}
		})
	}
	if authorizations != 0 || discoveries != 0 {
		t.Fatalf("construction invoked user code: authorizations=%d discoveries=%d", authorizations, discoveries)
	}
	if _, err := os.Stat(root); !os.IsNotExist(err) {
		t.Fatalf("construction touched root: %v", err)
	}
}

func TestInvalidLocalCompositionDoesNotPublishPartialServicesOrDiscover(t *testing.T) {
	reads := localSources{discover: func(string) (*config.Config, error) {
		t.Fatal("invalid composition discovered a repository")
		return nil, nil
	}}
	services, err := bindingsFor(reads).Compose(nil)
	if !errors.Is(err, core.ErrInvalidMutationPolicy) || services.Configuration != nil || services.Spaces != nil ||
		services.Overview != nil || services.Workspaces != nil || services.OpenPlanning != nil {
		t.Fatalf("invalid composition = %+v, %v", services, err)
	}
	if opened, err := openPlanning(reads, "unused", core.MutationPolicy{}); !errors.Is(err, core.ErrInvalidMutationPolicy) || opened.Service != nil {
		t.Fatalf("invalid deferred opening = %+v, %v", opened, err)
	}
}
